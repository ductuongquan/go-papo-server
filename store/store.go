// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package store

import (
	"bitbucket.org/enesyteam/papo-server/facebook_graph"
	"bitbucket.org/enesyteam/papo-server/model"
	"context"
	"time"
)

type StoreResult struct {
	Data interface{}
	Err  *model.AppError

	// NErr a temporary field used by the new code for the AppError migration. This will later become Err when the entire store is migrated.
	NErr error
}

type Store interface {
	Audit() AuditStore
	ClusterDiscovery() ClusterDiscoveryStore
	Compliance() ComplianceStore
	Command() CommandStore
	Preference() PreferenceStore
	License() LicenseStore
	Team() TeamStore
	Token() TokenStore
	OAuth() OAuthStore // oauth for apps
	Reaction() ReactionStore
	Role() RoleStore
	Job() JobStore
	Scheme() SchemeStore
	System() SystemStore
	User() UserStore
	UserAccessToken() UserAccessTokenStore
	TermsOfService() TermsOfServiceStore
	Fanpage() FanpageStore
	FanpageInitResult() FanpageInitResultStore
	FileInfo() FileInfoStore
	PageReplySnippet() PageReplySnippetStore
	AutoMessageTask() AutoMessageTaskStore
	FacebookConversation() FacebookConversationStore

	Order() OrderStore

	PageTag() PageTagStore
	ConversationTag() ConversationTagStore
	ConversationNote() ConversationNoteStore
	FacebookPost() FacebookPostStore
	FacebookUid() FacebookUidStore

	Session() SessionStore
	Status() StatusStore
	Webhook() WebhookStore
	Close()
	DropAllTables()
	RecycleDBConnections(d time.Duration)
	GetCurrentSchemaVersion() string
	GetDbVersion() (string, error)
	TotalMasterDbConnections() int
	TotalReadDbConnections() int
	TotalSearchDbConnections() int
	CheckIntegrity() <-chan model.IntegrityCheckResult
	SetContext(context context.Context)
	Context() context.Context
	LinkMetadata() LinkMetadataStore

	Plugin() PluginStore
	UploadSession() UploadSessionStore
	UserTermsOfService() UserTermsOfServiceStore
}

type UserTermsOfServiceStore interface {
	GetByUser(userId string) (*model.UserTermsOfService, error)
	Save(userTermsOfService *model.UserTermsOfService) (*model.UserTermsOfService, error)
	Delete(userId, termsOfServiceId string) error
}

type UploadSessionStore interface {
	Save(session *model.UploadSession) (*model.UploadSession, error)
	Update(session *model.UploadSession) error
	Get(id string) (*model.UploadSession, error)
	GetForUser(userId string) ([]*model.UploadSession, error)
	Delete(id string) error
}

type PluginStore interface {
	SaveOrUpdate(keyVal *model.PluginKeyValue) (*model.PluginKeyValue, error)
	CompareAndSet(keyVal *model.PluginKeyValue, oldValue []byte) (bool, error)
	CompareAndDelete(keyVal *model.PluginKeyValue, oldValue []byte) (bool, error)
	SetWithOptions(pluginId string, key string, value []byte, options model.PluginKVSetOptions) (bool, error)
	Get(pluginId, key string) (*model.PluginKeyValue, error)
	Delete(pluginId, key string) error
	DeleteAllForPlugin(PluginId string) error
	DeleteAllExpired() error
	List(pluginId string, page, perPage int) ([]string, error)
}

type OrderStore interface {
	Save(order *model.Order) (*model.Order, error)
	GetOrders(limit, offset int) ([]*model.Order, error)
}

type LicenseStore interface {
	Save(license *model.LicenseRecord) (*model.LicenseRecord, error)
	Get(id string) (*model.LicenseRecord, error)
}

type FanpageInitResultStore interface {
	Save(job *model.FanpageInitResult) (*model.FanpageInitResult, error)
	UpdateConversationCount(r *model.FanpageInitResult, newCount int64) (*model.FanpageInitResult, error)
	UpdatePostCount(r *model.FanpageInitResult, newCount int64) (*model.FanpageInitResult, error)
	UpdateMessageCount(r *model.FanpageInitResult, newCount int64) (*model.FanpageInitResult, error)
	UpdateCommentCount(r *model.FanpageInitResult, newCount int64) (*model.FanpageInitResult, error)
	UpdateEndAt(r *model.FanpageInitResult, endAt int64) (*model.FanpageInitResult, error)
}

type JobStore interface {
	Save(job *model.Job) (*model.Job, error)
	UpdateOptimistically(job *model.Job, currentStatus string) (bool, error)
	UpdateStatus(id string, status string) (*model.Job, error)
	UpdateStatusOptimistically(id string, currentStatus string, newStatus string) (bool, error)
	Get(id string) (*model.Job, error)
	GetAllPage(offset int, limit int) ([]*model.Job, error)
	GetAllByType(jobType string) ([]*model.Job, error)
	GetAllByTypePage(jobType string, offset int, limit int) ([]*model.Job, error)
	GetAllByStatus(status string) ([]*model.Job, error)
	GetNewestJobByStatusAndType(status string, jobType string) (*model.Job, error)
	GetNewestJobByStatusesAndType(statuses []string, jobType string) (*model.Job, error)
	GetCountByStatusAndType(status string, jobType string) (int64, error)
	Delete(id string) (string, error)
}

type ComplianceStore interface {
	Save(compliance *model.Compliance) (*model.Compliance, error)
	Update(compliance *model.Compliance) (*model.Compliance, error)
	Get(id string) (*model.Compliance, error)
	GetAll(offset, limit int) (model.Compliances, error)
	//ComplianceExport(compliance *model.Compliance) ([]*model.CompliancePost, error)
	//MessageExport(after int64, limit int) ([]*model.MessageExport, error)
}

