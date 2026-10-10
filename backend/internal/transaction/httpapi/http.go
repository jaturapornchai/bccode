package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"smlcloudplatform/internal/config"
	"smlcloudplatform/internal/doctype"
	gl "smlcloudplatform/internal/generalledger"
	glhttp "smlcloudplatform/internal/generalledger/httpapi"
	"smlcloudplatform/internal/goapi/language"
	"smlcloudplatform/internal/goapi/mypg"
	"smlcloudplatform/pkg/apperr"
	"smlcloudplatform/pkg/microservice"
)

type Http struct {
	ms    *microservice.Microservice
	store *Store
	// session ยืนยัน session + สิทธิ์รายเมนู (ฉีดได้เพื่อทดสอบโดยไม่ต้องมีฐานข้อมูลกลาง)
	session func(context.Context, microservice.IContext) (glhttp.SessionAccess, error)
}

func NewHttp(ms *microservice.Microservice, cfg config.IConfig) *Http {
	return &Http{
		ms:      ms,
		store:   NewStore(),
		session: glhttp.ResolveSessionAccess,
	}
}

func (h *Http) RegisterHttp() {
	h.ms.GET("/transaction/:module/list", h.list)
	h.ms.POST("/transaction/:module/approve-batch", h.approveBatch)
	h.ms.POST("/transaction/:module/unapprove-batch", h.unapproveBatch)
	h.ms.POST("/transaction/:module/cancel-batch", h.cancelBatch)
	h.ms.GET("/transaction/:module/:id", h.get)
	h.ms.POST("/transaction/:module", h.create)
	h.ms.PUT("/transaction/:module/:id", h.update)
	h.ms.DELETE("/transaction/:module/:id", h.delete)

	h.ms.POST("/transaction/:module/:id/approve", h.approve)
	h.ms.POST("/transaction/:module/:id/unapprove", h.unapprove)
	h.ms.POST("/transaction/:module/:id/cancel", h.cancel)
	h.ms.GET("/transaction/:module/:id/events", h.getEvents)

	h.ms.GET("/transaction/:module/delivery-dates", h.listDeliveryDates)
	h.ms.POST("/transaction/:module/delivery-dates", h.saveDeliveryDates)

	h.ms.GET("/transaction/goodsreceive/receipt-adjustments", h.listReceiptAdjustments)
	h.ms.POST("/transaction/goodsreceive/receipt-adjustments", h.processReceiptAdjustments)
	h.ms.GET("/transaction/:module/receipt-adjustments", h.listReceiptAdjustments)
	h.ms.POST("/transaction/:module/receipt-adjustments", h.processReceiptAdjustments)
}

// การกระทำที่ตรวจสิทธิ์: "" = เข้าจอ/อ่าน, อื่นๆ ต้องมีทั้ง <เมนู> และ <เมนู>:<action> (กฎเดียวกับ GL)
const (
	actionRead    = ""
	actionCreate  = "create"
	actionUpdate  = "update"
	actionDelete  = "delete"
	actionApprove = "approve"
	actionCancel  = "cancel"
)

// docAccess คือผลของ authorize: ประเภทเอกสารที่ถูกต้อง session ที่ยืนยันแล้ว และฐานข้อมูลของ holding
type docAccess struct {
	ctx   context.Context
	scope gl.Scope
	db    *sql.DB
	dt    doctype.DocType
}

