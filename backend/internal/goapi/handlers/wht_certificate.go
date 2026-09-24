package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	authmodels "smlcloudplatform/internal/authentication/models"
	"smlcloudplatform/internal/generalledger"
	"smlcloudplatform/internal/goapi/language"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/mypg"
	"smlcloudplatform/internal/mcptoken"
	orgpolicy "smlcloudplatform/internal/organization/access"
	"smlcloudplatform/internal/whtcert"
	msmodels "smlcloudplatform/pkg/microservice/models"

	"github.com/labstack/echo/v4"
)

// whtCompanyHeader - จุดสลับให้เทสระดับ handler ไม่ต้องต่อฐานข้อมูลควบคุมกลาง
var whtCompanyHeader = loadCompanyHeader

// whtRecord - รายการภาษีหักที่บันทึก + ทะเบียนคู่ค้าปัจจุบัน + สาขาของใบสำคัญ (ใช้ตรวจสิทธิ์สาขา)
type whtRecord struct {
	item    generalledger.SubledgerWithholding
	partner generalledger.SubledgerPartner
	branch  string
}

// whtRecordedItem - รายการภาษีหักที่บันทึกในใบสำคัญ (+ ทะเบียนคู่ค้าปัจจุบัน + สาขาของใบ) จากฐานข้อมูลของกลุ่มกิจการ; เทสสลับได้
// generalledger.RecordedWithholding กรองบริษัทแล้วแต่ไม่รู้สาขา — อ่าน branchcode ของใบเดียวกันเพื่อให้ handler ตรวจสิทธิ์สาขา
var whtRecordedItem = func(ctx context.Context, holdingCode, company, journalID, itemID string) (whtRecord, error) {
	db, err := mypg.PgSqlFastConnect(holdingCode) // pool กลางของกลุ่มกิจการ ห้าม Close
	if err != nil {
		return whtRecord{}, fmt.Errorf("connect holding database: %w", err)
	}
	if err := generalledger.EnsureSchema(ctx, db); err != nil {
		return whtRecord{}, fmt.Errorf("ensure GL schema: %w", err)
	}
	item, partner, err := generalledger.RecordedWithholding(ctx, db, company, journalID, itemID)
	if err != nil {
		return whtRecord{}, err
	}
	var branch string
	if err := db.QueryRowContext(ctx, `SELECT COALESCE(payload->>'branchcode', '') FROM gl_records WHERE company = $1 AND kind = 'journals' AND id = $2`,
		company, journalID).Scan(&branch); err != nil {
		return whtRecord{}, fmt.Errorf("read journal branch: %w", err)
	}
	return whtRecord{item: item, partner: partner, branch: branch}, nil
}

// errWhtScopeDenied - บริษัทหรือสาขาที่เลือกไม่อยู่ในสิทธิ์ของผู้ใช้ (access_scopes ของสมาชิกภาพ)
var errWhtScopeDenied = errors.New("wht certificate outside the caller's company/branch scope")

// whtAccess - สิทธิ์ของผู้เรียกในบริษัทที่เลือก ตาม access_scopes ของสมาชิกภาพในฐานควบคุมกลาง
type whtAccess struct {
	company string
	scopes  []authmodels.AccessScope
}

// allowsBranch - ใบสำคัญของสาขานี้อยู่ในสิทธิ์ (สิทธิ์ทั้งบริษัท หรือสาขาที่ได้รับ); ใบที่ไม่มีสาขาต้องมีสิทธิ์ทั้งบริษัท
func (a whtAccess) allowsBranch(branch string) bool {
	if strings.TrimSpace(branch) == "" {
		return orgpolicy.AllowsAllBranches(a.scopes, a.company)
	}
	return orgpolicy.AllowsBranch(a.scopes, a.company, branch)
}

// whtCallerAccess - ตรวจสมาชิกภาพจากฐานควบคุมกลางทุกครั้ง (ไม่เชื่อ token อย่างเดียว): ยังใช้งานได้และไม่หมดอายุ
// (orgpolicy.ErrAccessExpired), บริษัทที่เลือกอยู่ในสิทธิ์ และสาขาที่เลือก (ถ้ามี) อยู่ในสิทธิ์ — เทสสลับได้
var whtCallerAccess = func(ctx context.Context, user msmodels.UserInfo, company string) (whtAccess, error) {
	db, err := mypg.PgSqlFastConnect(mcptoken.ControlDatabase)
	if err != nil {
		return whtAccess{}, fmt.Errorf("connect control database: %w", err)
	}
	return whtAccessFor(ctx, db, user, company, time.Now())
}

