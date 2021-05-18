package app

import (
	"bitbucket.org/enesyteam/papo-server/mlog"
	"bitbucket.org/enesyteam/papo-server/model"
	"fmt"
)

func (app *App) CreateFanpageInitResult(fr *model.FanpageInitResult) (*model.FanpageInitResult, *model.AppError) {
	result := <-app.Srv.Store.FanpageInitResult().Save(fr)
	if result.Err != nil {
		mlog.Error(fmt.Sprintf("Couldn't save the page init result err=%v", result.Err))
		return nil, result.Err
	}
	r := result.Data.(*model.FanpageInitResult)

	return r, nil
}

func (app *App) UpdateFanpageInitResultConversationCount(fr *model.FanpageInitResult, newCount int64) (*model.FanpageInitResult, *model.AppError) {
	result := <-app.Srv.Store.FanpageInitResult().UpdateConversationCount(fr, newCount)
	if result.Err != nil {
		mlog.Error(fmt.Sprintf("Couldn't save the page init result err=%v", result.Err))
		return nil, result.Err
	}
	r := result.Data.(*model.FanpageInitResult)

	return r, nil
}

func (app *App) UpdateFanpageInitResultPostCount(fr *model.FanpageInitResult, newCount int64) (*model.FanpageInitResult, *model.AppError) {
	result := <-app.Srv.Store.FanpageInitResult().UpdatePostCount(fr, newCount)
	if result.Err != nil {
		mlog.Error(fmt.Sprintf("Couldn't save the page init result err=%v", result.Err))
		return nil, result.Err
	}
	r := result.Data.(*model.FanpageInitResult)

	return r, nil
}

func (app *App) UpdateFanpageInitResultCommentCount(fr *model.FanpageInitResult, newCount int64) (*model.FanpageInitResult, *model.AppError) {
	result := <-app.Srv.Store.FanpageInitResult().UpdateCommentCount(fr, newCount)
	if result.Err != nil {
		mlog.Error(fmt.Sprintf("Couldn't save the page init result err=%v", result.Err))
		return nil, result.Err
	}
	r := result.Data.(*model.FanpageInitResult)

	return r, nil
}

func (app *App) UpdateFanpageInitResultMessageCount(fr *model.FanpageInitResult, newCount int64) (*model.FanpageInitResult, *model.AppError) {
	result := <-app.Srv.Store.FanpageInitResult().UpdateMessageCount(fr, newCount)
	if result.Err != nil {
		mlog.Error(fmt.Sprintf("Couldn't save the page init result err=%v", result.Err))
		return nil, result.Err
	}
	r := result.Data.(*model.FanpageInitResult)

	return r, nil
}

func (app *App) UpdateFanpageInitResultEndAt(fr *model.FanpageInitResult, endAt int64) (*model.FanpageInitResult, *model.AppError) {
	result := <-app.Srv.Store.FanpageInitResult().UpdateEndAt(fr, endAt)
	if result.Err != nil {
		mlog.Error(fmt.Sprintf("Couldn't save the page init result err=%v", result.Err))
		return nil, result.Err
	}
	r := result.Data.(*model.FanpageInitResult)

	return r, nil
}