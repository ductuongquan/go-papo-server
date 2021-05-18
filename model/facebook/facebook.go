// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package oauthfacebook

import (
	"bitbucket.org/enesyteam/papo-server/einterfaces"
	"bitbucket.org/enesyteam/papo-server/mlog"
	"bitbucket.org/enesyteam/papo-server/model"
	"encoding/json"
	"io"
	"strings"
)

type FacebookProvider struct {
}

type FacebookUser struct {
	Id 			string		`json:"id"`
	Name 		string		`json:"name"`
	Email 		string		`json:"email"`
}

func init()  {
	//mlog.Info( "Facebook provider is initialized" )
	provider := &FacebookProvider{}
	einterfaces.RegisterOauthProvider( model.USER_AUTH_SERVICE_FACEBOOK, provider )
}

func userFromFacebookUser(fbu *FacebookUser) *model.User {
	user := &model.User{}
	name := fbu.Name
	if name == "" {
		name = fbu.Name
	}
	user.Username = model.CleanUsername(name)
	splitName := strings.Split(fbu.Name, " ")
	if len(splitName) == 2 {
		user.FirstName = splitName[0]
		user.LastName = splitName[1]
	} else if len(splitName) >= 2 {
		user.FirstName = splitName[0]
		user.LastName = strings.Join(splitName[1:], " ")
	} else {
		user.FirstName = fbu.Name
	}
	user.Email = fbu.Email //.Value
	userId := fbu.getAuthData()
	user.AuthData = &userId
	user.AuthService = model.USER_AUTH_SERVICE_FACEBOOK

	return user
}

func facebookUserFromJson(data io.Reader) *FacebookUser {
	var fbu *FacebookUser
	err := json.NewDecoder(data).Decode(&fbu)
	if err == nil {
		return fbu
	} else {
		return nil
	}
}

func (fbu *FacebookUser) getAuthData() string {
	return fbu.Id
}

func (fbu *FacebookUser) ToJson() string {
	b, err := json.Marshal(fbu)
	if err != nil {
		return ""
	} else {
		return string(b)
	}
}

func (fbu *FacebookUser) IsValid() bool {
	if len(fbu.Id) == 0 {
		return false
	}

	if len(fbu.Email) == 0 {
		mlog.Info( "email not found" )
		return false
	}

	return true
}

func (m *FacebookProvider) GetUserFromJson(data io.Reader) *model.User {
	fbu := facebookUserFromJson(data)
	if fbu.IsValid() {
		return userFromFacebookUser(fbu)
	}
	return &model.User{}
}