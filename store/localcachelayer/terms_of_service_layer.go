// Copyright (c) 2015-present Ladifire, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package localcachelayer

import (
	"bitbucket.org/enesyteam/papo-server/model"
	"bitbucket.org/enesyteam/papo-server/store"
)

const (
	LATEST_KEY = "latest"
)

type LocalCacheTermsOfServiceStore struct {
	store.TermsOfServiceStore
	rootStore *LocalCacheStore
}

func (s *LocalCacheTermsOfServiceStore) handleClusterInvalidateTermsOfService(msg *model.ClusterMessage) {
	if msg.Data == CLEAR_CACHE_MESSAGE_DATA {
		s.rootStore.termsOfServiceCache.Purge()
	} else {
		s.rootStore.termsOfServiceCache.Remove(msg.Data)
	}
}

func (s LocalCacheTermsOfServiceStore) ClearCaches() {
	s.rootStore.doClearCacheCluster(s.rootStore.termsOfServiceCache)

	if s.rootStore.metrics != nil {
		s.rootStore.metrics.IncrementMemCacheInvalidationCounter("Terms Of Service - Purge")
	}
}

func (s LocalCacheTermsOfServiceStore) Save(termsOfService *model.TermsOfService) (*model.TermsOfService, error) {
	tos, err := s.TermsOfServiceStore.Save(termsOfService)

	if err == nil {
		s.rootStore.doStandardAddToCache(s.rootStore.termsOfServiceCache, tos.Id, tos)
		s.rootStore.doInvalidateCacheCluster(s.rootStore.termsOfServiceCache, LATEST_KEY)
	}
	return tos, err
}

func (s LocalCacheTermsOfServiceStore) GetLatest(allowFromCache bool) (*model.TermsOfService, error) {
	if allowFromCache {
		if len, err := s.rootStore.termsOfServiceCache.Len(); err == nil && len != 0 {
			var cacheItem *model.TermsOfService
			if err := s.rootStore.doStandardReadCache(s.rootStore.termsOfServiceCache, LATEST_KEY, &cacheItem); err == nil {
				return cacheItem, nil
			}
		}
	}

	termsOfService, err := s.TermsOfServiceStore.GetLatest(allowFromCache)

	if allowFromCache && err == nil {
		s.rootStore.doStandardAddToCache(s.rootStore.termsOfServiceCache, termsOfService.Id, termsOfService)
		s.rootStore.doStandardAddToCache(s.rootStore.termsOfServiceCache, LATEST_KEY, termsOfService)
	}

	return termsOfService, err
}

func (s LocalCacheTermsOfServiceStore) Get(id string, allowFromCache bool) (*model.TermsOfService, error) {
	if allowFromCache {
		var cacheItem *model.TermsOfService
		if err := s.rootStore.doStandardReadCache(s.rootStore.termsOfServiceCache, id, &cacheItem); err == nil {
			return cacheItem, nil
		}
	}

	termsOfService, err := s.TermsOfServiceStore.Get(id, allowFromCache)

	if allowFromCache && err == nil {
		s.rootStore.doStandardAddToCache(s.rootStore.termsOfServiceCache, termsOfService.Id, termsOfService)
	}

	return termsOfService, err
}
