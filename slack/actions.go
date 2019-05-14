package slack

import (
	"fmt"
	"log"

	"github.com/keremk/challenge-bot/config"
	"github.com/keremk/challenge-bot/models"
	"github.com/keremk/challenge-bot/repo"
	"github.com/nlopes/slack"
	slackApi "github.com/nlopes/slack"
)

type slackActionContext struct {
	env     config.Environment
	repoCtx repo.ActionContext
	teamID  string
}

func newSlackActionContext(teamID string, env config.Environment) slackActionContext {
	return slackActionContext{
		env:     env,
		repoCtx: repo.NewActionContext(env),
		teamID:  teamID,
	}
}

func (s slackActionContext) createChallenge(challenge models.Challenge, candidate models.Candidate, targetChannel string) {
	token, err := getBotToken(s.env, s.teamID)
	if err != nil {
		return
	}

	slackClient := slackApi.New(token)

	if s.repoCtx.CheckUser(candidate.GithubAlias) == false {
		errorMsg := fmt.Sprintf("Github Alias %s for candidate %s is not correct", candidate.GithubAlias, candidate.Name)
		slackClient.PostMessage(targetChannel, slack.MsgOptionText(errorMsg, false))
		return
	}

	slackClient.PostMessage(targetChannel, slack.MsgOptionText("Please be patient, while I go create a coding challenge for you...", false))

	challengeURL, err := s.repoCtx.CreateChallenge(candidate, challenge)

	if err != nil {
		log.Println("[ERROR] Create challenge failed: ", err)
		errorMsg := fmt.Sprintf("Unable to create challenge for %s", candidate.Name)
		slackClient.PostMessage(targetChannel, slack.MsgOptionText(errorMsg, false))
		return
	}
	slackClient.PostMessage(targetChannel, newChallengeSummary(candidate, challengeURL, challenge.TrackingIssuesURL()))
}
