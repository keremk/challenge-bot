package slackops

import (
	"log"

	"github.com/keremk/challenge-bot/config"
	"github.com/keremk/challenge-bot/models"
)

func getUserToken(env config.Environment, userID string) (string, error) {
	user, err := models.GetSlackUser(env, userID)
	if err != nil {
		log.Println("[ERROR] Cannot retrieve user token ", err)
		return "", err
	}
	return user.Token, err
}

func getBotToken(env config.Environment, teamID string) (string, error) {
	team, err := models.GetSlackTeam(env, teamID)
	if err != nil {
		log.Println("[ERROR] Cannot retrieve bot token ", err)
		return "", err
	}
	return team.BotToken, err
}
