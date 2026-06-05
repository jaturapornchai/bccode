package purchasereceive_test

import (
	"smlcloudplatform/internal/transaction/transactionconsumer/purchasereceive"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPurchaseReceiveStockPhaser(t *testing.T) {
	giveMsg := `{
	"id": "000000000000000000000000",
	"holdingcode": "30LbRx3l0SLaK84gLpcF0W4x9Z0",
	"guidfixed": "32Y3x1r0sqYOww0mRXrh0ftiIxV",
	"docno": "PP2025091100001",
	"docdatetime": "2025-09-11T09:44:52.81Z",
	"guid_ref": "5335d4d3-66ff-4b5d-a2c4-b9fb67e4db7c",
	"shiftdocno": "",
	"devicename": "",
	"guidpos": "",
	"transflag": 310,
	"docreftype": 0,
	"docreferences": [
		{
			"guidfixed": "32Y3jQOpI9qn0tKdvaI38LBdERe",
			"docno": "PO2025091100003",
			"docdatetime": "2025-09-11T09:41:11.668Z"
		}
	],
	"docrefno": "",
	"docrefdate": "2025-09-11T09:44:52.81Z",
	"taxdocdate": "2025-09-11T09:44:52.81Z",
	"taxdocno": "",
	"doc_type": 0,
	"imageurl": "",
	"inquirytype": 0,
	"vat_type": 0,
	"vatrate": 7,
	"custcode": "AP0004",
	"custnames": [
		{
			"code": "th",
			"name": "นาย ซื้อเยอะ ค้างรับ",
			"isauto": false,
			"isdelete": false
		}
	],
	"getpoint": 0,
	"usepoint": 0,
	"pointdiscountamount": 0,
	"description": "",
	"discountword": "0",
	"totaldiscount": 0,
	"totalvalue": 10,
	"totalexceptvat": 0,
	"totalaftervat": 10.7,
	"totalbeforevat": 10,
	"totalvatvalue": 0.7,
	"total_amount": 10.7,
	"total_cost": 0,
	"posid": "",
	"cashiercode": "",
	"salecode": "",
	"salename": "",
	"membercode": "",
	"iscancel": false,
	"ismanualamount": false,
	"status": 0,
	"paymentdetail": {
		"cashamounttext": "",
		"cashamount": 0,
		"paymentcreditcards": [],
		"paymenttransfers": []
	},
	"paymentdetailraw": "[]",
	"paycashamount": 0,
	"paypointamount": 0,
	"branch": {
		"guidfixed": "30LbS0SnNu5luOq9DWkLp3vvoZK",
		"code": "00000",
		"names": [
			{
				"code": "th",
				"name": "สำนักงานใหญ่",
				"isauto": false,
				"isdelete": false
			},
			{
				"code": "en",
				"name": "Head Office",
				"isauto": false,
				"isdelete": false
			}
		]
	},
	"billtaxtype": 0,
	"canceldatetime": "",
	"cancelusercode": "",
	"cancelusername": "",
	"canceldescription": "",
	"cancelreason": "",
	"fullvataddress": "",
	"fullvatbranchnumber": "",
	"fullvatname": "",
	"fullvatdocnumber": "",
	"fullvattaxid": "",
	"fullvatprint": false,
	"isvatregister": false,
	"isclose": false,
	"printcopybilldatetime": [],
	"tablenumber": "",
	"tableopendatetime": "",
	"tableclosedatetime": "",
	"mancount": 0,
	"womancount": 0,
	"childcount": 0,
	"istableallacratemode": false,
	"buffetcode": "",
	"customertelephone": "",
	"totalqty": 1,
	"totaldiscountvatamount": 0,
	"totaldiscountexceptvatamount": 0,
	"cashiername": "",
	"paycashchange": 0,
	"sumqrcode": 0,
	"sumcreditcard": 0,
	"summoneytransfer": 0,
	"sumcheque": 0,
	"sumcoupon": 0,
	"sumdeposit": 0,
	"sumadvancepayment": 0,
	"coupons": [],
	"totalcouponamount": 0,
	"coupondiscountamount": 0,
	"couponcashamount": 0,
	"detaildiscountformula": "0",
	"detailtotalamount": 10.7,
	"detailtotaldiscount": 0,
	"roundamount": 0,
	"totalamountafterdiscount": 10.7,
	"detailtotalamountbeforediscount": 0,
	"sumcredit": 0,
	"details": [
		{
			"inquirytype": 0,
			"line_number": 1,
			"docdatetime": "2025-09-11T09:44:52.81Z",
			"docref": "PO2025091100003",
			"docrefdatetime": "2025-09-11T09:41:11.668Z",
			"calcflag": 1,
			"barcode": "885002",
			"itemcode": "ITEM01",
			"itemnames": [
				{
					"code": "th",
					"name": "โค้ก",
					"isauto": false,
					"isdelete": false
				}
			],
			"unitcode": "PAC",
			"unitnames": [
				{
					"code": "th",
					"name": "แพ็ค",
					"isauto": false,
					"isdelete": false
				},
				{
					"code": "en",
					"name": "PACK",
					"isauto": false,
					"isdelete": false
				}
			],
			"item_type": 0,
			"item_guid": "30LblYDognbOaiUNd79MFGPdbI0",
			"imageurl": "",
			"description": "",
			"qty": 1,
			"eventqty": 2,
			"reason": "",
			"totalqty": 1,
			"price": 10,
			"discount": "",
			"discountamount": 0,
			"totalvaluevat": 0.7,
			"priceexcludevat": 10,
			"sum_amount": 10,
			"sumamountexcludevat": 10,
			"refguid": "",
			"dividevalue": 1,
			"standvalue": 1,
			"vat_type": 0,
			"remark": "",
			"multiunit": true,
			"issumpoint": false,
			"sumofcost": 0,
			"averagecost": 0,
			"foodtype": 0,
			"laststatus": 0,
			"ischoice": 0,
			"ispos": 0,
			"tax_type": 0,
			"vatcal": 0,
			"whcode": "00000",
			"whnames": [
				{
					"code": "th",
					"name": "สำนักงานใหญ่",
					"isauto": false,
					"isdelete": false
				},
				{
					"code": "en",
					"name": "Head Office",
					"isauto": false,
					"isdelete": false
				}
			],
			"shelfcode": "",
			"locationcode": "",
			"locationnames": [],
			"towhcode": "00000",
			"towhnames": [
				{
					"code": "th",
					"name": "สำนักงานใหญ่",
					"isauto": false,
					"isdelete": false
				},
				{
					"code": "en",
					"name": "Head Office",
					"isauto": false,
					"isdelete": false
				}
			],
			"tolocationcode": "",
			"tolocationnames": [],
			"sku": "",
			"extrajson": "",
			"group_code": "",
			"group_names": null,
			"manufacturerguid": "",
			"manufacturercode": "",
			"manufacturernames": null,
			"sumamountchoice": 0
		}
	]
}`

	stockPhaser := &purchasereceive.PurchaseReceiveTransactionStockPhaser{}
	phaser := &purchasereceive.PurchaseReceiveTransactionPhaser{}

	txnDoc, err := phaser.PhaseSingleDoc(giveMsg)
	assert.NoError(t, err)
	assert.NotNil(t, txnDoc)

	got, err := stockPhaser.PhaseSingleDoc(*txnDoc)
	assert.NoError(t, err)
	assert.NotNil(t, got)

	// Assert main transaction fields
	assert.Equal(t, "PP2025091100001", got.DocNo)
	assert.Equal(t, "30LbRx3l0SLaK84gLpcF0W4x9Z0", got.HoldingCode)
	assert.Equal(t, "32Y3x1r0sqYOww0mRXrh0ftiIxV", got.GuidFixed)
	assert.Equal(t, "5335d4d3-66ff-4b5d-a2c4-b9fb67e4db7c", got.GuidRef)
	expectedDocDate := time.Date(2025, 9, 11, 9, 44, 52, 810000000, time.UTC)
	assert.Equal(t, expectedDocDate, got.DocDate)
	expectedDocRefDate := time.Date(2025, 9, 11, 9, 44, 52, 810000000, time.UTC)
	assert.Equal(t, expectedDocRefDate, got.DocRefDate)
	assert.Equal(t, int16(310), got.TransFlag)
	assert.Equal(t, int8(0), got.DocRefType)
	assert.Equal(t, "", got.DocRefNo)
	assert.Equal(t, int(0), got.InquiryType)
	assert.Equal(t, int8(0), got.VatType)
	assert.Equal(t, 7.0, got.VatRate)
	assert.Equal(t, "", got.Description)
	assert.Equal(t, "0", got.DiscountWord)
	assert.Equal(t, 0.0, got.TotalDiscount)
	assert.Equal(t, 10.0, got.TotalValue)
	assert.Equal(t, 0.0, got.TotalExceptVat)
	assert.Equal(t, 10.7, got.TotalAfterVat)
	assert.Equal(t, 10.0, got.TotalBeforeVat)
	assert.Equal(t, 0.7, got.TotalVatValue)
	assert.Equal(t, 10.7, got.TotalAmount)
	assert.Equal(t, false, got.IsCancel)

	// Assert details array
	assert.NotNil(t, got.Details)
	assert.Len(t, *got.Details, 1)

	// Assert first detail item
	detail := (*got.Details)[0]
	assert.Equal(t, "PP2025091100001", detail.DocNo)
	assert.Equal(t, "30LbRx3l0SLaK84gLpcF0W4x9Z0", detail.HoldingCode)
	assert.Equal(t, "PO2025091100003", detail.DocRef)
	assert.Equal(t, "885002", detail.Barcode)
	assert.Equal(t, int8(0), detail.ItemType)
	assert.Equal(t, "30LblYDognbOaiUNd79MFGPdbI0", detail.ItemGuid)
	assert.Equal(t, int8(0), detail.VatType)
	assert.Equal(t, int8(0), detail.TaxType)
	assert.Equal(t, "PAC", detail.UnitCode)
	assert.Equal(t, 1.0, detail.StandValue)
	assert.Equal(t, 1.0, detail.DivideValue)
	assert.Equal(t, "00000", detail.WhCode)
	assert.Equal(t, "", detail.LocationCode)
	assert.Equal(t, 1.0, detail.Qty)
	assert.Equal(t, 10.0, detail.Price)
	assert.Equal(t, 10.0, detail.PriceExcludeVat)
	assert.Equal(t, 0.7, detail.TotalValueVat)
	assert.Equal(t, 10.0, detail.SumAmount)
	assert.Equal(t, 10.0, detail.SumAmountExcludeVat)
	assert.Equal(t, "", detail.Discount)
	assert.Equal(t, 0.0, detail.DiscountAmount)
	assert.Equal(t, int8(1), detail.CalcFlag)
	assert.Equal(t, int8(1), detail.LineNumber)

}
