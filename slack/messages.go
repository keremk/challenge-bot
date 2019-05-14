package slack

import (
	"fmt"

	"github.com/keremk/challenge-bot/models"
	"github.com/nlopes/slack"
)

func newChallengeSummary(candidate models.Candidate, challengeURL string, trackingIssuesURL string) slack.MsgOption {
	// Header Section
	headerText := fmt.Sprintf("You have created a new coding challenge at:\n*<%s|%s>*", challengeURL, challengeURL)
	headerTextBlock := slack.NewTextBlockObject("mrkdwn", headerText, false, false)
	headerSection := slack.NewSectionBlock(headerTextBlock, nil, nil)

	// Fields
	candidateNameText := fmt.Sprintf("*Candidate Name:*\n<%s|%s>", candidate.ResumeURL, candidate.Name)
	candidateNameBlock := slack.NewTextBlockObject("mrkdwn", candidateNameText, false, false)
	githubAliasText := fmt.Sprintf("*Github Alias:*\n%s", candidate.GithubAlias)
	githubAliasBlock := slack.NewTextBlockObject("mrkdwn", githubAliasText, false, false)

	fieldSlice := make([]*slack.TextBlockObject, 0)
	fieldSlice = append(fieldSlice, candidateNameBlock)
	fieldSlice = append(fieldSlice, githubAliasBlock)
	fieldsSection := slack.NewSectionBlock(nil, fieldSlice, nil)

	// Footer Section
	footerText := fmt.Sprintf("You can track coding challenges at <%s>", trackingIssuesURL)
	footerBlock := slack.NewTextBlockObject("mrkdwn", footerText, false, false)
	footerSection := slack.NewSectionBlock(footerBlock, nil, nil)

	return slack.MsgOptionBlocks(
		headerSection,
		fieldsSection,
		footerSection,
	)
}
