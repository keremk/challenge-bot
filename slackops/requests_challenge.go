package slackops

import (
	"errors"
	"fmt"
	"log"
	"regexp"

	"github.com/keremk/challenge-bot/models"
	"github.com/keremk/challenge-bot/repo"
)

func (r request) handleSendChallenge() error {
	candidate, reviewers, err := r.parseSendDialogInput(r.icb.Submission)
	if err != nil {
		return err
	}

	challenge, err := models.GetChallengeSetupByName(r.ctx.Env, candidate.ChallengeName)
	if err != nil {
		return err
	}
	go r.sendChallenge(challenge, candidate, reviewers)

	return nil
}

func (r request) parseSendDialogInput(input map[string]string) (models.Candidate, []models.Reviewer, error) {
	candidate := models.NewCandidate(input)
	reviewers := make([]models.Reviewer, 0, 2)
	reviewerLabels := []string{"reviewer1_id", "reviewer2_id"}

	for _, label := range reviewerLabels {
		reviewer, err := r.resolveReviewer(input[label])
		if err == nil {
			reviewers = append(reviewers, reviewer)
		}
	}
	return candidate, reviewers, nil
}

func (r request) resolveReviewer(reviewerSlackID string) (models.Reviewer, error) {
	if reviewerSlackID == "" {
		log.Printf("[ERROR] Reviewer not specified")
		return models.Reviewer{}, errors.New("[ERROR] Reviewer not specified")
	}

	reviewer, err := models.GetReviewerBySlackID(r.ctx.Env, reviewerSlackID)
	if err != nil {
		log.Printf("[ERROR] Reviewer ID %s not found in database", reviewerSlackID)
		return reviewer, err
	}

	return reviewer, nil
}

func (r request) sendChallenge(challenge models.ChallengeSetup, candidate models.Candidate, reviewers []models.Reviewer) {
	repoCtx := repo.NewActionContext(r.ctx.Env, challenge)

	// Check the candidate
	if repoCtx.CheckUser(candidate.GithubAlias) == false {
		errorMsg := fmt.Sprintf("Github Alias %s for candidate %s is not correct", candidate.GithubAlias, candidate.Name)
		r.ctx.postMessage(r.icb.Channel.ID, toMsgOption(errorMsg))
		return
	}

	// Check the reviewers
	for _, reviewer := range reviewers {
		if repoCtx.CheckUser(reviewer.GithubAlias) == false {
			errorMsg := fmt.Sprintf("Github Alias %s for reviewer %s is not correct", reviewer.GithubAlias, reviewer.Name)
			r.ctx.postMessage(r.icb.Channel.ID, toMsgOption(errorMsg))
			return
		}
	}

	// log.Println("[INFO] Reviewer count = ", len(reviewers))

	// Create the challenge
	r.ctx.postMessage(r.icb.Channel.ID, toMsgOption("Please be patient, while I go create a coding challenge for you..."))
	challengeURL, err := repoCtx.CreateChallenge(candidate, challenge, reviewers)
	if err != nil {
		re := regexp.MustCompile(dreadedPrivateRepoError)
		var errorMsg string
		if re.FindStringIndex(err.Error()) != nil {
			errorMsg = fmt.Sprintf("Unable to create challenge. You need to cleanup private repositories, because you exceeded your allowed limit.")
		} else {
			errorMsg = fmt.Sprintf("Unable to create challenge for %s because of ", candidate.Name, err.Error())
		}
		r.ctx.postMessage(r.icb.Channel.ID, toMsgOption(errorMsg))
		return
	}
	r.ctx.postMessage(r.icb.Channel.ID, renderChallengeSummary(candidate, challengeURL, challenge.TrackingIssuesURL()))
}

func (r request) handleNewChallenge() error {
	challengeInput := r.icb.Submission
	challengeInput["team_id"] = r.icb.Team.ID

	challenge := models.NewChallenge(challengeInput)
	go r.updateChallenge(challenge)
	return nil
}

func (r request) handleEditChallenge() error {
	challengeInput := r.icb.Submission
	challengeInput["team_id"] = r.icb.Team.ID
	challengeID := r.icb.State

	challenge, err := models.EditChallenge(r.ctx.Env, challengeInput, challengeID)
	if err != nil {
		return err
	}
	go r.updateChallenge(challenge)
	return nil
}

func (r request) updateChallenge(challenge models.Challenge) {
	err := models.UpdateChallenge(r.ctx.Env, challenge)
	if err != nil {
		log.Println("[ERROR] Could not update challenge in db ", err)
		r.ctx.postMessage(r.icb.Channel.ID, toMsgOption("We were not able to create the new challenge"))
	}

	challengeSetup, err := models.GetChallengeSetupByName(r.ctx.Env, challenge.Name)
	if err != nil {
		log.Println("[ERROR] Could not create a valid challenge setup, perhaps the github repo name is not valid ", err)
		r.ctx.postMessage(r.icb.Channel.ID, toMsgOption("We were not able to create a valid challenge"))
	}
	msgText := fmt.Sprintf("We created a challenge named %s in our database. It is pointing to: %s", challengeSetup.Name, challengeSetup.TemplateRepositoryURL())
	r.ctx.postMessage(r.icb.Channel.ID, toMsgOption(msgText))
}
