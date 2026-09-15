package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"smlcloudplatform/internal/config"
	fa "smlcloudplatform/internal/fixedasset"
	"smlcloudplatform/internal/fixedasset/mcp"
	"smlcloudplatform/internal/goapi/mypg"
	access "smlcloudplatform/internal/organization/access"
	branchmodels "smlcloudplatform/internal/organization/branch/models"
	"smlcloudplatform/pkg/apperr"
	"smlcloudplatform/pkg/microservice"
)

type Http struct {
	ms         *microservice.Microservice
	pst        microservice.IPersisterMongo
	store      *fa.Store
	poster     *fa.GLPoster
	reporter   *fa.Reporter
	mcpHandler *mcp.MCPHandler
	mu         sync.Mutex
}

func NewHttp(ms *microservice.Microservice, cfg config.IConfig) *Http {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	return &Http{
		ms:  ms,
		pst: pst,
	}
}

func (h *Http) initialize(ctx context.Context, holding string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.store != nil {
		return nil
	}

	col, err := h.pst.Exec(ctx, fa.Asset{})
	if err != nil {
		return err
	}
	db := col.Database()

	var pgDB *sql.DB
	if holding != "" {
		if conn, err := mypg.PgSqlFastConnect(holding); err == nil {
			pgDB = conn
		}
	}

	h.store = fa.NewStore(db)
	h.poster = fa.NewGLPoster(db, pgDB)
	h.reporter = fa.NewReporter(db)
	h.mcpHandler = mcp.NewMCPHandler(h.store, h.poster, h.reporter)
	return nil
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
	u := request.UserInfo()
	result := fa.Scope{Holding: u.HoldingCode, Company: u.BusinessCode, Actor: u.UID}
	if u.HoldingCode == "" || u.BusinessCode == "" || u.CompanyUID == "" {
		return result, fmt.Errorf("กรุณาเลือกบริษัทก่อนใช้งานระบบสินทรัพย์")
	}
	membership, err := access.FindActiveMembership(ctx, h.pst, u, time.Now().UTC())
	if err != nil {
		return result, err
	}
	if err = access.RequireActiveHolding(ctx, h.pst, u.HoldingCode); err != nil {
		return result, err
	}
	if !access.AllowsCompany(membership.AccessScopes, u.CompanyUID) {
		return result, fmt.Errorf("ไม่มีสิทธิ์เข้าใช้บริษัทนี้")
	}
	if !access.AllowsAllBranches(membership.AccessScopes, u.CompanyUID) {
		if !access.AllowsBranch(membership.AccessScopes, u.CompanyUID, u.BranchUID) {
			return result, fmt.Errorf("ไม่มีสิทธิ์เข้าใช้สาขานี้")
		}
		var branch branchmodels.BranchDoc
		if err = h.pst.FindOne(ctx, branchmodels.BranchDoc{}, bson.M{"holdingcode": u.HoldingCode, "companyuid": u.CompanyUID, "branchuid": u.BranchUID, "isdeleted": false}, &branch); err != nil {
			return result, err
		}
		result.Branch = branch.Code
	}
	return result, nil
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

func (h *Http) command(request microservice.IContext) error {
	ctx, cancel := context.WithTimeout(request.Request().Context(), 30*time.Second)
	defer cancel()

	scope, err := h.scope(ctx, request)
	if err != nil {
		return fail(request, http.StatusForbidden, err.Error())
	}

	if err := h.initialize(ctx, scope.Holding); err != nil {
		return fail(request, http.StatusInternalServerError, "ไม่สามารถเชื่อมต่อฐานข้อมูลได้")
	}

	var cmd fa.Command
	if err := json.NewDecoder(request.Request().Body).Decode(&cmd); err != nil {
		return fail(request, http.StatusBadRequest, "รูปแบบคำสั่งไม่ถูกต้อง")
	}

	_ = h.store.EnsureIndexes(ctx)
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
				return fail(request, http.StatusBadRequest, err.Error())
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
				return fail(request, http.StatusBadRequest, err.Error())
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
		switch cmd.Action {
		case "post-gl":
			journal, err := h.poster.PostDepreciation(ctx, scope, cmd.FiscalYear, cmd.Period, cmd.Date, cmd.DocNo, now)
			if err != nil {
				return fail(request, http.StatusBadRequest, err.Error())
			}
			return success(request, map[string]interface{}{"success": true, "journal": journal})

		case "reverse-gl":
			err := h.poster.ReverseDepreciation(ctx, scope, cmd.DocNo, cmd.Reason, now)
			if err != nil {
				return fail(request, http.StatusBadRequest, err.Error())
			}
			return success(request, map[string]interface{}{"success": true})
		}

	case "disposals":
		if cmd.Action == "dispose" {
			if cmd.Disposal == nil {
				return fail(request, http.StatusBadRequest, "กรุณาส่งข้อมูลการจำหน่ายสินทรัพย์")
			}
			disp, journal, err := h.poster.DisposeAsset(ctx, scope, *cmd.Disposal, now)
			if err != nil {
				return fail(request, http.StatusBadRequest, err.Error())
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
	if err := h.initialize(ctx, scope.Holding); err != nil {
		return fail(request, http.StatusInternalServerError, "ไม่สามารถเชื่อมต่อฐานข้อมูลได้")
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
	if err := h.initialize(ctx, scope.Holding); err != nil {
		return fail(request, http.StatusInternalServerError, "ไม่สามารถเชื่อมต่อฐานข้อมูลได้")
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
	if err := h.initialize(ctx, scope.Holding); err != nil {
		return fail(request, http.StatusInternalServerError, "ไม่สามารถเชื่อมต่อฐานข้อมูลได้")
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
	if err := h.initialize(ctx, scope.Holding); err != nil {
		return fail(request, http.StatusInternalServerError, "ไม่สามารถเชื่อมต่อฐานข้อมูลได้")
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
	if err := h.initialize(ctx, scope.Holding); err != nil {
		return fail(request, http.StatusInternalServerError, "ไม่สามารถเชื่อมต่อฐานข้อมูลได้")
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
	if err := h.initialize(ctx, scope.Holding); err != nil {
		return fail(request, http.StatusInternalServerError, "ไม่สามารถเชื่อมต่อฐานข้อมูลได้")
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
	if err := h.initialize(ctx, scope.Holding); err != nil {
		return fail(request, http.StatusInternalServerError, "ไม่สามารถเชื่อมต่อฐานข้อมูลได้")
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
		res, err := h.mcpHandler.HandleToolCall(ctx, scope, name, args)
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
