package sqlstore

import (
	"bitbucket.org/enesyteam/papo-server/einterfaces"
	"bitbucket.org/enesyteam/papo-server/facebook_graph"
	"bitbucket.org/enesyteam/papo-server/model"
	"bitbucket.org/enesyteam/papo-server/store"
	"bitbucket.org/enesyteam/papo-server/utils"
	"database/sql"
	sq "github.com/Masterminds/squirrel"
	"net/http"
)

const (
	SU_PROFILES_IN_CHANNEL_CACHE_SIZE = model.CHANNEL_CACHE_SIZE
	SU_PROFILES_IN_CHANNEL_CACHE_SEC  = 900 // 15 mins
	SU_PROFILE_BY_IDS_CACHE_SIZE      = model.SESSION_CACHE_SIZE
	SU_PROFILE_BY_IDS_CACHE_SEC       = 900 // 15 mins
)

type sqlFacebookUidStore struct {
	SqlStore
	metrics einterfaces.MetricsInterface

	// usersQuery is a starting point for all queries that return one or more Users.
	usersQuery sq.SelectBuilder
}

var suProfilesInChannelCache *utils.Cache = utils.NewLru(SU_PROFILES_IN_CHANNEL_CACHE_SIZE)
var suProfileByIdsCache *utils.Cache = utils.NewLru(SU_PROFILE_BY_IDS_CACHE_SIZE)

func NewSqlFacebookUidStore(sqlStore SqlStore, metrics einterfaces.MetricsInterface) store.FacebookUidStore {
	fs := &sqlFacebookUidStore{
		SqlStore: sqlStore,
		metrics:  metrics,
	}

	fs.usersQuery = fs.getQueryBuilder().
		Select("facebookuids.*").
		From("facebookuids")

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(model.FacebookUid{}, "FacebookUids").SetKeys(false, "Id")
		table.ColMap("Id")
		table.ColMap("Name")
		table.ColMap("Email")
	}

	return fs
}

func (us sqlFacebookUidStore) Get(id string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		if obj, err := us.GetReplica().Get(model.FacebookUid{}, id); err != nil {
			result.Err = model.NewAppError("sqlFacebookUidStore.Get", "store.sql_user.get.app_error", nil, "id="+id+", "+err.Error(), http.StatusInternalServerError)
		} else if obj == nil {
			result.Err = model.NewAppError("sqlFacebookUidStore.Get", store.MISSING_FANPAGE_ERROR, nil, "id="+id, http.StatusNotFound)
		} else {
			result.Data = obj.(*model.FacebookUid)
		}
	})
}

func (us sqlFacebookUidStore) GetByIds(userIds []string, allowFromCache bool) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		users := []*model.FacebookUid{}
		remainingUserIds := make([]string, 0)

		if allowFromCache {
			for _, userId := range userIds {
				if cacheItem, ok := suProfileByIdsCache.Get(userId); ok {
					u := &model.FacebookUid{}
					*u = *cacheItem.(*model.FacebookUid)
					users = append(users, u)
				} else {
					remainingUserIds = append(remainingUserIds, userId)
				}
			}
			if us.metrics != nil {
				us.metrics.AddMemCacheHitCounter("Profile By Ids", float64(len(users)))
				us.metrics.AddMemCacheMissCounter("Profile By Ids", float64(len(remainingUserIds)))
			}
		} else {
			remainingUserIds = userIds
			if us.metrics != nil {
				us.metrics.AddMemCacheMissCounter("Profile By Ids", float64(len(remainingUserIds)))
			}
		}

		// If everything came from the cache then just return
		if len(remainingUserIds) == 0 {
			result.Data = users
			return
		}

		query := us.usersQuery.
			Where(map[string]interface{}{
				"facebookuids.Id": remainingUserIds,
			})

		queryString, args, err := query.ToSql()

		if err != nil {
			result.Err = model.NewAppError("SqlUserStore.GetProfileByIds", "store.sql_user.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, err := us.GetReplica().Select(&users, queryString, args...); err != nil {
			result.Err = model.NewAppError("SqlUserStore.GetProfileByIds", "store.sql_user.get_profiles.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		}

		for _, u := range users {
			//u.Sanitize(map[string]bool{})

			cpy := &model.FacebookUid{}
			*cpy = *u
			suProfileByIdsCache.AddWithExpiresInSecs(cpy.Id, cpy, SU_PROFILE_BY_IDS_CACHE_SEC)
		}

		result.Data = users
	})
}

