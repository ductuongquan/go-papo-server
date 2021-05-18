// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"bitbucket.org/enesyteam/papo-server/einterfaces"
	"bitbucket.org/enesyteam/papo-server/facebook_graph"
	"bitbucket.org/enesyteam/papo-server/mlog"
	"bitbucket.org/enesyteam/papo-server/model"
	"bitbucket.org/enesyteam/papo-server/store"
	"bitbucket.org/enesyteam/papo-server/utils"
	"bytes"
	b64 "encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	OAUTH_COOKIE_MAX_AGE_SECONDS = 30 * 60 // 30 minutes
	COOKIE_OAUTH                 = "MMOAUTH"
)

func (app *App) CreateOAuthApp(a *model.OAuthApp) (*model.OAuthApp, *model.AppError) {
	if !*app.Config().ServiceSettings.EnableOAuthServiceProvider {
		return nil, model.NewAppError("CreateOAuthApp", "api.oauth.register_oauth_app.turn_off.app_error", nil, "", http.StatusNotImplemented)
	}

	secret := model.NewId()
	a.ClientSecret = secret

	if result := <-app.Srv.Store.OAuth().SaveApp(a); result.Err != nil {
		return nil, result.Err
	} else {
		return result.Data.(*model.OAuthApp), nil
	}
}

func (app *App) GetOAuthApp(appId string) (*model.OAuthApp, *model.AppError) {
	if !*app.Config().ServiceSettings.EnableOAuthServiceProvider {
		return nil, model.NewAppError("GetOAuthApp", "api.oauth.allow_oauth.turn_off.app_error", nil, "", http.StatusNotImplemented)
	}

	if result := <-app.Srv.Store.OAuth().GetApp(appId); result.Err != nil {
		return nil, result.Err
	} else {
		return result.Data.(*model.OAuthApp), nil
	}
}

func (app *App) UpdateOauthApp(oldApp, updatedApp *model.OAuthApp) (*model.OAuthApp, *model.AppError) {
	if !*app.Config().ServiceSettings.EnableOAuthServiceProvider {
		return nil, model.NewAppError("UpdateOauthApp", "api.oauth.allow_oauth.turn_off.app_error", nil, "", http.StatusNotImplemented)
	}

	updatedApp.Id = oldApp.Id
	updatedApp.CreatorId = oldApp.CreatorId
	updatedApp.CreateAt = oldApp.CreateAt
	updatedApp.ClientSecret = oldApp.ClientSecret

	if result := <-app.Srv.Store.OAuth().UpdateApp(updatedApp); result.Err != nil {
		return nil, result.Err
	} else {
		return result.Data.([2]*model.OAuthApp)[0], nil
	}
}

func (app *App) DeleteOAuthApp(appId string) *model.AppError {
	if !*app.Config().ServiceSettings.EnableOAuthServiceProvider {
		return model.NewAppError("DeleteOAuthApp", "api.oauth.allow_oauth.turn_off.app_error", nil, "", http.StatusNotImplemented)
	}

	if err := (<-app.Srv.Store.OAuth().DeleteApp(appId)).Err; err != nil {
		return err
	}

	app.InvalidateAllCaches()

	return nil
}

func (app *App) GetOAuthApps(page, perPage int) ([]*model.OAuthApp, *model.AppError) {
	if !*app.Config().ServiceSettings.EnableOAuthServiceProvider {
		return nil, model.NewAppError("GetOAuthApps", "api.oauth.allow_oauth.turn_off.app_error", nil, "", http.StatusNotImplemented)
	}

	if result := <-app.Srv.Store.OAuth().GetApps(page*perPage, perPage); result.Err != nil {
		return nil, result.Err
	} else {
		return result.Data.([]*model.OAuthApp), nil
	}
}

func (app *App) GetOAuthAppsByCreator(userId string, page, perPage int) ([]*model.OAuthApp, *model.AppError) {
	if !*app.Config().ServiceSettings.EnableOAuthServiceProvider {
		return nil, model.NewAppError("GetOAuthAppsByUser", "api.oauth.allow_oauth.turn_off.app_error", nil, "", http.StatusNotImplemented)
	}

	if result := <-app.Srv.Store.OAuth().GetAppByUser(userId, page*perPage, perPage); result.Err != nil {
		return nil, result.Err
	} else {
		return result.Data.([]*model.OAuthApp), nil
	}
}

func (app *App) GetOAuthImplicitRedirect(userId string, authRequest *model.AuthorizeRequest) (string, *model.AppError) {
	session, err := app.GetOAuthAccessTokenForImplicitFlow(userId, authRequest)
	if err != nil {
		return "", err
	}

	values := &url.Values{}
	values.Add("access_token", session.Token)
	values.Add("token_type", "bearer")
	values.Add("expires_in", strconv.FormatInt((session.ExpiresAt-model.GetMillis())/1000, 10))
	values.Add("scope", authRequest.Scope)
	values.Add("state", authRequest.State)

	return fmt.Sprintf("%s#%s", authRequest.RedirectUri, values.Encode()), nil
}

