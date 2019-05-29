package slackops

import (
	"fmt"
	"log"

	"github.com/keremk/challenge-bot/config"
	"github.com/keremk/challenge-bot/models"

	"github.com/nlopes/slack"
)

func handleNewReviewer(env config.Environment, icb slack.InteractionCallback) error {
	addReviewerInput := icb.Submission
	// log.Println("[INFO] Reviewer data", addReviewerInput)

	user, err := getUserInfo(env, addReviewerInput["reviewer_id"], icb.Team.ID)
	if err != nil {
		return err
	}

	reviewer := models.NewReviewer(user.Name, addReviewerInput)
	// log.Println("[INFO] Reviewer is ", reviewer)

	err = models.UpdateReviewer(env, reviewer)
	if err != nil {
		log.Println("[ERROR] Could not update reviewer in db ", err)
		_ = postMessage(env, icb.Team.ID, icb.Channel.ID, toMsgOption("We were not able to create the new reviewer"))
		return err
	}

	msgText := fmt.Sprintf("We created a reviewer named %s in our database. They will be reviewing: %s, and their Github alias is: %s", reviewer.Name, reviewer.ChallengeName, reviewer.GithubAlias)
	_ = postMessage(env, icb.Team.ID, icb.Channel.ID, toMsgOption(msgText))
	return nil
}
