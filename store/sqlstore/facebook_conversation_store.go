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
	sq "github.com/Masterminds/squirrel"
	"net/http"
	"strings"
	"time"
)

const (
	ATTACHMENT_CACHE_SIZE = model.CHANNEL_CACHE_SIZE
	ATTACHMENT_CACHE_SEC  = 900 // 15 mins
	ATTACHMENT_BY_IDS_CACHE_SIZE      = model.SESSION_CACHE_SIZE
	ATTACHMENT_BY_IDS_CACHE_SEC       = 900 // 15 mins
)

var attachmentCache *utils.Cache = utils.NewLru(ATTACHMENT_CACHE_SIZE)
var attachmentByIdsCache *utils.Cache = utils.NewLru(ATTACHMENT_BY_IDS_CACHE_SIZE)

type sqlFacebookConversationStore struct {
	SqlStore
	metrics einterfaces.MetricsInterface

	attachmentsQuery sq.SelectBuilder
	conversationQuery sq.SelectBuilder
}

//func newConversationMessageFromModel(fm *model.FacebookConversationMessage) *model.FacebookConversationMessage {
//	return &model.FacebookConversationMessage{
//		FanpageId:   fm.FanpageId,
//		UserId:      fm.UserId,
//		Roles:       fm.Roles,
//		Accesstoken: fm.AccessToken,
//	}
//}

func NewSqlFacebookConversationStore(sqlStore SqlStore, metrics einterfaces.MetricsInterface) store.FacebookConversationStore {
	// Khởi tạo đối tượng FacebookConversation
	cv := &sqlFacebookConversationStore{
		SqlStore: sqlStore,
		metrics:  metrics,
	}

	cv.attachmentsQuery = cv.getQueryBuilder().
		Select("facebookattachmentimages.*").
		From("facebookattachmentimages")

	cv.conversationQuery = cv.getQueryBuilder().
		Select("FacebookConversations.*").
		From("FacebookConversations")

	for _, db := range sqlStore.GetAllConns() {
		// Khởi tạo table
		table := db.AddTableWithName(model.FacebookConversation{}, "FacebookConversations").SetKeys(false, "Id")
		table.ColMap("Id").SetMaxSize(26)
		// Thiết lập cho các columns
		table.ColMap("Type").SetMaxSize(12)
		//table.ColMap("Snippet").SetMaxSize(120) // chỉ lấy 120 ký tự

		// Khởi tạo các table con
		tablem := db.AddTableWithName(model.FacebookConversationMessage{}, "FacebookConversationMessages").SetKeys(false, "Id", "ConversationId", )

		tablem.ColMap("ConversationId").SetMaxSize(26)
		tablem.ColMap("Id").SetMaxSize(26)

		// image table
		tablei := db.AddTableWithName(model.FacebookAttachmentImage{}, "FacebookAttachmentImages").SetKeys(false, "Id", "MessageId", "PostId" )
		tablei.ColMap("MessageId").SetMaxSize(26)
		tablei.ColMap("Id").SetMaxSize(26)
	}

	return cv
}

func (fs sqlFacebookConversationStore) CreateIndexesIfNotExists() {
	fs.CreateIndexIfNotExists("idx_facebook_conversations_page_id", "FacebookConversations", "PageId")
	fs.CreateIndexIfNotExists("idx_facebook_conversations_type", "FacebookConversations", "Type")
	fs.CreateIndexIfNotExists("idx_facebook_conversations_seen", "FacebookConversations", "Seen")
	fs.CreateIndexIfNotExists("idx_facebook_conversations_update_at", "FacebookConversations", "UpdateAt")
	fs.CreateIndexIfNotExists("idx_facebook_conversations_updated_time", "FacebookConversations", "UpdatedTime")
	fs.CreateIndexIfNotExists("idx_facebook_conversations_create_at", "FacebookConversations", "CreateAt")
	fs.CreateIndexIfNotExists("idx_facebook_conversations_delete_at", "FacebookConversations", "DeleteAt")

	fs.CreateIndexIfNotExists("idx_facebook_conversations_messages_created_time", "FacebookConversationMessages", "CreatedTime")
	//fs.CreateIndexIfNotExists("idx_facebook_conversations_messages_conversation_id", "FacebookConversationMessages", "ConversationId")

	// còn nhiều thứ khác cần index
}

func (fs sqlFacebookConversationStore) Get(id string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		if obj, err := fs.GetReplica().Get(model.FacebookConversation{}, id); err != nil {
			result.Err = model.NewAppError("sqlFacebookConversationStore.Get", "store.sql_fanpage.get.app_error", nil, "fanpage_id="+id+", "+err.Error(), http.StatusInternalServerError)
		} else if obj == nil {
			result.Err = model.NewAppError("sqlFacebookConversationStore.Get", store.MISSING_FANPAGE_ERROR, nil, "page_id="+id, http.StatusNotFound)
		} else {
			result.Data = obj.(*model.FacebookConversation)
		}
	})
}

func (fs sqlFacebookConversationStore) GetMessage(id string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		if obj, err := fs.GetReplica().Get(model.FacebookConversationMessage{}, id); err != nil {
			result.Err = model.NewAppError("SqlFanpageStore.Get", "store.sql_fanpage.get.app_error", nil, "fanpage_id="+id+", "+err.Error(), http.StatusInternalServerError)
		} else if obj == nil {
			result.Err = model.NewAppError("SqlFanpageStore.Get", store.MISSING_FANPAGE_ERROR, nil, "page_id="+id, http.StatusNotFound)
		} else {
			result.Data = obj.(*model.FacebookConversationMessage)
		}
	})
}

func (fs sqlFacebookConversationStore) Save(conversation *model.FacebookConversation) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		conversation.PreSave()
		//if result.Err = fanpage.IsValid(); result.Err != nil {
		//	return
		//}
		//
		// insert vào database
		if err := fs.GetMaster().Insert(conversation); err != nil {
			// cần sửa
			result.Err = model.NewAppError("sqlFanpageStore.Save", "store.sql_fanpage.save.app_error", nil, "page_id="+conversation.Id+", "+err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = conversation
			//
		}
	})
}

func (fs sqlFacebookConversationStore) UpsertCommentConversation(conversation *model.FacebookConversation) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		var conversationData *model.FacebookConversation
		query := "SELECT * FROM facebookconversations a WHERE a.from = :uid AND a.postid = :postid"
		if err := fs.GetReplica().SelectOne(&conversationData, query, map[string]interface{}{"uid": conversation.From, "postid": conversation.PostId}); err != nil {
			if err == sql.ErrNoRows {
				conversation.PreSave()
				if err := fs.GetMaster().Insert(conversation); err != nil {
					result.Err = model.NewAppError("sqlFanpageStore.Save", "store.sql_fanpage.save.app_error", nil, "page_id="+conversation.Id+", "+err.Error(), http.StatusInternalServerError)
				} else {
					result.Data = conversation
				}
			} else {
				result.Err = model.NewAppError("sqlFanpageStore.Save", "store.sql_fanpage.save.app_error", nil, "page_id="+conversation.Id+", "+err.Error(), http.StatusInternalServerError)
			}
		} else {
			result.Data = conversationData
		}
	})
}

