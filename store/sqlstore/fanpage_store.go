// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package sqlstore

import (
	"bitbucket.org/enesyteam/papo-server/einterfaces"
	"bitbucket.org/enesyteam/papo-server/model"
	"bitbucket.org/enesyteam/papo-server/store"
	"bitbucket.org/enesyteam/papo-server/utils"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
)

type sqlFanpageStore struct {
	SqlStore
	metrics einterfaces.MetricsInterface
}

const (
	ALL_PAGE_MEMBERS_FOR_USER_CACHE_SIZE = model.SESSION_CACHE_SIZE
	ALL_PAGE_MEMBERS_FOR_USER_CACHE_SEC  = 900 // 15 mins

	ALL_CHANNEL_MEMBERS_NOTIFY_PROPS_FOR_CHANNEL_CACHE_SIZE = model.SESSION_CACHE_SIZE
	ALL_CHANNEL_MEMBERS_NOTIFY_PROPS_FOR_CHANNEL_CACHE_SEC  = 1800 // 30 mins

	PAGE_MEMBERS_COUNTS_CACHE_SIZE = model.PAGE_CACHE_SIZE
	CHANNEL_MEMBERS_COUNTS_CACHE_SEC  = 1800 // 30 mins

	CHANNEL_CACHE_SEC = 900 // 15 mins
)

var pageMemberCountsCache = utils.NewLru(PAGE_MEMBERS_COUNTS_CACHE_SIZE)
var allPageMembersForUserCache = utils.NewLru(ALL_PAGE_MEMBERS_FOR_USER_CACHE_SIZE)
var pageCache = utils.NewLru(model.PAGE_CACHE_SIZE)
var pageByNameCache = utils.NewLru(model.PAGE_CACHE_SIZE)

//type fanpageMember struct {
//	FanpageId   string
//	UserId      string
//	Roles       string
//	AccessToken string
//}

//func newFanpageMemberFromModel(fm *model.FanpageMember) *fanpageMember {
//	return &fanpageMember{
//		FanpageId:   fm.FanpageId,
//		UserId:      fm.UserId,
//		Roles:       fm.Roles,
//		AccessToken: fm.AccessToken,
//	}
//}

func (s sqlFanpageStore) ClearCaches() {
	pageMemberCountsCache.Purge()
	allPageMembersForUserCache.Purge()
	//allChannelMembersNotifyPropsForChannelCache.Purge()
	pageCache.Purge()
	pageByNameCache.Purge()

	if s.metrics != nil {
		s.metrics.IncrementMemCacheInvalidationCounter("Page Member Counts - Purge")
		s.metrics.IncrementMemCacheInvalidationCounter("All Page Members for User - Purge")
		s.metrics.IncrementMemCacheInvalidationCounter("All Page Members Notify Props for Channel - Purge")
		s.metrics.IncrementMemCacheInvalidationCounter("Page - Purge")
		s.metrics.IncrementMemCacheInvalidationCounter("Page By Name - Purge")
	}
}

func NewSqlFanpageStore(sqlStore SqlStore, metrics einterfaces.MetricsInterface) store.FanpageStore {
	// Khởi tạo đối tượng fanpage
	fs := &sqlFanpageStore{
		SqlStore: sqlStore,
		metrics:  metrics,
	}

	for _, db := range sqlStore.GetAllConns() {
		// Khởi tạo table
		table := db.AddTableWithName(model.Fanpage{}, "Fanpages").SetKeys(false, "Id")
		table.ColMap("Id").SetMaxSize(26)
		// Thiết lập cho các columns
		// các thiết lập này cố gắng SetMaxSize vừa đủ cho các column, tránh lãng phí tài nguyên không cần thiết
		table.ColMap("PageId").SetMaxSize(50).SetUnique(true) // có thể nhỏ hơn nhưng không rõ facebook có quy định nào không
		table.ColMap("Name").SetMaxSize(120)                   // tên fanpage được Facebook giới hạn dài tối đa 50 ký tự
		table.ColMap("Category").SetMaxSize(120)
		table.ColMap("Status").SetMaxSize(26)

		tablem := db.AddTableWithName(model.FanpageMember{}, "FanpageMembers").SetKeys(false, "FanpageId", "UserId")
		tablem.ColMap("FanpageId").SetMaxSize(26)
		tablem.ColMap("UserId").SetMaxSize(26)
		tablem.ColMap("Roles").SetMaxSize(64)
		tablem.ColMap("AccessToken").SetMaxSize(500)
		table.ColMap("Filenames").SetMaxSize(model.FANPAGE_FILENAMES_MAX_RUNES)
		table.ColMap("FileIds").SetMaxSize(150)
	}

	return fs
}

