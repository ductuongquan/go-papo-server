// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package api1

import (
	"bitbucket.org/enesyteam/papo-server/facebook_graph"
	"bitbucket.org/enesyteam/papo-server/model"
	"bitbucket.org/enesyteam/papo-server/utils"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

func (api *API) InitFacebookConversation() {
	api.BaseRoutes.Conversations.Handle("", api.ApiSessionRequired(getConversations)).Methods("GET")
	api.BaseRoutes.Conversations.Handle("", api.ApiSessionRequired(createConversations)).Methods("POST")
	api.BaseRoutes.Conversation.Handle("/{page_id:[A-Za-z0-9]+}/seen", api.ApiSessionRequired(updateSeen)).Methods("PUT")
	api.BaseRoutes.Conversation.Handle("/{page_id:[A-Za-z0-9]+}/unseen", api.ApiSessionRequired(updateUnSeen)).Methods("PUT")

	api.BaseRoutes.Conversation.Handle("", api.ApiSessionRequired(getConversationMessages)).Methods("GET")
	api.BaseRoutes.Conversations.Handle("/search", api.ApiSessionRequired(searchConversations)).Methods("POST")
	api.BaseRoutes.Conversation.Handle("/messages", api.ApiSessionRequired(addConversationMessage)).Methods("POST")

	api.BaseRoutes.Conversation.Handle("/reply", api.ApiSessionRequired(replyConversation)).Methods("POST")
}

type searchInput struct {
	PageIds []string `json:"pageIds"`
	Term 	string 		`json:"term"`
	Limit 	int 		`json:"limit"`
	Offset  int 		`json:"offset"`
}

func replyConversation(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireConversationId()
	if c.Err != nil {
		c.SetInvalidUrlParam("conversation_id")
		return
	}

	message := model.ConversationReplyFromJson(r.Body)
	if len(message.PageId) == 0 {
		c.Err = model.NewAppError("api.reply", "api.reply_conversation.missing_page_token.app_error", nil, "", http.StatusBadRequest)
		return
	}

	if len(message.PageToken) == 0 {
		c.Err = model.NewAppError("api.reply", "api.reply_conversation.missing_page_token.app_error", nil, "", http.StatusBadRequest)
		return
	}

	if len(message.UserToken) == 0 {
		c.Err = model.NewAppError("api.reply", "api.reply_conversation.missing_user_token.app_error", nil, "", http.StatusBadRequest)
		return
	}


	if message.Type == "comment" {
		if len(message.CommentId) > 0 {

			// thêm conversation message và emit về user, sau đó mới post lên facebook,
			// kết quả nhận được từ post facebook có 2 giai đoạn:
			// 1: Sau khi post xong, thành công sẽ trả về comment_id
			// 2: Webhook gửi về 1 kết quả sau khi post thành công
			conversationMessage := &model.FacebookConversationMessage{
				Type: "comment",
				From: message.PageId,
				PageId: message.PageId,
				Message: message.Message,
				ConversationId: c.Params.ConversationId,
				CreatedTime: time.Now().Format("2006-01-02T15:04:05-0700"),
			}

			conversationMessage.PreSave()

			// thêm message vào database
			var er *model.AppError
			var rms *model.FacebookConversationMessage
			if rms, _, er = c.App.AddMessage(conversationMessage, false, false, true); er != nil {
				// TODO: làm gì với lỗi này?

			} else {
				// update conversation snippet
				updateResult := <- c.App.Srv.Store.FacebookConversation().UpdateConversation(c.Params.ConversationId, utils.GetSnippet(message.Message), true, time.Now().Format("2006-01-02T15:04:05-0700"), 0, "")
				if updateResult.Err != nil {
					//mlog.Error(fmt.Sprintf("Couldn't update pages status err=%v", updateResult.Err))
				}
				// emit message to socket
				m := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_ADD_MESSAGE, "", message.PageId, "", nil)
				m.Add("page_id", message.PageId)
				m.Add("message", rms)
				c.App.Publish(m)
			}

			response, fErr, aErr := c.App.ReplyComment(message.CommentId, message)
			if fErr != nil {
				fmt.Println("fErr", fErr.Error)
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(fErr.Error.ToJson()))
				return
			} else if aErr != nil {
				fmt.Println("aErr: ", aErr)
				c.Err = aErr
				return
			} else {
				var resp *facebookgraph.FacebookReplyCommentResponse
				x, _ := ioutil.ReadAll(response)
				json.Unmarshal([]byte(x), &resp)

				// also add pending message id
				resp.PendingMessageId = message.PendingMessageId
				resp.ConversationId = c.Params.ConversationId

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(facebookgraph.FacebookReplyCommentResponseToJson(resp)))
				return
			}
		}
	} else if message.Type == "message" {
		if len(message.ThreadId) > 0 {
			// thêm conversation message và emit về user, sau đó mới post lên facebook,
			// kết quả nhận được từ post facebook có 2 giai đoạn:
			// 1: Sau khi post xong, thành công sẽ trả về comment_id
			// 2: Webhook gửi về 1 kết quả sau khi post thành công
			conversationMessage := &model.FacebookConversationMessage{
				Type: "message",
				From: message.PageId,
				PageId: message.PageId,
				Message: message.Message,
				ConversationId: c.Params.ConversationId,
				CreatedTime:  time.Now().Format("2006-01-02T15:04:05-0700"),
				UserId: c.App.Session.UserId,
			}

			conversationMessage.PreSave()

			var psId string
			// Maybe need get page scope id, then update to conversation
			if len(message.PageScopeId) == 0 {
				// get match Page Scope Id from App Scope Id
				psid, getPsIdErr, getAppErr := c.App.MatchPageScopeId(message.PageId, message.UserToken, message.To)
				if getPsIdErr != nil {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(getPsIdErr.Error.ToJson()))
					return
				}

				if getAppErr != nil {
					c.Err = getAppErr
					return
				}

				psId = psid

				// update conversation page_scope_id for later replying
				updatedErr := c.App.UpdateConversationPageScopeId(c.Params.ConversationId, psid)
				if updatedErr != nil {
					// TODO: Log this err to mlog
				}
			} else {
				psId = message.PageScopeId
			}


			// thêm message vào database
			var er *model.AppError
			//var rms *model.FacebookConversationMessage
			if _, _, er = c.App.AddMessage(conversationMessage, false, false, true); er != nil {
				// TODO: làm gì với lỗi này?

			} else {
				// update conversation snippet
				updateResult := <- c.App.Srv.Store.FacebookConversation().UpdateConversation(c.Params.ConversationId, utils.GetSnippet(message.Message), true, time.Now().Format("2006-01-02T15:04:05-0700"), 0, "")
				if updateResult.Err != nil {
					//mlog.Error(fmt.Sprintf("Couldn't update pages status err=%v", updateResult.Err))
				}
			}

			response, fErr, aErr := c.App.ReplyMessage(message.ThreadId, psId, message)
			if fErr != nil {
				fmt.Println("fErr", fErr.Error)
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(fErr.Error.ToJson()))
				return
			} else if aErr != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(aErr.ToJson()))
				return
			} else {
				var resp *facebookgraph.FacebookReplyCommentResponse
				x, _ := ioutil.ReadAll(response)
				json.Unmarshal([]byte(x), &resp)

				// sau khi post facebook thanh cong, ket qua tra ve gom co message_id
				// chung ta se cap nhat message_id
				conversationMessage.MessageId = resp.MessageId
				conversationMessage.Sent = true
				overWriteMessageResult := <- c.App.Srv.Store.FacebookConversation().OverwriteMessage(conversationMessage)
				if overWriteMessageResult.Err != nil {
					fmt.Println("da co loi xay ra khi cap nhat message id")
				} else {
					var updatedMessage *model.FacebookConversationMessage
					updatedMessage = overWriteMessageResult.Data.(*model.FacebookConversationMessage)
					if updatedMessage != nil {
						fmt.Println("da cap nhat messageId thanh cong: ", updatedMessage)
						// emit message to socket
						m := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_ADD_MESSAGE, "", message.PageId, "", nil)
						m.Add("page_id", message.PageId)
						m.Add("id", updatedMessage.Id)
						m.Add("message", updatedMessage)
						m.Add("pending_message_id", message.PendingMessageId)
						c.App.Publish(m)
					}
				}


				// also add pending message id
				resp.PendingMessageId = message.PendingMessageId
				resp.ConversationId = c.Params.ConversationId

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(facebookgraph.FacebookReplyCommentResponseToJson(resp)))
				return
			}
		}
	} else {
		//fmt.Println("missing type of reply")
		err := model.NewAppError("replyConversation", "api.conversation.reply_conversation.token_not_found.app_error", nil, "", http.StatusUnauthorized)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(err.ToJson()))
	}
}

