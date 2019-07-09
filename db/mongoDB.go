package db

import (
	"context"
	"log"
	"reflect"

	// "go.mongodb.org/mongo-driver/bson"
	"github.com/keremk/challenge-bot/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const Mongo = "MongoDB"

type MongoDB struct {
	env        config.Environment
	collection string
	database   string
}

func (s MongoDB) Update(key string, obj interface{}) error {
	client, ctx, err := s.getClient()
	if err != nil {
		return err
	}

	col := client.Database(s.database).Collection(s.collection)

	_, err = col.InsertOne(ctx, obj)
	if err != nil {
		log.Printf("[ERROR] Unable to insert to MongoDB - %s", err)
		return err
	}

	return nil
}

func (s MongoDB) Merge(key string, values map[string]interface{}) error {
	client, ctx, err := s.getClient()
	if err != nil {
		return err
	}

	col := client.Database(s.database).Collection(s.collection)

	filter := bson.D{{"ID", key}}

	updateValues := make(bson.D, len(values))
	for k, v := range values {
		updateValues = append(updateValues, bson.E{Key: k, Value: v})
	}

	update := bson.D{{
		"$set", values,
	}}
	_, err = col.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Printf("[ERROR] Unable to update document in MongoDB - %s", err)
	}
	return nil
}

func (s MongoDB) FindByID(id string, obj interface{}) error {

	return s.FindFirst("ID", id, obj)
}

func (s MongoDB) FindFirst(key, value string, obj interface{}) error {
	client, ctx, err := s.getClient()
	if err != nil {
		return err
	}

	col := client.Database(s.database).Collection(s.collection)

	filter := bson.D{{key, value}}
	err = col.FindOne(ctx, filter).Decode(obj)
	if err != nil {
		log.Printf("[ERROR] Cannot find the document with key - %s", err)
	}

	return err
}

func (s MongoDB) FindAll(itemType reflect.Type) (interface{}, error) {

	return s.find(itemType, "", "")
}

func (s MongoDB) FindAllWithKeyValue(itemType reflect.Type, key, value string) (interface{}, error) {

	return s.find(itemType, key, value)
}

func (s MongoDB) find(itemType reflect.Type, key, value string) (interface{}, error) {
	client, ctx, err := s.getClient()
	if err != nil {
		return nil, err
	}

	col := client.Database(s.database).Collection(s.collection)

	findOptions := options.Find()
	findOptions.SetLimit(100)

	var filter interface{}
	if key == "" {
		filter = bson.D{{}}
	} else {
		filter = bson.D{{key, value}}
	}

	results := reflect.MakeSlice(itemType, 0, 100)

	// Finding multiple documents returns a cursor
	cur, err := col.Find(ctx, filter, findOptions)
	if err != nil {
		log.Println("[ERROR] Cannot find any matching results - ", err)
		return nil, err
	}

	// Iterate through the cursor
	for cur.Next(ctx) {
		item := reflect.New(itemType.Elem())
		err := cur.Decode(item.Interface())
		if err != nil {
			log.Println("[ERROR] Cannot decode result - ", err)
		}

		results = reflect.Append(results, item.Elem())
	}

	return results.Interface(), err
}

func (s MongoDB) getClient() (*mongo.Client, context.Context, error) {
	ctx := context.Background()

	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	return client, ctx, err
}
