package db

import (
	"log"
	"reflect"
	"testing"

	"github.com/keremk/challenge-bot/config"
	"github.com/stretchr/testify/assert"
)

func getTestDB() MongoDB {
	env := config.NewEnvironment("production")
	mongoDB := MongoDB{
		env:        env,
		collection: "testcollection",
		database:   "test",
	}

	client, ctx, err := mongoDB.getClient()
	if err != nil {
		log.Fatal("[ERROR] Cannot connect")
	}

	err = client.Database(mongoDB.database).Drop(ctx)

	return mongoDB
}

func TestConnectingMongoDB(t *testing.T) {
	env := config.NewEnvironment("production")
	collection := "challengesettings"

	s := MongoDB{
		env:        env,
		collection: collection,
		database:   "test",
	}

	_, _, err := s.getClient()
	assert.Nil(t, err, "Expected the operation to be nil")
}

type testEntry struct {
	ID           string `bson:"ID"`
	WorkingTitle string `bson:"working_title"`
	Category     string
}

type testData struct {
	Name       string
	PricePoint int
	Amap       map[string]*testEntry
	Bmap       map[string][]string
}

func getUpdateFixture() testData {
	amap := map[string]*testEntry{
		"foo1": &testEntry{
			WorkingTitle: "Bar1",
			Category:     "Cat1",
		},
		"baz1": &testEntry{
			WorkingTitle: "Foobaz1",
			Category:     "Cat2",
		},
	}

	bmap := map[string][]string{
		"foo2": []string{"bar2", "bar3"},
		"baz2": []string{"foobaz2", "foobaz2"},
	}

	return testData{
		Name:       "Hello World",
		PricePoint: 100,
		Amap:       amap,
		Bmap:       bmap,
	}
}
func TestInsertingDocInMongoDB(t *testing.T) {
	db := getTestDB()

	err := db.Update("TestKey", getUpdateFixture())

	assert.Nil(t, err, "Expected the operation to be nil")

}

func addSearchableDocs(db MongoDB) error {
	client, ctx, err := db.getClient()
	if err != nil {
		return err
	}

	col := client.Database(db.database).Collection(db.collection)

	fixtures := []interface{}{
		&testEntry{ID: "100", WorkingTitle: "Foo", Category: "Cat1"},
		&testEntry{ID: "101", WorkingTitle: "Bar", Category: "Cat1"},
		&testEntry{ID: "102", WorkingTitle: "Baz", Category: "Cat2"},
	}

	_, err = col.InsertMany(ctx, fixtures)

	return err
}

func TestFindingOneDocInMongoDB(t *testing.T) {
	db := getTestDB()

	err := addSearchableDocs(db)
	assert.Nil(t, err, "Could not add the fixtures to test database")

	obj := testEntry{}
	err = db.FindFirst("working_title", "Bar", &obj)
	assert.Nil(t, err, "Could not find what I am looking for")
	assert.Equal(t, testEntry{ID: "101", WorkingTitle: "Bar", Category: "Cat1"}, obj)
}

func TestFindingManyDocsWithKeyValueInMongoDB(t *testing.T) {
	db := getTestDB()

	err := addSearchableDocs(db)
	assert.Nil(t, err, "Could not add the fixtures to test database")

	var all []testEntry
	result, err := db.FindAllWithKeyValue(reflect.TypeOf(all), "category", "Cat1")
	all, ok := result.([]testEntry)

	assert.Equal(t, true, ok)
	assert.Nil(t, err, "Could not find what I am looking for")
	assert.Equal(t, testEntry{ID: "100", WorkingTitle: "Foo", Category: "Cat1"}, all[0])
	assert.Equal(t, testEntry{ID: "101", WorkingTitle: "Bar", Category: "Cat1"}, all[1])
}

func TestFindingAllDocsInMongoDB(t *testing.T) {
	db := getTestDB()

	err := addSearchableDocs(db)
	assert.Nil(t, err, "Could not add the fixtures to test database")

	var all []testEntry
	result, err := db.FindAll(reflect.TypeOf(all))
	all, ok := result.([]testEntry)

	assert.Equal(t, true, ok)
	assert.Nil(t, err, "Could not find what I am looking for")
	assert.Equal(t, testEntry{ID: "100", WorkingTitle: "Foo", Category: "Cat1"}, all[0])
	assert.Equal(t, testEntry{ID: "101", WorkingTitle: "Bar", Category: "Cat1"}, all[1])
	assert.Equal(t, testEntry{ID: "102", WorkingTitle: "Baz", Category: "Cat2"}, all[2])
}

func TestUpdatingDocInMongoDB(t *testing.T) {
	db := getTestDB()

	err := addSearchableDocs(db)
	assert.Nil(t, err, "Could not add the fixtures to test database")

	values := map[string]interface{}{
		"working_title": "owner",
	}
	err = db.Merge("100", values)
	assert.Nil(t, err, "Could not update the document")

	var obj testEntry
	err = db.FindByID("100", &obj)
	assert.Nil(t, err, "Could not find the updated document")
	assert.Equal(t, testEntry{ID: "100", WorkingTitle: "owner", Category: "Cat1"}, obj)
}
