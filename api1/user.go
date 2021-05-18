// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package api1

import (
	"bitbucket.org/enesyteam/papo-server/app"
	"bitbucket.org/enesyteam/papo-server/mlog"
	"bitbucket.org/enesyteam/papo-server/model"
	"net/http"
	"strconv"
	"time"
)

func (api *API) InitUser() {
	api.BaseRoutes.Users.Handle("", api.ApiHandler(createUser)).Methods("POST")
	api.BaseRoutes.Users.Handle("", api.ApiSessionRequired(getUsers)).Methods("GET")
	api.BaseRoutes.User.Handle("", api.ApiSessionRequired(getUser)).Methods("GET")
	api.BaseRoutes.Users.Handle("/stats", api.ApiSessionRequired(getTotalUsersStats)).Methods("GET")
	api.BaseRoutes.Users.Handle("/login", api.ApiHandler(login)).Methods("POST")
	api.BaseRoutes.Users.Handle("/logout", api.ApiHandler(logout)).Methods("POST")

	api.BaseRoutes.User.Handle("/sessions", api.ApiSessionRequired(getSessions)).Methods("GET")
	api.BaseRoutes.User.Handle("/sessions/revoke", api.ApiSessionRequired(revokeSession)).Methods("POST")
	api.BaseRoutes.User.Handle("/sessions/revoke/all", api.ApiSessionRequired(revokeAllSessionsForUser)).Methods("POST")
	api.BaseRoutes.Users.Handle("/sessions/device", api.ApiSessionRequired(attachDeviceId)).Methods("PUT")
	api.BaseRoutes.User.Handle("/audits", api.ApiSessionRequired(getUserAudits)).Methods("GET")

	api.BaseRoutes.User.Handle("/tokens", api.ApiSessionRequired(createUserAccessToken)).Methods("POST")
	api.BaseRoutes.User.Handle("/tokens", api.ApiSessionRequired(getUserAccessTokensForUser)).Methods("GET")
	api.BaseRoutes.Users.Handle("/tokens", api.ApiSessionRequired(getUserAccessTokens)).Methods("GET")

	api.BaseRoutes.User.Handle("/patch", api.ApiSessionRequired(patchUser)).Methods("PUT")
	api.BaseRoutes.User.Handle("/locale", api.ApiSessionRequired(updateLocale)).Methods("PUT")
}

func getSessions(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireUserId()
	if c.Err != nil {
		return
	}

	if !c.App.SessionHasPermissionToUser(c.App.Session, c.Params.UserId) {
		c.SetPermissionError(model.PERMISSION_EDIT_OTHER_USERS)
		return
	}

	sessions, err := c.App.GetSessions(c.Params.UserId)
	if err != nil {
		c.Err = err
		return
	}

	for _, session := range sessions {
		session.Sanitize()
	}

	w.Write([]byte(model.SessionsToJson(sessions)))
}

func revokeSession(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireUserId()
	if c.Err != nil {
		return
	}

	if !c.App.SessionHasPermissionToUser(c.App.Session, c.Params.UserId) {
		c.SetPermissionError(model.PERMISSION_EDIT_OTHER_USERS)
		return
	}

	props := model.MapFromJson(r.Body)
	sessionId := props["session_id"]
	if sessionId == "" {
		c.SetInvalidParam("session_id")
		return
	}

	session, err := c.App.GetSessionById(sessionId)
	if err != nil {
		c.Err = err
		return
	}

	if session.UserId != c.Params.UserId {
		c.SetInvalidUrlParam("user_id")
		return
	}

	if err := c.App.RevokeSession(session); err != nil {
		c.Err = err
		return
	}

	ReturnStatusOK(w)
}

func revokeAllSessionsForUser(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireUserId()
	if c.Err != nil {
		return
	}

	if !c.App.SessionHasPermissionToUser(c.App.Session, c.Params.UserId) {
		c.SetPermissionError(model.PERMISSION_EDIT_OTHER_USERS)
		return
	}

	if err := c.App.RevokeAllSessions(c.Params.UserId); err != nil {
		c.Err = err
		return
	}

	ReturnStatusOK(w)
}

