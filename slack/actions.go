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
	repoCtx     *repo.ActionContext
	slackClient *slackApi.Client
}

func newSlackActionContext(challengeConfig *config.ChallengeConfig, slackClient *slackApi.Client) *slackActionContext {
	return &slackActionContext{
		repoCtx:     repo.NewActionContext(challengeConfig),
		slackClient: slackClient,
	}
}

func (ctx slackActionContext) createChallenge(challengeDesc *models.ChallengeDesc, targetChannel string) {
	if ctx.repoCtx.CheckUser(challengeDesc.GithubAlias) == false {
		errorMsg := fmt.Sprintf("Github Alias %s for candidate %s is not correct", challengeDesc.GithubAlias, challengeDesc.CandidateName)
		ctx.slackClient.PostMessage(targetChannel, slack.MsgOptionText(errorMsg, false))
		return
	}

	ctx.slackClient.PostMessage(targetChannel, slack.MsgOptionText("Please be patient, while I go create a coding challenge for you...", false))

	challengeURL, err := ctx.repoCtx.CreateChallenge(challengeDesc.GithubAlias, challengeDesc.ChallengeTemplate)

	if err != nil {
		log.Println("[ERROR] Create challenge failed: ", err)
		errorMsg := fmt.Sprintf("Unable to create challenge for %s", challengeDesc.CandidateName)
		ctx.slackClient.PostMessage(targetChannel, slack.MsgOptionText(errorMsg, false))
		return
	}
	challengeDesc.ChallengeURL = challengeURL
	ctx.slackClient.PostMessage(targetChannel, newChallengeSummary(challengeDesc))
}
