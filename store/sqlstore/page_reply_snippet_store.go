// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package sqlstore

import (
	"bitbucket.org/enesyteam/papo-server/model"
	"bitbucket.org/enesyteam/papo-server/store"
	"net/http"
)

type sqlPageReplySnippetStore struct {
	SqlStore
}

func NewSqlPageReplySnippetStore(sqlStore SqlStore) store.PageReplySnippetStore {

	fs := &sqlPageReplySnippetStore{
		SqlStore: sqlStore,
	}

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(model.ReplySnippet{}, "ReplySnippets").SetKeys(false, "Id")
		table.ColMap("Id").SetMaxSize(26)
	}

	return fs
}

func (fs sqlPageReplySnippetStore) CreateIndexesIfNotExists() {
	fs.CreateIndexIfNotExists("idx_reply_snippets_page_id", "ReplySnippets", "PageId")
	fs.CreateIndexIfNotExists("idx_reply_snippets_update_at", "ReplySnippets", "UpdateAt")
	fs.CreateIndexIfNotExists("idx_reply_snippets_create_at", "ReplySnippets", "CreateAt")
	fs.CreateIndexIfNotExists("idx_reply_snippets_delete_at", "ReplySnippets", "DeleteAt")
}

func (fs sqlPageReplySnippetStore) GetByPageId(pageId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var data []*model.ReplySnippet
		if _, err := fs.GetReplica().Select(&data, "SELECT ReplySnippets.* FROM ReplySnippets WHERE ReplySnippets.PageId = :PageId", map[string]interface{}{"PageId": pageId}); err != nil {
			result.Err = model.NewAppError("SqlTeamStore.GetTeamsByUserId", "store.sql_team.get_all.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		}
		result.Data = data
	})
}

// kiểm tra nếu 1 snippet có trigger là "abc" của 1 page đã được tạo hay chưa
func (fs sqlPageReplySnippetStore) CheckDuplicate(pageId string, trigger string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var data []*model.ReplySnippet
		if _, err := fs.GetReplica().Select(&data, "SELECT ReplySnippets.* FROM ReplySnippets WHERE ReplySnippets.PageId = :PageId AND ReplySnippets.Trigger = :Trigger", map[string]interface{}{"PageId": pageId, "Trigger": trigger}); err != nil {
			result.Err = model.NewAppError("SqlTeamStore.GetTeamsByUserId", "store.sql_team.get_all.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		} else {
			result.Data = data
		}
	})
}

func (fs sqlPageReplySnippetStore) Save(snippet *model.ReplySnippet) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		snippet.PreSave()

		if err := fs.GetMaster().Insert(snippet); err != nil {
			result.Err = model.NewAppError("sqlPageReplySnippetStore.Save", "store.sql_fanpage.save.app_error", nil, "page_id="+snippet.Id+", "+err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = snippet
		}
	})
}

func (s sqlPageReplySnippetStore) Update(snippet *model.ReplySnippet) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		snippet.PreUpdate()

		oldResult, err := s.GetMaster().Get(model.ReplySnippet{}, snippet.Id)
		if err != nil {
			result.Err = model.NewAppError("sqlPageReplySnippetStore.Update", "store.sql_team.update.finding.app_error", nil, "id="+snippet.Id+", "+err.Error(), http.StatusInternalServerError)
			return
		}

		if oldResult == nil {
			result.Err = model.NewAppError("sqlPageReplySnippetStore.Update", "store.sql_team.update.find.app_error", nil, "id="+snippet.Id, http.StatusBadRequest)
			return
		}

		oldSnippet := oldResult.(*model.ReplySnippet)
		snippet.CreateAt = oldSnippet.CreateAt
		snippet.UpdateAt = model.GetMillis()

		count, err := s.GetMaster().Update(snippet)
		if err != nil {
			result.Err = model.NewAppError("SqlTeamStore.Update", "store.sql_team.update.updating.app_error", nil, "id="+snippet.Id+", "+err.Error(), http.StatusInternalServerError)
			return
		}
		if count != 1 {
			result.Err = model.NewAppError("SqlTeamStore.Update", "store.sql_team.update.app_error", nil, "id="+snippet.Id, http.StatusInternalServerError)
			return
		}
		result.Data = snippet
	})
}