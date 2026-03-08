package models_test

import (
	commonModel "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
	"testing"

	"github.com/tj/assert"
)

var stockTransactionDoc models.StockTransaction
var stockTransactionPG models.StockTransaction

func init() {
	stockTransactionDoc = models.StockTransaction{
		ShopIdentity: commonModel.ShopIdentity{
			ShopID: "shoptester",
		},
		DocNo: "TRXTEST",
		Details: &[]models.StockTransactionDetail{
			{
				DocNo:   "TRXTEST",
				ShopID:  "shoptester",
				Barcode: "BAR1",
			},
			{
				DocNo:   "TRXTEST",
				ShopID:  "shoptester",
				Barcode: "BAR2",
			},
		},
	}

	stockTransactionPG = models.StockTransaction{

		ShopIdentity: commonModel.ShopIdentity{
			ShopID: "shoptester",
		},
		DocNo: "TRXTEST",
		Details: &[]models.StockTransactionDetail{
			{
				ID:          1,
				DocNo:       "TRXTEST",
				ShopID:      "shoptester",
				Barcode:     "BAR1",
				TotalCost:   1.0,
				CostPerUnit: 1.0,
			},
			{
				ID:          2,
				DocNo:       "TRXTEST",
				ShopID:      "shoptester",
				Barcode:     "BAR2",
				TotalCost:   2.0,
				CostPerUnit: 2.0,
			},
		},
	}
}

func TestCompareStockTransaction(t *testing.T) {
	isEqual := stockTransactionDoc.CompareTo(&stockTransactionPG)
	assert.Equal(t, isEqual, true)
}
