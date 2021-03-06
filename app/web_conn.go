// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"bitbucket.org/enesyteam/papo-server/mlog"
	"bitbucket.org/enesyteam/papo-server/model"

	"github.com/gorilla/websocket"
	goi18n "github.com/mattermost/go-i18n/i18n"
)

const (
	sendQueueSize          = 256
	sendSlowWarn           = (sendQueueSize * 50) / 100
	sendFullWarn           = (sendQueueSize * 95) / 100
	writeWaitTime          = 30 * time.Second
	pongWaitTime           = 100 * time.Second
	pingInterval           = (pongWaitTime * 6) / 10
	authCheckInterval      = 5 * time.Second
	webConnMemberCacheTime = 1000 * 60 * 30 // 30 minutes
)

// WebConn represents a single websocket connection to a user.
// It contains all the necesarry state to manage sending/receiving data to/from
// a websocket.
type WebConn struct {
	sessionExpiresAt 		int64 // This should stay at the top for 64-bit alignment of 64-bit words accessed atomically
	App              		*App
	WebSocket        		*websocket.Conn
	T                		goi18n.TranslateFunc
	Locale           		string
	Sequence         		int64
	UserId           		string

	AllPageMembers         	map[string]string
	LastAllPageMembersTime 	int64
	lastUserActivityAt        int64
	send                      chan model.WebSocketMessage
	sessionToken              atomic.Value
	session                   atomic.Value
	endWritePump              chan struct{}
	pumpFinished              chan struct{}
	PageIds 			   	[]string
}

// NewWebConn returns a new WebConn instance.
func (a *App) NewWebConn(ws *websocket.Conn, session model.Session, t goi18n.TranslateFunc, locale string) *WebConn {
	if session.UserId != "" {
		a.Srv().Go(func() {
			a.SetStatusOnline(session.UserId, false)
			a.UpdateLastActivityAtIfNeeded(session)
		})
	}

	wc := &WebConn{
		App:                a,
		send:               make(chan model.WebSocketMessage, sendQueueSize),
		WebSocket:          ws,
		lastUserActivityAt: model.GetMillis(),
		UserId:             session.UserId,
		T:                  t,
		Locale:             locale,
		endWritePump:       make(chan struct{}),
		pumpFinished:       make(chan struct{}),
	}

	wc.SetSession(&session)
	wc.SetSessionToken(session.Token)
	wc.SetSessionExpiresAt(session.ExpiresAt)

	return wc
}

// Close closes the WebConn.
func (wc *WebConn) Close() {
	wc.WebSocket.Close()
	<-wc.pumpFinished
}

// GetSessionExpiresAt returns the time at which the session expires.
func (wc *WebConn) GetSessionExpiresAt() int64 {
	return atomic.LoadInt64(&wc.sessionExpiresAt)
}

// SetSessionExpiresAt sets the time at which the session expires.
func (wc *WebConn) SetSessionExpiresAt(v int64) {
	atomic.StoreInt64(&wc.sessionExpiresAt, v)
}

// GetSessionToken returns the session token of the connection.
func (wc *WebConn) GetSessionToken() string {
	return wc.sessionToken.Load().(string)
}

// SetSessionToken sets the session token of the connection.
func (wc *WebConn) SetSessionToken(v string) {
	wc.sessionToken.Store(v)
}

// GetSession returns the session of the connection.
func (wc *WebConn) GetSession() *model.Session {
	return wc.session.Load().(*model.Session)
}

// SetSession sets the session of the connection.
func (wc *WebConn) SetSession(v *model.Session) {
	if v != nil {
		v = v.DeepCopy()
	}

	wc.session.Store(v)
}

// Pump starts the WebConn instance. After this, the websocket
// is ready to send/receive messages.
func (wc *WebConn) Pump() {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		wc.writePump()
	}()
	wc.readPump()
	close(wc.endWritePump)
	wg.Wait()
	wc.App.HubUnregister(wc)
	close(wc.pumpFinished)
}

func (wc *WebConn) readPump() {
	defer func() {
		wc.WebSocket.Close()
	}()
	wc.WebSocket.SetReadLimit(model.SOCKET_MAX_MESSAGE_SIZE_KB)
	wc.WebSocket.SetReadDeadline(time.Now().Add(pongWaitTime))
	wc.WebSocket.SetPongHandler(func(string) error {
		wc.WebSocket.SetReadDeadline(time.Now().Add(pongWaitTime))
		if wc.IsAuthenticated() {
			wc.App.Srv().Go(func() {
				wc.App.SetStatusAwayIfNeeded(wc.UserId, false)
			})
		}
		return nil
	})

	for {
		var req model.WebSocketRequest
		if err := wc.WebSocket.ReadJSON(&req); err != nil {
			wc.logSocketErr("websocket.read", err)
			return
		}
		wc.App.Srv().WebSocketRouter.ServeWebSocket(wc, &req)
	}
}

