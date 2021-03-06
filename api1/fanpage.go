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

	// cho ph??p t???o chi???n d???ch kh??ng ?????t t??n, h??? th???ng s??? t??? ?????t t??n cho chi???n d???ch
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

// kh???i t???o pages
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

		// Validation tr?????c khi kh???i t???o
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

			// c???p nh???t status = queued ????? lock t???t c??? c??c pages, kh??ng cho ph??p kh???i t???o n???u page ???? ???????c queued tr?????c ????
			if error, _ := c.App.UpdatePagesStatus(successLpi, model.PAGE_STATUS_QUEUED, c.App.Session.UserId); error != nil {
				mlog.Error("Kh??ng th??? c???p nh???t tr???ng th??i pages "+model.PAGE_STATUS_QUEUED)
			}

			for _, pageId := range successLpi.PageIds {

				// recheck to ensure page never initialized before
				if thisPage := <-c.App.Srv.Store.Fanpage().GetFanpageByPageID( pageId ); thisPage.Err != nil {
					// page ch??a ???????c th??m v??o db
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
						mlog.Error(pageId+": Kh??ng th??? c???p nh???t tr???ng th??i page "+model.PAGE_STATUS_ERROR)
					}

					break
					//w.WriteHeader(http.StatusBadRequest)
					//w.Write([]byte(facebookgraph.FacebookErrorToJson(err)))
					//return
				}
				if aerr != nil {
					mlog.Error(aerr.Error())
					if error, _ := c.App.UpdatePageStatus(pageId, model.PAGE_STATUS_ERROR, c.App.Session.UserId); error != nil {
						mlog.Error(pageId+": Kh??ng th??? c???p nh???t tr???ng th??i page "+model.PAGE_STATUS_ERROR)
					}
					break
				}

				// KH???I T???O PAGE
				// 1: Kh???i t???o k???t qu??? init page
				var pageInitResult *model.FanpageInitResult
				if rs, error := c.App.CreateFanpageInitResult(&model.FanpageInitResult{
					PageId: pageId,
					Creator: c.App.Session.UserId,
				}); error != nil {
					mlog.Error(error.Message)
				} else {
					pageInitResult = rs
				}

				// TODO: c???p nh???t tr???ng th??i page => initializing, emit socket th??ng b??o tr???ng th??i page
				if error, _ := c.App.UpdatePageStatus(pageId, model.PAGE_STATUS_INITIALIZING, c.App.Session.UserId); error != nil {
					mlog.Error(pageId+": Kh??ng th??? c???p nh???t tr???ng th??i page "+model.PAGE_STATUS_INITIALIZING)
				}

				start := time.Now()

				// ch???y 2 ti???n tr??nh song song: Kh???i t???o messages v?? kh???i t???o comments
				doneMessages := make(chan bool)
				doneComments := make(chan bool)

				go func() {
					// Graph messages
					fbErr, AppErr := c.App.GraphMessagesAndInit(pageId, pageToken, pageInitResult)
					if fbErr != nil {
						mlog.Error(fbErr.Error.Message)
						if error, _ := c.App.UpdatePageStatus(pageId, model.PAGE_STATUS_ERROR, c.App.Session.UserId); error != nil {
							mlog.Error(pageId+": Kh??ng th??? c???p nh???t tr???ng th??i page "+model.PAGE_STATUS_ERROR)
						}
					}

					if AppErr != nil {
						mlog.Error(AppErr.Error())
						if error, _ := c.App.UpdatePageStatus(pageId, model.PAGE_STATUS_ERROR, c.App.Session.UserId); error != nil {
							mlog.Error(pageId+": Kh??ng th??? c???p nh???t tr???ng th??i page "+model.PAGE_STATUS_ERROR)
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
								mlog.Error(pageId+": Kh??ng th??? c???p nh???t tr???ng th??i page "+model.PAGE_STATUS_ERROR)
							}
							break
						}
						if appErr != nil {
							mlog.Error(appErr.Error())
							if error, _ := c.App.UpdatePageStatus(pageId, model.PAGE_STATUS_ERROR, c.App.Session.UserId); error != nil {
								mlog.Error(pageId+": Kh??ng th??? c???p nh???t tr???ng th??i page "+model.PAGE_STATUS_ERROR)
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
					// TODO: c???p nh???t tr???ng th??i page => initialized, emit socket th??ng b??o tr???ng th??i page
					if error, _ := c.App.UpdatePageStatus(pageId, model.PAGE_STATUS_INITIALIZED, c.App.Session.UserId); error != nil {
						mlog.Error(pageId+": Kh??ng th??? c???p nh???t tr???ng th??i page "+model.PAGE_STATUS_INITIALIZED)
					}

					mlog.Info("Kh???i t???o Page th??nh c??ng " + pageId + " sau " + time.Since(start).String() + " gi??y")
				} else {
					// TODO: l???i kh???i t???o page
					mlog.Error("Kh???i t???o kh??ng th??nh c??ng: " + pageId)
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
		// danh s??ch page c?? trong DB
		dbPages, dbErr := c.App.GetFanpagesForUser(c.App.Session.UserId)
		//fmt.Println(dbPages)
		if dbErr != nil {
			// l???i ??? ????y l?? kh??ng t??m th???y page n??o ???? ???????c th??m v??o db
			// n???u query != nil th?? m???i return, ng?????c l???i s??? ti???p t???c ????? tr??? v??? t???t c??? pages c???a user
			// TODO: check query v?? return k???t qu??? theo query
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(dbErr.ToJson()))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(model.FanpagesIncludeInitResultToJson(dbPages)))
		return

		//query := r.URL.Query()
		//if len(query.Get("status")) > 0 {
		//	// t???m th???i tr??? v??? t???t c??? page c?? trong DB, kh??ng graph c??c fanpage
		//	w.WriteHeader(http.StatusOK)
		//	w.Write([]byte(model.FanpagesIncludeInitResultToJson(dbPages)))
		//	return
		//}

		//// danh s??ch t???t c??? pages, load t??? facbook
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
	// T???t c??? h??m API ?????u ???????c wrap trong APISessionRiquired
	// do ???? handle s??? x??? l?? check xem c?? authorization trong header hay kh??ng tr?????c
	// sau ???? m???i g???i h??m n??y. V?? th??? ??? ????y ta kh??ng c???n check token n???a
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
