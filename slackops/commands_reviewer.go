package slackops

import (
	"log"

	"github.com/keremk/challenge-bot/config"
	"github.com/keremk/challenge-bot/models"
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
	token, err := getBotToken(env, s.TeamID)
	if err != nil {
		return err
	}

	// Create the dialog and send a message to open it
	state := dialogState{
		channelID:     s.ChannelID,
		challengeName: c.arg,
	}
	challengeList, err := models.GetAllChallenges(env)
	if err != nil {
		return err
	}
	dialog := newAddReviewerDialog(s.TriggerID, state, challengeList)

	slackClient := slack.New(token)
	err = slackClient.OpenDialog(s.TriggerID, *dialog)
	if err != nil {
		log.Println("[ERROR] Cannot create the dialog ", err)
	}
	return err
}

func newAddReviewerDialog(triggerID string, state dialogState, options []models.Challenge) *slack.Dialog {
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
	return &slack.Dialog{
		TriggerID:      triggerID,
		CallbackID:     "new_reviewer",
		Title:          "Add Reviewer",
		SubmitLabel:    "Add",
		NotifyOnCancel: false,
		State:          state.string(),
		Elements:       elements,
	}
}
