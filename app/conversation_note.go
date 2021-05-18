// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"bitbucket.org/enesyteam/papo-server/mlog"
	"bitbucket.org/enesyteam/papo-server/model"
	"fmt"
)

func (app *App) CreateConversationNote(note *model.ConversationNote) (*model.ConversationNote, *model.AppError) {
	result := <-app.Srv.Store.ConversationNote().Save(note)
	if result.Err != nil {
		mlog.Error(fmt.Sprintf("Couldn't save the note err=%v", result.Err))
		return nil, result.Err
	}
	rnote := result.Data.(*model.ConversationNote)

	return rnote, nil

	// Có thể gửi một message đến tất cả thành viên của page bằng websocket
	//message := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_NEW_FANPAGE, "", "", "", nil)
	//message.Add("fanpage_id", rfanpage.Id)
	//a.Publish(message)
}

func (app *App) GetNote(id string) (*model.ConversationNote, *model.AppError) {
	result := <-app.Srv.Store.ConversationNote().Get(id)
	if result.Err != nil {
		return nil, result.Err
	}
	return result.Data.(*model.ConversationNote), nil
}

func (app *App) GetConversationNotes(conversationId string) ([]*model.ConversationNote, *model.AppError) {
	result := <-app.Srv.Store.ConversationNote().GetConversationNotes(conversationId)
	if result.Err != nil {
		return nil, result.Err
	}
	return result.Data.([]*model.ConversationNote), nil
}

func (app *App) UpdateConversationNote(note *model.ConversationNote) (*model.ConversationNote, *model.AppError) {
	result := <-app.Srv.Store.ConversationNote().Update(note)
	if result.Err != nil {
		mlog.Error(fmt.Sprintf("Couldn't update the conversation note err=%v", result.Err))
		return nil, result.Err
	}
	rNote := result.Data.(*model.ConversationNote)

	return rNote, nil
}