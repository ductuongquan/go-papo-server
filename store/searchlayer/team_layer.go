// Copyright (c) 2015-present Ladifire, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package searchlayer

import (
	model "bitbucket.org/enesyteam/papo-server/model"
	store "bitbucket.org/enesyteam/papo-server/store"
)

type SearchTeamStore struct {
	store.TeamStore
	rootStore *SearchStore
}

func (s SearchTeamStore) SaveMember(teamMember *model.TeamMember, maxUsersPerTeam int) (*model.TeamMember, error) {
	member, err := s.TeamStore.SaveMember(teamMember, maxUsersPerTeam)
	if err == nil {
		s.rootStore.indexUserFromID(member.UserId)
	}
	return member, err
}

func (s SearchTeamStore) UpdateMember(teamMember *model.TeamMember) (*model.TeamMember, error) {
	member, err := s.TeamStore.UpdateMember(teamMember)
	if err == nil {
		s.rootStore.indexUserFromID(member.UserId)
	}
	return member, err
}

func (s SearchTeamStore) RemoveMember(teamId string, userId string) error {
	err := s.TeamStore.RemoveMember(teamId, userId)
	if err == nil {
		s.rootStore.indexUserFromID(userId)
	}
	return err
}

func (s SearchTeamStore) RemoveAllMembersByUser(userId string) error {
	err := s.TeamStore.RemoveAllMembersByUser(userId)
	if err == nil {
		s.rootStore.indexUserFromID(userId)
	}
	return err
}