// Index để phục vụ cho tìm kiếm
func (fs sqlFanpageStore) CreateIndexesIfNotExists() {
	fs.CreateIndexIfNotExists("idx_fanpages_page_id", "Fanpages", "PageId")
	fs.CreateIndexIfNotExists("idx_fanpages_name", "Fanpages", "Name")
	fs.CreateIndexIfNotExists("idx_fanpages_update_at", "Fanpages", "UpdateAt")
	fs.CreateIndexIfNotExists("idx_fanpages_create_at", "Fanpages", "CreateAt")
	fs.CreateIndexIfNotExists("idx_fanpages_delete_at", "Fanpages", "DeleteAt")
	fs.CreateIndexIfNotExists("idx_fanpages_block_at", "Fanpages", "BlockAt")

	fs.CreateIndexIfNotExists("idx_fanpagemembers_team_id", "FanpageMembers", "FanpageId")
	fs.CreateIndexIfNotExists("idx_fanpagemembers_user_id", "FanpageMembers", "UserId")
}

// Save vào database
func (fs sqlFanpageStore) Save(fanpage *model.Fanpage) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		fanpage.PreSave()
		if result.Err = fanpage.IsValid(); result.Err != nil {
			return
		}

		// insert fanpage vào database
		if err := fs.GetMaster().Insert(fanpage); err != nil {
			// dự đoán rằng chỉ có lỗi trùng id do page đã được thêm vào trước đó
			if IsUniqueConstraintError(err, []string{"Fanpages", "fanpages_pageid_key", "idx_fanpages_pageid_unique"}) {
				result.Err = model.NewAppError("sqlFanpageStore.Save", "store.sql_fanpage.save.id_exists.app_error", nil, "fanpage_id="+fanpage.Id+", "+err.Error(), http.StatusBadRequest)
			} else {
				fmt.Println("fanpage loi: ", fanpage)
				// Một lỗi xảy ra và không biết nó là lỗi gì
				result.Err = model.NewAppError("sqlFanpageStore.Save", "store.sql_fanpage.save.app_error", nil, "page_id="+fanpage.Id+", "+err.Error(), http.StatusInternalServerError)
			}
		} else {
			result.Data = fanpage
			// thêm members
		}
	})
}

func (fs sqlFanpageStore) Update(newPage *model.Fanpage, oldPage *model.Fanpage) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

	})
}

func (fs sqlFanpageStore) ValidatePagesBeforeInit(pageIds *model.LoadPagesInput) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		validateResult := model.PagesInitValidationResult{}
		for _, pageId := range pageIds.PageIds {
			var page *model.Fanpage
			err := fs.GetReplica().SelectOne(&page, "SELECT fanpages.* FROM fanpages WHERE pageid = :pageId", map[string]interface{}{"pageId": pageId})
			if err != nil {
				if err == sql.ErrNoRows {
					validateResult.Error = append(validateResult.Error, model.PageInitValidationError{
						Id: pageId,
						ErrorCode: 1,
						Message: "Không thể khởi tạo Page chưa có trong cơ sở dữ liệu.",
					})
				} else {
					validateResult.Error = append(validateResult.Error, model.PageInitValidationError{
						Id: pageId,
						ErrorCode: 0,
						Message: "Không tìm thấy Page trong cơ sở dữ liệu. Lỗi không xác định.",
					})
				}

			} else {
				// khong cho phép khởi tạo các page đã khởi tạo từ trước hoặc đang trong hàng đợi khởi tạo, hoặc đang khởi tạo, hoặc đang blocked
				if page.Status == model.PAGE_STATUS_QUEUED {
					validateResult.Error = append(validateResult.Error, model.PageInitValidationError{
						Id: pageId,
						ErrorCode: 2,
						Message: "Trang đang chờ khởi tạo.",
					})
				} else if page.Status == model.PAGE_STATUS_INITIALIZING {
					validateResult.Error = append(validateResult.Error, model.PageInitValidationError{
						Id: pageId,
						ErrorCode: 3,
						Message: "Trang đang khởi tạo.",
					})
				} else if page.Status == model.PAGE_STATUS_INITIALIZED {
					validateResult.Error = append(validateResult.Error, model.PageInitValidationError{
						Id: pageId,
						ErrorCode: 2,
						Message: "Trang đã được khởi tạo thành công.",
					})
				} else if page.Status == model.PAGE_STATUS_BLOCKED {
					validateResult.Error = append(validateResult.Error, model.PageInitValidationError{
						Id: pageId,
						ErrorCode: 2,
						Message: "Trang đang bị khóa và không thể khởi tạo.",
					})
				} else {
					validateResult.Success = append(validateResult.Success, pageId)
				}
			}
		}

		result.Data = &validateResult
	})
}