// authorize ตรวจตามลำดับ module → session/สมาชิกภาพ → สิทธิ์รายเมนูของประเภทเอกสาร → เชื่อมฐานข้อมูล
// ก่อนแตะข้อมูลใดๆ; ปฏิเสธ = ตอบ response เองแล้วคืน nil (ผู้เรียกคืน nil ต่อทันที)
func (h *Http) authorize(request microservice.IContext, action string) *docAccess {
	ctx := request.Request().Context()
	lang := requestLanguage(request)

	dt, ok := lookupDocType(request.Param("module"))
	if !ok {
		_ = fail(request, http.StatusBadRequest, fmt.Sprintf("โมดูลไม่ถูกต้อง (%s)", request.Param("module")))
		return nil
	}
	session, err := h.session(ctx, request)
	if err != nil {
		status, body := glhttp.SessionFailure(err, lang)
		request.Response(status, body)
		return nil
	}
	if !allowsDocType(session, dt, action) {
		key := "erp_err_action_forbidden"
		if action == actionRead {
			key = "erp_err_read_forbidden"
		}
		_ = forbidden(request, key, lang)
		return nil
	}
	db, err := mypg.PgSqlFastConnect(session.Scope.Holding)
	if err != nil {
		_ = fail(request, http.StatusServiceUnavailable, fmt.Sprintf("เชื่อมต่อฐานข้อมูลล้มเหลว: %v", err))
		return nil
	}
	if err := ensureSchema(ctx, db); err != nil {
		_ = fail(request, http.StatusInternalServerError, fmt.Sprintf("เตรียมตารางเอกสารล้มเหลว: %v", err))
		return nil
	}
	// เอกสารที่ลงบัญชีเขียนใบสำคัญและบัญชีย่อยในธุรกรรมเดียวกัน ตารางบัญชีต้องพร้อมก่อน (อ่านอย่างเดียวไม่ต้อง)
	if action != actionRead && doctype.PostsToGL(dt.Flag) {
		if err := gl.EnsureSchema(ctx, db); err != nil {
			_ = fail(request, http.StatusInternalServerError, fmt.Sprintf("เตรียมตารางบัญชีล้มเหลว: %v", err))
			return nil
		}
	}
	return &docAccess{ctx: ctx, scope: session.Scope, db: db, dt: dt}
}

// menuAccess คือส่วนของ glhttp.SessionAccess ที่ใช้ตรวจสิทธิ์ (แยกไว้ให้ทดสอบโดยไม่ต้องมีฐานข้อมูลกลาง)
type menuAccess interface {
	Allows(screen, action string) bool
}

// allowsDocType: สิทธิ์เมนูใดเมนูหนึ่งที่เปิดจอของประเภทนี้ก็พอ (stock-count มี 2 จอ);
// ประเภทที่ไม่มี menu id = ปฏิเสธ (fail-closed)
func allowsDocType(session menuAccess, dt doctype.DocType, action string) bool {
	for _, screen := range doctype.Screens(dt.Key) {
		if screen != "" {
			if session.Allows(screen, action) {
				return true
			}
			if action == actionCancel && session.Allows(screen, actionDelete) {
				return true
			}
			if action == actionApprove && session.Allows(screen, actionUpdate) {
				return true
			}
			if action == actionApprove && isApproveScreen(screen) && session.Allows(screen, "") {
				return true
			}
			if action == actionCancel && isCancelScreen(screen) && session.Allows(screen, "") {
				return true
			}
		}
	}
	return false
}

func isApproveScreen(screen string) bool {
	return strings.HasSuffix(screen, "-approve") || strings.Contains(screen, "approve")
}

func isCancelScreen(screen string) bool {
	return strings.HasSuffix(screen, "-cancel") || strings.Contains(screen, "cancel")
}

func requestLanguage(request microservice.IContext) string {
	if lang := request.QueryParam("lang"); lang != "" {
		return lang
	}
	if lang := request.Header("Accept-Language"); lang != "" {
		return lang
	}
	return "th"
}

// forbidden ตอบ 403 พร้อมข้อความตามภาษาของคำขอ (key อยู่ใน languages.tsv)
func forbidden(request microservice.IContext, key, lang string) error {
	appErr := &apperr.AppError{
		Code:       "forbidden",
		Message:    language.Text(key, lang),
		ThaiMsg:    language.Text(key, "th"),
		HTTPStatus: http.StatusForbidden,
	}
	request.Response(http.StatusForbidden, appErr.ToResponse())
	return nil
}

func lookupDocType(module string) (doctype.DocType, bool) {
	return doctype.ByKey(strings.ToLower(strings.TrimSpace(module)))
}

