//go:build integration

package repositories_test

import (
	"os"
	"smlcloudplatform/internal/config"
	"smlcloudplatform/internal/product/productbarcode/models"
	"smlcloudplatform/internal/product/productbarcode/repositories"
	"smlcloudplatform/pkg/microservice"
	"strconv"
	"testing"
	"time"

	commonModel "smlcloudplatform/internal/models"

	"github.com/stretchr/testify/assert"
)

func newRealDBRepository(t *testing.T) (repositories.IProductBarcodePGRepository, *models.ProductBarcodePg) {
	t.Helper()
	if os.Getenv("BC_REAL_DB_TESTS") != "1" {
		t.Skip("set BC_REAL_DB_TESTS=1 to run PostgreSQL integration tests")
	}

	t.Setenv("MODE", "test")
	codeTh := "th"
	itemNameThai := "ทดสอบ"
	codeEn := "en"
	itemNameEng := "test"
	barcode := &models.ProductBarcodePg{
		HoldingCode: "shoptester",
		ItemCode:    "ITEM001",
		Barcode:     "1234567890",
		PartitionIdentity: commonModel.PartitionIdentity{
			ParID: "partitiontester",
		},
		Names: []commonModel.NameX{
			{
				Code: &codeTh,
				Name: &itemNameThai,
			},
			{
				Code: &codeEn,
				Name: &itemNameEng,
			},
		},
	}

	cfg := config.NewConfig()
	repo := microservice.NewPersister(cfg.PersisterConfig())
	return repositories.NewProductBarcodePGRepository(repo), barcode
}

func TestCreateProductBarcodeInRealDB(t *testing.T) {
	productBarcodeRepository, barcode := newRealDBRepository(t)

	err := productBarcodeRepository.Create(barcode)
	assert.NoError(t, err)
}

func TestGetBarcode(t *testing.T) {
	productBarcodeRepository, barcode := newRealDBRepository(t)

	bar, err := productBarcodeRepository.Get(barcode.HoldingCode, barcode.Barcode)
	assert.NoError(t, err)

	assert.Equal(t, barcode.HoldingCode, bar.HoldingCode)
}

func TestUpdateBarcode(t *testing.T) {
	productBarcodeRepository, barcode := newRealDBRepository(t)

	currentTime := time.Now()
	timeStr := currentTime.Format("20060201150405")
	barcode.BalanceQty, _ = strconv.ParseFloat(timeStr, 64)

	err := productBarcodeRepository.Update(barcode.HoldingCode, barcode.ItemCode, barcode.Barcode, barcode)
	assert.NoError(t, err)

	bar, err := productBarcodeRepository.Get(barcode.HoldingCode, barcode.Barcode)
	assert.NoError(t, err)

	assert.Equal(t, barcode.BalanceQty, bar.BalanceQty)
}

func TestGetBarcodeAssertNotFoundBarcode(t *testing.T) {
	productBarcodeRepository, _ := newRealDBRepository(t)

	doc, err := productBarcodeRepository.Get("999", "999")
	assert.NoError(t, err)

	assert.Nil(t, doc)
}

func TestDeleteProductBarcodeInRealDB(t *testing.T) {
	productBarcodeRepository, barcode := newRealDBRepository(t)

	err := productBarcodeRepository.Delete(barcode.HoldingCode, barcode.ItemCode, barcode.Barcode)
	assert.NoError(t, err)
}
