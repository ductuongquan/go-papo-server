// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package model

import (
	"encoding/json"
	"fmt"
	"io"
)

const (
	WEBSOCKET_EVENT_TYPING                  = "typing"
	WEBSOCKET_EVENT_POSTED                  = "posted"
	WEBSOCKET_EVENT_POST_EDITED             = "post_edited"
	WEBSOCKET_EVENT_POST_DELETED            = "post_deleted"
	WEBSOCKET_EVENT_CHANNEL_CONVERTED       = "channel_converted"
	WEBSOCKET_EVENT_CHANNEL_CREATED         = "channel_created"
	WEBSOCKET_EVENT_CHANNEL_DELETED         = "channel_deleted"
	WEBSOCKET_EVENT_CHANNEL_UPDATED         = "channel_updated"
	WEBSOCKET_EVENT_CHANNEL_MEMBER_UPDATED  = "channel_member_updated"
	WEBSOCKET_EVENT_DIRECT_ADDED            = "direct_added"
	WEBSOCKET_EVENT_GROUP_ADDED             = "group_added"
	WEBSOCKET_EVENT_NEW_USER                = "new_user"
	WEBSOCKET_EVENT_ADDED_TO_TEAM           = "added_to_team"
	WEBSOCKET_EVENT_LEAVE_TEAM              = "leave_team"
	WEBSOCKET_EVENT_UPDATE_TEAM             = "update_team"
	WEBSOCKET_EVENT_DELETE_TEAM             = "delete_team"
	WEBSOCKET_EVENT_USER_ADDED              = "user_added"
	WEBSOCKET_EVENT_USER_UPDATED            = "user_updated"
	WEBSOCKET_EVENT_USER_ROLE_UPDATED       = "user_role_updated"
	WEBSOCKET_EVENT_MEMBERROLE_UPDATED      = "memberrole_updated"
	WEBSOCKET_EVENT_USER_REMOVED            = "user_removed"
	WEBSOCKET_EVENT_PREFERENCE_CHANGED      = "preference_changed"
	WEBSOCKET_EVENT_PREFERENCES_CHANGED     = "preferences_changed"
	WEBSOCKET_EVENT_PREFERENCES_DELETED     = "preferences_deleted"
	WEBSOCKET_EVENT_EPHEMERAL_MESSAGE       = "ephemeral_message"
	WEBSOCKET_EVENT_STATUS_CHANGE           = "status_change"
	WEBSOCKET_EVENT_HELLO                   = "hello"
	WEBSOCKET_AUTHENTICATION_CHALLENGE      = "authentication_challenge"
	WEBSOCKET_EVENT_REACTION_ADDED          = "reaction_added"
	WEBSOCKET_EVENT_REACTION_REMOVED        = "reaction_removed"
	WEBSOCKET_EVENT_RESPONSE                = "response"
	WEBSOCKET_EVENT_EMOJI_ADDED             = "emoji_added"
	WEBSOCKET_EVENT_CHANNEL_VIEWED          = "channel_viewed"
	WEBSOCKET_EVENT_PLUGIN_STATUSES_CHANGED = "plugin_statuses_changed"
	WEBSOCKET_EVENT_PLUGIN_ENABLED          = "plugin_enabled"
	WEBSOCKET_EVENT_PLUGIN_DISABLED         = "plugin_disabled"
	WEBSOCKET_EVENT_ROLE_UPDATED            = "role_updated"
	WEBSOCKET_EVENT_LICENSE_CHANGED         = "license_changed"
	WEBSOCKET_EVENT_CONFIG_CHANGED          = "config_changed"

	// Papo
	WEBSOCKET_EVENT_PREFERENCES_ADDED 		= "preference_added"
	WEBSOCKET_EVENT_CONVERSATION_UNSEEN 	= "conversation_unseen"
	WEBSOCKET_EVENT_CONVERSATION_SEEN 		= "conversation_seen"
	CONVERSATION_REMOVED_TAG 				= "conversation_removed_tag"
	CONVERSATION_RECEIVED_TAG 				= "conversation_received_tag"
	PAGE_TAG_CREATED 						= "page_tag_created"

	RECEIVE_CONVERSATION_UPDATED			= "receive_conversation_update"
	RECEIVE_COMMENT_UPDATED					= "receive_comment_update"
	RECEIVE_COMMENT_DELETED					= "receive_comment_delete"
	WEBSOCKET_EVENT_PAGE_STATUS_UPDATED     = "page_status_updated"
	WEBSOCKET_EVENT_PAGES_STATUS_UPDATED    = "pages_status_updated"
	WEBSOCKET_EVENT_PAGES_INI_VALIDATION_RESULT    = "pages_init_validation_result"
	WEBSOCKET_EVENT_PAGE_INIT_UPDATE_VALUE    = "pages_init_update_value"

	WEBSOCKET_EVENT_ADD_MESSAGE    			= "conversation_add_message"
	MESSAGE_SENT 							= "message_sent"
	RECEIVE_CONVERSATION_READ 				= "read_watermark"
	ADDED_ORDER 							= "added_order"
	WEBSOCKET_WARN_METRIC_STATUS_RECEIVED                    = "warn_metric_status_received"
	WEBSOCKET_WARN_METRIC_STATUS_REMOVED                     = "warn_metric_status_removed"
	WEBSOCKET_EVENT_GUESTS_DEACTIVATED                       = "guests_deactivated"
	WEBSOCKET_EVENT_UPDATE_TEAM_SCHEME                       = "update_team_scheme"
	WEBSOCKET_EVENT_RESTORE_TEAM                             = "restore_team"
)

