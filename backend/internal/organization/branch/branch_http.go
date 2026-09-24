package branch

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"

	authModels "smlcloudplatform/internal/authentication/models"
	"smlcloudplatform/internal/centraldb"
	"smlcloudplatform/internal/config"
	common "smlcloudplatform/internal/models"
	orgaccess "smlcloudplatform/internal/organization"
	orgpolicy "smlcloudplatform/internal/organization/access"
	branchModels "smlcloudplatform/internal/organization/branch/models"
	"smlcloudplatform/pkg/apperr"
	"smlcloudplatform/pkg/microservice"
)

type BranchHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
}

func NewBranchHttp(ms *microservice.Microservice, cfg config.IConfig) BranchHttp {
	return BranchHttp{
		ms:  ms,
		cfg: cfg,
	}
}

func (h BranchHttp) RegisterHttp() {
	h.ms.POST("/organization/branch", h.CreateBranch)
	h.ms.GET("/organization/branch", h.SearchBranch)
	h.ms.GET("/organization/branch/list", h.SearchBranchStep)
	h.ms.GET("/organization/branch/:id", h.InfoBranch)
	h.ms.PUT("/organization/branch/:id", h.UpdateBranch)
}

var (
	errBranchParentChange = errors.New("branch parent cannot be changed")
	errBranchCodeChange   = errors.New("branch code cannot be changed by ordinary update")
	errBranchAmbiguous    = errors.New("branch code exists in more than one company; companyuid is required")
)

const branchColumns = `company_code, code, name, names, settings, is_active, created_at, updated_at, created_by, updated_by`

func scanBranch(row interface{ Scan(...interface{}) error }, holdingCode string) (branchModels.BranchOrgDoc, error) {
	var (
		doc      branchModels.BranchOrgDoc
		name     string
		names    []byte
		settings []byte
	)
	err := row.Scan(&doc.CompanyUID, &doc.Code, &name, &names, &settings, &doc.IsActive, &doc.CreatedAt, &doc.UpdatedAt, &doc.CreatedBy, &doc.UpdatedBy)
	if err != nil {
		return doc, err
	}
	if len(settings) > 0 {
		if err := json.Unmarshal(settings, &doc.BranchSettings); err != nil {
			return doc, fmt.Errorf("decode branch settings: %w", err)
		}
	}
	doc.Names = orgaccess.DecodeNames(names, name)
	setBranchIdentity(&doc, holdingCode, doc.CompanyUID, doc.Code)
	normalizeBranchSlices(&doc.BranchSettings)
	return doc, nil
}

// setBranchIdentity fills the identity aliases the clients read: in PostgreSQL the
// branch UID is its code and the company UID is the company code.
func setBranchIdentity(doc *branchModels.BranchOrgDoc, holdingCode, companyCode, code string) {
	doc.HoldingCode = holdingCode
	doc.HoldingUID = holdingCode
	doc.CompanyUID = companyCode
	doc.CompanyGuid = companyCode
	doc.BusinessCode = companyCode
	doc.Code = code
	doc.BranchCode = code
	doc.BranchUID = code
	doc.GuidFixed = code
}

func normalizeBranchSlices(settings *branchModels.BranchSettings) {
	if settings.BusinessTypes == nil {
		settings.BusinessTypes = []string{}
	}
	if settings.Addresses == nil {
		settings.Addresses = []branchModels.BranchAddress{}
	}
	if settings.DocumentFormats == nil {
		settings.DocumentFormats = []branchModels.BranchDocFormat{}
	}
}

type rowQuerier interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

