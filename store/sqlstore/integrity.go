// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package sqlstore

import (
	"bitbucket.org/enesyteam/papo-server/mlog"
	"bitbucket.org/enesyteam/papo-server/model"

	sq "github.com/Masterminds/squirrel"
)

type relationalCheckConfig struct {
	parentName         string
	parentIdAttr       string
	childName          string
	childIdAttr        string
	canParentIdBeEmpty bool
	sortRecords        bool
	filter             interface{}
}

func getOrphanedRecords(ss *SqlSupplier, cfg relationalCheckConfig) ([]model.OrphanedRecord, error) {
	var records []model.OrphanedRecord

	sub := ss.getQueryBuilder().
		Select("TRUE").
		From(cfg.parentName + " AS PT").
		Prefix("NOT EXISTS (").
		Suffix(")").
		Where("PT.id = CT." + cfg.parentIdAttr)

	main := ss.getQueryBuilder().
		Select().
		Column("CT." + cfg.parentIdAttr + " AS ParentId").
		From(cfg.childName + " AS CT").
		Where(sub)

	if cfg.childIdAttr != "" {
		main = main.Column("CT." + cfg.childIdAttr + " AS ChildId")
	}

	if cfg.canParentIdBeEmpty {
		main = main.Where(sq.NotEq{"CT." + cfg.parentIdAttr: ""})
	}

	if cfg.filter != nil {
		main = main.Where(cfg.filter)
	}

	if cfg.sortRecords {
		main = main.OrderBy("CT." + cfg.parentIdAttr)
	}

	query, args, _ := main.ToSql()

	_, err := ss.GetMaster().Select(&records, query, args...)

	return records, err
}

func checkParentChildIntegrity(ss *SqlSupplier, config relationalCheckConfig) model.IntegrityCheckResult {
	var result model.IntegrityCheckResult
	var data model.RelationalIntegrityCheckData

	config.sortRecords = true
	data.Records, result.Err = getOrphanedRecords(ss, config)
	if result.Err != nil {
		mlog.Error(result.Err.Error())
		return result
	}
	data.ParentName = config.parentName
	data.ChildName = config.childName
	data.ParentIdAttr = config.parentIdAttr
	data.ChildIdAttr = config.childIdAttr
	result.Data = data

	return result
}

func checkChannelsCommandWebhooksIntegrity(ss *SqlSupplier) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, relationalCheckConfig{
		parentName:   "Channels",
		parentIdAttr: "ChannelId",
		childName:    "CommandWebhooks",
		childIdAttr:  "Id",
	})
}

func checkChannelsChannelMemberHistoryIntegrity(ss *SqlSupplier) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, relationalCheckConfig{
		parentName:   "Channels",
		parentIdAttr: "ChannelId",
		childName:    "ChannelMemberHistory",
		childIdAttr:  "",
	})
}

func checkChannelsChannelMembersIntegrity(ss *SqlSupplier) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, relationalCheckConfig{
		parentName:   "Channels",
		parentIdAttr: "ChannelId",
		childName:    "ChannelMembers",
		childIdAttr:  "",
	})
}

func checkChannelsIncomingWebhooksIntegrity(ss *SqlSupplier) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, relationalCheckConfig{
		parentName:   "Channels",
		parentIdAttr: "ChannelId",
		childName:    "IncomingWebhooks",
		childIdAttr:  "Id",
	})
}

func checkChannelsOutgoingWebhooksIntegrity(ss *SqlSupplier) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, relationalCheckConfig{
		parentName:   "Channels",
		parentIdAttr: "ChannelId",
		childName:    "OutgoingWebhooks",
		childIdAttr:  "Id",
	})
}

func checkChannelsPostsIntegrity(ss *SqlSupplier) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, relationalCheckConfig{
		parentName:   "Channels",
		parentIdAttr: "ChannelId",
		childName:    "Posts",
		childIdAttr:  "Id",
	})
}

func checkCommandsCommandWebhooksIntegrity(ss *SqlSupplier) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, relationalCheckConfig{
		parentName:   "Commands",
		parentIdAttr: "CommandId",
		childName:    "CommandWebhooks",
		childIdAttr:  "Id",
	})
}

func checkPostsFileInfoIntegrity(ss *SqlSupplier) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, relationalCheckConfig{
		parentName:   "Posts",
		parentIdAttr: "PostId",
		childName:    "FileInfo",
		childIdAttr:  "Id",
	})
}

