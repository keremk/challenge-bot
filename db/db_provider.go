package db

import (
	"github.com/keremk/challenge-bot/config"
)

const SlackUsersCollection = "slackusers"
const SlackTeamsCollection = "slackteams"

type CrudOps interface {
	Update(key string, obj interface{}) error
	FindByID(id string, obj interface{}) error
}

func NewStore(env config.Environment, collection string) CrudOps {
	switch env.DbProvider {
	case Firestore:
		return FirestoreDb{
			env:        env,
			collection: collection,
		}
	default:
		return nil
	}
}