func parseTransFlag(module string) (int, string, error) {
	dt, ok := lookupDocType(module)
	if !ok {
		return 0, "", fmt.Errorf("โมดูลไม่ถูกต้อง (%s)", module)
	}
	return dt.Flag, dt.Prefix, nil
}

func fail(request microservice.IContext, status int, message string) error {
	appErr := &apperr.AppError{
		Code:       "error",
		Message:    message,
		ThaiMsg:    message,
		HTTPStatus: status,
	}
	request.Response(status, appErr.ToResponse())
	return nil
}

func success(request microservice.IContext, data interface{}) error {
	request.Response(http.StatusOK, data)
	return nil
}

func (h *Http) list(request microservice.IContext) error {
	access := h.authorize(request, actionRead)
	if access == nil {
		return nil
	}
	ctx, scope, db := access.ctx, access.scope, access.db
	transflag := access.dt.Flag

	q := request.QueryParam("q")
	status := request.QueryParam("status")
	limit, _ := strconv.Atoi(request.QueryParam("limit"))
	offset, _ := strconv.Atoi(request.QueryParam("offset"))

	docs, total, err := h.store.ListDocs(ctx, db, scope.Company, transflag, q, status, offset, limit)
	if err != nil {
		return fail(request, http.StatusInternalServerError, fmt.Sprintf("โหลดข้อมูลไม่สำเร็จ: %v", err))
	}

	page := 1
	if limit > 0 {
		page = (offset / limit) + 1
	}

	return success(request, DocListResponse{
		Success: true,
		Data:    docs,
		Pagination: Pagination{
			Total:  total,
			Page:   page,
			Limit:  limit,
			Offset: offset,
		},
	})
}

func (h *Http) get(request microservice.IContext) error {
	access := h.authorize(request, actionRead)
	if access == nil {
		return nil
	}
	ctx, scope, db := access.ctx, access.scope, access.db
	transflag := access.dt.Flag

	id := request.Param("id")
	if strings.TrimSpace(id) == "" {
		return fail(request, http.StatusBadRequest, "กรุณาระบุเลขที่เอกสาร")
	}

	doc, err := h.store.GetDoc(ctx, db, scope.Company, transflag, id)
	if err != nil {
		return fail(request, http.StatusInternalServerError, fmt.Sprintf("ดึงข้อมูลไม่สำเร็จ: %v", err))
	}
	if doc == nil {
		return fail(request, http.StatusNotFound, "ไม่พบเอกสารที่ต้องการ")
	}

	return success(request, DocItemResponse{
		Success: true,
		Data:    *doc,
	})
}

// docFailure ตอบความผิดพลาดของการบันทึก/ลบ: ความผิดจากกฎบัญชี/เอกสาร (ปีบัญชีปิด ตัดหนี้แล้ว เลขซ้ำ ...) ตอบแบบเดียวกับจอบัญชี
// คือข้อความที่ผู้ใช้แก้ได้ตามภาษาของคำขอ; ไม่พบเอกสาร = 404 (notFound ว่าง = ไม่ใช่กรณีนี้); อื่นๆ = 500
func docFailure(request microservice.IContext, err error, notFound, generic string) error {
	var ledger *ledgerError
	if _, isUser := gl.AsUserError(err); isUser || errors.As(err, &ledger) {
		status, body := glhttp.SessionFailure(err, requestLanguage(request))
		request.Response(status, body)
		return nil
	}
	if notFound != "" && errors.Is(err, sql.ErrNoRows) {
		return fail(request, http.StatusNotFound, notFound)
	}
	return fail(request, http.StatusInternalServerError, fmt.Sprintf("%s: %v", generic, err))
}

func (h *Http) create(request microservice.IContext) error {
	access := h.authorize(request, actionCreate)
	if access == nil {
		return nil
	}
	var doc ErpTransactionDoc
	if err := json.NewDecoder(request.Request().Body).Decode(&doc); err != nil {
		return fail(request, http.StatusBadRequest, "รูปแบบข้อมูลไม่ถูกต้อง")
	}

	saved, err := h.store.SaveDoc(access.ctx, access.db, access.scope, access.dt, &doc, false)
	if err != nil {
		return docFailure(request, err, "", "บันทึกเอกสารไม่สำเร็จ")
	}

	return success(request, DocItemResponse{
		Success: true,
		Data:    *saved,
		Message: "save_success",
	})
}