type ConversationTagStore interface {
	Save(tag *model.ConversationTag) (*model.ConversationTag, error)
	SaveOrRemove(tag *model.ConversationTag) (*model.ConversationTag, error)
	Get(id string) (*model.ConversationTag, error)
	GetConversationTags(conversationId string) ([]*model.ConversationTag, error)
	Delete(id string) (*model.ConversationTag, error)
}

type ConversationNoteStore interface {
	Save(tag *model.ConversationNote) (*model.ConversationNote, error)
	Get(id string) (*model.ConversationNote, error)
	GetConversationNotes(conversationId string) ([]*model.ConversationNote, error)
	Update(note *model.ConversationNote) (*model.ConversationNote, error)
	Delete(id string) (*model.ConversationNote, error)
}

type PageTagStore interface {
	Save(tag *model.PageTag) (*model.PageTag, error)
	Get(id string) (*model.PageTag, error)
	GetPageTags(pageId string) ([]*model.PageTag, error)
	Update(tag *model.PageTag) (*model.PageTag, error)
	Delete(id string) (*model.PageTag, error)
}

type FileInfoStore interface {
	Save(info *model.FileInfo) (*model.FileInfo, error)
	Upsert(info *model.FileInfo) (*model.FileInfo, error)
	Get(id string) (*model.FileInfo, error)
	GetByPath(path string) (*model.FileInfo, error)
	GetForUser(userId string) ([]*model.FileInfo, error)
	GetWithOptions(page, perPage int, opt *model.GetFileInfosOptions) ([]*model.FileInfo, error)
	GetForPage(pageId string, readFromMaster bool, allowFromCache bool, offset, limit int) ([]*model.FileInfo, error)
	AttachToMessage(fileId string, messageId string, creatorId string) (*model.FileInfo, error)
	InvalidateFileInfosForPageCache(pageId string)
	AttachToPost(fileId string, postId string, creatorId string) error
	DeleteForPost(postId string) (string, error)
	PermanentDelete(fileId string) error
	PermanentDeleteBatch(endTime int64, limit int64) (int64, error)
	PermanentDeleteByUser(userId string) (int64, error)
	ClearCaches()
}

type AutoMessageTaskStore interface {
	Save(messageTask *model.AutoMessageTask) (*model.AutoMessageTask, error)
	Update(messageTask *model.AutoMessageTask) (*model.AutoMessageTask, error)
	GetByPageId(pageId string) ([]*model.AutoMessageTask, error)
	Start(taskId string) (*model.AutoMessageTask, error)
	Pause(taskId string) (*model.AutoMessageTask, error)
	Delete(taskId string) (*model.AutoMessageTask, error)
}

type PageReplySnippetStore interface {
	Save(snippet *model.ReplySnippet) StoreChannel
	Update(snippet *model.ReplySnippet) StoreChannel
	GetByPageId(pageId string) StoreChannel
	CheckDuplicate(pageId string, trigger string) StoreChannel
}

type FacebookPostStore interface {
	Save(post *model.FacebookPost) StoreChannel
	Get(postId string) StoreChannel
	GetPagePosts(pageId string) StoreChannel
}

type FacebookConversationStore interface {
	Save(conversation *model.FacebookConversation) StoreChannel
	Get(conversationId string) StoreChannel
	GetMessage(messageId string) StoreChannel
	AddMessage(message *model.FacebookConversationMessage, shouldUpdateConversation bool, isFromPage bool) StoreChannel
	AddImage(image *model.FacebookAttachmentImage) StoreChannel
	GetMessagesByConversationId(conversationId string, offset, limit int) StoreChannel
	GetConversations(pageIds string, offset, limit int) StoreChannel
	GetConversationById(id string) StoreChannel
	Search(term string, pageIds string, offset, limit int) StoreChannel
	UpsertCommentConversation(conversation *model.FacebookConversation) StoreChannel
	UpdatePageScopeId(conversationId, pageScopeId string) StoreChannel
	UpdateLatestTime(conversationId string, time string, commentId string) StoreChannel
	UpdateSeen(id string, pageId string, userId string) StoreChannel
	UpdateUnSeen(id string, pageId string, userId string) StoreChannel
	GetPageConversationBySenderId(pageId, senderId, conversationType string) StoreChannel
	GetPageMessageByMid(pageId, mid string) StoreChannel
	AnalyticsConversationCountsByDay(pageId string) StoreChannel
	OverwriteMessage(message *model.FacebookConversationMessage) StoreChannel
	//GetConversationTypeComment(userId string, pageId string, postId string, commentId string) StoreChannel
	InsertConversationFromCommentIfNeed(parentId string, commentId string, pageId string, postId string, userId string, time string, message string) StoreChannel
	UpdateConversation(conversationId string, snippet string, isFromPage bool, updatedTime string, unreadCount int, lastUserMessageAt string) StoreChannel
	UpdateConversationUnread(conversationId string, isFromPage bool, unreadCount int, lastUserMessageAt string) StoreChannel
	GetFacebookAttachmentByIds(userIds []string, allowFromCache bool) StoreChannel
	UpdateMessageSent(messageId string) StoreChannel
	UpdateReadWatermark(conversationId, pageId string, timestamp int64) StoreChannel
	UpdateCommentByCommentId(commentId, newText string) StoreChannel
	DeleteCommentByCommentId(commentId, appScopedUserId string) StoreChannel
}

type FanpageStore interface {
	Save(fanpage *model.Fanpage) StoreChannel
	ValidatePagesBeforeInit(pageIds *model.LoadPagesInput) StoreChannel
	Update(newPage *model.Fanpage, oldPage *model.Fanpage) StoreChannel
	UpdateStatus(pageId string, status string) StoreChannel
	UpdatePagesStatus(pageIds *model.LoadPagesInput, status string) StoreChannel
	Get(fanpageId string) StoreChannel
	GetMember(teamId string, userId string) StoreChannel
	GetMemberByPageId(pageId string, userId string) StoreChannel
	SaveFanPageMember(member *model.FanpageMember) StoreChannel
	GetFanpagesByUserId(userId string) StoreChannel
	GetFanpageByPageID(pageId string) StoreChannel
	GetOneFanPageMember(fanpageId string) StoreChannel
	GetAllPageMembersForUser(userId string, allowFromCache bool, includeDeleted bool) StoreChannel
	//GetAllFanpages(offset int, limit int) StoreChannel
	//Delete(fanpageId string) StoreChannel
	//UpdateStatus(newStatus string) StoreChannel
	UpdateLastViewedAt(pageIds []string, userId string) StoreChannel
}

