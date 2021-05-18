package model

type Conversation struct {
	Id                       string                 `json:"id"`
	PageId                   string                 `json:"page_id"` // foreign key
	Type                     string                 `json:"type"`
	From                     string                 `json:"from"` // foreign key
	Snippet                  string                 `json:"snippet"`
	UpdatedTime              string                 `json:"updated_time"`
	PostId                   string                 `json:"post_id"`
	CommentId                string                 `json:"comment_id"`   // type = comment
	ThreadId                 string                 `json:"thread_id"`    // type = message
	Seen                     bool                   `json:"seen"`         // true nếu hội thợi đã đọc
	LastSeenBy               string                 `json:"last_seen_by"` // Ghi nhớ hội thoại được xem lần cuối bởi ai
	UnreadCount              int                    `json:"unread_count"` // ghi nhớ số tin nhắn chưa đọc
	CreateAt                 int64                  `json:"create_at"`
	UpdateAt                 int64                  `json:"update_at"`
	DeleteAt                 int64                  `json:"delete_at"`
	CanComment               bool                   `json:"can_comment"`         // chỉ có ở comment
	CanHide                  bool                   `json:"can_hide"`            // chỉ có ở comment
	CanLike                  bool                   `json:"can_like"`            // chỉ có ở comment
	CanReply                 bool                   `json:"can_reply"`           // chỉ có ở comment
	CanRemove                bool                   `json:"can_remove"`          // chỉ có ở comment
	CanReplyPrivately        bool                   `json:"can_reply_privately"` // chỉ có ở comment
	CreatedTime              string                 `json:"created_time"`        // chỉ có ở comment
	IsHidden                 bool                   `json:"is_hidden"`           // chỉ có ở comment
	IsPrivate                bool                   `json:"is_private"`          // chỉ có ở comment
	Message                  string                 `json:"message"`             // chỉ có ở comment, tương đương snippet của message
	PermalinkUrl             string                 `json:"permalink_url"`       // chỉ có ở comment
	PrivateReplyConversation map[string]interface{} `json:"private_reply_conversation,omitempty"`
}
