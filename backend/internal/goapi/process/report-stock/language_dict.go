package reportstock

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/labstack/gommon/log"
)

// ReportLabels represents the simplified language dictionary structure
type ReportLabels struct {
	ReportNames map[string]map[string]string `json:"report_names"`
	Headers     map[string]map[string]string `json:"headers"`
	Columns     map[string]map[string]string `json:"columns"`
	Common      map[string]map[string]string `json:"common"`
}

var reportLabels *ReportLabels

// GetText returns the text for a given key in the specified language
// รับข้อความตาม section, key และภาษาที่กำหนด
//
// Field Keys ที่ใช้ได้:
// - Report Names: "product_balance", "product_balance_by_warehouse", "product_balance_by_location", "product_stock_movement"
// - Headers: "report_as_of" (รายงาน ณ. วันที่)
// - Columns:
//   - "code" (รหัส/Barcode), "product_name" (ชื่อสินค้า), "unit" (หน่วยนับ)
//   - "warehouse" (คลัง), "location" (ตำแหน่ง), "quantity" (จำนวน)
//   - "cost" (ต้นทุน), "avg_cost" (ต้นทุนเฉลี่ย), "total_value" (มูลค่ารวม)
//   - "balance" (คงเหลือ), "balance_value" (มูลค่าคงเหลือ)
//   - "balance_word" (ยอดคงเหลือ ตัวหนังสือ), "balance_unit" (คงเหลือ หน่วย)
//   - "increase_qty" (จำนวนเพิ่ม), "decrease_qty" (จำนวนลด)
//   - "increase_value" (มูลค่าบวก), "decrease_value" (มูลค่าลบ)
//   - "date" (วันที่), "document" (เอกสาร), "type" (ประเภท)
//
// - Common: "total" (รวม), "subtotal" (รวมย่อย), "grand_total" (รวมทั้งสิ้น), "page" (หน้า), "of" (จาก)
func GetText(section, key, language string) string {
	if reportLabels == nil {
		err := LoadLanguageDict()
		if err != nil {
			log.Error("Failed to load language dictionary:", err)
			return key
		}
	}

	var sectionData map[string]map[string]string

	switch section {
	case "report_names":
		sectionData = reportLabels.ReportNames
	case "headers":
		sectionData = reportLabels.Headers
	case "columns":
		sectionData = reportLabels.Columns
	case "common":
		sectionData = reportLabels.Common
	default:
		return key
	}

	if langData, exists := sectionData[key]; exists {
		if text, exists := langData[language]; exists {
			return text
		}
		// Fallback to Thai if requested language not found
		if text, exists := langData["th"]; exists {
			return text
		}
	}

	return key
}

// GetReportName returns the report name in specified language
// รับชื่อรายงานตามภาษาที่กำหนด
func GetReportName(reportType, language string) string {
	return GetText("report_names", reportType, language)
}

// GetHeaderText returns the header text in specified language
// รับข้อความ header ตามภาษาที่กำหนด (เช่น "รายงาน ณ. วันที่")
func GetHeaderText(key, language string) string {
	return GetText("headers", key, language)
}

// GetColumnText returns the column text in specified language
// รับชื่อคอลัมน์ตามภาษาที่กำหนด (เช่น "รหัส/Barcode", "ชื่อสินค้า", "คลัง")
func GetColumnText(key, language string) string {
	return GetText("columns", key, language)
}

// GetCommonText returns common text in specified language
// รับข้อความทั่วไปตามภาษาที่กำหนด (เช่น "รวม", "หน้า")
func GetCommonText(key, language string) string {
	return GetText("common", key, language)
}

// FormatDateWithLanguage formats date according to language
func FormatDateWithLanguage(date time.Time, language string) string {
	switch language {
	case "th":
		months := []string{
			"มกราคม", "กุมภาพันธ์", "มีนาคม", "เมษายน", "พฤษภาคม", "มิถุนายน",
			"กรกฎาคม", "สิงหาคม", "กันยายน", "ตุลาคม", "พฤศจิกายน", "ธันวาคม",
		}
		return fmt.Sprintf("%d %s %d", date.Day(), months[date.Month()-1], date.Year()+543)
	case "zh":
		return date.Format("2006年01月02日")
	default: // en
		return date.Format("January 02, 2006")
	}
}

// LoadLanguageDict loads the language dictionary from JSON file
func LoadLanguageDict() error {
	if reportLabels != nil {
		return nil
	}

	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %v", err)
	}

	// Construct the path to the JSON file
	jsonPath := filepath.Join(wd, "process", "report-stock", "lang", "report_labels.json")

	// Read the JSON file
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return fmt.Errorf("failed to read language file: %v", err)
	}

	if len(data) == 0 {
		return fmt.Errorf("language file is empty")
	}

	// Parse JSON
	tmpReportLabels := &ReportLabels{}
	if err := json.Unmarshal(data, tmpReportLabels); err != nil {
		return fmt.Errorf("failed to parse language JSON: %v", err)
	}

	reportLabels = tmpReportLabels
	return nil
}

// GetReportLanguageDict returns simplified access to language data (for backward compatibility)
func GetReportLanguageDict(languageCode string) map[string]interface{} {
	if reportLabels == nil {
		LoadLanguageDict()
	}

	result := make(map[string]interface{})

	// Add report names
	if reportLabels.ReportNames != nil {
		reportNames := make(map[string]string)
		for key, langMap := range reportLabels.ReportNames {
			if text, exists := langMap[languageCode]; exists {
				reportNames[key] = text
			} else if text, exists := langMap["th"]; exists {
				reportNames[key] = text
			} else {
				reportNames[key] = key
			}
		}
		result["report_names"] = reportNames
	}

	// Add headers
	if reportLabels.Headers != nil {
		headers := make(map[string]string)
		for key, langMap := range reportLabels.Headers {
			if text, exists := langMap[languageCode]; exists {
				headers[key] = text
			} else if text, exists := langMap["th"]; exists {
				headers[key] = text
			} else {
				headers[key] = key
			}
		}
		result["headers"] = headers
	}

	// Add columns
	if reportLabels.Columns != nil {
		columns := make(map[string]string)
		for key, langMap := range reportLabels.Columns {
			if text, exists := langMap[languageCode]; exists {
				columns[key] = text
			} else if text, exists := langMap["th"]; exists {
				columns[key] = text
			} else {
				columns[key] = key
			}
		}
		result["columns"] = columns
	}

	// Add common
	if reportLabels.Common != nil {
		common := make(map[string]string)
		for key, langMap := range reportLabels.Common {
			if text, exists := langMap[languageCode]; exists {
				common[key] = text
			} else if text, exists := langMap["th"]; exists {
				common[key] = text
			} else {
				common[key] = key
			}
		}
		result["common"] = common
	}

	return result
}
