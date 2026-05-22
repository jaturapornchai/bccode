package microservice

import (
	"strings"
	"unicode"
)

var legacyStorageNameAliases = map[string]string{
	"advancepayment":                              "advance_payment",
	"advancepaymentrefund":                        "advance_payment_refund",
	"ap_advancepayment_refund_transaction":        "ap_advance_payment_refund_transaction",
	"ap_advancepayment_refund_transaction_detail": "ap_advance_payment_refund_transaction_detail",
	"ap_advancepayment_transaction":               "ap_advance_payment_transaction",
	"ap_advancepayment_transaction_detail":        "ap_advance_payment_transaction_detail",
	"ap_depositpayment_refund_transaction":        "ap_deposit_payment_refund_transaction",
	"ap_depositpayment_refund_transaction_detail": "ap_deposit_payment_refund_transaction_detail",
	"ap_depositpayment_transaction":               "ap_deposit_payment_transaction",
	"ap_depositpayment_transaction_detail":        "ap_deposit_payment_transaction_detail",
	"ap_purchasereceive_transaction":              "ap_purchase_receive_transaction",
	"ap_purchasereceive_transaction_detail":       "ap_purchase_receive_transaction_detail",
	"ar_advancepayment_refund_transaction":        "ar_advance_payment_refund_transaction",
	"ar_advancepayment_refund_transaction_detail": "ar_advance_payment_refund_transaction_detail",
	"ar_advancepayment_transaction":               "ar_advance_payment_transaction",
	"ar_advancepayment_transaction_detail":        "ar_advance_payment_transaction_detail",
	"ar_depositpayment_refund_transaction":        "ar_deposit_payment_refund_transaction",
	"ar_depositpayment_refund_transaction_detail": "ar_deposit_payment_refund_transaction_detail",
	"ar_depositpayment_transaction":               "ar_deposit_payment_transaction",
	"ar_depositpayment_transaction_detail":        "ar_deposit_payment_transaction_detail",
	"bankmaster":                                  "bank_master",
	"banktransferrecord":                          "bank_transfer_record",
	"bookbank":                                    "book_bank",
	"cartorder":                                   "cart_order",
	"cartorderdetail":                             "cart_order_detail",
	"chartofaccounts":                             "chart_of_accounts",
	"chequechange":                                "cheque_change",
	"chequedeposit":                               "cheque_deposit",
	"chequedisqualified":                          "cheque_disqualified",
	"chequepass":                                  "cheque_pass",
	"chequepaymentchange":                         "cheque_payment_change",
	"chequepaymentdeposit":                        "cheque_payment_deposit",
	"chequepaymentdisqualified":                   "cheque_payment_disqualified",
	"chequepaymentreturn":                         "cheque_payment_return",
	"chequerenew":                                 "cheque_renew",
	"chequereturn":                                "cheque_return",
	"costcenter":                                  "cost_center",
	"creditcardwithdrawal":                        "credit_card_withdrawal",
	"depositrecord":                               "deposit_record",
	"depositrefund":                               "deposit_refund",
	"docdetail":                                   "doc_detail",
	"docdetail_updated":                           "doc_detail_updated",
	"docpayment":                                  "doc_payment",
	"docref":                                      "doc_ref",
	"docwaitprocess":                              "doc_wait_process",
	"inventoryoptions":                            "inventory_options",
	"jobproject":                                  "job_project",
	"journalvats_details":                         "journal_vats_details",
	"journaltaxes_details":                        "journal_taxes_details",
	"ordertype":                                   "order_type",
	"paymentmaster":                               "payment_master",
	"paidadvance":                                 "paid_advance",
	"paidadvancerefund":                           "paid_advance_refund",
	"pickandpack":                                 "pick_and_pack",
	"processstock":                                "process_stock",
	"processstockcost":                            "process_stock_cost",
	"processstockdetail":                          "process_stock_detail",
	"processstocklot":                             "process_stock_lot",
	"productbarcode":                              "product_barcodes",
	"productbarcode_dict":                         "product_barcode_dict",
	"productbarcodeboms":                          "product_barcode_boms",
	"productbarcodeimport":                        "product_barcode_import",
	"productbarcodeprocess":                       "product_barcode_process",
	"productbarcoderef":                           "product_barcode_ref",
	"productcategory":                             "product_category",
	"producttype":                                 "product_type",
	"productunit":                                 "product_unit",
	"purchaseorder":                               "purchase_order",
	"purchaseorder_transaction":                   "purchase_order_transaction",
	"purchaseorder_transaction_detail":            "purchase_order_transaction_detail",
	"purchasepartial":                             "purchase_partial",
	"purchasereceive_transaction":                 "purchase_receive_transaction",
	"purchasereceive_transaction_detail":          "purchase_receive_transaction_detail",
	"purchaserequisition":                         "purchase_requisition",
	"purchasedebitnote":                           "purchase_debit_note",
	"receivedeposit":                              "receive_deposit",
	"receivedepositrefund":                        "receive_deposit_refund",
	"resultfordashboard":                          "result_for_dashboard",
	"qrpayment":                                   "qr_payment",
	"salechannel":                                 "sale_channel",
	"saledebitnote":                               "sale_debit_note",
	"saledebitnote_transaction":                   "sale_debit_note_transaction",
	"saledebitnote_transaction_detail":            "sale_debit_note_transaction_detail",
	"saleinvoice":                                 "sale_invoice",
	"saleinvoice_return_transaction":              "sale_invoice_return_transaction",
	"saleinvoice_return_transaction_detail":       "sale_invoice_return_transaction_detail",
	"saleinvoice_transaction":                     "sale_invoice_transaction",
	"saleinvoice_transaction_detail":              "sale_invoice_transaction_detail",
	"saleorder":                                   "sale_order",
	"saleorder_transaction":                       "sale_order_transaction",
	"saleorder_transaction_detail":                "sale_order_transaction_detail",
	"stockadjustment":                             "stock_adjustment",
	"stockbalanceimport":                          "stock_balance_import",
	"stockwaitprocess":                            "stock_wait_process",
	"userlogin":                                   "user_login",
	"withdrawalrecord":                            "withdrawal_record",
}

