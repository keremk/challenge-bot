package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/keremk/challenge-bot/config"
)

/*
 {
    "access_token": "xoxp-XXXXXXXX-XXXXXXXX-XXXXX",
    "scope": "incoming-webhook,commands,bot",
    "team_name": "Team Installing Your Hook",
    "team_id": "XXXXXXXXXX",
    "incoming_webhook": {
        "url": "https://hooks.slack.com/TXXXXX/BXXXXX/XXXXXXXXXX",
        "channel": "#channel-it-will-post-to",
        "configuration_url": "https://teamname.slack.com/services/BXXXXX"
    },
    "bot":{
        "bot_user_id":"UTTTTTTTTTTR",
        "bot_access_token":"xoxb-XXXXXXXXXXXX-TTTTTTTTTTTTTT"
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
	TeamName    string  `json:"team_name"`
	TeamID      string  `json:"team_id"`
	Bot         botInfo `json:"bot"`
}

type authHandler struct {
	env *config.Environment
}

func (handler authHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	url, err := url.Parse("https://slack.com/api/oauth.access?")
	if err != nil {
		log.Fatal(err)
	}
	query := url.Query()
	query.Set("code", r.URL.Query().Get("code"))
	query.Set("client_id", handler.env.SlackClientID)
	query.Set("client_secret", handler.env.SlackClientSecret)
	query.Set("redirect_uri", handler.env.SlackRedirectURI)

	url.RawQuery = query.Encode()

	resp, err := client.Get(url.String())

	if err != nil {
		log.Println("[ERROR] Cannot reach Slack for OAuth - ", err)
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		log.Println("[ERROR] Cannot read the response - ", err)
	}

	var response authResponse
	err = json.Unmarshal(buf.Bytes(), &response)
	if err != nil {
		log.Println("[ERROR] Cannot parse the json - ", err)
	}

	fmt.Println(response)
}
