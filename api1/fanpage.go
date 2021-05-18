// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package api1

import (
	"bitbucket.org/enesyteam/papo-server/facebook_graph"
	"bitbucket.org/enesyteam/papo-server/mlog"
	"bitbucket.org/enesyteam/papo-server/model"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

func (api *API) InitFanpage() {
	api.BaseRoutes.Fanpages.Handle("", api.ApiSessionRequired(getFanpages)).Methods("GET")
	api.BaseRoutes.Fanpages.Handle("/status", api.ApiSessionRequired(updatePagesStatus)).Methods("POST")

	// get a page
	api.BaseRoutes.Fanpage.Handle("", api.ApiSessionRequired(getFanpage)).Methods("GET")
	api.BaseRoutes.Fanpages.Handle("/members/{user_id:[A-Za-z0-9]+}/view", api.ApiSessionRequired(viewPage)).Methods("POST")
	// get page reply snippets
	api.BaseRoutes.Fanpage.Handle("/snippets", api.ApiSessionRequired(getPageSnippets)).Methods("GET")
	// create snippet
	api.BaseRoutes.Fanpage.Handle("/snippets", api.ApiSessionRequired(createSnippet)).Methods("POST")
	api.BaseRoutes.Fanpage.Handle("/status", api.ApiSessionRequired(updatePageStatus)).Methods("POST")
	// update snippet
	api.BaseRoutes.Fanpage.Handle("/snippets/{snippet_id:[A-Za-z0-9]+}/update", api.ApiSessionRequired(updateSnippet)).Methods("PUT")
	// initialize pages
	api.BaseRoutes.Fanpages.Handle("/initialize", api.ApiSessionRequired(initializeFanpages)).Methods("POST")

	// AUTO MESSAGE TASK
	api.BaseRoutes.Fanpage.Handle("/auto_message_tasks", api.ApiSessionRequired(createAutoMessageTask)).Methods("POST")
	// get page AUTO MESSAGE TASKS
	api.BaseRoutes.Fanpage.Handle("/auto_message_tasks", api.ApiSessionRequired(getPageAutoMessageTasks)).Methods("GET")
	// DEMO: GET api/v1/users/2132321dsfdf/fanpages
	api.BaseRoutes.FanpagesForUser.Handle("", api.ApiSessionRequired(getUserFanpages)).Methods("GET")

	// get page images
	api.BaseRoutes.Fanpage.Handle("/images", api.ApiSessionRequired(getFileInfosForPage)).Methods("GET")

}

func viewPage(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireUserId()
	if c.Err != nil {
		return
	}

	if !c.App.SessionHasPermissionToUser(c.App.Session, c.Params.UserId) {
		c.SetPermissionError(model.PERMISSION_EDIT_OTHER_USERS)
		return
	}

	view := model.PageViewFromJson(r.Body)
	if view == nil {
		c.SetInvalidParam("page_view")
		return
	}

	// Validate view struct
	// Check IDs are valid or blank. Blank IDs are used to denote focus loss or inital page view.
	if view.PageId != "" && !model.IsValidId(view.PageId) {
		c.SetInvalidParam("page_view.page_id")
		return
	}

	times, err := c.App.ViewPage(view, c.Params.UserId, c.App.Session.Id)
	if err != nil {
		c.Err = err
		return
	}

	c.App.UpdateLastActivityAtIfNeeded(c.App.Session)

	// Returning {"status": "OK", ...} for backwards compatibility
	resp := &model.PageViewResponse{
		Status:            "OK",
		LastViewedAtTimes: times,
	}

	w.Write([]byte(resp.ToJson()))
}

func updatePagesStatus(c *Context, w http.ResponseWriter, r *http.Request) {

	status := model.PagesStatusFromJson(r.Body)
	if len(status.Status) == 0 {
		c.SetInvalidParam("Status")
		return
	}

	if len(status.PageIds) == 0 {
		c.SetInvalidParam("Page Ids")
		return
	}

	lpi := model.LoadPagesInput{
		PageIds: status.PageIds,
	}

	err, _ := c.App.UpdatePagesStatus(&lpi, status.Status, c.App.Session.UserId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.ToJson()))
	}
	ReturnStatusOK(w)
}

