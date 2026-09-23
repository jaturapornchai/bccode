package company

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"smlcloudplatform/internal/centraldb"
	"smlcloudplatform/internal/config"
	common "smlcloudplatform/internal/models"
	orgaccess "smlcloudplatform/internal/organization"
	orgpolicy "smlcloudplatform/internal/organization/access"
	companyModels "smlcloudplatform/internal/organization/company/models"
	"smlcloudplatform/pkg/apperr"
	"smlcloudplatform/pkg/microservice"
)

type CompanyHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
}

func NewCompanyHttp(ms *microservice.Microservice, cfg config.IConfig) CompanyHttp {
	return CompanyHttp{
		ms:  ms,
		cfg: cfg,
	}
}

func (h CompanyHttp) RegisterHttp() {
	h.ms.POST("/organization/company", h.CreateCompany)
	h.ms.GET("/organization/company", h.SearchCompany)
	h.ms.GET("/organization/company/:id", h.InfoCompany)
	h.ms.PUT("/organization/company/:id", h.UpdateCompany)
}

const companyColumns = `code, name, names, COALESCE(tax_id, ''), logo_uri, is_active, created_at, updated_at, created_by, updated_by`

func scanCompany(row interface{ Scan(...interface{}) error }, holdingCode string) (companyModels.CompanyDoc, error) {
	var (
		doc   companyModels.CompanyDoc
		name  string
		names []byte
	)
	err := row.Scan(&doc.Code, &name, &names, &doc.TaxID, &doc.LogoURI, &doc.IsActive, &doc.CreatedAt, &doc.UpdatedAt, &doc.CreatedBy, &doc.UpdatedBy)
	if err != nil {
		return doc, err
	}
	doc.Names = orgaccess.DecodeNames(names, name)
	doc.HoldingCode = holdingCode
	doc.HoldingUID = holdingCode
	doc.GuidFixed = doc.Code
	doc.CompanyUID = doc.Code
	doc.ID = doc.Code
	return doc, nil
}

func (h CompanyHttp) CreateCompany(ctx microservice.IContext) error {
	holdingCode := ctx.UserInfo().HoldingCode
	input := ctx.ReadInput()

	var req companyModels.CompanyDoc
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
		return respondCompanyMembershipError(ctx, err)
	}
	holding, err := orgpolicy.FindActiveHolding(reqCtx, db, holdingCode)
	if err != nil {
		return respondCompanyMembershipError(ctx, err)
	}

	actorUID := strings.TrimSpace(ctx.UserInfo().UID)
	now := time.Now().UTC()
	if err := prepareCompanyCreate(&req, holding.Code, ctx.UserInfo().Username, now); err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}
	names, err := orgaccess.EncodeNames(req.Names)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	err = inTx(reqCtx, db, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(reqCtx, `
			INSERT INTO companies (holding_code, code, name, names, tax_id, logo_uri, is_active, created_by, updated_by, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, true, $7, '', $8, $8)`,
			req.HoldingCode, req.Code, orgaccess.PrimaryName(req.Names), names, req.TaxID, req.LogoURI, req.CreatedBy, now); err != nil {
			return err
		}
		return orgaccess.RecordAudit(reqCtx, tx, orgaccess.Audit{
			HoldingCode: req.HoldingCode, Action: "company.created", TargetType: "company", TargetCode: req.Code,
			CompanyCode: req.Code, ActorUID: actorUID,
			After:      map[string]interface{}{"code": req.Code, "isactive": true},
			OccurredAt: now,
		})
	})
	if err != nil {
		if centraldb.IsUniqueViolation(err) {
			return apperr.Respond(ctx, apperr.DuplicateCode("code", req.Code).
				WithMessage("company code is exists").WithThaiMessage("รหัสบริษัทนี้มีอยู่แล้วใน Holding").WithWrap(err))
		}
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}
	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
		ID:      req.CompanyUID,
		Data:    orgaccess.OrganizationCreateResponse{Entity: req},
	})
	return nil
}

