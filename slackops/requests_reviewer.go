package slackops

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/keremk/challenge-bot/config"
	"github.com/keremk/challenge-bot/models"
	"github.com/keremk/challenge-bot/scheduling"

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

func handleShowSchedule(env config.Environment, icb slack.InteractionCallback) error {
	scheduleInput := icb.Submission
	log.Println("[INFO] Reviewer data", scheduleInput)

	week, year := decodeWeekAndYear(scheduleInput["year_week"])
	log.Println("[INFO] Week ", week)

	state, err := stateFromString(icb.State)
	reviewerSlackID := state.argument
	log.Println("[INFO] Reviewer name", reviewerSlackID)

	reviewer, err := models.GetReviewerBySlackID(env, reviewerSlackID)
	if err != nil {
		log.Println("[ERROR] No such reviewer registered.")
		return err
	}

	challenge, err := models.GetChallengeSetup(env, reviewer.ChallengeName)
	if err != nil {
		log.Println("[ERROR] Reviewer did not register to a challenge.")
		return err
	}

	slots := scheduling.SlotsForWeek(week, year, reviewer, challenge)
	log.Println("[INFO] Slots available: ", slots)
	log.Println("[INFO] Reviewer is ", reviewer)

	headerMsgText := fmt.Sprintf("%s schedule for week #: %d", reviewer.Name, week)
	postMessage(env, icb.Team.ID, icb.Channel.ID, toMsgOption(headerMsgText))

	scheduleMsgBlock := renderSchedule(week, year, reviewer, slots)
	postMessage(env, icb.Team.ID, icb.Channel.ID, slack.MsgOptionBlocks(&scheduleMsgBlock))
	return nil
}

type updateMsg struct {
	ReplaceOriginal bool                `json:"replace_original,omitempty"`
	Blocks          []slack.ActionBlock `json:"blocks,omitempty"`
}

func handleUpdateSchedule(env config.Environment, icb slack.InteractionCallback) error {
	scheduleInfo, err := decodeScheduleActionInfo(icb.ActionCallback.BlockActions[0].ActionID)
	if err != nil {
		log.Println("[ERROR] Cannot decode schedule info - ", err)
		return err
	}

	reviewer, err := models.GetReviewerBySlackID(env, scheduleInfo.ReviewerID)
	if err != nil {
		log.Println("[ERROR] Cannot retrieve reviewer - ", err)
		return err
	}
	log.Println("[INFO] Reviewer is - ", reviewer)

	challenge, err := models.GetChallengeSetup(env, reviewer.ChallengeName)
	if err != nil {
		log.Println("[ERROR] Reviewer did not register to a challenge.", err)
		return err
	}
	log.Println("[INFO] Challenge is - ", challenge)

	slotChecked, err := strconv.ParseBool(icb.ActionCallback.BlockActions[0].Value)
	if err != nil {
		log.Println("[ERROR] value not properly encoded ", err)
		return err
	}
	slotChecked = !slotChecked

	reviewer, err = models.UpdateReviewerAvailability(env, reviewer, models.SlotReference{
		SlotID:    scheduleInfo.SlotID,
		WeekNo:    scheduleInfo.WeekNo,
		Year:      scheduleInfo.Year,
		Available: slotChecked,
	})
	if err != nil {
		log.Println("[ERROR] Update availability not successful - ", err)
		return err
	}
	log.Println("[INFO] Updated reviewer is - ", reviewer)

	slots := scheduling.SlotsForWeek(scheduleInfo.WeekNo, scheduleInfo.Year, reviewer, challenge)
	scheduleMsgBlock := renderSchedule(scheduleInfo.WeekNo, scheduleInfo.Year, reviewer, slots)

	updateMsg := updateMsg{
		ReplaceOriginal: true,
		Blocks:          []slack.ActionBlock{scheduleMsgBlock},
	}

	respJSON, err := json.Marshal(updateMsg)
	if err != nil {
		log.Println("[ERROR] Cannot marshal the json response - ", err)
		return err
	}
	log.Println(string(respJSON))

	sendDelayedResponse(icb.ResponseURL, string(respJSON))
	return nil
}
