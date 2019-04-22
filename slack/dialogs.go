package slack

import (
	slackApi "github.com/nlopes/slack"
)

func newChallengeOptionsDialog(triggerID string, channelID string, options []string) *slackApi.Dialog {
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
