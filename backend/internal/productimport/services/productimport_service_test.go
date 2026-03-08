package services_test

import (
	"fmt"
	"smlcloudplatform/internal/productimport/models"
	"smlcloudplatform/internal/productimport/services"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrepareProductBarcodes(t *testing.T) {
	svc := services.ProductImportService{}

	languageCode := "en"
	docs := []models.ProductImportDoc{}

	for i := 1; i <= 10; i++ {
		doc := models.ProductImportDoc{}
		doc.Barcode = fmt.Sprintf("barcode%d", i)
		doc.Name = fmt.Sprintf("name%d", i)
		doc.UnitCode = fmt.Sprintf("unitcode%d", i)
		doc.Price = 1000
		doc.PriceMember = 900

		docs = append(docs, doc)
	}

	resultDocs := svc.PrepareProductBarcodes(languageCode, docs)

	assert.Equal(t, len(resultDocs), 10)

	for i := 1; i <= 10; i++ {

		expectDoc := docs[i-1]
		resultDoc := resultDocs[i-1]

		assert.Equal(t, expectDoc.Barcode, resultDoc.Barcode)
		assert.Equal(t, expectDoc.UnitCode, resultDoc.ItemUnitCode)

		tempResultNames := *resultDoc.Names

		assert.Equal(t, 1, len(tempResultNames))
		assert.Equal(t, expectDoc.Name, *tempResultNames[0].Name)
		assert.Equal(t, languageCode, *tempResultNames[0].Code)

		tempResultPrices := *resultDoc.Prices

		assert.Equal(t, 2, len(tempResultPrices))

		assert.Equal(t, expectDoc.Price, tempResultPrices[0].Price)
		assert.Equal(t, expectDoc.PriceMember, tempResultPrices[1].Price)

	}

}
