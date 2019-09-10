package models

import (
	"github.com/keremk/challenge-bot/config"
	"github.com/keremk/challenge-bot/db"
)

type SlackTeam struct {
	ID        string `bson:"ID"`
	Name      string `bson:"Name"`
	BotToken  string `bson:"BotToken"`
	BotUserID string `bson:"BotUserID"`
}

func GetSlackTeam(env config.Environment, id string) (SlackTeam, error) {
	teamID := getTeamID(env, id)
	team := SlackTeam{}
	store, err := db.NewStore(env, db.SlackTeamsCollection)
	if err != nil {
		return team, err
	}

	err = store.FindByID(teamID, &team)
	return team, err
}

func UpdateSlackTeam(env config.Environment, team SlackTeam) error {
	teamID := getTeamID(env, team.ID)
	store, err := db.NewStore(env, db.SlackTeamsCollection)
	if err != nil {
		return err
	}
	return store.Update(teamID, team)
}

// This is for testing reasons
// In our test Slack app (ChallengeTest) we always use the hardcoded "ADMIN" as team ID to
// ensure we are not messing up with the production DB.
func getTeamID(env config.Environment, id string) string {
	var teamIDLookup string
	if env.DebugOn {
		teamIDLookup = "TGB941BGQ"
	} else {
		teamIDLookup = id
	}
	return teamIDLookup
}
