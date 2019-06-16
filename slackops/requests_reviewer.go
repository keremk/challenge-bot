package slackops

import (
	"fmt"
	"log"
	"strconv"

	"github.com/keremk/challenge-bot/models"
	"github.com/keremk/challenge-bot/scheduling"

	"github.com/nlopes/slack"
)

func (r request) handleNewReviewer() error {
	addReviewerInput := r.icb.Submission
	// log.Println("[INFO] Reviewer data", addReviewerInput)
	reviewerSlackID := addReviewerInput["reviewer_id"]

	go r.createNewReviewer(reviewerSlackID, addReviewerInput)
	return nil
}

func (r request) createNewReviewer(reviewerSlackID string, input map[string]string)  {
	user, err := r.ctx.getUserInfo(reviewerSlackID)
	if err != nil {
		log.Println("[ERROR] Could not update reviewer in db ", err)
		r.ctx.postMessage(r.icb.Channel.ID, toMsgOption("Cannot find the reviewer in Slack, please make sure reviewer is a Slack member"))
	}

	reviewer := models.NewReviewer(user.Name, input)
	// log.Println("[INFO] Reviewer is ", reviewer)

	err = models.UpdateReviewer(r.ctx.Env, reviewer)
	if err != nil {
		log.Println("[ERROR] Could not update reviewer in db ", err)
		r.ctx.postMessage(r.icb.Channel.ID, toMsgOption("We were not able to create the new reviewer"))
	}

	msgText := fmt.Sprintf("We created a reviewer <@%s> in our database. Their Github alias is: %s", reviewer.SlackID, reviewer.GithubAlias)
	r.ctx.postMessage(r.icb.Channel.ID, toMsgOption(msgText))
}

func (r request) handleEditReviewer() error {
	reviewer, err := models.EditReviewer(r.ctx.Env, r.icb.State, r.icb.Submission)
	// log.Println("[INFO] Reviewer is ", reviewer)

	if err != nil {
		log.Println("[ERROR] Could not update reviewer in db ", err)
		r.ctx.postMessage(r.icb.Channel.ID, toMsgOption("We were not able to create the new reviewer"))
		return err
	}

	msgText := fmt.Sprintf("We edited the reviewer <@%s> in our database.", reviewer.SlackID)
	r.ctx.postMessage(r.icb.Channel.ID, toMsgOption(msgText))
	return nil
}

func (r request) handleShowSchedule() error {
	scheduleInput := r.icb.Submission
	log.Println("[INFO] Reviewer data", scheduleInput)

	week, year := decodeWeekAndYear(scheduleInput["year_week"])
	log.Println("[INFO] Week ", week)

	reviewerSlackID := r.icb.State
	log.Println("[INFO] Reviewer ID", reviewerSlackID)

	go r.showSchedule(week, year, reviewerSlackID)

	return nil
}

