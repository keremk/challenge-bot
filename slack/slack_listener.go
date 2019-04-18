package slack

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/keremk/challenge-bot/config"
	slackApi "github.com/nlopes/slack"
)

func SetupSlackListeners() {
	env := config.NewEnvironment("production")
	challengeConfig, err := config.NewChallengeConfig(env, config.NewGithubChallengeReader())
	if err != nil {
		log.Fatalln("[ERROR] Configuration cannot be retrieved ", err)
	}

	client := slackApi.New(env.BotToken)

	http.Handle("/commands", &commandHandler{
		slackClient:     client,
		challengeConfig: challengeConfig,
		env:             env,
	})
	http.Handle("/requests", &requestHandler{
		slackClient:     client,
		challengeConfig: challengeConfig,
		env:             env,
	})

	http.Handle("/events", &eventsHandler{
		slackClient:     client,
		challengeConfig: challengeConfig,
		env:             env,
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("[INFO] Defaulting to port %s and listening", port)
	}

	log.Printf("[INFO] Listening on port %s", port)
	http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}
