package config

import (
	"fmt"
	"os"
	"sync"

	"github.com/kelseyhightower/envconfig"
)

type Environment struct {
	Port              string `envconfig:"PORT" default:"4390"`
	BotToken          string `envconfig:"BOT_TOKEN" required:"true"`
	VerificationToken string `envconfig:"VERIFICATION_TOKEN" required:"true"`
	GithubToken       string `envconfig:"GITHUB_TOKEN" required:"true"`
	GithubOwner       string `envconfig:"GITHUB_OWNER" required:"true"`
	GithubRepo        string `envconfig:"GITHUB_REPO" required:"true"`
}

var envInstance *Environment
var envOnce sync.Once

func GetEnvironment() *Environment {
	envOnce.Do(func() {
		envInstance = &Environment{}
		err := envconfig.Process("", envInstance)
		if err != nil {
			fmt.Println("[ERROR] Can not read environment variables ", err)
			os.Exit(1)
		}
	})
	return envInstance
}
