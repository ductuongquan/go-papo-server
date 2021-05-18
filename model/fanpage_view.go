// Copyright (c) 2016-present Papo, Inc. All Rights Reserved.
// See License.txt for license information.

package model

import (
	"encoding/json"
	"io"
)

type PageView struct {
	PageId     string `json:"page_id"`
}

func (o *PageView) ToJson() string {
	b, _ := json.Marshal(o)
	return string(b)
}

func PageViewFromJson(data io.Reader) *PageView {
	var o *PageView
	json.NewDecoder(data).Decode(&o)
	return o
}

type PageViewResponse struct {
	Status            string           `json:"status"`
	LastViewedAtTimes map[string]int64 `json:"last_viewed_at_times"`
}

func (o *PageViewResponse) ToJson() string {
	b, _ := json.Marshal(o)
	return string(b)
}

func PageViewResponseFromJson(data io.Reader) *PageViewResponse {
	var o *PageViewResponse
	json.NewDecoder(data).Decode(&o)
	return o
}
