// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package plugin

import (
	"github.com/pkg/errors"

	"bitbucket.org/enesyteam/papo-server/model"
	"bitbucket.org/enesyteam/papo-server/utils"
)

// CheckRequiredServerConfiguration implements Helpers.CheckRequiredServerConfiguration
func (p *HelpersImpl) CheckRequiredServerConfiguration(req *model.Config) (bool, error) {
	if req == nil {
		return true, nil
	}

	cfg := p.API.GetConfig()

	mc, err := utils.Merge(cfg, req, nil)
	if err != nil {
		return false, errors.Wrap(err, "could not merge configurations")
	}

	mergedCfg := mc.(model.Config)
	if mergedCfg.ToJson() != cfg.ToJson() {
		return false, nil
	}

	return true, nil
}
