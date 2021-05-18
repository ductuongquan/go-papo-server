// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package model

import (
	"encoding/json"
	"io"
)

type ConversationNote struct {
	Id 						string 			`json:"id"`
	ConversationId 			string 			`json:"conversation_id"`
	Message 				string 			`json:"message"`
	Creator 				string 			`json:"creator"`
	CreateAt 				int64 			`json:"create_at"`
	UpdateAt 				int64 			`json:"update_at"`
	DeleteAt 				int64 			`json:"delete_at"`
	DeleteBy 				string 			`json:"delete_by"`
	IsPrivate 				bool 			`json:"is_private"`
}

func (p *ConversationNote) PreSave() {
	if p.Id == "" {
		p.Id = NewId()
	}
	p.CreateAt = GetMillis()
	p.UpdateAt = p.CreateAt
	p.DeleteAt = 0
}

func (o *ConversationNote) PreUpdate() {
	o.UpdateAt = GetMillis()
}

func (p *ConversationNote) ToJson() string {
	b, _ := json.Marshal(p)
	return string(b)
}

func ConversationNoteFromJson(data io.Reader) *ConversationNote {
	var p *ConversationNote
	json.NewDecoder(data).Decode(&p)
	return p
}

func ConversationNotesToJson(p []*ConversationNote) string {
	b, _ := json.Marshal(p)
	return string(b)
}