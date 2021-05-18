// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"bitbucket.org/enesyteam/papo-server/facebook_graph"
	"bitbucket.org/enesyteam/papo-server/mlog"
	"bitbucket.org/enesyteam/papo-server/model"
	"bitbucket.org/enesyteam/papo-server/utils"
	"fmt"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"strings"
	"time"
)

func (app *App) HandleFacebookWebhook(hubEntries *facebookgraph.HubEntries) *model.AppError {
	if hubEntries != nil && len(hubEntries.Entry) > 0 {
		entry := hubEntries.Entry[0]
		if len(entry.Messaging) > 0 {
			event := entry.Messaging[0]

			if len(event.Message.Text) > 0 || len(event.Message.Attachments) > 0 {
				err := app.receiveMessage(event)

				if event.Message.QuickReply.Payload != nil {
					// at this time, we don't known what to do with quick reply message
					//fmt.Println("handle quick reply message")
					//fmt.Println(event.Message.QuickReply)
					return nil
				}

				if err != nil {
					return err
				}

				return nil

			} else if event.PostBack.Payload != nil {
				fmt.Println("Sender...", event.Sender)
				fmt.Println("Recipient...", event.Recipient)
				fmt.Println(event.PostBack)
				fmt.Println("handle message post back")
			} else if len(event.Delivery.Mids) > 0 {
				fmt.Println("Sender...", event.Sender)
				fmt.Println("Recipient...", event.Recipient)
				fmt.Println("handle message delivery")
				fmt.Println(event.Delivery)
				isUpdateError := false
				var updateError *model.AppError
				for _, mid := range event.Delivery.Mids {
					err := app.receiveMessageDelivery(event.Recipient.Id, "m_"+mid)
					if err != nil {
						updateError = err
					}
					isUpdateError = isUpdateError && err != nil
				}
				return updateError
			} else if event.Read.Watermark > 0 {
				fmt.Println("Sender...", event.Sender)
				fmt.Println("Recipient...", event.Recipient)
				fmt.Println("handle message read")
				fmt.Println(event.Read)
				err := app.receivedMessageRead(event.Recipient.Id, event.Sender.Id, event.Read.Watermark)
				if err != nil {
					return err
				}

				return nil

			} else {
				fmt.Println("Webhook received unhandle event: ", event)
				return nil
			}

		} else if len(entry.Changes) > 0 {

			if entry.Changes[0].Field == "feed" {
				// handle feed changes
				change := entry.Changes[0]
				fmt.Println("change", change)

				if change.Value.Item == "comment" {
					if err := app.receiveComment(change); err != nil {
						fmt.Println("----Loi webhook comment")
						fmt.Println(err)
						return err
					}
					return nil
				} else if change.Value.Item == "post" {
					fmt.Println("receive post")
					fmt.Println("change", change)
				} else if change.Value.Item == "status" || change.Value.Item == "photo" {
					fmt.Println("Page đăng trạng thái mới có thể đính kèm ảnh")
					//fmt.Println("change", change)
					app.receivePost(change)
				} else {
					fmt.Println("unhandle changes")
				}

				return nil

			} else if entry.Changes[0].Field == "conversations" {
				// do not handle anything here
				return nil
				//fmt.Println("handle message")
				//fmt.Println("entry.Changes[0].Value", entry.Changes[0].Value)
				//fmt.Println("From: ", entry.Id)
				//fmt.Println("entry.Changes[0].Value.PageId: ", entry.Changes[0].Value.PageId)
				//fmt.Println("entry.Changes[0].Value.ThreadId: ", entry.Changes[0].Value.ThreadId)
			} else {
				return nil
			}
		}
	}

	return nil
}

