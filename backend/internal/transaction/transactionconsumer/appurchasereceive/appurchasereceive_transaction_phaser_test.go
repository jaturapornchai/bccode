package appurchasereceive

import (
	"strings"
	"testing"

	"smlcloudplatform/internal/transaction/models"
)

// TestAPPurchaseReceiveTransactionPhaser_PhaseSingleDoc_Success tests the successful conversion
// of a valid JSON message with details to an AccrualReceiveTransactionPG struct.
// This test verifies the happy path where all required fields are present and valid.
func TestAPPurchaseReceiveTransactionPhaser_PhaseSingleDoc_Success(t *testing.T) {
	// ARRANGE - Set up test data and expected values
	phaser := APPurchaseReceiveTransactionPhaser{}

	validMsg := `{
		"id": "000000000000000000000000",
		"holdingcode": "2PrIIqTWxoBXv16K310sNwfHmfY",
		"guidfixed": "38mCBH7gD2Eiig00wVyUHzUfQSh",
		"docno": "PI2026012600001",
		"docdatetime": "2026-01-26T02:39:29.115Z",
		"guidref": "bed1331f-ce78-4b9e-9798-3f46304fc328",
		"transflag": 315,
		"docreftype": 0,
		"docrefno": "REF001",
		"docrefdate": "2026-01-26T02:39:29.115Z",
		"taxdocdate": "2026-01-26T02:39:29.115Z",
		"taxdocno": "TAX001",
		"inquirytype": 0,
		"vattype": 1,
		"vatrate": 7,
		"custcode": "AP-001",
		"custnames": [
			{
				"code": "th",
				"name": "เจ้าหนี้ทั่วไป",
				"isauto": false,
				"isdelete": false
			}
		],
		"description": "Test accrual receive",
		"discountword": "0",
		"totaldiscount": 0,
		"totalvalue": 50,
		"totalexceptvat": 0,
		"totalaftervat": 50,
		"totalbeforevat": 46.73,
		"totalvatvalue": 3.27,
		"totalamount": 50,
		"iscancel": false,
		"branch": {
			"guidfixed": "2Prp2MbDKqpDBAgSYBtqbVXODwT",
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
				"linenumber": 1,
				"docdatetime": "2026-01-26T02:39:29.115Z",
				"docref": "REF001",
				"docrefdatetime": "2026-01-26T02:39:52.015Z",
				"barcode": "BARCODE002",
				"itemnames": [
					{
						"code": "th",
						"name": "มาม่า",
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
				"itemtype": 0,
				"item_guid": "2PrfZIsQh3VoCxfyFOwOZF5qzux",
				"qty": 1,
				"price": 50,
				"discount": "",
				"discountamount": 0,
				"totalvaluevat": 3.27,
				"priceexcludevat": 46.73,
				"sumamount": 50,
				"sumamountexcludevat": 46.73,
				"dividevalue": 1,
				"standvalue": 6,
				"vattype": 1,
				"taxtype": 0,
				"whcode": "00000",
				"whnames": [
					{
						"code": "th",
						"name": "คลังสำนักงานใหญ่",
						"isauto": false,
						"isdelete": false
					}
				],
				"locationcode": "001",
				"locationnames": [
					{
						"code": "th",
						"name": "fyy",
						"isauto": false,
						"isdelete": false
					}
				],
				"groupcode": "GRP001",
				"groupnames": [
					{
						"code": "th",
						"name": "กลุ่มทดสอบ",
						"isauto": false,
						"isdelete": false
					}
				]
			}
		]
	}`

	// ACT - Execute the method under test
	result, err := phaser.PhaseSingleDoc(validMsg)

	// ASSERT - Verify the results
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result to not be nil")
	}

	// Verify header fields
	if result.GuidFixed != "38mCBH7gD2Eiig00wVyUHzUfQSh" {
		t.Errorf("Expected GuidFixed to be '38mCBH7gD2Eiig00wVyUHzUfQSh', got: %s", result.GuidFixed)
	}

	if result.GuidRef != "bed1331f-ce78-4b9e-9798-3f46304fc328" {
		t.Errorf("Expected GuidRef to be 'bed1331f-ce78-4b9e-9798-3f46304fc328', got: %s", result.GuidRef)
	}

	if result.HoldingCode != "2PrIIqTWxoBXv16K310sNwfHmfY" {
		t.Errorf("Expected HoldingCode to be '2PrIIqTWxoBXv16K310sNwfHmfY', got: %s", result.HoldingCode)
	}

	if result.DocNo != "PI2026012600001" {
		t.Errorf("Expected DocNo to be 'PI2026012600001', got: %s", result.DocNo)
	}

	if result.TransFlag != 12 {
		t.Errorf("Expected TransFlag to be 12, got: %d", result.TransFlag)
	}

	if result.BranchCode != "00000" {
		t.Errorf("Expected BranchCode to be '00000', got: %s", result.BranchCode)
	}

	if result.TaxDocNo != "TAX001" {
		t.Errorf("Expected TaxDocNo to be 'TAX001', got: %s", result.TaxDocNo)
	}

	if result.Description != "Test accrual receive" {
		t.Errorf("Expected Description to be 'Test accrual receive', got: %s", result.Description)
	}

	if result.VatType != 1 {
		t.Errorf("Expected VatType to be 1, got: %d", result.VatType)
	}

	if result.VatRate != 7 {
		t.Errorf("Expected VatRate to be 7, got: %f", result.VatRate)
	}

	if result.DocRefType != 0 {
		t.Errorf("Expected DocRefType to be 0, got: %d", result.DocRefType)
	}

	if result.DocRefNo != "REF001" {
		t.Errorf("Expected DocRefNo to be 'REF001', got: %s", result.DocRefNo)
	}

	if result.TotalValue != 50 {
		t.Errorf("Expected TotalValue to be 50, got: %f", result.TotalValue)
	}

	if result.TotalDiscount != 0 {
		t.Errorf("Expected TotalDiscount to be 0, got: %f", result.TotalDiscount)
	}

	if result.TotalBeforeVat != 46.73 {
		t.Errorf("Expected TotalBeforeVat to be 46.73, got: %f", result.TotalBeforeVat)
	}

	if result.TotalVatValue != 3.27 {
		t.Errorf("Expected TotalVatValue to be 3.27, got: %f", result.TotalVatValue)
	}

	if result.TotalExceptVat != 0 {
		t.Errorf("Expected TotalExceptVat to be 0, got: %f", result.TotalExceptVat)
	}

	if result.TotalAfterVat != 50 {
		t.Errorf("Expected TotalAfterVat to be 50, got: %f", result.TotalAfterVat)
	}

	if result.TotalAmount != 50 {
		t.Errorf("Expected TotalAmount to be 50, got: %f", result.TotalAmount)
	}

	if result.IsCancel != false {
		t.Errorf("Expected IsCancel to be false, got: %v", result.IsCancel)
	}

	// Verify creditor-specific fields
	if result.CreditorCode != "AP-001" {
		t.Errorf("Expected CreditorCode to be 'AP-001', got: %s", result.CreditorCode)
	}

	if len(result.CreditorNames) == 0 {
		t.Error("Expected CreditorNames to not be empty")
	} else if result.CreditorNames[0].Name == nil || *result.CreditorNames[0].Name != "เจ้าหนี้ทั่วไป" {
		t.Errorf("Expected first creditor name to be 'เจ้าหนี้ทั่วไป', got: %v", result.CreditorNames[0].Name)
	}

	// Verify details
	if result.Items == nil {
		t.Fatal("Expected Items to not be nil")
	}

	if len(*result.Items) != 1 {
		t.Errorf("Expected 1 item, got: %d", len(*result.Items))
	}

	item := (*result.Items)[0]
	if item.GuidFixed != "38mCBH7gD2Eiig00wVyUHzUfQSh" {
		t.Errorf("Expected item GuidFixed to be '38mCBH7gD2Eiig00wVyUHzUfQSh', got: %s", item.GuidFixed)
	}

	if item.DocRef != "REF001" {
		t.Errorf("Expected item DocRef to be 'REF001', got: %s", item.DocRef)
	}

	if item.DocNo != "PI2026012600001" {
		t.Errorf("Expected item DocNo to be 'PI2026012600001', got: %s", item.DocNo)
	}

	if item.HoldingCode != "2PrIIqTWxoBXv16K310sNwfHmfY" {
		t.Errorf("Expected item HoldingCode to be '2PrIIqTWxoBXv16K310sNwfHmfY', got: %s", item.HoldingCode)
	}

	if item.LineNumber != 1 {
		t.Errorf("Expected item LineNumber to be 1, got: %d", item.LineNumber)
	}

	if item.Barcode != "BARCODE002" {
		t.Errorf("Expected item Barcode to be 'BARCODE002', got: %s", item.Barcode)
	}

	if item.Qty != 1 {
		t.Errorf("Expected item Qty to be 1, got: %f", item.Qty)
	}

	if item.Price != 50 {
		t.Errorf("Expected item Price to be 50, got: %f", item.Price)
	}

	if item.PriceExcludeVat != 46.73 {
		t.Errorf("Expected item PriceExcludeVat to be 46.73, got: %f", item.PriceExcludeVat)
	}

	if item.DiscountAmount != 0 {
		t.Errorf("Expected item DiscountAmount to be 0, got: %f", item.DiscountAmount)
	}

	if item.SumAmount != 50 {
		t.Errorf("Expected item SumAmount to be 50, got: %f", item.SumAmount)
	}

	if item.SumAmountExcludeVat != 46.73 {
		t.Errorf("Expected item SumAmountExcludeVat to be 46.73, got: %f", item.SumAmountExcludeVat)
	}

	if item.TotalValueVat != 3.27 {
		t.Errorf("Expected item TotalValueVat to be 3.27, got: %f", item.TotalValueVat)
	}

	if item.WhCode != "00000" {
		t.Errorf("Expected item WhCode to be '00000', got: %s", item.WhCode)
	}

	if item.LocationCode != "001" {
		t.Errorf("Expected item LocationCode to be '001', got: %s", item.LocationCode)
	}

	if item.VatType != 1 {
		t.Errorf("Expected item VatType to be 1, got: %d", item.VatType)
	}

	if item.TaxType != 0 {
		t.Errorf("Expected item TaxType to be 0, got: %d", item.TaxType)
	}

	if item.StandValue != 6 {
		t.Errorf("Expected item StandValue to be 6, got: %f", item.StandValue)
	}

	if item.DivideValue != 1 {
		t.Errorf("Expected item DivideValue to be 1, got: %f", item.DivideValue)
	}

	if item.ItemType != 0 {
		t.Errorf("Expected item ItemType to be 0, got: %d", item.ItemType)
	}

	if item.ItemGuid != "2PrfZIsQh3VoCxfyFOwOZF5qzux" {
		t.Errorf("Expected item ItemGuid to be '2PrfZIsQh3VoCxfyFOwOZF5qzux', got: %s", item.ItemGuid)
	}

	if item.UnitCode != "PAC" {
		t.Errorf("Expected item UnitCode to be 'PAC', got: %s", item.UnitCode)
	}

	if item.GroupCode != "GRP001" {
		t.Errorf("Expected item GroupCode to be 'GRP001', got: %s", item.GroupCode)
	}

	// Verify item names arrays
	if len(item.UnitNames) == 0 {
		t.Error("Expected item UnitNames to not be empty")
	} else if item.UnitNames[0].Name == nil || *item.UnitNames[0].Name != "แพ็ค" {
		t.Errorf("Expected first unit name to be 'แพ็ค', got: %v", item.UnitNames[0].Name)
	}

	if len(item.ItemNames) == 0 {
		t.Error("Expected item ItemNames to not be empty")
	} else if item.ItemNames[0].Name == nil || *item.ItemNames[0].Name != "มาม่า" {
		t.Errorf("Expected first item name to be 'มาม่า', got: %v", item.ItemNames[0].Name)
	}

	if len(item.WhNames) == 0 {
		t.Error("Expected item WhNames to not be empty")
	} else if item.WhNames[0].Name == nil || *item.WhNames[0].Name != "คลังสำนักงานใหญ่" {
		t.Errorf("Expected first warehouse name to be 'คลังสำนักงานใหญ่', got: %v", item.WhNames[0].Name)
	}

	if len(item.LocationNames) == 0 {
		t.Error("Expected item LocationNames to not be empty")
	} else if item.LocationNames[0].Name == nil || *item.LocationNames[0].Name != "fyy" {
		t.Errorf("Expected first location name to be 'fyy', got: %v", item.LocationNames[0].Name)
	}

	if len(item.GroupNames) == 0 {
		t.Error("Expected item GroupNames to not be empty")
	} else if item.GroupNames[0].Name == nil || *item.GroupNames[0].Name != "กลุ่มทดสอบ" {
		t.Errorf("Expected first group name to be 'กลุ่มทดสอบ', got: %v", item.GroupNames[0].Name)
	}
}

