package generalledger

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (s *Store) apply(ctx context.Context, scope Scope, cmd Command, now time.Time) ([]Change, error) {
	switch cmd.Resource {
	case "processes":
		return s.processMutation(ctx, scope, cmd, now)
	case "accounts":
		return s.accountMutation(ctx, scope, cmd, now)
	case "fiscal-years":
		return s.yearMutation(ctx, scope, cmd, now)
	case "journals":
		return s.journalMutation(ctx, scope, cmd, now)
	default:
		return s.masterMutation(ctx, scope, cmd, now)
	}
}

func (s *Store) load(ctx context.Context, scope Scope, kind, id string, target interface{}) error {
	if collectionName(kind) == "" {
		return ErrNotFound
	}
	err := s.db.Collection(collectionName(kind)).FindOne(ctx, scopedID(scope, id)).Decode(target)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return ErrNotFound
	}
	return err
}
func checkVersion(cmd Command, old Identity) error {
	if old.IsDeleted {
		return ErrNotFound
	}
	if old.Version != cmd.Version || cmd.Version < 1 {
		return userError(CodeStaleVersion, "รายการนี้เปลี่ยนไปแล้ว กรุณาโหลดข้อมูลล่าสุดก่อนบันทึก")
	}
	return nil
}
func single(c Change, err error) ([]Change, error) {
	if err != nil {
		return nil, err
	}
	return []Change{c}, nil
}
func (s *Store) referenced(ctx context.Context, scope Scope, filter bson.M) (bool, error) {
	for k, v := range scopeFilter(scope) {
		filter[k] = v
	}
	n, err := s.db.Collection("gl_journals").CountDocuments(ctx, filter)
	return n > 0, err
}