func (app *App) receivedMessageRead(pageId, userId string, timestamp int64) *model.AppError {
	result := <-app.Srv.Store.FacebookConversation().GetPageConversationBySenderId(pageId, userId, "message")

	var conversations []*model.FacebookConversation
	var conversation *model.FacebookConversation

	if result.Err != nil {
		// không tìm thấy hội thoại nào từ user này => bỏ qua cập nhật
		return nil
	} else {
		conversations = result.Data.([]*model.FacebookConversation)
	}

	if len(conversations) > 0 {
		if len(conversations) > 1 {
			// lỗi, 1 page chỉ có duy nhất 1 hội thoại tin nhắn từ 1 user nào đó, bỏ qua cập nhật
			return nil
		} else {
			conversation = conversations[0]
		}
	} else {
		// không tìm thấy hội thoại => bỏ qua cập nhật
		return nil
	}

	if conversation != nil {
		if error := app.UpdateReadWatermark(conversation.Id, pageId, timestamp); error != nil {
			// bỏ qua lỗi
			return nil
		} else {
			webhookData := model.NewWebSocketEvent(model.RECEIVE_CONVERSATION_READ, "", pageId, "", nil)
			webhookData.Add("id", conversation.Id)
			webhookData.Add("read_watermark", timestamp)
			app.Publish(webhookData)
		}
	}

	return nil
}

func (app *App) receiveMessageDelivery(pageId, messageId string) *model.AppError {
	fmt.Println("receive message delivery: pageId: ", pageId, "-messageId: ", messageId )
	if len(messageId) > 0 {
		updateResult := <-app.Srv.Store.FacebookConversation().UpdateMessageSent(messageId)
		if updateResult.Err == nil {
			webhookData := model.NewWebSocketEvent(model.MESSAGE_SENT, "", pageId, "", nil)
			webhookData.Add("pageId", pageId)
			webhookData.Add("messageId", messageId)
			app.Publish(webhookData)

		} else {
			mlog.Error(fmt.Sprintf("Couldn't update message id err=%v", updateResult.Err))
		}
		return nil
	}
	return nil
}

func (app *App) receiveMessage(message facebookgraph.HubEntryMessaging) *model.AppError {
	fmt.Println("messagemessagemessage=", message)

	messageId := "m_" + message.Message.Mid
	senderId := message.Sender.Id
	receptionId := message.Recipient.Id
	messageText := message.Message.Text
	messageTime := message.Timestamp
	isEcho := message.Message.IsEcho
	stickerId := message.Message.StickerId
	attachments := message.Message.Attachments

	var from string
	var pageId string

	if isEcho {
		from = receptionId
		pageId = senderId
	} else {
		from = senderId
		pageId = receptionId
	}

	if len(from) == 0 || len(pageId) == 0 {
		return model.NewAppError("receiveTextMessage", "webhook.facebook_missing_information.app_error", nil, "", http.StatusBadRequest)
	}

	var attachmentType string
	if len(attachments) > 0 {

		if len(attachments) == 0 {
			attachmentType = attachments[0].Type
		} else {
			for _, attachment := range attachments {
				if len(attachmentType) == 0 {
					attachmentType = attachment.Type
				}
				// TODO: what to do if user send multiple attachments with difference types?
			}
		}
	}

	var snippet string
	if len(messageText) > 0 {
		snippet = utils.GetSnippet(messageText)
	} else {
		if len(attachments) > 0 {
			snippet = attachmentType
		} else {
			snippet = "[Error snippet]"
		}
	}

	var conversations []*model.FacebookConversation
	var conversation *model.FacebookConversation
	result := <-app.Srv.Store.FacebookConversation().GetPageConversationBySenderId(pageId, from, "message")
	if result.Err != nil {
		newConversation := model.FacebookConversation {
			Type: "message",
			PageId: pageId,
			From: senderId,
			UpdatedTime: time.Unix(messageTime/1000, 10).Format(time.RFC3339),
			Snippet: snippet,
		}

		cResult := <-app.Srv.Store.FacebookConversation().Save(&newConversation)
		if cResult.Err == nil {
			conversation = cResult.Data.(*model.FacebookConversation)
		} else {
			return model.NewAppError("receiveTextMessage", "webhook.facebook_add_new_conversation.app_error", nil, "", http.StatusBadRequest)
		}
	} else {
		conversations = result.Data.([]*model.FacebookConversation)
	}

	if len(conversations) > 0 {
		if len(conversations) > 1 {
			return model.NewAppError("receiveTextMessage", "webhook.facebook_duplicate_conversation_from_user.app_error", nil, pageId + from, http.StatusBadRequest)
		} else {
			conversation = conversations[0]
		}
	}

	// finnaly check to ensure conversation is not nil
	if conversation == nil {
		return model.NewAppError("receiveTextMessage", "webhook.facebook_general_error.app_error", nil, "pageId: " + pageId + " from: " + from, http.StatusBadRequest)
	}


	var conversationMessage *model.FacebookConversationMessage

	// message is sent by page, check if message sent from papo.
	// if a message sent from papo, it have created before the time received this webhook
	// so we need to find and update this message
	if isEcho {
		mResult := <-app.Srv.Store.FacebookConversation().GetPageMessageByMid(pageId, messageId)
		if mResult.Err == nil {
			conversationMessage = mResult.Data.(*model.FacebookConversationMessage)
		}
	}

	if conversationMessage == nil {
		newMessage := &model.FacebookConversationMessage{
			ConversationId: conversation.Id,
			Type: 			"message",
			PageId: 		pageId,
			MessageId:      messageId,
			Message:        messageText,
			CreatedTime:    time.Unix(messageTime/1000, 10).Format(time.RFC3339),
			From:           senderId,
			Sent: 			true,
		}

		if stickerId > 0 {
			newMessage.Sticker = attachments[0].Payload["url"].(string)
		} else if len(attachments) > 0 {
			newMessage.HasAttachments = true
			newMessage.AttachmentsCount = len(attachments)
			newMessage.AttachmentType = attachmentType
		}

		addedMessage, _, err := app.AddMessage(newMessage, true, true, isEcho)
		if err != nil {
			return model.NewAppError("receiveTextMessage", "webhook.facebook_add_message_error.app_error", nil, "", http.StatusBadRequest)
		} else {
			conversationMessage = addedMessage
		}
	}

	if conversationMessage != nil {

		var omitUsers map[string]bool
		if len(conversationMessage.UserId) > 0 {
			omitUsers = map[string]bool{
				conversationMessage.UserId: true,
			}
		}

		webhookData := model.NewWebSocketEvent(model.RECEIVE_CONVERSATION_UPDATED, "", pageId, "", omitUsers)
		webhookData.Add("id", conversation.Id)
		webhookData.Add("conversation", conversation)
		webhookData.Add("newMessage", conversationMessage)
		app.Publish(webhookData)
	}

	return nil
}

