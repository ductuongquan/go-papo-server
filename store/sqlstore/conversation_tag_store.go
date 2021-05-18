// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package sqlstore

import (
	"bitbucket.org/enesyteam/papo-server/model"
	"bitbucket.org/enesyteam/papo-server/store"
	"net/http"
)

type sqlConversationTagStore struct {
	SqlStore
}

func NewSqlConversationTagStore(sqlStore SqlStore) store.ConversationTagStore {
	fs := &sqlConversationTagStore{
		SqlStore: sqlStore,
	}

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(model.ConversationTag{}, "ConversationTags").SetKeys(false, "Id")
		table.ColMap("Id").SetMaxSize(26)
		table.ColMap("ConversationId").SetMaxSize(26)
		table.ColMap("TagId").SetMaxSize(26)
	}

	return fs
}

func (fs sqlConversationTagStore) CreateIndexesIfNotExists() {
	fs.CreateIndexIfNotExists("idx_conversation_tags_conversation_id", "ConversationTags", "ConversationId")
	fs.CreateIndexIfNotExists("idx_conversation_tags_tag_id", "ConversationTags", "TagId")
	fs.CreateIndexIfNotExists("idx_conversation_tags_creator", "ConversationTags", "Creator")
	fs.CreateIndexIfNotExists("idx_conversation_tags_create_at", "ConversationTags", "CreateAt")
}

func (fs sqlConversationTagStore) Get(id string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var tag model.ConversationTag

		if err := fs.GetReplica().SelectOne(&tag, "SELECT ConversationTags.* FROM ConversationTags WHERE id = :Id", map[string]interface{}{"Id": id}); err != nil {
			result.Err = model.NewAppError("sqlConversationTagStore.Get", "store.sql_team.get_all.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		}
		result.Data = tag
	})
}

func (fs sqlConversationTagStore) GetConversationTags(conversationId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var tags []*model.ConversationTag

		if _, err := fs.GetReplica().Select(&tags, "SELECT ConversationTags.* FROM ConversationTags WHERE ConversationId = :ConversationId", map[string]interface{}{"ConversationId": conversationId}); err != nil {
			result.Err = model.NewAppError("sqlConversationTagStore.Get", "store.sql_team.get_all.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		}
		result.Data = tags
	})
}

func (fs sqlConversationTagStore) Save(tag *model.ConversationTag) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		tag.PreSave()

		if err := fs.GetMaster().Insert(tag); err != nil {
			result.Err = model.NewAppError("sqlConversationTagStore.Save", "store.sql_fanpage.save.app_error", nil, "id="+tag.Id+", "+err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = tag
		}
	})
}

func (fs sqlConversationTagStore) SaveOrRemove(tag *model.ConversationTag) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var found []*model.ConversationTag
		if _, err := fs.GetReplica().Select(&found, "SELECT ConversationTags.* FROM ConversationTags WHERE ConversationTags.tagid = :TagId AND ConversationTags.conversationid = :ConversationId", map[string]interface{}{"TagId": tag.TagId, "ConversationId": tag.ConversationId}); err != nil {
			result.Err = model.NewAppError("sqlConversationTagStore.SaveOrRemove", "store.sql_user.permanent_delete.app_error", nil, "tag_id="+tag.TagId+", "+err.Error(), http.StatusInternalServerError)
		} else {
			// this tag has not attached to conversation, just add it
			if len(found) == 0 {
				tag.PreSave()

				if err := fs.GetMaster().Insert(tag); err != nil {
					result.Err = model.NewAppError("sqlConversationTagStore.Save", "store.sql_fanpage.save.app_error", nil, "tag_id="+tag.TagId+", "+err.Error(), http.StatusInternalServerError)
				} else {
					result.Data = tag
				}
			} else {
				// this tag has attached to conversation, we need to remove tag from conversation
				if _, err := fs.GetMaster().Exec("DELETE FROM ConversationTags WHERE tagid = :TagId", map[string]interface{}{"TagId": tag.TagId}); err != nil {
					result.Err = model.NewAppError("sqlConversationTagStore.Delete", "store.sql_user.permanent_delete.app_error", nil, "tag_id="+tag.TagId+", "+err.Error(), http.StatusInternalServerError)
				}
				result.Data = nil

			}
		}
	})
}

func (fs sqlConversationTagStore) Delete(id string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		if _, err := fs.GetMaster().Exec("DELETE FROM ConversationTags WHERE Id = :Id", map[string]interface{}{"Id": id}); err != nil {
			result.Err = model.NewAppError("sqlConversationTagStore.Delete", "store.sql_user.permanent_delete.app_error", nil, "id="+id+", "+err.Error(), http.StatusInternalServerError)
		}
	})
}