// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package model

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	HEADER_REQUEST_ID         = "X-Request-ID"
	HEADER_VERSION_ID         = "X-Version-ID"
	HEADER_CLUSTER_ID         = "X-Cluster-ID"
	HEADER_ETAG_SERVER        = "ETag"
	HEADER_ETAG_CLIENT        = "If-None-Match"
	HEADER_FORWARDED          = "X-Forwarded-For"
	HEADER_REAL_IP            = "X-Real-IP"
	HEADER_FORWARDED_PROTO    = "X-Forwarded-Proto"
	HEADER_TOKEN              = "token"
	HEADER_CSRF_TOKEN         = "X-CSRF-Token"
	HEADER_BEARER             = "BEARER"
	HEADER_AUTH               = "Authorization"
	HEADER_REQUESTED_WITH     = "X-Requested-With"
	HEADER_REQUESTED_WITH_XML = "XMLHttpRequest"
	STATUS                    = "status"
	STATUS_OK                 = "OK"
	STATUS_FAIL               = "FAIL"
	STATUS_REMOVE             = "REMOVE"

	CLIENT_DIR = "client"

	API_URL_SUFFIX_V1 = "/api/v1"
	API_URL_SUFFIX_V4 = "/api/v4"
	API_URL_SUFFIX    = API_URL_SUFFIX_V1
)

type Response struct {
	StatusCode    int
	Error         *AppError
	RequestId     string
	Etag          string
	ServerVersion string
	Header        http.Header
}

type Client4 struct {
	Url        string       // The location of the server, for example  "http://localhost:8065"
	ApiUrl     string       // The api location of the server, for example "http://localhost:8065/api/v4"
	HttpClient *http.Client // The http client
	AuthToken  string
	AuthType   string
	HttpHeader map[string]string // Headers to be copied over for each request
}

func closeBody(r *http.Response) {
	if r.Body != nil {
		ioutil.ReadAll(r.Body)
		r.Body.Close()
	}
}

// Must is a convenience function used for testing.
func (c *Client4) Must(result interface{}, resp *Response) interface{} {
	if resp.Error != nil {

		time.Sleep(time.Second)
		panic(resp.Error)
	}

	return result
}

func NewAPIv4Client(url string) *Client4 {
	return &Client4{url, url + API_URL_SUFFIX, &http.Client{}, "", "", map[string]string{}}
}

func BuildErrorResponse(r *http.Response, err *AppError) *Response {
	var statusCode int
	var header http.Header
	if r != nil {
		statusCode = r.StatusCode
		header = r.Header
	} else {
		statusCode = 0
		header = make(http.Header)
	}

	return &Response{
		StatusCode: statusCode,
		Error:      err,
		Header:     header,
	}
}

func BuildResponse(r *http.Response) *Response {
	return &Response{
		StatusCode:    r.StatusCode,
		RequestId:     r.Header.Get(HEADER_REQUEST_ID),
		Etag:          r.Header.Get(HEADER_ETAG_SERVER),
		ServerVersion: r.Header.Get(HEADER_VERSION_ID),
		Header:        r.Header,
	}
}

func (c *Client4) MockSession(sessionToken string) {
	c.AuthToken = sessionToken
	c.AuthType = HEADER_BEARER
}

func (c *Client4) SetOAuthToken(token string) {
	c.AuthToken = token
	c.AuthType = HEADER_TOKEN
}

func (c *Client4) ClearOAuthToken() {
	c.AuthToken = ""
	c.AuthType = HEADER_BEARER
}

func (c *Client4) GetUsersRoute() string {
	return fmt.Sprintf("/users")
}

func (c *Client4) GetUserRoute(userId string) string {
	return fmt.Sprintf(c.GetUsersRoute()+"/%v", userId)
}

func (c *Client4) GetUserAccessTokensRoute() string {
	return fmt.Sprintf(c.GetUsersRoute() + "/tokens")
}

func (c *Client4) GetUserAccessTokenRoute(tokenId string) string {
	return fmt.Sprintf(c.GetUsersRoute()+"/tokens/%v", tokenId)
}

func (c *Client4) GetUserByUsernameRoute(userName string) string {
	return fmt.Sprintf(c.GetUsersRoute()+"/username/%v", userName)
}

func (c *Client4) GetUserByEmailRoute(email string) string {
	return fmt.Sprintf(c.GetUsersRoute()+"/email/%v", email)
}

func (c *Client4) GetTeamsRoute() string {
	return fmt.Sprintf("/teams")
}

func (c *Client4) GetTeamRoute(teamId string) string {
	return fmt.Sprintf(c.GetTeamsRoute()+"/%v", teamId)
}

func (c *Client4) GetTeamAutoCompleteCommandsRoute(teamId string) string {
	return fmt.Sprintf(c.GetTeamsRoute()+"/%v/commands/autocomplete", teamId)
}

func (c *Client4) GetTeamByNameRoute(teamName string) string {
	return fmt.Sprintf(c.GetTeamsRoute()+"/name/%v", teamName)
}

func (c *Client4) GetTeamMemberRoute(teamId, userId string) string {
	return fmt.Sprintf(c.GetTeamRoute(teamId)+"/members/%v", userId)
}

func (c *Client4) GetTeamMembersRoute(teamId string) string {
	return fmt.Sprintf(c.GetTeamRoute(teamId) + "/members")
}

func (c *Client4) GetTeamStatsRoute(teamId string) string {
	return fmt.Sprintf(c.GetTeamRoute(teamId) + "/stats")
}

func (c *Client4) GetTeamImportRoute(teamId string) string {
	return fmt.Sprintf(c.GetTeamRoute(teamId) + "/import")
}

func (c *Client4) GetChannelsRoute() string {
	return fmt.Sprintf("/channels")
}

func (c *Client4) GetChannelsForTeamRoute(teamId string) string {
	return fmt.Sprintf(c.GetTeamRoute(teamId) + "/channels")
}

func (c *Client4) GetChannelRoute(channelId string) string {
	return fmt.Sprintf(c.GetChannelsRoute()+"/%v", channelId)
}

func (c *Client4) GetChannelByNameRoute(channelName, teamId string) string {
	return fmt.Sprintf(c.GetTeamRoute(teamId)+"/channels/name/%v", channelName)
}

func (c *Client4) GetChannelByNameForTeamNameRoute(channelName, teamName string) string {
	return fmt.Sprintf(c.GetTeamByNameRoute(teamName)+"/channels/name/%v", channelName)
}

func (c *Client4) GetChannelMembersRoute(channelId string) string {
	return fmt.Sprintf(c.GetChannelRoute(channelId) + "/members")
}

func (c *Client4) GetChannelMemberRoute(channelId, userId string) string {
	return fmt.Sprintf(c.GetChannelMembersRoute(channelId)+"/%v", userId)
}

func (c *Client4) GetPostsRoute() string {
	return fmt.Sprintf("/posts")
}

func (c *Client4) GetPostsEphemeralRoute() string {
	return fmt.Sprintf("/posts/ephemeral")
}

func (c *Client4) GetConfigRoute() string {
	return fmt.Sprintf("/config")
}

func (c *Client4) GetLicenseRoute() string {
	return fmt.Sprintf("/license")
}

func (c *Client4) GetPostRoute(postId string) string {
	return fmt.Sprintf(c.GetPostsRoute()+"/%v", postId)
}

func (c *Client4) GetFilesRoute() string {
	return fmt.Sprintf("/files")
}

func (c *Client4) GetFileRoute(fileId string) string {
	return fmt.Sprintf(c.GetFilesRoute()+"/%v", fileId)
}

func (c *Client4) GetPluginsRoute() string {
	return fmt.Sprintf("/plugins")
}

func (c *Client4) GetPluginRoute(pluginId string) string {
	return fmt.Sprintf(c.GetPluginsRoute()+"/%v", pluginId)
}

func (c *Client4) GetSystemRoute() string {
	return fmt.Sprintf("/system")
}

