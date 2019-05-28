package models

import (
	"errors"
	"reflect"

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

func UpdateReviewer(env config.Environment, reviewer Reviewer) error {
	store, err := db.NewStore(env, db.ReviewersCollection)
	if err != nil {
		return err
	}
	return store.Update(reviewer.ID, reviewer)
}
