// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"bitbucket.org/enesyteam/papo-server/app"
	"bitbucket.org/enesyteam/papo-server/model"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"strings"
)

var CommandCmd = &cobra.Command{
	Use:   "command",
	Short: "Management of slash commands",
}

var CommandCreateCmd = &cobra.Command{
	Use:     "create [team]",
	Short:   "Create a custom slash command",
	Long:    `Create a custom slash command for the specified team.`,
	Args:    cobra.MinimumNArgs(1),
	Example: `  command create myteam --title MyCommand --description "My Command Description" --trigger-word mycommand --url http://localhost:8000/my-slash-handler --creator myusername --response-username my-bot-username --icon http://localhost:8000/my-slash-handler-bot-icon.png --autocomplete --post`,
	RunE:    createCommandCmdF,
}

var CommandMoveCmd = &cobra.Command{
	Use:     "move",
	Short:   "Move a slash command to a different team",
	Long:    `Move a slash command to a different team. Commands can be specified by [team]:[command-trigger-word]. ie. myteam:trigger or by command ID.`,
	Example: `  command move newteam oldteam:command`,
	RunE:    moveCommandCmdF,
}

var CommandListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all commands on specified teams.",
	Long:    `List all commands on specified teams.`,
	Example: ` command list myteam`,
	RunE:    listCommandCmdF,
}

var CommandDeleteCmd = &cobra.Command{
	Use:     "delete",
	Short:   "Delete a slash command",
	Long:    `Delete a slash command. Commands can be specified by command ID.`,
	Example: `  command delete commandID`,
	Args:    cobra.MinimumNArgs(1),
	RunE:    deleteCommandCmdF,
}

func init() {
	CommandCreateCmd.Flags().String("title", "", "Command Title")
	CommandCreateCmd.Flags().String("description", "", "Command Description")
	CommandCreateCmd.Flags().String("trigger-word", "", "Command Trigger Word (required)")
	CommandCreateCmd.MarkFlagRequired("trigger-word")
	CommandCreateCmd.Flags().String("url", "", "Command Callback URL (required)")
	CommandCreateCmd.MarkFlagRequired("url")
	CommandCreateCmd.Flags().String("creator", "", "Command Creator's Username (required)")
	CommandCreateCmd.MarkFlagRequired("creator")
	CommandCreateCmd.Flags().String("response-username", "", "Command Response Username")
	CommandCreateCmd.Flags().String("icon", "", "Command Icon URL")
	CommandCreateCmd.Flags().Bool("autocomplete", false, "Show Command in autocomplete list")
	CommandCreateCmd.Flags().String("autocompleteDesc", "", "Short Command Description for autocomplete list")
	CommandCreateCmd.Flags().String("autocompleteHint", "", "Command Arguments displayed as help in autocomplete list")
	CommandCreateCmd.Flags().Bool("post", false, "Use POST method for Callback URL")

	CommandCmd.AddCommand(
		CommandCreateCmd,
		CommandMoveCmd,
		CommandListCmd,
		CommandDeleteCmd,
	)
	RootCmd.AddCommand(CommandCmd)
}

