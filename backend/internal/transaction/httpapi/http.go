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
	"smlcloudplatform/internal/goapi/mypg"
	"smlcloudplatform/pkg/apperr"
	"smlcloudplatform/pkg/microservice"
)

type Http struct {
	ms    *microservice.Microservice
	store *Store
}

func NewHttp(ms *microservice.Microservice, cfg config.IConfig) *Http {
	return &Http{
		ms:    ms,
		store: NewStore(),
	}
}

func (h *Http) RegisterHttp() {
	h.ms.GET("/transaction/:module/list", h.list)
	h.ms.GET("/transaction/:module/:id", h.get)
	h.ms.POST("/transaction/:module", h.create)
	h.ms.PUT("/transaction/:module/:id", h.update)
	h.ms.DELETE("/transaction/:module/:id", h.delete)
}

func (h *Http) scope(ctx context.Context, request microservice.IContext) (gl.Scope, *sql.DB, error) {
	scope, err := glhttp.ResolveSessionScope(ctx, request)
	if err != nil {
		return gl.Scope{}, nil, fmt.Errorf("กรุณาเลือกบริษัทที่มีสิทธิ์ก่อนใช้งาน: %w", err)
	}
	db, err := mypg.PgSqlFastConnect(scope.Holding)
	if err != nil {
		return scope, nil, fmt.Errorf("เชื่อมต่อฐานข้อมูลล้มเหลว: %w", err)
	}
	if err := ensureSchema(ctx, db); err != nil {
		return scope, nil, fmt.Errorf("เตรียมตารางเอกสารล้มเหลว: %w", err)
	}
	return scope, db, nil
}

func parseTransFlag(module string) (int, string, error) {
	dt, ok := doctype.ByKey(strings.ToLower(strings.TrimSpace(module)))
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
	ctx := request.Request().Context()
	scope, db, err := h.scope(ctx, request)
	if err != nil {
		return fail(request, http.StatusUnauthorized, err.Error())
	}

	module := request.Param("module")
	transflag, _, err := parseTransFlag(module)
	if err != nil {
		return fail(request, http.StatusBadRequest, err.Error())
	}

	q := request.QueryParam("q")
	limit, _ := strconv.Atoi(request.QueryParam("limit"))
	offset, _ := strconv.Atoi(request.QueryParam("offset"))

	docs, total, err := h.store.ListDocs(ctx, db, scope.Company, transflag, q, offset, limit)
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
	ctx := request.Request().Context()
	scope, db, err := h.scope(ctx, request)
	if err != nil {
		return fail(request, http.StatusUnauthorized, err.Error())
	}

	module := request.Param("module")
	transflag, _, err := parseTransFlag(module)
	if err != nil {
		return fail(request, http.StatusBadRequest, err.Error())
	}

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

func (h *Http) create(request microservice.IContext) error {
	ctx := request.Request().Context()
	scope, db, err := h.scope(ctx, request)
	if err != nil {
		return fail(request, http.StatusUnauthorized, err.Error())
	}

	module := request.Param("module")
	transflag, prefix, err := parseTransFlag(module)
	if err != nil {
		return fail(request, http.StatusBadRequest, err.Error())
	}

	var doc ErpTransactionDoc
	if err := json.NewDecoder(request.Request().Body).Decode(&doc); err != nil {
		return fail(request, http.StatusBadRequest, "รูปแบบข้อมูลไม่ถูกต้อง")
	}

	saved, err := h.store.SaveDoc(ctx, db, scope.Company, scope.Branch, transflag, prefix, &doc, scope.Actor, false)
	if err != nil {
		return fail(request, http.StatusInternalServerError, fmt.Sprintf("บันทึกเอกสารไม่สำเร็จ: %v", err))
	}

	return success(request, DocItemResponse{
		Success: true,
		Data:    *saved,
		Message: "save_success",
	})
}

func (h *Http) update(request microservice.IContext) error {
	ctx := request.Request().Context()
	scope, db, err := h.scope(ctx, request)
	if err != nil {
		return fail(request, http.StatusUnauthorized, err.Error())
	}

	module := request.Param("module")
	transflag, prefix, err := parseTransFlag(module)
	if err != nil {
		return fail(request, http.StatusBadRequest, err.Error())
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

	saved, err := h.store.SaveDoc(ctx, db, scope.Company, scope.Branch, transflag, prefix, &doc, scope.Actor, true)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fail(request, http.StatusNotFound, "ไม่พบเอกสารที่ต้องการแก้ไข")
		}
		return fail(request, http.StatusInternalServerError, fmt.Sprintf("แก้ไขเอกสารไม่สำเร็จ: %v", err))
	}

	return success(request, DocItemResponse{
		Success: true,
		Data:    *saved,
		Message: "save_success",
	})
}

func (h *Http) delete(request microservice.IContext) error {
	ctx := request.Request().Context()
	scope, db, err := h.scope(ctx, request)
	if err != nil {
		return fail(request, http.StatusUnauthorized, err.Error())
	}

	module := request.Param("module")
	transflag, _, err := parseTransFlag(module)
	if err != nil {
		return fail(request, http.StatusBadRequest, err.Error())
	}

	id := request.Param("id")
	if strings.TrimSpace(id) == "" {
		return fail(request, http.StatusBadRequest, "กรุณาระบุเลขที่เอกสาร")
	}

	if err := h.store.DeleteDoc(ctx, db, scope.Company, transflag, id, scope.Actor); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fail(request, http.StatusNotFound, "ไม่พบเอกสารที่ต้องการลบ")
		}
		return fail(request, http.StatusInternalServerError, fmt.Sprintf("ลบเอกสารไม่สำเร็จ: %v", err))
	}

	return success(request, map[string]any{
		"success": true,
		"message": "delete_success",
	})
}
