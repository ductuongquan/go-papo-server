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
	COMMENT_ATTACHMENT_BY_IDS_CACHE_SIZE      = model.COMMENT_ATTACHMENT_CACHE_SIZE
	COMMENT_ATTACHMENT_BY_IDS_CACHE_SEC       = 8 * 60 * 60 // 8 hours
)

var commentAttachmentByIdsCache *utils.Cache = utils.NewLru(COMMENT_ATTACHMENT_BY_IDS_CACHE_SIZE)

func (a *App) GetFacebookAttachmentByIds(ids []string) ([]*model.FacebookAttachmentImage, *model.AppError) {
	result := <-a.Srv.Store.FacebookConversation().GetFacebookAttachmentByIds(ids, true)
	if result.Err != nil {
		return nil, result.Err
	}
	return result.Data.([]*model.FacebookAttachmentImage), nil

	return nil, nil
}

func (app *App) GraphAttachmentTarget(targetId, attachmentType, token string) (*facebookgraph.FacebookError, *model.AppError, *model.TargetItem) {

	if cacheItem, ok := commentAttachmentByIdsCache.Get(targetId); ok {
		p := &model.TargetItem{}
		*p = *cacheItem.(*model.TargetItem)
		return nil, nil, p
	}

	var path string

	if attachmentType == "video_inline" {
		path = "/" + targetId + "?fields=picture,format"
	} else {
		path = "/" + targetId + "?fields=height,width,images,picture,can_backdate,can_delete,can_tag,created_time,event,icon,id,link,alt_text"
	}

	var body io.ReadCloser
	var err *facebookgraph.FacebookError
	var aerr *model.AppError

	body, err, aerr = app.request(token, path, "GET")

	if err != nil {
		mlog.Error(fmt.Sprint(err))
		return err, nil, nil
	} else if aerr != nil {
		mlog.Error(fmt.Sprint(aerr))
		return nil, aerr, nil
	} else {
		targetItem := *model.TargetItemFromJson(body)
		targetItem.Type = attachmentType

		cpy := &model.TargetItem{}
		*cpy = targetItem
		commentAttachmentByIdsCache.AddWithExpiresInSecs(targetId, cpy, COMMENT_ATTACHMENT_BY_IDS_CACHE_SEC)

		return nil, nil, &targetItem
	}
}

func (app *App) GraphMessageAttachments(messageId, token string) (*facebookgraph.FacebookError, *model.AppError, *model.MessageAttachmentsResponse) {

	if cacheItem, ok := commentAttachmentByIdsCache.Get(messageId); ok {
		fmt.Println("found message attachments from cache")
		p := &model.MessageAttachmentsResponse{}
		*p = *cacheItem.(*model.MessageAttachmentsResponse)
		return nil, nil, p
	}

	path := "/" + messageId + "?fields=attachments,message"

	var body io.ReadCloser
	var err *facebookgraph.FacebookError
	var aerr *model.AppError

	body, err, aerr = app.request(token, path, "GET")

	if err != nil {
		mlog.Error(fmt.Sprint(err))
		return err, nil, nil
	} else if aerr != nil {
		mlog.Error(fmt.Sprint(aerr))
		return nil, aerr, nil
	} else {
		messageAttachments := *model.MessageAttachmentsResponseFromJson(body)

		cpy := &model.MessageAttachmentsResponse{}
		*cpy = messageAttachments
		commentAttachmentByIdsCache.AddWithExpiresInSecs(messageId, cpy, COMMENT_ATTACHMENT_BY_IDS_CACHE_SEC)

		return nil, nil, &messageAttachments
	}
}

func (app *App) GraphCommentAttachments(commentId, token string) (*facebookgraph.FacebookError, *model.AppError, *model.CommentAttachmentsResponse) {

	if cacheItem, ok := commentAttachmentByIdsCache.Get(commentId); ok {
		fmt.Println("found a comment attachments in cache")
		p := &model.CommentAttachmentsResponse{}
		*p = *cacheItem.(*model.CommentAttachmentsResponse)
		return nil, nil, p
	}

	path := "/" + commentId + "?fields=attachment"

	var body io.ReadCloser
	var err *facebookgraph.FacebookError
	var aerr *model.AppError

	body, err, aerr = app.request(token, path, "GET")

	if err != nil {
		mlog.Error(fmt.Sprint(err))
		return err, nil, nil
	} else if aerr != nil {
		mlog.Error(fmt.Sprint(aerr))
		return nil, aerr, nil
	} else {
		commentAttachments := *model.CommentAttachmentsResponseFromJson(body)

		cpy := &model.CommentAttachmentsResponse{}
		*cpy = commentAttachments
		commentAttachmentByIdsCache.AddWithExpiresInSecs(commentId, cpy, COMMENT_ATTACHMENT_BY_IDS_CACHE_SEC)

		return nil, nil, &commentAttachments
	}
}