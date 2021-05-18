// Copyright (c) 2018-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package api1

import (
	"bitbucket.org/enesyteam/papo-server/model"
	"net/http"
	"strings"
)

func (api *API) InitRole() {
	api.BaseRoutes.Roles.Handle("/{role_id:[A-Za-z0-9]+}", api.ApiSessionRequiredTrustRequester(getRole)).Methods("GET")
	api.BaseRoutes.Roles.Handle("/name/{role_name:[a-z0-9_]+}", api.ApiSessionRequiredTrustRequester(getRoleByName)).Methods("GET")
	api.BaseRoutes.Roles.Handle("/names", api.ApiSessionRequiredTrustRequester(getRolesByNames)).Methods("POST")
	api.BaseRoutes.Roles.Handle("/{role_id:[A-Za-z0-9]+}/patch", api.ApiSessionRequired(patchRole)).Methods("PUT")

	api.BaseRoutes.Roles.Handle("", api.ApiSessionRequired(createRole)).Methods("POST")
	api.BaseRoutes.Team.Handle("/roles", api.ApiSessionRequired(getTeamRoles)).Methods("GET")

	api.BaseRoutes.ApiRoot.Handle("/permissions", api.ApiSessionRequired(getAvailablePermissions)).Methods("GET")
}

func getAvailablePermissions(c *Context, w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(model.PermissionsListToJson(model.PAPO_PERMISSIONS)))
}

func createRole(c *Context, w http.ResponseWriter, r *http.Request) {

	var err *model.AppError
	role := model.RoleFromJson(r.Body)
	if role == nil {
		c.SetInvalidParam("vai trò")
		return
	}

	role, err = c.App.CreateRole(role)

	if err != nil {
		c.Err = err
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(role.ToJson()))
}

func getTeamRoles(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireTeamId()
	if c.Err != nil {
		return
	}

	//if !c.App.SessionHasPermissionToTeam(c.App.Session, c.Params.TeamId, model.PERMISSION_MANAGE_TEAM_ROLES) {
	//	//fmt.Println("errrrrrr")
	//	c.SetPermissionError(model.PERMISSION_MANAGE_TEAM_ROLES)
	//	return
	//}

	roles, err := c.App.GetRolesByTeamId(c.Params.TeamId)
	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(model.RoleListToJson(roles)))
}

func getRole(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireRoleId()
	if c.Err != nil {
		return
	}

	role, err := c.App.GetRole(c.Params.RoleId)
	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(role.ToJson()))
}

func getRoleByName(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireRoleName()
	if c.Err != nil {
		return
	}

	role, err := c.App.GetRoleByName(c.Params.RoleName)
	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(role.ToJson()))
}

func getRolesByNames(c *Context, w http.ResponseWriter, r *http.Request) {
	rolenames := model.ArrayFromJson(r.Body)

	if len(rolenames) == 0 {
		c.SetInvalidParam("rolenames")
		return
	}

	var cleanedRoleNames []string
	for _, rolename := range rolenames {
		if strings.TrimSpace(rolename) == "" {
			continue
		}

		if !model.IsValidRoleName(rolename) {
			c.SetInvalidParam("rolename")
			return
		}

		cleanedRoleNames = append(cleanedRoleNames, rolename)
	}

	roles, err := c.App.GetRolesByNames(cleanedRoleNames)
	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(model.RoleListToJson(roles)))
}

func patchRole(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireRoleId()
	if c.Err != nil {
		return
	}

	patch := model.RolePatchFromJson(r.Body)
	if patch == nil {
		c.SetInvalidParam("role")
		return
	}

	oldRole, err := c.App.GetRole(c.Params.RoleId)
	if err != nil {
		c.Err = err
		return
	}

	if c.App.License() == nil && patch.Permissions != nil {
		allowedPermissions := []string{
			model.PERMISSION_VIEW_TEAM.Id,
			model.PERMISSION_CREATE_TEAM.Id,
			model.PERMISSION_MANAGE_TEAM_ROLES.Id,
			model.PERMISSION_MANAGE_TEAM.Id,
			model.PERMISSION_MANAGE_WEBHOOKS.Id,
			model.PERMISSION_MANAGE_SLASH_COMMANDS.Id,
			model.PERMISSION_MANAGE_OAUTH.Id,
			model.PERMISSION_MANAGE_SYSTEM_WIDE_OAUTH.Id,
			model.PERMISSION_MANAGE_EMOJIS.Id,
			model.PERMISSION_EDIT_OTHERS_POSTS.Id,

			model.PERMISSION_ADD_USER_TO_TEAM.Id,
			model.PERMISSION_ASSIGN_SYSTEM_ADMIN_ROLE.Id,
			model.PERMISSION_MANAGE_SYSTEM.Id,
			model.PERMISSION_EDIT_OTHER_USERS.Id,
			model.PERMISSION_PERMANENT_DELETE_USER.Id,

			model.PERMISSION_INVITE_USER.Id,
			model.PERMISSION_ADD_USER_TO_TEAM.Id,
			model.PERMISSION_ASSIGN_SYSTEM_ADMIN_ROLE.Id,
			model.PERMISSION_EDIT_OTHER_USERS.Id,
			model.PERMISSION_PERMANENT_DELETE_USER.Id,
			model.PERMISSION_MANAGE_ROLES.Id,
			model.PERMISSION_UPLOAD_FILE.Id,
			model.PERMISSION_VIEW_ADMIN_STATS.Id,
			model.PERMISSION_VIEW_ADMIN_ADVANCED_STATS.Id,
			model.PERMISSION_VIEW_ADMIN_MEMBER.Id,
			model.PERMISSION_VIEW_ORDERS.Id,
			model.PERMISSION_EDIT_ORDERS.Id,
			model.PERMISSION_VIEW_PRODUCTS.Id,
			model.PERMISSION_EDIT_PRODUCTS.Id,
			model.PERMISSION_VIEW_LOGISTICS.Id,
			model.PERMISSION_EDIT_LOGISTICS.Id,
			model.PERMISSION_VIEW_SHIPPING.Id,
			model.PERMISSION_EDIT_SHIPPING.Id,

		}

		changedPermissions := model.PermissionsChangedByPatch(oldRole, patch)
		for _, permission := range changedPermissions {
			allowed := false
			for _, allowedPermission := range allowedPermissions {
				if permission == allowedPermission {
					allowed = true
				}
			}

			if !allowed {
				c.Err = model.NewAppError("Api4.PatchRoles", "api.roles.patch_roles.license.error", nil, "", http.StatusNotImplemented)
				return
			}
		}
	}

	//if !c.App.SessionHasPermissionTo(c.App.Session, model.PERMISSION_MANAGE_SYSTEM) {
	//	c.SetPermissionError(model.PERMISSION_MANAGE_SYSTEM)
	//	return
	//}

	role, err := c.App.PatchRole(oldRole, patch)
	if err != nil {
		c.Err = err
		return
	}

	c.LogAudit("")
	w.Write([]byte(role.ToJson()))
}
