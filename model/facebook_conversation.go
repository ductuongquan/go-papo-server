// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package model

import (
	"bitbucket.org/enesyteam/papo-server/facebook_graph"
	"encoding/json"
	"io"
)

// Đơn vị của hội thoại, type này dùng chung cho cả conversations và comments
// mỗi hội thoại gồm một hay nhiều FacebookConversationMessage tạo nên
type FacebookConversationMessage struct {
	Id                       string                 `json:"id"`
	EditAt                 	 int64                  `json:"edit_at,omitempty"`
	EditBy					 string 				`json:"edit_by,omitempty"`
	DeleteAt                 int64                 	`json:"delete_at,omitempty"`
	DeleteBy				 string 				`json:"delete_by,omitempty"`
	Type                     string                 `json:"type"` // message Or comment
	HasAttachments 			 bool 					`json:"has_attachments,omitempty"` // Chỉ có ở message
	AttachmentsCount 		 int 					`json:"attachments_count,omitempty"` // chỉ có ở message
	AttachmentType 		 	 string 				`json:"attachment_type,omitempty"` // chỉ có ở message
	Sticker 				 string 				`json:"sticker,omitempty"` // chỉ có ở message
	CreatedTime              string                 `json:"created_time"`
	From                     string                 `json:"from"`
	PageId 					 string 				`json:"page_id"`
	Message                  string                 `json:"message,omitempty"`
	ConversationId           string                 `json:"conversation_id"`               // message này thuộc về conversation nào
	CommentId                string                 `json:"comment_id,omitempty"`        // chỉ có ở comment 1351651651_165654
	MessageId 				 string 				`json:"message_id,omitempty"` // chỉ có ở message m_xdfsdfsdfsdfsdff
	CanComment               bool                   `json:"can_comment,omitempty"`         // chỉ có ở comment
	CanHide                  bool                   `json:"can_hide,omitempty"`            // chỉ có ở comment
	CanLike                  bool                   `json:"can_like,omitempty"`            // chỉ có ở comment
	CanReply                 bool                   `json:"can_reply,omitempty"`           // chỉ có ở comment
	CanRemove                bool                   `json:"can_remove,omitempty"`          // chỉ có ở comment
	CanReplyPrivately        bool                   `json:"can_reply_privately,omitempty"` // chỉ có ở comment
	PrivateReplyConversation map[string]interface{} `json:"private_reply_conversation,omitempty"`
	//MessageTags 		 	[]StringMap 			`json:"message_tags,omitempty"` // với comment
	IsHidden  				bool 					`json:"is_hidden,omitempty"`  // chỉ có ở comment
	IsPrivate 				bool 					`json:"is_private,omitempty"` // chỉ có ở comment
	FileIds       			StringArray     		`json:"file_ids,omitempty"`
	AttachmentTargetIds     StringArray  			`json:"attachment_target_ids,omitempty"` // chỉ có ở comment
	AttachmentIds     		StringArray  			`json:"attachment_ids,omitempty"` // chỉ có ở message
	IsPinned   				bool   					`json:"is_pinned,omitempty"`
	UserId     				string 					`json:"user_id,omitempty"`
	Sent 					bool 					`json:"sent,omitempty"`
	Delivered 				bool 					`json:"delivered,omitempty"`
}

type PostImage struct {
	Width  int `json:"width"`
	Height int `json:"height"`

	// Format is the name of the image format as used by image/go such as "png", "gif", or "jpeg".
	Format string `json:"format"`

	// FrameCount stores the number of frames in this image, if it is an animated gif. It will be 0 for other formats.
	FrameCount int `json:"frame_count"`
}

func (o *PostImage) ToJson() string {
	b, _ := json.Marshal(o)
	return string(b)
}

// Deprecated, do not use this field any more
type FacebookAttachmentImage struct {
	Id               string `json:"id"`
	PostId           string `json:"post_id,omitempty"`           // nếu attachment trong post
	ConversationType string `json:"conversation_type"` // "message" OR "comment"
	MessageId        string `json:"message_id,omitempty"`
	Height           int    `json:"height"`
	Width            int    `json:"width"`
	Src              string `json:"src,omitempty"`
	Url              string `json:"url,omitempty"`               // không có ở comment attachment
	PreviewUrl       string `json:"preview_url,omitempty"`       // không có ở comment attachment
	ImageType        int    `json:"image_type,omitempty"`        // không có ở comment attachment
	RenderAsSticker  bool   `json:"render_as_sticker,omitempty"` // không có ở comment attachment
	TargetId 		 string `json:"target_id,omitempty"`
	TargetUrl 		 string `json:"target_url,omitempty"`
}