func (c *Client4) GetTestEmailRoute() string {
	return fmt.Sprintf("/email/test")
}

func (c *Client4) GetTestS3Route() string {
	return fmt.Sprintf("/file/s3_test")
}

func (c *Client4) GetDatabaseRoute() string {
	return fmt.Sprintf("/database")
}

func (c *Client4) GetCacheRoute() string {
	return fmt.Sprintf("/caches")
}

func (c *Client4) GetClusterRoute() string {
	return fmt.Sprintf("/cluster")
}

func (c *Client4) GetIncomingWebhooksRoute() string {
	return fmt.Sprintf("/hooks/incoming")
}

func (c *Client4) GetIncomingWebhookRoute(hookID string) string {
	return fmt.Sprintf(c.GetIncomingWebhooksRoute()+"/%v", hookID)
}

func (c *Client4) GetComplianceReportsRoute() string {
	return fmt.Sprintf("/compliance/reports")
}

func (c *Client4) GetComplianceReportRoute(reportId string) string {
	return fmt.Sprintf("/compliance/reports/%v", reportId)
}

func (c *Client4) GetOutgoingWebhooksRoute() string {
	return fmt.Sprintf("/hooks/outgoing")
}

func (c *Client4) GetOutgoingWebhookRoute(hookID string) string {
	return fmt.Sprintf(c.GetOutgoingWebhooksRoute()+"/%v", hookID)
}

func (c *Client4) GetPreferencesRoute(userId string) string {
	return fmt.Sprintf(c.GetUserRoute(userId) + "/preferences")
}

func (c *Client4) GetUserStatusRoute(userId string) string {
	return fmt.Sprintf(c.GetUserRoute(userId) + "/status")
}

func (c *Client4) GetUserStatusesRoute() string {
	return fmt.Sprintf(c.GetUsersRoute() + "/status")
}

func (c *Client4) GetSamlRoute() string {
	return fmt.Sprintf("/saml")
}

func (c *Client4) GetLdapRoute() string {
	return fmt.Sprintf("/ldap")
}

func (c *Client4) GetBrandRoute() string {
	return fmt.Sprintf("/brand")
}

func (c *Client4) GetDataRetentionRoute() string {
	return fmt.Sprintf("/data_retention")
}

func (c *Client4) GetElasticsearchRoute() string {
	return fmt.Sprintf("/elasticsearch")
}

func (c *Client4) GetCommandsRoute() string {
	return fmt.Sprintf("/commands")
}

func (c *Client4) GetCommandRoute(commandId string) string {
	return fmt.Sprintf(c.GetCommandsRoute()+"/%v", commandId)
}

func (c *Client4) GetEmojisRoute() string {
	return fmt.Sprintf("/emoji")
}

func (c *Client4) GetEmojiRoute(emojiId string) string {
	return fmt.Sprintf(c.GetEmojisRoute()+"/%v", emojiId)
}

func (c *Client4) GetEmojiByNameRoute(name string) string {
	return fmt.Sprintf(c.GetEmojisRoute()+"/name/%v", name)
}

func (c *Client4) GetReactionsRoute() string {
	return fmt.Sprintf("/reactions")
}

func (c *Client4) GetOAuthAppsRoute() string {
	return fmt.Sprintf("/oauth/apps")
}

func (c *Client4) GetOAuthAppRoute(appId string) string {
	return fmt.Sprintf("/oauth/apps/%v", appId)
}

func (c *Client4) GetOpenGraphRoute() string {
	return fmt.Sprintf("/opengraph")
}

func (c *Client4) GetJobsRoute() string {
	return fmt.Sprintf("/jobs")
}

func (c *Client4) GetRolesRoute() string {
	return fmt.Sprintf("/roles")
}

func (c *Client4) GetSchemesRoute() string {
	return fmt.Sprintf("/schemes")
}

func (c *Client4) GetSchemeRoute(id string) string {
	return c.GetSchemesRoute() + fmt.Sprintf("/%v", id)
}

func (c *Client4) GetAnalyticsRoute() string {
	return fmt.Sprintf("/analytics")
}

func (c *Client4) GetTimezonesRoute() string {
	return fmt.Sprintf(c.GetSystemRoute() + "/timezones")
}

func (c *Client4) GetChannelSchemeRoute(channelId string) string {
	return fmt.Sprintf(c.GetChannelsRoute()+"/%v/scheme", channelId)
}

func (c *Client4) GetTeamSchemeRoute(teamId string) string {
	return fmt.Sprintf(c.GetTeamsRoute()+"/%v/scheme", teamId)
}

func (c *Client4) GetTotalUsersStatsRoute() string {
	return fmt.Sprintf(c.GetUsersRoute() + "/stats")
}

func (c *Client4) GetRedirectLocationRoute() string {
	return fmt.Sprintf("/redirect_location")
}

func (c *Client4) GetRegisterTermsOfServiceRoute(userId string) string {
	return c.GetUserRoute(userId) + "/terms_of_service"
}

func (c *Client4) GetTermsOfServiceRoute() string {
	return "/terms_of_service"
}

func (c *Client4) DoApiGet(url string, etag string) (*http.Response, *AppError) {
	return c.DoApiRequest(http.MethodGet, c.ApiUrl+url, "", etag)
}

func (c *Client4) DoApiPost(url string, data string) (*http.Response, *AppError) {
	return c.DoApiRequest(http.MethodPost, c.ApiUrl+url, data, "")
}

func (c *Client4) doApiPostBytes(url string, data []byte) (*http.Response, *AppError) {
	return c.doApiRequestBytes(http.MethodPost, c.ApiUrl+url, data, "")
}

func (c *Client4) DoApiPut(url string, data string) (*http.Response, *AppError) {
	return c.DoApiRequest(http.MethodPut, c.ApiUrl+url, data, "")
}

func (c *Client4) DoApiDelete(url string) (*http.Response, *AppError) {
	return c.DoApiRequest(http.MethodDelete, c.ApiUrl+url, "", "")
}

func (c *Client4) DoApiRequest(method, url, data, etag string) (*http.Response, *AppError) {
	return c.doApiRequestReader(method, url, strings.NewReader(data), etag)
}

func (c *Client4) doApiRequestBytes(method, url string, data []byte, etag string) (*http.Response, *AppError) {
	return c.doApiRequestReader(method, url, bytes.NewReader(data), etag)
}

func (c *Client4) doApiRequestReader(method, url string, data io.Reader, etag string) (*http.Response, *AppError) {
	rq, _ := http.NewRequest(method, url, data)

	if len(etag) > 0 {
		rq.Header.Set(HEADER_ETAG_CLIENT, etag)
	}

	if len(c.AuthToken) > 0 {
		rq.Header.Set(HEADER_AUTH, c.AuthType+" "+c.AuthToken)
	}

	if c.HttpHeader != nil && len(c.HttpHeader) > 0 {

		for k, v := range c.HttpHeader {
			rq.Header.Set(k, v)
		}
	}

	if rp, err := c.HttpClient.Do(rq); err != nil || rp == nil {
		return nil, NewAppError(url, "model.client.connecting.app_error", nil, err.Error(), 0)
	} else if rp.StatusCode == 304 {
		return rp, nil
	} else if rp.StatusCode >= 300 {
		defer closeBody(rp)
		return rp, AppErrorFromJson(rp.Body)
	} else {
		return rp, nil
	}
}

func (c *Client4) DoUploadFile(url string, data []byte, contentType string) (*FileUploadResponse, *Response) {
	rq, _ := http.NewRequest("POST", c.ApiUrl+url, bytes.NewReader(data))
	rq.Header.Set("Content-Type", contentType)

	if len(c.AuthToken) > 0 {
		rq.Header.Set(HEADER_AUTH, c.AuthType+" "+c.AuthToken)
	}

	if rp, err := c.HttpClient.Do(rq); err != nil || rp == nil {
		return nil, BuildErrorResponse(rp, NewAppError(url, "model.client.connecting.app_error", nil, err.Error(), 0))
	} else {
		defer closeBody(rp)

		if rp.StatusCode >= 300 {
			return nil, BuildErrorResponse(rp, AppErrorFromJson(rp.Body))
		} else {
			return FileUploadResponseFromJson(rp.Body), BuildResponse(rp)
		}
	}
}

