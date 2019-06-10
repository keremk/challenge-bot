package slackops

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/url"
	"strings"

	"github.com/keremk/challenge-bot/config"

	"github.com/nlopes/slack"
)

const dreadedPrivateRepoError = "422 Visibility can't be private"

func HandleRequests(env config.Environment, readCloser io.ReadCloser) error {
	icb, err := parseInteractionCallback(readCloser, env.VerificationToken)
	if err != nil {
		return err
	}

	switch icb.Type {
	case "dialog_submission":
		err = handleDialogSubmission(env, icb)
	case "block_actions":
		err = handleBlockActions(env, icb, readCloser)
	default:
		err = errors.New("[ERROR] Unknown dialog response")
		log.Println("[ERROR] Unknown dialog response - ", icb.CallbackID)
	}

	return err
}

func handleDialogSubmission(env config.Environment, icb slack.InteractionCallback) error {
	var err error

	switch icb.CallbackID {
	case "send_challenge":
		err = handleSendChallenge(env, icb)
	case "new_challenge":
		err = handleNewChallenge(env, icb)
	case "new_reviewer":
		err = handleNewReviewer(env, icb)
	case "edit_reviewer":
		err = handleEditReviewer(env, icb)
	case "schedule_update":
		err = handleShowSchedule(env, icb)
	case "find_reviewers":
		err = handleFindReviewers(env, icb)
	default:
		err = errors.New("[ERROR] Unknown CallbackID")
		log.Println("[ERROR] Unknown CallbackID - ", icb.CallbackID)
	}
	return err
}

func handleBlockActions(env config.Environment, icb slack.InteractionCallback, readCloser io.ReadCloser) error {
	var err error

	// log.Println("State of message = ", icb.State)
	// log.Printf("Message Response URL %s", icb.ResponseURL)
	// log.Printf("Block actions %s", icb.ActionCallback.BlockActions)
	// log.Printf("Action ID of first %s", icb.ActionCallback.BlockActions[0].ActionID)
	// log.Printf("Action Text of first %s", icb.ActionCallback.BlockActions[0].Text)
	// log.Printf("Action Value of first %s", icb.ActionCallback.BlockActions[0].Value)
	// log.Printf("Action Type of first %s", icb.ActionCallback.BlockActions[0].Type)
	// log.Printf("Action BlockID of first %s", icb.ActionCallback.BlockActions[0].BlockID)

	action, encodedActionInfo, err := decodeAction(icb.ActionCallback.BlockActions[0].ActionID)
	log.Printf("Action %s, Encoded ActionInfo %s", action, encodedActionInfo)
	if err != nil {
		return err
	}
	switch action {
	case scheduleUpdate:
		err = handleUpdateSchedule(env, icb, encodedActionInfo)
	case showBookings:
		fallthrough
	case findReviewers:
		err = handleBookings(env, icb, encodedActionInfo)
	default:
		err = errors.New("[ERROR] Unknown action")
		log.Println("[ERROR] Unknown action - ", action)
	}
	return err
}

func parseInteractionCallback(readCloser io.ReadCloser, verificationToken string) (slack.InteractionCallback, error) {
	payload, err := payloadContents(readCloser)
	if err != nil {
		return slack.InteractionCallback{}, err
	}

	var icb slack.InteractionCallback
	err = json.Unmarshal([]byte(payload), &icb)
	if err != nil {
		log.Println("[ERROR] Unable to unmarshall json response", err)
		return slack.InteractionCallback{}, err
	}

	if icb.Token != verificationToken {
		log.Println("[ERROR] Unable to validate request ", err)
		return slack.InteractionCallback{}, ValidationError{}
	}

	return icb, nil
}

func payloadContents(readCloser io.ReadCloser) (string, error) {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(readCloser)
	if err != nil {
		log.Println("[ERROR] Unable to read the response body ", err)
		return "", err
	}

	response := buf.String()
	payload := strings.TrimLeft(response, "payload=")
	unescapedPayload, err := url.QueryUnescape(payload)
	if err != nil {
		log.Println("[ERROR] Unable to unescape the response body ", err)
		return "", err
	}

	return unescapedPayload, nil
}

func toMsgOption(text string) slack.MsgOption {
	return slack.MsgOptionText(text, false)
}

func postMessage(env config.Environment, teamID string, targetChannel string, msgOption slack.MsgOption) error {
	token, err := getBotToken(env, teamID)
	if err != nil {
		return err
	}

	slackClient := slack.New(token)
	slackClient.PostMessage(targetChannel, msgOption)
	return nil
}
