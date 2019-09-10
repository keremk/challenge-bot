package main

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"
	"github.com/keremk/challenge-bot/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type migration struct {
	env     config.Environment
	mClient *mongo.Client
	fClient *firestore.Client
	mCtx    context.Context
	fCtx    context.Context
}

func newMigration(env config.Environment) migration {
	fClient, fCtx, err := getFirestoreClient(env)
	if err != nil {
		log.Fatal("Cannot get Firestore client ", err)
	}
	mClient, mCtx, err := getMongoClient(env)
	if err != nil {
		log.Fatal("Cannot get MongoDB client ", err)
	}

	return migration{
		env:     env,
		mClient: mClient,
		fClient: fClient,
		mCtx:    mCtx,
		fCtx:    fCtx,
	}
}

type c2c struct {
	fc string
	mc string
}

func main() {
	c2cMap := []c2c{
		c2c{fc: "challengesettings", mc: "challengesettings"},
		c2c{fc: "githubaccounts", mc: "githubaccounts"},
		c2c{fc: "reviewers", mc: "reviewers"},
		c2c{fc: "slackteams", mc: "slackteams"},
		c2c{fc: "slackusers", mc: "slackusers"},
	}
	env := config.NewEnvironment("production")

	m := newMigration(env)

	for _, mapping := range c2cMap {
		log.Printf("Migrating %s to %s", mapping.fc, mapping.mc)
		rows, err := m.getColRows(mapping.fc)
		if err != nil {
			log.Fatalf("Getting rows failed for %s with error %s", mapping.fc, err)
		}
		if len(rows) == 0 {
			log.Printf("Zero rows found for %s", mapping.fc)
		}
		err = m.saveColRows(mapping.mc, rows)
		if err != nil {
			log.Fatalf("Savings rows failed for %s with error %s", mapping.mc, err)
		}
	}
}

func getFirestoreClient(env config.Environment) (*firestore.Client, context.Context, error) {
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, env.GCloudProjectID)
	if err != nil {
		log.Println("[ERROR] cannot connect to Firestore", err)
		return nil, ctx, err
	}

	return client, ctx, err
}

func getMongoClient(env config.Environment) (*mongo.Client, context.Context, error) {
	ctx := context.Background()

	clientOptions := options.Client().ApplyURI(env.MongoDBConnectionString)
	client, err := mongo.Connect(ctx, clientOptions)

	return client, ctx, err
}

func (m migration) getColRows(colName string) ([]map[string]interface{}, error) {
	docs, err := m.fClient.Collection(colName).Documents(m.fCtx).GetAll()
	if err != nil {
		return nil, err
	}
	rows := make([]map[string]interface{}, len(docs))
	for i, doc := range docs {
		row := doc.Data()
		rows[i] = row
	}

	return rows, nil
}

func (m migration) saveColRows(colName string, rows []map[string]interface{}) error {
	col := m.mClient.Database(m.env.MongoDBDatabaseName).Collection(colName)

	for _, row := range rows {
		_, err := col.InsertOne(m.mCtx, row)
		if err != nil {
			return err
		}
	}
	return nil
}