// TestAPPurchaseReceiveTransactionPhaser_PhaseSingleDoc_SuccessWithNilDetails tests the successful conversion
// of a valid JSON message with nil details to an AccrualReceiveTransactionPG struct.
// This test verifies that the method handles nil details correctly.
func TestAPPurchaseReceiveTransactionPhaser_PhaseSingleDoc_SuccessWithNilDetails(t *testing.T) {
	// ARRANGE - Set up test data with nil details
	phaser := APPurchaseReceiveTransactionPhaser{}

	validMsgWithNilDetails := `{
		"id": "000000000000000000000000",
		"holdingcode": "2PrIIqTWxoBXv16K310sNwfHmfY",
		"guidfixed": "38mCBH7gD2Eiig00wVyUHzUfQSh",
		"docno": "PI2026012600002",
		"docdatetime": "2026-01-26T02:39:29.115Z",
		"guidref": "bed1331f-ce78-4b9e-9798-3f46304fc328",
		"transflag": 315,
		"docreftype": 0,
		"docrefno": "",
		"docrefdate": "2026-01-26T02:39:29.115Z",
		"taxdocdate": "2026-01-26T02:39:29.115Z",
		"taxdocno": "",
		"inquirytype": 0,
		"vattype": 1,
		"vatrate": 7,
		"custcode": "AP-002",
		"custnames": [
			{
				"code": "th",
				"name": "เจ้าหนี้ทดสอบ",
				"isauto": false,
				"isdelete": false
			}
		],
		"description": "Test with nil details",
		"discountword": "0",
		"totaldiscount": 0,
		"totalvalue": 100,
		"totalexceptvat": 0,
		"totalaftervat": 100,
		"totalbeforevat": 93.46,
		"totalvatvalue": 6.54,
		"totalamount": 100,
		"iscancel": false,
		"branch": {
			"guidfixed": "2Prp2MbDKqpDBAgSYBtqbVXODwT",
			"code": "00001",
			"names": [
				{
					"code": "th",
					"name": "สาขาทดสอบ",
					"isauto": false,
					"isdelete": false
				}
			]
		},
		"details": null
	}`

	// ACT - Execute the method under test
	result, err := phaser.PhaseSingleDoc(validMsgWithNilDetails)

	// ASSERT - Verify the results
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result to not be nil")
	}

	// Verify header fields are correctly set
	if result.DocNo != "PI2026012600002" {
		t.Errorf("Expected DocNo to be 'PI2026012600002', got: %s", result.DocNo)
	}

	if result.CreditorCode != "AP-002" {
		t.Errorf("Expected CreditorCode to be 'AP-002', got: %s", result.CreditorCode)
	}

	if result.Description != "Test with nil details" {
		t.Errorf("Expected Description to be 'Test with nil details', got: %s", result.Description)
	}

	if result.BranchCode != "00001" {
		t.Errorf("Expected BranchCode to be '00001', got: %s", result.BranchCode)
	}

	if result.TotalValue != 100 {
		t.Errorf("Expected TotalValue to be 100, got: %f", result.TotalValue)
	}

	// Verify that Items is an empty slice, not nil
	if result.Items == nil {
		t.Error("Expected Items to be an empty slice, not nil")
	} else if len(*result.Items) != 0 {
		t.Errorf("Expected 0 items, got: %d", len(*result.Items))
	}
}