func (app *App) receivePost(change facebookgraph.HubEntryChange) *model.AppError {
	rawPost := change.Value
	if rawPost.Verb == "add" {
		postId := rawPost.PostId
		message := rawPost.Message
		createdTime := time.Unix(rawPost.CreatedTime, 10).String()
		picture := rawPost.Link

		post := model.FacebookPost{
			PostId: postId,
			Message: message,
			CreatedTime: createdTime,
			Picture: picture,
		}

		fmt.Println(post)

	}

	return nil
}

// see: https://developers.facebook.com/docs/graph-api/webhooks/reference/page/
func (app *App) receiveComment(change facebookgraph.HubEntryChange) *model.AppError {

	rawComment := change.Value
	fmt.Println(rawComment)

	// thêm user nếu cần
	from := <-app.Srv.Store.FacebookUid().UpsertFromMap(rawComment.From)

	userId := rawComment.From["id"].(string)
	postId := rawComment.PostId
	commentId := rawComment.CommentId
	commentTime := time.Unix(rawComment.CreatedTime, 10).Format(time.RFC3339)

	parentId := rawComment.ParentId

	// tat ca comment deu phai co post id, check neu postid nil
	if len(postId) == 0 {
		return model.NewAppError("HandleIncomingWebhook", "web.incoming_webhook.parse.app_error", nil, "", http.StatusBadRequest)
	}

	// postId phai co dang 5465465465_5651651651, chung ta se lay pageId tu postID nay
	s := strings.Split(postId, "_")

	if len(s) < 2 {
		return model.NewAppError("HandleIncomingWebhook", "web.incoming_webhook.parse.app_error", nil, "", http.StatusBadRequest)
	}
	pageId := s[0]

	var snippet string
	if len(rawComment.Message) == 0 {
		if len(rawComment.Attachment.Url) > 0 {
			snippet = "[Attachment]"
		} else if len(rawComment.Photo) > 0 {
			snippet = "[Photo]"
		}
	} else {
		snippet = utils.GetSnippet(rawComment.Message)
	}

	if rawComment.Verb == "add" {
		foundConversation := <-app.Srv.Store.FacebookConversation().InsertConversationFromCommentIfNeed(parentId, commentId, pageId, postId, userId, commentTime, snippet)
		if foundConversation.Data == nil {
			return model.NewAppError("HandleIncomingWebhook", "web.incoming_webhook.parse.app_error", nil, "", http.StatusBadRequest)
		} else {
			conversation := foundConversation.Data.(*model.UpsertConversationResult).Data
			isNew := foundConversation.Data.(*model.UpsertConversationResult).IsNew
			isFromPage := from.Data.(*model.FacebookUid).Id == pageId

			// chỉ cập nhật conversation nếu nó chưa được tạo mới
			if !isNew {
				updatedResult := <-app.Srv.Store.FacebookConversation().UpdateConversation(conversation.Id, snippet, isFromPage, commentTime, 1, commentTime)
				if updatedResult.Err != nil {
					return model.NewAppError("HandleIncomingWebhook", "web.incoming_webhook.parse.app_error", nil, "", http.StatusBadRequest)
				}
			}

			message := &model.FacebookConversationMessage{
				Type: "comment",
				PageId: pageId,
				From: from.Data.(*model.FacebookUid).Id,
				ConversationId: conversation.Id,
				Message: rawComment.Message,
				CreatedTime: commentTime,
				CommentId: rawComment.CommentId,
				CanComment: true,
				CanLike: true,
				CanReply: true,
				CanRemove: true,
				CanHide: true,
				CanReplyPrivately: true,
			}

			// nếu comment có chứa attachment thì:
			if len(rawComment.Message) == 0 || len(rawComment.Photo) > 0 || len(rawComment.Video) > 0 {
				message.HasAttachments = true
			}

			message.PreSave()

			var newMessage *model.FacebookConversationMessage

			if newMessage, _, _ = app.AddMessage(message, true, !isNew, isFromPage); newMessage == nil {
				return model.NewAppError("HandleIncomingWebhook", "web.incoming_webhook.parse.app_error", nil, "", http.StatusBadRequest)
			}

			if newMessage != nil {

				var omitUsers map[string]bool

				if len(newMessage.UserId) > 0 {
					omitUsers = map[string]bool{
						newMessage.UserId:         true,
					}
				}

				webhookData := model.NewWebSocketEvent(model.RECEIVE_CONVERSATION_UPDATED, "", pageId, "", omitUsers)
				webhookData.Add("id", conversation.Id)
				webhookData.Add("conversation", conversation)
				webhookData.Add("newMessage", newMessage)
				app.Publish(webhookData)
			}
		}
	} else if rawComment.Verb == "edit" {
		fmt.Println("edit comment")
	} else if rawComment.Verb == "edited" {
		// check xem comment đã có trong db chưa, đôi khi việc khởi tạo comment không graph được đầy đủ hoặc lỗi
		// phải luôn đảm bảo các cập nhật từ facebook cũng được cập nhật lên server
		// khi chỉnh sửa 1 comment, rất khó để nhận biết comment này chính là snippet của hội thoại.
		// vì thế chúng ta sẽ bỏ qua cập nhật snippet khi chỉnh sửa comment,
		// TODO: cần xem xét cập nhật snippet
		foundConversation := <-app.Srv.Store.FacebookConversation().InsertConversationFromCommentIfNeed(parentId, commentId, pageId, postId, userId, commentTime, snippet)
		if foundConversation.Data == nil {
			return model.NewAppError("HandleIncomingWebhook", "web.incoming_webhook.parse.app_error", nil, "", http.StatusBadRequest)
		} else {
			conversation := foundConversation.Data.(*model.UpsertConversationResult).Data

			// cập nhật comment
			updatedData := <-app.Srv.Store.FacebookConversation().UpdateCommentByCommentId(rawComment.CommentId, rawComment.Message)
			if updatedData.Err != nil {
				return updatedData.Err
			} else {
				updatedMessage := updatedData.Data.(*model.FacebookConversationMessage)
				webhookData := model.NewWebSocketEvent(model.RECEIVE_COMMENT_UPDATED, "", pageId, "", nil)
				webhookData.Add("id", conversation.Id)
				webhookData.Add("conversation", conversation)
				webhookData.Add("newMessage", updatedMessage)
				app.Publish(webhookData)
			}
			return nil
		}
	} else if rawComment.Verb == "delete" {
		fmt.Println("delete comment")
	} else if rawComment.Verb == "remove" {
		fmt.Println("remove comment")
		// cập nhật comment
		updatedData := <-app.Srv.Store.FacebookConversation().DeleteCommentByCommentId(rawComment.CommentId, "")
		if updatedData.Err != nil {
			return updatedData.Err
		} else {
			updatedMessage := updatedData.Data.(*model.FacebookConversationMessage)
			webhookData := model.NewWebSocketEvent(model.RECEIVE_COMMENT_DELETED, "", pageId, "", nil)
			webhookData.Add("deletedComment", updatedMessage)
			app.Publish(webhookData)
		}
		return nil

	} else if rawComment.Verb == "hide" {
		fmt.Println("hide comment")
	} else {
		// unhandle comment
		fmt.Println("received unhandled comment")
	}
	return nil

}

