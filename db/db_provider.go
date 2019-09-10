package db

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/keremk/challenge-bot/config"
)

const SlackUsersCollection = "slackusers"
const SlackTeamsCollection = "slackteams"
const SettingsCollection = "challengesettings"
const GithubAccountsCollection = "githubaccounts"
const ReviewersCollection = "reviewers"

type CrudOps interface {
	Update(key string, obj interface{}) error
	Merge(key string, values map[string]interface{}) error
	FindByID(id string, obj interface{}) error
	FindFirst(key, value string, obj interface{}) error
	FindAll(itemType reflect.Type) (interface{}, error)
	FindAllWithKeyValue(itemType reflect.Type, key, value string) (interface{}, error)
}

func NewStore(env config.Environment, collection string) (CrudOps, error) {
	switch env.DbProvider {
	case Firestore:
		return FirestoreDb{
			env:        env,
			collection: collection,
		}, nil
	case PostgreSQL:
		return PostgreSQLDB{
			env:   env,
			table: collection,
		}, nil
	case Mongo:
		return MongoDB{
			env:        env,
			collection: collection,
			database:   env.MongoDBDatabaseName,
		}, nil
	default:
		errMsg := fmt.Sprintf("[ERROR] db provider not known or unspecified - %s", env.DbProvider)
		return nil, errors.New(errMsg)
	}
}