// TestAPPurchaseReceiveTransactionPhaser_PhaseSingleDoc_InvalidJSON tests the error handling
// when the input message contains malformed JSON.
// This test verifies that the method returns an appropriate error for invalid JSON.
func TestAPPurchaseReceiveTransactionPhaser_PhaseSingleDoc_InvalidJSON(t *testing.T) {
	// ARRANGE - Set up test data with malformed JSON
	phaser := APPurchaseReceiveTransactionPhaser{}

	invalidJSON := `{
		"id": "000000000000000000000000",
		"holdingcode": "2PrIIqTWxoBXv16K310sNwfHmfY",
		"guidfixed": "38mCBH7gD2Eiig00wVyUHzUfQSh",
		"docno": "PI2026012600001",
		"docdatetime": "2026-01-26T02:39:29.115Z",
		"guidref": "bed1331f-ce78-4b9e-9798-3f46304fc328",
		"transflag": 315,
		"docreftype": 0,
		"docrefno": "REF001",
		"docrefdate": "2026-01-26T02:39:29.115Z",
		"taxdocdate": "2026-01-26T02:39:29.115Z",
		"taxdocno": "TAX001",
		"inquirytype": 0,
		"vattype": 1,
		"vatrate": 7,
		"custcode": "AP-001",
		"custnames": [
			{
				"code": "th",
				"name": "เจ้าหนี้ทั่วไป",
				"isauto": false,
				"isdelete": false
			}
		],
		"description": "Test accrual receive",
		"discountword": "0",
		"totaldiscount": 0,
		"totalvalue": 50,
		"totalexceptvat": 0,
		"totalaftervat": 50,
		"totalbeforevat": 46.73,
		"totalvatvalue": 3.27,
		"totalamount": 50,
		"iscancel": false,
		"branch": {
			"guidfixed": "2Prp2MbDKqpDBAgSYBtqbVXODwT",
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
				"linenumber": 1,
				"docdatetime": "2026-01-26T02:39:29.115Z",
				"docref": "REF001",
				"docrefdatetime": "2026-01-26T02:39:52.015Z",
				"barcode": "BARCODE002",
				"itemnames": [
					{
						"code": "th",
						"name": "มาม่า",
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
				"itemtype": 0,
				"item_guid": "2PrfZIsQh3VoCxfyFOwOZF5qzux",
				"qty": 1,
				"price": 50,
				"discount": "",
				"discountamount": 0,
				"totalvaluevat": 3.27,
				"priceexcludevat": 46.73,
				"sumamount": 50,
				"sumamountexcludevat": 46.73,
				"dividevalue": 1,
				"standvalue": 6,
				"vattype": 1,
				"taxtype": 0,
				"whcode": "00000",
				"whnames": [
					{
						"code": "th",
						"name": "คลังสำนักงานใหญ่",
						"isauto": false,
						"isdelete": false
					}
				],
				"locationcode": "001",
				"locationnames": [
					{
						"code": "th",
						"name": "fyy",
						"isauto": false,
						"isdelete": false
					}
				],
				"groupcode": "GRP001",
				"groupnames": [
					{
						"code": "th",
						"name": "กลุ่มทดสอบ",
						"isauto": false,
						"isdelete": false
					}
				]
			}
		]
	` // Missing closing brace

	// ACT - Execute the method under test
	result, err := phaser.PhaseSingleDoc(invalidJSON)

	// ASSERT - Verify the results
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}

	if result != nil {
		t.Error("Expected result to be nil for invalid JSON")
	}

	if !strings.Contains(err.Error(), "Cannot Unmarshal AccrualReceive Message") {
		t.Errorf("Expected error message to contain 'Cannot Unmarshal AccrualReceive Message', got: %v", err)
	}
}

