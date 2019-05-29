package controllers

import (
	"log"
	"net/http"

	"github.com/keremk/challenge-bot/config"
	"github.com/keremk/challenge-bot/slackops"
)

type optionsHandler struct {
	env config.Environment
}

func (h optionsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	respJSON, err := slackops.HandleOptions(h.env, r.Body)
	if err != nil {
		log.Println("[ERROR] error handling options - ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(respJSON)
}
