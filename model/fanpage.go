// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package model

import (
	"encoding/json"
	"io"
	"net/http"
)

const(
	FANPAGE_FILENAMES_MAX_RUNES    		= 4000
	PAGE_CACHE_SIZE             		= 25000
	PAGE_STATUS_READY 					= "ready"
	PAGE_STATUS_QUEUED 					= "queued"
	PAGE_STATUS_INITIALIZING			= "initializing"
	PAGE_STATUS_INITIALIZED 			= "initialized"
	PAGE_STATUS_ERROR 					= "error"
	PAGE_STATUS_BLOCKED 				= "blocked"
)
// Model Fanpage sẽ chỉ gồm các trường sau đây, một số trường trong Server cũ không phù hợp đã được loại bỏ
// Trường users trong server cũ là không cần thiết, vì hiếm có trường hợp nào cần query tất cả users của một
// fanpage, thường chỉ query tất cả các page của một user thôi
// Trường access_token chắc chắn phải bỏ đi, vì access_token không cố định cho 1 page và thay đổi theo từng user
// nên không thể lưu vào đây được. Access token của page sẽ được grant và xem xét sẽ lưu thành session
type Fanpage struct {
	Id       string        `json:"id"`
	PageId   string        `json:"page_id"`
	Name     string        `json:"name"`
	Category string        `json:"category"`
	Status   string        `json:"status"`
	CreateAt int64         `json:"create_at"`
	UpdateAt int64         `json:"update_at"`
	DeleteAt int64         `json:"delete_at"`
	BlockAt  int64         `json:"block_at"`
	Visible  bool          `json:"visible"`
	Filenames     StringArray     `json:"filenames,omitempty"` // Deprecated, do not use this field any more
	FileIds       StringArray     `json:"file_ids,omitempty"`// Ví dụ nếu 1 tài khoản chưa thanh toán có thể hệ thống sẽ cần phải khóa page lại
	//Member   FanpageMember `json:"member,omitempty"` // Hiển thị thông tin của member khi join 2 bảng với nhau, chủ yếu để hiển thị access token của member đó
}

type FanpageIncludeInitResult struct {
	Data 			map[string]interface{} 				`json:"data"`
	Init      		map[string]interface{}  			`json:"init"`
}

type LoadPagesInput struct {
	PageIds []string `json:"pageIds"`
}

// Mảng chứa các trạng thái của một Fanpage
var statuses = []string{
	"ready",        // Trạng thái sẵn sàng khởi tạo
	"waiting",      // đã được queued vào hàng và đang chờ để khởi tạo
	"initializing", // đang khởi tạo
	"initialized",  // khởi tạo thành công
	"error",        // khởi tạo lỗi
	"hidden",       // Chủ fanpage có thể ẩn page để các thành viên không thể nhìn thấy dữ liệu của page này
	"deleted",      // page cũng có thể bị xóa hoàn toàn
}

func (p *Fanpage) IsValid() *AppError {

	if len(p.Id) != 26 {
		return NewAppError("FanpageMember.IsValid", "model.fanpage_member.is_valid.fanpage_id.app_error", nil, "", http.StatusBadRequest)
	}

	return nil
}

// Hàm này phải được thực thi trước khi lưu dữ liệu vào database
// các trường CreateAt và UpdateAt sẽ được tự động thêm vào
func (p *Fanpage) PreSave() {
	if p.Id == "" {
		p.Id = NewId()
	}
	p.Status = "ready"
	p.Visible = true
	p.CreateAt = GetMillis()
	p.UpdateAt = p.CreateAt
	p.DeleteAt = 0
}

// Hàm này phải được thực thi trước khi thực hiện update page
// xem xét cần làm gì trước khi update Fanpage
func (p *Fanpage) PreUpdate() {
	// Tạm thời chưa nghĩ ra phải làm gì trước khi update Fanpage
	//
}

// Convert một Fanpage sang chuỗi json
func (p *Fanpage) ToJson() string {
	b, _ := json.Marshal(p)
	return string(b)
}

// decode input và trả về đối tượng Fanpage
func FanpageFromJson(data io.Reader) *Fanpage {
	var p *Fanpage
	json.NewDecoder(data).Decode(&p)
	return p
}

// Chuyển đổi một danh sách Fanpages sang chuỗi json
func FanpageListToJson(p []*Fanpage) string {
	b, _ := json.Marshal(p)
	return string(b)
}

func FanpagesIncludeInitResultToJson(p []*FanpageIncludeInitResult) string {
	b, _ := json.Marshal(p)
	return string(b)
}

// decode input và trả về danh sách các đối tượng Fanpages
func FanpageListFromJson(data io.Reader) []*Fanpage {
	var fanpages []*Fanpage
	json.NewDecoder(data).Decode(&fanpages)
	return fanpages
}

type PageStatus struct {
	Status string `json:"status"`
}

func PageStatusFromJson(data io.Reader) *PageStatus {
	var p *PageStatus
	json.NewDecoder(data).Decode(&p)
	return p
}

type PagesStatus struct {
	Status string `json:"status"`
	PageIds []string 	`json:"page_ids"`
}

type PageInitValidationError struct {
	Id 			string 		`json:"id"`
	ErrorCode 	int 		`json:"error_code"`
	Message 	string 		`json:"message"`
}

type PagesInitValidationResult struct {
	Success 	[]string 						`json:"success"`
	Error 		[]PageInitValidationError 		`json:"error"`
}

func PagesInitValidationResultToJson(p *PagesInitValidationResult) string {
	b, _ := json.Marshal(p)
	return string(b)
}

func PagesInitValidationToLPi(r *PagesInitValidationResult) *LoadPagesInput {
	lpi := LoadPagesInput{}
	for _, pageId := range r.Success {
		lpi.PageIds = append(lpi.PageIds, pageId)
	}
	return &lpi
}

func PagesStatusFromJson(data io.Reader) *PagesStatus {
	var p *PagesStatus
	json.NewDecoder(data).Decode(&p)
	return p
}