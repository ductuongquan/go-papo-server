// Copyright (c) 2016-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package app

import (
	"bitbucket.org/enesyteam/papo-server/mlog"
	"bitbucket.org/enesyteam/papo-server/model"
	"fmt"
	"net/http"
)

func (app *App) GetPreferencesForUser(userId string) (model.Preferences, *model.AppError) {
	result := <-app.Srv.Store.Preference().GetAll(userId)
	if result.Err != nil {
		result.Err.StatusCode = http.StatusBadRequest
		return nil, result.Err
	}
	return result.Data.(model.Preferences), nil
}

func (app *App) GetPreferenceByCategoryForUser(userId string, category string) (model.Preferences, *model.AppError) {
	result := <-app.Srv.Store.Preference().GetCategory(userId, category)
	if result.Err != nil {
		result.Err.StatusCode = http.StatusBadRequest
		return nil, result.Err
	}
	if len(result.Data.(model.Preferences)) == 0 {
		err := model.NewAppError("getPreferenceCategory", "api.preference.preferences_category.get.app_error", nil, "", http.StatusNotFound)
		return nil, err
	}
	return result.Data.(model.Preferences), nil
}

func (app *App) GetPreferenceByCategoryAndNameForUser(userId string, category string, preferenceName string) (*model.Preference, *model.AppError) {
	result := <-app.Srv.Store.Preference().Get(userId, category, preferenceName)
	if result.Err != nil {
		result.Err.StatusCode = http.StatusBadRequest
		return nil, result.Err
	}
	data := result.Data.(model.Preference)
	return &data, nil
}

func (app *App) UpdatePreferences(userId string, preferences model.Preferences) *model.AppError {
	for _, preference := range preferences {
		if userId != preference.UserId {
			return model.NewAppError("savePreferences", "api.preference.update_preferences.set.app_error", nil,
				"userId="+userId+", preference.UserId="+preference.UserId, http.StatusForbidden)
		}
	}

	if result := <-app.Srv.Store.Preference().Save(&preferences); result.Err != nil {
		result.Err.StatusCode = http.StatusBadRequest
		return result.Err
	}

	message := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_PREFERENCES_CHANGED, "", "", userId, nil)
	message.Add("preferences", preferences.ToJson())
	app.Publish(message)

	return nil
}

func (app *App) DeletePreferences(userId string, preferences model.Preferences) *model.AppError {
	for _, preference := range preferences {
		if userId != preference.UserId {
			err := model.NewAppError("deletePreferences", "api.preference.delete_preferences.delete.app_error", nil,
				"userId="+userId+", preference.UserId="+preference.UserId, http.StatusForbidden)
			return err
		}
	}

	for _, preference := range preferences {
		if result := <-app.Srv.Store.Preference().Delete(userId, preference.Category, preference.Name); result.Err != nil {
			result.Err.StatusCode = http.StatusBadRequest
			return result.Err
		}
	}

	message := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_PREFERENCES_DELETED, "", "", userId, nil)
	message.Add("preferences", preferences.ToJson())
	app.Publish(message)

	return nil
}

func (app *App) CreateUserSelectedPages(userId string, pageIds string) *model.AppError {

	pref := model.Preference{UserId: userId, Category: model.PREFERENCE_CATEGORY_SELECTED_PAGES, Name: userId, Value: pageIds}
	if presult := <-app.Srv.Store.Preference().Save(&model.Preferences{pref}); presult.Err != nil {
		mlog.Error(fmt.Sprintf("Encountered error saving selected pages preference, err=%v", presult.Err.Message))
	}

	message := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_PREFERENCES_ADDED, "", "", userId, nil)
	message.Add("preference", pref.ToJson())
	app.Publish(message)

	return nil
}