func updatePageStatus(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequirePageId()
	if c.Err != nil {
		return
	}

	status := model.PageStatusFromJson(r.Body)
	if len(status.Status) == 0 {
		c.SetInvalidParam("Status")
		return
	}

	if len(c.Params.PageId) > 0 {
		err, _ := c.App.UpdatePageStatus(c.Params.PageId, status.Status, c.App.Session.UserId)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.ToJson()))
		}
		ReturnStatusOK(w)
	}
}

func getFileInfosForPage(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequirePageId()
	if c.Err != nil {
		return
	}

	if !c.App.SessionHasPermissionToChannelByPost(c.App.Session, c.Params.PageId, model.PERMISSION_READ_CHANNEL) {
		c.SetPermissionError(model.PERMISSION_READ_CHANNEL)
		return
	}

	query := r.URL.Query()
	limit, _ := strconv.Atoi(query.Get("limit"))
	offset, _ := strconv.Atoi(query.Get("offset"))

	infos, err := c.App.GetFileInfosForPage(c.Params.PageId, false, offset, limit)
	if err != nil {
		c.Err = err
		return
	}

	//if c.HandleEtag(model.GetEtagForFileInfos(infos), "Get File Infos For Page", w, r) {
	//	return
	//}

	//fmt.Println(len(model.FileInfosToJson(infos)))
	//
	//w.Header().Set("Cache-Control", "max-age=2592000, public")
	//w.Header().Set(model.HEADER_ETAG_SERVER, model.GetEtagForFileInfos(infos))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(model.FileInfosToJson(infos)))
}

func getPageAutoMessageTasks(c *Context, w http.ResponseWriter, r *http.Request) {
	pageId := c.Params.PageId
	if len(pageId) == 0 {
		c.SetInvalidUrlParam("PageId")
		return
	}

	if p, err := c.App.GetPageAutoMessageTasks(pageId); err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.ToJson()))
		return
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(model.AutoMessageTasksToJson(p)))
	}
}

func createAutoMessageTask(c *Context, w http.ResponseWriter, r *http.Request) {
	task := model.AutoMessageTaskFromJson(r.Body)

	pageId := c.Params.PageId
	task.PageId = pageId
	task.Creator = c.App.Session.UserId

	if len(pageId) == 0 {
		c.SetInvalidUrlParam("PageId")
		return
	}

	if task == nil {
		c.SetInvalidParam("missing")
		return
	}

	if len(task.PageId) == 0 {
		c.SetInvalidParam("PageId")
		return
	}

	if len(task.Message) == 0 {
		c.SetInvalidParam("Message")
		return
	}

	if task.FilterFromDate == 0 {
		c.SetInvalidParam("FilterFromDate")
		return
	}

	if task.FilterToDate == 0 {
		c.SetInvalidParam("FilterToDate")
		return
	}

	// cho phép tạo chiến dịch không đặt tên, hệ thống sẽ tự đặt tên cho chiến dịch
	if len(task.Name) == 0 {
		task.Name = "No name " + strconv.FormatInt(model.GetMillis(), 10)
	}

	var rTask *model.AutoMessageTask
	var err *model.AppError

	rTask, err = c.App.CreateAutoMessageTask(task)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.ToJson()))
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(rTask.ToJson()))
}

func getPageSnippets(c *Context, w http.ResponseWriter, r *http.Request) {
	pageId := c.Params.PageId
	if len(pageId) > 0 {
		if p, err := c.App.GetPageSnippets(pageId); err != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(err.ToJson()))
			return
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(model.ReplySnippetListToJson(p)))
		}
	}
}