// findBranch loads one branch by code. Branch codes are unique per company only, so a
// code that exists in several companies needs the company code to be resolved.
func findBranch(ctx context.Context, q rowQuerier, holdingCode, companyCode, code string, lock bool) (branchModels.BranchOrgDoc, error) {
	query := `SELECT ` + branchColumns + ` FROM branches
		WHERE holding_code = $1 AND code = $2 AND ($3 = '' OR company_code = $3)
		ORDER BY company_code LIMIT 2`
	if lock {
		query += ` FOR UPDATE`
	}
	rows, err := q.QueryContext(ctx, query, holdingCode, lookupBranchCode(code), strings.TrimSpace(companyCode))
	if err != nil {
		return branchModels.BranchOrgDoc{}, err
	}
	defer rows.Close()
	found := make([]branchModels.BranchOrgDoc, 0, 2)
	for rows.Next() {
		doc, err := scanBranch(rows, holdingCode)
		if err != nil {
			return branchModels.BranchOrgDoc{}, err
		}
		found = append(found, doc)
	}
	if err := rows.Err(); err != nil {
		return branchModels.BranchOrgDoc{}, err
	}
	switch len(found) {
	case 0:
		return branchModels.BranchOrgDoc{}, sql.ErrNoRows
	case 1:
		return found[0], nil
	default:
		return branchModels.BranchOrgDoc{}, errBranchAmbiguous
	}
}

func lookupBranchCode(code string) string {
	if normalized, err := branchModels.NormalizeThaiTaxBranchCode(code); err == nil {
		return normalized
	}
	return strings.TrimSpace(code)
}

func requireActiveBranchCompany(ctx context.Context, q interface {
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}, holdingCode, companyCode string) error {
	var exists bool
	err := q.QueryRowContext(ctx, `SELECT true FROM companies WHERE holding_code = $1 AND code = $2 AND is_active = true`,
		holdingCode, strings.TrimSpace(companyCode)).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return branchModels.ErrBranchCompanyNotFound
	}
	return err
}

func (h BranchHttp) CreateBranch(ctx microservice.IContext) error {
	holdingCode := ctx.UserInfo().HoldingCode
	input := ctx.ReadInput()

	var req branchModels.BranchOrgDoc
	if err := json.Unmarshal([]byte(input), &req); err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	reqCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	db, err := centraldb.Open()
	if err != nil {
		return apperr.Respond(ctx, apperr.ErrInternal.WithWrap(err))
	}
	if authErr := orgaccess.RequireHoldingAdmin(reqCtx, db, ctx.UserInfo()); authErr != nil {
		return apperr.Respond(ctx, authErr)
	}
	if _, err := orgpolicy.FindActiveHoldingManager(reqCtx, db, ctx.UserInfo(), time.Now()); err != nil {
		return respondBranchMembershipError(ctx, err)
	}
	holding, err := orgpolicy.FindActiveHolding(reqCtx, db, holdingCode)
	if err != nil {
		return respondBranchMembershipError(ctx, err)
	}

	now := time.Now().UTC()
	if err := prepareBranchCreate(&req, holding.Code, ctx.UserInfo().Username, now); err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}
	names, err := orgaccess.EncodeNames(req.Names)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}
	settings, err := json.Marshal(req.BranchSettings)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	err = inTx(reqCtx, db, func(tx *sql.Tx) error {
		if err := requireActiveBranchCompany(reqCtx, tx, req.HoldingCode, req.CompanyUID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(reqCtx, `
			INSERT INTO branches (holding_code, company_code, code, name, names, settings, is_headquarters, is_active, created_by, updated_by, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, true, $8, '', $9, $9)`,
			req.HoldingCode, req.CompanyUID, req.Code, orgaccess.PrimaryName(req.Names), names, string(settings),
			isHeadquarters(req), req.CreatedBy, now); err != nil {
			return err
		}
		return orgaccess.RecordAudit(reqCtx, tx, orgaccess.Audit{
			HoldingCode: req.HoldingCode, Action: "branch.created", TargetType: "branch", TargetCode: req.Code,
			CompanyCode: req.CompanyUID, ActorUID: ctx.UserInfo().UID,
			After:      map[string]interface{}{"companyuid": req.CompanyUID, "branchcode": req.Code, "isactive": true},
			OccurredAt: now,
		})
	})
	if err != nil {
		if centraldb.IsUniqueViolation(err) {
			return apperr.Respond(ctx, apperr.DuplicateCode("code", req.Code).
				WithMessage("branch code is exists").WithThaiMessage("รหัสสาขานี้มีอยู่แล้วในบริษัทนี้").WithWrap(err))
		}
		responseBranchWriteError(ctx, err)
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
		ID:      req.BranchUID,
		Data:    orgaccess.OrganizationCreateResponse{Entity: req},
	})
	return nil
}

