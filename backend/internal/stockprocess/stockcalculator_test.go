package stockprocess_test

import (
	"fmt"
	"os"
	"smlcloudplatform/internal/config"
	"smlcloudplatform/internal/logger"
	productBarcodeModel "smlcloudplatform/internal/product/productbarcode/models"
	"smlcloudplatform/internal/stockprocess"
	stockModel "smlcloudplatform/internal/stockprocess/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockStockProcessPGRepository struct {
	mock.Mock
}

func (m *MockStockProcessPGRepository) GetStockTransactionList(shopID string, barcode string) ([]stockModel.StockData, error) {
	ret := m.Called(shopID, barcode)
	return ret.Get(0).([]stockModel.StockData), ret.Error(1)
}

func (m *MockStockProcessPGRepository) UpdateStockTransactionChange(stockData []stockModel.StockData) error {
	ret := m.Called(stockData)
	return ret.Error(0)
}

func (m *MockStockProcessPGRepository) ExecuteUpdateProductBarcodeStockBalance(shopID string, barcode string) error {
	ret := m.Called(shopID, barcode)
	return ret.Error(0)
}

type MockProductBarcodePGRepository struct {
	mock.Mock
}

func (m *MockProductBarcodePGRepository) FindByBarcode(shopID string, barcode string) (*productBarcodeModel.ProductBarcodePg, error) {
	ret := m.Called(shopID, barcode)
	return ret.Get(0).(*productBarcodeModel.ProductBarcodePg), ret.Error(1)
}

func (m *MockProductBarcodePGRepository) FindByBarcodes(shopID string, barcodes []string) ([]productBarcodeModel.ProductBarcodePg, error) {
	ret := m.Called(shopID, barcodes)
	return ret.Get(0).([]productBarcodeModel.ProductBarcodePg), ret.Error(1)
}

func (m *MockProductBarcodePGRepository) Get(shopID string, barcode string) (*productBarcodeModel.ProductBarcodePg, error) {
	ret := m.Called(shopID, barcode)
	return ret.Get(0).(*productBarcodeModel.ProductBarcodePg), ret.Error(1)
}

func (m *MockProductBarcodePGRepository) Create(doc *productBarcodeModel.ProductBarcodePg) error {
	ret := m.Called(doc)
	return ret.Error(0)
}

func (m *MockProductBarcodePGRepository) Update(shopID string, barcode string, doc *productBarcodeModel.ProductBarcodePg) error {
	ret := m.Called(shopID, barcode, doc)
	return ret.Error(0)
}

func (m *MockProductBarcodePGRepository) Delete(shopID string, barcode string) error {
	ret := m.Called(shopID, barcode)
	return ret.Error(0)
}

// var repo repositories.IStockProcessPGRepository

// func init() {
// 	cfg := config.NewConfig()
// 	persister := microservice.NewPersister(cfg.PersisterConfig())
// 	repo = repositories.NewStockProcessPGRepository(persister)
// }

