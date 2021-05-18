// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package model

import (
	"encoding/json"
	"io"
)

type ReplySnippet struct {
	Id       					string        	`json:"id"`
	PageId   					string        	`json:"page_id"`
	Trigger  					string 			`json:"trigger"`
	AutoCompleteDesc			string 			`json:"auto_complete_desc"`
	AutoComplete 				bool 			`json:"auto_complete"`
	Attachments 				string 			`json:"attachments"`
	CreateAt 					int64         	`json:"create_at"`
	UpdateAt 					int64         	`json:"update_at"`
	DeleteAt 					int64         	`json:"delete_at"`
	Visible  					bool          	`json:"visible"`
}

func (p *ReplySnippet) PreSave() {
	if p.Id == "" {
		p.Id = NewId()
	}
	p.Visible = true
	p.AutoComplete = true
	p.CreateAt = GetMillis()
	p.UpdateAt = p.CreateAt
	p.DeleteAt = 0
}

func (o *ReplySnippet) PreUpdate() {
	o.UpdateAt = GetMillis()
}

func (p *ReplySnippet) ToJson() string {
	b, _ := json.Marshal(p)
	return string(b)
}

func ReplySnippetFromJson(data io.Reader) *ReplySnippet {
	var p *ReplySnippet
	json.NewDecoder(data).Decode(&p)
	return p
}

func ReplySnippetListToJson(p []*ReplySnippet) string {
	b, _ := json.Marshal(p)
	return string(b)
}


func ReplySnippetListFromJson(data io.Reader) []*ReplySnippet {
	var snippets []*ReplySnippet
	json.NewDecoder(data).Decode(&snippets)
	return snippets
}
