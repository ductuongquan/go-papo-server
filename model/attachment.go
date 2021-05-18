package model

import (
	"bitbucket.org/enesyteam/papo-server/facebook_graph"
	"encoding/json"
	"io"
)

const (
	COMMENT_ATTACHMENT_CACHE_SIZE             = 50000
)

type TargetImage struct {
	Height 		int 		`json:"height"`
	Width 		int 		`json:"width"`
	Source 		string 		`json:"source"`
}

type TargetItem struct {
	Id 			string 						`json:"id"`
	Type 		string 						`json:"type"`
	Height 		int 						`json:"height,omitempty"`
	Width 		int 						`json:"width,omitempty"`
	Images 		[]TargetImage 				`json:"images,omitempty"`
	Picture 	string 						`json:"picture,omitempty"`
	CanDelete 	bool 						`json:"can_delete,omitempty"`
	CreatedTime string 						`json:"created_time,omitempty"`
	Link 		string 						`json:"link,omitempty"`
	AltText 	string 						`json:"alt_text,omitempty"`
	Format 		[]VideoTargetFormat 		`json:"format,omitempty"` // video
}

type VideoTargetFormat struct {
	EmbedHtml	string 		`json:"embed_html"`
	Filter		string 		`json:"filter"`
	Height 		int 		`json:"height"`
	Width 		int 		`json:"width"`
	Picture		string 		`json:"picture"`
}

type MessageAttachmentsResponse struct {
	Attachments facebookgraph.FacebookMessageItemAttachments 	`json:"attachments"`
	Paging 		facebookgraph.FacebookPaging        			`json:"paging"`
}

func MessageAttachmentsResponseFromJson(data io.Reader) *MessageAttachmentsResponse {
	var o *MessageAttachmentsResponse
	json.NewDecoder(data).Decode(&o)
	return o
}

func (p *MessageAttachmentsResponse) ToJson() string {
	b, _ := json.Marshal(p)
	return string(b)
}

type CommentAttachmentsResponse struct {
	Attachment facebookgraph.CommentAttachmentImage 	`json:"attachment"`
}

func CommentAttachmentsResponseFromJson(data io.Reader) *CommentAttachmentsResponse {
	var o *CommentAttachmentsResponse
	json.NewDecoder(data).Decode(&o)
	return o
}

func (p *CommentAttachmentsResponse) ToJson() string {
	b, _ := json.Marshal(p)
	return string(b)
}

func (p *TargetItem) ToJson() string {
	b, _ := json.Marshal(p)
	return string(b)
}

func TargetItemFromJson(data io.Reader) *TargetItem {
	var o *TargetItem
	json.NewDecoder(data).Decode(&o)
	return o
}

func TargetItemListToJson(p []*TargetItem) string {
	b, _ := json.Marshal(p)
	return string(b)
}
