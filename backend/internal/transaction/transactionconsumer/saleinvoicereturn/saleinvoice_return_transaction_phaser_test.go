package saleinvoicereturn_test

import (
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/saleinvoicereturn"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func SaleInvoiceReturnTransactionStruct() models.SaleInvoiceReturnTransactionPG {

	codeTh := "th"
	nameTh := "นาง เมย์ ไฟแรง"
	give := models.SaleInvoiceReturnTransactionPG{
		TransactionPG: models.TransactionPG{
			ShopIdentity: pkgModels.ShopIdentity{
				ShopID: "2PrIIqTWxoBXv16K310sNwfHmfY",
			},
			GuidFixed:      "2RFXUaW570MAWkgYgDduGM9WYIk",
			TransFlag:      48,
			DocNo:          "ST2023061500001",
			DocDate:        time.Date(2023, 6, 15, 16, 33, 32, 0, time.UTC),
			GuidRef:        "e846ae42-f506-4ee1-b588-7d395de13e7e",
			DocRefType:     4,
			DocRefNo:       "SALEINVOICE",
			DocRefDate:     time.Date(2023, 6, 15, 16, 33, 32, 0, time.UTC),
			TaxDocNo:       "ST2023061500001",
			TaxDocDate:     time.Date(2023, 6, 15, 16, 33, 32, 0, time.UTC),
			Description:    "CN",
			InquiryType:    22,
			VatRate:        7,
			VatType:        1,
			DiscountWord:   "9",
			TotalDiscount:  9,
			TotalValue:     280,
			TotalBeforeVat: 261.68224299065423,
			TotalExceptVat: 2,
			TotalVatValue:  18.317757009345794,
			TotalAfterVat:  280,
			TotalAmount:    280,
		},
		Items: &[]models.SaleInvoiceReturnTransactionDetailPG{
			{
				TransactionDetailPG: models.TransactionDetailPG{
					GuidFixed:           "2RFXUaW570MAWkgYgDduGM9WYIk",
					DocRef:              "ITEM001",
					DocRefDateTime:      time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC),
					DocNo:               "ST2023061500001",
					ShopID:              "2PrIIqTWxoBXv16K310sNwfHmfY",
					LineNumber:          1,
					ItemGuid:            "2PrfeYHTr2Sbo1oRU9rgdm9H35r",
					Barcode:             "BARCODE003",
					UnitCode:            "BOX",
					Qty:                 1,
					Price:               280,
					PriceExcludeVat:     261.68224299065423,
					Discount:            "7",
					DiscountAmount:      7,
					SumAmount:           280,
					SumAmountExcludeVat: 261.68224299065423,
					TotalValueVat:       18.317757009345794,
					WhCode:              "00000",
					LocationCode:        "LC001",
					VatType:             2,
					TaxType:             0,
					StandValue:          1,
					DivideValue:         1,
					ItemType:            0,
					Remark:              "-",
					VatCal:              0,
				},
			},
		},
		TotalPayCash:     5,
		TotalPayCredit:   0,
		TotalPayTransfer: 0,
		DebtorCode:       "AR002",
		DebtorNames: []pkgModels.NameX{
			{
				Code:     &codeTh,
				Name:     &nameTh,
				IsAuto:   false,
				IsDelete: false,
			},
		},
	}
	return give
}

