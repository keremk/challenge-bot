package slack

import (
	"log"
	"net/http"

	"github.com/keremk/challenge-bot/config"
	slackApi "github.com/nlopes/slack"
)

type dialogHandler struct {
	slackClient     *slackApi.Client
	env             *config.Environment
	challengeConfig *config.ChallengeConfig
}

func (handler *dialogHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	parser := &ResponseParser{
		VerificationToken: handler.env.VerificationToken,
	}

	challengeDesc, returnChannel, err := parser.DialogResponseParse(r.Body)
	if err != nil {
		switch err.(type) {
		case ValidationError:
			log.Println("[ERROR] Unable to validate request ", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		default:
			log.Println("[ERROR] Unexpected request ", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	slackActionCtx := newSlackActionContext(handler.challengeConfig, handler.slackClient)

	go slackActionCtx.createChallenge(challengeDesc, returnChannel)

	w.WriteHeader(http.StatusAccepted)
}