type FacebookConversation struct {
	Id                       string                 `json:"id"` // tự động sinh khi tạo
	PageId                   string                 `json:"page_id"`
	Type                     string                 `json:"type"` // "message" hoặc "comment"
	From                     string                 `json:"from"`
	Snippet                  string                 `json:"snippet"` // cần set max length = 120 ký tự
	UpdatedTime              string                 `json:"updated_time"`
	PostId                   string                 `json:"post_id,omitempty"`           // chỉ có ở comment
	CommentId                string                 `json:"comment_id,omitempty"`        // chỉ có ở comment 1351651651_165654
	PageScopeId              string                 `json:"page_scope_id,omitempty"`        // chỉ có ở message 3328924677161946
	ScopedThreadKey          string                 `json:"scoped_thread_key,omitempty"` // chỉ có ở tin nhắn t_6546541351651
	Seen                     bool                   `json:"seen"`              // true nếu hội thợi đã đọc
	LastSeenBy               string                 `json:"last_seen_by,omitempty"`      // Ghi nhớ hội thoại được xem lần cuối bởi ai
	Replied 				 bool 					`json:"replied"`		   // Cho biết hội thoại đã được trả lời hay chưa
	UnreadCount              int                    `json:"unread_count,omitempty"`      // ghi nhớ số tin nhắn chưa đọc
	CreateAt                 int64                  `json:"create_at"`
	UpdateAt                 int64                  `json:"update_at"`
	DeleteAt                 int64                  `json:"delete_at"`
	CanComment               bool                   `json:"can_comment,omitempty"`         // chỉ có ở comment
	CanHide                  bool                   `json:"can_hide,omitempty"`            // chỉ có ở comment
	CanLike                  bool                   `json:"can_like,omitempty"`            // chỉ có ở comment
	CanReply                 bool                   `json:"can_reply,omitempty"`           // chỉ có ở comment
	CanRemove                bool                   `json:"can_remove,omitempty"`          // chỉ có ở comment
	CanReplyPrivately        bool                   `json:"can_reply_privately,omitempty"` // chỉ có ở comment
	CreatedTime              string                 `json:"created_time"`        // chỉ có ở comment
	IsHidden                 bool                   `json:"is_hidden,omitempty"`           // chỉ có ở comment
	IsPrivate                bool                   `json:"is_private,omitempty"`          // chỉ có ở comment
	Message                  string                 `json:"message"`             // chỉ có ở comment, tương đương snippet của message
	PermalinkUrl             string                 `json:"permalink_url,omitempty"`       // chỉ có ở comment
	PrivateReplyConversation map[string]interface{} `json:"private_reply_conversation,omitempty"`
	LastUserMessageAt 		 string 				`json:"last_user_message_at,omitempty"` 		// Tin nhắn cuối cùng của khách hàng
	TagIds       			StringArray     		`json:"tag_ids,omitempty"`
	NoteIds       			StringArray     		`json:"note_ids,omitempty"`
	ReadWatermark 			int64 					`json:"read_watermark,omitempty"` // chỉ có ở message, cho biết người dùng đã đọc tất cả tin nhắn từ thời điểm này về trước
}

type UpsertConversationResult struct {
	IsNew 					bool 					`json:"is_new"`
	Data 					*FacebookConversation 	`json:"data"`
}

type ConversationResponse struct {
	Data 					map[string]interface{} 		`json:"data"`
	Tags 					json.RawMessage 			`json:"tags"`
	From 					map[string]interface{} 		`json:"from"`
}

type ConversationMessageResponse struct {
	Data 					map[string]interface{} 		`json:"data"`
	Messages 				json.RawMessage 			`json:"messages"`
	Notes 					json.RawMessage 			`json:"notes"`
	Tags 					json.RawMessage 			`json:"tags"`
}

// Tạo conversation từ kết quả của Facebook graph
// FacebookCommentItem là kết quả của graph
func FacebookConversationModelFromGraphCommentItem(g *facebookgraph.FacebookCommentItem) *FacebookConversation {
	return &FacebookConversation{
		Type:                     "comment",
		CanComment:               g.CanComment,
		CanHide:                  g.CanHide,
		CanLike:                  g.CanLike,
		CanReply:                 g.CanReply,
		CanRemove:                g.CanRemove,
		CanReplyPrivately:        g.CanReplyPrivately,
		PrivateReplyConversation: g.PrivateReplyConversation,
		CreatedTime:              g.CreatedTime,
		UpdatedTime:              g.CreatedTime,
		IsHidden:                 g.IsHidden,
		IsPrivate:                g.IsPrivate,
		Message:                  g.Message,
		PermalinkUrl:             g.PermalinkUrl,
		//CommentId:                g.CommentId,
		From:                     g.From["id"].(string),
		//MessageTags: 		g.MessageTags,
	}
}

func FacebookConversationModelFromGraphConversationItem(g *facebookgraph.FacebookConversationItem) *FacebookConversation {
	return &FacebookConversation{
		Type:            "message",
		CanReply:        g.CanReply,
		Snippet:         g.Snippet, // utils.GetSnippet(g.Snippet, 120),
		UpdatedTime:     g.UpdatedTime,
		ScopedThreadKey: g.ScopedThreadKey,
	}
}

