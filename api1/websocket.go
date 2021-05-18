// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package api1

import (
	"bitbucket.org/enesyteam/papo-server/mlog"
	"bitbucket.org/enesyteam/papo-server/model"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
)

func (api *API) InitWebSocket() {
	api.BaseRoutes.ApiRoot.Handle("/websocket", api.ApiHandlerTrustRequester(connectWebSocket)).Methods("GET")
}

func connectWebSocket(c *Context, w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  model.SOCKET_MAX_MESSAGE_SIZE_KB,
		WriteBufferSize: model.SOCKET_MAX_MESSAGE_SIZE_KB,
		CheckOrigin:     c.App.OriginChecker(),
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		mlog.Error(fmt.Sprintf("websocket connect err: %v", err))
		c.Err = model.NewAppError("connect", "api.web_socket.connect.upgrade.app_error", nil, "", http.StatusInternalServerError)
		return
	}

	query := r.URL.Query()
	pageIds := query["page_id"]

	wc := c.App.NewWebConn(ws, c.App.Session, c.App.T, "", pageIds)

	if len(c.App.Session.UserId) > 0 {
		c.App.HubRegister(wc)
	}

	wc.Pump()
}
