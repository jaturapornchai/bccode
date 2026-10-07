package generalledger

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"
)

type processReader interface {
	ProcessBalances(context.Context, Scope, string, string) (ProcessBalanceSnapshot, error)
	HasDraftJournals(context.Context, Scope, string, string) (bool, error)
	HasOpeningJournal(context.Context, Scope, string) (bool, error)
}

// preparedProcess is the plan planProcess builds: the vouchers to create (one per branch) and
// what the preview needs to explain them.
type preparedProcess struct {
	Sequence  int64
	Journals  []Journal
	Year      FiscalYear
	CloseYear bool
	// To is the last date included in the balances (the document date for a close).
	To string
	// PendingDrafts: drafts dated up to To still exist; the real command refuses, the preview only warns.
	PendingDrafts bool
	// TransferFrom: per branch, the index of the first profit-transfer line in that branch's voucher.
	// A branch without an entry has no transfer lines.
	TransferFrom map[string]int
	// NetProfit: per branch, credit minus debit of the income and expense lines being closed.
	NetProfit map[string]decimal.Decimal
}

func signedLine(code string, value decimal.Decimal, description string) (Line, error) {
	line := Line{AccountCode: code, Description: description}
	amount, err := ParseAmount(value.Abs().String())
	if err != nil {
		return line, fmt.Errorf("ยอดรวมเกินขนาดจำนวนเงินที่รองรับ กรุณาตรวจสอบบัญชี %s", code)
	}
	if value.IsNegative() {
		line.Credit = amount
	} else {
		line.Debit = amount
	}
	return line, nil
}