// TestAPPurchaseReceiveTransactionPhaser_PhaseSingleDoc_EmptyJSON tests the error handling
// when the input message is an empty JSON object.
// This test verifies that the method handles empty JSON gracefully.
func TestAPPurchaseReceiveTransactionPhaser_PhaseSingleDoc_EmptyJSON(t *testing.T) {
	// ARRANGE - Set up test data with empty JSON
	phaser := APPurchaseReceiveTransactionPhaser{}

	emptyJSON := `{}`

	// ACT - Execute the method under test
	result, err := phaser.PhaseSingleDoc(emptyJSON)

	// ASSERT - Verify the results
	// Empty JSON should still unmarshal successfully but with zero values
	if err != nil {
		t.Errorf("Expected no error for empty JSON, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result to not be nil for empty JSON")
	}

	// Verify that the result has default/zero values
	if result.GuidFixed != "" {
		t.Errorf("Expected GuidFixed to be empty, got: %s", result.GuidFixed)
	}

	if result.DocNo != "" {
		t.Errorf("Expected DocNo to be empty, got: %s", result.DocNo)
	}

	if result.HoldingCode != "" {
		t.Errorf("Expected HoldingCode to be empty, got: %s", result.HoldingCode)
	}

	if result.TransFlag != 12 {
		t.Errorf("Expected TransFlag to be 12, got: %d", result.TransFlag)
	}

	if result.CreditorCode != "" {
		t.Errorf("Expected CreditorCode to be empty, got: %s", result.CreditorCode)
	}

	// Items should be an empty slice
	if result.Items == nil {
		t.Error("Expected Items to be an empty slice, not nil")
	} else if len(*result.Items) != 0 {
		t.Errorf("Expected 0 items, got: %d", len(*result.Items))
	}
}

// TestAPPurchaseReceiveTransactionPhaser_PhaseSingleDoc_EmptyString tests the error handling
// when the input message is an empty string.
// This test verifies that the method returns an appropriate error for empty input.
func TestAPPurchaseReceiveTransactionPhaser_PhaseSingleDoc_EmptyString(t *testing.T) {
	// ARRANGE - Set up test data with empty string
	phaser := APPurchaseReceiveTransactionPhaser{}

	emptyString := ``

	// ACT - Execute the method under test
	result, err := phaser.PhaseSingleDoc(emptyString)

	// ASSERT - Verify the results
	if err == nil {
		t.Error("Expected error for empty string, got nil")
	}

	if result != nil {
		t.Error("Expected result to be nil for empty string")
	}

	if !strings.Contains(err.Error(), "Cannot Unmarshal AccrualReceive Message") {
		t.Errorf("Expected error message to contain 'Cannot Unmarshal AccrualReceive Message', got: %v", err)
	}
}

// TestAPPurchaseReceiveTransactionPhaser_PhaseSingleDoc_InvalidDataType tests the error handling
// when the input message contains invalid data types for fields.
// This test verifies that the method handles type mismatches correctly.
func TestAPPurchaseReceiveTransactionPhaser_PhaseSingleDoc_InvalidDataType(t *testing.T) {
	// ARRANGE - Set up test data with invalid data types
	phaser := APPurchaseReceiveTransactionPhaser{}

	invalidDataTypeJSON := `{
		"id": "000000000000000000000000",
		"holdingcode": "2PrIIqTWxoBXv16K310sNwfHmfY",
		"guidfixed": "38mCBH7gD2Eiig00wVyUHzUfQSh",
		"docno": "PI2026012600001",
		"docdatetime": "2026-01-26T02:39:29.115Z",
		"guidref": "bed1331f-ce78-4b9e-9798-3f46304fc328",
		"transflag": "invalid", // Should be a number, not a string
		"docreftype": 0,
		"docrefno": "REF001",
		"docrefdate": "2026-01-26T02:39:29.115Z",
		"taxdocdate": "2026-01-26T02:39:29.115Z",
		"taxdocno": "TAX001",
		"inquirytype": 0,
		"vattype": 1,
		"vatrate": 7,
		"custcode": "AP-001",
		"custnames": [
			{
				"code": "th",
				"name": "เจ้าหนี้ทั่วไป",
				"isauto": false,
				"isdelete": false
			}
		],
		"description": "Test accrual receive",
		"discountword": "0",
		"totaldiscount": 0,
		"totalvalue": 50,
		"totalexceptvat": 0,
		"totalaftervat": 50,
		"totalbeforevat": 46.73,
		"totalvatvalue": 3.27,
		"totalamount": 50,
		"iscancel": false,
		"branch": {
			"guidfixed": "2Prp2MbDKqpDBAgSYBtqbVXODwT",
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
				"linenumber": 1,
				"docdatetime": "2026-01-26T02:39:29.115Z",
				"docref": "REF001",
				"docrefdatetime": "2026-01-26T02:39:52.015Z",
				"barcode": "BARCODE002",
				"itemnames": [
					{
						"code": "th",
						"name": "มาม่า",
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
				"itemtype": 0,
				"item_guid": "2PrfZIsQh3VoCxfyFOwOZF5qzux",
				"qty": 1,
				"price": 50,
				"discount": "",
				"discountamount": 0,
				"totalvaluevat": 3.27,
				"priceexcludevat": 46.73,
				"sumamount": 50,
				"sumamountexcludevat": 46.73,
				"dividevalue": 1,
				"standvalue": 6,
				"vattype": 1,
				"taxtype": 0,
				"whcode": "00000",
				"whnames": [
					{
						"code": "th",
						"name": "คลังสำนักงานใหญ่",
						"isauto": false,
						"isdelete": false
					}
				],
				"locationcode": "001",
				"locationnames": [
					{
						"code": "th",
						"name": "fyy",
						"isauto": false,
						"isdelete": false
					}
				],
				"groupcode": "GRP001",
				"groupnames": [
					{
						"code": "th",
						"name": "กลุ่มทดสอบ",
						"isauto": false,
						"isdelete": false
					}
				]
			}
		]
	}`

	// ACT - Execute the method under test
	result, err := phaser.PhaseSingleDoc(invalidDataTypeJSON)

	// ASSERT - Verify the results
	if err == nil {
		t.Error("Expected error for invalid data type, got nil")
	}

	if result != nil {
		t.Error("Expected result to be nil for invalid data type")
	}

	if !strings.Contains(err.Error(), "Cannot Unmarshal AccrualReceive Message") {
		t.Errorf("Expected error message to contain 'Cannot Unmarshal AccrualReceive Message', got: %v", err)
	}
}

