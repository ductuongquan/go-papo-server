// Copyright (c) 2015-present Ladifire, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package localcachelayer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"bitbucket.org/enesyteam/papo-server/model"
	"bitbucket.org/enesyteam/papo-server/store/storetest"
	"bitbucket.org/enesyteam/papo-server/store/storetest/mocks"
)

func TestWebhookStore(t *testing.T) {
	StoreTest(t, storetest.TestWebhookStore)
}

func TestWebhookStoreCache(t *testing.T) {
	fakeWebhook := model.IncomingWebhook{Id: "123"}

	t.Run("first call not cached, second cached and returning same data", func(t *testing.T) {
		mockStore := getMockStore()
		mockCacheProvider := getMockCacheProvider()
		cachedStore := NewLocalCacheLayer(mockStore, nil, nil, mockCacheProvider)

		incomingWebhook, err := cachedStore.Webhook().GetIncoming("123", true)
		require.Nil(t, err)
		assert.Equal(t, incomingWebhook, &fakeWebhook)
		mockStore.Webhook().(*mocks.WebhookStore).AssertNumberOfCalls(t, "GetIncoming", 1)

		assert.Equal(t, incomingWebhook, &fakeWebhook)
		cachedStore.Webhook().GetIncoming("123", true)
		mockStore.Webhook().(*mocks.WebhookStore).AssertNumberOfCalls(t, "GetIncoming", 1)
	})

	t.Run("first call not cached, second force not cached", func(t *testing.T) {
		mockStore := getMockStore()
		mockCacheProvider := getMockCacheProvider()
		cachedStore := NewLocalCacheLayer(mockStore, nil, nil, mockCacheProvider)

		cachedStore.Webhook().GetIncoming("123", true)
		mockStore.Webhook().(*mocks.WebhookStore).AssertNumberOfCalls(t, "GetIncoming", 1)
		cachedStore.Webhook().GetIncoming("123", false)
		mockStore.Webhook().(*mocks.WebhookStore).AssertNumberOfCalls(t, "GetIncoming", 2)
	})

	t.Run("first call not cached, invalidate, and then not cached again", func(t *testing.T) {
		mockStore := getMockStore()
		mockCacheProvider := getMockCacheProvider()
		cachedStore := NewLocalCacheLayer(mockStore, nil, nil, mockCacheProvider)

		cachedStore.Webhook().GetIncoming("123", true)
		mockStore.Webhook().(*mocks.WebhookStore).AssertNumberOfCalls(t, "GetIncoming", 1)
		cachedStore.Webhook().InvalidateWebhookCache("123")
		cachedStore.Webhook().GetIncoming("123", true)
		mockStore.Webhook().(*mocks.WebhookStore).AssertNumberOfCalls(t, "GetIncoming", 2)
	})
}
