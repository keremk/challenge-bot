package models

import (
	"fmt"

	"github.com/keremk/challenge-bot/config"
)

type ChallengeSetup struct {
	Name            string
	GithubOwner     string
	GithubOrg       string
	GithubToken     string
	TemplateRepo    string
	RepoNameFormat  string
	CreatedByTeamID string
	Slots           []Slot
}

func GetChallengeSetup(env config.Environment, name string) (ChallengeSetup, error) {
	challenge, err := getChallenge(env, name)
	if err != nil {
		return ChallengeSetup{}, err
	}

	account, err := GetGithubAccount(env, challenge.GithubAccountName)
	if err != nil {
		return ChallengeSetup{}, err
	}

	return ChallengeSetup{
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