func (s *Store) accountMutation(ctx context.Context, scope Scope, cmd Command, now time.Time) ([]Change, error) {
	if !contains([]string{"create", "update", "delete"}, cmd.Action) {
		return nil, userError(CodeUnsupported, "คำสั่งผังบัญชีไม่ถูกต้อง")
	}
	var old Account
	if cmd.Action != "create" {
		if err := s.load(ctx, scope, "accounts", cmd.ID, &old); err != nil {
			return nil, err
		}
		if err := checkVersion(cmd, old.Identity); err != nil {
			return nil, err
		}
	}
	if cmd.Action == "delete" {
		if err := s.accountMasterReferences(ctx, scope, old.AccountCode); err != nil {
			return nil, err
		}
		used, err := s.referenced(ctx, scope, bson.M{"lines.accountcode": old.AccountCode})
		if err != nil {
			return nil, err
		}
		if used {
			return nil, userError(CodeReferenced, "บัญชีนี้มีข้อมูลอ้างอิงจากสมุดรายวัน ห้ามลบผังบัญชีเด็ดขาด กรุณาปิดใช้งานแทนการลบ")
		}
		f := scopeFilter(scope)
		f["parentaccountcode"] = old.AccountCode
		f["isdeleted"] = false
		n, err := s.db.Collection("chart_of_accounts").CountDocuments(ctx, f)
		if err != nil {
			return nil, err
		}
		if n > 0 {
			return nil, userError(CodeHasChildren, "บัญชีนี้มีบัญชีลูก กรุณาตรวจสอบผังบัญชีก่อน")
		}
		old.Identity = identityFor(scope, cmd, old.Identity, now)
		old.IsDeleted = true
		return single(s.save(ctx, scope, "accounts", old.ID, old.AccountCode, old, cmd.Version))
	}
	if cmd.Account == nil {
		return nil, userError(CodePayloadMissing, "ไม่พบข้อมูลผังบัญชี")
	}
	next := *cmd.Account
	if err := next.Validate(); err != nil {
		return nil, validationFailed(err)
	}
	if next.AccountGroup != "" {
		f := scopeFilter(scope)
		f["code"] = next.AccountGroup
		f["isdeleted"] = false
		f["isactive"] = true
		n, err := s.db.Collection(MasterCollections["account-groups"]).CountDocuments(ctx, f)
		if err != nil {
			return nil, err
		}
		if n == 0 {
			return nil, userError(CodeGroupNotFound, "ไม่พบกลุ่มผังบัญชีที่เปิดใช้งานในบริษัทนี้")
		}
	}
	if cmd.Action == "update" {
		if next.AccountCode != old.AccountCode {
			return nil, userError(CodeImmutableCode, "รหัสบัญชีแก้ไม่ได้ กรุณาสร้างบัญชีใหม่")
		}
		if next.AccountType != old.AccountType || next.NormalBalance != old.NormalBalance || next.IsCash != old.IsCash || next.ParentAccountCode != old.ParentAccountCode || next.AllowPosting != old.AllowPosting {
			used, err := s.referenced(ctx, scope, bson.M{"lines.accountcode": old.AccountCode, "status": bson.M{"$in": []string{"posted", "reversed"}}})
			if err != nil {
				return nil, err
			}
			if used {
				return nil, userError(CodePostedLocked, "บัญชีนี้ผ่านรายการแล้ว เปลี่ยนหมวด ด้านบัญชี หรือโครงสร้างไม่ได้")
			}
		}
	}
	seen := map[string]bool{next.AccountCode: true}
	parent := next.ParentAccountCode
	depth := 0
	parentLevel := 0
	for ; parent != ""; depth++ {
		if seen[parent] || depth >= 64 {
			return nil, userError(CodeTreeInvalid, "โครงสร้างบัญชีวนซ้ำหรือมีระดับมากเกินไป")
		}
		seen[parent] = true
		var a Account
		f := scopeFilter(scope)
		f["accountcode"] = parent
		f["isdeleted"] = false
		if err := s.db.Collection("chart_of_accounts").FindOne(ctx, f).Decode(&a); err != nil {
			return nil, userError(CodeParentNotFound, "ไม่พบบัญชีแม่ในบริษัทนี้")
		}
		if a.AllowPosting || !a.IsActive {
			return nil, userError(CodeParentInvalid, "บัญชีแม่ต้องเป็นบัญชีคุมที่เปิดใช้งาน")
		}
		if parentLevel == 0 {
			parentLevel = a.Level
			if parentLevel <= 0 {
				parentLevel = 1
			}
		}
		parent = a.ParentAccountCode
	}
	calculatedLevel := depth + 1
	if next.Level <= 0 {
		if parentLevel > 0 {
			next.Level = parentLevel + 1
		} else {
			next.Level = calculatedLevel
		}
	}
	if next.Level < 1 || next.Level > 12 {
		return nil, userError(CodeLevelRange, "ระดับบัญชีต้องอยู่ระหว่าง 1 ถึง 12")
	}
	if next.ParentAccountCode != "" && parentLevel > 0 && next.Level <= parentLevel {
		return nil, userError(CodeLevelShallow, "ระดับบัญชีต้องมากกว่าระดับของบัญชีแม่")
	}
	if cmd.Action == "update" && next.Level != old.Level {
		f := scopeFilter(scope)
		f["parentaccountcode"] = old.AccountCode
		f["isdeleted"] = false
		f["level"] = bson.M{"$lte": next.Level}
		n, err := s.db.Collection("chart_of_accounts").CountDocuments(ctx, f)
		if err != nil {
			return nil, err
		}
		if n > 0 {
			return nil, userError(CodeHasChildren, "ไม่สามารถกำหนดระดับบัญชีนี้ได้ เนื่องจากมีบัญชีลูกที่มีระดับน้อยกว่าหรือเท่ากัน")
		}
	}
	next.Identity = identityFor(scope, cmd, old.Identity, now)
	return single(s.save(ctx, scope, "accounts", next.ID, next.AccountCode, next, cmd.Version))
}

func (s *Store) yearByCode(ctx context.Context, scope Scope, code string) (FiscalYear, error) {
	var f FiscalYear
	q := scopeFilter(scope)
	q["code"] = code
	q["isdeleted"] = false
	err := s.db.Collection("fiscal_year").FindOne(ctx, q).Decode(&f)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return f, fmt.Errorf("กรุณาตั้งค่าปีบัญชีก่อนบันทึกรายการ")
	}
	return f, err
}