func (h *Http) update(request microservice.IContext) error {
	access := h.authorize(request, actionUpdate)
	if access == nil {
		return nil
	}
	id := request.Param("id")
	if strings.TrimSpace(id) == "" {
		return fail(request, http.StatusBadRequest, "กรุณาระบุเลขที่เอกสาร")
	}

	var doc ErpTransactionDoc
	if err := json.NewDecoder(request.Request().Body).Decode(&doc); err != nil {
		return fail(request, http.StatusBadRequest, "รูปแบบข้อมูลไม่ถูกต้อง")
	}
	if doc.DocNo == "" {
		doc.DocNo = id
	}

	saved, err := h.store.SaveDoc(access.ctx, access.db, access.scope, access.dt, &doc, true)
	if err != nil {
		return docFailure(request, err, "ไม่พบเอกสารที่ต้องการแก้ไข", "แก้ไขเอกสารไม่สำเร็จ")
	}

	return success(request, DocItemResponse{
		Success: true,
		Data:    *saved,
		Message: "save_success",
	})
}

func (h *Http) delete(request microservice.IContext) error {
	access := h.authorize(request, actionDelete)
	if access == nil {
		return nil
	}
	id := request.Param("id")
	if strings.TrimSpace(id) == "" {
		return fail(request, http.StatusBadRequest, "กรุณาระบุเลขที่เอกสาร")
	}

	if err := h.store.DeleteDoc(access.ctx, access.db, access.scope, access.dt, id); err != nil {
		return docFailure(request, err, "ไม่พบเอกสารที่ต้องการลบ", "ลบเอกสารไม่สำเร็จ")
	}

	return success(request, map[string]any{
		"success": true,
		"message": "delete_success",
	})
}

type reasonPayload struct {
	Reason string `json:"reason"`
}

type batchActionPayload struct {
	IDs    []string `json:"ids"`
	DocNos []string `json:"docnos"`
	Reason string   `json:"reason"`
}

func (p *batchActionPayload) docIDs() []string {
	if len(p.IDs) > 0 {
		return p.IDs
	}
	return p.DocNos
}

func (h *Http) approve(request microservice.IContext) error {
	access := h.authorize(request, actionApprove)
	if access == nil {
		return nil
	}
	id := request.Param("id")
	if strings.TrimSpace(id) == "" {
		return fail(request, http.StatusBadRequest, "กรุณาระบุเลขที่เอกสาร")
	}

	doc, err := h.store.ApproveDoc(access.ctx, access.db, access.scope, access.dt, id)
	if err != nil {
		return docFailure(request, err, "ไม่พบเอกสารที่ต้องการอนุมัติ", "อนุมัติเอกสารไม่สำเร็จ")
	}

	return success(request, DocItemResponse{
		Success: true,
		Data:    *doc,
		Message: "erp_msg_approved",
	})
}

func (h *Http) unapprove(request microservice.IContext) error {
	access := h.authorize(request, actionApprove)
	if access == nil {
		return nil
	}
	id := request.Param("id")
	if strings.TrimSpace(id) == "" {
		return fail(request, http.StatusBadRequest, "กรุณาระบุเลขที่เอกสาร")
	}

	var payload reasonPayload
	_ = json.NewDecoder(request.Request().Body).Decode(&payload)

	doc, err := h.store.UnapproveDoc(access.ctx, access.db, access.scope, access.dt, id, payload.Reason)
	if err != nil {
		return docFailure(request, err, "ไม่พบเอกสารที่ต้องการยกเลิกการอนุมัติ", "ยกเลิกการอนุมัติไม่สำเร็จ")
	}

	return success(request, DocItemResponse{
		Success: true,
		Data:    *doc,
		Message: "erp_msg_unapproved",
	})
}

