package models

import (
	"encoding/json"
	"testing"
)

func TestBranchAcceptsMovedCompanyAndBranchFields(t *testing.T) {
	payload := []byte(`{
		"code":"00000",
		"companyregistrationno":"0100000000000",
		"isvatregistered":true,
		"contact":{
			"countrycode":"TH",
			"provincecode":"10",
			"districtcode":"1001",
			"subdistrictcode":"100101",
			"zipcode":"10200",
			"phonenumber":"021234567"
		},
		"pos":{"taxid":"0100000000000"},
		"yeartype":"buddhist",
		"isretail":true,
		"isservice":true,
		"ismobileshop":true
	}`)

	var branch Branch
	if err := json.Unmarshal(payload, &branch); err != nil {
		t.Fatalf("unmarshal branch: %v", err)
	}

	if branch.CompanyRegistrationNo != "0100000000000" {
		t.Fatalf("company registration no = %q", branch.CompanyRegistrationNo)
	}
	if !branch.IsVatRegistered {
		t.Fatal("expected VAT registration flag")
	}
	if branch.Contact.PhoneNumber != "021234567" {
		t.Fatalf("phone number = %q", branch.Contact.PhoneNumber)
	}
	if branch.POS.TaxID != "0100000000000" {
		t.Fatalf("tax id = %q", branch.POS.TaxID)
	}
	if branch.YearType != "buddhist" {
		t.Fatalf("year type = %q", branch.YearType)
	}
	if !branch.IsRetail || !branch.IsService || !branch.IsMobileShop {
		t.Fatalf("business flags not preserved: retail=%v service=%v mobile=%v", branch.IsRetail, branch.IsService, branch.IsMobileShop)
	}
}

func TestNormalizeThaiTaxBranchCode(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "head office numeric", input: "00000", want: "00000"},
		{name: "head office thai", input: "สำนักงานใหญ่", want: "00000"},
		{name: "head office thai abbreviation", input: "สนญ", want: "00000"},
		{name: "branch one", input: "1", want: "00001"},
		{name: "branch one padded", input: "01", want: "00001"},
		{name: "branch existing five digits", input: "12345", want: "12345"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeThaiTaxBranchCode(tt.input)
			if err != nil {
				t.Fatalf("NormalizeThaiTaxBranchCode(%q) error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("NormalizeThaiTaxBranchCode(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormalizeThaiTaxBranchCodeRejectsInvalidValues(t *testing.T) {
	for _, input := range []string{"", "abc", "000000", "สาขาหนึ่ง"} {
		t.Run(input, func(t *testing.T) {
			if got, err := NormalizeThaiTaxBranchCode(input); err == nil {
				t.Fatalf("NormalizeThaiTaxBranchCode(%q) = %q, want error", input, got)
			}
		})
	}
}
