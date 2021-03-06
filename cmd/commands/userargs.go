// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"bitbucket.org/enesyteam/papo-server/app"
	"bitbucket.org/enesyteam/papo-server/model"
)

func getUsersFromUserArgs(a *app.App, userArgs []string) []*model.User {
	users := make([]*model.User, 0, len(userArgs))
	for _, userArg := range userArgs {
		user := getUserFromUserArg(a, userArg)
		users = append(users, user)
	}
	return users
}

func getUserFromUserArg(a *app.App, userArg string) *model.User {
	var user *model.User
	if result := <-a.Srv.Store.User().GetByEmail(userArg); result.Err == nil {
		user = result.Data.(*model.User)
	}

	if user == nil {
		if result := <-a.Srv.Store.User().GetByUsername(userArg); result.Err == nil {
			user = result.Data.(*model.User)
		}
	}

	if user == nil {
		if result := <-a.Srv.Store.User().Get(userArg); result.Err == nil {
			user = result.Data.(*model.User)
		}
	}

	return user
}