// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

// Tất cả giao tiếp với facebook graph phải được định nghĩa ở đây

package app

import (
	"bitbucket.org/enesyteam/papo-server/facebook_graph"
	"bitbucket.org/enesyteam/papo-server/mlog"
	"bitbucket.org/enesyteam/papo-server/model"
	"bitbucket.org/enesyteam/papo-server/utils"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	FACEBOOK_API_ROOT  = "https://graph.facebook.com/v3.2"
	HEADER_REQUEST_ID  = "X-Request-ID"
	HEADER_VERSION_ID  = "X-Version-ID"
	HEADER_ETAG_SERVER = "ETag"
	STATUS_OK          = "OK"
)

func closeBody(r *http.Response) {
	if r.Body != nil {
		ioutil.ReadAll(r.Body)
		r.Body.Close()
	}
}

type Response struct {
	StatusCode    int
	Error         *model.AppError
	RequestId     string
	Etag          string
	ServerVersion string
	Header        http.Header
}

func BuildErrorResponse(r *http.Response, err *model.AppError) *Response {
	var statusCode int
	var header http.Header
	if r != nil {
		statusCode = r.StatusCode
		header = r.Header
	} else {
		statusCode = 0
		header = make(http.Header)
	}

	return &Response{
		StatusCode: statusCode,
		Error:      err,
		Header:     header,
	}
}

func BuildResponse(r *http.Response) *Response {
	return &Response{
		StatusCode:    r.StatusCode,
		RequestId:     r.Header.Get(HEADER_REQUEST_ID),
		Etag:          r.Header.Get(HEADER_ETAG_SERVER),
		ServerVersion: r.Header.Get(HEADER_VERSION_ID),
		Header:        r.Header,
	}
}

// Kết quả trả về của Init Page
//type InitPageResponse struct {
//	StatusCode    int
//	Error         string
//}

func (app *App) doFacebookRequest(accessToken string, path string, method string, usingFullPath bool) (io.ReadCloser, *facebookgraph.FacebookError, *model.AppError) {
	if len(path) == 0 {
		return nil, nil, model.NewAppError("doFacebookRequest", "missing", nil, "", http.StatusBadRequest)
	}
	if len(method) == 0 {
		method = "GET"
	}

	reqUrl := FACEBOOK_API_ROOT + path
	if usingFullPath {
		reqUrl = path
	}
	req, _ := http.NewRequest(method, reqUrl, strings.NewReader(""))

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	if resp, err := app.HTTPService.MakeClient(true).Do(req); err != nil {
		// Cần xử lý thêm ở đoạn này
		return nil, nil, model.NewAppError("doFacebookRequest", "missing", nil, "", http.StatusBadRequest)
	} else {
		var bodyBytes []byte
		bodyBytes, _ = ioutil.ReadAll(resp.Body)
		resp.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

		if resp.StatusCode != http.StatusOK {
			return nil, facebookgraph.FacebookErrorFromJson(resp.Body), nil
		}

		return resp.Body, nil, nil
	}
}

// request với path có dạng "/me/accounts"
func (app *App) request(accessToken string, path string, method string) (io.ReadCloser, *facebookgraph.FacebookError, *model.AppError) {
	return app.doFacebookRequest(accessToken, path, method, false)
}

// request với path có dạng "https://graph.facebook.com/v3.2/t_170206597264241/messages?fields=created_.....
// thường là một URL trong paging của kết quả trả về, mà ta cần graph đầy đủ kết quả
func (app *App) requestUsingFullpath(accessToken string, path string, method string) (io.ReadCloser, *facebookgraph.FacebookError, *model.AppError) {
	return app.doFacebookRequest(accessToken, path, method, true)
}

// Trả về 2 tham số:
// 1. token mới sau khi extend
// 2. FacebookError nếu có lỗi
func (app *App) ExtendFacebookToken(token string) (*facebookgraph.FacebookToken, *facebookgraph.FacebookError, *model.AppError) {
	body, err, aerr := app.request(token, "/oauth/access_token?client_id=" + *app.Config().FacebookSettings.Id+"&client_secret=" + *app.Config().FacebookSettings.Secret+"&grant_type=fb_exchange_token&fb_exchange_token="+token, "GET")
	if err != nil || aerr != nil {
		return nil, err, aerr
	}
	var fbtk *facebookgraph.FacebookToken
	fbtk = facebookgraph.FacebookTokenFromJson(body)
	return fbtk, nil, nil
}

// TODO: Sửa lại không trả về access token
// Trả về danh sách các fanpages của user
func (app *App) GraphFanpages(token string) ([]facebookgraph.FacebookPage, *facebookgraph.FacebookError, *model.AppError) {
	var nextPage = ""
	var result []facebookgraph.FacebookPage
	body, err, aerr := app.request(token, "/me/accounts?fields=username,access_token,id,name,category,category_list,tasks,phone,is_published=true", "GET")
	if err != nil {
		return nil, err, nil
	}
	if aerr != nil {
		return nil, nil, aerr
	}

	pages := facebookgraph.FacebookPagesFromJson(body)
	for _, p := range pages.Data {
		result = append(result, p)
	}

	nextPage = pages.Paging.Next
	for len(nextPage) > 0 {
		body1, err1, aErr1 := app.requestUsingFullpath(token, nextPage, "GET")
		if err1 != nil {
			return nil, err1, nil
		}
		if aErr1 != nil {
			return nil, nil, aErr1
		}
		pages1 := facebookgraph.FacebookPagesFromJson(body1)

		for _, p1 := range pages1.Data {
			result = append(result, p1)
		}
		nextPage = pages1.Paging.Next
	}
	return result, nil, nil
}

func (app *App) GraphPageAndInit(pageId string, user *model.User, ) (string, *facebookgraph.FacebookError, *model.AppError) {
	body, err, aerr := app.request(user.FacebookToken, "/"+pageId+"?fields=name,category,access_token", "GET")
	if err != nil {
		return "", err, nil
	}

	if aerr != nil {
		// cần log lỗi từ facebook
		//mlog.Error(aerr.Error())
		return "", nil, aerr
	}

	fbPage := facebookgraph.FacebookPageFromJson(body)
	fmt.Println("xxxxxxxxxxxxxxxxxxx", fbPage)
	var page *model.Fanpage

	// Kiểm tra page đã tồn tại hay chưa, tìm theo PageId
	if pageResult := <-app.Srv.Store.Fanpage().GetFanpageByPageID(fbPage.Id); pageResult.Err != nil {
		// Page chưa có trong DB => Save page
		page = &model.Fanpage{PageId: fbPage.Id, Name: fbPage.Name, Category: fbPage.Category}
		if pageResult := <-app.Srv.Store.Fanpage().Save(page); pageResult.Err == nil {
			page = pageResult.Data.(*model.Fanpage)
		} else {
			mlog.Error(pageResult.Err.Error())
			return "", nil, pageResult.Err
		}
	} else {
		// page đã có trong DB
		page = pageResult.Data.(*model.Fanpage)
	}

	if page != nil {
		memberResult := <-app.Srv.Store.Fanpage().GetMember(page.Id, user.Id)
		if memberResult.Err != nil {

			pageMember := model.FanpageMember{FanpageId: page.Id, PageId: page.PageId, UserId: user.Id, AccessToken: fbPage.AccessToken}
			if memberResult := <-app.Srv.Store.Fanpage().SaveFanPageMember(&pageMember); memberResult.Err != nil {
				return "", nil, memberResult.Err
			} else {
				return memberResult.Data.(*model.FanpageMember).AccessToken, nil, nil
			}
		} else {
			return memberResult.Data.(*model.FanpageMember).AccessToken, nil, nil
		}
	}
	return "", nil, nil
}

