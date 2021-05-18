package sqlstore

import (
	"bitbucket.org/enesyteam/papo-server/model"
	"bitbucket.org/enesyteam/papo-server/store"
	"fmt"
	"net/http"
)

type sqlOrderStore struct {
	SqlStore
}

func NewSqlOrderStore(sqlStore SqlStore) store.OrderStore {

	fs := &sqlOrderStore{
		SqlStore: sqlStore,
	}

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(model.Order{}, "Orders").SetKeys(false, "Id")
		table.ColMap("Id").SetMaxSize(26)
	}

	return fs
}

func (fs sqlOrderStore) CreateIndexesIfNotExists() {
	//fs.CreateIndexIfNotExists("idx_fanpages_page_id", "Fanpages", "PageId")
}

func (fs sqlOrderStore) Save(order *model.Order) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		order.PreSave()

		if err := fs.GetMaster().Insert(order); err != nil {
			fmt.Print(err)
			result.Err = model.NewAppError("sqlOrderStore.Save", "store.sql_order.save.app_error", nil, "", http.StatusInternalServerError)
		} else {
			result.Data = order
		}
	})
}

func (fs sqlOrderStore) GetOrders(limit, offset int) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var orders []*model.Order
		query := `SELECT Orders.* FROM Orders LIMIT :Limit OFFSET :Offset`
		if _, err := fs.GetReplica().Select(&orders, query, map[string]interface{}{"Limit": limit, "Offset": offset}); err != nil {
			result.Err = model.NewAppError("sqlOrderStore.GetOrders", "store.sqlOrderStore.GetOrders.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		}
		result.Data = orders
	})
}