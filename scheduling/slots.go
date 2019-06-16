package scheduling

import (
	"fmt"
	"log"
	"strings"

	"github.com/keremk/challenge-bot/config"
	"github.com/keremk/challenge-bot/models"
)

type SlotInfo struct {
	Slot       models.Slot
	IsSelected bool
}

type SlotBooking struct {
	SlotID   string
	WeekNo   int
	Year     int
	IsBooked bool
}
type SlotReference struct {
	SlotID    string
	WeekNo    int
	Year      int
	Available bool
}

func SlotIndex(week, year int) string {
	var slotIndex string
	if week == 0 {
		// Return all weeks (general) setup
		slotIndex = "General"
	} else {
		// Return the specific week info
		slotIndex = fmt.Sprintf("%d-%d", week, year)
	}

	return slotIndex
}

func GetAvailableSlots(reviewer models.Reviewer, weekNo, year int) []SlotBooking {
	slotIndex := SlotIndex(weekNo, year)
	availableSlots := reviewer.Availability[slotIndex]
	if availableSlots == nil {
		availableSlots = reviewer.Availability["General"]
	}

	bookedSlots := reviewer.Bookings[slotIndex]

	var slotBookings = make([]SlotBooking, 0, len(availableSlots))
	for _, slotID := range availableSlots {
		isBooked := false
		for _, bookedID := range bookedSlots {
			if bookedID == slotID {
				isBooked = true
				break
			}
		}
		slotBookings = append(slotBookings, SlotBooking{
			SlotID:   slotID,
			WeekNo:   weekNo,
			Year:     year,
			IsBooked: isBooked,
		})
	}

	return slotBookings
}

func SlotsForWeek(week, year int, reviewer models.Reviewer, challenge models.ChallengeSetup) []SlotInfo {
	slotIndex := SlotIndex(week, year)

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

func FindAvailableReviewers(env config.Environment, challengeID string, tech string, weekNo, year int) (map[DayOfWeek]map[string]*SlotAvailability, error) {
	challenge, err := models.GetChallengeSetupByID(env, challengeID)
	if err != nil {
		log.Println("[ERROR] Cannot find the challenge setup - ", challengeID)
		return nil, err
	}

	reviewers, err := models.GetAllReviewersForChallenge(env, challenge.ID)
	if err != nil {
		log.Println("[ERROR] No reviewers for the challenge - ", challenge.Name)
		return nil, err
	}

	selectedReviewers := make(map[DayOfWeek]map[string]*SlotAvailability)
	for _, reviewer := range reviewers {
		if tech != "" && !strings.Contains(reviewer.TechnologyList, tech) {
			continue
		}

		slotBookings := GetAvailableSlots(reviewer, weekNo, year)

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

func UpdateReviewerAvailability(env config.Environment, reviewer models.Reviewer, ref SlotReference) (models.Reviewer, error) {
	slotIndex := SlotIndex(ref.WeekNo, ref.Year)
	slots := initializeSlots(slotIndex, reviewer)

	var newSlots []string
	if ref.Available {
		newSlots = addSlot(slots, ref.SlotID)
	} else {
		newSlots = removeSlot(slots, ref.SlotID)
	}

	reviewer.Availability[slotIndex] = newSlots
	err := models.UpdateReviewer(env, reviewer)

	return reviewer, err
}

func initializeSlots(slotIndex string, reviewer models.Reviewer) []string {
	var slots []string
	if slotIndex == "General" {
		// Get General or if not create empty
		slots = reviewer.Availability[slotIndex]
		if slots == nil {
			slots = make([]string, 0, 50)
		}
	} else {
		// Get the slot if not copy General or if general does not exist init with zero
		slots = reviewer.Availability[slotIndex]
		if slots == nil {
			slots = reviewer.Availability["General"]
			if slots == nil {
				slots = make([]string, 0, 50)
			}
		}
	}

	return slots
}

func UpdateReviewerBooking(env config.Environment, reviewer models.Reviewer, ref SlotBooking) (models.Reviewer, error) {
	slotIndex := SlotIndex(ref.WeekNo, ref.Year)
	slots := reviewer.Bookings[slotIndex]
	if slots == nil {
		slots = make([]string, 0, 20)
	}
	maxBookings := reviewer.BookingsPerWeek

	var newSlots []string
	if ref.IsBooked {
		if len(slots) >= maxBookings {
			return reviewer, MaxBookingsError{}
		}
		newSlots = addSlot(slots, ref.SlotID)
	} else {
		newSlots = removeSlot(slots, ref.SlotID)
	}

	reviewer.Bookings[slotIndex] = newSlots
	err := models.UpdateReviewer(env, reviewer)

	return reviewer, err
}

func addSlot(slots []string, newSlot string) []string {
	for _, slot := range slots {
		if slot == newSlot {
			return slots
		}
	}
	return append(slots, newSlot)
}

func removeSlot(slots []string, oldSlot string) []string {
	newSlots := make([]string, 0, len(slots))
	for _, slot := range slots {
		if slot != oldSlot {
			newSlots = append(newSlots, slot)
		}
	}
	return newSlots
}
