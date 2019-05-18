package controllers

import (
	"net/http"

	"github.com/keremk/challenge-bot/config"
	"github.com/keremk/challenge-bot/slackops"
)

type CommandHandler struct {
	env config.Environment
}

func (h CommandHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := slackops.ExecuteCommand(h.env, r)
	if err != nil {
		switch err.(type) {
		case slackops.ValidationError:
			w.WriteHeader(http.StatusUnauthorized)
			return
		default:
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusAccepted)
}