func TestStockProcess(t *testing.T) {

	FIX_SHOPID := "SHOPID"
	FIX_BARCODE := "BARCODE"
	// stockLists, err := repo.GetStockTransactionList("2IZS0jFeRXWPidSupyXN7zQIlaS", "888555")
	// assert.Nil(t, err)
	// assert.NotNil(t, stockLists)
	// assert.Equal(t, 2, len(stockLists))

	var stockDataLists []stockModel.StockData
	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:              FIX_SHOPID,
		Barcode:             FIX_BARCODE,
		DocNo:               "PU25030001",
		TransFlag:           12,
		CalcFlag:            1,
		CalcQty:             300,
		SumAmount:           2000,
		SumAmountExcludeVat: 1915.89,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:              FIX_SHOPID,
		Barcode:             FIX_BARCODE,
		DocNo:               "INV25030001",
		TransFlag:           44,
		CalcFlag:            -1,
		CalcQty:             3,
		SumAmount:           90,
		SumAmountExcludeVat: 81.77,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:              FIX_SHOPID,
		Barcode:             FIX_BARCODE,
		DocNo:               "INV25030002",
		TransFlag:           44,
		CalcFlag:            -1,
		CalcQty:             5,
		SumAmount:           90,
		SumAmountExcludeVat: 81.77,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:              FIX_SHOPID,
		Barcode:             FIX_BARCODE,
		DocNo:               "FG25030002",
		TransFlag:           60,
		CalcFlag:            1,
		CalcQty:             5,
		SumAmount:           0,
		SumAmountExcludeVat: 0,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:              FIX_SHOPID,
		Barcode:             FIX_BARCODE,
		DocNo:               "PU25030002",
		TransFlag:           12,
		CalcFlag:            1,
		CalcQty:             90,
		SumAmount:           650,
		SumAmountExcludeVat: 609.00,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:              FIX_SHOPID,
		Barcode:             FIX_BARCODE,
		DocNo:               "PT25030001",
		DocRef:              "PU25030001",
		TransFlag:           16,
		CalcFlag:            -1,
		CalcQty:             60,
		SumAmount:           650,
		SumAmountExcludeVat: 609.00,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:              FIX_SHOPID,
		Barcode:             FIX_BARCODE,
		DocNo:               "INV25030003",
		TransFlag:           44,
		CalcFlag:            -1,
		CalcQty:             60,
		SumAmount:           650,
		SumAmountExcludeVat: 609.00,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:     FIX_SHOPID,
		Barcode:    FIX_BARCODE,
		DocNo:      "TF25030001",
		TransFlag:  72,
		CalcFlag:   -1,
		CalcQty:    60,
		LineNumber: 0,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:     FIX_SHOPID,
		Barcode:    FIX_BARCODE,
		DocNo:      "TF25030001",
		TransFlag:  72,
		CalcFlag:   -1,
		CalcQty:    30,
		LineNumber: 1,
	})
	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:     FIX_SHOPID,
		Barcode:    FIX_BARCODE,
		DocNo:      "TF25030001",
		TransFlag:  72,
		CalcFlag:   -1,
		CalcQty:    30,
		LineNumber: 2,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:     FIX_SHOPID,
		Barcode:    FIX_BARCODE,
		DocNo:      "TF25030001",
		TransFlag:  72,
		CalcFlag:   -1,
		CalcQty:    30,
		LineNumber: 3,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:     FIX_SHOPID,
		Barcode:    FIX_BARCODE,
		DocNo:      "TF25030001",
		TransFlag:  72,
		CalcFlag:   -1,
		CalcQty:    30,
		LineNumber: 4,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:     FIX_SHOPID,
		Barcode:    FIX_BARCODE,
		DocNo:      "TF25030001",
		TransFlag:  72,
		CalcFlag:   -1,
		CalcQty:    30,
		LineNumber: 5,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:     FIX_SHOPID,
		Barcode:    FIX_BARCODE,
		DocNo:      "TF25030001",
		TransFlag:  72,
		CalcFlag:   1,
		CalcQty:    60,
		LineNumber: 0,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:     FIX_SHOPID,
		Barcode:    FIX_BARCODE,
		DocNo:      "TF25030001",
		TransFlag:  72,
		CalcFlag:   1,
		CalcQty:    30,
		LineNumber: 1,
	})
	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:     FIX_SHOPID,
		Barcode:    FIX_BARCODE,
		DocNo:      "TF25030001",
		TransFlag:  72,
		CalcFlag:   1,
		CalcQty:    30,
		LineNumber: 2,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:     FIX_SHOPID,
		Barcode:    FIX_BARCODE,
		DocNo:      "TF25030001",
		TransFlag:  72,
		CalcFlag:   1,
		CalcQty:    30,
		LineNumber: 3,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:     FIX_SHOPID,
		Barcode:    FIX_BARCODE,
		DocNo:      "TF25030001",
		TransFlag:  72,
		CalcFlag:   1,
		CalcQty:    30,
		LineNumber: 4,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:     FIX_SHOPID,
		Barcode:    FIX_BARCODE,
		DocNo:      "TF25030001",
		TransFlag:  72,
		CalcFlag:   1,
		CalcQty:    30,
		LineNumber: 5,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:              FIX_SHOPID,
		Barcode:             FIX_BARCODE,
		DocNo:               "INV25030004",
		TransFlag:           44,
		CalcFlag:            -1,
		CalcQty:             2,
		SumAmount:           12.78,
		SumAmountExcludeVat: 12.78,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:              FIX_SHOPID,
		Barcode:             FIX_BARCODE,
		DocNo:               "INV25030005",
		TransFlag:           44,
		CalcFlag:            -1,
		CalcQty:             3,
		SumAmount:           19.17,
		SumAmountExcludeVat: 19.17,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:              FIX_SHOPID,
		Barcode:             FIX_BARCODE,
		DocNo:               "INV25030006",
		TransFlag:           44,
		CalcFlag:            -1,
		CalcQty:             25,
		SumAmount:           159.75,
		SumAmountExcludeVat: 159.75,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:              FIX_SHOPID,
		Barcode:             FIX_BARCODE,
		DocNo:               "WID25030001",
		TransFlag:           56,
		CalcFlag:            -1,
		CalcQty:             15,
		SumAmount:           95.85,
		SumAmountExcludeVat: 95.85,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:              FIX_SHOPID,
		Barcode:             FIX_BARCODE,
		DocNo:               "FG25030001",
		TransFlag:           60,
		CalcFlag:            1,
		CalcQty:             30,
		SumAmount:           199,
		SumAmountExcludeVat: 199.00,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:              FIX_SHOPID,
		Barcode:             FIX_BARCODE,
		DocNo:               "PT25030002",
		TransFlag:           16,
		CalcFlag:            -1,
		CalcQty:             30,
		SumAmount:           650,
		SumAmountExcludeVat: 609.00,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:              FIX_SHOPID,
		Barcode:             FIX_BARCODE,
		DocNo:               "INV25030007",
		TransFlag:           44,
		CalcFlag:            -1,
		CalcQty:             5,
		SumAmount:           159.75,
		SumAmountExcludeVat: 159.75,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:              FIX_SHOPID,
		Barcode:             FIX_BARCODE,
		DocNo:               "INV25030008",
		TransFlag:           44,
		CalcFlag:            -1,
		CalcQty:             5,
		SumAmount:           159.75,
		SumAmountExcludeVat: 159.75,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:              FIX_SHOPID,
		Barcode:             FIX_BARCODE,
		DocNo:               "INV25030009",
		TransFlag:           44,
		CalcFlag:            -1,
		CalcQty:             5,
		SumAmount:           159.75,
		SumAmountExcludeVat: 159.75,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:              FIX_SHOPID,
		Barcode:             FIX_BARCODE,
		DocNo:               "INV250300011",
		TransFlag:           44,
		CalcFlag:            -1,
		CalcQty:             1,
		SumAmount:           159.75,
		SumAmountExcludeVat: 159.75,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:              FIX_SHOPID,
		Barcode:             FIX_BARCODE,
		DocNo:               "INV25030012",
		TransFlag:           44,
		CalcFlag:            -1,
		CalcQty:             60,
		SumAmount:           159.75,
		SumAmountExcludeVat: 159.75,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:              FIX_SHOPID,
		Barcode:             FIX_BARCODE,
		DocNo:               "WID25030002",
		TransFlag:           56,
		CalcFlag:            -1,
		CalcQty:             15,
		SumAmount:           95.85,
		SumAmountExcludeVat: 95.85,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:              FIX_SHOPID,
		Barcode:             FIX_BARCODE,
		DocNo:               "IS25030001",
		TransFlag:           68,
		CalcFlag:            -1,
		CalcQty:             2,
		SumAmount:           2.85,
		SumAmountExcludeVat: 2.85,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:              FIX_SHOPID,
		Barcode:             FIX_BARCODE,
		DocNo:               "IA25030001",
		TransFlag:           866,
		CalcFlag:            1,
		CalcQty:             0,
		SumAmount:           3.59,
		SumAmountExcludeVat: 3.59,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:              FIX_SHOPID,
		Barcode:             FIX_BARCODE,
		DocNo:               "TF25030002",
		TransFlag:           72,
		CalcFlag:            -1,
		CalcQty:             30,
		SumAmount:           95.85,
		SumAmountExcludeVat: 95.85,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:              FIX_SHOPID,
		Barcode:             FIX_BARCODE,
		DocNo:               "TF25030002",
		TransFlag:           72,
		CalcFlag:            1,
		CalcQty:             30,
		SumAmount:           0,
		SumAmountExcludeVat: 0,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:              FIX_SHOPID,
		Barcode:             FIX_BARCODE,
		DocNo:               "RIM25030001",
		DocRef:              "WID25030002",
		TransFlag:           58,
		CalcFlag:            1,
		CalcQty:             5,
		SumAmount:           95.85,
		SumAmountExcludeVat: 95.85,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:              FIX_SHOPID,
		Barcode:             FIX_BARCODE,
		DocNo:               "INV25030013",
		TransFlag:           44,
		CalcFlag:            -1,
		CalcQty:             6,
		SumAmount:           38.7,
		SumAmountExcludeVat: 38.7,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:              FIX_SHOPID,
		Barcode:             FIX_BARCODE,
		DocNo:               "CN25030001",
		DocRef:              "INV25030013",
		TransFlag:           48,
		CalcFlag:            1,
		CalcQty:             1,
		SumAmount:           6.45,
		SumAmountExcludeVat: 6.45,
	})

	// stockDataLists = append(stockDataLists, stockModel.StockData{
	// 	ShopID:              FIX_SHOPID,
	// 	Barcode:             FIX_BARCODE,
	// 	DocNo:               "DOC2",
	// 	TransFlag:           12,
	// 	CalcFlag:            1,
	// 	CalcQty:             2,
	// 	StandValue:          1,
	// 	DivideValue:         1,
	// 	SumAmount:           22,
	// 	SumAmountExcludeVat: 22,
	// })

	// stockDataLists = append(stockDataLists, stockModel.StockData{
	// 	ShopID:              FIX_SHOPID,
	// 	Barcode:             FIX_BARCODE,
	// 	DocNo:               "DOC2",
	// 	TransFlag:           12,
	// 	CalcFlag:            1,
	// 	CalcQty:             3,
	// 	StandValue:          1,
	// 	DivideValue:         1,
	// 	SumAmount:           30,
	// 	SumAmountExcludeVat: 30,
	// })

	os.Setenv("LOG_LEVEL", "debug")

	logger.NewAppLogger(config.NewLoggerConfig())

	giveBarcode := &productBarcodeModel.ProductBarcodePg{
		ShopID:     FIX_SHOPID,
		Barcode:    FIX_BARCODE,
		BalanceQty: 297,
	}

	repo := new(MockStockProcessPGRepository)
	repo.On("GetStockTransactionList", FIX_SHOPID, FIX_BARCODE).Return(stockDataLists, nil)
	repo.On("UpdateStockTransactionChange", mock.Anything).Return(nil)

	barcodeRepo := new(MockProductBarcodePGRepository)
	barcodeRepo.On("Get", FIX_SHOPID, FIX_BARCODE).Return(giveBarcode, nil)
	barcodeRepo.On("Update", FIX_SHOPID, FIX_BARCODE, giveBarcode).Return(nil)

	process := stockprocess.NewStockCalculator(repo, barcodeRepo)
	process.CalculatorStock(FIX_SHOPID, FIX_BARCODE)

	assert.Equal(t, stockDataLists[0].CostPerUnit, float64(6.39), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[0].DocNo))
	assert.Equal(t, stockDataLists[0].TotalCost, float64(1915.89), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[0].DocNo))

	assert.Equal(t, stockDataLists[1].CostPerUnit, float64(6.39), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[1].DocNo))
	assert.Equal(t, stockDataLists[1].TotalCost, float64(19.17), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[1].DocNo))

	assert.Equal(t, stockDataLists[2].CostPerUnit, float64(6.39), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[2].DocNo))
	assert.Equal(t, stockDataLists[2].TotalCost, float64(31.95), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[2].DocNo))

	assert.Equal(t, stockDataLists[3].CostPerUnit, float64(0), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[3].DocNo))
	assert.Equal(t, stockDataLists[3].TotalCost, float64(0), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[3].DocNo))

	assert.Equal(t, stockDataLists[4].CostPerUnit, float64(6.77), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[4].DocNo))
	assert.Equal(t, stockDataLists[4].TotalCost, float64(609), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[4].DocNo))

	assert.Equal(t, stockDataLists[5].CostPerUnit, float64(6.39), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[5].DocNo))
	assert.Equal(t, stockDataLists[5].TotalCost, float64(383.4), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[5].DocNo))

	assert.Equal(t, stockDataLists[6].CostPerUnit, float64(6.39), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[6].DocNo))
	assert.Equal(t, stockDataLists[6].TotalCost, float64(383.4), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[6].DocNo))

	assert.Equal(t, stockDataLists[7].CostPerUnit, float64(6.39), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[7].DocNo))
	assert.Equal(t, stockDataLists[7].TotalCost, float64(383.4), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[7].DocNo))

	assert.Equal(t, stockDataLists[8].CostPerUnit, float64(6.39), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[8].DocNo))
	assert.Equal(t, stockDataLists[8].TotalCost, float64(191.7), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[8].DocNo))

	assert.Equal(t, stockDataLists[9].CostPerUnit, float64(6.39), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[9].DocNo))
	assert.Equal(t, stockDataLists[9].TotalCost, float64(191.7), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[9].DocNo))

	assert.Equal(t, stockDataLists[10].CostPerUnit, float64(6.4), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[10].DocNo))
	assert.Equal(t, stockDataLists[10].TotalCost, float64(192), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[10].DocNo))

	assert.Equal(t, stockDataLists[11].CostPerUnit, float64(6.39), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[11].DocNo))
	assert.Equal(t, stockDataLists[11].TotalCost, float64(191.7), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[11].DocNo))

	assert.Equal(t, stockDataLists[12].CostPerUnit, float64(6.4), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[12].DocNo))
	assert.Equal(t, stockDataLists[12].TotalCost, float64(192), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[12].DocNo))

	assert.Equal(t, stockDataLists[13].CostPerUnit, float64(6.39), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[13].DocNo))
	assert.Equal(t, stockDataLists[13].TotalCost, float64(383.4), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[13].DocNo))

	assert.Equal(t, stockDataLists[14].CostPerUnit, float64(6.39), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[14].DocNo))
	assert.Equal(t, stockDataLists[14].TotalCost, float64(191.7), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[14].DocNo))

	assert.Equal(t, stockDataLists[15].CostPerUnit, float64(6.39), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[15].DocNo))
	assert.Equal(t, stockDataLists[15].TotalCost, float64(191.7), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[15].DocNo))

	assert.Equal(t, stockDataLists[16].CostPerUnit, float64(6.4), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[16].DocNo))
	assert.Equal(t, stockDataLists[16].TotalCost, float64(192), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[16].DocNo))

	assert.Equal(t, stockDataLists[17].CostPerUnit, float64(6.39), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[17].DocNo))
	assert.Equal(t, stockDataLists[17].TotalCost, float64(191.7), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[17].DocNo))

	assert.Equal(t, stockDataLists[18].CostPerUnit, float64(6.4), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[18].DocNo))
	assert.Equal(t, stockDataLists[18].TotalCost, float64(192), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[18].DocNo))

	assert.Equal(t, stockDataLists[19].CostPerUnit, float64(6.39), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[19].DocNo))
	assert.Equal(t, stockDataLists[19].TotalCost, float64(12.78), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[19].DocNo))

	assert.Equal(t, stockDataLists[20].CostPerUnit, float64(6.39), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[20].DocNo))
	assert.Equal(t, stockDataLists[20].TotalCost, float64(19.17), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[20].DocNo))

	assert.Equal(t, stockDataLists[21].CostPerUnit, float64(6.39), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[21].DocNo))
	assert.Equal(t, stockDataLists[21].TotalCost, float64(159.75), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[21].DocNo))

	assert.Equal(t, stockDataLists[22].CostPerUnit, float64(6.39), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[22].DocNo))
	assert.Equal(t, stockDataLists[22].TotalCost, float64(95.85), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[22].DocNo))

	assert.Equal(t, stockDataLists[23].CostPerUnit, float64(6.63), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[23].DocNo))
	assert.Equal(t, stockDataLists[23].TotalCost, float64(199), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[23].DocNo))

	assert.Equal(t, stockDataLists[24].CostPerUnit, float64(6.42), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[24].DocNo))
	assert.Equal(t, stockDataLists[24].TotalCost, float64(192.6), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[24].DocNo))

	assert.Equal(t, stockDataLists[25].CostPerUnit, float64(6.42), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[25].DocNo))
	assert.Equal(t, stockDataLists[25].TotalCost, float64(32.1), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[25].DocNo))

	assert.Equal(t, stockDataLists[26].CostPerUnit, float64(6.42), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[26].DocNo))
	assert.Equal(t, stockDataLists[26].TotalCost, float64(32.1), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[26].DocNo))

	assert.Equal(t, stockDataLists[27].CostPerUnit, float64(6.42), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[27].DocNo))
	assert.Equal(t, stockDataLists[27].TotalCost, float64(32.1), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[27].DocNo))

	assert.Equal(t, stockDataLists[28].CostPerUnit, float64(6.42), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[28].DocNo))
	assert.Equal(t, stockDataLists[28].TotalCost, float64(6.42), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[28].DocNo))

	assert.Equal(t, stockDataLists[29].CostPerUnit, float64(6.42), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[29].DocNo))
	assert.Equal(t, stockDataLists[29].TotalCost, float64(385.2), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[29].DocNo))

	assert.Equal(t, stockDataLists[30].CostPerUnit, float64(6.42), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[30].DocNo))
	assert.Equal(t, stockDataLists[30].TotalCost, float64(96.3), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[30].DocNo))

	assert.Equal(t, stockDataLists[31].CostPerUnit, float64(6.42), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[31].DocNo))
	assert.Equal(t, stockDataLists[31].TotalCost, float64(12.84), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[31].DocNo))

	assert.Equal(t, stockDataLists[32].CostPerUnit, float64(0), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[32].DocNo))
	assert.Equal(t, stockDataLists[32].TotalCost, float64(3.59), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[32].DocNo))

	assert.Equal(t, stockDataLists[33].CostPerUnit, float64(6.45), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[33].DocNo))
	assert.Equal(t, stockDataLists[33].TotalCost, float64(193.50), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[33].DocNo))

	assert.Equal(t, stockDataLists[34].CostPerUnit, float64(6.45), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[34].DocNo))
	assert.Equal(t, stockDataLists[34].TotalCost, float64(193.50), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[34].DocNo))

	assert.Equal(t, stockDataLists[35].CostPerUnit, float64(6.42), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[35].DocNo))
	assert.Equal(t, stockDataLists[35].TotalCost, float64(32.1), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[35].DocNo))

	assert.Equal(t, stockDataLists[36].CostPerUnit, float64(6.45), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[36].DocNo))
	assert.Equal(t, stockDataLists[36].TotalCost, float64(38.7), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[36].DocNo))

	assert.Equal(t, stockDataLists[37].CostPerUnit, float64(6.45), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[37].DocNo))
	assert.Equal(t, stockDataLists[37].TotalCost, float64(6.45), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[37].DocNo))

}

