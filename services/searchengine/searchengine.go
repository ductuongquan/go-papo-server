// Copyright (c) 2015-present Ladifire, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package searchengine

import (
	"bitbucket.org/enesyteam/papo-server/jobs"
	"bitbucket.org/enesyteam/papo-server/model"
)

func NewBroker(cfg *model.Config, jobServer *jobs.JobServer) *Broker {
	return &Broker{
		cfg:       cfg,
		jobServer: jobServer,
	}
}

func (seb *Broker) RegisterElasticsearchEngine(es SearchEngineInterface) {
	seb.ElasticsearchEngine = es
}

func (seb *Broker) RegisterBleveEngine(be SearchEngineInterface) {
	seb.BleveEngine = be
}

type Broker struct {
	cfg                 *model.Config
	jobServer           *jobs.JobServer
	ElasticsearchEngine SearchEngineInterface
	BleveEngine         SearchEngineInterface
}

func (seb *Broker) UpdateConfig(cfg *model.Config) *model.AppError {
	seb.cfg = cfg
	if seb.ElasticsearchEngine != nil {
		seb.ElasticsearchEngine.UpdateConfig(cfg)
	}

	if seb.BleveEngine != nil {
		seb.BleveEngine.UpdateConfig(cfg)
	}

	return nil
}

func (seb *Broker) GetActiveEngines() []SearchEngineInterface {
	engines := []SearchEngineInterface{}
	if seb.ElasticsearchEngine != nil && seb.ElasticsearchEngine.IsActive() {
		engines = append(engines, seb.ElasticsearchEngine)
	}
	if seb.BleveEngine != nil && seb.BleveEngine.IsActive() {
		engines = append(engines, seb.BleveEngine)
	}
	return engines
}