func createSnippet(c *Context, w http.ResponseWriter, r *http.Request) {

	snippet := model.ReplySnippetFromJson(r.Body)

	if snippet == nil {
		c.SetInvalidParam("missing")
		return
	}

	if len(snippet.Trigger) == 0 {
		c.SetInvalidParam("Trigger")
		return
	}

	if len(snippet.AutoCompleteDesc) == 0 {
		c.SetInvalidParam("AutoCompleteDesc")
		return
	}

	var rSnippet *model.ReplySnippet
	var err *model.AppError

	rSnippet, err = c.App.CreateReplySnippet(snippet)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.ToJson()))
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(rSnippet.ToJson()))
}

func updateSnippet(c *Context, w http.ResponseWriter, r *http.Request) {
	snippetId := c.Params.SnippetId

	snippet := model.ReplySnippetFromJson(r.Body)
	if snippet == nil {
		c.SetInvalidParam("something")
		return
	}
	snippet.Id = snippetId

	var rSnippet *model.ReplySnippet
	var err *model.AppError

	rSnippet, err = c.App.UpdateReplySnippet(snippet)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.ToJson()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(rSnippet.ToJson()))
}

func getFanpage(c *Context, w http.ResponseWriter, r *http.Request) {
	pageId := c.Params.PageId
	if len(pageId) > 0 {
		if p, err := c.App.GetFanpageByPageId(pageId); err != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(err.ToJson()))
			return
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(p.ToJson()))
		}
	}
}

