package sqlstore

import (
	"bitbucket.org/enesyteam/papo-server/mlog"
	"bitbucket.org/enesyteam/papo-server/model"
	"bitbucket.org/enesyteam/papo-server/store"
	"database/sql"
	"fmt"
	"net/http"
)

type SqlFanpageInitResultStore struct {
	SqlStore
}

func NewSqlFanpageInitResultStore(sqlStore SqlStore) store.FanpageInitResultStore {
	s := &SqlFanpageInitResultStore{sqlStore}

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(model.FanpageInitResult{}, "FanpageInitResults").SetKeys(false, "Id", "PageId")
		table.ColMap("Id").SetMaxSize(26)
		table.ColMap("Creator").SetMaxSize(26)
	}

	return s
}

func (fs SqlFanpageInitResultStore) Save(r *model.FanpageInitResult) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		r.PreSave()

		if result.Err != nil {
			return
		}

		if err := fs.GetMaster().Insert(r); err != nil {
			result.Err = model.NewAppError("SqlFanpageInitResultStore.Save", "store.SqlFanpageInitResultStore.save.app_error", nil, "id="+r.Id+", "+err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = r
		}
	})
}

func (fs SqlFanpageInitResultStore) UpdateConversationCount(r *model.FanpageInitResult, newCount int64) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		if _, err := fs.GetMaster().Exec("UPDATE FanpageInitResults SET ConversationCount = ConversationCount + :NewCount WHERE PageId = :PageId RETURNING *", map[string]interface{}{"PageId": r.PageId, "NewCount": newCount}); err != nil {
			result.Err = model.NewAppError("SqlFanpageInitResultStore.UpdateConversationCount", "store.SqlFanpageInitResultStore.UpdateConversationCount.app_error", nil, "page_id="+r.PageId, http.StatusInternalServerError)
		} else {
			var initResult *model.FanpageInitResult
			query := "SELECT a.* FROM FanpageInitResults a WHERE a.pageid = :PageId"
			err := fs.GetReplica().SelectOne(&initResult, query, map[string]interface{}{"PageId": r.PageId})
			if err != nil {
				if err == sql.ErrNoRows {
					result.Err = model.NewAppError("SqlFanpageInitResultStore.UpdateMessageCount", "store.SqlFanpageInitResultStore.UpdateMessageCount.app_error", nil, "page_id="+r.PageId, http.StatusInternalServerError)
				} else {
					mlog.Error(fmt.Sprintf("Couldn't find page init result with id, err=%v", result.Err))
					result.Err = model.NewAppError("SqlFanpageInitResultStore.UpdateMessageCount", "store.SqlFanpageInitResultStore.UpdateMessageCount.app_error", nil, "page_id="+r.PageId, http.StatusInternalServerError)
				}
			} else {
				result.Data = initResult
			}
		}
	})
}

func (fs SqlFanpageInitResultStore) UpdatePostCount(r *model.FanpageInitResult, newCount int64) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		if _, err := fs.GetMaster().Exec("UPDATE FanpageInitResults SET PostCount = PostCount + :NewCount WHERE PageId = :PageId RETURNING *", map[string]interface{}{"PageId": r.PageId, "NewCount": newCount}); err != nil {
			result.Err = model.NewAppError("SqlFanpageInitResultStore.UpdatePostCount", "store.SqlFanpageInitResultStore.UpdatePostCount.app_error", nil, "page_id="+r.PageId, http.StatusInternalServerError)
		} else {
			var initResult *model.FanpageInitResult
			query := "SELECT a.* FROM FanpageInitResults a WHERE a.pageid = :PageId"
			err := fs.GetReplica().SelectOne(&initResult, query, map[string]interface{}{"PageId": r.PageId})
			if err != nil {
				if err == sql.ErrNoRows {
					result.Err = model.NewAppError("SqlFanpageInitResultStore.UpdateMessageCount", "store.SqlFanpageInitResultStore.UpdateMessageCount.app_error", nil, "page_id="+r.PageId, http.StatusInternalServerError)
				} else {
					mlog.Error(fmt.Sprintf("Couldn't find page init result with id, err=%v", result.Err))
					result.Err = model.NewAppError("SqlFanpageInitResultStore.UpdateMessageCount", "store.SqlFanpageInitResultStore.UpdateMessageCount.app_error", nil, "page_id="+r.PageId, http.StatusInternalServerError)
				}
			} else {
				result.Data = initResult
			}
		}
	})
}