type WebSocketMessage interface {
	ToJson() string
	IsValid() bool
	EventType() string
}

type WebsocketBroadcast struct {
	OmitUsers             map[string]bool `json:"omit_users"` // broadcast is omitted for users listed here
	UserId                string          `json:"user_id"`    // broadcast only occurs for this user
	PageId             		string          `json:"page_id"` // broadcast only occurs for users in this channel
	TeamId                string          `json:"team_id"`    // broadcast only occurs for users in this team
	ContainsSanitizedData bool            `json:"-"`
	ContainsSensitiveData bool            `json:"-"`
}

type precomputedWebSocketEventJSON struct {
	Event     json.RawMessage
	Data      json.RawMessage
	Broadcast json.RawMessage
}

type WebSocketEvent struct {
	Event     string                 `json:"event"`
	Data      map[string]interface{} `json:"data"`
	Broadcast *WebsocketBroadcast    `json:"broadcast"`
	Sequence  int64                  `json:"seq"`

	precomputedJSON *precomputedWebSocketEventJSON
}

// PrecomputeJSON precomputes and stores the serialized JSON for all fields other than Sequence.
// This makes ToJson much more efficient when sending the same event to multiple connections.
func (ev *WebSocketEvent) PrecomputeJSON() *WebSocketEvent {
	copy := ev.Copy()
	event, _ := json.Marshal(copy.Event)
	data, _ := json.Marshal(copy.Data)
	broadcast, _ := json.Marshal(copy.Broadcast)
	copy.precomputedJSON = &precomputedWebSocketEventJSON{
		Event:     json.RawMessage(event),
		Data:      json.RawMessage(data),
		Broadcast: json.RawMessage(broadcast),
	}
	return copy
}

func (m *WebSocketEvent) Add(key string, value interface{}) {
	m.Data[key] = value
}

func NewWebSocketEvent(event, teamId, pageId, userId string, omitUsers map[string]bool) *WebSocketEvent {
	return &WebSocketEvent{Event: event, Data: make(map[string]interface{}),
		Broadcast: &WebsocketBroadcast{TeamId: teamId, PageId: pageId, UserId: userId, OmitUsers: omitUsers}}
}

func (ev *WebSocketEvent) Copy() *WebSocketEvent {
	copy := &WebSocketEvent{
		Event:           ev.Event,
		Data:            ev.Data,
		Broadcast:       ev.Broadcast,
		Sequence:        ev.Sequence,
		precomputedJSON: ev.precomputedJSON,
	}
	return copy
}

func (ev *WebSocketEvent) GetData() map[string]interface{} {
	return ev.Data
}

func (ev *WebSocketEvent) GetBroadcast() *WebsocketBroadcast {
	return ev.Broadcast
}

func (ev *WebSocketEvent) GetSequence() int64 {
	return ev.Sequence
}

func (ev *WebSocketEvent) SetEvent(event string) *WebSocketEvent {
	copy := ev.Copy()
	copy.Event = event
	return copy
}

func (ev *WebSocketEvent) SetData(data map[string]interface{}) *WebSocketEvent {
	copy := ev.Copy()
	copy.Data = data
	return copy
}

func (ev *WebSocketEvent) SetBroadcast(broadcast *WebsocketBroadcast) *WebSocketEvent {
	copy := ev.Copy()
	copy.Broadcast = broadcast
	return copy
}

func (ev *WebSocketEvent) SetSequence(seq int64) *WebSocketEvent {
	copy := ev.Copy()
	copy.Sequence = seq
	return copy
}

func (o *WebSocketEvent) IsValid() bool {
	return o.Event != ""
}

func (o *WebSocketEvent) EventType() string {
	return o.Event
}

func (o *WebSocketEvent) ToJson() string {
	if o.precomputedJSON != nil {
		return fmt.Sprintf(`{"event": %s, "data": %s, "broadcast": %s, "seq": %d}`, o.precomputedJSON.Event, o.precomputedJSON.Data, o.precomputedJSON.Broadcast, o.Sequence)
	}
	b, _ := json.Marshal(o)
	return string(b)
}

func WebSocketEventFromJson(data io.Reader) *WebSocketEvent {
	var o *WebSocketEvent
	json.NewDecoder(data).Decode(&o)
	return o
}

type WebSocketResponse struct {
	Status   string                 `json:"status"`
	SeqReply int64                  `json:"seq_reply,omitempty"`
	Data     map[string]interface{} `json:"data,omitempty"`
	Error    *AppError              `json:"error,omitempty"`
}

func (m *WebSocketResponse) Add(key string, value interface{}) {
	m.Data[key] = value
}

func NewWebSocketResponse(status string, seqReply int64, data map[string]interface{}) *WebSocketResponse {
	return &WebSocketResponse{Status: status, SeqReply: seqReply, Data: data}
}

func NewWebSocketError(seqReply int64, err *AppError) *WebSocketResponse {
	return &WebSocketResponse{Status: STATUS_FAIL, SeqReply: seqReply, Error: err}
}

func (o *WebSocketResponse) IsValid() bool {
	return o.Status != ""
}

func (o *WebSocketResponse) EventType() string {
	return WEBSOCKET_EVENT_RESPONSE
}

func (o *WebSocketResponse) ToJson() string {
	b, _ := json.Marshal(o)
	return string(b)
}

func WebSocketResponseFromJson(data io.Reader) *WebSocketResponse {
	var o *WebSocketResponse
	json.NewDecoder(data).Decode(&o)
	return o
}
