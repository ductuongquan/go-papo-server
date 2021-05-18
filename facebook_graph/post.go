// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package facebookgraph

import (
	"encoding/json"
	"io"
	"io/ioutil"
)

type FacebookGraphPost struct {
	Attachments  FacebookAttachments `json:"attachments"`
	CreatedTime  string              `json:"created_time"`
	From         map[string]interface{} 	`json:"from"`
	Picture      string              `json:"picture"`
	UpdatedTime  string              `json:"updated_time"`
	IsHidden     bool                `json:"is_hidden"`
	PermalinkUrl string              `json:"permalink_url"`
	Message      string              `json:"message"`
	Id           string              `json:"id"`
	Story 		 string 			 `json:"story"`
}

type FacebookGraphPosts struct {
	Data   []FacebookGraphPost `json:"data"`
	Paging FacebookPaging `json:"paging"`
}

func FacebookPostFromJson(data io.Reader) *FacebookGraphPost {
	var u *FacebookGraphPost
	json.NewDecoder(data).Decode(&u)
	return u
}

func FacebookPostsFromJson(data io.Reader) *FacebookGraphPosts {
	var fbps *FacebookGraphPosts
	x, _ := ioutil.ReadAll(data)
	json.Unmarshal([]byte(x), &fbps)
	return fbps
}

func FacebookPostsToJson(p *FacebookGraphPosts) string {
	b, _ := json.Marshal(p)
	return string(b)
}

func FacebookPostToJson(p *FacebookGraphPosts) string {
	b, _ := json.Marshal(p)
	return string(b)
}