type FacebookUidStore interface {
	Get(id string) StoreChannel
	UpsertFromMap(data map[string]interface{}) StoreChannel
	UpsertFromFbUser(fbUser facebookgraph.FacebookUser) StoreChannel
	GetByIds(userIds []string, allowFromCache bool) StoreChannel
	UpdatePageId(id, pageId string) StoreChannel
	UpdatePageScopeId(id, pageScopeId string) StoreChannel
}

type PreferenceStore interface {
	//Save(preferences *model.Preferences) StoreChannel
	//Get(userId string, category string, name string) StoreChannel
	//GetCategory(userId string, category string) StoreChannel
	//GetAll(userId string) StoreChannel
	//Delete(userId, category, name string) StoreChannel
	//DeleteCategory(userId string, category string) StoreChannel
	//DeleteCategoryAndName(category string, name string) StoreChannel
	//PermanentDeleteByUser(userId string) StoreChannel
	//IsFeatureEnabled(feature, userId string) StoreChannel
	//CleanupFlagsBatch(limit int64) StoreChannel

	Save(preferences *model.Preferences) error
	GetCategory(userId string, category string) (model.Preferences, error)
	Get(userId string, category string, name string) (*model.Preference, error)
	GetAll(userId string) (model.Preferences, error)
	Delete(userId, category, name string) error
	DeleteCategory(userId string, category string) error
	DeleteCategoryAndName(category string, name string) error
	PermanentDeleteByUser(userId string) error
	//IsFeatureEnabled(feature, userId string) bool // TODO: Need this ? (from old version)
	CleanupFlagsBatch(limit int64) (int64, error)
}

type ClusterDiscoveryStore interface {
	//Save(discovery *model.ClusterDiscovery) StoreChannel
	//Delete(discovery *model.ClusterDiscovery) StoreChannel
	//Exists(discovery *model.ClusterDiscovery) StoreChannel
	//GetAll(discoveryType, clusterName string) StoreChannel
	//SetLastPingAt(discovery *model.ClusterDiscovery) StoreChannel
	//Cleanup() StoreChannel

	Save(discovery *model.ClusterDiscovery) error
	Delete(discovery *model.ClusterDiscovery) (bool, error)
	Exists(discovery *model.ClusterDiscovery) (bool, error)
	GetAll(discoveryType, clusterName string) ([]*model.ClusterDiscovery, error)
	SetLastPingAt(discovery *model.ClusterDiscovery) error
	Cleanup() error
}

type SchemeStore interface {
	//Save(scheme *model.Scheme) StoreChannel
	//Get(schemeId string) StoreChannel
	//GetByName(schemeName string) StoreChannel
	//GetAllPage(scope string, offset int, limit int) StoreChannel
	//Delete(schemeId string) StoreChannel
	//PermanentDeleteAll() StoreChannel

	Save(scheme *model.Scheme) (*model.Scheme, error)
	Get(schemeId string) (*model.Scheme, error)
	GetByName(schemeName string) (*model.Scheme, error)
	GetAllPage(scope string, offset int, limit int) ([]*model.Scheme, error)
	Delete(schemeId string) (*model.Scheme, error)
	PermanentDeleteAll() error
	CountByScope(scope string) (int64, error)
	CountWithoutPermission(scope, permissionID string, roleScope model.RoleScope, roleType model.RoleType) (int64, error)
}

type RoleStore interface {
	//Save(role *model.Role) StoreChannel
	//Get(roleId string) StoreChannel
	//GetByName(name string) StoreChannel
	//GetByNames(names []string) StoreChannel
	//GetByTeamId(teamId string) StoreChannel
	//Delete(roldId string) StoreChannel
	//PermanentDeleteAll() StoreChannel

	Save(role *model.Role) (*model.Role, error)
	Get(roleId string) (*model.Role, error)
	GetAll() ([]*model.Role, error)
	GetByName(name string) (*model.Role, error)
	GetByNames(names []string) ([]*model.Role, error)
	//GetByTeamId(teamId string) ([]*model.Role, error)
	Delete(roleId string) (*model.Role, error)
	PermanentDeleteAll() error
}

type ReactionStore interface {
	//Save(reaction *model.Reaction) StoreChannel
	//Delete(reaction *model.Reaction) StoreChannel
	//GetForPost(postId string, allowFromCache bool) StoreChannel
	//DeleteAllWithEmojiName(emojiName string) StoreChannel
	//PermanentDeleteBatch(endTime int64, limit int64) StoreChannel
	Save(reaction *model.Reaction) (*model.Reaction, error)
	Delete(reaction *model.Reaction) (*model.Reaction, error)
	GetForPost(postId string, allowFromCache bool) ([]*model.Reaction, error)
	DeleteAllWithEmojiName(emojiName string) error
	PermanentDeleteBatch(endTime int64, limit int64) (int64, error)
	BulkGetForPosts(postIds []string) ([]*model.Reaction, error)
}

type TokenStore interface {
	//Save(recovery *model.Token) StoreChannel
	//Delete(token string) StoreChannel
	//GetByToken(token string) StoreChannel
	//Cleanup()
	Save(recovery *model.Token) error
	Delete(token string) error
	GetByToken(token string) (*model.Token, error)
	Cleanup()
	RemoveAllTokensByType(tokenType string) error
}

