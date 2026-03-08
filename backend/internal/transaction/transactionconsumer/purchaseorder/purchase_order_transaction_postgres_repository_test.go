package purchaseorder_test

import (
	"errors"
	"os"
	"smlcloudplatform/internal/config"
	"smlcloudplatform/internal/logger"
	microModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/purchaseorder"
	"smlcloudplatform/mock"
	"smlcloudplatform/pkg/microservice"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// Helper function to create test data
func createTestPurchaseOrderTransaction() models.PurchaseOrderTransactionPG {

	thaiLang := "th"
	branchNameThai := "สาขาทดสอบ"
	apNameThai := "เจ้าหนี้ทดสอบ"
	productNameThai := "สินค้าทดสอบ"
	warehouseNameThai := "คลังสินค้าหลัก"
	locationNameThai := "ตำแหน่ง A"
	docDate := time.Now()

	return models.PurchaseOrderTransactionPG{
		TransactionPG: models.TransactionPG{
			GuidFixed: "test-guid-fixed",
			ShopIdentity: microModels.ShopIdentity{
				ShopID: "TEST001",
			},
			InquiryType: 1,
			TransFlag:   1,
			DocNo:       "PO001",
			DocDate:     docDate,
			DocRefType:  1,
			DocRefNo:    "REF001",
			DeviceName:  "TEST_DEVICE",
			GuidPos:     "pos-guid",
			DocRefDate:  time.Now(),
			BranchCode:  "BR001",
			BranchNames: microModels.JSONB{
				microModels.NameX{
					Code: &thaiLang,
					Name: &branchNameThai,
				},
			},
			Description:    "Test purchase order",
			TaxDocNo:       "TAX001",
			TaxDocDate:     time.Now(),
			IsCancel:       false,
			IsBom:          false,
			Status:         1,
			VatType:        1,
			VatRate:        7.0,
			TotalValue:     1000.0,
			DiscountWord:   "10%",
			DeliveryAmount: 50.0,
			TotalDiscount:  100.0,
			TotalBeforeVat: 900.0,
			TotalVatValue:  63.0,
			TotalExceptVat: 837.0,
			TotalAfterVat:  963.0,
			TotalAmount:    1013.0,
			GuidRef:        "ref-guid",
		},
		CreditorCode: "CRED001",
		CreditorNames: microModels.JSONB{
			microModels.NameX{
				Code: &thaiLang,
				Name: &apNameThai,
			},
		},
		Items: &[]models.PurchaseOrderDetailTransactionPG{
			{
				TransactionDetailPG: models.TransactionDetailPG{
					ID:         1,
					ShopID:     "TEST001",
					DocNo:      "PO001",
					LineNumber: 1,
					Barcode:    "123456789",
					ItemNames: microModels.JSONB{
						microModels.NameX{
							Code: &thaiLang,
							Name: &productNameThai,
						},
					},
					UnitCode:            "PCS",
					Qty:                 10.0,
					Price:               100.0,
					PriceExcludeVat:     93.46,
					Discount:            "5%",
					DiscountAmount:      50.0,
					SumAmount:           950.0,
					SumAmountExcludeVat: 887.85,
					SumAmountChoice:     950.0,
					RefGuid:             "item-ref-guid",
					WhCode:              "WH001",
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
					VatCal:         1,
					FoodType:       1,
					VatType:        1,
					TaxType:        1,
					IsChoice:       0,
					StandValue:     100.0,
					DivideValue:    1.0,
					ItemType:       1,
					ItemGuid:       "item-guid",
					TotalValueVat:  66.54,
					DocRef:         "DOC-REF-001",
					DocRefDateTime: time.Now(),
					Remark:         "Test remark",
					DocDate:        docDate,
				},
			},
		},
	}
}

func TestCreatPurchaseOrder(t *testing.T) {

	if os.Getenv("SERVERLESS") == "serverless" {
		t.Skip()
	}

	_ = logger.NewAppLogger(config.NewLoggerConfig())
	dbConfig := mock.NewPersisterPostgresqlConfig()
	persisterPG := microservice.NewPersister(dbConfig)
	repo := purchaseorder.NewPurchaseOrderTransactionRepository(persisterPG)

	t.Run("Create_Success", func(t *testing.T) {
		testData := createTestPurchaseOrderTransaction()

		err := repo.Create(testData)
		assert.NoError(t, err)
	})

	t.Run("Create_WithDatabaseError", func(t *testing.T) {
		// Test with invalid data that would cause database error
		testData := models.PurchaseOrderTransactionPG{
			TransactionPG: models.TransactionPG{
				DocNo: "", // Empty DocNo should cause validation error
			},
		}

		err := repo.Create(testData)
		// We expect an error due to validation constraints
		assert.Error(t, err)
	})

}

func TestGetPurchaseOrder(t *testing.T) {

	if os.Getenv("SERVERLESS") == "serverless" {
		t.Skip()
	}

	_ = logger.NewAppLogger(config.NewLoggerConfig())
	dbConfig := mock.NewPersisterPostgresqlConfig()
	persisterPG := microservice.NewPersister(dbConfig)
	repo := purchaseorder.NewPurchaseOrderTransactionRepository(persisterPG)

	t.Run("Get_Success", func(t *testing.T) {
		// First create a test record
		// testData := createTestPurchaseOrderTransaction()
		//testData.DocNo = "PO_GET_TEST_001"
		//testData.ShopID = "TEST001"

		// err := repo.Create(testData)
		// assert.NoError(t, err)

		// Now try to get it
		result, err := repo.Get("TEST001", "PO001")
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "PO001", result.DocNo)
		assert.Equal(t, "TEST001", result.ShopID)
		assert.Equal(t, "CRED001", result.CreditorCode)
	})

	t.Run("Get_NotFound", func(t *testing.T) {
		result, err := repo.Get("NONEXISTENT", "NONEXISTENT")
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
	})

	t.Run("Get_EmptyParameters", func(t *testing.T) {
		result, err := repo.Get("", "")
		assert.Error(t, err)
		assert.Nil(t, result)
	})

}