func (c *Client4) DoEmojiUploadFile(url string, data []byte, contentType string) (*Emoji, *Response) {
	rq, _ := http.NewRequest("POST", c.ApiUrl+url, bytes.NewReader(data))
	rq.Header.Set("Content-Type", contentType)

	if len(c.AuthToken) > 0 {
		rq.Header.Set(HEADER_AUTH, c.AuthType+" "+c.AuthToken)
	}

	if rp, err := c.HttpClient.Do(rq); err != nil || rp == nil {
		return nil, BuildErrorResponse(rp, NewAppError(url, "model.client.connecting.app_error", nil, err.Error(), 0))
	} else {
		defer closeBody(rp)

		if rp.StatusCode >= 300 {
			return nil, BuildErrorResponse(rp, AppErrorFromJson(rp.Body))
		} else {
			return EmojiFromJson(rp.Body), BuildResponse(rp)
		}
	}
}

func (c *Client4) DoUploadImportTeam(url string, data []byte, contentType string) (map[string]string, *Response) {
	rq, _ := http.NewRequest("POST", c.ApiUrl+url, bytes.NewReader(data))
	rq.Header.Set("Content-Type", contentType)

	if len(c.AuthToken) > 0 {
		rq.Header.Set(HEADER_AUTH, c.AuthType+" "+c.AuthToken)
	}

	if rp, err := c.HttpClient.Do(rq); err != nil || rp == nil {
		return nil, BuildErrorResponse(rp, NewAppError(url, "model.client.connecting.app_error", nil, err.Error(), 0))
	} else {
		defer closeBody(rp)

		if rp.StatusCode >= 300 {
			return nil, BuildErrorResponse(rp, AppErrorFromJson(rp.Body))
		} else {
			return MapFromJson(rp.Body), BuildResponse(rp)
		}
	}
}

// CheckStatusOK is a convenience function for checking the standard OK response
// from the web service.
func CheckStatusOK(r *http.Response) bool {
	m := MapFromJson(r.Body)
	defer closeBody(r)

	if m != nil && m[STATUS] == STATUS_OK {
		return true
	}

	return false
}

// Authentication Section

// LoginById authenticates a user by user id and password.
func (c *Client4) LoginById(id string, password string) (*User, *Response) {
	m := make(map[string]string)
	m["id"] = id
	m["password"] = password
	return c.login(m)
}

// Login authenticates a user by login id, which can be username, email or some sort
// of SSO identifier based on server configuration, and a password.
func (c *Client4) Login(loginId string, password string) (*User, *Response) {
	m := make(map[string]string)
	m["login_id"] = loginId
	m["password"] = password
	return c.login(m)
}

// LoginByLdap authenticates a user by LDAP id and password.
func (c *Client4) LoginByLdap(loginId string, password string) (*User, *Response) {
	m := make(map[string]string)
	m["login_id"] = loginId
	m["password"] = password
	m["ldap_only"] = "true"
	return c.login(m)
}

// LoginWithDevice authenticates a user by login id (username, email or some sort
// of SSO identifier based on configuration), password and attaches a device id to
// the session.
func (c *Client4) LoginWithDevice(loginId string, password string, deviceId string) (*User, *Response) {
	m := make(map[string]string)
	m["login_id"] = loginId
	m["password"] = password
	m["device_id"] = deviceId
	return c.login(m)
}

func (c *Client4) login(m map[string]string) (*User, *Response) {
	if r, err := c.DoApiPost("/users/login", MapToJson(m)); err != nil {
		return nil, BuildErrorResponse(r, err)
	} else {
		c.AuthToken = r.Header.Get(HEADER_TOKEN)
		c.AuthType = HEADER_BEARER
		defer closeBody(r)
		return UserFromJson(r.Body), BuildResponse(r)
	}
}

// Logout terminates the current user's session.
func (c *Client4) Logout() (bool, *Response) {
	if r, err := c.DoApiPost("/users/logout", ""); err != nil {
		return false, BuildErrorResponse(r, err)
	} else {
		c.AuthToken = ""
		c.AuthType = HEADER_BEARER

		defer closeBody(r)
		return CheckStatusOK(r), BuildResponse(r)
	}
}

// SwitchAccountType changes a user's login type from one type to another.
func (c *Client4) SwitchAccountType(switchRequest *SwitchRequest) (string, *Response) {
	if r, err := c.DoApiPost(c.GetUsersRoute()+"/login/switch", switchRequest.ToJson()); err != nil {
		return "", BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return MapFromJson(r.Body)["follow_link"], BuildResponse(r)
	}
}

// User Section

// CreateUser creates a user in the system based on the provided user struct.
func (c *Client4) CreateUser(user *User) (*User, *Response) {
	if r, err := c.DoApiPost(c.GetUsersRoute(), user.ToJson()); err != nil {
		return nil, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return UserFromJson(r.Body), BuildResponse(r)
	}
}

// CreateUserWithToken creates a user in the system based on the provided tokenId.
func (c *Client4) CreateUserWithToken(user *User, tokenId string) (*User, *Response) {
	var query string
	if tokenId != "" {
		query = fmt.Sprintf("?t=%v", tokenId)
	} else {
		err := NewAppError("MissingHashOrData", "api.user.create_user.missing_token.app_error", nil, "", http.StatusBadRequest)
		return nil, &Response{StatusCode: err.StatusCode, Error: err}
	}
	if r, err := c.DoApiPost(c.GetUsersRoute()+query, user.ToJson()); err != nil {
		return nil, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return UserFromJson(r.Body), BuildResponse(r)
	}
}

// CreateUserWithInviteId creates a user in the system based on the provided invited id.
func (c *Client4) CreateUserWithInviteId(user *User, inviteId string) (*User, *Response) {
	var query string
	if inviteId != "" {
		query = fmt.Sprintf("?iid=%v", url.QueryEscape(inviteId))
	} else {
		err := NewAppError("MissingInviteId", "api.user.create_user.missing_invite_id.app_error", nil, "", http.StatusBadRequest)
		return nil, &Response{StatusCode: err.StatusCode, Error: err}
	}
	if r, err := c.DoApiPost(c.GetUsersRoute()+query, user.ToJson()); err != nil {
		return nil, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return UserFromJson(r.Body), BuildResponse(r)
	}
}

// GetMe returns the logged in user.
func (c *Client4) GetMe(etag string) (*User, *Response) {
	if r, err := c.DoApiGet(c.GetUserRoute(ME), etag); err != nil {
		return nil, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return UserFromJson(r.Body), BuildResponse(r)
	}
}

// GetUser returns a user based on the provided user id string.
func (c *Client4) GetUser(userId, etag string) (*User, *Response) {
	if r, err := c.DoApiGet(c.GetUserRoute(userId), etag); err != nil {
		return nil, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return UserFromJson(r.Body), BuildResponse(r)
	}
}

// GetUserByUsername returns a user based on the provided user name string.
func (c *Client4) GetUserByUsername(userName, etag string) (*User, *Response) {
	if r, err := c.DoApiGet(c.GetUserByUsernameRoute(userName), etag); err != nil {
		return nil, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return UserFromJson(r.Body), BuildResponse(r)
	}
}

// GetUserByEmail returns a user based on the provided user email string.
func (c *Client4) GetUserByEmail(email, etag string) (*User, *Response) {
	if r, err := c.DoApiGet(c.GetUserByEmailRoute(email), etag); err != nil {
		return nil, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return UserFromJson(r.Body), BuildResponse(r)
	}
}