func TestProcessStockResultNAN(t *testing.T) {

	FIX_SHOPID := "SHOPID"
	FIX_BARCODE := "01-150"
	// stockLists, err := repo.GetStockTransactionList("2IZS0jFeRXWPidSupyXN7zQIlaS", "888555")
	// assert.Nil(t, err)
	// assert.NotNil(t, stockLists)
	// assert.Equal(t, 2, len(stockLists))

	var stockDataLists []stockModel.StockData
	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:              FIX_SHOPID,
		Barcode:             FIX_BARCODE,
		DocNo:               "TF2025020800004",
		TransFlag:           72,
		CalcFlag:            -1,
		CalcQty:             500,
		SumAmount:           0,
		SumAmountExcludeVat: 0,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:              FIX_SHOPID,
		Barcode:             FIX_BARCODE,
		DocNo:               "TF2025020800004",
		TransFlag:           72,
		CalcFlag:            1,
		CalcQty:             500,
		SumAmount:           0,
		SumAmountExcludeVat: 0,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:              FIX_SHOPID,
		Barcode:             FIX_BARCODE,
		DocNo:               "TF2025021100001",
		TransFlag:           72,
		CalcFlag:            -1,
		CalcQty:             500,
		SumAmount:           0,
		SumAmountExcludeVat: 0,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:              FIX_SHOPID,
		Barcode:             FIX_BARCODE,
		DocNo:               "TF2025021100001",
		TransFlag:           72,
		CalcFlag:            1,
		CalcQty:             500,
		SumAmount:           0,
		SumAmountExcludeVat: 0,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:              FIX_SHOPID,
		Barcode:             FIX_BARCODE,
		DocNo:               "TF2025021200003",
		TransFlag:           72,
		CalcFlag:            -1,
		CalcQty:             500,
		SumAmount:           0,
		SumAmountExcludeVat: 0,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:              FIX_SHOPID,
		Barcode:             FIX_BARCODE,
		DocNo:               "TF2025021200003",
		TransFlag:           72,
		CalcFlag:            1,
		CalcQty:             500,
		SumAmount:           0,
		SumAmountExcludeVat: 0,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:              FIX_SHOPID,
		Barcode:             FIX_BARCODE,
		DocNo:               "IF2025021700002",
		TransFlag:           60,
		CalcFlag:            1,
		CalcQty:             1000,
		SumAmount:           0,
		SumAmountExcludeVat: 0,
	})

	os.Setenv("LOG_LEVEL", "debug")

	logger.NewAppLogger(config.NewLoggerConfig())

	giveBarcode := &productBarcodeModel.ProductBarcodePg{
		ShopID:     FIX_SHOPID,
		Barcode:    FIX_BARCODE,
		BalanceQty: 0,
	}

	repo := new(MockStockProcessPGRepository)
	repo.On("GetStockTransactionList", FIX_SHOPID, FIX_BARCODE).Return(stockDataLists, nil)
	repo.On("UpdateStockTransactionChange", mock.Anything).Return(nil)

	barcodeRepo := new(MockProductBarcodePGRepository)
	barcodeRepo.On("Get", FIX_SHOPID, FIX_BARCODE).Return(giveBarcode, nil)
	barcodeRepo.On("Update", FIX_SHOPID, FIX_BARCODE, giveBarcode).Return(nil)

	process := stockprocess.NewStockCalculator(repo, barcodeRepo)
	process.CalculatorStock(FIX_SHOPID, FIX_BARCODE)

	assert.Equal(t, stockDataLists[0].CostPerUnit, float64(0), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[0].DocNo))
	assert.Equal(t, stockDataLists[0].TotalCost, float64(0), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[0].DocNo))
	assert.Equal(t, stockDataLists[0].BalanceQty, float64(-500), fmt.Sprintf("Balance Qty DocNo: %s", stockDataLists[0].DocNo))
	assert.Equal(t, stockDataLists[0].BalanceAverage, float64(0), fmt.Sprintf("Balance Average DocNo: %s", stockDataLists[0].DocNo))
	assert.Equal(t, stockDataLists[0].BalanceAmount, float64(0), fmt.Sprintf("Balance Amount DocNo: %s", stockDataLists[0].DocNo))

	assert.Equal(t, stockDataLists[1].CostPerUnit, float64(0), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[1].DocNo))
	assert.Equal(t, stockDataLists[1].TotalCost, float64(0), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[1].DocNo))
	assert.Equal(t, stockDataLists[1].BalanceQty, float64(0), fmt.Sprintf("Balance Qty DocNo: %s", stockDataLists[1].DocNo))
	assert.Equal(t, stockDataLists[1].BalanceAverage, float64(0), fmt.Sprintf("Balance Average DocNo: %s", stockDataLists[1].DocNo))
	assert.Equal(t, stockDataLists[1].BalanceAmount, float64(0), fmt.Sprintf("Balance Amount DocNo: %s", stockDataLists[1].DocNo))

	assert.Equal(t, stockDataLists[2].CostPerUnit, float64(0), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[2].DocNo))
	assert.Equal(t, stockDataLists[2].TotalCost, float64(0), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[2].DocNo))
	assert.Equal(t, stockDataLists[2].BalanceQty, float64(-500), fmt.Sprintf("Balance Qty DocNo: %s", stockDataLists[2].DocNo))
	assert.Equal(t, stockDataLists[2].BalanceAverage, float64(0), fmt.Sprintf("Balance Average DocNo: %s", stockDataLists[2].DocNo))
	assert.Equal(t, stockDataLists[2].BalanceAmount, float64(0), fmt.Sprintf("Balance Amount DocNo: %s", stockDataLists[2].DocNo))

	assert.Equal(t, stockDataLists[3].CostPerUnit, float64(0), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[3].DocNo))
	assert.Equal(t, stockDataLists[3].TotalCost, float64(0), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[3].DocNo))
	assert.Equal(t, stockDataLists[3].BalanceQty, float64(0), fmt.Sprintf("Balance Qty DocNo: %s", stockDataLists[3].DocNo))
	assert.Equal(t, stockDataLists[3].BalanceAverage, float64(0), fmt.Sprintf("Balance Average DocNo: %s", stockDataLists[3].DocNo))
	assert.Equal(t, stockDataLists[3].BalanceAmount, float64(0), fmt.Sprintf("Balance Amount DocNo: %s", stockDataLists[3].DocNo))

	assert.Equal(t, stockDataLists[4].CostPerUnit, float64(0), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[4].DocNo))
	assert.Equal(t, stockDataLists[4].TotalCost, float64(0), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[4].DocNo))
	assert.Equal(t, stockDataLists[4].BalanceQty, float64(-500), fmt.Sprintf("Balance Qty DocNo: %s", stockDataLists[4].DocNo))
	assert.Equal(t, stockDataLists[4].BalanceAverage, float64(0), fmt.Sprintf("Balance Average DocNo: %s", stockDataLists[4].DocNo))
	assert.Equal(t, stockDataLists[4].BalanceAmount, float64(0), fmt.Sprintf("Balance Amount DocNo: %s", stockDataLists[4].DocNo))

	assert.Equal(t, stockDataLists[5].CostPerUnit, float64(0), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[5].DocNo))
	assert.Equal(t, stockDataLists[5].TotalCost, float64(0), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[5].DocNo))
	assert.Equal(t, stockDataLists[5].BalanceQty, float64(0), fmt.Sprintf("Balance Qty DocNo: %s", stockDataLists[5].DocNo))
	assert.Equal(t, stockDataLists[5].BalanceAverage, float64(0), fmt.Sprintf("Balance Average DocNo: %s", stockDataLists[5].DocNo))
	assert.Equal(t, stockDataLists[5].BalanceAmount, float64(0), fmt.Sprintf("Balance Amount DocNo: %s", stockDataLists[5].DocNo))

	assert.Equal(t, stockDataLists[6].CostPerUnit, float64(0), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[6].DocNo))
	assert.Equal(t, stockDataLists[6].TotalCost, float64(0), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[6].DocNo))
	assert.Equal(t, stockDataLists[6].BalanceQty, float64(1000), fmt.Sprintf("Balance Qty DocNo: %s", stockDataLists[6].DocNo))
	assert.Equal(t, stockDataLists[6].BalanceAverage, float64(0), fmt.Sprintf("Balance Average DocNo: %s", stockDataLists[6].DocNo))
	assert.Equal(t, stockDataLists[6].BalanceAmount, float64(0), fmt.Sprintf("Balance Amount DocNo: %s", stockDataLists[6].DocNo))
}