func (fs sqlFacebookConversationStore) UpdatePageScopeId(conversationId, pageScopeId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		query := "UPDATE facebookconversations SET PageScopeId = :PageScopeId WHERE id = :ConversationId"
		_, err := fs.GetMaster().Exec(query, map[string]interface{}{"PageScopeId": pageScopeId, "ConversationId": conversationId})
		if err != nil {
			result.Err = model.NewAppError("sqlFacebookConversationStore.UpdatePageScopeId", "store.sql_fanpage.save.app_error", nil, "page_id="+conversationId+", "+err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = true
		}
	})
}

func (fs sqlFacebookConversationStore) UpdateLatestTime(conversationId string, time string, commentId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		query := "UPDATE facebookconversations SET updatedtime = :updated_time WHERE id = :conversation_id"
		_, err := fs.GetMaster().Exec(query, map[string]interface{}{"updated_time": time, "comment_id": commentId, "conversation_id": conversationId})
		if err != nil {
			result.Err = model.NewAppError("sqlFanpageStore.Save", "store.sql_fanpage.save.app_error", nil, "page_id="+conversationId+", "+err.Error(), http.StatusInternalServerError)
		} else {
			var conversationData *model.FacebookConversation
			err := fs.GetReplica().SelectOne(&conversationData, "SELECT * FROM facebookconversations WHERE id = :conversation_id", map[string]interface{}{"conversation_id": conversationId})
			if err != nil {
				result.Err = model.NewAppError("sqlFanpageStore.Save", "store.sql_fanpage.save.app_error", nil, "page_id="+conversationId+", "+err.Error(), http.StatusInternalServerError)
			} else {
				result.Data = conversationData
			}
		}
	})
}

//// UnReadCount là số đếm tổng số tin nhắn từ khách hàng kể từ tin nhắn mới nhất của page
//// Mỗi khi Page Reaply 1 tin nhắn => reset bộ đếm UnReadCount về 0
//// Ta không thể reset bộ đếm này về 0 mỗi khi Conversation Seen được set = true
//// Lý do là vì chúng ta muốn cập nhật lại UnReadCount mỗi khi đánh dấu hội thoại là chưa đọc
//func (fs sqlFacebookConversationStore) UpdateConversationUnreadCount(conversationId string, reset bool) store.StoreChannel {
//	return store.Do(func(result *store.StoreResult) {
//		query := "UPDATE facebookconversations SET UnreadCount = UnreadCount + 1 WHERE id = :ConversationId"
//		if reset {
//			query = "UPDATE facebookconversations SET UnreadCount = 0 WHERE id = :ConversationId"
//		}
//		_, err := fs.GetMaster().Exec(query, map[string]interface{}{"ConversationId": conversationId})
//		if err != nil {
//			result.Err = model.NewAppError("sqlFacebookConversationStore.UpdateConversationUnreadCount", "store.sqlFacebookConversationStore.save.app_error", nil, "conversation_id="+conversationId+", "+err.Error(), http.StatusInternalServerError)
//		} else {
//			result.Data = nil
//		}
//	})
//}

// cập nhật conversation mỗi khi có một tin nhắn mới
// Tham số isFromPage cho ta biết tin nhắn này được gửi từ page hay từ khách hàng
func (fs sqlFacebookConversationStore) UpdateConversation(conversationId string, snippet string, isFromPage bool, updatedTime string, unreadCount int, lastUserMessageAt string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var query string
		if isFromPage {
			query = "UPDATE facebookconversations SET UnreadCount = 0, Snippet = :Snippet, UpdatedTime = :UpdatedTime, Seen = true, Replied = true  WHERE id = :ConversationId"
		} else {
			query = "UPDATE facebookconversations SET UnreadCount = UnreadCount + :UnreadCount, Snippet = :Snippet, UpdatedTime = :UpdatedTime, Seen = false, Replied = false, LastUserMessageAt = :LastUserMessageAt WHERE id = :ConversationId"
		}

		_, err := fs.GetMaster().Exec(query, map[string]interface{}{"ConversationId": conversationId, "Snippet": snippet, "UpdatedTime": updatedTime, "UnreadCount": unreadCount, "LastUserMessageAt": lastUserMessageAt})
		if err != nil {
			fmt.Println(err)
			result.Err = model.NewAppError("sqlFacebookConversationStore.UpdateConversation", "store.UpdateConversation.update_snippet.app_error", nil, "conversation_id="+conversationId+", "+err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = nil
		}
	})
}

func (fs sqlFacebookConversationStore) UpdateConversationUnread(conversationId string, isFromPage bool, unreadCount int, lastUserMessageAt string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var query string
		if isFromPage {
			query = "UPDATE facebookconversations SET UnreadCount = 0, Seen = true, Replied = true  WHERE id = :ConversationId"
		} else {
			query = "UPDATE facebookconversations SET UnreadCount = :UnreadCount, Seen = false, Replied = false, LastUserMessageAt = :LastUserMessageAt WHERE id = :ConversationId"
		}

		_, err := fs.GetMaster().Exec(query, map[string]interface{}{"ConversationId": conversationId, "UnreadCount": unreadCount, "LastUserMessageAt": lastUserMessageAt})
		if err != nil {
			fmt.Println(err)
			result.Err = model.NewAppError("sqlFacebookConversationStore.UpdateConversation", "store.UpdateConversation.update_snippet.app_error", nil, "conversation_id="+conversationId+", "+err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = nil
		}
	})
}

func (s sqlFacebookConversationStore) GetMessagesByConversationId(conversationId string, offset, limit int) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		if limit > 250 {
			result.Err = model.NewAppError("sqlFacebookConversationStore.GetMessagesByConversationId", "store.sql_conversation.get_messages.app_error", nil, "conversationId="+conversationId, http.StatusBadRequest)
			return
		}

		query := 	`SELECT * 
					FROM "facebookconversationmessages"
					WHERE facebookconversationmessages.conversationid = :ConversationId
					ORDER BY
						 facebookconversationmessages.CreatedTime DESC
					LIMIT :Limit OFFSET :Offset`

		var messages []*model.FacebookConversationMessage
		_, err := s.GetReplica().Select(&messages, query, map[string]interface{}{"ConversationId": conversationId, "Offset": offset, "Limit": limit})
		if err != nil {
			result.Err = model.NewAppError("SqlPostStore.GetLinearPosts", "store.sql_post.get_root_posts.app_error", nil, "conversationId="+conversationId+err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = messages
		}
	})
}

