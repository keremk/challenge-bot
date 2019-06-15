package slackops

import (
	"log"

	"github.com/keremk/challenge-bot/config"
	"github.com/keremk/challenge-bot/models"
	"github.com/nlopes/slack"
)

type commCtx struct {
	Env    config.Environment
	UserID string
	TeamID string
	AsUser bool
}

func newCommCtx(env config.Environment, userID, teamID string, asUser bool) commCtx {
	return commCtx{
		Env:    env,
		UserID: userID,
		TeamID: teamID,
		AsUser: asUser,
	}
}

func (c commCtx) postMessage(targetChannel string, msgOption slack.MsgOption) error {
	token, err := c.getToken()
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

func (c commCtx) updateMessage(targetChannel, messageTs string, msgOption slack.MsgOption) error {
	token, err := c.getToken()
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

func (c commCtx) showDialog(triggerID string, dialog slack.Dialog) error {
	token, err := c.getToken()
	if err != nil {
		return err
	}
	slackClient := slack.New(token)
	err = slackClient.OpenDialog(triggerID, dialog)
	if err != nil {
		log.Println("[ERROR] Cannot create the dialog ", err)
	}
	return err
}

func (c commCtx) getUserInfo(userID string) (slack.User, error) {
	token, err := c.getToken()
	if err != nil {
		return slack.User{}, err
	}

	slackClient := slack.New(token)
	user, err := slackClient.GetUserInfo(userID)
	if err != nil {
		log.Println("[ERROR] User info can't be retrieved - ", err)
		return slack.User{}, err
	}

	return *user, nil
}

func (c commCtx) getToken() (string, error) {
	if c.AsUser {
		return getUserToken(c.Env, c.UserID)
	} else {
		return getBotToken(c.Env, c.TeamID)
	}
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