// TestAPPurchaseReceiveTransactionPhaser_PhaseSingleDoc_MultipleDetails tests the successful conversion
// of a valid JSON message with multiple details to an AccrualReceiveTransactionPG struct.
// This test verifies that the method correctly handles multiple line items.
func TestAPPurchaseReceiveTransactionPhaser_PhaseSingleDoc_MultipleDetails(t *testing.T) {
	// ARRANGE - Set up test data with multiple details
	phaser := APPurchaseReceiveTransactionPhaser{}

	validMsgWithMultipleDetails := `{
		"id": "000000000000000000000000",
		"holdingcode": "2PrIIqTWxoBXv16K310sNwfHmfY",
		"guidfixed": "38mCBH7gD2Eiig00wVyUHzUfQSh",
		"docno": "PI2026012600003",
		"docdatetime": "2026-01-26T02:39:29.115Z",
		"guidref": "bed1331f-ce78-4b9e-9798-3f46304fc328",
		"transflag": 315,
		"docreftype": 0,
		"docrefno": "REF001",
		"docrefdate": "2026-01-26T02:39:29.115Z",
		"taxdocdate": "2026-01-26T02:39:29.115Z",
		"taxdocno": "TAX001",
		"inquirytype": 0,
		"vattype": 1,
		"vatrate": 7,
		"custcode": "AP-001",
		"custnames": [
			{
				"code": "th",
				"name": "เจ้าหนี้ทั่วไป",
				"isauto": false,
				"isdelete": false
			}
		],
		"description": "Test with multiple details",
		"discountword": "0",
		"totaldiscount": 0,
		"totalvalue": 150,
		"totalexceptvat": 0,
		"totalaftervat": 150,
		"totalbeforevat": 140.19,
		"totalvatvalue": 9.81,
		"totalamount": 150,
		"iscancel": false,
		"branch": {
			"guidfixed": "2Prp2MbDKqpDBAgSYBtqbVXODwT",
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
				"linenumber": 1,
				"docdatetime": "2026-01-26T02:39:29.115Z",
				"docref": "REF001",
				"docrefdatetime": "2026-01-26T02:39:52.015Z",
				"barcode": "BARCODE001",
				"itemnames": [
					{
						"code": "th",
						"name": "สินค้าที่ 1",
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
				"itemtype": 0,
				"item_guid": "2PrfZIsQh3VoCxfyFOwOZF5qzux",
				"qty": 1,
				"price": 50,
				"discount": "",
				"discountamount": 0,
				"totalvaluevat": 3.27,
				"priceexcludevat": 46.73,
				"sumamount": 50,
				"sumamountexcludevat": 46.73,
				"dividevalue": 1,
				"standvalue": 6,
				"vattype": 1,
				"taxtype": 0,
				"whcode": "00000",
				"whnames": [
					{
						"code": "th",
						"name": "คลังสำนักงานใหญ่",
						"isauto": false,
						"isdelete": false
					}
				],
				"locationcode": "001",
				"locationnames": [
					{
						"code": "th",
						"name": "fyy",
						"isauto": false,
						"isdelete": false
					}
				],
				"groupcode": "GRP001",
				"groupnames": [
					{
						"code": "th",
						"name": "กลุ่มทดสอบ",
						"isauto": false,
						"isdelete": false
					}
				]
			},
			{
				"inquirytype": 0,
				"linenumber": 2,
				"docdatetime": "2026-01-26T02:39:29.115Z",
				"docref": "REF001",
				"docrefdatetime": "2026-01-26T02:39:52.015Z",
				"barcode": "BARCODE002",
				"itemnames": [
					{
						"code": "th",
						"name": "สินค้าที่ 2",
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
				"itemtype": 0,
				"item_guid": "2PrfZIsQh3VoCxfyFOwOZF5qzuy",
				"qty": 2,
				"price": 50,
				"discount": "",
				"discountamount": 0,
				"totalvaluevat": 6.54,
				"priceexcludevat": 46.73,
				"sumamount": 100,
				"sumamountexcludevat": 93.46,
				"dividevalue": 1,
				"standvalue": 6,
				"vattype": 1,
				"taxtype": 0,
				"whcode": "00000",
				"whnames": [
					{
						"code": "th",
						"name": "คลังสำนักงานใหญ่",
						"isauto": false,
						"isdelete": false
					}
				],
				"locationcode": "001",
				"locationnames": [
					{
						"code": "th",
						"name": "fyy",
						"isauto": false,
						"isdelete": false
					}
				],
				"groupcode": "GRP001",
				"groupnames": [
					{
						"code": "th",
						"name": "กลุ่มทดสอบ",
						"isauto": false,
						"isdelete": false
					}
				]
			}
		]
	}`

	// ACT - Execute the method under test
	result, err := phaser.PhaseSingleDoc(validMsgWithMultipleDetails)

	// ASSERT - Verify the results
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result to not be nil")
	}

	// Verify header fields
	if result.DocNo != "PI2026012600003" {
		t.Errorf("Expected DocNo to be 'PI2026012600003', got: %s", result.DocNo)
	}

	if result.TotalValue != 150 {
		t.Errorf("Expected TotalValue to be 150, got: %f", result.TotalValue)
	}

	if result.TotalAmount != 150 {
		t.Errorf("Expected TotalAmount to be 150, got: %f", result.TotalAmount)
	}

	// Verify details count
	if result.Items == nil {
		t.Fatal("Expected Items to not be nil")
	}

	if len(*result.Items) != 2 {
		t.Errorf("Expected 2 items, got: %d", len(*result.Items))
	}

	// Verify first item
	item1 := (*result.Items)[0]
	if item1.LineNumber != 1 {
		t.Errorf("Expected first item LineNumber to be 1, got: %d", item1.LineNumber)
	}

	if item1.Barcode != "BARCODE001" {
		t.Errorf("Expected first item Barcode to be 'BARCODE001', got: %s", item1.Barcode)
	}

	if item1.Qty != 1 {
		t.Errorf("Expected first item Qty to be 1, got: %f", item1.Qty)
	}

	if item1.SumAmount != 50 {
		t.Errorf("Expected first item SumAmount to be 50, got: %f", item1.SumAmount)
	}

	// Verify second item
	item2 := (*result.Items)[1]
	if item2.LineNumber != 2 {
		t.Errorf("Expected second item LineNumber to be 2, got: %d", item2.LineNumber)
	}

	if item2.Barcode != "BARCODE002" {
		t.Errorf("Expected second item Barcode to be 'BARCODE002', got: %s", item2.Barcode)
	}

	if item2.Qty != 2 {
		t.Errorf("Expected second item Qty to be 2, got: %f", item2.Qty)
	}

	if item2.SumAmount != 100 {
		t.Errorf("Expected second item SumAmount to be 100, got: %f", item2.SumAmount)
	}
}

