package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"smlcloudplatform/pkg/microservice"
)

// prepareMenu คือ menu id ของจอบันทึกเตรียมจ่ายชำระหนี้ (frontend menu-data.ts) — สิทธิ์ <เมนู> + <เมนู>:<action> แบบเดียวกับเอกสาร ERP
const prepareMenu = "ap-payment-prepare"

// registerPrepare: route group ของตัวเอง ไม่ผ่าน /transaction/:module เพราะใบเตรียมจ่ายไม่ใช่ doctype (design §4.4)
func (h *Http) registerPrepare() {
	h.ms.GET("/ap-payment-prepare/list", h.listPrepares)
	h.ms.GET("/ap-payment-prepare/outstanding", h.prepareOutstanding)
	h.ms.GET("/ap-payment-prepare/creditors", h.listPrepareCreditors)
	h.ms.GET("/ap-payment-prepare/:docno", h.getPrepare)
	h.ms.GET("/ap-payment-prepare/:docno/outstanding", h.prepareOutstanding)
	h.ms.GET("/ap-payment-prepare/:docno/events", h.prepareEvents)
	h.ms.POST("/ap-payment-prepare", h.createPrepare)
	h.ms.PUT("/ap-payment-prepare/:docno", h.updatePrepare)
	h.ms.POST("/ap-payment-prepare/:docno/cancel", h.cancelPrepare)
}

// allowsPrepare: ยกเลิกใช้สิทธิ์ cancel หรือ delete ของเมนูได้ (ปุ่ม Del ของ Champ = ยกเลิกระดับหัวเอกสาร)
func allowsPrepare(session menuAccess, action string) bool {
	return session.Allows(prepareMenu, action) || (action == actionCancel && session.Allows(prepareMenu, actionDelete))
}

// prepareAccess: ทุกคำขอ (รวมอ่าน) ต้องมีตารางบัญชี เพราะยอดค้างจ่ายอ่านจากบัญชีย่อยเจ้าหนี้
func (h *Http) prepareAccess(request microservice.IContext, action string) *docAccess {
	return h.openSession(request, action, func(session menuAccess) bool { return allowsPrepare(session, action) }, true)
}