// tham số đầu tiên trong kết quả trả về cho biết page đã tồn tại hay chưa
func (app *App) UpsertPageFromFacebookPage(fbPage facebookgraph.FacebookPage) (bool, *model.Fanpage, *model.AppError) {
	var newPage *model.Fanpage
	isExist := false

	if pageResult := <-app.Srv.Store.Fanpage().GetFanpageByPageID(fbPage.Id); pageResult.Err != nil {
		// Page chưa có trong DB => Save page
		// khởi tạo facebookuids để tránh việc phải khởi tạo lại chính user này khi khởi tạo page
		fUser := facebookgraph.FacebookUser{
			Id: fbPage.Id,
			Name: fbPage.Name,
			PageId: fbPage.Id,
		}

		//app.Srv.Store.FacebookUid().UpsertFromFbUser(fUser)

		if r := <-app.Srv.Store.FacebookUid().UpsertFromFbUser(fUser); r.Err != nil {
			// Lỗi thêm page user vào hệ thống, xử lý sau
			//return nil, r.Err
		}

		newPage = &model.Fanpage{PageId: fbPage.Id, Name: fbPage.Name, Category: fbPage.Category}
		if pageResult := <-app.Srv.Store.Fanpage().Save(newPage); pageResult.Err == nil {
			newPage = pageResult.Data.(*model.Fanpage)
		} else {
			mlog.Error(pageResult.Err.Error())
			return false, nil, pageResult.Err
		}
	} else {
		// page đã có trong DB
		isExist = true
		newPage = pageResult.Data.(*model.Fanpage)
	}

	return isExist, newPage, nil
}

func (app *App) GraphPageAndInit2(pageId string, user *model.User, pages []facebookgraph.FacebookPage) (string, bool, *facebookgraph.FacebookError, *model.AppError) {
	var fbPage facebookgraph.FacebookPage
	for _, page := range pages {
		if page.Id == pageId {
			fbPage = page
			break
		}
	}

	if &fbPage != nil {
		var newPage *model.Fanpage
		// Kiểm tra page đã tồn tại hay chưa, tìm theo PageId
		if pageResult := <-app.Srv.Store.Fanpage().GetFanpageByPageID(fbPage.Id); pageResult.Err != nil {
			// Page chưa có trong DB => Save page
			// khởi tạo facebookuids để tránh việc phải khởi tạo lại chính user này khi khởi tạo page
			fUser := facebookgraph.FacebookUser{
				Id: fbPage.Id,
				Name: fbPage.Name,
				PageId: fbPage.Id,
			}

			app.Srv.Store.FacebookUid().UpsertFromFbUser(fUser)

			if r := <-app.Srv.Store.FacebookUid().UpsertFromFbUser(fUser); r.Err != nil {
				return "", false, nil, r.Err
			}

			newPage = &model.Fanpage{PageId: fbPage.Id, Name: fbPage.Name, Category: fbPage.Category}
			if pageResult := <-app.Srv.Store.Fanpage().Save(newPage); pageResult.Err == nil {
				newPage = pageResult.Data.(*model.Fanpage)
			} else {
				mlog.Error(pageResult.Err.Error())
				return "", false, nil, pageResult.Err
			}
		} else {
			// page đã có trong DB, kiểm tra nếu page đã khởi tạo hoặc đang khoi tao thì không cho phép khởi tạo
			newPage = pageResult.Data.(*model.Fanpage)
			if newPage.Status == model.PAGE_STATUS_INITIALIZING ||
				newPage.Status == model.PAGE_STATUS_INITIALIZED ||
				newPage.Status == model.PAGE_STATUS_ERROR ||
				newPage.Status == model.PAGE_STATUS_BLOCKED {
				return "", true, nil, nil
			}
		}

		if newPage != nil {
			memberResult := <-app.Srv.Store.Fanpage().GetMember(newPage.Id, user.Id)
			if memberResult.Err != nil {
				pageMember := model.FanpageMember{FanpageId: newPage.Id, PageId: newPage.PageId, UserId: user.Id, AccessToken: fbPage.AccessToken}
				if memberResult := <-app.Srv.Store.Fanpage().SaveFanPageMember(&pageMember); memberResult.Err != nil {
					return "", false, nil, memberResult.Err
				} else {
					return memberResult.Data.(*model.FanpageMember).AccessToken, false, nil, nil
				}
			} else {
				return memberResult.Data.(*model.FanpageMember).AccessToken, false, nil, nil
			}
		}
		return "", false, nil, nil
	} else {
		return "", false, nil, nil
	}
}

// Trả về page token cho user
func (app *App) GraphPageToken(userToken string, pageId string) (*facebookgraph.FacebookToken, *facebookgraph.FacebookError, *model.AppError) {
	body, err, aerr := app.request(userToken, "/"+pageId+"?fields=name,category,access_token", "GET")
	if err != nil || aerr != nil {
		return nil, err, aerr
	}
	var fbtk *facebookgraph.FacebookToken
	fbtk = facebookgraph.FacebookTokenFromJson(body)

	return fbtk, nil, nil
}

func (app *App) ListFanpagePosts(pageId string, pageToken string) (*facebookgraph.FacebookGraphPosts, *facebookgraph.FacebookError) {
	body, err, aerr := app.request(pageToken, "/"+pageId+"/posts?fields=attachments{media,target,type,url},admin_creator,child_attachments,created_time,from,name,picture,updated_time,likes,is_hidden,story,permalink_url,message", "GET")

	if err != nil || aerr != nil {
		return nil, err
	}

	var posts *facebookgraph.FacebookGraphPosts
	posts = facebookgraph.FacebookPostsFromJson(body)
	return posts, nil
}

func (app *App) AddPagePost(post *model.FacebookPost, pageToken string) (*model.FacebookPost, *model.AppError) {
	postResult := <-app.Srv.Store.FacebookPost().Save(post)

	if postResult.Err != nil {
		mlog.Error(fmt.Sprintf("Couldn't save the post err=%v", postResult.Err))
		return nil, postResult.Err
	}
	rPost := postResult.Data.(*model.FacebookPost)
	return rPost, nil
}

