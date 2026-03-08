package stockprocess_test

import (
	"smlcloudplatform/internal/config"
	productbarcoderepository "smlcloudplatform/internal/product/productbarcode/repositories"
	"smlcloudplatform/internal/stockprocess"
	"smlcloudplatform/internal/stockprocess/repositories"
	"smlcloudplatform/pkg/microservice"
	"testing"
)

var repo repositories.IStockProcessPGRepository
var productBarcodePGRepository productbarcoderepository.IProductBarcodePGRepository

func init() {
	cfg := config.NewConfig()
	persister := microservice.NewPersister(cfg.PersisterConfig())
	repo = repositories.NewStockProcessPGRepository(persister)
	productBarcodePGRepository = productbarcoderepository.NewProductBarcodePGRepository(persister)
}

func TestStockProcessRealDBTest(t *testing.T) {

	// stockLists, err := repo.GetStockTransactionList("2IZS0jFeRXWPidSupyXN7zQIlaS", "888555")
	// assert.Nil(t, err)
	// assert.NotNil(t, stockLists)
	// assert.Equal(t, 2, len(stockLists))

	process := stockprocess.NewStockCalculator(repo, productBarcodePGRepository)
	process.CalculatorStock("2IZS0jFeRXWPidSupyXN7zQIlaS", "888555")
}
