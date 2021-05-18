package app

import (
	"bitbucket.org/enesyteam/papo-server/model"
)

func (app *App) CreateTermsOfService(text, userId string) (*model.TermsOfService, *model.AppError) {
	termsOfService := &model.TermsOfService{
		Text:   text,
		UserId: userId,
	}

	if _, err := app.GetUser(userId); err != nil {
		return nil, err
	}

	result := <-app.Srv.Store.TermsOfService().Save(termsOfService)
	if result.Err != nil {
		return nil, result.Err
	}

	termsOfService = result.Data.(*model.TermsOfService)
	return termsOfService, nil
}

func (app *App) GetLatestTermsOfService() (*model.TermsOfService, *model.AppError) {
	if result := <-app.Srv.Store.TermsOfService().GetLatest(true); result.Err != nil {
		return nil, result.Err
	} else {
		termsOfService := result.Data.(*model.TermsOfService)
		return termsOfService, nil
	}
}

func (app *App) GetTermsOfService(id string) (*model.TermsOfService, *model.AppError) {
	if result := <-app.Srv.Store.TermsOfService().Get(id, true); result.Err != nil {
		return nil, result.Err
	} else {
		termsOfService := result.Data.(*model.TermsOfService)
		return termsOfService, nil
	}
}