// Nhận vào 3 tham số: pageId, pageToken và requestUrl
// tham số requestURL nếu có thì hàm sẽ graph theo requestUrl
// Lý do có requestURL là vì API GET /feed của facebook mỗi lần chỉ return về 25 kết quả
// đồng thời trả về next link trong paging, nếu thực hiện request tiếp sau lần thứ nhất
// thì sẽ truyền next link này như requestURL
// Nếu không truyền requestURL thì hàm chỉ thực hiện 1 lần duy nhất và trả về 25 kết quả
// Trả về 3 tham số: tham số thứ nhất là next link
// 2 tham số cuối là lỗi từ facebook và lỗi từ ứng dụng
func (app *App) GraphPagePostsAndInit(pageId string, pageToken string, requestURL string, initResult *model.FanpageInitResult) (string, *facebookgraph.FacebookError, *model.AppError) {

	isFirstGraph := false
	if len(requestURL) == 0 {
		isFirstGraph = true
		requestURL = "/" + pageId + "/feed?fields=is_hidden,story,permalink_url,message,admin_creator,child_attachments&include_hidden=true&limit=100"
	}

	var body io.ReadCloser
	var err *facebookgraph.FacebookError
	var aerr *model.AppError
	nextPageLink := ""

	var pr *facebookgraph.FacebookGraphPosts

	if isFirstGraph {
		body, err, aerr = app.request(pageToken, requestURL, "GET")
		if err != nil {
			// cần log lỗi từ facebook
			return "", err, nil
		}
		if aerr != nil {
			mlog.Error(aerr.Error())
			return "", nil, aerr
		}
	} else {
		body, err, aerr = app.requestUsingFullpath(pageToken, requestURL, "GET")
	}

	pr = facebookgraph.FacebookPostsFromJson(body)

	// ============= counter
	var postCount int64
	var conversationCount int64
	var commentCount int64
	// ============= end counter

	if len(pr.Data) > 0 {
		for _, post := range pr.Data {
			// kiểm tra nếu có attachments hoặc subattachments
			if len(post.Attachments.Data) > 0 && len(post.Attachments.Data[0].SubAttachments.Data) > 0 {
				for _, sub := range post.Attachments.Data[0].SubAttachments.Data {

					subPostId := pageId + "_" + sub.Target.Id
					subRequestUrl := "/" + subPostId + "?include_hidden=true&fields=is_hidden,story,permalink_url,message,admin_creator,child_attachments&include_hidden=true&limit=100"

					if subBody, err, aerr := app.request(pageToken, subRequestUrl, "GET"); err != nil {
						fmt.Println(err)
					} else if aerr != nil {
						fmt.Println(aerr)
					} else {
						subPost := model.FacebookPostFromJson(subBody)
						subPost.PageId = pageId
						//
						//addedPost, err := app.AddPagePost(subPost, pageToken)
						//if err != nil {
						//	return "", nil, err
						//}
						conversationAdded, commentAdded, e, aErr := app.InitConversationsFromPost(subPost.PostId, pageId, pageToken)
						if e != nil {
							return "", e, nil
						}
						if aErr != nil {
							return "", nil, aErr
						}

						conversationCount += conversationAdded
						commentCount += commentAdded
					}
				}
			}

			// add attachments if there'is only one image in attachment
			//if len(post.Attachments.Data) > 0 {
			//	//fmt.Println("bai viet chi co 1 anh")
			//	for _, media := range post.Attachments.Data {
			//		if media.Type == "photo" {
			//			//fmt.Println("media", media)
			//			image := &model.FacebookAttachmentImage{
			//				PostId: post.Id,
			//				Height: media.Media.Image.Height,
			//				Width: media.Media.Image.Width,
			//				Src: media.Media.Image.Src,
			//				Url: media.Url,
			//				TargetId: media.Target.Id,
			//				TargetUrl: media.Target.Url,
			//			}
			//
			//			image.PreSave()
			//
			//			if _, err := app.AddImage(image); err != nil {
			//				fmt.Println("không thể thêm post image vào db, lỗi: ", err)
			//			}
			//		}
			//	}
			//
			//	if len(post.Attachments.Data[0].SubAttachments.Data) > 0 {
			//		//fmt.Println("bai viet co nhieu anh")
			//		for _, media := range post.Attachments.Data[0].SubAttachments.Data {
			//			if media.Type == "photo" {
			//				//fmt.Println("media", media)
			//				image := &model.FacebookAttachmentImage{
			//					PostId: post.Id,
			//					Height: media.Media.Image.Height,
			//					Width: media.Media.Image.Width,
			//					Src: media.Media.Image.Src,
			//					Url: media.Url,
			//					TargetId: media.Target.Id,
			//					TargetUrl: media.Target.Url,
			//				}
			//
			//				image.PreSave()
			//
			//				if _, err := app.AddImage(image); err != nil {
			//					fmt.Println("không thể thêm post image vào db, lỗi: ", err)
			//				}
			//			}
			//		}
			//	}
			//}


			postToAdd := model.FacebookPostFromPostGraphResult(&post)
			// luôn chắc chắn phải thêm pageId
			postToAdd.PageId = pageId
			// thêm post vào DB
			//addedPost, err := app.AddPagePost(postToAdd, pageToken)
			//if err != nil {
			//	return "", nil, err
			//}

			postCount ++

			conversationAdded1, commentAdded1, e, aErr := app.InitConversationsFromPost(postToAdd.PostId, pageId, pageToken)
			if e != nil {
				return "", e, nil
			}
			if aErr != nil {
				return "", nil, aErr
			}

			conversationCount += conversationAdded1
			commentCount += commentAdded1
		}

		nextPageLink = pr.Paging.Next
	} else {
		nextPageLink = ""
	}

	// update counter
	if initResult != nil && len(initResult.Creator) > 0 {
		app.UpdateFanpageInitResultPostCount(initResult, postCount)
		app.UpdateFanpageInitResultConversationCount(initResult, conversationCount)
		updatedCommentCount, _ := app.UpdateFanpageInitResultCommentCount(initResult, commentCount)

		//message := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_PAGE_INIT_UPDATE_VALUE, "", "", initResult.Creator, nil)
		//message.Add("data", updatedPostCount)
		//app.Publish(message)
		//
		//message2 := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_PAGE_INIT_UPDATE_VALUE, "", "", initResult.Creator, nil)
		//message2.Add("data", updatedConversationCount)
		//app.Publish(message2)

		message3 := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_PAGE_INIT_UPDATE_VALUE, "", "", initResult.Creator, nil)
		message3.Add("data", updatedCommentCount)
		app.Publish(message3)
	}

	if len(nextPageLink) == 0 {
		updatedEndAt, _ := app.UpdateFanpageInitResultEndAt(initResult, time.Now().Unix())
		message := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_PAGE_INIT_UPDATE_VALUE, "", "", initResult.Creator, nil)
		message.Add("data", updatedEndAt)
		message.Add("status", "finished")
		message.Add("page_id", pageId)
		app.Publish(message)
	}

	return nextPageLink, nil, nil
}
//func (app *App) GraphPagePostsAndInit(pageId string, pageToken string, requestURL string, initResult *model.FanpageInitResult) (string, *facebookgraph.FacebookError, *model.AppError) {
//
//	isFirstGraph := false
//	if len(requestURL) == 0 {
//		isFirstGraph = true
//		requestURL = "/" + pageId + "/feed?include_hidden=true&fields=attachments{media,target,type,url,subattachments},admin_creator,child_attachments,created_time,from,name,picture,updated_time,likes,is_hidden,story,permalink_url,message&limit=100"
//	}
//
//	var body io.ReadCloser
//	var err *facebookgraph.FacebookError
//	var aerr *model.AppError
//	nextPageLink := ""
//
//	var pr *facebookgraph.FacebookGraphPosts
//
//	if isFirstGraph {
//		body, err, aerr = app.request(pageToken, requestURL, "GET")
//		if err != nil {
//			// cần log lỗi từ facebook
//			return "", err, nil
//		}
//		if aerr != nil {
//			mlog.Error(aerr.Error())
//			return "", nil, aerr
//		}
//	} else {
//		body, err, aerr = app.requestUsingFullpath(pageToken, requestURL, "GET")
//	}
//
//	pr = facebookgraph.FacebookPostsFromJson(body)
//
//	// ============= counter
//	var postCount int64
//	var conversationCount int64
//	var commentCount int64
//	// ============= end counter
//
//	if len(pr.Data) > 0 {
//		for _, post := range pr.Data {
//			// kiểm tra nếu có attachments hoặc subattachments
//			if len(post.Attachments.Data) > 0 && len(post.Attachments.Data[0].SubAttachments.Data) > 0 {
//				for _, sub := range post.Attachments.Data[0].SubAttachments.Data {
//
//					subPostId := pageId + "_" + sub.Target.Id
//					subRequestUrl := "/" + subPostId + "?include_hidden=true&fields=attachments{media,target,type,url,subattachments},admin_creator,child_attachments,created_time,from,name,picture,updated_time,likes,is_hidden,story,permalink_url,message&limit=100"
//
//					if subBody, err, aerr := app.request(pageToken, subRequestUrl, "GET"); err != nil {
//						fmt.Println(err)
//					} else if aerr != nil {
//						fmt.Println(aerr)
//					} else {
//						subPost := model.FacebookPostFromJson(subBody)
//						subPost.PageId = pageId
//
//						addedPost, err := app.AddPagePost(subPost, pageToken)
//						if err != nil {
//							return "", nil, err
//						}
//						conversationAdded, commentAdded, e, aErr := app.InitConversationsFromPost(addedPost.PostId, pageId, pageToken)
//						if e != nil {
//							return "", e, nil
//						}
//						if aErr != nil {
//							return "", nil, aErr
//						}
//
//						conversationCount += conversationAdded
//						commentCount += commentAdded
//					}
//				}
//			}
//
//			// add attachments if there'is only one image in attachment
//			if len(post.Attachments.Data) > 0 {
//				//fmt.Println("bai viet chi co 1 anh")
//				for _, media := range post.Attachments.Data {
//					if media.Type == "photo" {
//						//fmt.Println("media", media)
//						image := &model.FacebookAttachmentImage{
//							PostId: post.Id,
//							Height: media.Media.Image.Height,
//							Width: media.Media.Image.Width,
//							Src: media.Media.Image.Src,
//							Url: media.Url,
//							TargetId: media.Target.Id,
//							TargetUrl: media.Target.Url,
//						}
//
//						image.PreSave()
//
//						if _, err := app.AddImage(image); err != nil {
//							fmt.Println("không thể thêm post image vào db, lỗi: ", err)
//						}
//					}
//				}
//
//				if len(post.Attachments.Data[0].SubAttachments.Data) > 0 {
//					//fmt.Println("bai viet co nhieu anh")
//					for _, media := range post.Attachments.Data[0].SubAttachments.Data {
//						if media.Type == "photo" {
//							//fmt.Println("media", media)
//							image := &model.FacebookAttachmentImage{
//								PostId: post.Id,
//								Height: media.Media.Image.Height,
//								Width: media.Media.Image.Width,
//								Src: media.Media.Image.Src,
//								Url: media.Url,
//								TargetId: media.Target.Id,
//								TargetUrl: media.Target.Url,
//							}
//
//							image.PreSave()
//
//							if _, err := app.AddImage(image); err != nil {
//								fmt.Println("không thể thêm post image vào db, lỗi: ", err)
//							}
//						}
//					}
//				}
//			}
//
//
//			postToAdd := model.FacebookPostFromPostGraphResult(&post)
//			// luôn chắc chắn phải thêm pageId
//			postToAdd.PageId = pageId
//			// thêm post vào DB
//			addedPost, err := app.AddPagePost(postToAdd, pageToken)
//			if err != nil {
//				return "", nil, err
//			}
//
//			postCount ++
//
//			conversationAdded1, commentAdded1, e, aErr := app.InitConversationsFromPost(addedPost.PostId, pageId, pageToken)
//			if e != nil {
//				return "", e, nil
//			}
//			if aErr != nil {
//				return "", nil, aErr
//			}
//
//			conversationCount += conversationAdded1
//			commentCount += commentAdded1
//		}
//
//		nextPageLink = pr.Paging.Next
//	} else {
//		nextPageLink = ""
//	}
//
//	// update counter
//	if initResult != nil && len(initResult.Creator) > 0 {
//		app.UpdateFanpageInitResultPostCount(initResult, postCount)
//		app.UpdateFanpageInitResultConversationCount(initResult, conversationCount)
//		updatedCommentCount, _ := app.UpdateFanpageInitResultCommentCount(initResult, commentCount)
//
//		//message := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_PAGE_INIT_UPDATE_VALUE, "", "", initResult.Creator, nil)
//		//message.Add("data", updatedPostCount)
//		//app.Publish(message)
//		//
//		//message2 := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_PAGE_INIT_UPDATE_VALUE, "", "", initResult.Creator, nil)
//		//message2.Add("data", updatedConversationCount)
//		//app.Publish(message2)
//
//		message3 := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_PAGE_INIT_UPDATE_VALUE, "", "", initResult.Creator, nil)
//		message3.Add("data", updatedCommentCount)
//		app.Publish(message3)
//	}
//
//	if len(nextPageLink) == 0 {
//		updatedEndAt, _ := app.UpdateFanpageInitResultEndAt(initResult, time.Now().Unix())
//		message := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_PAGE_INIT_UPDATE_VALUE, "", "", initResult.Creator, nil)
//		message.Add("data", updatedEndAt)
//		message.Add("status", "finished")
//		message.Add("page_id", pageId)
//		app.Publish(message)
//	}
//
//	return nextPageLink, nil, nil
//}