func attachDeviceId(c *Context, w http.ResponseWriter, r *http.Request) {
	props := model.MapFromJson(r.Body)

	deviceId := props["device_id"]
	if len(deviceId) == 0 {
		c.SetInvalidParam("device_id")
		return
	}

	// A special case where we logout of all other sessions with the same device id
	if err := c.App.RevokeSessionsForDeviceId(c.App.Session.UserId, deviceId, c.App.Session.Id); err != nil {
		c.Err = err
		return
	}

	c.App.ClearSessionCacheForUser(c.App.Session.UserId)
	c.App.Session.SetExpireInDays(*c.App.Config().ServiceSettings.SessionLengthMobileInDays)

	maxAge := *c.App.Config().ServiceSettings.SessionLengthMobileInDays * 60 * 60 * 24

	secure := false
	if app.GetProtocol(r) == "https" {
		secure = true
	}

	expiresAt := time.Unix(model.GetMillis()/1000+int64(maxAge), 0)
	sessionCookie := &http.Cookie{
		Name:     model.SESSION_COOKIE_TOKEN,
		Value:    c.App.Session.Token,
		Path:     "/",
		MaxAge:   maxAge,
		Expires:  expiresAt,
		HttpOnly: true,
		Domain:   c.App.GetCookieDomain(),
		Secure:   secure,
	}

	http.SetCookie(w, sessionCookie)

	if err := c.App.AttachDeviceId(c.App.Session.Id, deviceId, c.App.Session.ExpiresAt); err != nil {
		c.Err = err
		return
	}

	c.LogAudit("")
	ReturnStatusOK(w)
}

func getUserAudits(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireUserId()
	if c.Err != nil {
		return
	}

	if !c.App.SessionHasPermissionToUser(c.App.Session, c.Params.UserId) {
		c.SetPermissionError(model.PERMISSION_EDIT_OTHER_USERS)
		return
	}

	audits, err := c.App.GetAuditsPage(c.Params.UserId, c.Params.Page, c.Params.PerPage)
	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(audits.ToJson()))
}

func createUser(c *Context, w http.ResponseWriter, r *http.Request) {
	user := model.UserFromJson(r.Body)
	if user == nil {
		c.SetInvalidParam("user")
		return
	}

	tokenId := r.URL.Query().Get("t")
	inviteId := r.URL.Query().Get("iid")

	// No permission check required

	var ruser *model.User
	var err *model.AppError
	if len(tokenId) > 0 {
		mlog.Info( "Tạo user từ access_token" )
		//ruser, err = c.App.CreateUserWithToken(user, tokenId)
	} else if len(inviteId) > 0 {
		mlog.Info( "Tạo user từ một lời mời" )
		//ruser, err = c.App.CreateUserWithInviteId(user, inviteId)
	} else if c.IsSystemAdmin() {
		mlog.Info( "Tạo system_user" )
		ruser, err = c.App.CreateUserAsAdmin(user)
	} else {
		ruser, err = c.App.CreateUserFromSignup(user)
	}

	if err != nil {
		c.Err = err
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(ruser.ToJson()))
}

func getUser(c *Context, w http.ResponseWriter, r *http.Request) {
	mlog.Warn( "called" )
	c.RequireUserId()
	if c.Err != nil {
		return
	}

	// No permission check required

	var user *model.User
	var err *model.AppError

	if user, err = c.App.GetUser(c.Params.UserId); err != nil {
		c.Err = err
		return
	}

	etag := user.Etag(*c.App.Config().PrivacySettings.ShowFullName, *c.App.Config().PrivacySettings.ShowEmailAddress)

	if c.HandleEtag(etag, "Get User", w, r) {
		return
	}

	if c.App.Session.UserId == user.Id {
		user.Sanitize(map[string]bool{})
	} else {
		c.App.SanitizeProfile(user, c.IsSystemAdmin())
	}
	c.App.UpdateLastActivityAtIfNeeded(c.App.Session)
	w.Header().Set(model.HEADER_ETAG_SERVER, etag)
	w.Write([]byte(user.ToJson()))
}

