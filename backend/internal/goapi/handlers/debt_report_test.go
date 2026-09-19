package handlers

import (
	"strings"
	"testing"
	"time"
)

func TestBuildDebtReportQuery(t *testing.T) {
	fromDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	toDate := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		reportCode string
		partyCode  string
		hasDate    bool
		wantSub    string
		wantArgs   int
	}{
		{
			reportCode: "ap_movement",
			partyCode:  "AP001",
			hasDate:    true,
			wantSub:    "public.creditortransaction",
			wantArgs:   5, // fromDate, toDate, partyCode, limit, offset
		},
		{
			reportCode: "ap_status",
			partyCode:  "AP002",
			hasDate:    false,
			wantSub:    "public.creditor",
			wantArgs:   3, // partyCode, limit, offset
		},
		{
			reportCode: "ap_outstanding",
			partyCode:  "",
			hasDate:    false,
			wantSub:    "t.balanceamount > 0",
			wantArgs:   2, // limit, offset
		},
		{
			reportCode: "ap_daily_payment",
			partyCode:  "AP003",
			hasDate:    true,
			wantSub:    "t.paidamount > 0",
			wantArgs:   5, // fromDate, toDate, partyCode, limit, offset
		},
		{
			reportCode: "ar_movement",
			partyCode:  "AR001",
			hasDate:    true,
			wantSub:    "public.debtortransaction",
			wantArgs:   5, // fromDate, toDate, partyCode, limit, offset
		},
		{
			reportCode: "ar_status",
			partyCode:  "AR002",
			hasDate:    false,
			wantSub:    "public.debtor",
			wantArgs:   3, // partyCode, limit, offset
		},
		{
			reportCode: "ar_outstanding",
			partyCode:  "",
			hasDate:    false,
			wantSub:    "t.balanceamount > 0",
			wantArgs:   2, // limit, offset
		},
		{
			reportCode: "ar_credit_limit",
			partyCode:  "AR003",
			hasDate:    false,
			wantSub:    "creditday",
			wantArgs:   3, // partyCode, limit, offset
		},
	}

	for _, tt := range tests {
		t.Run(tt.reportCode, func(t *testing.T) {
			fd := time.Time{}
			td := time.Time{}
			if tt.hasDate {
				fd = fromDate
				td = toDate
			}

			q, args, err := BuildDebtReportQuery(tt.reportCode, "biz01", fd, td, tt.partyCode, 100, 0)
			if err != nil {
				t.Fatalf("unexpected error for %s: %v", tt.reportCode, err)
			}
			if !strings.Contains(q, tt.wantSub) {
				t.Errorf("%s: query missing substring %q\nQuery: %s", tt.reportCode, tt.wantSub, q)
			}
			if len(args) != tt.wantArgs {
				t.Errorf("%s: expected %d args, got %d", tt.reportCode, tt.wantArgs, len(args))
			}
		})
	}

	// Test invalid report code
	_, _, err := BuildDebtReportQuery("invalid_report", "biz01", time.Time{}, time.Time{}, "", 100, 0)
	if err == nil {
		t.Fatal("expected error for invalid report code, got nil")
	}
}