func (fs sqlFanpageStore) UpdateStatus(pageId string, status string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		// Kiểm tra trạng thái page hiện tại, không cho phép update trạng thái từ initializing, initialized sang
		// queued, ready,
		if _, err := fs.GetMaster().Exec("UPDATE Fanpages SET Status = :Status WHERE PageId = :PageId", map[string]interface{}{"PageId": pageId, "Status": status}); err != nil {
			result.Err = model.NewAppError("sqlFanpageStore.UpdateStatus", "store.sqlFanpageStore.update_status_error.app_error", nil, "page_id="+pageId+"&status="+status, http.StatusInternalServerError)
		} else {
			result.Data = status
		}
	})
}

func (fs sqlFanpageStore) UpdatePagesStatus(pageIds *model.LoadPagesInput, status string) store.StoreChannel {

	inPage := ` WHERE PageId IN (`
	i := 0
	for _, id := range pageIds.PageIds  {
		inPage += "'" + id + "'"
		if i < len(pageIds.PageIds) - 1 {
			inPage += ","
		}
		i++
	}
	inPage += ")"

	query := `UPDATE Fanpages SET Status = :Status ` + inPage

	fmt.Println(query)

	return store.Do(func(result *store.StoreResult) {
		if _, err := fs.GetMaster().Exec(query, map[string]interface{}{"Status": status}); err != nil {
			result.Err = model.NewAppError("sqlFanpageStore.UpdatePagesStatus", "store.sqlFanpageStore.update_status_error.app_error", nil,"&status="+status, http.StatusInternalServerError)
		} else {
			result.Data = status
		}
	})
}

func (fs sqlFanpageStore) Get(id string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		if obj, err := fs.GetReplica().Get(model.Fanpage{}, id); err != nil {
			result.Err = model.NewAppError("SqlFanpageStore.Get", "store.sql_fanpage.get.app_error", nil, "fanpage_id="+id+", "+err.Error(), http.StatusInternalServerError)
		} else if obj == nil {
			result.Err = model.NewAppError("SqlFanpageStore.Get", store.MISSING_FANPAGE_ERROR, nil, "page_id="+id, http.StatusNotFound)
		} else {
			result.Data = obj.(*model.Fanpage)
		}
	})
}

