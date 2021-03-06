// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"bitbucket.org/enesyteam/papo-server/model"
	"bitbucket.org/enesyteam/papo-server/store"
	"bitbucket.org/enesyteam/papo-server/utils"
	"errors"
	goi18n "github.com/mattermost/go-i18n/i18n"
	"net/http"
	"net/url"
	"strings"
)

type CommandProvider interface {
	GetTrigger() string
	GetCommand(a *App, T goi18n.TranslateFunc) *model.Command
	DoCommand(a *App, args *model.CommandArgs, message string) *model.CommandResponse
}

var commandProviders = make(map[string]CommandProvider)

func RegisterCommandProvider(newProvider CommandProvider) {
	commandProviders[newProvider.GetTrigger()] = newProvider
}

func GetCommandProvider(name string) CommandProvider {
	provider, ok := commandProviders[name]
	if ok {
		return provider
	}

	return nil
}

// @openTracingParams teamId, skipSlackParsing
func (a *App) CreateCommandPost(post *model.Post, teamId string, response *model.CommandResponse, skipSlackParsing bool) (*model.Post, *model.AppError) {
	return nil, nil
	//if skipSlackParsing {
	//	post.Message = response.Text
	//} else {
	//	post.Message = model.ParseSlackLinksToMarkdown(response.Text)
	//}
	//
	//post.CreateAt = model.GetMillis()
	//
	//if strings.HasPrefix(post.Type, model.POST_SYSTEM_MESSAGE_PREFIX) {
	//	err := model.NewAppError("CreateCommandPost", "api.context.invalid_param.app_error", map[string]interface{}{"Name": "post.type"}, "", http.StatusBadRequest)
	//	return nil, err
	//}
	//
	//if response.Attachments != nil {
	//	model.ParseSlackAttachment(post, response.Attachments)
	//}
	//
	//if response.ResponseType == model.COMMAND_RESPONSE_TYPE_IN_CHANNEL {
	//	return a.CreatePostMissingChannel(post, true)
	//}
	//
	//if (response.ResponseType == "" || response.ResponseType == model.COMMAND_RESPONSE_TYPE_EPHEMERAL) && (response.Text != "" || response.Attachments != nil) {
	//	post.ParentId = ""
	//	a.SendEphemeralPost(post.UserId, post)
	//}
	//
	//return post, nil
}

// @openTracingParams teamId
// previous ListCommands now ListAutocompleteCommands
func (a *App) ListAutocompleteCommands(teamId string, T goi18n.TranslateFunc) ([]*model.Command, *model.AppError) {
	commands := make([]*model.Command, 0, 32)
	seen := make(map[string]bool)

	for _, cmd := range a.PluginCommandsForTeam(teamId) {
		if cmd.AutoComplete && !seen[cmd.Trigger] {
			seen[cmd.Trigger] = true
			commands = append(commands, cmd)
		}
	}

	if *a.Config().ServiceSettings.EnableCommands {
		teamCmds, err := a.Srv().Store.Command().GetByTeam(teamId)
		if err != nil {
			return nil, model.NewAppError("ListAutocompleteCommands", "app.command.listautocompletecommands.internal_error", nil, err.Error(), http.StatusInternalServerError)
		}

		for _, cmd := range teamCmds {
			if cmd.AutoComplete && !seen[cmd.Trigger] {
				cmd.Sanitize()
				seen[cmd.Trigger] = true
				commands = append(commands, cmd)
			}
		}
	}

	for _, value := range commandProviders {
		if cmd := value.GetCommand(a, T); cmd != nil {
			cpy := *cmd
			if cpy.AutoComplete && !seen[cpy.Trigger] {
				cpy.Sanitize()
				seen[cpy.Trigger] = true
				commands = append(commands, &cpy)
			}
		}
	}

	return commands, nil
}