func TestUpdatePurchaseOrder(t *testing.T) {

	if os.Getenv("SERVERLESS") == "serverless" {
		t.Skip()
	}

	_ = logger.NewAppLogger(config.NewLoggerConfig())
	dbConfig := mock.NewPersisterPostgresqlConfig()
	persisterPG := microservice.NewPersister(dbConfig)
	repo := purchaseorder.NewPurchaseOrderTransactionRepository(persisterPG)

	t.Run("Update_Success", func(t *testing.T) {
		// First create a test record
		result, err := repo.Get("TEST001", "PO001")
		assert.NoError(t, err)
		testData := *result

		// Modify the data
		testData.Description = "Updated description"
		testData.TotalAmount = 2000.0
		testData.CreditorCode = "CRED002"

		// update detail
		(*testData.Items)[0].Remark = "Updated item remark"

		// Update the record
		err = repo.Update(result.ShopID, result.DocNo, testData)
		assert.NoError(t, err)

		// Verify the update
		result, err = repo.Get(result.ShopID, result.DocNo)
		assert.NoError(t, err)
		assert.Equal(t, "Updated description", result.Description)
		assert.Equal(t, 2000.0, result.TotalAmount)
		assert.Equal(t, "CRED002", result.CreditorCode)
	})

	t.Run("Update_NonExistentRecord", func(t *testing.T) {
		testData := createTestPurchaseOrderTransaction()
		testData.DocNo = "NONEXISTENT"
		testData.ShopID = "NONEXISTENT"

		err := repo.Update("NONEXISTENT", "NONEXISTENT", testData)
		// Update should not return error even if no records are affected
		assert.NoError(t, err)
	})

	t.Run("Update_EmptyParameters", func(t *testing.T) {
		testData := createTestPurchaseOrderTransaction()

		err := repo.Update("", "", testData)
		// Update with empty parameters should still execute but won't find any records
		assert.NoError(t, err)
	})

}

func TestDeleteData(t *testing.T) {

	if os.Getenv("SERVERLESS") == "serverless" {
		t.Skip()
	}

	_ = logger.NewAppLogger(config.NewLoggerConfig())
	dbConfig := mock.NewPersisterPostgresqlConfig()
	persisterPG := microservice.NewPersister(dbConfig)
	repo := purchaseorder.NewPurchaseOrderTransactionRepository(persisterPG)

	t.Run("Delete_Success", func(t *testing.T) {
		// First create a test record with details
		// testData := createTestPurchaseOrderTransaction()
		// testData.DocNo = "PO_DELETE_TEST_001"
		// testData.ShopID = "TEST001"

		// err := repo.Create(testData)
		// assert.NoError(t, err)

		// Verify it exists
		result, err := repo.Get("TEST001", "PO001")
		assert.NoError(t, err)
		assert.NotNil(t, result)

		// Delete the record
		err = repo.DeleteData("TEST001", result.DocNo, *result)
		assert.NoError(t, err)

		// Verify it's deleted
		result, err = repo.Get("TEST001", result.DocNo)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
	})

	// t.Run("Delete_NonExistentRecord", func(t *testing.T) {
	// 	testData := createTestPurchaseOrderTransaction()

	// 	err := repo.DeleteData("NONEXISTENT", "NONEXISTENT", testData)
	// 	// Delete should not return error even if no records are found
	// 	assert.NoError(t, err)
	// })

	// t.Run("Delete_WithMultipleDetails", func(t *testing.T) {

	// 	thaiLang := "th"
	// 	productNameThai := "สินค้าทดสอบ"
	// 	// Create a test record with multiple detail lines
	// 	testData := createTestPurchaseOrderTransaction()
	// 	testData.DocNo = "PO_DELETE_MULTI_001"
	// 	testData.ShopID = "TEST001"

	// 	// Add more detail items
	// 	additionalItem := models.PurchaseOrderDetailTransactionPG{
	// 		TransactionDetailPG: models.TransactionDetailPG{
	// 			ID:         2,
	// 			ShopID:     "TEST001",
	// 			DocNo:      "PO_DELETE_MULTI_001",
	// 			LineNumber: 2,
	// 			Barcode:    "987654321",
	// 			ItemNames: microModels.JSONB{
	// 				microModels.NameX{
	// 					Code: &thaiLang,
	// 					Name: &productNameThai,
	// 				},
	// 			},
	// 			UnitCode:  "PCS",
	// 			Qty:       5.0,
	// 			Price:     200.0,
	// 			SumAmount: 1000.0,
	// 		},
	// 	}
	// 	*testData.Items = append(*testData.Items, additionalItem)

	// 	err := repo.Create(testData)
	// 	assert.NoError(t, err)

	// 	// Delete the record (should delete both header and all details)
	// 	err = repo.DeleteData("TEST001", "PO_DELETE_MULTI_001", testData)
	// 	assert.NoError(t, err)

	// 	// Verify it's deleted
	// 	result, err := repo.Get("TEST001", "PO_DELETE_MULTI_001")
	// 	assert.Error(t, err)
	// 	assert.Nil(t, result)
	// 	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
	// })

	// t.Run("Delete_EmptyParameters", func(t *testing.T) {
	// 	testData := createTestPurchaseOrderTransaction()

	// 	err := repo.DeleteData("", "", testData)
	// 	// Delete with empty parameters should execute but won't find records
	// 	assert.NoError(t, err)
	// })
}
