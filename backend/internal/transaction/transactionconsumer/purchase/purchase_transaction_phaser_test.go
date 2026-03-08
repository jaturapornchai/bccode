package purchase_test

import (
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/purchase"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func PurchaseTransactionStruct() models.PurchaseTransactionPG {

	codeTh := "th"
	nameTh := "เจ้าหนี้ทั่วไป"

	want := models.PurchaseTransactionPG{
		TransactionPG: models.TransactionPG{
			ShopIdentity: pkgModels.ShopIdentity{
				ShopID: "2PrIIqTWxoBXv16K310sNwfHmfY",
			},
			GuidFixed:  "2RYA2Yri2HRKDF5JFnKpwuGmydO",
			TransFlag:  12,
			DocNo:      "PU2023062200001",
			DocDate:    time.Date(2023, 6, 22, 6, 46, 25, 0, time.UTC),
			GuidRef:    "bba805ec-f6aa-4568-b644-63147cd6cbcf",
			DocRefType: 5,
			DocRefNo:   "REFNO",
			DocRefDate: time.Date(2023, 6, 22, 6, 46, 25, 0, time.UTC),
			BranchCode: "branch01",
			BranchNames: []pkgModels.NameX{
				*pkgModels.NewNameXWithCodeName("th", "สาขาที่ 1"),
			},
			TaxDocNo:       "TAXPU2023062200001",
			TaxDocDate:     time.Date(2023, 6, 22, 6, 46, 25, 0, time.UTC),
			Description:    "Purchase Remark",
			InquiryType:    1,
			VatRate:        7,
			VatType:        1,
			DiscountWord:   "30",
			TotalDiscount:  30,
			TotalValue:     50,
			TotalBeforeVat: 46.728971962616825,
			TotalExceptVat: 0,
			TotalVatValue:  3.2710280373831777,
			TotalAfterVat:  50,
			TotalAmount:    50,
		},
		Items: &[]models.PurchaseTransactionDetailPG{
			{
				TransactionDetailPG: models.TransactionDetailPG{
					GuidFixed:           "2RYA2Yri2HRKDF5JFnKpwuGmydO",
					DocRef:              "detail doc ref",
					DocRefDateTime:      time.Date(2023, 6, 22, 6, 46, 43, 0, time.UTC),
					DocNo:               "PU2023062200001",
					ShopID:              "2PrIIqTWxoBXv16K310sNwfHmfY",
					LineNumber:          0,
					ItemGuid:            "2PrfDoufKF7KF0Ua2V6sbHBlm2R",
					Barcode:             "BARCODE001",
					UnitCode:            "ENV",
					Qty:                 10,
					Price:               5,
					PriceExcludeVat:     4.672897196261682,
					Discount:            "2",
					DiscountAmount:      2,
					SumAmount:           50,
					SumAmountExcludeVat: 46.728971962616825,
					TotalValueVat:       3.2710280373831777,
					WhCode:              "00000",
					LocationCode:        "LC001",
					VatType:             1,
					TaxType:             0,
					StandValue:          1,
					DivideValue:         1,
					ItemType:            0,
					Remark:              "detail remark",
					VatCal:              0,
				},
			},
		},
		TotalPayCash:     20,
		TotalPayCredit:   20,
		TotalPayTransfer: 10,
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

func TestPurchaseTransactionPhaser(t *testing.T) {

	giveInput := `{
		"id": "6493ee72a7408bc3e6035632",
		"shopid": "2PrIIqTWxoBXv16K310sNwfHmfY",
		"guidfixed": "2RYA2Yri2HRKDF5JFnKpwuGmydO",
		"docno": "PU2023062200001",
		"docdatetime": "2023-06-22T06:46:25.000Z",
		"guidref": "bba805ec-f6aa-4568-b644-63147cd6cbcf",
		"docreftype": 5,
		"docrefno": "REFNO",
		"docrefdate": "2023-06-22T06:46:25.000Z",
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
		"taxdocno": "TAXPU2023062200001",
		"taxdocdate": "2023-06-22T06:46:25.000Z",
		"description": "Purchase Remark",		
		"doctype": 0,
		"inquirytype": 1,
		"vattype": 1,
		"vatrate": 7,
		"discountword": "30",
		"totaldiscount": 30,
		"totalvalue": 50,
		"totalbeforevat": 46.728971962616825,
		"totalexceptvat": 0,		
		"totalvatvalue": 3.2710280373831777,
		"totalaftervat": 50,
		"totalamount": 50,
		"totalcost": 0,
		"transflag": 12,
		"posid": "",
		"cashiercode": "",
		"salecode": "",
		"salename": "",
		"membercode": "",
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
			"docref": "detail doc ref",
			"docrefdatetime": "2023-06-22T06:46:43.000Z",
			"docdatetime": "2023-06-22T06:46:43.000Z",
			"linenumber": 0,			
			"inquirytype": 1,
			"itemguid": "2PrfDoufKF7KF0Ua2V6sbHBlm2R",
			"barcode": "BARCODE001",
			"itemcode": "ITEM001",
			"unitcode": "ENV",
			"multiunit": true,
			"qty": 10,
			"price": 5,
			"discount": "2",
			"discountamount": 2,
			"priceexcludevat": 4.672897196261682,
			"sumamount": 50,
			"sumamountexcludevat": 46.728971962616825,
			"totalvaluevat": 3.2710280373831777,
			"whcode": "00000",			
			"locationcode": "LC001",
			"vattype": 1,
			"vatcal": 0,
			"taxtype": 0,
			"standvalue": 1,
			"dividevalue": 1,
			"itemtype": 0,
			"remark": "detail remark",
			"totalqty": 10,
			"calcflag": 1,
			"towhcode": "00000",
			"tolocationcode": "LC001",
			"shelfcode": "",
			"towhnames": [
			{
				"code": "th",
					"name": "คลังสำนักงานใหญ่",
					"isauto": false,
					"isdelete": false
				}
			],
			"locationnames": [
			  {
				"code": "th",
				"name": "ที่เก็บ สนง.",
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
			"whnames": [
			  {
				"code": "th",
				"name": "คลังสำนักงานใหญ่",
				"isauto": false,
				"isdelete": false
			  }
			],
			"tolocationnames": [
			  {
				"code": "th",
				"name": "ที่เก็บ สนง.",
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
			"averagecost": 0,
			"sumofcost": 0,
			"ispos": 0,
			"laststatus": 0
		  }
		],
		"paymentdetail": {
		  "cashamounttext": "",
		  "cashamount": 20,
		  "paymentcreditcards": [
			{
			  "docdatetime": "2023-06-22T09:39:49.18Z",
			  "cardnumber": "456452",
			  "amount": 20,
			  "chargeword": "0",
			  "chargevalue": 0,
			  "totalnetworth": 20
			}
		  ],
		  "paymenttransfers": [
			{
			  "docdatetime": "2023-06-22T09:39:40.942Z",
			  "bankcode": "SCB",
			  "banknames": [
				{
				  "code": "th",
				  "name": "ธ.ไทยพาณิชย์",
				  "isauto": false,
				  "isdelete": false
				}
			  ],
			  "accountnumber": "01234567890",
			  "amount": 10
			}
		  ]
		},
		"status": 0,
		"iscancel": false,
		"ismanualamount": false,
		"paymentdetailraw": ""
	  }`

	phaser := purchase.PurchaseTransactionPhaser{}
	want := PurchaseTransactionStruct()
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
	assert.Equal(t, (*get.Items)[0].WhCode, (*want.Items)[0].WhCode, "item.whcode")
	assert.Equal(t, (*get.Items)[0].LocationCode, (*want.Items)[0].LocationCode, "item.unitname")
	assert.Equal(t, (*get.Items)[0].Qty, (*want.Items)[0].Qty, "item.qty")
	assert.Equal(t, (*get.Items)[0].Price, (*want.Items)[0].Price, "item.price")
	assert.Equal(t, (*get.Items)[0].PriceExcludeVat, (*want.Items)[0].PriceExcludeVat, "item.PriceExcludeVat")
	assert.Equal(t, (*get.Items)[0].Discount, (*want.Items)[0].Discount, "item.discount")
	assert.Equal(t, (*get.Items)[0].DiscountAmount, (*want.Items)[0].DiscountAmount, "item.discountamount")
	assert.Equal(t, (*get.Items)[0].SumAmount, (*want.Items)[0].SumAmount, "item.sumamount")
	assert.Equal(t, (*get.Items)[0].SumAmountExcludeVat, (*want.Items)[0].SumAmountExcludeVat, "item.sumamountexcludevat")
	assert.Equal(t, (*get.Items)[0].StandValue, (*want.Items)[0].StandValue, "item.standvalue")
	assert.Equal(t, (*get.Items)[0].DivideValue, (*want.Items)[0].DivideValue, "item.dividevalue")
	assert.Equal(t, *((*get.Items)[0].ItemNames[0]).Name, "มาม่า", "item.ItemNames")
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
