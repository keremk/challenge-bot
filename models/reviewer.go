package models

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/keremk/challenge-bot/config"
	"github.com/keremk/challenge-bot/db"
	"github.com/keremk/challenge-bot/util"
)

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

type Reviewer struct {
	ID             string
	Name           string
	GithubAlias    string
	SlackID        string
	TechnologyList string
	ChallengeName  string
	Availability   map[string][]string
	Bookings       map[string][]string
}

func NewReviewer(name string, input map[string]string) Reviewer {
	id := fmt.Sprintf("%s-%s", name, util.RandomString(8))
	return Reviewer{
		ID:             id,
		Name:           name,
		GithubAlias:    input["github_alias"],
		SlackID:        input["reviewer_id"],
		TechnologyList: input["technology_list"],
		ChallengeName:  input["challenge_name"],
		Availability:   make(map[string][]string),
		Bookings:       make(map[string][]string),
	}
}

func GetReviewerBySlackID(env config.Environment, slackID string) (Reviewer, error) {
	reviewer := Reviewer{}
	store, err := db.NewStore(env, db.ReviewersCollection)
	if err != nil {
		return reviewer, err
	}

	err = store.FindFirst("SlackID", slackID, &reviewer)
	return reviewer, err
}

func GetAllReviewers(env config.Environment) ([]Reviewer, error) {
	store, err := db.NewStore(env, db.ReviewersCollection)
	if err != nil {
		return nil, err
	}

	var all []Reviewer
	result, err := store.FindAll(reflect.TypeOf(all))
	all, ok := result.([]Reviewer)
	if !ok {
		return nil, errors.New("[ERROR] Cannot convert")
	}
	return all, err
}

func GetAllReviewersForChallenge(env config.Environment, challengeName string) ([]Reviewer, error) {
	store, err := db.NewStore(env, db.ReviewersCollection)
	if err != nil {
		return nil, err
	}

	var all []Reviewer
	result, err := store.FindAllWithKeyValue(reflect.TypeOf(all), "ChallengeName", challengeName)
	all, ok := result.([]Reviewer)
	if !ok {
		return nil, errors.New("[ERROR] Cannot convert")
	}
	return all, err
}

func GetAvailableSlots(reviewer Reviewer, weekNo, year int) []SlotBooking {
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

func UpdateReviewer(env config.Environment, reviewer Reviewer) error {
	store, err := db.NewStore(env, db.ReviewersCollection)
	if err != nil {
		return err
	}
	return store.Update(reviewer.ID, reviewer)
}

func UpdateReviewerAvailability(env config.Environment, reviewer Reviewer, ref SlotReference) (Reviewer, error) {
	slotIndex := SlotIndex(ref.WeekNo, ref.Year)
	slots := initializeSlots(slotIndex, reviewer)

	var newSlots []string
	if ref.Available {
		newSlots = addSlot(slots, ref.SlotID)
	} else {
		newSlots = removeSlot(slots, ref.SlotID)
	}

	reviewer.Availability[slotIndex] = newSlots
	err := UpdateReviewer(env, reviewer)

	return reviewer, err
}

func initializeSlots(slotIndex string, reviewer Reviewer) []string {
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

func UpdateReviewerBooking(env config.Environment, reviewer Reviewer, ref SlotBooking) (Reviewer, error) {
	slotIndex := SlotIndex(ref.WeekNo, ref.Year)
	slots := reviewer.Bookings[slotIndex]
	if slots == nil {
		slots = make([]string, 0, 20)
	}
	var newSlots []string
	if ref.IsBooked {
		newSlots = addSlot(slots, ref.SlotID)
	} else {
		newSlots = removeSlot(slots, ref.SlotID)
	}

	reviewer.Bookings[slotIndex] = newSlots
	err := UpdateReviewer(env, reviewer)

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