// TestAPPurchaseReceiveTransactionPhaser_PhaseSingleDoc_SpecialCharacters tests the successful conversion
// of a valid JSON message with special characters in string fields.
// This test verifies that the method correctly handles special characters and Unicode.
func TestAPPurchaseReceiveTransactionPhaser_PhaseSingleDoc_SpecialCharacters(t *testing.T) {
	// ARRANGE - Set up test data with special characters
	phaser := APPurchaseReceiveTransactionPhaser{}

	validMsgWithSpecialChars := `{
		"id": "000000000000000000000000",
		"holdingcode": "2PrIIqTWxoBXv16K310sNwfHmfY",
		"guidfixed": "38mCBH7gD2Eiig00wVyUHzUfQSh",
		"docno": "PI2026012600004",
		"docdatetime": "2026-01-26T02:39:29.115Z",
		"guidref": "bed1331f-ce78-4b9e-9798-3f46304fc328",
		"transflag": 315,
		"docreftype": 0,
		"docrefno": "REF-001",
		"docrefdate": "2026-01-26T02:39:29.115Z",
		"taxdocdate": "2026-01-26T02:39:29.115Z",
		"taxdocno": "TAX-001",
		"inquirytype": 0,
		"vattype": 1,
		"vatrate": 7,
		"custcode": "AP-001",
		"custnames": [
			{
				"code": "th",
				"name": "เจ้าหนี้ทั่วไป & พิเศษ",
				"isauto": false,
				"isdelete": false
			}
		],
		"description": "Test with special characters: <>&\"'",
		"discountword": "0",
		"totaldiscount": 0,
		"totalvalue": 50,
		"totalexceptvat": 0,
		"totalaftervat": 50,
		"totalbeforevat": 46.73,
		"totalvatvalue": 3.27,
		"totalamount": 50,
		"iscancel": false,
		"branch": {
			"guidfixed": "2Prp2MbDKqpDBAgSYBtqbVXODwT",
			"code": "00000",
			"names": [
				{
					"code": "th",
					"name": "สำนักงานใหญ่ (HQ)",
					"isauto": false,
					"isdelete": false
				}
			]
		},
		"details": [
			{
				"inquirytype": 0,
				"linenumber": 1,
				"docdatetime": "2026-01-26T02:39:29.115Z",
				"docref": "REF-001",
				"docrefdatetime": "2026-01-26T02:39:52.015Z",
				"barcode": "BARCODE-001",
				"itemnames": [
					{
						"code": "th",
						"name": "มาม่า (รสหมู)",
						"isauto": false,
						"isdelete": false
					}
				],
				"unitcode": "PAC",
				"unitnames": [
					{
						"code": "th",
						"name": "แพ็ค (Pack)",
						"isauto": false,
						"isdelete": false
					}
				],
				"itemtype": 0,
				"item_guid": "2PrfZIsQh3VoCxfyFOwOZF5qzux",
				"qty": 1,
				"price": 50,
				"discount": "",
				"discountamount": 0,
				"totalvaluevat": 3.27,
				"priceexcludevat": 46.73,
				"sumamount": 50,
				"sumamountexcludevat": 46.73,
				"dividevalue": 1,
				"standvalue": 6,
				"vattype": 1,
				"taxtype": 0,
				"whcode": "00000",
				"whnames": [
					{
						"code": "th",
						"name": "คลังสำนักงานใหญ่ (HQ)",
						"isauto": false,
						"isdelete": false
					}
				],
				"locationcode": "001",
				"locationnames": [
					{
						"code": "th",
						"name": "fyy (location)",
						"isauto": false,
						"isdelete": false
					}
				],
				"groupcode": "GRP-001",
				"groupnames": [
					{
						"code": "th",
						"name": "กลุ่มทดสอบ (Test Group)",
						"isauto": false,
						"isdelete": false
					}
				]
			}
		]
	}`

	// ACT - Execute the method under test
	result, err := phaser.PhaseSingleDoc(validMsgWithSpecialChars)

	// ASSERT - Verify the results
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result to not be nil")
	}

	// Verify that special characters are preserved
	expectedDescription := "Test with special characters: <>&\"'"
	if result.Description != expectedDescription {
		t.Errorf("Expected Description to be '%s', got: %s", expectedDescription, result.Description)
	}

	if result.DocRefNo != "REF-001" {
		t.Errorf("Expected DocRefNo to be 'REF-001', got: %s", result.DocRefNo)
	}

	if result.TaxDocNo != "TAX-001" {
		t.Errorf("Expected TaxDocNo to be 'TAX-001', got: %s", result.TaxDocNo)
	}

	if len(result.CreditorNames) == 0 {
		t.Error("Expected CreditorNames to not be empty")
	} else if result.CreditorNames[0].Name == nil || *result.CreditorNames[0].Name != "เจ้าหนี้ทั่วไป & พิเศษ" {
		t.Errorf("Expected first creditor name to be 'เจ้าหนี้ทั่วไป & พิเศษ', got: %v", result.CreditorNames[0].Name)
	}

	// Verify item with special characters
	item := (*result.Items)[0]
	if item.Barcode != "BARCODE-001" {
		t.Errorf("Expected item Barcode to be 'BARCODE-001', got: %s", item.Barcode)
	}

	if len(item.ItemNames) == 0 {
		t.Error("Expected item ItemNames to not be empty")
	} else if item.ItemNames[0].Name == nil || *item.ItemNames[0].Name != "มาม่า (รสหมู)" {
		t.Errorf("Expected first item name to be 'มาม่า (รสหมู)', got: %v", item.ItemNames[0].Name)
	}

	if item.GroupCode != "GRP-001" {
		t.Errorf("Expected item GroupCode to be 'GRP-001', got: %s", item.GroupCode)
	}
}

