package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockChallengeConfigReader struct {
	T                 *testing.T
	CheckConfigURL    bool
	CheckTaskURL      bool
	ExpectedConfigURL string
	ExpectedTaskURL   string
}

func (c *MockChallengeConfigReader) Read(url string, token string) ([]byte, error) {
	var output string
	if strings.HasSuffix(url, "challenge.yaml") {
		if c.CheckConfigURL {
			assert.Equal(c.T, c.ExpectedConfigURL, url, "Config URL is not correct")
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
		if c.CheckTaskURL {
			assert.Equal(c.T, c.ExpectedTaskURL, url, "Task URL is not correct")
		}
		// Mock returns a task description
		output = `
## Task
Your first task consists of reading this document. Please read it!
`
	}

	return []byte(output), nil
}