// API /conversations không trả về trường From, do đó ta phải
// filter trong mảng senders để lấy ra from
// TODO: nếu hội thoại nhóm thì xử lý thế nào?????
func getConversationFromFromSenders(pageId string, senders []facebookgraph.FacebookUser) map[string]interface{} {
	for _, s := range senders {
		if s.Id != pageId {
			from := make(map[string]interface{})
			from["name"] = s.Name
			from["id"] = s.Id
			//from["email"] = s.Email // chưa cần
			return from
		}
	}
	return nil
}

// Graph messages của một page và insert vào DB
// trả về 3 tham số
// tham số thứ nhất là kết quả init, 2 tham số sau là lỗi phát sinh từ facebook và lỗi phát sinh từ ứng dụng
func (app *App) GraphMessagesAndInit(pageId string, pageToken string, initResult *model.FanpageInitResult) (*facebookgraph.FacebookError, *model.AppError) {

	path := "/" + pageId + "/conversations?fields=can_reply,name,snippet,subject,updated_time,messages{created_time,from,id,message,sticker,attachments},id,scoped_thread_key,senders&limit=100"

	nextPageLink := "next"
	currentPageNumber := 1

	for len(nextPageLink) > 0 {
		// reset nextPageLink
		//nextPageLink = ""
		// ============= counter
		var conversationCount int64
		var messageCount int64
		// ============= end counter

		var url = path

		var body io.ReadCloser
		var err *facebookgraph.FacebookError
		var aerr *model.AppError

		if currentPageNumber > 1 && len(nextPageLink) > 0 {
			url = nextPageLink
			body, err, aerr = app.requestUsingFullpath(pageToken, url, "GET")
		} else {
			body, err, aerr = app.request(pageToken, url, "GET")
		}

		if err != nil {
			mlog.Error(fmt.Sprint(err))
			return err, nil
		} else if aerr != nil {
			mlog.Error(fmt.Sprint(aerr))
			return nil, aerr
		}

		cvs := *facebookgraph.FacebookConversationsFromJson(body)
		// lặp từng hội thoại và load full tin nhắn của hội thoại đó
		for _, c := range cvs.Data {
			conversationCount ++
			// thêm conversation vào database
			fc := model.FacebookConversationModelFromGraphConversationItem(&c)
			fc.From = getConversationFromFromSenders(pageId, c.Senders.Data)["id"].(string)
			fc.PageId = pageId

			// cập nhật user
			for _, u := range c.Senders.Data {
				if u.Id != pageId {
					u.PageId = pageId
					if r := <-app.Srv.Store.FacebookUid().UpsertFromFbUser(u); r.Err != nil {
						return nil, r.Err
					}
				}
			}

			updatedTime, _ := time.Parse("2006-01-02T15:04:05-0700", fc.UpdatedTime)

			// Biến này cho biết hội thoại đã được trả lời hay chưa
			replied := false

			// Biến này cho biết trong hội thoại có bao nhiêu tin nhắn chưa xem,
			// Tin nhắn được coi là chưa xem nếu nó được gửi từ User tới page và phải nằm sau tin nhắn cuối cùng từ Page
			// Nếu tin nhắn cuối cùng trong hội thoại là từ Page thì unreadCount = 0
			unreadCount := 0

			// Biến này lưu giá trị thời gian của tin nhắn mới nhất trong hội thoại của User
			var lastUserMessageAt string
			var lastPageMessageAt string

			// thêm conversation vào db
			var rfc *model.FacebookConversation
			var err *model.AppError // cần
			if rfc, err = app.CreateConversation(fc); err != nil {
				mlog.Error(fmt.Sprint(err))
				return nil, err
			}

			// lặp và thêm messages vào database
			for _, m := range c.Messages.Data {

				mes := &model.FacebookConversationMessage{
					ConversationId: rfc.Id,
					Type: 			"message",
					PageId: 		pageId,
					MessageId:      m.Id,
					Message:        m.Message,
					CreatedTime:    m.CreatedTime,
					From:           m.From["id"].(string),
					Sent: 			true,
				}

				if len(m.Sticker) > 0 {
					mes.Sticker = m.Sticker
				}

				// nếu message có attachments
				if len(m.Attachments.Data) > 0 {
					mes.HasAttachments = true
					mes.AttachmentsCount = len(m.Attachments.Data)

					// chúng ta cần biết type của attachment để client có thể hiển thị placeholder tương ứng với loại attachment
					// trong trường hợp chỉ có 1 attachment thì Attachment type luôn đúng với m.Attachments.Data[0].MimeType
					// tuy nhiên khi có nhiều attachment type, xem xét có trường hợp nào mà người dùng gửi đồng thời nhiều file
					// với định dạng khác nhau không?
					mes.AttachmentType = m.Attachments.Data[0].MimeType
				}

				var er *model.AppError

				if _, _, er = app.AddMessage(mes, false, false, mes.From == pageId); er != nil {
					mlog.Error(fmt.Sprint(er))
					return nil, er
				}

				messageCount ++

				// FSE122 Kiểm tra và cập nhật unread_count, replied cho hội thoại
				from := m.From["id"].(string)
				messageTime, _ := time.Parse("2006-01-02T15:04:05-0700", m.CreatedTime)

				if messageTime.Equal(updatedTime) {
					if from == pageId {
						// Tin nhan tu page
						replied = true
						lastPageMessageAt = m.CreatedTime
					} else {
						// Tin nhan tu user
						lastUserMessageAt = m.CreatedTime
						replied = false
					}
				} else {
					if len(lastUserMessageAt) == 0 {
						// chưa tìm thấy tin nhắn nào từ user trước đó
						if from != pageId {
							lastUserMessageAt = m.CreatedTime
						}
					}

					if len(lastPageMessageAt) == 0 {
						// chưa tìm thấy tin nhắn mới nhất từ page
						if from == pageId {
							lastPageMessageAt = m.CreatedTime
						}
					}
				}

				if !replied {
					lastPageMessageAtTime, _ := time.Parse("2006-01-02T15:04:05-0700", lastPageMessageAt)
					messageTime, _ := time.Parse("2006-01-02T15:04:05-0700", m.CreatedTime)
					if messageTime.After(lastPageMessageAtTime) {
						unreadCount += 1
					}
				}
				// END: FSE122
			}

			//fmt.Println("Conversation mới được tạo: "+rfc.Id)
			// nếu conversation này có paging thì lặp paging và tiếp tục thêm vào database
			nextPage := c.Messages.Paging.Next
			for len(nextPage) > 0 {
				b, e, ae := app.requestUsingFullpath(pageToken, nextPage, "GET")
				if e != nil || ae != nil {
					mlog.Error(fmt.Sprint(e))
					return e, ae
				}
				pagingMessages := facebookgraph.FacebookConversationMessagesFromJson(b)

				// lặp qua lần lượt từng message và thêm vào database
				// đầu tiên lưu attachment
				for _, m := range pagingMessages.Data {
					message := &model.FacebookConversationMessage{
						ConversationId: rfc.Id,
						Type: 			"message",
						PageId: 		pageId,
						MessageId:      m.Id,
						Message:        m.Message,
						CreatedTime:    m.CreatedTime,
						From:           m.From["id"].(string),
						Sent: 			true,
					}

					// nếu message có attachments
					if len(m.Attachments.Data) > 0 {
						message.HasAttachments = true
						message.AttachmentsCount = len(m.Attachments.Data)

						// chúng ta cần biết type của attachment để client có thể hiển thị placeholder tương ứng với loại attachment
						// trong trường hợp chỉ có 1 attachment thì Attachment type luôn đúng với m.Attachments.Data[0].MimeType
						// tuy nhiên khi có nhiều attachment type, xem xét có trường hợp nào mà người dùng gửi đồng thời nhiều file
						// với định dạng khác nhau không?
						message.AttachmentType = m.Attachments.Data[0].MimeType
					}

					// thêm message vào database
					var er *model.AppError
					if _, _, er = app.AddMessage(message, false, false, message.From == pageId); er != nil {
						mlog.Error(fmt.Sprint(er))
						return nil, er
					}

					messageCount ++

					// FSE122 Kiểm tra và cập nhật unread_count, replied cho hội thoại
					from := m.From["id"].(string)
					messageTime, _ := time.Parse("2006-01-02T15:04:05-0700", m.CreatedTime)

					if messageTime.Equal(updatedTime) {
						if from == pageId {
							// Tin nhan tu page
							replied = true
							lastPageMessageAt = m.CreatedTime
						} else {
							// Tin nhan tu user
							lastUserMessageAt = m.CreatedTime
							replied = false
						}
					} else {
						if len(lastUserMessageAt) == 0 {
							// chưa tìm thấy tin nhắn nào từ user trước đó
							if from != pageId {
								lastUserMessageAt = m.CreatedTime
							}
						}

						if len(lastPageMessageAt) == 0 {
							// chưa tìm thấy tin nhắn mới nhất từ page
							if from == pageId {
								lastPageMessageAt = m.CreatedTime
							}
						}
					}

					if !replied {
						lastPageMessageAtTime, _ := time.Parse("2006-01-02T15:04:05-0700", lastPageMessageAt)
						messageTime, _ := time.Parse("2006-01-02T15:04:05-0700", m.CreatedTime)
						if messageTime.After(lastPageMessageAtTime) {
							unreadCount += 1
						}
					}
					// END: FSE122
				}
				nextPage = pagingMessages.Paging.Next
			}

			// do not update snippet
			// cập nhật hội thoại gốc
			updateResult := <-app.Srv.Store.FacebookConversation().UpdateConversationUnread(rfc.Id, replied, unreadCount, lastUserMessageAt)
			if updateResult.Err != nil {
				mlog.Error(fmt.Sprintf("Couldn't update pages status err=%v", updateResult.Err))
			}
		}

		currentPageNumber = currentPageNumber + 1
		nextPageLink = cvs.Paging.Next

		// update counter
		if initResult != nil && len(initResult.Creator) > 0 {
			app.UpdateFanpageInitResultConversationCount(initResult, conversationCount)
			updatedMessageCount, _ := app.UpdateFanpageInitResultMessageCount(initResult, messageCount)

			message := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_PAGE_INIT_UPDATE_VALUE, "", "", initResult.Creator, nil)
			message.Add("data", updatedMessageCount)
			app.Publish(message)

			//message2 := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_PAGE_INIT_UPDATE_VALUE, "", "", initResult.Creator, nil)
			//message2.Add("data", updatedConversationCount)
			//app.Publish(message2)
		}
	}

	return nil, nil // chúng ta cần return kết quả init
}

