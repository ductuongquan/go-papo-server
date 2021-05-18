// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package einterfaces

import "bitbucket.org/enesyteam/papo-server/model"

type SamlInterface interface {
	ConfigureSP() *model.AppError
	BuildRequest(relayState string) (*model.SamlAuthRequest, *model.AppError)
	DoLogin(encodedXML string, relayState map[string]string) (*model.User, *model.AppError)
	GetMetadata() (string, *model.AppError)
}