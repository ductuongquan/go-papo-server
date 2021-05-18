// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package model

import (
	"encoding/json"
	"io"
)

type PageTag struct {
	Id 				string 				`json:"id"`
	PageId 			string 				`json:"page_id"`
	Name 			string 				`json:"name"`
	Color 			string 				`json:"color"`
	Creator 		string 				`json:"creator"`
	CreateAt		int64 				`json:"create_at"`
	UpdateAt 		int64 				`json:"update_at"`
	DeleteAt 		int64 				`json:"delete_at"`
	IsPrivate 		bool 				`json:"is_private"`
	Visible 		bool 				`json:"visible"`
}

func (p *PageTag) PreSave() {
	if p.Id == "" {
		p.Id = NewId()
	}
	p.Visible = true
	p.CreateAt = GetMillis()
	p.UpdateAt = p.CreateAt
	p.DeleteAt = 0
}

func (o *PageTag) PreUpdate() {
	o.UpdateAt = GetMillis()
}

func (p *PageTag) ToJson() string {
	b, _ := json.Marshal(p)
	return string(b)
}

func PageTagFromJson(data io.Reader) *PageTag {
	var p *PageTag
	json.NewDecoder(data).Decode(&p)
	return p
}

func PageTagsToJson(p []*PageTag) string {
	b, _ := json.Marshal(p)
	return string(b)
}