func (app *App) GetOAuthCodeRedirect(userId string, authRequest *model.AuthorizeRequest) (string, *model.AppError) {
	authData := &model.AuthData{UserId: userId, ClientId: authRequest.ClientId, CreateAt: model.GetMillis(), RedirectUri: authRequest.RedirectUri, State: authRequest.State, Scope: authRequest.Scope}
	authData.Code = model.NewId() + model.NewId()

	if result := <-app.Srv.Store.OAuth().SaveAuthData(authData); result.Err != nil {
		return authRequest.RedirectUri + "?error=server_error&state=" + authRequest.State, nil
	}

	return authRequest.RedirectUri + "?code=" + url.QueryEscape(authData.Code) + "&state=" + url.QueryEscape(authData.State), nil
}

func (app *App) AllowOAuthAppAccessToUser(userId string, authRequest *model.AuthorizeRequest) (string, *model.AppError) {
	if !*app.Config().ServiceSettings.EnableOAuthServiceProvider {
		return "", model.NewAppError("AllowOAuthAppAccessToUser", "api.oauth.allow_oauth.turn_off.app_error", nil, "", http.StatusNotImplemented)
	}

	if len(authRequest.Scope) == 0 {
		authRequest.Scope = model.DEFAULT_SCOPE
	}

	var oauthApp *model.OAuthApp
	if result := <-app.Srv.Store.OAuth().GetApp(authRequest.ClientId); result.Err != nil {
		return "", result.Err
	} else {
		oauthApp = result.Data.(*model.OAuthApp)
	}

	if !oauthApp.IsValidRedirectURL(authRequest.RedirectUri) {
		return "", model.NewAppError("AllowOAuthAppAccessToUser", "api.oauth.allow_oauth.redirect_callback.app_error", nil, "", http.StatusBadRequest)
	}

	var redirectURI string
	var err *model.AppError

	switch authRequest.ResponseType {
	case model.AUTHCODE_RESPONSE_TYPE:
		redirectURI, err = app.GetOAuthCodeRedirect(userId, authRequest)
	case model.IMPLICIT_RESPONSE_TYPE:
		redirectURI, err = app.GetOAuthImplicitRedirect(userId, authRequest)
	default:
		return authRequest.RedirectUri + "?error=unsupported_response_type&state=" + authRequest.State, nil
	}

	if err != nil {
		mlog.Error(err.Error())
		return authRequest.RedirectUri + "?error=server_error&state=" + authRequest.State, nil
	}

	// this saves the OAuth2 app as authorized
	authorizedApp := model.Preference{
		UserId:   userId,
		Category: model.PREFERENCE_CATEGORY_AUTHORIZED_OAUTH_APP,
		Name:     authRequest.ClientId,
		Value:    authRequest.Scope,
	}

	if result := <-app.Srv.Store.Preference().Save(&model.Preferences{authorizedApp}); result.Err != nil {
		mlog.Error(result.Err.Error())
		return authRequest.RedirectUri + "?error=server_error&state=" + authRequest.State, nil
	}

	return redirectURI, nil
}

func (app *App) GetOAuthAccessTokenForImplicitFlow(userId string, authRequest *model.AuthorizeRequest) (*model.Session, *model.AppError) {
	if !*app.Config().ServiceSettings.EnableOAuthServiceProvider {
		return nil, model.NewAppError("GetOAuthAccessToken", "api.oauth.get_access_token.disabled.app_error", nil, "", http.StatusNotImplemented)
	}

	var oauthApp *model.OAuthApp
	oauthApp, err := app.GetOAuthApp(authRequest.ClientId)
	if err != nil {
		return nil, model.NewAppError("GetOAuthAccessToken", "api.oauth.get_access_token.credentials.app_error", nil, "", http.StatusNotFound)
	}

	user, err := app.GetUser(userId)
	if err != nil {
		return nil, err
	}

	session, err := app.newSession(oauthApp.Name, user)
	if err != nil {
		return nil, err
	}

	accessData := &model.AccessData{ClientId: authRequest.ClientId, UserId: user.Id, Token: session.Token, RefreshToken: "", RedirectUri: authRequest.RedirectUri, ExpiresAt: session.ExpiresAt, Scope: authRequest.Scope}

	if result := <-app.Srv.Store.OAuth().SaveAccessData(accessData); result.Err != nil {
		mlog.Error(fmt.Sprint(result.Err))
		return nil, model.NewAppError("GetOAuthAccessToken", "api.oauth.get_access_token.internal_saving.app_error", nil, "", http.StatusInternalServerError)
	}

	return session, nil
}

