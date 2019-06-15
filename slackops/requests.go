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

type request struct {
	ctx commCtx
	icb *slack.InteractionCallback
}

func newRequest(env config.Environment, icb *slack.InteractionCallback) request {
	ctx := newCommCtx(env, icb.User.ID, icb.Team.ID, false)

	return request{
		ctx: ctx,
		icb: icb,
	}
}

func HandleRequests(env config.Environment, readCloser io.ReadCloser) error {
	icb, err := parseInteractionCallback(readCloser, env.VerificationToken)
	if err != nil {
		return err
	}

	r := newRequest(env, icb)

	switch icb.Type {
	case "dialog_submission":
		err = r.handleDialogSubmission()
	case "block_actions":
		err = r.handleBlockActions()
	default:
		err = errors.New("[ERROR] Unknown dialog response")
		log.Println("[ERROR] Unknown dialog response - ", icb.CallbackID)
	}

	return err
}

func (r request) handleDialogSubmission() error {
	var err error

	switch r.icb.CallbackID {
	case "send_challenge":
		err = r.handleSendChallenge()
	case "new_challenge":
		err = r.handleNewChallenge()
	case "new_reviewer":
		err = r.handleNewReviewer()
	case "edit_reviewer":
		err = r.handleEditReviewer()
	case "schedule_update":
		err = r.handleShowSchedule()
	case "find_reviewers":
		err = r.handleFindReviewers()
	default:
		err = errors.New("[ERROR] Unknown CallbackID")
		log.Println("[ERROR] Unknown CallbackID - ", r.icb.CallbackID)
	}
	return err
}

func (r request) handleBlockActions() error {
	var err error

	action, encodedActionInfo, err := decodeAction(r.icb.ActionCallback.BlockActions[0].ActionID)
	//log.Printf("Action %s, Encoded ActionInfo %s", action, encodedActionInfo)
	if err != nil {
		return err
	}
	switch action {
	case scheduleUpdate:
		err = r.handleUpdateSchedule(encodedActionInfo)
	case showBookings:
		fallthrough
	case findReviewers:
		err = r.handleBookings(encodedActionInfo)
	default:
		err = errors.New("[ERROR] Unknown action")
		log.Println("[ERROR] Unknown action - ", action)
	}
	return err
}

func parseInteractionCallback(readCloser io.ReadCloser, verificationToken string) (*slack.InteractionCallback, error) {
	payload, err := payloadContents(readCloser)
	if err != nil {
		return nil, err
	}

	var icb slack.InteractionCallback
	err = json.Unmarshal([]byte(payload), &icb)
	if err != nil {
		log.Println("[ERROR] Unable to unmarshall json response", err)
		return nil, err
	}

	if icb.Token != verificationToken {
		log.Println("[ERROR] Unable to validate request ", err)
		return nil, ValidationError{}
	}

	return &icb, nil
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
