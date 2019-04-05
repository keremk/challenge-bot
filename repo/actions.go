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
	"io/ioutil"

	"github.com/keremk/challenge-bot/config"
)

// Creates a coding challenge for a given candidate and challenge type.
// The coding challenge is created based on the configuration settings the .challenge.yaml file
func CreateChallenge(candidateName string, discipline string, challengeConfig config.ChallengeConfig, token string) (string, error) {
	fmt.Println("Creating coding challenge")

	repoName := generateChallengeRepositoryName(candidateName, discipline)
	fmt.Println("Challenge repo name: ", repoName)
	fmt.Println("Organization: ", challengeConfig.Organization)
	fmt.Println("Owner: ", challengeConfig.Owner)
	fmt.Println("Tracking Repo Name:", challengeConfig.TrackingRepoName)

	// createRepository call always take the Organization name, it's implementation takes into account
	// if the organization is an empty string and creates a different url altogether
	challengeRepoURL, err := createRepository(repoName, challengeConfig.Organization, token)
	if err != nil {
		fmt.Println("Cannot create the challenge repository")
		fmt.Println(err)
		return "", err
	}
	fmt.Println("Created: ", challengeRepoURL)

	challenge, err := challengeConfig.FindChallenge(discipline)
	if err != nil {
		fmt.Println("Invalid challenge discipline ", discipline)
		return challengeRepoURL, err
	}

	templateRepoURL := generateTemplateRepositoryURL(challengeConfig.Owner, challengeConfig.Organization, challenge.TemplateRepoName)
	fmt.Println(templateRepoURL)

	err = pushStarterProject(templateRepoURL, challengeRepoURL, token)

	if err != nil {
		fmt.Println("Could not push the starter project")
		fmt.Println(err)
		return challengeRepoURL, err
	}

	err = createCandidateTask(candidateName, discipline, 0, challengeConfig, token)
	if err != nil {
		fmt.Println("Could not create candidate task")
		fmt.Println(err)
		return challengeRepoURL, err
	}

	err = createTrackingIssue(candidateName, discipline, challengeRepoURL, challengeConfig, token)
	if err != nil {
		fmt.Println("Could not create tracking issue")
		fmt.Println(err)
		return challengeRepoURL, err
	}

	accountName := ownerOrOrganization(challengeConfig.Owner, challengeConfig.Organization)
	err = addCollaborator(candidateName, accountName, repoName, token)

	if err != nil {
		fmt.Println("Cannot add the candidate as a collaborator ", candidateName)
		fmt.Println(err)
		return challengeRepoURL, err
	}

	fmt.Println("Challenge created successfully.")
	return challengeRepoURL, nil
}

func createCandidateTask(candidateName string, discipline string, level int, challengeConfig config.ChallengeConfig, token string) error {
	fmt.Println("Creating candidate task")

	challenge, err := challengeConfig.FindChallenge(discipline)
	if err != nil {
		fmt.Println("Invalid challenge discipline ", discipline)
		return err
	}

	if level >= len(challenge.Tasks) {
		fmt.Println("No task specified for the level ", level)
		return err
	}

	task := challenge.Tasks[level]
	descriptionFilePath, err := generateTaskDescriptionFilePath(challenge.Tasks[level].DescriptionFile)
	if err != nil {
		fmt.Println("Cannot create the description file path")
		fmt.Println(err)
		return err
	}

	description, err := readDescription(descriptionFilePath)
	if err != nil {
		fmt.Println("Aborting task creation")
		fmt.Println(err)
		return err
	}

	issue := Issue{
		Title:       task.Title,
		Discipline:  discipline,
		Description: description,
	}

	repoName := generateChallengeRepositoryName(candidateName, discipline)
	accountName := ownerOrOrganization(challengeConfig.Owner, challengeConfig.Organization)
	err = createIssue(issue, accountName, repoName, token)

	if err != nil {
		fmt.Println("Could not create a candidate task at ", repoName)
		fmt.Println(err)
		return err
	}
	fmt.Println("Candidate task created at: ", repoName)
	return nil
}

func createTrackingIssue(candidateName string, discipline string, challengeRepoURL string, challengeConfig config.ChallengeConfig, token string) error {
	title := "Coding Challenge for: " + candidateName
	description := "Coding challenge is located at: " + challengeRepoURL

	issue := Issue{
		Title:       title,
		Discipline:  discipline,
		Description: description,
	}
	accountName := ownerOrOrganization(challengeConfig.Owner, challengeConfig.Organization)
	err := createIssue(issue, accountName, challengeConfig.TrackingRepoName, token)

	if err != nil {
		fmt.Println("Could not create a tracking issue at ", challengeConfig.TrackingRepoName)
		fmt.Println(err)
		return err
	}
	fmt.Println("Tracking issue created at: ", challengeConfig.TrackingRepoName)
	return nil
}

func readDescription(filePath string) (string, error) {
	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Println("There is no file or cannot read in location: ", filePath)
		return "", err
	}

	return string(bytes[:]), nil
}
