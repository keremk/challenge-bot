package slack

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/url"
	"strings"

	"github.com/keremk/challenge-bot/config"
	"github.com/keremk/challenge-bot/models"
	slackApi "github.com/nlopes/slack"
)

func HandleDialogResponse(env config.Environment, readCloser io.ReadCloser, challenge *config.ChallengeConfig) error {
	icb, err := parseChallengeStart(readCloser, env.VerificationToken)
	if err != nil {
		return err
	}

	challengeDesc := models.NewChallengeDesc(icb.Submission)
	returnChannel := icb.State
	teamID := icb.Team.ID

	slackActionCtx := newSlackActionContext(challenge, teamID, env)

	go slackActionCtx.createChallenge(challengeDesc, returnChannel)

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

func newChallengeOptionsDialog(triggerID string, channelID string, options []string) *slackApi.Dialog {
	candidateNameElement := slackApi.NewTextInput("candidate_name", "Candidate Name", "")
	githubNameElement := slackApi.NewTextInput("github_alias", "Github Alias", "")
	resumeURLElement := slackApi.NewTextInput("resume_URL", "Resume URL", "")
	selectOptions := make([]slackApi.DialogSelectOption, len(options))
	for i, v := range options {
		selectOptions[i] = slackApi.DialogSelectOption{
			Label: v,
			Value: v,
		}
	}
	disciplineSelectElement := slackApi.NewStaticSelectDialogInput("challenge_template", "Challenge Template", selectOptions)

	elements := []slackApi.DialogElement{
		candidateNameElement,
		githubNameElement,
		resumeURLElement,
		disciplineSelectElement,
	}

	return &slackApi.Dialog{
		TriggerID:      triggerID,
		CallbackID:     "challenge_67e2d0",
		Title:          "Create Coding Challenge",
		SubmitLabel:    "Create",
		NotifyOnCancel: false,
		State:          channelID,
		Elements:       elements,
	}
}
