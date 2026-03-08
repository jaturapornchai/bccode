package purchaseorder_test

import (
	"smlcloudplatform/internal/transaction/transactionconsumer/purchaseorder"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPhasePurchaseOrderMsg(t *testing.T) {

	giveMsg := `{
	"id": "000000000000000000000000",
	"shopid": "30LbRx3l0SLaK84gLpcF0W4x9Z0",
	"guidfixed": "32sPNAFfVGHiVjgN6amwMr0i1mS",
	"docno": "PO2025091800001",
	"docdatetime": "2025-09-18T14:37:15.211Z",
	"guidref": "3e52d89b-0621-4c4b-a5f2-ae1e950fe710",
	"shiftdocno": "",
	"devicename": "",
	"guidpos": "",
	"transflag": 6,
	"docreftype": 0,
	"docrefno": "",
	"docrefdate": "2025-09-18T14:37:15.211Z",
	"taxdocdate": "2025-09-18T14:37:15.211Z",
	"taxdocno": "",
	"doctype": 0,
	"imageurl": "",
	"inquirytype": 0,
	"vattype": 0,
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
	"totalqty": 4,
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
	"detailtotalamount": 0,
	"detailtotaldiscount": 0,
	"roundamount": 0,
	"totalamountafterdiscount": 0,
	"detailtotalamountbeforediscount": 0,
	"sumcredit": 0,
	"details": [
		{
			"inquirytype": 0,
			"linenumber": 1,
			"docdatetime": "2025-09-18T14:37:15.211Z",
			"docref": "",
			"docrefdatetime": "2025-09-18T14:37:24.505Z",
			"calcflag": 1,
			"barcode": "885001",
			"itemcode": "ITEM01",
			"itemnames": [
				{
					"code": "th",
					"name": "โค้ก",
					"isauto": false,
					"isdelete": false
				}
			],
			"unitcode": "CN",
			"unitnames": [
				{
					"code": "th",
					"name": "กระป๋อง",
					"isauto": false,
					"isdelete": false
				},
				{
					"code": "en",
					"name": "CAN",
					"isauto": false,
					"isdelete": false
				}
			],
			"itemtype": 0,
			"itemguid": "30LbfsvYmdlp5hzcz8qznVjzArZ",
			"imageurl": "",
			"description": "",
			"qty": 1,
			"eventqty": 0,
			"reason": "",
			"totalqty": 1,
			"price": 0,
			"discount": "",
			"discountamount": 0,
			"totalvaluevat": 0,
			"priceexcludevat": 0,
			"sumamount": 0,
			"sumamountexcludevat": 0,
			"refguid": "",
			"dividevalue": 1,
			"standvalue": 1,
			"vattype": 0,
			"remark": "",
			"multiunit": true,
			"issumpoint": false,
			"sumofcost": 0,
			"averagecost": 0,
			"foodtype": 0,
			"laststatus": 0,
			"ischoice": 0,
			"ispos": 0,
			"taxtype": 0,
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
			"locationcode": "TEST1",
			"locationnames": [
				{
					"code": "th",
					"name": "test",
					"isauto": false,
					"isdelete": false
				}
			],
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
			"tolocationcode": "TEST1",
			"tolocationnames": [
				{
					"code": "th",
					"name": "test",
					"isauto": false,
					"isdelete": false
				}
			],
			"sku": "",
			"extrajson": "",
			"groupcode": "",
			"groupnames": null,
			"manufacturerguid": "",
			"manufacturercode": "",
			"manufacturernames": null,
			"sumamountchoice": 0
		},
		{
			"inquirytype": 0,
			"linenumber": 2,
			"docdatetime": "2025-09-18T14:37:15.211Z",
			"docref": "",
			"docrefdatetime": "2025-09-18T14:37:32.557Z",
			"calcflag": 1,
			"barcode": "889001",
			"itemcode": "ITEM03",
			"itemnames": [
				{
					"code": "th",
					"name": "สินค้ายกเว้นภาษี",
					"isauto": false,
					"isdelete": false
				}
			],
			"unitcode": "TIM",
			"unitnames": [
				{
					"code": "th",
					"name": "ครั้ง",
					"isauto": false,
					"isdelete": false
				},
				{
					"code": "en",
					"name": "TIME",
					"isauto": false,
					"isdelete": false
				}
			],
			"itemtype": 0,
			"itemguid": "30rUdC5X1IKUpHWMcy1tBdb2jy8",
			"imageurl": "",
			"description": "",
			"qty": 1,
			"eventqty": 0,
			"reason": "",
			"totalqty": 1,
			"price": 0,
			"discount": "",
			"discountamount": 0,
			"totalvaluevat": 0,
			"priceexcludevat": 0,
			"sumamount": 0,
			"sumamountexcludevat": 0,
			"refguid": "",
			"dividevalue": 1,
			"standvalue": 1,
			"vattype": 0,
			"remark": "",
			"multiunit": true,
			"issumpoint": false,
			"sumofcost": 0,
			"averagecost": 0,
			"foodtype": 0,
			"laststatus": 0,
			"ischoice": 0,
			"ispos": 0,
			"taxtype": 0,
			"vatcal": 1,
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
			"locationcode": "TEST1",
			"locationnames": [
				{
					"code": "th",
					"name": "test",
					"isauto": false,
					"isdelete": false
				}
			],
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
			"tolocationcode": "TEST1",
			"tolocationnames": [
				{
					"code": "th",
					"name": "test",
					"isauto": false,
					"isdelete": false
				}
			],
			"sku": "",
			"extrajson": "",
			"groupcode": "",
			"groupnames": null,
			"manufacturerguid": "",
			"manufacturercode": "",
			"manufacturernames": null,
			"sumamountchoice": 0
		},
		{
			"inquirytype": 0,
			"linenumber": 3,
			"docdatetime": "2025-09-18T14:37:15.211Z",
			"docref": "",
			"docrefdatetime": "2025-09-18T14:37:36.398Z",
			"calcflag": 1,
			"barcode": "885006",
			"itemcode": "",
			"itemnames": [
				{
					"code": "th",
					"name": "ขนม",
					"isauto": false,
					"isdelete": false
				}
			],
			"unitcode": "BAG",
			"unitnames": [
				{
					"code": "th",
					"name": "ถุง",
					"isauto": false,
					"isdelete": false
				},
				{
					"code": "en",
					"name": "BAG",
					"isauto": false,
					"isdelete": false
				}
			],
			"itemtype": 0,
			"itemguid": "32XOdHh35AEvNhf1Am2S4fcpjy4",
			"imageurl": "",
			"description": "",
			"qty": 1,
			"eventqty": 0,
			"reason": "",
			"totalqty": 1,
			"price": 0,
			"discount": "",
			"discountamount": 0,
			"totalvaluevat": 0,
			"priceexcludevat": 0,
			"sumamount": 0,
			"sumamountexcludevat": 0,
			"refguid": "",
			"dividevalue": 1,
			"standvalue": 1,
			"vattype": 0,
			"remark": "",
			"multiunit": true,
			"issumpoint": false,
			"sumofcost": 0,
			"averagecost": 0,
			"foodtype": 0,
			"laststatus": 0,
			"ischoice": 0,
			"ispos": 0,
			"taxtype": 0,
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
			"locationcode": "TEST1",
			"locationnames": [
				{
					"code": "th",
					"name": "test",
					"isauto": false,
					"isdelete": false
				}
			],
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
			"tolocationcode": "TEST1",
			"tolocationnames": [
				{
					"code": "th",
					"name": "test",
					"isauto": false,
					"isdelete": false
				}
			],
			"sku": "",
			"extrajson": "",
			"groupcode": "",
			"groupnames": null,
			"manufacturerguid": "",
			"manufacturercode": "",
			"manufacturernames": null,
			"sumamountchoice": 0
		},
		{
			"inquirytype": 0,
			"linenumber": 4,
			"docdatetime": "2025-09-18T14:37:15.211Z",
			"docref": "",
			"docrefdatetime": "2025-09-18T14:37:39.036Z",
			"calcflag": 1,
			"barcode": "885007",
			"itemcode": "",
			"itemnames": [
				{
					"code": "th",
					"name": "เค้ก",
					"isauto": false,
					"isdelete": false
				}
			],
			"unitcode": "LMP",
			"unitnames": [
				{
					"code": "th",
					"name": "ก้อน",
					"isauto": false,
					"isdelete": false
				},
				{
					"code": "en",
					"name": "",
					"isauto": false,
					"isdelete": false
				}
			],
			"itemtype": 0,
			"itemguid": "32XOiH1Yhuzus2StYicM1rtnlSj",
			"imageurl": "",
			"description": "",
			"qty": 1,
			"eventqty": 0,
			"reason": "",
			"totalqty": 1,
			"price": 0,
			"discount": "",
			"discountamount": 0,
			"totalvaluevat": 0,
			"priceexcludevat": 0,
			"sumamount": 0,
			"sumamountexcludevat": 0,
			"refguid": "",
			"dividevalue": 1,
			"standvalue": 1,
			"vattype": 0,
			"remark": "",
			"multiunit": true,
			"issumpoint": false,
			"sumofcost": 0,
			"averagecost": 0,
			"foodtype": 0,
			"laststatus": 0,
			"ischoice": 0,
			"ispos": 0,
			"taxtype": 0,
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
			"locationcode": "TEST1",
			"locationnames": [
				{
					"code": "th",
					"name": "test",
					"isauto": false,
					"isdelete": false
				}
			],
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
			"tolocationcode": "TEST1",
			"tolocationnames": [
				{
					"code": "th",
					"name": "test",
					"isauto": false,
					"isdelete": false
				}
			],
			"sku": "",
			"extrajson": "",
			"groupcode": "",
			"groupnames": null,
			"manufacturerguid": "",
			"manufacturercode": "",
			"manufacturernames": null,
			"sumamountchoice": 0
		}
	]
}`

	txnPhaser := &purchaseorder.PurchaseOrderTransactionPhaser{}

	got, err := txnPhaser.PhaseSingleDoc(giveMsg)

	assert.NoError(t, err)
	assert.NotNil(t, got)

	// Main transaction fields
	assert.Equal(t, "32sPNAFfVGHiVjgN6amwMr0i1mS", got.GuidFixed)
	assert.Equal(t, "30LbRx3l0SLaK84gLpcF0W4x9Z0", got.ShopID)
	assert.Equal(t, int16(6), got.TransFlag)
	assert.Equal(t, "PO2025091800001", got.DocNo)
	assert.Equal(t, "3e52d89b-0621-4c4b-a5f2-ae1e950fe710", got.GuidRef)
	assert.Equal(t, int8(0), got.DocRefType)
	assert.Equal(t, "", got.DocRefNo)
	assert.Equal(t, "", got.TaxDocNo)
	assert.Equal(t, "", got.Description)
	assert.Equal(t, 0, got.InquiryType)
	assert.Equal(t, int8(0), got.VatType)
	assert.Equal(t, 7.0, got.VatRate)
	assert.Equal(t, "0", got.DiscountWord)
	assert.Equal(t, 0.0, got.TotalDiscount)
	assert.Equal(t, 0.0, got.TotalValue)
	assert.Equal(t, 0.0, got.TotalExceptVat)
	assert.Equal(t, 0.0, got.TotalAfterVat)
	assert.Equal(t, 0.0, got.TotalBeforeVat)
	assert.Equal(t, 0.0, got.TotalVatValue)
	assert.Equal(t, 0.0, got.TotalAmount)
	assert.Equal(t, false, got.IsCancel)

	// Branch information
	assert.Equal(t, "00000", got.BranchCode)
	assert.NotNil(t, got.BranchNames)
	assert.Len(t, got.BranchNames, 2)

	// Check first branch name (Thai)
	assert.Equal(t, "th", *got.BranchNames[0].Code)
	assert.Equal(t, "สำนักงานใหญ่", *got.BranchNames[0].Name)

	// Check second branch name (English)
	assert.Equal(t, "en", *got.BranchNames[1].Code)
	assert.Equal(t, "Head Office", *got.BranchNames[1].Name)

	// Creditor information
	assert.Equal(t, "AP0004", got.CreditorCode)
	assert.NotNil(t, got.CreditorNames)
	assert.Len(t, got.CreditorNames, 1)
	assert.Equal(t, "th", *got.CreditorNames[0].Code)
	assert.Equal(t, "นาย ซื้อเยอะ ค้างรับ", *got.CreditorNames[0].Name)

	// Items
	assert.NotNil(t, got.Items)
	assert.Len(t, *got.Items, 4)

	// First item assertions
	item1 := (*got.Items)[0]
	assert.Equal(t, "32sPNAFfVGHiVjgN6amwMr0i1mS", item1.GuidFixed)
	assert.Equal(t, "PO2025091800001", item1.DocNo)
	assert.Equal(t, "30LbRx3l0SLaK84gLpcF0W4x9Z0", item1.ShopID)
	assert.Equal(t, int8(1), item1.LineNumber)
	assert.Equal(t, "885001", item1.Barcode)
	assert.Equal(t, "30LbfsvYmdlp5hzcz8qznVjzArZ", item1.ItemGuid)
	assert.Equal(t, "CN", item1.UnitCode)
	assert.Equal(t, 1.0, item1.Qty)
	assert.Equal(t, 0.0, item1.Price)
	assert.Equal(t, "", item1.Discount)
	assert.Equal(t, 0.0, item1.DiscountAmount)
	assert.Equal(t, 0.0, item1.SumAmount)
	assert.Equal(t, 0.0, item1.SumAmountExcludeVat)
	assert.Equal(t, 0.0, item1.TotalValueVat)
	assert.Equal(t, 0.0, item1.PriceExcludeVat)
	assert.Equal(t, "00000", item1.WhCode)
	assert.Equal(t, "TEST1", item1.LocationCode)
	assert.Equal(t, int8(0), item1.VatType)
	assert.Equal(t, int8(0), item1.TaxType)
	assert.Equal(t, 1.0, item1.StandValue)
	assert.Equal(t, 1.0, item1.DivideValue)
	assert.Equal(t, int8(0), item1.ItemType)
	assert.Equal(t, "", item1.Remark)

	// Check first item names
	assert.NotNil(t, item1.ItemNames)
	assert.Len(t, item1.ItemNames, 1)
	assert.Equal(t, "th", *item1.ItemNames[0].Code)
	assert.Equal(t, "โค้ก", *item1.ItemNames[0].Name)

	// Check first item unit names
	assert.NotNil(t, item1.UnitNames)
	assert.Len(t, item1.UnitNames, 2)
	assert.Equal(t, "th", *item1.UnitNames[0].Code)
	assert.Equal(t, "กระป๋อง", *item1.UnitNames[0].Name)
	assert.Equal(t, "en", *item1.UnitNames[1].Code)
	assert.Equal(t, "CAN", *item1.UnitNames[1].Name)

	// Second item assertions (key fields)
	item2 := (*got.Items)[1]
	assert.Equal(t, int8(2), item2.LineNumber)
	assert.Equal(t, "889001", item2.Barcode)
	assert.Equal(t, "30rUdC5X1IKUpHWMcy1tBdb2jy8", item2.ItemGuid)
	assert.Equal(t, "TIM", item2.UnitCode)
	assert.Equal(t, 1.0, item2.Qty)
	assert.Equal(t, int8(1), item2.VatCal)

	// Check second item names
	assert.NotNil(t, item2.ItemNames)
	assert.Len(t, item2.ItemNames, 1)
	assert.Equal(t, "th", *item2.ItemNames[0].Code)
	assert.Equal(t, "สินค้ายกเว้นภาษี", *item2.ItemNames[0].Name)

	// Third item assertions (key fields)
	item3 := (*got.Items)[2]
	assert.Equal(t, int8(3), item3.LineNumber)
	assert.Equal(t, "885006", item3.Barcode)
	assert.Equal(t, "32XOdHh35AEvNhf1Am2S4fcpjy4", item3.ItemGuid)
	assert.Equal(t, "BAG", item3.UnitCode)
	assert.Equal(t, 1.0, item3.Qty)
	assert.Equal(t, int8(0), item3.VatCal)

	// Check third item names
	assert.NotNil(t, item3.ItemNames)
	assert.Len(t, item3.ItemNames, 1)
	assert.Equal(t, "th", *item3.ItemNames[0].Code)
	assert.Equal(t, "ขนม", *item3.ItemNames[0].Name)

	// Fourth item assertions (key fields)
	item4 := (*got.Items)[3]
	assert.Equal(t, int8(4), item4.LineNumber)
	assert.Equal(t, "885007", item4.Barcode)
	assert.Equal(t, "32XOiH1Yhuzus2StYicM1rtnlSj", item4.ItemGuid)
	assert.Equal(t, "LMP", item4.UnitCode)
	assert.Equal(t, 1.0, item4.Qty)
	assert.Equal(t, int8(0), item4.VatCal)

	// Check fourth item names
	assert.NotNil(t, item4.ItemNames)
	assert.Len(t, item4.ItemNames, 1)
	assert.Equal(t, "th", *item4.ItemNames[0].Code)
	assert.Equal(t, "เค้ก", *item4.ItemNames[0].Name)

	// Additional warehouse and location assertions for all items
	for i, item := range *got.Items {
		assert.Equal(t, "00000", item.WhCode, "Item %d WhCode", i+1)
		assert.NotNil(t, item.WhNames, "Item %d WhNames", i+1)
		assert.Len(t, item.WhNames, 2, "Item %d WhNames length", i+1)
		assert.Equal(t, "th", *item.WhNames[0].Code, "Item %d WhNames[0].Code", i+1)
		assert.Equal(t, "สำนักงานใหญ่", *item.WhNames[0].Name, "Item %d WhNames[0].Name", i+1)
		assert.Equal(t, "en", *item.WhNames[1].Code, "Item %d WhNames[1].Code", i+1)
		assert.Equal(t, "Head Office", *item.WhNames[1].Name, "Item %d WhNames[1].Name", i+1)

		assert.Equal(t, "TEST1", item.LocationCode, "Item %d LocationCode", i+1)
		assert.NotNil(t, item.LocationNames, "Item %d LocationNames", i+1)
		assert.Len(t, item.LocationNames, 1, "Item %d LocationNames length", i+1)
		assert.Equal(t, "th", *item.LocationNames[0].Code, "Item %d LocationNames[0].Code", i+1)
		assert.Equal(t, "test", *item.LocationNames[0].Name, "Item %d LocationNames[0].Name", i+1)

		// Common item assertions
		assert.Equal(t, got.GuidFixed, item.GuidFixed, "Item %d GuidFixed", i+1)
		assert.Equal(t, got.DocNo, item.DocNo, "Item %d DocNo", i+1)
		assert.Equal(t, got.ShopID, item.ShopID, "Item %d ShopID", i+1)
		assert.Equal(t, 0.0, item.Price, "Item %d Price", i+1)
		assert.Equal(t, "", item.Discount, "Item %d Discount", i+1)
		assert.Equal(t, 0.0, item.DiscountAmount, "Item %d DiscountAmount", i+1)
		assert.Equal(t, 0.0, item.SumAmount, "Item %d SumAmount", i+1)
		assert.Equal(t, 0.0, item.SumAmountExcludeVat, "Item %d SumAmountExcludeVat", i+1)
		assert.Equal(t, 0.0, item.TotalValueVat, "Item %d TotalValueVat", i+1)
		assert.Equal(t, 0.0, item.PriceExcludeVat, "Item %d PriceExcludeVat", i+1)
		assert.Equal(t, int8(0), item.VatType, "Item %d VatType", i+1)
		assert.Equal(t, int8(0), item.TaxType, "Item %d TaxType", i+1)
		assert.Equal(t, 1.0, item.StandValue, "Item %d StandValue", i+1)
		assert.Equal(t, 1.0, item.DivideValue, "Item %d DivideValue", i+1)
		assert.Equal(t, int8(0), item.ItemType, "Item %d ItemType", i+1)
		assert.Equal(t, "", item.Remark, "Item %d Remark", i+1)
	}

	// Verify specific DocRef and DocRefDateTime for each item (from giveMsg)
	assert.Equal(t, "", (*got.Items)[0].DocRef)
	assert.Equal(t, "", (*got.Items)[1].DocRef)
	assert.Equal(t, "", (*got.Items)[2].DocRef)
	assert.Equal(t, "", (*got.Items)[3].DocRef)

}