func (app *App) GetOAuthAccessTokenForCodeFlow(clientId, grantType, redirectUri, code, secret, refreshToken string) (*model.AccessResponse, *model.AppError) {
	if !*app.Config().ServiceSettings.EnableOAuthServiceProvider {
		return nil, model.NewAppError("GetOAuthAccessToken", "api.oauth.get_access_token.disabled.app_error", nil, "", http.StatusNotImplemented)
	}

	var oauthApp *model.OAuthApp
	if result := <-app.Srv.Store.OAuth().GetApp(clientId); result.Err != nil {
		return nil, model.NewAppError("GetOAuthAccessToken", "api.oauth.get_access_token.credentials.app_error", nil, "", http.StatusNotFound)
	} else {
		oauthApp = result.Data.(*model.OAuthApp)
	}

	if oauthApp.ClientSecret != secret {
		return nil, model.NewAppError("GetOAuthAccessToken", "api.oauth.get_access_token.credentials.app_error", nil, "", http.StatusForbidden)
	}

	var user *model.User
	var accessData *model.AccessData
	var accessRsp *model.AccessResponse
	if grantType == model.ACCESS_TOKEN_GRANT_TYPE {

		var authData *model.AuthData
		if result := <-app.Srv.Store.OAuth().GetAuthData(code); result.Err != nil {
			return nil, model.NewAppError("GetOAuthAccessToken", "api.oauth.get_access_token.expired_code.app_error", nil, "", http.StatusInternalServerError)
		} else {
			authData = result.Data.(*model.AuthData)
		}

		if authData.IsExpired() {
			<-app.Srv.Store.OAuth().RemoveAuthData(authData.Code)
			return nil, model.NewAppError("GetOAuthAccessToken", "api.oauth.get_access_token.expired_code.app_error", nil, "", http.StatusForbidden)
		}

		if authData.RedirectUri != redirectUri {
			return nil, model.NewAppError("GetOAuthAccessToken", "api.oauth.get_access_token.redirect_uri.app_error", nil, "", http.StatusBadRequest)
		}

		if result := <-app.Srv.Store.User().Get(authData.UserId); result.Err != nil {
			return nil, model.NewAppError("GetOAuthAccessToken", "api.oauth.get_access_token.internal_user.app_error", nil, "", http.StatusNotFound)
		} else {
			user = result.Data.(*model.User)
		}

		if result := <-app.Srv.Store.OAuth().GetPreviousAccessData(user.Id, clientId); result.Err != nil {
			return nil, model.NewAppError("GetOAuthAccessToken", "api.oauth.get_access_token.internal.app_error", nil, "", http.StatusInternalServerError)
		} else if result.Data != nil {
			accessData := result.Data.(*model.AccessData)
			if accessData.IsExpired() {
				if access, err := app.newSessionUpdateToken(oauthApp.Name, accessData, user); err != nil {
					return nil, err
				} else {
					accessRsp = access
				}
			} else {
				//return the same token and no need to create a new session
				accessRsp = &model.AccessResponse{
					AccessToken:  accessData.Token,
					TokenType:    model.ACCESS_TOKEN_TYPE,
					RefreshToken: accessData.RefreshToken,
					ExpiresIn:    int32((accessData.ExpiresAt - model.GetMillis()) / 1000),
				}
			}
		} else {
			// create a new session and return new access token
			var session *model.Session
			if result, err := app.newSession(oauthApp.Name, user); err != nil {
				return nil, err
			} else {
				session = result
			}

			accessData = &model.AccessData{ClientId: clientId, UserId: user.Id, Token: session.Token, RefreshToken: model.NewId(), RedirectUri: redirectUri, ExpiresAt: session.ExpiresAt, Scope: authData.Scope}

			if result := <-app.Srv.Store.OAuth().SaveAccessData(accessData); result.Err != nil {
				mlog.Error(fmt.Sprint(result.Err))
				return nil, model.NewAppError("GetOAuthAccessToken", "api.oauth.get_access_token.internal_saving.app_error", nil, "", http.StatusInternalServerError)
			}

			accessRsp = &model.AccessResponse{
				AccessToken:  session.Token,
				TokenType:    model.ACCESS_TOKEN_TYPE,
				RefreshToken: accessData.RefreshToken,
				ExpiresIn:    int32(*app.Config().ServiceSettings.SessionLengthSSOInDays * 60 * 60 * 24),
			}
		}

		<-app.Srv.Store.OAuth().RemoveAuthData(authData.Code)
	} else {
		// when grantType is refresh_token
		if result := <-app.Srv.Store.OAuth().GetAccessDataByRefreshToken(refreshToken); result.Err != nil {
			return nil, model.NewAppError("GetOAuthAccessToken", "api.oauth.get_access_token.refresh_token.app_error", nil, "", http.StatusNotFound)
		} else {
			accessData = result.Data.(*model.AccessData)
		}

		if result := <-app.Srv.Store.User().Get(accessData.UserId); result.Err != nil {
			return nil, model.NewAppError("GetOAuthAccessToken", "api.oauth.get_access_token.internal_user.app_error", nil, "", http.StatusNotFound)
		} else {
			user = result.Data.(*model.User)
		}

		if access, err := app.newSessionUpdateToken(oauthApp.Name, accessData, user); err != nil {
			return nil, err
		} else {
			accessRsp = access
		}
	}

	return accessRsp, nil
}

