// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"bitbucket.org/enesyteam/papo-server/mlog"
	"bitbucket.org/enesyteam/papo-server/model"
	"fmt"
	"net/http"
)

func (app *App) GetFanpage(fanpageId string) (*model.Fanpage, *model.AppError) {
	result := <-app.Srv.Store.Fanpage().Get(fanpageId)
	if result.Err != nil {
		return nil, result.Err
	}
	return result.Data.(*model.Fanpage), nil
}

func (app *App) GetFanpageByPageId(pageId string) (*model.Fanpage, *model.AppError) {
	result := <-app.Srv.Store.Fanpage().GetFanpageByPageID(pageId)
	if result.Err != nil {
		return nil, result.Err
	}
	return result.Data.(*model.Fanpage), nil
}

func (app *App) GetFanpagesForUser(userId string) ([]*model.FanpageIncludeInitResult, *model.AppError) {
	result := <-app.Srv.Store.Fanpage().GetFanpagesByUserId(userId)
	if result.Err != nil {
		return nil, result.Err
	}
	return result.Data.([]*model.FanpageIncludeInitResult), nil
}

func (app *App) CreateFanpage(fanpage *model.Fanpage) (*model.Fanpage, *model.AppError) {
	result := <-app.Srv.Store.Fanpage().Save(fanpage)
	if result.Err != nil {
		mlog.Error(fmt.Sprintf("Couldn't save the page err=%v", result.Err))
		return nil, result.Err
	}
	rfanpage := result.Data.(*model.Fanpage)

	return rfanpage, nil

	// Có thể gửi một message đến tất cả thành viên của page bằng websocket
	//message := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_NEW_FANPAGE, "", "", "", nil)
	//message.Add("fanpage_id", rfanpage.Id)
	//a.Publish(message)
}

// user tạo mới fanpage
func (app *App) CreateFanpageWithUser(fanpage *model.Fanpage, userId string) (*model.Fanpage, *model.AppError) {
	user, err := app.GetUser(userId)
	if err != nil {
		return nil, err
	}

	rfanpage, err := app.CreateFanpage(fanpage)
	if err != nil {
		return nil, err
	}

	// thêm user là thành viên của fanpage
	if err = app.AddUserToFanpage(rfanpage, user, ""); err != nil {
		return nil, err
	}

	// emit to socket
	message := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_CHANNEL_CREATED, "", "", userId, nil)
	message.Add("channel_id", "213213")
	message.Add("team_id", "6516489498")
	app.Publish(message)

	return rfanpage, nil
}

// Trả về 3 giá trị
// 1. Con trỏ tới fanpage member nếu thêm thành công
// 2. bool: true nếu user đã được thêm vào fanpage từ trước, fale nếu user chưa được thêm
// 3. Contrỏ tới 1 AppError nếu xảy ra lỗi
func (app *App) addUserToFanpage(fanpage *model.Fanpage, user *model.User) (*model.FanpageMember, bool, *model.AppError) {
	fm := &model.FanpageMember{
		FanpageId: fanpage.Id,
		UserId:    user.Id,
	}

	efmr := <-app.Srv.Store.Fanpage().GetMember(fanpage.Id, user.Id)
	if efmr.Err != nil {
		// membership chưa có, thêm ngay
		fmr := <-app.Srv.Store.Fanpage().SaveFanPageMember(fm)
		if fmr.Err != nil {
			return nil, false, fmr.Err
		}
		return fmr.Data.(*model.FanpageMember), false, nil
	}

	// Membership already exists.  Check if deleted and and update, otherwise do nothing
	rfm := efmr.Data.(*model.FanpageMember)

	return rfm, true, nil
}

func (app *App) AddUserToFanpage(fanpage *model.Fanpage, user *model.User, userRequestorId string) *model.AppError {
	_, alreadyAdded, err := app.addUserToFanpage(fanpage, user)
	if err != nil {
		return err
	}
	if alreadyAdded {
		return nil
	}

	app.ClearSessionCacheForUser(user.Id)
	app.InvalidateCacheForUser(user.Id)

	return nil
}

func (a *App) ViewPage(view *model.PageView, userId string, currentSessionId string) (map[string]int64, *model.AppError) {
	if err := a.SetActivePage(userId, view.PageId); err != nil {
		return nil, err
	}

	pageIds := []string{}

	if len(view.PageId) > 0 {
		pageIds = append(pageIds, view.PageId)
	}

	if len(pageIds) == 0 {
		return map[string]int64{}, nil
	}

	return a.MarkPagesAsViewed(pageIds, userId, currentSessionId)
}

