package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/kelseyhightower/envconfig"
	"github.com/nlopes/slack"
	"github.com/nlopes/slack/slackevents"
)

type config struct {
	Port              string `envconfig:"PORT" default:"4390"`
	BotToken          string `envconfig:"BOT_TOKEN" required:"true"`
	VerificationToken string `envconfig:"VERIFICATION_TOKEN" required:"true"`
}

var env config

func exampleOne() slack.MsgOption {

	// Header Section
	headerText := slack.NewTextBlockObject("mrkdwn", "You have a new request:\n*<fakeLink.toEmployeeProfile.com|Fred Enriquez - New device request>*", false, false)
	headerSection := slack.NewSectionBlock(headerText, nil, nil)

	// Fields
	typeField := slack.NewTextBlockObject("mrkdwn", "*Type:*\nComputer (laptop)", false, false)
	whenField := slack.NewTextBlockObject("mrkdwn", "*When:*\nSubmitted Aut 10", false, false)
	lastUpdateField := slack.NewTextBlockObject("mrkdwn", "*Last Update:*\nMar 10, 2015 (3 years, 5 months)", false, false)
	reasonField := slack.NewTextBlockObject("mrkdwn", "*Reason:*\nAll vowel keys aren't working.", false, false)
	specsField := slack.NewTextBlockObject("mrkdwn", "*Specs:*\n\"Cheetah Pro 15\" - Fast, really fast\"", false, false)

	fieldSlice := make([]*slack.TextBlockObject, 0)
	fieldSlice = append(fieldSlice, typeField)
	fieldSlice = append(fieldSlice, whenField)
	fieldSlice = append(fieldSlice, lastUpdateField)
	fieldSlice = append(fieldSlice, reasonField)
	fieldSlice = append(fieldSlice, specsField)

	fieldsSection := slack.NewSectionBlock(nil, fieldSlice, nil)

	// Approve and Deny Buttons
	approveBtnTxt := slack.NewTextBlockObject("plain_text", "Approve", false, false)
	approveBtn := slack.NewButtonBlockElement("", "click_me_123", approveBtnTxt)

	denyBtnTxt := slack.NewTextBlockObject("plain_text", "Deny", false, false)
	denyBtn := slack.NewButtonBlockElement("", "click_me_123", denyBtnTxt)

	actionBlock := slack.NewActionBlock("", approveBtn, denyBtn)

	return slack.MsgOptionBlocks(
		headerSection,
		fieldsSection,
		actionBlock,
	)
}

func slackEventHandler(w http.ResponseWriter, r *http.Request) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	body := buf.String()
	eventsAPIEvent, e := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionVerifyToken(&slackevents.TokenComparator{VerificationToken: env.VerificationToken}))
	if e != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	if eventsAPIEvent.Type == slackevents.URLVerification {
		var r *slackevents.ChallengeResponse
		err := json.Unmarshal([]byte(body), &r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Header().Set("Content-Type", "text")
		w.Write([]byte(r.Challenge))
	}
	if eventsAPIEvent.Type == slackevents.CallbackEvent {
		innerEvent := eventsAPIEvent.InnerEvent
		switch ev := innerEvent.Data.(type) {
		case *slackevents.AppMentionEvent:
			api := slack.New(env.BotToken)
			api.PostMessage(ev.Channel, exampleOne())
		}
	}
}

func slackCommandHandler(w http.ResponseWriter, r *http.Request) {
	s, err := slack.SlashCommandParse(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !s.ValidateToken(env.VerificationToken) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	switch s.Command {
	case "/challenge":
		if s.TriggerID == "" {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Println("No trigger ID given")
			return
		}

		githubNameElement := slack.NewTextInput("githubAlias", "Github Alias", "")
		options := []slack.DialogSelectOption{
			{Label: "android", Value: "android"},
			{Label: "ios", Value: "ios"},
			{Label: "backend", Value: "backend"},
		}
		disciplineSelectElement := slack.NewStaticSelectDialogInput("discipline", "Discipline", options)

		elements := []slack.DialogElement{
			githubNameElement,
			disciplineSelectElement,
		}

		dialog := &slack.Dialog{
			TriggerID:      s.TriggerID,
			CallbackID:     "challenge_67e2d0",
			Title:          "Create Coding Challenge",
			SubmitLabel:    "Create",
			NotifyOnCancel: false,
			Elements:       elements,
		}

		w.WriteHeader(http.StatusOK)
		api := slack.New(env.BotToken)
		err := api.OpenDialog(s.TriggerID, *dialog)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("TriggerID", s.TriggerID)
		return
	default:
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func slackRequestHandler(w http.ResponseWriter, r *http.Request) {
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

	if interactionCB.Token != env.VerificationToken {
		fmt.Println("Invalid token ", interactionCB.Token)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	fmt.Println(interactionCB.Submission)
	w.WriteHeader(http.StatusAccepted)
}

func main() {
	err := envconfig.Process("", &env)
	if err != nil {
		fmt.Println("[ERROR] Can not read environment variables")
		os.Exit(1)
	}

	fmt.Println("BOT_TOKEN=", env.BotToken)
	fmt.Println("VERIFICATION_TOKEN=", env.VerificationToken)

	http.HandleFunc("/events-endpoint", slackEventHandler)
	http.HandleFunc("/commands-endpoint", slackCommandHandler)
	http.HandleFunc("/requests", slackRequestHandler)
	fmt.Println("[INFO] Server listening")
	http.ListenAndServe(":4390", nil)
}
