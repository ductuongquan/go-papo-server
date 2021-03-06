// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"bitbucket.org/enesyteam/papo-server/model"
	"bitbucket.org/enesyteam/papo-server/store"
	"fmt"
	"github.com/avct/uasurfer"
	"net/http"
	"strings"
	"time"
)

func (app *App) CheckForClienSideCert(r *http.Request) (string, string, string) {
	pem := r.Header.Get("X-SSL-Client-Cert")                // mapped to $ssl_client_cert from nginx
	subject := r.Header.Get("X-SSL-Client-Cert-Subject-DN") // mapped to $ssl_client_s_dn from nginx
	email := ""

	if len(subject) > 0 {
		for _, v := range strings.Split(subject, "/") {
			kv := strings.Split(v, "=")
			if len(kv) == 2 && kv[0] == "emailAddress" {
				email = kv[1]
			}
		}
	}

	return pem, subject, email
}

func (app *App) AuthenticateUserForLogin(id, loginId, password, mfaToken string, ldapOnly bool) (user *model.User, err *model.AppError) {
	// Do statistics
	defer func() {
		if app.Metrics != nil {
			if user == nil || err != nil {
				app.Metrics.IncrementLoginFail()
			} else {
				app.Metrics.IncrementLogin()
			}
		}
	}()

	if len(password) == 0 {
		err := model.NewAppError("AuthenticateUserForLogin", "api.user.login.blank_pwd.app_error", nil, "", http.StatusBadRequest)
		return nil, err
	}

	// Get the MM user we are trying to login
	if user, err = app.GetUserForLogin(id, loginId); err != nil {
		return nil, err
	}

	// If client side cert is enable and it's checking as a primary source
	// then trust the proxy and cert that the correct user is supplied and allow
	// them access
	if *app.Config().ExperimentalSettings.ClientSideCertEnable && *app.Config().ExperimentalSettings.ClientSideCertCheck == model.CLIENT_SIDE_CERT_CHECK_PRIMARY_AUTH {
		return user, nil
	}

	// and then authenticate them
	if user, err = app.authenticateUser(user, password, mfaToken); err != nil {
		return nil, err
	}

	//if a.PluginsReady() {
	//	var rejectionReason string
	//	pluginContext := &plugin.Context{}
	//	a.Plugins.RunMultiPluginHook(func(hooks plugin.Hooks) bool {
	//		rejectionReason = hooks.UserWillLogIn(pluginContext, user)
	//		return rejectionReason == ""
	//	}, plugin.UserWillLogInId)
	//
	//	if rejectionReason != "" {
	//		return nil, model.NewAppError("AuthenticateUserForLogin", "Login rejected by plugin: "+rejectionReason, nil, "", http.StatusBadRequest)
	//	}
	//
	//	a.Go(func() {
	//		pluginContext := &plugin.Context{}
	//		a.Plugins.RunMultiPluginHook(func(hooks plugin.Hooks) bool {
	//			hooks.UserHasLoggedIn(pluginContext, user)
	//			return true
	//		}, plugin.UserHasLoggedInId)
	//	})
	//}

	return user, nil
}

func (app *App) GetUserForLogin(id, loginId string) (*model.User, *model.AppError) {
	enableUsername := *app.Config().EmailSettings.EnableSignInWithUsername
	enableEmail := *app.Config().EmailSettings.EnableSignInWithEmail

	// If we are given a userID then fail if we can't find a user with that ID
	if len(id) != 0 {
		user, err := app.GetUser(id)
		if err != nil {
			if err.Id != store.MISSING_ACCOUNT_ERROR {
				err.StatusCode = http.StatusInternalServerError
				return nil, err
			}
			err.StatusCode = http.StatusBadRequest
			return nil, err
		}
		return user, nil
	}

	// Try to get the user by username/email
	if result := <-app.Srv.Store.User().GetForLogin(loginId, enableUsername, enableEmail); result.Err == nil {
		return result.Data.(*model.User), nil
	}

	// Try to get the user with LDAP if enabled
	if *app.Config().LdapSettings.Enable && app.Ldap != nil {
		if ldapUser, err := app.Ldap.GetUser(loginId); err == nil {
			if user, err := app.GetUserByAuth(ldapUser.AuthData, model.USER_AUTH_SERVICE_LDAP); err == nil {
				return user, nil
			}
			return ldapUser, nil
		}
	}

	return nil, model.NewAppError("GetUserForLogin", "store.sql_user.get_for_login.app_error", nil, "", http.StatusBadRequest)
}

func (app *App) DoLogin(w http.ResponseWriter, r *http.Request, user *model.User, deviceId string) (*model.Session, *model.AppError) {
	session := &model.Session{UserId: user.Id, Roles: user.GetRawRoles(), DeviceId: deviceId, IsOAuth: false}
	session.GenerateCSRF()
	maxAge := *app.Config().ServiceSettings.SessionLengthWebInDays * 60 * 60 * 24

	if len(deviceId) > 0 {
		session.SetExpireInDays(*app.Config().ServiceSettings.SessionLengthMobileInDays)

		// A special case where we logout of all other sessions with the same Id
		if err := app.RevokeSessionsForDeviceId(user.Id, deviceId, ""); err != nil {
			err.StatusCode = http.StatusInternalServerError
			return nil, err
		}
	} else {
		session.SetExpireInDays(*app.Config().ServiceSettings.SessionLengthWebInDays)
	}

	ua := uasurfer.Parse(r.UserAgent())

	plat := getPlatformName(ua)
	os := getOSName(ua)
	bname := getBrowserName(ua, r.UserAgent())
	bversion := getBrowserVersion(ua, r.UserAgent())

	session.AddProp(model.SESSION_PROP_PLATFORM, plat)
	session.AddProp(model.SESSION_PROP_OS, os)
	session.AddProp(model.SESSION_PROP_BROWSER, fmt.Sprintf("%v/%v", bname, bversion))

	var err *model.AppError
	if session, err = app.CreateSession(session); err != nil {
		err.StatusCode = http.StatusInternalServerError
		return nil, err
	}

	w.Header().Set(model.HEADER_TOKEN, session.Token)

	secure := false
	if GetProtocol(r) == "https" {
		secure = true
	}

	domain := app.GetCookieDomain()
	expiresAt := time.Unix(model.GetMillis()/1000+int64(maxAge), 0)
	sessionCookie := &http.Cookie{
		Name:     model.SESSION_COOKIE_TOKEN,
		Value:    session.Token,
		Path:     "/",
		MaxAge:   maxAge,
		Expires:  expiresAt,
		HttpOnly: true,
		Domain:   domain,
		Secure:   secure,
	}

	userCookie := &http.Cookie{
		Name:    model.SESSION_COOKIE_USER,
		Value:   user.Id,
		Path:    "/",
		MaxAge:  maxAge,
		Expires: expiresAt,
		Domain:  domain,
		Secure:  secure,
	}

	csrfCookie := &http.Cookie{
		Name:    model.SESSION_COOKIE_CSRF,
		Value:   session.GetCSRF(),
		Path:    "/",
		MaxAge:  maxAge,
		Expires: expiresAt,
		Domain:  domain,
		Secure:  secure,
	}

	http.SetCookie(w, sessionCookie)
	http.SetCookie(w, userCookie)
	http.SetCookie(w, csrfCookie)

	return session, nil
}

func GetProtocol(r *http.Request) string {
	if r.Header.Get(model.HEADER_FORWARDED_PROTO) == "https" || r.TLS != nil {
		return "https"
	}
	return "http"
}