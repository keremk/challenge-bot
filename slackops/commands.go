package slackops

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/keremk/challenge-bot/config"
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
			go executeChallengeHelp(c)
		case "new":
			go executeNewChallenge(env, c)
		case "send":
			go executeSendChallenge(env, c)
		}
	case "/reviewer":
		fallthrough
	case "/reviewertest":
		switch c.subCommand {
		case "help":
			log.Println("[INFO] HELP is called here")
			go executeReviewerHelp(c)
		case "new":
			go executeNewReviewer(env, c)
		case "edit":
			go executeEditReviewer(env, c)
		case "schedule":
			go executeSchedule(env, c)
		case "find":
			go executeFindReviewers(env, c)
		case "bookings":
			go executeShowBookings(env, c)
		default:
			log.Println("[ERROR] Unexpected Command ", c.mainCommand)
			return errors.New("Unexpected command")
		}
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

func showDialog(env config.Environment, teamID, triggerID string, dialog slack.Dialog) error {
	token, err := getBotToken(env, teamID)
	if err != nil {
		return err
	}
	slackClient := slack.New(token)
	err = slackClient.OpenDialog(triggerID, dialog)
	if err != nil {
		log.Println("[ERROR] Cannot create the dialog ", err)
	}
	return err
}

func newExternalOptionsDialogInput(name, label, value string, optional bool) *slack.DialogInputSelect {
	return &slack.DialogInputSelect{
		DialogInput: slack.DialogInput{
			Type:     slack.InputTypeSelect,
			Name:     name,
			Label:    label,
			Optional: optional,
		},
		DataSource: slack.DialogDataSourceExternal,
	}
}

func newStaticOptionsDialogInput(name, label, value string, optional bool, options []slack.DialogSelectOption) *slack.DialogInputSelect {
	return &slack.DialogInputSelect{
		DialogInput: slack.DialogInput{
			Type:     slack.InputTypeSelect,
			Name:     name,
			Label:    label,
			Optional: optional,
		},
		DataSource: slack.DialogDataSourceStatic,
		Options:    options,
		Value: 			value,
	}
}