func (app *App) newSession(appName string, user *model.User) (*model.Session, *model.AppError) {
	// set new token an session
	session := &model.Session{UserId: user.Id, Roles: user.Roles, IsOAuth: true}
	session.GenerateCSRF()
	session.SetExpireInDays(*app.Config().ServiceSettings.SessionLengthSSOInDays)
	session.AddProp(model.SESSION_PROP_PLATFORM, appName)
	session.AddProp(model.SESSION_PROP_OS, "OAuth2")
	session.AddProp(model.SESSION_PROP_BROWSER, "OAuth2")

	if result := <-app.Srv.Store.Session().Save(session); result.Err != nil {
		return nil, model.NewAppError("newSession", "api.oauth.get_access_token.internal_session.app_error", nil, "", http.StatusInternalServerError)
	} else {
		session = result.Data.(*model.Session)
		app.AddSessionToCache(session)
	}

	return session, nil
}

func (app *App) newSessionUpdateToken(appName string, accessData *model.AccessData, user *model.User) (*model.AccessResponse, *model.AppError) {
	var session *model.Session
	<-app.Srv.Store.Session().Remove(accessData.Token) //remove the previous session

	if result, err := app.newSession(appName, user); err != nil {
		return nil, err
	} else {
		session = result
	}

	accessData.Token = session.Token
	accessData.RefreshToken = model.NewId()
	accessData.ExpiresAt = session.ExpiresAt
	if result := <-app.Srv.Store.OAuth().UpdateAccessData(accessData); result.Err != nil {
		mlog.Error(fmt.Sprint(result.Err))
		return nil, model.NewAppError("newSessionUpdateToken", "web.get_access_token.internal_saving.app_error", nil, "", http.StatusInternalServerError)
	}
	accessRsp := &model.AccessResponse{
		AccessToken:  session.Token,
		RefreshToken: accessData.RefreshToken,
		TokenType:    model.ACCESS_TOKEN_TYPE,
		ExpiresIn:    int32(*app.Config().ServiceSettings.SessionLengthSSOInDays * 60 * 60 * 24),
	}

	return accessRsp, nil
}

func (app *App) GetOAuthLoginEndpoint(w http.ResponseWriter, r *http.Request, service, teamId, action, redirectTo, loginHint string) (string, *model.AppError) {
	stateProps := map[string]string{}
	stateProps["action"] = action
	if len(teamId) != 0 {
		stateProps["team_id"] = teamId
	}

	if len(redirectTo) != 0 {
		stateProps["redirect_to"] = redirectTo
	}

	if authUrl, err := app.GetAuthorizationCode(w, r, service, stateProps, loginHint); err != nil {
		return "", err
	} else {
		fmt.Println("authUrl", authUrl)
		return authUrl, nil
	}
}

func (app *App) GetOAuthSignupEndpoint(w http.ResponseWriter, r *http.Request, service, teamId string) (string, *model.AppError) {
	stateProps := map[string]string{}
	stateProps["action"] = model.OAUTH_ACTION_SIGNUP
	if len(teamId) != 0 {
		stateProps["team_id"] = teamId
	}

	if authUrl, err := app.GetAuthorizationCode(w, r, service, stateProps, ""); err != nil {
		return "", err
	} else {
		return authUrl, nil
	}
}

func (app *App) GetAuthorizedAppsForUser(userId string, page, perPage int) ([]*model.OAuthApp, *model.AppError) {
	if !*app.Config().ServiceSettings.EnableOAuthServiceProvider {
		return nil, model.NewAppError("GetAuthorizedAppsForUser", "api.oauth.allow_oauth.turn_off.app_error", nil, "", http.StatusNotImplemented)
	}

	if result := <-app.Srv.Store.OAuth().GetAuthorizedApps(userId, page*perPage, perPage); result.Err != nil {
		return nil, result.Err
	} else {
		apps := result.Data.([]*model.OAuthApp)
		for k, a := range apps {
			a.Sanitize()
			apps[k] = a
		}

		return apps, nil
	}
}

func (app *App) DeauthorizeOAuthAppForUser(userId, appId string) *model.AppError {
	if !*app.Config().ServiceSettings.EnableOAuthServiceProvider {
		return model.NewAppError("DeauthorizeOAuthAppForUser", "api.oauth.allow_oauth.turn_off.app_error", nil, "", http.StatusNotImplemented)
	}

	// revoke app sessions
	if result := <-app.Srv.Store.OAuth().GetAccessDataByUserForApp(userId, appId); result.Err != nil {
		return result.Err
	} else {
		accessData := result.Data.([]*model.AccessData)

		for _, ad := range accessData {
			if err := app.RevokeAccessToken(ad.Token); err != nil {
				return err
			}

			if rad := <-app.Srv.Store.OAuth().RemoveAccessData(ad.Token); rad.Err != nil {
				return rad.Err
			}
		}
	}

	// Deauthorize the app
	if err := (<-app.Srv.Store.Preference().Delete(userId, model.PREFERENCE_CATEGORY_AUTHORIZED_OAUTH_APP, appId)).Err; err != nil {
		return err
	}

	return nil
}

func (app *App) RegenerateOAuthAppSecret(a *model.OAuthApp) (*model.OAuthApp, *model.AppError) {
	if !*app.Config().ServiceSettings.EnableOAuthServiceProvider {
		return nil, model.NewAppError("RegenerateOAuthAppSecret", "api.oauth.allow_oauth.turn_off.app_error", nil, "", http.StatusNotImplemented)
	}

	a.ClientSecret = model.NewId()
	if update := <-app.Srv.Store.OAuth().UpdateApp(a); update.Err != nil {
		return nil, update.Err
	}

	return a, nil
}