func addConversationMessage(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireConversationId()
	if c.Err != nil {
		return
	}

	message := model.FacebookConversationMessageFromJson(r.Body)
	if message == nil {
		c.SetInvalidParam("Nội dung tin nhắn")
		return
	}

	if len(message.Type) == 0 {
		c.SetInvalidParam("Kiểu tin nhắn")
		return
	}

	if len(message.From) == 0 {
		c.SetInvalidParam("Tham số người gửi")
		return
	}

	if len(message.PageId) == 0 {
		c.SetInvalidParam("ID trang")
		return
	}

	if len(message.Message) == 0 {
		c.SetInvalidParam("Nội dung tin nhắn")
		return
	}

	if len(message.ConversationId) == 0 {
		c.SetInvalidParam("ID hội thoại")
		return
	}

	fmt.Println(message)
}

func searchConversations(c *Context, w http.ResponseWriter, r *http.Request) {
	var si *searchInput
	x, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal([]byte(x), &si)

	if si == nil {
		c.SetInvalidParam("Nội dung tìm kiếm")
		return
	}

	var inPages string
	if len(si.PageIds) > 0 {
		inPages += "("
		i := 1
		for _, pageId := range si.PageIds {
			inPages += "'" + pageId + "'"
			if i < len(si.PageIds) {
				inPages += ","
			}
			i += 1
		}
		inPages += ")"
	}

	if len(si.Term) == 0 {
		c.SetInvalidParam("Từ khóa tìm kiếm")
		return
	}

	limit := si.Limit
	offset:= si.Offset

	if limit == 0 {
		limit = 30
	}
	cvs, err := c.App.SearchConversations(si.Term, inPages, limit, offset)
	if err != nil {
		c.Err = err
		return
	}
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Write([]byte(model.ConversationResponseListToJson(cvs)))
}

