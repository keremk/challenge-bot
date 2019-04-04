package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/keremk/challenge-bot/config"
	"github.com/nlopes/slack"
	slackApi "github.com/nlopes/slack"
)

type requestHandler struct {
	slackClient     *slackApi.Client
	env             *config.Environment
	challengeConfig *config.ChallengeConfig
}

func (handler *requestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	response := buf.String()
	payload := strings.TrimLeft(response, "payload=")
	unescapedPayload, _ := url.QueryUnescape(payload)

	var interactionCB slack.InteractionCallback
	err = json.Unmarshal([]byte(unescapedPayload), &interactionCB)
	if err != nil {
		fmt.Println(err, unescapedPayload)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if interactionCB.Token != handler.env.VerificationToken {
		fmt.Println("Invalid token ", interactionCB.Token)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	fmt.Println(interactionCB.Submission)
	w.WriteHeader(http.StatusAccepted)
}
