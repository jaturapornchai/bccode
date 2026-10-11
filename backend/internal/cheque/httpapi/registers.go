package httpapi

import (
	"net/http"
	"strconv"
	"strings"

	"smlcloudplatform/internal/cheque"
	"smlcloudplatform/internal/doctype"
	glhttp "smlcloudplatform/internal/generalledger/httpapi"
	"smlcloudplatform/pkg/microservice"
)

// ทะเบียนเช็ค + เอกสารเปลี่ยนสถานะ 8 ตัว + คำนวณยอดเช็คใหม่: สิทธิ์ตามเมนูของจอนั้น (menu-data.ts)

// registerScreens คือเมนูทะเบียนของเช็คแต่ละชนิด
var registerScreens = map[int]string{cheque.TypeReceived: "cheque-received", cheque.TypeIssued: "cheque-issued"}

const recalculateScreen = "cheque-recalculate"

// docScreen คือเมนูของเอกสารสถานะ (screen ของ doctype) — ok = false เมื่อไม่ใช่ transflag ของเอกสารสถานะ
func docScreen(flag int) (string, bool) {
	if _, ok := cheque.DocChequeType(flag); !ok {
		return "", false
	}
	dt, ok := doctype.ByFlag(flag)
	return dt.Screen, ok
}

// canSeeType: เปิดดูเช็คชนิดหนึ่งได้จากจอทะเบียนของชนิดนั้น หรือจอเอกสารสถานะที่เลือกเช็คชนิดนั้น
func canSeeType(s glhttp.SessionAccess, chequeType int) bool {
	if s.Allows(registerScreens[chequeType], actionRead) {
		return true
	}
	for _, flag := range cheque.DocFlags() {
		if t, _ := cheque.DocChequeType(flag); t != chequeType {
			continue
		}
		if screen, ok := docScreen(flag); ok && s.Allows(screen, actionRead) {
			return true
		}
	}
	return false
}

func canSeeAnyCheque(s glhttp.SessionAccess) bool {
	return canSeeType(s, cheque.TypeReceived) || canSeeType(s, cheque.TypeIssued)
}

// allowsWrite: ยกเลิกได้ด้วยสิทธิ์ cancel หรือ delete ของเมนู (กติกาเดียวกับเอกสาร ERP)
func allowsWrite(s glhttp.SessionAccess, screen, action string) bool {
	if action == actionCancel {
		return s.Allows(screen, actionCancel) || s.Allows(screen, actionDelete)
	}
	return s.Allows(screen, action)
}

// chequeTypeValue รับชนิดเช็ค "0" (รับ) หรือ "1" (จ่าย)
func chequeTypeValue(value string) (int, bool) {
	switch strings.TrimSpace(value) {
	case "0":
		return cheque.TypeReceived, true
	case "1":
		return cheque.TypeIssued, true
	}
	return 0, false
}

// optionalInt: ว่าง = ไม่กรอง
func optionalInt(value string) (*int, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, true
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return nil, false
	}
	return &n, true
}

func boolParam(request microservice.IContext, name string) bool {
	v := strings.TrimSpace(request.QueryParam(name))
	return v == "1" || strings.EqualFold(v, "true")
}

func pageBody[T any](items []T, total, page, limit int) map[string]any {
	return map[string]any{"items": items, "total": total, "page": page, "limit": limit}
}

// GET /cheque/v1/registers?type=0|1&status=&partner=&book=&due_from=&due_to=&q=&page=&limit=
func (h *Http) listRegisters(request microservice.IContext) error {
	chequeType, ok := chequeTypeValue(request.QueryParam("type"))
	if !ok {
		return fail(request, http.StatusBadRequest, "cheque_err_type_invalid", "type")
	}
	status, ok := optionalInt(request.QueryParam("status"))
	if !ok {
		return fail(request, http.StatusBadRequest, "cheque_err_status_invalid", "status")
	}
	a := h.authorizeLedger(request, false, func(s glhttp.SessionAccess) bool { return s.Allows(registerScreens[chequeType], actionRead) })
	if a == nil {
		return nil
	}
	page, err := cheque.ListRegisters(a.ctx, a.db, a.scope.Company, cheque.RegisterFilter{
		ChequeType: chequeType, Status: status,
		Partner: request.QueryParam("partner"), Book: request.QueryParam("book"),
		DueFrom: request.QueryParam("due_from"), DueTo: request.QueryParam("due_to"), Query: request.QueryParam("q"),
		Page: intParam(request, "page"), Limit: intParam(request, "limit"),
	})
	if err != nil {
		return failure(request, err)
	}
	return respond(request, pageBody(page.Items, page.Total, page.Page, page.Limit))
}

