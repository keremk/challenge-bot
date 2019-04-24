package config

import (
	"errors"
	"fmt"
	"log"

	yaml "gopkg.in/yaml.v2"
)

type Task struct {
	Level           int8
	Title           string
	DescriptionFile string `yaml:"descriptionFile"`
}

type Challenge struct {
	Discipline       string
	TemplateRepoName string `yaml:"templateRepoName"`
	Reviewers        []string
	Tasks            []Task
}

type ChallengeConfig struct {
	Organization           string
	Owner                  string
	TrackingRepoName       string `yaml:"trackingRepoName"`
	GithubToken            string
	SlackBotToken          string
	SlackVerificationToken string
	Challenges             []Challenge
	reader                 ChallengeReader
}

type ChallengeReader interface {
	Read(url string, token string) ([]byte, error)
}

func NewChallengeConfig(env *Environment, reader ChallengeReader) (*ChallengeConfig, error) {
	url := challengeURL(env.GithubOwner, env.GithubRepo)
	contents, err := reader.Read(url, env.GithubToken)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	challengeConfig := &ChallengeConfig{}
	err = yaml.Unmarshal(contents, challengeConfig)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	challengeConfig.Owner = env.GithubOwner
	challengeConfig.TrackingRepoName = env.GithubRepo
	challengeConfig.GithubToken = env.GithubToken
	challengeConfig.SlackBotToken = env.BotToken
	challengeConfig.SlackVerificationToken = env.VerificationToken
	challengeConfig.reader = reader
	return challengeConfig, err
}

func (config *ChallengeConfig) FindChallenge(discipline string) (*Challenge, error) {
	for _, challenge := range config.Challenges {
		if challenge.Discipline == discipline {
			return &challenge, nil
		}
	}
	return nil, errors.New("Unknown discipline")
}

func (config *ChallengeConfig) AllDisciplines() []string {
	disciplines := make([]string, len(config.Challenges))
	for i, v := range config.Challenges {
		disciplines[i] = v.Discipline
	}
	return disciplines
}

func (config *ChallengeConfig) AccountName() string {
	if config.Organization != "" {
		return config.Organization
	} else {
		return config.Owner
	}
}

func (config *ChallengeConfig) TemplateRepositoryURL(discipline string) (string, error) {
	challenge, err := config.FindChallenge(discipline)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("https://github.com/%v/%v.git", config.AccountName(), challenge.TemplateRepoName)
	return url, nil
}

func (config *ChallengeConfig) LoadTask(discipline string, level int) (string, string, error) {
	challenge, err := config.FindChallenge(discipline)
	if err != nil {
		log.Println("[ERROR] Invalid challenge discipline ", discipline)
		return "", "", err
	}

	if level >= len(challenge.Tasks) {
		log.Println("[ERROR] No task specified for the level ", level)
		return "", "", err
	}

	task := challenge.Tasks[level]
	url := taskDescriptionURL(config.Owner, config.TrackingRepoName, task.DescriptionFile)
	taskContents, err := config.reader.Read(url, config.GithubToken)
	if err != nil {
		log.Println("[ERROR] Cannot read task contents ", err)
		return "", task.Title, err
	}
	return string(taskContents), task.Title, nil
}

func challengeURL(owner string, repo string) string {
	// Ex: https://api.github.com/repos/xing/coding-challenges/contents/challenge.yaml
	return fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/challenge.yaml", owner, repo)
}

func taskDescriptionURL(owner string, repo string, filepath string) string {
	return fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", owner, repo, filepath)
}
