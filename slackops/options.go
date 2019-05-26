package slackops

import (
	"encoding/json"
	"errors"
	"io"
	"log"

	"github.com/keremk/challenge-bot/config"
	"github.com/keremk/challenge-bot/models"
	"github.com/nlopes/slack"
)

type option struct {
	Label string `json:"label,omitempty"`
	Value string `json:"value,omitempty"`
}

type options struct {
	Options []option `json:"options,omitempty"`
}

func HandleOptions(env config.Environment, readCloser io.ReadCloser) ([]byte, error) {
	icb, err := parseInteractionCallback(readCloser, env.VerificationToken)
	if err != nil {
		return nil, err
	}

	switch icb.CallbackID {
	case "send_challenge":
		respJSON, err := handleSendChallengeOptions(env, icb)
		return respJSON, err
	case "new_challenge":
		respJSON, err := handleNewChallengeOptions(env, icb)
		return respJSON, err
	default:
		err = errors.New("[ERROR] Unknown dialog response")
		log.Println("[ERROR] Unknown dialog response")
	}
	return nil, err
}

func handleSendChallengeOptions(env config.Environment, icb slack.InteractionCallback) ([]byte, error) {
	switch icb.Name {
	case "challenge_name":
		js, err := getChallengeList(env)
		if err != nil {
			return nil, err
		}
		return js, nil
	case "reviewer1_id":
		fallthrough
	case "reviewer2_id":
		js, err := getReviewerList(env)
		if err != nil {
			return nil, err
		}
		return js, nil
	default:
		return nil, nil
	}
}

func getChallengeList(env config.Environment) ([]byte, error) {
	challengeList, err := models.GetAllChallenges(env)
	if err != nil {
		return nil, err
	}
	optionList := make([]option, 0, len(challengeList))
	for _, challenge := range challengeList {
		optionList = append(optionList, option{
			Label: challenge.Name,
			Value: challenge.Name,
		})
	}

	options := options{
		Options: optionList,
	}

	js, err := json.Marshal(options)
	if err != nil {
		return nil, err
	}
	return js, nil
}

func getReviewerList(env config.Environment) ([]byte, error) {
	reviewerList, err := models.GetAllReviewers(env)
	if err != nil {
		return nil, err
	}
	optionList := make([]option, 0, len(reviewerList))
	for _, reviewer := range reviewerList {
		optionList = append(optionList, option{
			Label: reviewer.Name,
			Value: reviewer.ID,
		})
	}

	options := options{
		Options: optionList,
	}

	js, err := json.Marshal(options)
	if err != nil {
		return nil, err
	}
	return js, nil
}

func handleNewChallengeOptions(env config.Environment, icb slack.InteractionCallback) ([]byte, error) {
	switch icb.Name {
	case "github_account":
		js, err := getAccountsList(env)
		if err != nil {
			return nil, err
		}
		return js, nil
	default:
		return nil, nil
	}
}

func getAccountsList(env config.Environment) ([]byte, error) {
	accountList, err := models.GetAllAccounts(env)
	if err != nil {
		return nil, err
	}
	optionList := make([]option, 0, len(accountList))
	for _, account := range accountList {
		optionList = append(optionList, option{
			Label: account.Name,
			Value: account.Name,
		})
	}

	options := options{
		Options: optionList,
	}

	js, err := json.Marshal(options)
	if err != nil {
		return nil, err
	}
	return js, nil
}