func getTotalUsersStats(c *Context, w http.ResponseWriter, r *http.Request) {
	if c.Err != nil {
		return
	}

	stats, err := c.App.GetTotalUsersStats()
	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(stats.ToJson()))
}

func login(c *Context, w http.ResponseWriter, r *http.Request) {
	// For hardened mode, translate all login errors to generic.
	defer func() {
		if *c.App.Config().ServiceSettings.ExperimentalEnableHardenedMode && c.Err != nil {
			c.Err = model.NewAppError("login", "api.user.login.invalid_credentials", nil, "", http.StatusUnauthorized)
		}
	}()

	props := model.MapFromJson(r.Body)

	id := props["id"]
	loginId := props["login_id"]
	password := props["password"]
	mfaToken := props["token"]
	deviceId := props["device_id"]
	ldapOnly := props["ldap_only"] == "true"

	if *c.App.Config().ExperimentalSettings.ClientSideCertEnable {
		if license := c.App.License(); license == nil || !*license.Features.SAML {
			c.Err = model.NewAppError("ClientSideCertNotAllowed", "api.user.login.client_side_cert.license.app_error", nil, "", http.StatusBadRequest)
			return
		} else {
			certPem, certSubject, certEmail := c.App.CheckForClienSideCert(r)
			mlog.Debug("Client Cert", mlog.String("cert_subject", certSubject), mlog.String("cert_email", certEmail))

			if len(certPem) == 0 || len(certEmail) == 0 {
				c.Err = model.NewAppError("ClientSideCertMissing", "api.user.login.client_side_cert.certificate.app_error", nil, "", http.StatusBadRequest)
				return
			} else if *c.App.Config().ExperimentalSettings.ClientSideCertCheck == model.CLIENT_SIDE_CERT_CHECK_PRIMARY_AUTH {
				loginId = certEmail
				password = "certificate"
			}
		}
	}

	c.LogAuditWithUserId(id, "attempt - login_id="+loginId)
	user, err := c.App.AuthenticateUserForLogin(id, loginId, password, mfaToken, ldapOnly)

	if err != nil {
		c.LogAuditWithUserId(id, "failure - login_id="+loginId)
		c.Err = err
		return
	}

	c.LogAuditWithUserId(user.Id, "authenticated")

	var session *model.Session
	session, err = c.App.DoLogin(w, r, user, deviceId)
	if err != nil {
		c.Err = err
		return
	}

	c.LogAuditWithUserId(user.Id, "success")

	c.App.Session = *session

	user.Sanitize(map[string]bool{})

	w.Write([]byte(user.ToJson()))
}

func logout(c *Context, w http.ResponseWriter, r *http.Request) {
	Logout(c, w, r)
}

func Logout(c *Context, w http.ResponseWriter, r *http.Request) {
	c.LogAudit("")
	c.RemoveSessionCookie(w, r)
	if c.App.Session.Id != "" {
		if err := c.App.RevokeSessionById(c.App.Session.Id); err != nil {
			c.Err = err
			return
		}
	}

	ReturnStatusOK(w)
}

func createUserAccessToken(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireUserId()
	if c.Err != nil {
		return
	}

	if c.App.Session.IsOAuth {
		c.SetPermissionError(model.PERMISSION_CREATE_USER_ACCESS_TOKEN)
		c.Err.DetailedError += ", attempted access by oauth app"
		return
	}

	accessToken := model.UserAccessTokenFromJson(r.Body)
	if accessToken == nil {
		c.SetInvalidParam("user_access_token")
		return
	}

	if accessToken.Description == "" {
		c.SetInvalidParam("description")
		return
	}

	c.LogAudit("")

	if !c.App.SessionHasPermissionTo(c.App.Session, model.PERMISSION_CREATE_USER_ACCESS_TOKEN) {
		c.SetPermissionError(model.PERMISSION_CREATE_USER_ACCESS_TOKEN)
		return
	}

	if !c.App.SessionHasPermissionToUser(c.App.Session, c.Params.UserId) {
		c.SetPermissionError(model.PERMISSION_EDIT_OTHER_USERS)
		return
	}

	accessToken.UserId = c.Params.UserId
	accessToken.Token = ""

	var err *model.AppError
	accessToken, err = c.App.CreateUserAccessToken(accessToken)
	if err != nil {
		c.Err = err
		return
	}

	c.LogAudit("success - token_id=" + accessToken.Id)
	w.Write([]byte(accessToken.ToJson()))
}

