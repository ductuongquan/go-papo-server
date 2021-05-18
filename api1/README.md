# Chỉ dẫn cho Developer

## Cấu trúc thư mục trong Papo server

Cấu trúc thư mục trong server rất rõ ràng và dễ hiểu, đại khái là:
- *`./api1/`* chứa tất cả API và code liên quan đến App
- *`./model`* chứa tất cả định nghĩa dữ liệu của model và các hàm điều khiển
- *`./store`* chứa tất cả code query database
- *`./utils`* các hàm phụ trợ
- *`./i18n`* các file localization, CHÚ Ý: Tất cả chuỗi string dùng trong server này phải được khai báo ở đây
- *`./mlog`* package dùng để log dữ liệu, CHÚ Ý: Tất cả những gì log ra phải dùng package này
- *`./config`* chứa tất cả thiết lập của server, CHÚ Ý: Tất cả thiết lập phải được khai báo ở đây
- *`./cmd`* chứa tất cả commands để chạy server
- *`./app`* package app

## Tạo Model

Xem file: `./model/fanpage.go` trong đó đã được giải thích rất chi tiết

## Khai báo Store (tương tác với dữ liệu của database)

Tại file `./store/store.go` bổ sung vào các interface như sau:

`...type Store interface {` bổ sung:
    ```
    Fanpage() FanpageStore
    ```
Đồng thời khai báo interface FanpageStore vào ngay bên dưới của file này:

`type FanpageStore interface {....` << Xem trong file `./store/store.go`

Tiếp theo khai báo các dòng này vào file `./store/layered_store.go`:

```
func (s *LayeredStore) Fanpage() FanpageStore {
	return s.DatabaseLayer.Fanpage()
}
```

Tiếp theo thêm vào file: `./store/sqlstore/superlier.go` như sau:

tại struct thêm dòng:
```fanpage 			 store.FanpageStore```

tại `NewSqlSupplier` thêm dòng:
```supplier.oldStores.fanpage = NewSqlFanpageStore(supplier)```

và dòng:
```supplier.oldStores.fanpage.(*sqlFanpageStore).CreateIndexesIfNotExists()```

và khai báo thêm phương thức Fanpage()
```
func (ss *SqlSupplier) Fanpage() store.FanpageStore {
	return ss.oldStores.fanpage
}
```

Các bước khai báo đã được giải thích rất rõ trong file `./store/sqlstore/fanpage_store.go`

## Viết hàm điều khiển tại App

Đây là các hàm điều khiển việc tương tác với dữ liệu database ở cấp độ App

Xem file `./app/fanpage.go`

## Viết hàm API

Mở file `./api1/api.go` và thêm dòng:

```
    Fanpages     	*mux.Router // 'api/v1/fanpages'
	Fanpage     	*mux.Router // 'api/v1/fanpages/{fanpage_id:[A-Za-z0-9]+}'
```

vào `type Routes struct {`. Ta dự định sẽ quản lý fanpage bằng 2 nhóm routes, nhóm thứ nhất xử lý
danh sách fanpage và nhóm thứ 2 xử lý riêng từng fanpage, với params là pageId

Thêm các dòng này vào Hàm Init()

    ```
    // fanpages
    api.BaseRoutes.Fanpages = api.BaseRoutes.ApiRoot.PathPrefix("/fanpages").Subrouter()
    api.BaseRoutes.Fanpage = api.BaseRoutes.ApiRoot.PathPrefix("/{fanpage_id:[A-Za-z0-9]+}").Subrouter()
    ```

Và: gọi hàm Init() của Api fanpage mới tạo này, hàm Init này sẽ được viết trong file `./api1/fanpage.go`

```
api.InitFanpage()
```
## Xử lý APIs cho fanpage

Tạo file `./api1/fanpage.go`

Các chỉ dẫn xem trong file này nhé!
======================================================
OK, khởi động lại server và mở Navicat để xem, ta thấy một table mới là `fanpages` đã được khởi tạo

bây giờ test Postman xem thế nào:

API Post 'http://localhost:8065/api/v1/fanpages'

Ban đầu thử không truyền Body thì sẽ trả về lỗi:
```
{
    "id": "api.context.invalid_body_param.app_error",
    "message": "Invalid or missing fanpage in request body",
    "detailed_error": "",
    "request_id": "cu7wkzxz63ddmb1tcgghetxufa",
    "status_code": 400
}
```
Thử truyền vào body này:
Body:
```
{
	"id": "123456789",
	"name": "Page của tôi",
	"category": "Artist",
	"status": "ready"
}
```

Kết quả:
```
{
    "id": "123456789",
    "name": "Page của tôi",
    "category": "Artist",
    "status": "ready",
    "create_at": 1541619395164,
    "update_at": 1541619395164,
    "delete_at": 0,
    "block_at": 0,
    "visible": false
}
```

Thử post tiếp request này:
Lỗi:
```
{
    "id": "store.sql_fanpage.save.app_error",
    "message": "Đã có lỗi xảy ra, không thể lưu dữ liệu page.",
    "detailed_error": "",
    "request_id": "czo18fyd5td9tk5uhgkmrwe6iy",
    "status_code": 500
}
```

## API `GET /fanpages/id`

API GET đòi hỏi phải khai báo params trong file `./web/params.go`. Ta sẽ thêm params là PageId vào như sau:
Trong file này thêm vào `PageId         string` tại struct




================================================
Enjoy!