// preparePage อ่าน page (เริ่ม 1) + limit (ค่าเริ่ม 50, สูงสุด 200) เป็น limit/offset
func preparePage(request microservice.IContext) (limit, offset, page int) {
	limit, _ = strconv.Atoi(request.QueryParam("limit"))
	if limit < 1 || limit > 200 {
		limit = 50
	}
	page, _ = strconv.Atoi(request.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	return limit, (page - 1) * limit, page
}

type prepareListResponse struct {
	Success    bool        `json:"success"`
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

type prepareItemResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

func pageResponse(data interface{}, total, limit, offset, page int) prepareListResponse {
	return prepareListResponse{Success: true, Data: data, Pagination: Pagination{Total: total, Page: page, Limit: limit, Offset: offset}}
}

func (h *Http) listPrepares(request microservice.IContext) error {
	access := h.prepareAccess(request, actionRead)
	if access == nil {
		return nil
	}
	limit, offset, page := preparePage(request)
	status := request.QueryParam("status")
	switch status {
	case "", "open", "paid", "cancelled":
	default:
		return fail(request, http.StatusBadRequest, "erp_err_invalid_payload")
	}
	rows, total, err := h.store.ListPrepares(access.ctx, access.db, access.scope.Company, apPrepareFilter{
		Q: request.QueryParam("q"), Partner: request.QueryParam("partner"), FromDate: request.QueryParam("fromdate"),
		ToDate: request.QueryParam("todate"), Status: status, Limit: limit, Offset: offset,
	})
	if err != nil {
		return docFailure(request, err, false)
	}
	return success(request, pageResponse(rows, total, limit, offset, page))
}

// prepareOutstanding: /outstanding = ใบใหม่, /:docno/outstanding = แก้ใบเดิม (ไม่นับยอดจองของตัวเอง); todate ว่าง = วันนี้
func (h *Http) prepareOutstanding(request microservice.IContext) error {
	access := h.prepareAccess(request, actionRead)
	if access == nil {
		return nil
	}
	limit, offset, page := preparePage(request)
	asOf := strings.TrimSpace(request.QueryParam("todate"))
	if asOf == "" {
		asOf = businessDay(time.Now())
	}
	rows, total, err := h.store.PrepareOutstanding(access.ctx, access.db, access.scope, request.QueryParam("partner"), asOf,
		request.QueryParam("q"), strings.TrimSpace(request.Param("docno")), limit, offset)
	if err != nil {
		return docFailure(request, err, false)
	}
	return success(request, pageResponse(rows, total, limit, offset, page))
}

func (h *Http) listPrepareCreditors(request microservice.IContext) error {
	access := h.prepareAccess(request, actionRead)
	if access == nil {
		return nil
	}
	limit, offset, page := preparePage(request)
	rows, total, err := h.store.ListCreditors(access.ctx, access.db, access.scope.Company, request.QueryParam("q"), limit, offset)
	if err != nil {
		return docFailure(request, err, false)
	}
	return success(request, pageResponse(rows, total, limit, offset, page))
}

func (h *Http) getPrepare(request microservice.IContext) error {
	access := h.prepareAccess(request, actionRead)
	if access == nil {
		return nil
	}
	doc, err := h.store.GetPrepare(access.ctx, access.db, access.scope.Company, strings.TrimSpace(request.Param("docno")))
	if err != nil {
		return docFailure(request, err, false)
	}
	if doc == nil {
		return docFailure(request, prepareNotFound(), false)
	}
	return success(request, prepareItemResponse{Success: true, Data: doc})
}

func prepareNotFound() error {
	e := erpError("ap_prepare_not_found", "docno")
	e.Status = http.StatusNotFound
	return e
}

func (h *Http) prepareEvents(request microservice.IContext) error {
	access := h.prepareAccess(request, actionRead)
	if access == nil {
		return nil
	}
	events, err := h.store.PrepareEvents(access.ctx, access.db, access.scope.Company, strings.TrimSpace(request.Param("docno")))
	if err != nil {
		return docFailure(request, err, false)
	}
	return success(request, prepareItemResponse{Success: true, Data: events})
}

// ขนาดคำขอและความยาวข้อความสูงสุดของใบเตรียมจ่าย (Rule 10)
const (
	maxPrepareBodyBytes = 1 << 20
	maxPrepareLines     = 1000
	maxPrepareCodeRunes = 50
	maxPrepareTextRunes = 500
)

// readPrepare อ่านคำขอบันทึก/แก้: ผิดรูปแบบ ใหญ่เกิน หรือข้อความยาวเกิน = 400 erp_err_invalid_payload (ตอบเองแล้วคืน false)
func readPrepare(request microservice.IContext, dst interface{}) bool {
	body := http.MaxBytesReader(request.ResponseWriter(), request.Request().Body, maxPrepareBodyBytes)
	if err := json.NewDecoder(body).Decode(dst); err != nil {
		_ = fail(request, http.StatusBadRequest, "erp_err_invalid_payload")
		return false
	}
	return true
}

func prepareTextsFit(in apPrepareInput) bool {
	if len(in.Lines) > maxPrepareLines || utf8.RuneCountInString(in.Remark) > maxPrepareTextRunes {
		return false
	}
	for _, code := range []string{in.DocNo, in.PayDate, in.PartnerCode, in.ProjectCode, in.DepartmentCode, in.DiscountAmount} {
		if utf8.RuneCountInString(code) > maxPrepareCodeRunes {
			return false
		}
	}
	for _, line := range in.Lines {
		if utf8.RuneCountInString(line.DebtDocumentID) > maxPrepareCodeRunes*2 || utf8.RuneCountInString(line.PayAmount) > maxPrepareCodeRunes ||
			utf8.RuneCountInString(line.Description) > maxPrepareTextRunes {
			return false
		}
	}
	return true
}

func (h *Http) createPrepare(request microservice.IContext) error {
	return h.savePrepare(request, actionCreate, "")
}

func (h *Http) updatePrepare(request microservice.IContext) error {
	docNo := strings.TrimSpace(request.Param("docno"))
	if docNo == "" {
		return fail(request, http.StatusBadRequest, "erp_err_docno_required")
	}
	return h.savePrepare(request, actionUpdate, docNo)
}

func (h *Http) savePrepare(request microservice.IContext, action, editNo string) error {
	access := h.prepareAccess(request, action)
	if access == nil {
		return nil
	}
	var in apPrepareInput
	if !readPrepare(request, &in) {
		return nil
	}
	if !prepareTextsFit(in) {
		return fail(request, http.StatusBadRequest, "erp_err_invalid_payload")
	}
	doc, err := h.store.SavePrepare(access.ctx, access.db, access.scope, in, editNo)
	if err != nil {
		return docFailure(request, err, false)
	}
	status := http.StatusOK
	if editNo == "" {
		status = http.StatusCreated
	}
	request.Response(status, prepareItemResponse{Success: true, Data: doc})
	return nil
}

func (h *Http) cancelPrepare(request microservice.IContext) error {
	access := h.prepareAccess(request, actionCancel)
	if access == nil {
		return nil
	}
	var in struct {
		Reason  string `json:"reason"`
		Version int64  `json:"version"`
	}
	if !readPrepare(request, &in) {
		return nil
	}
	if utf8.RuneCountInString(in.Reason) > maxPrepareTextRunes {
		return fail(request, http.StatusBadRequest, "erp_err_invalid_payload")
	}
	docNo := strings.TrimSpace(request.Param("docno"))
	if err := h.store.CancelPrepare(access.ctx, access.db, access.scope, docNo, in.Reason, in.Version); err != nil {
		return docFailure(request, err, false)
	}
	return success(request, prepareItemResponse{Success: true, Data: map[string]interface{}{"docno": docNo, "iscancel": true}})
}