// khởi tạo pages
func initializeFanpages(c *Context, w http.ResponseWriter, r *http.Request) {
	if result := <-c.App.Srv.Store.User().GetById(c.App.Session.UserId); result.Err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(result.Err.ToJson()))
		return
	} else {
		var lpi *model.LoadPagesInput
		x, _ := ioutil.ReadAll(r.Body)
		json.Unmarshal([]byte(x), &lpi)

		user := result.Data.(*model.User)

		// Validation trước khi khởi tạo
		validationResult := c.App.ValidationPagesBeforeInit(lpi, c.App.Session.UserId)
		successLpi := model.PagesInitValidationToLPi(validationResult)
		if len(successLpi.PageIds) > 0 {
			pages, graphErr, appErr := c.App.GraphFanpages(user.FacebookToken)
			if graphErr != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(facebookgraph.FacebookErrorToJson(graphErr)))
				return
			} else if appErr != nil {
				c.Err = appErr
				return
			}

			// cập nhật status = queued để lock tất cả các pages, không cho phép khởi tạo nếu page đã được queued trước đó
			if error, _ := c.App.UpdatePagesStatus(successLpi, model.PAGE_STATUS_QUEUED, c.App.Session.UserId); error != nil {
				mlog.Error("Không thể cập nhật trạng thái pages "+model.PAGE_STATUS_QUEUED)
			}

			for _, pageId := range successLpi.PageIds {

				// recheck to ensure page never initialized before
				if thisPage := <-c.App.Srv.Store.Fanpage().GetFanpageByPageID( pageId ); thisPage.Err != nil {
					// page chưa được thêm vào db
				} else {
					page := thisPage.Data.(*model.Fanpage)
					if page.Status == model.PAGE_STATUS_INITIALIZING ||
						page.Status == model.PAGE_STATUS_INITIALIZED ||
						page.Status == model.PAGE_STATUS_ERROR ||
						page.Status == model.PAGE_STATUS_BLOCKED {

						break
					}
				}

				pageToken, disabled, err, aerr := c.App.GraphPageAndInit2(pageId, user, pages)

				if disabled {
					break
				}

				if err != nil {
					mlog.Error(err.Error.Message)
					if error, _ := c.App.UpdatePageStatus(pageId, model.PAGE_STATUS_ERROR, c.App.Session.UserId); error != nil {
						mlog.Error(pageId+": Không thể cập nhật trạng thái page "+model.PAGE_STATUS_ERROR)
					}

					break
					//w.WriteHeader(http.StatusBadRequest)
					//w.Write([]byte(facebookgraph.FacebookErrorToJson(err)))
					//return
				}
				if aerr != nil {
					mlog.Error(aerr.Error())
					if error, _ := c.App.UpdatePageStatus(pageId, model.PAGE_STATUS_ERROR, c.App.Session.UserId); error != nil {
						mlog.Error(pageId+": Không thể cập nhật trạng thái page "+model.PAGE_STATUS_ERROR)
					}
					break
				}

				// KHỞI TẠO PAGE
				// 1: Khởi tạo kết quả init page
				var pageInitResult *model.FanpageInitResult
				if rs, error := c.App.CreateFanpageInitResult(&model.FanpageInitResult{
					PageId: pageId,
					Creator: c.App.Session.UserId,
				}); error != nil {
					mlog.Error(error.Message)
				} else {
					pageInitResult = rs
				}

				// TODO: cập nhật trạng thái page => initializing, emit socket thông báo trạng thái page
				if error, _ := c.App.UpdatePageStatus(pageId, model.PAGE_STATUS_INITIALIZING, c.App.Session.UserId); error != nil {
					mlog.Error(pageId+": Không thể cập nhật trạng thái page "+model.PAGE_STATUS_INITIALIZING)
				}

				start := time.Now()

				// chạy 2 tiến trình song song: Khởi tạo messages và khởi tạo comments
				doneMessages := make(chan bool)
				doneComments := make(chan bool)

				go func() {
					// Graph messages
					fbErr, AppErr := c.App.GraphMessagesAndInit(pageId, pageToken, pageInitResult)
					if fbErr != nil {
						mlog.Error(fbErr.Error.Message)
						if error, _ := c.App.UpdatePageStatus(pageId, model.PAGE_STATUS_ERROR, c.App.Session.UserId); error != nil {
							mlog.Error(pageId+": Không thể cập nhật trạng thái page "+model.PAGE_STATUS_ERROR)
						}
					}

					if AppErr != nil {
						mlog.Error(AppErr.Error())
						if error, _ := c.App.UpdatePageStatus(pageId, model.PAGE_STATUS_ERROR, c.App.Session.UserId); error != nil {
							mlog.Error(pageId+": Không thể cập nhật trạng thái page "+model.PAGE_STATUS_ERROR)
						}
					}

					if fbErr == nil && AppErr == nil {
						doneMessages <- true
					}

					close(doneMessages)
				}()

				go func() {
					// Graph comments
					nextLink := "next"
					var fbErr2 *facebookgraph.FacebookError
					var appErr *model.AppError
					for len(nextLink) > 0 {
						nextLink = ""
						nextLink, fbErr2, appErr = c.App.GraphPagePostsAndInit(pageId, pageToken, nextLink, pageInitResult)
						if fbErr2 != nil {
							mlog.Error(fbErr2.Error.Message)
							if error, _ := c.App.UpdatePageStatus(pageId, model.PAGE_STATUS_ERROR, c.App.Session.UserId); error != nil {
								mlog.Error(pageId+": Không thể cập nhật trạng thái page "+model.PAGE_STATUS_ERROR)
							}
							break
						}
						if appErr != nil {
							mlog.Error(appErr.Error())
							if error, _ := c.App.UpdatePageStatus(pageId, model.PAGE_STATUS_ERROR, c.App.Session.UserId); error != nil {
								mlog.Error(pageId+": Không thể cập nhật trạng thái page "+model.PAGE_STATUS_ERROR)
							}
						}
					}

					if fbErr2 == nil && appErr == nil {
						doneComments <- true
					}

					close(doneComments)
				}()

				finishedMessages := <- doneMessages
				finishedComments := <- doneComments

				if finishedMessages && finishedComments {
					// TODO: cập nhật trạng thái page => initialized, emit socket thông báo trạng thái page
					if error, _ := c.App.UpdatePageStatus(pageId, model.PAGE_STATUS_INITIALIZED, c.App.Session.UserId); error != nil {
						mlog.Error(pageId+": Không thể cập nhật trạng thái page "+model.PAGE_STATUS_INITIALIZED)
					}

					mlog.Info("Khởi tạo Page thành công " + pageId + " sau " + time.Since(start).String() + " giây")
				} else {
					// TODO: lỗi khởi tạo page
					mlog.Error("Khởi tạo không thành công: " + pageId)
				}
			}
		}

		ReturnStatusOK(w)
	}
}