func (s *Store) yearMutation(ctx context.Context, scope Scope, cmd Command, now time.Time) ([]Change, error) {
	if !contains([]string{"create", "update", "delete"}, cmd.Action) {
		return nil, fmt.Errorf("คำสั่งปีบัญชีไม่ถูกต้อง")
	}
	var old FiscalYear
	if cmd.Action != "create" {
		if err := s.load(ctx, scope, "fiscal-years", cmd.ID, &old); err != nil {
			return nil, err
		}
		if err := checkVersion(cmd, old.Identity); err != nil {
			return nil, err
		}
	}
	if old.Closed {
		return nil, fmt.Errorf("ปีบัญชีนี้ปิดแล้ว เปลี่ยนข้อมูลไม่ได้")
	}
	used, err := s.referenced(ctx, scope, bson.M{"fiscalyear": old.Code, "isdeleted": false})
	if err != nil {
		return nil, err
	}
	if cmd.Action == "delete" {
		if used {
			return nil, fmt.Errorf("ปีบัญชีนี้มีรายการอ้างอิง ลบไม่ได้")
		}
		if err := s.yearMasterReferences(ctx, scope, old, nil); err != nil {
			return nil, err
		}
		old.Identity = identityFor(scope, cmd, old.Identity, now)
		old.IsDeleted = true
		return single(s.save(ctx, scope, "fiscal-years", old.ID, old.Code, old, cmd.Version))
	}
	if cmd.FiscalYear == nil {
		return nil, fmt.Errorf("ไม่พบข้อมูลปีบัญชี")
	}
	next := *cmd.FiscalYear
	next.Closed = false
	if err := next.Validate(); err != nil {
		return nil, err
	}
	if cmd.Action == "update" && (next.Code != old.Code || (used && (next.StartDate != old.StartDate || next.EndDate != old.EndDate || next.Scale != old.Scale || next.Currency != old.Currency || !next.IsActive))) {
		return nil, fmt.Errorf("ปีบัญชีมีรายการแล้ว เปลี่ยนช่วงวัน สกุลเงิน หรือทศนิยมไม่ได้")
	}
	f := scopeFilter(scope)
	f["isdeleted"] = false
	f["startdate"] = bson.M{"$lte": next.EndDate}
	f["enddate"] = bson.M{"$gte": next.StartDate}
	if cmd.Action != "create" {
		f["_id"] = bson.M{"$ne": cmd.ID}
	}
	n, err := s.db.Collection("fiscal_year").CountDocuments(ctx, f)
	if err != nil {
		return nil, err
	}
	if n > 0 {
		return nil, fmt.Errorf("ช่วงปีบัญชีทับซ้อนกับปีที่มีอยู่")
	}
	if cmd.Action == "update" {
		if err := s.yearMasterReferences(ctx, scope, old, &next); err != nil {
			return nil, err
		}
	}
	for _, code := range []string{next.ProfitLossAccount, next.RetainedEarningsAccount} {
		if code != "" {
			accounts, err := s.accounts(ctx, scope, []string{code})
			if err != nil {
				return nil, err
			}
			a, ok := accounts[code]
			if !ok || a.AccountType != "equity" || !a.AllowPosting || !a.IsActive {
				return nil, fmt.Errorf("บัญชีปิดผลกำไรต้องเป็นบัญชีทุนที่เปิดรับรายการ")
			}
		}
	}
	if next.ProfitLossAccount != "" && next.ProfitLossAccount == next.RetainedEarningsAccount {
		return nil, fmt.Errorf("บัญชีกำไรขาดทุนและกำไรสะสมต้องเป็นคนละบัญชี")
	}
	next.Identity = identityFor(scope, cmd, old.Identity, now)
	return single(s.save(ctx, scope, "fiscal-years", next.ID, next.Code, next, cmd.Version))
}

