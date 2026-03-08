package shop_test

import (
	"context"
	"fmt"
	"os"
	"smlcloudplatform/internal/authentication/models"
	"smlcloudplatform/internal/shop"
	"smlcloudplatform/mock"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestCreateShopUser(t *testing.T) {
	if os.Getenv("SERVERLESS") == "serverless" {
		t.Skip()
	}

	mongoPersisterConfig := mock.NewPersisterMongoConfig()
	mongoPersister := microservice.NewPersisterMongo(mongoPersisterConfig)
	repo := shop.NewShopUserRepository(mongoPersister)

	err := repo.Save(context.TODO(), "25H2pZ8v2jRVGwjOLKBAzSaHgOA", "dev01", models.ROLE_USER)

	if err != nil {
		t.Error(err.Error())
		return
	}

	memUser, err := repo.FindByShopIDAndUsername(context.TODO(), "25H2pZ8v2jRVGwjOLKBAzSaHgOA", "dev01")

	if err != nil {
		t.Error(err.Error())
		return
	}

	fmt.Printf("%v \n", memUser)
	if memUser.ID == primitive.NilObjectID {
		t.Error("find error")
		return
	}

}

func TestFindByUsernamePage(t *testing.T) {

	if os.Getenv("SERVERLESS") == "serverless" {
		t.Skip()
	}
	mongoPersisterConfig := mock.NewPersisterMongoConfig()
	t.Log(mongoPersisterConfig.DB())
	mongoPersister := microservice.NewPersisterMongo(mongoPersisterConfig)
	repo := shop.NewShopUserRepository(mongoPersister)
	pageable := micromodels.Pageable{
		Page:  1,
		Limit: 20,
		Query: "",
	}
	docList, paginated, err := repo.FindByUsernamePage(context.TODO(), "dev01", pageable)

	if err != nil {
		t.Error(err.Error())
		return
	}

	t.Log(paginated.TotalPage)
	t.Log(paginated.Total)
	t.Log(docList)

}