//func (fs sqlFacebookConversationStore) GetMessagesByConversationId(conversationId string, offset, limit int) store.StoreChannel {
//
//	return store.Do(func(result *store.StoreResult) {
//		if limit > 250 {
//			result.Err = model.NewAppError("sqlFacebookConversationStore.GetMessagesByConversationId", "store.sql_conversation.get_messages.app_error", nil, "conversationId="+conversationId, http.StatusBadRequest)
//			return
//		}
//
//		var data *model.ConversationMessageResponse
//		//var query string
//		childQuery :=
//			`
//				SELECT
//				   z.conversationid,
//				   json_agg(z.message) as messages
//				FROM
//				   (
//					  SELECT
//						 a.conversationid,
//						 json_build_object('id', a.id, 'comment_id', a.commentid, 'conversation_id', a.conversationid, 'page_id', a.pageid, 'type', a.type, 'message', a.message, 'created_time', a.createdtime, 'can_comment', a.cancomment, 'can_hide', a.canhide, 'can_like', a.canlike, 'can_reply', a.canreply, 'can_remove', a.canremove, 'can_reply_privately', a.canreplyprivately, 'private_reply_conversation', a.privatereplyconversation, 'is_hidden', a.ishidden, 'is_private', a.isprivate, 'attachments', COALESCE(NULLIF(d.attachments::TEXT, '[null]'), '[]')::JSON, 'from', json_build_object('id', e.id, 'name', e.name)) as message
//					  FROM
//						 FacebookConversationMessages a
//						 LEFT OUTER JOIN
//							(
//							   SELECT
//								  MessageId,
//								  json_agg(json_build_object('id', b.id, 'url', b.url, 'src', b.src, 'width', b.width, 'height', b.height, 'preview_url', b.previewurl, 'render_as_sticker', b.renderassticker, 'target_id', b.targetid, 'target_url', b.targeturl)) as attachments
//							   FROM
//								  FacebookAttachmentImages b
//								  INNER JOIN
//									 FacebookConversationMessages c
//									 ON b.MessageId = c.id
//							   WHERE
//								  c.ConversationId = :ConversationId
//							   GROUP BY
//								  b.MessageId
//							)
//							d
//							ON a.Id = d.MessageId
//						 INNER JOIN
//							facebookuids e
//							ON a.
//					  from
//						 = e.id
//					  WHERE
//						 a.ConversationId = :ConversationId
//					  ORDER BY
//						 a.CreatedTime DESC
//					  LIMIT :Limit OFFSET :Offset
//				   )
//				   z
//				GROUP BY
//				   z.conversationid
//			`
//
//		parentQuery := `
//			SELECT
//				json_build_object(
//					'page_id', f.pageid,
//					'id', f.id,
//					'can_comment', f.cancomment,
//					'comment_id', f.commentid,
//					'scoped_thread_key', f.ScopedthreadKey,
//					'created_time', f.createdtime,
//					'can_comment', f.cancomment,
//					'can_hide', f.canhide,
//					'can_like', f.canlike,
//					'can_reply', f.canreply,
//					'can_remove', f.canremove,
//					'can_reply_privately', f.canreplyprivately,
//					'private_reply_conversation', f.privatereplyconversation,
//					'is_hidden', f.ishidden,
//					'is_private', f.isprivate,
//					'seen', f.seen,
//					'last_seen_by', f.lastseenby,
//					'unread_count', f.unreadcount,
//					'snippet', f.snippet,
//					'last_user_message_at', f.LastUserMessageAt,
//					'replied', f.Replied
//				) as data,
//				g.messages,
//				COALESCE(NULLIF(j.notes::TEXT, '[null]'), '[]')::JSON as notes
//			FROM facebookconversations f INNER JOIN (` + childQuery +
//			`
//				) g ON g.conversationid = f.id
//				LEFT OUTER JOIN (
//						SELECT
//							conversationid ,
//							json_agg(conversationnotes.*) as notes
//						FROM conversationnotes
//						WHERE conversationid = :ConversationId
//						GROUP BY conversationid
//				) j ON j.conversationid = f.id
//			`
//
//		//if offset == 0 {
//		//	query = parentQuery
//		//} else {
//		//	query = parentQuery
//		//}
//
//		if err := fs.GetReplica().SelectOne(&data, parentQuery, map[string]interface{}{"ConversationId": conversationId, "Limit": limit, "Offset": offset}); err != nil {
//			//fmt.Println("err", err.Error())
//			//result.Data = nil
//			////result.Err = model.NewAppError("SqlTeamStore.GetTeamsByUserId", "store.sql_team.get_all.app_error", nil, err.Error(), http.StatusInternalServerError)
//			//return
//		}
//		result.Data = data
//	})
//}

func (fs sqlFacebookConversationStore) getConversations(pageIds string, conversationId string, offset, limit int) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		if limit > 250 {
			result.Err = model.NewAppError("sqlFacebookConversationStore.GetMessagesByConversationId", "store.sql_conversation.get_messages.app_error", nil, "conversationId="+conversationId, http.StatusBadRequest)
			return
		}

		query := `SELECT * FROM facebookconversations `

		if len(pageIds) > 0 {
			s := strings.Split(pageIds, ",")
			inPage := ` WHERE facebookconversations.pageid IN (`
			i := 0
			for _, id := range s  {
				inPage += "'" + id + "'"
				if i < cap(s) - 1 {
					inPage += ","
				}
				i++
			}
			inPage += ")"

			query += inPage
		} else if len(conversationId) > 0 {
			query += ` WHERE facebookconversations.id = :ConversationId `
		}

		query += `	ORDER BY
						facebookconversations.Seen,
						facebookconversations.UpdatedTime DESC
					LIMIT :Limit OFFSET :Offset
				`
		var data []*model.FacebookConversation

		if _, err := fs.GetReplica().Select(&data, query, map[string]interface{}{"Limit": limit, "Offset": offset, "ConversationId": conversationId}); err != nil {
			//fmt.Println(err.Error())
			result.Err = model.NewAppError("SqlTeamStore.getConversations", "store.sqlFacebookConversationStore.getConversations.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		}
		result.Data = data
	})
}

