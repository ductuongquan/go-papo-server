// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package facebookgraph

import (
	"encoding/json"
	"io"
	"io/ioutil"
)

type FacebookConversationMessages struct {
	Data 							[]FacebookMessageItem	`json:"data"`
	Paging 							FacebookPaging 				`json:"paging"`
}

type FacebookConversationItem struct {
	CanReply						bool 					`json:"can_reply"`
	Snippet							string 					`json:"snippet"`
	UpdatedTime						string 					`json:"updated_time"`
	Messages						FacebookConversationMessages 	`json:"messages"`
	Attachments 					FacebookMessageItemAttachments `json:"attachments,omitempty"`
	ScopedThreadKey 				string 					`json:"scoped_thread_key,omitempty"` // đối với tin nhắn t_5165156
	CommentId 						string 					`json:"comment_id"` // 51651651651_51651651 với comment
	From 							map[string]interface{} 	`json:"from"`
	PrivateReplyConversation 		map[string]interface{} 	`json:"private_reply_conversation,omitempty"`
	MessageTags 					[]map[string]interface{} 	`json:"message_tags"` // với comment
	Senders 						ConversationSenders 	`json:"senders"`
}

type ConversationSenders struct {
	Data 							[]FacebookUser 			`json:"data"`
}

type FacebookConversations struct {
	Data 							[]FacebookConversationItem `json:"data"`
	Paging 							FacebookPaging 				`json:"paging"`
}

func FacebookConversationsFromJson(data io.Reader) *FacebookConversations {
	var fbps *FacebookConversations
	x, _ := ioutil.ReadAll(data)
	json.Unmarshal([]byte(x), &fbps)
	return fbps
}

func FacebookConversationsToJson(p *FacebookConversations) string {
	b, _ := json.Marshal(p)
	return string(b)
}

func FacebookUserToJson(u *FacebookUser) string {
	b, _ := json.Marshal(u)
	return string(b)
}

func FacebookConversationMessagesFromJson(data io.Reader) *FacebookConversationMessages {
	var fbps *FacebookConversationMessages
	x, _ := ioutil.ReadAll(data)
	json.Unmarshal([]byte(x), &fbps)
	return fbps
}

func FacebookConversationMessagesToJson(p *FacebookConversationMessages) string {
	b, _ := json.Marshal(p)
	return string(b)
}