func (fs sqlFanpageStore) GetFanpagesByUserId(userId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var data []*model.FanpageIncludeInitResult
		query := `
			SELECT 
				json_build_object(
					'id', f.id,
					'page_id', f.pageid,
					'name', f.name,
					'category', f.category,
					'status', f.status,
					'access_token', fanpagemembers.accesstoken
				) as data,
				COALESCE(NULLIF(m.init::TEXT, '{}'), '{}')::JSON as init
			FROM 
				fanpages f
			LEFT JOIN 
					fanpagemembers
					ON  fanpagemembers.pageid = f.pageid
			LEFT OUTER JOIN (
	SELECT 
			pageid, json_build_object('id', x.id, 'page_id', x.pageid, 'post_count', x.postcount, 'conversation_count', x.conversationcount, 'message_count', x.messagecount, 'comment_count', x.commentcount, 'start_at', x.startat, 'end_at', x.endat, 'creator', x.creator) as init
	FROM
			fanpageinitresults x
)  m ON m.pageid = f.pageid
			WHERE 
				fanpagemembers.UserId =	:UserId
			AND
				fanpagemembers.pageid = f.pageid
		`
		//query2 := "SELECT Fanpages.* FROM Fanpages, FanpageMembers WHERE FanpageMembers.FanpageId = Fanpages.Id AND FanpageMembers.UserId = :UserId"
		if _, err := fs.GetReplica().Select(&data, query, map[string]interface{}{"UserId": userId}); err != nil {
			result.Err = model.NewAppError("SqlTeamStore.GetFanpagesByUserId", "store.sqlFanpageStore.GetFanpagesByUserId.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		}
		result.Data = data
	})
}

func (fs sqlFanpageStore) GetFanpageByPageID(pageId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var page *model.Fanpage
		err := fs.GetReplica().SelectOne(&page, "SELECT fanpages.* FROM fanpages WHERE pageid = :pageId", map[string]interface{}{"pageId": pageId})
		if err != nil {
			if err == sql.ErrNoRows {
				result.Err = model.NewAppError("SqlFanpageStore.GetOneFanpageByUserId", "store.sql_fanpage.get_one.missing.app_error", nil, "pageId="+pageId+" "+err.Error(), http.StatusNotFound)
				return
			}
			result.Err = model.NewAppError("SqlFanpageStore.GetOneFanpageByUserId", "store.sql_fanpage.get_one.app_error", nil, "pageId="+pageId+" "+err.Error(), http.StatusInternalServerError)
			return
		}
		result.Data = page
	})
}

func (fs sqlFanpageStore) SaveFanPageMember(member *model.FanpageMember) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		if result.Err = member.IsValid(); result.Err != nil {
			return
		}

		// Kiểm tra nếu FanpageMember chưa có => thêm vào db
		// nếu đã tồn tại => cập nhật token
		var dbMember *model.FanpageMember
		err := fs.GetReplica().SelectOne(&dbMember, "SELECT * FROM fanpagemembers WHERE fanpageid = :FanpageId AND userid = :UserId", map[string]interface{}{"FanpageId": member.FanpageId, "UserId": member.UserId})

		if err != nil {
			if err == sql.ErrNoRows {
				if err := fs.GetMaster().Insert(member); err != nil {
					result.Err = model.NewAppError("SqlFanpageStore.SaveMember", "store.sql_fanpage.save_member.save.app_error", nil, "fanpage_id="+member.FanpageId+", user_id="+member.UserId+", "+err.Error(), http.StatusInternalServerError)
					return
				}
			} else {
				result.Data = member
				fmt.Println("Đã thêm mới fanpage member: ", member.UserId, " & page: ", member.PageId)
				//result.Err = model.NewAppError("SqlFanpageStore.SaveMember", "store.sql_fanpage.save_member.save.app_error", nil, "fanpage_id="+member.FanpageId+", user_id="+member.UserId+", "+err.Error(), http.StatusInternalServerError)
				return
			}
		}

		// FanpageMember đã tồn tại => cập nhật page token
		query := "UPDATE fanpagemembers SET accesstoken = :AccessToken WHERE userid = :UserId AND pageid = :PageId"
		_, updateError := fs.GetMaster().Exec(query, map[string]interface{}{"AccessToken": member.AccessToken, "UserId": member.UserId, "PageId": member.PageId})
		if updateError != nil {
			result.Err = model.NewAppError("sqlFanpageStore.SaveFanPageMember", "store.sqlFanpageStore.update_fanpage_member_token.app_error", nil, "page_id="+member.PageId+ "&user_id="+ member.UserId +", "+err.Error(), http.StatusInternalServerError)
			return
		} else {
			fmt.Println("Đã cập nhật page token của user: ", member.UserId, " & page: ", member.PageId)
			result.Data = member
		}
	})
}