//func (fs sqlFacebookConversationStore) getConversations(pageIds string, conversationId string, offset, limit int) store.StoreChannel {
//	query1 := `
//		SELECT
//			json_build_object(
//				'id', a.id,
//				'page_id', a.PageId,
//				'type', a.Type,
//				'snippet', a.Snippet,
//				'updated_time', a.UpdatedTime,
//				'seen', a.Seen,
//				'replied', a.Replied,
//				'unread_count', a.UnreadCount
//			) as data, json_build_object('id', b.id, 'name', b.Name) as from,
//			COALESCE(NULLIF(d.tags::TEXT, '[null]'), '[]')::JSON as tags
//		FROM
//			FacebookConversations a INNER JOIN FacebookUids b
//		ON
//			a.From = b.Id
//		LEFT OUTER JOIN (
//			SELECT b.ConversationId, json_agg(json_build_object('id', c.id, 'color', c.color, 'name', c.name)) as tags
//			FROM ConversationTags b INNER JOIN PageTags c ON b.TagId = c.id
//			GROUP BY b.ConversationId
//			) d ON  d.ConversationId = a.id
//	`
//
//	query2 :=
//		`	ORDER BY
//			a.Seen,
//			a.UpdatedTime DESC
//		LIMIT :Limit OFFSET :Offset
//	`
//	var query string
//
//	if len(conversationId) > 0 {
//		query = query1 + ` WHERE a.id = :ConversationId `
//	} else if len(pageIds) > 0 {
//		s := strings.Split(pageIds, ",")
//		inPage := ` WHERE a.PageId IN (`
//		i := 0
//		for _, id := range s  {
//			inPage += "'" + id + "'"
//			if i < cap(s) - 1 {
//				inPage += ","
//			}
//			i++
//		}
//		inPage += ")"
//
//		query = query1 + inPage + query2
//	} else {
//		query = query1 + query2
//	}
//
//	return store.Do(func(result *store.StoreResult) {
//
//		if limit > 250 {
//			result.Err = model.NewAppError("sqlFacebookConversationStore.GetMessagesByConversationId", "store.sql_conversation.get_messages.app_error", nil, "conversationId="+conversationId, http.StatusBadRequest)
//			return
//		}
//
//		var data []*model.ConversationResponse
//
//		if _, err := fs.GetReplica().Select(&data, query, map[string]interface{}{"Limit": limit, "Offset": offset, "ConversationId": conversationId}); err != nil {
//			//fmt.Println(err.Error())
//			result.Err = model.NewAppError("SqlTeamStore.getConversations", "store.sqlFacebookConversationStore.getConversations.app_error", nil, err.Error(), http.StatusInternalServerError)
//			return
//		}
//		result.Data = data
//	})
//}

func (s *sqlFacebookConversationStore) AnalyticsConversationCountsByDay(pageId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		query :=
			`SELECT
			        DATE(FacebookConversations.UpdatedTime) AS Name,
			        COUNT(FacebookConversations.Id) AS Value
			    FROM FacebookConversations`

		if len(pageId) > 0 {
			query += " INNER JOIN Fanpages ON FacebookConversations.PageId = Fanpages.PageId AND Fanpages.PageId = :PageId AND"
		} else {
			query += " WHERE"
		}

		query += ` FacebookConversations.UpdatedTime::date <= :EndTime
			            AND FacebookConversations.UpdatedTime::date >= :StartTime
			GROUP BY DATE(FacebookConversations.UpdatedTime)
			ORDER BY Name DESC
			LIMIT 30`

		if s.DriverName() == model.DATABASE_DRIVER_POSTGRES {
			query =
				`SELECT
					TO_CHAR(DATE(FacebookConversations.UpdatedTime), 'YYYY-MM-DD') AS Name, Count(FacebookConversations.Id) AS Value
				FROM FacebookConversations`

			if len(pageId) > 0 {
				query += " INNER JOIN Fanpages ON FacebookConversations.PageId = Fanpages.PageId  AND Fanpages.PageId = :PageId AND"
			} else {
				query += " WHERE"
			}

			query += ` FacebookConversations.UpdatedTime <= :EndTime
				            AND FacebookConversations.UpdatedTime >= :StartTime
				GROUP BY DATE(FacebookConversations.UpdatedTime)
				ORDER BY Name DESC
				LIMIT 30`
		}

		end := utils.EndOfDay(utils.Yesterday()).Format("2006-01-02T15:04:05-0700")
		fmt.Println("end", end)
		start := utils.StartOfDay(utils.Yesterday().AddDate(0, 0, -31)).Format("2006-01-02T15:04:05-0700")

		var rows model.AnalyticsRows
		_, err := s.GetReplica().Select(
			&rows,
			query,
			map[string]interface{}{"PageId": pageId, "StartTime": start, "EndTime": end})
		if err != nil {
			result.Err = model.NewAppError("sqlFacebookConversationStore.AnalyticsConversationCountsByDay", "store.sql_post.analytics_posts_count_by_day.app_error", nil, err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = rows
		}
	})
	//return store.Do(func(result *store.StoreResult) {
	//	query :=
	//		`SELECT
	//		        DATE(FROM_UNIXTIME(FacebookConversations.CreateAt / 1000)) AS Name,
	//		        COUNT(FacebookConversations.Id) AS Value
	//		    FROM FacebookConversations`
	//
	//	if len(pageId) > 0 {
	//		query += " INNER JOIN Fanpages ON FacebookConversations.PageId = Fanpages.PageId AND Fanpages.PageId = :PageId AND"
	//	} else {
	//		query += " WHERE"
	//	}
	//
	//	query += ` FacebookConversations.CreateAt <= :EndTime
	//		            AND FacebookConversations.CreateAt >= :StartTime
	//		GROUP BY DATE(FROM_UNIXTIME(FacebookConversations.CreateAt / 1000))
	//		ORDER BY Name DESC
	//		LIMIT 30`
	//
	//	if s.DriverName() == model.DATABASE_DRIVER_POSTGRES {
	//		query =
	//			`SELECT
	//				TO_CHAR(DATE(TO_TIMESTAMP(FacebookConversations.CreateAt / 1000)), 'YYYY-MM-DD') AS Name, Count(FacebookConversations.Id) AS Value
	//			FROM FacebookConversations`
	//
	//		if len(pageId) > 0 {
	//			query += " INNER JOIN Fanpages ON FacebookConversations.PageId = Fanpages.PageId  AND Fanpages.PageId = :PageId AND"
	//		} else {
	//			query += " WHERE"
	//		}
	//
	//		query += ` FacebookConversations.CreateAt <= :EndTime
	//			            AND FacebookConversations.CreateAt >= :StartTime
	//			GROUP BY DATE(TO_TIMESTAMP(FacebookConversations.CreateAt / 1000))
	//			ORDER BY Name DESC
	//			LIMIT 30`
	//	}
	//
	//	end := utils.MillisFromTime(utils.EndOfDay(utils.Yesterday()))
	//	start := utils.MillisFromTime(utils.StartOfDay(utils.Yesterday().AddDate(0, 0, -31)))
	//
	//	var rows model.AnalyticsRows
	//	_, err := s.GetReplica().Select(
	//		&rows,
	//		query,
	//		map[string]interface{}{"PageId": pageId, "StartTime": start, "EndTime": end})
	//	if err != nil {
	//		result.Err = model.NewAppError("sqlFacebookConversationStore.AnalyticsConversationCountsByDay", "store.sql_post.analytics_posts_count_by_day.app_error", nil, err.Error(), http.StatusInternalServerError)
	//	} else {
	//		result.Data = rows
	//	}
	//})
}

