package app

import (
	"bitbucket.org/enesyteam/papo-server/facebook_graph"
	"bitbucket.org/enesyteam/papo-server/mlog"
	"bitbucket.org/enesyteam/papo-server/model"
	"bitbucket.org/enesyteam/papo-server/utils"
	"fmt"
	"io"
)

const (
	POST_BY_IDS_CACHE_SIZE      = model.POST_CACHE_SIZE
	POST_BY_IDS_CACHE_SEC       = 8 * 60 * 60 // 8 hours
)

var postByIdsCache *utils.Cache = utils.NewLru(POST_BY_IDS_CACHE_SIZE)

func (app *App) GetPagePosts(pageId string) ([]*model.FacebookPostResponse, *model.AppError) {
	result := <-app.Srv.Store.FacebookPost().GetPagePosts(pageId)
	if result.Err != nil {
		return nil, result.Err
	}
	return result.Data.([]*model.FacebookPostResponse), nil
}

func (app *App) FetchPost(postId, pageToken string) (*facebookgraph.FacebookError, *model.AppError, *model.Post) {

	if cacheItem, ok := postByIdsCache.Get(postId); ok {
		p := &model.Post{}
		*p = *cacheItem.(*model.Post)
		return nil, nil, p
	}

	path := "/" + postId + "?fields=picture,message,story,attachments,likes.summary(true).limit(0),comments.summary(true).limit(0),created_time,permalink_url"
	var body io.ReadCloser
	var err *facebookgraph.FacebookError
	var aerr *model.AppError

	body, err, aerr = app.request(pageToken, path, "GET")

	if err != nil {
		mlog.Error(fmt.Sprint(err))
		return err, nil, nil
	} else if aerr != nil {
		mlog.Error(fmt.Sprint(aerr))
		return nil, aerr, nil
	} else {
		post := *model.PostFromJson(body)

		cpy := &model.Post{}
		*cpy = post
		postByIdsCache.AddWithExpiresInSecs(cpy.Id, cpy, POST_BY_IDS_CACHE_SEC)

		return nil, nil, &post
	}
}