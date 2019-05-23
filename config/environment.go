package config

import (
	"log"

	"github.com/kelseyhightower/envconfig"
)

type Environment struct {
	Port                     string `envconfig:"PORT" default:"4390"`
	VerificationToken        string `envconfig:"VERIFICATION_TOKEN" required:"true"`
	GithubToken              string `envconfig:"GITHUB_TOKEN" required:"true"`
	SlackClientID            string `envconfig:"SLACK_CLIENT_ID" required:"true"`
	SlackClientSecret        string `envconfig:"SLACK_CLIENT_SECRET" required:"true"`
	SlackRedirectURI         string `envconfig:"SLACK_REDIRECT_URI" required:"true"`
	GithubClientID           string `envconfig:"GITHUB_CLIENT_ID" required:"true"`
	GithubClientSecret       string `envconfig:"GITHUB_CLIENT_SECRET" required:"true"`
	GithubRedirectURI        string `envconfig:"GITHUB_REDIRECT_URI" required:"true"`
	GithubPrivateKeyFilename string `envconfig:"GITHUB_PRIVATEKEYFILENAME" required:"true"`
	GCloudProjectID          string `envconfig:"GCLOUD_PROJECT_ID" required:"true"`
	DbProvider               string `envconfig:"DB_PROVIDER" required:"true"`
	DebugOn                  bool   `envconfig:"DEBUG_ON" required:"true"`
}

func NewEnvironment(params ...string) Environment {
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

func getProductionEnv() Environment {
	env := Environment{}
	err := envconfig.Process("", &env)
	if err != nil {
		log.Fatalln("[ERROR] Can not read environment variables ", err)
	}
	log.Println("[INFO] DB Provider is: ", env.DbProvider)
	log.Println("[INFO] GCloud Project ID is: ", env.GCloudProjectID)
	return env
}

func getUnitTestEnv() Environment {
	return Environment{
		Port:              "4390",
		VerificationToken: "FakeToken",
		GithubToken:       "FakeToken",
		SlackClientID:     "Fake",
		SlackClientSecret: "Fake",
		SlackRedirectURI:  "http://example.com",
		GCloudProjectID:   "coding-challenge-bot",
	}
}

func getIntegrationTestEnv() Environment {
	env := getProductionEnv()
	return env
}