func (app *App) RevokeAccessToken(token string) *model.AppError {
	session, _ := app.GetSession(token)
	schan := app.Srv.Store.Session().Remove(token)

	if result := <-app.Srv.Store.OAuth().GetAccessData(token); result.Err != nil {
		return model.NewAppError("RevokeAccessToken", "api.oauth.revoke_access_token.get.app_error", nil, "", http.StatusBadRequest)
	}

	tchan := app.Srv.Store.OAuth().RemoveAccessData(token)

	if result := <-tchan; result.Err != nil {
		return model.NewAppError("RevokeAccessToken", "api.oauth.revoke_access_token.del_token.app_error", nil, "", http.StatusInternalServerError)
	}

	if result := <-schan; result.Err != nil {
		return model.NewAppError("RevokeAccessToken", "api.oauth.revoke_access_token.del_session.app_error", nil, "", http.StatusInternalServerError)
	}

	if session != nil {
		app.ClearSessionCacheForUser(session.UserId)
	}

	return nil
}

func (app *App) CompleteOAuth(service string, body io.ReadCloser, facebookToken string, teamId string, props map[string]string) (*model.User, *facebookgraph.FacebookError, *model.AppError) {
	defer body.Close()

	action := props["action"]

	switch action {
	case model.OAUTH_ACTION_SIGNUP:
		// đăng ký thành viên mới. Hiện tại không áp dụng mà xử lý trong case tiếp theo dưới đây
		// người dùng chỉ cần bấm vào login là sẽ được đăng ký nếu chưa có tài khoản
		return app.CreateOAuthUser(service, body, facebookToken, teamId)
	case model.OAUTH_ACTION_LOGIN:
		return app.LoginByOAuth(service, body, facebookToken, teamId)
	case model.OAUTH_ACTION_EMAIL_TO_SSO:
		return app.CompleteSwitchWithOAuth(service, body, props["email"])
	case model.OAUTH_ACTION_SSO_TO_EMAIL:
		return app.LoginByOAuth(service, body, facebookToken, teamId)
	default:
		return app.LoginByOAuth(service, body, facebookToken, teamId)
	}
}

func (app *App) LoginByOAuth(service string, userData io.Reader, facebookToken string, teamId string) (*model.User, *facebookgraph.FacebookError, *model.AppError) {

	// extend facebook token
	if longLiveToken, err, _ := app.ExtendFacebookToken(facebookToken); err != nil {
		// không thể extend token, chúng ta sẽ bỏ qua trường hợp này
		// và lưu token ngắn hạn hiện tại vào database
	} else {
		facebookToken = longLiveToken.AccessToken
	}

	buf := bytes.Buffer{}
	buf.ReadFrom(userData)

	authData := ""
	provider := einterfaces.GetOauthProvider(service)
	if provider == nil {
		return nil, nil, model.NewAppError("LoginByOAuth", "api.user.login_by_oauth.not_available.app_error",
			map[string]interface{}{"Service": strings.Title(service)}, "", http.StatusNotImplemented)
	} else {
		authUser := provider.GetUserFromJson(bytes.NewReader(buf.Bytes()))

		if authUser.AuthData != nil {
			authData = *authUser.AuthData
		} else {
			authData = ""
		}
	}

	if len(authData) == 0 {
		return nil, nil, model.NewAppError("LoginByOAuth", "api.user.login_by_oauth.parse.app_error",
			map[string]interface{}{"Service": service}, "", http.StatusBadRequest)
	}

	var user *model.User
	var err *model.AppError

	isExist := false

	user, err = app.GetUserByAuth(&authData, service)
	if err != nil {
		if err.Id == store.MISSING_AUTH_ACCOUNT_ERROR {
			// chưa có tài khoản => tạo mới user
			fmt.Println("User chưa từng đăng nhập trên hệ thống, tạo mới tài khoản")
			if newUser, _, newErr := app.CreateOAuthUser(service, bytes.NewReader(buf.Bytes()), facebookToken, teamId); newErr != nil {
				// không thể tạo mới user => trả về lỗi
				return nil, nil, newErr
			} else {
				user = newUser
			}
		} else {
			// Một lỗi không xác định => trả về lỗi
			return nil, nil, err
		}
	}

	isExist = true

	// đã có tài khoản => update một số attribute của user nếu có thay đổi
	// ví dụ người dùng facebook có thể đổi tên, chúng ta phải cập nhật tên mới của họ vào
	// hay mối lần đăng nhập đều có 1 access_token mới từ facebook
	// chúng ta cũng cần cập nhật access token này vào database
	if isExist {
		if err = app.UpdateOAuthUserAttrs(bytes.NewReader(buf.Bytes()), user, provider, service, facebookToken); err != nil {
			return nil, nil, err
		}
	}

	// Thêm tất cả các pages của user vào db nếu chưa có
	fbPages, fbErr, aErr := app.GraphFanpages(facebookToken)

	if fbErr != nil {
		fmt.Println("Loi lay danh sach fanpages")
	}

	if aErr != nil {
		fmt.Println("Loi he thong")
	}

	//fmt.Println("fbPages", fbPages)
	for _, fPage := range fbPages {
		// thêm fanpage vào db
		isExist, rPage, err := app.UpsertPageFromFacebookPage(fPage)
		if err != nil {
			// Lỗi không thêm được page vào db
			fmt.Println("err", err)
		}

		if isExist && rPage != nil {
			// page đã tồn tại, update page token
			fmt.Println("page đã tồn tại, update page token ", rPage.PageId)
		}

		// đồng thời thêm member page
		pageMember := model.FanpageMember{FanpageId: rPage.Id, PageId: rPage.PageId, UserId: user.Id, AccessToken: fPage.AccessToken}
		if memberResult := <-app.Srv.Store.Fanpage().SaveFanPageMember(&pageMember); memberResult.Err != nil {
			//return "", nil, memberResult.Err
			fmt.Println("Không thể thêm fanpage member, lỗi: ", memberResult.Err.Message)
		} else {
			fmt.Println("Lưu fanpage member thành công")
		}
	}

	return user, nil, nil
}

