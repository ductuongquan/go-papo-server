// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package facebookgraph

import (
	"encoding/json"
	"io"
	"io/ioutil"
)

type FacebookCommentItem struct {
	CanComment               bool                     `json:"can_comment"`
	CommentId 				 string 				  `json:"comment_id"`
	CanHide                  bool                     `json:"can_hide"`
	CanLike                  bool                     `json:"can_like"`
	CanRemove                bool                     `json:"can_remove"`
	CanReplyPrivately        bool                     `json:"can_reply_privately"`
	CanReply                 bool                     `json:"can_reply"`
	CreatedTime              string                   `json:"created_time"`
	Id                       string                   `json:"id"`
	IsHidden                 bool                     `json:"is_hidden"`
	IsPrivate                bool                     `json:"is_private"`
	Message                  string                   `json:"message"`
	MessageTags              []map[string]interface{} `json:"message_tags"`
	PermalinkUrl             string                   `json:"permalink_url"`
	From                     map[string]interface{}   `json:"from"`
	PrivateReplyConversation map[string]interface{}   `json:"private_reply_conversation"`
	Comments                 SubComments              `json:"Comments"`
	Attachment               CommentAttachmentImage   `json:"attachment"`
	Parent 					ParentItem 				  `json:"parent"`
}

type SubComments struct {
	Data   		[]FacebookCommentItem 		`json:"data"`
	Paging 		FacebookPaging        		`json:"paging"`
	Parent 		ParentItem 					`json:"parent"`
}

type ParentItem struct {
	CreatedTime              string                   `json:"created_time"`
	From 					 FacebookUser 			  `json:"from"`
	Message 				 string 				  `json:"message"`
	Id                       string                   `json:"id"`
}

type ImageMedia struct {
	Image 		FacebookImageData 		`json:"image"`
}

type CommentAttachmentImage struct {
	Media 		ImageMedia 		`json:"media"`
	Target 		TargetItem 		`json:"target"`
	Type  		string     		`json:"type"`
	Url   		string     		`json:"url"`
}

func CommentAttachmentFromJson(data io.Reader) *CommentAttachmentImage {
	var o *CommentAttachmentImage
	json.NewDecoder(data).Decode(&o)
	return o
}

func (p *CommentAttachmentImage) ToJson() string {
	b, _ := json.Marshal(p)
	return string(b)
}

type TargetItem struct {
	Id 		string 		`json:"id"`
	Url 	string 		`json:"url"`
}

type FacebookCommentTag struct {
	Id     	string 		`json:"id"`
	Length 	int    		`json:"length"` // độ dài tag
	Name   	string 		`json:"name"`   // tên được tag
	Offset 	int    		`json:"offset"` // vị trí từ 0 đến điểm bắt đầu xuất hiện tag
	Type   	string 		`json:"type"`   // "page" "person"
}

type FacebookPrivateReplyConversation struct {
	Id          string `json:"id"`
	Link        string `json:"link"`
	UpdatedTime string `json:"updated_time"`
}

type FacebookGraphPostCommentsResponseData struct {
	Data   []FacebookCommentItem `json:"data"`
	Paging FacebookPaging        `json:"paging"`
}

type FacebookGraphPostCommentsResponse struct {
	Comments FacebookGraphPostCommentsResponseData `json:"comments"`
}

type FacebookReplyCommentResponse struct {
	Id 					string 		`json:"id"`
	PendingMessageId 	string 		`json:"pending_message_id"` // chỉ dùng khi gửi về client
	ConversationId 		string 		`json:"conversation_id"` // chỉ dùng khi gửi về client
	MessageId			string 		`json:"message_id"`
}

func FacebookReplyCommentResponseToJson(t *FacebookReplyCommentResponse) string {
	b, _ := json.Marshal(t)
	return string(b)
}

func FacebookGraphPostCommentsResponseFromJson(data io.Reader) *FacebookGraphPostCommentsResponse {
	var fbps *FacebookGraphPostCommentsResponse
	x, _ := ioutil.ReadAll(data)
	json.Unmarshal([]byte(x), &fbps)
	return fbps
}

// response từ page thứ 2 trở đi có chút khác biệt
func FacebookGraphPostCommentsResponseNextPageFromJson(data io.Reader) *FacebookGraphPostCommentsResponseData {
	var fbps *FacebookGraphPostCommentsResponseData
	x, _ := ioutil.ReadAll(data)
	json.Unmarshal([]byte(x), &fbps)
	return fbps
}

func FacebookCommentItemFromJson(data io.Reader) *FacebookCommentItem {
	var u *FacebookCommentItem
	json.NewDecoder(data).Decode(&u)
	return u
}