// AutocompleteUsersInTeam returns the users on a team based on search term.
func (c *Client4) AutocompleteUsersInTeam(teamId string, username string, limit int, etag string) (*UserAutocomplete, *Response) {
	query := fmt.Sprintf("?in_team=%v&name=%v&limit=%d", teamId, username, limit)
	if r, err := c.DoApiGet(c.GetUsersRoute()+"/autocomplete"+query, etag); err != nil {
		return nil, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return UserAutocompleteFromJson(r.Body), BuildResponse(r)
	}
}

// AutocompleteUsersInChannel returns the users in a channel based on search term.
func (c *Client4) AutocompleteUsersInChannel(teamId string, channelId string, username string, limit int, etag string) (*UserAutocomplete, *Response) {
	query := fmt.Sprintf("?in_team=%v&in_channel=%v&name=%v&limit=%d", teamId, channelId, username, limit)
	if r, err := c.DoApiGet(c.GetUsersRoute()+"/autocomplete"+query, etag); err != nil {
		return nil, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return UserAutocompleteFromJson(r.Body), BuildResponse(r)
	}
}

// AutocompleteUsers returns the users in the system based on search term.
func (c *Client4) AutocompleteUsers(username string, limit int, etag string) (*UserAutocomplete, *Response) {
	query := fmt.Sprintf("?name=%v&limit=%d", username, limit)
	if r, err := c.DoApiGet(c.GetUsersRoute()+"/autocomplete"+query, etag); err != nil {
		return nil, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return UserAutocompleteFromJson(r.Body), BuildResponse(r)
	}
}

// GetDefaultProfileImage gets the default user's profile image. Must be logged in.
func (c *Client4) GetDefaultProfileImage(userId string) ([]byte, *Response) {
	r, appErr := c.DoApiGet(c.GetUserRoute(userId)+"/image/default", "")
	if appErr != nil {
		return nil, BuildErrorResponse(r, appErr)
	}
	defer closeBody(r)

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, BuildErrorResponse(r, NewAppError("GetDefaultProfileImage", "model.client.read_file.app_error", nil, err.Error(), r.StatusCode))
	}

	return data, BuildResponse(r)
}

// GetProfileImage gets user's profile image. Must be logged in.
func (c *Client4) GetProfileImage(userId, etag string) ([]byte, *Response) {
	if r, err := c.DoApiGet(c.GetUserRoute(userId)+"/image", etag); err != nil {
		return nil, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)

		if data, err := ioutil.ReadAll(r.Body); err != nil {
			return nil, BuildErrorResponse(r, NewAppError("GetProfileImage", "model.client.read_file.app_error", nil, err.Error(), r.StatusCode))
		} else {
			return data, BuildResponse(r)
		}
	}
}

// GetUsers returns a page of users on the system. Page counting starts at 0.
func (c *Client4) GetUsers(page int, perPage int, etag string) ([]*User, *Response) {
	query := fmt.Sprintf("?page=%v&per_page=%v", page, perPage)
	if r, err := c.DoApiGet(c.GetUsersRoute()+query, etag); err != nil {
		return nil, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return UserListFromJson(r.Body), BuildResponse(r)
	}
}

// GetUsersInTeam returns a page of users on a team. Page counting starts at 0.
func (c *Client4) GetUsersInTeam(teamId string, page int, perPage int, etag string) ([]*User, *Response) {
	query := fmt.Sprintf("?in_team=%v&page=%v&per_page=%v", teamId, page, perPage)
	if r, err := c.DoApiGet(c.GetUsersRoute()+query, etag); err != nil {
		return nil, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return UserListFromJson(r.Body), BuildResponse(r)
	}
}

// GetNewUsersInTeam returns a page of users on a team. Page counting starts at 0.
func (c *Client4) GetNewUsersInTeam(teamId string, page int, perPage int, etag string) ([]*User, *Response) {
	query := fmt.Sprintf("?sort=create_at&in_team=%v&page=%v&per_page=%v", teamId, page, perPage)
	if r, err := c.DoApiGet(c.GetUsersRoute()+query, etag); err != nil {
		return nil, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return UserListFromJson(r.Body), BuildResponse(r)
	}
}

// GetRecentlyActiveUsersInTeam returns a page of users on a team. Page counting starts at 0.
func (c *Client4) GetRecentlyActiveUsersInTeam(teamId string, page int, perPage int, etag string) ([]*User, *Response) {
	query := fmt.Sprintf("?sort=last_activity_at&in_team=%v&page=%v&per_page=%v", teamId, page, perPage)
	if r, err := c.DoApiGet(c.GetUsersRoute()+query, etag); err != nil {
		return nil, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return UserListFromJson(r.Body), BuildResponse(r)
	}
}

// GetUsersNotInTeam returns a page of users who are not in a team. Page counting starts at 0.
func (c *Client4) GetUsersNotInTeam(teamId string, page int, perPage int, etag string) ([]*User, *Response) {
	query := fmt.Sprintf("?not_in_team=%v&page=%v&per_page=%v", teamId, page, perPage)
	if r, err := c.DoApiGet(c.GetUsersRoute()+query, etag); err != nil {
		return nil, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return UserListFromJson(r.Body), BuildResponse(r)
	}
}

// GetUsersInChannel returns a page of users in a channel. Page counting starts at 0.
func (c *Client4) GetUsersInChannel(channelId string, page int, perPage int, etag string) ([]*User, *Response) {
	query := fmt.Sprintf("?in_channel=%v&page=%v&per_page=%v", channelId, page, perPage)
	if r, err := c.DoApiGet(c.GetUsersRoute()+query, etag); err != nil {
		return nil, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return UserListFromJson(r.Body), BuildResponse(r)
	}
}

// GetUsersInChannelByStatus returns a page of users in a channel. Page counting starts at 0. Sorted by Status
func (c *Client4) GetUsersInChannelByStatus(channelId string, page int, perPage int, etag string) ([]*User, *Response) {
	query := fmt.Sprintf("?in_channel=%v&page=%v&per_page=%v&sort=status", channelId, page, perPage)
	if r, err := c.DoApiGet(c.GetUsersRoute()+query, etag); err != nil {
		return nil, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return UserListFromJson(r.Body), BuildResponse(r)
	}
}

// GetUsersNotInChannel returns a page of users not in a channel. Page counting starts at 0.
func (c *Client4) GetUsersNotInChannel(teamId, channelId string, page int, perPage int, etag string) ([]*User, *Response) {
	query := fmt.Sprintf("?in_team=%v&not_in_channel=%v&page=%v&per_page=%v", teamId, channelId, page, perPage)
	if r, err := c.DoApiGet(c.GetUsersRoute()+query, etag); err != nil {
		return nil, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return UserListFromJson(r.Body), BuildResponse(r)
	}
}

// GetUsersWithoutTeam returns a page of users on the system that aren't on any teams. Page counting starts at 0.
func (c *Client4) GetUsersWithoutTeam(page int, perPage int, etag string) ([]*User, *Response) {
	query := fmt.Sprintf("?without_team=1&page=%v&per_page=%v", page, perPage)
	if r, err := c.DoApiGet(c.GetUsersRoute()+query, etag); err != nil {
		return nil, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return UserListFromJson(r.Body), BuildResponse(r)
	}
}

// GetUsersByIds returns a list of users based on the provided user ids.
func (c *Client4) GetUsersByIds(userIds []string) ([]*User, *Response) {
	if r, err := c.DoApiPost(c.GetUsersRoute()+"/ids", ArrayToJson(userIds)); err != nil {
		return nil, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return UserListFromJson(r.Body), BuildResponse(r)
	}
}

// GetUsersByUsernames returns a list of users based on the provided usernames.
func (c *Client4) GetUsersByUsernames(usernames []string) ([]*User, *Response) {
	if r, err := c.DoApiPost(c.GetUsersRoute()+"/usernames", ArrayToJson(usernames)); err != nil {
		return nil, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return UserListFromJson(r.Body), BuildResponse(r)
	}
}