func (s *Store) accounts(ctx context.Context, scope Scope, codes []string) (map[string]Account, error) {
	f := scopeFilter(scope)
	f["accountcode"] = bson.M{"$in": codes}
	f["isdeleted"] = false
	cursor, err := s.db.Collection("chart_of_accounts").Find(ctx, f)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	result := map[string]Account{}
	for cursor.Next(ctx) {
		var a Account
		if err := cursor.Decode(&a); err != nil {
			return nil, err
		}
		result[a.AccountCode] = a
	}
	return result, cursor.Err()
}

func (s *Store) masterMutation(ctx context.Context, scope Scope, cmd Command, now time.Time) ([]Change, error) {
	if MasterCollections[cmd.Resource] == "" {
		return nil, ErrNotFound
	}
	var old Master
	old.Kind = cmd.Resource
	if cmd.Action != "create" {
		if err := s.load(ctx, scope, cmd.Resource, cmd.ID, &old); err != nil {
			return nil, err
		}
		if err := checkVersion(cmd, old.Identity); err != nil {
			return nil, err
		}
		old.Kind = cmd.Resource
		if old.FiscalYear != "" && (cmd.Action == "update" || cmd.Action == "delete") {
			year, err := s.yearByCode(ctx, scope, old.FiscalYear)
			if err != nil {
				return nil, err
			}
			if year.Closed {
				return nil, fmt.Errorf("ปีบัญชีปิดแล้ว แก้ไข ย้ายปี หรือลบรายการอ้างอิงไม่ได้")
			}
		}
		if scope.Branch != "" && (cmd.Resource == "budgets" || cmd.Resource == "forecast") && old.BranchCode != scope.Branch {
			return nil, ErrNotFound
		}
	}
	if cmd.Action == "lock" || cmd.Action == "unlock" {
		if cmd.Resource != "periods" || strings.TrimSpace(cmd.Reason) == "" {
			return nil, fmt.Errorf("กรุณาระบุเหตุผลการเปลี่ยนสถานะงวด")
		}
		year, err := s.yearByCode(ctx, scope, old.FiscalYear)
		if err != nil {
			return nil, err
		}
		if year.Closed {
			return nil, fmt.Errorf("ปีบัญชีปิดแล้ว เปลี่ยนงวดไม่ได้")
		}
		if cmd.Action == "lock" {
			pending, err := s.referenced(ctx, scope, bson.M{"fiscalyear": old.FiscalYear, "date": bson.M{"$gte": old.StartDate, "$lte": old.EndDate}, "status": "draft", "isdeleted": false})
			if err != nil {
				return nil, err
			}
			if pending {
				return nil, fmt.Errorf("งวดนี้มีรายการร่าง กรุณาผ่านรายการหรือยกเลิกก่อนล็อกงวด")
			}
		}
		old.Locked = cmd.Action == "lock"
		old.Identity = identityFor(scope, cmd, old.Identity, now)
		return single(s.save(ctx, scope, cmd.Resource, old.ID, old.Code, old, cmd.Version))
	}
	if !contains([]string{"create", "update", "delete"}, cmd.Action) {
		return nil, fmt.Errorf("คำสั่งข้อมูลหลักไม่ถูกต้อง")
	}
	if cmd.Action == "delete" {
		if err := s.masterDeleteReferences(ctx, scope, old); err != nil {
			return nil, err
		}
		if old.Locked {
			return nil, fmt.Errorf("งวดที่ล็อกอยู่ลบไม่ได้")
		}
		old.Identity = identityFor(scope, cmd, old.Identity, now)
		old.IsDeleted = true
		return single(s.save(ctx, scope, cmd.Resource, old.ID, old.Code, old, cmd.Version))
	}
	if cmd.Master == nil {
		return nil, fmt.Errorf("ไม่พบข้อมูลที่ต้องการบันทึก")
	}
	next := *cmd.Master
	next.Kind = cmd.Resource
	next.Locked = old.Locked
	if scope.Branch != "" && (cmd.Resource == "budgets" || cmd.Resource == "forecast") && next.BranchCode != scope.Branch {
		return nil, fmt.Errorf("รายการอยู่นอกสาขาที่เลือก")
	}
	if !validCode(next.Code) || strings.TrimSpace(next.Name) == "" || len(next.Name) > 300 {
		return nil, fmt.Errorf("กรุณาระบุรหัสและชื่อรายการให้ครบ")
	}
	if cmd.Action == "update" && next.Code != old.Code {
		return nil, fmt.Errorf("รหัสรายการแก้ไม่ได้")
	}
	if next.Kind == "periods" || next.Kind == "budgets" || next.Kind == "forecast" {
		y, err := s.yearByCode(ctx, scope, next.FiscalYear)
		if err != nil {
			return nil, err
		}
		if y.Closed || !validDate(next.StartDate) || !validDate(next.EndDate) || next.StartDate > next.EndDate || next.StartDate < y.StartDate || next.EndDate > y.EndDate {
			return nil, fmt.Errorf("ช่วงวันต้องอยู่ในปีบัญชีที่เปิดใช้งาน")
		}
		if err := next.Amount.ValidateScale(y.Scale); err != nil {
			return nil, err
		}
		if next.Amount.Decimal().IsNegative() {
			return nil, fmt.Errorf("จำนวนเงินต้องไม่ติดลบ")
		}
		if next.Kind == "periods" && cmd.Action == "update" && (next.StartDate != old.StartDate || next.EndDate != old.EndDate || next.FiscalYear != old.FiscalYear || !next.IsActive) {
			if err := s.masterDeleteReferences(ctx, scope, old); err != nil {
				return nil, err
			}
		}
		if next.Kind == "periods" {
			if old.Locked && (next.StartDate != old.StartDate || next.EndDate != old.EndDate || next.FiscalYear != old.FiscalYear) {
				return nil, fmt.Errorf("งวดที่ล็อกอยู่เปลี่ยนช่วงวันไม่ได้")
			}
			f := scopeFilter(scope)
			f["isdeleted"] = false
			f["startdate"] = bson.M{"$lte": next.EndDate}
			f["enddate"] = bson.M{"$gte": next.StartDate}
			if cmd.ID != "" {
				f["_id"] = bson.M{"$ne": cmd.ID}
			}
			n, err := s.db.Collection("gl_periods").CountDocuments(ctx, f)
			if err != nil {
				return nil, err
			}
			if n > 0 {
				return nil, fmt.Errorf("ช่วงงวดบัญชีทับซ้อนกัน")
			}
		}
	}
	codes := []string{}
	if next.AccountCode != "" {
		codes = append(codes, next.AccountCode)
	}
	if next.Kind == "budgets" && next.AccountCode == "" {
		return nil, fmt.Errorf("กรุณาเลือกบัญชีงบประมาณ")
	}
	if next.Kind == "forecast" && !contains([]string{"in", "out"}, next.Direction) {
		return nil, fmt.Errorf("กรุณาเลือกรายรับหรือรายจ่ายตามแผน")
	}
	if next.Kind == "product-account-groups" {
		if next.ItemAccount == "" || next.CostAccount == "" || next.RevenueAccount == "" {
			return nil, fmt.Errorf("กรุณากำหนดบัญชีสินค้า ต้นทุน และรายได้")
		}
		codes = append(codes, next.ItemAccount, next.CostAccount, next.RevenueAccount)
	}
	if next.Kind == "mappings" {
		if len(next.Rules) < 2 || len(next.Rules) > 100 || !contains([]string{"JV", "UV", "SV", "RV", "PV"}, next.BookCode) {
			return nil, fmt.Errorf("กรุณากำหนดสมุดรายวันและรูปแบบเชื่อมอย่างน้อย 2 บรรทัด")
		}
		for _, rule := range next.Rules {
			if !contains([]string{"debit", "credit"}, rule.Side) || !contains([]string{"net", "tax", "total"}, rule.Source) {
				return nil, fmt.Errorf("รูปแบบเชื่อมต้องเลือกด้านบัญชีและยอดก่อนภาษี ภาษี หรือยอดรวม")
			}
			codes = append(codes, rule.AccountCode)
		}
	}
	if next.Kind == "statement-templates" {
		if next.StatementType == "" {
			next.StatementType = "custom"
		}
	}
	accounts, err := s.accounts(ctx, scope, codes)
	if err != nil {
		return nil, err
	}
	for _, code := range codes {
		a, ok := accounts[code]
		if !ok || !a.IsActive || !a.AllowPosting {
			return nil, fmt.Errorf("ไม่พบบัญชีที่เปิดรับรายการในบริษัทนี้")
		}
	}
	next.Identity = identityFor(scope, cmd, old.Identity, now)
	return single(s.save(ctx, scope, cmd.Resource, next.ID, next.Code, next, cmd.Version))
}

