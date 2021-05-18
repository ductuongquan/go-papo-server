// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package model

import (
	"encoding/json"
	"io"
)

type ConversationTag struct {
	Id 							string 				`json:"id"`
	ConversationId 				string 				`json:"conversation_id"`
	TagId 						string				`json:"tag_id"`
	Creator 					string 				`json:"creator"`
	CreateAt 					int64 				`json:"create_at"`
}

func (p *ConversationTag) PreSave() {
	if p.Id == "" {
		p.Id = NewId()
	}

	p.CreateAt = GetMillis()
}

func (p *ConversationTag) ToJson() string {
	b, _ := json.Marshal(p)
	return string(b)
}

func ConversationTagFromJson(data io.Reader) *ConversationTag {
	var p *ConversationTag
	json.NewDecoder(data).Decode(&p)
	return p
}

func ConversationTagsToJson(p []*ConversationTag) string {
	b, _ := json.Marshal(p)
	return string(b)
}