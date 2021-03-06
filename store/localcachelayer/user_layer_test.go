// Copyright (c) 2015-present Ladifire, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package localcachelayer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"bitbucket.org/enesyteam/papo-server/model"
	"bitbucket.org/enesyteam/papo-server/store"
	"bitbucket.org/enesyteam/papo-server/store/storetest"
	"bitbucket.org/enesyteam/papo-server/store/storetest/mocks"
)

func TestUserStore(t *testing.T) {
	StoreTestWithSqlSupplier(t, storetest.TestUserStore)
}

func TestUserStoreCache(t *testing.T) {
	fakeUserIds := []string{"123"}
	fakeUser := []*model.User{{
		Id:          "123",
		AuthData:    model.NewString("authData"),
		AuthService: "authService",
	}}

	t.Run("first call not cached, second cached and returning same data", func(t *testing.T) {
		mockStore := getMockStore()
		mockCacheProvider := getMockCacheProvider()
		cachedStore := NewLocalCacheLayer(mockStore, nil, nil, mockCacheProvider)

		gotUser, err := cachedStore.User().GetProfileByIds(fakeUserIds, &store.UserGetByIdsOpts{}, true)
		require.Nil(t, err)
		assert.Equal(t, fakeUser, gotUser)
		mockStore.User().(*mocks.UserStore).AssertNumberOfCalls(t, "GetProfileByIds", 1)

		_, _ = cachedStore.User().GetProfileByIds(fakeUserIds, &store.UserGetByIdsOpts{}, true)
		mockStore.User().(*mocks.UserStore).AssertNumberOfCalls(t, "GetProfileByIds", 1)
	})

	t.Run("first call not cached, second force not cached", func(t *testing.T) {
		mockStore := getMockStore()
		mockCacheProvider := getMockCacheProvider()
		cachedStore := NewLocalCacheLayer(mockStore, nil, nil, mockCacheProvider)

		gotUser, err := cachedStore.User().GetProfileByIds(fakeUserIds, &store.UserGetByIdsOpts{}, true)
		require.Nil(t, err)
		assert.Equal(t, fakeUser, gotUser)
		mockStore.User().(*mocks.UserStore).AssertNumberOfCalls(t, "GetProfileByIds", 1)

		_, _ = cachedStore.User().GetProfileByIds(fakeUserIds, &store.UserGetByIdsOpts{}, false)
		mockStore.User().(*mocks.UserStore).AssertNumberOfCalls(t, "GetProfileByIds", 2)
	})

	t.Run("first call not cached, invalidate, and then not cached again", func(t *testing.T) {
		mockStore := getMockStore()
		mockCacheProvider := getMockCacheProvider()
		cachedStore := NewLocalCacheLayer(mockStore, nil, nil, mockCacheProvider)

		gotUser, err := cachedStore.User().GetProfileByIds(fakeUserIds, &store.UserGetByIdsOpts{}, true)
		require.Nil(t, err)
		assert.Equal(t, fakeUser, gotUser)

		cachedStore.User().InvalidateProfileCacheForUser("123")

		_, _ = cachedStore.User().GetProfileByIds(fakeUserIds, &store.UserGetByIdsOpts{}, true)
		mockStore.User().(*mocks.UserStore).AssertNumberOfCalls(t, "GetProfileByIds", 2)
	})

	t.Run("should always return a copy of the stored data", func(t *testing.T) {
		mockStore := getMockStore()
		mockCacheProvider := getMockCacheProvider()
		cachedStore := NewLocalCacheLayer(mockStore, nil, nil, mockCacheProvider)

		storedUsers, err := mockStore.User().GetProfileByIds(fakeUserIds, &store.UserGetByIdsOpts{}, false)
		require.Nil(t, err)

		originalProps := make([]model.StringMap, len(storedUsers))

		for i := 0; i < len(storedUsers); i++ {
			originalProps[i] = storedUsers[i].NotifyProps
			storedUsers[i].NotifyProps = map[string]string{}
			storedUsers[i].NotifyProps["key"] = "somevalue"
		}

		cachedUsers, err := cachedStore.User().GetProfileByIds(fakeUserIds, &store.UserGetByIdsOpts{}, true)
		require.Nil(t, err)

		for i := 0; i < len(storedUsers); i++ {
			assert.Equal(t, storedUsers[i].Id, cachedUsers[i].Id)
		}

		cachedUsers, err = cachedStore.User().GetProfileByIds(fakeUserIds, &store.UserGetByIdsOpts{}, true)
		require.Nil(t, err)
		for i := 0; i < len(storedUsers); i++ {
			storedUsers[i].Props = model.StringMap{}
			storedUsers[i].Timezone = model.StringMap{}
			assert.Equal(t, storedUsers[i], cachedUsers[i])
			if storedUsers[i] == cachedUsers[i] {
				assert.Fail(t, "should be different pointers")
			}
			cachedUsers[i].NotifyProps["key"] = "othervalue"
			assert.NotEqual(t, storedUsers[i], cachedUsers[i])
		}

		for i := 0; i < len(storedUsers); i++ {
			storedUsers[i].NotifyProps = originalProps[i]
		}
	})
}

