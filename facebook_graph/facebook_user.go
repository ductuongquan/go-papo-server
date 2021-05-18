// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package facebookgraph

import (
	"encoding/json"
	"io"
)

type FacebookUser struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Email string `json:"email"`
	PageId string `json:"page_id"`
}

func FacebookUserFromJson(data io.Reader) *FacebookUser {
	var u *FacebookUser
	json.NewDecoder(data).Decode(&u)
	return u
}
