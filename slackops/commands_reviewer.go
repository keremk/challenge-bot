package slackops

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"time"

	"github.com/keremk/challenge-bot/models"
	"github.com/keremk/challenge-bot/scheduling"
	"github.com/nlopes/slack"
)

func (c command) executeReviewerHelp() error {
	return c.ctx.postMessage(c.slashCmd.ChannelID, renderReviewerHelp())
}

func (c command) executeNewReviewer() error {
	dialog := newAddReviewerDialog(c.slashCmd.TriggerID)

	return c.ctx.showDialog(c.slashCmd.TriggerID, dialog)
}

func (c command) executeEditReviewer() error {
	var reviewerSlackID string
	if c.arg == "" {
		reviewerSlackID = c.slashCmd.UserID
	} else {
		reviewerSlackID = parseSlackIDFromString(c.arg)
	}
	reviewer, err := models.GetReviewerBySlackID(c.ctx.Env, reviewerSlackID)
	if err != nil {
		log.Println("[ERROR] No such reviewer registered.", err)
		errorMsg := fmt.Sprintf("Reviewer <@%s> is not registered. Please register first using /reviewer new command.", reviewerSlackID)
		c.ctx.postMessage(c.slashCmd.ChannelID, toMsgOption(errorMsg))
		return err
	}

	dialog := newEditReviewerDialog(c.slashCmd.TriggerID, reviewer)

	return c.ctx.showDialog(c.slashCmd.TriggerID, dialog)
}

func newAddReviewerDialog(triggerID string) slack.Dialog {
	return slack.Dialog{
		TriggerID:      triggerID,
		CallbackID:     "new_reviewer",
		Title:          "Add Reviewer",
		SubmitLabel:    "Add",
		NotifyOnCancel: false,
		State:          "",
		Elements:       reviewerDialogElements(models.Reviewer{}, false),
	}
}

func newEditReviewerDialog(triggerID string, reviewer models.Reviewer) slack.Dialog {
	return slack.Dialog{
		TriggerID:      triggerID,
		CallbackID:     "edit_reviewer",
		Title:          "Edit Reviewer",
		SubmitLabel:    "Edit",
		NotifyOnCancel: false,
		State:          reviewer.SlackID,
		Elements:       reviewerDialogElements(reviewer, true),
	}
}

func reviewerDialogElements(reviewer models.Reviewer, editMode bool) []slack.DialogElement {
	elements := make([]slack.DialogElement, 0, 10)
	if !editMode {
		reviewerEl := slack.NewUsersSelect("reviewer_id", "Reviewer")
		elements = append(elements, reviewerEl)
	}

	githubNameEl := slack.NewTextInput("github_alias", "Github Alias", reviewer.GithubAlias)
	challengeNameEl := newExternalOptionsDialogInput("challenge_name", "Challenge Name", "", false)
	technologyListEl := slack.NewTextInput("technology_list", "Technology List", reviewer.TechnologyList)
	experienceLevel := strconv.Itoa(reviewer.Experience)
	experienceLevelEl := newStaticOptionsDialogInput("experience", "Experience Level", experienceLevel, true, experienceOptions())
	bookingsPerWeek := strconv.Itoa(reviewer.BookingsPerWeek)
	bookingsPerWeekEl := newStaticOptionsDialogInput("bookings_week", "# Bookings per Week", bookingsPerWeek, true,
		bookingsOptions())

	return append(elements,
		githubNameEl,
		challengeNameEl,
		technologyListEl,
		experienceLevelEl,
		bookingsPerWeekEl,
	)
}

func (c command) executeSchedule() error {
	var reviewerSlackID string
	if c.arg == "" {
		reviewerSlackID = c.slashCmd.UserID
	} else {
		reviewerSlackID = parseSlackIDFromString(c.arg)
	}

	dialog := newScheduleDialog(c.slashCmd.TriggerID, reviewerSlackID)

	return c.ctx.showDialog(c.slashCmd.TriggerID, dialog)
}

func newScheduleDialog(triggerID string, reviewerSlackID string) slack.Dialog {
	weekOfYearDefault := encodeWeekAndYear(0, time.Now().Year())
	weekOfYearEl := newStaticOptionsDialogInput("year_week", "Week of the Year", weekOfYearDefault, true, weekOfYearOptions(true))

	elements := []slack.DialogElement{
		weekOfYearEl,
	}
	return slack.Dialog{
		TriggerID:      triggerID,
		CallbackID:     "schedule_update",
		Title:          "Update Schedule",
		SubmitLabel:    "Update",
		NotifyOnCancel: false,
		State:          reviewerSlackID,
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

func bookingsOptions() []slack.DialogSelectOption {
	selectOptions := make([]slack.DialogSelectOption, 0, 7)
	for i := 1; i < 7; i++ {
		label := fmt.Sprintf("Max %d times/week", i)
		selectOptions = append(selectOptions, slack.DialogSelectOption{
			Label: label,
			Value: strconv.Itoa(i),
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

func (c command) executeFindReviewers() error {
	var reviewerSlackID string
	if c.arg == "" {
		reviewerSlackID = c.slashCmd.UserID
	} else {
		reviewerSlackID = parseSlackIDFromString(c.arg)
	}

	dialog := newFindDialog(c.slashCmd.TriggerID, reviewerSlackID)

	return c.ctx.showDialog(c.slashCmd.TriggerID, dialog)
}

func newFindDialog(triggerID string, reviewerSlackID string) slack.Dialog {
	defaultYear, defaultWeekNo := time.Now().ISOWeek()
	weekOfYearDefault := encodeWeekAndYear(defaultWeekNo, defaultYear)
	weekOfYearEl := newStaticOptionsDialogInput("year_week", "Week of the Year", weekOfYearDefault, true, weekOfYearOptions(false))
	defaultDay := "Monday"
	dayEl := newStaticOptionsDialogInput("day", "Day of Week", defaultDay, true, dayOptions())
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
		State:          reviewerSlackID,
		Elements:       elements,
	}
}

type sectionMsg struct {
	ReplaceOriginal bool          `json:"replace_original,omitempty"`
	Blocks          []slack.Block `json:"blocks,omitempty"`
}

func (c command) executeShowBookings() error {
	var reviewerSlackID string
	if c.arg == "" {
		reviewerSlackID = c.slashCmd.UserID
	} else {
		reviewerSlackID = parseSlackIDFromString(c.arg)
	}

	reviewer, err := models.GetReviewerBySlackID(c.ctx.Env, reviewerSlackID)
	if err != nil {
		log.Println("[ERROR] No such reviewer registered.", err)
		errorMsg := fmt.Sprintf("Reviewer <@%s> is not registered. Please register first using /reviewer new command.", reviewerSlackID)
		c.ctx.postMessage(c.slashCmd.ChannelID, toMsgOption(errorMsg))
		return err
	}
	challenge, err := models.GetChallengeSetup(c.ctx.Env, reviewer.ChallengeName)
	if err != nil {
		log.Println("[ERROR] Invalid challenge for reviewer", err)
		errorMsg := fmt.Sprintf("Reviewer <@%s> does not seem to have a valid challenge they registered. Please use /reviewer edit to register a challenge.", reviewerSlackID)
		c.ctx.postMessage(c.slashCmd.ChannelID, toMsgOption(errorMsg))
		return err
	}

	sections := renderBookings(reviewer, challenge)

	return c.ctx.postMessage(c.slashCmd.ChannelID, slack.MsgOptionBlocks(sections...))
}