func (a *App) ListTeamCommands(teamId string) ([]*model.Command, *model.AppError) {
	if !*a.Config().ServiceSettings.EnableCommands {
		return nil, model.NewAppError("ListTeamCommands", "api.command.disabled.app_error", nil, "", http.StatusNotImplemented)
	}

	teamCmds, err := a.Srv().Store.Command().GetByTeam(teamId)
	if err != nil {
		return nil, model.NewAppError("ListTeamCommands", "app.command.listteamcommands.internal_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return teamCmds, nil
}

func (a *App) ListAllCommands(teamId string, T goi18n.TranslateFunc) ([]*model.Command, *model.AppError) {
	commands := make([]*model.Command, 0, 32)
	seen := make(map[string]bool)
	for _, value := range commandProviders {
		if cmd := value.GetCommand(a, T); cmd != nil {
			cpy := *cmd
			if cpy.AutoComplete && !seen[cpy.Trigger] {
				cpy.Sanitize()
				seen[cpy.Trigger] = true
				commands = append(commands, &cpy)
			}
		}
	}

	for _, cmd := range a.PluginCommandsForTeam(teamId) {
		if !seen[cmd.Trigger] {
			seen[cmd.Trigger] = true
			commands = append(commands, cmd)
		}
	}

	if *a.Config().ServiceSettings.EnableCommands {
		teamCmds, err := a.Srv().Store.Command().GetByTeam(teamId)
		if err != nil {
			return nil, model.NewAppError("ListAllCommands", "app.command.listallcommands.internal_error", nil, err.Error(), http.StatusInternalServerError)
		}
		for _, cmd := range teamCmds {
			if !seen[cmd.Trigger] {
				cmd.Sanitize()
				seen[cmd.Trigger] = true
				commands = append(commands, cmd)
			}
		}
	}

	return commands, nil
}

// @openTracingParams args
func (a *App) ExecuteCommand(args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	return nil, nil
	//trigger := ""
	//message := ""
	//index := strings.IndexFunc(args.Command, unicode.IsSpace)
	//if index != -1 {
	//	trigger = args.Command[:index]
	//	message = args.Command[index+1:]
	//} else {
	//	trigger = args.Command
	//}
	//trigger = strings.ToLower(trigger)
	//if !strings.HasPrefix(trigger, "/") {
	//	return nil, model.NewAppError("command", "api.command.execute_command.format.app_error", map[string]interface{}{"Trigger": trigger}, "", http.StatusBadRequest)
	//}
	//trigger = strings.TrimPrefix(trigger, "/")
	//
	//clientTriggerId, triggerId, appErr := model.GenerateTriggerId(args.UserId, a.AsymmetricSigningKey())
	//if appErr != nil {
	//	mlog.Error("error occurred in generating trigger Id for a user ", mlog.Err(appErr))
	//}
	//
	//args.TriggerId = triggerId
	//
	//// Plugins can override built in and custom commands
	//cmd, response, appErr := a.tryExecutePluginCommand(args)
	//if appErr != nil {
	//	return nil, appErr
	//} else if cmd != nil && response != nil {
	//	response.TriggerId = clientTriggerId
	//	return a.HandleCommandResponse(cmd, args, response, true)
	//}
	//
	//// Custom commands can override built ins
	//cmd, response, appErr = a.tryExecuteCustomCommand(args, trigger, message)
	//if appErr != nil {
	//	return nil, appErr
	//} else if cmd != nil && response != nil {
	//	response.TriggerId = clientTriggerId
	//	return a.HandleCommandResponse(cmd, args, response, false)
	//}
	//
	//cmd, response = a.tryExecuteBuiltInCommand(args, trigger, message)
	//if cmd != nil && response != nil {
	//	return a.HandleCommandResponse(cmd, args, response, true)
	//}
	//
	//return nil, model.NewAppError("command", "api.command.execute_command.not_found.app_error", map[string]interface{}{"Trigger": trigger}, "", http.StatusNotFound)
}

// MentionsToTeamMembers returns all the @ mentions found in message that
// belong to users in the specified team, linking them to their users
func (a *App) MentionsToTeamMembers(message, teamId string) model.UserMentionMap {
	//type mentionMapItem struct {
	//	Name string
	//	Id   string
	//}
	//
	//possibleMentions := model.PossibleAtMentions(message)
	//mentionChan := make(chan *mentionMapItem, len(possibleMentions))
	//
	//var wg sync.WaitGroup
	//for _, mention := range possibleMentions {
	//	wg.Add(1)
	//	go func(mention string) {
	//		defer wg.Done()
	//		user, err := a.Srv().Store.User().GetByUsername(mention)
	//
	//		if err != nil && err.StatusCode != http.StatusNotFound {
	//			mlog.Warn("Failed to retrieve user @"+mention, mlog.Err(err))
	//			return
	//		}
	//
	//		// If it's a http.StatusNotFound error, check for usernames in substrings
	//		// without trailing punctuation
	//		if err != nil {
	//			trimmed, ok := model.TrimUsernameSpecialChar(mention)
	//			for ; ok; trimmed, ok = model.TrimUsernameSpecialChar(trimmed) {
	//				userFromTrimmed, userErr := a.Srv().Store.User().GetByUsername(trimmed)
	//				if userErr != nil && err.StatusCode != http.StatusNotFound {
	//					return
	//				}
	//
	//				if userErr != nil {
	//					continue
	//				}
	//
	//				_, err = a.GetTeamMember(teamId, userFromTrimmed.Id)
	//				if err != nil {
	//					// The user is not in the team, so we should ignore it
	//					return
	//				}
	//
	//				mentionChan <- &mentionMapItem{trimmed, userFromTrimmed.Id}
	//				return
	//			}
	//
	//			return
	//		}
	//
	//		_, err = a.GetTeamMember(teamId, user.Id)
	//		if err != nil {
	//			// The user is not in the team, so we should ignore it
	//			return
	//		}
	//
	//		mentionChan <- &mentionMapItem{mention, user.Id}
	//	}(mention)
	//}
	//
	//wg.Wait()
	//close(mentionChan)
	//
	//atMentionMap := make(model.UserMentionMap)
	//for mention := range mentionChan {
	//	atMentionMap[mention.Name] = mention.Id
	//}
	//
	//return atMentionMap
	return nil
}

// MentionsToPublicChannels returns all the mentions to public channels,
// linking them to their channels
func (a *App) MentionsToPublicChannels(message, teamId string) model.ChannelMentionMap {
	//type mentionMapItem struct {
	//	Name string
	//	Id   string
	//}
	//
	//channelMentions := model.ChannelMentions(message)
	//mentionChan := make(chan *mentionMapItem, len(channelMentions))
	//
	//var wg sync.WaitGroup
	//for _, channelName := range channelMentions {
	//	wg.Add(1)
	//	go func(channelName string) {
	//		defer wg.Done()
	//		channel, err := a.GetChannelByName(channelName, teamId, false)
	//		if err != nil {
	//			return
	//		}
	//
	//		if !channel.IsOpen() {
	//			return
	//		}
	//
	//		mentionChan <- &mentionMapItem{channelName, channel.Id}
	//	}(channelName)
	//}
	//
	//wg.Wait()
	//close(mentionChan)
	//
	//channelMentionMap := make(model.ChannelMentionMap)
	//for mention := range mentionChan {
	//	channelMentionMap[mention.Name] = mention.Id
	//}
	//
	//return channelMentionMap
	return nil
}

// tryExecuteBuiltInCommand attempts to run a built in command based on the given arguments. If no such command can be
// found, returns nil for all arguments.
func (a *App) tryExecuteBuiltInCommand(args *model.CommandArgs, trigger string, message string) (*model.Command, *model.CommandResponse) {
	provider := GetCommandProvider(trigger)
	if provider == nil {
		return nil, nil
	}

	cmd := provider.GetCommand(a, args.T)
	if cmd == nil {
		return nil, nil
	}

	return cmd, provider.DoCommand(a, args, message)
}

// tryExecuteCustomCommand attempts to run a custom command based on the given arguments. If no such command can be
// found, returns nil for all arguments.
//func (a *App) tryExecuteCustomCommand(args *model.CommandArgs, trigger string, message string) (*model.Command, *model.CommandResponse, *model.AppError) {
//	//// Handle custom commands
//	//if !*a.Config().ServiceSettings.EnableCommands {
//	//	return nil, nil, model.NewAppError("ExecuteCommand", "api.command.disabled.app_error", nil, "", http.StatusNotImplemented)
//	//}
//	//
//	//chanChan := make(chan store.StoreResult, 1)
//	//go func() {
//	//	channel, err := a.Srv().Store.Channel().Get(args.ChannelId, true)
//	//	chanChan <- store.StoreResult{Data: channel, NErr: err}
//	//	close(chanChan)
//	//}()
//	//
//	//teamChan := make(chan store.StoreResult, 1)
//	//go func() {
//	//	team, err := a.Srv().Store.Team().Get(args.TeamId)
//	//	teamChan <- store.StoreResult{Data: team, NErr: err}
//	//	close(teamChan)
//	//}()
//	//
//	//userChan := make(chan store.StoreResult, 1)
//	//go func() {
//	//	user, err := a.Srv().Store.User().Get(args.UserId)
//	//	userChan <- store.StoreResult{Data: user, Err: err}
//	//	close(userChan)
//	//}()
//	//
//	//teamCmds, err := a.Srv().Store.Command().GetByTeam(args.TeamId)
//	//if err != nil {
//	//	return nil, nil, model.NewAppError("tryExecuteCustomCommand", "app.command.tryexecutecustomcommand.internal_error", nil, err.Error(), http.StatusInternalServerError)
//	//}
//	//
//	//tr := <-teamChan
//	//if tr.NErr != nil {
//	//	var nfErr *store.ErrNotFound
//	//	switch {
//	//	case errors.As(tr.NErr, &nfErr):
//	//		return nil, nil, model.NewAppError("tryExecuteCustomCommand", "app.team.get.find.app_error", nil, nfErr.Error(), http.StatusNotFound)
//	//	default:
//	//		return nil, nil, model.NewAppError("tryExecuteCustomCommand", "app.team.get.finding.app_error", nil, tr.NErr.Error(), http.StatusInternalServerError)
//	//	}
//	//}
//	//team := tr.Data.(*model.Team)
//	//
//	//ur := <-userChan
//	//if ur.Err != nil {
//	//	return nil, nil, ur.Err
//	//}
//	//user := ur.Data.(*model.User)
//	//
//	//cr := <-chanChan
//	//if cr.NErr != nil {
//	//	var nfErr *store.ErrNotFound
//	//	switch {
//	//	case errors.As(cr.NErr, &nfErr):
//	//		return nil, nil, model.NewAppError("tryExecuteCustomCommand", "app.channel.get.existing.app_error", nil, nfErr.Error(), http.StatusNotFound)
//	//	default:
//	//		return nil, nil, model.NewAppError("tryExecuteCustomCommand", "app.channel.get.find.app_error", nil, cr.NErr.Error(), http.StatusInternalServerError)
//	//	}
//	//}
//	//channel := cr.Data.(*model.Channel)
//	//
//	//var cmd *model.Command
//	//
//	//for _, teamCmd := range teamCmds {
//	//	if trigger == teamCmd.Trigger {
//	//		cmd = teamCmd
//	//	}
//	//}
//	//
//	//if cmd == nil {
//	//	return nil, nil, nil
//	//}
//	//
//	//mlog.Debug("Executing command", mlog.String("command", trigger), mlog.String("user_id", args.UserId))
//	//
//	//p := url.Values{}
//	//p.Set("token", cmd.Token)
//	//
//	//p.Set("team_id", cmd.TeamId)
//	//p.Set("team_domain", team.Name)
//	//
//	//p.Set("channel_id", args.ChannelId)
//	//p.Set("channel_name", channel.Name)
//	//
//	//p.Set("user_id", args.UserId)
//	//p.Set("user_name", user.Username)
//	//
//	//p.Set("command", "/"+trigger)
//	//p.Set("text", message)
//	//
//	//p.Set("trigger_id", args.TriggerId)
//	//
//	//userMentionMap := a.MentionsToTeamMembers(message, team.Id)
//	//for key, values := range userMentionMap.ToURLValues() {
//	//	p[key] = values
//	//}
//	//
//	//channelMentionMap := a.MentionsToPublicChannels(message, team.Id)
//	//for key, values := range channelMentionMap.ToURLValues() {
//	//	p[key] = values
//	//}
//	//
//	//hook, appErr := a.CreateCommandWebhook(cmd.Id, args)
//	//if appErr != nil {
//	//	return cmd, nil, model.NewAppError("command", "api.command.execute_command.failed.app_error", map[string]interface{}{"Trigger": trigger}, appErr.Error(), http.StatusInternalServerError)
//	//}
//	//p.Set("response_url", args.SiteURL+"/hooks/commands/"+hook.Id)
//	//
//	//return a.DoCommandRequest(cmd, p)
//	return nil, nil
//}

func (a *App) DoCommandRequest(cmd *model.Command, p url.Values) (*model.Command, *model.CommandResponse, *model.AppError) {
	//// Prepare the request
	//var req *http.Request
	//var err error
	//if cmd.Method == model.COMMAND_METHOD_GET {
	//	req, err = http.NewRequest(http.MethodGet, cmd.URL, nil)
	//} else {
	//	req, err = http.NewRequest(http.MethodPost, cmd.URL, strings.NewReader(p.Encode()))
	//}
	//
	//if err != nil {
	//	return cmd, nil, model.NewAppError("command", "api.command.execute_command.failed.app_error", map[string]interface{}{"Trigger": cmd.Trigger}, err.Error(), http.StatusInternalServerError)
	//}
	//
	//if cmd.Method == model.COMMAND_METHOD_GET {
	//	if req.URL.RawQuery != "" {
	//		req.URL.RawQuery += "&"
	//	}
	//	req.URL.RawQuery += p.Encode()
	//}
	//
	//req.Header.Set("Accept", "application/json")
	//req.Header.Set("Authorization", "Token "+cmd.Token)
	//if cmd.Method == model.COMMAND_METHOD_POST {
	//	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	//}
	//
	//// Send the request
	//resp, err := a.HTTPService().MakeClient(false).Do(req)
	//if err != nil {
	//	return cmd, nil, model.NewAppError("command", "api.command.execute_command.failed.app_error", map[string]interface{}{"Trigger": cmd.Trigger}, err.Error(), http.StatusInternalServerError)
	//}
	//
	//defer resp.Body.Close()
	//
	//// Handle the response
	//body := io.LimitReader(resp.Body, MaxIntegrationResponseSize)
	//
	//if resp.StatusCode != http.StatusOK {
	//	// Ignore the error below because the resulting string will just be the empty string if bodyBytes is nil
	//	bodyBytes, _ := ioutil.ReadAll(body)
	//
	//	return cmd, nil, model.NewAppError("command", "api.command.execute_command.failed_resp.app_error", map[string]interface{}{"Trigger": cmd.Trigger, "Status": resp.Status}, string(bodyBytes), http.StatusInternalServerError)
	//}
	//
	//response, err := model.CommandResponseFromHTTPBody(resp.Header.Get("Content-Type"), body)
	//if err != nil {
	//	return cmd, nil, model.NewAppError("command", "api.command.execute_command.failed.app_error", map[string]interface{}{"Trigger": cmd.Trigger}, err.Error(), http.StatusInternalServerError)
	//} else if response == nil {
	//	return cmd, nil, model.NewAppError("command", "api.command.execute_command.failed_empty.app_error", map[string]interface{}{"Trigger": cmd.Trigger}, "", http.StatusInternalServerError)
	//}
	//
	//return cmd, response, nil
	return nil, nil, nil
}

func (a *App) HandleCommandResponse(command *model.Command, args *model.CommandArgs, response *model.CommandResponse, builtIn bool) (*model.CommandResponse, *model.AppError) {
	//trigger := ""
	//if len(args.Command) != 0 {
	//	parts := strings.Split(args.Command, " ")
	//	trigger = parts[0][1:]
	//	trigger = strings.ToLower(trigger)
	//}
	//
	//var lastError *model.AppError
	//_, err := a.HandleCommandResponsePost(command, args, response, builtIn)
	//
	//if err != nil {
	//	mlog.Error("error occurred in handling command response post", mlog.Err(err))
	//	lastError = err
	//}
	//
	//if response.ExtraResponses != nil {
	//	for _, resp := range response.ExtraResponses {
	//		_, err := a.HandleCommandResponsePost(command, args, resp, builtIn)
	//
	//		if err != nil {
	//			mlog.Error("error occurred in handling command response post", mlog.Err(err))
	//			lastError = err
	//		}
	//	}
	//}
	//
	//if lastError != nil {
	//	return response, model.NewAppError("command", "api.command.execute_command.create_post_failed.app_error", map[string]interface{}{"Trigger": trigger}, "", http.StatusInternalServerError)
	//}
	//
	//return response, nil
	return nil, nil
}

func (a *App) HandleCommandResponsePost(command *model.Command, args *model.CommandArgs, response *model.CommandResponse, builtIn bool) (*model.Post, *model.AppError) {
	//post := &model.Post{}
	//post.ChannelId = args.ChannelId
	//post.RootId = args.RootId
	//post.ParentId = args.ParentId
	//post.UserId = args.UserId
	//post.Type = response.Type
	//post.SetProps(response.Props)
	//
	//if len(response.ChannelId) != 0 {
	//	_, err := a.GetChannelMember(response.ChannelId, args.UserId)
	//	if err != nil {
	//		err = model.NewAppError("HandleCommandResponsePost", "api.command.command_post.forbidden.app_error", nil, err.Error(), http.StatusForbidden)
	//		return nil, err
	//	}
	//	post.ChannelId = response.ChannelId
	//}
	//
	//isBotPost := !builtIn
	//
	//if *a.Config().ServiceSettings.EnablePostUsernameOverride {
	//	if len(command.Username) != 0 {
	//		post.AddProp("override_username", command.Username)
	//		isBotPost = true
	//	} else if len(response.Username) != 0 {
	//		post.AddProp("override_username", response.Username)
	//		isBotPost = true
	//	}
	//}
	//
	//if *a.Config().ServiceSettings.EnablePostIconOverride {
	//	if len(command.IconURL) != 0 {
	//		post.AddProp("override_icon_url", command.IconURL)
	//		isBotPost = true
	//	} else if len(response.IconURL) != 0 {
	//		post.AddProp("override_icon_url", response.IconURL)
	//		isBotPost = true
	//	} else {
	//		post.AddProp("override_icon_url", "")
	//	}
	//}
	//
	//if isBotPost {
	//	post.AddProp("from_webhook", "true")
	//}
	//
	//// Process Slack text replacements if the response does not contain "skip_slack_parsing": true.
	//if !response.SkipSlackParsing {
	//	response.Text = a.ProcessSlackText(response.Text)
	//	response.Attachments = a.ProcessSlackAttachments(response.Attachments)
	//}
	//
	//if _, err := a.CreateCommandPost(post, args.TeamId, response, response.SkipSlackParsing); err != nil {
	//	return post, err
	//}
	//
	//return post, nil
	return nil, nil
}

func (a *App) CreateCommand(cmd *model.Command) (*model.Command, *model.AppError) {
	if !*a.Config().ServiceSettings.EnableCommands {
		return nil, model.NewAppError("CreateCommand", "api.command.disabled.app_error", nil, "", http.StatusNotImplemented)
	}

	return a.createCommand(cmd)
}

func (a *App) createCommand(cmd *model.Command) (*model.Command, *model.AppError) {
	cmd.Trigger = strings.ToLower(cmd.Trigger)

	teamCmds, err := a.Srv().Store.Command().GetByTeam(cmd.TeamId)
	if err != nil {
		return nil, model.NewAppError("CreateCommand", "app.command.createcommand.internal_error", nil, err.Error(), http.StatusInternalServerError)
	}

	for _, existingCommand := range teamCmds {
		if cmd.Trigger == existingCommand.Trigger {
			return nil, model.NewAppError("CreateCommand", "api.command.duplicate_trigger.app_error", nil, "", http.StatusBadRequest)
		}
	}

	for _, builtInProvider := range commandProviders {
		builtInCommand := builtInProvider.GetCommand(a, utils.T)
		if builtInCommand != nil && cmd.Trigger == builtInCommand.Trigger {
			return nil, model.NewAppError("CreateCommand", "api.command.duplicate_trigger.app_error", nil, "", http.StatusBadRequest)
		}
	}

	command, nErr := a.Srv().Store.Command().Save(cmd)
	if nErr != nil {
		var appErr *model.AppError
		switch {
		case errors.As(nErr, &appErr):
			return nil, appErr
		default:
			return nil, model.NewAppError("CreateCommand", "app.command.createcommand.internal_error", nil, nErr.Error(), http.StatusInternalServerError)
		}
	}

	return command, nil
}

func (a *App) GetCommand(commandId string) (*model.Command, *model.AppError) {
	if !*a.Config().ServiceSettings.EnableCommands {
		return nil, model.NewAppError("GetCommand", "api.command.disabled.app_error", nil, "", http.StatusNotImplemented)
	}

	command, err := a.Srv().Store.Command().Get(commandId)
	if err != nil {
		var nfErr *store.ErrNotFound
		switch {
		case errors.As(err, &nfErr):
			return nil, model.NewAppError("SqlCommandStore.Get", "store.sql_command.get.missing.app_error", map[string]interface{}{"command_id": commandId}, "", http.StatusNotFound)
		default:
			return nil, model.NewAppError("GetCommand", "app.command.getcommand.internal_error", nil, err.Error(), http.StatusInternalServerError)
		}
	}
	return command, nil
}

func (a *App) UpdateCommand(oldCmd, updatedCmd *model.Command) (*model.Command, *model.AppError) {
	if !*a.Config().ServiceSettings.EnableCommands {
		return nil, model.NewAppError("UpdateCommand", "api.command.disabled.app_error", nil, "", http.StatusNotImplemented)
	}

	updatedCmd.Trigger = strings.ToLower(updatedCmd.Trigger)
	updatedCmd.Id = oldCmd.Id
	updatedCmd.Token = oldCmd.Token
	updatedCmd.CreateAt = oldCmd.CreateAt
	updatedCmd.UpdateAt = model.GetMillis()
	updatedCmd.DeleteAt = oldCmd.DeleteAt
	updatedCmd.CreatorId = oldCmd.CreatorId
	updatedCmd.PluginId = oldCmd.PluginId
	updatedCmd.TeamId = oldCmd.TeamId

	command, err := a.Srv().Store.Command().Update(updatedCmd)
	if err != nil {
		var nfErr *store.ErrNotFound
		var appErr *model.AppError
		switch {
		case errors.As(err, &nfErr):
			return nil, model.NewAppError("SqlCommandStore.Update", "store.sql_command.update.missing.app_error", map[string]interface{}{"command_id": updatedCmd.Id}, "", http.StatusNotFound)
		case errors.As(err, &appErr):
			return nil, appErr
		default:
			return nil, model.NewAppError("UpdateCommand", "app.command.updatecommand.internal_error", nil, err.Error(), http.StatusInternalServerError)
		}
	}

	return command, nil
}

func (a *App) MoveCommand(team *model.Team, command *model.Command) *model.AppError {
	command.TeamId = team.Id

	_, err := a.Srv().Store.Command().Update(command)
	if err != nil {
		var nfErr *store.ErrNotFound
		var appErr *model.AppError
		switch {
		case errors.As(err, &nfErr):
			return model.NewAppError("SqlCommandStore.Update", "store.sql_command.update.missing.app_error", map[string]interface{}{"command_id": command.Id}, "", http.StatusNotFound)
		case errors.As(err, &appErr):
			return appErr
		default:
			return model.NewAppError("MoveCommand", "app.command.movecommand.internal_error", nil, err.Error(), http.StatusInternalServerError)
		}
	}

	return nil
}

func (a *App) RegenCommandToken(cmd *model.Command) (*model.Command, *model.AppError) {
	if !*a.Config().ServiceSettings.EnableCommands {
		return nil, model.NewAppError("RegenCommandToken", "api.command.disabled.app_error", nil, "", http.StatusNotImplemented)
	}

	cmd.Token = model.NewId()

	command, err := a.Srv().Store.Command().Update(cmd)
	if err != nil {
		var nfErr *store.ErrNotFound
		var appErr *model.AppError
		switch {
		case errors.As(err, &nfErr):
			return nil, model.NewAppError("SqlCommandStore.Update", "store.sql_command.update.missing.app_error", map[string]interface{}{"command_id": cmd.Id}, "", http.StatusNotFound)
		case errors.As(err, &appErr):
			return nil, appErr
		default:
			return nil, model.NewAppError("RegenCommandToken", "app.command.regencommandtoken.internal_error", nil, err.Error(), http.StatusInternalServerError)
		}
	}

	return command, nil
}

func (a *App) DeleteCommand(commandId string) *model.AppError {
	if !*a.Config().ServiceSettings.EnableCommands {
		return model.NewAppError("DeleteCommand", "api.command.disabled.app_error", nil, "", http.StatusNotImplemented)
	}

	err := a.Srv().Store.Command().Delete(commandId, model.GetMillis())
	if err != nil {
		return model.NewAppError("DeleteCommand", "app.command.deletecommand.internal_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}
