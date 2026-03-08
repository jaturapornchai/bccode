package purchasereturn_test

import (
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/purchasereturn"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func PurchaseReturnTransactionStruct() models.PurchaseReturnTransactionPG {

	codeTh := "th"
	nameTh := "เจ้าหนี้ทั่วไป"

	branchNames := []pkgModels.NameX{
		*pkgModels.NewNameXWithCodeName("th", "สาขาที่ 1"),
	}

	want := models.PurchaseReturnTransactionPG{

		TransactionPG: models.TransactionPG{
			ShopIdentity: pkgModels.ShopIdentity{
				ShopID: "2PrIIqTWxoBXv16K310sNwfHmfY",
			},
			GuidFixed:      "2PxduUIwAoptr2OTwROegQ98Uvq",
			TransFlag:      16,
			DocNo:          "PO23050616392C90",
			DocDate:        time.Date(2023, 5, 6, 9, 41, 21, 0, time.UTC),
			GuidRef:        "d4d2eddd-2f36-424f-92d2-3d0cb6c50b3f",
			DocRefType:     3,
			DocRefNo:       "PO2305051637AAD9",
			DocRefDate:     time.Date(2023, 5, 5, 9, 37, 25, 0, time.UTC),
			BranchCode:     "branch01",
			BranchNames:    branchNames,
			TaxDocNo:       "TAXXXXX",
			TaxDocDate:     time.Date(2023, 5, 18, 9, 37, 21, 0, time.UTC),
			Description:    "remark",
			InquiryType:    1,
			VatRate:        7,
			VatType:        2,
			DiscountWord:   "3%",
			TotalDiscount:  15,
			TotalValue:     20,
			TotalBeforeVat: 20,
			TotalExceptVat: 0,
			TotalVatValue:  0,
			TotalAfterVat:  20,
			TotalAmount:    20,
		},
		Items: &[]models.PurchaseReturnTransactionDetailPG{
			{
				TransactionDetailPG: models.TransactionDetailPG{
					GuidFixed:           "2PxduUIwAoptr2OTwROegQ98Uvq",
					DocNo:               "PO23050616392C90",
					ShopID:              "2PrIIqTWxoBXv16K310sNwfHmfY",
					LineNumber:          0,
					ItemGuid:            "2Pxcf33JyR8jRXiKpH2cNc9lH9v",
					Barcode:             "BARCODE015",
					UnitCode:            "PCE",
					Qty:                 2,
					Price:               10,
					PriceExcludeVat:     10,
					Discount:            "2%",
					DiscountAmount:      5,
					SumAmount:           20,
					SumAmountExcludeVat: 20,
					TotalValueVat:       20,
					WhCode:              "00000",
					LocationCode:        "",
					VatType:             2,
					TaxType:             0,
					StandValue:          1,
					DivideValue:         1,
					ItemType:            3,
					Remark:              "detail remark",
					DocRef:              "DetailDocRef",
					DocRefDateTime:      time.Date(2029, 1, 1, 0, 0, 0, 0, time.UTC),
					VatCal:              1,
				},
			},
		},
		TotalPayCash:     5,
		TotalPayCredit:   25,
		TotalPayTransfer: 20,
		CreditorCode:     "AP001",
		CreditorNames: []pkgModels.NameX{
			{
				Code:     &codeTh,
				Name:     &nameTh,
				IsAuto:   false,
				IsDelete: false,
			},
		},
	}
	return want
}

