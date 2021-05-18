// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"bitbucket.org/enesyteam/papo-server/mlog"
	"bitbucket.org/enesyteam/papo-server/model"
	"fmt"
)


func (a *App) AddOrRemoveConversationTag(tag *model.ConversationTag, pageId string, pageTag *model.PageTag) (*model.ConversationTag, *model.AppError) {

	result := <-a.Srv.Store.ConversationTag().SaveOrRemove(tag)

	if result.Err != nil {
		mlog.Error(fmt.Sprintf("Couldn't add or remove the tag from conversation err=%v", result.Err))
		return nil, result.Err
	}

	if result.Data != nil {
		rtag := result.Data.(*model.ConversationTag)
		message := model.NewWebSocketEvent(model.CONVERSATION_RECEIVED_TAG, "", pageId, "", nil)
		message.Add("new_tag", rtag)
		message.Add("page_tag", pageTag)
		a.Publish(message)

		return rtag, nil
	} else {
		message := model.NewWebSocketEvent(model.CONVERSATION_REMOVED_TAG, "", pageId, "", nil)
		message.Add("tag_id", tag.TagId)
		message.Add("page_tag", pageTag)
		message.Add("conversation_id", tag.ConversationId)
		message.Add("page_id", pageId)
		a.Publish(message)

		return nil, nil
	}
}

func (app *App) GetConversationTags(conversationId string) ([]*model.ConversationTag, *model.AppError) {
	result := <-app.Srv.Store.ConversationTag().GetConversationTags(conversationId)
	if result.Err != nil {
		return nil, result.Err
	}
	return result.Data.([]*model.ConversationTag), nil
}

// need method delete here