func createCommandCmdF(command *cobra.Command, args []string) error {
	a, err := InitDBCommandContextCobra(command)
	if err != nil {
		return err
	}
	defer a.Shutdown()

	team := getTeamFromTeamArg(a, args[0])
	if team == nil {
		return errors.New("unable to find team '" + args[0] + "'")
	}

	title, _ := command.Flags().GetString("title")
	description, _ := command.Flags().GetString("description")
	trigger, _ := command.Flags().GetString("trigger-word")

	if strings.HasPrefix(trigger, "/") {
		return errors.New("a trigger word cannot begin with a /")
	}
	if strings.Contains(trigger, " ") {
		return errors.New("a trigger word must not contain spaces")
	}

	url, _ := command.Flags().GetString("url")
	creator, _ := command.Flags().GetString("creator")
	user := getUserFromUserArg(a, creator)
	if user == nil {
		return errors.New("unable to find user '" + creator + "'")
	}
	responseUsername, _ := command.Flags().GetString("response-username")
	icon, _ := command.Flags().GetString("icon")
	autocomplete, _ := command.Flags().GetBool("autocomplete")
	autocompleteDesc, _ := command.Flags().GetString("autocompleteDesc")
	autocompleteHint, _ := command.Flags().GetString("autocompleteHint")
	post, errp := command.Flags().GetBool("post")
	method := "P"
	if errp != nil || post == false {
		method = "G"
	}

	newCommand := &model.Command{
		CreatorId:        user.Id,
		TeamId:           team.Id,
		Trigger:          trigger,
		Method:           method,
		Username:         responseUsername,
		IconURL:          icon,
		AutoComplete:     autocomplete,
		AutoCompleteDesc: autocompleteDesc,
		AutoCompleteHint: autocompleteHint,
		DisplayName:      title,
		Description:      description,
		URL:              url,
	}

	if _, err := a.CreateCommand(newCommand); err != nil {
		return errors.New("unable to create command '" + newCommand.Trigger + "'. " + err.Error())
	}
	CommandPrettyPrintln("created command '" + newCommand.Trigger + "'")

	return nil
}

func moveCommandCmdF(command *cobra.Command, args []string) error {
	a, err := InitDBCommandContextCobra(command)
	if err != nil {
		return err
	}
	defer a.Shutdown()

	if len(args) < 2 {
		return errors.New("Enter the destination team and at least one comamnd to move.")
	}

	team := getTeamFromTeamArg(a, args[0])
	if team == nil {
		return errors.New("Unable to find destination team '" + args[0] + "'")
	}

	commands := getCommandsFromCommandArgs(a, args[1:])
	CommandPrintErrorln(commands)
	for i, command := range commands {
		if command == nil {
			CommandPrintErrorln("Unable to find command '" + args[i+1] + "'")
			continue
		}
		if err := moveCommand(a, team, command); err != nil {
			CommandPrintErrorln("Unable to move command '" + command.Trigger + "' error: " + err.Error())
		} else {
			CommandPrettyPrintln("Moved command '" + command.Trigger + "'")
		}
	}

	return nil
}

func moveCommand(a *app.App, team *model.Team, command *model.Command) *model.AppError {
	return a.MoveCommand(team, command)
}

func listCommandCmdF(command *cobra.Command, args []string) error {
	a, err := InitDBCommandContextCobra(command)
	if err != nil {
		return err
	}
	defer a.Shutdown()

	var teams []*model.Team
	if len(args) < 1 {
		teamList, err := a.GetAllTeams()
		if err != nil {
			return err
		}
		teams = teamList
	} else {
		teams = getTeamsFromTeamArgs(a, args)
	}

	for i, team := range teams {
		if team == nil {
			CommandPrintErrorln("Unable to find team '" + args[i] + "'")
			continue
		}
		result := <-a.Srv.Store.Command().GetByTeam(team.Id)
		if result.Err != nil {
			CommandPrintErrorln("Unable to list commands for '" + args[i] + "'")
			continue
		}
		commands := result.Data.([]*model.Command)
		for _, command := range commands {
			commandListItem := fmt.Sprintf("%s: %s (team: %s)", command.Id, command.DisplayName, team.Name)
			CommandPrettyPrintln(commandListItem)
		}
	}
	return nil
}

func deleteCommandCmdF(command *cobra.Command, args []string) error {
	a, err := InitDBCommandContextCobra(command)
	if err != nil {
		return err
	}
	defer a.Shutdown()

	commandID := args[0]

	deleteErr := a.DeleteCommand(commandID)
	if deleteErr != nil {
		CommandPrintErrorln("Unable to delete command '" + commandID + "' error: " + deleteErr.Error())
		return deleteErr
	}
	CommandPrettyPrintln("Deleted command '" + commandID + "'")
	return nil
}