func checkPostsPostsParentIdIntegrity(ss *SqlSupplier) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, relationalCheckConfig{
		parentName:         "Posts",
		parentIdAttr:       "ParentId",
		childName:          "Posts",
		childIdAttr:        "Id",
		canParentIdBeEmpty: true,
	})
}

func checkPostsPostsRootIdIntegrity(ss *SqlSupplier) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, relationalCheckConfig{
		parentName:         "Posts",
		parentIdAttr:       "RootId",
		childName:          "Posts",
		childIdAttr:        "Id",
		canParentIdBeEmpty: true,
	})
}

func checkPostsReactionsIntegrity(ss *SqlSupplier) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, relationalCheckConfig{
		parentName:   "Posts",
		parentIdAttr: "PostId",
		childName:    "Reactions",
		childIdAttr:  "",
	})
}

func checkSchemesChannelsIntegrity(ss *SqlSupplier) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, relationalCheckConfig{
		parentName:         "Schemes",
		parentIdAttr:       "SchemeId",
		childName:          "Channels",
		childIdAttr:        "Id",
		canParentIdBeEmpty: true,
	})
}

func checkSchemesTeamsIntegrity(ss *SqlSupplier) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, relationalCheckConfig{
		parentName:         "Schemes",
		parentIdAttr:       "SchemeId",
		childName:          "Teams",
		childIdAttr:        "Id",
		canParentIdBeEmpty: true,
	})
}

func checkSessionsAuditsIntegrity(ss *SqlSupplier) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, relationalCheckConfig{
		parentName:         "Sessions",
		parentIdAttr:       "SessionId",
		childName:          "Audits",
		childIdAttr:        "Id",
		canParentIdBeEmpty: true,
	})
}

func checkTeamsChannelsIntegrity(ss *SqlSupplier) model.IntegrityCheckResult {
	res1 := checkParentChildIntegrity(ss, relationalCheckConfig{
		parentName:   "Teams",
		parentIdAttr: "TeamId",
		childName:    "Channels",
		childIdAttr:  "Id",
		filter:       sq.NotEq{"CT.Type": []string{model.CHANNEL_DIRECT, model.CHANNEL_GROUP}},
	})
	res2 := checkParentChildIntegrity(ss, relationalCheckConfig{
		parentName:         "Teams",
		parentIdAttr:       "TeamId",
		childName:          "Channels",
		childIdAttr:        "Id",
		canParentIdBeEmpty: true,
		filter:             sq.Eq{"CT.Type": []string{model.CHANNEL_DIRECT, model.CHANNEL_GROUP}},
	})
	data1 := res1.Data.(model.RelationalIntegrityCheckData)
	data2 := res2.Data.(model.RelationalIntegrityCheckData)
	data1.Records = append(data1.Records, data2.Records...)
	res1.Data = data1
	return res1
}

func checkTeamsCommandsIntegrity(ss *SqlSupplier) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, relationalCheckConfig{
		parentName:   "Teams",
		parentIdAttr: "TeamId",
		childName:    "Commands",
		childIdAttr:  "Id",
	})
}

func checkTeamsIncomingWebhooksIntegrity(ss *SqlSupplier) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, relationalCheckConfig{
		parentName:   "Teams",
		parentIdAttr: "TeamId",
		childName:    "IncomingWebhooks",
		childIdAttr:  "Id",
	})
}

func checkTeamsOutgoingWebhooksIntegrity(ss *SqlSupplier) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, relationalCheckConfig{
		parentName:   "Teams",
		parentIdAttr: "TeamId",
		childName:    "OutgoingWebhooks",
		childIdAttr:  "Id",
	})
}

func checkTeamsTeamMembersIntegrity(ss *SqlSupplier) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, relationalCheckConfig{
		parentName:   "Teams",
		parentIdAttr: "TeamId",
		childName:    "TeamMembers",
		childIdAttr:  "",
	})
}

func checkUsersAuditsIntegrity(ss *SqlSupplier) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, relationalCheckConfig{
		parentName:         "Users",
		parentIdAttr:       "UserId",
		childName:          "Audits",
		childIdAttr:        "Id",
		canParentIdBeEmpty: true,
	})
}

func checkUsersCommandWebhooksIntegrity(ss *SqlSupplier) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, relationalCheckConfig{
		parentName:   "Users",
		parentIdAttr: "UserId",
		childName:    "CommandWebhooks",
		childIdAttr:  "Id",
	})
}