type CommandStore interface {
	//Save(webhook *model.Command) StoreChannel
	//Get(id string) StoreChannel
	//GetByTeam(teamId string) StoreChannel
	//GetByTrigger(teamId string, trigger string) StoreChannel
	//Delete(commandId string, time int64) StoreChannel
	//PermanentDeleteByTeam(teamId string) StoreChannel
	//PermanentDeleteByUser(userId string) StoreChannel
	//Update(hook *model.Command) StoreChannel
	//AnalyticsCommandCount(teamId string) StoreChannel
	Save(webhook *model.Command) (*model.Command, error)
	GetByTrigger(teamId string, trigger string) (*model.Command, error)
	Get(id string) (*model.Command, error)
	GetByTeam(teamId string) ([]*model.Command, error)
	Delete(commandId string, time int64) error
	PermanentDeleteByTeam(teamId string) error
	PermanentDeleteByUser(userId string) error
	Update(hook *model.Command) (*model.Command, error)
	AnalyticsCommandCount(teamId string) (int64, error)
}

type AuditStore interface {
	//Save(audit *model.Audit) StoreChannel
	//Get(user_id string, offset int, limit int) StoreChannel
	//PermanentDeleteByUser(userId string) StoreChannel
	//PermanentDeleteBatch(endTime int64, limit int64) StoreChannel
	Save(audit *model.Audit) error
	Get(user_id string, offset int, limit int) (model.Audits, error)
	PermanentDeleteByUser(userId string) error
}

type OAuthStore interface {
	//SaveApp(app *model.OAuthApp) StoreChannel
	//UpdateApp(app *model.OAuthApp) StoreChannel
	//GetApp(id string) StoreChannel
	//GetAppByUser(userId string, offset, limit int) StoreChannel
	//GetApps(offset, limit int) StoreChannel
	//GetAuthorizedApps(userId string, offset, limit int) StoreChannel
	//DeleteApp(id string) StoreChannel
	//SaveAuthData(authData *model.AuthData) StoreChannel
	//GetAuthData(code string) StoreChannel
	//RemoveAuthData(code string) StoreChannel
	//PermanentDeleteAuthDataByUser(userId string) StoreChannel
	//SaveAccessData(accessData *model.AccessData) StoreChannel
	//UpdateAccessData(accessData *model.AccessData) StoreChannel
	//GetAccessData(token string) StoreChannel
	//GetAccessDataByUserForApp(userId, clientId string) StoreChannel
	//GetAccessDataByRefreshToken(token string) StoreChannel
	//GetPreviousAccessData(userId, clientId string) StoreChannel
	//RemoveAccessData(token string) StoreChannel
	SaveApp(app *model.OAuthApp) (*model.OAuthApp, error)
	UpdateApp(app *model.OAuthApp) (*model.OAuthApp, error)
	GetApp(id string) (*model.OAuthApp, error)
	GetAppByUser(userId string, offset, limit int) ([]*model.OAuthApp, error)
	GetApps(offset, limit int) ([]*model.OAuthApp, error)
	GetAuthorizedApps(userId string, offset, limit int) ([]*model.OAuthApp, error)
	DeleteApp(id string) error
	SaveAuthData(authData *model.AuthData) (*model.AuthData, error)
	GetAuthData(code string) (*model.AuthData, error)
	RemoveAuthData(code string) error
	PermanentDeleteAuthDataByUser(userId string) error
	SaveAccessData(accessData *model.AccessData) (*model.AccessData, error)
	UpdateAccessData(accessData *model.AccessData) (*model.AccessData, error)
	GetAccessData(token string) (*model.AccessData, error)
	GetAccessDataByUserForApp(userId, clientId string) ([]*model.AccessData, error)
	GetAccessDataByRefreshToken(token string) (*model.AccessData, error)
	GetPreviousAccessData(userId, clientId string) (*model.AccessData, error)
	RemoveAccessData(token string) error
	RemoveAllAccessData() error
}

type SystemStore interface {
	//Save(system *model.System) StoreChannel
	//SaveOrUpdate(system *model.System) StoreChannel
	//Update(system *model.System) StoreChannel
	//Get() StoreChannel
	//GetByName(name string) StoreChannel
	//PermanentDeleteByName(name string) StoreChannel
	Save(system *model.System) error
	SaveOrUpdate(system *model.System) error
	Update(system *model.System) error
	Get() (model.StringMap, error)
	GetByName(name string) (*model.System, error)
	PermanentDeleteByName(name string) (*model.System, error)
	InsertIfExists(system *model.System) (*model.System, error)
	SaveOrUpdateWithWarnMetricHandling(system *model.System) error
}
type UserAccessTokenStore interface {
	//Save(token *model.UserAccessToken) StoreChannel
	//Delete(tokenId string) StoreChannel
	//DeleteAllForUser(userId string) StoreChannel
	//Get(tokenId string) StoreChannel
	//GetAll(offset int, limit int) StoreChannel
	//GetByToken(tokenString string) StoreChannel
	//GetByUser(userId string, page, perPage int) StoreChannel
	//Search(term string) StoreChannel
	//UpdateTokenEnable(tokenId string) StoreChannel
	//UpdateTokenDisable(tokenId string) StoreChannel
	Save(token *model.UserAccessToken) (*model.UserAccessToken, error)
	DeleteAllForUser(userId string) error
	Delete(tokenId string) error
	Get(tokenId string) (*model.UserAccessToken, error)
	GetAll(offset int, limit int) ([]*model.UserAccessToken, error)
	GetByToken(tokenString string) (*model.UserAccessToken, error)
	GetByUser(userId string, page, perPage int) ([]*model.UserAccessToken, error)
	Search(term string) ([]*model.UserAccessToken, error)
	UpdateTokenEnable(tokenId string) error
	UpdateTokenDisable(tokenId string) error
}

