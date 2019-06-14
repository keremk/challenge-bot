package slackops

import (
	"github.com/keremk/challenge-bot/config"
	"github.com/nlopes/slack"
)

func executeChallengeHelp(env config.Environment, c command) error {
	s := c.slashCommand

	err := postMessage(env, s.TeamID, s.ChannelID, renderChallengeHelp())
	return err
}

func executeSendChallenge(env config.Environment, c command) error {
	s := c.slashCommand

	dialog := sendChallengeDialog(s.TriggerID, c.arg)

	return showDialog(env, s.TeamID, s.TriggerID, dialog)
}

func executeNewChallenge(env config.Environment, c command) error {
	s := c.slashCommand

	dialog := newChallengeDialog(s.TriggerID)

	return showDialog(env, s.TeamID, s.TriggerID, dialog)
}

func sendChallengeDialog(triggerID string, challengeName string) slack.Dialog {
	candidateNameElement := slack.NewTextInput("candidate_name", "Candidate Name", "")
	githubNameElement := slack.NewTextInput("github_alias", "Github Alias", "")
	resumeURLElement := slack.NewTextInput("resume_URL", "Resume URL", "")
	challengeNameElement := newExternalOptionsDialogInput("challenge_name", "Challenge Name", "", false)

	reviewer1OptionsElement := newExternalOptionsDialogInput("reviewer1_id", "Reviewer 1", challengeName, true)
	reviewer2OptionsElement := newExternalOptionsDialogInput("reviewer2_id", "Reviewer 2", challengeName, true)

	elements := []slack.DialogElement{
		candidateNameElement,
		githubNameElement,
		resumeURLElement,
		challengeNameElement,
		reviewer1OptionsElement,
		reviewer2OptionsElement,
	}

	return slack.Dialog{
		TriggerID:      triggerID,
		CallbackID:     "send_challenge",
		Title:          "Send Coding Challenge",
		SubmitLabel:    "Send",
		NotifyOnCancel: false,
		State:          challengeName,
		Elements:       elements,
	}
}

func newChallengeDialog(triggerID string) slack.Dialog {
	challengeNameEl := slack.NewTextInput("challenge_name", "Challenge Name", "")
	templateRepoNameEl := slack.NewTextInput("template_repo", "Template Repo Name", "")
	repoNameFormatEl := slack.NewTextInput("repo_name_format", "Repo Name Format", "test_CHALLENGENAME-GITHUBALIAS")

	githubAccountEl := newExternalOptionsDialogInput("github_account", "Github Account Name", "", false)
	elements := []slack.DialogElement{
		challengeNameEl,
		templateRepoNameEl,
		repoNameFormatEl,
		githubAccountEl,
	}
	return slack.Dialog{
		TriggerID:      triggerID,
		CallbackID:     "new_challenge",
		Title:          "New Coding Challenge",
		SubmitLabel:    "Create",
		NotifyOnCancel: false,
		Elements:       elements,
	}
}