// graph comments của một post và insert vào database
// nhận vào 4 tham số, postId, pageId, pageToken và url
// nếu không truyền url thì request url sẽ được build trong hàm, nếu truyền vào url thì sẽ dùng url để request
// trả về 3 tham số: tham số thứ nhất là nextPageLink nếu kết quả graph vẫn còn next page
// tham số thứ 2 là lỗi trong quá trình graph facebook
// tham số thứ 3 là lỗi ứng dụng
// tham số thứ 4 là số lượng conversation mới đã thêm
// tham số thứ 5 là số lượng comment đã thêm
func (app *App) DoGraphCommentsAndInit(postId string, pageId string, pageToken string, url string) (string, *facebookgraph.FacebookError, *model.AppError, int64, int64) {
	isFirstGraph := false
	if len(url) == 0 {
		isFirstGraph = true
		url = "/" + postId + "?filter=stream&fields=comments{attachment,can_comment,can_hide,can_like,can_remove,can_reply_privately,created_time,id,is_hidden,is_private,message,parent,permalink_url,private_reply_conversation,comments{attachment,can_comment,can_hide,can_like,can_remove,can_reply_privately,created_time,from,id,is_hidden,is_private,message,message_tags,parent,permalink_url,private_reply_conversation},admin_creator,message_tags,from}&limit=100"
	}

	var body io.ReadCloser
	var err *facebookgraph.FacebookError
	var aerr *model.AppError

	var cms1 *facebookgraph.FacebookGraphPostCommentsResponse
	var cms2 *facebookgraph.FacebookGraphPostCommentsResponseData // kết quả trả về lần graph thứ 2 trở đi có chút khác
	var comments []facebookgraph.FacebookCommentItem
	nextPageLink := ""

	if isFirstGraph {
		body, err, aerr = app.request(pageToken, url, "GET")
		if err != nil || aerr != nil {
			return "", err, aerr, 0, 0
		}
		cms1 = facebookgraph.FacebookGraphPostCommentsResponseFromJson(body)
		comments = cms1.Comments.Data
		nextPageLink = cms1.Comments.Paging.Next
	} else {
		body, err, aerr = app.requestUsingFullpath(pageToken, url, "GET")
		if err != nil || aerr != nil {
			return "", err, aerr, 0, 0
		}
		cms2 = facebookgraph.FacebookGraphPostCommentsResponseNextPageFromJson(body)
		comments = cms2.Data
		nextPageLink = cms2.Paging.Next
	}

	// ============= counter
	var conversationCount int64
	var commentCount int64
	// ============= end counter

	for _, c := range comments {

		// Nếu app không có quyền truy cập trường comments->from => kết thúc quá trình khởi tạo
		if c.From == nil {
			break
		}

		// append Page id to from
		c.From["page_id"] = pageId

		// cập nhật uid
		app.Srv.Store.FacebookUid().UpsertFromMap(c.From)

		var snippet string
		var updatedTime string
		replied := false
		unreadCount := 0
		var lastUserMessageAt string

		needUpdate := false

		// thêm conversation vào db
		var newConversation *model.FacebookConversation

		// tim kiem hoac them hoi thoai neu chua co
		foundConversation := <-app.Srv.Store.FacebookConversation().InsertConversationFromCommentIfNeed(c.Parent.Id, c.Id, pageId, postId, c.From["id"].(string), c.CreatedTime, c.Message)
		if foundConversation.Data == nil {
			return "", nil, model.NewAppError("HandleIncomingWebhook", "web.incoming_webhook.parse.app_error", nil, "", http.StatusBadRequest), 0, 0
		} else {
			newConversation = foundConversation.Data.(*model.UpsertConversationResult).Data

			if foundConversation.Data.(*model.UpsertConversationResult).IsNew {
				conversationCount ++
			}
			//unreadCount = newConversation.UnreadCount
			//replied = newConversation.Replied
			lastUserMessageAt = newConversation.LastUserMessageAt
		}

		lastUserMessageAt = newConversation.LastUserMessageAt
		//lastUserMessageAtFromConversation, _ := time.Parse("2006-01-02T15:04:05-0700", lastUserMessageAt)
		lastUserMessageAtTime, _ := time.Parse("2006-01-02T15:04:05-0700", lastUserMessageAt)

		//Mỗi comment đồng thời cũng là 1 FacebookConversationMessage thêm message này vào DB
		rootMessage := model.FacebookConversationMessageModelFromCommentItem(&c)
		rootMessage.ConversationId = newConversation.Id
		rootMessage.Type = "comment"
		rootMessage.PageId = pageId
		rootMessage.CommentId = c.Id

		// Kiểm tra nếu có attachment
		// mỗi comment chỉ có duy nhất 1 attachment, attachment trong comment luôn có target.
		// chúng ta sẽ lưu target này vào message.attachment_target_ids
		// mà không cần tạo attachment item vào db
		// sau này client sẽ query dựa vào target id này
		if c.Attachment != (facebookgraph.CommentAttachmentImage{}) {
			rootMessage.AttachmentType = c.Attachment.Type

			if len(c.Attachment.Target.Id) > 0 {
				var attachedIds []string

				attachedIds = append(attachedIds, c.Attachment.Target.Id)

				rootMessage.AttachmentTargetIds = attachedIds
			}

			if len(rootMessage.Message) == 0 {
				snippet = "[Photo]" + snippet
			}

			if c.Attachment.Type == "sticker" {
				rootMessage.Sticker = c.Attachment.Url
			}

			//if c.Attachment.Type == "sticker" {
			//	rootMessage.Sticker = c.Attachment.Url
			//} else if c.Attachment.Type == "animated_image_share" {
			//
			//} else if c.Attachment.Type == "animated_image_video" {
			//
			//} else if c.Attachment.Type == "photo" {
			//
			//} else if c.Attachment.Type == "video_inline" {
			//
			//}
		}

		// thêm message vào database
		var er *model.AppError
		var rms *model.FacebookConversationMessage
		if rms, _, er = app.AddMessage(rootMessage, false, false, rootMessage.From == pageId); er != nil {
			return "", nil, er, 0, 0
		}

		commentCount ++

		timeFromComment, _ := time.Parse("2006-01-02T15:04:05-0700", rootMessage.CreatedTime)
		timeFromConversation, _ := time.Parse("2006-01-02T15:04:05-0700", newConversation.UpdatedTime)

		if timeFromComment.After(timeFromConversation) || timeFromComment.Equal(timeFromConversation) {
			needUpdate = true
			snippet = utils.GetSnippet(rootMessage.Message)
			updatedTime = rootMessage.CreatedTime
			replied = rms.From == pageId

			// kiểm tra và cập nhật lastUserMessageAt cho conversation
			if timeFromComment.After(lastUserMessageAtTime) {
				if rms.From != pageId {
					lastUserMessageAt = rootMessage.CreatedTime
					unreadCount += 1
				} else {
					unreadCount = 0
				}
			}

			if rms.From == pageId {
				unreadCount = 0
			}
		}

		if len(c.Comments.Data) > 0 {
			// lặp qua tất cả subcomments của comment gốc và thêm vào DB
			for _, cm := range c.Comments.Data {
				// cập nhật uid
				cm.From["page_id"] = pageId
				app.Srv.Store.FacebookUid().UpsertFromMap(cm.From)

				subMessage := model.FacebookConversationMessageModelFromCommentItem(&cm)
				subMessage.ConversationId = newConversation.Id
				subMessage.Type = "comment"
				subMessage.PageId = pageId
				subMessage.CommentId = cm.Id

				// Kiểm tra nếu có attachment
				// mỗi comment chỉ có duy nhất 1 attachment, attachment trong comment luôn có target.
				// chúng ta sẽ lưu target này vào message.attachment_target_ids
				// mà không cần tạo attachment item vào db
				// sau này client sẽ query dựa vào target id này
				if cm.Attachment != (facebookgraph.CommentAttachmentImage{}) {
					subMessage.AttachmentType = cm.Attachment.Type

					if len(cm.Attachment.Target.Id) > 0 {
						var attachedIds []string

						attachedIds = append(attachedIds, cm.Attachment.Target.Id)

						subMessage.AttachmentTargetIds = attachedIds
					}

					if len(subMessage.Message) == 0 {
						snippet = "[Photo]" + snippet
					}

					if cm.Attachment.Type == "sticker" {
						subMessage.Sticker = cm.Attachment.Url
					}

					//if c.Attachment.Type == "sticker" {
					//	rootMessage.Sticker = c.Attachment.Url
					//} else if c.Attachment.Type == "animated_image_share" {
					//
					//} else if c.Attachment.Type == "animated_image_video" {
					//
					//} else if c.Attachment.Type == "photo" {
					//
					//} else if c.Attachment.Type == "video_inline" {
					//
					//}
				}

				// thêm message vào database
				var er *model.AppError
				var x *model.FacebookConversationMessage
				if x, _, er = app.AddMessage(subMessage, false, false, subMessage.From == pageId); er != nil {
					return "", nil, er, 0, 0
				}


				timeFromComment, _ := time.Parse("2006-01-02T15:04:05-0700", x.CreatedTime)

				if timeFromComment.After(timeFromConversation) {
					needUpdate = true
					snippet = utils.GetSnippet(subMessage.Message)
					updatedTime = subMessage.CreatedTime
					replied = x.From == pageId

					// kiểm tra và cập nhật lastUserMessageAt cho conversation
					lastUserMessageAtTime, _ := time.Parse("2006-01-02T15:04:05-0700", lastUserMessageAt)
					if timeFromComment.After(lastUserMessageAtTime) || timeFromComment.Equal(lastUserMessageAtTime) {
						if x.From != pageId {
							lastUserMessageAt = x.CreatedTime
							unreadCount += 1
						} else {
							unreadCount = 0
						}
					}

					if x.From == pageId {
						unreadCount = 0
					}
				}

				//if cm.Attachment != (facebookgraph.CommentAttachmentImage{}) {
				//	snippet = "[Photo]" + snippet
				//	// thêm attachment vào db
				//	image := &model.FacebookAttachmentImage{
				//		ConversationType: "comment",
				//		MessageId:        x.Id,
				//		Url:              cm.Attachment.Url,
				//		Src:              cm.Attachment.Media.Image.Src,
				//		Height:           cm.Attachment.Media.Image.Height,
				//		Width:            cm.Attachment.Media.Image.Width,
				//		TargetId: 		  cm.Attachment.Target.Id,
				//		TargetUrl: 		  cm.Attachment.Target.Url,
				//	}
				//
				//	// thêm image vào db
				//	if _, err := app.AddImage(image); err != nil {
				//		mlog.Error(fmt.Sprint(err))
				//		return "", nil, err, 0, 0
				//	}
				//}
			}

			// paging trong sub comments
			nextSubCommentsLink := c.Comments.Paging.Next
			for len(nextSubCommentsLink) > 0 {
				b, e, ae := app.requestUsingFullpath(pageToken, nextSubCommentsLink, "GET")
				if e != nil || ae != nil {
					return "", e, ae, 0, 0
				}
				cms3 := facebookgraph.FacebookGraphPostCommentsResponseNextPageFromJson(b)
				for _, s := range cms3.Data {
					// cập nhật uid
					s.From["page_id"] = pageId
					app.Srv.Store.FacebookUid().UpsertFromMap(s.From)

					sbm := model.FacebookConversationMessageModelFromCommentItem(&s)
					sbm.ConversationId = newConversation.Id
					sbm.CommentId = s.Id

					// Kiểm tra nếu có attachment
					// mỗi comment chỉ có duy nhất 1 attachment, attachment trong comment luôn có target.
					// chúng ta sẽ lưu target này vào message.attachment_target_ids
					// mà không cần tạo attachment item vào db
					// sau này client sẽ query dựa vào target id này
					if s.Attachment != (facebookgraph.CommentAttachmentImage{}) {
						sbm.AttachmentType = s.Attachment.Type

						if len(s.Attachment.Target.Id) > 0 {
							var attachedIds []string

							attachedIds = append(attachedIds, s.Attachment.Target.Id)

							sbm.AttachmentTargetIds = attachedIds
						}

						if len(sbm.Message) == 0 {
							snippet = "[Photo]" + snippet
						}

						if s.Attachment.Type == "sticker" {
							sbm.Sticker = s.Attachment.Url
						}

						//if c.Attachment.Type == "sticker" {
						//	rootMessage.Sticker = c.Attachment.Url
						//} else if c.Attachment.Type == "animated_image_share" {
						//
						//} else if c.Attachment.Type == "animated_image_video" {
						//
						//} else if c.Attachment.Type == "photo" {
						//
						//} else if c.Attachment.Type == "video_inline" {
						//
						//}
					}

					// thêm message vào database
					var er *model.AppError
					var scms *model.FacebookConversationMessage
					if scms, _, er = app.AddMessage(sbm, false, false, sbm.From == pageId); er != nil {
						return "", nil, er, 0, 0
					}


					timeFromComment, _ := time.Parse("2006-01-02T15:04:05-0700", scms.CreatedTime)

					if timeFromComment.After(timeFromConversation) {
						needUpdate = true
						snippet = utils.GetSnippet(sbm.Message)
						updatedTime = sbm.CreatedTime
						replied = scms.From == pageId

						lastUserMessageAtTime, _ := time.Parse("2006-01-02T15:04:05-0700", lastUserMessageAt)
						if timeFromComment.After(lastUserMessageAtTime) || timeFromComment.Equal(lastUserMessageAtTime) {
							if scms.From != pageId {
								lastUserMessageAt = scms.CreatedTime
								unreadCount += 1
							} else {
								unreadCount = 0
							}
						}

						if scms.From == pageId {
							unreadCount = 0
						}
					}

					//if s.Attachment != (facebookgraph.CommentAttachmentImage{}) {
					//	snippet = "[Photo]" + snippet
					//	// thêm attachment vào db
					//	image := &model.FacebookAttachmentImage{
					//		ConversationType: "comment",
					//		MessageId:        scms.Id,
					//		Url:              s.Attachment.Url,
					//		Src:              s.Attachment.Media.Image.Src,
					//		Height:           s.Attachment.Media.Image.Height,
					//		Width:            s.Attachment.Media.Image.Width,
					//		TargetId: 		  s.Attachment.Target.Id,
					//		TargetUrl: 		  s.Attachment.Target.Url,
					//	}
					//
					//	// thêm image vào db
					//	if _, err := app.AddImage(image); err != nil {
					//		mlog.Error(fmt.Sprint(err))
					//		return "", nil, err, 0, 0
					//	}
					//}
				}
				nextSubCommentsLink = cms3.Paging.Next
			}
		}
		// Cần xem lại ở đây có paging không?????????????????????????

		if needUpdate {
			// cập nhật hội thoại gốc
			updateResult := <-app.Srv.Store.FacebookConversation().UpdateConversation(newConversation.Id, snippet, replied, updatedTime, unreadCount, lastUserMessageAt)
			if updateResult.Err != nil {
				mlog.Error(fmt.Sprintf("Couldn't update pages status err=%v", updateResult.Err))
			}
		}
	}

	return nextPageLink, nil, nil, conversationCount, commentCount
}

