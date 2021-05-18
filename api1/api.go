// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package api1

import (
	"bitbucket.org/enesyteam/papo-server/app"
	"bitbucket.org/enesyteam/papo-server/model"
	"bitbucket.org/enesyteam/papo-server/services/configservice"
	"bitbucket.org/enesyteam/papo-server/web"
	"github.com/gorilla/mux"
	_ "github.com/nicksnyder/go-i18n/i18n"
	"net/http"
)

type Routes struct {
	Root    					*mux.Router // ''
	ApiRoot 					*mux.Router // 'api/v1'

	Users 						*mux.Router // 'api/v1/users'
	User  						*mux.Router // 'api/v1/users/{user_id:[A-Za-z0-9]+}'

	Teams              			*mux.Router // 'api/v1/teams'
	TeamsForUser       			*mux.Router // 'api/v1/users/{user_id:[A-Za-z0-9]+}/teams'
	Team               			*mux.Router // 'api/v1/teams/{team_id:[A-Za-z0-9]+}'
	TeamForUser        			*mux.Router // 'api/v1/users/{user_id:[A-Za-z0-9]+}/teams/{team_id:[A-Za-z0-9]+}'
	TeamByName         			*mux.Router // 'api/v1/teams/name/{team_name:[A-Za-z0-9_-]+}'
	TeamMembers        			*mux.Router // 'api/v1/teams/{team_id:[A-Za-z0-9_-]+}/members'
	TeamMember         			*mux.Router // 'api/v1/teams/{team_id:[A-Za-z0-9_-]+}/members/{user_id:[A-Za-z0-9_-]+}'
	TeamMembersForUser 			*mux.Router // 'api/v1/users/{user_id:[A-Za-z0-9]+}/teams/members'

	OAuth     					*mux.Router // 'api/v1/oauth'
	OAuthApps 					*mux.Router // 'api/v1/oauth/apps'
	OAuthApp  					*mux.Router // 'api/v1/oauth/apps/{app_id:[A-Za-z0-9]+}'

	Orders						*mux.Router // 'api/v1/orders'
	Order 						*mux.Router // 'api/v1/orders/{order_id:[A-Za-z0-9_-]+}

	Elasticsearch 				*mux.Router // 'api/v1/elasticsearch'

	FacebookUsers        		*mux.Router // 'api/v1/facebookusers'
	Attachments 				*mux.Router // 'api/v1/attachments'

	Fanpages        			*mux.Router // 'api/v1/fanpages'
	FanpagesForUser 			*mux.Router // 'api/v1/users/{user_id:[A-Za-z0-9]+}/fanpages'
	Fanpage         			*mux.Router // 'api/v1/fanpages/{fanpage_id:[A-Za-z0-9]+}'

	Posts 						*mux.Router // 'api/v1/posts
	Post 						*mux.Router // 'api/v1/posts/{post_id:[A-Za-z0-9]+}'
	PagePosts 					*mux.Router // 'api/v1/fanpages

	Conversations 				*mux.Router // 'api/v1/conversations'
	Conversation  				*mux.Router // 'api/v1/conversations/{conversation_id:[A-Za-z0-9]+}'

	Test 						*mux.Router

	PageTags 					*mux.Router // 'api/v1/fanpages/tags'
	PageTag 					*mux.Router // 'api/v1/fanpages/{page_tag_id:[A-Za-z0-9]+}'

	Roles   					*mux.Router // 'api/v1/roles'
	Schemes 					*mux.Router // 'api/v1/schemes'

	ConversationNote 			*mux.Router // 'api/v1/conversations/notes/{note_id:[A-Za-z0-9]+}'

	System 						*mux.Router // 'api/v1/system'
	Jobs 						*mux.Router // 'api/v1/jobs'

	Files 						*mux.Router // 'api/v1/files'
	File  						*mux.Router // 'api/v1/files/{file_id:[A-Za-z0-9]+}'
	PublicFile 					*mux.Router // 'files/{file_id:[A-Za-z0-9]+}/public'
	Preferences 				*mux.Router // 'api/v1/users/{user_id:[A-Za-z0-9]+}/preferences'
	License 					*mux.Router // 'api/v1/license'

	OpenGraph 					*mux.Router // 'api/v1/opengraph'

	WebHooks					*mux.Router //api/v1/webhooks
}