func (s *Store) checkJournal(ctx context.Context, scope Scope, j *Journal) error {
	if scope.Branch != "" && j.BranchCode != scope.Branch {
		return fmt.Errorf("รายการบัญชีอยู่นอกสาขาที่เลือก")
	}
	year, err := s.yearByCode(ctx, scope, j.FiscalYear)
	if err != nil {
		return err
	}
	codes := make([]string, len(j.Lines))
	for i, line := range j.Lines {
		codes[i] = line.AccountCode
	}
	accounts, err := s.accounts(ctx, scope, codes)
	if err != nil {
		return err
	}
	if err := j.Validate(year, accounts); err != nil {
		return err
	}
	if j.Kind == "opening" && j.Date != year.StartDate {
		return fmt.Errorf("ยอดยกมาต้องลงวันที่เริ่มปีบัญชี")
	}
	f := scopeFilter(scope)
	f["fiscalyear"] = year.Code
	f["isdeleted"] = false
	f["isactive"] = true
	f["startdate"] = bson.M{"$lte": j.Date}
	f["enddate"] = bson.M{"$gte": j.Date}
	var period Master
	if err := s.db.Collection("gl_periods").FindOne(ctx, f).Decode(&period); err != nil {
		return fmt.Errorf("วันที่นี้ยังไม่มีงวดบัญชี กรุณากำหนดงวดก่อนบันทึก")
	}
	if period.Locked {
		return fmt.Errorf("งวดบัญชีนี้ถูกล็อก กรุณาเลือกวันที่ในงวดเปิด")
	}
	for i := range j.Lines {
		j.Lines[i].AccountName = accounts[j.Lines[i].AccountCode].ThaiName()
	}
	return nil
}

