package slack

import (
	"fmt"
	"net/http"

	"github.com/keremk/challenge-bot/config"
	slackApi "github.com/nlopes/slack"
)

func SetupSlackListeners() {
	challengeConfig := config.GetChallengeConfig()
	env := config.GetEnvironment()

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
	fmt.Println("[INFO] Server listening")
	http.ListenAndServe(":4390", nil)
}