func (app *App) CompleteSwitchWithOAuth(service string, userData io.ReadCloser, email string) (*model.User, *facebookgraph.FacebookError, *model.AppError) {
	authData := ""
	ssoEmail := ""
	provider := einterfaces.GetOauthProvider(service)
	if provider == nil {
		return nil, nil, model.NewAppError("CompleteSwitchWithOAuth", "api.user.complete_switch_with_oauth.unavailable.app_error",
			map[string]interface{}{"Service": strings.Title(service)}, "", http.StatusNotImplemented)
	} else {
		ssoUser := provider.GetUserFromJson(userData)
		ssoEmail = ssoUser.Email

		if ssoUser.AuthData != nil {
			authData = *ssoUser.AuthData
		}
	}

	if len(authData) == 0 {
		return nil, nil, model.NewAppError("CompleteSwitchWithOAuth", "api.user.complete_switch_with_oauth.parse.app_error",
			map[string]interface{}{"Service": service}, "", http.StatusBadRequest)
	}

	if len(email) == 0 {
		return nil, nil, model.NewAppError("CompleteSwitchWithOAuth", "api.user.complete_switch_with_oauth.blank_email.app_error", nil, "", http.StatusBadRequest)
	}

	var user *model.User
	if result := <-app.Srv.Store.User().GetByEmail(email); result.Err != nil {
		return nil, nil, result.Err
	} else {
		user = result.Data.(*model.User)
	}

	if err := app.RevokeAllSessions(user.Id); err != nil {
		return nil, nil, err
	}

	if result := <-app.Srv.Store.User().UpdateAuthData(user.Id, service, &authData, ssoEmail, true); result.Err != nil {
		return nil, nil, result.Err
	}

	//a.Go(func() {
	//	if err := a.SendSignInChangeEmail(user.Email, strings.Title(service)+" SSO", user.Locale, a.GetSiteURL()); err != nil {
	//		mlog.Error(err.Error())
	//	}
	//})

	return user, nil, nil
}

func (app *App) CreateOAuthStateToken(extra string) (*model.Token, *model.AppError) {
	token := model.NewToken(model.TOKEN_TYPE_OAUTH, extra)

	if result := <-app.Srv.Store.Token().Save(token); result.Err != nil {
		return nil, result.Err
	}

	return token, nil
}

func (app *App) GetOAuthStateToken(token string) (*model.Token, *model.AppError) {
	if result := <-app.Srv.Store.Token().GetByToken(token); result.Err != nil {
		return nil, model.NewAppError("GetOAuthStateToken", "api.oauth.invalid_state_token.app_error", nil, result.Err.Error(), http.StatusBadRequest)
	} else {
		token := result.Data.(*model.Token)
		if token.Type != model.TOKEN_TYPE_OAUTH {
			return nil, model.NewAppError("GetOAuthStateToken", "api.oauth.invalid_state_token.app_error", nil, "", http.StatusBadRequest)
		}

		return token, nil
	}
}

func generateOAuthStateTokenExtra(email, action, cookie string) string {
	return email + ":" + action + ":" + cookie
}

