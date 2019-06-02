package models

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/keremk/challenge-bot/config"
	"github.com/keremk/challenge-bot/db"
	"github.com/keremk/challenge-bot/util"
)

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
}

func NewReviewer(name string, input map[string]string) Reviewer {
	return Reviewer{
		ID:             util.RandomString(16),
		Name:           name,
		GithubAlias:    input["github_alias"],
		SlackID:        input["reviewer_id"],
		TechnologyList: input["technology_list"],
		ChallengeName:  input["challenge_name"],
	}
}

func GetReviewer(env config.Environment, id string) (Reviewer, error) {
	reviewer := Reviewer{}
	store, err := db.NewStore(env, db.ReviewersCollection)
	if err != nil {
		return reviewer, err
	}

	err = store.FindByID(id, &reviewer)
	return reviewer, err
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

func UpdateReviewer(env config.Environment, reviewer Reviewer) error {
	store, err := db.NewStore(env, db.ReviewersCollection)
	if err != nil {
		return err
	}
	return store.Update(reviewer.ID, reviewer)
}

func UpdateReviewerAvailability(env config.Environment, reviewer Reviewer, ref SlotReference) (Reviewer, error) {
	slotIndex := SlotIndex(ref.WeekNo, ref.Year)
	slots := reviewer.Availability[slotIndex]
	if slots == nil {
		slots = reviewer.Availability["General"]
	}
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
