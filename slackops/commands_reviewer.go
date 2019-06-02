package slackops

import (
	"fmt"
	"regexp"
	"time"

	"github.com/keremk/challenge-bot/config"
	"github.com/nlopes/slack"
)

func executeReviewerHelp(c command) error {
	helpText := `
{
	"blocks": [
		{
			"type": "section", 
			"text": {
				"type": "mrkdwn",
				"text": "Hello and welcome to the coding challenge tool. You can use the following commands:"
			} 
		},
		{
			"type": "section",
			"text": {
				"type": "mrkdwn",
				"text": "*/reviewer help* : Displays this message"
			}
		}, 
		{
			"type": "section",
			"text": {
				"type": "mrkdwn",
				"text": "*/reviewer new* : Opens a dialog to register a reviewer"
			}
		}
	]
}
`
	err := sendDelayedResponse(c.slashCommand.ResponseURL, helpText)
	return err
}

func executeNewReviewer(env config.Environment, c command) error {
	s := c.slashCommand

	// Create the dialog and send a message to open it
	state := dialogState{
		channelID: s.ChannelID,
		argument:  c.arg,
	}
	dialog := newAddReviewerDialog(s.TriggerID, state)

	return showDialog(env, s.TeamID, s.TriggerID, dialog)
}

func newAddReviewerDialog(triggerID string, state dialogState) slack.Dialog {
	reviewerEl := slack.NewUsersSelect("reviewer_id", "Reviewer")
	githubNameEl := slack.NewTextInput("github_alias", "Github Alias", "")
	challengeNameEl := newExternalOptionsDialogInput("challenge_name", "Challenge Name", "", false)
	technologyListEl := slack.NewTextInput("technology_list", "Technology List", "")
	elements := []slack.DialogElement{
		reviewerEl,
		githubNameEl,
		challengeNameEl,
		technologyListEl,
	}
	return slack.Dialog{
		TriggerID:      triggerID,
		CallbackID:     "new_reviewer",
		Title:          "Add Reviewer",
		SubmitLabel:    "Add",
		NotifyOnCancel: false,
		State:          state.string(),
		Elements:       elements,
	}
}

func executeSchedule(env config.Environment, c command) error {
	s := c.slashCommand

	var reviewer string
	if c.arg == "" {
		reviewer = s.UserID
	} else {
		reviewer = parseSlackIDFromString(c.arg)
	}
	// Create the dialog and send a message to open it
	state := dialogState{
		channelID: s.ChannelID,
		argument:  reviewer,
	}

	dialog := newScheduleDialog(s.TriggerID, state)

	return showDialog(env, s.TeamID, s.TriggerID, dialog)
}

func newScheduleDialog(triggerID string, state dialogState) slack.Dialog {
	// allOrSelectWeekEl := newStaticOptionsDialogInput("all_or_select", "Entire Schedule or By Week", false, allOrSelectWeekOptions())
	weekOfYearEl := newStaticOptionsDialogInput("year_week", "Week of the Year", true, weekOfYearOptions())

	elements := []slack.DialogElement{
		// allOrSelectWeekEl,
		weekOfYearEl,
	}
	return slack.Dialog{
		TriggerID:      triggerID,
		CallbackID:     "schedule_update",
		Title:          "Update Schedule",
		SubmitLabel:    "Update",
		NotifyOnCancel: false,
		State:          state.string(),
		Elements:       elements,
	}
}

// func allOrSelectWeekOptions() []slack.DialogSelectOption {
// 	options := []string{"All Weeks", "Select Week"}
// 	selectOptions := make([]slack.DialogSelectOption, len(options))
// 	for i, v := range options {
// 		selectOptions[i] = slack.DialogSelectOption{
// 			Label: v,
// 			Value: v,
// 		}
// 	}
// 	return selectOptions
// }

func weekOfYearOptions() []slack.DialogSelectOption {
	week := firstDayOfWeek(time.Now())
	year, weekNo := week.ISOWeek()
	selectOptions := make([]slack.DialogSelectOption, 25)

	selectOptions[0] = slack.DialogSelectOption{
		Label: "All Weeks",
		Value: encodeWeekAndYear(0, year),
	}
	for i := 0; i < 24; i++ {
		beginWeekMonth := week.Month().String()
		beginWeekDay := week.Day()
		endWeek := week.AddDate(0, 0, 4)
		endWeekMonth := endWeek.Month().String()
		endWeekDay := endWeek.Day()
		weekLabel := fmt.Sprintf("Week %d : %s %d - %s %d", weekNo, beginWeekMonth, beginWeekDay, endWeekMonth, endWeekDay)
		selectOptions[i+1] = slack.DialogSelectOption{
			Label: weekLabel,
			Value: encodeWeekAndYear(weekNo, year),
		}
		week = week.AddDate(0, 0, 7)
		year, weekNo = week.ISOWeek()
	}
	return selectOptions
}

func firstDayOfWeek(day time.Time) time.Time {
	var firstDay time.Time
	weekDay := int(day.Weekday())
	if weekDay == 0 {
		// Sunday -> Add one more day
		firstDay = day.AddDate(0, 0, 1)
	} else if weekDay == 1 {
		firstDay = day
	} else {
		firstDay = day.AddDate(0, 0, -(weekDay - 1))
	}

	return firstDay
}

func parseSlackIDFromString(combinedID string) string {
	// Format is <@U1234|user>
	match := "([A-Z])\\w+"
	re := regexp.MustCompile(match)

	return re.FindString(combinedID)
}
