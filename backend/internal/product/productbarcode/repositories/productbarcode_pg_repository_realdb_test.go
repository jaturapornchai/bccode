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

var barcode = &models.ProductBarcodePg{}
var productBarcodeRepository repositories.IProductBarcodePGRepository

func init() {

	os.Setenv("MODE", "test")
	cfg := config.NewConfig()

	repo := microservice.NewPersister(cfg.PersisterConfig())

	productBarcodeRepository = repositories.NewProductBarcodePGRepository(repo)

	codeTh := "th"
	itemNameThai := "ทดสอบ"
	codeEn := "en"
	itemNameEng := "test"
	barcode = &models.ProductBarcodePg{
		ShopID:  "shoptester",
		Barcode: "1234567890",
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
}

func TestCreateProductBarcodeInRealDB(t *testing.T) {

	if os.Getenv("SERVERLESS") == "serverless" {
		t.Skip()
	}

	err := productBarcodeRepository.Create(barcode)
	assert.NoError(t, err)
}

func TestGetBarcode(t *testing.T) {

	if os.Getenv("SERVERLESS") == "serverless" {
		t.Skip()
	}

	bar, err := productBarcodeRepository.Get(barcode.ShopID, barcode.Barcode)
	assert.NoError(t, err)

	assert.Equal(t, barcode.ShopID, bar.ShopID)
}

func TestUpdateBarcode(t *testing.T) {

	if os.Getenv("SERVERLESS") == "serverless" {
		t.Skip()
	}

	currentTime := time.Now()
	timeStr := currentTime.Format("20060201150405")
	barcode.BalanceQty, _ = strconv.ParseFloat(timeStr, 64)

	err := productBarcodeRepository.Update(barcode.ShopID, barcode.Barcode, barcode)
	assert.NoError(t, err)

	bar, err := productBarcodeRepository.Get(barcode.ShopID, barcode.Barcode)
	assert.NoError(t, err)

	assert.Equal(t, barcode.BalanceQty, bar.BalanceQty)
}

func TestGetBarcodeAssertNotFoundBarcode(t *testing.T) {

	if os.Getenv("SERVERLESS") == "serverless" {
		t.Skip()
	}

	doc, err := productBarcodeRepository.Get("999", "999")
	assert.NoError(t, err)

	assert.Nil(t, doc)
}

func TestDeleteProductBarcodeInRealDB(t *testing.T) {

	if os.Getenv("SERVERLESS") == "serverless" {
		t.Skip()
	}

	err := productBarcodeRepository.Delete(barcode.ShopID, barcode.Barcode)
	assert.NoError(t, err)
}