// GET /cheque/v1/registers/selectable?transflag=&doc_date=&book=&cancel_type=&cancel_passed=&q=&page=&limit=
func (h *Http) listSelectable(request microservice.IContext) error {
	flag, _ := strconv.Atoi(request.QueryParam("transflag"))
	screen, ok := docScreen(flag)
	if !ok {
		return fail(request, http.StatusBadRequest, "cheque_err_transflag_invalid", "transflag", request.QueryParam("transflag"))
	}
	cancelType, ok := optionalInt(request.QueryParam("cancel_type"))
	if !ok {
		return fail(request, http.StatusBadRequest, "cheque_err_cancel_type_invalid", "cancel_type")
	}
	a := h.authorizeLedger(request, false, func(s glhttp.SessionAccess) bool { return s.Allows(screen, actionRead) })
	if a == nil {
		return nil
	}
	f := cheque.SelectableFilter{
		Flag: flag, DocDate: request.QueryParam("doc_date"), Book: request.QueryParam("book"),
		CancelType: -1, CancelPassed: boolParam(request, "cancel_passed"), Query: request.QueryParam("q"),
		Page: intParam(request, "page"), Limit: intParam(request, "limit"),
	}
	if cancelType != nil {
		f.CancelType = *cancelType
	}
	page, err := cheque.ListSelectable(a.ctx, a.db, a.scope.Company, f)
	if err != nil {
		return failure(request, err)
	}
	return respond(request, pageBody(page.Items, page.Total, page.Page, page.Limit))
}

// GET /cheque/v1/registers/:id — เช็คหนึ่งใบ + การปรับปรุงล่าสุด + เอกสารที่อ้างถึง
func (h *Http) getRegister(request microservice.IContext) error {
	a := h.authorizeLedger(request, false, canSeeAnyCheque)
	if a == nil {
		return nil
	}
	detail, err := cheque.GetRegister(a.ctx, a.db, a.scope.Company, request.Param("id"))
	if err != nil {
		return failure(request, err)
	}
	if !canSeeType(a.session, detail.ChequeType) {
		return fail(request, http.StatusForbidden, "cheque_err_read_forbidden", "")
	}
	return respond(request, map[string]any{"data": detail})
}

// docFlagParam อ่าน :transflag แล้วคืนเมนูของเอกสาร
func docFlagParam(request microservice.IContext) (int, string, bool) {
	flag, _ := strconv.Atoi(request.Param("transflag"))
	screen, ok := docScreen(flag)
	if !ok {
		_ = fail(request, http.StatusBadRequest, "cheque_err_transflag_invalid", "transflag", request.Param("transflag"))
	}
	return flag, screen, ok
}

// GET /cheque/v1/docs/:transflag?q=&date_from=&date_to=&status=&page=&limit=
func (h *Http) listDocs(request microservice.IContext) error {
	flag, screen, ok := docFlagParam(request)
	if !ok {
		return nil
	}
	a := h.authorizeLedger(request, false, func(s glhttp.SessionAccess) bool { return s.Allows(screen, actionRead) })
	if a == nil {
		return nil
	}
	page, err := cheque.ListDocs(a.ctx, a.db, a.scope.Company, flag, cheque.DocFilter{
		Query: request.QueryParam("q"), DateFrom: request.QueryParam("date_from"), DateTo: request.QueryParam("date_to"),
		Status: request.QueryParam("status"), Page: intParam(request, "page"), Limit: intParam(request, "limit"),
	})
	if err != nil {
		return failure(request, err)
	}
	return respond(request, pageBody(page.Items, page.Total, page.Page, page.Limit))
}

