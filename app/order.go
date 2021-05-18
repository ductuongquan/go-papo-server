package app

import (
	"bitbucket.org/enesyteam/papo-server/model"
	"fmt"
)

func (app *App) CreateOrder(order *model.Order) (*model.AppError, *model.Order) {

	fmt.Println("order", order)

	result := <-app.Srv.Store.Order().Save(order)
	if result.Err != nil {
		return result.Err, nil
	}

	addedOrder := result.Data.(*model.Order)

	// socket ve client
	webhookData := model.NewWebSocketEvent(model.ADDED_ORDER, "", "461787374317076", "", nil)
	webhookData.Add("orders", addedOrder)
	webhookData.Add("test", "test1")
	app.Publish(webhookData)

	//return model.NewAppError("CreateOrder", "order.create_new_order.app_error", nil, "", http.StatusBadRequest), nil

	return nil, addedOrder
}

func (app *App) GetOrders(limit, offset int) (*model.AppError, []*model.Order) {
	if limit == 0 {
		limit = 25
	}
	result := <-app.Srv.Store.Order().GetOrders(limit, offset)
	if result.Err != nil {
		return result.Err, nil
	}
	return nil, result.Data.([]*model.Order)
}