func (a *App) MarkPagesAsViewed(pageIds []string, userId string, currentSessionId string) (map[string]int64, *model.AppError) {
	result := <-a.Srv.Store.Fanpage().UpdateLastViewedAt(pageIds, userId)
	if result.Err != nil {
		return nil, result.Err
	}

	times := result.Data.(map[string]int64)
	if *a.Config().ServiceSettings.EnableChannelViewedMessages {
		for _, pageId := range pageIds {
			message := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_CHANNEL_VIEWED, "", "", userId, nil)
			message.Add("page_id", pageId)
			a.Publish(message)
		}
	}

	return times, nil
}

func (a *App) SetActivePage(userId string, pageId string) *model.AppError {
	status, err := a.GetStatus(userId)

	oldStatus := model.STATUS_OFFLINE

	if err != nil {
		status = &model.Status{
			UserId: userId,
			Status: model.STATUS_ONLINE,
			Manual: false,
			LastActivityAt: model.GetMillis(),
			ActivePage: pageId,
		}
	} else {
		oldStatus = status.Status
		status.ActivePage = pageId
		if !status.Manual && pageId != "" {
			status.Status = model.STATUS_ONLINE
		}
		status.LastActivityAt = model.GetMillis()
	}

	a.AddStatusCache(status)

	if status.Status != oldStatus {
		a.BroadcastStatus(status)
	}

	return nil
}

func (app *App) SanitizeFanpage(session model.Session, fanpage *model.Fanpage) *model.Fanpage {
	// cần kiểm tra xem user có quyền truy cập page không
	// tạm thời chưa dùng đến
	//if !a.SessionHasPermissionToFanpage(session, fanpage.Id, model.PERMISSION_MANAGE_FANPAGE) {
	//	fanpage.Sanitize()
	//}

	return fanpage
}

func (app *App) SanitizeFanpages(session model.Session, fanpages []*model.Fanpage) []*model.Fanpage {
	for _, fanpage := range fanpages {
		app.SanitizeFanpage(session, fanpage)
	}
	return fanpages
}

func (app *App) CreateReplySnippet(snippet *model.ReplySnippet) (*model.ReplySnippet, *model.AppError) {
	r := <-app.Srv.Store.PageReplySnippet().CheckDuplicate(snippet.PageId, snippet.Trigger)
	if r.Err == nil && len(r.Data.([]*model.ReplySnippet)) == 0 {
		// snippet chưa có trong db
		result := <-app.Srv.Store.PageReplySnippet().Save(snippet)
		if result.Err != nil {
			mlog.Error(fmt.Sprintf("Couldn't save the reply snippet err=%v", result.Err))
			return nil, result.Err
		}
		rSnippet := result.Data.(*model.ReplySnippet)

		return rSnippet, nil
	} else {
		return nil, model.NewAppError("CreateReplySnippet", "dupplicate", nil, "Câu trả lời với ký tự tắt này đã được thêm cho page", http.StatusNotAcceptable)
	}
}

func (app *App) UpdateReplySnippet(snippet *model.ReplySnippet) (*model.ReplySnippet, *model.AppError) {
	result := <-app.Srv.Store.PageReplySnippet().Update(snippet)
	if result.Err != nil {
		mlog.Error(fmt.Sprintf("Couldn't update the reply snippet err=%v", result.Err))
		return nil, result.Err
	}
	rSnippet := result.Data.(*model.ReplySnippet)

	return rSnippet, nil
}

func (app *App) GetPageSnippets(pageId string) ([]*model.ReplySnippet, *model.AppError) {
	result := <-app.Srv.Store.PageReplySnippet().GetByPageId(pageId)
	if result.Err != nil {
		return nil, result.Err
	}
	return result.Data.([]*model.ReplySnippet), nil
}

///////////////////////AUTO MESSAGE TASK
func (app *App) CreateAutoMessageTask(task *model.AutoMessageTask) (*model.AutoMessageTask, *model.AppError) {
	// snippet chưa có trong db
	result := <-app.Srv.Store.AutoMessageTask().Save(task)
	if result.Err != nil {
		mlog.Error(fmt.Sprintf("Couldn't save the auto message task err=%v", result.Err))
		return nil, result.Err
	}
	rTask := result.Data.(*model.AutoMessageTask)

	return rTask, nil
}

