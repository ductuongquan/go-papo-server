package app

import "bitbucket.org/enesyteam/papo-server/model"

func (a *App) GetFacebookUsersById(id string) (*model.FacebookUid, *model.AppError) {
	result := <-a.Srv.Store.FacebookUid().Get(id)
	if result.Err != nil {
		return nil, result.Err
	}
	return result.Data.(*model.FacebookUid), nil
}

func (a *App) GetFacebookUsersByIds(userIds []string, asAdmin bool) ([]*model.FacebookUid, *model.AppError) {
	result := <-a.Srv.Store.FacebookUid().GetByIds(userIds, true)
	if result.Err != nil {
		return nil, result.Err
	}
	return result.Data.([]*model.FacebookUid), nil
}

func (a *App) UpdateFacebookUserPageId(id, pageId string) (bool, *model.AppError) {
	result := <-a.Srv.Store.FacebookUid().UpdatePageId(id, pageId)
	if result.Err != nil {
		return false, result.Err
	}
	return true, nil
}

func (a *App) UpdateFacebookUserPageScopeId(id, pageScopeId string) (bool, *model.AppError) {
	result := <-a.Srv.Store.FacebookUid().UpdatePageScopeId(id, pageScopeId)
	if result.Err != nil {
		return false, result.Err
	}
	return true, nil
}
