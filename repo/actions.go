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
	ChallengeConfig *config.ChallengeConfig
	RepoOps         repoOps
}

const challengeRepoFormat = "test_%s_%s"
const starterTaskNo = 0

func NewActionContext(challengeConfig *config.ChallengeConfig) *ActionContext {
	ops := &githubOps{
		token: challengeConfig.GithubToken,
	}
	return &ActionContext{
		ChallengeConfig: challengeConfig,
		RepoOps:         ops,
	}
}

func (ctx ActionContext) CheckUser(githubAlias string) bool {
	return ctx.RepoOps.checkUser(githubAlias)
}

// Creates a coding challenge for a given candidate and challenge type.
// The coding challenge is created based on the configuration settings the .challenge.yaml file
func (ctx ActionContext) CreateChallenge(candidateName string, discipline string) (string, error) {
	repoName := fmt.Sprintf(challengeRepoFormat, candidateName, discipline)
	challengeRepoURL, err := ctx.createStarterRepo(repoName, discipline)
	if err != nil {
		log.Println("[ERROR] Cannot create starter repo for candidate ", candidateName)
		return "", err
	}

	err = ctx.createCandidateTask(repoName, discipline, starterTaskNo)
	if err != nil {
		log.Println("[ERROR] Can not create candidate task for ", candidateName)
		return challengeRepoURL, err
	}

	err = ctx.createTrackingIssue(candidateName, discipline, challengeRepoURL)
	if err != nil {
		log.Println("[ERROR] Could not create tracking issue for ", candidateName)
		return challengeRepoURL, err
	}

	err = ctx.addCollaborator(candidateName, repoName)
	if err != nil {
		log.Println("[ERROR] Cannot add the candidate as a collaborator ", candidateName)
		return challengeRepoURL, err
	}

	return challengeRepoURL, nil
}

func (ctx ActionContext) createStarterRepo(repoName string, discipline string) (string, error) {
	templateRepoURL, err := ctx.ChallengeConfig.TemplateRepositoryURL(discipline)
	if err != nil {
		log.Println("[ERROR] Cannot find template repository url for ", discipline)
		return "", err
	}

	organization := ctx.ChallengeConfig.Organization

	challengeRepoURL, err := ctx.RepoOps.createRepository(repoName, organization)
	if err != nil {
		log.Println("[ERROR] Cannot create a new repository, ", err)
		return "", err
	}

	err = ctx.RepoOps.pushStarterRepo(templateRepoURL, challengeRepoURL)
	if err != nil {
		log.Println("[ERROR] Could not push the starter repository, ", err)
		return challengeRepoURL, err
	}

	return challengeRepoURL, nil
}

func (ctx ActionContext) createCandidateTask(repoName string, discipline string, level int) error {
	description, title, err := ctx.ChallengeConfig.LoadTask(discipline, level)
	if err != nil {
		return err
	}

	issue := Issue{
		Title:       title,
		Discipline:  discipline,
		Description: description,
	}

	accountName := ctx.ChallengeConfig.AccountName()

	err = ctx.RepoOps.createIssue(issue, accountName, repoName)
	if err != nil {
		return err
	}
	return nil
}

func (ctx ActionContext) createTrackingIssue(candidateName string, discipline string, challengeRepoURL string) error {
	title := "Coding Challenge for: " + candidateName
	descriptionFormat := `
Coding challenge is located at: %s
`

	description := fmt.Sprintf(descriptionFormat, challengeRepoURL)
	issue := Issue{
		Title:       title,
		Discipline:  discipline,
		Description: description,
	}
	trackingRepoName := ctx.ChallengeConfig.TrackingRepoName
	accountName := ctx.ChallengeConfig.AccountName()

	err := ctx.RepoOps.createIssue(issue, accountName, trackingRepoName)
	if err != nil {
		log.Println("[ERROR] Could not create a tracking issue at ", trackingRepoName)
		return err
	}
	return nil
}

func (ctx ActionContext) addCollaborator(githubName string, repoName string) error {
	accountName := ctx.ChallengeConfig.AccountName()
	return ctx.RepoOps.addCollaborator(githubName, accountName, repoName)
}
