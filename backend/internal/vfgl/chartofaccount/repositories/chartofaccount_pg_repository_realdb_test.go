//go:build integration

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

func newRealDBRepository(t *testing.T) repositories.ChartOfAccountPgRepository {
	t.Helper()
	if os.Getenv("BC_REAL_DB_TESTS") != "1" {
		t.Skip("set BC_REAL_DB_TESTS=1 to run PostgreSQL integration tests")
	}

	persisterConfig := mock.NewPersisterPostgresqlConfig()
	pst := microservice.NewPersister(persisterConfig)
	return repositories.NewChartOfAccountPgRepository(pst)
}

func TestChartOfAccountRepositoryCreateInRealDB(t *testing.T) {
	repo := newRealDBRepository(t)

	assert := assert.New(t)
	assert.NotNil(repo)

	// give := &vfgl.ChartOfAccountPG{
	// 	HoldingCodeentity: models.HoldingCodeentity{
	// 		HoldingCode: "SHOPTEST",
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
	repo := newRealDBRepository(t)

	assert := assert.New(t)
	assert.NotNil(repo)

	get, err := repo.Get("SHOPTEST", "10099")
	assert.ErrorIs(err, gorm.ErrRecordNotFound, "Assert Not Found Record is Not Match")
	assert.Nil(get)
}
