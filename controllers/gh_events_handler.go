package controllers

import (
	"log"
	"net/http"

	"github.com/google/go-github/github"
	"github.com/keremk/challenge-bot/config"
)

type ghEventsHandler struct {
	env config.Environment
}

func (gh ghEventsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	payload, err := github.ValidatePayload(r, nil)
	if err != nil {
		w.WriteHeader(400)
		log.Println("[ERROR] Could not validate payload - ", err)
		return
	}

	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		w.WriteHeader(500)
		log.Println(w, "[ERROR] Cannot parse webhook contents - ", err)
		return
	}

	log.Println("Event received")

	switch event := event.(type) {
	case *github.PullRequestEvent:
		if *event.Action == "opened" {
			log.Println("PR event")
		}
	case *github.InstallationEvent:
		if *event.Action == "created" {
			log.Printf("Installation successful with id = %d", *event.Installation.ID)
		}
	}
}
