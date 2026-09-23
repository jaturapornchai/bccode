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

type preparedProcess struct {
	Sequence  int64
	Journals  []Journal
	Year      FiscalYear
	CloseYear bool
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
