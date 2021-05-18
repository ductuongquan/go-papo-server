package api1

import (
	"bitbucket.org/enesyteam/papo-server/facebook_graph"
	"bitbucket.org/enesyteam/papo-server/model"
	"fmt"
	"net/http"
)

func (api *API) InitAttachment() {
	api.BaseRoutes.Attachments.Handle("/facebook/ids", api.ApiSessionRequired(getFacebookAttachmentByIds)).Methods("POST")
	api.BaseRoutes.Attachments.Handle("/facebook/target/ids", api.ApiSessionRequired(getFacebookAttachmentTargetByIds)).Methods("POST")
	api.BaseRoutes.Attachments.Handle("/facebook/{page_id:[A-Za-z0-9]+}/{message_id:[A-Za-z0-9\\_\\-\\.$+]+}", api.ApiSessionRequired(getFacebookMessageAttachments)).Methods("GET")
	api.BaseRoutes.Attachments.Handle("/facebook/{comment_id:[A-Za-z0-9\\_\\-\\.$+]+}/attachments", api.ApiSessionRequired(getFacebookCommentAttachments)).Methods("GET")
}

func getFacebookCommentAttachments(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireCommentId()
	if c.Err != nil {
		return
	}

	// check if client sent a query with page access token
	var pageAccessToken string
	query := r.URL.Query()
	if len(query.Get("page_access_token")) > 0 {
		pageAccessToken = query.Get("page_access_token")
		//fmt.Println(pageAccessToken)
	} else {
		fmt.Println("must get page access token")
	}

	if len(pageAccessToken) == 0 {
		c.Err = model.NewAppError("getFacebookAttachmentTargetByIds", "api.get_facebook_attachment_target_by_ids.app_error", nil, "", http.StatusUnauthorized)
		return
	}

	fErr, err, attachments := c.App.GraphCommentAttachments(c.Params.CommentId, pageAccessToken)
	if err != nil {
		c.Err = err
		return
	}

	if fErr != nil {
		w.Write([]byte(facebookgraph.FacebookErrorToJson(fErr)))
		return
	}

	w.Write([]byte(attachments.ToJson()))
}

func getFacebookMessageAttachments(c *Context, w http.ResponseWriter, r *http.Request) {

	c.RequirePageId()
	if c.Err != nil {
		return
	}

	c.RequireMessageId()
	if c.Err != nil {
		return
	}

	// check if client sent a query with page access token
	var pageAccessToken string
	query := r.URL.Query()
	if len(query.Get("page_access_token")) > 0 {
		pageAccessToken = query.Get("page_access_token")
		//fmt.Println(pageAccessToken)
	} else {
		fmt.Println("must get page access token")
	}

	if len(pageAccessToken) == 0 {
		c.Err = model.NewAppError("getFacebookAttachmentTargetByIds", "api.get_facebook_attachment_target_by_ids.app_error", nil, "", http.StatusUnauthorized)
		return
	}

	fErr, err, attachments := c.App.GraphMessageAttachments(c.Params.MessageId, pageAccessToken)
	if err != nil {
		c.Err = err
		return
	}

	if fErr != nil {
		w.Write([]byte(facebookgraph.FacebookErrorToJson(fErr)))
		return
	}

	w.Write([]byte(attachments.ToJson()))
}

func getFacebookAttachmentTargetByIds(c *Context, w http.ResponseWriter, r *http.Request) {
	// No permission check required

	ids := model.ArrayFromJson(r.Body)

	if len(ids) == 0 {
		c.SetInvalidParam("ids")
		return
	}

	if result := <-c.App.Srv.Store.User().GetById(c.App.Session.UserId); result.Err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(result.Err.ToJson()))
		return
	} else {
		user := result.Data.(*model.User)
		token := user.FacebookToken
		if len(token) == 0 {
			c.Err = model.NewAppError("getFacebookAttachmentTargetByIds", "api.get_facebook_attachment_target_by_ids.app_error", nil, "", http.StatusUnauthorized)
		}

		var attachmentType string
		query := r.URL.Query()
		attachmentType = query.Get("type")

		fErr, err, targetItem := c.App.GraphAttachmentTarget(ids[0], attachmentType, user.FacebookToken)
		if err != nil {
			c.Err = err
			return
		}

		if fErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(facebookgraph.FacebookErrorToJson(fErr)))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(targetItem.ToJson()))
	}
}

func getFacebookAttachmentByIds(c *Context, w http.ResponseWriter, r *http.Request) {
	ids := model.ArrayFromJson(r.Body)

	if len(ids) == 0 {
		c.SetInvalidParam("ids")
		return
	}

	// No permission check required

	attachments, err := c.App.GetFacebookAttachmentByIds(ids)
	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(model.FacebookAttachmentImageListToJson(attachments)))
}
