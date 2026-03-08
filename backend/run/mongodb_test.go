package run_test

import (
	"context"
	"smlcloudplatform/internal/authentication/models"
	"testing"

	"github.com/tj/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

// func TestMain(m *testing.M) {
// 	// All tests that use mtest.Setup() are expected to be integration tests, so skip them when the
// 	// -short flag is included in the "go test" command. Also, we have to parse flags here to use
// 	// testing.Short() because flags aren't parsed before TestMain() is called.
// 	flag.Parse()
// 	if testing.Short() {
// 		log.Print("skipping mtest integration test in short mode")
// 		return
// 	}

// 	if err := mtest.Setup(); err != nil {
// 		log.Fatal(err)
// 	}
// 	defer os.Exit(m.Run())
// 	if err := mtest.Teardown(); err != nil {
// 		log.Fatal(err)
// 	}
// }

func TestMongodbCreateData(t *testing.T) {

	// ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	// defer cancel()

	// client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mtest.ClusterURI()))
	// require.NoError(t, err)
	// defer client.Disconnect(ctx)

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()
	mt.Run("Success", func(mt *mtest.T) {
		// test code
		userCollection := mt.DB.Collection("TEST")
		id := primitive.NewObjectID()

		mt.AddMockResponses(mtest.CreateSuccessResponse())

		giveUser := models.UserDoc{}

		giveUser.ID = id
		giveUser.Username = "john"

		insertedUser, err := userCollection.InsertOne(context.Background(), giveUser)

		assert.Nil(t, err)
		assert.Equal(t, giveUser.ID, insertedUser.InsertedID)
	})
}
