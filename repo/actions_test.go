package repo

import (
	"fmt"
	"testing"

	"github.com/keremk/challenge-bot/config"
	"github.com/keremk/challenge-bot/models"
	"github.com/stretchr/testify/assert"
)

// Integration test
func TestCreatingChallenge(t *testing.T) {
	env := config.NewEnvironment("production")
	candidate := models.Candidate{
		Name:          "Foo bar",
		GithubAlias:   "keremktest1",
		ResumeURL:     "http://foo.com",
		ChallengeName: "Test123",
	}
	challenge, _ := models.GetChallengeSetup(env, candidate.ChallengeName)
	fmt.Println(challenge)
	ctx := NewActionContext(env, challenge)

	repoURL, err := ctx.CreateChallenge(candidate, challenge)

	assert.Nil(t, err, "Expected the operation to be nil")
	assert.Equal(t, "https://github.com/xingtesting/test_Test123-keremktest1.git", repoURL, "Unexpected repo url")
}

// type mockRepoOps struct {
// 	t                         *testing.T
// 	assertInputs              bool
// 	createRepositoryAssertion func(repoName, organization string)
// 	createRepositoryCalled    int
// 	createIssueAssertion1     func(issue Issue, accountName string, repoName string)
// 	createIssueAssertion2     func(issue Issue, accountName string, repoName string)
// 	createIssueCalled         int
// 	addCollaboratorAssertion  func(githubName, accountName, repoName string)
// 	addCollaboratorCalled     int
// 	pushStarterRepoAssertion  func(templateRepoURL, remoteRepoURL string)
// 	pushStarterRepoCalled     int
// }

// func (ctx *mockRepoOps) createRepository(repoName string, organization string) (string, error) {
// 	if ctx.assertInputs {
// 		ctx.createRepositoryAssertion(repoName, organization)
// 	}

// 	ctx.createRepositoryCalled++

// 	return "http://github.com/example", nil
// }

// func (ctx *mockRepoOps) createIssue(issue Issue, accountName string, repoName string) error {
// 	if ctx.assertInputs {
// 		if ctx.createIssueCalled == 0 {
// 			ctx.createIssueAssertion1(issue, accountName, repoName)
// 		} else if ctx.createIssueCalled == 1 {
// 			ctx.createIssueAssertion2(issue, accountName, repoName)
// 		}
// 	}

// 	ctx.createIssueCalled++
// 	return nil
// }

// func (ctx *mockRepoOps) addCollaborator(githubName string, accountName string, repoName string) error {
// 	if ctx.assertInputs {
// 		ctx.addCollaboratorAssertion(githubName, accountName, repoName)
// 	}
// 	ctx.addCollaboratorCalled++
// 	return nil
// }

// func (ctx *mockRepoOps) pushStarterRepo(templateRepoURL string, remoteRepoURL string) error {
// 	if ctx.assertInputs {
// 		ctx.pushStarterRepoAssertion(templateRepoURL, remoteRepoURL)
// 	}
// 	ctx.pushStarterRepoCalled++
// 	return nil
// }

// func (ctx *mockRepoOps) checkUser(githubAlias string) bool {

// 	return false
// }

// func NewMockChallengeConfig(t *testing.T) (*config.ChallengeConfig, error) {
// 	env := config.NewEnvironment("unittest")
// 	mockReader := &config.MockChallengeConfigReader{
// 		T:              t,
// 		CheckConfigURL: false,
// 		CheckTaskURL:   false,
// 	}

// 	return config.NewChallengeConfig(env, mockReader)
// }

// func TestCreatingChallenge(t *testing.T) {
// 	mockOps := &mockRepoOps{
// 		t:            t,
// 		assertInputs: true,
// 		createRepositoryAssertion: func(repoName, organization string) {
// 			assert.Equal(t, "test_testuser_android", repoName, "reponame is not correct")
// 			assert.Equal(t, "ORG", organization, "organization is not correct")
// 		},
// 		createRepositoryCalled: 0,
// 		createIssueAssertion1: func(issue Issue, accountName string, repoName string) {
// 			expectedIssue := Issue{
// 				Title: "Do this first task",
// 				Description: `
// ## Task
// Your first task consists of reading this document. Please read it!
// `,
// 				Discipline: "android",
// 			}
// 			assert.Equal(t, expectedIssue, issue, "candidate task is not correct")
// 			assert.Equal(t, "ORG", accountName, "accountname should be the organization name")
// 			assert.Equal(t, "test_testuser_android", repoName, "repo name is not correct")
// 		},
// 		createIssueAssertion2: func(issue Issue, accountName string, repoName string) {
// 			expectedIssue := Issue{
// 				Title: "Coding Challenge for: Test User",
// 				Description: `
// Github Alias: testuser
// Coding Challenge Link: http://github.com/example
// Resume Link: http://example.com/testuser
// `,
// 				Discipline: "android",
// 			}
// 			assert.Equal(t, expectedIssue, issue, "tracking issue is not correct")
// 			assert.Equal(t, "ORG", accountName, "accountname should be the organization name")
// 			assert.Equal(t, "challenge-bot", repoName, "repo name is not correct")
// 		},
// 		createIssueCalled: 0,
// 		addCollaboratorAssertion: func(githubName, accountName, repoName string) {
// 			assert.Equal(t, "ORG", accountName, "accountname should be the organization name")
// 			assert.Equal(t, "testuser", githubName, "repo name is not correct")
// 			assert.Equal(t, "test_testuser_android", repoName, "repo name is not correct")
// 		},
// 		addCollaboratorCalled: 0,
// 		pushStarterRepoAssertion: func(templateRepoURL, remoteRepoURL string) {
// 			assert.Equal(t, "https://github.com/ORG/challenge-test.git", templateRepoURL, "template repo URL is not correct")
// 			assert.Equal(t, "http://github.com/example", remoteRepoURL, "remote repo url is not correct")
// 		},
// 		pushStarterRepoCalled: 0,
// 	}
// 	challengeConfig, _ := NewMockChallengeConfig(t)

// 	ctx := ActionContext{
// 		ChallengeConfig: challengeConfig,
// 		RepoOps:         mockOps,
// 	}

// 	challengeDesc := models.ChallengeDesc{
// 		CandidateName:     "Test User",
// 		ChallengeTemplate: "android",
// 		GithubAlias:       "testuser",
// 		ResumeURL:         "http://example.com/testuser",
// 	}

// 	url, err := ctx.CreateChallenge(challengeDesc)
// 	assert.Nil(t, err, "Unexpected error")
// 	assert.Equal(t, "http://github.com/example", url, "url is not correct")
// 	assert.Equal(t, 1, mockOps.createRepositoryCalled, "create repository operation not called")
// 	assert.Equal(t, 2, mockOps.createIssueCalled, "create issue needs to be called twice")
// 	assert.Equal(t, 1, mockOps.addCollaboratorCalled, "add collaborator is not called")
// 	assert.Equal(t, 1, mockOps.pushStarterRepoCalled, "push starter repo not called")
// }