func (fs sqlFanpageStore) GetMember(fanpageId string, userId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var dbMember *model.FanpageMember
		err := fs.GetReplica().SelectOne(&dbMember, "SELECT * FROM fanpagemembers WHERE fanpageid = :FanpageId AND userid = :UserId", map[string]interface{}{"FanpageId": fanpageId, "UserId": userId})

		if err != nil {
			if err == sql.ErrNoRows {
				result.Err = model.NewAppError("SqlFanpageStore.GetMember", "store.sql_fanpage.get_member.missing.app_error", nil, "fanpageId="+fanpageId+" userId="+userId+" "+err.Error(), http.StatusNotFound)
				return
			}
			result.Err = model.NewAppError("SqlFanpageStore.GetMember", "store.sql_fanpage.get_member.app_error", nil, "fanpageId="+fanpageId+" userId="+userId+" "+err.Error(), http.StatusInternalServerError)
			return
		}
		result.Data = dbMember
	})
}

func (fs sqlFanpageStore) GetMemberByPageId(pageId string, userId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var dbMember *model.FanpageMember
		err := fs.GetReplica().SelectOne(&dbMember, "SELECT * FROM fanpagemembers WHERE pageid = :PageId AND userid = :UserId", map[string]interface{}{"PageId": pageId, "UserId": userId})

		if err != nil {
			if err == sql.ErrNoRows {
				result.Err = model.NewAppError("SqlFanpageStore.GetMember", "store.sql_fanpage.get_member.missing.app_error", nil, "pageId="+pageId+" userId="+userId+" "+err.Error(), http.StatusNotFound)
				return
			}
			result.Err = model.NewAppError("SqlFanpageStore.GetMember", "store.sql_fanpage.get_member.app_error", nil, "pageId="+pageId+" userId="+userId+" "+err.Error(), http.StatusInternalServerError)
			return
		}
		result.Data = dbMember
	})
}

func (fs sqlFanpageStore) GetOneFanPageMember(pageiId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var fanpageMember *model.FanpageMember
		err := fs.GetReplica().SelectOne(&fanpageMember, "SELECT a.* FROM fanpagemembers a INNER JOIN fanpages b on a.fanpageid = b.id  WHERE b.pageid = :pageId LIMIT 1", map[string]interface{}{"pageId": pageiId})
		if err != nil {
			if err == sql.ErrNoRows {
				result.Err = model.NewAppError("SqlFanpageStore.GetMember", "store.sql_fanpage.get_member.missing.app_error", nil, "fanpageId="+pageiId+" "+err.Error(), http.StatusNotFound)
				return
			}
			result.Err = model.NewAppError("SqlFanpageStore.GetMember", "store.sql_fanpage.get_member.app_error", nil, "fanpageId="+pageiId+" "+err.Error(), http.StatusInternalServerError)
		}
		result.Data = fanpageMember
	})
}

func (s sqlFanpageStore) UpdateLastViewedAt(pageIds []string, userId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		//props := make(map[string]interface{})
		//
		//updateIdQuery := ""
		//for index, pageId := range pageIds {
		//	if len(updateIdQuery) > 0 {
		//		updateIdQuery += " OR "
		//	}
		//
		//	props["pageId"+strconv.Itoa(index)] = pageId
		//	updateIdQuery += "PageId = :pageId" + strconv.Itoa(index)
		//}
		//
		//times := map[string]int64{}
		//
		//var updateQuery string
		//updateQuery = `UPDATE
		//		ChannelMembers
		//	SET
		//	    LastViewedAt = CAST(CASE ChannelId ` + lastViewedQuery + ` END AS BIGINT),
		//	    LastUpdateAt = CAST(CASE ChannelId ` + lastViewedQuery + ` END AS BIGINT)
		//	WHERE
		//	        UserId = :UserId
		//	        AND (` + updateIdQuery + `)`
	})
}

type allPageMember struct {
	PageId                     	  string
	FanpageId 					  string
	UserId 						  string
	Roles                         string
	SchemeUser                    sql.NullBool
	SchemeAdmin                   sql.NullBool
	TeamSchemeDefaultUserRole     sql.NullString
	TeamSchemeDefaultAdminRole    sql.NullString
	ChannelSchemeDefaultUserRole  sql.NullString
	ChannelSchemeDefaultAdminRole sql.NullString
}

