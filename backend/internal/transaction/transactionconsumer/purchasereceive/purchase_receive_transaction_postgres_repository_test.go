package purchasereceive_test

import (
	"errors"
	"os"
	"smlcloudplatform/internal/config"
	"smlcloudplatform/internal/logger"
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/purchasereceive"
	"smlcloudplatform/mock"
	"smlcloudplatform/pkg/microservice"

	microModels "smlcloudplatform/internal/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestPurcahseReceiveRepository(t *testing.T) {

	if os.Getenv("SERVERLESS") == "serverless" {
		t.Skip()
	}

	_ = logger.NewAppLogger(config.NewLoggerConfig())
	dbConfig := mock.NewPersisterPostgresqlConfig()
	persisterPG := microservice.NewPersister(dbConfig)
	repo := purchasereceive.NewPurchaseReceiveTransactionPGRepository(persisterPG)

	doc := purchaesReceiveDoc()

	t.Run("Create Purchase Receive", func(t *testing.T) {

		CleanUpData(persisterPG)

		err := repo.Create(doc)
		if err != nil {
			t.Errorf("Error creating purchase receive: %v", err)
		}

		result, err := repo.Get("TESTSHOP", doc.DocNo)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "TEST_PI0001", result.DocNo)
		assert.Equal(t, 2, len(*result.Items))
	})

	t.Run("Update Purchase Receive", func(t *testing.T) {

		CleanUpData(persisterPG)

		err := repo.Create(doc)
		if err != nil {
			t.Errorf("Error creating purchase receive: %v", err)
		}

		result, err := repo.Get("TESTSHOP", doc.DocNo)
		assert.NoError(t, err)

		result.CreditorCode = "CRED002"
		(*result.Items)[0].Qty = 10
		(*result.Items)[0].SumAmount = 1000

		err = repo.Update("TESTSHOP", doc.DocNo, *result)
		assert.NoError(t, err)

		updatedResult, err := repo.Get("TESTSHOP", doc.DocNo)
		assert.NoError(t, err)
		assert.Equal(t, "CRED002", updatedResult.CreditorCode)
		assert.Equal(t, float64(10), (*updatedResult.Items)[0].Qty)
		assert.Equal(t, float64(1000), (*updatedResult.Items)[0].SumAmount)
	})

	t.Run("Delete Purchase Receive", func(t *testing.T) {

		CleanUpData(persisterPG)

		err := repo.Create(doc)
		if err != nil {
			t.Errorf("Error creating purchase receive: %v", err)
		}

		result, err := repo.Get("TESTSHOP", doc.DocNo)
		assert.NoError(t, err)
		assert.NotNil(t, result)

		err = repo.DeleteData("TESTSHOP", doc.DocNo, *result)
		assert.NoError(t, err)

		deletedResult, err := repo.Get("TESTSHOP", doc.DocNo)
		assert.Error(t, err)
		assert.Nil(t, deletedResult)
		assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
	})
}

func purchaesReceiveDoc() models.PurchaseReceiveTransactionPG {

	testDocNo := "TEST_PI0001"
	thaiLang := "th"
	branchNameThai := "สาขาทดสอบ"
	apNameThai := "เจ้าหนี้ทดสอบ"
	productNameThai := "สินค้าทดสอบ"
	warehouseNameThai := "คลังสินค้าหลัก"
	locationNameThai := "ตำแหน่ง A"
	docDate := time.Now()

	return models.PurchaseReceiveTransactionPG{
		TransactionPG: models.TransactionPG{
			GuidFixed: "test-guid-fixed",
			HoldingCodeentity: microModels.HoldingCodeentity{
				HoldingCode: "TESTSHOP",
			},
			InquiryType: 1,
			TransFlag:   1,
			DocNo:       testDocNo,
			DocDate:     docDate,
			BranchCode:  "BR001",
			BranchNames: microModels.JSONB{
				microModels.NameX{
					Code: &thaiLang,
					Name: &branchNameThai,
				},
			},
		},
		CreditorCode: "CRED001",
		CreditorNames: microModels.JSONB{
			microModels.NameX{
				Code: &thaiLang,
				Name: &apNameThai,
			},
		},
		Items: &[]models.PurchaseReceiveTransactionDetailPG{
			{
				TransactionDetailPG: models.TransactionDetailPG{
					HoldingCode: "TESTSHOP",
					DocNo:       testDocNo,
					Barcode:     "ITEM001",
					ItemNames: microModels.JSONB{
						microModels.NameX{
							Code: &thaiLang,
							Name: &productNameThai,
						},
					},
					Qty:       5,
					Price:     100,
					SumAmount: 500,
					WhCode:    "WH001",
					WhNames: microModels.JSONB{
						microModels.NameX{
							Code: &thaiLang,
							Name: &warehouseNameThai,
						},
					},
					LocationCode: "LOC001",
					LocationNames: microModels.JSONB{
						microModels.NameX{
							Code: &thaiLang,
							Name: &locationNameThai,
						},
					},
					DocDate: docDate,
				},
			},
			{
				TransactionDetailPG: models.TransactionDetailPG{
					HoldingCode: "TESTSHOP",
					DocNo:       "TEST_001",
					Barcode:     "ITEM002",
					ItemNames: microModels.JSONB{
						microModels.NameX{
							Code: &thaiLang,
							Name: &productNameThai,
						},
					},
					Qty:       1,
					Price:     100,
					SumAmount: 100,
					WhCode:    "WH002",
					WhNames: microModels.JSONB{
						microModels.NameX{
							Code: &thaiLang,
							Name: &warehouseNameThai,
						},
					},
					LocationCode: "LOC002",
					LocationNames: microModels.JSONB{
						microModels.NameX{
							Code: &thaiLang,
							Name: &locationNameThai,
						},
					},
					DocDate: docDate,
				},
			},
		},
	}

}

func CleanUpData(pst *microservice.Persister) {

	_ = pst.DBClient().Exec("DELETE FROM purchasereceive_transaction_detail WHERE holdingcode='TESTSHOP' AND docno like 'TEST_%'").Error
	_ = pst.DBClient().Exec("DELETE FROM purchasereceive_transaction WHERE holdingcode='TESTSHOP' AND docno like 'TEST_%'").Error

}
