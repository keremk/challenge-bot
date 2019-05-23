package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/keremk/challenge-bot/models"

	"github.com/keremk/challenge-bot/config"
)

/*
{
  "ok": true,
  "access_token": "xoxp-XXXXX-XXXXX-XXXXX-XXXX",
  "scope": "identify,bot,commands",
  "user_id": "WG7ALQ7JA",
  "team_name": "Lime",
  "team_id": "TGB941BGQ",
  "bot": {
    "bot_user_id": "WHTE8CSH1",
    "bot_access_token": "xoxb-XXXXX-XXXXXX-XXXXXX"
  }
}
*/

type botInfo struct {
	BotUserID      string `json:"bot_user_id"`
	BotAccessToken string `json:"bot_access_token"`
}

type slAuthResponse struct {
	Ok          bool    `json:"Ok"`
	Error       string  `json:"error"`
	AccessToken string  `json:"access_token"`
	Scope       string  `json:"scope"`
	UserID      string  `json:"user_id"`
	TeamName    string  `json:"team_name"`
	TeamID      string  `json:"team_id"`
	Bot         botInfo `json:"bot"`
}

type authHandler struct {
	env config.Environment
}

func (h authHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	slackCode := r.URL.Query().Get("code")
	if slackCode == "" {
		log.Println("[ERROR] No code received from slack")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp, err := h.callbackSlack(slackCode)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	authResp, err := h.readAndParse(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Println("[INFO] User ID: ", authResp.UserID)
	log.Println("[INFO] Team ID: ", authResp.TeamID)
	log.Println("[INFO] Team Name: ", authResp.TeamName)

	err = h.saveToDB(authResp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/auth/slack/success.html", http.StatusFound)
}

func (h authHandler) callbackSlack(slackCode string) (*http.Response, error) {
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	uri, err := url.Parse("https://slack.com/api/oauth.access?")
	if err != nil {
		log.Fatal("[ERROR] Unexpected error in parsing hard coded URL?!?", err)
	}
	query := uri.Query()
	query.Set("code", url.QueryEscape(slackCode))
	query.Set("client_id", url.QueryEscape(h.env.SlackClientID))
	query.Set("client_secret", url.QueryEscape(h.env.SlackClientSecret))
	query.Set("redirect_uri", h.env.SlackRedirectURI)

	log.Println("[INFO] Redirect URI is: ", h.env.SlackRedirectURI)

	uri.RawQuery = query.Encode()
	uriString := uri.String()

	// IMPORTANT: If you log this, regenerate the Client Secret after diagnosing and comment it out again.
	// log.Println("[INFO] The request we are sending to Slack: ", uriString)
	resp, err := client.Get(uriString)
	if err != nil {
		log.Println("[ERROR] Cannot reach Slack for OAuth - ", err)
	}
	return resp, err
}

func (h authHandler) readAndParse(resp *http.Response) (slAuthResponse, error) {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(resp.Body)
	if err != nil {
		log.Println("[ERROR] Cannot read the response - ", err)
		return slAuthResponse{}, err
	}

	var authResp slAuthResponse
	err = json.Unmarshal(buf.Bytes(), &authResp)
	if err != nil {
		log.Println("[ERROR] Cannot parse the json - ", err)
		return slAuthResponse{}, err
	}

	if authResp.Ok == false {
		log.Println("[ERROR] Did not get expected auth response - ", buf.String())
		return slAuthResponse{}, errors.New("[ERROR] Did not get expected auth response")
	}

	return authResp, nil
}

func (h authHandler) saveToDB(resp slAuthResponse) error {
	slackUser := models.SlackUser{
		ID:    resp.UserID,
		Token: resp.AccessToken,
	}
	err := models.UpdateSlackUser(h.env, slackUser)
	if err != nil {
		return err
	}

	slackTeam := models.SlackTeam{
		BotToken:  resp.Bot.BotAccessToken,
		BotUserID: resp.Bot.BotUserID,
		ID:        resp.TeamID,
		Name:      resp.TeamName,
	}
	return models.UpdateSlackTeam(h.env, slackTeam)
}
