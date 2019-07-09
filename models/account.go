package models

import (
	"errors"
	"reflect"

	"github.com/keremk/challenge-bot/config"
	"github.com/keremk/challenge-bot/db"
)

type GithubAccount struct {
	Name           string 	`bson:"Name"`
	Owner          string 	`bson:"Owner"`
	Org            string 	`bson:"Org"`
	InstallationID string 	`bson:"InstallationID"`
	AccessToken    string 	`bson:"AccessToken"`
}

func NewGithubAccount(installationID, token string) GithubAccount {
	return GithubAccount{
		AccessToken:    token,
		InstallationID: installationID,
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

func CreateGithubAccount(env config.Environment, account GithubAccount) error {
	store, err := db.NewStore(env, db.GithubAccountsCollection)
	if err != nil {
		return err
	}
	return store.Update(account.InstallationID, account)
}

func EditGithubAccount(env config.Environment, installationID, org, owner, name string) error {
	store, err := db.NewStore(env, db.GithubAccountsCollection)
	if err != nil {
		return err
	}

	account := map[string]interface{}{
		"Owner": owner,
		"Org":   org,
		"Name":  name,
	}
	return store.Merge(installationID, account)
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