// TestAPPurchaseReceiveTransactionPhaser_PhaseSingleDoc_IsCancelTrue tests the successful conversion
// of a valid JSON message with IsCancel set to true.
// This test verifies that the method correctly handles cancelled transactions.
func TestAPPurchaseReceiveTransactionPhaser_PhaseSingleDoc_IsCancelTrue(t *testing.T) {
	// ARRANGE - Set up test data with IsCancel true
	phaser := APPurchaseReceiveTransactionPhaser{}

	cancelledMsg := `{
		"id": "000000000000000000000000",
		"holdingcode": "2PrIIqTWxoBXv16K310sNwfHmfY",
		"guidfixed": "38mCBH7gD2Eiig00wVyUHzUfQSh",
		"docno": "PI2026012600005",
		"docdatetime": "2026-01-26T02:39:29.115Z",
		"guidref": "bed1331f-ce78-4b9e-9798-3f46304fc328",
		"transflag": 315,
		"docreftype": 0,
		"docrefno": "REF001",
		"docrefdate": "2026-01-26T02:39:29.115Z",
		"taxdocdate": "2026-01-26T02:39:29.115Z",
		"taxdocno": "TAX001",
		"inquirytype": 0,
		"vattype": 1,
		"vatrate": 7,
		"custcode": "AP-001",
		"custnames": [
			{
				"code": "th",
				"name": "เจ้าหนี้ทั่วไป",
				"isauto": false,
				"isdelete": false
			}
		],
		"description": "Cancelled transaction",
		"discountword": "0",
		"totaldiscount": 0,
		"totalvalue": 50,
		"totalexceptvat": 0,
		"totalaftervat": 50,
		"totalbeforevat": 46.73,
		"totalvatvalue": 3.27,
		"totalamount": 50,
		"iscancel": true,
		"branch": {
			"guidfixed": "2Prp2MbDKqpDBAgSYBtqbVXODwT",
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
				"linenumber": 1,
				"docdatetime": "2026-01-26T02:39:29.115Z",
				"docref": "REF001",
				"docrefdatetime": "2026-01-26T02:39:52.015Z",
				"barcode": "BARCODE002",
				"itemnames": [
					{
						"code": "th",
						"name": "มาม่า",
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
				"itemtype": 0,
				"item_guid": "2PrfZIsQh3VoCxfyFOwOZF5qzux",
				"qty": 1,
				"price": 50,
				"discount": "",
				"discountamount": 0,
				"totalvaluevat": 3.27,
				"priceexcludevat": 46.73,
				"sumamount": 50,
				"sumamountexcludevat": 46.73,
				"dividevalue": 1,
				"standvalue": 6,
				"vattype": 1,
				"taxtype": 0,
				"whcode": "00000",
				"whnames": [
					{
						"code": "th",
						"name": "คลังสำนักงานใหญ่",
						"isauto": false,
						"isdelete": false
					}
				],
				"locationcode": "001",
				"locationnames": [
					{
						"code": "th",
						"name": "fyy",
						"isauto": false,
						"isdelete": false
					}
				],
				"groupcode": "GRP001",
				"groupnames": [
					{
						"code": "th",
						"name": "กลุ่มทดสอบ",
						"isauto": false,
						"isdelete": false
					}
				]
			}
		]
	}`

	// ACT - Execute the method under test
	result, err := phaser.PhaseSingleDoc(cancelledMsg)

	// ASSERT - Verify the results
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result to not be nil")
	}

	// Verify that IsCancel is correctly set
	if result.IsCancel != true {
		t.Errorf("Expected IsCancel to be true, got: %v", result.IsCancel)
	}

	if result.Description != "Cancelled transaction" {
		t.Errorf("Expected Description to be 'Cancelled transaction', got: %s", result.Description)
	}
}

// TestAPPurchaseReceiveTransactionPhaser_PhaseSingleDoc_ZeroValues tests the successful conversion
// of a valid JSON message with zero values for numeric fields.
// This test verifies that the method correctly handles zero values.
func TestAPPurchaseReceiveTransactionPhaser_PhaseSingleDoc_ZeroValues(t *testing.T) {
	// ARRANGE - Set up test data with zero values
	phaser := APPurchaseReceiveTransactionPhaser{}

	zeroValuesMsg := `{
		"id": "000000000000000000000000",
		"holdingcode": "2PrIIqTWxoBXv16K310sNwfHmfY",
		"guidfixed": "38mCBH7gD2Eiig00wVyUHzUfQSh",
		"docno": "PI2026012600006",
		"docdatetime": "2026-01-26T02:39:29.115Z",
		"guidref": "bed1331f-ce78-4b9e-9798-3f46304fc328",
		"transflag": 315,
		"docreftype": 0,
		"docrefno": "",
		"docrefdate": "2026-01-26T02:39:29.115Z",
		"taxdocdate": "2026-01-26T02:39:29.115Z",
		"taxdocno": "",
		"inquirytype": 0,
		"vattype": 0,
		"vatrate": 0,
		"custcode": "AP-001",
		"custnames": [
			{
				"code": "th",
				"name": "เจ้าหนี้ทั่วไป",
				"isauto": false,
				"isdelete": false
			}
		],
		"description": "Zero values test",
		"discountword": "0",
		"totaldiscount": 0,
		"totalvalue": 0,
		"totalexceptvat": 0,
		"totalaftervat": 0,
		"totalbeforevat": 0,
		"totalvatvalue": 0,
		"totalamount": 0,
		"iscancel": false,
		"branch": {
			"guidfixed": "2Prp2MbDKqpDBAgSYBtqbVXODwT",
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
				"linenumber": 1,
				"docdatetime": "2026-01-26T02:39:29.115Z",
				"docref": "",
				"docrefdatetime": "2026-01-26T02:39:52.015Z",
				"barcode": "BARCODE002",
				"itemnames": [
					{
						"code": "th",
						"name": "มาม่า",
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
				"itemtype": 0,
				"item_guid": "2PrfZIsQh3VoCxfyFOwOZF5qzux",
				"qty": 0,
				"price": 0,
				"discount": "",
				"discountamount": 0,
				"totalvaluevat": 0,
				"priceexcludevat": 0,
				"sumamount": 0,
				"sumamountexcludevat": 0,
				"dividevalue": 0,
				"standvalue": 0,
				"vattype": 0,
				"taxtype": 0,
				"whcode": "00000",
				"whnames": [
					{
						"code": "th",
						"name": "คลังสำนักงานใหญ่",
						"isauto": false,
						"isdelete": false
					}
				],
				"locationcode": "001",
				"locationnames": [
					{
						"code": "th",
						"name": "fyy",
						"isauto": false,
						"isdelete": false
					}
				],
				"groupcode": "",
				"groupnames": null
			}
		]
	}`

	// ACT - Execute the method under test
	result, err := phaser.PhaseSingleDoc(zeroValuesMsg)

	// ASSERT - Verify the results
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result to not be nil")
	}

	// Verify that zero values are correctly set
	if result.VatType != 0 {
		t.Errorf("Expected VatType to be 0, got: %d", result.VatType)
	}

	if result.VatRate != 0 {
		t.Errorf("Expected VatRate to be 0, got: %f", result.VatRate)
	}

	if result.TotalValue != 0 {
		t.Errorf("Expected TotalValue to be 0, got: %f", result.TotalValue)
	}

	if result.TotalDiscount != 0 {
		t.Errorf("Expected TotalDiscount to be 0, got: %f", result.TotalDiscount)
	}

	if result.TotalBeforeVat != 0 {
		t.Errorf("Expected TotalBeforeVat to be 0, got: %f", result.TotalBeforeVat)
	}

	if result.TotalVatValue != 0 {
		t.Errorf("Expected TotalVatValue to be 0, got: %f", result.TotalVatValue)
	}

	if result.TotalExceptVat != 0 {
		t.Errorf("Expected TotalExceptVat to be 0, got: %f", result.TotalExceptVat)
	}

	if result.TotalAfterVat != 0 {
		t.Errorf("Expected TotalAfterVat to be 0, got: %f", result.TotalAfterVat)
	}

	if result.TotalAmount != 0 {
		t.Errorf("Expected TotalAmount to be 0, got: %f", result.TotalAmount)
	}

	// Verify item zero values
	item := (*result.Items)[0]
	if item.Qty != 0 {
		t.Errorf("Expected item Qty to be 0, got: %f", item.Qty)
	}

	if item.Price != 0 {
		t.Errorf("Expected item Price to be 0, got: %f", item.Price)
	}

	if item.SumAmount != 0 {
		t.Errorf("Expected item SumAmount to be 0, got: %f", item.SumAmount)
	}

	if item.DivideValue != 0 {
		t.Errorf("Expected item DivideValue to be 0, got: %f", item.DivideValue)
	}

	if item.StandValue != 0 {
		t.Errorf("Expected item StandValue to be 0, got: %f", item.StandValue)
	}
}