// SearchUsers returns a list of users based on some search criteria.
func (c *Client4) SearchUsers(search *UserSearch) ([]*User, *Response) {
	if r, err := c.doApiPostBytes(c.GetUsersRoute()+"/search", search.ToJson()); err != nil {
		return nil, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return UserListFromJson(r.Body), BuildResponse(r)
	}
}

// UpdateUser updates a user in the system based on the provided user struct.
func (c *Client4) UpdateUser(user *User) (*User, *Response) {
	if r, err := c.DoApiPut(c.GetUserRoute(user.Id), user.ToJson()); err != nil {
		return nil, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return UserFromJson(r.Body), BuildResponse(r)
	}
}

// PatchUser partially updates a user in the system. Any missing fields are not updated.
func (c *Client4) PatchUser(userId string, patch *UserPatch) (*User, *Response) {
	if r, err := c.DoApiPut(c.GetUserRoute(userId)+"/patch", patch.ToJson()); err != nil {
		return nil, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return UserFromJson(r.Body), BuildResponse(r)
	}
}

// UpdateUserAuth updates a user AuthData (uthData, authService and password) in the system.
func (c *Client4) UpdateUserAuth(userId string, userAuth *UserAuth) (*UserAuth, *Response) {
	if r, err := c.DoApiPut(c.GetUserRoute(userId)+"/auth", userAuth.ToJson()); err != nil {
		return nil, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return UserAuthFromJson(r.Body), BuildResponse(r)
	}
}

// UpdateUserMfa activates multi-factor authentication for a user if activate
// is true and a valid code is provided. If activate is false, then code is not
// required and multi-factor authentication is disabled for the user.
func (c *Client4) UpdateUserMfa(userId, code string, activate bool) (bool, *Response) {
	requestBody := make(map[string]interface{})
	requestBody["activate"] = activate
	requestBody["code"] = code

	if r, err := c.DoApiPut(c.GetUserRoute(userId)+"/mfa", StringInterfaceToJson(requestBody)); err != nil {
		return false, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return CheckStatusOK(r), BuildResponse(r)
	}
}

// CheckUserMfa checks whether a user has MFA active on their account or not based on the
// provided login id.
func (c *Client4) CheckUserMfa(loginId string) (bool, *Response) {
	requestBody := make(map[string]interface{})
	requestBody["login_id"] = loginId

	if r, err := c.DoApiPost(c.GetUsersRoute()+"/mfa", StringInterfaceToJson(requestBody)); err != nil {
		return false, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		data := StringInterfaceFromJson(r.Body)
		if mfaRequired, ok := data["mfa_required"].(bool); !ok {
			return false, BuildResponse(r)
		} else {
			return mfaRequired, BuildResponse(r)
		}
	}
}

// GenerateMfaSecret will generate a new MFA secret for a user and return it as a string and
// as a base64 encoded image QR code.
func (c *Client4) GenerateMfaSecret(userId string) (*MfaSecret, *Response) {
	if r, err := c.DoApiPost(c.GetUserRoute(userId)+"/mfa/generate", ""); err != nil {
		return nil, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return MfaSecretFromJson(r.Body), BuildResponse(r)
	}
}

// UpdateUserPassword updates a user's password. Must be logged in as the user or be a system administrator.
func (c *Client4) UpdateUserPassword(userId, currentPassword, newPassword string) (bool, *Response) {
	requestBody := map[string]string{"current_password": currentPassword, "new_password": newPassword}
	if r, err := c.DoApiPut(c.GetUserRoute(userId)+"/password", MapToJson(requestBody)); err != nil {
		return false, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return CheckStatusOK(r), BuildResponse(r)
	}
}

// UpdateUserRoles updates a user's roles in the system. A user can have "system_user" and "system_admin" roles.
func (c *Client4) UpdateUserRoles(userId, roles string) (bool, *Response) {
	requestBody := map[string]string{"roles": roles}
	if r, err := c.DoApiPut(c.GetUserRoute(userId)+"/roles", MapToJson(requestBody)); err != nil {
		return false, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return CheckStatusOK(r), BuildResponse(r)
	}
}

// UpdateUserActive updates status of a user whether active or not.
func (c *Client4) UpdateUserActive(userId string, active bool) (bool, *Response) {
	requestBody := make(map[string]interface{})
	requestBody["active"] = active

	if r, err := c.DoApiPut(c.GetUserRoute(userId)+"/active", StringInterfaceToJson(requestBody)); err != nil {
		return false, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return CheckStatusOK(r), BuildResponse(r)
	}
}

// DeleteUser deactivates a user in the system based on the provided user id string.
func (c *Client4) DeleteUser(userId string) (bool, *Response) {
	if r, err := c.DoApiDelete(c.GetUserRoute(userId)); err != nil {
		return false, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return CheckStatusOK(r), BuildResponse(r)
	}
}

// SendPasswordResetEmail will send a link for password resetting to a user with the
// provided email.
func (c *Client4) SendPasswordResetEmail(email string) (bool, *Response) {
	requestBody := map[string]string{"email": email}
	if r, err := c.DoApiPost(c.GetUsersRoute()+"/password/reset/send", MapToJson(requestBody)); err != nil {
		return false, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return CheckStatusOK(r), BuildResponse(r)
	}
}

// ResetPassword uses a recovery code to update reset a user's password.
func (c *Client4) ResetPassword(token, newPassword string) (bool, *Response) {
	requestBody := map[string]string{"token": token, "new_password": newPassword}
	if r, err := c.DoApiPost(c.GetUsersRoute()+"/password/reset", MapToJson(requestBody)); err != nil {
		return false, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return CheckStatusOK(r), BuildResponse(r)
	}
}

// GetSessions returns a list of sessions based on the provided user id string.
func (c *Client4) GetSessions(userId, etag string) ([]*Session, *Response) {
	if r, err := c.DoApiGet(c.GetUserRoute(userId)+"/sessions", etag); err != nil {
		return nil, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return SessionsFromJson(r.Body), BuildResponse(r)
	}
}

// RevokeSession revokes a user session based on the provided user id and session id strings.
func (c *Client4) RevokeSession(userId, sessionId string) (bool, *Response) {
	requestBody := map[string]string{"session_id": sessionId}
	if r, err := c.DoApiPost(c.GetUserRoute(userId)+"/sessions/revoke", MapToJson(requestBody)); err != nil {
		return false, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return CheckStatusOK(r), BuildResponse(r)
	}
}

// RevokeAllSessions revokes all sessions for the provided user id string.
func (c *Client4) RevokeAllSessions(userId string) (bool, *Response) {
	if r, err := c.DoApiPost(c.GetUserRoute(userId)+"/sessions/revoke/all", ""); err != nil {
		return false, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return CheckStatusOK(r), BuildResponse(r)
	}
}

// AttachDeviceId attaches a mobile device ID to the current session.
func (c *Client4) AttachDeviceId(deviceId string) (bool, *Response) {
	requestBody := map[string]string{"device_id": deviceId}
	if r, err := c.DoApiPut(c.GetUsersRoute()+"/sessions/device", MapToJson(requestBody)); err != nil {
		return false, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return CheckStatusOK(r), BuildResponse(r)
	}
}

// VerifyUserEmail will verify a user's email using the supplied token.
func (c *Client4) VerifyUserEmail(token string) (bool, *Response) {
	requestBody := map[string]string{"token": token}
	if r, err := c.DoApiPost(c.GetUsersRoute()+"/email/verify", MapToJson(requestBody)); err != nil {
		return false, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return CheckStatusOK(r), BuildResponse(r)
	}
}

// SendVerificationEmail will send an email to the user with the provided email address, if
// that user exists. The email will contain a link that can be used to verify the user's
// email address.
func (c *Client4) SendVerificationEmail(email string) (bool, *Response) {
	requestBody := map[string]string{"email": email}
	if r, err := c.DoApiPost(c.GetUsersRoute()+"/email/verify/send", MapToJson(requestBody)); err != nil {
		return false, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return CheckStatusOK(r), BuildResponse(r)
	}
}

