package slackops

import (
	"regexp"
	"strconv"
	"time"

	"github.com/keremk/challenge-bot/config"
	"github.com/keremk/challenge-bot/scheduling"
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
	experienceLevelEl := newStaticOptionsDialogInput("experience", "Experience Level", true, experienceOptions())
	elements := []slack.DialogElement{
		reviewerEl,
		githubNameEl,
		challengeNameEl,
		technologyListEl,
		experienceLevelEl,
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
	weekOfYearEl := newStaticOptionsDialogInput("year_week", "Week of the Year", true, weekOfYearOptions(true))

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

func weekOfYearOptions(includeAllWeeks bool) []slack.DialogSelectOption {
	week := scheduling.FirstDayOfWeek(time.Now())
	year, weekNo := week.ISOWeek()
	selectOptions := make([]slack.DialogSelectOption, 0, 25)

	if includeAllWeeks {
		selectOptions = append(selectOptions, slack.DialogSelectOption{
			Label: "All Weeks",
			Value: encodeWeekAndYear(0, year),
		})
	}
	for i := 0; i < 24; i++ {
		weekLabel := scheduling.WeekDescription(week)
		selectOptions = append(selectOptions, slack.DialogSelectOption{
			Label: weekLabel,
			Value: encodeWeekAndYear(weekNo, year),
		})
		week = week.AddDate(0, 0, 7)
		year, weekNo = week.ISOWeek()
	}
	return selectOptions
}

func experienceOptions() []slack.DialogSelectOption {
	experienceLevel := []string{"Low", "Mid", "High"}
	selectOptions := make([]slack.DialogSelectOption, 0, len(experienceLevel))
	for i, level := range experienceLevel {
		selectOptions = append(selectOptions, slack.DialogSelectOption{
			Label: level,
			Value: strconv.Itoa(i),
		})
	}
	return selectOptions
}

func dayOptions() []slack.DialogSelectOption {
	daysOfWeek := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday"}

	selectOptions := make([]slack.DialogSelectOption, 0, len(daysOfWeek))
	for _, day := range daysOfWeek {
		selectOptions = append(selectOptions, slack.DialogSelectOption{
			Label: day,
			Value: day,
		})
	}
	return selectOptions
}

func parseSlackIDFromString(combinedID string) string {
	// Format is <@U1234|user>
	match := "([A-Z])\\w+"
	re := regexp.MustCompile(match)

	return re.FindString(combinedID)
}

func executeFindReviewers(env config.Environment, c command) error {
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

	dialog := newFindDialog(s.TriggerID, state)

	return showDialog(env, s.TeamID, s.TriggerID, dialog)
}

func newFindDialog(triggerID string, state dialogState) slack.Dialog {
	weekOfYearEl := newStaticOptionsDialogInput("year_week", "Week of the Year", true, weekOfYearOptions(false))
	dayEl := newStaticOptionsDialogInput("day", "Day of Week", true, dayOptions())
	challengeNameEl := newExternalOptionsDialogInput("challenge_name", "Challenge Name", "", false)
	technologyEl := slack.NewTextInput("technology", "Technology List", "")
	elements := []slack.DialogElement{
		weekOfYearEl,
		dayEl,
		challengeNameEl,
		technologyEl,
	}
	return slack.Dialog{
		TriggerID:      triggerID,
		CallbackID:     "find_reviewers",
		Title:          "Find Reviewers",
		SubmitLabel:    "Search",
		NotifyOnCancel: false,
		State:          state.string(),
		Elements:       elements,
	}
}
