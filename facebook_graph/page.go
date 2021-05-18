// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package facebookgraph

import (
	"encoding/json"
	"io"
	"io/ioutil"
)

type FacebookPageCategoryItem struct {
	Id 			string 	`json:"id"`
	Name 		string 	`json:"name"`
}

type FacebookPage struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	AccessToken string `json:"access_token"`
	CategoryList []*FacebookPageCategoryItem	`json:"category_list"`
	Task 		[]string `json:"tasks"`
	UserName 	string `json:"username"`
	Phone 		string `json:"phone"`
}

type FacebookPages struct {
	Data   []FacebookPage `json:"data"`
	Paging FacebookPaging `json:"paging"`
}

// Facebook Page
func FacebookPageFromJson(data io.Reader) *FacebookPage {
	var u *FacebookPage
	json.NewDecoder(data).Decode(&u)
	return u
}

func FacebookPageToJson(p *FacebookPage) string {
	b, _ := json.Marshal(p)
	return string(b)
}

// Facebook Pages
func FacebookPagesFromJson(data io.Reader) *FacebookPages {
	var pages *FacebookPages
	x, _ := ioutil.ReadAll(data)
	json.Unmarshal([]byte(x), &pages)
	return pages
}

func FacebookPageListToJson(p []FacebookPage) string {
	b, _ := json.Marshal(p)
	return string(b)
}