type UserStore interface {
	//Save(user *model.User) StoreChannel
	//Update(user *model.User, allowRoleUpdate bool) StoreChannel
	//UpdateLastPictureUpdate(userId string) StoreChannel
	//ResetLastPictureUpdate(userId string) StoreChannel
	//UpdateUpdateAt(userId string) StoreChannel
	//UpdateLocale(userId string, locale string) StoreChannel
	//UpdatePassword(userId, newPassword string) StoreChannel
	//UpdateAuthData(userId string, service string, authData *string, email string, resetMfa bool) StoreChannel
	//UpdateMfaSecret(userId, secret string) StoreChannel
	//UpdateMfaActive(userId string, active bool) StoreChannel
	//Get(id string) StoreChannel
	//GetAll() StoreChannel
	//ClearCaches()
	//InvalidateProfilesInChannelCacheByUser(userId string)
	//InvalidateProfilesInChannelCache(channelId string)
	//GetProfilesInChannel(channelId string, offset int, limit int) StoreChannel
	//GetProfilesInChannelByStatus(channelId string, offset int, limit int) StoreChannel
	//GetAllProfilesInChannel(channelId string, allowFromCache bool) StoreChannel
	//GetProfilesNotInChannel(teamId string, channelId string, offset int, limit int) StoreChannel
	//GetProfilesWithoutTeam(offset int, limit int) StoreChannel
	//GetProfilesByUsernames(usernames []string, teamId string) StoreChannel
	//GetAllProfiles(options *model.UserGetOptions) StoreChannel
	//GetProfiles(options *model.UserGetOptions) StoreChannel
	//GetProfileByIds(userId []string, allowFromCache bool) StoreChannel
	//InvalidatProfileCacheForUser(userId string)
	//GetByEmail(email string) StoreChannel
	//GetById(id string) StoreChannel
	//GetByAuth(authData *string, authService string) StoreChannel
	//GetAllUsingAuthService(authService string) StoreChannel
	//GetByUsername(username string) StoreChannel
	//GetForLogin(loginId string, allowSignInWithUsername, allowSignInWithEmail bool) StoreChannel
	//VerifyEmail(userId string) StoreChannel
	//GetEtagForAllProfiles() StoreChannel
	//GetEtagForProfiles(teamId string) StoreChannel
	//UpdateFailedPasswordAttempts(userId string, attempts int) StoreChannel
	//GetTotalUsersCount() StoreChannel
	//GetSystemAdminProfiles() StoreChannel
	//PermanentDelete(userId string) StoreChannel
	//AnalyticsUniqueUserCount(teamId string) StoreChannel
	//AnalyticsActiveCount(time int64) StoreChannel
	//AnalyticsActiveCountForPeriod(startTime int64, endTime int64, options model.UserCountOptions) (int64, error)
	//GetUnreadCount(userId string) StoreChannel
	//GetUnreadCountForChannel(userId string, channelId string) StoreChannel
	//GetAnyUnreadPostCountForChannel(userId string, channelId string) StoreChannel
	//GetRecentlyActiveUsersForTeam(teamId string, offset, limit int) StoreChannel
	//GetNewUsersForTeam(teamId string, offset, limit int) StoreChannel
	//Search(teamId string, term string, options *model.UserSearchOptions) StoreChannel
	//SearchNotInTeam(notInTeamId string, term string, options *model.UserSearchOptions) StoreChannel
	//SearchInChannel(channelId string, term string, options *model.UserSearchOptions) StoreChannel
	//SearchNotInChannel(teamId string, channelId string, term string, options *model.UserSearchOptions) StoreChannel
	//SearchWithoutTeam(term string, options *model.UserSearchOptions) StoreChannel
	//AnalyticsGetInactiveUsersCount() StoreChannel
	//AnalyticsGetSystemAdminCount() StoreChannel
	//GetProfilesNotInTeam(teamId string, offset int, limit int) StoreChannel
	//GetEtagForProfilesNotInTeam(teamId string) StoreChannel
	//ClearAllCustomRoleAssignments() StoreChannel
	//InferSystemInstallDate() StoreChannel
	//GetAllAfter(limit int, afterId string) StoreChannel
	//Count(options model.UserCountOptions) StoreChannel
	Save(user *model.User) (*model.User, *model.AppError)
	Update(user *model.User, allowRoleUpdate bool) (*model.UserUpdate, *model.AppError)
	UpdateLastPictureUpdate(userId string) *model.AppError
	ResetLastPictureUpdate(userId string) *model.AppError
	UpdatePassword(userId, newPassword string) *model.AppError
	UpdateUpdateAt(userId string) (int64, *model.AppError)
	UpdateAuthData(userId string, service string, authData *string, email string, resetMfa bool) (string, *model.AppError)
	UpdateMfaSecret(userId, secret string) *model.AppError
	UpdateMfaActive(userId string, active bool) *model.AppError
	Get(id string) (*model.User, *model.AppError)
	GetAll() ([]*model.User, *model.AppError)
	ClearCaches()
	InvalidateProfilesInChannelCacheByUser(userId string)
	InvalidateProfilesInChannelCache(channelId string)
	GetProfilesInChannel(options *model.UserGetOptions) ([]*model.User, *model.AppError)
	GetProfilesInChannelByStatus(options *model.UserGetOptions) ([]*model.User, *model.AppError)
	GetAllProfilesInChannel(channelId string, allowFromCache bool) (map[string]*model.User, *model.AppError)
	GetProfilesNotInChannel(teamId string, channelId string, groupConstrained bool, offset int, limit int, viewRestrictions *model.ViewUsersRestrictions) ([]*model.User, *model.AppError)
	GetProfilesWithoutTeam(options *model.UserGetOptions) ([]*model.User, *model.AppError)
	GetProfilesByUsernames(usernames []string, viewRestrictions *model.ViewUsersRestrictions) ([]*model.User, *model.AppError)
	GetAllProfiles(options *model.UserGetOptions) ([]*model.User, *model.AppError)
	GetProfiles(options *model.UserGetOptions) ([]*model.User, *model.AppError)
	GetProfileByIds(userIds []string, options *UserGetByIdsOpts, allowFromCache bool) ([]*model.User, *model.AppError)
	GetProfileByGroupChannelIdsForUser(userId string, channelIds []string) (map[string][]*model.User, *model.AppError)
	InvalidateProfileCacheForUser(userId string)
	GetByEmail(email string) (*model.User, *model.AppError)
	GetByAuth(authData *string, authService string) (*model.User, *model.AppError)
	GetAllUsingAuthService(authService string) ([]*model.User, *model.AppError)
	GetAllNotInAuthService(authServices []string) ([]*model.User, *model.AppError)
	GetByUsername(username string) (*model.User, *model.AppError)
	GetForLogin(loginId string, allowSignInWithUsername, allowSignInWithEmail bool) (*model.User, *model.AppError)
	VerifyEmail(userId, email string) (string, *model.AppError)
	GetEtagForAllProfiles() string
	GetEtagForProfiles(teamId string) string
	UpdateFailedPasswordAttempts(userId string, attempts int) *model.AppError
	GetSystemAdminProfiles() (map[string]*model.User, *model.AppError)
	PermanentDelete(userId string) *model.AppError
	AnalyticsActiveCount(time int64, options model.UserCountOptions) (int64, *model.AppError)
	AnalyticsActiveCountForPeriod(startTime int64, endTime int64, options model.UserCountOptions) (int64, error)
	GetUnreadCount(userId string) (int64, *model.AppError)
	GetUnreadCountForChannel(userId string, channelId string) (int64, *model.AppError)
	GetAnyUnreadPostCountForChannel(userId string, channelId string) (int64, *model.AppError)
	GetRecentlyActiveUsersForTeam(teamId string, offset, limit int, viewRestrictions *model.ViewUsersRestrictions) ([]*model.User, *model.AppError)
	GetNewUsersForTeam(teamId string, offset, limit int, viewRestrictions *model.ViewUsersRestrictions) ([]*model.User, *model.AppError)
	Search(teamId string, term string, options *model.UserSearchOptions) ([]*model.User, *model.AppError)
	SearchNotInTeam(notInTeamId string, term string, options *model.UserSearchOptions) ([]*model.User, *model.AppError)
	SearchInChannel(channelId string, term string, options *model.UserSearchOptions) ([]*model.User, *model.AppError)
	SearchNotInChannel(teamId string, channelId string, term string, options *model.UserSearchOptions) ([]*model.User, *model.AppError)
	SearchWithoutTeam(term string, options *model.UserSearchOptions) ([]*model.User, *model.AppError)
	SearchInGroup(groupID string, term string, options *model.UserSearchOptions) ([]*model.User, *model.AppError)
	AnalyticsGetInactiveUsersCount() (int64, *model.AppError)
	AnalyticsGetExternalUsers(hostDomain string) (bool, *model.AppError)
	AnalyticsGetSystemAdminCount() (int64, *model.AppError)
	AnalyticsGetGuestCount() (int64, *model.AppError)
	GetProfilesNotInTeam(teamId string, groupConstrained bool, offset int, limit int, viewRestrictions *model.ViewUsersRestrictions) ([]*model.User, *model.AppError)
	GetEtagForProfilesNotInTeam(teamId string) string
	ClearAllCustomRoleAssignments() *model.AppError
	InferSystemInstallDate() (int64, *model.AppError)
	GetAllAfter(limit int, afterId string) ([]*model.User, *model.AppError)
	GetUsersBatchForIndexing(startTime, endTime int64, limit int) ([]*model.UserForIndexing, *model.AppError)
	Count(options model.UserCountOptions) (int64, *model.AppError)
	GetTeamGroupUsers(teamID string) ([]*model.User, *model.AppError)
	GetChannelGroupUsers(channelID string) ([]*model.User, *model.AppError)
	PromoteGuestToUser(userID string) *model.AppError
	DemoteUserToGuest(userID string) *model.AppError
	DeactivateGuests() ([]string, *model.AppError)
	AutocompleteUsersInChannel(teamId, channelId, term string, options *model.UserSearchOptions) (*model.UserAutocompleteInChannel, *model.AppError)
	GetKnownUsers(userID string) ([]string, *model.AppError)
}

