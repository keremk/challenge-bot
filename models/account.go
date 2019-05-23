package models

import (
	"errors"
	"reflect"

	"github.com/keremk/challenge-bot/config"
	"github.com/keremk/challenge-bot/db"
	"github.com/keremk/challenge-bot/util"
)

type GithubAccount struct {
	ID              string
	Name            string
	Owner           string
	Org             string
	InstallationID  string
	AccessToken     string
	CreatedByTeamID string
}

func NewGithubAccount(input map[string]string, token string) GithubAccount {
	return GithubAccount{
		ID:              util.RandomString(16),
		Name:            input["name"],
		Owner:           input["owner"],
		Org:             input["org"],
		AccessToken:     token,
		InstallationID:  input["installation_id"],
		CreatedByTeamID: input["team_id"],
	}
}

func HardcodedGithubAccount(token, installationID string) GithubAccount {
	return GithubAccount{
		ID:             "93b8c538c36f67b3e1db343ade0ef",
		Name:           "Hardcoded",
		Owner:          "xingtesting",
		Org:            "",
		InstallationID: installationID,
		AccessToken:    token,
	}
}

func GetGithubAccount(env config.Environment, name string) (GithubAccount, error) {
	account := GithubAccount{}
	store, err := db.NewStore(env, db.GithubAccountsCollection)
	if err != nil {
		return account, err
	}

	err = store.FindFirst("Name", name, &account)
	return account, err
}

func UpdateGithubAccount(env config.Environment, account GithubAccount) error {
	store, err := db.NewStore(env, db.GithubAccountsCollection)
	if err != nil {
		return err
	}
	return store.Update(account.ID, account)
}

func GetAllAccounts(env config.Environment) ([]GithubAccount, error) {
	store, err := db.NewStore(env, db.GithubAccountsCollection)
	if err != nil {
		return nil, err
	}

	var all []GithubAccount
	result, err := store.FindAll(reflect.TypeOf(all))
	all, ok := result.([]GithubAccount)
	if !ok {
		return nil, errors.New("[ERROR] Cannot convert")
	}
	return all, err
}