func FacebookConversationMessageModelFromCommentItem(g *facebookgraph.FacebookCommentItem) *FacebookConversationMessage {
	return &FacebookConversationMessage{
		Type:                     "comment",
		CanComment:               g.CanComment,
		CanHide:                  g.CanHide,
		CanLike:                  g.CanLike,
		CanReply:                 g.CanReply,
		CanRemove:                g.CanRemove,
		CanReplyPrivately:        g.CanReplyPrivately,
		PrivateReplyConversation: g.PrivateReplyConversation,
		CreatedTime:              g.CreatedTime,
		IsHidden:                 g.IsHidden,
		IsPrivate:                g.IsPrivate,
		Message:                  g.Message,
		From:                     g.From["id"].(string),
		//MessageTags: 		g.MessageTags,
	}
}

type FacebookUser struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	PageId string `json:"page_id"`
}

type ConversationReply struct {
	PageId 				string 		`json:"page_id"`
	Message 			string 		`json:"message"`
	AttachmentUrl 		string 		`json:"attachment_url"`
	CommentId 			string 		`json:"comment_id"`
	ThreadId 			string 		`json:"thread_id"`
	Type 				string 		`json:"type"`
	PendingMessageId 	string 		`json:"pending_message_id"`
	To 					string 		`json:"to"`
	PageScopeId 		string 		`json:"page_scope_id"`
	PageToken 			string 		`json:"page_token"`
	UserToken 			string 		`json:"user_token"`
}

type MessageGraphReply struct {
	MessagingType 		string 		`json:"messaging_type,omitempty"` //https://developers.facebook.com/docs/messenger-platform/send-messages/#messaging_types
	Recipient 			*Recipient 	`json:"recipient"`
	Message 			*Message 	`json:"message"`
}

type Recipient struct {
	Id 					string 		`json:"id"`
}

func (p *Recipient) ToJson() string {
	b, _ := json.Marshal(p)
	return string(b)
}

type Message struct {
	Text 				string 		`json:"text,omitempty"`
	Attachment   		*Attachment 	`json:"attachment,omitempty"`
}

func (p *Message) ToJson() string {
	b, _ := json.Marshal(p)
	return string(b)
}

type Attachment struct {
	Type 			string 			`json:"type,omitempty"` //audio, video, image, file
	Payload 		*Payload 		`json:"payload,omitempty"`
}

type Payload struct {
	Url 			string 			`json:"url,omitempty"`
	IsReusable      bool 			`json:"is_reusable,omitempty"`
	AttachmentId	string 			`json:"attachment_id,omitempty"`
}

func ConversationReplyFromJson(data io.Reader) *ConversationReply {
	var u *ConversationReply
	json.NewDecoder(data).Decode(&u)
	return u
}

func FacebookUserFromJson(data io.Reader) *FacebookUser {
	var u *FacebookUser
	json.NewDecoder(data).Decode(&u)
	return u
}

// Hàm này phải được thực thi trước khi lưu dữ liệu vào database
// các trường CreateAt và UpdateAt sẽ được tự động thêm vào
func (p *FacebookConversationMessage) PreSave() {
	if p.Id == "" {
		p.Id = NewId()
	}
}

func FacebookConversationMessageToJson(p []*FacebookConversationMessage) string {
	b, _ := json.Marshal(p)
	return string(b)
}

// Hàm này phải được thực thi trước khi lưu dữ liệu vào database
// các trường CreateAt và UpdateAt sẽ được tự động thêm vào
func (p *FacebookConversation) PreSave() {
	if p.Id == "" {
		p.Id = NewId()
	}
	p.Seen = false
	p.CreateAt = GetMillis()
	p.UpdateAt = p.CreateAt
}

func (p *FacebookAttachmentImage) PreSave() {
	if p.Id == "" {
		p.Id = NewId()
	}
}

// Convert một Fanpage sang chuỗi json
func (p *FacebookConversation) ToJson() string {
	b, _ := json.Marshal(p)
	return string(b)
}

// Chuyển đổi một danh sách FacebookConversations sang chuỗi json
func FacebookConversationListToJson(p []*FacebookConversation) string {
	b, _ := json.Marshal(p)
	return string(b)
}

func FacebookConversationMessagesToJson(p []*FacebookConversationMessage) string {
	b, _ := json.Marshal(p)
	return string(b)
}

func ConversationResponseListToJson(p []*ConversationResponse) string {
	b, _ := json.Marshal(p)
	return string(b)
}

func ConversationMessageResponseListToJson(p *ConversationMessageResponse) string {
	b, _ := json.Marshal(p)
	return string(b)
}

// decode input và trả về danh sách các đối tượng FacebookConversations
func FacebookConversationListFromJson(data io.Reader) []*FacebookConversation {
	var cvs []*FacebookConversation
	json.NewDecoder(data).Decode(&cvs)
	return cvs
}

func FacebookConversationMessageFromJson(data io.Reader) *FacebookConversationMessage {
	var p *FacebookConversationMessage
	json.NewDecoder(data).Decode(&p)
	return p
}
