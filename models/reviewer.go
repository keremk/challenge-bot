package models

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strconv"

	"github.com/keremk/challenge-bot/config"
	"github.com/keremk/challenge-bot/db"
	"github.com/keremk/challenge-bot/util"
)

type Reviewer struct {
	ID              string              `bson:"ID"`
	Name            string              `bson:"Name"`
	GithubAlias     string              `bson:"GithubAlias"`
	SlackID         string              `bson:"SlackID"`
	TechnologyList  string              `bson:"TechnologyList"`
	ChallengeName   string              `bson:"ChallengeName"`
	ChallengeID     string              `bson:"ChallengeID"`
	Experience      int                 `bson:"Experience"`
	BookingsPerWeek int                 `bson:"BookingsPerWeek"`
	Availability    map[string][]string `bson:"Availability"`
	Bookings        map[string][]string `bson:"Bookings"`
}

func NewReviewer(name string, input map[string]string) Reviewer {
	id := fmt.Sprintf("%s-%s", name, util.RandomString(8))

	reviewer := Reviewer{
		ID:           id,
		SlackID:      input["reviewer_id"],
		Name:         name,
		Availability: make(map[string][]string),
		Bookings:     make(map[string][]string),
	}
	return reviewerFromInput(reviewer, input)
}

func reviewerFromInput(reviewer Reviewer, input map[string]string) Reviewer {
	reviewer.GithubAlias = input["github_alias"]
	reviewer.TechnologyList = input["technology_list"]
	reviewer.ChallengeID = input["challenge_id"]

	experience, err := strconv.Atoi(input["experience"])
	if err != nil {
		log.Println("[ERROR] Experience level not properly encoded, assuming lowest", err)
		experience = 0
	}
	reviewer.Experience = experience

	bookingsPerWeek, err := strconv.Atoi(input["bookings_week"])
	if err != nil {
		log.Println("[ERROR] Bookings per week not properly encoded, assuming 1", err)
		bookingsPerWeek = 1
	}
	reviewer.BookingsPerWeek = bookingsPerWeek
	return reviewer
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

func GetAllReviewersForChallenge(env config.Environment, challengeID string) ([]Reviewer, error) {
	store, err := db.NewStore(env, db.ReviewersCollection)
	if err != nil {
		return nil, err
	}

	var all []Reviewer
	result, err := store.FindAllWithKeyValue(reflect.TypeOf(all), "ChallengeID", challengeID)
	all, ok := result.([]Reviewer)
	if !ok {
		return nil, errors.New("[ERROR] Cannot convert")
	}
	return all, err
}

func EditReviewer(env config.Environment, slackID string, input map[string]string) (Reviewer, error) {
	reviewer, err := GetReviewerBySlackID(env, slackID)
	if err != nil {
		return reviewer, err
	}

	reviewer = reviewerFromInput(reviewer, input)
	err = UpdateReviewer(env, reviewer)
	return reviewer, err
}

func UpdateReviewer(env config.Environment, reviewer Reviewer) error {
	store, err := db.NewStore(env, db.ReviewersCollection)
	if err != nil {
		return err
	}
	return store.Update(reviewer.ID, reviewer)
}
