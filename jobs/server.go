// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package jobs

import (
	ejobs "bitbucket.org/enesyteam/papo-server/einterfaces/jobs"
	tjobs "bitbucket.org/enesyteam/papo-server/jobs/interfaces"
	"bitbucket.org/enesyteam/papo-server/model"
	"bitbucket.org/enesyteam/papo-server/services/configservice"
	"bitbucket.org/enesyteam/papo-server/store"
)

type JobServer struct {
	ConfigService configservice.ConfigService
	Store         store.Store
	Workers       *Workers
	Schedulers    *Schedulers

	DataRetentionJob        ejobs.DataRetentionJobInterface
	MessageExportJob        ejobs.MessageExportJobInterface
	ElasticsearchAggregator ejobs.ElasticsearchAggregatorInterface
	ElasticsearchIndexer    tjobs.IndexerJobInterface
	LdapSync                ejobs.LdapSyncInterface
	Migrations              tjobs.MigrationsJobInterface
	Plugins                 tjobs.PluginsJobInterface
	BleveIndexer            tjobs.IndexerJobInterface
	ExpiryNotify            tjobs.ExpiryNotifyJobInterface
	ProductNotices          tjobs.ProductNoticesJobInterface
	ActiveUsers             tjobs.ActiveUsersJobInterface
}

func NewJobServer(configService configservice.ConfigService, store store.Store) *JobServer {
	return &JobServer{
		ConfigService: configService,
		Store:         store,
	}
}

func (srv *JobServer) Config() *model.Config {
	return srv.ConfigService.Config()
}

func (srv *JobServer) StartWorkers() {
	srv.Workers = srv.Workers.Start()
}

func (srv *JobServer) StartSchedulers() {
	srv.Schedulers = srv.Schedulers.Start()
}

func (srv *JobServer) StopWorkers() {
	if srv.Workers != nil {
		srv.Workers.Stop()
	}
}

func (srv *JobServer) StopSchedulers() {
	if srv.Schedulers != nil {
		srv.Schedulers.Stop()
	}
}
