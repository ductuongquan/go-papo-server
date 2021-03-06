// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package sqlstore

import (
	"bitbucket.org/enesyteam/papo-server/store"
	sq "github.com/Masterminds/squirrel"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/mattermost/gorp"
)

type SqlStore interface {
	DriverName() string
	GetCurrentSchemaVersion() string
	GetMaster() *gorp.DbMap
	GetSearchReplica() *gorp.DbMap
	GetReplica() *gorp.DbMap
	TotalMasterDbConnections() int
	TotalReadDbConnections() int
	TotalSearchDbConnections() int
	MarkSystemRanUnitTests()
	DoesTableExist(tablename string) bool
	DoesColumnExist(tableName string, columName string) bool
	DoesTriggerExist(triggerName string) bool
	CreateColumnIfNotExists(tableName string, columnName string, mySqlColType string, postgresColType string, defaultValue string) bool
	CreateColumnIfNotExistsNoDefault(tableName string, columnName string, mySqlColType string, postgresColType string) bool
	RemoveColumnIfExists(tableName string, columnName string) bool
	RemoveTableIfExists(tableName string) bool
	RenameColumnIfExists(tableName string, oldColumnName string, newColumnName string, colType string) bool
	GetMaxLengthOfColumnIfExists(tableName string, columnName string) string
	AlterColumnTypeIfExists(tableName string, columnName string, mySqlColType string, postgresColType string) bool
	CreateUniqueIndexIfNotExists(indexName string, tableName string, columnName string) bool
	CreateIndexIfNotExists(indexName string, tableName string, columnName string) bool
	CreateCompositeIndexIfNotExists(indexName string, tableName string, columnNames []string) bool
	CreateFullTextIndexIfNotExists(indexName string, tableName string, columnName string) bool
	RemoveIndexIfExists(indexName string, tableName string) bool
	GetAllConns() []*gorp.DbMap
	Close()
	LockToMaster()
	UnlockFromMaster()
	Team() store.TeamStore
	User() store.UserStore
	Audit() store.AuditStore
	ClusterDiscovery() store.ClusterDiscoveryStore
	Session() store.SessionStore
	OAuth() store.OAuthStore
	System() store.SystemStore
	Webhook() store.WebhookStore
	Command() store.CommandStore
	Preference() store.PreferenceStore
	Token() store.TokenStore
	Status() store.StatusStore
	Reaction() store.ReactionStore
	UserAccessToken() store.UserAccessTokenStore
	Role() store.RoleStore
	Scheme() store.SchemeStore
	TermsOfService() store.TermsOfServiceStore
	getQueryBuilder() sq.StatementBuilderType
}

