package slack

import (
	"io"
	"net/url"
	"strings"
	"testing"

	"github.com/keremk/challenge-bot/models"

	"github.com/stretchr/testify/assert"
)

type MockReadCloser struct {
	*strings.Reader
}

func (m *MockReadCloser) Close() error {
	return nil
}

func MockDialogResponse() io.ReadCloser {
	jsonPayload := `
{
	"type": "dialog_submission",
	"submission": {
			"candidate_name": "Sigourney Dreamweaver",
			"github_alias": "sigdre",
			"resume_URL": "https://example.com",
			"challenge_template": "android_repo"
	},
	"callback_id": "foo_1138b",
	"state": "coding_challenge_channel",
	"team": {
			"id": "T1ABCD2E12",
			"domain": "coverbands"
	},
	"user": {
			"id": "W12A3BCDEF",
			"name": "dreamweaver"
	},
	"channel": {
			"id": "C1AB2C3DE",
			"name": "coverthon-1999"
	},
	"action_ts": "936893340.702759",
	"token": "M1AqUUw3FqayAbqNtsGMch72",
	"response_url": "https://hooks.slack.com/app/T012AB0A1/123456789/JpmK0yzoZDeRiqfeduTBYXWQ"
}		
`

	jsonEscaped := url.QueryEscape(jsonPayload)
	output := "payload=" + jsonEscaped
	reader := strings.NewReader(output)

	return &MockReadCloser{reader}
}

func TestDialogResponseParser(t *testing.T) {
	expectedDesc := models.ChallengeDesc{
		CandidateName:     "Sigourney Dreamweaver",
		GithubAlias:       "sigdre",
		ResumeURL:         "https://example.com",
		ChallengeTemplate: "android_repo",
	}

	icb, err := parseChallengeStart(MockDialogResponse(), "M1AqUUw3FqayAbqNtsGMch72")
	challengeDesc := models.NewChallengeDesc(icb.Submission)
	returnChannel := icb.State
	teamID := icb.Team.ID

	assert.Nil(t, err, "Unexpected error")
	assert.Equal(t, "coding_challenge_channel", returnChannel, "Target channel not correct.")
	assert.Equal(t, "T1ABCD2E12", teamID, "Team ID is not correct.")
	assert.Equal(t, expectedDesc, challengeDesc, "Challenge description not correct.")
}

func TestDialogResponseFailedVerification(t *testing.T) {

	_, err := parseChallengeStart(MockDialogResponse(), "InvalidToken")
	assert.NotNil(t, err, "Error expected")
	assert.IsType(t, ValidationError{}, err, "Validation Error expected")
}
