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
	"log"
	"os"

	git "gopkg.in/src-d/go-git.v4"
	gitConfig "gopkg.in/src-d/go-git.v4/config"
	auth "gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

type gitOps struct {
	token string
}

func (ctx gitOps) pushStarterRepo(templateRepoURL string, remoteRepoURL string) error {
	repository, err := ctx.cloneRepository(templateRepoURL)
	if err != nil {
		log.Println("Cannot clone repository")
		return err
	}

	return ctx.createAndPushToRemote(remoteRepoURL, repository)
}

func (ctx gitOps) cloneRepository(repoURL string) (*git.Repository, error) {
	repository, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		Auth: &auth.BasicAuth{
			Username: "abc123", // yes, this can be anything except an empty string
			Password: ctx.token,
		},
		URL:      repoURL,
		Progress: os.Stdout,
	})

	return repository, err
}

func (ctx gitOps) createAndPushToRemote(remoteRepoURL string, repository *git.Repository) error {
	newRepo, err := repository.CreateRemote(&gitConfig.RemoteConfig{
		Name: "candidate",
		URLs: []string{remoteRepoURL},
	})

	if err != nil {
		log.Println("Cannot create a remote")
		return err
	}

	return newRepo.Push(&git.PushOptions{
		RemoteName: "candidate",
		Auth: &auth.BasicAuth{
			Username: "abc123", // yes, this can be anything except an empty string
			Password: ctx.token,
		},
		Progress: os.Stdout,
	})
}