func searchUserAccessTokens(c *Context, w http.ResponseWriter, r *http.Request) {
	if !c.App.SessionHasPermissionTo(c.App.Session, model.PERMISSION_MANAGE_SYSTEM) {
		c.SetPermissionError(model.PERMISSION_MANAGE_SYSTEM)
		return
	}
	props := model.UserAccessTokenSearchFromJson(r.Body)
	if props == nil {
		c.SetInvalidParam("user_access_token_search")
		return
	}

	if len(props.Term) == 0 {
		c.SetInvalidParam("term")
		return
	}
	accessTokens, err := c.App.SearchUserAccessTokens(props.Term)
	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(model.UserAccessTokenListToJson(accessTokens)))
}

func getUserAccessTokens(c *Context, w http.ResponseWriter, r *http.Request) {
	if !c.App.SessionHasPermissionTo(c.App.Session, model.PERMISSION_MANAGE_SYSTEM) {
		c.SetPermissionError(model.PERMISSION_MANAGE_SYSTEM)
		return
	}

	accessTokens, err := c.App.GetUserAccessTokens(c.Params.Page, c.Params.PerPage)
	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(model.UserAccessTokenListToJson(accessTokens)))
}

func getUserAccessTokensForUser(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireUserId()
	if c.Err != nil {
		return
	}

	if !c.App.SessionHasPermissionTo(c.App.Session, model.PERMISSION_READ_USER_ACCESS_TOKEN) {
		c.SetPermissionError(model.PERMISSION_READ_USER_ACCESS_TOKEN)
		return
	}

	if !c.App.SessionHasPermissionToUser(c.App.Session, c.Params.UserId) {
		c.SetPermissionError(model.PERMISSION_EDIT_OTHER_USERS)
		return
	}

	accessTokens, err := c.App.GetUserAccessTokensForUser(c.Params.UserId, c.Params.Page, c.Params.PerPage)
	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(model.UserAccessTokenListToJson(accessTokens)))
}

func getUserAccessToken(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireTokenId()
	if c.Err != nil {
		return
	}

	if !c.App.SessionHasPermissionTo(c.App.Session, model.PERMISSION_READ_USER_ACCESS_TOKEN) {
		c.SetPermissionError(model.PERMISSION_READ_USER_ACCESS_TOKEN)
		return
	}

	accessToken, err := c.App.GetUserAccessToken(c.Params.TokenId, true)
	if err != nil {
		c.Err = err
		return
	}

	if !c.App.SessionHasPermissionToUser(c.App.Session, accessToken.UserId) {
		c.SetPermissionError(model.PERMISSION_EDIT_OTHER_USERS)
		return
	}

	w.Write([]byte(accessToken.ToJson()))
}

func revokeUserAccessToken(c *Context, w http.ResponseWriter, r *http.Request) {
	props := model.MapFromJson(r.Body)
	tokenId := props["token_id"]

	if tokenId == "" {
		c.SetInvalidParam("token_id")
	}

	c.LogAudit("")

	if !c.App.SessionHasPermissionTo(c.App.Session, model.PERMISSION_REVOKE_USER_ACCESS_TOKEN) {
		c.SetPermissionError(model.PERMISSION_REVOKE_USER_ACCESS_TOKEN)
		return
	}

	accessToken, err := c.App.GetUserAccessToken(tokenId, false)
	if err != nil {
		c.Err = err
		return
	}

	if !c.App.SessionHasPermissionToUser(c.App.Session, accessToken.UserId) {
		c.SetPermissionError(model.PERMISSION_EDIT_OTHER_USERS)
		return
	}

	err = c.App.RevokeUserAccessToken(accessToken)
	if err != nil {
		c.Err = err
		return
	}

	c.LogAudit("success - token_id=" + accessToken.Id)
	ReturnStatusOK(w)
}

