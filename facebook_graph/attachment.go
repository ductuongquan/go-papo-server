// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package facebookgraph

import (
	"encoding/json"
	"io"
)

type FacebookAttachmentImage struct {
	Height int    `json:"height"`
	Src    string `json:"src"`
	Width  int    `json:"width"`
}

type FacebookAttachmentImageItem struct {
	Image FacebookAttachmentImage `json:"image"`
}

type FacebookSubAttachmentItem struct {
	Media  FacebookAttachmentImageItem `json:"media"`
	Url            	string                  `json:"url"`
	Target 			TargetItem 				`json:"target"`
	Type   			string                  `json:"type"`
}

type FacebookSubAttachments struct {
	Data []FacebookSubAttachmentItem `json:"data"`
}

type FacebookAttachmentItem struct {
	Media          FacebookAttachmentImageItem `json:"media"`
	Target 			TargetItem 				`json:"target"`
	Type           string                  `json:"type"`
	Url            string                  `json:"url"`
	SubAttachments FacebookSubAttachments  `json:"subattachments"`
}

type FacebookAttachments struct {
	Data 		[]FacebookAttachmentItem `json:"data"`
	Type 		string                   `json:"type"`
	Url  		string                   `json:"url"`
	Target 		TargetItem 				`json:"target"`
}
//SubAttachments SubAttachmentsItem `json:"subattachments"`

type SubAttachmentsItem struct {
	Data []FacebookAttachmentItem `json:"data"`
}

func FacebookAttachmentImageFromJson(data io.Reader) *FacebookAttachmentImage {
	var u *FacebookAttachmentImage
	json.NewDecoder(data).Decode(&u)
	return u
}

func FacebookAttachmentItemFromJson(data io.Reader) *FacebookAttachmentItem {
	var u *FacebookAttachmentItem
	json.NewDecoder(data).Decode(&u)
	return u
}

func FacebookAttachmentsFromJson(data io.Reader) *FacebookAttachments {
	var u *FacebookAttachments
	json.NewDecoder(data).Decode(&u)
	return u
}
