// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"bitbucket.org/enesyteam/papo-server/mlog"
	"bitbucket.org/enesyteam/papo-server/model"
	"fmt"
	"time"
)

const (
	PENDING_MESSAGE_IDS_CACHE_SIZE = 25000
	PENDING_MESSAGE_IDS_CACHE_TTL  = 30 * time.Second
	PAGE_DEFAULT                = 0
)

func (app *App) CreateConversation(conversation *model.FacebookConversation) (*model.FacebookConversation, *model.AppError) {
	result := <-app.Srv.Store.FacebookConversation().Save(conversation)
	if result.Err != nil {
		mlog.Error(fmt.Sprintf("Couldn't save the conversation err=%v", result.Err))
		return nil, result.Err
	}
	rcvs := result.Data.(*model.FacebookConversation)

	return rcvs, nil

	// Websocket khi tạo xong conversation
	//	//message := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_NEW_FANPAGE, "", "", "", nil)
	//	//message.Add("fanpage_id", rfanpage.Id)
	//	//a.Publish(message)
}

func (app *App) UpsertConversation(conversation *model.FacebookConversation) (*model.FacebookConversation, *model.AppError) {
	result := <-app.Srv.Store.FacebookConversation().UpsertCommentConversation(conversation)
	if result.Err != nil {
		mlog.Error(fmt.Sprintf("Couldn't save the conversation err=%v", result.Err))
		return nil, result.Err
	}
	rcvs := result.Data.(*model.FacebookConversation)

	return rcvs, nil
}

// Cần 1 method thêm message vào Conversation
// AddMessageToConversation()

// Trả về 3 giá trị
// 1. Con trỏ tới fanpage member nếu thêm thành công
// 2. bool: true nếu user đã được thêm vào fanpage từ trước, fale nếu user chưa được thêm
// 3. Contrỏ tới 1 AppError nếu xảy ra lỗi
func (app *App) addMessageToConversation(mesage *model.FacebookConversationMessage, shouldUpdateConversation bool, isFromPage bool) (*model.FacebookConversationMessage, bool, *model.AppError) {

	efmr := <-app.Srv.Store.FacebookConversation().GetMessage(mesage.Id)
	if efmr.Err != nil {
		// chưa có, thêm ngay
		fmr := <-app.Srv.Store.FacebookConversation().AddMessage(mesage, shouldUpdateConversation, isFromPage)
		if fmr.Err != nil {
			return nil, false, fmr.Err
		}
		return fmr.Data.(*model.FacebookConversationMessage), false, nil
	}

	// already exists.  Check if deleted and and update, otherwise do nothing
	rfm := efmr.Data.(*model.FacebookConversationMessage)

	return rfm, true, nil
}

func (a *App) UpdateConversationPageScopeId(conversationId, pageScopeId string) *model.AppError {
	result := <-a.Srv.Store.FacebookConversation().UpdatePageScopeId(conversationId, pageScopeId)
	if result.Err != nil {
		return result.Err
	}

	// TODO: Emit socket to clients to update page scope id for replying
	//message := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_CONVERSATION_SEEN, "", pageId, "", nil)
	//message.Add("conversation_id", id)
	//a.Publish(message)

	return nil
}


func (a *App) UpdateSeen(id string, pageId string, userId string) (string, *model.AppError) {
	result := <-a.Srv.Store.FacebookConversation().UpdateSeen(id, pageId, userId)
	if result.Err != nil {
		return "", result.Err
	}

	message := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_CONVERSATION_SEEN, "", pageId, "", nil)
	message.Add("conversation_id", id)
	a.Publish(message)

	return result.Data.(string), nil
}

func (app *App) UpdateUnSeen(id string, pageId string, userId string) (*model.FacebookConversation, *model.AppError) {
	result := <-app.Srv.Store.FacebookConversation().UpdateUnSeen(id, pageId, userId)
	if result.Err != nil {
		return nil, result.Err
	}

	// Websocket
	message := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_CONVERSATION_UNSEEN, "", pageId, "", nil)
	message.Add("conversation", result.Data.(*model.FacebookConversation))
	app.Publish(message)

	return result.Data.(*model.FacebookConversation), nil
}


func (app *App) UpdateReadWatermark(id string, pageId string, timestamp int64) *model.AppError {
	result := <-app.Srv.Store.FacebookConversation().UpdateReadWatermark(id, pageId, timestamp)
	if result.Err != nil {
		return result.Err
	}
	return nil
}

