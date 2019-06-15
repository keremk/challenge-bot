package slackops

import (
	"fmt"
	"log"

	"github.com/keremk/challenge-bot/models"
	"github.com/nlopes/slack"
)

func (c command) executeChallengeHelp() error {
	return c.ctx.postMessage(c.slashCmd.ChannelID, renderChallengeHelp())
}

func (c command) executeSendChallenge() error {
	dialog := sendChallengeDialog(c.slashCmd.TriggerID, c.arg)

	return c.ctx.showDialog(c.slashCmd.TriggerID, dialog)
}

func (c command) executeNewChallenge() error {
	dialog := newChallengeDialog(c.slashCmd.TriggerID)

	return c.ctx.showDialog(c.slashCmd.TriggerID, dialog)
}

func (c command) executeEditChallenge() error {
	var challengeName string
	if c.arg == "" {
		c.ctx.postMessage(c.slashCmd.ChannelID, toMsgOption("You need to provide a challenge name. Please try /challenge edit CHALLENGENAME"))
	} else {
		challengeName = c.arg
	}
	challenge, err := models.GetChallengeSetupByName(c.ctx.Env, challengeName)
	if err != nil {
		log.Println("[ERROR] No such challenge is registered.", err)
		errorMsg := fmt.Sprintf("Challenge named %s is not registered. Please register first using /challenge new command.", challengeName)
		c.ctx.postMessage(c.slashCmd.ChannelID, toMsgOption(errorMsg))
		return err
	}

	dialog := editChallengeDialog(c.slashCmd.TriggerID, challenge)

	return c.ctx.showDialog(c.slashCmd.TriggerID, dialog)
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
	return slack.Dialog{
		TriggerID:      triggerID,
		CallbackID:     "new_challenge",
		Title:          "New Coding Challenge",
		SubmitLabel:    "Create",
		NotifyOnCancel: false,
		Elements: challengeDialogElements(models.ChallengeSetup{
			RepoNameFormat: "test_CHALLENGENAME-GITHUBALIAS",
		}),
	}
}

func editChallengeDialog(triggerID string, challenge models.ChallengeSetup) slack.Dialog {
	return slack.Dialog{
		TriggerID:      triggerID,
		CallbackID:     "edit_challenge",
		Title:          "Edit Coding Challenge",
		SubmitLabel:    "Edit",
		State:          challenge.ID,
		NotifyOnCancel: false,
		Elements:       challengeDialogElements(challenge),
	}
}

func challengeDialogElements(challenge models.ChallengeSetup) []slack.DialogElement {
	challengeNameEl := slack.NewTextInput("challenge_name", "Challenge Name", challenge.Name)
	templateRepoNameEl := slack.NewTextInput("template_repo", "Template Repo Name", challenge.TemplateRepo)
	repoNameFormatEl := slack.NewTextInput("repo_name_format", "Repo Name Format", challenge.RepoNameFormat)

	githubAccountEl := newExternalOptionsDialogInput("github_account", "Github Account Name", "", false)
	return []slack.DialogElement{
		challengeNameEl,
		templateRepoNameEl,
		repoNameFormatEl,
		githubAccountEl,
	}
}
