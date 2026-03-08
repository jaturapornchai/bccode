package shop_test

import (
	"context"
	"os"
	"smlcloudplatform/internal/authentication/models"
	"smlcloudplatform/internal/shop"
	"smlcloudplatform/pkg/microservice"
	"testing"
)

type TestPersisterMongoConfig struct{}

func (TestPersisterMongoConfig) MongodbURI() string {
	return "mongodb://root:rootx@localhost:27017/"
}

func (TestPersisterMongoConfig) DB() string {
	return "micro_test"
}

func (TestPersisterMongoConfig) Debug() bool {
	return true
}

var shopUserRepo shop.IShopUserRepository

func setup() {
	pst := microservice.NewPersisterMongo(&TestPersisterMongoConfig{})

	shopUserRepo = shop.NewShopUserRepository(pst)
}

func TestShopMemberSave(t *testing.T) {

	if os.Getenv("SERVERLESS") == "serverless" {
		t.Skip()
	}
	setup()

	err := shopUserRepo.Save(context.TODO(), "mx1", "ux3", models.ROLE_OWNER)

	if err != nil {
		t.Error(err)
	}
}

func TestShopMemberFindRole(t *testing.T) {
	if os.Getenv("SERVERLESS") == "serverless" {
		t.Skip()
	}
	setup()

	role, err := shopUserRepo.FindRole(context.TODO(), "mx1", "ux3")

	if err != nil {
		t.Error(err)
	}

	t.Log(role)
}

func TestShopMemberFindByShop(t *testing.T) {
	if os.Getenv("SERVERLESS") == "serverless" {
		t.Skip()
	}
	setup()

	members, err := shopUserRepo.FindByShopID(context.TODO(), "mx1")

	if err != nil {
		t.Error(err)
		return
	}

	t.Log(members)
}

func TestShopMemberFindByUsername(t *testing.T) {
	if os.Getenv("SERVERLESS") == "serverless" {
		t.Skip()
	}
	setup()

	members, err := shopUserRepo.FindByUsername(context.TODO(), "ux1")

	if err != nil {
		t.Error(err)
		return
	}

	t.Log(members)
}
