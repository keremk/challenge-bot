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
		encodedScheduleAction := encodeScheduleActionInfo(scheduleActionInfo{
			SlotID:     slot.Slot.ID,
			ReviewerID: reviewer.SlackID,
			WeekNo:     weekNo,
			Year:       year,
		})
		encodedID := encodeAction(scheduleUpdate, encodedScheduleAction)
		blockEl := slack.NewButtonBlockElement(encodedID, encodedValue, buttonTextBlock)
		blockEls = append(blockEls, blockEl)
	}

	slotsBlock := newActionBlock("interview_slots", blockEls)
	return slotsBlock
}

func renderReviewers(weekNo, year int, slots map[string]*scheduling.SlotAvailability) slack.MsgOption {
	sections := make([]slack.Block, 0, 50)
	weekDescription, err := scheduling.WeekDescriptionFromWeekNo(weekNo, year)
	if err != nil {
		log.Println("[ERROR] Week number not valid", err)
	}
	headerText := fmt.Sprintf("*Interviewer List For Week:* %s ", weekDescription)
	headerEl := slack.NewTextBlockObject("mrkdwn", headerText, false, false)
	sections = append(sections, slack.NewSectionBlock(headerEl, nil, nil))

	for slotID, slotAvailability := range slots {
		slotHeaderText := fmt.Sprintf("*Interview Slot:* %s ", slotAvailability.Slot.Name)
		slotHeaderEl := slack.NewTextBlockObject("mrkdwn", slotHeaderText, false, false)
		sections = append(sections, slack.NewSectionBlock(slotHeaderEl, nil, nil))
		for _, reviewerInfo := range slotAvailability.Reviewers {
			sections = append(sections, renderReviewer(reviewerInfo, slotID, weekNo, year))
		}
	}
	return slack.MsgOptionBlocks(sections...)
}

func renderReviewer(reviewerInfo scheduling.ReviewerInfo, slotID string, weekNo, year int) *slack.SectionBlock {
	reviewer := reviewerInfo.Reviewer
	isBooked := reviewerInfo.IsBooked

	reviewerNameText := fmt.Sprintf("*<@%s|%s>* (%s)", reviewer.SlackID, reviewer.Name, reviewer.TechnologyList)
	reviewerNameEl := slack.NewTextBlockObject("mrkdwn", reviewerNameText, false, false)

	encodedScheduleAction := encodeScheduleActionInfo(scheduleActionInfo{
		SlotID:     slotID,
		WeekNo:     weekNo,
		Year:       year,
		ReviewerID: reviewer.SlackID,
	})

	var buttonText string
	if isBooked {
		buttonText = "Unbook"
	} else {
		buttonText = "Book"
	}
	buttonTextBlock := slack.NewTextBlockObject("plain_text", buttonText, false, false)
	encodedID := encodeAction(findReviewers, encodedScheduleAction)
	encodedValue := strconv.FormatBool(isBooked)

	buttonEl := slack.NewButtonBlockElement(encodedID, encodedValue, buttonTextBlock)

	accessory := slack.NewAccessory(buttonEl)
	return slack.NewSectionBlock(reviewerNameEl, nil, accessory)
}

func renderBookings(reviewer models.Reviewer, challenge models.ChallengeSetup) []slack.Block {
	bookings := reviewer.Bookings

	sections := make([]slack.Block, 0, 50)
	reviewerNameText := fmt.Sprintf("All bookings for *<@%s>* (%s)", reviewer.SlackID, reviewer.TechnologyList)
	reviewerNameEl := slack.NewTextBlockObject("mrkdwn", reviewerNameText, false, false)
	sections = append(sections, slack.NewSectionBlock(reviewerNameEl, nil, nil))

	currentYear, currentWeek := time.Now().ISOWeek()
	noBookingsFound := true
	for weekInfo, bookingsPerWeek := range bookings {
		if len(bookingsPerWeek) == 0 {
			continue
		}
		weekNo, year := decodeWeekAndYear(weekInfo)
		if weekNo < currentWeek || year < currentYear {
			continue
		}
		weekDescription, err := scheduling.WeekDescriptionFromWeekNo(weekNo, year)
		if err != nil {
			log.Println("[ERROR] Week description is not retrieved")
			weekDescription = fmt.Sprintf("Week # %d - ", weekNo)
		}
		weekDescriptionEl := slack.NewTextBlockObject("mrkdwn", weekDescription, false, false)
		sections = append(sections, slack.NewSectionBlock(weekDescriptionEl, nil, nil))

		for _, bookingKey := range bookingsPerWeek {
			bookingSlot := challenge.Slots[bookingKey]
			sections = append(sections, renderBooking(*bookingSlot, weekNo, year, reviewer))
		}
		noBookingsFound = false
	}
	if noBookingsFound {
		noBookingsFoundEl := slack.NewTextBlockObject("mrkdwn", "No bookings found.", false, false)
		sections = append(sections, slack.NewSectionBlock(noBookingsFoundEl, nil, nil))
	}
	return sections
}

func renderBooking(slot models.Slot, weekNo, year int, reviewer models.Reviewer) *slack.SectionBlock {
	slotDescriptionText := fmt.Sprintf("*Slot*: %s", slot.Name)
	slotDescriptionEl := slack.NewTextBlockObject("mrkdwn", slotDescriptionText, false, false)

	encodedScheduleAction := encodeScheduleActionInfo(scheduleActionInfo{
		SlotID:     slot.ID,
		WeekNo:     weekNo,
		Year:       year,
		ReviewerID: reviewer.SlackID,
	})

	buttonTextBlock := slack.NewTextBlockObject("plain_text", "Unbook", false, false)
	encodedID := encodeAction(showBookings, encodedScheduleAction)
	encodedValue := strconv.FormatBool(true)

	buttonEl := slack.NewButtonBlockElement(encodedID, encodedValue, buttonTextBlock)

	accessory := slack.NewAccessory(buttonEl)
	return slack.NewSectionBlock(slotDescriptionEl, nil, accessory)
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
