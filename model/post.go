package model

import (
	"bitbucket.org/enesyteam/papo-server/facebook_graph"
	"encoding/json"
	"io"
)

const (
	POST_CACHE_SIZE             = 50000
)

type Post struct {
	Id 					string 			`json:"id"`
	Picture				string 			`json:"picture"`
	Message 			string 			`json:"message"`
	Attachments 		PostAttachment 	`json:"attachments"`
	Likes 				PostLike 		`json:"likes"`
	CreatedTime			string 			`json:"created_time"`
	Comments 			PostComments 	`json:"comments"`
	PermalinkUrl		string 			`json:"permalink_url"`
	Story 				string 			`json:"story"`
}

type PostComments struct {
	//Data
	//Paging
	Summary 			CommentSummary 	`json:"summary"`
}

type CommentSummary struct {
	Order 				string 			`json:"order"`
	TotalCount 			int 			`json:"total_count"`
	CanComment 			bool 			`json:"can_comment"`
}

type PostLike struct {
	//Data
	//Paging
	Summary 			PostLikeSummary 	`json:"summary"`
}

type PostLikeSummary struct {
	TotalCount 			int 			`json:"total_count"`
	CanLike				bool 			`json:"can_like"`
	HasLiked			bool 			`json:"has_liked"`
}

type PostAttachment struct {
	Data 				[]AttachmentItem 	`json:"data"`
}

type AttachmentItem struct {
	SubAttachments 				SubAttachmentItem 	`json:"subattachments"`
}

type SubAttachmentItem struct {
	Data 				[]facebookgraph.CommentAttachmentImage 		`json:"data"`
}

func PostFromJson(data io.Reader) *Post {
	var p *Post
	json.NewDecoder(data).Decode(&p)
	return p
}

func (p *Post) ToJson() string {
	b, _ := json.Marshal(p)
	return string(b)
}