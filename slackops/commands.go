package slackops

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/keremk/challenge-bot/config"
	"github.com/keremk/challenge-bot/models"
	"github.com/nlopes/slack"
)

type command struct {
	mainCommand  string
	subCommand   string
	arg          string
	slashCommand *slack.SlashCommand
}

func ExecuteCommand(env config.Environment, request *http.Request) error {
	slashCommand, err := parsePayload(request, env.VerificationToken)
	if err != nil {
		return err
	}

	c := parseSlashCommand(slashCommand)
	log.Printf("[INFO] Main Command %s, Sub Command %s, Text %s", c.mainCommand, c.subCommand, c.arg)

	switch c.mainCommand {
	case "/challenge":
		log.Println("[INFO] Challenge command is invoked")
		fallthrough
	case "/challengetest":
		switch c.subCommand {
		case "help":
			log.Println("[INFO] HELP is called here")
			go executeHelp(c)
		case "new":
			go executeNewChallenge(env, c)
		case "send":
			go executeSendChallenge(env, c)
		case "reviewer":
			go executeAddReviewer(env, c)
		}
	default:
		log.Println("[ERROR] Unexpected Command ", c.mainCommand)
		return errors.New("Unexpected command")
	}
	return nil
}

func parsePayload(request *http.Request, verificationToken string) (*slack.SlashCommand, error) {
	s, err := slack.SlashCommandParse(request)
	if err != nil {
		log.Println("[ERROR] Unable to parse command ", err)
		return nil, err
	}

	if !s.ValidateToken(verificationToken) {
		log.Println("[ERROR] Unable to validate command ", err)
		return nil, ValidationError{}
	}
	return &s, nil
}

func parseSlashCommand(slashCommand *slack.SlashCommand) command {
	helpCommand := command{
		mainCommand:  slashCommand.Command,
		subCommand:   "help",
		arg:          "",
		slashCommand: slashCommand,
	}

	if len(slashCommand.Text) == 0 {
		return helpCommand
	}

	c := strings.Split(slashCommand.Text, " ")
	switch len(c) {
	case 1:
		return command{
			mainCommand:  slashCommand.Command,
			subCommand:   c[0],
			arg:          "",
			slashCommand: slashCommand,
		}
	case 2:
		return command{
			mainCommand:  slashCommand.Command,
			subCommand:   c[0],
			arg:          c[1],
			slashCommand: slashCommand,
		}
	default:
		return helpCommand
	}
}

func executeHelp(c command) error {
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
				"text": "*/challenge help* : Displays this message"
			}
		}, 
		{
			"type": "section",
			"text": {
				"type": "mrkdwn",
				"text": "*/challenge new* : Opens a dialog to create a new challenge"
			}
		},
		{
			"type": "section",
			"text": {
				"type": "mrkdwn",
				"text": "*/challenge send* : Opens a dialog to send a challenge to a candidate"
			}
		},
		{
			"type": "section",
			"text": {
				"type": "mrkdwn",
				"text": "*/challenge reviewer* : Opens a dialog to register a reviewer for a challenge"
			}
		}
	]
}
`
	err := sendDelayedResponse(c.slashCommand.ResponseURL, helpText)
	return err
}

func executeSendChallenge(env config.Environment, c command) error {
	s := c.slashCommand
	token, err := getBotToken(env, s.TeamID)
	if err != nil {
		return err
	}

	// Create the dialog and send a message to open it
	state := dialogState{
		channelID:    s.ChannelID,
		settingsName: c.arg,
	}
	dialog := sendChallengeDialog(s.TriggerID, state)

	slackClient := slack.New(token)
	err = slackClient.OpenDialog(s.TriggerID, *dialog)
	if err != nil {
		log.Println("[ERROR] Cannot create the dialog ", err)
	}
	return err
}

func executeNewChallenge(env config.Environment, c command) error {
	s := c.slashCommand
	token, err := getBotToken(env, s.TeamID)
	if err != nil {
		return err
	}

	// Create the dialog and send a message to open it
	state := dialogState{
		channelID:    s.ChannelID,
		settingsName: c.arg,
	}
	dialog := newChallengeDialog(s.TriggerID, state)

	slackClient := slack.New(token)
	err = slackClient.OpenDialog(s.TriggerID, *dialog)
	if err != nil {
		log.Println("[ERROR] Cannot create the dialog ", err)
	}
	return err
}

func executeAddReviewer(env config.Environment, c command) error {
	s := c.slashCommand
	token, err := getBotToken(env, s.TeamID)
	if err != nil {
		return err
	}

	// Create the dialog and send a message to open it
	state := dialogState{
		channelID:    s.ChannelID,
		settingsName: c.arg,
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

func sendChallengeDialog(triggerID string, state dialogState) *slack.Dialog {
	candidateNameElement := slack.NewTextInput("candidate_name", "Candidate Name", "")
	githubNameElement := slack.NewTextInput("github_alias", "Github Alias", "")
	resumeURLElement := slack.NewTextInput("resume_URL", "Resume URL", "")
	challengeNameElement := newExternalOptionsDialogInput("challenge_name", "Challenge Name")
	reviewer1OptionsElement := newExternalOptionsDialogInput("reviewer1_id", "Reviewer 1")
	reviewer2OptionsElement := newExternalOptionsDialogInput("reviewer2_id", "Reviewer 2")

	elements := []slack.DialogElement{
		candidateNameElement,
		githubNameElement,
		resumeURLElement,
		challengeNameElement,
		reviewer1OptionsElement,
		reviewer2OptionsElement,
	}

	return &slack.Dialog{
		TriggerID:      triggerID,
		CallbackID:     "send_challenge",
		Title:          "Send Coding Challenge",
		SubmitLabel:    "Send",
		NotifyOnCancel: false,
		State:          state.string(),
		Elements:       elements,
	}
}

func newExternalOptionsDialogInput(name, label string) *slack.DialogInputSelect {
	return &slack.DialogInputSelect{
		DialogInput: slack.DialogInput{
			Type:     slack.InputTypeSelect,
			Name:     name,
			Label:    label,
			Optional: true,
		},
		DataSource: slack.DialogDataSourceExternal,
	}
}

func newChallengeDialog(triggerID string, state dialogState) *slack.Dialog {
	challengeNameEl := slack.NewTextInput("challenge_name", "Challenge Name", "")
	templateRepoNameEl := slack.NewTextInput("template_repo", "Template Repo Name", "")
	repoNameFormatEl := slack.NewTextInput("repo_name_format", "Repo Name Format", "test_CHALLENGENAME-GITHUBALIAS")

	githubAccountEl := newExternalOptionsDialogInput("github_account", "Github Account Name")
	elements := []slack.DialogElement{
		challengeNameEl,
		templateRepoNameEl,
		repoNameFormatEl,
		githubAccountEl,
	}
	return &slack.Dialog{
		TriggerID:      triggerID,
		CallbackID:     "new_challenge",
		Title:          "New Coding Challenge",
		SubmitLabel:    "Create",
		NotifyOnCancel: false,
		State:          state.string(),
		Elements:       elements,
	}
}

func newAddReviewerDialog(triggerID string, state dialogState, options []models.Challenge) *slack.Dialog {
	reviewerEl := slack.NewUsersSelect("reviewer_id", "Reviewer")
	githubNameEl := slack.NewTextInput("github_alias", "Github Alias", "")
	challengeNameEl := newExternalOptionsDialogInput("challenge_name", "Challenge Name")
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
