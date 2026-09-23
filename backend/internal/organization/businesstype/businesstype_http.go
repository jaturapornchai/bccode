package businesstype

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"smlcloudplatform/internal/centraldb"
	"smlcloudplatform/internal/config"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/organization/businesstype/models"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/apperr"
	"smlcloudplatform/pkg/microservice"
)

type IBusinessTypeHttp interface{}

type BusinessTypeHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
}

func NewBusinessTypeHttp(ms *microservice.Microservice, cfg config.IConfig) BusinessTypeHttp {
	return BusinessTypeHttp{
		ms:  ms,
		cfg: cfg,
	}
}

func (h BusinessTypeHttp) RegisterHttp() {
	h.ms.GET("/organization/business-type", h.SearchBusinessTypePage)
	h.ms.GET("/organization/business-type/list", h.SearchBusinessTypeStep)
	h.ms.POST("/organization/business-type", h.CreateBusinessType)
	h.ms.GET("/organization/business-type/:id", h.InfoBusinessType)
	h.ms.GET("/organization/business-type/default", h.InfoBusinessTypeDefault)
	h.ms.GET("/organization/business-type/code/:code", h.InfoBusinessTypeByCode)
	h.ms.PUT("/organization/business-type/:id", h.UpdateBusinessType)
	h.ms.DELETE("/organization/business-type/:id", h.DeleteBusinessType)
	h.ms.DELETE("/organization/business-type", h.DeleteBusinessTypeByGUIDs)
}

func (h BusinessTypeHttp) searchBusinessTypeStepPostgres(ctx microservice.IContext, db *sql.DB, holdingCode string, offset int, limit int, q string) error {
	holdingCode = strings.TrimSpace(holdingCode)
	if holdingCode == "" {
		ctx.Response(http.StatusOK, common.ApiResponse{Success: true, Data: []models.BusinessTypeInfo{}, Total: 0})
		return nil
	}
	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	q = strings.TrimSpace(q)
	countQuery := `SELECT COUNT(*) FROM business_types WHERE LOWER(holding_code) = LOWER($1) AND is_active = true`
	dataQuery := `SELECT id, code, names, is_default FROM business_types WHERE LOWER(holding_code) = LOWER($1) AND is_active = true`
	args := []interface{}{holdingCode}

	if q != "" {
		countQuery += ` AND (code ILIKE $2 OR names::text ILIKE $2)`
		dataQuery += ` AND (code ILIKE $2 OR names::text ILIKE $2)`
		args = append(args, "%"+q+"%")
	}

	qCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var total int64
	if err := db.QueryRowContext(qCtx, countQuery, args...).Scan(&total); err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	dataQuery += fmt.Sprintf(` ORDER BY code LIMIT %d OFFSET %d`, limit, offset)
	rows, err := db.QueryContext(qCtx, dataQuery, args...)
	if err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}
	defer rows.Close()

	list := make([]models.BusinessTypeInfo, 0)
	for rows.Next() {
		var (
			id        string
			code      string
			rawNames  []byte
			isDefault bool
		)
		if err := rows.Scan(&id, &code, &rawNames, &isDefault); err != nil {
			continue
		}
		var namesList []common.NameX
		if len(rawNames) > 0 {
			_ = json.Unmarshal(rawNames, &namesList)
		}
		list = append(list, models.BusinessTypeInfo{
			DocIdentity: common.DocIdentity{
				GuidFixed: id,
			},
			BusinessType: models.BusinessType{
				Code:      code,
				Names:     &namesList,
				IsDefault: isDefault,
			},
		})
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    list,
		Total:   total,
	})
	return nil
}

func (h BusinessTypeHttp) infoBusinessTypePostgres(ctx microservice.IContext, db *sql.DB, holdingCode string, id string) error {
	holdingCode = strings.TrimSpace(holdingCode)
	id = strings.TrimSpace(id)

	query := `SELECT id, code, names, is_default FROM business_types WHERE LOWER(holding_code) = LOWER($1) AND (id = $2 OR LOWER(code) = LOWER($2)) AND is_active = true LIMIT 1`
	var (
		guidFixed string
		code      string
		rawNames  []byte
		isDefault bool
	)
	err := db.QueryRowContext(context.Background(), query, holdingCode, id).Scan(&guidFixed, &code, &rawNames, &isDefault)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.ResponseError(http.StatusNotFound, "ไม่พบประเภทธุรกิจ")
			return err
		}
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	var namesList []common.NameX
	if len(rawNames) > 0 {
		_ = json.Unmarshal(rawNames, &namesList)
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data: models.BusinessTypeInfo{
			DocIdentity: common.DocIdentity{
				GuidFixed: guidFixed,
			},
			BusinessType: models.BusinessType{
				Code:      code,
				Names:     &namesList,
				IsDefault: isDefault,
			},
		},
	})
	return nil
}