// SearchBranch lists the active branches the caller's access scopes allow (workspace
// selector). management=true lists the whole structure for Holding managers.
func (h BranchHttp) SearchBranch(ctx microservice.IContext) error {
	management := strings.EqualFold(strings.TrimSpace(ctx.QueryParam("management")), "true")
	holdingCode := strings.TrimSpace(ctx.UserInfo().HoldingCode)
	companyCode := strings.TrimSpace(ctx.QueryParam("companyguid"))

	reqCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	db, err := centraldb.Open()
	if err != nil {
		return apperr.Respond(ctx, apperr.ErrInternal.WithWrap(err))
	}

	query := `SELECT ` + branchColumns + ` FROM branches b WHERE b.holding_code = $1 AND ($2 = '' OR b.company_code = $2)`
	args := []interface{}{holdingCode, companyCode}
	if management {
		if _, err := orgpolicy.FindActiveHoldingManager(reqCtx, db, ctx.UserInfo(), time.Now()); err != nil {
			return respondBranchMembershipError(ctx, err)
		}
		if err := orgpolicy.RequireActiveHolding(reqCtx, db, holdingCode); err != nil {
			return respondBranchMembershipError(ctx, err)
		}
	} else {
		membership, err := orgpolicy.FindActiveMembership(reqCtx, db, ctx.UserInfo(), time.Now())
		if err != nil {
			return respondBranchMembershipError(ctx, err)
		}
		if companyCode != "" && !orgpolicy.AllowsCompany(membership.AccessScopes, companyCode) {
			return apperr.Respond(ctx, apperr.ErrForbidden.WithMessage("company access denied").WithThaiMessage("ไม่มีสิทธิ์ใช้งานบริษัทนี้"))
		}
		var predicate string
		predicate, args = scopedBranchPredicate(membership.AccessScopes, args)
		query += ` AND b.is_active = true AND ` + predicate
	}
	query += ` ORDER BY b.created_at, b.company_code, b.code`

	list, err := queryBranches(reqCtx, db, holdingCode, query, args...)
	if err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}
	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    list,
	})
	return nil
}

// SearchBranchStep is the paged, searchable list of the branches the caller may use.
func (h BranchHttp) SearchBranchStep(ctx microservice.IContext) error {
	holdingCode := strings.TrimSpace(ctx.UserInfo().HoldingCode)
	q := strings.TrimSpace(ctx.QueryParam("q"))
	offset := parseNonNegative(ctx.QueryParam("offset"), 0)
	limit := parseNonNegative(ctx.QueryParam("limit"), 100)
	if limit == 0 || limit > 500 {
		limit = 100
	}

	reqCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	db, err := centraldb.Open()
	if err != nil {
		return apperr.Respond(ctx, apperr.ErrInternal.WithWrap(err))
	}
	membership, err := orgpolicy.FindActiveMembership(reqCtx, db, ctx.UserInfo(), time.Now())
	if err != nil {
		return respondBranchMembershipError(ctx, err)
	}

	where := ` FROM branches b WHERE b.holding_code = $1 AND b.is_active = true AND ($2 = '' OR b.code ILIKE $2 OR b.name ILIKE $2)`
	args := []interface{}{holdingCode, likePattern(q)}
	predicate, args := scopedBranchPredicate(membership.AccessScopes, args)
	where += ` AND ` + predicate

	var total int64
	if err := db.QueryRowContext(reqCtx, `SELECT COUNT(*)`+where, args...).Scan(&total); err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}
	pageArgs := append(args, limit, offset)
	query := `SELECT ` + branchColumns + where +
		fmt.Sprintf(` ORDER BY b.created_at, b.company_code, b.code LIMIT $%d OFFSET $%d`, len(args)+1, len(args)+2)
	list, err := queryBranches(reqCtx, db, holdingCode, query, pageArgs...)
	if err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}
	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    list,
		Total:   total,
	})
	return nil
}