// SetDefaultProfileImage resets the profile image to a default generated one
func (c *Client4) SetDefaultProfileImage(userId string) (bool, *Response) {
	r, err := c.DoApiDelete(c.GetUserRoute(userId) + "/image")
	if err != nil {
		return false, BuildErrorResponse(r, err)
	}
	return CheckStatusOK(r), BuildResponse(r)
}

// SetProfileImage sets profile image of the user
func (c *Client4) SetProfileImage(userId string, data []byte) (bool, *Response) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if part, err := writer.CreateFormFile("image", "profile.png"); err != nil {
		return false, &Response{Error: NewAppError("SetProfileImage", "model.client.set_profile_user.no_file.app_error", nil, err.Error(), http.StatusBadRequest)}
	} else if _, err = io.Copy(part, bytes.NewBuffer(data)); err != nil {
		return false, &Response{Error: NewAppError("SetProfileImage", "model.client.set_profile_user.no_file.app_error", nil, err.Error(), http.StatusBadRequest)}
	}

	if err := writer.Close(); err != nil {
		return false, &Response{Error: NewAppError("SetProfileImage", "model.client.set_profile_user.writer.app_error", nil, err.Error(), http.StatusBadRequest)}
	}

	rq, _ := http.NewRequest("POST", c.ApiUrl+c.GetUserRoute(userId)+"/image", bytes.NewReader(body.Bytes()))
	rq.Header.Set("Content-Type", writer.FormDataContentType())

	if len(c.AuthToken) > 0 {
		rq.Header.Set(HEADER_AUTH, c.AuthType+" "+c.AuthToken)
	}

	if rp, err := c.HttpClient.Do(rq); err != nil || rp == nil {
		// set to http.StatusForbidden(403)
		return false, &Response{StatusCode: http.StatusForbidden, Error: NewAppError(c.GetUserRoute(userId)+"/image", "model.client.connecting.app_error", nil, err.Error(), 403)}
	} else {
		defer closeBody(rp)

		if rp.StatusCode >= 300 {
			return false, BuildErrorResponse(rp, AppErrorFromJson(rp.Body))
		} else {
			return CheckStatusOK(rp), BuildResponse(rp)
		}
	}
}

// CreateUserAccessToken will generate a user access token that can be used in place
// of a session token to access the REST API. Must have the 'create_user_access_token'
// permission and if generating for another user, must have the 'edit_other_users'
// permission. A non-blank description is required.
func (c *Client4) CreateUserAccessToken(userId, description string) (*UserAccessToken, *Response) {
	requestBody := map[string]string{"description": description}
	if r, err := c.DoApiPost(c.GetUserRoute(userId)+"/tokens", MapToJson(requestBody)); err != nil {
		return nil, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return UserAccessTokenFromJson(r.Body), BuildResponse(r)
	}
}

// GetUserAccessTokens will get a page of access tokens' id, description, is_active
// and the user_id in the system. The actual token will not be returned. Must have
// the 'manage_system' permission.
func (c *Client4) GetUserAccessTokens(page int, perPage int) ([]*UserAccessToken, *Response) {
	query := fmt.Sprintf("?page=%v&per_page=%v", page, perPage)
	if r, err := c.DoApiGet(c.GetUserAccessTokensRoute()+query, ""); err != nil {
		return nil, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return UserAccessTokenListFromJson(r.Body), BuildResponse(r)
	}
}

// GetUserAccessToken will get a user access tokens' id, description, is_active
// and the user_id of the user it is for. The actual token will not be returned.
// Must have the 'read_user_access_token' permission and if getting for another
// user, must have the 'edit_other_users' permission.
func (c *Client4) GetUserAccessToken(tokenId string) (*UserAccessToken, *Response) {
	if r, err := c.DoApiGet(c.GetUserAccessTokenRoute(tokenId), ""); err != nil {
		return nil, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return UserAccessTokenFromJson(r.Body), BuildResponse(r)
	}
}

// GetUserAccessTokensForUser will get a paged list of user access tokens showing id,
// description and user_id for each. The actual tokens will not be returned. Must have
// the 'read_user_access_token' permission and if getting for another user, must have the
// 'edit_other_users' permission.
func (c *Client4) GetUserAccessTokensForUser(userId string, page, perPage int) ([]*UserAccessToken, *Response) {
	query := fmt.Sprintf("?page=%v&per_page=%v", page, perPage)
	if r, err := c.DoApiGet(c.GetUserRoute(userId)+"/tokens"+query, ""); err != nil {
		return nil, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return UserAccessTokenListFromJson(r.Body), BuildResponse(r)
	}
}

// RevokeUserAccessToken will revoke a user access token by id. Must have the
// 'revoke_user_access_token' permission and if revoking for another user, must have the
// 'edit_other_users' permission.
func (c *Client4) RevokeUserAccessToken(tokenId string) (bool, *Response) {
	requestBody := map[string]string{"token_id": tokenId}
	if r, err := c.DoApiPost(c.GetUsersRoute()+"/tokens/revoke", MapToJson(requestBody)); err != nil {
		return false, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return CheckStatusOK(r), BuildResponse(r)
	}
}

// SearchUserAccessTokens returns user access tokens matching the provided search term.
func (c *Client4) SearchUserAccessTokens(search *UserAccessTokenSearch) ([]*UserAccessToken, *Response) {
	if r, err := c.DoApiPost(c.GetUsersRoute()+"/tokens/search", search.ToJson()); err != nil {
		return nil, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return UserAccessTokenListFromJson(r.Body), BuildResponse(r)
	}
}

// DisableUserAccessToken will disable a user access token by id. Must have the
// 'revoke_user_access_token' permission and if disabling for another user, must have the
// 'edit_other_users' permission.
func (c *Client4) DisableUserAccessToken(tokenId string) (bool, *Response) {
	requestBody := map[string]string{"token_id": tokenId}
	if r, err := c.DoApiPost(c.GetUsersRoute()+"/tokens/disable", MapToJson(requestBody)); err != nil {
		return false, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return CheckStatusOK(r), BuildResponse(r)
	}
}

// EnableUserAccessToken will enable a user access token by id. Must have the
// 'create_user_access_token' permission and if enabling for another user, must have the
// 'edit_other_users' permission.
func (c *Client4) EnableUserAccessToken(tokenId string) (bool, *Response) {
	requestBody := map[string]string{"token_id": tokenId}
	if r, err := c.DoApiPost(c.GetUsersRoute()+"/tokens/enable", MapToJson(requestBody)); err != nil {
		return false, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return CheckStatusOK(r), BuildResponse(r)
	}
}

// File Section

// UploadFile will upload a file to a channel using a multipart request, to be later attached to a post.
// This method is functionally equivalent to Client4.UploadFileAsRequestBody.
func (c *Client4) UploadFile(data []byte, channelId string, filename string) (*FileUploadResponse, *Response) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if part, err := writer.CreateFormFile("files", filename); err != nil {
		return nil, &Response{Error: NewAppError("UploadPostAttachment", "model.client.upload_post_attachment.file.app_error", nil, err.Error(), http.StatusBadRequest)}
	} else if _, err = io.Copy(part, bytes.NewBuffer(data)); err != nil {
		return nil, &Response{Error: NewAppError("UploadPostAttachment", "model.client.upload_post_attachment.file.app_error", nil, err.Error(), http.StatusBadRequest)}
	}

	if part, err := writer.CreateFormField("channel_id"); err != nil {
		return nil, &Response{Error: NewAppError("UploadPostAttachment", "model.client.upload_post_attachment.channel_id.app_error", nil, err.Error(), http.StatusBadRequest)}
	} else if _, err = io.Copy(part, strings.NewReader(channelId)); err != nil {
		return nil, &Response{Error: NewAppError("UploadPostAttachment", "model.client.upload_post_attachment.channel_id.app_error", nil, err.Error(), http.StatusBadRequest)}
	}

	if err := writer.Close(); err != nil {
		return nil, &Response{Error: NewAppError("UploadPostAttachment", "model.client.upload_post_attachment.writer.app_error", nil, err.Error(), http.StatusBadRequest)}
	}

	return c.DoUploadFile(c.GetFilesRoute(), body.Bytes(), writer.FormDataContentType())
}