func TestUserStoreProfilesInChannelCache(t *testing.T) {
	fakeChannelId := "123"
	fakeUserId := "456"
	fakeMap := map[string]*model.User{
		fakeUserId: {Id: "456"},
	}

	t.Run("first call not cached, second cached and returning same data", func(t *testing.T) {
		mockStore := getMockStore()
		mockCacheProvider := getMockCacheProvider()
		cachedStore := NewLocalCacheLayer(mockStore, nil, nil, mockCacheProvider)

		gotMap, err := cachedStore.User().GetAllProfilesInChannel(fakeChannelId, true)
		require.Nil(t, err)
		assert.Equal(t, fakeMap, gotMap)
		mockStore.User().(*mocks.UserStore).AssertNumberOfCalls(t, "GetAllProfilesInChannel", 1)

		_, _ = cachedStore.User().GetAllProfilesInChannel(fakeChannelId, true)
		mockStore.User().(*mocks.UserStore).AssertNumberOfCalls(t, "GetAllProfilesInChannel", 1)
	})

	t.Run("first call not cached, second force not cached", func(t *testing.T) {
		mockStore := getMockStore()
		mockCacheProvider := getMockCacheProvider()
		cachedStore := NewLocalCacheLayer(mockStore, nil, nil, mockCacheProvider)

		gotMap, err := cachedStore.User().GetAllProfilesInChannel(fakeChannelId, true)
		require.Nil(t, err)
		assert.Equal(t, fakeMap, gotMap)
		mockStore.User().(*mocks.UserStore).AssertNumberOfCalls(t, "GetAllProfilesInChannel", 1)

		_, _ = cachedStore.User().GetAllProfilesInChannel(fakeChannelId, false)
		mockStore.User().(*mocks.UserStore).AssertNumberOfCalls(t, "GetAllProfilesInChannel", 2)
	})

	t.Run("first call not cached, invalidate by channel, and then not cached again", func(t *testing.T) {
		mockStore := getMockStore()
		mockCacheProvider := getMockCacheProvider()
		cachedStore := NewLocalCacheLayer(mockStore, nil, nil, mockCacheProvider)

		gotMap, err := cachedStore.User().GetAllProfilesInChannel(fakeChannelId, true)
		require.Nil(t, err)
		assert.Equal(t, fakeMap, gotMap)
		mockStore.User().(*mocks.UserStore).AssertNumberOfCalls(t, "GetAllProfilesInChannel", 1)

		cachedStore.User().InvalidateProfilesInChannelCache("123")

		_, _ = cachedStore.User().GetAllProfilesInChannel(fakeChannelId, true)
		mockStore.User().(*mocks.UserStore).AssertNumberOfCalls(t, "GetAllProfilesInChannel", 2)
	})

	t.Run("first call not cached, invalidate by user, and then not cached again", func(t *testing.T) {
		mockStore := getMockStore()
		mockCacheProvider := getMockCacheProvider()
		cachedStore := NewLocalCacheLayer(mockStore, nil, nil, mockCacheProvider)

		gotMap, err := cachedStore.User().GetAllProfilesInChannel(fakeChannelId, true)
		require.Nil(t, err)
		assert.Equal(t, fakeMap, gotMap)
		mockStore.User().(*mocks.UserStore).AssertNumberOfCalls(t, "GetAllProfilesInChannel", 1)

		cachedStore.User().InvalidateProfilesInChannelCacheByUser("456")

		_, _ = cachedStore.User().GetAllProfilesInChannel(fakeChannelId, true)
		mockStore.User().(*mocks.UserStore).AssertNumberOfCalls(t, "GetAllProfilesInChannel", 2)
	})
}

