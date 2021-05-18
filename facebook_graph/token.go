// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package facebookgraph

import (
	"encoding/json"
	"io"
)

type FacebookToken struct {
	AccessToken string `json:"access_token"`
	TokenType   int64  `json:"token_type"`
}

func FacebookTokenFromJson(data io.Reader) *FacebookToken {
	var token *FacebookToken
	json.NewDecoder(data).Decode(&token)
	return token
}

func FacebookTokenToJson(t *FacebookToken) string {
	b, _ := json.Marshal(t)
	return string(b)
}