func (fs sqlFacebookUidStore) UpsertFromMap(data map[string]interface{}) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		fbId := model.FacebookUid{}
		err := fs.GetReplica().SelectOne(&fbId, "SELECT * from facebookuids WHERE id = :id", map[string]interface{}{"id": data["id"].(string)})
		if err != nil {
			if err == sql.ErrNoRows {
				newUid := model.FacebookUid{
					Id: data["id"].(string),
					Name: data["name"].(string),
				}

				if err := fs.GetMaster().Insert(&newUid); err != nil {
					result.Err = model.NewAppError("SqlFanpageStore.GetMember", "store.sql_fanpage.get_member.app_error", nil, "uid="+data["id"].(string)+" "+err.Error(), http.StatusInternalServerError)
				} else {
					result.Data = &newUid
				}
			} else {
				result.Err = model.NewAppError("SqlFanpageStore.GetMember", "store.sql_fanpage.get_member.app_error", nil, "uid="+data["id"].(string)+" "+err.Error(), http.StatusInternalServerError)
			}
		} else {
			result.Data = &fbId
		}
	})
}

func (fs sqlFacebookUidStore) UpsertFromFbUser(fbUser facebookgraph.FacebookUser) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var fbId *model.FacebookUid
		err := fs.GetReplica().SelectOne(&fbId, "SELECT * from facebookuids WHERE id = :id", map[string]interface{}{"id": fbUser.Id})
		if err != nil {
			if err == sql.ErrNoRows {
				var newUid model.FacebookUid
				newUid.Id = fbUser.Id
				newUid.Name = fbUser.Name
				newUid.PageId = fbUser.PageId

				if err := fs.GetMaster().Insert(&newUid); err != nil {
					result.Err = model.NewAppError("SqlFanpageStore.GetMember", "store.sql_fanpage.get_member.app_error", nil, "uid="+fbUser.Id+" "+err.Error(), http.StatusInternalServerError)
				} else {
					result.Data = model.NewId()
				}
			} else {
				result.Err = model.NewAppError("SqlFanpageStore.GetMember", "store.sql_fanpage.get_member.app_error", nil, "uid="+fbUser.Id+" "+err.Error(), http.StatusInternalServerError)
			}
		} else {
			result.Data = fbId
		}
	})
}

func (fs sqlFacebookUidStore) UpdatePageId(id, pageId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		if _, err := fs.GetMaster().Exec("UPDATE facebookuids SET PageId = :PageId WHERE Id = :Id", map[string]interface{}{"PageId": pageId, "Id": id}); err != nil {
			result.Err = model.NewAppError("sqlFacebookUidStore.UpdatePageId", "store.sql_fanpage.update_page_id.app_error", nil, "page_id="+pageId, http.StatusInternalServerError)
		} else {
			result.Data = true
		}
	})
}

func (fs sqlFacebookUidStore) UpdatePageScopeId(id, pageScopeId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		if _, err := fs.GetMaster().Exec("UPDATE facebookuids SET PageScopeId = :PageScopeId WHERE Id = :Id", map[string]interface{}{"PageScopeId": pageScopeId, "Id": id}); err != nil {
			result.Err = model.NewAppError("sqlFacebookUidStore.UpdatePageScopeId", "store.sql_fanpage.update_page_id.app_error", nil, "page_scope_id="+pageScopeId, http.StatusInternalServerError)
		} else {
			result.Data = true
		}
	})
}