type TermsOfServiceStore interface {
	//Save(termsOfService *model.TermsOfService) StoreChannel
	//GetLatest(allowFromCache bool) StoreChannel
	//Get(id string, allowFromCache bool) StoreChannel
	Save(termsOfService *model.TermsOfService) (*model.TermsOfService, error)
	GetLatest(allowFromCache bool) (*model.TermsOfService, error)
	Get(id string, allowFromCache bool) (*model.TermsOfService, error)
}

type SessionStore interface {
	//Save(session *model.Session) StoreChannel
	//Get(sessionIdOrToken string) StoreChannel
	//GetSessions(userId string) StoreChannel
	//GetSessionsWithActiveDeviceIds(userId string) StoreChannel
	//Remove(sessionIdOrToken string) StoreChannel
	//RemoveAllSessions() StoreChannel
	//PermanentDeleteSessionsByUser(teamId string) StoreChannel
	//UpdateLastActivityAt(sessionId string, time int64) StoreChannel
	//UpdateRoles(userId string, roles string) StoreChannel
	//UpdateDeviceId(id string, deviceId string, expiresAt int64) StoreChannel
	//AnalyticsSessionCount() StoreChannel
	//Cleanup(expiryTime int64, batchSize int64)
	Get(sessionIdOrToken string) (*model.Session, error)
	Save(session *model.Session) (*model.Session, error)
	GetSessions(userId string) ([]*model.Session, error)
	GetSessionsWithActiveDeviceIds(userId string) ([]*model.Session, error)
	GetSessionsExpired(thresholdMillis int64, mobileOnly bool, unnotifiedOnly bool) ([]*model.Session, error)
	UpdateExpiredNotify(sessionid string, notified bool) error
	Remove(sessionIdOrToken string) error
	RemoveAllSessions() error
	PermanentDeleteSessionsByUser(teamId string) error
	UpdateExpiresAt(sessionId string, time int64) error
	UpdateLastActivityAt(sessionId string, time int64) error
	UpdateRoles(userId string, roles string) (string, error)
	UpdateDeviceId(id string, deviceId string, expiresAt int64) (string, error)
	UpdateProps(session *model.Session) error
	AnalyticsSessionCount() (int64, error)
	Cleanup(expiryTime int64, batchSize int64)
}

