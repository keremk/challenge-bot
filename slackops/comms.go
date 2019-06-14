package slackops

import (
	"log"

	"github.com/keremk/challenge-bot/config"
	"github.com/keremk/challenge-bot/models"
	"github.com/nlopes/slack"
)

func postMessage(env config.Environment, teamID string, targetChannel string, msgOption slack.MsgOption) error {
	token, err := getBotToken(env, teamID)
	if err != nil {
		return err
	}

	slackClient := slack.New(token)
	_, _, err = slackClient.PostMessage(targetChannel, msgOption)
	// log.Printf("[INFO]Message TS is : %s", messageTs)
	if err != nil {
		return err
	}
	return nil
}

func updateMessage(env config.Environment, teamID, targetChannel, messageTs string, msgOption slack.MsgOption) error {
	token, err := getBotToken(env, teamID)
	if err != nil {
		return err
	}

	// log.Printf("[INFO]Update message TS is: %s", messageTs)
	slackClient := slack.New(token)
	_, _, _, err = slackClient.UpdateMessage(targetChannel, messageTs, msgOption)
	if err != nil {
		log.Println("[ERROR] cannot update message - ", err)
		return err
	}
	// log.Printf("[INFO] Channel %s, timestamp %s, test %s", channel, timestamp, text)
	return nil
}

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