func (app *App) GetAuthorizationCode(w http.ResponseWriter, r *http.Request, service string, props map[string]string, loginHint string) (string, *model.AppError) {
	sso := app.Config().GetSSOService(service)
	if sso == nil || !*sso.Enable {
		return "", model.NewAppError("GetAuthorizationCode", "api.user.get_authorization_code.unsupported.app_error", nil, "service="+service, http.StatusNotImplemented)
	}

	secure := false
	if GetProtocol(r) == "https" {
		secure = true
	}

	cookieValue := model.NewId()
	expiresAt := time.Unix(model.GetMillis()/1000+int64(OAUTH_COOKIE_MAX_AGE_SECONDS), 0)
	oauthCookie := &http.Cookie{
		Name:     COOKIE_OAUTH,
		Value:    cookieValue,
		Path:     "/",
		MaxAge:   OAUTH_COOKIE_MAX_AGE_SECONDS,
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   secure,
	}

	http.SetCookie(w, oauthCookie)

	clientId := *sso.Id
	endpoint := *sso.AuthEndpoint
	scope := *sso.Scope

	tokenExtra := generateOAuthStateTokenExtra(props["email"], props["action"], cookieValue)
	stateToken, err := app.CreateOAuthStateToken(tokenExtra)
	if err != nil {
		return "", err
	}

	props["token"] = stateToken.Token
	state := b64.StdEncoding.EncodeToString([]byte(model.MapToJson(props)))

	siteUrl := app.GetSiteURL()
	if strings.TrimSpace(siteUrl) == "" {
		siteUrl = GetProtocol(r) + "://" + r.Host
	}

	redirectUri := siteUrl + "/signup/" + service + "/complete"

	authUrl := endpoint + "?response_type=code&client_id=" + clientId + "&redirect_uri=" + url.QueryEscape(redirectUri) + "&state=" + url.QueryEscape(state)

	if len(scope) > 0 {
		authUrl += "&scope=" + utils.UrlEncode(scope)
	}

	if len(loginHint) > 0 {
		authUrl += "&login_hint=" + utils.UrlEncode(loginHint)
	}

	return authUrl, nil
}

