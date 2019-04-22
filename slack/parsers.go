package slack

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/keremk/challenge-bot/models"
	slackApi "github.com/nlopes/slack"
)

type ValidationError struct{}

func (e ValidationError) Error() string {
	return "Invalid validation token received from Slack server"
}

type ResponseParser struct {
	VerificationToken string
}

func (r ResponseParser) DialogResponseParse(readCloser io.ReadCloser) (*models.ChallengeDesc, string, error) {
	payload, err := payloadContents(readCloser)
	if err != nil {
		return nil, "", err
	}

	var icb slackApi.InteractionCallback
	err = json.Unmarshal([]byte(payload), &icb)
	if err != nil {
		return nil, "", err
	}

	if icb.Token != r.VerificationToken {
		return nil, "", ValidationError{}
	}

	return models.NewChallengeDesc(icb.Submission), icb.State, nil
}

func (r ResponseParser) SlashCommandParse(input *http.Request) (*slackApi.SlashCommand, error) {
	s, err := slackApi.SlashCommandParse(input)
	if err != nil {
		return nil, err
	}

	if !s.ValidateToken(r.VerificationToken) {
		return nil, ValidationError{}
	}
	return &s, nil
}

func payloadContents(readCloser io.ReadCloser) (string, error) {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(readCloser)
	if err != nil {
		return "", err
	}

	response := buf.String()
	payload := strings.TrimLeft(response, "payload=")
	unescapedPayload, err := url.QueryUnescape(payload)
	if err != nil {
		return "", err
	}

	return unescapedPayload, nil
}
