package reportstock

import (
	"fmt"
	"smlcloudplatform/internal/goapi/language"
	"time"
)

var reportNameKeys = map[string]string{
	"product_balance":              "report_product_balance",
	"product_balance_by_location":  "report_product_balance_by_location",
	"product_balance_by_warehouse": "report_product_balance_by_warehouse",
	"product_stock_movement":       "report_product_stock_movement",
}

var headerKeys = map[string]string{
	"report_as_of": "report_as_of",
}

var columnKeys = map[string]string{
	"avg_cost":       "avg_cost",
	"balance":        "balance",
	"balance_unit":   "balance_unit",
	"balance_value":  "balance_value",
	"balance_word":   "balance_word",
	"barcode_list":   "barcode_list",
	"code":           "code",
	"cost":           "cost",
	"date":           "date",
	"decrease_qty":   "decrease_qty",
	"decrease_value": "decrease_value",
	"document":       "document",
	"increase_qty":   "increase_qty",
	"increase_value": "increase_value",
	"item_code":      "item_code",
	"location":       "location",
	"product_name":   "product_name",
	"quantity":       "quantity",
	"total_value":    "total_value",
	"type":           "type",
	"unit":           "unit",
	"warehouse":      "warehouse",
}

var commonKeys = map[string]string{
	"grand_total": "grand_total",
	"of":          "of",
	"page":        "page",
	"subtotal":    "subtotal",
	"total":       "total",
}

// GetText returns translated report text from the shared backend language TSV.
func GetText(section, key, languageCode string) string {
	var tsvKey string
	switch section {
	case "report_names":
		tsvKey = reportNameKeys[key]
	case "headers":
		tsvKey = headerKeys[key]
	case "columns":
		tsvKey = columnKeys[key]
	case "common":
		tsvKey = commonKeys[key]
	default:
		return key
	}
	if tsvKey == "" {
		return key
	}
	return language.Text(tsvKey, languageCode)
}

// GetReportName returns the report name in specified language.
func GetReportName(reportType, languageCode string) string {
	return GetText("report_names", reportType, languageCode)
}

// GetHeaderText returns the header text in specified language.
func GetHeaderText(key, languageCode string) string {
	return GetText("headers", key, languageCode)
}

// GetColumnText returns the column text in specified language.
func GetColumnText(key, languageCode string) string {
	return GetText("columns", key, languageCode)
}

// GetCommonText returns common text in specified language.
func GetCommonText(key, languageCode string) string {
	return GetText("common", key, languageCode)
}

// FormatDateWithLanguage formats a date according to the report language.
func FormatDateWithLanguage(date time.Time, languageCode string) string {
	switch language.Normalize(languageCode) {
	case "th":
		months := []string{
			"มกราคม", "กุมภาพันธ์", "มีนาคม", "เมษายน", "พฤษภาคม", "มิถุนายน",
			"กรกฎาคม", "สิงหาคม", "กันยายน", "ตุลาคม", "พฤศจิกายน", "ธันวาคม",
		}
		return fmt.Sprintf("%d %s %d", date.Day(), months[date.Month()-1], date.Year()+543)
	case "cn":
		return date.Format("2006年01月02日")
	case "ja":
		return date.Format("2006年01月02日")
	case "ko":
		return date.Format("2006년 01월 02일")
	default:
		return date.Format("January 02, 2006")
	}
}

// LoadLanguageDict is kept for backward compatibility with old report code.
func LoadLanguageDict() error {
	return language.Load()
}

// GetReportLanguageDict returns simplified access to report labels for one language.
func GetReportLanguageDict(languageCode string) map[string]interface{} {
	return map[string]interface{}{
		"common":       sectionDictionary(commonKeys, languageCode),
		"columns":      sectionDictionary(columnKeys, languageCode),
		"headers":      sectionDictionary(headerKeys, languageCode),
		"report_names": sectionDictionary(reportNameKeys, languageCode),
	}
}

func sectionDictionary(keys map[string]string, languageCode string) map[string]string {
	result := make(map[string]string, len(keys))
	for key, tsvKey := range keys {
		result[key] = language.Text(tsvKey, languageCode)
	}
	return result
}
