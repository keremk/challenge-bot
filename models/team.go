package models

import (
	"github.com/keremk/challenge-bot/config"
	"github.com/keremk/challenge-bot/db"
)

type SlackTeam struct {
	ID        string
	Name      string
	BotToken  string
	BotUserID string
}

func GetSlackTeam(env config.Environment, id string) (SlackTeam, error) {
	team := SlackTeam{}
	store, err := db.NewStore(env, db.SlackTeamsCollection)
	if err != nil {
		return team, err
	}

	err = store.FindByID(id, &team)
	return team, err
}

func UpdateSlackTeam(env config.Environment, team SlackTeam) error {
	store, err := db.NewStore(env, db.SlackTeamsCollection)
	if err != nil {
		return err
	}
	return store.Update(team.ID, team)
}