// Graph comments từ một bài viết của page và insert vào DB
// postId "465465165165165165_6516849849984984984"
func (app *App) InitConversationsFromPost(postId string, pageId string, pageToken string) (int64, int64, *facebookgraph.FacebookError, *model.AppError) {
	// request lần đầu
	var conversationAdded int64
	var commentAdded int64
	nextLink, err, aerr, c1, c2 := app.DoGraphCommentsAndInit(postId, pageId, pageToken, "")
	if err != nil || aerr != nil {
		return 0, 0, err, aerr
	}

	conversationAdded += c1
	commentAdded += c2

	// nếu kết quả trả về còn nextLink thì tiếp tục request
	for len(nextLink) > 0 {
		var c3 int64
		var c4 int64
		nextLink, err, aerr, c3, c4 = app.DoGraphCommentsAndInit(postId, pageId, pageToken, nextLink)
		if err != nil || aerr != nil {
			return 0, 0, err, aerr
		}

		conversationAdded += c3
		commentAdded += c4
	}
	return conversationAdded, commentAdded, nil, nil
}

func (app *App) DoGraphConversationMessage(pageAccessToken string, messageId string) *facebookgraph.FacebookMessageItem {
	body, err, aerr := app.request(pageAccessToken, "/"+messageId+"?fields=created_time,from,id,message,sticker,tags,to,attachments,shares", "GET")
	if err != nil {
		fmt.Println(err)
		return nil
	} else if aerr != nil {
		fmt.Println(aerr)
		return nil
	} else {
		messageItem := facebookgraph.FacebookMessageItemFromJson(body)
		return messageItem
	}
}

