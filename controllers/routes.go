package controllers

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/keremk/challenge-bot/config"
)

func SetupRoutes() {
	env := config.NewEnvironment("production")

	setupSlackListeners(env)
	setupGithubListeners(env)

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		log.Println("[INFO] Health ok")
		w.WriteHeader(http.StatusOK)
	})

	http.Handle("/auth/", http.StripPrefix("/auth/", http.FileServer(http.Dir("./static"))))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("[INFO] Defaulting to port %s and listening", port)
	}

	log.Printf("[INFO] Listening on port %s", port)
	http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}

func setupSlackListeners(env config.Environment) {
	http.Handle("/commands", &commandHandler{
		env: env,
	})
	http.Handle("/requests", &requestsHandler{
		env: env,
	})
	http.Handle("/options", &optionsHandler{
		env: env,
	})

	http.Handle("/auth/slack/redirect", &authHandler{
		env: env,
	})
}

func setupGithubListeners(env config.Environment) {
	http.Handle("/auth/github/redirect", &ghAuthHandler{
		env: env,
	})
	http.Handle("/auth/github/setup", &ghSetupHandler{
		env: env,
	})
	http.Handle("/github/events", &ghEventsHandler{
		env: env,
	})
	http.Handle("/auth/github/createaccount", &ghAccountHandler{
		env: env,
	})
}