func disableUserAccessToken(c *Context, w http.ResponseWriter, r *http.Request) {
	props := model.MapFromJson(r.Body)
	tokenId := props["token_id"]

	if tokenId == "" {
		c.SetInvalidParam("token_id")
	}

	c.LogAudit("")

	// No separate permission for this action for now
	if !c.App.SessionHasPermissionTo(c.App.Session, model.PERMISSION_REVOKE_USER_ACCESS_TOKEN) {
		c.SetPermissionError(model.PERMISSION_REVOKE_USER_ACCESS_TOKEN)
		return
	}

	accessToken, err := c.App.GetUserAccessToken(tokenId, false)
	if err != nil {
		c.Err = err
		return
	}

	if !c.App.SessionHasPermissionToUser(c.App.Session, accessToken.UserId) {
		c.SetPermissionError(model.PERMISSION_EDIT_OTHER_USERS)
		return
	}

	err = c.App.DisableUserAccessToken(accessToken)
	if err != nil {
		c.Err = err
		return
	}

	c.LogAudit("success - token_id=" + accessToken.Id)
	ReturnStatusOK(w)
}

func enableUserAccessToken(c *Context, w http.ResponseWriter, r *http.Request) {
	props := model.MapFromJson(r.Body)
	tokenId := props["token_id"]

	if tokenId == "" {
		c.SetInvalidParam("token_id")
	}

	c.LogAudit("")

	// No separate permission for this action for now
	if !c.App.SessionHasPermissionTo(c.App.Session, model.PERMISSION_CREATE_USER_ACCESS_TOKEN) {
		c.SetPermissionError(model.PERMISSION_CREATE_USER_ACCESS_TOKEN)
		return
	}

	accessToken, err := c.App.GetUserAccessToken(tokenId, false)
	if err != nil {
		c.Err = err
		return
	}

	if !c.App.SessionHasPermissionToUser(c.App.Session, accessToken.UserId) {
		c.SetPermissionError(model.PERMISSION_EDIT_OTHER_USERS)
		return
	}

	err = c.App.EnableUserAccessToken(accessToken)
	if err != nil {
		c.Err = err
		return
	}

	c.LogAudit("success - token_id=" + accessToken.Id)
	ReturnStatusOK(w)
}

