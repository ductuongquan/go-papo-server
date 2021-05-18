package model

import (
	"encoding/json"
	"io"
)

type Order struct {
	Id 					string 			`json:"id"`
	CustomerName 		string 			`json:"customer_name"`
	CreateAt 			int64 			`json:"create_at"`
}

func OrderFromJson(data io.Reader) *Order {
	var p *Order
	json.NewDecoder(data).Decode(&p)
	return p
}

func OrderListToJson(p []*Order) string {
	b, _ := json.Marshal(p)
	return string(b)
}

func (o *Order) PreSave() {
	if o.Id == "" {
		o.Id = NewId()
	}

	o.CreateAt = GetMillis()
}

func (p *Order) ToJson() string {
	b, _ := json.Marshal(p)
	return string(b)
}

//type Order struct {
//	Id 					string 			`json:"id"`
//	PageId 				string 			`json:"page_id"`
//	ExternalSource 		string 			`json:"external_source"`
//	ExternalLink 		string			`json:"external_link"`
//	TeamId 				string 			`json:"team_id"`
//	CreateAt 			int64         	`json:"create_at"`
//	UpdateAt 			int64         	`json:"update_at"`
//	DeleteAt 			int64         	`json:"delete_at"`
//	BlockAt  			int64         	`json:"block_at"`
//	AssignTo			string 			`json:"assign_to"`
//	AssignBy 			string 			`json:"assign_by"`
//	LastAssignAt 		int64 			`json:"last_assign_at"`
//	ProductId 			string 			`json:"product_id"`
//	CustomerName		string 			`json:"customer_name"`
//	CustomerMobile 		string 			`json:"customer_mobile"`
//	CustomerMobileExt  	string 			`json:"customer_mobile_ext"`
//	CustomerAddress 	string 			`json:"customer_address"`
//	CustomerAddressExt 	string 			`json:"customer_address_ext"`
//	CustomerBirthday 	int64 			`json:"customer_birthday"`
//	CustomerLunarBirthday 	int64 		`json:"customer_lunar_birthday"`
//	CustomerNotes 		string 			`json:"customer_notes"`
//	CreatorNotes 		string 			`json:"creator_notes"`
//	StatusId 			string 			`json:"status_id"`
//}