func (fs sqlFacebookConversationStore) GetConversations(pageIds string, offset, limit int) store.StoreChannel {
	return fs.getConversations(pageIds, "", offset, limit)
}

func (fs sqlFacebookConversationStore) GetConversationById(id string) store.StoreChannel {
	return fs.getConversations("", id, 0, 1)
}

func (fs sqlFacebookConversationStore) UpdateSeen(id string, pageId string, userId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		//curTime := model.GetMillis()

		if _, err := fs.GetMaster().Exec("UPDATE FacebookConversations SET Seen = true, LastSeenBy = :UserId WHERE Id = :Id", map[string]interface{}{"Id": id, "UserId": userId}); err != nil {
			result.Err = model.NewAppError("sqlFacebookConversationStore.UpdateSeen", "store.sql_user.update_last_picture_update.app_error", nil, "user_id="+userId, http.StatusInternalServerError)
		} else {
			result.Data = id
		}
	})
}

func (fs sqlFacebookConversationStore) UpdateUnSeen(id string, pageId string, userId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		//curTime := model.GetMillis()

		if sqlResult, err := fs.GetMaster().Exec("UPDATE FacebookConversations SET Seen = false, LastSeenBy = :UserId WHERE Id = :Id RETURNING *", map[string]interface{}{"Id": id, "UserId": userId}); err != nil {
			result.Err = model.NewAppError("sqlFacebookConversationStore.UpdateSeen", "store.sql_user.update_last_picture_update.app_error", nil, "user_id="+userId, http.StatusInternalServerError)
		} else {
			if _, err1 := sqlResult.RowsAffected(); err1 != nil {
				result.Err = model.NewAppError("sqlFacebookConversationStore.PermanentDeleteBatchForChannel", "store.sql_channel_member_history.permanent_delete_batch.app_error", nil, err.Error(), http.StatusInternalServerError)
			} else {
				if updated, err := fs.GetReplica().Get(model.FacebookConversation{}, id); err != nil {
					result.Err = model.NewAppError("sqlFacebookConversationStore.PermanentDeleteBatchForChannel", "store.sql_channel_member_history.permanent_delete_batch.app_error", nil, err.Error(), http.StatusInternalServerError)
				} else if updated == nil {
					result.Err = model.NewAppError("SqlUserStore.Get", store.MISSING_ACCOUNT_ERROR, nil, "user_id="+id, http.StatusNotFound)
				} else {
					result.Data = updated.(*model.FacebookConversation)
				}
			}
		}
	})
}

func (fs sqlFacebookConversationStore) AddMessage(message *model.FacebookConversationMessage, shouldUpdateConversation bool, isFromPage bool) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		message.PreSave()
		if err := fs.GetMaster().Insert(message); err != nil {
			// cần sửa
			result.Err = model.NewAppError("sqlFanpageStore.Save", "store.sql_fanpage.save.app_error", nil, "page_id="+message.Id+", "+err.Error(), http.StatusInternalServerError)
		} else {
			if shouldUpdateConversation {
				var snippet string
				if len(message.Message) == 0 && message.HasAttachments {
					snippet = "[Attachment]"
				} else {
					snippet = utils.GetSnippet(message.Message)
				}
				fs.UpdateConversation(message.ConversationId, snippet, isFromPage, message.CreatedTime, 1, message.CreatedTime)
			}

			result.Data = message
		}
	})
}

func (fs sqlFacebookConversationStore) AddImage(image *model.FacebookAttachmentImage) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		image.PreSave()
		// validate
		// kiểm tra nếu image

		if err := fs.GetMaster().Insert(image); err != nil {
			// cần sửa
			result.Err = model.NewAppError("sqlFanpageStore.Save", "store.sql_fanpage.save.app_error", nil, "page_id="+image.Id+", "+err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = image
			//
		}
	})
}

