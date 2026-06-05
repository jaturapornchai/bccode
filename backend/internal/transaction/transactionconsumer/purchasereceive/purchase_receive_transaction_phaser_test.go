package purchasereceive_test

import (
	"smlcloudplatform/internal/transaction/transactionconsumer/purchasereceive"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPurchaseReceiveTransactionPhaser(t *testing.T) {
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

	phaser := purchasereceive.PurchaseReceiveTransactionPhaser{}
	gotDoc, err := phaser.PhaseSingleDoc(giveMsg)

	// Basic test for no error and not nil
	assert.NoError(t, err)
	assert.NotNil(t, gotDoc)

	// Test header fields from TransactionPG
	assert.Equal(t, "32Y3x1r0sqYOww0mRXrh0ftiIxV", gotDoc.GuidFixed)
	assert.Equal(t, "30LbRx3l0SLaK84gLpcF0W4x9Z0", gotDoc.HoldingCode)
	assert.Equal(t, int16(310), gotDoc.TransFlag)
	assert.Equal(t, "PP2025091100001", gotDoc.DocNo)

	// Test parsed datetime
	expectedDocDate, _ := time.Parse("2006-01-02T15:04:05.999Z", "2025-09-11T09:44:52.81Z")
	assert.Equal(t, expectedDocDate, gotDoc.DocDate)

	assert.Equal(t, "5335d4d3-66ff-4b5d-a2c4-b9fb67e4db7c", gotDoc.GuidRef)
	assert.Equal(t, int8(0), gotDoc.DocRefType)
	assert.Equal(t, "", gotDoc.DocRefNo)

	// Test parsed doc ref date
	expectedDocRefDate, _ := time.Parse("2006-01-02T15:04:05.999Z", "2025-09-11T09:44:52.81Z")
	assert.Equal(t, expectedDocRefDate, gotDoc.DocRefDate)

	// Test branch information
	assert.Equal(t, "00000", gotDoc.BranchCode)

	// Test branch names (should be converted to proper format)
	branchNames := gotDoc.BranchNames
	assert.NotNil(t, branchNames)

	// Test tax document fields
	assert.Equal(t, "", gotDoc.TaxDocNo)
	expectedTaxDocDate, _ := time.Parse("2006-01-02T15:04:05.999Z", "2025-09-11T09:44:52.81Z")
	assert.Equal(t, expectedTaxDocDate, gotDoc.TaxDocDate)

	// Test transaction amounts
	assert.Equal(t, "", gotDoc.Description)
	assert.Equal(t, 0, gotDoc.InquiryType)
	assert.Equal(t, int8(0), gotDoc.VatType)
	assert.Equal(t, float64(7), gotDoc.VatRate)
	assert.Equal(t, "0", gotDoc.DiscountWord)
	assert.Equal(t, float64(0), gotDoc.TotalDiscount)
	assert.Equal(t, float64(10), gotDoc.TotalValue)
	assert.Equal(t, float64(0), gotDoc.TotalExceptVat)
	assert.Equal(t, float64(10.7), gotDoc.TotalAfterVat)
	assert.Equal(t, float64(10), gotDoc.TotalBeforeVat)
	assert.Equal(t, float64(0.7), gotDoc.TotalVatValue)
	assert.Equal(t, float64(10.7), gotDoc.TotalAmount)
	assert.Equal(t, false, gotDoc.IsCancel)

	// Test PurchaseReceiveTransactionPG specific fields
	assert.Equal(t, "AP0004", gotDoc.CreditorCode)

	// Test creditor names (should be converted to proper format)
	creditorNames := gotDoc.CreditorNames
	assert.NotNil(t, creditorNames)

	// Test items array
	assert.NotNil(t, gotDoc.Items)
	assert.Equal(t, 1, len(*gotDoc.Items))

	// Test first item details
	item := (*gotDoc.Items)[0]
	assert.Equal(t, "32Y3x1r0sqYOww0mRXrh0ftiIxV", item.GuidFixed)
	assert.Equal(t, "PO2025091100003", item.DocRef)

	// Test item doc ref datetime
	expectedItemDocRefDate, _ := time.Parse("2006-01-02T15:04:05.999Z", "2025-09-11T09:41:11.668Z")
	assert.Equal(t, expectedItemDocRefDate, item.DocRefDateTime)

	assert.Equal(t, "PP2025091100001", item.DocNo)
	assert.Equal(t, "30LbRx3l0SLaK84gLpcF0W4x9Z0", item.HoldingCode)
	assert.Equal(t, int8(1), item.LineNumber)
	assert.Equal(t, "885002", item.Barcode)
	assert.Equal(t, float64(1), item.Qty)
	assert.Equal(t, float64(10), item.Price)
	assert.Equal(t, float64(10), item.PriceExcludeVat)
	assert.Equal(t, "", item.Discount)
	assert.Equal(t, float64(0), item.DiscountAmount)
	assert.Equal(t, float64(10), item.SumAmount)
	assert.Equal(t, float64(10), item.SumAmountExcludeVat)
	assert.Equal(t, float64(0.7), item.TotalValueVat)
	assert.Equal(t, "00000", item.WhCode)
	assert.Equal(t, "", item.LocationCode)
	assert.Equal(t, int8(0), item.VatType)
	assert.Equal(t, int8(0), item.TaxType)
	assert.Equal(t, float64(1), item.StandValue)
	assert.Equal(t, float64(1), item.DivideValue)
	assert.Equal(t, int8(0), item.ItemType)
	assert.Equal(t, "30LblYDognbOaiUNd79MFGPdbI0", item.ItemGuid)
	assert.Equal(t, "", item.Remark)
	assert.Equal(t, "PAC", item.UnitCode)
	assert.Equal(t, "", item.GroupCode)

	// Test item datetime
	expectedItemDocDate, _ := time.Parse("2006-01-02T15:04:05.999Z", "2025-09-11T09:44:52.81Z")
	assert.Equal(t, expectedItemDocDate, item.DocDate)

	// Test item names arrays (should be converted to proper format)
	assert.NotNil(t, item.ItemNames)
	assert.NotNil(t, item.UnitNames)
	assert.NotNil(t, item.WhNames)
	assert.NotNil(t, item.LocationNames)
	assert.NotNil(t, item.GroupNames)
}

func TestPurchaseReceiveTransactionPhaser_InvalidJSON(t *testing.T) {
	invalidJSONMsg := `{
		"id": "000000000000000000000000",
		"holdingcode": "30LbRx3l0SLaK84gLpcF0W4x9Z0",
		"guidfixed": "invalid json structure`

	phaser := purchasereceive.PurchaseReceiveTransactionPhaser{}
	gotDoc, err := phaser.PhaseSingleDoc(invalidJSONMsg)

	assert.Error(t, err)
	assert.Nil(t, gotDoc)
	assert.Contains(t, err.Error(), "Cannot Unmarshal PurchaseDoc Message")
}

func TestPurchaseReceiveTransactionPhaser_EmptyDetails(t *testing.T) {
	emptyDetailsMsg := `{
		"id": "000000000000000000000000",
		"holdingcode": "30LbRx3l0SLaK84gLpcF0W4x9Z0",
		"guidfixed": "32Y3x1r0sqYOww0mRXrh0ftiIxV",
		"docno": "PP2025091100001",
		"docdatetime": "2025-09-11T09:44:52.81Z",
		"guid_ref": "5335d4d3-66ff-4b5d-a2c4-b9fb67e4db7c",
		"transflag": 310,
		"docreftype": 0,
		"docrefno": "",
		"docrefdate": "2025-09-11T09:44:52.81Z",
		"taxdocdate": "2025-09-11T09:44:52.81Z",
		"taxdocno": "",
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
		"description": "",
		"discountword": "0",
		"totaldiscount": 0,
		"totalvalue": 0,
		"totalexceptvat": 0,
		"totalaftervat": 0,
		"totalbeforevat": 0,
		"totalvatvalue": 0,
		"total_amount": 0,
		"iscancel": false,
		"ismanualamount": false,
		"branch": {
			"guidfixed": "30LbS0SnNu5luOq9DWkLp3vvoZK",
			"code": "00000",
			"names": [
				{
					"code": "th",
					"name": "สำนักงานใหญ่",
					"isauto": false,
					"isdelete": false
				}
			]
		},
		"details": []
	}`

	phaser := purchasereceive.PurchaseReceiveTransactionPhaser{}
	gotDoc, err := phaser.PhaseSingleDoc(emptyDetailsMsg)

	assert.NoError(t, err)
	assert.NotNil(t, gotDoc)
	assert.Equal(t, "PP2025091100001", gotDoc.DocNo)
	assert.Equal(t, "AP0004", gotDoc.CreditorCode)
	assert.NotNil(t, gotDoc.Items)
	assert.Equal(t, 0, len(*gotDoc.Items))
}

func TestPurchaseReceiveTransactionPhaser_MultipleItems(t *testing.T) {
	multipleItemsMsg := `{
		"id": "000000000000000000000000",
		"holdingcode": "30LbRx3l0SLaK84gLpcF0W4x9Z0",
		"guidfixed": "32Y3x1r0sqYOww0mRXrh0ftiIxV",
		"docno": "PP2025091100001",
		"docdatetime": "2025-09-11T09:44:52.81Z",
		"guid_ref": "5335d4d3-66ff-4b5d-a2c4-b9fb67e4db7c",
		"transflag": 310,
		"docreftype": 0,
		"docrefno": "",
		"docrefdate": "2025-09-11T09:44:52.81Z",
		"taxdocdate": "2025-09-11T09:44:52.81Z",
		"taxdocno": "",
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
		"description": "",
		"discountword": "0",
		"totaldiscount": 0,
		"totalvalue": 25,
		"totalexceptvat": 0,
		"totalaftervat": 26.75,
		"totalbeforevat": 25,
		"totalvatvalue": 1.75,
		"total_amount": 26.75,
		"iscancel": false,
		"ismanualamount": false,
		"branch": {
			"guidfixed": "30LbS0SnNu5luOq9DWkLp3vvoZK",
			"code": "00000",
			"names": [
				{
					"code": "th",
					"name": "สำนักงานใหญ่",
					"isauto": false,
					"isdelete": false
				}
			]
		},
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
					}
				],
				"item_type": 0,
				"item_guid": "30LblYDognbOaiUNd79MFGPdbI0",
				"qty": 1,
				"price": 10,
				"discount": "",
				"discountamount": 0,
				"totalvaluevat": 0.7,
				"priceexcludevat": 10,
				"sum_amount": 10,
				"sumamountexcludevat": 10,
				"dividevalue": 1,
				"standvalue": 1,
				"vat_type": 0,
				"remark": "",
				"tax_type": 0,
				"whcode": "00000",
				"whnames": [
					{
						"code": "th",
						"name": "สำนักงานใหญ่",
						"isauto": false,
						"isdelete": false
					}
				],
				"locationcode": "",
				"locationnames": [],
				"group_code": "",
				"group_names": null,
				"sumamountchoice": 0
			},
			{
				"inquirytype": 0,
				"line_number": 2,
				"docdatetime": "2025-09-11T09:44:52.81Z",
				"docref": "PO2025091100003",
				"docrefdatetime": "2025-09-11T09:41:11.668Z",
				"calcflag": 1,
				"barcode": "885003",
				"itemcode": "ITEM02",
				"itemnames": [
					{
						"code": "th",
						"name": "เป๊ปซี่",
						"isauto": false,
						"isdelete": false
					}
				],
				"unitcode": "BTL",
				"unitnames": [
					{
						"code": "th",
						"name": "ขวด",
						"isauto": false,
						"isdelete": false
					}
				],
				"item_type": 0,
				"item_guid": "30LblYDognbOaiUNd79MFGPdbI1",
				"qty": 1,
				"price": 15,
				"discount": "",
				"discountamount": 0,
				"totalvaluevat": 1.05,
				"priceexcludevat": 15,
				"sum_amount": 15,
				"sumamountexcludevat": 15,
				"dividevalue": 1,
				"standvalue": 1,
				"vat_type": 0,
				"remark": "",
				"tax_type": 0,
				"whcode": "00000",
				"whnames": [
					{
						"code": "th",
						"name": "สำนักงานใหญ่",
						"isauto": false,
						"isdelete": false
					}
				],
				"locationcode": "",
				"locationnames": [],
				"group_code": "",
				"group_names": null,
				"sumamountchoice": 0
			}
		]
	}`

	phaser := purchasereceive.PurchaseReceiveTransactionPhaser{}
	gotDoc, err := phaser.PhaseSingleDoc(multipleItemsMsg)

	assert.NoError(t, err)
	assert.NotNil(t, gotDoc)
	assert.Equal(t, "PP2025091100001", gotDoc.DocNo)
	assert.Equal(t, "AP0004", gotDoc.CreditorCode)

	// Test multiple items
	assert.NotNil(t, gotDoc.Items)
	assert.Equal(t, 2, len(*gotDoc.Items))

	// Test first item
	item1 := (*gotDoc.Items)[0]
	assert.Equal(t, int8(1), item1.LineNumber)
	assert.Equal(t, "885002", item1.Barcode)
	assert.Equal(t, "30LblYDognbOaiUNd79MFGPdbI0", item1.ItemGuid)
	assert.Equal(t, "PAC", item1.UnitCode)
	assert.Equal(t, float64(1), item1.Qty)
	assert.Equal(t, float64(10), item1.Price)
	assert.Equal(t, float64(10), item1.SumAmount)

	// Test second item
	item2 := (*gotDoc.Items)[1]
	assert.Equal(t, int8(2), item2.LineNumber)
	assert.Equal(t, "885003", item2.Barcode)
	assert.Equal(t, "30LblYDognbOaiUNd79MFGPdbI1", item2.ItemGuid)
	assert.Equal(t, "BTL", item2.UnitCode)
	assert.Equal(t, float64(1), item2.Qty)
	assert.Equal(t, float64(15), item2.Price)
	assert.Equal(t, float64(15), item2.SumAmount)

	// Test total amounts with multiple items
	assert.Equal(t, float64(25), gotDoc.TotalValue)
	assert.Equal(t, float64(26.75), gotDoc.TotalAmount)
	assert.Equal(t, float64(1.75), gotDoc.TotalVatValue)
}

func TestPurchaseReceiveTransactionPhaser_DateTimeParsing(t *testing.T) {
	dateTimeMsg := `{
		"id": "000000000000000000000000",
		"holdingcode": "30LbRx3l0SLaK84gLpcF0W4x9Z0",
		"guidfixed": "32Y3x1r0sqYOww0mRXrh0ftiIxV",
		"docno": "PP2025091100001",
		"docdatetime": "2025-12-31T23:59:59.999Z",
		"guid_ref": "5335d4d3-66ff-4b5d-a2c4-b9fb67e4db7c",
		"transflag": 310,
		"docreftype": 0,
		"docrefno": "",
		"docrefdate": "2025-01-01T00:00:00.000Z",
		"taxdocdate": "2025-06-15T12:30:45.123Z",
		"taxdocno": "TAX001",
		"inquirytype": 0,
		"vat_type": 0,
		"vatrate": 7,
		"custcode": "AP0004",
		"custnames": [
			{
				"code": "th",
				"name": "Test Customer",
				"isauto": false,
				"isdelete": false
			}
		],
		"description": "Date Time Test",
		"discountword": "0",
		"totaldiscount": 0,
		"totalvalue": 100,
		"totalexceptvat": 0,
		"totalaftervat": 107,
		"totalbeforevat": 100,
		"totalvatvalue": 7,
		"total_amount": 107,
		"iscancel": false,
		"ismanualamount": false,
		"branch": {
			"guidfixed": "30LbS0SnNu5luOq9DWkLp3vvoZK",
			"code": "00000",
			"names": [
				{
					"code": "th",
					"name": "สำนักงานใหญ่",
					"isauto": false,
					"isdelete": false
				}
			]
		},
		"details": [
			{
				"inquirytype": 0,
				"line_number": 1,
				"docdatetime": "2025-03-15T14:20:30.456Z",
				"docref": "PO2025091100003",
				"docrefdatetime": "2025-02-28T08:15:22.789Z",
				"calcflag": 1,
				"barcode": "885002",
				"itemcode": "ITEM01",
				"itemnames": [
					{
						"code": "th",
						"name": "Test Item",
						"isauto": false,
						"isdelete": false
					}
				],
				"unitcode": "PCS",
				"unitnames": [
					{
						"code": "th",
						"name": "ชิ้น",
						"isauto": false,
						"isdelete": false
					}
				],
				"item_type": 0,
				"item_guid": "30LblYDognbOaiUNd79MFGPdbI0",
				"qty": 1,
				"price": 100,
				"discount": "",
				"discountamount": 0,
				"totalvaluevat": 7,
				"priceexcludevat": 100,
				"sum_amount": 100,
				"sumamountexcludevat": 100,
				"dividevalue": 1,
				"standvalue": 1,
				"vat_type": 0,
				"remark": "DateTime test item",
				"tax_type": 0,
				"whcode": "00000",
				"whnames": [
					{
						"code": "th",
						"name": "สำนักงานใหญ่",
						"isauto": false,
						"isdelete": false
					}
				],
				"locationcode": "",
				"locationnames": [],
				"group_code": "",
				"group_names": null,
				"sumamountchoice": 0
			}
		]
	}`

	phaser := purchasereceive.PurchaseReceiveTransactionPhaser{}
	gotDoc, err := phaser.PhaseSingleDoc(dateTimeMsg)

	assert.NoError(t, err)
	assert.NotNil(t, gotDoc)

	// Test various date/time parsing scenarios

	// Test document date - end of year
	expectedDocDate, _ := time.Parse("2006-01-02T15:04:05.999Z", "2025-12-31T23:59:59.999Z")
	assert.Equal(t, expectedDocDate, gotDoc.DocDate)

	// Test document reference date - start of year
	expectedDocRefDate, _ := time.Parse("2006-01-02T15:04:05.999Z", "2025-01-01T00:00:00.000Z")
	assert.Equal(t, expectedDocRefDate, gotDoc.DocRefDate)

	// Test tax document date - middle of year with milliseconds
	expectedTaxDocDate, _ := time.Parse("2006-01-02T15:04:05.999Z", "2025-06-15T12:30:45.123Z")
	assert.Equal(t, expectedTaxDocDate, gotDoc.TaxDocDate)

	// Test item level date/time parsing
	item := (*gotDoc.Items)[0]

	// Test item document date
	expectedItemDocDate, _ := time.Parse("2006-01-02T15:04:05.999Z", "2025-03-15T14:20:30.456Z")
	assert.Equal(t, expectedItemDocDate, item.DocDate)

	// Test item document reference date
	expectedItemDocRefDate, _ := time.Parse("2006-01-02T15:04:05.999Z", "2025-02-28T08:15:22.789Z")
	assert.Equal(t, expectedItemDocRefDate, item.DocRefDateTime)

	// Test other fields are still parsed correctly
	assert.Equal(t, "Date Time Test", gotDoc.Description)
	assert.Equal(t, "TAX001", gotDoc.TaxDocNo)
	assert.Equal(t, "DateTime test item", item.Remark)
}