func (app *App) DoGraphOnePost(pageAccessToken string, postId string) *facebookgraph.FacebookGraphPost {
	body, err, aerr := app.request(pageAccessToken, "/"+postId+"?fields=attachments{media,target,type,url,subattachments},admin_creator,child_attachments,created_time,from,name,picture,updated_time,likes,is_hidden,story,permalink_url,message", "GET")
	if err != nil {
		fmt.Println(err)
		return nil
	} else if aerr != nil {
		fmt.Println(aerr)
		return nil
	} else {
		fbPost := facebookgraph.FacebookPostFromJson(body)
		return fbPost
	}
}

func (app *App) DoGraphOneComment(pageAccessToken string, commentId string) *facebookgraph.FacebookCommentItem {
	body, err, aerr := app.request(pageAccessToken, "/"+commentId+"?fields=from,attachment,can_comment,can_hide,can_like,can_remove,can_reply_privately,created_time,id,is_hidden,is_private,message,parent,permalink_url,private_reply_conversation", "GET")
	if err != nil {
		fmt.Println(err)
		return nil
	} else if aerr != nil {
		fmt.Println(aerr)
		return nil
	} else {
		graphComment := facebookgraph.FacebookCommentItemFromJson(body)
		return graphComment
	}
}

func (app *App) doFacebookPostRequest(token string, path string, data *model.ConversationReply) (io.ReadCloser, *facebookgraph.FacebookError, *model.AppError) {
	reqUrl := FACEBOOK_API_ROOT + path

	p := url.Values{}
	p.Set("message", data.Message)
	p.Set("attachment_url", data.AttachmentUrl)

	req, _ := http.NewRequest("POST", reqUrl, strings.NewReader(p.Encode()))

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	if resp, err := app.HTTPService.MakeClient(true).Do(req); err != nil {
		mlog.Error(err.Error())
		fmt.Println("===================================================================")
		fmt.Println("err", err)
		fmt.Println("===================================================================")
		// Cần xử lý thêm ở đoạn này
		return nil, nil, model.NewAppError("doFacebookRequest", "missing", nil, "", http.StatusBadRequest)
	} else {
		var bodyBytes []byte
		bodyBytes, _ = ioutil.ReadAll(resp.Body)
		resp.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

		if resp.StatusCode != http.StatusOK {
			return nil, facebookgraph.FacebookErrorFromJson(resp.Body), nil
		}

		return resp.Body, nil, nil
	}
}

