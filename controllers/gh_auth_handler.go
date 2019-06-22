package controllers

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/keremk/challenge-bot/config"
	"github.com/keremk/challenge-bot/models"
)

type ghAuthResponse struct {
	AccessToken string
	TokenType   string
}

type ghAuthHandler struct {
	env config.Environment
}

func (gh ghAuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("[INFO] Github Auth Handler called with - ", r.URL.String())

	code := r.URL.Query().Get("code")
	if code == "" {
		log.Println("[ERROR] No code received from github")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp, err := gh.callbackGithub(code)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	authResponse, err := gh.readAndParse(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Comment out below for production code
	// log.Println("Token retrieved = ", authResponse.AccessToken)

	installationID := r.URL.Query().Get("state")

	log.Println("[INFO] Installation ID - ", installationID)
	err = gh.saveToDB(authResponse, installationID)
	if err != nil {
		log.Println("[ERROR] Installation not registered - Installation ID - ", installationID)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	url := fmt.Sprintf("/auth/github/account.html?installationID=%s", installationID)
	http.Redirect(w, r, url, http.StatusFound)
}

func (gh ghAuthHandler) callbackGithub(code string) (*http.Response, error) {
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	uri, err := url.Parse("https://github.com/login/oauth/access_token?")
	if err != nil {
		log.Fatal("[ERROR] Unexpected error in parsing hard coded URL?!?", err)
	}
	query := uri.Query()
	query.Set("code", url.QueryEscape(code))
	query.Set("client_id", url.QueryEscape(gh.env.GithubClientID))
	query.Set("client_secret", url.QueryEscape(gh.env.GithubClientSecret))
	query.Set("redirect_uri", gh.env.GithubRedirectURI)

	log.Println("[INFO] Redirect URI is: ", gh.env.GithubRedirectURI)

	uri.RawQuery = query.Encode()
	uriString := uri.String()

	// IMPORTANT: If you log this, regenerate the Client Secret after diagnosing and comment it out again.
	// log.Println("[INFO] The request we are sending to Slack: ", uriString)
	resp, err := client.Post(uriString, "application/vnd.github.machine-man-preview+json", nil)
	if err != nil {
		log.Println("[ERROR] Cannot reach Github for OAuth - ", err)
	}
	return resp, err
}

func (gh ghAuthHandler) readAndParse(resp *http.Response) (ghAuthResponse, error) {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(resp.Body)
	if err != nil {
		log.Println("[ERROR] Cannot read the response - ", err)
		return ghAuthResponse{}, err
	}

	r := strings.Split(buf.String(), "&")
	if len(r) < 2 {
		log.Println("[ERROR] Unexpected auth response")
		return ghAuthResponse{}, errors.New("[ERROR] Unexpected auth response")
	}

	tokenPair := strings.Split(r[0], "=")
	if len(tokenPair) < 2 {
		log.Println("[ERROR] Unexpected auth response")
		return ghAuthResponse{}, errors.New("[ERROR] Unexpected auth response")
	}

	tokenTypePair := strings.Split(r[1], "=")
	if len(tokenTypePair) < 2 {
		log.Println("[ERROR] Unexpected auth response")
		return ghAuthResponse{}, errors.New("[ERROR] Unexpected auth response")
	}

	return ghAuthResponse{
		AccessToken: tokenPair[1],
		TokenType:   tokenTypePair[1],
	}, nil
}

func (gh ghAuthHandler) saveToDB(resp ghAuthResponse, installationID string) error {
	account := models.NewGithubAccount(installationID, resp.AccessToken)

	err := models.CreateGithubAccount(gh.env, account)
	if err != nil {
		return err
	}
	return nil
}