func (h BranchHttp) InfoBranch(ctx microservice.IContext) error {
	holdingCode := strings.TrimSpace(ctx.UserInfo().HoldingCode)
	id := ctx.Param("id")
	companyCode := firstNonBlank(ctx.QueryParam("companyuid"), ctx.QueryParam("companyguid"))

	reqCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	db, err := centraldb.Open()
	if err != nil {
		return apperr.Respond(ctx, apperr.ErrInternal.WithWrap(err))
	}
	membership, err := orgpolicy.FindActiveMembership(reqCtx, db, ctx.UserInfo(), time.Now())
	if err != nil {
		return respondBranchMembershipError(ctx, err)
	}

	data, err := findBranch(reqCtx, db, holdingCode, companyCode, id, false)
	if err != nil {
		return respondBranchLookupError(ctx, err)
	}
	if !orgpolicy.AllowsBranch(membership.AccessScopes, data.CompanyUID, data.BranchUID) {
		return apperr.Respond(ctx, apperr.ErrForbidden.WithMessage("branch access denied").WithThaiMessage("ไม่มีสิทธิ์ใช้งานสาขานี้"))
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    data,
	})
	return nil
}

func (h BranchHttp) UpdateBranch(ctx microservice.IContext) error {
	holdingCode := strings.TrimSpace(ctx.UserInfo().HoldingCode)
	authUsername := ctx.UserInfo().Username
	id := ctx.Param("id")
	input := ctx.ReadInput()

	var req branchModels.BranchOrgDoc
	if err := json.Unmarshal([]byte(input), &req); err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}
	requestedCompany := strings.TrimSpace(req.CompanyUID)
	if requestedCompany == "" {
		requestedCompany = strings.TrimSpace(req.CompanyGuid)
	}

	reqCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	db, err := centraldb.Open()
	if err != nil {
		return apperr.Respond(ctx, apperr.ErrInternal.WithWrap(err))
	}
	if _, err := orgpolicy.FindActiveHoldingManager(reqCtx, db, ctx.UserInfo(), time.Now()); err != nil {
		return respondBranchMembershipError(ctx, err)
	}
	if err := orgpolicy.RequireActiveHolding(reqCtx, db, holdingCode); err != nil {
		return respondBranchMembershipError(ctx, err)
	}

	var responseErr error
	respond := func(status int, err error) error {
		ctx.ResponseError(status, err.Error())
		responseErr = err
		return err
	}
	err = inTx(reqCtx, db, func(tx *sql.Tx) error {
		existing, err := findBranch(reqCtx, tx, holdingCode, requestedCompany, id, true)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) || errors.Is(err, errBranchAmbiguous) {
				responseErr = err
				return respondBranchLookupError(ctx, err)
			}
			return err
		}
		if err := preserveBranchParent(&req, existing); err != nil {
			return respond(http.StatusBadRequest, err)
		}
		if err := prepareBranchUpdate(&req); err != nil {
			return respond(http.StatusBadRequest, err)
		}
		if err := requireUnchangedBranchCode(existing, req.BranchCode); err != nil {
			return respond(http.StatusConflict, err)
		}
		if err := requireActiveBranchCompany(reqCtx, tx, holdingCode, existing.CompanyUID); err != nil {
			if errors.Is(err, branchModels.ErrBranchCompanyNotFound) {
				return respond(http.StatusBadRequest, err)
			}
			return err
		}
		requestedStatus, statusChanged, err := orgaccess.ResolveRequestedActiveStatus(input, existing.IsActive)
		if err != nil {
			return respond(http.StatusBadRequest, err)
		}
		reason := ""
		if statusChanged {
			if reason, err = orgaccess.StatusChangeReason(input); err != nil {
				return respond(http.StatusBadRequest, err)
			}
		}
		names, err := orgaccess.EncodeNames(req.Names)
		if err != nil {
			return respond(http.StatusBadRequest, err)
		}
		settings, err := json.Marshal(req.BranchSettings)
		if err != nil {
			return respond(http.StatusBadRequest, err)
		}
		now := time.Now().UTC()
		result, err := tx.ExecContext(reqCtx, `
			UPDATE branches SET name = $4, names = $5, settings = $6, is_headquarters = $7, is_active = $8, updated_by = $9, updated_at = $10
			WHERE holding_code = $1 AND company_code = $2 AND code = $3 AND is_active = $11`,
			holdingCode, existing.CompanyUID, existing.Code, orgaccess.PrimaryName(req.Names), names, string(settings),
			isHeadquarters(req), requestedStatus, authUsername, now, existing.IsActive)
		if err != nil {
			return err
		}
		if affected, err := result.RowsAffected(); err != nil || affected != 1 {
			return orgaccess.ErrStatusChangeConflict
		}
		if !statusChanged {
			return nil
		}
		return orgaccess.RecordAudit(reqCtx, tx, orgaccess.Audit{
			HoldingCode: holdingCode, Action: "organization.status.changed", TargetType: "branch",
			TargetCode: existing.Code, CompanyCode: existing.CompanyUID, ActorUID: ctx.UserInfo().UID, Reason: reason,
			Before: map[string]bool{"isactive": existing.IsActive}, After: map[string]bool{"isactive": requestedStatus},
			OccurredAt: now,
		})
	})
	if responseErr != nil {
		return responseErr
	}
	if err != nil {
		if errors.Is(err, orgaccess.ErrStatusChangeConflict) {
			ctx.ResponseError(http.StatusConflict, err.Error())
			return err
		}
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		ID:      id,
	})
	return nil
}