type StatusStore interface {
	//SaveOrUpdate(status *model.Status) StoreChannel
	//Get(userId string) StoreChannel
	//GetByIds(userIds []string) StoreChannel
	//GetOnlineAway() StoreChannel
	//GetOnline() StoreChannel
	//GetAllFromTeam(teamId string) StoreChannel
	//ResetAll() StoreChannel
	//GetTotalActiveUsersCount() StoreChannel
	//UpdateLastActivityAt(userId string, lastActivityAt int64) StoreChannel
	SaveOrUpdate(status *model.Status) error
	Get(userId string) (*model.Status, error)
	GetByIds(userIds []string) ([]*model.Status, error)
	ResetAll() error
	GetTotalActiveUsersCount() (int64, error)
	UpdateLastActivityAt(userId string, lastActivityAt int64) error
}

type TeamStore interface {
	//Save(team *model.Team) StoreChannel
	//Update(team *model.Team) StoreChannel
	//UpdateDisplayName(name string, teamId string) StoreChannel
	//Get(id string) StoreChannel
	//GetByName(name string) StoreChannel
	//SearchByName(name string) StoreChannel
	//SearchAll(term string) StoreChannel
	//SearchOpen(term string) StoreChannel
	//GetAll() StoreChannel
	//GetAllPage(offset int, limit int) StoreChannel
	//GetAllTeamListing() StoreChannel
	//GetAllTeamPageListing(offset int, limit int) StoreChannel
	//GetTeamsByUserId(userId string) StoreChannel
	//GetByInviteId(inviteId string) StoreChannel
	//PermanentDelete(teamId string) StoreChannel
	//AnalyticsTeamCount() StoreChannel
	//SaveMember(member *model.TeamMember, maxUsersPerTeam int) StoreChannel
	//UpdateMember(member *model.TeamMember) StoreChannel
	//GetMember(teamId string, userId string) StoreChannel
	//GetMembers(teamId string, offset int, limit int) StoreChannel
	//GetMembersByIds(teamId string, userIds []string) StoreChannel
	//GetTotalMemberCount(teamId string) StoreChannel
	//GetActiveMemberCount(teamId string) StoreChannel
	//GetTeamsForUser(userId string) StoreChannel
	//GetChannelUnreadsForAllTeams(excludeTeamId, userId string) StoreChannel
	//GetChannelUnreadsForTeam(teamId, userId string) StoreChannel
	//RemoveMember(teamId string, userId string) StoreChannel
	//RemoveAllMembersByTeam(teamId string) StoreChannel
	//RemoveAllMembersByUser(userId string) StoreChannel
	//UpdateLastTeamIconUpdate(teamId string, curTime int64) StoreChannel
	//GetTeamsByScheme(schemeId string, offset int, limit int) StoreChannel
	//MigrateTeamMembers(fromTeamId string, fromUserId string) StoreChannel
	//ResetAllTeamSchemes() StoreChannel
	//ClearAllCustomRoleAssignments() StoreChannel
	//AnalyticsGetTeamCountForScheme(schemeId string) StoreChannel
	//GetAllForExportAfter(limit int, afterId string) StoreChannel
	//GetTeamMembersForExport(userId string) StoreChannel
	Save(team *model.Team) (*model.Team, error)
	Update(team *model.Team) (*model.Team, error)
	Get(id string) (*model.Team, error)
	GetByName(name string) (*model.Team, error)
	GetByNames(name []string) ([]*model.Team, error)
	SearchAll(term string, opts *model.TeamSearch) ([]*model.Team, error)
	SearchAllPaged(term string, opts *model.TeamSearch) ([]*model.Team, int64, error)
	SearchOpen(term string) ([]*model.Team, error)
	SearchPrivate(term string) ([]*model.Team, error)
	GetAll() ([]*model.Team, error)
	GetAllPage(offset int, limit int) ([]*model.Team, error)
	GetAllPrivateTeamListing() ([]*model.Team, error)
	GetAllPrivateTeamPageListing(offset int, limit int) ([]*model.Team, error)
	GetAllPublicTeamPageListing(offset int, limit int) ([]*model.Team, error)
	GetAllTeamListing() ([]*model.Team, error)
	GetAllTeamPageListing(offset int, limit int) ([]*model.Team, error)
	GetTeamsByUserId(userId string) ([]*model.Team, error)
	GetByInviteId(inviteId string) (*model.Team, error)
	PermanentDelete(teamId string) error
	AnalyticsTeamCount(includeDeleted bool) (int64, error)
	AnalyticsPublicTeamCount() (int64, error)
	AnalyticsPrivateTeamCount() (int64, error)
	SaveMultipleMembers(members []*model.TeamMember, maxUsersPerTeam int) ([]*model.TeamMember, error)
	SaveMember(member *model.TeamMember, maxUsersPerTeam int) (*model.TeamMember, error)
	UpdateMember(member *model.TeamMember) (*model.TeamMember, error)
	UpdateMultipleMembers(members []*model.TeamMember) ([]*model.TeamMember, error)
	GetMember(teamId string, userId string) (*model.TeamMember, error)
	GetMembers(teamId string, offset int, limit int, teamMembersGetOptions *model.TeamMembersGetOptions) ([]*model.TeamMember, error)
	GetMembersByIds(teamId string, userIds []string, restrictions *model.ViewUsersRestrictions) ([]*model.TeamMember, error)
	GetTotalMemberCount(teamId string, restrictions *model.ViewUsersRestrictions) (int64, error)
	GetActiveMemberCount(teamId string, restrictions *model.ViewUsersRestrictions) (int64, error)
	GetTeamsForUser(userId string) ([]*model.TeamMember, error)
	GetTeamsForUserWithPagination(userId string, page, perPage int) ([]*model.TeamMember, error)
	GetChannelUnreadsForAllTeams(excludeTeamId, userId string) ([]*model.ChannelUnread, error)
	GetChannelUnreadsForTeam(teamId, userId string) ([]*model.ChannelUnread, error)
	RemoveMember(teamId string, userId string) error
	RemoveMembers(teamId string, userIds []string) error
	RemoveAllMembersByTeam(teamId string) error
	RemoveAllMembersByUser(userId string) error
	UpdateLastTeamIconUpdate(teamId string, curTime int64) error
	GetTeamsByScheme(schemeId string, offset int, limit int) ([]*model.Team, error)
	MigrateTeamMembers(fromTeamId string, fromUserId string) (map[string]string, error)
	ResetAllTeamSchemes() error
	ClearAllCustomRoleAssignments() error
	AnalyticsGetTeamCountForScheme(schemeId string) (int64, error)
	GetAllForExportAfter(limit int, afterId string) ([]*model.TeamForExport, error)
	GetTeamMembersForExport(userId string) ([]*model.TeamMemberForExport, error)
	UserBelongsToTeams(userId string, teamIds []string) (bool, error)
	GetUserTeamIds(userId string, allowFromCache bool) ([]string, error)
	InvalidateAllTeamIdsForUser(userId string)
	ClearCaches()

	// UpdateMembersRole sets all of the given team members to admins and all of the other members of the team to
	// non-admin members.
	UpdateMembersRole(teamID string, userIDs []string) error

	// GroupSyncedTeamCount returns the count of non-deleted group-constrained teams.
	GroupSyncedTeamCount() (int64, error)
}

