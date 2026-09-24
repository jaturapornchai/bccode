package fixedasset

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"regexp"
	"strconv"
	"strings"
	"testing"

	gl "smlcloudplatform/internal/generalledger"
)

func faUserError(t *testing.T, err error) *gl.UserError {
	t.Helper()
	user, ok := gl.AsUserError(err)
	if !ok {
		t.Fatalf("expected a user error with code and field, got %v", err)
	}
	return user
}

// Accounts come from the asset, then its asset type; nothing is guessed (bug 2026-09-24).
func TestResolveAssetAccountsNeverGuesses(t *testing.T) {
	machinery := &AssetType{AssetAccountCode: "12110", AccumDeprecAccountCode: "12120", DeprecExpenseAccountCode: "52100"}
	got := resolveAssetAccounts(Asset{AssetAccountCode: " 12130 "}, machinery)
	if want := (assetAccounts{cost: "12130", accum: "12120", expense: "52100"}); got != want {
		t.Fatalf("asset account must win, type must fill the blanks: got %+v want %+v", got, want)
	}
	for _, assetType := range []*AssetType{nil, {}} {
		if got := resolveAssetAccounts(Asset{AssetCode: "FA-01"}, assetType); got != (assetAccounts{}) {
			t.Fatalf("an account set nowhere must stay blank (never a default code), got %+v", got)
		}
	}
}

func TestAssetAccountMissingNamesFieldAndAssets(t *testing.T) {
	codes := make([]string, 12)
	for i := range codes {
		codes[i] = fmt.Sprintf("FA-%02d", i+1)
	}
	user := faUserError(t, assetAccountMissing("deprecexpenseaccountcode", labelExpenseAccount, codes))
	if user.Code != "fa_account_required" || user.Field != "deprecexpenseaccountcode" || user.HTTPStatus() != 400 {
		t.Fatalf("got %#v", user)
	}
	for _, want := range []string{labelExpenseAccount, "FA-01", "FA-10", "และอีก 2 รายการ", "ประเภทสินทรัพย์"} {
		if !strings.Contains(user.Message, want) {
			t.Fatalf("message %q lacks %q", user.Message, want)
		}
	}
	if strings.Contains(user.Message, "FA-11") {
		t.Fatalf("message must list at most 10 assets: %q", user.Message)
	}
}

// doc_no is VARCHAR(30) counted in runes: Thai vowels and tone marks are runes of their own.
func TestCheckJournalDocNoCountsRunes(t *testing.T) {
	thai30 := strings.Repeat("กี", 15) // 30 runes, 90 bytes
	if err := checkJournalDocNo(thai30, "journaldocno", true); err != nil {
		t.Fatalf("30 runes must pass: %v", err)
	}
	long := "GJ-DISP-" + strings.Repeat("ที่", 8) // 8 + 24 = 32 runes
	user := faUserError(t, checkJournalDocNo(long, "journaldocno", true))
	if user.Code != "code_too_long" || user.Field != "journaldocno" {
		t.Fatalf("generated overflow = %#v", user)
	}
	for _, want := range []string{long, "32", "30", "ระบุเลขที่ใบสำคัญเอง"} {
		if !strings.Contains(user.Message, want) {
			t.Fatalf("message %q lacks %q", user.Message, want)
		}
	}
	typed := faUserError(t, checkJournalDocNo(long, "docno", false))
	if typed.Code != "code_too_long" || typed.Field != "docno" || strings.Contains(typed.Message, "ระบบสร้างให้") {
		t.Fatalf("a number the user typed keeps the GL message: %#v", typed)
	}
	spaced := faUserError(t, checkJournalDocNo("GJ-DISP-FA 01", "journaldocno", true))
	if spaced.Code != "code_has_space" || !strings.Contains(spaced.Message, "ระบบสร้างให้") {
		t.Fatalf("generated number with a space = %#v", spaced)
	}
}

