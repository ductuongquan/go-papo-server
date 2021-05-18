// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"bitbucket.org/enesyteam/papo-server/mlog"
	"bitbucket.org/enesyteam/papo-server/model"
	"fmt"
)

func (app *App) CreatePageTag(tag *model.PageTag) (*model.PageTag, *model.AppError) {
	result := <-app.Srv.Store.PageTag().Save(tag)
	if result.Err != nil {
		mlog.Error(fmt.Sprintf("Couldn't save the tag err=%v", result.Err))
		return nil, result.Err
	}
	rtag := result.Data.(*model.PageTag)

	message := model.NewWebSocketEvent(model.PAGE_TAG_CREATED, "", tag.PageId, "", nil)
	message.Add("new_page_tag", rtag)
	app.Publish(message)

	return rtag, nil
}

func (app *App) GetTag(id string) (*model.PageTag, *model.AppError) {
	result := <-app.Srv.Store.PageTag().Get(id)
	if result.Err != nil {
		return nil, result.Err
	}
	return result.Data.(*model.PageTag), nil
}

func (app *App) GetPageTags(pageId string) ([]*model.PageTag, *model.AppError) {
	result := <-app.Srv.Store.PageTag().GetPageTags(pageId)
	if result.Err != nil {
		return nil, result.Err
	}
	return result.Data.([]*model.PageTag), nil
}

func (app *App) UpdatePageTag(tag *model.PageTag) (*model.PageTag, *model.AppError) {
	result := <-app.Srv.Store.PageTag().Update(tag)
	if result.Err != nil {
		mlog.Error(fmt.Sprintf("Couldn't update the conversation tag err=%v", result.Err))
		return nil, result.Err
	}
	rTag := result.Data.(*model.PageTag)

	return rTag, nil
}