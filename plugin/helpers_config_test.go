// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package plugin_test

import (
	"testing"

	"bitbucket.org/enesyteam/papo-server/model"
	"bitbucket.org/enesyteam/papo-server/plugin"
	"bitbucket.org/enesyteam/papo-server/plugin/plugintest"
	"github.com/stretchr/testify/assert"
)

func TestCheckRequiredServerConfiguration(t *testing.T) {
	for name, test := range map[string]struct {
		SetupAPI     func(*plugintest.API) *plugintest.API
		Input        *model.Config
		ShouldReturn bool
		ShouldError  bool
	}{
		"no required config therefore it should be compatible": {
			SetupAPI: func(api *plugintest.API) *plugintest.API {
				return api
			},
			Input:        nil,
			ShouldReturn: true,
			ShouldError:  false,
		},
		"contains required configuration": {
			SetupAPI: func(api *plugintest.API) *plugintest.API {
				api.On("GetConfig").Return(&model.Config{
					ServiceSettings: model.ServiceSettings{
						EnableCommands: model.NewBool(true),
					},
					TeamSettings: model.TeamSettings{
						EnableUserCreation: model.NewBool(true),
					},
				})

				return api
			},
			Input: &model.Config{
				ServiceSettings: model.ServiceSettings{
					EnableCommands: model.NewBool(true),
				},
			},
			ShouldReturn: true,
			ShouldError:  false,
		},
		"does not contain required configuration": {
			SetupAPI: func(api *plugintest.API) *plugintest.API {
				api.On("GetConfig").Return(&model.Config{
					ServiceSettings: model.ServiceSettings{
						EnableCommands: model.NewBool(true),
					},
				})

				return api
			},
			Input: &model.Config{
				ServiceSettings: model.ServiceSettings{
					EnableCommands: model.NewBool(true),
				},
				TeamSettings: model.TeamSettings{
					EnableUserCreation: model.NewBool(true),
				},
			},
			ShouldReturn: false,
			ShouldError:  false,
		},
		"different configurations": {
			SetupAPI: func(api *plugintest.API) *plugintest.API {
				api.On("GetConfig").Return(&model.Config{
					ServiceSettings: model.ServiceSettings{
						EnableCommands: model.NewBool(false),
					},
				})

				return api
			},
			Input: &model.Config{
				ServiceSettings: model.ServiceSettings{
					EnableCommands: model.NewBool(true),
				},
			},
			ShouldReturn: false,
			ShouldError:  false,
		},
		"non-existent configuration": {
			SetupAPI: func(api *plugintest.API) *plugintest.API {
				api.On("GetConfig").Return(&model.Config{})

				return api
			},
			Input: &model.Config{
				ServiceSettings: model.ServiceSettings{
					EnableCommands: model.NewBool(true),
				},
			},
			ShouldReturn: false,
			ShouldError:  false,
		},
	} {
		t.Run(name, func(t *testing.T) {
			api := test.SetupAPI(&plugintest.API{})
			defer api.AssertExpectations(t)

			p := &plugin.HelpersImpl{}
			p.API = api

			ok, err := p.CheckRequiredServerConfiguration(test.Input)

			assert.Equal(t, test.ShouldReturn, ok)
			if test.ShouldError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
