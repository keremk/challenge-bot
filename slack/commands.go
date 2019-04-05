package slack

import (
	"fmt"
	"net/http"

	"github.com/keremk/challenge-bot/config"
	slackApi "github.com/nlopes/slack"
)

type commandHandler struct {
	slackClient     *slackApi.Client
	env             *config.Environment
	challengeConfig *config.ChallengeConfig
}

func (handler *commandHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s, err := slackApi.SlashCommandParse(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !s.ValidateToken(handler.env.VerificationToken) {
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
		candidateNameElement := slackApi.NewTextInput("candidate_name", "Candidate Name", "")
		githubNameElement := slackApi.NewTextInput("github_alias", "Github Alias", "")
		resumeURLElement := slackApi.NewTextInput("resume_URL", "Resume URL", "")
		options := []slackApi.DialogSelectOption{
			{Label: "android", Value: "android"},
			{Label: "ios", Value: "ios"},
			{Label: "backend", Value: "backend"},
		}
		disciplineSelectElement := slackApi.NewStaticSelectDialogInput("challenge_template", "Challenge Template", options)

		elements := []slackApi.DialogElement{
			candidateNameElement,
			githubNameElement,
			resumeURLElement,
			disciplineSelectElement,
		}

		dialog := &slackApi.Dialog{
			TriggerID:      s.TriggerID,
			CallbackID:     "challenge_67e2d0",
			Title:          "Create Coding Challenge",
			SubmitLabel:    "Create",
			NotifyOnCancel: false,
			State:          s.ChannelID,
			Elements:       elements,
		}

		w.WriteHeader(http.StatusOK)
		err := handler.slackClient.OpenDialog(s.TriggerID, *dialog)
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
