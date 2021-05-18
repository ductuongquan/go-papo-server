package api1

import (
	"bitbucket.org/enesyteam/papo-server/model"
	"net/http"
	"strconv"
)

func (api *API) InitOrders() {
	api.BaseRoutes.Orders.Handle("", api.ApiSessionRequired(getOrders)).Methods("GET")
	api.BaseRoutes.Orders.Handle("", api.ApiSessionRequired(createOrder)).Methods("POST")
	api.BaseRoutes.Order.Handle("", api.ApiSessionRequired(getOrder)).Methods("GET")
}

func getOrders(c *Context, w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	limit, _ := strconv.Atoi(query.Get("limit"))
	offset, _ := strconv.Atoi(query.Get("offset"))

	err, orders := c.App.GetOrders(limit, offset)
	if err != nil {
		c.Err = err
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(model.OrderListToJson(orders)))
}

func getOrder(c *Context, w http.ResponseWriter, r *http.Request) {


}

func createOrder(c *Context, w http.ResponseWriter, r *http.Request) {
	order := model.OrderFromJson(r.Body)
	if order == nil {
		c.SetInvalidParam("order")
		return
	}

	if err, addedOrder := c.App.CreateOrder(order); err != nil {
		c.Err = err
		return
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(addedOrder.ToJson()))
	}
}