func checkUsersChannelMemberHistoryIntegrity(ss *SqlSupplier) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, relationalCheckConfig{
		parentName:   "Users",
		parentIdAttr: "UserId",
		childName:    "ChannelMemberHistory",
		childIdAttr:  "",
	})
}

func checkUsersChannelMembersIntegrity(ss *SqlSupplier) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, relationalCheckConfig{
		parentName:   "Users",
		parentIdAttr: "UserId",
		childName:    "ChannelMembers",
		childIdAttr:  "",
	})
}

func checkUsersChannelsIntegrity(ss *SqlSupplier) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, relationalCheckConfig{
		parentName:         "Users",
		parentIdAttr:       "CreatorId",
		childName:          "Channels",
		childIdAttr:        "Id",
		canParentIdBeEmpty: true,
	})
}

func checkUsersCommandsIntegrity(ss *SqlSupplier) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, relationalCheckConfig{
		parentName:   "Users",
		parentIdAttr: "CreatorId",
		childName:    "Commands",
		childIdAttr:  "Id",
	})
}

func checkUsersCompliancesIntegrity(ss *SqlSupplier) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, relationalCheckConfig{
		parentName:   "Users",
		parentIdAttr: "UserId",
		childName:    "Compliances",
		childIdAttr:  "Id",
	})
}

func checkUsersEmojiIntegrity(ss *SqlSupplier) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, relationalCheckConfig{
		parentName:   "Users",
		parentIdAttr: "CreatorId",
		childName:    "Emoji",
		childIdAttr:  "Id",
	})
}

func checkUsersFileInfoIntegrity(ss *SqlSupplier) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, relationalCheckConfig{
		parentName:   "Users",
		parentIdAttr: "CreatorId",
		childName:    "FileInfo",
		childIdAttr:  "Id",
	})
}

func checkUsersIncomingWebhooksIntegrity(ss *SqlSupplier) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, relationalCheckConfig{
		parentName:   "Users",
		parentIdAttr: "UserId",
		childName:    "IncomingWebhooks",
		childIdAttr:  "Id",
	})
}

func checkUsersOAuthAccessDataIntegrity(ss *SqlSupplier) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, relationalCheckConfig{
		parentName:   "Users",
		parentIdAttr: "UserId",
		childName:    "OAuthAccessData",
		childIdAttr:  "Token",
	})
}

func checkUsersOAuthAppsIntegrity(ss *SqlSupplier) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, relationalCheckConfig{
		parentName:   "Users",
		parentIdAttr: "CreatorId",
		childName:    "OAuthApps",
		childIdAttr:  "Id",
	})
}

func checkUsersOAuthAuthDataIntegrity(ss *SqlSupplier) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, relationalCheckConfig{
		parentName:   "Users",
		parentIdAttr: "UserId",
		childName:    "OAuthAuthData",
		childIdAttr:  "Code",
	})
}

func checkUsersOutgoingWebhooksIntegrity(ss *SqlSupplier) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, relationalCheckConfig{
		parentName:   "Users",
		parentIdAttr: "CreatorId",
		childName:    "OutgoingWebhooks",
		childIdAttr:  "Id",
	})
}

func checkUsersPostsIntegrity(ss *SqlSupplier) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, relationalCheckConfig{
		parentName:   "Users",
		parentIdAttr: "UserId",
		childName:    "Posts",
		childIdAttr:  "Id",
	})
}

func checkUsersPreferencesIntegrity(ss *SqlSupplier) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, relationalCheckConfig{
		parentName:   "Users",
		parentIdAttr: "UserId",
		childName:    "Preferences",
		childIdAttr:  "",
	})
}

func checkUsersReactionsIntegrity(ss *SqlSupplier) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, relationalCheckConfig{
		parentName:   "Users",
		parentIdAttr: "UserId",
		childName:    "Reactions",
		childIdAttr:  "",
	})
}

func checkUsersSessionsIntegrity(ss *SqlSupplier) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, relationalCheckConfig{
		parentName:   "Users",
		parentIdAttr: "UserId",
		childName:    "Sessions",
		childIdAttr:  "Id",
	})
}

func checkUsersStatusIntegrity(ss *SqlSupplier) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, relationalCheckConfig{
		parentName:   "Users",
		parentIdAttr: "UserId",
		childName:    "Status",
		childIdAttr:  "",
	})
}

