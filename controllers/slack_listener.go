package controllers

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/keremk/challenge-bot/config"
)

func SetupSlackListeners() {
	env := config.NewEnvironment("production")

	http.Handle("/commands", &CommandHandler{
		env: *env,
	})
	http.Handle("/requests", &dialogHandler{
		env: *env,
	})
	http.Handle("/auth/", http.StripPrefix("/auth/", http.FileServer(http.Dir("./static"))))
	http.Handle("/auth/redirect", &authHandler{
		env: env,
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		log.Println("[INFO] Health ok")
		w.WriteHeader(http.StatusOK)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("[INFO] Defaulting to port %s and listening", port)
	}

	log.Printf("[INFO] Listening on port %s", port)
	http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}