func updateSeen(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireConversationId()
	c.RequirePageId()
	if c.Err != nil {
		return
	}

	_, err := c.App.UpdateSeen(c.Params.ConversationId, c.Params.PageId, c.App.Session.UserId)

	if err != nil {
		c.Err = err
		return
	}

	ReturnStatusOK(w)
}

func updateUnSeen(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireConversationId()
	c.RequirePageId()
	if c.Err != nil {
		return
	}

	result, err := c.App.UpdateUnSeen(c.Params.ConversationId, c.Params.PageId, c.App.Session.UserId)

	if err != nil {
		c.Err = err
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(result.ToJson()))
}

func testGetMessages(c *Context, w http.ResponseWriter, r *http.Request)  {
	//query := r.URL.Query()
	//pageId := query.Get("pageId")
	//pageToken := query.Get("pageToken")
	//
	//if len(pageId) == 0 || len(pageToken) == 0 {
	//	fmt.Println("Thiếu thông tin")
	//	return
	//}
	//
	//err, aerr := c.App.GraphMessagesAndInit(pageId, pageToken)
	//if err != nil {
	//	w.WriteHeader(http.StatusBadRequest)
	//	w.Write([]byte(facebookgraph.FacebookErrorToJson(err)))
	//} else if aerr != nil {
	//	w.WriteHeader(http.StatusBadRequest)
	//	w.Write([]byte(aerr.ToJson()))
	//} else {
	//	w.WriteHeader(http.StatusOK)
	//}

}

func testGetComments(c *Context, w http.ResponseWriter, r *http.Request)  {
	//query := r.URL.Query()
	//pageId := query.Get("pageId")
	//pageToken := query.Get("pageToken")
	//postId := query.Get("postId")
	//
	//if len(pageId) == 0 || len(pageToken) == 0 || len(postId) == 0 {
	//	fmt.Println("Thiếu thông tin")
	//	return
	//}
	//
	//err, aerr := c.App.InitConversationsFromPost(postId, pageId, pageToken)
	//if err != nil {
	//	w.WriteHeader(http.StatusBadRequest)
	//	w.Write([]byte(facebookgraph.FacebookErrorToJson(err)))
	//} else if aerr != nil {
	//	w.WriteHeader(http.StatusBadRequest)
	//	w.Write([]byte(aerr.ToJson()))
	//} else {
	//	w.WriteHeader(http.StatusOK)
	//}
}

func createConversations(c *Context, w http.ResponseWriter, r *http.Request) {

}

func getConversations(c *Context, w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	limit, _ := strconv.Atoi(query.Get("limit"))
	offset, _ := strconv.Atoi(query.Get("offset"))
	pageIds := query.Get("pageIds")
	if limit == 0 {
		limit = 30
	}
	cvs, err := c.App.GetConversations(pageIds, offset, limit)
	if err != nil {
		c.Err = err
		return
	}
	w.Write([]byte(model.FacebookConversationListToJson(cvs)))
}

func getConversationMessages(c *Context, w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	limit, _ := strconv.Atoi(query.Get("limit"))
	offset, _ := strconv.Atoi(query.Get("offset"))
	if limit == 0 {
		limit = 20
	}
	//GetMessagesByConversationId
	//fmt.Println("called me ", limit, offset)
	conversationId := c.Params.ConversationId

	if len(conversationId) == 0 {
		fmt.Println("Thiếu conversation id")
		return
	}

	messages, err := c.App.GetConversationMessages(conversationId, offset, limit)
	if err != nil {
		c.Err = err
		return
	}
	//c.App.SanitizeMessages(c.App.Session, messages)
	w.Write([]byte(model.FacebookConversationMessagesToJson(messages)))
}
