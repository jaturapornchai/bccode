package main

import "testing"

func TestNormalizeStorageName(t *testing.T) {
	cases := map[string]string{
		"shopUsers":                       "shop_users",
		"shopUserAccessLogs":              "shop_user_access_logs",
		"productBarcodeBOMs":              "product_barcode_boms",
		"kbDocumentMetadata":              "kb_document_metadata",
		" mixed-Collection Name ":         "mixed_collection_name",
		"productbarcode":                  "product_barcodes",
		"productbarcode_dict":             "product_barcode_dict",
		"productbarcodeimport":            "product_barcode_import",
		"saleinvoice_transaction":         "sale_invoice_transaction",
		"purchasereceive_transaction":     "purchase_receive_transaction",
		"chartofaccounts":                 "chart_of_accounts",
		"docdetail":                       "doc_detail",
		"transactionSaleinvoiceBOMPrices": "transaction_saleinvoice_bom_prices",
	}

	for input, expected := range cases {
		if actual := normalizeStorageName(input); actual != expected {
			t.Fatalf("normalizeStorageName(%q) = %q, want %q", input, actual, expected)
		}
	}
}

func TestNormalizeStorageFieldName(t *testing.T) {
	cases := map[string]string{
		"docDate":       "doc_date",
		"docdate":       "doc_date",
		"custCode":      "cust_code",
		"custcode":      "cust_code",
		"holding_code":  "holding_code",
		"holdingCode":   "holding_code",
		"transFlag":     "trans_flag",
		"itemcode":      "item_code",
		"unitCode":      "unit_code",
		"already_snake": "already_snake",
	}

	for input, expected := range cases {
		if actual := normalizeStorageFieldName(input); actual != expected {
			t.Fatalf("normalizeStorageFieldName(%q) = %q, want %q", input, actual, expected)
		}
	}
}