//func (app *App) HandleWebHookFeed(entry *facebookgraph.HubEntryChange, fanpageMember *model.FanpageMember, pageId string) {
//
//	if entry.Value.Item == "comment" {
//		var fbPost *model.FacebookPost
//		postId := entry.Value.PostId
//
//		// tìm kiếm bài viết, nếu chưa có thì save vào db
//		result := <-app.Srv.Store.FacebookPost().Get(postId)
//		if result.Data != nil {
//			fbPost = result.Data.(*model.FacebookPost)
//		} else {
//			graphPost := app.DoGraphOnePost(fanpageMember.AccessToken, postId)
//			fbPost = model.FacebookPostFromPostGraphResult(graphPost)
//			fbPost, _ = app.AddPagePost(fbPost, "")
//		}
//
//		graphComment := app.DoGraphOneComment(fanpageMember.AccessToken, entry.Value.CommentId)
//		if graphComment != nil {
//
//			// cập nhật uid
//			app.Srv.Store.FacebookUid().UpsertFromMap(graphComment.From)
//
//			// TODO: Kiểm tra nếu comment này là subcomment, thêm message vào hội thoại gốc và không tạo hội thoại mới
//			//if len(graphComment.Parent.Id) > 0 {
//			//	fmt.Println("Có hội thoại gốc")
//			//	fmt.Println(graphComment.Parent)
//			//}
//
//			// tìm kiếm conversation
//			result := <-app.Srv.Store.FacebookConversation().GetConversationTypeComment(graphComment.From["id"].(string), pageId, fbPost.Id, graphComment.Id)
//			if result.Data != nil {
//				fbConversation := result.Data.(model.FacebookConversation)
//
//				messageData := model.FacebookConversationMessageModelFromCommentItem(graphComment)
//				messageData.ConversationId = fbConversation.Id
//
//				// thêm message vào db
//				var newMessage *model.FacebookConversationMessage
//				if newMessage, _ = app.AddMessage(messageData); newMessage != nil {
//
//					// thêm attachments
//					if graphComment.Attachment != (facebookgraph.CommentAttachmentImage{}) {
//						image := &model.FacebookAttachmentImage{
//							ConversationType: "comment",
//							MessageId:        newMessage.Id,
//							Url:              graphComment.Attachment.Url,
//							Src:              graphComment.Attachment.Media.Image.Src,
//							Height:           graphComment.Attachment.Media.Image.Height,
//							Width:            graphComment.Attachment.Media.Image.Width,
//						}
//						if _, err := app.AddImage(image); err != nil {
//							fmt.Println(err)
//							return
//						}
//
//						// TODO: cập nhật updated_time cho conversation
//						app.Srv.Store.FacebookConversation().UpdateLatestTime(fbConversation.Id, graphComment.CreatedTime, graphComment.Id)
//					}
//				}
//			}
//		}
//	}
//}

