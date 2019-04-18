package config

import (
	"log"

	"github.com/kelseyhightower/envconfig"
)

type Environment struct {
	Port               string `envconfig:"PORT" default:"4390"`
	BotToken           string `envconfig:"BOT_TOKEN" required:"true"`
	VerificationToken  string `envconfig:"VERIFICATION_TOKEN" required:"true"`
	GithubToken        string `envconfig:"GITHUB_TOKEN" required:"true"`
	GithubOwner        string `envconfig:"GITHUB_OWNER" required:"true"`
	GithubOrganization string `envconfig:"GITHUB_ORG" required:"false"`
	GithubRepo         string `envconfig:"GITHUB_REPO" required:"true"`
}

func NewEnvironment(params ...string) *Environment {
	if len(params) == 0 {
		log.Println("[Info] using production environment")
		return getProductionEnv()
	}

	switch requiredEnv := params[0]; requiredEnv {
	case "production":
		log.Println("[Info] using production environment")
		return getProductionEnv()
	case "unittest":
		log.Println("[Info] using unit test environment")
		return getUnitTestEnv()
	case "integrationtest":
		log.Println("[Info] using integration test environment")
		return getIntegrationTestEnv()
	default:
		log.Println("[Info] using production environment")
		return getProductionEnv()
	}
}

func getProductionEnv() *Environment {
	env := &Environment{}
	err := envconfig.Process("", env)
	if err != nil {
		log.Fatalln("[Error] Can not read environment variables ", err)
	}
	return env
}

func getUnitTestEnv() *Environment {
	return &Environment{
		Port:               "4390",
		BotToken:           "FakeToken",
		VerificationToken:  "FakeToken",
		GithubToken:        "FakeToken",
		GithubOwner:        "Owner",
		GithubOrganization: "ORG",
		GithubRepo:         "challenge-bot",
	}
}

func getIntegrationTestEnv() *Environment {
	env := getProductionEnv()
	env.GithubRepo = "challenge-sample"
	env.GithubOwner = "keremk"
	env.GithubOrganization = ""
	return env
}
