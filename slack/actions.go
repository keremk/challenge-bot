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
	repoCtx *repo.ActionContext
	teamID  string
}

func newSlackActionContext(challengeConfig *config.ChallengeConfig, teamID string, env config.Environment) *slackActionContext {
	return &slackActionContext{
		env:     env,
		repoCtx: repo.NewActionContext(challengeConfig),
		teamID:  teamID,
	}
}

func (s slackActionContext) createChallenge(challengeDesc models.ChallengeDesc, targetChannel string) {
	token, err := getBotToken(s.env, s.teamID)
	if err != nil {
		return
	}

	slackClient := slackApi.New(token)

	if s.repoCtx.CheckUser(challengeDesc.GithubAlias) == false {
		errorMsg := fmt.Sprintf("Github Alias %s for candidate %s is not correct", challengeDesc.GithubAlias, challengeDesc.CandidateName)
		slackClient.PostMessage(targetChannel, slack.MsgOptionText(errorMsg, false))
		return
	}

	slackClient.PostMessage(targetChannel, slack.MsgOptionText("Please be patient, while I go create a coding challenge for you...", false))

	challengeURL, err := s.repoCtx.CreateChallenge(challengeDesc)

	if err != nil {
		log.Println("[ERROR] Create challenge failed: ", err)
		errorMsg := fmt.Sprintf("Unable to create challenge for %s", challengeDesc.CandidateName)
		slackClient.PostMessage(targetChannel, slack.MsgOptionText(errorMsg, false))
		return
	}
	slackClient.PostMessage(targetChannel, newChallengeSummary(challengeDesc, challengeURL))
}
