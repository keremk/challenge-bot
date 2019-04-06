package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"

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
	Organization     string
	Owner            string
	TrackingRepoName string `yaml:"trackingRepoName"`
	Challenges       []Challenge
}

var configInstance *ChallengeConfig
var configOnce sync.Once

func GetChallengeConfig() *ChallengeConfig {
	configOnce.Do(func() {
		configInstance = &ChallengeConfig{}
		err := configInstance.readContentsFromGithub()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	})
	return configInstance
}

func (config *ChallengeConfig) readContentsFromGithub() error {
	client := &http.Client{}
	env := GetEnvironment()

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/challenge.yaml", env.GithubOwner, env.GithubRepo)

	// Ex: https://api.github.com/repos/xing/coding-challenges/contents/challenge.yaml
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err)
		return err
	}

	token := fmt.Sprintf("token %s", env.GithubToken)
	req.Header.Add("Authorization", token)
	req.Header.Add("Accept", "application/vnd.github.v3.raw")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(string(body))

	err = yaml.Unmarshal(body, config)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(config.Owner)
	return err
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
