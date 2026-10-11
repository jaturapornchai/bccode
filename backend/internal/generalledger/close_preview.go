package generalledger

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/shopspring/decimal"
)

// ClosePreviewAction is the processes command that only shows the closing entries.
const ClosePreviewAction = "close-preview"

// Champ 5504 shows the closing entries before they are saved. ClosePreview is that table for processes action
// "close-preview" (nothing is saved): the vouchers the real "close" command would create as drafts, per branch.
// Amounts are decimal strings with the fiscal-year scale; Balanced proves debit = credit over every line.
type ClosePreview struct {
	FiscalYear    string               `json:"fiscalyear"`
	Date          string               `json:"date"`
	Scale         int                  `json:"scale"`
	PendingDrafts bool                 `json:"pendingdrafts"`
	NetProfit     string               `json:"netprofit"`
	TotalDebit    string               `json:"totaldebit"`
	TotalCredit   string               `json:"totalcredit"`
	Balanced      bool                 `json:"balanced"`
	Branches      []ClosePreviewBranch `json:"branches"`
}

// ClosePreviewBranch is one voucher. DocNo is empty until the user has typed a document number.
type ClosePreviewBranch struct {
	BranchCode  string             `json:"branchcode"`
	DocNo       string             `json:"docno"`
	NetProfit   string             `json:"netprofit"`
	TotalDebit  string             `json:"totaldebit"`
	TotalCredit string             `json:"totalcredit"`
	Lines       []ClosePreviewLine `json:"lines"`
}

// ClosePreviewLine: Transfer marks the lines that move the net profit (loss) to retained earnings.
type ClosePreviewLine struct {
	AccountCode    string `json:"accountcode"`
	AccountName    string `json:"accountname"`
	DepartmentCode string `json:"departmentcode"`
	ProjectCode    string `json:"projectcode"`
	AllocateCode   string `json:"allocatecode"`
	JobCode        string `json:"jobcode"`
	PartCode       string `json:"partcode"`
	SideCode       string `json:"sidecode"`
	Debit          string `json:"debit"`
	Credit         string `json:"credit"`
	Transfer       bool   `json:"transfer"`
}

// previewClose runs planProcess — the generator behind the real "close" — in a read-only transaction
// and shows its vouchers; the database refuses any write, so nothing can be saved. Document number
// and reason are optional here (the user types them after seeing the entries).
func (s *PostgresStore) previewClose(ctx context.Context, scope Scope, cmd Command) (Result, error) {
	if scope.Branch != "" {
		return Result{}, fmt.Errorf("การประมวลผลต้องใช้สิทธิ์ครอบคลุมทุกสาขา")
	}
	if cmd.ID == "" {
		return Result{}, fmt.Errorf("กรุณาเลือกรายการบัญชี")
	}
	docNoGiven := cmd.DocNo != ""
	if !docNoGiven {
		cmd.DocNo = "PREVIEW"
	}
	cmd.Action = "close"
	db, err := s.pg.database(ctx, scope.Holding)
	if err != nil {
		return Result{}, err
	}
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return Result{}, err
	}
	defer tx.Rollback()
	plan, err := s.planProcess(ctx, tx, scope, cmd, true)
	if err != nil {
		return Result{}, err
	}
	preview := closePreviewOf(plan, docNoGiven)
	return Result{ClosePreview: &preview}, nil
}

func closePreviewOf(plan *preparedProcess, docNoGiven bool) ClosePreview {
	scale := int32(plan.Year.Scale)
	preview := ClosePreview{FiscalYear: plan.Year.Code, Date: plan.To, Scale: plan.Year.Scale, PendingDrafts: plan.PendingDrafts, Branches: make([]ClosePreviewBranch, 0, len(plan.Journals))}
	totalDebit, totalCredit, totalProfit := decimal.Zero, decimal.Zero, decimal.Zero
	for _, j := range plan.Journals {
		branch := ClosePreviewBranch{BranchCode: j.BranchCode, Lines: make([]ClosePreviewLine, 0, len(j.Lines))}
		if docNoGiven {
			branch.DocNo = j.DocNo
		}
		transferFrom, hasTransfer := plan.TransferFrom[j.BranchCode]
		debit, credit := decimal.Zero, decimal.Zero
		for i, line := range j.Lines {
			d, c := line.Debit.Decimal(), line.Credit.Decimal()
			debit, credit = debit.Add(d), credit.Add(c)
			branch.Lines = append(branch.Lines, ClosePreviewLine{AccountCode: line.AccountCode, AccountName: line.AccountName, DepartmentCode: line.DepartmentCode, ProjectCode: line.ProjectCode, AllocateCode: line.AllocateCode, JobCode: line.JobCode, PartCode: line.PartCode, SideCode: line.SideCode, Debit: d.StringFixed(scale), Credit: c.StringFixed(scale), Transfer: hasTransfer && i >= transferFrom})
		}
		profit := plan.NetProfit[j.BranchCode]
		branch.NetProfit, branch.TotalDebit, branch.TotalCredit = profit.StringFixed(scale), debit.StringFixed(scale), credit.StringFixed(scale)
		totalDebit, totalCredit, totalProfit = totalDebit.Add(debit), totalCredit.Add(credit), totalProfit.Add(profit)
		preview.Branches = append(preview.Branches, branch)
	}
	preview.NetProfit, preview.TotalDebit, preview.TotalCredit = totalProfit.StringFixed(scale), totalDebit.StringFixed(scale), totalCredit.StringFixed(scale)
	preview.Balanced = totalDebit.Equal(totalCredit)
	return preview
}
