package models

import (
	"fmt"
	"sort"

	"github.com/keremk/challenge-bot/config"
)

type ChallengeSetup struct {
	ID              string
	Name            string
	GithubOwner     string
	GithubOrg       string
	GithubToken     string
	TemplateRepo    string
	RepoNameFormat  string
	CreatedByTeamID string
	Slots           map[SlotID]*Slot
}

func GetChallengeSetupByID(env config.Environment, id string) (ChallengeSetup, error) {
	challenge, err := getChallengeByID(env, id)
	if err != nil {
		return ChallengeSetup{}, err
	}

	return getChallengeSetup(env, challenge)
}

func GetChallengeSetupByName(env config.Environment, name string) (ChallengeSetup, error) {
	challenge, err := getChallengeByName(env, name)
	if err != nil {
		return ChallengeSetup{}, err
	}

	return getChallengeSetup(env, challenge)
}

func getChallengeSetup(env config.Environment, challenge Challenge) (ChallengeSetup, error) {
	account, err := GetGithubAccount(env, challenge.GithubAccountName)
	if err != nil {
		return ChallengeSetup{}, err
	}

	return ChallengeSetup{
		ID:              challenge.ID,
		Name:            challenge.Name,
		GithubOwner:     account.Owner,
		GithubOrg:       account.Org,
		GithubToken:     account.AccessToken,
		TemplateRepo:    challenge.TemplateRepo,
		RepoNameFormat:  challenge.RepoNameFormat,
		CreatedByTeamID: challenge.CreatedByTeamID,
		Slots:           challenge.Slots,
	}, nil
}

func (s ChallengeSetup) OrgOrOwner() string {
	if s.GithubOrg != "" {
		return s.GithubOrg
	} else {
		return s.GithubOwner
	}
}

func (s ChallengeSetup) TemplateRepositoryURL() string {
	return fmt.Sprintf("https://github.com/%v/%v.git", s.OrgOrOwner(), s.TemplateRepo)
}

func (s ChallengeSetup) TrackingIssuesURL() string {
	return fmt.Sprintf("https://github.com/%v/%v/issues", s.OrgOrOwner(), s.TemplateRepo)
}

func (s ChallengeSetup) GetSlotsInOrder() []*Slot {
	slots := make([]*Slot, 0, len(s.Slots))
	for _, slot := range s.Slots {
		slots = append(slots, slot)
	}

	sort.Slice(slots, func(i, j int) bool { return slots[i].Ordinal < slots[j].Ordinal })

	return slots
}