func (s *Store) journalMutation(ctx context.Context, scope Scope, cmd Command, now time.Time) ([]Change, error) {
	if !contains([]string{"create", "update", "delete", "post", "reverse"}, cmd.Action) {
		return nil, fmt.Errorf("คำสั่งสมุดรายวันไม่ถูกต้อง")
	}
	var old Journal
	if cmd.Action != "create" {
		if err := s.load(ctx, scope, "journals", cmd.ID, &old); err != nil {
			return nil, err
		}
		if err := checkVersion(cmd, old.Identity); err != nil {
			return nil, err
		}
		if scope.Branch != "" && scope.Branch != old.BranchCode {
			return nil, ErrNotFound
		}
	}
	if cmd.Action == "reverse" {
		if old.Status != "posted" || (old.Kind == "reversal" || old.Kind == "closing") || strings.TrimSpace(cmd.Reason) == "" || !validCode(cmd.DocNo) || !validDate(cmd.Date) || cmd.Date < old.Date {
			return nil, fmt.Errorf("กลับรายการได้เฉพาะรายการที่ผ่านแล้ว โดยระบุเลขที่ วันที่ไม่ก่อนต้นฉบับ และเหตุผล")
		}
		reversal := old
		reversal.Lines = append([]Line(nil), old.Lines...)
		newCmd := cmd
		newCmd.Action = "create"
		newCmd.RequestID += "-reversal"
		reversal.Identity = identityFor(scope, newCmd, Identity{}, now)
		reversal.DocNo = cmd.DocNo
		reversal.Date = cmd.Date
		reversal.Kind = "reversal"
		reversal.Status = "posted"
		reversal.ReversalOf = old.ID
		reversal.Reference = old.DocNo
		reversal.Reason = cmd.Reason
		reversal.Description = "กลับรายการ " + old.DocNo + ": " + cmd.Reason
		// Reversal may be posted into a later open fiscal year of the same currency.
		f := scopeFilter(scope)
		f["startdate"] = bson.M{"$lte": cmd.Date}
		f["enddate"] = bson.M{"$gte": cmd.Date}
		f["isdeleted"] = false
		f["isactive"] = true
		var y FiscalYear
		if err := s.db.Collection("fiscal_year").FindOne(ctx, f).Decode(&y); err != nil {
			return nil, fmt.Errorf("ไม่พบปีบัญชีของวันที่กลับรายการ")
		}
		reversal.FiscalYear = y.Code
		for i := range reversal.Lines {
			reversal.Lines[i].Debit, reversal.Lines[i].Credit = reversal.Lines[i].Credit, reversal.Lines[i].Debit
		}
		if err := s.checkJournal(ctx, scope, &reversal); err != nil {
			return nil, err
		}
		reversal.PostedAt = &now
		reversal.PostedBy = scope.Actor
		old.Identity = identityFor(scope, cmd, old.Identity, now)
		old.Status = "reversed"
		originalChange, err := s.save(ctx, scope, "journals", old.ID, old.DocNo, old, cmd.Version)
		if err != nil {
			return nil, err
		}
		reverseChange, err := s.save(ctx, scope, "journals", reversal.ID, reversal.DocNo, reversal, 0)
		if err != nil {
			return nil, err
		}
		return []Change{originalChange, reverseChange}, nil
	}
	if (cmd.Action == "update" || cmd.Action == "delete") && strings.HasPrefix(old.Reference, "YEAR-END:") {
		return nil, fmt.Errorf("ยอดยกมาจากการปิดปีแก้ไขหรือลบไม่ได้ กรุณาผ่านรายการแล้วบันทึกปรับปรุงแยก")
	}
	if cmd.Action != "create" && old.Status != "draft" {
		return nil, fmt.Errorf("รายการที่ผ่านบัญชีแล้วแก้ไขหรือลบไม่ได้ กรุณาใช้กลับรายการ")
	}
	if cmd.Action == "delete" {
		old.Identity = identityFor(scope, cmd, old.Identity, now)
		old.IsDeleted = true
		old.Status = "void"
		return single(s.save(ctx, scope, "journals", old.ID, old.DocNo, old, cmd.Version))
	}
	next := old
	if cmd.Action == "create" || cmd.Action == "update" {
		if cmd.Journal == nil {
			return nil, fmt.Errorf("ไม่พบข้อมูลสมุดรายวัน")
		}
		next = *cmd.Journal
		if cmd.Action == "update" && (next.BookCode != old.BookCode || next.Kind != old.Kind) {
			return nil, fmt.Errorf("เปลี่ยนสมุดรายวันหรือประเภทรายการเดิมไม่ได้")
		}
		if !contains([]string{"manual", "opening"}, next.Kind) {
			return nil, fmt.Errorf("ประเภทรายการนี้ต้องสร้างจากขั้นตอนประมวลผล")
		}
		next.Status = "draft"
		next.ReversalOf = ""
		next.PostedAt = nil
		next.PostedBy = ""
		next.Reason = ""
	}
	if err := s.checkJournal(ctx, scope, &next); err != nil {
		return nil, err
	}
	if cmd.Action == "post" {
		next.Status = "posted"
		next.PostedAt = &now
		next.PostedBy = scope.Actor
	}
	next.Identity = identityFor(scope, cmd, old.Identity, now)
	return single(s.save(ctx, scope, "journals", next.ID, next.DocNo, next, cmd.Version))
}
