// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package web

import (
	"bitbucket.org/enesyteam/papo-server/facebook_graph"
	"bitbucket.org/enesyteam/papo-server/mlog"
	"bitbucket.org/enesyteam/papo-server/model"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

func (w *Web) InitWebhooks() {
	w.MainRouter.Handle("/hooks/commands/{id:[A-Za-z0-9]+}", w.NewHandler(commandWebhook)).Methods("POST")
	w.MainRouter.Handle("/hooks/{id:[A-Za-z0-9]+}", w.NewHandler(incomingWebhook)).Methods("POST")

	//w.MainRouter.Handle("/hooks", w.NewHandler(getWebhook)).Methods("GET")
	//w.MainRouter.Handle("/hooks", w.NewHandler(postWebhook)).Methods("POST")

	w.MainRouter.Handle("/webhooks/facebook", w.NewHandler(verifyWebhook)).Methods("GET")
	w.MainRouter.Handle("/webhooks/facebook", w.NewHandler(handleFacebookWebhook)).Methods("POST")
}

func verifyWebhook(c *Context, w http.ResponseWriter, r *http.Request) {
	if len(r.URL.Query().Get("hub.verify_token")) > 0 && r.URL.Query().Get("hub.verify_token") == *c.App.Config().FacebookAPISettings.WebhookToken {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(r.URL.Query().Get("hub.challenge")))
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error"))
	}
}

func handleFacebookWebhook(c *Context, w http.ResponseWriter, r *http.Request) {
	var z *facebookgraph.HubEntries
	x, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(x, &z)

	if &z.Entry[0] != nil {
		err := c.App.HandleFacebookWebhook(z)
		if err == nil {
			w.WriteHeader(http.StatusOK)
			return
		} else {
			fmt.Println("================error when receive wehbook" + err.Message + err.DetailedError)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	} else {
		// return status 200 for other webhook that we don't handle
		w.WriteHeader(http.StatusOK)
		return
	}
}

func incomingWebhook(c *Context, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	r.ParseForm()

	var err *model.AppError
	incomingWebhookPayload := &model.IncomingWebhookRequest{}
	contentType := r.Header.Get("Content-Type")
	if strings.Split(contentType, "; ")[0] == "application/x-www-form-urlencoded" {
		payload := strings.NewReader(r.FormValue("payload"))

		incomingWebhookPayload, err = decodePayload(payload)
		if err != nil {
			c.Err = err
			return
		}
	} else if strings.HasPrefix(contentType, "multipart/form-data") {
		r.ParseMultipartForm(0)

		decoder := schema.NewDecoder()
		err := decoder.Decode(incomingWebhookPayload, r.PostForm)

		if err != nil {
			c.Err = model.NewAppError("incomingWebhook", "api.webhook.incoming.error", nil, err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		incomingWebhookPayload, err = decodePayload(r.Body)
		if err != nil {
			c.Err = err
			return
		}
	}

	if *c.App.Config().LogSettings.EnableWebhookDebugging {
		mlog.Debug(fmt.Sprintf("Incoming webhook received. Id=%s Content=%s", id, incomingWebhookPayload.ToJson()))
	}

	err = c.App.HandleIncomingWebhook(id, incomingWebhookPayload)
	if err != nil {
		c.Err = err
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("ok"))
}

func commandWebhook(c *Context, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	response, err := model.CommandResponseFromHTTPBody(r.Header.Get("Content-Type"), r.Body)
	if err != nil {
		c.Err = model.NewAppError("commandWebhook", "web.command_webhook.parse.app_error", nil, err.Error(), http.StatusBadRequest)
		return
	}

	appErr := c.App.HandleCommandWebhook(id, response)
	if appErr != nil {
		c.Err = appErr
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("ok"))
}

func decodePayload(payload io.Reader) (*model.IncomingWebhookRequest, *model.AppError) {
	incomingWebhookPayload, decodeError := model.IncomingWebhookRequestFromJson(payload)

	if decodeError != nil {
		return nil, decodeError
	}

	return incomingWebhookPayload, nil
}