func (fs SqlFanpageInitResultStore) UpdateMessageCount(r *model.FanpageInitResult, newCount int64) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		if _, err := fs.GetMaster().Exec("UPDATE FanpageInitResults SET MessageCount = MessageCount + :NewCount WHERE PageId = :PageId RETURNING *", map[string]interface{}{"PageId": r.PageId, "NewCount": newCount}); err != nil {
			result.Err = model.NewAppError("SqlFanpageInitResultStore.UpdateMessageCount", "store.SqlFanpageInitResultStore.UpdateMessageCount.app_error", nil, "page_id="+r.PageId, http.StatusInternalServerError)
		} else {
			var initResult *model.FanpageInitResult
			query := "SELECT a.* FROM FanpageInitResults a WHERE a.pageid = :PageId"
			err := fs.GetReplica().SelectOne(&initResult, query, map[string]interface{}{"PageId": r.PageId})
			if err != nil {
				if err == sql.ErrNoRows {
					result.Err = model.NewAppError("SqlFanpageInitResultStore.UpdateMessageCount", "store.SqlFanpageInitResultStore.UpdateMessageCount.app_error", nil, "page_id="+r.PageId, http.StatusInternalServerError)
				} else {
					mlog.Error(fmt.Sprintf("Couldn't find page init result with id, err=%v", result.Err))
					result.Err = model.NewAppError("SqlFanpageInitResultStore.UpdateMessageCount", "store.SqlFanpageInitResultStore.UpdateMessageCount.app_error", nil, "page_id="+r.PageId, http.StatusInternalServerError)
				}
			} else {
				result.Data = initResult
			}
		}
	})
}

func (fs SqlFanpageInitResultStore) UpdateCommentCount(r *model.FanpageInitResult, newCount int64) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		if _, err := fs.GetMaster().Exec("UPDATE FanpageInitResults SET CommentCount = CommentCount + :NewCount WHERE PageId = :PageId RETURNING *", map[string]interface{}{"PageId": r.PageId, "NewCount": newCount}); err != nil {
			result.Err = model.NewAppError("SqlFanpageInitResultStore.UpdateMessageCount", "store.SqlFanpageInitResultStore.UpdateMessageCount.app_error", nil, "page_id="+r.PageId, http.StatusInternalServerError)
		} else {
			var initResult *model.FanpageInitResult
			query := "SELECT a.* FROM FanpageInitResults a WHERE a.pageid = :PageId"
			err := fs.GetReplica().SelectOne(&initResult, query, map[string]interface{}{"PageId": r.PageId})
			if err != nil {
				if err == sql.ErrNoRows {
					result.Err = model.NewAppError("SqlFanpageInitResultStore.UpdateMessageCount", "store.SqlFanpageInitResultStore.UpdateMessageCount.app_error", nil, "page_id="+r.PageId, http.StatusInternalServerError)
				} else {
					mlog.Error(fmt.Sprintf("Couldn't find page init result with id, err=%v", result.Err))
					result.Err = model.NewAppError("SqlFanpageInitResultStore.UpdateMessageCount", "store.SqlFanpageInitResultStore.UpdateMessageCount.app_error", nil, "page_id="+r.PageId, http.StatusInternalServerError)
				}
			} else {
				result.Data = initResult
			}
		}
	})
}

func (fs SqlFanpageInitResultStore) UpdateEndAt(r *model.FanpageInitResult, endAt int64) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		if _, err := fs.GetMaster().Exec("UPDATE FanpageInitResults SET EndAt = :NewCount WHERE PageId = :PageId RETURNING *", map[string]interface{}{"PageId": r.PageId, "EndAt": endAt}); err != nil {
			result.Err = model.NewAppError("SqlFanpageInitResultStore.UpdateMessageCount", "store.SqlFanpageInitResultStore.UpdateMessageCount.app_error", nil, "page_id="+r.PageId, http.StatusInternalServerError)
		} else {
			var initResult *model.FanpageInitResult
			query := "SELECT a.* FROM FanpageInitResults a WHERE a.pageid = :PageId"
			err := fs.GetReplica().SelectOne(&initResult, query, map[string]interface{}{"PageId": r.PageId})
			if err != nil {
				if err == sql.ErrNoRows {
					result.Err = model.NewAppError("SqlFanpageInitResultStore.UpdateMessageCount", "store.SqlFanpageInitResultStore.UpdateMessageCount.app_error", nil, "page_id="+r.PageId, http.StatusInternalServerError)
				} else {
					mlog.Error(fmt.Sprintf("Couldn't find page init result with id, err=%v", result.Err))
					result.Err = model.NewAppError("SqlFanpageInitResultStore.UpdateMessageCount", "store.SqlFanpageInitResultStore.UpdateMessageCount.app_error", nil, "page_id="+r.PageId, http.StatusInternalServerError)
				}
			} else {
				result.Data = initResult
			}
		}
	})
}