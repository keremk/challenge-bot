package slack

import (
	"errors"
	"log"
	"net/http"

	"github.com/keremk/challenge-bot/config"
	slackApi "github.com/nlopes/slack"
)

func ExecuteCommand(env config.Environment, request *http.Request) error {
	s, err := parseCommand(request, env.VerificationToken)
	if err != nil {
		return err
	}

	switch s.Command {
	case "/challenge":
		log.Println("[INFO] Challenge command is invoked")
		fallthrough
	case "/challengetest":
		go executeChallengeCmd(env, s)
		return nil
	default:
		log.Println("[ERROR] Unexpected Command ", s.Command)
		return errors.New("Unexpected command")
	}
}

func parseCommand(request *http.Request, verificationToken string) (*slackApi.SlashCommand, error) {
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

func executeChallengeCmd(env config.Environment, s *slackApi.SlashCommand) error {
	token, err := getBotToken(env, s.TeamID)
	if err != nil {
		return err
	}

	// Create the dialog and send a message to open it
	dialog := newChallengeOptionsDialog(s.TriggerID, s.ChannelID, allDisciplines())

	slackClient := slackApi.New(token)
	err = slackClient.OpenDialog(s.TriggerID, *dialog)
	if err != nil {
		log.Println("[ERROR] Cannot create the dialog ", err)
	}
	return err
}

// TODO: Temp filler until we move the challenges to DB
func allDisciplines() []string {
	return []string{"backend", "frontend"}
}
