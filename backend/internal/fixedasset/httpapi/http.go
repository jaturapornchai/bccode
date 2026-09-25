package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"smlcloudplatform/internal/config"
	fa "smlcloudplatform/internal/fixedasset"
	"smlcloudplatform/internal/fixedasset/mcp"
	gl "smlcloudplatform/internal/generalledger"
	glhttp "smlcloudplatform/internal/generalledger/httpapi"
	"smlcloudplatform/internal/goapi/mypg"
	"smlcloudplatform/pkg/apperr"
	"smlcloudplatform/pkg/microservice"
)

// validHoldingRegex guards the holding code before it is interpolated into a
// PostgreSQL connection string by mypg.PgSqlFastConnect (mirrors the same
// check in generalledger/httpapi).
var validHoldingRegex = regexp.MustCompile(`^[A-Za-z0-9_]{1,63}$`)

type Http struct {
	ms       *microservice.Microservice
	store    *fa.Store
	reporter *fa.Reporter
	poster   *fa.GLPoster
}

// NewHttp wires fixed assets to the holding PostgreSQL database; depreciation and
// disposal journals post through the same PostgreSQL general ledger engine.
func NewHttp(ms *microservice.Microservice, cfg config.IConfig) *Http {
	connect := func(holding string) (*sql.DB, error) {
		if !validHoldingRegex.MatchString(holding) {
			return nil, fmt.Errorf("รหัสกลุ่มบริษัทไม่ถูกต้อง")
		}
		return mypg.PgSqlFastConnect(holding)
	}
	ledger := gl.NewPostgresStore(gl.NewPostgres(connect))
	return &Http{
		ms:       ms,
		store:    fa.NewStore(connect),
		reporter: fa.NewReporter(connect),
		poster:   fa.NewGLPoster(connect, ledger, checkJournalBranch),
	}
}

// checkJournalBranch holds depreciation and disposal vouchers to the GL screen's header-branch
// rule (audit 2026-09-24: this module posted without it).
func checkJournalBranch(ctx context.Context, scope fa.Scope, branch string) error {
	return glhttp.CheckJournalBranch(ctx, mypg.PgSqlFastConnect, gl.Scope{Holding: scope.Holding, Company: scope.Company, Branch: scope.Branch, Actor: scope.Actor}, branch)
}

func (h *Http) RegisterHttp() {
	h.ms.POST("/fa/v2/command", h.command)
	h.ms.POST("/fa/v2/mcp", h.mcpRPC)
	h.ms.GET("/fa/v2/assets", h.listAssets)
	h.ms.GET("/fa/v2/assets/:id", h.getAsset)
	h.ms.GET("/fa/v2/assets/:id/schedule", h.getAssetSchedule)
	h.ms.GET("/fa/v2/types", h.listTypes)
	h.ms.GET("/fa/v2/reports/schedule", h.reportSchedule)
	h.ms.GET("/fa/v2/reports/tax-reconciliation", h.reportTaxReconciliation)
}

