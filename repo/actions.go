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
	"fmt"
	"log"

	"github.com/keremk/challenge-bot/models"

	"github.com/keremk/challenge-bot/config"
)

type repoOps interface {
	createRepository(repoName string, organization string) (string, error)
	pushStarterRepo(templateRepoURL string, remoteRepoURL string) error
	addCollaborator(githubName string, accountName string, repoName string) error
	createIssue(issue Issue, accountName string, repoName string) error
	checkUser(githubAlias string) bool
}

type ActionContext struct {
	ops githubOps
}

const challengeRepoFormat = "test_%s_%s"
const starterTaskNo = 0

func NewActionContext(env config.Environment) ActionContext {
	ops := githubOps{
		token: env.GithubToken,
	}
	return ActionContext{
		ops: ops,
	}
}

func (ctx ActionContext) CheckUser(githubAlias string) bool {
	return ctx.ops.checkUser(githubAlias)
}

// Creates a coding challenge for a given candidate and challenge type.
// The coding challenge is created based on the configuration settings the .challenge.yaml file
func (ctx ActionContext) CreateChallenge(candidate models.Candidate, challenge models.Challenge) (string, error) {
	repoName := fmt.Sprintf(challengeRepoFormat, challenge.Name, candidate.GithubAlias)
	challengeRepoURL, err := ctx.createStarterRepo(repoName, challenge)
	if err != nil {
		log.Println("[ERROR] Cannot create starter repo for candidate ", candidate.GithubAlias)
		return "", err
	}

	err = ctx.createTrackingIssue(candidate, challengeRepoURL, challenge)
	if err != nil {
		log.Println("[ERROR] Could not create tracking issue for ", candidate.GithubAlias)
		return challengeRepoURL, err
	}

	err = ctx.addCollaborator(candidate.GithubAlias, repoName, challenge.OrgOrOwner())
	if err != nil {
		log.Println("[ERROR] Cannot add the candidate as a collaborator ", candidate.GithubAlias)
		return challengeRepoURL, err
	}

	log.Println("[INFO] Challenge repo is successfully created and user added.")
	return challengeRepoURL, nil
}

func (ctx ActionContext) createStarterRepo(repoName string, challenge models.Challenge) (string, error) {
	templateRepoURL := challenge.TemplateRepositoryURL()
	organization := challenge.GithubOrg

	challengeRepoURL, err := ctx.ops.createRepository(repoName, organization)
	if err != nil {
		log.Println("[ERROR] Cannot create a new repository, ", err)
		return "", err
	}

	err = ctx.ops.pushStarterRepo(templateRepoURL, challengeRepoURL)
	if err != nil {
		log.Println("[ERROR] Could not push the starter repository, ", err)
		return challengeRepoURL, err
	}

	return challengeRepoURL, nil
}

func (ctx ActionContext) createTrackingIssue(candidate models.Candidate, challengeRepoURL string, challenge models.Challenge) error {
	title := "Coding Challenge for: " + candidate.Name
	descriptionFormat := `
Github Alias: %s
Coding Challenge Link: %s
Resume Link: %s
`

	description := fmt.Sprintf(descriptionFormat, candidate.GithubAlias, challengeRepoURL, candidate.ResumeURL)
	issue := Issue{
		Title:       title,
		Discipline:  challenge.Name,
		Description: description,
	}
	trackingRepoName := challenge.TemplateRepo

	err := ctx.ops.createIssue(issue, challenge.OrgOrOwner(), trackingRepoName)
	if err != nil {
		log.Println("[ERROR] Could not create a tracking issue at ", trackingRepoName)
		return err
	}
	return nil
}

func (ctx ActionContext) addCollaborator(githubAlias string, repoName string, orgOrOwner string) error {
	return ctx.ops.addCollaborator(githubAlias, orgOrOwner, repoName)
}