func TestPurchaseReturnTransactionPhaser(t *testing.T) {

	giveInput := `{
		"id": "6465f2c8dfc8097596db4215",
		"shopid": "2PrIIqTWxoBXv16K310sNwfHmfY",
		"guidfixed": "2PxduUIwAoptr2OTwROegQ98Uvq",
		"transflag": 16,
		"docno": "PO23050616392C90",
		"docdatetime": "2023-05-06T09:41:21.000Z",
		"guidref": "d4d2eddd-2f36-424f-92d2-3d0cb6c50b3f",
		"docreftype": 3,
		"docrefno": "PO2305051637AAD9",
		"docrefdate": "2023-05-05T09:37:25.000Z",
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
		"taxdocno": "TAXXXXX",
		"taxdocdate": "2023-05-18T09:37:21.000Z",
		"description": "remark",
		"doctype": 1,
		"inquirytype": 1,
		"vatrate": 7,
		"vattype": 2,
		"discountword": "3%",
		"totaldiscount": 15,
		"totalvalue": 20,
		"totalbeforevat": 20,
		"totalexceptvat": 0,
		"totalvatvalue": 0,
		"totalaftervat": 20,
		"totalamount": 20,		
		"membercode": "",
		"cashiercode": "",
		"salecode": "",
		"salename": "",
		"custcode": "AP001",
		"custnames": [
			{
				"code": "th",
				"name": "เจ้าหนี้ทั่วไป",
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
		"details": [
			{
				"inquirytype": 0,
				"itemguid": "2Pxcf33JyR8jRXiKpH2cNc9lH9v",
				"barcode": "BARCODE015",
				"unitcode": "PCE",
				"price": 10,
				"discount": "2%",
				"discountamount": 5,
				"sumamount": 20,
				"sumamountexcludevat": 20,
				"standvalue": 1,
				"dividevalue": 1,
				"whcode": "00000",
				"whnames": [
					{
						"code": "th",
						"name": "คลังสำนักงานใหญ่",
						"isauto": false,
						"isdelete": false
					}
				],
				"locationcode": "",
				"locationnames": [],
				"totalvaluevat": 0,
				"towhcode": "00000",
				"towhnames": [
					{
						"code": "th",
						"name": "คลังสำนักงานใหญ่",
						"isauto": false,
						"isdelete": false
					}
				],
				"tolocationcode": "",
				"tolocationnames": [],
				"shelfcode": "",
				"totalqty": 2,
				"calcflag": 1,
				"vattype": 2,
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
				"linenumber": 0,				
				"averagecost": 0,
				"laststatus": 0,
				"taxtype": 0,
				"itemcode": "",
				"ispos": 0,
				"multiunit": false,
				"priceexcludevat": 10,
				"itemtype": 3,
				"remark": "detail remark",
				"qty": 2,

				"docdatetime": "2023-05-18T09:37:43.42Z",
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
				"sumofcost": 0,
				"vatcal": 1,
				"docref": "DetailDocRef",
				"docrefdatetime": "2029-01-01T00:00:00Z"
			}
		],
		
		
		"totalcost": 0,
		"posid": "",
		"status": 0,
		"iscancel": false,
		"paymentdetail": {
			"cashamounttext": "",
			"cashamount": 5,
			"paymentcreditcards": [
			{
				"docdatetime": "2023-06-22T09:23:44.309Z",
				"cardnumber": "11111",
				"amount": 25,
				"chargeword": "0",
				"chargevalue": 0,
				"totalnetworth": 25
			}
			],
			"paymenttransfers": [
			  {
				"docdatetime": "2023-06-22T09:23:31.269Z",
				"bankcode": "SCB",
				"banknames": [
				  {
					"code": "th",
					"name": "ธ.ไทยพาณิชย์",
					"isauto": false,
					"isdelete": false
				  }
				],
				"accountnumber": "987654321",
				"amount": 20
			  }
			]
		},
		"ismanualamount": false,
		"paymentdetailraw": ""
	}`

	phaser := purchasereturn.PurchaseReturnTransactionPhaser{}

	want := PurchaseReturnTransactionStruct()
	get, err := phaser.PhaseSingleDoc(giveInput)
	// diff := cmp.Diff(get, want,
	// 	cmpopts.IgnoreFields(models.PurchaseReturnTransactionPG{}, "CreditorNames"),
	// 	cmpopts.IgnoreFields(models.PurchaseReturnTransactionDetailPG{}, "ID"),
	// )
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

	assert.Equal(t, get.CreditorCode, want.CreditorCode, "creditorcode")
	assert.Equal(t, *get.CreditorNames[0].Name, "เจ้าหนี้ทั่วไป", "creditorname")

	// detail
	assert.Equal(t, (*get.Items)[0].GuidFixed, (*want.Items)[0].GuidFixed, "item.guidfixed")
	assert.Equal(t, (*get.Items)[0].DocNo, (*want.Items)[0].DocNo, "item.docno")
	assert.Equal(t, (*get.Items)[0].ShopID, (*want.Items)[0].ShopID, "item.shopid")
	assert.Equal(t, (*get.Items)[0].LineNumber, (*want.Items)[0].LineNumber, "item.linenumber")
	assert.Equal(t, (*get.Items)[0].ItemGuid, (*want.Items)[0].ItemGuid, "item.itemguid")
	assert.Equal(t, (*get.Items)[0].Barcode, (*want.Items)[0].Barcode, "item.barcode")
	assert.Equal(t, (*get.Items)[0].UnitCode, (*want.Items)[0].UnitCode, "item.unitcode")
	assert.Equal(t, (*get.Items)[0].WhCode, (*want.Items)[0].WhCode, "item.unitname")
	assert.Equal(t, (*get.Items)[0].LocationCode, (*want.Items)[0].LocationCode, "item.unitname")
	assert.Equal(t, (*get.Items)[0].Qty, (*want.Items)[0].Qty, "item.unitname")
	assert.Equal(t, (*get.Items)[0].Price, (*want.Items)[0].Price, "item.price")
	assert.Equal(t, (*get.Items)[0].PriceExcludeVat, (*want.Items)[0].PriceExcludeVat, "item.price")
	assert.Equal(t, (*get.Items)[0].Discount, (*want.Items)[0].Discount, "item.discount")
	assert.Equal(t, (*get.Items)[0].DiscountAmount, (*want.Items)[0].DiscountAmount, "item.discountamount")
	assert.Equal(t, (*get.Items)[0].SumAmount, (*want.Items)[0].SumAmount, "item.sumamount")
	assert.Equal(t, (*get.Items)[0].SumAmountExcludeVat, (*want.Items)[0].SumAmountExcludeVat, "item.sumamountexcludevat")
	assert.Equal(t, (*get.Items)[0].StandValue, (*want.Items)[0].StandValue, "item.standvalue")
	assert.Equal(t, (*get.Items)[0].DivideValue, (*want.Items)[0].DivideValue, "item.dividevalue")
	assert.Equal(t, *((*get.Items)[0].ItemNames[0]).Name, "สินค้า ทดสอบคำนวณสต๊อก", "item.ItemNames")
	assert.Equal(t, (*get.Items)[0].ItemType, (*want.Items)[0].ItemType, "item.itemtype")
	assert.Equal(t, (*get.Items)[0].Remark, (*want.Items)[0].Remark, "item.remark")
	assert.Equal(t, (*get.Items)[0].DocRef, (*want.Items)[0].DocRef, "item.docref")
	assert.Equal(t, (*get.Items)[0].DocRefDateTime, (*want.Items)[0].DocRefDateTime, "item.docrefdatetime")
	assert.Equal(t, (*get.Items)[0].VatCal, (*want.Items)[0].VatCal, "item.vatcal")

	// payment
	assert.Equal(t, get.TotalPayCash, want.TotalPayCash, "totalpaycash")
	assert.Equal(t, get.TotalPayCredit, want.TotalPayCredit, "totalpaycredit")
	assert.Equal(t, get.TotalPayTransfer, want.TotalPayTransfer, "totalpaytransfer")

	// branch
	assert.Equal(t, want.BranchCode, get.BranchCode, "branchcode")
	assert.Equal(t, want.BranchNames, get.BranchNames, "branchnames")

}