func queryBranches(ctx context.Context, db *sql.DB, holdingCode, query string, args ...interface{}) ([]branchModels.BranchOrgDoc, error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]branchModels.BranchOrgDoc, 0)
	for rows.Next() {
		doc, err := scanBranch(rows, holdingCode)
		if err != nil {
			return nil, err
		}
		list = append(list, doc)
	}
	return list, rows.Err()
}

// scopedBranchPredicate appends the caller's access scopes as parameters and returns
// the matching SQL condition on alias b: whole companies, or explicit (company, branch)
// pairs. No scope matches nothing (fail closed).
func scopedBranchPredicate(scopes []authModels.AccessScope, args []interface{}) (string, []interface{}) {
	allCompanies := []string{}
	pairCompanies := []string{}
	pairBranches := []string{}
	for _, companyUID := range orgpolicy.AllowedCompanyUIDs(scopes) {
		if orgpolicy.AllowsAllBranches(scopes, companyUID) {
			allCompanies = append(allCompanies, companyUID)
			continue
		}
		for _, branchUID := range orgpolicy.AllowedBranchUIDs(scopes, companyUID) {
			pairCompanies = append(pairCompanies, companyUID)
			pairBranches = append(pairBranches, branchUID)
		}
	}
	n := len(args)
	args = append(args, pq.Array(allCompanies), pq.Array(pairCompanies), pq.Array(pairBranches))
	predicate := fmt.Sprintf(`(b.company_code = ANY($%d::text[]) OR (b.company_code, b.code) IN (SELECT * FROM unnest($%d::text[], $%d::text[])))`, n+1, n+2, n+3)
	return predicate, args
}

func isHeadquarters(req branchModels.BranchOrgDoc) bool {
	return strings.EqualFold(strings.TrimSpace(req.BranchType), "head") || branchModels.IsThaiHeadOfficeBranchCode(req.Code)
}

func likePattern(q string) string {
	if q == "" {
		return ""
	}
	replacer := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return "%" + replacer.Replace(q) + "%"
}

func parseNonNegative(value string, fallback int) int {
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || parsed < 0 {
		return fallback
	}
	return parsed
}

func firstNonBlank(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}

func inTx(ctx context.Context, db *sql.DB, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}

func responseBranchWriteError(ctx microservice.IContext, err error) {
	switch {
	case errors.Is(err, branchModels.ErrBranchCompanyGuidRequired), errors.Is(err, branchModels.ErrBranchCompanyNotFound):
		ctx.ResponseError(http.StatusBadRequest, err.Error())
	default:
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
	}
}

func respondBranchLookupError(ctx microservice.IContext, err error) error {
	switch {
	case errors.Is(err, sql.ErrNoRows):
		ctx.ResponseError(http.StatusNotFound, "Branch not found")
		return err
	case errors.Is(err, errBranchAmbiguous):
		ctx.ResponseError(http.StatusConflict, err.Error())
		return err
	default:
		return apperr.Respond(ctx, apperr.ErrInternal.WithWrap(err))
	}
}

