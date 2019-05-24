package controllers

import (
	"log"
	"net/http"
	"net/url"

	"github.com/keremk/challenge-bot/config"
)

type ghSetupHandler struct {
	env config.Environment
}

func (gh ghSetupHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("[INFO] Setup called with - ", r.URL.String())
	installationID := r.URL.Query().Get("installation_id")
	if installationID == "" {
		log.Println("[ERROR] No installation_id received from github")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	userAuthURI := gh.getUserAuthURI(installationID)

	http.Redirect(w, r, userAuthURI, http.StatusFound)
}

func (gh ghSetupHandler) getUserAuthURI(installationID string) string {
	uri, err := url.Parse("https://github.com/login/oauth/authorize?")
	if err != nil {
		log.Fatal("[ERROR] Unexpected error in parsing hard coded URL?!?", err)
	}
	query := uri.Query()
	query.Set("client_id", url.QueryEscape(gh.env.GithubClientID))
	query.Set("redirect_uri", gh.env.GithubRedirectURI)
	query.Set("state", installationID)

	uri.RawQuery = query.Encode()
	return uri.String()
}