func (fs sqlFacebookConversationStore) GetPageConversationBySenderId(pageId, senderId, conversationType string) store.StoreChannel {

	return store.Do(func(result *store.StoreResult) {
		var query sq.SelectBuilder

		if len(conversationType) > 0 {
			if conversationType != "comment" && conversationType != "message" {
				result.Err = model.NewAppError("sqlFacebookConversationStore.GetPageConversationBySenderId", "store.sql_fanpage.save.app_error", nil, "page_id="+pageId, http.StatusInternalServerError)
				return
			}
			if conversationType == "message" {
				query = fs.conversationQuery.
					Where(map[string]interface{}{
						"FacebookConversations.PageId": pageId,
						"FacebookConversations.PageScopeId": senderId,
						"FacebookConversations.Type": conversationType,
					})
			}
			// TODO: other types?
		} else {
			query = fs.conversationQuery.
				Where(map[string]interface{}{
					"FacebookConversations.PageId": pageId,
					"FacebookConversations.From": senderId,
				})
		}

		conversations := []*model.FacebookConversation{}

		queryString, args, err := query.ToSql()

		if err != nil {
			result.Err = model.NewAppError("sqlFacebookConversationStore.GetPageConversationBySenderId", "store.sql_conversation.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, err := fs.GetReplica().Select(&conversations, queryString, args...); err != nil {
			result.Err = model.NewAppError("sqlFacebookConversationStore.GetPageConversationBySenderId", "store.sql_conversation.get_attachments.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		}

		result.Data = conversations
	})
}
//
//func (fs sqlFacebookConversationStore) GetConversationTypeMessage(userId string, pageId string, updatedTime string) store.StoreChannel {
//	return store.Do(func(result *store.StoreResult) {
//		var conversation *model.FacebookConversation
//
//		query := "SELECT a.* FROM facebookconversations a WHERE a.type = 'message' AND a.from = :userId AND a.pageid = :pageId"
//
//		err := fs.GetReplica().SelectOne(&conversation, query, map[string]interface{}{"userId": userId, "pageId": pageId})
//		if err != nil {
//			if err == sql.ErrNoRows {
//				var newConversation model.FacebookConversation
//				newConversation.Type = "message"
//				newConversation.UpdatedTime = updatedTime
//				newConversation.PageId = pageId
//				newConversation.From = userId
//				newConversation.PreSave()
//
//				err := fs.GetMaster().Insert(&newConversation)
//				if err != nil {
//					fmt.Println(err)
//				} else {
//					result.Data = newConversation
//				}
//			} else {
//				fmt.Println(err)
//			}
//		} else {
//			result.Data = conversation
//		}
//	})
//}

func (fs sqlFacebookConversationStore) InsertConversationFromCommentIfNeed(parentId string, commentId string, pageId string, postId string, userId string, time string, message string) store.StoreChannel {
	unReadCount := 0
	replied := true
	var lastUserMessage string
	if pageId != userId {
		unReadCount = 1
		replied = false
		lastUserMessage = time
	}

	isNew := false

	return store.Do(func(result *store.StoreResult) {

		var query string
		// tìm kiếm hoặc thêm hội thoại
		// nếu parentId != "" tức là user reply vào 1 comment gốc =>

		// Đầu tiên kiểm tra xem nếu cũng trong bài viết này, user đã từng comment trước đó
		// Cần nhóm tất cả comment gốc của 1 user trong 1 bài viết vào 1 hội thoại duy nhất
		// nhưng nếu user reply vào 1 comment của người khác thì comment này sẽ thuộc về hội thoại của người đó

		conversation := model.FacebookConversation{}

		if len(parentId) == 0 || parentId == postId {
			// Nhóm tất cả comment của cùng 1 user trên 1 post vào 1 hội thoại duy nhất
			query = "SELECT a.* FROM facebookconversations a WHERE a.type = 'comment' AND a.from = :UserId AND a.postid = :PostId"

			err := fs.GetReplica().SelectOne(&conversation, query, map[string]interface{}{"UserId": userId, "PostId": postId})
			if err != nil {
				if err == sql.ErrNoRows {
					// User chưa từng comment vào bài viết này trước đó => Tạo hội thoại mới
					newConversation := model.FacebookConversation {
						Type: "comment",
						PageId: pageId,
						PostId: postId,
						CommentId: commentId,
						From: userId,
						CreatedTime: time,
						UpdatedTime: time,
						Snippet: message,
						CanComment: true,
						CanLike: true,
						CanReply: true,
						CanRemove: true,
						CanHide: true,
						CanReplyPrivately: true,
						UnreadCount: unReadCount,
						Replied: replied,
						Seen: replied,
						LastUserMessageAt: lastUserMessage,
					}

					isNew = true

					newConversation.PreSave()
					newConversation.UpdatedTime = time

					err := fs.GetMaster().Insert(&newConversation)
					if err != nil {
						fmt.Println("loi 1")
						fmt.Println(err)
						result.Err = model.NewAppError("sqlFacebookConversationStore.InsertConversationFromCommentIfNeed", "not_found", nil, "page_id="+pageId, http.StatusNotFound)
					} else {
						result.Data = &model.UpsertConversationResult{
							IsNew: isNew,
							Data: &newConversation,
						}
					}
				} else {
					result.Err = model.NewAppError("sqlFacebookConversationStore.InsertConversationFromCommentIfNeed", "not_found", nil, "page_id="+pageId, http.StatusNotFound)
				}
			} else {
				result.Data = &model.UpsertConversationResult{
					IsNew: isNew,
					Data: &conversation,
				}
			}

		} else if len(parentId) > 0 {
			//461787374317076_544889279340218
			// Nếu parent_id chứa phần 1 là page_id chứng tỏ user comment vào subPost của 1 post gốc (thường là comment vào ảnh con trong 1 bài viết
			// gồm nhiều ảnh

			// trong trường hợp này nếu tìm kiếm theo parent_id sẽ không có kết quả, khi đó ta chỉ cần nhóm hội thoại của cùng 1 user là được
			s := strings.Split(parentId, "_")

			if len(s) < 2 {
				result.Err = model.NewAppError("sqlFacebookConversationStore.InsertConversationFromCommentIfNeed", "not_found", nil, "page_id="+pageId, http.StatusNotFound)
			}

			if s[0] == pageId {
				// tìm hội thoại của user này
				query = "SELECT a.* FROM facebookconversations a WHERE a.type = 'comment' AND a.from = :UserId AND a.postid = :PostId"

				err := fs.GetReplica().SelectOne(&conversation, query, map[string]interface{}{"UserId": userId, "PostId": postId})
				if err != nil {
					if err == sql.ErrNoRows {
						// User chưa từng comment vào bài viết này trước đó => Tạo hội thoại mới
						newConversation := model.FacebookConversation {
							Type: "comment",
							PageId: pageId,
							PostId: postId,
							CommentId: commentId,
							From: userId,
							CreatedTime: time,
							UpdatedTime: time,
							Snippet: message,
							CanComment: true,
							CanLike: true,
							CanReply: true,
							CanRemove: true,
							CanHide: true,
							CanReplyPrivately: true,
							UnreadCount: unReadCount,
							Replied: replied,
							Seen: replied,
							LastUserMessageAt: lastUserMessage,
						}

						isNew = true

						newConversation.PreSave()
						newConversation.UpdatedTime = time

						err := fs.GetMaster().Insert(&newConversation)
						if err != nil {
							result.Err = model.NewAppError("sqlFacebookConversationStore.InsertConversationFromCommentIfNeed", "not_found", nil, "page_id="+pageId, http.StatusNotFound)
						} else {
							result.Data = &model.UpsertConversationResult{
								IsNew: isNew,
								Data: &newConversation,
							}
						}
					} else {
						result.Err = model.NewAppError("sqlFacebookConversationStore.InsertConversationFromCommentIfNeed", "not_found", nil, "page_id="+pageId, http.StatusNotFound)
					}
				} else {
					result.Data = &model.UpsertConversationResult{
						IsNew: isNew,
						Data: &conversation,
					}
				}

			} else {
				// có 2 trường hợp xảy ra:
				// 1: User comment vào parent comment của chính họ
				// 2: User comment vào parent comment của người khác
				query = "SELECT a.* FROM facebookconversations a WHERE a.type = 'comment' AND a.commentid = :ParentId AND a.postid = :PostId"

				err := fs.GetReplica().SelectOne(&conversation, query, map[string]interface{}{"UserId": userId, "PostId": postId, "ParentId": parentId})
				if err != nil {
					if err == sql.ErrNoRows {
						// User comment vào 1 comment cha nhưng không phải comment gốc của hội thoại
						// Ta sẽ tìm kiếm hội thoại dựa trên comment cha
						messageQuery := "SELECT a.* FROM facebookconversationmessages a WHERE a.type = 'comment' AND a.commentid = :CommentId"
						conversationMessage := model.FacebookConversationMessage{}
						err2 := fs.GetReplica().SelectOne(&conversationMessage, messageQuery, map[string]interface{}{"CommentId": parentId})
						if err2 != nil {
							if err2 == sql.ErrNoRows {
								fmt.Println("khong tim thay conversationMessage nao tu: ", parentId)
								newConversation2 := model.FacebookConversation {
									Type: "comment",
									PageId: pageId,
									PostId: postId,
									CommentId: commentId,
									From: userId,
									CreatedTime: time,
									UpdatedTime: time,
									Snippet: message,
									CanComment: true,
									CanLike: true,
									CanReply: true,
									CanRemove: true,
									CanHide: true,
									CanReplyPrivately: true,
									UnreadCount: unReadCount,
									Replied: replied,
									Seen: replied,
									LastUserMessageAt: lastUserMessage,
								}

								isNew = true


								newConversation2.PreSave()
								newConversation2.UpdatedTime = time

								err := fs.GetMaster().Insert(&newConversation2)
								if err != nil {
									fmt.Println("loi 1")
									fmt.Println(err)
								} else {
									result.Data = &model.UpsertConversationResult{
										IsNew: isNew,
										Data: &newConversation2,
									}
								}
							} else {
								fmt.Println("no rows in result set")
								fmt.Println(err2)
							}

						} else {
							// tim thay tin nhan, => get hoi thoai co conversation_id = conversation_id cua tin nhan nay
							//fmt.Println("tim thay tin nhan, => get hoi thoai co conversation_id = conversation_id cua tin nhan nay")
							//result.Data = <-fs.Get(conversationMessage.ConversationId)

							if obj, err := fs.GetReplica().Get(model.FacebookConversation{}, conversationMessage.ConversationId); err != nil {
								//result.Err = model.NewAppError("sqlFacebookConversationStore.Get", "store.sql_fanpage.get.app_error", nil, "fanpage_id="+id+", "+err.Error(), http.StatusInternalServerError)
							} else if obj == nil {
								result.Err = model.NewAppError("sqlFacebookConversationStore.Get", store.MISSING_FANPAGE_ERROR, nil, "page_id="+pageId, http.StatusNotFound)
							} else {
								result.Data = &model.UpsertConversationResult{
									IsNew: isNew,
									Data: obj.(*model.FacebookConversation),
								}
							}
						}
					} else {
						// Một lỗi xảy ra và không biết được lỗi gì
						result.Err = model.NewAppError("sqlFacebookConversationStore.InsertConversationFromCommentIfNeed", "not_found", nil, "page_id="+pageId, http.StatusNotFound)
					}
				} else {
					result.Data = &model.UpsertConversationResult{
						IsNew: isNew,
						Data: &conversation,
					}
				}
			}
		}
	})
}

func (fs *sqlFacebookConversationStore) Search(term string, pageIds string, limit, offset int) store.StoreChannel {

	var inPages string
	if len(pageIds) > 0 {
		inPages += " WHERE a.pageid IN " + pageIds + " "
	}
	query := `
			SELECT 
					json_build_object(
						'id', a.id, 
						'page_id', a.PageId,
						'post_id', a.PostId,
						'type', a.Type,
						'message', a.Message,
						'snippet', a.Snippet, 
						'updated_time', a.UpdatedTime, 
						'comment_id', a.CommentId, 
						'scoped_thread_key', a.ScopedthreadKey, 
						'seen', a.Seen, 
						'last_seen_by', a.LastseenBy, 
						'unread_count', a.UnreadCount, 
						'can_comment', a.CanComment, 
						'can_like', a.CanLike, 
						'can_hide', a.CanHide, 
						'can_reply', a.CanReply, 
						'can_reply_privately', a.CanReplyPrivately, 
						'is_hidden', a.IsHidden, 
						'is_private', a.IsPrivate, 
						'private_reply_conversation', a.PrivatereplyConversation
					) as data, json_build_object('id', b.id, 'name', b.Name) as from,
					COALESCE(NULLIF(d.tags::TEXT, '[null]'), '[]')::JSON as tags
				FROM 
					FacebookConversations a INNER JOIN FacebookUids b
				ON 
					a.From = b.Id
				LEFT OUTER JOIN ( -- 1 và 2 và 3 chỉ trong 1 câu query nhé
					SELECT b.ConversationId, json_agg(c.id) as tags
					FROM ConversationTags b INNER JOIN PageTags c ON b.TagId = c.id
					GROUP BY b.ConversationId
					) d ON  d.ConversationId = a.id
				INNER JOIN (
					SELECT DISTINCT g.conversationid FROM facebookconversationmessages g INNER JOIN facebookuids h
					ON h.id = g.from
					WHERE g.message ILIKE '%` + term + `%' OR h.name ILIKE '%` + term + `%'
				) f ON f.conversationid = a.id ` + inPages + `
			
			ORDER BY
					a.UpdatedTime DESC
				LIMIT :Limit OFFSET :Offset
		`

	return store.Do(func(result *store.StoreResult) {
		var data []*model.ConversationResponse
		if _, err := fs.GetReplica().Select(&data, query, map[string]interface{}{"Limit": limit, "Offset": offset}); err != nil {
			result.Err = model.NewAppError("sqlFacebookConversationStore.Search", "store.sql_team.get_all.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		}
		result.Data = data
	})
}

func (s *sqlFacebookConversationStore) OverwriteMessage(message *model.FacebookConversationMessage) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		//message.UpdateAt = model.GetMillis()

		//var maxPostSize int
		//if result := <-s.GetMaxPostSize(); result.Err != nil {
		//	result.Err = model.NewAppError("SqlPostStore.Save", "store.sql_post.overwrite.app_error", nil, "id="+post.Id+", "+result.Err.Error(), http.StatusInternalServerError)
		//	return
		//} else {
		//	maxPostSize = result.Data.(int)
		//}
		//
		//if result.Err = post.IsValid(maxPostSize); result.Err != nil {
		//	return
		//}

		if _, err := s.GetMaster().Update(message); err != nil {
			result.Err = model.NewAppError("sqlFacebookConversationStore.OverwriteMessage", "store.sql_message.overwrite.app_error", nil, "id="+message.Id+", "+err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = message
		}
	})
}

//GetFacebookAttachmentByIds
func (us sqlFacebookConversationStore) GetFacebookAttachmentByIds(ids []string, allowFromCache bool) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		attachments := []*model.FacebookAttachmentImage{}
		remainingIds := make([]string, 0)

		if allowFromCache {
			for _, id := range ids {
				if cacheItem, ok := attachmentByIdsCache.Get(id); ok {
					u := &model.FacebookAttachmentImage{}
					*u = *cacheItem.(*model.FacebookAttachmentImage)
					attachments = append(attachments, u)
				} else {
					remainingIds = append(remainingIds, id)
				}
			}
			if us.metrics != nil {
				us.metrics.AddMemCacheHitCounter("Attachment By Ids", float64(len(attachments)))
				us.metrics.AddMemCacheMissCounter("Attachment By Ids", float64(len(remainingIds)))
			}
		} else {
			remainingIds = ids
			if us.metrics != nil {
				us.metrics.AddMemCacheMissCounter("Attachment By Ids", float64(len(remainingIds)))
			}
		}

		// If everything came from the cache then just return
		if len(remainingIds) == 0 {
			result.Data = attachments
			return
		}

		query := us.attachmentsQuery.
			Where(map[string]interface{}{
				"facebookattachmentimages.Id": remainingIds,
			})

		queryString, args, err := query.ToSql()

		if err != nil {
			result.Err = model.NewAppError("sqlFacebookConversationStore.GetFacebookAttachmentByIds", "store.sql_conversation.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, err := us.GetReplica().Select(&attachments, queryString, args...); err != nil {
			result.Err = model.NewAppError("sqlFacebookConversationStore.GetFacebookAttachmentByIds", "store.sql_conversation.get_attachments.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		}

		for _, u := range attachments {
			//u.Sanitize(map[string]bool{})

			cpy := &model.FacebookAttachmentImage{}
			*cpy = *u
			attachmentByIdsCache.AddWithExpiresInSecs(cpy.Id, cpy, ATTACHMENT_BY_IDS_CACHE_SEC)
		}

		result.Data = attachments
	})
}

func (s sqlFacebookConversationStore) GetPageMessageByMid(pageId, mid string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var message model.FacebookConversationMessage
		err := s.GetReplica().SelectOne(&message,
			`SELECT
				FacebookConversationMessages.*
			FROM
				FacebookConversationMessages
			WHERE
				PageId = :PageId
                AND MessageId = :MessageId`,
			map[string]interface{}{"PageId": pageId, "MessageId": mid})

		if err != nil {
			result.Err = model.NewAppError("sqlFacebookConversationStore.GetPageMessageByMid", "store.sql_message.get_unread.app_error", nil, "pageId="+pageId+" "+err.Error(), http.StatusInternalServerError)
			if err == sql.ErrNoRows {
				result.Err.StatusCode = http.StatusNotFound
			}
		} else {
			result.Data = &message
		}
	})
}

func (fs sqlFacebookConversationStore) UpdateMessageSent(messageId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		query := "UPDATE FacebookConversationMessages SET Delivered = true WHERE MessageId = :MessageId"
		_, err := fs.GetMaster().Exec(query, map[string]interface{}{"MessageId": messageId})
		if err != nil {
			result.Err = model.NewAppError("sqlFacebookConversationStore.UpdateMessageSent", "store.sql_conversations.save.app_error", nil, "message_id="+messageId+", "+err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = nil
		}
	})
}

