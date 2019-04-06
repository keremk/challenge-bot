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
			fmt.Println("[Error] No trigger ID given")
			return
		}

		// Immediately return
		w.WriteHeader(http.StatusOK)

		// Create the dialog and send a message to open it
		dialog := createDialogWithChallengeOptions(s.TriggerID, s.ChannelID, handler.challengeConfig.AllDisciplines())
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

func createDialogWithChallengeOptions(triggerID string, channelID string, options []string) *slackApi.Dialog {
	candidateNameElement := slackApi.NewTextInput("candidate_name", "Candidate Name", "")
	githubNameElement := slackApi.NewTextInput("github_alias", "Github Alias", "")
	resumeURLElement := slackApi.NewTextInput("resume_URL", "Resume URL", "")
	selectOptions := make([]slackApi.DialogSelectOption, len(options))
	for i, v := range options {
		selectOptions[i] = slackApi.DialogSelectOption{
			Label: v,
			Value: v,
		}
	}
	disciplineSelectElement := slackApi.NewStaticSelectDialogInput("challenge_template", "Challenge Template", selectOptions)

	elements := []slackApi.DialogElement{
		candidateNameElement,
		githubNameElement,
		resumeURLElement,
		disciplineSelectElement,
	}

	return &slackApi.Dialog{
		TriggerID:      triggerID,
		CallbackID:     "challenge_67e2d0",
		Title:          "Create Coding Challenge",
		SubmitLabel:    "Create",
		NotifyOnCancel: false,
		State:          channelID,
		Elements:       elements,
	}
}