type WebhookStore interface {
	//SaveIncoming(webhook *model.IncomingWebhook) StoreChannel
	//GetIncoming(id string, allowFromCache bool) StoreChannel
	//GetIncomingList(offset, limit int) StoreChannel
	//GetIncomingByTeam(teamId string, offset, limit int) StoreChannel
	//UpdateIncoming(webhook *model.IncomingWebhook) StoreChannel
	//GetIncomingByChannel(channelId string) StoreChannel
	//DeleteIncoming(webhookId string, time int64) StoreChannel
	//PermanentDeleteIncomingByChannel(channelId string) StoreChannel
	//PermanentDeleteIncomingByUser(userId string) StoreChannel
	//
	//SaveOutgoing(webhook *model.OutgoingWebhook) StoreChannel
	//GetOutgoing(id string) StoreChannel
	//GetOutgoingList(offset, limit int) StoreChannel
	//GetOutgoingByChannel(channelId string, offset, limit int) StoreChannel
	//GetOutgoingByTeam(teamId string, offset, limit int) StoreChannel
	//DeleteOutgoing(webhookId string, time int64) StoreChannel
	//PermanentDeleteOutgoingByChannel(channelId string) StoreChannel
	//PermanentDeleteOutgoingByUser(userId string) StoreChannel
	//UpdateOutgoing(hook *model.OutgoingWebhook) StoreChannel
	//
	//AnalyticsIncomingCount(teamId string) StoreChannel
	//AnalyticsOutgoingCount(teamId string) StoreChannel
	//InvalidateWebhookCache(webhook string)
	//ClearCaches()
	SaveIncoming(webhook *model.IncomingWebhook) (*model.IncomingWebhook, error)
	GetIncoming(id string, allowFromCache bool) (*model.IncomingWebhook, error)
	GetIncomingList(offset, limit int) ([]*model.IncomingWebhook, error)
	GetIncomingListByUser(userId string, offset, limit int) ([]*model.IncomingWebhook, error)
	GetIncomingByTeam(teamId string, offset, limit int) ([]*model.IncomingWebhook, error)
	GetIncomingByTeamByUser(teamId string, userId string, offset, limit int) ([]*model.IncomingWebhook, error)
	UpdateIncoming(webhook *model.IncomingWebhook) (*model.IncomingWebhook, error)
	GetIncomingByChannel(channelId string) ([]*model.IncomingWebhook, error)
	DeleteIncoming(webhookId string, time int64) error
	PermanentDeleteIncomingByChannel(channelId string) error
	PermanentDeleteIncomingByUser(userId string) error

	SaveOutgoing(webhook *model.OutgoingWebhook) (*model.OutgoingWebhook, error)
	GetOutgoing(id string) (*model.OutgoingWebhook, error)
	GetOutgoingByChannel(channelId string, offset, limit int) ([]*model.OutgoingWebhook, error)
	GetOutgoingByChannelByUser(channelId string, userId string, offset, limit int) ([]*model.OutgoingWebhook, error)
	GetOutgoingList(offset, limit int) ([]*model.OutgoingWebhook, error)
	GetOutgoingListByUser(userId string, offset, limit int) ([]*model.OutgoingWebhook, error)
	GetOutgoingByTeam(teamId string, offset, limit int) ([]*model.OutgoingWebhook, error)
	GetOutgoingByTeamByUser(teamId string, userId string, offset, limit int) ([]*model.OutgoingWebhook, error)
	DeleteOutgoing(webhookId string, time int64) error
	PermanentDeleteOutgoingByChannel(channelId string) error
	PermanentDeleteOutgoingByUser(userId string) error
	UpdateOutgoing(hook *model.OutgoingWebhook) (*model.OutgoingWebhook, error)

	AnalyticsIncomingCount(teamId string) (int64, error)
	AnalyticsOutgoingCount(teamId string) (int64, error)
	InvalidateWebhookCache(webhook string)
	ClearCaches()
}

type LinkMetadataStore interface {
	Save(linkMetadata *model.LinkMetadata) (*model.LinkMetadata, error)
	Get(url string, timestamp int64) (*model.LinkMetadata, error)
}

type UserGetByIdsOpts struct {
	// IsAdmin tracks whether or not the request is being made by an administrator. Does nothing when provided by a client.
	IsAdmin bool

	// Restrict to search in a list of teams and channels. Does nothing when provided by a client.
	ViewRestrictions *model.ViewUsersRestrictions

	// Since filters the users based on their UpdateAt timestamp.
	Since int64
}