// UploadFileAsRequestBody will upload a file to a channel as the body of a request, to be later attached
// to a post. This method is functionally equivalent to Client4.UploadFile.
func (c *Client4) UploadFileAsRequestBody(data []byte, channelId string, filename string) (*FileUploadResponse, *Response) {
	return c.DoUploadFile(c.GetFilesRoute()+fmt.Sprintf("?channel_id=%v&filename=%v", url.QueryEscape(channelId), url.QueryEscape(filename)), data, http.DetectContentType(data))
}

// GetFile gets the bytes for a file by id.
func (c *Client4) GetFile(fileId string) ([]byte, *Response) {
	if r, err := c.DoApiGet(c.GetFileRoute(fileId), ""); err != nil {
		return nil, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)

		if data, err := ioutil.ReadAll(r.Body); err != nil {
			return nil, BuildErrorResponse(r, NewAppError("GetFile", "model.client.read_file.app_error", nil, err.Error(), r.StatusCode))
		} else {
			return data, BuildResponse(r)
		}
	}
}

// DownloadFile gets the bytes for a file by id, optionally adding headers to force the browser to download it
func (c *Client4) DownloadFile(fileId string, download bool) ([]byte, *Response) {
	if r, err := c.DoApiGet(c.GetFileRoute(fileId)+fmt.Sprintf("?download=%v", download), ""); err != nil {
		return nil, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)

		if data, err := ioutil.ReadAll(r.Body); err != nil {
			return nil, BuildErrorResponse(r, NewAppError("DownloadFile", "model.client.read_file.app_error", nil, err.Error(), r.StatusCode))
		} else {
			return data, BuildResponse(r)
		}
	}
}

// GetFileThumbnail gets the bytes for a file by id.
func (c *Client4) GetFileThumbnail(fileId string) ([]byte, *Response) {
	if r, err := c.DoApiGet(c.GetFileRoute(fileId)+"/thumbnail", ""); err != nil {
		return nil, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)

		if data, err := ioutil.ReadAll(r.Body); err != nil {
			return nil, BuildErrorResponse(r, NewAppError("GetFileThumbnail", "model.client.read_file.app_error", nil, err.Error(), r.StatusCode))
		} else {
			return data, BuildResponse(r)
		}
	}
}

// DownloadFileThumbnail gets the bytes for a file by id, optionally adding headers to force the browser to download it.
func (c *Client4) DownloadFileThumbnail(fileId string, download bool) ([]byte, *Response) {
	if r, err := c.DoApiGet(c.GetFileRoute(fileId)+fmt.Sprintf("/thumbnail?download=%v", download), ""); err != nil {
		return nil, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)

		if data, err := ioutil.ReadAll(r.Body); err != nil {
			return nil, BuildErrorResponse(r, NewAppError("DownloadFileThumbnail", "model.client.read_file.app_error", nil, err.Error(), r.StatusCode))
		} else {
			return data, BuildResponse(r)
		}
	}
}

// GetFileLink gets the public link of a file by id.
func (c *Client4) GetFileLink(fileId string) (string, *Response) {
	if r, err := c.DoApiGet(c.GetFileRoute(fileId)+"/link", ""); err != nil {
		return "", BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)

		return MapFromJson(r.Body)["link"], BuildResponse(r)
	}
}

// GetFilePreview gets the bytes for a file by id.
func (c *Client4) GetFilePreview(fileId string) ([]byte, *Response) {
	if r, err := c.DoApiGet(c.GetFileRoute(fileId)+"/preview", ""); err != nil {
		return nil, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)

		if data, err := ioutil.ReadAll(r.Body); err != nil {
			return nil, BuildErrorResponse(r, NewAppError("GetFilePreview", "model.client.read_file.app_error", nil, err.Error(), r.StatusCode))
		} else {
			return data, BuildResponse(r)
		}
	}
}

// DownloadFilePreview gets the bytes for a file by id.
func (c *Client4) DownloadFilePreview(fileId string, download bool) ([]byte, *Response) {
	if r, err := c.DoApiGet(c.GetFileRoute(fileId)+fmt.Sprintf("/preview?download=%v", download), ""); err != nil {
		return nil, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)

		if data, err := ioutil.ReadAll(r.Body); err != nil {
			return nil, BuildErrorResponse(r, NewAppError("DownloadFilePreview", "model.client.read_file.app_error", nil, err.Error(), r.StatusCode))
		} else {
			return data, BuildResponse(r)
		}
	}
}

// GetFileInfo gets all the file info objects.
func (c *Client4) GetFileInfo(fileId string) (*FileInfo, *Response) {
	if r, err := c.DoApiGet(c.GetFileRoute(fileId)+"/info", ""); err != nil {
		return nil, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return FileInfoFromJson(r.Body), BuildResponse(r)
	}
}

// GetFileInfosForPost gets all the file info objects attached to a post.
func (c *Client4) GetFileInfosForPost(postId string, etag string) ([]*FileInfo, *Response) {
	if r, err := c.DoApiGet(c.GetPostRoute(postId)+"/files/info", etag); err != nil {
		return nil, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return FileInfosFromJson(r.Body), BuildResponse(r)
	}
}

// General/System Section

// GetPing will return ok if the running goRoutines are below the threshold and unhealthy for above.
func (c *Client4) GetPing() (string, *Response) {
	if r, err := c.DoApiGet(c.GetSystemRoute()+"/ping", ""); r != nil && r.StatusCode == 500 {
		defer r.Body.Close()
		return "unhealthy", BuildErrorResponse(r, err)
	} else if err != nil {
		return "", BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return MapFromJson(r.Body)["status"], BuildResponse(r)
	}
}

// TestEmail will attempt to connect to the configured SMTP server.
func (c *Client4) TestEmail(config *Config) (bool, *Response) {
	if r, err := c.DoApiPost(c.GetTestEmailRoute(), config.ToJson()); err != nil {
		return false, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return CheckStatusOK(r), BuildResponse(r)
	}
}

// TestS3Connection will attempt to connect to the AWS S3.
func (c *Client4) TestS3Connection(config *Config) (bool, *Response) {
	if r, err := c.DoApiPost(c.GetTestS3Route(), config.ToJson()); err != nil {
		return false, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return CheckStatusOK(r), BuildResponse(r)
	}
}

// GetConfig will retrieve the server config with some sanitized items.
func (c *Client4) GetConfig() (*Config, *Response) {
	if r, err := c.DoApiGet(c.GetConfigRoute(), ""); err != nil {
		return nil, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return ConfigFromJson(r.Body), BuildResponse(r)
	}
}

// ReloadConfig will reload the server configuration.
func (c *Client4) ReloadConfig() (bool, *Response) {
	if r, err := c.DoApiPost(c.GetConfigRoute()+"/reload", ""); err != nil {
		return false, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return CheckStatusOK(r), BuildResponse(r)
	}
}

// GetOldClientConfig will retrieve the parts of the server configuration needed by the
// client, formatted in the old format.
func (c *Client4) GetOldClientConfig(etag string) (map[string]string, *Response) {
	if r, err := c.DoApiGet(c.GetConfigRoute()+"/client?format=old", etag); err != nil {
		return nil, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return MapFromJson(r.Body), BuildResponse(r)
	}
}

