// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package facebookgraph

import (
	"encoding/json"
	"io"
	"io/ioutil"
)

// Cần nghiên cứu thêm xem facebook API có kiểu lỗi trả về nào khác kiểu dưới đây không
type FacebookError struct {
	Error Error `json:"error"`
}

type Error struct {
	Message   			string 		`json:"message"`
	Type      			string 		`json:"type"`
	Code      			int    		`json:"code"`
	ErrorSubcode 		int 		`json:"error_subcode"`
	FbtraceId 			string 		`json:"fbtrace_id"`
}

func (p *FacebookError) ToJson() string {
	b, _ := json.Marshal(p)
	return string(b)
}

func (p *Error) ToJson() string {
	b, _ := json.Marshal(p)
	return string(b)
}

func FacebookErrorFromJson(data io.Reader) *FacebookError {
	var fbe *FacebookError
	x, _ := ioutil.ReadAll(data)
	json.Unmarshal([]byte(x), &fbe)
	return fbe
}

func FacebookErrorToJson(t *FacebookError) string {
	b, _ := json.Marshal(t)
	return string(b)
}

//func (p *FacebookError) ToAppError() *model.AppError {
//	return nil // model.NewAppError("Facebook error code: "+strconv.Itoa(p.Error.Code), p.Error.Message, nil, "", http.StatusBadRequest)
//}

type getPageScopeIdResponseDataItemPage struct {
	Id                  string           `json:"id"`
	Name                string           `json:"name"`
}

type getPageScopeIdResponseDataItem struct {
	Id 					string 					`json:"id"`
	Page 				*getPageScopeIdResponseDataItemPage 					`json:"page"`
}

type GetPageScopeIdResponse struct {
	Data 				[]*getPageScopeIdResponseDataItem 		`json:"data"`
}

func GetPageScopeIdResponseFromJson(data io.Reader) *GetPageScopeIdResponse {
	var content *GetPageScopeIdResponse
	json.NewDecoder(data).Decode(&content)
	return content
}
