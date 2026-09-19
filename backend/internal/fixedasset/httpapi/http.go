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
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"smlcloudplatform/internal/config"
	fa "smlcloudplatform/internal/fixedasset"
	"smlcloudplatform/internal/fixedasset/mcp"
	gl "smlcloudplatform/internal/generalledger"
	"smlcloudplatform/internal/goapi/mypg"
	access "smlcloudplatform/internal/organization/access"
	branchmodels "smlcloudplatform/internal/organization/branch/models"
	"smlcloudplatform/pkg/apperr"
	"smlcloudplatform/pkg/microservice"
)

// validHoldingRegex guards the holding code before it is interpolated into a
// PostgreSQL connection string by mypg.PgSqlFastConnect (mirrors the same
// check in generalledger/httpapi).
var validHoldingRegex = regexp.MustCompile(`^[A-Za-z0-9_]{1,63}$`)

type Http struct {
	ms        *microservice.Microservice
	cfg       config.IConfig
	pst       microservice.IPersisterMongo
	store     *fa.Store
	reporter  *fa.Reporter
	poster    *fa.GLPoster
	posterErr error
	mu        sync.Mutex
}

func NewHttp(ms *microservice.Microservice, cfg config.IConfig) *Http {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	h := &Http{
		ms:  ms,
		cfg: cfg,
		pst: pst,
	}
	return h
}

// initialize wires the Mongo-backed store and reporter. It has no PostgreSQL
// or Kafka dependency, so asset CRUD and reports keep working even when the
// general ledger broker is unreachable; only ensurePoster (used for the
// depreciations/disposals GL-posting commands) needs Kafka.
func (h *Http) initialize(ctx context.Context) error {
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
	h.store = fa.NewStore(db)
	h.reporter = fa.NewReporter(db)
	return nil
}

// ensurePoster wires the general ledger poster: fixed-asset journals are
// posted through generalledger.Store.Execute, which resolves the PostgreSQL
// projection per holding on every call (no cross-tenant *sql.DB caching,
// fixing P0-3) and requires the shared GL Kafka outbox to be configured.
func (h *Http) ensurePoster(ctx context.Context) (*fa.GLPoster, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.poster != nil {
		return h.poster, nil
	}
	if h.posterErr != nil {
		return nil, h.posterErr
	}
	col, err := h.pst.Exec(ctx, fa.Asset{})
	if err != nil {
		return nil, err
	}
	db := col.Database()
	projection := gl.NewPostgres(func(holding string) (*sql.DB, error) {
		if !validHoldingRegex.MatchString(holding) {
			return nil, fmt.Errorf("รหัสกลุ่มบริษัทไม่ถูกต้อง")
		}
		return mypg.PgSqlFastConnect(holding)
	})
	store := gl.NewPostgresStore(projection)
	h.poster = fa.NewGLPoster(db, store)
	return h.poster, nil
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

	if err := h.initialize(ctx); err != nil {
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
		poster, err := h.ensurePoster(ctx)
		if err != nil {
			return fail(request, http.StatusServiceUnavailable, "ไม่สามารถเชื่อมต่อระบบบัญชีแยกประเภทได้")
		}
		switch cmd.Action {
		case "post-gl":
			journal, err := poster.PostDepreciation(ctx, scope, cmd.FiscalYear, cmd.Period, cmd.Date, cmd.DocNo, now)
			if err != nil {
				return fail(request, http.StatusBadRequest, err.Error())
			}
			return success(request, map[string]interface{}{"success": true, "journal": journal})

		case "reverse-gl":
			err := poster.ReverseDepreciation(ctx, scope, cmd.DocNo, cmd.Reason, now)
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
			poster, err := h.ensurePoster(ctx)
			if err != nil {
				return fail(request, http.StatusServiceUnavailable, "ไม่สามารถเชื่อมต่อระบบบัญชีแยกประเภทได้")
			}
			disp, journal, err := poster.DisposeAsset(ctx, scope, *cmd.Disposal, now)
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
	if err := h.initialize(ctx); err != nil {
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
	if err := h.initialize(ctx); err != nil {
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
	if err := h.initialize(ctx); err != nil {
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
	if err := h.initialize(ctx); err != nil {
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
	if err := h.initialize(ctx); err != nil {
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
	if err := h.initialize(ctx); err != nil {
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
	if err := h.initialize(ctx); err != nil {
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
		var poster *fa.GLPoster
		if name == "fa_post_depreciation_to_gl" || name == "fa_dispose_asset" {
			poster, err = h.ensurePoster(ctx)
			if err != nil {
				return fail(request, http.StatusServiceUnavailable, "ไม่สามารถเชื่อมต่อระบบบัญชีแยกประเภทได้")
			}
		}
		mcpHandler := mcp.NewMCPHandler(h.store, poster, h.reporter)
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
