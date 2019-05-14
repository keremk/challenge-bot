package slack

import (
	"github.com/keremk/challenge-bot/config"
	"github.com/keremk/challenge-bot/models"
)

func getUserToken(env config.Environment, userID string) (string, error) {
	user, err := models.GetSlackUser(env, userID)
	if err != nil {
		return "", err
	}
	return user.Token, err
}

func getBotToken(env config.Environment, teamID string) (string, error) {
	team, err := models.GetSlackTeam(env, teamID)
	if err != nil {
		return "", err
	}
	return team.BotToken, err
}