// GET /cheque/v1/docs/:transflag/:doc_no
func (h *Http) getDoc(request microservice.IContext) error {
	flag, screen, ok := docFlagParam(request)
	if !ok {
		return nil
	}
	a := h.authorizeLedger(request, false, func(s glhttp.SessionAccess) bool { return s.Allows(screen, actionRead) })
	if a == nil {
		return nil
	}
	doc, err := cheque.GetDoc(a.ctx, a.db, a.scope.Company, flag, request.Param("doc_no"))
	if err != nil {
		return failure(request, err)
	}
	return respond(request, map[string]any{"data": doc})
}

// GET /cheque/v1/docs/:transflag/:doc_no/events — ประวัติ event ของเอกสาร (รวมรอบแก้ไขที่ถูกยกเลิก)
func (h *Http) listDocEvents(request microservice.IContext) error {
	flag, screen, ok := docFlagParam(request)
	if !ok {
		return nil
	}
	a := h.authorizeLedger(request, false, func(s glhttp.SessionAccess) bool { return s.Allows(screen, actionRead) })
	if a == nil {
		return nil
	}
	items, err := cheque.ListDocEvents(a.ctx, a.db, a.scope.Company, flag, strings.TrimSpace(request.Param("doc_no")))
	if err != nil {
		return failure(request, err)
	}
	return respond(request, map[string]any{"items": items, "total": len(items)})
}

// GET /cheque/v1/recalculate — ตรวจสถานะในทะเบียนเทียบประวัติ (ไม่เขียน)
func (h *Http) checkRecalculate(request microservice.IContext) error {
	a := h.authorizeLedger(request, false, func(s glhttp.SessionAccess) bool { return s.Allows(recalculateScreen, actionRead) })
	if a == nil {
		return nil
	}
	result, err := cheque.CheckRecalculate(a.ctx, a.db, a.scope.Company)
	if err != nil {
		return failure(request, err)
	}
	return respond(request, map[string]any{"data": result})
}

// GET /cheque/v1/balances/partners?partner=&page=&limit= — ยอดเช็คค้างรายผู้ติดต่อ (คำนวณจากทะเบียน)
func (h *Http) partnerBalances(request microservice.IContext) error {
	a := h.authorizeLedger(request, false, func(s glhttp.SessionAccess) bool {
		return s.Allows(registerScreens[cheque.TypeReceived], actionRead) || s.Allows(registerScreens[cheque.TypeIssued], actionRead)
	})
	if a == nil {
		return nil
	}
	page, err := cheque.PartnerBalances(a.ctx, a.db, a.scope.Company, cheque.BalanceFilter{
		Partner: request.QueryParam("partner"), Page: intParam(request, "page"), Limit: intParam(request, "limit"),
	})
	if err != nil {
		return failure(request, err)
	}
	return respond(request, pageBody(page.Items, page.Total, page.Page, page.Limit))
}

// GET /cheque/v1/balances/books?book=&date_from=&date_to= — เงินเข้า/ออกสมุดเงินฝากจากเอกสารสถานะเช็ค
func (h *Http) bookMovements(request microservice.IContext) error {
	a := h.authorizeLedger(request, false, func(s glhttp.SessionAccess) bool {
		return s.Allows(bookBankScreen, actionRead) || canSeeAnyCheque(s)
	})
	if a == nil {
		return nil
	}
	items, err := cheque.BookMovements(a.ctx, a.db, a.scope.Company, request.QueryParam("book"), request.QueryParam("date_from"), request.QueryParam("date_to"))
	if err != nil {
		return failure(request, err)
	}
	return respond(request, map[string]any{"items": items, "total": len(items)})
}

