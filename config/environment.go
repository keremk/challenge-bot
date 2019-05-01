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
	SlackClientID      string `envconfig:"SLACK_CLIENT_ID" required:"true"`
	SlackClientSecret  string `envconfig:"SLACK_CLIENT_SECRET" required:"true"`
	SlackRedirectURI   string `envconfig:"SLACK_REDIRECT_URI" required:"true"`
	GCloudProjectID    string `envconfig:"GCLOUD_PROJECT_ID" required:"true"`
}

func NewEnvironment(params ...string) *Environment {
	if len(params) == 0 {
		log.Println("[INFO] using production environment")
		return getProductionEnv()
	}

	switch requiredEnv := params[0]; requiredEnv {
	case "production":
		log.Println("[INFO] using production environment")
		return getProductionEnv()
	case "unittest":
		log.Println("[INFO] using unit test environment")
		return getUnitTestEnv()
	case "integrationtest":
		log.Println("[INFO] using integration test environment")
		return getIntegrationTestEnv()
	default:
		log.Println("[INFO] using production environment")
		return getProductionEnv()
	}
}

func getProductionEnv() *Environment {
	env := &Environment{}
	err := envconfig.Process("", env)
	if err != nil {
		log.Fatalln("[ERROR] Can not read environment variables ", err)
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
		SlackClientID:      "Fake",
		SlackClientSecret:  "Fake",
		SlackRedirectURI:   "http://example.com",
		GCloudProjectID:    "coding-challenge-bot",
	}
}

func getIntegrationTestEnv() *Environment {
	env := getProductionEnv()
	env.GithubRepo = "challenge-sample"
	env.GithubOwner = "keremk"
	env.GithubOrganization = ""
	return env
}
