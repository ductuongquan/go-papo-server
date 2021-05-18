// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package model

import (
	"bitbucket.org/enesyteam/papo-server/facebook_graph"
	"encoding/json"
	"io"
)

type FacebookPost struct {
	Id           string                 `json:"id"`
	CreatedTime  string                 `json:"created_time"`
	From         map[string]interface{} `json:"from"`
	Picture      string                 `json:"picture"`
	UpdatedTime  string                 `json:"updated_time"`
	IsHidden     bool                   `json:"is_hidden"`
	Story        string                 `json:"story"`
	PermalinkUrl string                 `json:"permalink_url"`
	Message      string                 `json:"message"`
	PostId       string                 `json:"post_id"`
	PageId       string                 `json:"page_id"`
}

func FacebookPostFromPostGraphResult(p *facebookgraph.FacebookGraphPost) *FacebookPost {
	return &FacebookPost{
		PostId:       p.Id,
		CreatedTime:  p.CreatedTime,
		From:         p.From,
		Picture:      p.Picture,
		UpdatedTime:  p.UpdatedTime,
		IsHidden:     p.IsHidden,
		Story:        p.Story,
		PermalinkUrl: p.PermalinkUrl,
		Message:      p.Message,
	}
}

type FacebookPostResponse struct {
	Data 					map[string]interface{} 		`json:"data"`
	Attachments 					json.RawMessage 			`json:"attachments"`
}
func FacebookPostResponseListToJson(p []*FacebookPostResponse) string {
	b, _ := json.Marshal(p)
	return string(b)
}


func (p *FacebookPost) PreSave() {
	if p.Id == "" {
		p.Id = NewId()
	}
}

func FacebookPostToJson(p []*FacebookConversationMessage) string {
	b, _ := json.Marshal(p)
	return string(b)
}

func FacebookPostFromJson(data io.Reader) *FacebookPost {
	var p *FacebookPost
	json.NewDecoder(data).Decode(&p)

	p.PostId = p.Id
	p.Id = ""

	return p
}

func FacebookPostsToJson(p []*FacebookPost) string {
	b, _ := json.Marshal(p)
	return string(b)
}
