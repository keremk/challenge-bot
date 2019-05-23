// Copyright © 2019 Kerem Karatal
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package repo

import (
	"context"
	"log"
	"net/http"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type githubOps struct {
	usingOAuth bool
	token      string
	transport  *ghinstallation.Transport
}

func newGithubOps(token, privateKeyFilename string) (githubOps, error) {
	// Shared transport to reuse TCP connections.
	tr := http.DefaultTransport

	// Wrap the shared transport for use with the integration ID 1 authenticating with installation ID 99.
	itr, err := ghinstallation.NewKeyFromFile(tr, 1, 1010483, privateKeyFilename)
	if err != nil {
		log.Fatal(err)
	}

	return githubOps{
		usingOAuth: true,
		token:      token,
		transport:  itr,
	}, nil
}

func (ctx githubOps) checkUser(githubAlias string) bool {
	client, context := ctx.getClient()
	_, response, err := client.Users.Get(context, githubAlias)
	if err != nil {
		log.Println(err)
	}
	if response.StatusCode == 404 {
		return false
	}
	return true
}

// createRepository call always take the Organization name, it's implementation takes into account
// if the organization is an empty string and creates a different url altogether
func (ctx githubOps) createRepository(repoName string, organization string) (string, error) {
	private := true
	repositoryInput := github.Repository{
		Name:    &repoName,
		Private: &private,
	}

	client, context := ctx.getClient()
	repository, _, err := client.Repositories.Create(context, organization, &repositoryInput)
	if err != nil {
		return "", err
	}

	return *repository.CloneURL, nil
}

func (ctx githubOps) createIssue(issue Issue, accountName string, repoName string) error {
	issueRequest := github.IssueRequest{
		Title:  &issue.Title,
		Body:   &issue.Description,
		Labels: &[]string{issue.Discipline},
	}

	client, context := ctx.getClient()
	_, _, err := client.Issues.Create(context, accountName, repoName, &issueRequest)
	return err
}

func (ctx githubOps) addCollaborator(githubName string, accountName string, repoName string) error {
	client, context := ctx.getClient()
	options := github.RepositoryAddCollaboratorOptions{
		Permission: "push",
	}

	_, err := client.Repositories.AddCollaborator(context, accountName, repoName, githubName, &options)
	return err
}

func (ctx githubOps) pushStarterRepo(templateRepoURL string, remoteRepoURL string) error {
	gitops := &gitOps{
		token: ctx.token,
	}

	return gitops.pushStarterRepo(templateRepoURL, remoteRepoURL)
}

func (ctx githubOps) getClient() (*github.Client, context.Context) {
	if ctx.usingOAuth {
		return ctx.getClientWithToken()
	} else {
		return ctx.getClientWithTransport()
	}
}

func (ctx githubOps) getClientWithToken() (*github.Client, context.Context) {
	context := context.Background()
	tokenService := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: ctx.token},
	)
	tokenClient := oauth2.NewClient(context, tokenService)
	return github.NewClient(tokenClient), context
}

func (ctx githubOps) getClientWithTransport() (*github.Client, context.Context) {
	context := context.Background()

	return github.NewClient(&http.Client{Transport: ctx.transport}), context
}