func (h BusinessTypeHttp) createBusinessTypePostgres(ctx microservice.IContext, db *sql.DB, holdingCode string, docReq models.BusinessType) error {
	holdingCode = strings.TrimSpace(holdingCode)
	code := strings.TrimSpace(docReq.Code)
	if code == "" {
		ctx.ResponseError(http.StatusBadRequest, "code is required")
		return nil
	}

	rawNames, err := json.Marshal(docReq.Names)
	if err != nil {
		rawNames = []byte("[]")
	}

	newID := uuid.New().String()
	query := `INSERT INTO business_types (id, holding_code, code, names, is_default, is_active)
	          VALUES ($1, $2, $3, $4, $5, true)
	          ON CONFLICT (holding_code, code) DO UPDATE SET
	            names = EXCLUDED.names,
	            is_default = EXCLUDED.is_default,
	            is_active = true`
	_, err = db.ExecContext(context.Background(), query, newID, holdingCode, code, rawNames, docReq.IsDefault)
	if err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
		ID:      newID,
	})
	return nil
}

func (h BusinessTypeHttp) updateBusinessTypePostgres(ctx microservice.IContext, db *sql.DB, holdingCode string, id string, docReq models.BusinessType) error {
	holdingCode = strings.TrimSpace(holdingCode)
	id = strings.TrimSpace(id)

	rawNames, err := json.Marshal(docReq.Names)
	if err != nil {
		rawNames = []byte("[]")
	}

	query := `UPDATE business_types SET names = $1, is_default = $2, updated_at = now()
	          WHERE LOWER(holding_code) = LOWER($3) AND (id = $4 OR LOWER(code) = LOWER($4))`
	_, err = db.ExecContext(context.Background(), query, rawNames, docReq.IsDefault, holdingCode, id)
	if err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		ID:      id,
	})
	return nil
}

func (h BusinessTypeHttp) deleteBusinessTypePostgres(ctx microservice.IContext, db *sql.DB, holdingCode string, id string) error {
	holdingCode = strings.TrimSpace(holdingCode)
	id = strings.TrimSpace(id)

	query := `UPDATE business_types SET is_active = false, updated_at = now()
	          WHERE LOWER(holding_code) = LOWER($1) AND (id = $2 OR LOWER(code) = LOWER($2))`
	_, err := db.ExecContext(context.Background(), query, holdingCode, id)
	if err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})
	return nil
}

func (h BusinessTypeHttp) deleteBusinessTypeByGUIDsPostgres(ctx microservice.IContext, db *sql.DB, holdingCode string, ids []string) error {
	holdingCode = strings.TrimSpace(holdingCode)
	if len(ids) == 0 {
		ctx.Response(http.StatusOK, common.ApiResponse{Success: true})
		return nil
	}
	query := `UPDATE business_types SET is_active = false, updated_at = now()
	          WHERE LOWER(holding_code) = LOWER($1) AND (id = ANY($2) OR code = ANY($2))`
	_, err := db.ExecContext(context.Background(), query, holdingCode, pq.Array(ids))
	if err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}
	ctx.Response(http.StatusOK, common.ApiResponse{Success: true})
	return nil
}

func (h BusinessTypeHttp) infoBusinessTypeDefaultPostgres(ctx microservice.IContext, db *sql.DB, holdingCode string) error {
	holdingCode = strings.TrimSpace(holdingCode)

	query := `SELECT id, code, names, is_default FROM business_types WHERE LOWER(holding_code) = LOWER($1) AND is_active = true ORDER BY is_default DESC, code ASC LIMIT 1`
	var (
		guidFixed string
		code      string
		rawNames  []byte
		isDefault bool
	)
	err := db.QueryRowContext(context.Background(), query, holdingCode).Scan(&guidFixed, &code, &rawNames, &isDefault)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.ResponseError(http.StatusNotFound, "ไม่พบประเภทธุรกิจ")
			return err
		}
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	var namesList []common.NameX
	if len(rawNames) > 0 {
		_ = json.Unmarshal(rawNames, &namesList)
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data: models.BusinessTypeInfo{
			DocIdentity: common.DocIdentity{
				GuidFixed: guidFixed,
			},
			BusinessType: models.BusinessType{
				Code:      code,
				Names:     &namesList,
				IsDefault: isDefault,
			},
		},
	})
	return nil
}

// withCentralDB runs a handler against the central database (business types are
// Holding-level master data).
func withCentralDB(ctx microservice.IContext, fn func(db *sql.DB) error) error {
	db, err := centraldb.Open()
	if err != nil {
		return apperr.Respond(ctx, apperr.ErrInternal.WithWrap(err))
	}
	return fn(db)
}

