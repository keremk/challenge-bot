package slack

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
	"github.com/keremk/challenge-bot/models"
	slackApi "github.com/nlopes/slack"
)

type dialogState struct {
	channelID    string
	settingsName string
}

func stateFromString(s string) (dialogState, error) {
	x := strings.Split(s, ",")
	if len(x) < 2 {
		return dialogState{}, errors.New("[ERROR] state persisted incorrectly")
	}

	return dialogState{
		channelID:    x[0],
		settingsName: x[1],
	}, nil
}

func (d dialogState) string() string {
	return fmt.Sprintf("%s,%s", d.channelID, d.settingsName)
}

func HandleDialogResponse(env config.Environment, readCloser io.ReadCloser) error {
	icb, err := parseChallengeStart(readCloser, env.VerificationToken)
	if err != nil {
		return err
	}

	candidate := models.NewCandidate(icb.Submission)
	state, err := stateFromString(icb.State)
	if err != nil {
		return err
	}
	returnChannel := state.channelID
	teamID := icb.Team.ID

	challenge, err := models.GetChallenge(env, candidate.ChallengeName)
	if err != nil {
		return err
	}
	slackActionCtx := newSlackActionContext(teamID, env)
	go slackActionCtx.createChallenge(challenge, candidate, returnChannel)

	return nil
}

func parseChallengeStart(readCloser io.ReadCloser, verificationToken string) (slackApi.InteractionCallback, error) {
	payload, err := payloadContents(readCloser)
	if err != nil {
		return slackApi.InteractionCallback{}, err
	}

	var icb slackApi.InteractionCallback
	err = json.Unmarshal([]byte(payload), &icb)
	if err != nil {
		log.Println("[ERROR] Unable to unmarshall json response", err)
		return slackApi.InteractionCallback{}, err
	}

	if icb.Token != verificationToken {
		log.Println("[ERROR] Unable to validate request ", err)
		return slackApi.InteractionCallback{}, ValidationError{}
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
