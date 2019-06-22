package db

import (
	"context"
	"errors"
	"log"
	"reflect"

	"cloud.google.com/go/firestore"
	"github.com/keremk/challenge-bot/config"
)

const Firestore = "Firestore"

type FirestoreDb struct {
	env        config.Environment
	collection string
}

func (s FirestoreDb) Update(key string, obj interface{}) error {
	client, ctx, err := s.getClient()
	if err != nil {
		return err
	}
	defer client.Close()

	_, err = client.Collection(s.collection).Doc(key).Set(ctx, obj)
	if err != nil {
		log.Printf("[ERROR] cannot merge data to Firestore for key %s - %s", key, err)
	}
	return err
}

func (s FirestoreDb) Merge(key string, values map[string]interface{}) error {
	client, ctx, err := s.getClient()
	if err != nil {
		return err
	}
	defer client.Close()

	opts := firestore.MergeAll

	_, err = client.Collection(s.collection).Doc(key).Set(ctx, values, opts)
	if err != nil {
		log.Printf("[ERROR] cannot merge data to Firestore for key %s - %s", key, err)
	}
	return err
}

func (s FirestoreDb) FindByID(id string, obj interface{}) error {
	client, ctx, err := s.getClient()
	if err != nil {
		return err
	}
	defer client.Close()

	data, err := client.Collection(s.collection).Doc(id).Get(ctx)
	if err != nil {
		log.Println("[ERROR] cannot find object with id=", id, err)
		return err
	}
	return data.DataTo(obj)
}

func (s FirestoreDb) FindFirst(key, value string, obj interface{}) error {
	client, ctx, err := s.getClient()
	if err != nil {
		return err
	}
	defer client.Close()

	docs, err := client.Collection(s.collection).Where(key, "==", value).Documents(ctx).GetAll()
	if err != nil {
		log.Println("[ERROR] cannot find object - ", err)
		return err
	}
	if len(docs) < 1 {
		err = errors.New("Empty collection returned")
		log.Println("[ERROR] cannot find object - ", err)
		return err
	}
	return docs[0].DataTo(obj)
}

type searchFunc = func(client *firestore.Client, ctx context.Context) ([]*firestore.DocumentSnapshot, error)

func (s FirestoreDb) FindAllWithKeyValue(itemType reflect.Type, key, value string) (interface{}, error) {
	return s.find(itemType, func(client *firestore.Client, ctx context.Context) ([]*firestore.DocumentSnapshot, error) {
		docs, err := client.Collection(s.collection).Where(key, "==", value).Documents(ctx).GetAll()
		return docs, err
	})
}

func (s FirestoreDb) FindAll(itemType reflect.Type) (interface{}, error) {
	return s.find(itemType, func(client *firestore.Client, ctx context.Context) ([]*firestore.DocumentSnapshot, error) {
		docs, err := client.Collection(s.collection).Documents(ctx).GetAll()
		return docs, err
	})
}

func (s FirestoreDb) find(itemType reflect.Type, searchFunc searchFunc) (interface{}, error) {
	if itemType.Kind() != reflect.Slice {
		panic("FindAll is expecting a type of kind slice")
	}

	client, ctx, err := s.getClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()

	docs, err := searchFunc(client, ctx)
	if err != nil || (len(docs) < 1) {
		log.Println("[ERROR] cannot find object", err)
		return nil, err
	}

	slice := reflect.MakeSlice(itemType, 0, 100)
	for _, doc := range docs {
		item := reflect.New(itemType.Elem())
		err := doc.DataTo(item.Interface())
		if err != nil {
			return nil, err
		}
		slice = reflect.Append(slice, item.Elem())
	}

	return slice.Interface(), err
}

func (s FirestoreDb) getClient() (*firestore.Client, context.Context, error) {
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, s.env.GCloudProjectID)
	if err != nil {
		log.Println("[ERROR] cannot connect to Firestore", err)
		return nil, ctx, err
	}

	return client, ctx, err
}
