// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package einterfaces

import (
	"bitbucket.org/enesyteam/papo-server/model"
)

type LdapInterface interface {
	DoLogin(id string, password string) (*model.User, *model.AppError)
	GetUser(id string) (*model.User,  *model.AppError)
	ValidateFilter(filter string) *model.AppError
}