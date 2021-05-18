package model

import "encoding/json"

type FacebookUid struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	PageId string `json:"page_id"`
	PageScopeId string `json:"page_scope_id"`
}

func FacebookUserListToJson(p []*FacebookUid) string {
	b, _ := json.Marshal(p)
	return string(b)
}
