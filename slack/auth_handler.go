package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/keremk/challenge-bot/db"
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

type authResponse struct {
	AccessToken string  `json:"access_token"`
	Scope       string  `json:"scope"`
	UserID      string  `json:"user_id"`
	TeamName    string  `json:"team_name"`
	TeamID      string  `json:"team_id"`
	Bot         botInfo `json:"bot"`
}

type authHandler struct {
	env *config.Environment
}

func (handler authHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	slackCode := r.URL.Query().Get("code")
	resp, err := handler.callbackSlack(slackCode)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	authResponse, err := handler.readAndParse(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fmt.Println(authResponse)

	err = handler.saveToDB(authResponse)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (handler authHandler) callbackSlack(slackCode string) (*http.Response, error) {
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	url, err := url.Parse("https://slack.com/api/oauth.access?")
	if err != nil {
		log.Fatal("[ERROR] Unexpected error in parsing hard coded URL?!?", err)
	}
	query := url.Query()
	query.Set("code", slackCode)
	query.Set("client_id", handler.env.SlackClientID)
	query.Set("client_secret", handler.env.SlackClientSecret)
	query.Set("redirect_uri", handler.env.SlackRedirectURI)

	url.RawQuery = query.Encode()

	resp, err := client.Get(url.String())
	if err != nil {
		log.Println("[ERROR] Cannot reach Slack for OAuth - ", err)
	}
	return resp, err
}

func (handler authHandler) readAndParse(resp *http.Response) (authResponse, error) {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(resp.Body)
	if err != nil {
		log.Println("[ERROR] Cannot read the response - ", err)
		return authResponse{}, err
	}

	log.Println(buf.String())

	var response authResponse
	err = json.Unmarshal(buf.Bytes(), &response)
	if err != nil {
		log.Println("[ERROR] Cannot parse the json - ", err)
		return authResponse{}, err
	}

	return response, nil
}

func (handler authHandler) saveToDB(resp authResponse) error {
	slackUser := &models.SlackUser{
		SlackID:    resp.UserID,
		SlackToken: resp.AccessToken,
	}
	slackTeam := &models.SlackTeam{
		SlackBotToken:  resp.Bot.BotAccessToken,
		SlackBotUserID: resp.Bot.BotUserID,
		SlackID:        resp.TeamID,
		Name:           resp.TeamName,
	}

	usersDb := db.NewStore(*handler.env, db.SlackUsersCollection)
	err := usersDb.Update(slackUser.SlackID, slackUser)
	if err != nil {
		return err
	}

	teamsDb := db.NewStore(*handler.env, db.SlackTeamsCollection)
	err = teamsDb.Update(slackTeam.SlackID, slackTeam)
	return err
}