func (app *App) AddMessage(message *model.FacebookConversationMessage, shouldSendWebhook bool, shouldUpdateConversation bool, isFromPage bool) (*model.FacebookConversationMessage, *model.FacebookConversation, *model.AppError) {
	if len(message.ConversationId) == 0 {
		fmt.Println("Kiểm tra lại, không cho phép lưu message không thuộc về conversation nào")
	}
	ms, _, err := app.addMessageToConversation(message, shouldUpdateConversation, isFromPage)
	if err != nil {
		return nil, nil, err
	}

	//else {
	//	if shouldSendWebhook {
	//		rConversation := <-app.Srv.Store.FacebookConversation().Get(message.ConversationId)
	//		if rConversation.Data != nil {
	//			foundConversation := rConversation.Data.(*model.FacebookConversation)
	//			if foundConversation != nil {
	//				webhookData := model.NewWebSocketEvent(model.RECEIVE_CONVERSATION_UPDATED, "", message.PageId, "", nil)
	//				webhookData.Add("id", foundConversation.Id)
	//				webhookData.Add("conversation", foundConversation)
	//				webhookData.Add("newMessage", message)
	//				app.Publish(webhookData)
	//				return ms, foundConversation, nil
	//			}
	//		} else {
	//			return ms, nil, nil
	//		}
	//	}
	//}

	return ms, nil, nil
}

func (a *App) attachFilesToMessage(message *model.FacebookConversationMessage) *model.AppError {
	var attachedIds []string
	for _, fileId := range message.FileIds {
		result := <-a.Srv.Store.FileInfo().AttachToMessage(fileId, message.Id, message.UserId)
		if result.Err != nil {
			mlog.Warn("Failed to attach file to message", mlog.String("file_id", fileId), mlog.String("message_id", message.Id), mlog.Err(result.Err))
			continue
		}

		attachedIds = append(attachedIds, fileId)
	}

	if len(message.FileIds) != len(attachedIds) {
		// We couldn't attach all files to the post, so ensure that post.FileIds reflects what was actually attached
		message.FileIds = attachedIds

		result := <-a.Srv.Store.FacebookConversation().OverwriteMessage(message)
		if result.Err != nil {
			return result.Err
		}
	}

	return nil
}

func (app *App) AddImage(image *model.FacebookAttachmentImage) (*model.FacebookAttachmentImage, *model.AppError) {
	result := <-app.Srv.Store.FacebookConversation().AddImage(image)
	if result.Err != nil {
		return nil, result.Err
	}
	return result.Data.(*model.FacebookAttachmentImage), nil
}

//func (app *App) GetConversationMessages(conversationId string, offset, limit int) (*model.ConversationMessageResponse, *model.AppError) {
//	result := <-app.Srv.Store.FacebookConversation().GetMessagesByConversationId(conversationId, offset, limit)
//	if result.Err != nil {
//		return nil, result.Err
//	}
//	return result.Data.(*model.ConversationMessageResponse), nil
//}

func (app *App) GetConversationMessages(conversationId string, offset, limit int) ([]*model.FacebookConversationMessage, *model.AppError) {
	result := <-app.Srv.Store.FacebookConversation().GetMessagesByConversationId(conversationId, offset, limit)
	if result.Err != nil {
		return nil, result.Err
	}
	return result.Data.([]*model.FacebookConversationMessage), nil
}

func (a *App) GetConversations(pageIds string, offset, limit int) ([]*model.FacebookConversation, *model.AppError) {
	result := <-a.Srv.Store.FacebookConversation().GetConversations(pageIds, offset, limit)

	if result.Err != nil {
		return nil, result.Err
	}
	return result.Data.([]*model.FacebookConversation), nil
}

func (a *App) SearchConversations(term string, pageIds string, limit, offset int) ([]*model.ConversationResponse, *model.AppError) {
	result := <-a.Srv.Store.FacebookConversation().Search(term, pageIds, limit, offset)
	if result.Err != nil {
		return nil, result.Err
	}
	return result.Data.([]*model.ConversationResponse), nil
}

func (a *App) SanitizeMessage(session model.Session, message *model.FacebookConversationMessage) *model.FacebookConversationMessage {
	// cần kiểm tra xem user có quyền truy cập page không
	// tạm thời chưa dùng đến
	//if !a.SessionHasPermissionToFanpage(session, fanpage.Id, model.PERMISSION_MANAGE_FANPAGE) {
	//	fanpage.Sanitize()
	//}

	return message
}

func (app *App) SanitizeMessages(session model.Session, messages []*model.FacebookConversationMessage) []*model.FacebookConversationMessage {
	for _, message := range messages {
		app.SanitizeMessage(session, message)
	}
	return messages
}
