// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package expirynotify

import (
	"time"

	"bitbucket.org/enesyteam/papo-server/app"
	"bitbucket.org/enesyteam/papo-server/model"
)

const (
	SchedFreqMinutes = 10
)

type Scheduler struct {
	App *app.App
}

func (m *ExpiryNotifyJobInterfaceImpl) MakeScheduler() model.Scheduler {
	return &Scheduler{m.App}
}

func (scheduler *Scheduler) Name() string {
	return JobName + "Scheduler"
}

func (scheduler *Scheduler) JobType() string {
	return model.JOB_TYPE_EXPIRY_NOTIFY
}

func (scheduler *Scheduler) Enabled(cfg *model.Config) bool {
	// Only enabled when ExtendSessionLengthWithActivity is enabled.
	return *cfg.ServiceSettings.ExtendSessionLengthWithActivity
}

func (scheduler *Scheduler) NextScheduleTime(cfg *model.Config, now time.Time, pendingJobs bool, lastSuccessfulJob *model.Job) *time.Time {
	nextTime := time.Now().Add(SchedFreqMinutes * time.Minute)
	return &nextTime
}

func (scheduler *Scheduler) ScheduleJob(cfg *model.Config, pendingJobs bool, lastSuccessfulJob *model.Job) (*model.Job, *model.AppError) {
	data := map[string]string{}

	if job, err := scheduler.App.Srv().Jobs.CreateJob(model.JOB_TYPE_EXPIRY_NOTIFY, data); err != nil {
		return nil, err
	} else {
		return job, nil
	}
}