func TestProcessStockBalanceAmountInfinity(t *testing.T) {
	FIX_SHOPID := "SHOPID"
	FIX_BARCODE := "885001"

	var stockDataLists []stockModel.StockData
	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:              FIX_SHOPID,
		Barcode:             FIX_BARCODE,
		DocNo:               "PU00",
		TransFlag:           12,
		CalcFlag:            1,
		CalcQty:             1,
		SumAmount:           10,
		SumAmountExcludeVat: 10,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:              FIX_SHOPID,
		Barcode:             FIX_BARCODE,
		DocNo:               "PU01",
		TransFlag:           12,
		CalcFlag:            1,
		CalcQty:             6,
		SumAmount:           30,
		SumAmountExcludeVat: 30,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:              FIX_SHOPID,
		Barcode:             FIX_BARCODE,
		DocNo:               "SALE02",
		TransFlag:           44,
		CalcFlag:            -1,
		CalcQty:             1,
		SumAmount:           16,
		SumAmountExcludeVat: 14.95,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:              FIX_SHOPID,
		Barcode:             FIX_BARCODE,
		DocNo:               "SALE03",
		TransFlag:           44,
		CalcFlag:            -1,
		CalcQty:             6,
		SumAmount:           35,
		SumAmountExcludeVat: 35,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:              FIX_SHOPID,
		Barcode:             FIX_BARCODE,
		DocNo:               "SALE04",
		TransFlag:           44,
		CalcFlag:            -1,
		CalcQty:             1,
		SumAmount:           16,
		SumAmountExcludeVat: 16,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:              FIX_SHOPID,
		Barcode:             FIX_BARCODE,
		DocNo:               "SALE05",
		TransFlag:           44,
		CalcFlag:            -1,
		CalcQty:             1,
		SumAmount:           16,
		SumAmountExcludeVat: 160,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:              FIX_SHOPID,
		Barcode:             FIX_BARCODE,
		DocNo:               "SALE06",
		TransFlag:           44,
		CalcFlag:            -1,
		CalcQty:             10,
		SumAmount:           16,
		SumAmountExcludeVat: 160,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:              FIX_SHOPID,
		Barcode:             FIX_BARCODE,
		DocNo:               "PU07",
		TransFlag:           12,
		CalcFlag:            1,
		CalcQty:             6,
		SumAmount:           30,
		SumAmountExcludeVat: 30,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:              FIX_SHOPID,
		Barcode:             FIX_BARCODE,
		DocNo:               "SALE08",
		TransFlag:           44,
		CalcFlag:            -1,
		CalcQty:             1,
		SumAmount:           30,
		SumAmountExcludeVat: 30,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:              FIX_SHOPID,
		Barcode:             FIX_BARCODE,
		DocNo:               "SALE09",
		TransFlag:           44,
		CalcFlag:            -1,
		CalcQty:             1,
		SumAmount:           30,
		SumAmountExcludeVat: 30,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:              FIX_SHOPID,
		Barcode:             FIX_BARCODE,
		DocNo:               "PU10",
		TransFlag:           12,
		CalcFlag:            1,
		CalcQty:             6,
		SumAmount:           30,
		SumAmountExcludeVat: 30,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:              FIX_SHOPID,
		Barcode:             FIX_BARCODE,
		DocNo:               "PU11",
		TransFlag:           12,
		CalcFlag:            1,
		CalcQty:             6,
		SumAmount:           30,
		SumAmountExcludeVat: 30,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:              FIX_SHOPID,
		Barcode:             FIX_BARCODE,
		DocNo:               "SALE12",
		TransFlag:           44,
		CalcFlag:            -1,
		CalcQty:             1,
		SumAmount:           30,
		SumAmountExcludeVat: 30,
	})

	os.Setenv("LOG_LEVEL", "debug")

	logger.NewAppLogger(config.NewLoggerConfig())

	giveBarcode := &productBarcodeModel.ProductBarcodePg{
		ShopID:     FIX_SHOPID,
		Barcode:    FIX_BARCODE,
		BalanceQty: 0,
	}

	repo := new(MockStockProcessPGRepository)
	repo.On("GetStockTransactionList", FIX_SHOPID, FIX_BARCODE).Return(stockDataLists, nil)
	repo.On("UpdateStockTransactionChange", mock.Anything).Return(nil)

	barcodeRepo := new(MockProductBarcodePGRepository)
	barcodeRepo.On("Get", FIX_SHOPID, FIX_BARCODE).Return(giveBarcode, nil)
	barcodeRepo.On("Update", FIX_SHOPID, FIX_BARCODE, giveBarcode).Return(nil)

	process := stockprocess.NewStockCalculator(repo, barcodeRepo)
	process.CalculatorStock(FIX_SHOPID, FIX_BARCODE)

	assert.Equal(t, stockDataLists[0].CalcQty, float64(1), fmt.Sprintf("Doc Qty of DocNo: %s", stockDataLists[0].DocNo))
	assert.Equal(t, stockDataLists[0].CostPerUnit, float64(10), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[0].DocNo))
	assert.Equal(t, stockDataLists[0].TotalCost, float64(10), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[0].DocNo))
	assert.Equal(t, stockDataLists[0].BalanceQty, float64(1), fmt.Sprintf("Balance Qty DocNo: %s", stockDataLists[0].DocNo))
	assert.Equal(t, stockDataLists[0].BalanceAverage, float64(10), fmt.Sprintf("Balance Average DocNo: %s", stockDataLists[0].DocNo))
	assert.Equal(t, stockDataLists[0].BalanceAmount, float64(10), fmt.Sprintf("Balance Amount DocNo: %s", stockDataLists[0].DocNo))

	assert.Equal(t, stockDataLists[1].CalcQty, float64(6), fmt.Sprintf("Doc Qty of DocNo: %s", stockDataLists[1].DocNo))
	assert.Equal(t, stockDataLists[1].BalanceQty, float64(7), fmt.Sprintf("Balance Qty DocNo: %s", stockDataLists[1].DocNo))
	assert.Equal(t, stockDataLists[1].BalanceAmount, float64(40), fmt.Sprintf("Balance Amount DocNo: %s", stockDataLists[1].DocNo))
	assert.Equal(t, stockDataLists[1].TotalCost, float64(30), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[1].DocNo))
	assert.Equal(t, stockDataLists[1].CostPerUnit, float64(5), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[1].DocNo))
	assert.Equal(t, stockDataLists[1].BalanceAverage, float64(5.71), fmt.Sprintf("Balance Average DocNo: %s", stockDataLists[1].DocNo))

	assert.Equal(t, stockDataLists[2].BalanceQty, float64(6), fmt.Sprintf("Balance Qty DocNo: %s", stockDataLists[2].DocNo))
	assert.Equal(t, stockDataLists[2].BalanceAmount, float64(34.29), fmt.Sprintf("Balance Amount DocNo: %s", stockDataLists[2].DocNo))
	assert.Equal(t, stockDataLists[2].CostPerUnit, float64(5.71), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[2].DocNo))
	assert.Equal(t, stockDataLists[2].TotalCost, float64(5.71), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[2].DocNo))
	assert.Equal(t, stockDataLists[2].BalanceAverage, float64(5.72), fmt.Sprintf("Balance Average DocNo: %s", stockDataLists[2].DocNo))

	assert.Equal(t, stockDataLists[3].CalcQty, float64(6), fmt.Sprintf("Doc Qty of DocNo: %s", stockDataLists[3].DocNo))
	assert.Equal(t, stockDataLists[3].BalanceQty, float64(0), fmt.Sprintf("Balance Qty DocNo: %s", stockDataLists[3].DocNo))
	assert.Equal(t, stockDataLists[3].BalanceAmount, float64(0), fmt.Sprintf("Balance Amount DocNo: %s", stockDataLists[3].DocNo))
	assert.Equal(t, stockDataLists[3].CostPerUnit, float64(5.72), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[3].DocNo))
	assert.Equal(t, stockDataLists[3].TotalCost, float64(34.29), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[3].DocNo))
	assert.Equal(t, stockDataLists[3].BalanceAverage, float64(0), fmt.Sprintf("Balance Average DocNo: %s", stockDataLists[3].DocNo))

	assert.Equal(t, stockDataLists[4].CalcQty, float64(1), fmt.Sprintf("Doc Qty of DocNo: %s", stockDataLists[4].DocNo))
	assert.Equal(t, stockDataLists[4].BalanceQty, float64(-1), fmt.Sprintf("Balance Qty DocNo: %s", stockDataLists[4].DocNo))
	assert.Equal(t, stockDataLists[4].BalanceAmount, float64(0), fmt.Sprintf("Balance Amount DocNo: %s", stockDataLists[4].DocNo))
	assert.Equal(t, stockDataLists[4].CostPerUnit, float64(0), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[4].DocNo))
	assert.Equal(t, stockDataLists[4].TotalCost, float64(0), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[4].DocNo))
	assert.Equal(t, stockDataLists[4].BalanceAverage, float64(0), fmt.Sprintf("Balance Average DocNo: %s", stockDataLists[4].DocNo))

	assert.Equal(t, stockDataLists[5].CalcQty, float64(1), fmt.Sprintf("Doc Qty of DocNo: %s", stockDataLists[5].DocNo))
	assert.Equal(t, stockDataLists[5].BalanceQty, float64(-2), fmt.Sprintf("Balance Qty DocNo: %s", stockDataLists[5].DocNo))
	assert.Equal(t, stockDataLists[5].BalanceAmount, float64(0), fmt.Sprintf("Balance Amount DocNo: %s", stockDataLists[5].DocNo))
	assert.Equal(t, stockDataLists[5].CostPerUnit, float64(0), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[5].DocNo))
	assert.Equal(t, stockDataLists[5].TotalCost, float64(0), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[5].DocNo))
	assert.Equal(t, stockDataLists[5].BalanceAverage, float64(0), fmt.Sprintf("Balance Average DocNo: %s", stockDataLists[5].DocNo))

	assert.Equal(t, stockDataLists[6].CalcQty, float64(10), fmt.Sprintf("Doc Qty of DocNo: %s", stockDataLists[6].DocNo))
	assert.Equal(t, stockDataLists[6].BalanceQty, float64(-12), fmt.Sprintf("Balance Qty DocNo: %s", stockDataLists[6].DocNo))
	assert.Equal(t, stockDataLists[6].BalanceAmount, float64(0), fmt.Sprintf("Balance Amount DocNo: %s", stockDataLists[6].DocNo))
	assert.Equal(t, stockDataLists[6].CostPerUnit, float64(0), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[6].DocNo))
	assert.Equal(t, stockDataLists[6].TotalCost, float64(0), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[6].DocNo))
	assert.Equal(t, stockDataLists[6].BalanceAverage, float64(0), fmt.Sprintf("Balance Average DocNo: %s", stockDataLists[6].DocNo))

	assert.Equal(t, stockDataLists[7].CalcQty, float64(6), fmt.Sprintf("Doc Qty of DocNo: %s", stockDataLists[7].DocNo))
	assert.Equal(t, stockDataLists[7].BalanceQty, float64(-6), fmt.Sprintf("Balance Qty DocNo: %s", stockDataLists[7].DocNo))
	assert.Equal(t, stockDataLists[7].BalanceAmount, float64(30), fmt.Sprintf("Balance Amount DocNo: %s", stockDataLists[7].DocNo))
	assert.Equal(t, stockDataLists[7].CostPerUnit, float64(5), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[7].DocNo))
	assert.Equal(t, stockDataLists[7].TotalCost, float64(30), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[7].DocNo))
	assert.Equal(t, stockDataLists[7].BalanceAverage, float64(5), fmt.Sprintf("Balance Average DocNo: %s", stockDataLists[7].DocNo))

	assert.Equal(t, stockDataLists[8].CalcQty, float64(1), fmt.Sprintf("Doc Qty of DocNo: %s", stockDataLists[8].DocNo))
	assert.Equal(t, stockDataLists[8].BalanceQty, float64(-7), fmt.Sprintf("Balance Qty DocNo: %s", stockDataLists[8].DocNo))
	assert.Equal(t, stockDataLists[8].BalanceAmount, float64(25), fmt.Sprintf("Balance Amount DocNo: %s", stockDataLists[8].DocNo))
	assert.Equal(t, stockDataLists[8].CostPerUnit, float64(5), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[8].DocNo))
	assert.Equal(t, stockDataLists[8].TotalCost, float64(5), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[8].DocNo))
	assert.Equal(t, stockDataLists[8].BalanceAverage, float64(3.57), fmt.Sprintf("Balance Average DocNo: %s", stockDataLists[8].DocNo))

	assert.Equal(t, stockDataLists[9].CalcQty, float64(1), fmt.Sprintf("Doc Qty of DocNo: %s", stockDataLists[9].DocNo))
	assert.Equal(t, stockDataLists[9].BalanceQty, float64(-8), fmt.Sprintf("Balance Qty DocNo: %s", stockDataLists[9].DocNo))
	assert.Equal(t, stockDataLists[9].BalanceAmount, float64(21.43), fmt.Sprintf("Balance Amount DocNo: %s", stockDataLists[9].DocNo))
	assert.Equal(t, stockDataLists[9].CostPerUnit, float64(3.57), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[9].DocNo))
	assert.Equal(t, stockDataLists[9].TotalCost, float64(3.57), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[9].DocNo))
	assert.Equal(t, stockDataLists[9].BalanceAverage, float64(2.68), fmt.Sprintf("Balance Average DocNo: %s", stockDataLists[9].DocNo))

	assert.Equal(t, stockDataLists[10].CalcQty, float64(6), fmt.Sprintf("Doc Qty of DocNo: %s", stockDataLists[10].DocNo))
	assert.Equal(t, stockDataLists[10].BalanceQty, float64(-2), fmt.Sprintf("Balance Qty DocNo: %s", stockDataLists[10].DocNo))
	assert.Equal(t, stockDataLists[10].BalanceAmount, float64(51.43), fmt.Sprintf("Balance Amount DocNo: %s", stockDataLists[10].DocNo))
	assert.Equal(t, stockDataLists[10].CostPerUnit, float64(5), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[10].DocNo))
	assert.Equal(t, stockDataLists[10].TotalCost, float64(30), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[10].DocNo))
	assert.Equal(t, stockDataLists[10].BalanceAverage, float64(25.72), fmt.Sprintf("Balance Average DocNo: %s", stockDataLists[10].DocNo))

	assert.Equal(t, stockDataLists[11].CalcQty, float64(6), fmt.Sprintf("Doc Qty of DocNo: %s", stockDataLists[11].DocNo))
	assert.Equal(t, stockDataLists[11].BalanceQty, float64(4), fmt.Sprintf("Balance Qty DocNo: %s", stockDataLists[11].DocNo))
	assert.Equal(t, stockDataLists[11].BalanceAmount, float64(81.43), fmt.Sprintf("Balance Amount DocNo: %s", stockDataLists[11].DocNo))
	assert.Equal(t, stockDataLists[11].CostPerUnit, float64(5), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[11].DocNo))
	assert.Equal(t, stockDataLists[11].TotalCost, float64(30), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[11].DocNo))
	assert.Equal(t, stockDataLists[11].BalanceAverage, float64(20.36), fmt.Sprintf("Balance Average DocNo: %s", stockDataLists[11].DocNo))

	assert.Equal(t, stockDataLists[12].CalcQty, float64(1), fmt.Sprintf("Doc Qty of DocNo: %s", stockDataLists[12].DocNo))
	assert.Equal(t, stockDataLists[12].BalanceQty, float64(3), fmt.Sprintf("Balance Qty DocNo: %s", stockDataLists[12].DocNo))
	assert.Equal(t, stockDataLists[12].BalanceAmount, float64(61.07), fmt.Sprintf("Balance Amount DocNo: %s", stockDataLists[12].DocNo))
	assert.Equal(t, stockDataLists[12].CostPerUnit, float64(20.36), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[12].DocNo))
	assert.Equal(t, stockDataLists[12].TotalCost, float64(20.36), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[12].DocNo))
	assert.Equal(t, stockDataLists[12].BalanceAverage, float64(20.36), fmt.Sprintf("Balance Average DocNo: %s", stockDataLists[12].DocNo))
}