func checkUsersTeamMembersIntegrity(ss *SqlSupplier) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, relationalCheckConfig{
		parentName:   "Users",
		parentIdAttr: "UserId",
		childName:    "TeamMembers",
		childIdAttr:  "",
	})
}

func checkUsersUserAccessTokensIntegrity(ss *SqlSupplier) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, relationalCheckConfig{
		parentName:   "Users",
		parentIdAttr: "UserId",
		childName:    "UserAccessTokens",
		childIdAttr:  "Id",
	})
}

func checkChannelsIntegrity(ss *SqlSupplier, results chan<- model.IntegrityCheckResult) {
	results <- checkChannelsCommandWebhooksIntegrity(ss)
	results <- checkChannelsChannelMemberHistoryIntegrity(ss)
	results <- checkChannelsChannelMembersIntegrity(ss)
	results <- checkChannelsIncomingWebhooksIntegrity(ss)
	results <- checkChannelsOutgoingWebhooksIntegrity(ss)
	results <- checkChannelsPostsIntegrity(ss)
}

func checkCommandsIntegrity(ss *SqlSupplier, results chan<- model.IntegrityCheckResult) {
	results <- checkCommandsCommandWebhooksIntegrity(ss)
}

func checkPostsIntegrity(ss *SqlSupplier, results chan<- model.IntegrityCheckResult) {
	results <- checkPostsFileInfoIntegrity(ss)
	results <- checkPostsPostsParentIdIntegrity(ss)
	results <- checkPostsPostsRootIdIntegrity(ss)
	results <- checkPostsReactionsIntegrity(ss)
}

func checkSchemesIntegrity(ss *SqlSupplier, results chan<- model.IntegrityCheckResult) {
	results <- checkSchemesChannelsIntegrity(ss)
	results <- checkSchemesTeamsIntegrity(ss)
}

func checkSessionsIntegrity(ss *SqlSupplier, results chan<- model.IntegrityCheckResult) {
	results <- checkSessionsAuditsIntegrity(ss)
}

func checkTeamsIntegrity(ss *SqlSupplier, results chan<- model.IntegrityCheckResult) {
	results <- checkTeamsChannelsIntegrity(ss)
	results <- checkTeamsCommandsIntegrity(ss)
	results <- checkTeamsIncomingWebhooksIntegrity(ss)
	results <- checkTeamsOutgoingWebhooksIntegrity(ss)
	results <- checkTeamsTeamMembersIntegrity(ss)
}

func checkUsersIntegrity(ss *SqlSupplier, results chan<- model.IntegrityCheckResult) {
	results <- checkUsersAuditsIntegrity(ss)
	results <- checkUsersCommandWebhooksIntegrity(ss)
	results <- checkUsersChannelMemberHistoryIntegrity(ss)
	results <- checkUsersChannelMembersIntegrity(ss)
	results <- checkUsersChannelsIntegrity(ss)
	results <- checkUsersCommandsIntegrity(ss)
	results <- checkUsersCompliancesIntegrity(ss)
	results <- checkUsersEmojiIntegrity(ss)
	results <- checkUsersFileInfoIntegrity(ss)
	results <- checkUsersIncomingWebhooksIntegrity(ss)
	results <- checkUsersOAuthAccessDataIntegrity(ss)
	results <- checkUsersOAuthAppsIntegrity(ss)
	results <- checkUsersOAuthAuthDataIntegrity(ss)
	results <- checkUsersOutgoingWebhooksIntegrity(ss)
	results <- checkUsersPostsIntegrity(ss)
	results <- checkUsersPreferencesIntegrity(ss)
	results <- checkUsersReactionsIntegrity(ss)
	results <- checkUsersSessionsIntegrity(ss)
	results <- checkUsersStatusIntegrity(ss)
	results <- checkUsersTeamMembersIntegrity(ss)
	results <- checkUsersUserAccessTokensIntegrity(ss)
}

func CheckRelationalIntegrity(ss *SqlSupplier, results chan<- model.IntegrityCheckResult) {
	mlog.Info("Starting relational integrity checks...")
	checkChannelsIntegrity(ss, results)
	checkCommandsIntegrity(ss, results)
	checkPostsIntegrity(ss, results)
	checkSchemesIntegrity(ss, results)
	checkSessionsIntegrity(ss, results)
	checkTeamsIntegrity(ss, results)
	checkUsersIntegrity(ss, results)
	mlog.Info("Done with relational integrity checks")
	close(results)
}
