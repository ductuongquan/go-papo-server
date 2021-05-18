// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package model

import (
	"encoding/json"
	"io"
)

// Tự động gửi tin nhắn đến các hội thoại theo bộ lọc được chỉ định
type AutoMessageTask struct {
	Id 						string 			`json:"id"`
	Name 					string 			`json:"name"`
	Description 			string 			`json:"description"`
	PageId 					string 			`json:"page_id"` // TASK của page nào
	Message 				string 			`json:"message"` // nội dung tin nhắn
	Attachments 			string 			`json:"attachments"` // gửi đính kèm
	Creator 				string 			`json:"creator"`
	Status 					string 			`json:"status"` // trạng thái của TASK = 'created', 'pending', 'active', 'succes', 'error'
	CreateAt 				int64 			`json:"create_at"`
	UpdateAt 				int64 			`json:"update_at"`
	StartAt 				int64 			`json:"start_at"`
	EndAt 					int64 			`json:"end_at"`
	DeleteAt 				int64 			`json:"delete_at"`
	FilterFromDate			int64 			`json:"filter_from_date"` // từ hội thoại ngày ...
	FilterToDate 			int64 			`json:"filter_to_date"` // đến hội thoại ngày ...
	FilterTags 				string 			`json:"filter_tags"` // gửi cho hội thoại có chứa tags
	FilterHasPhone 			bool 			`json:"filter_has_phone"` // true => gửi cho hội thoại có số đt
	FilterNotHasPhone 		bool 			`json:"filter_not_has_phone"` // true => gửi cho hội thoại k có số đt
	FilterHasOrder 			bool 			`json:"filter_has_order"` // true => gửi cho hội thoại đã có order
	FilterNotHasOrder 		bool 			`json:"filter_not_has_order"` // true => gửi cho hội thoại chưa có order
	FilterReceptions 		string 			`json:"filer_receptions"` // nếu có thì sẽ chỉ gửi tin nhắn cho những người này
}

func (p *AutoMessageTask) PreSave() {
	if p.Id == "" {
		p.Id = NewId()
	}
	p.Status = "created"
	p.CreateAt = GetMillis()
	p.UpdateAt = p.CreateAt
	p.DeleteAt = 0
}

func (o *AutoMessageTask) PreUpdate() {
	o.UpdateAt = GetMillis()
}

func (p *AutoMessageTask) ToJson() string {
	b, _ := json.Marshal(p)
	return string(b)
}

func AutoMessageTaskFromJson(data io.Reader) *AutoMessageTask {
	var p *AutoMessageTask
	json.NewDecoder(data).Decode(&p)
	return p
}

func AutoMessageTasksToJson(p []*AutoMessageTask) string {
	b, _ := json.Marshal(p)
	return string(b)
}

func AutoMessageTasksFromJson(data io.Reader) []*AutoMessageTask {
	var tasks []*AutoMessageTask
	json.NewDecoder(data).Decode(&tasks)
	return tasks
}

