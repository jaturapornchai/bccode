package repositories_test

import (
	"os"
	"smlcloudplatform/internal/config"
	"smlcloudplatform/internal/stockprocess/repositories"
	"smlcloudplatform/pkg/microservice"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newRealDBRepository(t *testing.T) repositories.IStockProcessPGRepository {
	t.Helper()
	if os.Getenv("BC_REAL_DB_TESTS") != "1" {
		t.Skip("set BC_REAL_DB_TESTS=1 to run PostgreSQL integration tests")
	}

	cfg := config.NewConfig()
	persister := microservice.NewPersister(cfg.PersisterConfig())
	return repositories.NewStockProcessPGRepository(persister)
}

func TestGetStockProcessList(t *testing.T) {
	repo := newRealDBRepository(t)

	stockLists, err := repo.GetStockTransactionList("2IZS0jFeRXWPidSupyXN7zQIlaS", "888555")
	assert.Nil(t, err)
	assert.NotNil(t, stockLists)
	assert.Equal(t, 2, len(stockLists))
}

func TestUpdateStockTransaction(t *testing.T) {
	repo := newRealDBRepository(t)

	stockLists, err := repo.GetStockTransactionList("2VsCV0xYjghds3Tjru425QKGkY1", "8851753098736")
	assert.Nil(t, err)

	for i, _ := range stockLists {
		stockLists[i].CostPerUnit = float64(1)
		stockLists[i].TotalCost = float64(1)
		stockLists[i].BalanceQty = float64(1)
		stockLists[i].BalanceAmount = float64(1)
		stockLists[i].BalanceAverage = float64(1)
	}

	err = repo.UpdateStockTransactionChange(stockLists)
	assert.Nil(t, err)
}
