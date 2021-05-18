// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package facebookgraph

type FacebookCursors struct {
	Before string `json:"before"`
	After  string `json:"after"`
}

type FacebookPaging struct {
	Cursors  FacebookCursors `json:"cursors"`
	Next     string          `json:"next"`
	Previous string          `json:"previous"`
}