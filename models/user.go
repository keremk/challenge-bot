package models

import (
	"github.com/keremk/challenge-bot/config"
	"github.com/keremk/challenge-bot/db"
)

type SlackUser struct {
	ID    string `bson:"ID"`
	Token string `bson:"Token"`
}

func GetSlackUser(env config.Environment, id string) (SlackUser, error) {
	user := SlackUser{}
	store, err := db.NewStore(env, db.SlackUsersCollection)
	if err != nil {
		return user, err
	}

	err = store.FindByID(id, &user)
	return user, err
}

func UpdateSlackUser(env config.Environment, user SlackUser) error {
	store, err := db.NewStore(env, db.SlackUsersCollection)
	if err != nil {
		return err
	}
	return store.Update(user.ID, user)
}
