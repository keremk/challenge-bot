package controllers

import (
	"log"
	"net/http"

	"github.com/keremk/challenge-bot/slackops"
	"github.com/keremk/challenge-bot/config"
)

type dialogHandler struct {
	env config.Environment
}

func (h *dialogHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := slackops.HandleDialogResponse(h.env, r.Body)

	if err != nil {
		switch err.(type) {
		case slackops.ValidationError:
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
