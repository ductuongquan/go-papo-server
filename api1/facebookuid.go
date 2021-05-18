package api1

import (
	"bitbucket.org/enesyteam/papo-server/model"
	"net/http"
)

func (api *API) InitFacebookUid() {
	api.BaseRoutes.FacebookUsers.Handle("/ids", api.ApiSessionRequired(getUsersByIds)).Methods("POST")
}

func getUsersByIds(c *Context, w http.ResponseWriter, r *http.Request) {
	userIds := model.ArrayFromJson(r.Body)

	if len(userIds) == 0 {
		c.SetInvalidParam("user_ids")
		return
	}

	// No permission check required

	users, err := c.App.GetFacebookUsersByIds(userIds, c.IsSystemAdmin())
	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(model.FacebookUserListToJson(users)))
}