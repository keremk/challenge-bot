package db

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"
	"github.com/keremk/challenge-bot/config"
)

const Firestore = "Firestore"

type FirestoreDb struct {
	env        config.Environment
	collection string
}

func (s FirestoreDb) Update(key string, obj interface{}) error {
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, s.env.GCloudProjectID)
	if err != nil {
		log.Println("[ERROR] cannot connect to Firestore", err)
		return err
	}
	defer client.Close()

	_, err = client.Collection(s.collection).Doc(key).Set(ctx, obj)
	if err != nil {
		log.Println("[ERROR] cannot write data to Firestore", err)
	}
	return err
}