func (fs sqlFacebookConversationStore) UpdateReadWatermark(conversationId, pageId string, timestamp int64) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		query := "UPDATE FacebookConversations SET ReadWatermark = :ReadWatermark WHERE Id = :Id AND PageId = :PageId"
		_, err := fs.GetMaster().Exec(query, map[string]interface{}{"ReadWatermark": timestamp, "Id": conversationId, "PageId": pageId})
		if err != nil {
			result.Err = model.NewAppError("sqlFacebookConversationStore.UpdateReadWatermark", "store.sql_conversations.save.app_error", nil, "conversation_id="+conversationId+", "+err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = nil
		}
	})
}

func (fs sqlFacebookConversationStore) UpdateCommentByCommentId(commentId, newText string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		updatedTime := time.Now().UnixNano() / int64(time.Millisecond)
		updateQuery := "UPDATE FacebookConversationMessages SET Message = :Message, EditAt = :EditTime WHERE Type = 'comment' AND CommentId = :CommentId"
		_, err := fs.GetMaster().Exec(updateQuery, map[string]interface{}{"CommentId": commentId, "Message": newText, "EditTime": updatedTime})
		if err == nil {
			messageQuery := "SELECT a.* FROM FacebookConversationMessages a WHERE a.Type = 'comment' AND a.CommentId = :CommentId"
			conversationMessage := model.FacebookConversationMessage{}
			err2 := fs.GetReplica().SelectOne(&conversationMessage, messageQuery, map[string]interface{}{"CommentId": commentId})
			if err2 != nil {
				result.Err = model.NewAppError("sqlFacebookConversationStore.UpdateCommentByCommentId", "store.sql_conversations.save.app_error", nil, "comment_id="+commentId+", "+err2.Error(), http.StatusInternalServerError)
			} else {
				result.Data = &conversationMessage
			}
		} else {
			result.Err = model.NewAppError("sqlFacebookConversationStore.UpdateCommentByCommentId", "store.sql_conversations.save.app_error", nil, "comment_id="+commentId+", "+err.Error(), http.StatusInternalServerError)
		}
	})
}