func getUsers(c *Context, w http.ResponseWriter, r *http.Request) {
	inTeamId := r.URL.Query().Get("in_team")
	notInTeamId := r.URL.Query().Get("not_in_team")
	inChannelId := r.URL.Query().Get("in_channel")
	notInChannelId := r.URL.Query().Get("not_in_channel")
	withoutTeam := r.URL.Query().Get("without_team")
	inactive := r.URL.Query().Get("inactive")
	role := r.URL.Query().Get("role")
	sort := r.URL.Query().Get("sort")

	if len(notInChannelId) > 0 && len(inTeamId) == 0 {
		c.SetInvalidUrlParam("team_id")
		return
	}

	if sort != "" && sort != "last_activity_at" && sort != "create_at" && sort != "status" {
		c.SetInvalidUrlParam("sort")
		return
	}

	// Currently only supports sorting on a team
	// or sort="status" on inChannelId
	if (sort == "last_activity_at" || sort == "create_at") && (inTeamId == "" || notInTeamId != "" || inChannelId != "" || notInChannelId != "" || withoutTeam != "") {
		c.SetInvalidUrlParam("sort")
		return
	}
	if sort == "status" && inChannelId == "" {
		c.SetInvalidUrlParam("sort")
		return
	}

	withoutTeamBool, _ := strconv.ParseBool(withoutTeam)
	inactiveBool, _ := strconv.ParseBool(inactive)

	userGetOptions := &model.UserGetOptions{
		InTeamId:       inTeamId,
		InPageId:    inChannelId,
		NotInTeamId:    notInTeamId,
		NotInPagelId: notInChannelId,
		WithoutTeam:    withoutTeamBool,
		Inactive:       inactiveBool,
		Role:           role,
		Sort:           sort,
		Page:           c.Params.Page,
		PerPage:        c.Params.PerPage,
	}

	var profiles []*model.TeamMemberProfile
	var err *model.AppError
	etag := ""


	if len(inTeamId) > 0 {
		if !c.App.SessionHasPermissionToTeam(c.App.Session, inTeamId, model.PERMISSION_VIEW_TEAM) {
			//fmt.Println("errrrrrr")
			c.SetPermissionError(model.PERMISSION_VIEW_TEAM)
			return
		}

		etag = c.App.GetUsersInTeamEtag(inTeamId)
		if c.HandleEtag(etag, "Get Users in Team", w, r) {
			return
		}
		profiles, err = c.App.GetUsersInTeamPage(userGetOptions, c.IsSystemAdmin())
	}

	//if withoutTeamBool, _ := strconv.ParseBool(withoutTeam); withoutTeamBool {
	//
	//	// Use a special permission for now
	//	if !c.App.SessionHasPermissionTo(c.App.Session, model.PERMISSION_LIST_USERS_WITHOUT_TEAM) {
	//		c.SetPermissionError(model.PERMISSION_LIST_USERS_WITHOUT_TEAM)
	//		return
	//	}
	//
	//	profiles, err = c.App.GetUsersWithoutTeamPage(c.Params.Page, c.Params.PerPage, c.IsSystemAdmin())
	//} else if len(notInChannelId) > 0 {
	//	if !c.App.SessionHasPermissionToChannel(c.App.Session, notInChannelId, model.PERMISSION_READ_CHANNEL) {
	//		c.SetPermissionError(model.PERMISSION_READ_CHANNEL)
	//		return
	//	}
	//
	//	profiles, err = c.App.GetUsersNotInChannelPage(inTeamId, notInChannelId, c.Params.Page, c.Params.PerPage, c.IsSystemAdmin())
	//} else if len(notInTeamId) > 0 {
	//	if !c.App.SessionHasPermissionToTeam(c.App.Session, notInTeamId, model.PERMISSION_VIEW_TEAM) {
	//		c.SetPermissionError(model.PERMISSION_VIEW_TEAM)
	//		return
	//	}
	//
	//	etag = c.App.GetUsersNotInTeamEtag(inTeamId)
	//	if c.HandleEtag(etag, "Get Users Not in Team", w, r) {
	//		return
	//	}
	//
	//	profiles, err = c.App.GetUsersNotInTeamPage(notInTeamId, c.Params.Page, c.Params.PerPage, c.IsSystemAdmin())
	//} else if len(inTeamId) > 0 {
	//	if !c.App.SessionHasPermissionToTeam(c.App.Session, inTeamId, model.PERMISSION_VIEW_TEAM) {
	//		//fmt.Println("errrrrrr")
	//		c.SetPermissionError(model.PERMISSION_VIEW_TEAM)
	//		return
	//	}
	//
	//	if sort == "last_activity_at" {
	//		profiles, err = c.App.GetRecentlyActiveUsersForTeamPage(inTeamId, c.Params.Page, c.Params.PerPage, c.IsSystemAdmin())
	//	} else if sort == "create_at" {
	//		profiles, err = c.App.GetNewUsersForTeamPage(inTeamId, c.Params.Page, c.Params.PerPage, c.IsSystemAdmin())
	//	} else {
	//		etag = c.App.GetUsersInTeamEtag(inTeamId)
	//		if c.HandleEtag(etag, "Get Users in Team", w, r) {
	//			return
	//		}
	//		profiles, err = c.App.GetUsersInTeamPage(userGetOptions, c.IsSystemAdmin())
	//	}
	//} else if len(inChannelId) > 0 {
	//	if !c.App.SessionHasPermissionToChannel(c.App.Session, inChannelId, model.PERMISSION_READ_CHANNEL) {
	//		c.SetPermissionError(model.PERMISSION_READ_CHANNEL)
	//		return
	//	}
	//	if sort == "status" {
	//		profiles, err = c.App.GetUsersInChannelPageByStatus(inChannelId, c.Params.Page, c.Params.PerPage, c.IsSystemAdmin())
	//	} else {
	//		profiles, err = c.App.GetUsersInChannelPage(inChannelId, c.Params.Page, c.Params.PerPage, c.IsSystemAdmin())
	//	}
	//} else {
	//	// No permission check required
	//
	//	etag = c.App.GetUsersEtag()
	//	if c.HandleEtag(etag, "Get Users", w, r) {
	//		return
	//	}
	//	profiles, err = c.App.GetUsersPage(userGetOptions, c.IsSystemAdmin())
	//}

	if err != nil {
		c.Err = err
		return
	}

	if len(etag) > 0 {
		w.Header().Set(model.HEADER_ETAG_SERVER, etag)
	}
	c.App.UpdateLastActivityAtIfNeeded(c.App.Session)
	w.Write([]byte(model.TeamMemberListToJson(profiles)))
}