// GetEnvironmentConfig will retrieve a map mirroring the server configuration where fields
// are set to true if the corresponding config setting is set through an environment variable.
// Settings that haven't been set through environment variables will be missing from the map.
func (c *Client4) GetEnvironmentConfig() (map[string]interface{}, *Response) {
	if r, err := c.DoApiGet(c.GetConfigRoute()+"/environment", ""); err != nil {
		return nil, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return StringInterfaceFromJson(r.Body), BuildResponse(r)
	}
}

// GetOldClientLicense will retrieve the parts of the server license needed by the
// client, formatted in the old format.
func (c *Client4) GetOldClientLicense(etag string) (map[string]string, *Response) {
	if r, err := c.DoApiGet(c.GetLicenseRoute()+"/client?format=old", etag); err != nil {
		return nil, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return MapFromJson(r.Body), BuildResponse(r)
	}
}

// DatabaseRecycle will recycle the connections. Discard current connection and get new one.
func (c *Client4) DatabaseRecycle() (bool, *Response) {
	if r, err := c.DoApiPost(c.GetDatabaseRoute()+"/recycle", ""); err != nil {
		return false, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return CheckStatusOK(r), BuildResponse(r)
	}
}

// InvalidateCaches will purge the cache and can affect the performance while is cleaning.
func (c *Client4) InvalidateCaches() (bool, *Response) {
	if r, err := c.DoApiPost(c.GetCacheRoute()+"/invalidate", ""); err != nil {
		return false, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return CheckStatusOK(r), BuildResponse(r)
	}
}

// UpdateConfig will update the server configuration.
func (c *Client4) UpdateConfig(config *Config) (*Config, *Response) {
	if r, err := c.DoApiPut(c.GetConfigRoute(), config.ToJson()); err != nil {
		return nil, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return ConfigFromJson(r.Body), BuildResponse(r)
	}
}

// UploadLicenseFile will add a license file to the system.
func (c *Client4) UploadLicenseFile(data []byte) (bool, *Response) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if part, err := writer.CreateFormFile("license", "test-license.mattermost-license"); err != nil {
		return false, &Response{Error: NewAppError("UploadLicenseFile", "model.client.set_profile_user.no_file.app_error", nil, err.Error(), http.StatusBadRequest)}
	} else if _, err = io.Copy(part, bytes.NewBuffer(data)); err != nil {
		return false, &Response{Error: NewAppError("UploadLicenseFile", "model.client.set_profile_user.no_file.app_error", nil, err.Error(), http.StatusBadRequest)}
	}

	if err := writer.Close(); err != nil {
		return false, &Response{Error: NewAppError("UploadLicenseFile", "model.client.set_profile_user.writer.app_error", nil, err.Error(), http.StatusBadRequest)}
	}

	rq, _ := http.NewRequest("POST", c.ApiUrl+c.GetLicenseRoute(), bytes.NewReader(body.Bytes()))
	rq.Header.Set("Content-Type", writer.FormDataContentType())

	if len(c.AuthToken) > 0 {
		rq.Header.Set(HEADER_AUTH, c.AuthType+" "+c.AuthToken)
	}

	if rp, err := c.HttpClient.Do(rq); err != nil || rp == nil {
		return false, &Response{StatusCode: http.StatusForbidden, Error: NewAppError(c.GetLicenseRoute(), "model.client.connecting.app_error", nil, err.Error(), http.StatusForbidden)}
	} else {
		defer closeBody(rp)

		if rp.StatusCode >= 300 {
			return false, BuildErrorResponse(rp, AppErrorFromJson(rp.Body))
		} else {
			return CheckStatusOK(rp), BuildResponse(rp)
		}
	}
}

// RemoveLicenseFile will remove the server license it exists. Note that this will
// disable all enterprise features.
func (c *Client4) RemoveLicenseFile() (bool, *Response) {
	if r, err := c.DoApiDelete(c.GetLicenseRoute()); err != nil {
		return false, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return CheckStatusOK(r), BuildResponse(r)
	}
}

// SAML Section

// GetSamlMetadata returns metadata for the SAML configuration.
func (c *Client4) GetSamlMetadata() (string, *Response) {
	if r, err := c.DoApiGet(c.GetSamlRoute()+"/metadata", ""); err != nil {
		return "", BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		return buf.String(), BuildResponse(r)
	}
}

func samlFileToMultipart(data []byte, filename string) ([]byte, *multipart.Writer, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if part, err := writer.CreateFormFile("certificate", filename); err != nil {
		return nil, nil, err
	} else if _, err = io.Copy(part, bytes.NewBuffer(data)); err != nil {
		return nil, nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, nil, err
	}

	return body.Bytes(), writer, nil
}

// UploadSamlIdpCertificate will upload an IDP certificate for SAML and set the config to use it.
func (c *Client4) UploadSamlIdpCertificate(data []byte, filename string) (bool, *Response) {
	body, writer, err := samlFileToMultipart(data, filename)
	if err != nil {
		return false, &Response{Error: NewAppError("UploadSamlIdpCertificate", "model.client.upload_saml_cert.app_error", nil, err.Error(), http.StatusBadRequest)}
	}

	_, resp := c.DoUploadFile(c.GetSamlRoute()+"/certificate/idp", body, writer.FormDataContentType())
	return resp.Error == nil, resp
}

// UploadSamlPublicCertificate will upload a public certificate for SAML and set the config to use it.
func (c *Client4) UploadSamlPublicCertificate(data []byte, filename string) (bool, *Response) {
	body, writer, err := samlFileToMultipart(data, filename)
	if err != nil {
		return false, &Response{Error: NewAppError("UploadSamlPublicCertificate", "model.client.upload_saml_cert.app_error", nil, err.Error(), http.StatusBadRequest)}
	}

	_, resp := c.DoUploadFile(c.GetSamlRoute()+"/certificate/public", body, writer.FormDataContentType())
	return resp.Error == nil, resp
}

// UploadSamlPrivateCertificate will upload a private key for SAML and set the config to use it.
func (c *Client4) UploadSamlPrivateCertificate(data []byte, filename string) (bool, *Response) {
	body, writer, err := samlFileToMultipart(data, filename)
	if err != nil {
		return false, &Response{Error: NewAppError("UploadSamlPrivateCertificate", "model.client.upload_saml_cert.app_error", nil, err.Error(), http.StatusBadRequest)}
	}

	_, resp := c.DoUploadFile(c.GetSamlRoute()+"/certificate/private", body, writer.FormDataContentType())
	return resp.Error == nil, resp
}

// DeleteSamlIdpCertificate deletes the SAML IDP certificate from the server and updates the config to not use it and disable SAML.
func (c *Client4) DeleteSamlIdpCertificate() (bool, *Response) {
	if r, err := c.DoApiDelete(c.GetSamlRoute() + "/certificate/idp"); err != nil {
		return false, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return CheckStatusOK(r), BuildResponse(r)
	}
}

// DeleteSamlPublicCertificate deletes the SAML IDP certificate from the server and updates the config to not use it and disable SAML.
func (c *Client4) DeleteSamlPublicCertificate() (bool, *Response) {
	if r, err := c.DoApiDelete(c.GetSamlRoute() + "/certificate/public"); err != nil {
		return false, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return CheckStatusOK(r), BuildResponse(r)
	}
}

// DeleteSamlPrivateCertificate deletes the SAML IDP certificate from the server and updates the config to not use it and disable SAML.
func (c *Client4) DeleteSamlPrivateCertificate() (bool, *Response) {
	if r, err := c.DoApiDelete(c.GetSamlRoute() + "/certificate/private"); err != nil {
		return false, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return CheckStatusOK(r), BuildResponse(r)
	}
}

// GetSamlCertificateStatus returns metadata for the SAML configuration.
func (c *Client4) GetSamlCertificateStatus() (*SamlCertificateStatus, *Response) {
	if r, err := c.DoApiGet(c.GetSamlRoute()+"/certificate/status", ""); err != nil {
		return nil, BuildErrorResponse(r, err)
	} else {
		defer closeBody(r)
		return SamlCertificateStatusFromJson(r.Body), BuildResponse(r)
	}
}

