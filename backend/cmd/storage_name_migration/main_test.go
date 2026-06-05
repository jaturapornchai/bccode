package main

import "testing"

func TestNormalizeStorageName(t *testing.T) {
	cases := map[string]string{
		"shopUsers":                       "shopusers",
		"shopUserAccessLogs":              "shopuseraccesslogs",
		"productBarcodeBOMs":              "productbarcodeboms",
		"kbDocumentMetadata":              "kbdocumentmetadata",
		" mixed-Collection Name ":         "mixedcollectionname",
		"productbarcode":                  "productbarcodes",
		"productbarcodedict":              "productbarcodedict",
		"productbarcodeimport":            "productbarcodeimport",
		"saleinvoicetransaction":          "saleinvoicetransaction",
		"purchasereceivetransaction":      "purchasereceivetransaction",
		"chartofaccounts":                 "chartofaccounts",
		"docdetail":                       "docdetail",
		"transactionSaleinvoiceBOMPrices": "transactionsaleinvoicebomprices",
	}

	for input, expected := range cases {
		if actual := normalizeStorageName(input); actual != expected {
			t.Fatalf("normalizeStorageName(%q) = %q, want %q", input, actual, expected)
		}
	}
}

func TestNormalizeStorageFieldName(t *testing.T) {
	cases := map[string]string{
		"docDate":      "docdate",
		"docdate":      "docdate",
		"custCode":     "custcode",
		"custcode":     "custcode",
		"holdingcode":  "holdingcode",
		"holdingCode":  "holdingcode",
		"transFlag":    "transflag",
		"itemcode":     "itemcode",
		"unitCode":     "unitcode",
		"alreadysnake": "alreadysnake",
	}

	for input, expected := range cases {
		if actual := normalizeStorageFieldName(input); actual != expected {
			t.Fatalf("normalizeStorageFieldName(%q) = %q, want %q", input, actual, expected)
		}
	}
}
