package slackops

import (
	"errors"
	"fmt"
	"log"
	"regexp"

	"github.com/keremk/challenge-bot/config"
	"github.com/keremk/challenge-bot/models"
	"github.com/keremk/challenge-bot/repo"

	"github.com/nlopes/slack"
)

func handleSendChallenge(env config.Environment, icb slack.InteractionCallback) error {
	candidate, reviewers, err := parseSendDialogInput(env, icb.Submission)
	if err != nil {
		return err
	}

	state, err := stateFromString(icb.State)
	if err != nil {
		return err
	}
	returnChannel := state.channelID
	teamID := icb.Team.ID

	challenge, err := models.GetChallengeSetup(env, candidate.ChallengeName)
	if err != nil {
		return err
	}
	go sendChallenge(env, challenge, candidate, reviewers, returnChannel, teamID)

	return nil
}

func parseSendDialogInput(env config.Environment, input map[string]string) (models.Candidate, []models.Reviewer, error) {
	candidate := models.NewCandidate(input)
	reviewers := make([]models.Reviewer, 0, 2)
	reviewerLabels := []string{"reviewer1_id", "reviewer2_id"}

	for _, label := range reviewerLabels {
		reviewer, err := resolveReviewer(env, input[label])
		if err == nil {
			reviewers = append(reviewers, reviewer)
		}
	}
	return candidate, reviewers, nil
}

func resolveReviewer(env config.Environment, reviewerSlackID string) (models.Reviewer, error) {
	if reviewerSlackID == "" {
		log.Printf("[ERROR] Reviewer not specified")
		return models.Reviewer{}, errors.New("[ERROR] Reviewer not specified")
	}

	reviewer, err := models.GetReviewerBySlackID(env, reviewerSlackID)
	if err != nil {
		log.Printf("[ERROR] Reviewer ID %s not found in database", reviewerSlackID)
		return reviewer, err
	}

	return reviewer, nil
}

func sendChallenge(env config.Environment, challenge models.ChallengeSetup, candidate models.Candidate, reviewers []models.Reviewer, targetChannel, teamID string) {
	repoCtx := repo.NewActionContext(env, challenge)

	// Check the candidate
	if repoCtx.CheckUser(candidate.GithubAlias) == false {
		errorMsg := fmt.Sprintf("Github Alias %s for candidate %s is not correct", candidate.GithubAlias, candidate.Name)
		postMessage(env, teamID, targetChannel, toMsgOption(errorMsg))
		return
	}

	// Check the reviewers
	for _, reviewer := range reviewers {
		if repoCtx.CheckUser(reviewer.GithubAlias) == false {
			errorMsg := fmt.Sprintf("Github Alias %s for reviewer %s is not correct", reviewer.GithubAlias, reviewer.Name)
			postMessage(env, teamID, targetChannel, toMsgOption(errorMsg))
			return
		}
	}

	// log.Println("[INFO] Reviewer count = ", len(reviewers))

	// Create the challenge
	postMessage(env, teamID, targetChannel, toMsgOption("Please be patient, while I go create a coding challenge for you..."))
	challengeURL, err := repoCtx.CreateChallenge(candidate, challenge, reviewers)
	if err != nil {
		re := regexp.MustCompile(dreadedPrivateRepoError)
		var errorMsg string
		if re.FindStringIndex(err.Error()) != nil {
			errorMsg = fmt.Sprintf("Unable to create challenge. You need to cleanup private repositories, because you exceeded your allowed limit.")
		} else {
			errorMsg = fmt.Sprintf("Unable to create challenge for %s because of ", candidate.Name, err.Error())
		}
		postMessage(env, teamID, targetChannel, toMsgOption(errorMsg))
		return
	}
	postMessage(env, teamID, targetChannel, newChallengeSummary(candidate, challengeURL, challenge.TrackingIssuesURL()))
}

func handleNewChallenge(env config.Environment, icb slack.InteractionCallback) error {
	challengeInput := icb.Submission
	challengeInput["team_id"] = icb.Team.ID

	go createNewChallenge(env, icb.Submission, icb.Team.ID, icb.Channel.ID)
	return nil
}

func createNewChallenge(env config.Environment, params map[string]string, teamID, channelID string) {
	challenge := models.NewChallenge(params)
	err := models.UpdateChallenge(env, challenge)
	if err != nil {
		log.Println("[ERROR] Could not update challenge in db ", err)
		postMessage(env, teamID, channelID, toMsgOption("We were not able to create the new challenge"))
	}

	challengeSetup, err := models.GetChallengeSetup(env, challenge.Name)
	if err != nil {
		log.Println("[ERROR] Could not create a valid challenge setup, perhaps the github repo name is not valid ", err)
		postMessage(env, teamID, channelID, toMsgOption("We were not able to create a valid challenge"))
	}
	msgText := fmt.Sprintf("We created a challenge named %s in our database. It is pointing to: %s", challengeSetup.Name, challengeSetup.TemplateRepositoryURL())
	postMessage(env, teamID, channelID, toMsgOption(msgText))
}

func getUserInfo(env config.Environment, id, teamID string) (slack.User, error) {
	token, err := getBotToken(env, teamID)
	if err != nil {
		return slack.User{}, err
	}

	slackClient := slack.New(token)
	user, err := slackClient.GetUserInfo(id)
	if err != nil {
		log.Println("[ERROR] User info can't be retrieved - ", err)
		return slack.User{}, err
	}

	return *user, nil
}