func (wc *WebConn) writePump() {
	ticker := time.NewTicker(pingInterval)
	authTicker := time.NewTicker(authCheckInterval)

	defer func() {
		ticker.Stop()
		authTicker.Stop()
		wc.WebSocket.Close()
	}()

	for {
		select {
		case msg, ok := <-wc.send:
			if !ok {
				wc.WebSocket.SetWriteDeadline(time.Now().Add(writeWaitTime))
				wc.WebSocket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			evt, evtOk := msg.(*model.WebSocketEvent)

			skipSend := false
			if len(wc.send) >= sendSlowWarn {
				// When the pump starts to get slow we'll drop non-critical messages
				switch msg.EventType() {
				case model.WEBSOCKET_EVENT_TYPING,
					model.WEBSOCKET_EVENT_STATUS_CHANGE,
					model.WEBSOCKET_EVENT_CHANNEL_VIEWED:
					mlog.Warn(
						"websocket.slow: dropping message",
						mlog.String("user_id", wc.UserId),
						mlog.String("type", msg.EventType()),
						mlog.String("page_id", evt.GetBroadcast().PageId),
					)
					skipSend = true
				}
			}

			if skipSend {
				continue
			}

			var msgBytes []byte
			if evtOk {
				cpyEvt := evt.SetSequence(wc.Sequence)
				msgBytes = []byte(cpyEvt.ToJson())
				wc.Sequence++
			} else {
				msgBytes = []byte(msg.ToJson())
			}

			if len(wc.send) >= sendFullWarn {
				logData := []mlog.Field{
					mlog.String("user_id", wc.UserId),
					mlog.String("type", msg.EventType()),
					mlog.Int("size", len(msgBytes)),
				}
				if evtOk {
					logData = append(logData, mlog.String("page_id", evt.GetBroadcast().PageId))
				}

				mlog.Warn("websocket.full", logData...)
			}

			wc.WebSocket.SetWriteDeadline(time.Now().Add(writeWaitTime))
			if err := wc.WebSocket.WriteMessage(websocket.TextMessage, msgBytes); err != nil {
				wc.logSocketErr("websocket.send", err)
				return
			}

			if wc.App.Metrics() != nil {
				wc.App.Metrics().IncrementWebSocketBroadcast(msg.EventType())
			}
		case <-ticker.C:
			wc.WebSocket.SetWriteDeadline(time.Now().Add(writeWaitTime))
			if err := wc.WebSocket.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				wc.logSocketErr("websocket.ticker", err)
				return
			}

		case <-wc.endWritePump:
			return

		case <-authTicker.C:
			if wc.GetSessionToken() == "" {
				mlog.Debug("websocket.authTicker: did not authenticate", mlog.Any("ip_address", wc.WebSocket.RemoteAddr()))
				return
			}
			authTicker.Stop()
		}
	}
}

// InvalidateCache resets all internal data of the WebConn.
func (wc *WebConn) InvalidateCache() {
	wc.AllPageMembers = nil
	wc.LastAllPageMembersTime = 0
	wc.SetSession(nil)
	wc.SetSessionExpiresAt(0)
}

// IsAuthenticated returns whether the given WebConn is authenticated or not.
func (wc *WebConn) IsAuthenticated() bool {
	// Check the expiry to see if we need to check for a new session
	if wc.GetSessionExpiresAt() < model.GetMillis() {
		if wc.GetSessionToken() == "" {
			return false
		}

		session, err := wc.App.GetSession(wc.GetSessionToken())
		if err != nil {
			mlog.Error("Invalid session.", mlog.Err(err))
			wc.SetSessionToken("")
			wc.SetSession(nil)
			wc.SetSessionExpiresAt(0)
			return false
		}

		wc.SetSession(session)
		wc.SetSessionExpiresAt(session.ExpiresAt)
	}

	return true
}

func (wc *WebConn) createHelloMessage() *model.WebSocketEvent {
	msg := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_HELLO, "", "", wc.UserId, nil)
	msg.Add("server_version", fmt.Sprintf("%v.%v.%v.%v", model.CurrentVersion, model.BuildNumber, wc.App.ClientConfigHash(), wc.App.Srv().License() != nil))
	return msg
}

