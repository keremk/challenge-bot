package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/keremk/challenge-bot/config"
	"github.com/keremk/challenge-bot/repo"
	"github.com/nlopes/slack"
	slackApi "github.com/nlopes/slack"
)

type requestHandler struct {
	slackClient     *slackApi.Client
	env             *config.Environment
	challengeConfig *config.ChallengeConfig
}

func (handler *requestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	response := buf.String()
	payload := strings.TrimLeft(response, "payload=")
	unescapedPayload, _ := url.QueryUnescape(payload)

	var interactionCB slack.InteractionCallback
	err = json.Unmarshal([]byte(unescapedPayload), &interactionCB)
	if err != nil {
		fmt.Println(err, unescapedPayload)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if interactionCB.Token != handler.env.VerificationToken {
		fmt.Println("Invalid token ", interactionCB.Token)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	fmt.Println(interactionCB.Submission)
	challengeDesc := createChallengeDesc(interactionCB.Submission)
	go handler.createChallenge(challengeDesc, interactionCB.State)

	w.WriteHeader(http.StatusAccepted)
}

func createChallengeDesc(input map[string]string) *config.ChallengeDesc {
	return &config.ChallengeDesc{
		CandidateName:     input["candidate_name"],
		GithubAlias:       input["github_alias"],
		ResumeURL:         input["resume_URL"],
		ChallengeTemplate: input["challenge_template"],
	}
}

func (handler *requestHandler) createChallenge(challengeDesc *config.ChallengeDesc, channel string) {
	if repo.CheckUser(challengeDesc.GithubAlias, handler.env.GithubToken) == false {
		fmt.Println("Github user does not exist")
		errorMsg := fmt.Sprintf("Github Alias %s for candidate %s is not correct", challengeDesc.GithubAlias, challengeDesc.CandidateName)
		handler.slackClient.PostMessage(channel, slack.MsgOptionText(errorMsg, false))
		return
	}

	handler.slackClient.PostMessage(channel, slack.MsgOptionText("Please be patient, while a go create a coding challenge for you...", false))

	challengeURL, err := repo.CreateChallenge(challengeDesc.GithubAlias, challengeDesc.ChallengeTemplate, *handler.challengeConfig, handler.env.GithubToken)

	if err != nil {
		fmt.Println(err)
		errorMsg := fmt.Sprintf("Unable to create challenge for %s", challengeDesc.CandidateName)
		handler.slackClient.PostMessage(channel, slack.MsgOptionText(errorMsg, false))
		return
	}
	challengeDesc.ChallengeURL = challengeURL
	handler.slackClient.PostMessage(channel, createChallengeSummary(challengeDesc))
}
