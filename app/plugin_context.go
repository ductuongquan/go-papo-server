// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import "bitbucket.org/enesyteam/papo-server/plugin"

func (a *App) PluginContext() *plugin.Context {
	context := &plugin.Context{
		RequestId:      a.RequestId(),
		SessionId:      a.Session().Id,
		IpAddress:      a.IpAddress(),
		AcceptLanguage: a.AcceptLanguage(),
		UserAgent:      a.UserAgent(),
	}
	return context
}