type API struct {
	ConfigService       configservice.ConfigService
	GetGlobalAppOptions app.AppOptionCreator
	BaseRoutes          *Routes
}

func Init(configservice configservice.ConfigService, globalOptionsFunc app.AppOptionCreator, root *mux.Router) *API {
	api := &API{
		ConfigService:       configservice,
		GetGlobalAppOptions: globalOptionsFunc,
		BaseRoutes:          &Routes{},
	}

	api.BaseRoutes.Root = root
	api.BaseRoutes.ApiRoot = root.PathPrefix(model.API_URL_SUFFIX).Subrouter()

	api.BaseRoutes.Users = api.BaseRoutes.ApiRoot.PathPrefix("/users").Subrouter()
	api.BaseRoutes.User = api.BaseRoutes.ApiRoot.PathPrefix("/users/{user_id:[A-Za-z0-9]+}").Subrouter()

	api.BaseRoutes.Teams = api.BaseRoutes.ApiRoot.PathPrefix("/teams").Subrouter()
	api.BaseRoutes.TeamsForUser = api.BaseRoutes.User.PathPrefix("/teams").Subrouter()
	api.BaseRoutes.Team = api.BaseRoutes.Teams.PathPrefix("/{team_id:[A-Za-z0-9]+}").Subrouter()
	api.BaseRoutes.TeamForUser = api.BaseRoutes.TeamsForUser.PathPrefix("/{team_id:[A-Za-z0-9]+}").Subrouter()
	api.BaseRoutes.TeamByName = api.BaseRoutes.Teams.PathPrefix("/name/{team_name:[A-Za-z0-9_-]+}").Subrouter()
	api.BaseRoutes.TeamMembers = api.BaseRoutes.Team.PathPrefix("/members").Subrouter()
	api.BaseRoutes.TeamMember = api.BaseRoutes.TeamMembers.PathPrefix("/{user_id:[A-Za-z0-9]+}").Subrouter()
	api.BaseRoutes.TeamMembersForUser = api.BaseRoutes.User.PathPrefix("/teams/members").Subrouter()

	api.BaseRoutes.OAuth = api.BaseRoutes.ApiRoot.PathPrefix("/oauth").Subrouter()
	api.BaseRoutes.OAuthApps = api.BaseRoutes.OAuth.PathPrefix("/apps").Subrouter()
	api.BaseRoutes.OAuthApp = api.BaseRoutes.OAuthApps.PathPrefix("/{app_id:[A-Za-z0-9]+}").Subrouter()

	// fanpages
	api.BaseRoutes.Fanpages = api.BaseRoutes.ApiRoot.PathPrefix("/fanpages").Subrouter()
	api.BaseRoutes.FanpagesForUser = api.BaseRoutes.User.PathPrefix("/fanpages").Subrouter()
	api.BaseRoutes.Fanpage = api.BaseRoutes.Fanpages.PathPrefix("/{page_id:[A-Za-z0-9]+}").Subrouter()

	// orders
	api.BaseRoutes.Orders = api.BaseRoutes.ApiRoot.PathPrefix("/orders").Subrouter()
	api.BaseRoutes.Order = api.BaseRoutes.Orders.PathPrefix("/{order_id:[A-Za-z0-9]+}").Subrouter()

	// posts
	//api.BaseRoutes.Posts = api.BaseRoutes.ApiRoot.PathPrefix("/posts").Subrouter()
	//api.BaseRoutes.Post = api.BaseRoutes.Fanpage.PathPrefix("/posts/{post_id:[A-Za-z0-9_-]+}").Subrouter()

	api.BaseRoutes.Files = api.BaseRoutes.ApiRoot.PathPrefix("/files").Subrouter()
	api.BaseRoutes.File = api.BaseRoutes.Files.PathPrefix("/{file_id:[A-Za-z0-9]+}").Subrouter()
	api.BaseRoutes.PublicFile = api.BaseRoutes.Root.PathPrefix("/files/{file_id:[A-Za-z0-9]+}/public").Subrouter()

	// facebook users
	api.BaseRoutes.FacebookUsers = api.BaseRoutes.ApiRoot.PathPrefix("/facebookusers").Subrouter()

	// attachments
	api.BaseRoutes.Attachments = api.BaseRoutes.ApiRoot.PathPrefix("/attachments").Subrouter()

	// conversations
	api.BaseRoutes.Conversations = api.BaseRoutes.ApiRoot.PathPrefix("/conversations").Subrouter()
	api.BaseRoutes.Conversation = api.BaseRoutes.Conversations.PathPrefix("/{conversation_id:[A-Za-z0-9]+}").Subrouter()

	// page tag
	api.BaseRoutes.PageTags = api.BaseRoutes.Fanpage.PathPrefix("/tags").Subrouter()
	api.BaseRoutes.PageTag = api.BaseRoutes.PageTags.PathPrefix("/{tag_id:[A-Za-z0-9]+}").Subrouter()

	api.BaseRoutes.ConversationNote = api.BaseRoutes.Conversation.PathPrefix("/{note_id:[A-Za-z0-9]+}").Subrouter()
	// test
	api.BaseRoutes.Test = api.BaseRoutes.ApiRoot.PathPrefix("/test").Subrouter()

	api.BaseRoutes.System = api.BaseRoutes.ApiRoot.PathPrefix("/system").Subrouter()
	api.BaseRoutes.Jobs = api.BaseRoutes.ApiRoot.PathPrefix("/jobs").Subrouter()
	api.BaseRoutes.Preferences = api.BaseRoutes.User.PathPrefix("/preferences").Subrouter()
	api.BaseRoutes.License = api.BaseRoutes.ApiRoot.PathPrefix("/license").Subrouter()

	api.BaseRoutes.Roles = api.BaseRoutes.ApiRoot.PathPrefix("/roles").Subrouter()
	api.BaseRoutes.Schemes = api.BaseRoutes.ApiRoot.PathPrefix("/schemes").Subrouter()

	api.BaseRoutes.Elasticsearch = api.BaseRoutes.ApiRoot.PathPrefix("/elasticsearch").Subrouter()

	api.BaseRoutes.OpenGraph = api.BaseRoutes.ApiRoot.PathPrefix("/opengraph").Subrouter()

	api.BaseRoutes.WebHooks = api.BaseRoutes.ApiRoot.PathPrefix("/webhooks").Subrouter()

	// init
	api.InitUser()
	api.InitTeam()
	api.InitSystem()
	api.InitOAuth()
	api.InitFile()

	api.InitFacebookUid()
	api.InitAttachment()
	api.InitFanpage()
	api.InitPost()
	api.InitFacebookConversation()
	api.InitPageTag()
	api.InitConversationTag()
	api.InitConversationNote()
	api.InitPreference()
	api.InitWebSocket()
	api.InitRole()
	api.InitScheme()
	api.InitJob()
	api.InitStatus()
	api.InitElasticsearch()
	api.InitOrders()
	api.InitOpenGraph()
	//api.InitWebHooks()

	root.Handle("/api/v1/{anything:.*}", http.HandlerFunc(api.Handle404))
	return api
}

func (api *API) Handle404(w http.ResponseWriter, r *http.Request) {
	web.Handle404(api.ConfigService, w, r)
}

var ReturnStatusOK = web.ReturnStatusOK
