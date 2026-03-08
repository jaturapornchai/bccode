package repositories_test

import (
	"context"
	"os"
	"smlcloudplatform/internal/authentication/models"
	"smlcloudplatform/internal/authentication/repositories"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/mock"
	"smlcloudplatform/pkg/microservice"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var repoMock repositories.IAuthenticationRepository

func init() {
	mongoPersisterConfig := mock.NewPersisterMongoConfig()
	mongoPersister := microservice.NewPersisterMongo(mongoPersisterConfig)
	repoMock = repositories.NewAuthenticationRepository(mongoPersister)
}

// mock Persister

func TestFindUser(t *testing.T) {

	if os.Getenv("SERVERLESS") == "serverless" {
		t.Skip()
	}
	password, _ := utils.HashPassword("test")

	username := models.UsernameField{
		Username: "test",
	}

	userPass := models.UserPassword{
		Password: password,
	}

	userDetail := models.UserDetail{
		Name: "test",
	}

	createAt := time.Now()
	give := &models.UserDoc{
		UsernameField: username,
		UserPassword:  userPass,
		UserDetail:    userDetail,
		CreatedAt:     createAt,
	}

	want := &models.UserDoc{
		UsernameField: username,
		UserPassword:  userPass,
		UserDetail:    userDetail,
		CreatedAt:     createAt,
	}

	createUserID, err := repoMock.CreateUser(context.TODO(), *give)
	if err != nil {
		t.Error(err.Error())
		return
	}

	require.NotEqual(t, createUserID, primitive.NilObjectID, "Create User Failed")

	getUser, err := repoMock.FindUser(context.TODO(), want.Username)

	if err != nil {
		t.Error(err.Error())
		return
	}

	assert.Equal(t, getUser.Username, want.Username, "Create User And Find Not Match")

}