func (wc *WebConn) shouldSendEventToGuest(msg *model.WebSocketEvent) bool {
	//var userId string
	//var canSee bool
	//
	//switch msg.EventType() {
	//case model.WEBSOCKET_EVENT_USER_UPDATED:
	//	user, ok := msg.GetData()["user"].(*model.User)
	//	if !ok {
	//		mlog.Error("webhub.shouldSendEvent: user not found in message", mlog.Any("user", msg.GetData()["user"]))
	//		return false
	//	}
	//	userId = user.Id
	//case model.WEBSOCKET_EVENT_NEW_USER:
	//	userId = msg.GetData()["user_id"].(string)
	//default:
	//	return true
	//}
	//
	//canSee, err := wc.App.UserCanSeeOtherUser(wc.UserId, userId)
	//if err != nil {
	//	mlog.Error("webhub.shouldSendEvent.", mlog.Err(err))
	//	return false
	//}
	//
	//return canSee

	// TODO: check this
	return false
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// shouldSendEvent returns whether the message should be sent or not.
func (wc *WebConn) shouldSendEvent(msg *model.WebSocketEvent) bool {
	// IMPORTANT: Do not send event if WebConn does not have a session
	if !wc.IsAuthenticated() {
		return false
	}

	// If the event contains sanitized data, only send to users that don't have permission to
	// see sensitive data. Prevents admin clients from receiving events with bad data
	var hasReadPrivateDataPermission *bool
	if msg.GetBroadcast().ContainsSanitizedData {
		hasReadPrivateDataPermission = model.NewBool(wc.App.RolesGrantPermission(wc.GetSession().GetUserRoles(), model.PERMISSION_MANAGE_SYSTEM.Id))

		if *hasReadPrivateDataPermission {
			return false
		}
	}

	// If the event contains sensitive data, only send to users with permission to see it
	if msg.GetBroadcast().ContainsSensitiveData {
		if hasReadPrivateDataPermission == nil {
			hasReadPrivateDataPermission = model.NewBool(wc.App.RolesGrantPermission(wc.GetSession().GetUserRoles(), model.PERMISSION_MANAGE_SYSTEM.Id))
		}

		if !*hasReadPrivateDataPermission {
			return false
		}
	}

	// If the event is destined to a specific user
	if msg.GetBroadcast().UserId != "" {
		return wc.UserId == msg.GetBroadcast().UserId
	}

	// if the user is omitted don't send the message
	if len(msg.GetBroadcast().OmitUsers) > 0 {
		if _, ok := msg.GetBroadcast().OmitUsers[wc.UserId]; ok {
			return false
		}
	}

	// Only report events to users who are in the channel for the event
	if msg.GetBroadcast().PageId != "" {
		if model.GetMillis()-wc.LastAllPageMembersTime > webConnMemberCacheTime {
			wc.AllPageMembers = nil
			wc.LastAllPageMembersTime = 0
		}

		if contains(wc.PageIds, msg.Broadcast.PageId) {
			return true
		}

		return false

		//if wc.AllPageMembers == nil {
		//	result, err := wc.App.Srv().Store.Channel().GetAllChannelMembersForUser(wc.UserId, true, false)
		//	if err != nil {
		//		mlog.Error("webhub.shouldSendEvent.", mlog.Err(err))
		//		return false
		//	}
		//	wc.allChannelMembers = result
		//	wc.lastAllChannelMembersTime = model.GetMillis()
		//}
		//
		//if _, ok := wc.AllPageMembers[msg.GetBroadcast().ChannelId]; ok {
		//	return true
		//}
		//return false
	}

	// Only report events to users who are in the team for the event
	if msg.GetBroadcast().TeamId != "" {
		return wc.isMemberOfTeam(msg.GetBroadcast().TeamId)
	}

	if wc.GetSession().Props[model.SESSION_PROP_IS_GUEST] == "true" {
		return wc.shouldSendEventToGuest(msg)
	}

	return true
}

// IsMemberOfTeam returns whether the user of the WebConn
// is a member of the given teamId or not.
func (wc *WebConn) isMemberOfTeam(teamId string) bool {
	currentSession := wc.GetSession()

	if currentSession == nil || currentSession.Token == "" {
		session, err := wc.App.GetSession(wc.GetSessionToken())
		if err != nil {
			mlog.Error("Invalid session.", mlog.Err(err))
			return false
		}
		wc.SetSession(session)
		currentSession = session
	}

	return currentSession.GetTeamByTeamId(teamId) != nil
}

func (wc *WebConn) logSocketErr(source string, err error) {
	// browsers will appear as CloseNoStatusReceived
	if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseNoStatusReceived) {
		mlog.Debug(source+": client side closed socket", mlog.String("user_id", wc.UserId))
	} else {
		mlog.Debug(source+": closing websocket", mlog.String("user_id", wc.UserId), mlog.Err(err))
	}
}
