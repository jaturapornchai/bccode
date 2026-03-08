package stockadjustment_test

import (
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/stockadjustment"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func wantStockAdjustmentTransactionStruct() models.StockAdjustmentTransactionPG {

	branchNames := []pkgModels.NameX{
		*pkgModels.NewNameXWithCodeName("th", "สาขาที่ 1"),
	}

	want := models.StockAdjustmentTransactionPG{
		TransactionPG: models.TransactionPG{
			ShopIdentity: pkgModels.ShopIdentity{
				ShopID: "2PrIIqTWxoBXv16K310sNwfHmfY",
			},
			GuidFixed:      "2PxeTSlssQvMZS8MViihtgYOC0w",
			TransFlag:      66,
			DocNo:          "PO2305111645E3FE",
			DocDate:        time.Date(2023, 5, 11, 9, 45, 46, 0, time.UTC),
			GuidRef:        "1a14456e-c859-422e-8f3c-ca9787202f0c",
			DocRefType:     0,
			DocRefNo:       "",
			DocRefDate:     time.Date(2023, 5, 18, 9, 45, 39, 0, time.UTC),
			BranchCode:     "branch01",
			BranchNames:    branchNames,
			TaxDocNo:       "",
			TaxDocDate:     time.Date(2023, 5, 18, 9, 45, 39, 0, time.UTC),
			Description:    "",
			InquiryType:    0,
			VatRate:        0,
			VatType:        0,
			DiscountWord:   "",
			TotalDiscount:  0,
			TotalValue:     0,
			TotalBeforeVat: 0,
			TotalExceptVat: 0,
			TotalVatValue:  0,
			TotalAfterVat:  0,
			TotalAmount:    0,
		},
		Items: &[]models.StockAdjustmentTransactionDetailPG{
			{
				TransactionDetailPG: models.TransactionDetailPG{
					GuidFixed:           "2PxeTSlssQvMZS8MViihtgYOC0w",
					DocRef:              "",
					DocRefDateTime:      time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC),
					DocNo:               "PO2305111645E3FE",
					ShopID:              "2PrIIqTWxoBXv16K310sNwfHmfY",
					LineNumber:          0,
					ItemGuid:            "2Pxcf33JyR8jRXiKpH2cNc9lH9v",
					Barcode:             "BARCODE015",
					UnitCode:            "PCE",
					Qty:                 5,
					Price:               0,
					PriceExcludeVat:     0,
					Discount:            "",
					DiscountAmount:      0,
					SumAmount:           0,
					SumAmountExcludeVat: 0,
					TotalValueVat:       0,
					WhCode:              "00000",
					LocationCode:        "",
					VatType:             0,
					TaxType:             0,
					StandValue:          1,
					DivideValue:         1,
					ItemType:            0,
					Remark:              "",
					VatCal:              0,
				},
			},
		},
	}

	return want
}

