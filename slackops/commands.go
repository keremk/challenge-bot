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
	ctx      commCtx
	command  string
	sub      string
	arg      string
	slashCmd *slack.SlashCommand
}

func newCommand(env config.Environment, slashCommand *slack.SlashCommand) command {
	sub, arg := parseSlashCommand(slashCommand)
	if sub == "" {
		sub = "help"
	}
	ctx := newCommCtx(env, slashCommand.UserID, slashCommand.TeamID, false)

	return command{
		ctx:      ctx,
		command:  slashCommand.Command,
		sub:      sub,
		arg:      arg,
		slashCmd: slashCommand,
	}
}

func ExecuteCommand(env config.Environment, request *http.Request) error {
	slashCommand, err := parsePayload(request, env.VerificationToken)
	if err != nil {
		return err
	}

	c := newCommand(env, slashCommand)
	log.Printf("[INFO] Main Command %s, Sub Command %s, Text %s", c.command, c.sub, c.arg)

	switch c.command {
	case "/challenge":
		fallthrough
	case "/challengetest":
		switch c.sub {
		case "help":
			go c.executeChallengeHelp()
		case "new":
			go c.executeNewChallenge()
		case "edit":
			go c.executeEditChallenge()
		case "send":
			go c.executeSendChallenge()
		}
	case "/reviewer":
		fallthrough
	case "/reviewertest":
		switch c.sub {
		case "help":
			go c.executeReviewerHelp()
		case "new":
			go c.executeNewReviewer()
		case "edit":
			go c.executeEditReviewer()
		case "schedule":
			go c.executeSchedule()
		case "find":
			go c.executeFindReviewers()
		case "bookings":
			go c.executeShowBookings()
		default:
			log.Println("[ERROR] Unexpected Command ", c.command)
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

func parseSlashCommand(slashCommand *slack.SlashCommand) (string, string) {
	if len(slashCommand.Text) == 0 {
		return "", ""
	}

	c := strings.Split(slashCommand.Text, " ")
	switch len(c) {
	case 1:
		return c[0], ""
	case 2:
		return c[0], c[1]
	default:
		return "", ""
	}
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
		Value:      value,
	}
}
