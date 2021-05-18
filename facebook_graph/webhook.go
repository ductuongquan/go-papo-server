// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package facebookgraph

type HubEntries struct {
	Entry []HubEntry `json:"entry"`
	Object string 	`json:"object"`
}

type HubEntry struct {
	Id        string              `json:"id"`
	Time      int                 `json:"time"`
	Messaging []HubEntryMessaging `json:"messaging"`
	Changes   []HubEntryChange    `json:"changes"`
	UserId 	  string 			  `json:"uid"`
}

/* Message Hook */
type HubEntryMessaging struct {
	Sender    	MessagingSender    					`json:"sender"`
	Recipient 	MessagingRecipient 					`json:"recipient"`
	Message   	MessagingMessage   					`json:"message"`
	Timestamp 	int64 			 					`json:"timestamp"`
	PostBack  	MessagingMessageAttachment    		`json:"post_back"`
	Delivery   	MessageAction 						`json:"delivery"`
	Read		MessageAction						`json:"read"`
}

type MessagingMessage struct {
	Sender    	MessagingSender    					`json:"sender"`
	Recipient 	MessagingRecipient 					`json:"recipient"`
	Mid         string                       		`json:"mid"`
	Text        string                       		`json:"text"`
	StickerId   int64                          		`json:"sticker_id"`
	Attachments []MessagingMessageAttachment 		`json:"attachments"`
	IsEcho 	  	bool 				 				`json:"is_echo"`
	QuickReply  MessagingMessageAttachment  		`json:"quick_reply"`
}

type MessageAction struct {
	Mids 		[]string 							`json:"mids"`
	Watermark  	int64 								`json:"watermark"`
	Seq 		int 								`json:"seq"`
}

type MessagingMessageAttachment struct {
	Type    	string                 				`json:"type"`
	Payload 	map[string]interface{} 				`json:"payload"`
	Url     	string 								`json:"url"`
}

type MessagingSender struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type MessagingRecipient struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

/* Comment Hook */
type HubEntryChange struct {
	Field string              `json:"field"`
	Value HubEntryChangeValue `json:"value"`
	//Value        map[string]interface{} `json:"value"`
}

//type HubEntryChangeValue struct {
//	Id 			string 					`json:"id"`
//	From        map[string]interface{} `json:"from"`
//	Item        string                 `json:"item"`
//	PostId      string                 `json:"post_id"` // bai viet
//	Link 		string 					`json:"link"` // bai viet co 1 anh
//	Published 	int 					`json:"published"` // bai viet
//	PhotoId 	string 					`json:"photo_id"` // bai viet co 1 anh
//	Photos 		[]string 				`json:"photos"` // bai viet co nhieu anh
//	CommentId   string                 `json:"comment_id"`
//	ParentId    string                 `json:"parent_id"`
//	Verb        string                 `json:"verb"`
//	CreatedTime int64                 `json:"created_time"`
//	IsHidden    bool                   `json:"is_hidden"`
//	Message     string                 `json:"message"`
//	PageId 		string 					`json:"page_id"`
//	ThreadId 	string 					`json:"thread_id"`
//	Attachment  CommentAttachmentImage   `json:"attachment"`
//	Photo 		string 					`json:"photo"`
//	EditedTime  string 					`json:"edited_time"`
//	PhotoIds    []string  				`json:"photo_ids"` //The IDs of the photos that were added to an album
//	Video  		string 					`json:"video"` //The link to an attached video
//	Action  	string 					`json:"action"` //action
//	ReactionType string 				`json:"reaction_type"`
//	RecipientId	string 					`json:"recipient_id"`
//	ShareId 	string 					`json:"share_id"`
//}

type HubEntryChangeValue struct {
	EditedTime			string 								`json:"edited_time,omitempty"` // edited_time
	From        		map[string]interface{} 				`json:"from"`	// The sender information
	Post 				map[string]interface{}				`json:"post,omitempty"` // Provide additional content about a post such as 'type' (e.g. photo, video), 'status_type' (e.g. added_photos), 'is_published' (e.g. true/false), 'updated_time', 'permalink_url', 'promotion_status' (e.g. inactive, extendable)
	IsHidden			bool 								`json:"is_hidden,omitempty"` // Whether the post is hidden or not
	Link				string 								`json:"link,omitempty"` // The link to attached content
	Message				string 								`json:"message"` // The message that is part of the content
	Photo				string 								`json:"photo"` // The link to an attached photo
	PhotoIds 			[]string 							`json:"photo_ids"` //The IDs of the photos that were added to an album
	Photos 				[]string 							`json:"photos"` //The links to any attached photos
	PostId				string 								`json:"post_id"` //The post ID
	Story				string 								`json:"story"` //The story—only logged for a milestone item
	Title				string 								`json:"title"` //The title—only logged for a milestone item
	Video 				string 								`json:"video"` //The link to an attached video
	VideoFlagReason		string 								`json:"video_flag_reason"` //The code why the video is flagged. Available when a video is blocked, muted, or unblocked. (Reason Code 1: Your video may contain content that belongs to someone else)
	Action				string 								`json:"action"` //action
	AlbumId				int64 								`json:"album_id"`
	CommentId			string 								`json:"comment_id"`
	CreatedTime			int64 								`json:"created_time,omitempty"`
	EventId				string 								`json:"event_id"`
	Item				string 								`json:"item"`
	OpenGraphStoryId	int64 								`json:"open_graph_story_id"`
	ParentId			string 								`json:"parent_id"`
	PhotoId				int64 								`json:"photo_id"`
	ReactionType		string 								`json:"reaction_type"`
	Published			int32 								`json:"published"`
	RecipientId			int64 								`json:"recipient_id"`
	ShareId				int64								`json:"share_id"`
	Verb				string 								`json:"verb"` // enum {add, block, edit, edited, delete, follow, hide, mute, remove, unblock, unhide, update}
	VideoId				int64 								`json:"video_id"`
	Attachment  		CommentAttachmentImage   			`json:"attachment"`
}
