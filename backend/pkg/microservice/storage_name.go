package microservice

import (
	"strings"
	"unicode"
)

var legacyStorageNameAliases = map[string]string{
	"advancepayment":                          "advancepayment",
	"advancepaymentrefund":                    "advancepaymentrefund",
	"apadvancepaymentrefundtransaction":       "apadvancepaymentrefundtransaction",
	"apadvancepaymentrefundtransactiondetail": "apadvancepaymentrefundtransactiondetail",
	"apadvancepaymenttransaction":             "apadvancepaymenttransaction",
	"apadvancepaymenttransactiondetail":       "apadvancepaymenttransactiondetail",
	"apdepositpaymentrefundtransaction":       "apdepositpaymentrefundtransaction",
	"apdepositpaymentrefundtransactiondetail": "apdepositpaymentrefundtransactiondetail",
	"apdepositpaymenttransaction":             "apdepositpaymenttransaction",
	"apdepositpaymenttransactiondetail":       "apdepositpaymenttransactiondetail",
	"appurchasereceivetransaction":            "appurchasereceivetransaction",
	"appurchasereceivetransactiondetail":      "appurchasereceivetransactiondetail",
	"aradvancepaymentrefundtransaction":       "aradvancepaymentrefundtransaction",
	"aradvancepaymentrefundtransactiondetail": "aradvancepaymentrefundtransactiondetail",
	"aradvancepaymenttransaction":             "aradvancepaymenttransaction",
	"aradvancepaymenttransactiondetail":       "aradvancepaymenttransactiondetail",
	"ardepositpaymentrefundtransaction":       "ardepositpaymentrefundtransaction",
	"ardepositpaymentrefundtransactiondetail": "ardepositpaymentrefundtransactiondetail",
	"ardepositpaymenttransaction":             "ardepositpaymenttransaction",
	"ardepositpaymenttransactiondetail":       "ardepositpaymenttransactiondetail",
	"bankmaster":                              "bankmaster",
	"banktransferrecord":                      "banktransferrecord",
	"bookbank":                                "bookbank",
	"cartorder":                               "cartorder",
	"cartorderdetail":                         "cartorderdetail",
	"chartofaccounts":                         "chartofaccounts",
	"chequechange":                            "chequechange",
	"chequedeposit":                           "chequedeposit",
	"chequedisqualified":                      "chequedisqualified",
	"chequepass":                              "chequepass",
	"chequepaymentchange":                     "chequepaymentchange",
	"chequepaymentdeposit":                    "chequepaymentdeposit",
	"chequepaymentdisqualified":               "chequepaymentdisqualified",
	"chequepaymentreturn":                     "chequepaymentreturn",
	"chequerenew":                             "chequerenew",
	"chequereturn":                            "chequereturn",
	"costcenter":                              "costcenter",
	"creditcardwithdrawal":                    "creditcardwithdrawal",
	"depositrecord":                           "depositrecord",
	"depositrefund":                           "depositrefund",
	"docdetail":                               "docdetail",
	"docdetailupdated":                        "docdetailupdated",
	"docpayment":                              "docpayment",
	"docref":                                  "docref",
	"docwaitprocess":                          "docwaitprocess",
	"inventoryoptions":                        "inventoryoptions",
	"jobproject":                              "jobproject",
	"journalvatsdetails":                      "journalvatsdetails",
	"journaltaxesdetails":                     "journaltaxesdetails",
	"ordertype":                               "ordertype",
	"paymentmaster":                           "paymentmaster",
	"paidadvance":                             "paidadvance",
	"paidadvancerefund":                       "paidadvancerefund",
	"pickandpack":                             "pickandpack",
	"processstock":                            "processstock",
	"processstockcost":                        "processstockcost",
	"processstockdetail":                      "processstockdetail",
	"processstocklot":                         "processstocklot",
	"productbarcode":                          "productbarcodes",
	"productbarcodedict":                      "productbarcodedict",
	"productbarcodeboms":                      "productbarcodeboms",
	"productbarcodeimport":                    "productbarcodeimport",
	"productbarcodeprocess":                   "productbarcodeprocess",
	"productbarcoderef":                       "productbarcoderef",
	"productcategory":                         "productcategory",
	"producttype":                             "producttype",
	"productunit":                             "productunit",
	"purchaseorder":                           "purchaseorder",
	"purchaseordertransaction":                "purchaseordertransaction",
	"purchaseordertransactiondetail":          "purchaseordertransactiondetail",
	"purchasepartial":                         "purchasepartial",
	"purchasereceivetransaction":              "purchasereceivetransaction",
	"purchasereceivetransactiondetail":        "purchasereceivetransactiondetail",
	"purchaserequisition":                     "purchaserequisition",
	"purchasedebitnote":                       "purchasedebitnote",
	"receivedeposit":                          "receivedeposit",
	"receivedepositrefund":                    "receivedepositrefund",
	"resultfordashboard":                      "resultfordashboard",
	"qrpayment":                               "qrpayment",
	"salechannel":                             "salechannel",
	"saledebitnote":                           "saledebitnote",
	"saledebitnotetransaction":                "saledebitnotetransaction",
	"saledebitnotetransactiondetail":          "saledebitnotetransactiondetail",
	"saleinvoice":                             "saleinvoice",
	"saleinvoicereturntransaction":            "saleinvoicereturntransaction",
	"saleinvoicereturntransactiondetail":      "saleinvoicereturntransactiondetail",
	"saleinvoicetransaction":                  "saleinvoicetransaction",
	"saleinvoicetransactiondetail":            "saleinvoicetransactiondetail",
	"saleorder":                               "saleorder",
	"saleordertransaction":                    "saleordertransaction",
	"saleordertransactiondetail":              "saleordertransactiondetail",
	"stockadjustment":                         "stockadjustment",
	"stockbalanceimport":                      "stockbalanceimport",
	"stockwaitprocess":                        "stockwaitprocess",
	"userlogin":                               "userlogin",
	"withdrawalrecord":                        "withdrawalrecord",
}

var legacyStorageFieldNameAliases = map[string]string{
	"custcode":       "custcode",
	"docdate":        "docdate",
	"docdatetime":    "docdatetime",
	"docno":          "docno",
	"itemcode":       "itemcode",
	"locationcode":   "locationcode",
	"holdingcode":    "holdingcode",
	"transflag":      "transflag",
	"unitcode":       "unitcode",
	"warehousecode":  "warehousecode",
	"whcode":         "whcode",
	"branchcode":     "branchcode",
	"departmentcode": "departmentcode",
}

// NormalizeStorageName keeps database object names in lowercase no-underscore.
func NormalizeStorageName(name string) string {
	return normalizeStorageIdentifier(name, legacyStorageNameAliases)
}

// NormalizeStorageFieldName keeps database field and column names in lowercase no-underscore.
func NormalizeStorageFieldName(name string) string {
	return normalizeStorageIdentifier(name, legacyStorageFieldNameAliases)
}

func normalizeStorageIdentifier(name string, aliases map[string]string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return name
	}

	var builder strings.Builder
	for _, current := range name {
		if current == '_' || current == '-' || unicode.IsSpace(current) {
			continue
		}

		if unicode.IsUpper(current) {
			builder.WriteRune(unicode.ToLower(current))
			continue
		}

		builder.WriteRune(unicode.ToLower(current))
	}

	normalized := builder.String()
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