// whtAccessFor - เนื้อของ whtCallerAccess บนฐานควบคุมกลางที่ส่งมา (เทส integration เรียกตรง)
func whtAccessFor(ctx context.Context, db orgpolicy.Querier, user msmodels.UserInfo, company string, now time.Time) (whtAccess, error) {
	membership, err := orgpolicy.FindActiveMembership(ctx, db, user, now)
	if err != nil {
		return whtAccess{}, err
	}
	access := whtAccess{company: company, scopes: membership.AccessScopes}
	if !orgpolicy.AllowsCompany(access.scopes, company) {
		return whtAccess{}, errWhtScopeDenied
	}
	if branch := strings.TrimSpace(user.BranchUID); branch != "" && !orgpolicy.AllowsBranch(access.scopes, company, branch) {
		return whtAccess{}, errWhtScopeDenied
	}
	return access, nil
}

// WhtCertificateRequest - ข้อมูลใบ 50 ทวิ ที่หน้าจอเตรียมจากรายงานภาษีหัก ณ ที่จ่าย (นักบัญชีตรวจ/แก้ได้ก่อนสั่งพิมพ์)
type WhtCertificateRequest struct {
	HoldingCode  string `json:"holdingcode"`
	BusinessCode string `json:"businesscode"`
	// ใบสำคัญ + id รายการใน details.withholdings (ไม่บังคับ แต่ต้องส่งคู่กัน): มีค่า = ผู้รับ/ผู้จ่ายฝั่งคู่ค้า เล่มที่ และเลขที่
	// ใช้ค่าที่บันทึกในรายการก่อน → ทะเบียนคู่ค้า → ค่าที่หน้าจอส่งมา; ฝั่งบริษัทยึดทะเบียนบริษัทก่อน (applyRecordedWithholding);
	// ยอดเงิน ภาษี วันที่จ่าย ประเภทเงินได้ แบบยื่น เงื่อนไข พิมพ์ตามรายการเสมอ (applyRecordedFigures)
	JournalID     string              `json:"journalid,omitempty"`
	WithholdingID string              `json:"withholdingid,omitempty"`
	Certificate   whtcert.Certificate `json:"certificate"`
}

