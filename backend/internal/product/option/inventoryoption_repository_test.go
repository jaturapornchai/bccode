package option_test

import (
	"context"
	"os"
	"smlcloudplatform/internal/product/option"
	"smlcloudplatform/internal/product/option/models"
	"smlcloudplatform/mock"
	"smlcloudplatform/pkg/microservice"
	"testing"
)

func getInventoryOptionMainRepo() option.OptionRepository {
	mongoPersisterConfig := mock.NewPersisterMongoConfig()
	mongoPersister := microservice.NewPersisterMongo(mongoPersisterConfig)
	repo := option.NewOptionRepository(mongoPersister)
	return *repo
}

func TestCreateInventoryOptionMain(t *testing.T) {

	if os.Getenv("SERVERLESS") == "serverless" {
		t.Skip()
	}
	repo := getInventoryOptionMainRepo()

	give := models.InventoryOptionMainDoc{}

	give.GuidFixed = "fx01"
	give.ShopID = "mx01"
	give.Code = "code001"

	_, err := repo.Create(context.TODO(), give)

	if err != nil {
		t.Error(err)
		return
	}

	findDoc, err := repo.FindByGuid(context.TODO(), "mx01", "fx01")

	if err != nil {
		t.Error(err)
		return
	}

	t.Log(findDoc)
}
