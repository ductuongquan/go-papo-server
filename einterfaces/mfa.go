// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package einterfaces

import "bitbucket.org/enesyteam/papo-server/model"

type MfaInterface interface {
	GenerateSecret(user *model.User) (string, []byte, *model.AppError)
	Activate(user *model.User, token string) *model.AppError
	Deactivate(userId string) *model.AppError
	ValidateToken(secret, token string) (bool, *model.AppError)
}