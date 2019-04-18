package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockChallengeConfigReader struct {
	t                 *testing.T
	checkConfigURL    bool
	checkTaskURL      bool
	expectedConfigURL string
	expectedTaskURL   string
}

func (c *MockChallengeConfigReader) Read(url string, token string) ([]byte, error) {
	var output string
	if strings.HasSuffix(url, "challenge.yaml") {
		if c.checkConfigURL {
			assert.Equal(c.t, c.expectedConfigURL, url, "Config URL is not correct")
		}
		// Mock returns the challenge.yaml
		output = `
organization: "ORG"
challenges:
  - discipline: android
    templateRepoName: challenge-test
    reviewers:
      - reviewer1
      - reviewer2
    tasks:
      - level: 1
        title: "Do this first task"
        descriptionFile: "test/android/task-1.md"
      - level: 2
        title: "Now do the second task"
        descriptionFile: "test/android/task-2.md"
  - discipline: ios
    templateRepoName: challenge-test
`
	} else {
		// Assert the url
		if c.checkTaskURL {
			assert.Equal(c.t, c.expectedTaskURL, url, "Task URL is not correct")
		}
		// Mock returns a task description
		output = `
## Task
Your first task consists of reading this document. Please read it!
`
	}

	return []byte(output), nil
}

func TestLoadingChallengeConfig(t *testing.T) {
	env := NewEnvironment("unittest")
	mockReader := &MockChallengeConfigReader{
		t:                 t,
		checkConfigURL:    true,
		checkTaskURL:      false,
		expectedConfigURL: "https://api.github.com/repos/Owner/challenge-bot/contents/challenge.yaml",
	}

	challengeConfig, err := NewChallengeConfig(env, mockReader)

	assert.Nil(t, err, "Unexpected error")
	assert.NotNil(t, challengeConfig, "No challenge configuration")

	assert.Equal(t, "Owner", challengeConfig.Owner, "Owner not configured")
	assert.Equal(t, "ORG", challengeConfig.Organization, "Organization not configured")
	assert.Equal(t, "challenge-bot", challengeConfig.TrackingRepoName, "TrackingRepoName not configured")
	assert.Equal(t, "FakeToken", challengeConfig.GithubToken, "GithubToken not configured")
	assert.Equal(t, "FakeToken", challengeConfig.SlackBotToken, "SlackBotToken not configured")
	assert.Equal(t, "FakeToken", challengeConfig.SlackVerificationToken, "SlackVerificationToken not configured")

	assert.Equal(t, 2, len(challengeConfig.Challenges))
	assert.Equal(t, "android", challengeConfig.Challenges[0].Discipline)
	assert.Equal(t, "challenge-test", challengeConfig.Challenges[0].TemplateRepoName)

	assert.Equal(t, 2, len(challengeConfig.Challenges[0].Reviewers))
	assert.Equal(t, "reviewer1", challengeConfig.Challenges[0].Reviewers[0])
	assert.Equal(t, "reviewer2", challengeConfig.Challenges[0].Reviewers[1])

	assert.Equal(t, 2, len(challengeConfig.Challenges[0].Tasks))
	assert.Equal(t, int8(1), challengeConfig.Challenges[0].Tasks[0].Level)
	assert.Equal(t, "Do this first task", challengeConfig.Challenges[0].Tasks[0].Title)
	assert.Equal(t, "test/android/task-1.md", challengeConfig.Challenges[0].Tasks[0].DescriptionFile)
}

func TestFindingChallenge(t *testing.T) {
	env := NewEnvironment("unittest")
	mockReader := &MockChallengeConfigReader{
		t:              t,
		checkConfigURL: false,
		checkTaskURL:   false,
	}

	challengeConfig, err := NewChallengeConfig(env, mockReader)

	challenge, err := challengeConfig.FindChallenge("android")

	assert.Nil(t, err, "Unexpected error")
	assert.NotNil(t, challenge, "Cannot find the challenge")

	challenge, err = challengeConfig.FindChallenge("foo")
	assert.NotNil(t, err, "Error was expected")
	assert.Nil(t, challenge, "There should be no challenge called foo")
}

func TestGettingAllDisciplines(t *testing.T) {
	env := NewEnvironment("unittest")
	mockReader := &MockChallengeConfigReader{
		t:              t,
		checkConfigURL: false,
		checkTaskURL:   false,
	}

	challengeConfig, _ := NewChallengeConfig(env, mockReader)

	disciplines := challengeConfig.AllDisciplines()

	assert.Equal(t, 2, len(disciplines))
	assert.Equal(t, "android", disciplines[0])
	assert.Equal(t, "ios", disciplines[1])
}

func TestChallengeRepoURL(t *testing.T) {
	env := NewEnvironment("unittest")
	mockReader := &MockChallengeConfigReader{
		t:              t,
		checkConfigURL: false,
		checkTaskURL:   false,
	}
	challengeConfig, _ := NewChallengeConfig(env, mockReader)

	url, err := challengeConfig.TemplateRepositoryURL("android")
	assert.Nil(t, err, "Unexpected error")
	assert.Equal(t, "https://github.com/ORG/challenge-test.git", url, "Template Repository URL is not correct")
}

func TestLoadingTask(t *testing.T) {
	env := NewEnvironment("unittest")
	mockReader := &MockChallengeConfigReader{
		t:               t,
		checkConfigURL:  false,
		checkTaskURL:    true,
		expectedTaskURL: "https://api.github.com/repos/Owner/challenge-bot/contents/test/android/task-1.md",
	}
	challengeConfig, _ := NewChallengeConfig(env, mockReader)

	expectedContents := `
## Task
Your first task consists of reading this document. Please read it!
`
	expectedTitle := "Do this first task"
	contents, title, err := challengeConfig.LoadTask("android", 0)

	assert.Nil(t, err, "Unexpected error")
	assert.Equal(t, expectedTitle, title, "Task Title is not correct")
	assert.Equal(t, expectedContents, contents, "Task contents is not correct")
}
