package slack

import (
	"log"
	"net/http"

	"github.com/keremk/challenge-bot/config"
	slackApi "github.com/nlopes/slack"
)

type commandHandler struct {
	slackClient     *slackApi.Client
	env             *config.Environment
	challengeConfig *config.ChallengeConfig
}

func (handler *commandHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	parser := &ResponseParser{
		VerificationToken: handler.env.VerificationToken,
	}

	s, err := parser.SlashCommandParse(r)
	if err != nil {
		switch err.(type) {
		case ValidationError:
			log.Println("[ERROR] Unable to validate command ", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		default:
			log.Println("[ERROR] Unable to parse command ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	switch s.Command {
	case "/challenge":
	case "/challengetest":
		// Immediately return
		w.WriteHeader(http.StatusOK)

		// Create the dialog and send a message to open it
		dialog := newChallengeOptionsDialog(s.TriggerID, s.ChannelID, handler.challengeConfig.AllDisciplines())
		err := handler.slackClient.OpenDialog(s.TriggerID, *dialog)
		if err != nil {
			log.Println("[ERROR] Cannot create the dialog ", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	default:
		log.Println("[ERROR] Unexcepted Command ", s.Command)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