func TestJournalBranchFollowsTheChecker(t *testing.T) {
	ctx := context.Background()
	if _, err := (&GLPoster{}).journalBranch(ctx, Scope{Branch: "00000"}, ""); !errors.Is(err, errBranchCheckUnavailable) {
		t.Fatalf("no checker must fail closed, got %v", err)
	}
	var checked []string
	p := &GLPoster{checkBranch: func(_ context.Context, _ Scope, branch string) error {
		checked = append(checked, branch)
		return nil
	}}
	if got, err := p.journalBranch(ctx, Scope{Branch: "00000"}, ""); err != nil || got != "00000" {
		t.Fatalf("blank must become the session branch: %q, %v", got, err)
	}
	if got, err := p.journalBranch(ctx, Scope{}, " 00001 "); err != nil || got != "00001" {
		t.Fatalf("explicit branch = %q, %v", got, err)
	}
	if strings.Join(checked, ",") != ",00001" {
		t.Fatalf("checker must see the normalized header branch, saw %q", checked)
	}
	p.checkBranch = func(context.Context, Scope, string) error { return gl.JournalBranchRequired() }
	if user := faUserError(t, func() error { _, err := p.journalBranch(ctx, Scope{}, ""); return err }()); user.Code != "journal_branch_required" {
		t.Fatalf("checker error must pass through, got %#v", user)
	}
}

func TestDisposalBranchMessagesNameTheAsset(t *testing.T) {
	ctx := context.Background()
	down := errors.New("registry down")
	p := &GLPoster{checkBranch: func(_ context.Context, s Scope, branch string) error {
		switch {
		case branch == "BROKEN":
			return down
		case branch == "" && s.Branch == "":
			return gl.JournalBranchRequired()
		case branch != "" && s.Branch != "" && branch != s.Branch:
			return gl.JournalBranchOutsideSession(branch, s.Branch)
		}
		return nil
	}}
	required := faUserError(t, func() error { _, err := p.disposalBranch(ctx, Scope{}, Asset{AssetCode: "FA-01"}); return err }())
	if required.Code != "journal_branch_required" || required.Field != "branchcode" || !strings.Contains(required.Message, "FA-01") || !strings.Contains(required.Message, "กำหนดสาขาในข้อมูลสินทรัพย์") {
		t.Fatalf("company session + asset without branch = %#v", required)
	}
	outside := faUserError(t, func() error {
		_, err := p.disposalBranch(ctx, Scope{Branch: "00000"}, Asset{AssetCode: "FA-02", BranchCode: "00001"})
		return err
	}())
	for _, want := range []string{"FA-02", "00001", "00000"} {
		if outside.Code != "journal_branch_outside_session" || !strings.Contains(outside.Message, want) {
			t.Fatalf("asset in another branch = %#v (want %q)", outside, want)
		}
	}
	if got, err := p.disposalBranch(ctx, Scope{Branch: "00000"}, Asset{AssetCode: "FA-03"}); err != nil || got != "00000" {
		t.Fatalf("asset without branch in a branch session = %q, %v", got, err)
	}
	if _, err := p.disposalBranch(ctx, Scope{}, Asset{AssetCode: "FA-04", BranchCode: "BROKEN"}); !errors.Is(err, down) {
		t.Fatalf("other checker errors must pass through, got %v", err)
	}
}

// Regression guard for the 2026-09-24 bug: no default account codes, no "JV" document
// prefix and no float conversion of money anywhere in the poster.
func TestGLPosterSourceHasNoGuessedAccountsOrFloatMoney(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "gl_poster.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	accountLike := regexp.MustCompile(`^[0-9]{4,}$`)
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.BasicLit:
			if node.Kind != token.STRING {
				return true
			}
			value, err := strconv.Unquote(node.Value)
			if err != nil {
				return true
			}
			if accountLike.MatchString(value) || strings.HasPrefix(value, "JV") {
				t.Errorf("gl_poster.go must not hard-code %q (charts and book codes differ per company)", value)
			}
		case *ast.SelectorExpr:
			if strings.HasPrefix(node.Sel.Name, "InexactFloat") {
				t.Errorf("gl_poster.go must not convert money to float (%s)", node.Sel.Name)
			}
		}
		return true
	})
}
