// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package bleveengine

import (
	"fmt"

	"bitbucket.org/enesyteam/papo-server/model"
)

func createPost(userId string, channelId string, message string) *model.Post {
	post := &model.Post{
		Message:       message,
		ChannelId:     channelId,
		PendingPostId: model.NewId() + ":" + fmt.Sprint(model.GetMillis()),
		UserId:        userId,
		CreateAt:      1000000,
	}
	post.PreSave()

	return post
}
