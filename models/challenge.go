package models

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/keremk/challenge-bot/config"
	"github.com/keremk/challenge-bot/db"
	"github.com/keremk/challenge-bot/util"
)

type SlotID = string

type Slot struct {
	ID        string
	Ordinal   int
	Name      string
	Day       string
	StartTime string
	EndTime   string
}

type Challenge struct {
	ID                string
	Name              string
	GithubAccountName string
	TemplateRepo      string
	RepoNameFormat    string
	CreatedByTeamID   string
	Slots             map[SlotID]*Slot
}

func NewChallenge(input map[string]string) Challenge {
	name := input["challenge_name"]
	id := fmt.Sprintf("%s-%s", name, util.RandomString(8))

	return Challenge{
		ID:                id,
		Name:              name,
		GithubAccountName: input["github_account"],
		TemplateRepo:      input["template_repo"],
		RepoNameFormat:    input["repo_name_format"],
		CreatedByTeamID:   input["team_id"],
	}
}

func EditChallenge(env config.Environment, input map[string]string, challengeID string) (Challenge, error) {
	challenge, err := getChallengeByID(env, challengeID)
	if err != nil {
		return Challenge{}, err
	}

	return Challenge{
		ID:                challengeID,
		Name:              input["challenge_name"],
		GithubAccountName: input["github_account"],
		TemplateRepo:      input["template_repo"],
		RepoNameFormat:    input["repo_name_format"],
		CreatedByTeamID:   input["team_id"],
		Slots:             challenge.Slots,
	}, nil
}

func getChallengeByName(env config.Environment, name string) (Challenge, error) {
	challenge := Challenge{}
	store, err := db.NewStore(env, db.SettingsCollection)
	if err != nil {
		return challenge, err
	}

	err = store.FindFirst("Name", name, &challenge)
	return challenge, err
}

func getChallengeByID(env config.Environment, id string) (Challenge, error) {
	challenge := Challenge{}
	store, err := db.NewStore(env, db.SettingsCollection)
	if err != nil {
		return challenge, err
	}

	err = store.FindByID(id, &challenge)
	return challenge, err
}

func GetAllChallenges(env config.Environment) ([]Challenge, error) {
	store, err := db.NewStore(env, db.SettingsCollection)
	if err != nil {
		return nil, err
	}

	var all []Challenge
	result, err := store.FindAll(reflect.TypeOf(all))
	all, ok := result.([]Challenge)
	if !ok {
		return nil, errors.New("[ERROR] Cannot convert")
	}
	return all, err
}

func UpdateChallenge(env config.Environment, challenge Challenge) error {
	store, err := db.NewStore(env, db.SettingsCollection)
	if err != nil {
		return err
	}
	if challenge.Slots == nil || len(challenge.Slots) == 0 {
		challenge.Slots = defaultSlots()
	}
	return store.Update(challenge.ID, challenge)
}

func defaultSlots() map[SlotID]*Slot {
	slots := make(map[SlotID]*Slot)
	ordinal := 0
	for i := 0; i < 5; i++ {
		day := time.Weekday(i + 1)
		slotID := fmt.Sprintf("%sMorning", day.String())
		slots[slotID] = &Slot{
			ID:        slotID,
			Ordinal:   ordinal,
			Name:      fmt.Sprintf("%s Morning", day.String()),
			Day:       day.String(),
			StartTime: "9:00",
			EndTime:   "11:00",
		}

		slotID = fmt.Sprintf("%sAfternoon", day.String())
		ordinal++
		slots[slotID] = &Slot{
			ID:        slotID,
			Ordinal:   ordinal,
			Name:      fmt.Sprintf("%s Afternoon", day.String()),
			Day:       day.String(),
			StartTime: "16:30",
			EndTime:   "18:30",
		}
		ordinal++
	}
	return slots
}
