package stockpickupproduct_test

import (
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/stockpickupproduct"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func wantStockPickupTransferStruct() models.StockPickUpTransactionPG {

	branchNames := []pkgModels.NameX{
		*pkgModels.NewNameXWithCodeName("th", "สาขาที่ 1"),
	}

	want := models.StockPickUpTransactionPG{
		TransactionPG: models.TransactionPG{
			HoldingCodeentity: pkgModels.HoldingCodeentity{
				HoldingCode: "2PrIIqTWxoBXv16K310sNwfHmfY",
			},
			GuidFixed:      "2PxeyrNxaW54VXcEOeyGcrMskcv",
			TransFlag:      56,
			DocNo:          "PO23051516490133",
			DocDate:        time.Date(2023, 5, 15, 9, 49, 58, 0, time.UTC),
			GuidRef:        "a71e8f3e-1efe-4222-994a-47c869599735",
			DocRefType:     0,
			DocRefNo:       "",
			DocRefDate:     time.Date(2023, 5, 18, 9, 49, 54, 0, time.UTC),
			BranchCode:     "branch01",
			BranchNames:    branchNames,
			TaxDocNo:       "",
			TaxDocDate:     time.Date(2023, 5, 18, 9, 49, 54, 0, time.UTC),
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
		Items: &[]models.StockPickUpTransactionDetailPG{
			{
				TransactionDetailPG: models.TransactionDetailPG{
					GuidFixed:           "2PxeyrNxaW54VXcEOeyGcrMskcv",
					DocRef:              "",
					DocRefDateTime:      time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC),
					DocNo:               "PO23051516490133",
					HoldingCode:         "2PrIIqTWxoBXv16K310sNwfHmfY",
					LineNumber:          0,
					ItemGuid:            "2Pxcf33JyR8jRXiKpH2cNc9lH9v",
					Barcode:             "BARCODE015",
					UnitCode:            "PCE",
					Qty:                 15,
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

func TestStockPickupTransactionDoc(t *testing.T) {

	giveInput := `{
		"id": "6465f4d9dfc8097596db4227",
		"holdingcode": "2PrIIqTWxoBXv16K310sNwfHmfY",
		"guidfixed": "2PxeyrNxaW54VXcEOeyGcrMskcv",
		"docno": "PO23051516490133",
		"docdatetime": "2023-05-15T09:49:58.000Z",
		"guid_ref": "a71e8f3e-1efe-4222-994a-47c869599735",
		"transflag": 56,
		"docreftype": 0,
		"docrefno": "",
		"docrefdate": "2023-05-18T09:49:54.000Z",
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
		"taxdocdate": "2023-05-18T09:49:54.000Z",
		"taxdocno": "",
		"doc_type": 0,
		"inquirytype": 0,
		"vat_type": 0,
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
		"total_amount": 0,
		"total_cost": 0,
		"posid": "",
		"cashiercode": "",
		"salecode": "SUP001",
		"salename": "นาง ดูแลดี",
		"membercode": "",
		"iscancel": false,
		"ismanualamount": false,
		"status": 0,
		"details": [
			{
				"inquirytype": 0,
				"line_number": 0,
				"docdatetime": "2023-05-18T09:50:09.934Z",
				"docref": "",
				"docrefdatetime": "0001-01-01T00:00:00Z",
				"calcflag": -1,
				"barcode": "BARCODE015",
				"itemcode": "",
				"unitcode": "PCE",
				"item_type": 0,
				"item_guid": "2Pxcf33JyR8jRXiKpH2cNc9lH9v",
				"qty": 15,
				"totalqty": 15,
				"price": 0,
				"discount": "",
				"discountamount": 0,
				"totalvaluevat": 0,
				"priceexcludevat": 0,
				"sum_amount": 0,
				"sumamountexcludevat": 0,
				"dividevalue": 1,
				"standvalue": 1,
				"vat_type": 0,
				"remark": "",
				"multiunit": false,
				"sumofcost": 0,
				"averagecost": 0,
				"laststatus": 0,
				"ispos": 0,
				"tax_type": 0,
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
	want := wantStockPickupTransferStruct()

	phaser := stockpickupproduct.StockPickUpTransactionPhaser{}
	get, err := phaser.PhaseSingleDoc(giveInput)

	assert.Nil(t, err)

	assert.Equal(t, get.HoldingCode, want.HoldingCode, "holdingcode")
	assert.Equal(t, get.GuidFixed, want.GuidFixed, "guidfixed")
	assert.Equal(t, get.TransFlag, want.TransFlag, "transflag")
	assert.Equal(t, get.DocNo, want.DocNo, "docno")
	assert.Equal(t, get.DocDate, want.DocDate, "docdate")
	assert.Equal(t, get.GuidRef, want.GuidRef, "guid_ref")
	assert.Equal(t, get.DocRefType, want.DocRefType, "docreftype")
	assert.Equal(t, get.DocRefNo, want.DocRefNo, "docrefno")
	assert.Equal(t, get.DocRefDate, want.DocRefDate, "docrefdate")
	assert.Equal(t, get.TaxDocNo, want.TaxDocNo, "taxdocno")
	assert.Equal(t, get.TaxDocDate, want.TaxDocDate, "taxdocdate")
	assert.Equal(t, get.Description, want.Description, "description")
	assert.Equal(t, get.InquiryType, want.InquiryType, "inquirytype")
	assert.Equal(t, get.VatRate, want.VatRate, "vatrate")
	assert.Equal(t, get.VatType, want.VatType, "vat_type")
	assert.Equal(t, get.DiscountWord, want.DiscountWord, "discountword")
	assert.Equal(t, get.TotalDiscount, want.TotalDiscount, "totaldiscount")
	assert.Equal(t, get.TotalValue, want.TotalValue, "totalvalue")
	assert.Equal(t, get.TotalBeforeVat, want.TotalBeforeVat, "totalbeforevat")
	assert.Equal(t, get.TotalExceptVat, want.TotalExceptVat, "totalexceptvat")
	assert.Equal(t, get.TotalVatValue, want.TotalVatValue, "totalvatvalue")
	assert.Equal(t, get.TotalAfterVat, want.TotalAfterVat, "totalaftervat")
	assert.Equal(t, get.TotalAmount, want.TotalAmount, "total_amount")

	// detail
	assert.Equal(t, (*get.Items)[0].GuidFixed, (*want.Items)[0].GuidFixed, "item.guidfixed")
	assert.Equal(t, (*get.Items)[0].DocNo, (*want.Items)[0].DocNo, "item.docno")
	assert.Equal(t, (*get.Items)[0].HoldingCode, (*want.Items)[0].HoldingCode, "item.holdingcode")
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

	wantEqual := want.CompareTo(&want)
	assert.Equal(t, wantEqual, true, "compare")

	// branch
	assert.Equal(t, want.BranchCode, get.BranchCode, "branchcode")
	assert.Equal(t, want.BranchNames, get.BranchNames, "branchnames")
}
