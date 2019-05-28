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

	// log.Println("[INFO] ICB state is - ", icb.State)
	switch icb.CallbackID {
	case "send_challenge":
		respJSON, err := handleSendChallengeOptions(env, icb)
		return respJSON, err
	case "new_challenge":
		respJSON, err := handleNewChallengeOptions(env, icb)
		return respJSON, err
	case "new_reviewer":
		respJSON, err := handleNewReviewerOptions(env, icb)
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
		state, err := stateFromString(icb.State)
		if err != nil {
			panic("Unknown error handling state")
		}
		js, err := getReviewerList(env, state.challengeName)
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

func getReviewerList(env config.Environment, challengeName string) ([]byte, error) {
	var reviewerList []models.Reviewer
	var err error
	if challengeName == "" {
		reviewerList, err = models.GetAllReviewers(env)
	} else {
		reviewerList, err = models.GetAllReviewersForChallenge(env, challengeName)
	}
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

func handleNewReviewerOptions(env config.Environment, icb slack.InteractionCallback) ([]byte, error) {
	switch icb.Name {
	case "challenge_name":
		js, err := getChallengeList(env)
		if err != nil {
			return nil, err
		}
		return js, nil
	default:
		return nil, nil
	}
}