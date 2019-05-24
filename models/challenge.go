package models

import (
	"errors"
	"reflect"

	"github.com/keremk/challenge-bot/config"
	"github.com/keremk/challenge-bot/db"
	"github.com/keremk/challenge-bot/util"
)

type Challenge struct {
	ID                string
	Name              string
	GithubOwner       string
	GithubOrg         string
	GithubAccountName string
	GithubToken       string
	TemplateRepo      string
	RepoNameFormat    string
	CreatedByTeamID   string
}

func NewChallenge(input map[string]string) Challenge {
	return Challenge{
		ID:                util.RandomString(16),
		Name:              input["challenge_name"],
		GithubAccountName: input["github_account"],
		TemplateRepo:      input["template_repo"],
		RepoNameFormat:    input["repo_name_format"],
		CreatedByTeamID:   input["team_id"],
	}
}

func getChallenge(env config.Environment, name string) (Challenge, error) {
	challenge := Challenge{}
	store, err := db.NewStore(env, db.SettingsCollection)
	if err != nil {
		return challenge, err
	}

	err = store.FindFirst("Name", name, &challenge)
	return challenge, err
}

func GetAllChallenges(env config.Environment) ([]Challenge, error) {
	store, err := db.NewStore(env, db.SettingsCollection)
	if err != nil {
		return nil, err
	}

	var all []Challenge
	result, err := store.FindAll(reflect.TypeOf(all))
	all, ok := result.([]Challenge)
	if !ok {
		return nil, errors.New("[ERROR] Cannot convert")
	}
	return all, err
}

func UpdateChallenge(env config.Environment, challenge Challenge) error {
	store, err := db.NewStore(env, db.SettingsCollection)
	if err != nil {
		return err
	}

	return store.Update(challenge.ID, challenge)
}