func TestStockAdjustmentTransactionPhaser(t *testing.T) {

	giveInput := `{
		"id": "6465f3dfdfc8097596db421e",
		"shopid": "2PrIIqTWxoBXv16K310sNwfHmfY",
		"guidfixed": "2PxeTSlssQvMZS8MViihtgYOC0w",
		"docno": "PO2305111645E3FE",
		"docdatetime": "2023-05-11T09:45:46.000Z",
		"guidref": "1a14456e-c859-422e-8f3c-ca9787202f0c",
		"transflag": 66,
		"docreftype": 0,
		"docrefno": "",
		"docrefdate": "2023-05-18T09:45:39.000Z",
		"branch": {
			"code": "branch01",
			"names": [
				{
					"code": "th",
					"name": "สาขาที่ 1",
					"isauto": false,
					"isdelete": false
				}
			]
		},
		"taxdocdate": "2023-05-18T09:45:39.000Z",
		"taxdocno": "",
		"doctype": 0,
		"inquirytype": 0,
		"vattype": 0,
		"vatrate": 0,
		"custcode": "",
		"custnames": [],
		"description": "",
		"discountword": "",
		"totaldiscount": 0,
		"totalvalue": 0,
		"totalexceptvat": 0,
		"totalaftervat": 0,
		"totalbeforevat": 0,
		"totalvatvalue": 0,
		"totalamount": 0,
		"totalcost": 0,
		"posid": "",
		"cashiercode": "",
		"salecode": "",
		"salename": "",
		"membercode": "",
		"iscancel": false,
		"ismanualamount": false,
		"status": 0,
		"details": [
			{
				"inquirytype": 0,
				"linenumber": 0,
				"docdatetime": "2023-05-18T09:45:58.652Z",
				"docref": "",
				"docrefdatetime": "0001-01-01T00:00:00Z",
				"calcflag": 1,
				"barcode": "BARCODE015",
				"itemcode": "",
				"unitcode": "PCE",
				"itemtype": 0,
				"itemguid": "2Pxcf33JyR8jRXiKpH2cNc9lH9v",
				"qty": 5,
				"totalqty": 5,
				"price": 0,
				"discount": "",
				"discountamount": 0,
				"totalvaluevat": 0,
				"priceexcludevat": 0,
				"sumamount": 0,
				"sumamountexcludevat": 0,
				"dividevalue": 1,
				"standvalue": 1,
				"vattype": 0,
				"remark": "",
				"multiunit": false,
				"sumofcost": 0,
				"averagecost": 0,
				"laststatus": 0,
				"ispos": 0,
				"taxtype": 0,
				"vatcal": 0,
				"whcode": "00000",
				"shelfcode": "",
				"locationcode": "",
				"towhcode": "00000",
				"tolocationcode": "",
				"itemnames": [
					{
						"code": "th",
						"name": "สินค้า ทดสอบคำนวณสต๊อก",
						"isauto": false,
						"isdelete": false
					},
					{
						"code": "en",
						"name": "",
						"isauto": false,
						"isdelete": false
					},
					{
						"code": "ja",
						"name": "",
						"isauto": false,
						"isdelete": false
					}
				],
				"unitnames": [
					{
						"code": "th",
						"name": "ชิ้น",
						"isauto": false,
						"isdelete": false
					},
					{
						"code": "en",
						"name": "Piece",
						"isauto": false,
						"isdelete": false
					},
					{
						"code": "ja",
						"name": "",
						"isauto": false,
						"isdelete": false
					}
				],
				"whnames": [
					{
						"code": "th",
						"name": "คลังสำนักงานใหญ่",
						"isauto": false,
						"isdelete": false
					}
				],
				"locationnames": [],
				"towhnames": [
					{
						"code": "th",
						"name": "คลังสำนักงานใหญ่",
						"isauto": false,
						"isdelete": false
					}
				],
				"tolocationnames": []
			}
		],
		"paymentdetail": {
			"cashamounttext": "",
			"cashamount": 0,
			"paymentcreditcards": [],
			"paymenttransfers": []
		},
		"paymentdetailraw": ""
	}`
	want := wantStockAdjustmentTransactionStruct()

	phaser := stockadjustment.StockAdjustmentTransactionPhaser{}
	get, err := phaser.PhaseSingleDoc(giveInput)

	assert.Nil(t, err)

	assert.Equal(t, get.ShopID, want.ShopID, "shopid")
	assert.Equal(t, get.GuidFixed, want.GuidFixed, "guidfixed")
	assert.Equal(t, get.TransFlag, want.TransFlag, "transflag")
	assert.Equal(t, get.DocNo, want.DocNo, "docno")
	assert.Equal(t, get.DocDate, want.DocDate, "docdate")
	assert.Equal(t, get.GuidRef, want.GuidRef, "guidref")
	assert.Equal(t, get.DocRefType, want.DocRefType, "docreftype")
	assert.Equal(t, get.DocRefNo, want.DocRefNo, "docrefno")
	assert.Equal(t, get.DocRefDate, want.DocRefDate, "docrefdate")
	assert.Equal(t, get.TaxDocNo, want.TaxDocNo, "taxdocno")
	assert.Equal(t, get.TaxDocDate, want.TaxDocDate, "taxdocdate")
	assert.Equal(t, get.Description, want.Description, "description")
	assert.Equal(t, get.InquiryType, want.InquiryType, "inquirytype")
	assert.Equal(t, get.VatRate, want.VatRate, "vatrate")
	assert.Equal(t, get.VatType, want.VatType, "vattype")
	assert.Equal(t, get.DiscountWord, want.DiscountWord, "discountword")
	assert.Equal(t, get.TotalDiscount, want.TotalDiscount, "totaldiscount")
	assert.Equal(t, get.TotalValue, want.TotalValue, "totalvalue")
	assert.Equal(t, get.TotalBeforeVat, want.TotalBeforeVat, "totalbeforevat")
	assert.Equal(t, get.TotalExceptVat, want.TotalExceptVat, "totalexceptvat")
	assert.Equal(t, get.TotalVatValue, want.TotalVatValue, "totalvatvalue")
	assert.Equal(t, get.TotalAfterVat, want.TotalAfterVat, "totalaftervat")
	assert.Equal(t, get.TotalAmount, want.TotalAmount, "totalamount")

	// detail
	assert.Equal(t, (*get.Items)[0].GuidFixed, (*want.Items)[0].GuidFixed, "item.guidfixed")
	assert.Equal(t, (*get.Items)[0].DocNo, (*want.Items)[0].DocNo, "item.docno")
	assert.Equal(t, (*get.Items)[0].ShopID, (*want.Items)[0].ShopID, "item.shopid")
	assert.Equal(t, (*get.Items)[0].LineNumber, (*want.Items)[0].LineNumber, "item.linenumber")
	assert.Equal(t, (*get.Items)[0].ItemGuid, (*want.Items)[0].ItemGuid, "item.itemguid")
	assert.Equal(t, (*get.Items)[0].Barcode, (*want.Items)[0].Barcode, "item.barcode")
	assert.Equal(t, (*get.Items)[0].UnitCode, (*want.Items)[0].UnitCode, "item.unitcode")
	assert.Equal(t, (*get.Items)[0].WhCode, (*want.Items)[0].WhCode, "item.whcode")
	assert.Equal(t, (*get.Items)[0].LocationCode, (*want.Items)[0].LocationCode, "item.locationcode")
	assert.Equal(t, (*get.Items)[0].Qty, (*want.Items)[0].Qty, "item.qty")
	assert.Equal(t, (*get.Items)[0].Price, (*want.Items)[0].Price, "item.price")
	assert.Equal(t, (*get.Items)[0].PriceExcludeVat, (*want.Items)[0].PriceExcludeVat, "item.PriceExcludeVat")
	assert.Equal(t, (*get.Items)[0].Discount, (*want.Items)[0].Discount, "item.discount")
	assert.Equal(t, (*get.Items)[0].DiscountAmount, (*want.Items)[0].DiscountAmount, "item.discountamount")
	assert.Equal(t, (*get.Items)[0].SumAmount, (*want.Items)[0].SumAmount, "item.sumamount")
	assert.Equal(t, (*get.Items)[0].SumAmountExcludeVat, (*want.Items)[0].SumAmountExcludeVat, "item.sumamountexcludevat")
	assert.Equal(t, (*get.Items)[0].StandValue, (*want.Items)[0].StandValue, "item.standvalue")
	assert.Equal(t, (*get.Items)[0].DivideValue, (*want.Items)[0].DivideValue, "item.dividevalue")
	// assert.Equal(t, *((*get.Items)[0].ItemNames[0]).Name, "โอวัลติน ซอง", "item.ItemNames")
	assert.Equal(t, (*get.Items)[0].ItemType, (*want.Items)[0].ItemType, "item.itemtype")
	assert.Equal(t, (*get.Items)[0].Remark, (*want.Items)[0].Remark, "item.remark")
	assert.Equal(t, (*get.Items)[0].DocRef, (*want.Items)[0].DocRef, "item.docref")
	assert.Equal(t, (*get.Items)[0].DocRefDateTime, (*want.Items)[0].DocRefDateTime, "item.docrefdatetime")
	assert.Equal(t, (*get.Items)[0].VatCal, (*want.Items)[0].VatCal, "item.vatcal")

	// branch
	assert.Equal(t, want.BranchCode, get.BranchCode, "branchcode")
	assert.Equal(t, want.BranchNames, get.BranchNames, "branchnames")

	wantEqual := want.CompareTo(&want)
	assert.Equal(t, wantEqual, true, "compare")
}
