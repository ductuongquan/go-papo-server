// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package model

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

type FanpageMember struct {
	FanpageId   string `json:"fanpage_id"`
	PageId 		string `json:"page_id"`
	UserId      string `json:"user_id"`
	Roles       string `json:"roles,omitempty"`
	AccessToken string `json:"access_token"`
	LastViewedAt  int64     `json:"last_viewed_at"`
	MsgCount      int64     `json:"msg_count"`
	NotifyProps   StringMap `json:"notify_props"`
	LastUpdateAt  int64     `json:"last_update_at"`
}

func (o *FanpageMember) ToJson() string {
	b, _ := json.Marshal(o)
	return string(b)
}

func FanpageMemberFromJson(data io.Reader) *FanpageMember {
	var o *FanpageMember
	json.NewDecoder(data).Decode(&o)
	return o
}

func FanpageMemberToJson(o []*FanpageMember) string {
	if b, err := json.Marshal(o); err != nil {
		return "[]"
	} else {
		return string(b)
	}
}

func FanpageMembersFromJson(data io.Reader) []*FanpageMember {
	var o []*FanpageMember
	json.NewDecoder(data).Decode(&o)
	return o
}

func (o *FanpageMember) PreUpdate() {
}

func (o *FanpageMember) IsValid() *AppError {

	if len(o.FanpageId) != 26 {
		return NewAppError("FanpageMember.IsValid", "model.fanpage_member.is_valid.fanpage_id.app_error", nil, "", http.StatusBadRequest)
	}

	if len(o.UserId) != 26 {
		return NewAppError("FanpageMember.IsValid", "model.fanpage_member.is_valid.user_id.app_error", nil, "", http.StatusBadRequest)
	}

	return nil
}

func (o *FanpageMember) GetRoles() []string {
	return strings.Fields(o.Roles)
}
