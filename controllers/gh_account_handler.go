package controllers

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/keremk/challenge-bot/config"
	"github.com/keremk/challenge-bot/models"
)

type ghAccountHandler struct {
	env config.Environment
}

type ghAccountInfo struct {
	installationID string
	org            string
	owner          string
	name           string
}

func (gh ghAccountHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	info, err := gh.parseAndValidate(r)
	if err != nil {
		url := fmt.Sprintf("/auth/github/error.html?error=%s", err)
		http.Redirect(w, r, url, http.StatusFound)
		return
	}
	err = gh.saveToDB(info)
	if err != nil {
		url := fmt.Sprintf("/auth/github/error.html?error=%s", err)
		http.Redirect(w, r, url, http.StatusFound)
		return
	}
	http.Redirect(w, r, "/auth/github/success.html", http.StatusFound)
}

func (gh ghAccountHandler) parseAndValidate(r *http.Request) (ghAccountInfo, error) {
	r.ParseForm()
	log.Println("[INFO] Github Account Handler called with - ", r.Form)
	values := r.Form
	if len(values["installationID"]) == 0 {
		return ghAccountInfo{}, errors.New("Installation ID not provided")
	}
	installationID := values["installationID"][0]
	var org string
	if len(values["organizationName"]) > 0 {
		org = values["organizationName"][0]
	}
	var owner string
	if len(values["accountName"]) > 0 {
		owner = values["accountName"][0]
	}

	if owner == "" && org == "" {
		return ghAccountInfo{}, errors.New("Either organization or account name needs to be provided")
	}

	var name string
	if org != "" {
		name = org
	} else {
		name = owner
	}

	return ghAccountInfo{
		installationID: installationID,
		owner:          owner,
		org:            org,
		name:           name,
	}, nil
}

func (gh ghAccountHandler) saveToDB(info ghAccountInfo) error {
	return models.EditGithubAccount(gh.env, info.installationID, info.org, info.owner, info.name)
}
