package controllers

import (
	"log"
	"net/http"

	"github.com/keremk/challenge-bot/slack"

	"github.com/keremk/challenge-bot/config"
)

type dialogHandler struct {
	env             config.Environment
	challengeConfig *config.ChallengeConfig
}

func (h *dialogHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := slack.HandleDialogResponse(h.env, r.Body, h.challengeConfig)

	if err != nil {
		switch err.(type) {
		case slack.ValidationError:
			w.WriteHeader(http.StatusUnauthorized)
			return
		default:
			log.Println("[ERROR] Unexpected request ", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	w.WriteHeader(http.StatusAccepted)
}