//func (app *App) HandleWebhookMessage(entry *facebookgraph.HubEntry, fanpageMember *model.FanpageMember) {
//	messageId := entry.Messaging[0].Message.Mid
//
//	fmt.Println("======entry========")
//	fmt.Println(entry)
//
//	messageFB := app.DoGraphConversationMessage(fanpageMember.AccessToken, "m_"+messageId)
//	if messageFB != nil {
//		uid := app.GetUidFromMessage(messageFB, entry.Id)
//		result := <-app.Srv.Store.FacebookConversation().GetConversationTypeMessage(uid, entry.Id, messageFB.CreatedTime)
//		if result.Data != nil {
//			fbConversation := result.Data.(*model.FacebookConversation)
//
//			fbConversationMessageData := &model.FacebookConversationMessage{
//				ConversationId: fbConversation.Id,
//				Message:        messageFB.Message,
//				From:           messageFB.From["id"].(string),
//				CreatedTime:    messageFB.CreatedTime,
//			}
//
//			var mr *model.FacebookConversationMessage
//			if mr, _ = app.AddMessage(fbConversationMessageData); mr != nil {
//				for _, im := range messageFB.Attachments.Data {
//					image := &model.FacebookAttachmentImage{
//						ConversationType: "message",
//						MessageId:        mr.Id,
//						Height:           im.ImageData.Height,
//						Width:            im.ImageData.Width,
//						Src:              im.ImageData.Src,
//						Url:              im.ImageData.Url,
//						PreviewUrl:       im.ImageData.PreviewUrl,
//						ImageType:        im.ImageData.ImageType,
//						RenderAsSticker:  im.ImageData.RenderAsSticker,
//					}
//
//					// thêm image vào db
//					if _, err := app.AddImage(image); err != nil {
//						fmt.Println(err)
//					}
//				}
//
//				// TODO: cập nhật updated_time và snippet tại đây
//				app.Srv.Store.FacebookConversation().UpdateLatestTime(fbConversation.Id, mr.CreatedTime, "")
//			}
//		}
//	}
//}