func TestUserStoreGetCache(t *testing.T) {
	fakeUserId := "123"
	fakeUser := &model.User{
		Id:          "123",
		AuthData:    model.NewString("authData"),
		AuthService: "authService",
	}
	t.Run("first call not cached, second cached and returning same data", func(t *testing.T) {
		mockStore := getMockStore()
		mockCacheProvider := getMockCacheProvider()
		cachedStore := NewLocalCacheLayer(mockStore, nil, nil, mockCacheProvider)

		gotUser, err := cachedStore.User().Get(fakeUserId)
		require.Nil(t, err)
		assert.Equal(t, fakeUser, gotUser)
		mockStore.User().(*mocks.UserStore).AssertNumberOfCalls(t, "Get", 1)

		_, _ = cachedStore.User().Get(fakeUserId)
		mockStore.User().(*mocks.UserStore).AssertNumberOfCalls(t, "Get", 1)
	})

	t.Run("first call not cached, invalidate, and then not cached again", func(t *testing.T) {
		mockStore := getMockStore()
		mockCacheProvider := getMockCacheProvider()
		cachedStore := NewLocalCacheLayer(mockStore, nil, nil, mockCacheProvider)

		gotUser, err := cachedStore.User().Get(fakeUserId)
		require.Nil(t, err)
		assert.Equal(t, fakeUser, gotUser)
		mockStore.User().(*mocks.UserStore).AssertNumberOfCalls(t, "Get", 1)

		cachedStore.User().InvalidateProfileCacheForUser("123")

		_, _ = cachedStore.User().Get(fakeUserId)
		mockStore.User().(*mocks.UserStore).AssertNumberOfCalls(t, "Get", 2)
	})

	t.Run("should always return a copy of the stored data", func(t *testing.T) {
		mockStore := getMockStore()
		mockCacheProvider := getMockCacheProvider()
		cachedStore := NewLocalCacheLayer(mockStore, nil, nil, mockCacheProvider)

		storedUser, err := mockStore.User().Get(fakeUserId)
		require.Nil(t, err)
		originalProps := storedUser.NotifyProps

		storedUser.NotifyProps = map[string]string{}
		storedUser.NotifyProps["key"] = "somevalue"

		cachedUser, err := cachedStore.User().Get(fakeUserId)
		require.Nil(t, err)
		assert.Equal(t, storedUser, cachedUser)

		storedUser.Props = model.StringMap{}
		storedUser.Timezone = model.StringMap{}
		cachedUser, err = cachedStore.User().Get(fakeUserId)
		require.Nil(t, err)
		assert.Equal(t, storedUser, cachedUser)
		if storedUser == cachedUser {
			assert.Fail(t, "should be different pointers")
		}
		cachedUser.NotifyProps["key"] = "othervalue"
		assert.NotEqual(t, storedUser, cachedUser)

		storedUser.NotifyProps = originalProps
	})
}
