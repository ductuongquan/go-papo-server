// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package model

import "encoding/json"

func FacebookAttachmentImageListToJson(p []*FacebookAttachmentImage) string {
	b, _ := json.Marshal(p)
	return string(b)
}