var legacyStorageFieldNameAliases = map[string]string{
	"custcode":       "cust_code",
	"docdate":        "doc_date",
	"docdatetime":    "doc_datetime",
	"docno":          "doc_no",
	"itemcode":       "item_code",
	"locationcode":   "location_code",
	"shopid":         "shop_id",
	"transflag":      "trans_flag",
	"unitcode":       "unit_code",
	"warehousecode":  "warehouse_code",
	"whcode":         "wh_code",
	"branchcode":     "branch_code",
	"departmentcode": "department_code",
}

// NormalizeStorageName keeps database object names in lower snake_case.
func NormalizeStorageName(name string) string {
	return normalizeStorageIdentifier(name, legacyStorageNameAliases)
}

// NormalizeStorageFieldName keeps database field and column names in lower snake_case.
func NormalizeStorageFieldName(name string) string {
	return normalizeStorageIdentifier(name, legacyStorageFieldNameAliases)
}

func normalizeStorageIdentifier(name string, aliases map[string]string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return name
	}

	runes := []rune(name)
	var builder strings.Builder
	for index, current := range runes {
		if current == '_' || current == '-' || unicode.IsSpace(current) {
			if builder.Len() > 0 && !strings.HasSuffix(builder.String(), "_") {
				builder.WriteRune('_')
			}
			continue
		}

		if unicode.IsUpper(current) {
			if index > 0 && builder.Len() > 0 && !strings.HasSuffix(builder.String(), "_") {
				previous := runes[index-1]
				var next rune
				if index+1 < len(runes) {
					next = runes[index+1]
				}
				isLastSingleLowerAfterAcronym := unicode.IsUpper(previous) && next != 0 && unicode.IsLower(next) && index+1 == len(runes)-1
				if unicode.IsLower(previous) || unicode.IsDigit(previous) || (unicode.IsUpper(previous) && next != 0 && unicode.IsLower(next) && !isLastSingleLowerAfterAcronym) {
					builder.WriteRune('_')
				}
			}
			builder.WriteRune(unicode.ToLower(current))
			continue
		}

		builder.WriteRune(unicode.ToLower(current))
	}

	normalized := strings.Trim(builder.String(), "_")
	if alias, ok := aliases[normalized]; ok {
		return alias
	}
	return normalized
}

func NormalizeMongoCollectionName(name string) string {
	return NormalizeStorageName(name)
}

func NormalizePostgresTableName(name string) string {
	return NormalizeStorageName(name)
}

func NormalizeClickHouseTableName(name string) string {
	return NormalizeStorageName(name)
}