func (app *App) GetUidFromMessage(message *facebookgraph.FacebookMessageItem, pageId string) string {
	if message.From["id"].(string) != pageId {
		return message.From["id"].(string)
	} else {
		return message.To.Data[0]["id"].(string)
	}
}

func (app *App) HandleIncomingWebhook(hookId string, req *model.IncomingWebhookRequest) *model.AppError {
	if !*app.Config().ServiceSettings.EnableIncomingWebhooks {
		//return model.NewAppError("HandleIncomingWebhook", "web.incoming_webhook.disabled.app_error", nil, "", http.StatusNotImplemented)
		return model.NewAppError("HandleIncomingWebhook", "web.incoming_webhook.disabled.app_error", nil, "", http.StatusNotImplemented)
	}

	//hchan := a.Srv.Store.Webhook().GetIncoming(hookId, true)

	if req == nil {
		return model.NewAppError("HandleIncomingWebhook", "web.incoming_webhook.parse.app_error", nil, "", http.StatusBadRequest)
	}

	text := req.Text
	if len(text) == 0 && req.Attachments == nil {
		return model.NewAppError("HandleIncomingWebhook", "web.incoming_webhook.text.app_error", nil, "", http.StatusBadRequest)
	}

	// just test
	return model.NewAppError("HandleIncomingWebhook", "web.incoming_webhook.disabled.app_error", nil, "", http.StatusNotImplemented)

	//channelName := req.ChannelName
	//webhookType := req.Type
	//
	//var hook *model.IncomingWebhook
	//if result := <-hchan; result.Err != nil {
	//	return model.NewAppError("HandleIncomingWebhook", "web.incoming_webhook.invalid.app_error", nil, "err="+result.Err.Message, http.StatusBadRequest)
	//} else {
	//	hook = result.Data.(*model.IncomingWebhook)
	//}
	//
	//uchan := a.Srv.Store.User().Get(hook.UserId)
	//
	//if len(req.Props) == 0 {
	//	req.Props = make(model.StringInterface)
	//}
	//
	//req.Props["webhook_display_name"] = hook.DisplayName
	//
	//text = a.ProcessSlackText(text)
	//req.Attachments = a.ProcessSlackAttachments(req.Attachments)
	//// attachments is in here for slack compatibility
	//if len(req.Attachments) > 0 {
	//	req.Props["attachments"] = req.Attachments
	//	webhookType = model.POST_SLACK_ATTACHMENT
	//}
	//
	//var channel *model.Channel
	//var cchan store.StoreChannel
	//
	//if len(channelName) != 0 {
	//	if channelName[0] == '@' {
	//		if result := <-a.Srv.Store.User().GetByUsername(channelName[1:]); result.Err != nil {
	//			return model.NewAppError("HandleIncomingWebhook", "web.incoming_webhook.user.app_error", nil, "err="+result.Err.Message, http.StatusBadRequest)
	//		} else {
	//			if ch, err := a.GetDirectChannel(hook.UserId, result.Data.(*model.User).Id); err != nil {
	//				return err
	//			} else {
	//				channel = ch
	//			}
	//		}
	//	} else if channelName[0] == '#' {
	//		cchan = a.Srv.Store.Channel().GetByName(hook.TeamId, channelName[1:], true)
	//	} else {
	//		cchan = a.Srv.Store.Channel().GetByName(hook.TeamId, channelName, true)
	//	}
	//} else {
	//	cchan = a.Srv.Store.Channel().Get(hook.ChannelId, true)
	//}
	//
	//if channel == nil {
	//	result := <-cchan
	//	if result.Err != nil {
	//		return model.NewAppError("HandleIncomingWebhook", "web.incoming_webhook.channel.app_error", nil, "err="+result.Err.Message, result.Err.StatusCode)
	//	} else {
	//		channel = result.Data.(*model.Channel)
	//	}
	//}
	//
	//if hook.ChannelLocked && hook.ChannelId != channel.Id {
	//	return model.NewAppError("HandleIncomingWebhook", "web.incoming_webhook.channel_locked.app_error", nil, "", http.StatusForbidden)
	//}
	//
	//var user *model.User
	//if result := <-uchan; result.Err != nil {
	//	return model.NewAppError("HandleIncomingWebhook", "web.incoming_webhook.user.app_error", nil, "err="+result.Err.Message, http.StatusForbidden)
	//} else {
	//	user = result.Data.(*model.User)
	//}
	//
	//if a.License() != nil && *a.Config().TeamSettings.ExperimentalTownSquareIsReadOnly &&
	//	channel.Name == model.DEFAULT_CHANNEL && !a.RolesGrantPermission(user.GetRoles(), model.PERMISSION_MANAGE_SYSTEM.Id) {
	//	return model.NewAppError("HandleIncomingWebhook", "api.post.create_post.town_square_read_only", nil, "", http.StatusForbidden)
	//}
	//
	//if channel.Type != model.CHANNEL_OPEN && !a.HasPermissionToChannel(hook.UserId, channel.Id, model.PERMISSION_READ_CHANNEL) {
	//	return model.NewAppError("HandleIncomingWebhook", "web.incoming_webhook.permissions.app_error", nil, "", http.StatusForbidden)
	//}
	//
	//overrideUsername := hook.Username
	//if req.Username != "" {
	//	overrideUsername = req.Username
	//}
	//
	//overrideIconUrl := hook.IconURL
	//if req.IconURL != "" {
	//	overrideIconUrl = req.IconURL
	//}
	//
	//_, err := a.CreateWebhookPost(hook.UserId, channel, text, overrideUsername, overrideIconUrl, req.Props, webhookType, "")
	//return err
}