// trả về 4 tham số
// 1. io.ReadCloser
// 2. string: facebook access token
// 3. string: team id, hiện tại chưa dùng cái này nhưng sau này có trường hợp phải dùng đến, ví dụ mời user vào 1 team
// map[string]string: stateProps
// *model.AppError nếu có lỗi
func (a *App) AuthorizeOAuthUser(w http.ResponseWriter, r *http.Request, service, code, state, redirectUri string) (io.ReadCloser, string, string, map[string]string, *model.AppError) {
	sso := a.Config().GetSSOService(service)
	if sso == nil || !*sso.Enable {
		return nil, "", "", nil, model.NewAppError("AuthorizeOAuthUser", "api.user.authorize_oauth_user.unsupported.app_error", nil, "service="+service, http.StatusNotImplemented)
	}

	b, strErr := b64.StdEncoding.DecodeString(state)
	if strErr != nil {
		return nil, "", "", nil, model.NewAppError("AuthorizeOAuthUser", "api.user.authorize_oauth_user.invalid_state.app_error", nil, strErr.Error(), http.StatusBadRequest)
	}

	stateStr := string(b)

	stateProps := model.MapFromJson(strings.NewReader(stateStr))

	fmt.Println("==========================================================")

	expectedToken, appErr := a.GetOAuthStateToken(stateProps["token"])
	if appErr != nil {
		return nil, "", "", stateProps, appErr
	}

	stateEmail := stateProps["email"]
	stateAction := stateProps["action"]
	if stateAction == model.OAUTH_ACTION_EMAIL_TO_SSO && stateEmail == "" {
		return nil, "", "", stateProps, model.NewAppError("AuthorizeOAuthUser", "api.user.authorize_oauth_user.invalid_state.app_error", nil, "", http.StatusBadRequest)
	}

	cookie, cookieErr := r.Cookie(COOKIE_OAUTH)
	if cookieErr != nil {
		return nil, "", "", stateProps, model.NewAppError("AuthorizeOAuthUser", "api.user.authorize_oauth_user.invalid_state.app_error", nil, "", http.StatusBadRequest)
	}

	expectedTokenExtra := generateOAuthStateTokenExtra(stateEmail, stateAction, cookie.Value)
	if expectedTokenExtra != expectedToken.Extra {
		return nil, "", "", stateProps, model.NewAppError("AuthorizeOAuthUser", "api.user.authorize_oauth_user.invalid_state.app_error", nil, "", http.StatusBadRequest)
	}

	appErr = a.DeleteToken(expectedToken)
	if appErr != nil {
		mlog.Error(appErr.Error())
	}

	httpCookie := &http.Cookie{
		Name:     COOKIE_OAUTH,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	}

	http.SetCookie(w, httpCookie)

	teamId := stateProps["team_id"]

	p := url.Values{}
	p.Set("client_id", *sso.Id)
	p.Set("client_secret", *sso.Secret)
	p.Set("code", code)
	p.Set("grant_type", model.ACCESS_TOKEN_GRANT_TYPE)
	p.Set("redirect_uri", redirectUri)

	req, requestErr := http.NewRequest("POST", *sso.TokenEndpoint, strings.NewReader(p.Encode()))
	if requestErr != nil {
		return nil, "", "", stateProps, model.NewAppError("AuthorizeOAuthUser", "api.user.authorize_oauth_user.token_failed.app_error", nil, requestErr.Error(), http.StatusInternalServerError)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := a.HTTPService.MakeClient(true).Do(req)
	if err != nil {
		return nil, "", "", stateProps, model.NewAppError("AuthorizeOAuthUser", "api.user.authorize_oauth_user.token_failed.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	defer resp.Body.Close()

	var buf bytes.Buffer
	tee := io.TeeReader(resp.Body, &buf)
	ar := model.AccessResponseFromJson(tee)

	if ar == nil || resp.StatusCode != http.StatusOK {
		return nil, "", "", stateProps, model.NewAppError("AuthorizeOAuthUser", "api.user.authorize_oauth_user.bad_response.app_error", nil, "response_body="+buf.String(), http.StatusInternalServerError)
	}

	if strings.ToLower(ar.TokenType) != model.ACCESS_TOKEN_TYPE {
		return nil, "", "", stateProps, model.NewAppError("AuthorizeOAuthUser", "api.user.authorize_oauth_user.bad_token.app_error", nil, "token_type="+ar.TokenType+", response_body="+buf.String(), http.StatusInternalServerError)
	}

	if len(ar.AccessToken) == 0 {
		return nil, "", "", stateProps, model.NewAppError("AuthorizeOAuthUser", "api.user.authorize_oauth_user.missing.app_error", nil, "response_body="+buf.String(), http.StatusInternalServerError)
	}

	p = url.Values{}
	p.Set("access_token", ar.AccessToken)
	req, requestErr = http.NewRequest("GET", *sso.UserApiEndpoint, strings.NewReader(""))
	if requestErr != nil {
		return nil, "", "", stateProps, model.NewAppError("AuthorizeOAuthUser", "api.user.authorize_oauth_user.service.app_error", map[string]interface{}{"Service": service}, requestErr.Error(), http.StatusInternalServerError)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+ar.AccessToken)

	resp, err = a.HTTPService.MakeClient(true).Do(req)
	if err != nil {
		return nil, "", "", stateProps, model.NewAppError("AuthorizeOAuthUser", "api.user.authorize_oauth_user.service.app_error", map[string]interface{}{"Service": service}, err.Error(), http.StatusInternalServerError)
	} else if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()

		// Ignore the error below because the resulting string will just be the empty string if bodyBytes is nil
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		bodyString := string(bodyBytes)

		mlog.Error("Error getting OAuth user: " + bodyString)

		if service == model.SERVICE_GITLAB && resp.StatusCode == http.StatusForbidden && strings.Contains(bodyString, "Terms of Service") {
			// Return a nicer error when the user hasn't accepted GitLab's terms of service
			return nil, "", "", stateProps, model.NewAppError("AuthorizeOAuthUser", "oauth.gitlab.tos.error", nil, "", http.StatusBadRequest)
		}

		return nil, "", "", stateProps, model.NewAppError("AuthorizeOAuthUser", "api.user.authorize_oauth_user.response.app_error", nil, "response_body="+bodyString, http.StatusInternalServerError)
	}

	// Note that resp.Body is not closed here, so it must be closed by the caller
	return resp.Body, ar.AccessToken, teamId, stateProps, nil

}

func (app *App) SwitchEmailToOAuth(w http.ResponseWriter, r *http.Request, email, password, code, service string) (string, *model.AppError) {
	if app.License() != nil && !*app.Config().ServiceSettings.ExperimentalEnableAuthenticationTransfer {
		return "", model.NewAppError("emailToOAuth", "api.user.email_to_oauth.not_available.app_error", nil, "", http.StatusForbidden)
	}

	var user *model.User
	var err *model.AppError
	if user, err = app.GetUserByEmail(email); err != nil {
		return "", err
	}

	if err := app.CheckPasswordAndAllCriteria(user, password, code); err != nil {
		return "", err
	}

	stateProps := map[string]string{}
	stateProps["action"] = model.OAUTH_ACTION_EMAIL_TO_SSO
	stateProps["email"] = email

	if service == model.USER_AUTH_SERVICE_SAML {
		return app.GetSiteURL() + "/login/sso/saml?action=" + model.OAUTH_ACTION_EMAIL_TO_SSO + "&email=" + utils.UrlEncode(email), nil
	} else {
		if authUrl, err := app.GetAuthorizationCode(w, r, service, stateProps, ""); err != nil {
			return "", err
		} else {
			return authUrl, nil
		}
	}
}

func (app *App) SwitchOAuthToEmail(email, password, requesterId string) (string, *model.AppError) {
	if app.License() != nil && !*app.Config().ServiceSettings.ExperimentalEnableAuthenticationTransfer {
		return "", model.NewAppError("oauthToEmail", "api.user.oauth_to_email.not_available.app_error", nil, "", http.StatusForbidden)
	}

	var user *model.User
	var err *model.AppError
	if user, err = app.GetUserByEmail(email); err != nil {
		return "", err
	}

	if user.Id != requesterId {
		return "", model.NewAppError("SwitchOAuthToEmail", "api.user.oauth_to_email.context.app_error", nil, "", http.StatusForbidden)
	}

	if err := app.UpdatePassword(user, password); err != nil {
		return "", err
	}

	//T := utils.GetUserTranslations(user.Locale)

	//a.Go(func() {
	//	if err := a.SendSignInChangeEmail(user.Email, T("api.templates.signin_change_email.body.method_email"), user.Locale, a.GetSiteURL()); err != nil {
	//		mlog.Error(err.Error())
	//	}
	//})

	if err := app.RevokeAllSessions(requesterId); err != nil {
		return "", err
	}

	return "/login?extra=signin_change", nil
}
