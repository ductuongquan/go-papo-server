// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package sqlstore

import (
	"bitbucket.org/enesyteam/papo-server/model"
	"bitbucket.org/enesyteam/papo-server/store"
	"net/http"
)

type sqlFacebookPostStore struct {
	SqlStore
}

func NewSqlFacebookPostStore(sqlStore SqlStore) store.FacebookPostStore {
	fs := &sqlFacebookPostStore{
		SqlStore: sqlStore,
	}

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(model.FacebookPost{}, "FacebookPosts").SetKeys(false, "Id")
		table.ColMap("Id").SetMaxSize(26)

		// cần các thiết lập khác ở đây

	}
	return fs
}

func (fs sqlFacebookPostStore) CreateIndexesIfNotExists() {
	fs.CreateIndexIfNotExists("idx_facebook_posts_post_id", "FacebookPosts", "PostId")
	fs.CreateIndexIfNotExists("idx_facebook_posts_created_time", "FacebookPosts", "CreatedTime")
}

func (fs sqlFacebookPostStore) Save(post *model.FacebookPost) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		post.PreSave()

		if err := fs.GetMaster().Insert(post); err != nil {
			result.Err = model.NewAppError("sqlFanpageStore.Save", "store.sql_fanpage.save.app_error", nil, "page_id="+post.Id+", "+err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = post
		}
	})
}

func (fs sqlFacebookPostStore) Get(postId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		var fbPost *model.FacebookPost

		err := fs.GetReplica().SelectOne(&fbPost, "SELECT * FROM facebookposts WHERE postid = :postId", map[string]interface{}{"postId": postId})
		if err != nil {
			result.Err = model.NewAppError("sqlFanpageStore.Save", "store.sql_fanpage.save.app_error", nil, "post_id="+postId+", "+err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = fbPost
		}
	})
}

func (fs sqlFacebookPostStore) GetPagePosts(pageId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var data []*model.FacebookPostResponse

		query := `
SELECT 
		json_build_object(
			'id', f.id,
			'created_time', f.createdtime,
			'from', f.from,
			'picture', f.picture,
			'updated_time', f.updatedtime,
			'is_hidden', f.ishidden,
			'story', f.story,
			'permalink_url', f.permalinkurl,
			'message', f.message,
			'post_id', f.postid,
			'page_id', f.pageid
		) as data,
		COALESCE(NULLIF(m.attachments::TEXT, '[null]'), '[]')::JSON as attachments
FROM 
	facebookposts f
LEFT OUTER JOIN (
	SELECT 
			postid, json_agg(facebookattachmentimages.*) as attachments
	FROM
			facebookattachmentimages
	GROUP BY postid
)  m ON m.postid = f.postid
WHERE pageid = :PageId`

		if _, err := fs.GetReplica().Select(&data, query, map[string]interface{}{"PageId": pageId}); err != nil {
			result.Err = model.NewAppError("sqlFanpageStore.GetPagePosts", "store.sql_team.get_all.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		}
		result.Data = data
	})
}
