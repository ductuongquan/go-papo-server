// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package sqlstore

import (
	"bitbucket.org/enesyteam/papo-server/model"
	"bitbucket.org/enesyteam/papo-server/store"

	"github.com/pkg/errors"
)

type SqlComplianceStore struct {
	SqlStore
}

func newSqlComplianceStore(sqlStore SqlStore) store.ComplianceStore {
	s := &SqlComplianceStore{sqlStore}

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(model.Compliance{}, "Compliances").SetKeys(false, "Id")
		table.ColMap("Id").SetMaxSize(26)
		table.ColMap("UserId").SetMaxSize(26)
		table.ColMap("Status").SetMaxSize(64)
		table.ColMap("Desc").SetMaxSize(512)
		table.ColMap("Type").SetMaxSize(64)
		table.ColMap("Keywords").SetMaxSize(512)
		table.ColMap("Emails").SetMaxSize(1024)
	}

	return s
}

func (s SqlComplianceStore) createIndexesIfNotExists() {
}

func (s SqlComplianceStore) Save(compliance *model.Compliance) (*model.Compliance, error) {
	compliance.PreSave()
	if err := compliance.IsValid(); err != nil {
		return nil, err
	}

	if err := s.GetMaster().Insert(compliance); err != nil {
		return nil, errors.Wrap(err, "failed to save Compliance")
	}
	return compliance, nil
}

func (s SqlComplianceStore) Update(compliance *model.Compliance) (*model.Compliance, error) {
	if err := compliance.IsValid(); err != nil {
		return nil, err
	}

	if _, err := s.GetMaster().Update(compliance); err != nil {
		return nil, errors.Wrap(err, "failed to update Compliance")
	}
	return compliance, nil
}

func (s SqlComplianceStore) GetAll(offset, limit int) (model.Compliances, error) {
	query := "SELECT * FROM Compliances ORDER BY CreateAt DESC LIMIT :Limit OFFSET :Offset"

	var compliances model.Compliances
	if _, err := s.GetReplica().Select(&compliances, query, map[string]interface{}{"Offset": offset, "Limit": limit}); err != nil {
		return nil, errors.Wrap(err, "failed to find all Compliances")
	}
	return compliances, nil
}

func (s SqlComplianceStore) Get(id string) (*model.Compliance, error) {
	obj, err := s.GetReplica().Get(model.Compliance{}, id)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get Compliance with id=%s", id)
	}
	if obj == nil {
		return nil, store.NewErrNotFound("Compliance", id)
	}
	return obj.(*model.Compliance), nil
}