type allPageMembers []allPageMember

func (db allPageMember) Process() (string, string) {
	roles := strings.Fields(db.Roles)

	// Add any scheme derived roles that are not in the Roles field due to being Implicit from the Scheme, and add
	// them to the Roles field for backwards compatibility reasons.
	var schemeImpliedRoles []string
	if db.SchemeUser.Valid && db.SchemeUser.Bool {
		if db.ChannelSchemeDefaultUserRole.Valid && db.ChannelSchemeDefaultUserRole.String != "" {
			schemeImpliedRoles = append(schemeImpliedRoles, db.ChannelSchemeDefaultUserRole.String)
		} else if db.TeamSchemeDefaultUserRole.Valid && db.TeamSchemeDefaultUserRole.String != "" {
			schemeImpliedRoles = append(schemeImpliedRoles, db.TeamSchemeDefaultUserRole.String)
		} else {
			schemeImpliedRoles = append(schemeImpliedRoles, model.CHANNEL_USER_ROLE_ID)
		}
	}
	if db.SchemeAdmin.Valid && db.SchemeAdmin.Bool {
		if db.ChannelSchemeDefaultAdminRole.Valid && db.ChannelSchemeDefaultAdminRole.String != "" {
			schemeImpliedRoles = append(schemeImpliedRoles, db.ChannelSchemeDefaultAdminRole.String)
		} else if db.TeamSchemeDefaultAdminRole.Valid && db.TeamSchemeDefaultAdminRole.String != "" {
			schemeImpliedRoles = append(schemeImpliedRoles, db.TeamSchemeDefaultAdminRole.String)
		} else {
			schemeImpliedRoles = append(schemeImpliedRoles, model.CHANNEL_ADMIN_ROLE_ID)
		}
	}
	for _, impliedRole := range schemeImpliedRoles {
		alreadyThere := false
		for _, role := range roles {
			if role == impliedRole {
				alreadyThere = true
			}
		}
		if !alreadyThere {
			roles = append(roles, impliedRole)
		}
	}

	return db.PageId, strings.Join(roles, " ")
}

func (db allPageMembers) ToMapStringString() map[string]string {
	result := make(map[string]string)

	for _, item := range db {
		key, value := item.Process()
		result[key] = value
	}

	return result
}

func (s sqlFanpageStore) GetAllPageMembersForUser(userId string, allowFromCache bool, includeDeleted bool) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		cache_key := userId
		if includeDeleted {
			cache_key += "_deleted"
		}
		if allowFromCache {
			if cacheItem, ok := allPageMembersForUserCache.Get(cache_key); ok {
				if s.metrics != nil {
					s.metrics.IncrementMemCacheHitCounter("All Page Members for User")
				}
				result.Data = cacheItem.(map[string]string)
				return
			}
		}

		if s.metrics != nil {
			s.metrics.IncrementMemCacheMissCounter("All Page Members for User")
		}

		var deletedClause string
		if !includeDeleted {
			deletedClause = "Fanpages.DeleteAt = 0"
		}

		var data allPageMembers
		_, err := s.GetReplica().Select(&data, `
			SELECT
				FanpageMembers.PageId, FanpageMembers.UserId, FanpageMembers.Roles
			FROM
				FanpageMembers
			INNER JOIN
				Fanpages ON FanpageMembers.PageId = Fanpages.PageId
			WHERE
				`+deletedClause+`
			`, map[string]interface{}{"UserId": userId})

		if err != nil {
			fmt.Println(err.Error())
			result.Err = model.NewAppError("sqlFanpageStore.GetAllChannelMembersForUser", "store.sql_channel.get_channels.get.app_error", nil, "userId="+userId+", err="+err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Println("data", data)
		ids := data.ToMapStringString()
		fmt.Println("ids", ids)
		result.Data = ids

		if allowFromCache {
			allPageMembersForUserCache.AddWithExpiresInSecs(cache_key, ids, ALL_PAGE_MEMBERS_FOR_USER_CACHE_SEC)
		}
	})
}
