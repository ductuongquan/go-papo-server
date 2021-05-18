// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package api1

import (
	"bitbucket.org/enesyteam/papo-server/model"
	"net/http"
)

func (api *API) InitConversationTag() {
	api.BaseRoutes.Conversation.Handle("/tags", api.ApiSessionRequired(getConversationTags)).Methods("GET")
	api.BaseRoutes.Fanpage.Handle("/{conversation_id:[A-Za-z0-9]+}/tags", api.ApiSessionRequired(addOrRemoveConversationTag)).Methods("POST")
}

func getConversationTags(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireConversationId()
	if c.Err != nil {
		return
	}

	conversationId := c.Params.ConversationId
	if p, err := c.App.GetConversationTags(conversationId); err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.ToJson()))
		return
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(model.ConversationTagsToJson(p)))
	}
}

func addOrRemoveConversationTag(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireConversationId()
	c.RequirePageId()
	if c.Err != nil {
		return
	}

	tag := model.PageTagFromJson(r.Body)

	if tag == nil {
		c.SetInvalidParam("Nội dung")
		return
	}


	cTag := model.ConversationTag{
		ConversationId: c.Params.ConversationId,
		Creator: c.App.Session.UserId,
		TagId: tag.Id,
	}

	_, err := c.App.AddOrRemoveConversationTag(&cTag, c.Params.PageId, tag)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.ToJson()))
		return
	}

	ReturnStatusOK(w)
}