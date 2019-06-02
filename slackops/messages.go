package slackops

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/keremk/challenge-bot/models"
	"github.com/keremk/challenge-bot/scheduling"
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

func renderSchedule(weekNo, year int, reviewer models.Reviewer, slots []scheduling.SlotInfo) slack.ActionBlock {
	// Schedule Action Blocks
	blockEls := make([]slack.BlockElement, 0, len(slots))
	for _, slot := range slots {
		var buttonText string
		if slot.IsSelected {
			buttonText = fmt.Sprintf("\u2713 %s : %s - %s", slot.Slot.Day, slot.Slot.StartTime, slot.Slot.EndTime)
		} else {
			buttonText = fmt.Sprintf("\u2717 %s : %s - %s", slot.Slot.Day, slot.Slot.StartTime, slot.Slot.EndTime)
		}
		buttonTextBlock := slack.NewTextBlockObject("plain_text", buttonText, false, false)
		encodedValue := strconv.FormatBool(slot.IsSelected)
		encodedID := encodeScheduleActionInfo(scheduleActionInfo{
			SlotID:     slot.Slot.ID,
			ReviewerID: reviewer.SlackID,
			WeekNo:     weekNo,
			Year:       year,
		})
		blockEl := slack.NewButtonBlockElement(encodedID, encodedValue, buttonTextBlock)
		blockEls = append(blockEls, blockEl)
	}

	slotsBlock := newActionBlock("interview_slots", blockEls)
	return slotsBlock
}

func newActionBlock(blockID string, elements []slack.BlockElement) slack.ActionBlock {
	return slack.ActionBlock{
		Type:    slack.MBTAction,
		BlockID: blockID,
		Elements: slack.BlockElements{
			ElementSet: elements,
		},
	}
}

func sendDelayedResponse(url string, json string) error {
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	body := strings.NewReader(json)
	_, err := client.Post(url, "application/json", body)
	if err != nil {
		log.Println("[ERROR] Failed to send delayed response")
		return err
	}
	return nil
}