// registers create/update/cancel: สิทธิ์ตามเมนูทะเบียนของชนิดเช็คในคำขอ
func (h *Http) registersCommand(request microservice.IContext, body commandBody) error {
	action := strings.TrimSpace(body.Action)
	if action != actionCreate && action != actionUpdate && action != actionCancel {
		return fail(request, http.StatusBadRequest, "cheque_err_action_invalid", "action", body.Action)
	}
	if body.ChequeType == nil || registerScreens[*body.ChequeType] == "" {
		return fail(request, http.StatusBadRequest, "cheque_err_type_invalid", "cheque_type")
	}
	chequeType := *body.ChequeType
	a := h.authorizeLedger(request, true, func(s glhttp.SessionAccess) bool { return allowsWrite(s, registerScreens[chequeType], action) })
	if a == nil {
		return nil
	}
	var (
		saved cheque.Register
		err   error
	)
	switch action {
	case actionCancel:
		saved, err = cheque.CancelRegister(a.ctx, a.db, a.scope, body.RequestID, chequeType, body.ID, body.Version)
	default:
		input := cheque.RegisterInput{
			ID: body.ID, Version: body.Version, ChequeType: chequeType, DocNo: body.DocNo, DocDate: body.DocDate,
			DueDate: body.DueDate, PartnerCode: body.PartnerCode, ChequeNo: body.ChequeNo, BankCode: body.BankCode,
			BankBranchCode: body.BankBranchCode, BookCode: body.BookCode, Status: body.Status, ExtStatusCode: body.ExtStatusCode,
			CurrencyCode: body.CurrencyCode, ExchangeRate: body.ExchangeRate, Amount: body.Amount, Description: body.Description,
			SaleCode: body.SaleCode, DepartmentCode: body.DepartmentCode, Confirm: body.Confirm,
		}
		if action == actionCreate {
			saved, err = cheque.CreateRegister(a.ctx, a.db, a.scope, body.RequestID, input)
		} else {
			saved, err = cheque.UpdateRegister(a.ctx, a.db, a.scope, body.RequestID, input)
		}
	}
	if err != nil {
		return failure(request, err)
	}
	return respond(request, map[string]any{"data": saved})
}

// docs create/update/cancel: สิทธิ์ตามเมนูของเอกสารสถานะ (transflag)
func (h *Http) docsCommand(request microservice.IContext, body commandBody) error {
	action := strings.TrimSpace(body.Action)
	if action != actionCreate && action != actionUpdate && action != actionCancel {
		return fail(request, http.StatusBadRequest, "cheque_err_action_invalid", "action", body.Action)
	}
	screen, ok := docScreen(body.Transflag)
	if !ok {
		return fail(request, http.StatusBadRequest, "cheque_err_transflag_invalid", "transflag", strconv.Itoa(body.Transflag))
	}
	a := h.authorizeLedger(request, true, func(s glhttp.SessionAccess) bool { return allowsWrite(s, screen, action) })
	if a == nil {
		return nil
	}
	var (
		saved cheque.Doc
		err   error
	)
	switch action {
	case actionCancel:
		saved, err = cheque.CancelDoc(a.ctx, a.db, h.ledger, a.scope, body.RequestID, body.Transflag, body.DocNo, body.Version, body.Reason)
	default:
		input := cheque.DocInput{
			Flag: body.Transflag, DocNo: body.DocNo, Version: body.Version, DocDate: body.DocDate, BookCode: body.BookCode,
			CancelType: body.CancelType, CancelPassed: body.CancelPassed, Description: body.Description,
			Lines: make([]cheque.DocLineInput, len(body.Lines)),
		}
		for i, l := range body.Lines {
			input.Lines[i] = cheque.DocLineInput{ChequeID: l.ChequeID, FeeAmount: l.FeeAmount}
		}
		if t := body.Totals; t != nil {
			input.Totals = &cheque.DocTotals{Total: t.Total, Fee: t.Fee, Net: t.Net}
		}
		if action == actionCreate {
			saved, err = cheque.CreateDoc(a.ctx, a.db, h.ledger, a.scope, body.RequestID, input)
		} else {
			saved, err = cheque.UpdateDoc(a.ctx, a.db, h.ledger, a.scope, body.RequestID, input)
		}
	}
	if err != nil {
		return failure(request, err)
	}
	return respond(request, map[string]any{"data": saved})
}

// recalculate apply: ปรับสถานะในทะเบียนให้ตรงประวัติ (ต้องมีเหตุผล) ใช้สิทธิ์ update ของเมนู
func (h *Http) recalculateCommand(request microservice.IContext, body commandBody) error {
	if strings.TrimSpace(body.Action) != "apply" {
		return fail(request, http.StatusBadRequest, "cheque_err_action_invalid", "action", body.Action)
	}
	a := h.authorizeLedger(request, true, func(s glhttp.SessionAccess) bool { return s.Allows(recalculateScreen, actionUpdate) })
	if a == nil {
		return nil
	}
	result, err := cheque.ApplyRecalculate(a.ctx, a.db, a.scope, body.RequestID, body.Reason)
	if err != nil {
		return failure(request, err)
	}
	return respond(request, map[string]any{"data": result})
}
