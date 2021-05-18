package api1

import (
	"bitbucket.org/enesyteam/papo-server/facebook_graph"
	"bitbucket.org/enesyteam/papo-server/model"
	"fmt"
	"net/http"
)

func (api *API) InitPost() {
	api.BaseRoutes.Fanpage.Handle("/posts", api.ApiSessionRequired(getPagePosts)).Methods("GET")
	api.BaseRoutes.Fanpage.Handle("/posts/{post_id:[A-Za-z0-9_-]+}", api.ApiSessionRequired(fetchPost)).Methods("GET")
}

func fetchPost(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequirePostId()
	if c.Err != nil {
		return
	}

	// check if client sent a query with page access token
	var pageAccessToken string
	query := r.URL.Query()
	if len(query.Get("page_access_token")) > 0 {
		pageAccessToken = query.Get("page_access_token")
	} else {
		fmt.Println("must get page access token")
	}

	fErr, aErr, post := c.App.FetchPost(c.Params.PostId, pageAccessToken)
	if aErr != nil {
		c.Err = aErr
		return
	}

	if fErr != nil {
		w.Write([]byte(facebookgraph.FacebookErrorToJson(fErr)))
		return
	}

	w.Write([]byte(post.ToJson()))
}

func getPagePosts(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequirePageId()
	if c.Err != nil {
		return
	}

	posts, err := c.App.GetPagePosts(c.Params.PageId)
	if err != nil {
		c.Err = err
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(model.FacebookPostResponseListToJson(posts)))
}
