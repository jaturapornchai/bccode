package stockreceiveproduct_test

import (
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/stockreceiveproduct"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func wantStockReceiveProductTransactionPGStruct() models.StockReceiveProductTransactionPG {
	branchNames := []pkgModels.NameX{
		*pkgModels.NewNameXWithCodeName("th", "สาขาที่ 1"),
	}

	want := models.StockReceiveProductTransactionPG{
		TransactionPG: models.TransactionPG{
			ShopIdentity: pkgModels.ShopIdentity{
				ShopID: "2PrIIqTWxoBXv16K310sNwfHmfY",
			},
			GuidFixed:      "2QoOKOZ7Bv8dSUv38VTkkS2AZ0J",
			TransFlag:      60,
			DocNo:          "IF2023060100001",
			DocDate:        time.Date(2023, 6, 1, 1, 52, 44, 0, time.UTC),
			GuidRef:        "665f61bf-a49a-420b-9140-9824a704ef15",
			DocRefType:     0,
			DocRefNo:       "",
			DocRefDate:     time.Date(2023, 6, 6, 1, 36, 47, 0, time.UTC),
			BranchCode:     "branch01",
			BranchNames:    branchNames,
			TaxDocNo:       "",
			TaxDocDate:     time.Date(2023, 5, 31, 17, 0, 0, 0, time.UTC),
			Description:    "",
			InquiryType:    0,
			VatRate:        7,
			VatType:        0,
			DiscountWord:   "",
			TotalDiscount:  0,
			TotalValue:     12,
			TotalBeforeVat: 12,
			TotalExceptVat: 0,
			TotalVatValue:  0.84,
			TotalAfterVat:  12.84,
			TotalAmount:    12.84,
		},
		Items: &[]models.StockReceiveProductTransactionDetailPG{
			{
				TransactionDetailPG: models.TransactionDetailPG{
					GuidFixed:           "2QoOKOZ7Bv8dSUv38VTkkS2AZ0J",
					DocRef:              "",
					DocRefDateTime:      time.Date(2023, 6, 6, 1, 40, 28, 0, time.UTC),
					DocNo:               "IF2023060100001",
					ShopID:              "2PrIIqTWxoBXv16K310sNwfHmfY",
					LineNumber:          0,
					ItemGuid:            "2PrfDoufKF7KF0Ua2V6sbHBlm2R",
					Barcode:             "BARCODE001",
					UnitCode:            "ENV",
					Qty:                 1,
					Price:               12,
					PriceExcludeVat:     12,
					Discount:            "",
					DiscountAmount:      0,
					SumAmount:           12,
					SumAmountExcludeVat: 12,
					TotalValueVat:       0.84,
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

func TestStockReceiveProductTransactionPhaser(t *testing.T) {
	giveInput := `{
		"id": "647e919242c9a6d317361d78",
		"shopid": "2PrIIqTWxoBXv16K310sNwfHmfY",
		"guidfixed": "2QoOKOZ7Bv8dSUv38VTkkS2AZ0J",
		"docno": "IF2023060100001",
		"docdatetime": "2023-06-01T01:52:44.000Z",
		"guidref": "665f61bf-a49a-420b-9140-9824a704ef15",
		"transflag": 60,
		"docreftype": 0,
		"docrefno": "",
		"docrefdate": "2023-06-06T01:36:47.000Z",
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
		"taxdocno": "",
		"taxdocdate": "2023-05-31T17:00:00Z",
		"doctype": 0,
		"inquirytype": 0,
		"vattype": 0,
		"vatrate": 7,
		"custcode": "",
		"custnames": [],
		"description": "",
		"discountword": "",
		"totaldiscount": 0,
		"totalvalue": 12,
		"totalexceptvat": 0,
		"totalaftervat": 12.84,
		"totalbeforevat": 12,
		"totalvatvalue": 0.84,
		"totalamount": 12.84,
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
				"docdatetime": "2023-06-06T01:40:28.334Z",
				"docref": "",
				"docrefdatetime": "2023-06-06T01:40:28.000Z",
				"calcflag": 1,
				"barcode": "BARCODE001",
				"itemcode": "ITEM001",
				"unitcode": "ENV",
				"itemtype": 0,
				"itemguid": "2PrfDoufKF7KF0Ua2V6sbHBlm2R",
				"qty": 1,
				"totalqty": 1,
				"price": 12,
				"discount": "",
				"discountamount": 0,
				"totalvaluevat": 0.84,
				"priceexcludevat": 12,
				"sumamount": 12,
				"sumamountexcludevat": 12,
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
						"name": "มาม่า",
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
						"name": "ซอง",
						"isauto": false,
						"isdelete": false
					},
					{
						"code": "en",
						"name": "Envelope",
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

	want := wantStockReceiveProductTransactionPGStruct()

	phaser := stockreceiveproduct.StockReceiveTransactionPhaser{}
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