// SearchCompany lists the active companies the caller's access scopes allow
// (workspace selector). management=true lists the whole structure for Holding managers.
func (h CompanyHttp) SearchCompany(ctx microservice.IContext) error {
	management := strings.EqualFold(strings.TrimSpace(ctx.QueryParam("management")), "true")
	holdingCode := strings.TrimSpace(ctx.UserInfo().HoldingCode)

	reqCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	db, err := centraldb.Open()
	if err != nil {
		return apperr.Respond(ctx, apperr.ErrInternal.WithWrap(err))
	}

	query := `SELECT ` + companyColumns + ` FROM companies WHERE holding_code = $1`
	var allowed map[string]bool
	if management {
		if _, err := orgpolicy.FindActiveHoldingManager(reqCtx, db, ctx.UserInfo(), time.Now()); err != nil {
			return respondCompanyMembershipError(ctx, err)
		}
		if err := orgpolicy.RequireActiveHolding(reqCtx, db, holdingCode); err != nil {
			return respondCompanyMembershipError(ctx, err)
		}
	} else {
		membership, err := orgpolicy.FindActiveMembership(reqCtx, db, ctx.UserInfo(), time.Now())
		if err != nil {
			return respondCompanyMembershipError(ctx, err)
		}
		allowed = map[string]bool{}
		for _, uid := range orgpolicy.AllowedCompanyUIDs(membership.AccessScopes) {
			allowed[uid] = true
		}
		query += ` AND is_active = true`
	}
	query += ` ORDER BY created_at, code`

	rows, err := db.QueryContext(reqCtx, query, holdingCode)
	if err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}
	defer rows.Close()

	list := make([]companyModels.CompanyDoc, 0)
	for rows.Next() {
		doc, err := scanCompany(rows, holdingCode)
		if err != nil {
			ctx.ResponseError(http.StatusInternalServerError, err.Error())
			return err
		}
		if allowed != nil && !allowed[doc.CompanyUID] {
			continue
		}
		list = append(list, doc)
	}
	if err := rows.Err(); err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    list,
	})
	return nil
}

func findCompany(ctx context.Context, q interface {
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}, holdingCode, code string, lock bool) (companyModels.CompanyDoc, error) {
	query := `SELECT ` + companyColumns + ` FROM companies WHERE holding_code = $1 AND code = $2`
	if lock {
		query += ` FOR UPDATE`
	}
	return scanCompany(q.QueryRowContext(ctx, query, holdingCode, strings.TrimSpace(code)), holdingCode)
}

func (h CompanyHttp) InfoCompany(ctx microservice.IContext) error {
	holdingCode := strings.TrimSpace(ctx.UserInfo().HoldingCode)
	id := ctx.Param("id")

	reqCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	db, err := centraldb.Open()
	if err != nil {
		return apperr.Respond(ctx, apperr.ErrInternal.WithWrap(err))
	}
	membership, err := orgpolicy.FindActiveMembership(reqCtx, db, ctx.UserInfo(), time.Now())
	if err != nil {
		return respondCompanyMembershipError(ctx, err)
	}
	if len(orgpolicy.AllowedCompanyUIDs(membership.AccessScopes)) == 0 {
		return apperr.Respond(ctx, apperr.ErrForbidden.WithMessage("company access denied").WithThaiMessage("ไม่มีสิทธิ์ใช้งานบริษัทนี้"))
	}

	data, err := findCompany(reqCtx, db, holdingCode, id, false)
	if errors.Is(err, sql.ErrNoRows) {
		ctx.ResponseError(http.StatusNotFound, "Company not found")
		return err
	}
	if err != nil {
		return apperr.Respond(ctx, apperr.ErrInternal.WithWrap(err))
	}
	if !orgpolicy.AllowsCompany(membership.AccessScopes, data.CompanyUID) {
		return apperr.Respond(ctx, apperr.ErrForbidden.WithMessage("company access denied").WithThaiMessage("ไม่มีสิทธิ์ใช้งานบริษัทนี้"))
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    data,
	})
	return nil
}

func respondCompanyMembershipError(ctx microservice.IContext, err error) error {
	if errors.Is(err, orgpolicy.ErrActiveMembershipRequired) || errors.Is(err, orgpolicy.ErrHoldingManagerRequired) {
		return apperr.Respond(ctx, apperr.ErrForbidden.WithMessage("active Holding membership is required").WithThaiMessage("ไม่มี Membership ที่ใช้งานได้ใน Holding นี้"))
	}
	if errors.Is(err, orgpolicy.ErrActiveHoldingRequired) {
		return apperr.Respond(ctx, apperr.ErrForbidden.WithMessage("active Holding is required").WithThaiMessage("Holding นี้ไม่ได้เปิดใช้งาน"))
	}
	return apperr.Respond(ctx, apperr.ErrInternal.WithWrap(err))
}