func patchUser(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireUserId()
	if c.Err != nil {
		return
	}

	patch := model.UserPatchFromJson(r.Body)
	if patch == nil {
		c.SetInvalidParam("user")
		return
	}

	if !c.App.SessionHasPermissionToUser(c.App.Session, c.Params.UserId) {
		c.SetPermissionError(model.PERMISSION_EDIT_OTHER_USERS)
		return
	}

	ouser, err := c.App.GetUser(c.Params.UserId)
	if err != nil {
		c.SetInvalidParam("user_id")
		return
	}

	if c.App.Session.IsOAuth && patch.Email != nil {
		if err != nil {
			c.Err = err
			return
		}

		if ouser.Email != *patch.Email {
			c.SetPermissionError(model.PERMISSION_EDIT_OTHER_USERS)
			c.Err.DetailedError += ", attempted email update by oauth app"
			return
		}
	}

	// If eMail update is attempted by the currently logged in user, check if correct password was provided
	//if patch.Email != nil && ouser.Email != *patch.Email && c.App.Session.UserId == c.Params.UserId {
	//	if patch.Password == nil {
	//		c.SetInvalidParam("password")
	//		return
	//	}
	//
	//	if err = c.App.DoubleCheckPassword(ouser, *patch.Password); err != nil {
	//		c.Err = err
	//		return
	//	}
	//}

	ruser, err := c.App.PatchUser(c.Params.UserId, patch, c.IsSystemAdmin())
	if err != nil {
		c.Err = err
		return
	}

	c.App.SanitizeProfile(ruser, c.IsSystemAdmin())

	//c.App.SetAutoResponderStatus(ruser, ouser.NotifyProps)
	c.LogAudit("")
	w.Write([]byte(ruser.ToJson()))
}

func updateLocale(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireUserId()
	if c.Err != nil {
		return
	}

	patch := model.UserPatchFromJson(r.Body)
	if patch == nil {
		c.SetInvalidParam("user")
		return
	}

	if !c.App.SessionHasPermissionToUser(c.App.Session, c.Params.UserId) {
		c.SetPermissionError(model.PERMISSION_EDIT_OTHER_USERS)
		return
	}

	_, err := c.App.UpdateLocale(c.Params.UserId, *patch.Locale)
	if err != nil {
		c.Err = err
		return
	}

	//fmt.Println(ruser)

	//c.App.SetAutoResponderStatus(ruser, ouser.NotifyProps)
	//c.LogAudit("")
	//w.Write([]byte(ruser.ToJson()))
	ReturnStatusOK(w)
}