func TestCalcStockSaleAndReturnMustBeNotNAN(t *testing.T) {

	FIX_SHOPID := "SHOPID"
	FIX_BARCODE := "885001"

	var stockDataLists []stockModel.StockData
	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:              FIX_SHOPID,
		Barcode:             FIX_BARCODE,
		DocNo:               "SALE01",
		TransFlag:           44,
		CalcFlag:            -1,
		CalcQty:             1,
		SumAmount:           69,
		SumAmountExcludeVat: 69,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:              FIX_SHOPID,
		Barcode:             FIX_BARCODE,
		DocNo:               "CN02",
		DocRef:              "SALE01",
		TransFlag:           48,
		CalcFlag:            1,
		CalcQty:             1,
		SumAmount:           69,
		SumAmountExcludeVat: 69,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:              FIX_SHOPID,
		Barcode:             FIX_BARCODE,
		DocNo:               "SALE03",
		TransFlag:           44,
		CalcFlag:            -1,
		CalcQty:             1,
		SumAmount:           69,
		SumAmountExcludeVat: 69,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:              FIX_SHOPID,
		Barcode:             FIX_BARCODE,
		DocNo:               "PU04",
		TransFlag:           12,
		CalcFlag:            1,
		CalcQty:             1,
		SumAmount:           100,
		SumAmountExcludeVat: 100,
	})

	os.Setenv("LOG_LEVEL", "debug")

	logger.NewAppLogger(config.NewLoggerConfig())

	giveBarcode := &productBarcodeModel.ProductBarcodePg{
		ShopID:     FIX_SHOPID,
		Barcode:    FIX_BARCODE,
		BalanceQty: 0,
	}

	repo := new(MockStockProcessPGRepository)
	repo.On("GetStockTransactionList", FIX_SHOPID, FIX_BARCODE).Return(stockDataLists, nil)
	repo.On("UpdateStockTransactionChange", mock.Anything).Return(nil)

	barcodeRepo := new(MockProductBarcodePGRepository)
	barcodeRepo.On("Get", FIX_SHOPID, FIX_BARCODE).Return(giveBarcode, nil)
	barcodeRepo.On("Update", FIX_SHOPID, FIX_BARCODE, giveBarcode).Return(nil)

	process := stockprocess.NewStockCalculator(repo, barcodeRepo)
	process.CalculatorStock(FIX_SHOPID, FIX_BARCODE)

	assert.Equal(t, stockDataLists[0].CalcQty, float64(1), fmt.Sprintf("Doc Qty of DocNo: %s", stockDataLists[0].DocNo))
	assert.Equal(t, stockDataLists[0].BalanceQty, float64(-1), fmt.Sprintf("Balance Qty DocNo: %s", stockDataLists[0].DocNo))
	assert.Equal(t, stockDataLists[0].BalanceAmount, float64(0), fmt.Sprintf("Balance Amount DocNo: %s", stockDataLists[0].DocNo))
	assert.Equal(t, stockDataLists[0].CostPerUnit, float64(0), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[0].DocNo))
	assert.Equal(t, stockDataLists[0].TotalCost, float64(0), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[0].DocNo))
	assert.Equal(t, stockDataLists[0].BalanceAverage, float64(0), fmt.Sprintf("Balance Average DocNo: %s", stockDataLists[0].DocNo))

	assert.Equal(t, stockDataLists[1].CalcQty, float64(1), fmt.Sprintf("Doc Qty of DocNo: %s", stockDataLists[1].DocNo))
	assert.Equal(t, stockDataLists[1].BalanceQty, float64(0), fmt.Sprintf("Balance Qty DocNo: %s", stockDataLists[1].DocNo))
	assert.Equal(t, stockDataLists[1].BalanceAmount, float64(0), fmt.Sprintf("Balance Amount DocNo: %s", stockDataLists[1].DocNo))
	assert.Equal(t, stockDataLists[1].TotalCost, float64(0), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[1].DocNo))
	assert.Equal(t, stockDataLists[1].CostPerUnit, float64(0), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[1].DocNo))
	assert.Equal(t, stockDataLists[1].BalanceAverage, float64(0), fmt.Sprintf("Balance Average DocNo: %s", stockDataLists[1].DocNo))

	assert.Equal(t, stockDataLists[2].CalcQty, float64(1), fmt.Sprintf("Doc Qty of DocNo: %s", stockDataLists[2].DocNo))
	assert.Equal(t, stockDataLists[2].BalanceQty, float64(-1), fmt.Sprintf("Balance Qty DocNo: %s", stockDataLists[2].DocNo))
	assert.Equal(t, stockDataLists[2].BalanceAmount, float64(0), fmt.Sprintf("Balance Amount DocNo: %s", stockDataLists[2].DocNo))
	assert.Equal(t, stockDataLists[2].CostPerUnit, float64(0), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[2].DocNo))
	assert.Equal(t, stockDataLists[2].TotalCost, float64(0), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[2].DocNo))
	assert.Equal(t, stockDataLists[2].BalanceAverage, float64(0), fmt.Sprintf("Balance Average DocNo: %s", stockDataLists[2].DocNo))

	assert.Equal(t, stockDataLists[3].CalcQty, float64(1), fmt.Sprintf("Doc Qty of DocNo: %s", stockDataLists[3].DocNo))
	assert.Equal(t, stockDataLists[3].BalanceQty, float64(0), fmt.Sprintf("Balance Qty DocNo: %s", stockDataLists[3].DocNo))
	assert.Equal(t, stockDataLists[3].BalanceAmount, float64(0), fmt.Sprintf("Balance Amount DocNo: %s", stockDataLists[3].DocNo))
	assert.Equal(t, stockDataLists[3].CostPerUnit, float64(100), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[3].DocNo))
	assert.Equal(t, stockDataLists[3].TotalCost, float64(100), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[3].DocNo))
	assert.Equal(t, stockDataLists[3].BalanceAverage, float64(0), fmt.Sprintf("Balance Average DocNo: %s", stockDataLists[3].DocNo))
}

