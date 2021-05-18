// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package sqlstore

import (
	"bitbucket.org/enesyteam/papo-server/model"
	"bitbucket.org/enesyteam/papo-server/store"
	"net/http"
)

type sqlPageTagStore struct {
	SqlStore
}

func NewSqlPageTagStore(sqlStore SqlStore) store.PageTagStore {
	fs := &sqlPageTagStore{
		SqlStore: sqlStore,
	}

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(model.PageTag{}, "PageTags").SetKeys(false, "Id")
		table.ColMap("Id").SetMaxSize(26)
		table.ColMap("PageId").SetMaxSize(26)
		table.ColMap("Creator").SetMaxSize(26)
		table.ColMap("Name").SetMaxSize(64)
	}

	return fs
}

func (fs sqlPageTagStore) CreateIndexesIfNotExists() {
	fs.CreateIndexIfNotExists("idx_page_tags_name_unique", "PageTags", "Name")
	fs.CreateIndexIfNotExists("idx_page_tags_page_id", "PageTags", "PageId")
	fs.CreateIndexIfNotExists("idx_page_tags_creator", "PageTags", "Creator")
	fs.CreateIndexIfNotExists("idx_page_tags_update_at", "PageTags", "UpdateAt")
	fs.CreateIndexIfNotExists("idx_page_tags_create_at", "PageTags", "CreateAt")
	fs.CreateIndexIfNotExists("idx_page_tags_delete_at", "PageTags", "DeleteAt")
}

func (fs sqlPageTagStore) Get(id string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var tag model.PageTag

		if err := fs.GetReplica().SelectOne(&tag, "SELECT PageTags.* FROM PageTags WHERE id = :Id", map[string]interface{}{"Id": id}); err != nil {
			result.Err = model.NewAppError("sqlPageTagStore.Get", "store.sql_team.get_all.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		}
		result.Data = tag
	})
}

func (fs sqlPageTagStore) GetPageTags(pageId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var tags []*model.PageTag

		if _, err := fs.GetReplica().Select(&tags, "SELECT PageTags.* FROM PageTags WHERE pageId = :PageId", map[string]interface{}{"PageId": pageId}); err != nil {
			result.Err = model.NewAppError("sqlPageTagStore.Get", "store.sql_team.get_all.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		}
		result.Data = tags
	})
}

func (fs sqlPageTagStore) Save(tag *model.PageTag) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		tag.PreSave()

		if err := fs.GetMaster().Insert(tag); err != nil {
			if IsUniqueConstraintError(err, []string{"PageTags", "page_tag_name_key", "idx_page_tags_name_unique"}) {
				result.Err = model.NewAppError("sqlPageTagStore.Save", "store.sql_fanpage.save.id_exists.app_error", nil, "tag_id="+tag.Id+", "+err.Error(), http.StatusBadRequest)
			} else {
				// Một lỗi xảy ra và không biết nó là lỗi gì
				result.Err = model.NewAppError("sqlPageTagStore.Save", "store.sql_fanpage.save.app_error", nil, "id="+tag.Id+", "+err.Error(), http.StatusInternalServerError)
			}
		} else {
			result.Data = tag
		}
	})
}

func (s sqlPageTagStore) Update(tag *model.PageTag) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		tag.PreUpdate()

		oldResult, err := s.GetMaster().Get(model.PageTag{}, tag.Id)
		if err != nil {
			result.Err = model.NewAppError("sqlPageTagStore.Update", "store.sql_team.update.finding.app_error", nil, "id="+tag.Id+", "+err.Error(), http.StatusInternalServerError)
			return
		}

		if oldResult == nil {
			result.Err = model.NewAppError("sqlPageTagStore.Update", "store.sql_team.update.find.app_error", nil, "id="+tag.Id, http.StatusBadRequest)
			return
		}

		oldTag := oldResult.(*model.PageTag)
		tag.CreateAt = oldTag.CreateAt
		tag.UpdateAt = model.GetMillis()

		if len(tag.Name) == 0 {
			tag.Name = oldTag.Name
		}

		if len(tag.Color) == 0 {
			tag.Color = oldTag.Color
		}

		count, err := s.GetMaster().Update(tag)
		if err != nil {
			result.Err = model.NewAppError("sqlPageTagStore.Update", "store.sql_team.update.updating.app_error", nil, "id="+tag.Id+", "+err.Error(), http.StatusInternalServerError)
			return
		}
		if count != 1 {
			result.Err = model.NewAppError("sqlPageTagStore.Update", "store.sql_team.update.app_error", nil, "id="+tag.Id, http.StatusInternalServerError)
			return
		}
		result.Data = tag
	})
}

func (fs sqlPageTagStore) Delete(id string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		if _, err := fs.GetMaster().Exec("DELETE FROM PageTags WHERE Id = :Id", map[string]interface{}{"Id": id}); err != nil {
			result.Err = model.NewAppError("sqlPageTagStore.Delete", "store.sql_user.permanent_delete.app_error", nil, "id="+id+", "+err.Error(), http.StatusInternalServerError)
		}
	})
}