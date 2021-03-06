// Code generated by mockery v1.0.0. DO NOT EDIT.

// Regenerate this file using `make store-mocks`.

package mocks

import (
	model "bitbucket.org/enesyteam/papo-server/model"
	mock "github.com/stretchr/testify/mock"
)

// SessionStore is an autogenerated mock type for the SessionStore type
type SessionStore struct {
	mock.Mock
}

// AnalyticsSessionCount provides a mock function with given fields:
func (_m *SessionStore) AnalyticsSessionCount() (int64, error) {
	ret := _m.Called()

	var r0 int64
	if rf, ok := ret.Get(0).(func() int64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int64)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Cleanup provides a mock function with given fields: expiryTime, batchSize
func (_m *SessionStore) Cleanup(expiryTime int64, batchSize int64) {
	_m.Called(expiryTime, batchSize)
}

// Get provides a mock function with given fields: sessionIdOrToken
func (_m *SessionStore) Get(sessionIdOrToken string) (*model.Session, error) {
	ret := _m.Called(sessionIdOrToken)

	var r0 *model.Session
	if rf, ok := ret.Get(0).(func(string) *model.Session); ok {
		r0 = rf(sessionIdOrToken)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Session)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(sessionIdOrToken)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetSessions provides a mock function with given fields: userId
func (_m *SessionStore) GetSessions(userId string) ([]*model.Session, error) {
	ret := _m.Called(userId)

	var r0 []*model.Session
	if rf, ok := ret.Get(0).(func(string) []*model.Session); ok {
		r0 = rf(userId)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Session)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(userId)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetSessionsExpired provides a mock function with given fields: thresholdMillis, mobileOnly, unnotifiedOnly
func (_m *SessionStore) GetSessionsExpired(thresholdMillis int64, mobileOnly bool, unnotifiedOnly bool) ([]*model.Session, error) {
	ret := _m.Called(thresholdMillis, mobileOnly, unnotifiedOnly)

	var r0 []*model.Session
	if rf, ok := ret.Get(0).(func(int64, bool, bool) []*model.Session); ok {
		r0 = rf(thresholdMillis, mobileOnly, unnotifiedOnly)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Session)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(int64, bool, bool) error); ok {
		r1 = rf(thresholdMillis, mobileOnly, unnotifiedOnly)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetSessionsWithActiveDeviceIds provides a mock function with given fields: userId
func (_m *SessionStore) GetSessionsWithActiveDeviceIds(userId string) ([]*model.Session, error) {
	ret := _m.Called(userId)

	var r0 []*model.Session
	if rf, ok := ret.Get(0).(func(string) []*model.Session); ok {
		r0 = rf(userId)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Session)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(userId)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PermanentDeleteSessionsByUser provides a mock function with given fields: teamId
func (_m *SessionStore) PermanentDeleteSessionsByUser(teamId string) error {
	ret := _m.Called(teamId)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(teamId)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Remove provides a mock function with given fields: sessionIdOrToken
func (_m *SessionStore) Remove(sessionIdOrToken string) error {
	ret := _m.Called(sessionIdOrToken)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(sessionIdOrToken)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RemoveAllSessions provides a mock function with given fields:
func (_m *SessionStore) RemoveAllSessions() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Save provides a mock function with given fields: session
func (_m *SessionStore) Save(session *model.Session) (*model.Session, error) {
	ret := _m.Called(session)

	var r0 *model.Session
	if rf, ok := ret.Get(0).(func(*model.Session) *model.Session); ok {
		r0 = rf(session)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Session)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*model.Session) error); ok {
		r1 = rf(session)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateDeviceId provides a mock function with given fields: id, deviceId, expiresAt
func (_m *SessionStore) UpdateDeviceId(id string, deviceId string, expiresAt int64) (string, error) {
	ret := _m.Called(id, deviceId, expiresAt)

	var r0 string
	if rf, ok := ret.Get(0).(func(string, string, int64) string); ok {
		r0 = rf(id, deviceId, expiresAt)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, int64) error); ok {
		r1 = rf(id, deviceId, expiresAt)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateExpiredNotify provides a mock function with given fields: sessionid, notified
func (_m *SessionStore) UpdateExpiredNotify(sessionid string, notified bool) error {
	ret := _m.Called(sessionid, notified)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, bool) error); ok {
		r0 = rf(sessionid, notified)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateExpiresAt provides a mock function with given fields: sessionId, time
func (_m *SessionStore) UpdateExpiresAt(sessionId string, time int64) error {
	ret := _m.Called(sessionId, time)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, int64) error); ok {
		r0 = rf(sessionId, time)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateLastActivityAt provides a mock function with given fields: sessionId, time
func (_m *SessionStore) UpdateLastActivityAt(sessionId string, time int64) error {
	ret := _m.Called(sessionId, time)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, int64) error); ok {
		r0 = rf(sessionId, time)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateProps provides a mock function with given fields: session
func (_m *SessionStore) UpdateProps(session *model.Session) error {
	ret := _m.Called(session)

	var r0 error
	if rf, ok := ret.Get(0).(func(*model.Session) error); ok {
		r0 = rf(session)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateRoles provides a mock function with given fields: userId, roles
func (_m *SessionStore) UpdateRoles(userId string, roles string) (string, error) {
	ret := _m.Called(userId, roles)

	var r0 string
	if rf, ok := ret.Get(0).(func(string, string) string); ok {
		r0 = rf(userId, roles)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(userId, roles)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
