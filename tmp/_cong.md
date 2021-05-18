# Mục lục

1. [import trong Golang](#markdown-header-import)
2. [export trong Golang](#markdown-header-export)
3. [hàm init()](#markdown-header-ham-init)

## import

**Chú ý rằng các câu lệnh `import` sẽ phải khai báo đầy đủ đường dẫn của file, tính từ src trở đi**
Ví dụ: 

Với `nodejs` thường khai báo import:
```import * from './commands';```

Thì trong dự án này sẽ sử dụng cấu trúc import như sau:
```import "bitbucket.org/enesyteam/papo/cmd/commands"```

Chỗ này được hiểu là khi biên dịch, trình biên dịch của GO sẽ thêm `GOPATH/src` vào trước `bitbucket.org/enesyteam/papo/cmd/commands` 

## export

Golang không có export, để export một hàm, bạn chỉ cần gõ in hoa chữ cái đầu của tên hàm là được

Ví dụ:
```
func Hello() {

}
```

## Hàm init()


##
```
func updateUserStatus(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireUserId()
	if c.Err != nil {
		return
	}

	status := model.StatusFromJson(r.Body)
	if status == nil {
		c.SetInvalidParam("status")
		return
	}

	// The user being updated in the payload must be the same one as indicated in the URL.
	if status.UserId != c.Params.UserId {
		c.SetInvalidParam("user_id")
		return
	}

	if !c.App.SessionHasPermissionToUser(c.Session, c.Params.UserId) {
		c.SetPermissionError(model.PERMISSION_EDIT_OTHER_USERS)
		return
	}

	currentStatus, err := c.App.GetStatus(c.Params.UserId)
	if err == nil && currentStatus.Status == model.STATUS_OUT_OF_OFFICE && status.Status != model.STATUS_OUT_OF_OFFICE {
		c.App.DisableAutoResponder(c.Params.UserId, c.IsSystemAdmin())
	}

	switch status.Status {
	case "online":
		c.App.SetStatusOnline(c.Params.UserId, true)
	case "offline":
		c.App.SetStatusOffline(c.Params.UserId, true)
	case "away":
		c.App.SetStatusAwayIfNeeded(c.Params.UserId, true)
	case "dnd":
		c.App.SetStatusDoNotDisturb(c.Params.UserId)
	default:
		c.SetInvalidParam("status")
		return
	}

	getUserStatus(c, w, r)
}
```

```
// Notify if you're not looking in the right channel or when
        // the window itself is not active
        const activeChannel = getCurrentChannel(state);
        const channelId = channel ? channel.id : null;
        const notify = (activeChannel && activeChannel.id !== channelId) || !state.views.browser.focused;

        if (notify) {
            Utils.notifyMe(title, body, channel, teamId, !sound);

            //Don't add extra sounds on native desktop clients
            if (sound && !isWindowsApp() && !isMacApp() && !isMobileApp()) {
                Utils.ding();
            }
        }
        ```