func respondBranchMembershipError(ctx microservice.IContext, err error) error {
	if expired := orgaccess.AccessExpiredError(err, orgaccess.RequestLanguage(ctx)); expired != nil {
		return apperr.Respond(ctx, expired)
	}
	if errors.Is(err, orgpolicy.ErrActiveMembershipRequired) || errors.Is(err, orgpolicy.ErrHoldingManagerRequired) {
		return apperr.Respond(ctx, apperr.ErrForbidden.WithMessage("active Holding membership is required").WithThaiMessage("ไม่มี Membership ที่ใช้งานได้ใน Holding นี้"))
	}
	if errors.Is(err, orgpolicy.ErrActiveHoldingRequired) {
		return apperr.Respond(ctx, apperr.ErrForbidden.WithMessage("active Holding is required").WithThaiMessage("Holding นี้ไม่ได้เปิดใช้งาน"))
	}
	return apperr.Respond(ctx, apperr.ErrInternal.WithWrap(err))
}

func prepareBranchCreate(req *branchModels.BranchOrgDoc, holdingCode, authUsername string, now time.Time) error {
	if err := prepareBranchUpdate(req); err != nil {
		return err
	}
	holdingCode = strings.TrimSpace(holdingCode)
	if holdingCode == "" {
		return errors.New("stable Branch identity is required")
	}
	now = now.UTC()
	setBranchIdentity(req, holdingCode, req.CompanyUID, req.Code)
	normalizeBranchSlices(&req.BranchSettings)
	req.CreatedAt = now
	req.UpdatedAt = now
	req.IsActive = true
	req.CreatedBy = strings.TrimSpace(authUsername)
	req.UpdatedBy = ""
	return nil
}

func prepareBranchUpdate(req *branchModels.BranchOrgDoc) error {
	companyUID := strings.TrimSpace(req.CompanyUID)
	companyGuid := strings.TrimSpace(req.CompanyGuid)
	if companyUID != "" && companyGuid != "" && companyUID != companyGuid {
		return errBranchParentChange
	}
	if companyUID == "" {
		companyUID = companyGuid
	}
	if companyUID == "" {
		return branchModels.ErrBranchCompanyGuidRequired
	}
	req.CompanyUID = companyUID
	req.CompanyGuid = companyUID

	inputCode := strings.TrimSpace(req.Code)
	inputBranchCode := strings.TrimSpace(req.BranchCode)
	if inputCode != "" && inputBranchCode != "" && inputCode != inputBranchCode {
		return errors.New("branch code aliases do not match")
	}
	if inputCode == "" {
		inputCode = inputBranchCode
	}
	normalizedCode, err := branchModels.NormalizeThaiTaxBranchCode(inputCode)
	if err != nil {
		return err
	}
	req.Code = normalizedCode
	req.BranchCode = normalizedCode
	normalizeBranchSlices(&req.BranchSettings)
	normalizedTimezone, err := branchModels.NormalizeIANATimezone(req.Timezone)
	if err != nil {
		return err
	}
	req.Timezone = normalizedTimezone
	req.Addresses = sanitizeBranchAddresses(req.Addresses)
	req.CountryCode = strings.TrimSpace(req.CountryCode)
	req.ProvinceCode = strings.TrimSpace(req.ProvinceCode)
	req.DistrictCode = strings.TrimSpace(req.DistrictCode)
	req.SubDistrictCode = strings.TrimSpace(req.SubDistrictCode)
	req.ZipCode = strings.TrimSpace(req.ZipCode)
	if err := normalizeAndValidateDocFormats(req); err != nil {
		return err
	}
	return validateBranchNames(req.Names)
}

func preserveBranchParent(req *branchModels.BranchOrgDoc, existing branchModels.BranchOrgDoc) error {
	existingHoldingUID := strings.TrimSpace(existing.HoldingUID)
	existingUID := strings.TrimSpace(existing.CompanyUID)
	existingBranchUID := strings.TrimSpace(existing.BranchUID)
	if existingHoldingUID == "" || existingUID == "" || existingBranchUID == "" {
		return errors.New("stable Branch parent identity is required")
	}
	for _, identity := range []struct{ requested, existing string }{
		{req.HoldingUID, existingHoldingUID},
		{req.HoldingCode, strings.TrimSpace(existing.HoldingCode)},
		{req.BranchUID, existingBranchUID},
		{req.GuidFixed, existingBranchUID},
	} {
		if value := strings.TrimSpace(identity.requested); value != "" && value != identity.existing {
			return errBranchParentChange
		}
	}
	requestedUID := strings.TrimSpace(req.CompanyUID)
	requestedGuid := strings.TrimSpace(req.CompanyGuid)
	if requestedUID != "" && requestedGuid != "" && requestedUID != requestedGuid {
		return errBranchParentChange
	}
	if requestedUID == "" {
		requestedUID = requestedGuid
	}
	if requestedUID != "" && requestedUID != existingUID {
		return errBranchParentChange
	}
	req.CompanyUID = existingUID
	req.CompanyGuid = existingUID
	req.HoldingUID = existingHoldingUID
	req.HoldingCode = strings.TrimSpace(existing.HoldingCode)
	req.BranchUID = existingBranchUID
	req.GuidFixed = existingBranchUID
	return nil
}