func (fs sqlFacebookConversationStore) DeleteCommentByCommentId(commentId, appScopedUserId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		updatedTime := time.Now().UnixNano() / int64(time.Millisecond)
		var deleteQuery string
		if len(appScopedUserId) > 0 {
			deleteQuery = "UPDATE FacebookConversationMessages SET DeleteAt = :DeleteTime, DeleteBy = :DeleteBy WHERE Type = 'comment' AND CommentId = :CommentId"
		} else {
			deleteQuery = "UPDATE FacebookConversationMessages SET DeleteAt = :DeleteTime WHERE Type = 'comment' AND CommentId = :CommentId"
		}
		_, err := fs.GetMaster().Exec(deleteQuery, map[string]interface{}{"CommentId": commentId, "DeleteBy": appScopedUserId, "DeleteTime": updatedTime})
		if err == nil {
			messageQuery := "SELECT a.* FROM FacebookConversationMessages a WHERE a.Type = 'comment' AND a.CommentId = :CommentId"
			conversationMessage := model.FacebookConversationMessage{}
			err2 := fs.GetReplica().SelectOne(&conversationMessage, messageQuery, map[string]interface{}{"CommentId": commentId})
			if err2 != nil {
				result.Err = model.NewAppError("sqlFacebookConversationStore.DeleteCommentByCommentId", "store.sql_conversations.save.app_error", nil, "comment_id="+commentId+", "+err2.Error(), http.StatusInternalServerError)
			} else {
				result.Data = &conversationMessage
			}
		} else {
			result.Err = model.NewAppError("sqlFacebookConversationStore.DeleteCommentByCommentId", "store.sql_conversations.save.app_error", nil, "comment_id="+commentId+", "+err.Error(), http.StatusInternalServerError)
		}
	})
}
