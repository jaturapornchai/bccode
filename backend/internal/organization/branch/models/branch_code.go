package models

import (
	"fmt"
	"strings"
)

const ThaiHeadOfficeBranchCode = "00000"

// NormalizeThaiTaxBranchCode returns the five-digit Thai VAT branch code used in
// tax invoices and VAT reports. Head office is always 00000; 00001 is branch 1.
func NormalizeThaiTaxBranchCode(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("branch code is required")
	}

	compact := strings.ToLower(strings.Join(strings.Fields(trimmed), ""))
	switch compact {
	case "สำนักงานใหญ่", "สํานักงานใหญ่", "สนญ", "head-office", "headoffice", "hq", "ho":
		return ThaiHeadOfficeBranchCode, nil
	}

	for _, r := range compact {
		if r < '0' || r > '9' {
			return "", fmt.Errorf("branch code must be numeric and no more than 5 digits")
		}
	}
	if len(compact) > 5 {
		return "", fmt.Errorf("branch code must be no more than 5 digits")
	}

	return fmt.Sprintf("%05s", compact), nil
}

func IsThaiHeadOfficeBranchCode(value string) bool {
	normalized, err := NormalizeThaiTaxBranchCode(value)
	return err == nil && normalized == ThaiHeadOfficeBranchCode
}
