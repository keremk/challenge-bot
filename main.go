package main

import (
	"fmt"

	"github.com/keremk/challenge-bot/config"
	"github.com/keremk/challenge-bot/slack"
)

func main() {
	c := config.GetChallengeConfig()
	fmt.Println(c)

	slack.SetupSlackListeners()
}
