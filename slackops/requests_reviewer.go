package slackops

import (
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
	reviewerSlackID := addReviewerInput["reviewer_id"]

	user, err := getUserInfo(env, reviewerSlackID, icb.Team.ID)
	if err != nil {
		return err
	}

	reviewer := models.NewReviewer(user.Name, addReviewerInput)
	// log.Println("[INFO] Reviewer is ", reviewer)

	err = models.UpdateReviewer(env, reviewer)
	if err != nil {
		log.Println("[ERROR] Could not update reviewer in db ", err)
		postMessage(env, icb.Team.ID, icb.Channel.ID, toMsgOption("We were not able to create the new reviewer"))
		return err
	}

	msgText := fmt.Sprintf("We created a reviewer named %s in our database. They will be reviewing: %s, and their Github alias is: %s", reviewer.Name, reviewer.ChallengeName, reviewer.GithubAlias)
	postMessage(env, icb.Team.ID, icb.Channel.ID, toMsgOption(msgText))
	return nil
}

func handleEditReviewer(env config.Environment, icb slack.InteractionCallback) error {
	reviewer, err := models.EditReviewer(env, icb.State, icb.Submission)
	// log.Println("[INFO] Reviewer is ", reviewer)

	if err != nil {
		log.Println("[ERROR] Could not update reviewer in db ", err)
		postMessage(env, icb.Team.ID, icb.Channel.ID, toMsgOption("We were not able to create the new reviewer"))
		return err
	}

	msgText := fmt.Sprintf("We edited the reviewer named %s in our database.", reviewer.Name)
	postMessage(env, icb.Team.ID, icb.Channel.ID, toMsgOption(msgText))
	return nil
}

func handleShowSchedule(env config.Environment, icb slack.InteractionCallback) error {
	scheduleInput := icb.Submission
	log.Println("[INFO] Reviewer data", scheduleInput)

	week, year := decodeWeekAndYear(scheduleInput["year_week"])
	log.Println("[INFO] Week ", week)

	reviewerSlackID := icb.State
	log.Println("[INFO] Reviewer ID", reviewerSlackID)

	go showSchedule(env, week, year, reviewerSlackID, icb.Team.ID, icb.Channel.ID)

	return nil
}

func showSchedule(env config.Environment, week, year int, reviewerSlackID, teamID, channelID string) {
	reviewer, err := models.GetReviewerBySlackID(env, reviewerSlackID)
	// log.Println("INFO: Reviewer - ", reviewer)
	// log.Println("INFO: Error - ", err)
	if err != nil {
		log.Println("[ERROR] No such reviewer registered.", err)
		errorMsg := fmt.Sprintf("Reviewer <@%s> is not registered. Please register first using /reviewer new command.", reviewerSlackID)
		postMessage(env, teamID, channelID, toMsgOption(errorMsg))
		return
	}

	challenge, err := models.GetChallengeSetup(env, reviewer.ChallengeName)
	if err != nil {
		log.Println("[ERROR] Reviewer did not register to a challenge.", err)
		errorMsg := fmt.Sprintf("Reviewer <%s> did not register for a specific challenge.", reviewer.Name)
		postMessage(env, teamID, channelID, toMsgOption(errorMsg))
		return
	}

	slots := scheduling.SlotsForWeek(week, year, reviewer, challenge)
	// log.Println("[INFO] Slots available: ", slots)
	// log.Println("[INFO] Reviewer is ", reviewer)

	weekDescription, err := scheduling.WeekDescriptionFromWeekNo(week, year)
	if err != nil {
		log.Println("[ERROR] Week number not valid", err)
		return
	}

	headerMsgText := fmt.Sprintf("<@%s>'s schedule in %s", reviewer.SlackID, weekDescription)
	err = postMessage(env, teamID, channelID, toMsgOption(headerMsgText))
	if err != nil {
		log.Println("[ERROR] Cannot send the reviewer schedule header - ", err)
		return
	}

	scheduleMsgBlock := renderSchedule(week, year, reviewer, slots)
	err = postMessage(env, teamID, channelID, slack.MsgOptionBlocks(&scheduleMsgBlock))
	if err != nil {
		log.Println("[ERROR] Cannot send the reviewer schedule details - ", err)
	}
}

type updateMsg struct {
	ReplaceOriginal bool                `json:"replace_original,omitempty"`
	Blocks          []slack.ActionBlock `json:"blocks,omitempty"`
}

func handleUpdateSchedule(env config.Environment, icb slack.InteractionCallback, encodedActionInfo string) error {
	scheduleInfo, err := decodeScheduleActionInfo(encodedActionInfo)
	if err != nil {
		log.Println("[ERROR] Cannot decode schedule info - ", err)
		return err
	}

	slotChecked, err := strconv.ParseBool(icb.ActionCallback.BlockActions[0].Value)
	if err != nil {
		log.Println("[ERROR] value not properly encoded ", err)
		return err
	}

	updateSchedule(env, icb.Team.ID, icb.Channel.ID, icb.Message.Timestamp, icb.ResponseURL, slotChecked, scheduleInfo)
	return nil
}

