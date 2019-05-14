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
	"github.com/keremk/challenge-bot/repo"
	"github.com/nlopes/slack"
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

	switch icb.CallbackID {
	case "send_challenge":
		err = handleSendChallenge(env, icb)
	case "new_challenge":
		err = handleNewChallenge(env, icb)
	default:
		err = errors.New("[ERROR] Unknown dialog response")
		log.Println("[ERROR] Unknown dialog response")
	}
	return err
}

func handleSendChallenge(env config.Environment, icb slack.InteractionCallback) error {
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
	// slackActionCtx := newSlackActionContext(teamID, env)
	go sendChallenge(env, challenge, candidate, returnChannel, teamID)

	return nil
}

func sendChallenge(env config.Environment, challenge models.Challenge, candidate models.Candidate, targetChannel, teamID string) {
	repoCtx := repo.NewActionContext(env)

	if repoCtx.CheckUser(candidate.GithubAlias) == false {
		errorMsg := fmt.Sprintf("Github Alias %s for candidate %s is not correct", candidate.GithubAlias, candidate.Name)
		postMessage(env, teamID, targetChannel, toMsgOption(errorMsg))
		return
	}

	postMessage(env, teamID, targetChannel, toMsgOption("Please be patient, while I go create a coding challenge for you..."))

	challengeURL, err := repoCtx.CreateChallenge(candidate, challenge)

	if err != nil {
		log.Println("[ERROR] Create challenge failed: ", err)
		errorMsg := fmt.Sprintf("Unable to create challenge for %s", candidate.Name)
		postMessage(env, teamID, targetChannel, toMsgOption(errorMsg))
		return
	}
	postMessage(env, teamID, targetChannel, newChallengeSummary(candidate, challengeURL, challenge.TrackingIssuesURL()))
}

func handleNewChallenge(env config.Environment, icb slack.InteractionCallback) error {
	challengeInput := icb.Submission
	challengeInput["team_id"] = icb.Team.ID
	challenge := models.NewChallenge(icb.Submission)
	err := models.UpdateChallenge(env, challenge)
	if err != nil {
		log.Println("[ERROR] Could not update challenge in db ", err)
		_ = postMessage(env, icb.Team.ID, icb.Channel.ID, toMsgOption("We were not able to create the new challenge"))
		return err
	}

	msgText := fmt.Sprintf("We created a challenge named %s in our database. It is pointing to: %s", challenge.Name, challenge.TemplateRepositoryURL())
	_ = postMessage(env, icb.Team.ID, icb.Channel.ID, toMsgOption(msgText))
	return nil
}

func toMsgOption(text string) slack.MsgOption {
	return slack.MsgOptionText(text, false)
}

func postMessage(env config.Environment, teamID string, targetChannel string, msgOption slack.MsgOption) error {
	token, err := getBotToken(env, teamID)
	if err != nil {
		return err
	}

	slackClient := slackApi.New(token)
	slackClient.PostMessage(targetChannel, msgOption)
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