func (app *App) GetPageAutoMessageTasks(pageId string) ([]*model.AutoMessageTask, *model.AppError) {
	result := <-app.Srv.Store.AutoMessageTask().GetByPageId(pageId)
	if result.Err != nil {
		return nil, result.Err
	}
	return result.Data.([]*model.AutoMessageTask), nil
}

func (app *App) UpdateAutoMessageTask(task *model.AutoMessageTask) (*model.AutoMessageTask, *model.AppError) {
	result := <-app.Srv.Store.AutoMessageTask().Update(task)
	if result.Err != nil {
		mlog.Error(fmt.Sprintf("Couldn't update the auto message task err=%v", result.Err))
		return nil, result.Err
	}
	rTask := result.Data.(*model.AutoMessageTask)

	return rTask, nil
}

// Khởi động một task, trả về thời điểm bắt đầu chạy
func (app *App) StartAutoMessageTask(taskId string) (string, *model.AppError) {

	return "", nil
}

// Pause một task, trả về thời điểm paused
func (app *App) PauseAutoMessageTask(taskId string) (string, *model.AppError) {

	return "", nil
}

// Delete một task, trả về thời điểm deleted
func (app *App) DeleteAutoMessageTask(taskId string) (string, *model.AppError) {

	return "", nil
}

func (app *App) GetFileInfosForPage(pageId string, readFromMaster bool, offset, limit int) ([]*model.FileInfo, *model.AppError) {
	pchan := app.Srv.Store.Fanpage().GetFanpageByPageID(pageId)

	result := <-app.Srv.Store.FileInfo().GetForPage(pageId, readFromMaster, false, offset, limit)
	if result.Err != nil {
		return nil, result.Err
	}
	infos := result.Data.([]*model.FileInfo)

	if len(infos) == 0 {
		// No FileInfos were returned so check if they need to be created for this post
		result := <-pchan
		if result.Err != nil {
			return nil, result.Err
		}
		page := result.Data.(*model.Fanpage)

		if len(page.Filenames) > 0 {
			app.Srv.Store.FileInfo().InvalidateFileInfosForPageCache(pageId)
			// The post has Filenames that need to be replaced with FileInfos
			infos = app.MigrateFilenamesToFileInfos(page)
		}
	}

	return infos, nil
}

func (app *App) UpdatePageStatus(pageId string, status string, userId string) (*model.AppError, string) {
	udpateResult := <-app.Srv.Store.Fanpage().UpdateStatus(pageId, status)

	if udpateResult.Err != nil {
		mlog.Error(fmt.Sprintf("Couldn't update page status err=%v", udpateResult.Err))
		return udpateResult.Err, ""
	}
	statusResult := udpateResult.Data.(string)

	message := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_PAGE_STATUS_UPDATED, "", "", userId, nil)
	message.Add("page_id", pageId)
	message.Add("status", status)
	app.Publish(message)

	return nil, statusResult
}

func (app *App) UpdatePagesStatus(pageIds *model.LoadPagesInput, status string, userId string) (*model.AppError, string) {
	udpateResult := <-app.Srv.Store.Fanpage().UpdatePagesStatus(pageIds, status)

	if udpateResult.Err != nil {
		mlog.Error(fmt.Sprintf("Couldn't update pages status err=%v", udpateResult.Err))
		return udpateResult.Err, ""
	}
	statusResult := udpateResult.Data.(string)

	// emit to socket
	message := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_PAGES_STATUS_UPDATED, "", "", userId, nil)
	message.Add("page_ids", pageIds.PageIds)
	message.Add("status", status)
	app.Publish(message)

	return nil, statusResult
}

func (app *App) ValidationPagesBeforeInit(pageIds *model.LoadPagesInput, userId string) *model.PagesInitValidationResult {
	validationResult := <-app.Srv.Store.Fanpage().ValidatePagesBeforeInit(pageIds)
	// emit socket thong bao cho user biet ket qua validation
	message := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_PAGES_INI_VALIDATION_RESULT, "", "", userId, nil)
	message.Add("result", validationResult)
	app.Publish(message)

	return validationResult.Data.(*model.PagesInitValidationResult)
}