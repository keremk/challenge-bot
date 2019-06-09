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
	ID             string
	Name           string
	GithubAlias    string
	SlackID        string
	TechnologyList string
	ChallengeName  string
	Experience     int
	Availability   map[string][]string
	Bookings       map[string][]string
}

func NewReviewer(name string, input map[string]string) Reviewer {
	id := fmt.Sprintf("%s-%s", name, util.RandomString(8))
	experience, err := strconv.Atoi(input["experience"])
	if err != nil {
		log.Println("[ERROR] Experience level not properly encoded, assuming lowest", err)
		experience = 0
	}
	return Reviewer{
		ID:             id,
		Name:           name,
		GithubAlias:    input["github_alias"],
		SlackID:        input["reviewer_id"],
		TechnologyList: input["technology_list"],
		ChallengeName:  input["challenge_name"],
		Experience:     experience,
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

func UpdateReviewer(env config.Environment, reviewer Reviewer) error {
	store, err := db.NewStore(env, db.ReviewersCollection)
	if err != nil {
		return err
	}
	return store.Update(reviewer.ID, reviewer)
}
