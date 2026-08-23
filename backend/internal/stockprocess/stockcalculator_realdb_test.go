//go:build integration

package stockprocess_test

import (
	"os"
	"smlcloudplatform/internal/config"
	productbarcoderepository "smlcloudplatform/internal/product/productbarcode/repositories"
	"smlcloudplatform/internal/stockprocess"
	"smlcloudplatform/internal/stockprocess/repositories"
	"smlcloudplatform/pkg/microservice"
	"testing"
)

func newRealDBRepositories(t *testing.T) (repositories.IStockProcessPGRepository, productbarcoderepository.IProductBarcodePGRepository) {
	t.Helper()
	if os.Getenv("BC_REAL_DB_TESTS") != "1" {
		t.Skip("set BC_REAL_DB_TESTS=1 to run PostgreSQL integration tests")
	}

	cfg := config.NewConfig()
	persister := microservice.NewPersister(cfg.PersisterConfig())
	return repositories.NewStockProcessPGRepository(persister), productbarcoderepository.NewProductBarcodePGRepository(persister)
}

func TestStockProcessRealDBTest(t *testing.T) {
	repo, productBarcodePGRepository := newRealDBRepositories(t)

	// stockLists, err := repo.GetStockTransactionList("2IZS0jFeRXWPidSupyXN7zQIlaS", "888555")
	// assert.Nil(t, err)
	// assert.NotNil(t, stockLists)
	// assert.Equal(t, 2, len(stockLists))

	process := stockprocess.NewStockCalculator(repo, productBarcodePGRepository)
	process.CalculatorStock("2IZS0jFeRXWPidSupyXN7zQIlaS", "888555")
}
