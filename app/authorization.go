// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"bitbucket.org/enesyteam/papo-server/mlog"
	"bitbucket.org/enesyteam/papo-server/model"
	"bitbucket.org/enesyteam/papo-server/utils"
	"fmt"
	"strings"
)

func (app *App) SessionHasPermissionTo(session model.Session, permission *model.Permission) bool {
	return app.RolesGrantPermission(session.GetUserRoles(), permission.Id)
}

/// DO NOT USE: LEGACY
func (app *App) SessionHasPermissionToTeam(session model.Session, teamId string, permission *model.Permission) bool {
	if teamId == "" {
		return false
	}

	teamMember := session.GetTeamByTeamId(teamId)

	if teamMember != nil {

		if app.RolesGrantPermission(teamMember.GetRoles(), permission.Id) {
			return true
		}
	}

	return app.RolesGrantPermission(session.GetUserRoles(), permission.Id)
}

func (app *App) SessionHasPermissionToChannel(session model.Session, channelId string, permission *model.Permission) bool {
	//if channelId == "" {
	//	return false
	//}
	//
	//cmc := a.Srv.Store.Channel().GetAllChannelMembersForUser(session.UserId, true, true)
	//
	//var channelRoles []string
	//if cmcresult := <-cmc; cmcresult.Err == nil {
	//	ids := cmcresult.Data.(map[string]string)
	//	if roles, ok := ids[channelId]; ok {
	//		channelRoles = strings.Fields(roles)
	//		if a.RolesGrantPermission(channelRoles, permission.Id) {
	//			return true
	//		}
	//	}
	//}
	//
	//channel, err := a.GetChannel(channelId)
	//if err == nil && channel.TeamId != "" {
	//	return a.SessionHasPermissionToTeam(session, channel.TeamId, permission)
	//}
	//
	//if err != nil && err.StatusCode == http.StatusNotFound {
	//	return false
	//}
	//
	//return a.SessionHasPermissionTo(session, permission)
	return  true
}

func (app *App) SessionHasPermissionToChannelByPost(session model.Session, postId string, permission *model.Permission) bool {
	//var channelMember *model.ChannelMember
	//if result := <-a.Srv.Store.Channel().GetMemberForPost(postId, session.UserId); result.Err == nil {
	//	channelMember = result.Data.(*model.ChannelMember)
	//
	//	if a.RolesGrantPermission(channelMember.GetRoles(), permission.Id) {
	//		return true
	//	}
	//}
	//
	//if result := <-a.Srv.Store.Channel().GetForPost(postId); result.Err == nil {
	//	channel := result.Data.(*model.Channel)
	//	if channel.TeamId != "" {
	//		return a.SessionHasPermissionToTeam(session, channel.TeamId, permission)
	//	}
	//}
	//
	//return a.SessionHasPermissionTo(session, permission)
	return true
}

func (app *App) SessionHasPermissionToUser(session model.Session, userId string) bool {
	//return true
	if userId == "" {
		return false
	}

	if session.UserId == userId {
		return true
	}

	if app.SessionHasPermissionTo(session, model.PERMISSION_EDIT_OTHER_USERS) {
		return true
	}

	return false
}

func (a *App) HasPermissionTo(askingUserId string, permission *model.Permission) bool {
	user, err := a.GetUser(askingUserId)
	if err != nil {
		return false
	}

	roles := user.GetRoles()

	return a.RolesGrantPermission(roles, permission.Id)
	//return true
}

func (a *App) HasPermissionToTeam(askingUserId string, teamId string, permission *model.Permission) bool {
	if teamId == "" || askingUserId == "" {
		return false
	}

	teamMember, err := a.GetTeamMember(teamId, askingUserId)
	if err != nil {
		return false
	}

	roles := teamMember.GetRoles()

	if a.RolesGrantPermission(roles, permission.Id) {
		return true
	}

	return a.HasPermissionTo(askingUserId, permission)
	//return true
}

func (app *App) HasPermissionToChannel(askingUserId string, channelId string, permission *model.Permission) bool {
	//if channelId == "" || askingUserId == "" {
	//	return false
	//}
	//
	//channelMember, err := a.GetChannelMember(channelId, askingUserId)
	//if err == nil {
	//	roles := channelMember.GetRoles()
	//	if a.RolesGrantPermission(roles, permission.Id) {
	//		return true
	//	}
	//}
	//
	//var channel *model.Channel
	//channel, err = a.GetChannel(channelId)
	//if err == nil {
	//	return a.HasPermissionToTeam(askingUserId, channel.TeamId, permission)
	//}
	//
	//return a.HasPermissionTo(askingUserId, permission)
	return true
}

func (app *App) HasPermissionToChannelByPost(askingUserId string, postId string, permission *model.Permission) bool {
	//var channelMember *model.ChannelMember
	//if result := <-a.Srv.Store.Channel().GetMemberForPost(postId, askingUserId); result.Err == nil {
	//	channelMember = result.Data.(*model.ChannelMember)
	//
	//	if a.RolesGrantPermission(channelMember.GetRoles(), permission.Id) {
	//		return true
	//	}
	//}
	//
	//if result := <-a.Srv.Store.Channel().GetForPost(postId); result.Err == nil {
	//	channel := result.Data.(*model.Channel)
	//	return a.HasPermissionToTeam(askingUserId, channel.TeamId, permission)
	//}
	//
	//return a.HasPermissionTo(askingUserId, permission)
	return true
}

func (app *App) HasPermissionToUser(askingUserId string, userId string) bool {
	if askingUserId == userId {
		return true
	}

	if app.HasPermissionTo(askingUserId, model.PERMISSION_EDIT_OTHER_USERS) {
		return true
	}

	return false
}

func (app *App) RolesGrantPermission(roleNames []string, permissionId string) bool {
	//return true
	//fmt.Println("roleNames", roleNames)
	isAdmin := utils.Contains(roleNames, "team_admin")
	if isAdmin {
		return true
	}

	roles, err := app.GetRolesByNames(roleNames)
	//fmt.Println(roleNames)

	if err != nil {
		// This should only happen if something is very broken. We can't realistically
		// recover the situation, so deny permission and log an error.
		mlog.Error("Failed to get roles from database with role names: " + strings.Join(roleNames, ","))
		mlog.Error(fmt.Sprint(err))
		return false
	}

	for _, role := range roles {
		if role.DeleteAt != 0 {
			continue
		}

		permissions := role.Permissions

		for _, permission := range permissions {
			//fmt.Println(permission == permissionId)
			if permission == permissionId {
				return true
			}
		}
	}

	return false
}