func (app *App) replyMessage(token string, path string, data *model.MessageGraphReply) (io.ReadCloser, *facebookgraph.FacebookError, *model.AppError) {
	reqUrl := FACEBOOK_API_ROOT + path

	p := url.Values{}
	p.Set("message", data.Message.ToJson())
	p.Set("recipient", data.Recipient.ToJson())

	req, _ := http.NewRequest("POST", reqUrl, strings.NewReader(p.Encode()))

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	if resp, err := app.HTTPService.MakeClient(true).Do(req); err != nil {
		mlog.Error(err.Error())
		fmt.Println("===================================================================")
		fmt.Println("err", err)
		fmt.Println("===================================================================")
		// Cần xử lý thêm ở đoạn này
		return nil, nil, model.NewAppError("doFacebookRequest", "missing", nil, "", http.StatusBadRequest)
	} else {
		var bodyBytes []byte
		bodyBytes, _ = ioutil.ReadAll(resp.Body)
		resp.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

		if resp.StatusCode != http.StatusOK {
			return nil, facebookgraph.FacebookErrorFromJson(resp.Body), nil
		}

		return resp.Body, nil, nil
	}
}

func (app *App) ReplyComment(commentId string, replyItem *model.ConversationReply) (io.ReadCloser, *facebookgraph.FacebookError, *model.AppError) {

	response, fErr, aErr := app.doFacebookPostRequest(replyItem.PageToken, "/"+commentId+"/comments", replyItem)
	if fErr != nil || aErr != nil {
		fmt.Println("fErr", fErr)
		fmt.Println("aErr", aErr)
		return nil, fErr, aErr
	}

	return response, nil, nil
}

func (app *App) ReplyMessage(threadId, psId string, replyItem *model.ConversationReply) (io.ReadCloser, *facebookgraph.FacebookError, *model.AppError) {

	to := replyItem.To

	if len(to) == 0 || len(psId) == 0 {
		return nil, nil, model.NewAppError("ReplyMessage", "web.reply_message.missing_user_id.app_error", nil, "", http.StatusBadRequest)
	}
	//var psid string
	//var getPsIdErr *facebookgraph.FacebookError
	//var getAppErr *model.AppError
	//if fbUser != nil && len(fbUser.PageScopeId) > 0 {
	//	psid = fbUser.PageScopeId
	//} else {
	//	// get match Page Scope Id from App Scope Id
	//	psid, getPsIdErr, getAppErr = app.MatchPageScopeId(replyItem.PageId, pageAccessToken, to)
	//	if getPsIdErr != nil || getAppErr != nil {
	//		return nil, getPsIdErr, getAppErr
	//	}
	//
	//	// Update pageId if need (for old version)
	//	if len(fbUser.PageId) == 0 {
	//		_, updatedErr := app.UpdateFacebookUserPageId(to, replyItem.PageId)
	//		if updatedErr != nil {
	//			// TODO: Log this err to mlog
	//		}
	//	}
	//
	//	// update page scope id
	//	_, updatedErr := app.UpdateFacebookUserPageScopeId(to, psid)
	//	if updatedErr != nil {
	//		// TODO: Log this err to mlog
	//	}
	//}

	message := &model.MessageGraphReply{
		Recipient: &model.Recipient{Id: psId},
		Message: &model.Message{
			Text: replyItem.Message,
			//Attachment: &model.Attachment{
			//	Type: "image",
			//	Payload: &model.Payload{
			//		Url: "https://wordtracker-swoop-uploads.s3.amazonaws.com/uploads/ckeditor/pictures/1217/content_emoji_smiles.jpg",
			//		IsReusable: true,
			//	},
			//},
		},
	}

	response, fErr, aErr := app.replyMessage(replyItem.PageToken, "/me/messages", message)
	if fErr != nil || aErr != nil {
		fmt.Println("fErr", fErr)
		fmt.Println("aErr", aErr)
		return nil, fErr, aErr
	}

	return response, nil, nil
}

// Get page_scope_id from app_scope_id and update to database
func (a *App) MatchPageScopeId(pageId, pageAccessToken, appScopeId string) (string, *facebookgraph.FacebookError, *model.AppError) {
	appToken := *a.Config().FacebookSettings.AppToken
	appSecretProof := *a.Config().FacebookSettings.AppSecretProof

	fmt.Println("appToken", appToken)
	fmt.Println("appSecretProof", appSecretProof)
	body, err, aerr := a.request(pageAccessToken, "/"+appScopeId+"/ids_for_pages?page="+pageId+"&access_token="+appToken+"&appsecret_proof="+appSecretProof, "GET")
	if err != nil {
		fmt.Println(err)
		return "", err, nil
	} else if aerr != nil {
		return "", nil, aerr
	} else {
		response := facebookgraph.GetPageScopeIdResponseFromJson(body)
		if response != nil && len(response.Data) > 0 {
			return response.Data[0].Id, nil, nil
		}

		// TODO: return an error here
		return "", nil, nil
	}
}
