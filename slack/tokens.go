package slack

import (
	"github.com/keremk/challenge-bot/config"
	"github.com/keremk/challenge-bot/db"
	"github.com/keremk/challenge-bot/models"
)

func getUserToken(env config.Environment, userID string) (string, error) {
	store := db.NewStore(env, db.SlackUsersCollection)

	user := models.SlackUser{}
	err := store.FindByID(userID, &user)
	return user.Token, err
}

func getBotToken(env config.Environment, teamID string) (string, error) {
	store := db.NewStore(env, db.SlackTeamsCollection)

	team := models.SlackTeam{}
	err := store.FindByID(teamID, &team)
	return team.BotToken, err
}
