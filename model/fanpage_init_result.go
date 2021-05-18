package model

import (
	"encoding/json"
	"io"
)

type FanpageInitResult struct {
	Id 						string 				`json:"id"`
	PageId 					string 				`json:"page_id"`
	PostCount 				int64 				`json:"post_count"`
	ConversationCount 		int64 				`json:"conversation_count"`
	MessageCount 			int64 				`json:"message_count"`
	CommentCount 			int64 				`json:"comment_count"`
	StartAt 				int64 				`json:"start_at"`
	EndAt 					int64 				`json:"end_at"`
	Creator 				string 				`json:"creator"`
}

func (p *FanpageInitResult) PreSave() {
	if p.Id == "" {
		p.Id = NewId()
	}
	p.StartAt = GetMillis()
}

func (p *FanpageInitResult) ToJson() string {
	b, _ := json.Marshal(p)
	return string(b)
}

func FanpageInitResultFromJson(data io.Reader) *FanpageInitResult {
	var p *FanpageInitResult
	json.NewDecoder(data).Decode(&p)
	return p
}

func FanpageInitResultListToJson(p []*FanpageInitResult) string {
	b, _ := json.Marshal(p)
	return string(b)
}