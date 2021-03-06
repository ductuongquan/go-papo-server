// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package wsapi

import (
	"bitbucket.org/enesyteam/papo-server/app"
)

type API struct {
	App    *app.App
	Router *app.WebSocketRouter
}

func Init(a *app.App, router *app.WebSocketRouter) {
	api := &API{
		App:    a,
		Router: router,
	}

	api.InitUser()
	api.InitSystem()
	api.InitStatus()

	a.HubStart()
}