func TestDebugStockNotCalc(t *testing.T) {

	FIX_SHOPID := "SHOPID"
	FIX_BARCODE := "885001"

	var stockDataLists []stockModel.StockData
	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:              FIX_SHOPID,
		Barcode:             FIX_BARCODE,
		DocNo:               "SALE01",
		TransFlag:           310,
		CalcFlag:            1,
		CalcQty:             10,
		SumAmount:           977.5,
		SumAmountExcludeVat: 913.5514018691589,
		CostPerUnit:         91.35,
		TotalCost:           913.55,
		BalanceAmount:       913.55,
		BalanceAverage:      91.35,
		BalanceQty:          10,
	})

	stockDataLists = append(stockDataLists, stockModel.StockData{
		ShopID:              FIX_SHOPID,
		Barcode:             FIX_BARCODE,
		DocNo:               "POS01",
		DocRef:              "",
		TransFlag:           44,
		CalcFlag:            -1,
		CalcQty:             1,
		SumAmount:           170,
		SumAmountExcludeVat: 158.8785046728972,
	})

	os.Setenv("LOG_LEVEL", "debug")

	logger.NewAppLogger(config.NewLoggerConfig())

	giveBarcode := &productBarcodeModel.ProductBarcodePg{
		ShopID:     FIX_SHOPID,
		Barcode:    FIX_BARCODE,
		BalanceQty: 0,
	}

	repo := new(MockStockProcessPGRepository)
	repo.On("GetStockTransactionList", FIX_SHOPID, FIX_BARCODE).Return(stockDataLists, nil)
	repo.On("UpdateStockTransactionChange", mock.Anything).Return(nil)

	barcodeRepo := new(MockProductBarcodePGRepository)
	barcodeRepo.On("Get", FIX_SHOPID, FIX_BARCODE).Return(giveBarcode, nil)
	barcodeRepo.On("Update", FIX_SHOPID, FIX_BARCODE, giveBarcode).Return(nil)

	process := stockprocess.NewStockCalculator(repo, barcodeRepo)
	process.CalculatorStock(FIX_SHOPID, FIX_BARCODE)

	assert.Equal(t, stockDataLists[0].CalcQty, float64(10), fmt.Sprintf("Doc Qty of DocNo: %s", stockDataLists[0].DocNo))
	assert.Equal(t, stockDataLists[0].BalanceQty, float64(10), fmt.Sprintf("Balance Qty DocNo: %s", stockDataLists[0].DocNo))
	assert.Equal(t, stockDataLists[0].BalanceAmount, float64(913.55), fmt.Sprintf("Balance Amount DocNo: %s", stockDataLists[0].DocNo))
	assert.Equal(t, stockDataLists[0].CostPerUnit, float64(91.35), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[0].DocNo))
	assert.Equal(t, stockDataLists[0].TotalCost, float64(913.55), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[0].DocNo))
	assert.Equal(t, stockDataLists[0].BalanceAverage, float64(91.35), fmt.Sprintf("Balance Average DocNo: %s", stockDataLists[0].DocNo))

	assert.Equal(t, stockDataLists[1].CalcQty, float64(1), fmt.Sprintf("Doc Qty of DocNo: %s", stockDataLists[1].DocNo))
	assert.Equal(t, stockDataLists[1].BalanceQty, float64(9), fmt.Sprintf("Balance Qty DocNo: %s", stockDataLists[1].DocNo))
	assert.Equal(t, stockDataLists[1].TotalCost, float64(91.35), fmt.Sprintf("Total Cost DocNo: %s", stockDataLists[1].DocNo))
	assert.Equal(t, stockDataLists[1].BalanceAmount, float64(822.2), fmt.Sprintf("Balance Amount DocNo: %s", stockDataLists[1].DocNo))
	assert.Equal(t, stockDataLists[1].CostPerUnit, float64(91.35), fmt.Sprintf("Cost Per Unit DocNo: %s", stockDataLists[1].DocNo))
	assert.Equal(t, stockDataLists[1].BalanceAverage, float64(91.36), fmt.Sprintf("Balance Average DocNo: %s", stockDataLists[1].DocNo))

}
