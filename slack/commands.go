package slack

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/keremk/challenge-bot/models"

	"github.com/keremk/challenge-bot/config"
	slackApi "github.com/nlopes/slack"
)

type command struct {
	mainCommand  string
	subCommand   string
	arg          string
	slashCommand *slackApi.SlashCommand
}

func ExecuteCommand(env config.Environment, request *http.Request) error {
	slashCommand, err := parsePayload(request, env.VerificationToken)
	if err != nil {
		return err
	}

	c := parseSlashCommand(slashCommand)
	log.Println("[INFO] Challenge command")
	log.Println("[INFO] Main Command", c.mainCommand)
	log.Println("[INFO] Sub Command", c.subCommand)
	log.Println("[INFO] Text", c.arg)

	switch c.mainCommand {
	case "/challenge":
		log.Println("[INFO] Challenge command is invoked")
		fallthrough
	case "/challengetest":
		switch c.subCommand {
		case "help":
			go executeHelp(c.slashCommand.ResponseURL)
		case "new":
			go executeNewChallenge(env, c)
		case "send":
			go executeSendChallenge(env, c)
		}
	default:
		log.Println("[ERROR] Unexpected Command ", c.mainCommand)
		return errors.New("Unexpected command")
	}
	return nil
}

func parsePayload(request *http.Request, verificationToken string) (*slackApi.SlashCommand, error) {
	s, err := slackApi.SlashCommandParse(request)
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

func parseSlashCommand(slashCommand *slackApi.SlashCommand) command {
	helpCommand := command{
		mainCommand:  slashCommand.Command,
		subCommand:   "help",
		arg:          "",
		slashCommand: slashCommand,
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

func executeHelp(responseURL string) error {

	return nil
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
	challengeList, err := challengeNames(env)
	if err != nil {
		return err
	}
	dialog := newChallengeOptionsDialog(s.TriggerID, state, challengeList)

	slackClient := slackApi.New(token)
	err = slackClient.OpenDialog(s.TriggerID, *dialog)
	if err != nil {
		log.Println("[ERROR] Cannot create the dialog ", err)
	}
	return err
}

func executeNewChallenge(env config.Environment, c command) error {
	return nil
}

func challengeNames(env config.Environment) ([]string, error) {
	settings, err := models.GetAllChallenges(env)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, setting := range settings {
		names = append(names, setting.Name)
	}
	return names, nil
}

func newChallengeOptionsDialog(triggerID string, state dialogState, options []string) *slackApi.Dialog {
	candidateNameElement := slackApi.NewTextInput("candidate_name", "Candidate Name", "")
	githubNameElement := slackApi.NewTextInput("github_alias", "Github Alias", "")
	resumeURLElement := slackApi.NewTextInput("resume_URL", "Resume URL", "")
	selectOptions := make([]slackApi.DialogSelectOption, len(options))
	for i, v := range options {
		selectOptions[i] = slackApi.DialogSelectOption{
			Label: v,
			Value: v,
		}
	}
	challengeNameElement := slackApi.NewStaticSelectDialogInput("challenge_name", "Challenge Name", selectOptions)

	elements := []slackApi.DialogElement{
		candidateNameElement,
		githubNameElement,
		resumeURLElement,
		challengeNameElement,
	}

	return &slackApi.Dialog{
		TriggerID:      triggerID,
		CallbackID:     "challenge_67e2d0",
		Title:          "Create Coding Challenge",
		SubmitLabel:    "Create",
		NotifyOnCancel: false,
		State:          state.string(),
		Elements:       elements,
	}
}
