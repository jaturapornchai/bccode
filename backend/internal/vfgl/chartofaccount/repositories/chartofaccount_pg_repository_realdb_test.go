package repositories_test

import (
	"os"
	"smlcloudplatform/internal/vfgl/chartofaccount/repositories"
	"smlcloudplatform/mock"
	"smlcloudplatform/pkg/microservice"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var repo repositories.ChartOfAccountPgRepository

func init() {
	persisterConfig := mock.NewPersisterPostgresqlConfig()
	pst := microservice.NewPersister(persisterConfig)
	repo = repositories.NewChartOfAccountPgRepository(pst)
}

func TestChartOfAccountRepositoryCreateInRealDB(t *testing.T) {

	if os.Getenv("SERVERLESS") == "serverless" {
		t.Skip()
	}
	assert := assert.New(t)
	assert.NotNil(repo)

	// give := &vfgl.ChartOfAccountPG{
	// 	ShopIdentity: models.ShopIdentity{
	// 		ShopID: "SHOPTEST",
	// 	},
	// 	AccountCode: "10000",
	// 	AccountName: "เงินสด",
	// }

	// err := repo.Create(*give)
	// assert.Nil(err)

	get, err := repo.Get("SHOPTEST", "10099")
	assert.Nil(err)
	assert.NotNil(get)

}

func TestChartOfAccountRepositoryGetDataInRealDBFirstAssertErrorNotFound(t *testing.T) {

	if os.Getenv("SERVERLESS") == "serverless" {
		t.Skip()
	}
	assert := assert.New(t)
	assert.NotNil(repo)

	get, err := repo.Get("SHOPTEST", "10099")
	assert.ErrorIs(err, gorm.ErrRecordNotFound, "Assert Not Found Record is Not Match")
	assert.Nil(get)
}
