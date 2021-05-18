// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package sqlstore

import (
	"bitbucket.org/enesyteam/papo-server/model"
	"bitbucket.org/enesyteam/papo-server/store"
	"net/http"
)

type sqlAutoMessageTaskStore struct {
	SqlStore
}

func NewSqlAutoMessageTaskStore(sqlStore SqlStore) store.AutoMessageTaskStore {

	fs := &sqlAutoMessageTaskStore{
		SqlStore: sqlStore,
	}

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(model.AutoMessageTask{}, "AutoMessageTasks").SetKeys(false, "Id")
		table.ColMap("Id").SetMaxSize(26)
	}

	return fs
}

func (fs sqlAutoMessageTaskStore) CreateIndexesIfNotExists() {
	fs.CreateIndexIfNotExists("idx_auto_message_tasks_page_id", "AutoMessageTasks", "PageId")
	fs.CreateIndexIfNotExists("idx_auto_message_tasks_creator", "AutoMessageTasks", "Creator")
	fs.CreateIndexIfNotExists("idx_auto_message_tasks_start_at", "AutoMessageTasks", "StartAt")
	fs.CreateIndexIfNotExists("idx_auto_message_tasks_end_at", "AutoMessageTasks", "EndAt")
	fs.CreateIndexIfNotExists("idx_auto_message_tasks_update_at", "AutoMessageTasks", "UpdateAt")
	fs.CreateIndexIfNotExists("idx_auto_message_tasks_create_at", "AutoMessageTasks", "CreateAt")
	fs.CreateIndexIfNotExists("idx_auto_message_tasks_delete_at", "AutoMessageTasks", "DeleteAt")
}

func (fs sqlAutoMessageTaskStore) GetByPageId(pageId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var data []*model.AutoMessageTask
		if _, err := fs.GetReplica().Select(&data, "SELECT AutoMessageTasks.* FROM AutoMessageTasks WHERE AutoMessageTasks.PageId = :PageId", map[string]interface{}{"PageId": pageId}); err != nil {
			result.Err = model.NewAppError("sqlAutoMessageTaskStore.GetByPageId", "store.sql_team.get_all.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		}
		result.Data = data
	})
}

func (fs sqlAutoMessageTaskStore) Save(task *model.AutoMessageTask) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		task.PreSave()

		if err := fs.GetMaster().Insert(task); err != nil {
			result.Err = model.NewAppError("sqlAutoMessageTaskStore.Save", "store.sql_fanpage.save.app_error", nil, "page_id="+task.Id+", "+err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = task
		}
	})
}

func (s sqlAutoMessageTaskStore) Update(task *model.AutoMessageTask) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		task.PreUpdate()

		oldResult, err := s.GetMaster().Get(model.AutoMessageTask{}, task.Id)
		if err != nil {
			result.Err = model.NewAppError("sqlAutoMessageTaskStore.Update", "store.sql_team.update.finding.app_error", nil, "id="+task.Id+", "+err.Error(), http.StatusInternalServerError)
			return
		}

		if oldResult == nil {
			result.Err = model.NewAppError("sqlPageReplySnippetStore.Update", "store.sql_team.update.find.app_error", nil, "id="+task.Id, http.StatusBadRequest)
			return
		}

		oldTask := oldResult.(*model.AutoMessageTask)
		task.CreateAt = oldTask.CreateAt
		task.UpdateAt = model.GetMillis()

		count, err := s.GetMaster().Update(task)
		if err != nil {
			result.Err = model.NewAppError("sqlAutoMessageTaskStore.Update", "store.sql_team.update.updating.app_error", nil, "id="+task.Id+", "+err.Error(), http.StatusInternalServerError)
			return
		}
		if count != 1 {
			result.Err = model.NewAppError("sqlAutoMessageTaskStore.Update", "store.sql_team.update.app_error", nil, "id="+task.Id, http.StatusInternalServerError)
			return
		}
		result.Data = task
	})
}

func (s sqlAutoMessageTaskStore) Delete(taskId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

	})
}

func (s sqlAutoMessageTaskStore) Start(taskId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

	})
}

func (s sqlAutoMessageTaskStore) Pause(taskId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

	})
}