func validateBranchNames(names common.JSONB) error {
	for _, name := range names {
		if name.Code != nil && name.Name != nil && strings.TrimSpace(*name.Code) != "" && strings.TrimSpace(*name.Name) != "" {
			return nil
		}
	}
	return errors.New("branch name is required")
}

func storedBranchCode(branch branchModels.BranchOrgDoc) string {
	value := strings.TrimSpace(branch.BranchCode)
	if value == "" {
		value = strings.TrimSpace(branch.Code)
	}
	normalized, err := branchModels.NormalizeThaiTaxBranchCode(value)
	if err != nil {
		return value
	}
	return normalized
}

func requireUnchangedBranchCode(existing branchModels.BranchOrgDoc, requestedCode string) error {
	normalized, err := branchModels.NormalizeThaiTaxBranchCode(requestedCode)
	if err != nil {
		return err
	}
	if storedBranchCode(existing) != normalized {
		return errBranchCodeChange
	}
	return nil
}

var docFormatPrefixAllowed = regexp.MustCompile(`[^A-Z0-9]+`)

// normalizeAndValidateDocFormats uppercases/trims each document-number format's
// DocType and Prefix (Prefix keeps only A-Z0-9), then validates that every
// entry has a non-empty DocType and Prefix and that no Prefix repeats anywhere
// in the branch (across all doctypes/formats). Validation runs on every entry
// sent, regardless of Enabled.
func normalizeAndValidateDocFormats(req *branchModels.BranchOrgDoc) error {
	seen := make(map[string]struct{}, len(req.DocumentFormats))
	for i := range req.DocumentFormats {
		f := &req.DocumentFormats[i]
		f.DocType = strings.ToUpper(strings.TrimSpace(f.DocType))
		f.Prefix = strings.TrimSpace(docFormatPrefixAllowed.ReplaceAllString(strings.ToUpper(f.Prefix), ""))
		if f.DocType == "" {
			return fmt.Errorf("ประเภทเอกสาร (doctype) ของรูปแบบเลขที่เอกสารห้ามว่าง")
		}
		if f.Prefix == "" {
			return fmt.Errorf("คำนำหน้าเลขที่เอกสารห้ามว่าง (doctype: %s)", f.DocType)
		}
		if _, dup := seen[f.Prefix]; dup {
			return fmt.Errorf("คำนำหน้าเลขที่เอกสารซ้ำ: %s", f.Prefix)
		}
		seen[f.Prefix] = struct{}{}
	}
	return nil
}

// sanitizeBranchAddresses trims each per-language address, caps its length, and
// drops entries with an empty code or address. Branch handlers don't run
// ctx.Validate, so this is the defense-in-depth guard against oversized/garbage
// direct-API payloads. Length is capped by rune count to avoid splitting a
// multibyte (Thai) character mid-address.
func sanitizeBranchAddresses(in []branchModels.BranchAddress) []branchModels.BranchAddress {
	const maxAddressRunes = 1000
	out := make([]branchModels.BranchAddress, 0, len(in))
	for _, a := range in {
		code := strings.TrimSpace(a.Code)
		addr := strings.TrimSpace(a.Address)
		if code == "" || addr == "" {
			continue
		}
		if r := []rune(addr); len(r) > maxAddressRunes {
			addr = string(r[:maxAddressRunes])
		}
		out = append(out, branchModels.BranchAddress{Code: code, Address: addr})
	}
	return out
}
