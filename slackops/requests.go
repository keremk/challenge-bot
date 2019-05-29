package slackops

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"strings"

	"github.com/keremk/challenge-bot/config"

	"github.com/nlopes/slack"
)

const dreadedPrivateRepoError = "422 Visibility can't be private"

type dialogState struct {
	channelID     string
	challengeName string
}

func stateFromString(s string) (dialogState, error) {
	x := strings.Split(s, ",")
	if len(x) < 2 {
		return dialogState{}, errors.New("[ERROR] state persisted incorrectly")
	}

	return dialogState{
		channelID:     x[0],
		challengeName: x[1],
	}, nil
}

func (d dialogState) string() string {
	return fmt.Sprintf("%s,%s", d.channelID, d.challengeName)
}

func HandleRequests(env config.Environment, readCloser io.ReadCloser) error {
	icb, err := parseInteractionCallback(readCloser, env.VerificationToken)
	if err != nil {
		return err
	}

	switch icb.CallbackID {
	case "send_challenge":
		err = handleSendChallenge(env, icb)
	case "new_challenge":
		err = handleNewChallenge(env, icb)
	case "new_reviewer":
		err = handleNewReviewer(env, icb)
	default:
		err = errors.New("[ERROR] Unknown dialog response")
		log.Println("[ERROR] Unknown dialog response")
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