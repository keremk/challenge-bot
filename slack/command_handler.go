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
			w.WriteHeader(http.StatusUnauthorized)
			return
		default:
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	switch s.Command {
	case "/challenge":
		// Immediately return
		w.WriteHeader(http.StatusOK)

		// Create the dialog and send a message to open it
		dialog := newChallengeOptionsDialog(s.TriggerID, s.ChannelID, handler.challengeConfig.AllDisciplines())
		err := handler.slackClient.OpenDialog(s.TriggerID, *dialog)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	default:
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