func (h *Http) cancel(request microservice.IContext) error {
	access := h.authorize(request, actionCancel)
	if access == nil {
		return nil
	}
	id := request.Param("id")
	if strings.TrimSpace(id) == "" {
		return fail(request, http.StatusBadRequest, "กรุณาระบุเลขที่เอกสาร")
	}

	var payload reasonPayload
	if err := json.NewDecoder(request.Request().Body).Decode(&payload); err != nil {
		return fail(request, http.StatusBadRequest, "รูปแบบข้อมูลไม่ถูกต้อง")
	}

	doc, err := h.store.CancelDoc(access.ctx, access.db, access.scope, access.dt, id, payload.Reason)
	if err != nil {
		return docFailure(request, err, "ไม่พบเอกสารที่ต้องการยกเลิก", "ยกเลิกเอกสารไม่สำเร็จ")
	}

	return success(request, DocItemResponse{
		Success: true,
		Data:    *doc,
		Message: "erp_msg_cancelled",
	})
}

func (h *Http) approveBatch(request microservice.IContext) error {
	access := h.authorize(request, actionApprove)
	if access == nil {
		return nil
	}

	var payload batchActionPayload
	if err := json.NewDecoder(request.Request().Body).Decode(&payload); err != nil {
		return fail(request, http.StatusBadRequest, "รูปแบบข้อมูลไม่ถูกต้อง")
	}

	results, err := h.store.ApproveBatchDocs(access.ctx, access.db, access.scope, access.dt, payload.docIDs())
	if err != nil {
		return docFailure(request, err, "", "อนุมัติเอกสารไม่สำเร็จ")
	}

	return success(request, BatchApproveResponse{
		Success: true,
		Data:    results,
	})
}

func (h *Http) unapproveBatch(request microservice.IContext) error {
	access := h.authorize(request, actionApprove)
	if access == nil {
		return nil
	}

	var payload batchActionPayload
	if err := json.NewDecoder(request.Request().Body).Decode(&payload); err != nil {
		return fail(request, http.StatusBadRequest, "รูปแบบข้อมูลไม่ถูกต้อง")
	}

	results, err := h.store.UnapproveBatchDocs(access.ctx, access.db, access.scope, access.dt, payload.docIDs(), payload.Reason)
	if err != nil {
		return docFailure(request, err, "", "ยกเลิกการอนุมัติเอกสารไม่สำเร็จ")
	}

	return success(request, BatchApproveResponse{
		Success: true,
		Data:    results,
	})
}

func (h *Http) cancelBatch(request microservice.IContext) error {
	access := h.authorize(request, actionCancel)
	if access == nil {
		return nil
	}

	var payload batchActionPayload
	if err := json.NewDecoder(request.Request().Body).Decode(&payload); err != nil {
		return fail(request, http.StatusBadRequest, "รูปแบบข้อมูลไม่ถูกต้อง")
	}
	reason := strings.TrimSpace(payload.Reason)
	if reason == "" {
		return fail(request, http.StatusBadRequest, "erp_err_cancel_reason_required")
	}

	results, err := h.store.CancelBatchDocs(access.ctx, access.db, access.scope, access.dt, payload.docIDs(), reason)
	if err != nil {
		return docFailure(request, err, "", "ยกเลิกเอกสารไม่สำเร็จ")
	}

	return success(request, BatchApproveResponse{
		Success: true,
		Data:    results,
	})
}

func (h *Http) getEvents(request microservice.IContext) error {
	access := h.authorize(request, actionRead)
	if access == nil {
		return nil
	}
	id := request.Param("id")
	if strings.TrimSpace(id) == "" {
		return fail(request, http.StatusBadRequest, "กรุณาระบุเลขที่เอกสาร")
	}

	events, err := h.store.GetDocEvents(access.ctx, access.db, access.scope.Company, access.dt.Flag, id)
	if err != nil {
		return docFailure(request, err, "ไม่พบเอกสาร", "ดึงประวัติเอกสารไม่สำเร็จ")
	}

	return success(request, DocEventsResponse{
		Success: true,
		Data:    events,
	})
}
