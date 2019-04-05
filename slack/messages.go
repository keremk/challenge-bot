package slack

import (
	"fmt"

	"github.com/keremk/challenge-bot/config"
	"github.com/nlopes/slack"
)

func createChallengeSummary(challengeDesc *config.ChallengeDesc) slack.MsgOption {
	// Header Section
	headerText := fmt.Sprintf("You have created a new coding challenge at:\n*<%s|%s>*", challengeDesc.ChallengeURL, challengeDesc.ChallengeURL)
	headerTextBlock := slack.NewTextBlockObject("mrkdwn", headerText, false, false)
	headerSection := slack.NewSectionBlock(headerTextBlock, nil, nil)

	// Fields
	candidateNameText := fmt.Sprintf("*Candidate Name:*\n<%s|%s>", challengeDesc.ResumeURL, challengeDesc.CandidateName)
	candidateNameBlock := slack.NewTextBlockObject("mrkdwn", candidateNameText, false, false)
	githubAliasText := fmt.Sprintf("*Github Alias:*\n%s", challengeDesc.GithubAlias)
	githubAliasBlock := slack.NewTextBlockObject("mrkdwn", githubAliasText, false, false)
	challengeTemplateText := fmt.Sprintf("*Challenge Template:*\n%s", challengeDesc.ChallengeTemplate)
	challengeTemplateBlock := slack.NewTextBlockObject("mrkdwn", challengeTemplateText, false, false)

	fieldSlice := make([]*slack.TextBlockObject, 0)
	fieldSlice = append(fieldSlice, candidateNameBlock)
	fieldSlice = append(fieldSlice, githubAliasBlock)
	fieldSlice = append(fieldSlice, challengeTemplateBlock)
	fieldsSection := slack.NewSectionBlock(nil, fieldSlice, nil)

	// Footer Section
	footerText := fmt.Sprintf("You can track coding challenges <https://github.com/xing/coding-challenges/projects/1>")
	footerBlock := slack.NewTextBlockObject("mrkdwn", footerText, false, false)
	footerSection := slack.NewSectionBlock(footerBlock, nil, nil)

	return slack.MsgOptionBlocks(
		headerSection,
		fieldsSection,
		footerSection,
	)
}
