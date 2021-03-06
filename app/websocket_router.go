// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"net/http"

	"bitbucket.org/enesyteam/papo-server/mlog"
	"bitbucket.org/enesyteam/papo-server/model"
	"bitbucket.org/enesyteam/papo-server/utils"
)

type webSocketHandler interface {
	ServeWebSocket(*WebConn, *model.WebSocketRequest)
}

type WebSocketRouter struct {
	server   *Server
	app      *App
	handlers map[string]webSocketHandler
}

func (wr *WebSocketRouter) Handle(action string, handler webSocketHandler) {
	wr.handlers[action] = handler
}

func (wr *WebSocketRouter) ServeWebSocket(conn *WebConn, r *model.WebSocketRequest) {
	wr.app.InitServer()

	if r.Action == "" {
		err := model.NewAppError("ServeWebSocket", "api.web_socket_router.no_action.app_error", nil, "", http.StatusBadRequest)
		returnWebSocketError(wr.app, conn, r, err)
		return
	}

	if r.Seq <= 0 {
		err := model.NewAppError("ServeWebSocket", "api.web_socket_router.bad_seq.app_error", nil, "", http.StatusBadRequest)
		returnWebSocketError(wr.app, conn, r, err)
		return
	}

	if r.Action == model.WEBSOCKET_AUTHENTICATION_CHALLENGE {
		if conn.GetSessionToken() != "" {
			return
		}

		token, ok := r.Data["token"].(string)
		if !ok {
			conn.WebSocket.Close()
			return
		}

		session, err := wr.app.GetSession(token)
		if err != nil {
			conn.WebSocket.Close()
			return
		}

		conn.SetSession(session)
		conn.SetSessionToken(session.Token)
		conn.UserId = session.UserId

		wr.app.HubRegister(conn)

		wr.app.Srv().Go(func() {
			wr.app.SetStatusOnline(session.UserId, false)
			wr.app.UpdateLastActivityAtIfNeeded(*session)
		})

		resp := model.NewWebSocketResponse(model.STATUS_OK, r.Seq, nil)
		hub := wr.app.GetHubForUserId(conn.UserId)
		if hub == nil {
			return
		}
		hub.SendMessage(conn, resp)

		return
	}

	if !conn.IsAuthenticated() {
		err := model.NewAppError("ServeWebSocket", "api.web_socket_router.not_authenticated.app_error", nil, "", http.StatusUnauthorized)
		returnWebSocketError(wr.app, conn, r, err)
		return
	}

	handler, ok := wr.handlers[r.Action]
	if !ok {
		err := model.NewAppError("ServeWebSocket", "api.web_socket_router.bad_action.app_error", nil, "", http.StatusInternalServerError)
		returnWebSocketError(wr.app, conn, r, err)
		return
	}

	handler.ServeWebSocket(conn, r)
}

func returnWebSocketError(app *App, conn *WebConn, r *model.WebSocketRequest, err *model.AppError) {
	mlog.Error(
		"websocket routing error.",
		mlog.Int64("seq", r.Seq),
		mlog.String("user_id", conn.UserId),
		mlog.String("system_message", err.SystemMessage(utils.T)),
		mlog.Err(err),
	)

	hub := app.GetHubForUserId(conn.UserId)
	if hub == nil {
		return
	}

	err.DetailedError = ""
	errorResp := model.NewWebSocketError(r.Seq, err)
	hub.SendMessage(conn, errorResp)
}