func (h CompanyHttp) UpdateCompany(ctx microservice.IContext) error {
	holdingCode := strings.TrimSpace(ctx.UserInfo().HoldingCode)
	authUsername := ctx.UserInfo().Username
	id := ctx.Param("id")
	input := ctx.ReadInput()

	var req companyModels.CompanyDoc
	if err := json.Unmarshal([]byte(input), &req); err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}
	req.Code = companyModels.NormalizeCompanyCode(req.Code)
	if req.Code == "" {
		ctx.ResponseError(http.StatusBadRequest, "company code is required")
		return errors.New("company code is required")
	}
	if err := validateCompanyNames(req.Names); err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	reqCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	db, err := centraldb.Open()
	if err != nil {
		return apperr.Respond(ctx, apperr.ErrInternal.WithWrap(err))
	}
	if _, err := orgpolicy.FindActiveHoldingManager(reqCtx, db, ctx.UserInfo(), time.Now()); err != nil {
		return respondCompanyMembershipError(ctx, err)
	}
	if err := orgpolicy.RequireActiveHolding(reqCtx, db, holdingCode); err != nil {
		return respondCompanyMembershipError(ctx, err)
	}

	var responseErr error
	err = inTx(reqCtx, db, func(tx *sql.Tx) error {
		existing, err := findCompany(reqCtx, tx, holdingCode, id, true)
		if errors.Is(err, sql.ErrNoRows) {
			ctx.ResponseError(http.StatusNotFound, "Company not found")
			responseErr = err
			return err
		}
		if err != nil {
			return err
		}
		if err := requireUnchangedCompanyCode(existing.Code, req.Code); err != nil {
			ctx.ResponseError(http.StatusConflict, err.Error())
			responseErr = err
			return err
		}
		requestedStatus, statusChanged, err := orgaccess.ResolveRequestedActiveStatus(input, existing.IsActive)
		if err != nil {
			ctx.ResponseError(http.StatusBadRequest, err.Error())
			responseErr = err
			return err
		}
		reason := ""
		if statusChanged {
			if reason, err = orgaccess.StatusChangeReason(input); err != nil {
				ctx.ResponseError(http.StatusBadRequest, err.Error())
				responseErr = err
				return err
			}
		}
		names, err := orgaccess.EncodeNames(req.Names)
		if err != nil {
			return err
		}
		now := time.Now().UTC()
		result, err := tx.ExecContext(reqCtx, `
			UPDATE companies SET name = $3, names = $4, tax_id = $5, logo_uri = $6, is_active = $7, updated_by = $8, updated_at = $9
			WHERE holding_code = $1 AND code = $2 AND is_active = $10`,
			holdingCode, existing.Code, orgaccess.PrimaryName(req.Names), names, req.TaxID, req.LogoURI,
			requestedStatus, authUsername, now, existing.IsActive)
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
			HoldingCode: holdingCode, Action: "organization.status.changed", TargetType: "company",
			TargetCode: existing.Code, CompanyCode: existing.Code, ActorUID: ctx.UserInfo().UID, Reason: reason,
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

func prepareCompanyCreate(req *companyModels.CompanyDoc, holdingCode, actor string, now time.Time) error {
	req.Code = companyModels.NormalizeCompanyCode(req.Code)
	if req.Code == "" {
		return errors.New("company code is required")
	}
	if err := validateCompanyNames(req.Names); err != nil {
		return err
	}
	holdingCode = strings.TrimSpace(holdingCode)
	if holdingCode == "" {
		return errors.New("stable Company parent identity is required")
	}
	now = now.UTC()
	req.ID = req.Code
	req.HoldingCode = holdingCode
	req.HoldingUID = holdingCode
	req.GuidFixed = req.Code
	req.CompanyUID = req.Code
	req.IsDeleted = false
	req.Version = 0
	req.IsActive = true
	req.CreatedAt = now
	req.UpdatedAt = now
	req.DeletedAt = nil
	req.CreatedBy = strings.TrimSpace(actor)
	req.UpdatedBy = ""
	req.DeletedBy = ""
	return nil
}

func validateCompanyNames(names common.JSONB) error {
	for _, name := range names {
		if name.Code != nil && name.Name != nil && strings.TrimSpace(*name.Code) != "" && strings.TrimSpace(*name.Name) != "" {
			return nil
		}
	}
	return errors.New("company name is required")
}

var errCompanyCodeChange = errors.New("company code cannot be changed by ordinary update")

func requireUnchangedCompanyCode(existingCode, requestedCode string) error {
	if companyModels.NormalizeCompanyCode(existingCode) != companyModels.NormalizeCompanyCode(requestedCode) {
		return errCompanyCodeChange
	}
	return nil
}
