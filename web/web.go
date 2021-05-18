// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package web

import (
	"bitbucket.org/enesyteam/papo-server/app"
	"bitbucket.org/enesyteam/papo-server/mlog"
	"bitbucket.org/enesyteam/papo-server/model"
	"bitbucket.org/enesyteam/papo-server/services/configservice"
	"bitbucket.org/enesyteam/papo-server/utils"
	"fmt"
	"github.com/avct/uasurfer"
	"github.com/gorilla/mux"
	"net/http"
	"path"
	"strings"
)

type Web struct {
	GetGlobalAppOptions app.AppOptionCreator
	ConfigService       configservice.ConfigService
	MainRouter          *mux.Router
}

func New(config configservice.ConfigService, globalOptions app.AppOptionCreator, root *mux.Router) *Web {
	mlog.Debug("Initializing web routes")

	web := &Web{
		GetGlobalAppOptions: globalOptions,
		ConfigService:       config,
		MainRouter:          root,
	}

	web.InitWebhooks()
	web.InitSaml()
	web.InitStatic()

	return web
}

// Due to the complexities of UA detection and the ramifications of a misdetection
// only older Safari and IE browsers throw incompatibility errors.
// Map should be of minimum required browser version.
var browserMinimumSupported = map[string]int{
	"BrowserIE":     11,
	"BrowserSafari": 9,
}

func CheckClientCompatability(agentString string) bool {
	ua := uasurfer.Parse(agentString)

	if version, exist := browserMinimumSupported[ua.Browser.Name.String()]; exist && ua.Browser.Version.Major < version {
		return false
	}

	return true
}

func Handle404(config configservice.ConfigService, w http.ResponseWriter, r *http.Request) {
	err := model.NewAppError("Handle404", "api.context.404.app_error", nil, "", http.StatusNotFound)

	mlog.Debug(fmt.Sprintf("%v: code=404 ip=%v", r.URL.Path, utils.GetIpAddress(r)))

	if IsApiCall(config, r) {
		w.WriteHeader(err.StatusCode)
		err.DetailedError = "There doesn't appear to be an api call for the url='" + r.URL.Path + "'.  Typo? are you missing a team_id or user_id as part of the url?"
		w.Write([]byte(err.ToJson()))
	} else {
		utils.RenderWebAppError(config.Config(), w, r, err, config.AsymmetricSigningKey())
	}
}

func IsApiCall(config configservice.ConfigService, r *http.Request) bool {
	subpath, _ := utils.GetSubpathFromConfig(config.Config())

	return strings.HasPrefix(r.URL.Path, path.Join(subpath, "api")+"/")
}

func IsWebhookCall(a *app.App, r *http.Request) bool {
	subpath, _ := utils.GetSubpathFromConfig(a.Config())

	return strings.HasPrefix(r.URL.Path, path.Join(subpath, "hooks")+"/")
}

func ReturnStatusOK(w http.ResponseWriter) {
	m := make(map[string]string)
	m[model.STATUS] = model.STATUS_OK
	w.Write([]byte(model.MapToJson(m)))
}