// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"

	"bitbucket.org/enesyteam/papo-server/mlog"
	"bitbucket.org/enesyteam/papo-server/model"
	"bitbucket.org/enesyteam/papo-server/store"
)

func getKeyHash(key string) string {
	hash := sha256.New()
	hash.Write([]byte(key))
	return base64.StdEncoding.EncodeToString(hash.Sum(nil))
}

func (a *App) SetPluginKey(pluginId string, key string, value []byte) *model.AppError {
	return a.SetPluginKeyWithExpiry(pluginId, key, value, 0)
}

func (a *App) SetPluginKeyWithExpiry(pluginId string, key string, value []byte, expireInSeconds int64) *model.AppError {
	options := model.PluginKVSetOptions{
		ExpireInSeconds: expireInSeconds,
	}
	_, err := a.SetPluginKeyWithOptions(pluginId, key, value, options)
	return err
}

func (a *App) CompareAndSetPluginKey(pluginId string, key string, oldValue, newValue []byte) (bool, *model.AppError) {
	options := model.PluginKVSetOptions{
		Atomic:   true,
		OldValue: oldValue,
	}
	return a.SetPluginKeyWithOptions(pluginId, key, newValue, options)
}

func (a *App) SetPluginKeyWithOptions(pluginId string, key string, value []byte, options model.PluginKVSetOptions) (bool, *model.AppError) {
	if err := options.IsValid(); err != nil {
		mlog.Error("Failed to set plugin key value with options", mlog.String("plugin_id", pluginId), mlog.String("key", key), mlog.Err(err))
		return false, err
	}

	updated, err := a.Srv().Store.Plugin().SetWithOptions(pluginId, key, value, options)
	if err != nil {
		mlog.Error("Failed to set plugin key value with options", mlog.String("plugin_id", pluginId), mlog.String("key", key), mlog.Err(err))
		var appErr *model.AppError
		switch {
		case errors.As(err, &appErr):
			return false, appErr
		default:
			return false, model.NewAppError("SetPluginKeyWithOptions", "app.plugin_store.save.app_error", nil, err.Error(), http.StatusInternalServerError)
		}
	}

	// Clean up a previous entry using the hashed key, if it exists.
	if err := a.Srv().Store.Plugin().Delete(pluginId, getKeyHash(key)); err != nil {
		mlog.Error("Failed to clean up previously hashed plugin key value", mlog.String("plugin_id", pluginId), mlog.String("key", key), mlog.Err(err))
	}

	return updated, nil
}

func (a *App) CompareAndDeletePluginKey(pluginId string, key string, oldValue []byte) (bool, *model.AppError) {
	kv := &model.PluginKeyValue{
		PluginId: pluginId,
		Key:      key,
	}

	deleted, err := a.Srv().Store.Plugin().CompareAndDelete(kv, oldValue)
	if err != nil {
		mlog.Error("Failed to compare and delete plugin key value", mlog.String("plugin_id", pluginId), mlog.String("key", key), mlog.Err(err))
		var appErr *model.AppError
		switch {
		case errors.As(err, &appErr):
			return deleted, appErr
		default:
			return false, model.NewAppError("CompareAndDeletePluginKey", "app.plugin_store.delete.app_error", nil, err.Error(), http.StatusInternalServerError)
		}
	}

	// Clean up a previous entry using the hashed key, if it exists.
	if err := a.Srv().Store.Plugin().Delete(pluginId, getKeyHash(key)); err != nil {
		mlog.Error("Failed to clean up previously hashed plugin key value", mlog.String("plugin_id", pluginId), mlog.String("key", key), mlog.Err(err))
	}

	return deleted, nil
}

func (a *App) GetPluginKey(pluginId string, key string) ([]byte, *model.AppError) {
	if kv, err := a.Srv().Store.Plugin().Get(pluginId, key); err == nil {
		return kv.Value, nil
	} else if nfErr := new(store.ErrNotFound); !errors.As(err, &nfErr) {
		mlog.Error("Failed to query plugin key value", mlog.String("plugin_id", pluginId), mlog.String("key", key), mlog.Err(err))
		return nil, model.NewAppError("GetPluginKey", "app.plugin_store.get.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	// Lookup using the hashed version of the key for keys written prior to v5.6.
	if kv, err := a.Srv().Store.Plugin().Get(pluginId, getKeyHash(key)); err == nil {
		return kv.Value, nil
	} else if nfErr := new(store.ErrNotFound); !errors.As(err, &nfErr) {
		mlog.Error("Failed to query plugin key value using hashed key", mlog.String("plugin_id", pluginId), mlog.String("key", key), mlog.Err(err))
		return nil, model.NewAppError("GetPluginKey", "app.plugin_store.get.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil, nil
}

func (a *App) DeletePluginKey(pluginId string, key string) *model.AppError {
	if err := a.Srv().Store.Plugin().Delete(pluginId, getKeyHash(key)); err != nil {
		mlog.Error("Failed to delete plugin key value", mlog.String("plugin_id", pluginId), mlog.String("key", key), mlog.Err(err))
		return model.NewAppError("DeletePluginKey", "app.plugin_store.delete.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	// Also delete the key without hashing
	if err := a.Srv().Store.Plugin().Delete(pluginId, key); err != nil {
		mlog.Error("Failed to delete plugin key value using hashed key", mlog.String("plugin_id", pluginId), mlog.String("key", key), mlog.Err(err))
		return model.NewAppError("DeletePluginKey", "app.plugin_store.delete.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (a *App) DeleteAllKeysForPlugin(pluginId string) *model.AppError {
	if err := a.Srv().Store.Plugin().DeleteAllForPlugin(pluginId); err != nil {
		mlog.Error("Failed to delete all plugin key values", mlog.String("plugin_id", pluginId), mlog.Err(err))
		return model.NewAppError("DeleteAllKeysForPlugin", "app.plugin_store.delete.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (a *App) DeleteAllExpiredPluginKeys() *model.AppError {
	if a.Srv() == nil {
		return nil
	}

	if err := a.Srv().Store.Plugin().DeleteAllExpired(); err != nil {
		mlog.Error("Failed to delete all expired plugin key values", mlog.Err(err))
		return model.NewAppError("DeleteAllExpiredPluginKeys", "app.plugin_store.delete.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (a *App) ListPluginKeys(pluginId string, page, perPage int) ([]string, *model.AppError) {
	data, err := a.Srv().Store.Plugin().List(pluginId, page*perPage, perPage)

	if err != nil {
		mlog.Error("Failed to list plugin key values", mlog.Int("page", page), mlog.Int("perPage", perPage), mlog.Err(err))
		return nil, model.NewAppError("ListPluginKeys", "app.plugin_store.list.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return data, nil
}