func (h *Http) scope(ctx context.Context, request microservice.IContext) (fa.Scope, error) {
	scope, err := glhttp.ResolveSessionScope(ctx, request)
	if err != nil {
		return fa.Scope{}, fmt.Errorf("กรุณาเลือกบริษัทที่มีสิทธิ์ก่อนใช้งานระบบสินทรัพย์")
	}
	return fa.Scope{Holding: scope.Holding, Company: scope.Company, Branch: scope.Branch, Actor: scope.Actor}, nil
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

// failure keeps the machine code and field of ledger/field errors (e.g. fa_account_required on
// deprecexpenseaccountcode) so the screen can point at the input to fix; others stay 400.
func failure(request microservice.IContext, err error) error {
	if user, ok := gl.AsUserError(err); ok {
		appErr := user.ToAppError()
		request.Response(appErr.HTTPStatus, appErr.ToResponse())
		return nil
	}
	return fail(request, http.StatusBadRequest, err.Error())
}

func success(request microservice.IContext, data interface{}) error {
	request.Response(http.StatusOK, data)
	return nil
}

func (h *Http) command(request microservice.IContext) error {
	ctx, cancel := context.WithTimeout(request.Request().Context(), 30*time.Second)
	defer cancel()

	scope, err := h.scope(ctx, request)
	if err != nil {
		return fail(request, http.StatusForbidden, err.Error())
	}

	var cmd fa.Command
	if err := json.NewDecoder(request.Request().Body).Decode(&cmd); err != nil {
		return fail(request, http.StatusBadRequest, "รูปแบบคำสั่งไม่ถูกต้อง")
	}

	now := time.Now().UTC()

	switch cmd.Resource {
	case "assets":
		switch cmd.Action {
		case "create":
			if cmd.Asset == nil {
				return fail(request, http.StatusBadRequest, "กรุณาส่งข้อมูลสินทรัพย์")
			}
			created, err := h.store.CreateAsset(ctx, scope, *cmd.Asset, now)
			if err != nil {
				return fail(request, http.StatusBadRequest, err.Error())
			}
			return success(request, map[string]interface{}{"success": true, "data": created})

		case "update":
			if cmd.Asset == nil || cmd.ID == "" {
				return fail(request, http.StatusBadRequest, "กรุณาระบุรหัสและข้อมูลสินทรัพย์")
			}
			updated, err := h.store.UpdateAsset(ctx, scope, cmd.ID, *cmd.Asset, cmd.Version, now)
			if err != nil {
				return failure(request, err)
			}
			return success(request, map[string]interface{}{"success": true, "data": updated})

		case "delete":
			if cmd.ID == "" {
				return fail(request, http.StatusBadRequest, "กรุณาระบุรหัสสินทรัพย์ที่ต้องการลบ")
			}
			err := h.store.DeleteAsset(ctx, scope, cmd.ID, cmd.Version, now)
			if err != nil {
				return fail(request, http.StatusBadRequest, err.Error())
			}
			return success(request, map[string]interface{}{"success": true})

		case "recalculate":
			if cmd.AssetCode == "" {
				return fail(request, http.StatusBadRequest, "กรุณาระบุรหัสสินทรัพย์")
			}
			err := h.store.RecalculateAssetSchedule(ctx, scope, cmd.AssetCode, now)
			if err != nil {
				return failure(request, err)
			}
			return success(request, map[string]interface{}{"success": true})
		}

	case "types":
		if cmd.Action == "create" {
			if cmd.AssetType == nil {
				return fail(request, http.StatusBadRequest, "กรุณาส่งข้อมูลประเภทสินทรัพย์")
			}
			created, err := h.store.CreateAssetType(ctx, scope, *cmd.AssetType, now)
			if err != nil {
				return fail(request, http.StatusBadRequest, err.Error())
			}
			return success(request, map[string]interface{}{"success": true, "data": created})
		}

	case "depreciations":
		poster := h.poster
		switch cmd.Action {
		case "post-gl":
			journal, err := poster.PostDepreciation(ctx, scope, cmd.FiscalYear, cmd.Period, cmd.Date, cmd.DocNo, cmd.BranchCode, now)
			if err != nil {
				return failure(request, err)
			}
			return success(request, map[string]interface{}{"success": true, "journal": journal})

		case "reverse-gl":
			err := poster.ReverseDepreciation(ctx, scope, cmd.DocNo, cmd.Reason, now)
			if err != nil {
				return failure(request, err)
			}
			return success(request, map[string]interface{}{"success": true})
		}

	case "disposals":
		if cmd.Action == "dispose" {
			if cmd.Disposal == nil {
				return fail(request, http.StatusBadRequest, "กรุณาส่งข้อมูลการจำหน่ายสินทรัพย์")
			}
			poster := h.poster
			disp, journal, err := poster.DisposeAsset(ctx, scope, *cmd.Disposal, now)
			if err != nil {
				return failure(request, err)
			}
			return success(request, map[string]interface{}{"success": true, "disposal": disp, "journal": journal})
		}
	}

	return fail(request, http.StatusBadRequest, "คำสั่งไม่ถูกต้องหรือไม่รองรับ")
}

func (h *Http) listAssets(request microservice.IContext) error {
	ctx, cancel := context.WithTimeout(request.Request().Context(), 30*time.Second)
	defer cancel()

	scope, err := h.scope(ctx, request)
	if err != nil {
		return fail(request, http.StatusForbidden, err.Error())
	}

	q := request.QueryParam("q")
	typeCode := request.QueryParam("type")
	status := request.QueryParam("status")
	page, _ := strconv.Atoi(request.QueryParam("page"))
	limit, _ := strconv.Atoi(request.QueryParam("limit"))

	items, total, err := h.store.ListAssets(ctx, scope, q, typeCode, status, page, limit)
	if err != nil {
		return fail(request, http.StatusInternalServerError, err.Error())
	}

	return success(request, map[string]interface{}{
		"success": true,
		"items":   items,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

func (h *Http) getAsset(request microservice.IContext) error {
	ctx, cancel := context.WithTimeout(request.Request().Context(), 30*time.Second)
	defer cancel()

	scope, err := h.scope(ctx, request)
	if err != nil {
		return fail(request, http.StatusForbidden, err.Error())
	}

	id := request.Param("id")
	asset, err := h.store.GetAsset(ctx, scope, id)
	if err != nil {
		if errors.Is(err, fa.ErrNotFound) {
			return fail(request, http.StatusNotFound, "ไม่พบสินทรัพย์")
		}
		return fail(request, http.StatusInternalServerError, err.Error())
	}
	return success(request, map[string]interface{}{"success": true, "data": asset})
}

func (h *Http) getAssetSchedule(request microservice.IContext) error {
	ctx, cancel := context.WithTimeout(request.Request().Context(), 30*time.Second)
	defer cancel()

	scope, err := h.scope(ctx, request)
	if err != nil {
		return fail(request, http.StatusForbidden, err.Error())
	}

	id := request.Param("id")
	items, err := h.store.GetAssetDepreciationSchedule(ctx, scope, id)
	if err != nil {
		return fail(request, http.StatusInternalServerError, err.Error())
	}
	return success(request, map[string]interface{}{"success": true, "items": items})
}

func (h *Http) listTypes(request microservice.IContext) error {
	ctx, cancel := context.WithTimeout(request.Request().Context(), 30*time.Second)
	defer cancel()

	scope, err := h.scope(ctx, request)
	if err != nil {
		return fail(request, http.StatusForbidden, err.Error())
	}

	items, err := h.store.ListAssetTypes(ctx, scope)
	if err != nil {
		return fail(request, http.StatusInternalServerError, err.Error())
	}
	return success(request, map[string]interface{}{"success": true, "items": items})
}

func (h *Http) reportSchedule(request microservice.IContext) error {
	ctx, cancel := context.WithTimeout(request.Request().Context(), 30*time.Second)
	defer cancel()

	scope, err := h.scope(ctx, request)
	if err != nil {
		return fail(request, http.StatusForbidden, err.Error())
	}

	fiscalYear := request.QueryParam("fiscalyear")
	period, _ := strconv.Atoi(request.QueryParam("period"))
	typeCode := request.QueryParam("type")

	rep, err := h.reporter.GetAssetScheduleReport(ctx, scope, fiscalYear, period, typeCode)
	if err != nil {
		return fail(request, http.StatusInternalServerError, err.Error())
	}
	return success(request, map[string]interface{}{"success": true, "report": rep})
}

func (h *Http) reportTaxReconciliation(request microservice.IContext) error {
	ctx, cancel := context.WithTimeout(request.Request().Context(), 30*time.Second)
	defer cancel()

	scope, err := h.scope(ctx, request)
	if err != nil {
		return fail(request, http.StatusForbidden, err.Error())
	}

	fiscalYear := request.QueryParam("fiscalyear")
	rep, err := h.reporter.GetTaxReconciliationReport(ctx, scope, fiscalYear)
	if err != nil {
		return fail(request, http.StatusInternalServerError, err.Error())
	}
	return success(request, map[string]interface{}{"success": true, "report": rep})
}

type mcpRequest struct {
	Method string                 `json:"method"`
	Params map[string]interface{} `json:"params"`
}

func (h *Http) mcpRPC(request microservice.IContext) error {
	ctx, cancel := context.WithTimeout(request.Request().Context(), 30*time.Second)
	defer cancel()

	scope, err := h.scope(ctx, request)
	if err != nil {
		return fail(request, http.StatusForbidden, err.Error())
	}

	var req mcpRequest
	if err := json.NewDecoder(request.Request().Body).Decode(&req); err != nil {
		return fail(request, http.StatusBadRequest, "Invalid JSON-RPC request")
	}

	switch req.Method {
	case "tools/list":
		return success(request, map[string]interface{}{
			"tools": mcp.Tools,
		})
	case "tools/call":
		name, _ := req.Params["name"].(string)
		args, _ := req.Params["arguments"].(map[string]interface{})
		mcpHandler := mcp.NewMCPHandler(h.store, h.poster, h.reporter)
		res, err := mcpHandler.HandleToolCall(ctx, scope, name, args)
		if err != nil {
			return success(request, map[string]interface{}{
				"isError": true,
				"content": []map[string]string{{"type": "text", "text": err.Error()}},
			})
		}
		rawJSON, _ := json.Marshal(res)
		return success(request, map[string]interface{}{
			"content": []map[string]string{{"type": "text", "text": string(rawJSON)}},
		})
	default:
		return fail(request, http.StatusBadRequest, "Method not supported")
	}
}
