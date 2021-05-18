// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package sqlstore

import (
	"bitbucket.org/enesyteam/papo-server/model"
	"bitbucket.org/enesyteam/papo-server/store"
	"net/http"
)

type sqlConversationNoteStore struct {
	SqlStore
}

func NewSqlConversationNoteStore(sqlStore SqlStore) store.ConversationNoteStore {
	fs := &sqlConversationNoteStore{
		SqlStore: sqlStore,
	}

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(model.ConversationNote{}, "ConversationNotes").SetKeys(false, "Id")
		table.ColMap("Id").SetMaxSize(26)
		table.ColMap("ConversationId").SetMaxSize(26)
		table.ColMap("Creator").SetMaxSize(26)
		table.ColMap("Message").SetMaxSize(1000)
	}

	return fs
}

func (fs sqlConversationNoteStore) CreateIndexesIfNotExists() {
	fs.CreateIndexIfNotExists("idx_conversation_notes_conversation_id", "ConversationNotes", "ConversationId")
	fs.CreateIndexIfNotExists("idx_conversation_notes_creator", "ConversationNotes", "Creator")
	fs.CreateIndexIfNotExists("idx_conversation_notes_update_at", "ConversationNotes", "UpdateAt")
	fs.CreateIndexIfNotExists("idx_conversation_notes_create_at", "ConversationNotes", "CreateAt")
	fs.CreateIndexIfNotExists("idx_conversation_notes_delete_at", "ConversationNotes", "DeleteAt")
}

func (fs sqlConversationNoteStore) Get(id string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var tag model.ConversationNote

		if err := fs.GetReplica().SelectOne(&tag, "SELECT ConversationNotes.* FROM ConversationNotes WHERE id = :Id", map[string]interface{}{"Id": id}); err != nil {
			result.Err = model.NewAppError("sqlConversationNoteStore.Get", "store.sql_team.get_all.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		}
		result.Data = tag
	})
}

func (fs sqlConversationNoteStore) GetConversationNotes(conversationId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var tags []*model.ConversationNote

		if _, err := fs.GetReplica().Select(&tags, "SELECT ConversationNotes.* FROM ConversationNotes WHERE ConversationId = :ConversationId", map[string]interface{}{"ConversationId": conversationId}); err != nil {
			result.Err = model.NewAppError("sqlConversationNoteStore.Get", "store.sql_team.get_all.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		}
		result.Data = tags
	})
}

func (fs sqlConversationNoteStore) Save(note *model.ConversationNote) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		note.PreSave()

		if err := fs.GetMaster().Insert(note); err != nil {
			result.Err = model.NewAppError("sqlConversationNoteStore.Save", "store.sql_fanpage.save.app_error", nil, "id="+note.Id+", "+err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = note
		}
	})
}

func (s sqlConversationNoteStore) Update(note *model.ConversationNote) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		note.PreUpdate()

		oldResult, err := s.GetMaster().Get(model.ConversationNote{}, note.Id)
		if err != nil {
			result.Err = model.NewAppError("sqlConversationNoteStore.Update", "store.sql_team.update.finding.app_error", nil, "id="+note.Id+", "+err.Error(), http.StatusInternalServerError)
			return
		}

		if oldResult == nil {
			result.Err = model.NewAppError("sqlConversationNoteStore.Update", "store.sql_team.update.find.app_error", nil, "id="+note.Id, http.StatusBadRequest)
			return
		}

		oldNote := oldResult.(*model.ConversationNote)
		note.CreateAt = oldNote.CreateAt
		note.UpdateAt = model.GetMillis()

		count, err := s.GetMaster().Update(note)
		if err != nil {
			result.Err = model.NewAppError("sqlConversationNoteStore.Update", "store.sql_team.update.updating.app_error", nil, "id="+note.Id+", "+err.Error(), http.StatusInternalServerError)
			return
		}
		if count != 1 {
			result.Err = model.NewAppError("sqlConversationNoteStore.Update", "store.sql_team.update.app_error", nil, "id="+note.Id, http.StatusInternalServerError)
			return
		}
		result.Data = note
	})
}

func (fs sqlConversationNoteStore) Delete(id string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		if _, err := fs.GetMaster().Exec("DELETE FROM ConversationNotes WHERE Id = :Id", map[string]interface{}{"Id": id}); err != nil {
			result.Err = model.NewAppError("sqlConversationNoteStore.Delete", "store.sql_user.permanent_delete.app_error", nil, "id="+id+", "+err.Error(), http.StatusInternalServerError)
		}
	})
}