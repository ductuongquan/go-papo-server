// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package facebookgraph

import (
	"encoding/json"
	"io"
)

type FacebookMessageItem struct {
	CreatedTime string                         `json:"created_time"`
	From        map[string]interface{}         `json:"from"`
	To          FacebookMessageToData          `json:"to"`
	Id          string                         `json:"id"`
	Message     string                         `json:"message"`
	Sticker     string                         `json:"sticker"`
	Attachments FacebookMessageItemAttachments `json:"attachments"`
}

type FacebookMessageAttachmentItem struct {
	Id        string            `json:"id"`
	MimeType  string            `json:"mime_type"`
	Name      string            `json:"name"`
	ImageData FacebookImageData `json:"image_data,omitempty"`
	VideoData FacebookVideoData `json:"video_data,omitempty"`
	Size 	  int64 			`json:"size,omitempty"`
	FileUrl   string 			`json:"file_url,omitempty"`

}

type FacebookMessageItemAttachments struct {
	Data []FacebookMessageAttachmentItem `json:"data,omitempty"`
}

type FacebookMessageToData struct {
	Data []map[string]interface{} `json:"data"`
}

func FacebookMessageItemFromJson(data io.Reader) *FacebookMessageItem {
	var u *FacebookMessageItem
	json.NewDecoder(data).Decode(&u)
	return u
}
