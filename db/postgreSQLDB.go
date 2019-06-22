package db

import (
	"reflect"

	"github.com/keremk/challenge-bot/config"
)

const PostgreSQL = "PostgreSQL"

type PostgreSQLDB struct {
	env   config.Environment
	table string
}

func (s PostgreSQLDB) Update(key string, obj interface{}) error {

	return nil
}

func (s PostgreSQLDB) Merge(key string, values map[string]interface{}) error {

	return nil
}

func (s PostgreSQLDB) FindByID(id string, obj interface{}) error {

	return nil
}

func (s PostgreSQLDB) FindFirst(key, value string, obj interface{}) error {

	return nil
}

func (s PostgreSQLDB) FindAll(itemType reflect.Type) (interface{}, error) {

	return nil, nil
}

func (s PostgreSQLDB) FindAllWithKeyValue(itemType reflect.Type, key, value string) (interface{}, error) {

	return nil, nil
}