// WhtCertificateHandler - POST /api/report/tax/wht/certificate → application/pdf
// backend สร้าง PDF บนแบบฟอร์มของกรมสรรพากรเอง (ตรวจเลข/ยอด คำนวณยอดรวมและตัวอักษร) หน้าจอแค่แสดงผล
// ไม่อ้างรายการที่บันทึก: ชื่อและเลขผู้เสียภาษีของผู้หักภาษียึดทะเบียนบริษัทเมื่อทะเบียนมีค่า (กันพิมพ์ในนามบริษัทอื่น)
// อ้างรายการที่บันทึก (journalid + withholdingid): ฝั่งคู่ค้าใช้ snapshot ที่บันทึกพร้อม audit ก่อน, ฝั่งบริษัทยังยึดทะเบียนบริษัท,
// ตัวเลข แบบยื่น และเงื่อนไขพิมพ์จากรายการ (ตัวเลขในคำขอใช้เฉพาะหนังสือรับรองที่ไม่อ้างรายการ) — applyRecordedWithholding + applyRecordedFigures
// สิทธิ์: สมาชิกภาพที่ยังใช้งานได้ + บริษัท/สาขาที่เลือกอยู่ในสิทธิ์ (whtCallerAccess) และใบสำคัญที่อ้างต้องอยู่ในสาขาที่มีสิทธิ์
func WhtCertificateHandler(c echo.Context) error {
	// ภาษาเดียวกับรายงานภาษีอื่น (?lang ก่อน แล้ว Accept-Language) — เดิมอ่านแค่ header จึงไม่ตรงกับจอที่ส่ง ?lang
	lang := taxRequestLanguage(c)
	fail := func(status int, key, field string) error {
		return c.JSON(status, map[string]any{"success": false, "code": key, "field": field, "message": language.Text(key, lang)})
	}

	var req WhtCertificateRequest
	if err := c.Bind(&req); err != nil {
		return fail(http.StatusBadRequest, "wht_cert_payload_invalid", "")
	}
	holdingCode, businessCode, scopeErr := authenticatedCompanyContext(c, req.HoldingCode, req.BusinessCode)
	if scopeErr != nil {
		// ข้อความของ authenticatedCompanyContext เป็นภาษาเดียว — แปลผ่าน languages.tsv แบบเดียวกับรายงานภาษี
		return taxScopeFail(c, scopeErr)
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 15*time.Second)
	defer cancel()
	user, _ := c.Get("UserInfo").(msmodels.UserInfo) // authenticatedCompanyContext ตรวจแล้วว่ามี
	access, err := whtCallerAccess(ctx, user, businessCode)
	switch {
	case errors.Is(err, orgpolicy.ErrAccessExpired):
		return fail(http.StatusForbidden, "user_access_expired", "")
	case errors.Is(err, errWhtScopeDenied), errors.Is(err, orgpolicy.ErrActiveMembershipRequired):
		return fail(http.StatusForbidden, "wht_cert_scope_denied", "")
	case err != nil:
		logger.Error("WhtCertificate: caller access: %v", err)
		return fail(http.StatusInternalServerError, "wht_cert_render_failed", "")
	}
	company, err := whtCompanyHeader(ctx, holdingCode, businessCode)
	if err != nil {
		logger.Error("WhtCertificate: company header: %v", err)
		return fail(http.StatusInternalServerError, "wht_cert_render_failed", "")
	}
	cert := req.Certificate
	journalID, itemID := strings.TrimSpace(req.JournalID), strings.TrimSpace(req.WithholdingID)
	if journalID == "" && itemID == "" {
		// ไม่อ้างรายการที่บันทึก (แบบเดิม): ผู้หักภาษีทิศทางเราหักยึดทะเบียนบริษัท
		cert.Payer = partyOr(whtcert.Party{Name: company.Name, TaxID: company.TaxID}, cert.Payer)
	} else {
		if journalID == "" {
			return fail(http.StatusBadRequest, "wht_cert_record_not_found", "journalid")
		}
		if itemID == "" {
			return fail(http.StatusBadRequest, "wht_cert_record_not_found", "withholdingid")
		}
		record, err := whtRecordedItem(ctx, holdingCode, businessCode, journalID, itemID)
		if errors.Is(err, generalledger.ErrNotFound) {
			return fail(http.StatusNotFound, "wht_cert_record_not_found", "withholdingid")
		}
		if err != nil {
			logger.Error("WhtCertificate: recorded withholding: %v", err)
			return fail(http.StatusInternalServerError, "wht_cert_render_failed", "")
		}
		if !access.allowsBranch(record.branch) {
			return fail(http.StatusForbidden, "wht_cert_scope_denied", "journalid")
		}
		cert = applyRecordedWithholding(cert, record.item, record.partner, company)
		cert = applyRecordedFigures(cert, record.item)
	}

	pdf, err := whtcert.Render(cert)
	var invalid *whtcert.ValidationError
	if errors.As(err, &invalid) {
		return fail(http.StatusBadRequest, invalid.Key, invalid.Field)
	}
	if err != nil {
		logger.Error("WhtCertificate: render: %v", err)
		return fail(http.StatusInternalServerError, "wht_cert_render_failed", "")
	}

	c.Response().Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="50tawi-%s.pdf"`, safeFileToken(cert.RunNo)))
	c.Response().Header().Set("Cache-Control", "no-store")
	return c.Blob(http.StatusOK, "application/pdf", pdf)
}

// applyRecordedWithholding - ช่องของผู้จ่าย/ผู้รับเงินทีละช่อง
// ฝั่งคู่ค้า: snapshot ที่บันทึกในรายการ → ทะเบียนคู่ค้าปัจจุบัน → ค่าที่หน้าจอส่งมา
// ฝั่งบริษัท: ชื่อ/เลขผู้เสียภาษีจากทะเบียนบริษัทเมื่อทะเบียนมีค่า (กันพิมพ์ในนามบริษัทอื่น — snapshot ฝั่งบริษัทแก้ได้ทาง API/reconcile
// และค้างค่าที่พิมพ์ไว้ตอนทะเบียนยังว่าง) → snapshot → ค่าที่หน้าจอส่งมา; ที่อยู่ไม่มีในทะเบียนจึงใช้ snapshot ก่อน
// ทิศทาง 1 (เราหักภาษีผู้รับเงิน) ผู้จ่าย = บริษัท ผู้รับ = คู่ค้า; ทิศทาง 2 สลับกัน (wht.sql: payer=ผู้จ่าย/ผู้หัก เสมอ)
func applyRecordedWithholding(cert whtcert.Certificate, item generalledger.SubledgerWithholding, partner generalledger.SubledgerPartner, company CompanyHeader) whtcert.Certificate {
	companyParty := whtcert.Party{Name: company.Name, TaxID: company.TaxID}
	// ชื่อ/ที่อยู่เต็มแบบเดียวกับ snapshot (fillPartnerSnapshot) — แบบ 50 ทวิ ให้ระบุอำเภอ/เขต จังหวัด ในที่อยู่
	partnerParty := whtcert.Party{Name: generalledger.PartnerFullName(partner.TitleName, partner.Name), TaxID: partner.TaxID,
		Address: generalledger.PartnerFullAddress(partner.Address, partner.AddrDistrict, partner.AddrProvince, partner.AddrPostcode)}
	payer, payee := item.Payer(), item.Payee()
	payerSnap := whtcert.Party{Name: payer.Name, TaxID: payer.TaxID, Address: payer.Address}
	payeeSnap := whtcert.Party{Name: payee.Name, TaxID: payee.TaxID, Address: payee.Address}
	if item.PartnerIsPayer() {
		cert.Payer = partyOr(payerSnap, partyOr(partnerParty, cert.Payer))
		cert.Payee = partyOr(companyParty, partyOr(payeeSnap, cert.Payee))
	} else {
		cert.Payer = partyOr(companyParty, partyOr(payerSnap, cert.Payer))
		cert.Payee = partyOr(payeeSnap, partyOr(partnerParty, cert.Payee))
	}
	cert.BookNo = firstNonBlank(item.BookNo, cert.BookNo)
	cert.RunNo = firstNonBlank(item.CertificateNo, cert.RunNo)
	return cert
}

// recordedForms / recordedConditions - ค่าที่บันทึกในรายการภาษีหัก → ค่าบนแบบ 50 ทวิ
var (
	recordedForms      = map[string]string{"PND2": whtcert.FormPND2, "PND3": whtcert.FormPND3, "PND53": whtcert.FormPND53}
	recordedConditions = map[int]string{1: whtcert.ConditionWithhold, 2: whtcert.ConditionAlways, 3: whtcert.ConditionOnce}
)

// applyRecordedFigures - อ้างรายการที่บันทึก: พิมพ์ตามรายการเสมอ (ยอดเดียวกับ ภ.ง.ด./ไฟล์ยื่น) — ประเภทเงินได้ วันที่จ่าย
// จำนวนเงิน ภาษี แบบยื่น เงื่อนไข และวันที่ออกหนังสือรับรอง (ถ้าบันทึกไว้) มาจากรายการ; ตัวเลขในคำขอใช้เฉพาะหนังสือรับรองที่ไม่อ้างรายการ
// (review 2026-09-24: เดิมปฏิเสธเมื่อค่าไม่ตรง แต่ยังรับตัวเลขจากคำขอเมื่อช่องว่าง/รายการซ้ำ — ตอนนี้ไม่มีทางพิมพ์ยอดอื่นจากรายการ)
// ช่อง "ระบุ" ของบรรทัดที่มี (อัตราเงินปันผล/อื่น ๆ) ใช้ข้อความจากหน้าจอของประเภทเดียวกัน ถ้าว่างใช้คำอธิบายเงินได้ที่บันทึก;
// เงื่อนไขในรายการมีแค่ 1–3 จึงไม่มีรายละเอียด "อื่น ๆ"; เงินกองทุน (กบข./ประกันสังคม/สำรองเลี้ยงชีพ) เป็นเรื่องเงินเดือน รายการไม่มี → ว่าง
func applyRecordedFigures(cert whtcert.Certificate, item generalledger.SubledgerWithholding) whtcert.Certificate {
	tax := ""
	if item.TaxAmount != nil {
		tax = string(*item.TaxAmount)
	}
	in := whtcert.Income{Type: item.IncomeType, PaidDate: item.PaymentDate, Amount: string(item.BaseAmount), Tax: tax}
	if whtcert.IncomeHasNote(item.IncomeType) {
		screenNote := ""
		for _, requested := range cert.Incomes {
			if requested.Type == item.IncomeType {
				screenNote = requested.Note
				break
			}
		}
		in.Note = firstNonBlank(screenNote, item.Description)
	}
	cert.Incomes = []whtcert.Income{in}
	cert.Form = recordedForms[item.FormType]
	cert.Condition = recordedConditions[item.Condition]
	cert.ConditionNote = ""
	cert.FundGPF, cert.FundSSF, cert.FundPVD = "", "", ""
	cert.IssueDate = firstNonBlank(item.CertificateDate, cert.IssueDate)
	return cert
}

// partyOr - ใช้ช่องของ preferred ที่ไม่ว่าง ช่องที่ว่างใช้ของ fallback (TaxID10 มาจาก fallback เสมอ — ไม่มีใน snapshot)
func partyOr(preferred, fallback whtcert.Party) whtcert.Party {
	return whtcert.Party{
		Name:    firstNonBlank(preferred.Name, fallback.Name),
		Address: firstNonBlank(preferred.Address, fallback.Address),
		TaxID:   firstNonBlank(preferred.TaxID, fallback.TaxID),
		TaxID10: firstNonBlank(preferred.TaxID10, fallback.TaxID10),
	}
}

func firstNonBlank(values ...string) string {
	for _, v := range values {
		if v = strings.TrimSpace(v); v != "" {
			return v
		}
	}
	return ""
}

// safeFileToken - เลขที่เอกสารสำหรับชื่อไฟล์ (ตัด / และอักขระพิเศษ กัน header injection)
func safeFileToken(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= '0' && r <= '9', r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r == '-':
			b.WriteRune(r)
		case r == '/' || r == '_' || r == ' ':
			b.WriteByte('-')
		}
	}
	if b.Len() == 0 {
		return "certificate"
	}
	return b.String()
}
