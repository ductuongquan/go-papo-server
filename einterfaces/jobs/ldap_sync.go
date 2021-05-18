// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package jobs

import (
	"bitbucket.org/enesyteam/papo-server/model"
)

type LdapSyncInterface interface {
	MakeWorker() model.Worker
	MakeScheduler() model.Scheduler
}