func getUserFanpages(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireUserId()
	if c.Err != nil {
		return
	}

	if c.App.Session.UserId != c.Params.UserId && !c.App.SessionHasPermissionTo(c.App.Session, model.PERMISSION_MANAGE_SYSTEM) {
		c.SetPermissionError(model.PERMISSION_MANAGE_SYSTEM)
		return
	}

	fanpages, err := c.App.GetFanpagesForUser(c.Params.UserId)
	if err != nil {
		c.Err = err
		return
	}
	//c.App.SanitizeFanpages(c.App.Session, fanpages)
	w.Write([]byte(model.FanpagesIncludeInitResultToJson(fanpages)))
}

// query: ?statuses=ready,waiting,initializing,initialized,error,hidden,deleted
func getFanpages(c *Context, w http.ResponseWriter, r *http.Request) {
	//var user *model.User
	if result := <-c.App.Srv.Store.User().GetById(c.App.Session.UserId); result.Err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(result.Err.ToJson()))
		return
	} else {
		//user = result.Data.(*model.User)
		// danh sách page có trong DB
		dbPages, dbErr := c.App.GetFanpagesForUser(c.App.Session.UserId)
		//fmt.Println(dbPages)
		if dbErr != nil {
			// lỗi ở đây là không tìm thấy page nào đã được thêm vào db
			// nếu query != nil thì mới return, ngược lại sẽ tiếp tục để trả về tất cả pages của user
			// TODO: check query và return kết quả theo query
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(dbErr.ToJson()))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(model.FanpagesIncludeInitResultToJson(dbPages)))
		return

		//query := r.URL.Query()
		//if len(query.Get("status")) > 0 {
		//	// tạm thời trả về tất cả page có trong DB, không graph các fanpage
		//	w.WriteHeader(http.StatusOK)
		//	w.Write([]byte(model.FanpagesIncludeInitResultToJson(dbPages)))
		//	return
		//}

		//// danh sách tất cả pages, load từ facbook
		//fbPages, fbErr, _ := c.App.GraphFanpages(user.FacebookToken)
		//if fbErr != nil {
		//	w.WriteHeader(http.StatusBadRequest)
		//	w.Write([]byte(facebookgraph.FacebookErrorToJson(fbErr)))
		//	return
		//}
		//
		//var result []*model.Fanpage
		//for _, fPage := range fbPages {
		//	pageItem := model.Fanpage{PageId: fPage.Id, Category: fPage.Category, Name: fPage.Name}
		//	for _, p := range dbPages {
		//		if p.Data["page_id"] == fPage.Id {
		//			pageItem.Id = p.Data.Id
		//			pageItem.Status = p.Data.Status
		//		}
		//	}
		//	result = append(result, &pageItem)
		//}
		//
		//w.WriteHeader(http.StatusOK)
		//w.Write([]byte(model.FanpageListToJson(result)))
	}
}

func createFanpage(c *Context, w http.ResponseWriter, r *http.Request) {
	// Tất cả hàm API đều được wrap trong APISessionRiquired
	// do đó handle sẽ xử lý check xem có authorization trong header hay không trước
	// sau đó mới gọi hàm này. Vì thế ở đây ta không cần check token nữa
	fanpage := model.FanpageFromJson(r.Body)
	if fanpage == nil {
		c.SetInvalidParam("fanpage")
		return
	}

	var rfanpage *model.Fanpage
	var err *model.AppError

	rfanpage, err = c.App.CreateFanpageWithUser(fanpage, c.App.Session.UserId)

	if err != nil {
		c.Err = err
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(rfanpage.ToJson()))
}