func (app *App) HandleCommandWebhook(hookId string, response *model.CommandResponse) *model.AppError {
	// just test
	return model.NewAppError("HandleCommandWebhook", "web.command_webhook.parse.app_error", nil, "", http.StatusBadRequest)

	//if response == nil {
	//	return model.NewAppError("HandleCommandWebhook", "web.command_webhook.parse.app_error", nil, "", http.StatusBadRequest)
	//}

	//var hook *model.CommandWebhook
	//if result := <-a.Srv.Store.CommandWebhook().Get(hookId); result.Err != nil {
	//	return model.NewAppError("HandleCommandWebhook", "web.command_webhook.invalid.app_error", nil, "err="+result.Err.Message, result.Err.StatusCode)
	//} else {
	//	hook = result.Data.(*model.CommandWebhook)
	//}
	//
	//var cmd *model.Command
	//if result := <-a.Srv.Store.Command().Get(hook.CommandId); result.Err != nil {
	//	return model.NewAppError("HandleCommandWebhook", "web.command_webhook.command.app_error", nil, "err="+result.Err.Message, http.StatusBadRequest)
	//} else {
	//	cmd = result.Data.(*model.Command)
	//}
	//
	//args := &model.CommandArgs{
	//	UserId:    hook.UserId,
	//	ChannelId: hook.ChannelId,
	//	TeamId:    cmd.TeamId,
	//	RootId:    hook.RootId,
	//	ParentId:  hook.ParentId,
	//}
	//
	//if result := <-a.Srv.Store.CommandWebhook().TryUse(hook.Id, 5); result.Err != nil {
	//	return model.NewAppError("HandleCommandWebhook", "web.command_webhook.invalid.app_error", nil, "err="+result.Err.Message, result.Err.StatusCode)
	//}
	//
	//_, err := a.HandleCommandResponse(cmd, args, response, false)
	//return err
}
