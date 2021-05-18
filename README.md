# [![Papo Server](https://github.com/enesyteam/papo-server/blob/master/uploads/papo-server-logo.png?raw=true)](https://papoweb.org)

Papo Server được viết bởi Golang, database sử dụng Postgresql.

## Thiết lập `GOPATH`

Nếu sử dụng Windows, sau khi cài đặt Go, cần phải thiết lập biến môi trường `GOPATH`, mặc định là `C:\Users\ducnv\go`.

`GOPATH` rất quan trọng, sau này các câu lệnh `go get ....` (Tương tự như `npm install ...`) sẽ cài đặt các thư viện vào đây.

(Cũng có thể thiết lập `GOPATH` bằng câu lệnh `set GOPATH=C:\...` ngay khi bạn đang ở thư mục project làm việc.)

### Tải source code Papo Server

Mở `terminal`, gõ:
```
go get bitbucket.org/enesyteam/papo-server
```
Source code sẽ được clone về: `GOPATH\src\bitbucket.org\enesyteam\papo-server`

### Cài đặt dependencies

cd đến thư mục này, cài các dependencies:

```
dep ensure
```

Việc cài bằng `dep ensure` mất khá nhiều thời gian, hãy kiên nhẫn chờ đợi

### Run

Chạy ứng dụng bằng lệnh

```
go run cmd/main.go
```

Enjoy!

**TỪ ĐOẠN NÀY TRỞ ĐI LÀ NHỮNG ĐIỀU HAY HO MÀ CHÚNG TA HỌC ĐƯỢC VÀ MUỐN GHI CHÉP RA ĐÂY ĐỂ CHIA SẺ VỚI NHAU**

* [Ghi chép của Công](tmp/ducnv.md)
* [Ghi chép của Cường](tmp/_cuong.md)

https://graph.facebook.com/{graph-api-version}/oauth/access_token?  
    grant_type=fb_exchange_token&          
    client_id=1461602217298862&
    client_secret=06896b52987d1eb5d81a7fc31fa22464&
    fb_exchange_token=EAAUxUcj3C64BAGo3GltL9imhzGqy2JVHYwjVe8fhChQFWC29pZBFmF8l1ZBftD70XIqFamMu9UqqZBlx3sfbRR6cqPBZC2uwDiX0yEgNlT5Vte8Nb603nibm8TdiZCPeqZBZBqFr0sKT3Pvj0SSnonTcwJS96QNkmV1xZBBZA6MYUDnjmTQzfrLyZB15ETyJrsEnNp2ydgdkCDuvazoIqvK4iZA


Webhook: https://6f1f77568371.ngrok.io/webhooks/facebook
references: https://github.com/Schmavery/facebook-chat-api/blob/master/src/listen.js