func TestSaleInvoiceTransactionPhaser(t *testing.T) {
	giveInput := `{
		"id": "648b3d633d5e36f1165454b5",
		"shopid": "2PrIIqTWxoBXv16K310sNwfHmfY",
		"guidfixed": "2RFXUaW570MAWkgYgDduGM9WYIk",
		"docno": "ST2023061500001",
		"guidref": "e846ae42-f506-4ee1-b588-7d395de13e7e",
		"docdatetime": "2023-06-15T16:33:32.000Z",
		"docrefno": "SALEINVOICE",
		"docrefdate": "2023-06-15T16:33:32.000Z",
		"taxdocno": "ST2023061500001",
		"taxdocdate": "2023-06-15T16:33:32.000Z",
		"doctype": 0,
		"inquirytype": 22,
		"discountword": "9",
		"totalcost": 0,
		"totaldiscount": 9,
		"totalbeforevat": 261.68224299065423,
		"totalvatvalue": 18.317757009345794,
		"totalexceptvat": 2,
		"totalamount": 280,
		"salecode": "",
		"posid": "",
		"salename": "",
		"membercode": "",
		"vatrate": 7,
		"totalvalue": 280,
		"docreftype": 4,
		"vattype": 1,
		"totalaftervat": 280,
		"transflag": 48,
		"status": 0,
		"iscancel": false,
		"description": "CN",
		"ismanualamount": false,
		"branch": {
			"code": "b01",
			"names": [
				{
					"code": "th",
					"name": "สาขาที่ 1",
					"isauto": false,
					"isdelete": false
				}
			]
		},
		"custcode": "AR002",
		"custnames": [
			{
				"code": "th",
				"name": "นาง เมย์ ไฟแรง",
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
		"cashiercode": "",
		"details": [
			{
				"linenumber": 1,
				"docdatetime": "2023-06-15T16:33:35.735Z",
				"itemguid": "2PrfeYHTr2Sbo1oRU9rgdm9H35r",
				"barcode": "BARCODE003",
				"unitcode": "BOX",
				"itemcode": "ITEM001",
				"qty": 1,
				"discount": "7",
				"discountamount": 7,				
				"price": 280,
				"priceexcludevat": 261.68224299065423,
				"sumamount": 280,				
				"sumamountexcludevat": 261.68224299065423,
				"totalvaluevat": 18.317757009345794,
				"dividevalue": 1,
				"standvalue": 1,
				"inquirytype": 0,
				"totalqty": 1,
				"calcflag": 1,
				"vattype": 2,
				"averagecost": 0,
				"taxtype": 0,
				"ispos": 0,
				"multiunit": true,
				"itemtype": 0,
				"remark": "-",
				"sumofcost": 0,
				"docref": "ITEM001",
				"docrefdatetime": "2023-06-15T16:33:35.735Z",
				"vatcal": 0,
				"laststatus": 0,
				"locationcode": "LC001",
				"whcode": "00000",
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
						"name": "กล่อง",
						"isauto": false,
						"isdelete": false
					},
					{
						"code": "en",
						"name": "Box",
						"isauto": false,
						"isdelete": false
					},
					{
						"code": "ja",
						"name": "",
						"isauto": false,
						"isdelete": false
					}
				]
			}
		],		
		"paymentdetail": {
			"cashamounttext": "",
			"cashamount": 5,
			"paymentcreditcards": [],
			"paymenttransfers": []
		},
		"paymentdetailraw": "",
		"ispos": false,
		"couponno": "",
		"couponamount": 0,
		"coupondescription": "",
		"qrcode": "",
		"qrcodeamount": 0,
		"chequeno": "",
		"chequebooknumber": "",
		"chequebookcode": "",
		"chequeduedate": "",
		"chequeamount": 0
	}`

	phaser := saleinvoicereturn.SaleInvoiceReturnTransactionPhaser{}
	want := SaleInvoiceReturnTransactionStruct()

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

	assert.Equal(t, get.DebtorCode, want.DebtorCode, "creditorcode")
	assert.Equal(t, *get.DebtorNames[0].Name, *want.DebtorNames[0].Name, "creditorname")

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

	// payment
	assert.Equal(t, get.TotalPayCash, want.TotalPayCash, "totalpaycash")
	assert.Equal(t, get.TotalPayCredit, want.TotalPayCredit, "totalpaycredit")
	assert.Equal(t, get.TotalPayTransfer, want.TotalPayTransfer, "totalpaytransfer")

	// branch
	assert.Equal(t, "b01", get.BranchCode, "branchc code")
	assert.Equal(t, "สาขาที่ 1", *(get.BranchNames[0].Name), "branch name")
}
