// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package api1

import (
	"bitbucket.org/enesyteam/papo-server/model"
	"net/http"
)

func (api *API) InitPageTag() {
	api.BaseRoutes.Fanpage.Handle("/tags", api.ApiSessionRequired(getPageTags)).Methods("GET")
	api.BaseRoutes.Fanpage.Handle("/tags", api.ApiSessionRequired(createPageTag)).Methods("POST")
	api.BaseRoutes.PageTag.Handle("/update", api.ApiSessionRequired(updateTag)).Methods("PUT")
}

func getPageTags(c *Context, w http.ResponseWriter, r *http.Request) {
	pageId := c.Params.PageId
	if len(pageId) > 0 {
		if p, err := c.App.GetPageTags(pageId); err != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(err.ToJson()))
			return
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(model.PageTagsToJson(p)))
		}
	}
}

func createPageTag(c *Context, w http.ResponseWriter, r *http.Request) {
	tag := model.PageTagFromJson(r.Body)

	if tag == nil {
		c.SetInvalidParam("Nội dung")
		return
	}

	pageId := c.Params.PageId
	tag.PageId = pageId
	tag.Creator = c.App.Session.UserId

	if len(pageId) == 0 {
		c.SetInvalidUrlParam("PageId")
		return
	}

	if len(tag.Name) == 0 {
		c.SetInvalidParam("Tên nhãn")
		return
	}

	if len(tag.Color) == 0 {
		c.SetInvalidParam("Màu sắc")
		return
	}

	var rTag *model.PageTag
	var err *model.AppError

	rTag, err = c.App.CreatePageTag(tag)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.ToJson()))
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(rTag.ToJson()))
}

// cần fix lại, không cho phép update tag của page khác
// => check page_id
func updateTag(c *Context, w http.ResponseWriter, r *http.Request) {
	tagId := c.Params.TagId

	tag := model.PageTagFromJson(r.Body)
	if tag == nil {
		c.SetInvalidParam("something")
		return
	}
	tag.Id = tagId

	pageId := c.Params.PageId
	tag.PageId = pageId

	if len(tag.Name) == 0 {

	}

	if len(tag.Color) == 0 {

	}

	var rTag *model.PageTag
	var err *model.AppError

	rTag, err = c.App.UpdatePageTag(tag)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.ToJson()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(rTag.ToJson()))
}