func (r request) showSchedule(week, year int, reviewerSlackID string) {
	reviewer, err := models.GetReviewerBySlackID(r.ctx.Env, reviewerSlackID)
	// log.Println("INFO: Reviewer - ", reviewer)
	// log.Println("INFO: Error - ", err)
	if err != nil {
		log.Println("[ERROR] No such reviewer registered.", err)
		errorMsg := fmt.Sprintf("Reviewer <@%s> is not registered. Please register first using /reviewer new command.", reviewerSlackID)
		r.ctx.postMessage(r.icb.Channel.ID, toMsgOption(errorMsg))
		return
	}

	challenge, err := models.GetChallengeSetupByName(r.ctx.Env, reviewer.ChallengeName)
	if err != nil {
		log.Println("[ERROR] Reviewer did not register to a challenge.", err)
		errorMsg := fmt.Sprintf("Reviewer <%s> did not register for a specific challenge.", reviewer.Name)
		r.ctx.postMessage(r.icb.Channel.ID, toMsgOption(errorMsg))
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
	err = r.ctx.postMessage(r.icb.Channel.ID, toMsgOption(headerMsgText))
	if err != nil {
		log.Println("[ERROR] Cannot send the reviewer schedule header - ", err)
		return
	}

	scheduleMsgBlock := renderSchedule(week, year, reviewer, slots)
	err = r.ctx.postMessage(r.icb.Channel.ID, slack.MsgOptionBlocks(&scheduleMsgBlock))
	if err != nil {
		log.Println("[ERROR] Cannot send the reviewer schedule details - ", err)
	}
}

func (r request) handleUpdateSchedule(encodedActionInfo string) error {
	scheduleInfo, err := decodeScheduleActionInfo(encodedActionInfo)
	if err != nil {
		log.Println("[ERROR] Cannot decode schedule info - ", err)
		return err
	}

	slotChecked, err := strconv.ParseBool(r.icb.ActionCallback.BlockActions[0].Value)
	if err != nil {
		log.Println("[ERROR] value not properly encoded ", err)
		return err
	}

	r.updateSchedule(slotChecked, scheduleInfo)
	return nil
}

func (r request) updateSchedule(slotChecked bool, scheduleInfo scheduleActionInfo) {
	reviewer, err := models.GetReviewerBySlackID(r.ctx.Env, scheduleInfo.ReviewerID)
	if err != nil {
		log.Println("[ERROR] No such reviewer registered.", err)
		errorMsg := fmt.Sprintf("Reviewer <%s> is not registered.", scheduleInfo.ReviewerID)
		r.ctx.postMessage(r.icb.Channel.ID, toMsgOption(errorMsg))
	}
	// log.Println("[INFO] Reviewer is - ", reviewer)

	challenge, err := models.GetChallengeSetupByName(r.ctx.Env, reviewer.ChallengeName)
	if err != nil {
		log.Println("[ERROR] Reviewer did not register to a challenge.", err)
		errorMsg := fmt.Sprintf("Reviewer <%s> did not register for a specific challenge.", reviewer.Name)
		r.ctx.postMessage(r.icb.Channel.ID, toMsgOption(errorMsg))
	}
	// log.Println("[INFO] Challenge is - ", challenge)

	slotChecked = !slotChecked

	reviewer, err = scheduling.UpdateReviewerAvailability(r.ctx.Env, reviewer, scheduling.SlotReference{
		SlotID:    scheduleInfo.SlotID,
		WeekNo:    scheduleInfo.WeekNo,
		Year:      scheduleInfo.Year,
		Available: slotChecked,
	})
	if err != nil {
		log.Println("[ERROR] Update availability not successful - ", err)
		errorMsg := fmt.Sprintf("There was an error. Availability cannot be updated.")
		r.ctx.postMessage(r.icb.Channel.ID, toMsgOption(errorMsg))
	}
	// log.Println("[INFO] Updated reviewer is - ", reviewer)

	slots := scheduling.SlotsForWeek(scheduleInfo.WeekNo, scheduleInfo.Year, reviewer, challenge)
	scheduleMsgBlock := renderSchedule(scheduleInfo.WeekNo, scheduleInfo.Year, reviewer, slots)

	msg := slack.MsgOptionBlocks(&scheduleMsgBlock)

	r.ctx.updateMessage(r.icb.Channel.ID, r.icb.Message.Timestamp, msg)

	// respJSON, err := json.Marshal(scheduleMsgBlock)
	// if err != nil {
	// 	log.Println("[ERROR] Cannot marshal the json response - ", err)
	// }
	// log.Println(string(respJSON))
}

func (r request) handleFindReviewers() error {
	scheduleInput := r.icb.Submission
	// log.Println("[INFO] Reviewer data", scheduleInput)

	week, year := decodeWeekAndYear(scheduleInput["year_week"])
	day := scheduleInput["day"]
	// log.Println("[INFO] Day ", day)
	// log.Println("[INFO] Week ", week)

	challengeName := scheduleInput["challenge_name"]
	technology := scheduleInput["technology"]

	go r.findAvailableReviewers(challengeName, technology, day, week, year)

	return nil
}

func (r request) findAvailableReviewers(challengeName, technology, day string, week, year int) {
	availableReviewers, err := scheduling.FindAvailableReviewers(r.ctx.Env, challengeName, technology, week, year)
	if err != nil {
		log.Println("[ERROR] Found no results", err)
	}

	scheduleInfo := availableReviewers[day]
	if scheduleInfo == nil {
		errorMsg := fmt.Sprintf("No reviewers available for %s on the week of %d, %d", day, week, year)
		r.ctx.postMessage(r.icb.Channel.ID, toMsgOption(errorMsg))
	}
	scheduleMsg := renderReviewers(week, year, scheduleInfo)

	r.ctx.postMessage(r.icb.Channel.ID, scheduleMsg)
}

func (r request) handleBookings(encodedActionInfo string) error {
	scheduleInfo, err := decodeScheduleActionInfo(encodedActionInfo)
	if err != nil {
		log.Println("[ERROR] Cannot decode schedule info - ", err)
		return err
	}

	isBooked, err := strconv.ParseBool(r.icb.ActionCallback.BlockActions[0].Value)
	if err != nil {
		log.Println("[ERROR] value not properly encoded ", err)
		return err
	}

	r.updateBooking(isBooked, scheduleInfo)
	return nil
}

func (r request) updateBooking(isBooked bool, scheduleInfo scheduleActionInfo) {
	reviewer, err := models.GetReviewerBySlackID(r.ctx.Env, scheduleInfo.ReviewerID)
	if err != nil {
		log.Println("[ERROR] No such reviewer registered.", err)
		errorMsg := fmt.Sprintf("Reviewer <%s> is not registered.", scheduleInfo.ReviewerID)
		r.ctx.postMessage(r.icb.Channel.ID, toMsgOption(errorMsg))
	}
	// log.Println("[INFO] Reviewer is - ", reviewer)
	isBooked = !isBooked // Toggle booking

	reviewer, err = scheduling.UpdateReviewerBooking(r.ctx.Env, reviewer, scheduling.SlotBooking{
		SlotID:   scheduleInfo.SlotID,
		WeekNo:   scheduleInfo.WeekNo,
		Year:     scheduleInfo.Year,
		IsBooked: isBooked,
	})
	if err != nil {
		switch err.(type) {
		case scheduling.MaxBookingsError:
			errorMsg := fmt.Sprintf("Reviewer can only be booked a maximum of %d times/week. Please unbook another appointment in that week.", reviewer.BookingsPerWeek)
			r.ctx.postMessage(r.icb.Channel.ID, toMsgOption(errorMsg))
			return
		default:
			log.Println("[ERROR] Update booking not successful - ", err)
			errorMsg := fmt.Sprintf("There was an error. Booking cannot be updated.")
			r.ctx.postMessage(r.icb.Channel.ID, toMsgOption(errorMsg))
			return
		}
	}

	var msg string
	if isBooked {
		msg = fmt.Sprintf("<@%s|%s> is now booked for the slot %s on week %d", reviewer.SlackID, reviewer.Name, scheduleInfo.SlotID, scheduleInfo.WeekNo)
	} else {
		msg = fmt.Sprintf("<@%s|%s> is now free for the slot %s on week %d", reviewer.SlackID, reviewer.Name, scheduleInfo.SlotID, scheduleInfo.WeekNo)
	}
	r.ctx.postMessage(r.icb.Channel.ID, toMsgOption(msg))
}