func readBusinessType(ctx microservice.IContext) (models.BusinessType, error) {
	docReq := models.BusinessType{}
	if err := json.Unmarshal([]byte(ctx.ReadInput()), &docReq); err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return docReq, err
	}
	if err := ctx.Validate(&docReq); err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return docReq, err
	}
	return docReq, nil
}

// CreateBusinessType godoc
// @Router /organization/business-type [post]
func (h BusinessTypeHttp) CreateBusinessType(ctx microservice.IContext) error {
	docReq, err := readBusinessType(ctx)
	if err != nil {
		return err
	}
	return withCentralDB(ctx, func(db *sql.DB) error {
		return h.createBusinessTypePostgres(ctx, db, ctx.UserInfo().HoldingCode, docReq)
	})
}

// UpdateBusinessType godoc
// @Router /organization/business-type/{id} [put]
func (h BusinessTypeHttp) UpdateBusinessType(ctx microservice.IContext) error {
	docReq, err := readBusinessType(ctx)
	if err != nil {
		return err
	}
	return withCentralDB(ctx, func(db *sql.DB) error {
		return h.updateBusinessTypePostgres(ctx, db, ctx.UserInfo().HoldingCode, ctx.Param("id"), docReq)
	})
}

// DeleteBusinessType godoc
// @Router /organization/business-type/{id} [delete]
func (h BusinessTypeHttp) DeleteBusinessType(ctx microservice.IContext) error {
	return withCentralDB(ctx, func(db *sql.DB) error {
		return h.deleteBusinessTypePostgres(ctx, db, ctx.UserInfo().HoldingCode, ctx.Param("id"))
	})
}

// DeleteBusinessTypeByGUIDs godoc
// @Router /organization/business-type [delete]
func (h BusinessTypeHttp) DeleteBusinessTypeByGUIDs(ctx microservice.IContext) error {
	docReq := []string{}
	if err := json.Unmarshal([]byte(ctx.ReadInput()), &docReq); err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}
	return withCentralDB(ctx, func(db *sql.DB) error {
		return h.deleteBusinessTypeByGUIDsPostgres(ctx, db, ctx.UserInfo().HoldingCode, docReq)
	})
}

// InfoBusinessType godoc
// @Router /organization/business-type/{id} [get]
func (h BusinessTypeHttp) InfoBusinessType(ctx microservice.IContext) error {
	return withCentralDB(ctx, func(db *sql.DB) error {
		return h.infoBusinessTypePostgres(ctx, db, ctx.UserInfo().HoldingCode, ctx.Param("id"))
	})
}

// InfoBusinessTypeDefault godoc
// @Router /organization/business-type/default [get]
func (h BusinessTypeHttp) InfoBusinessTypeDefault(ctx microservice.IContext) error {
	return withCentralDB(ctx, func(db *sql.DB) error {
		return h.infoBusinessTypeDefaultPostgres(ctx, db, ctx.UserInfo().HoldingCode)
	})
}

// InfoBusinessTypeByCode godoc
// @Router /organization/business-type/code/{code} [get]
func (h BusinessTypeHttp) InfoBusinessTypeByCode(ctx microservice.IContext) error {
	return withCentralDB(ctx, func(db *sql.DB) error {
		return h.infoBusinessTypePostgres(ctx, db, ctx.UserInfo().HoldingCode, ctx.Param("code"))
	})
}

// SearchBusinessTypePage godoc
// @Router /organization/business-type [get]
func (h BusinessTypeHttp) SearchBusinessTypePage(ctx microservice.IContext) error {
	pageable := utils.GetPageable(ctx.QueryParam)
	limit := 100
	if pageable.Limit > 0 {
		limit = pageable.Limit
	}
	q := ctx.QueryParam("q")
	if pageable.Query != "" {
		q = pageable.Query
	}
	return withCentralDB(ctx, func(db *sql.DB) error {
		return h.searchBusinessTypeStepPostgres(ctx, db, ctx.UserInfo().HoldingCode, pageable.GetOffest(), limit, q)
	})
}

// SearchBusinessTypeStep godoc
// @Router /organization/business-type/list [get]
func (h BusinessTypeHttp) SearchBusinessTypeStep(ctx microservice.IContext) error {
	pageableStep := utils.GetPageableStep(ctx.QueryParam)
	limit := 100
	if pageableStep.Limit > 0 {
		limit = pageableStep.Limit
	}
	q := ctx.QueryParam("q")
	if pageableStep.Query != "" {
		q = pageableStep.Query
	}
	return withCentralDB(ctx, func(db *sql.DB) error {
		return h.searchBusinessTypeStepPostgres(ctx, db, ctx.UserInfo().HoldingCode, pageableStep.Skip, limit, q)
	})
}