// TestAPPurchaseReceiveTransactionPhaser_PhaseSingleDoc_ReturnsCorrectType tests that the method
// returns the correct type (*models.AccrualReceiveTransactionPG).
// This test verifies type safety and correct return type.
func TestAPPurchaseReceiveTransactionPhaser_PhaseSingleDoc_ReturnsCorrectType(t *testing.T) {
	// ARRANGE - Set up test data
	phaser := APPurchaseReceiveTransactionPhaser{}

	validMsg := `{
		"id": "000000000000000000000000",
		"holdingcode": "2PrIIqTWxoBXv16K310sNwfHmfY",
		"guidfixed": "38mCBH7gD2Eiig00wVyUHzUfQSh",
		"docno": "PI2026012600007",
		"docdatetime": "2026-01-26T02:39:29.115Z",
		"guidref": "bed1331f-ce78-4b9e-9798-3f46304fc328",
		"transflag": 315,
		"docreftype": 0,
		"docrefno": "REF001",
		"docrefdate": "2026-01-26T02:39:29.115Z",
		"taxdocdate": "2026-01-26T02:39:29.115Z",
		"taxdocno": "TAX001",
		"inquirytype": 0,
		"vattype": 1,
		"vatrate": 7,
		"custcode": "AP-001",
		"custnames": [
			{
				"code": "th",
				"name": "เจ้าหนี้ทั่วไป",
				"isauto": false,
				"isdelete": false
			}
		],
		"description": "Type test",
		"discountword": "0",
		"totaldiscount": 0,
		"totalvalue": 50,
		"totalexceptvat": 0,
		"totalaftervat": 50,
		"totalbeforevat": 46.73,
		"totalvatvalue": 3.27,
		"totalamount": 50,
		"iscancel": false,
		"branch": {
			"guidfixed": "2Prp2MbDKqpDBAgSYBtqbVXODwT",
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
				"linenumber": 1,
				"docdatetime": "2026-01-26T02:39:29.115Z",
				"docref": "REF001",
				"docrefdatetime": "2026-01-26T02:39:52.015Z",
				"barcode": "BARCODE002",
				"itemnames": [
					{
						"code": "th",
						"name": "มาม่า",
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
				"itemtype": 0,
				"item_guid": "2PrfZIsQh3VoCxfyFOwOZF5qzux",
				"qty": 1,
				"price": 50,
				"discount": "",
				"discountamount": 0,
				"totalvaluevat": 3.27,
				"priceexcludevat": 46.73,
				"sumamount": 50,
				"sumamountexcludevat": 46.73,
				"dividevalue": 1,
				"standvalue": 6,
				"vattype": 1,
				"taxtype": 0,
				"whcode": "00000",
				"whnames": [
					{
						"code": "th",
						"name": "คลังสำนักงานใหญ่",
						"isauto": false,
						"isdelete": false
					}
				],
				"locationcode": "001",
				"locationnames": [
					{
						"code": "th",
						"name": "fyy",
						"isauto": false,
						"isdelete": false
					}
				],
				"groupcode": "GRP001",
				"groupnames": [
					{
						"code": "th",
						"name": "กลุ่มทดสอบ",
						"isauto": false,
						"isdelete": false
					}
				]
			}
		]
	}`

	// ACT - Execute the method under test
	result, err := phaser.PhaseSingleDoc(validMsg)

	// ASSERT - Verify the results
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result to not be nil")
	}

	// Verify the type is correct
	var _ *models.APPurchaseReceivePG = result

	// Verify the embedded TransactionPG fields are accessible
	if result.GuidFixed == "" {
		t.Error("Expected GuidFixed to be set")
	}

	if result.HoldingCode == "" {
		t.Error("Expected HoldingCode to be set")
	}

	// Verify the AccrualReceiveTransactionPG specific fields are accessible
	if result.CreditorCode == "" {
		t.Error("Expected CreditorCode to be set")
	}

	if result.Items == nil {
		t.Error("Expected Items to be set")
	}
}