func updateSchedule(env config.Environment, teamID, channelID, messageTs, responseURL string, slotChecked bool, scheduleInfo scheduleActionInfo) {
	reviewer, err := models.GetReviewerBySlackID(env, scheduleInfo.ReviewerID)
	if err != nil {
		log.Println("[ERROR] No such reviewer registered.", err)
		errorMsg := fmt.Sprintf("Reviewer <%s> is not registered.", scheduleInfo.ReviewerID)
		postMessage(env, teamID, channelID, toMsgOption(errorMsg))
	}
	// log.Println("[INFO] Reviewer is - ", reviewer)

	challenge, err := models.GetChallengeSetup(env, reviewer.ChallengeName)
	if err != nil {
		log.Println("[ERROR] Reviewer did not register to a challenge.", err)
		errorMsg := fmt.Sprintf("Reviewer <%s> did not register for a specific challenge.", reviewer.Name)
		postMessage(env, teamID, channelID, toMsgOption(errorMsg))
	}
	// log.Println("[INFO] Challenge is - ", challenge)

	slotChecked = !slotChecked

	reviewer, err = scheduling.UpdateReviewerAvailability(env, reviewer, scheduling.SlotReference{
		SlotID:    scheduleInfo.SlotID,
		WeekNo:    scheduleInfo.WeekNo,
		Year:      scheduleInfo.Year,
		Available: slotChecked,
	})
	if err != nil {
		log.Println("[ERROR] Update availability not successful - ", err)
		errorMsg := fmt.Sprintf("There was an error. Availability cannot be updated.")
		postMessage(env, teamID, channelID, toMsgOption(errorMsg))
	}
	// log.Println("[INFO] Updated reviewer is - ", reviewer)

	slots := scheduling.SlotsForWeek(scheduleInfo.WeekNo, scheduleInfo.Year, reviewer, challenge)
	scheduleMsgBlock := renderSchedule(scheduleInfo.WeekNo, scheduleInfo.Year, reviewer, slots)

	msg := slack.MsgOptionBlocks(&scheduleMsgBlock)

	updateMessage(env, teamID, channelID, messageTs, msg)

	// respJSON, err := json.Marshal(scheduleMsgBlock)
	// if err != nil {
	// 	log.Println("[ERROR] Cannot marshal the json response - ", err)
	// }
	// log.Println(string(respJSON))
}

func handleFindReviewers(env config.Environment, icb slack.InteractionCallback) error {
	scheduleInput := icb.Submission
	log.Println("[INFO] Reviewer data", scheduleInput)

	week, year := decodeWeekAndYear(scheduleInput["year_week"])
	day := scheduleInput["day"]
	log.Println("[INFO] Day ", day)
	log.Println("[INFO] Week ", week)

	challengeName := scheduleInput["challenge_name"]
	technology := scheduleInput["technology"]

	go findAvailableReviewers(env, challengeName, technology, day, week, year, icb.Team.ID, icb.Channel.ID)

	return nil
}

func findAvailableReviewers(env config.Environment, challengeName, technology, day string, week, year int, teamID, channelID string) {
	availableReviewers, err := scheduling.FindAvailableReviewers(env, challengeName, technology, week, year)
	if err != nil {
		log.Println("[ERROR] Found no results", err)
	}

	scheduleInfo := availableReviewers[day]
	if scheduleInfo == nil {
		errorMsg := fmt.Sprintf("No reviewers available for %s on the week of %d, %d", day, week, year)
		postMessage(env, teamID, channelID, toMsgOption(errorMsg))
	}
	scheduleMsg := renderReviewers(week, year, scheduleInfo)

	postMessage(env, teamID, channelID, scheduleMsg)
}

func handleBookings(env config.Environment, icb slack.InteractionCallback, encodedActionInfo string) error {
	scheduleInfo, err := decodeScheduleActionInfo(encodedActionInfo)
	if err != nil {
		log.Println("[ERROR] Cannot decode schedule info - ", err)
		return err
	}

	isBooked, err := strconv.ParseBool(icb.ActionCallback.BlockActions[0].Value)
	if err != nil {
		log.Println("[ERROR] value not properly encoded ", err)
		return err
	}

	updateBooking(env, icb.Team.ID, icb.Channel.ID, icb.ResponseURL, isBooked, scheduleInfo)
	return nil
}

func updateBooking(env config.Environment, teamID, channelID, responseURL string, isBooked bool, scheduleInfo scheduleActionInfo) {
	reviewer, err := models.GetReviewerBySlackID(env, scheduleInfo.ReviewerID)
	if err != nil {
		log.Println("[ERROR] No such reviewer registered.", err)
		errorMsg := fmt.Sprintf("Reviewer <%s> is not registered.", scheduleInfo.ReviewerID)
		postMessage(env, teamID, channelID, toMsgOption(errorMsg))
	}
	// log.Println("[INFO] Reviewer is - ", reviewer)
	isBooked = !isBooked // Toggle booking

	reviewer, err = scheduling.UpdateReviewerBooking(env, reviewer, scheduling.SlotBooking{
		SlotID:   scheduleInfo.SlotID,
		WeekNo:   scheduleInfo.WeekNo,
		Year:     scheduleInfo.Year,
		IsBooked: isBooked,
	})
	if err != nil {
		switch err.(type) {
		case scheduling.MaxBookingsError:
			errorMsg := fmt.Sprintf("Reviewer can only be booked a maximum of %d times/week. Please unbook another appointment in that week.", reviewer.BookingsPerWeek)
			postMessage(env, teamID, channelID, toMsgOption(errorMsg))
			return
		default:
			log.Println("[ERROR] Update booking not successful - ", err)
			errorMsg := fmt.Sprintf("There was an error. Booking cannot be updated.")
			postMessage(env, teamID, channelID, toMsgOption(errorMsg))
			return
		}
	}

	var msg string
	if isBooked {
		msg = fmt.Sprintf("<@%s|%s> is now booked for the slot %s on week %d", reviewer.SlackID, reviewer.Name, scheduleInfo.SlotID, scheduleInfo.WeekNo)
	} else {
		msg = fmt.Sprintf("<@%s|%s> is now free for the slot %s on week %d", reviewer.SlackID, reviewer.Name, scheduleInfo.SlotID, scheduleInfo.WeekNo)
	}
	postMessage(env, teamID, channelID, toMsgOption(msg))
}
