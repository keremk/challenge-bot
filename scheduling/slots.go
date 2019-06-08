package scheduling

import (
	"log"
	"strings"

	"github.com/keremk/challenge-bot/config"
	"github.com/keremk/challenge-bot/models"
)

type SlotInfo struct {
	Slot       models.Slot
	IsSelected bool
}

func SlotsForWeek(week, year int, reviewer models.Reviewer, challenge models.ChallengeSetup) []SlotInfo {
	slotIndex := models.SlotIndex(week, year)

	slots := make([]SlotInfo, 0, len(challenge.Slots))
	availableSlots := reviewer.Availability[slotIndex]
	challengeSlots := challenge.GetSlotsInOrder()

	if availableSlots == nil || len(availableSlots) == 0 {
		for _, slot := range challengeSlots {
			slots = append(slots, SlotInfo{
				Slot:       *slot,
				IsSelected: isSelected(*slot, reviewer.Availability["General"]),
			})
		}
	} else {
		for _, slot := range challengeSlots {
			slots = append(slots, SlotInfo{
				Slot:       *slot,
				IsSelected: isSelected(*slot, availableSlots),
			})
		}
	}

	return slots
}

func isSelected(slot models.Slot, availableSlots []string) bool {
	for _, slotID := range availableSlots {
		if slotID == slot.ID {
			return true
		}
	}
	return false
}

type DayOfWeek = string

type ReviewerInfo struct {
	Reviewer models.Reviewer
	IsBooked bool
}
type SlotAvailability struct {
	Slot      *models.Slot
	Reviewers []ReviewerInfo
}

func newSlotAvailability(slot *models.Slot) *SlotAvailability {
	return &SlotAvailability{
		Slot:      slot,
		Reviewers: make([]ReviewerInfo, 0, 100),
	}
}

func FindAvailableReviewers(env config.Environment, challengeName string, tech string, weekNo, year int) (map[DayOfWeek]map[string]*SlotAvailability, error) {
	challenge, err := models.GetChallengeSetup(env, challengeName)
	if err != nil {
		log.Println("[ERROR] Cannot find the challenge setup - ", challengeName)
		return nil, err
	}

	reviewers, err := models.GetAllReviewersForChallenge(env, challengeName)
	if err != nil {
		log.Println("[ERROR] No reviewers for the challenge - ", challengeName)
		return nil, err
	}

	selectedReviewers := make(map[DayOfWeek]map[string]*SlotAvailability)
	for _, reviewer := range reviewers {
		if tech != "" && !strings.Contains(reviewer.TechnologyList, tech) {
			continue
		}

		slotBookings := models.GetAvailableSlots(reviewer, weekNo, year)

		for _, slotBooking := range slotBookings {
			slotID := slotBooking.SlotID
			isBooked := slotBooking.IsBooked
			slot := challenge.Slots[slotID]
			if slot == nil {
				log.Println("[ERROR] Invalid slot name for reviewer - ", slotID)
				continue
			}
			if selectedReviewers[slot.Day] == nil {
				selectedReviewers[slot.Day] = make(map[string]*SlotAvailability)
			}
			if selectedReviewers[slot.Day][slotID] == nil {
				selectedReviewers[slot.Day][slotID] = newSlotAvailability(slot)
			}

			reviewerInfo := ReviewerInfo{
				Reviewer: reviewer,
				IsBooked: isBooked,
			}
			selectedReviewers[slot.Day][slotID].Reviewers = append(selectedReviewers[slot.Day][slotID].Reviewers, reviewerInfo)
		}
	}

	return selectedReviewers, nil
}
