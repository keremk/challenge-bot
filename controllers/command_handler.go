package controllers

import (
	"net/http"

	"github.com/keremk/challenge-bot/config"
	"github.com/keremk/challenge-bot/slack"
)

type CommandHandler struct {
	env config.Environment
}

func (h CommandHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := slack.ExecuteCommand(h.env, r)
	if err != nil {
		switch err.(type) {
		case slack.ValidationError:
			w.WriteHeader(http.StatusUnauthorized)
			return
		default:
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusAccepted)
}
