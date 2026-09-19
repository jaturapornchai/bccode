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
	"smlcloudplatform/internal/config"
	mypg "smlcloudplatform/internal/goapi/mypg"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/organization/businesstype/models"
	"smlcloudplatform/internal/organization/businesstype/repositories"
	"smlcloudplatform/internal/organization/businesstype/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
)

type IBusinessTypeHttp interface{}

type BusinessTypeHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IBusinessTypeHttpService
}

func NewBusinessTypeHttp(ms *microservice.Microservice, cfg config.IConfig) BusinessTypeHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())

	repo := repositories.NewBusinessTypeRepository(pst)

	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewBusinessTypeHttpService(repo, masterSyncCacheRepo)

	return BusinessTypeHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h BusinessTypeHttp) RegisterHttp() {

	h.ms.POST("/organization/business-type/bulk", h.SaveBulk)

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

// Create BusinessType godoc
// @Description Create BusinessType
// @Tags		BusinessType
// @Param		BusinessType  body      models.BusinessType  true  "BusinessType"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /organization/business-type [post]
func (h BusinessTypeHttp) CreateBusinessType(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	holdingCode := ctx.UserInfo().HoldingCode
	input := ctx.ReadInput()

	docReq := &models.BusinessType{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if db, err := mypg.PgSqlFastConnect("bcai_projection"); err == nil && db != nil {
		return h.createBusinessTypePostgres(ctx, db, holdingCode, *docReq)
	}

	idx, err := h.svc.CreateBusinessType(holdingCode, authUsername, *docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
		ID:      idx,
	})
	return nil
}

// Update BusinessType godoc
// @Description Update BusinessType
// @Tags		BusinessType
// @Param		id  path      string  true  "BusinessType ID"
// @Param		BusinessType  body      models.BusinessType  true  "BusinessType"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /organization/business-type/{id} [put]
func (h BusinessTypeHttp) UpdateBusinessType(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.BusinessType{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if db, err := mypg.PgSqlFastConnect("bcai_projection"); err == nil && db != nil {
		return h.updateBusinessTypePostgres(ctx, db, holdingCode, id, *docReq)
	}

	err = h.svc.UpdateBusinessType(holdingCode, id, authUsername, *docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
		ID:      id,
	})

	return nil
}

// Delete BusinessType godoc
// @Description Delete BusinessType
// @Tags		BusinessType
// @Param		id  path      string  true  "BusinessType ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /organization/business-type/{id} [delete]
func (h BusinessTypeHttp) DeleteBusinessType(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username

	id := ctx.Param("id")

	if db, err := mypg.PgSqlFastConnect("bcai_projection"); err == nil && db != nil {
		return h.deleteBusinessTypePostgres(ctx, db, holdingCode, id)
	}

	err := h.svc.DeleteBusinessType(holdingCode, id, authUsername)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		ID:      id,
	})

	return nil
}

// Delete BusinessType godoc
// @Description Delete BusinessType
// @Tags		BusinessType
// @Param		BusinessType  body      []string  true  "BusinessType GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /organization/business-type [delete]
func (h BusinessTypeHttp) DeleteBusinessTypeByGUIDs(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username

	input := ctx.ReadInput()

	docReq := []string{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if db, err := mypg.PgSqlFastConnect("bcai_projection"); err == nil && db != nil {
		return h.deleteBusinessTypeByGUIDsPostgres(ctx, db, holdingCode, docReq)
	}

	err = h.svc.DeleteBusinessTypeByGUIDs(holdingCode, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get BusinessType godoc
// @Description get BusinessType info by guidfixed
// @Tags		BusinessType
// @Param		id  path      string  true  "BusinessType guidfixed"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /organization/business-type/{id} [get]
func (h BusinessTypeHttp) InfoBusinessType(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get BusinessType %v", id)

	if db, err := mypg.PgSqlFastConnect("bcai_projection"); err == nil && db != nil {
		return h.infoBusinessTypePostgres(ctx, db, holdingCode, id)
	}

	doc, err := h.svc.InfoBusinessType(holdingCode, id)

	if err != nil {
		h.ms.Logger.Errorf("Error getting document %v: %v", id, err)
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    doc,
	})
	return nil
}

// Get BusinessType default godoc
// @Description get BusinessType info default
// @Tags		BusinessType
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /organization/business-type/default [get]
func (h BusinessTypeHttp) InfoBusinessTypeDefault(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	if db, err := mypg.PgSqlFastConnect("bcai_projection"); err == nil && db != nil {
		return h.infoBusinessTypeDefaultPostgres(ctx, db, holdingCode)
	}

	doc, err := h.svc.InfoBusinessTypeDefault(holdingCode)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    doc,
	})
	return nil
}

// Get BusinessType By Code godoc
// @Description get BusinessType info by Code
// @Tags		BusinessType
// @Param		code  path      string  true  "BusinessType Code"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /organization/business-type/code/{code} [get]
func (h BusinessTypeHttp) InfoBusinessTypeByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	code := ctx.Param("code")

	if db, err := mypg.PgSqlFastConnect("bcai_projection"); err == nil && db != nil {
		return h.infoBusinessTypePostgres(ctx, db, holdingCode, code)
	}

	doc, err := h.svc.InfoBusinessTypeByCode(holdingCode, code)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    doc,
	})
	return nil
}

// List BusinessType step godoc
// @Description get list step
// @Tags		BusinessType
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /organization/business-type [get]
func (h BusinessTypeHttp) SearchBusinessTypePage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	pageable := utils.GetPageable(ctx.QueryParam)

	if db, err := mypg.PgSqlFastConnect("bcai_projection"); err == nil && db != nil {
		limit := 100
		offset := pageable.GetOffest()
		if pageable.Limit > 0 {
			limit = pageable.Limit
		}
		q := ctx.QueryParam("q")
		if pageable.Query != "" {
			q = pageable.Query
		}
		return h.searchBusinessTypeStepPostgres(ctx, db, holdingCode, offset, limit, q)
	}

	docList, pagination, err := h.svc.SearchBusinessType(holdingCode, map[string]interface{}{}, pageable)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success:    true,
		Data:       docList,
		Pagination: pagination,
	})
	return nil
}

// List BusinessType godoc
// @Description search limit offset
// @Tags		BusinessType
// @Param		q		query	string		false  "Search Value"
// @Param		offset	query	integer		false  "offset"
// @Param		limit	query	integer		false  "limit"
// @Param		lang	query	string		false  "lang"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /organization/business-type/list [get]
func (h BusinessTypeHttp) SearchBusinessTypeStep(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	pageableStep := utils.GetPageableStep(ctx.QueryParam)

	if db, err := mypg.PgSqlFastConnect("bcai_projection"); err == nil && db != nil {
		limit := 100
		offset := pageableStep.Skip
		if pageableStep.Limit > 0 {
			limit = pageableStep.Limit
		}
		q := ctx.QueryParam("q")
		if pageableStep.Query != "" {
			q = pageableStep.Query
		}
		return h.searchBusinessTypeStepPostgres(ctx, db, holdingCode, offset, limit, q)
	}

	lang := ctx.QueryParam("lang")

	docList, total, err := h.svc.SearchBusinessTypeStep(holdingCode, lang, pageableStep)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    docList,
		Total:   total,
	})
	return nil
}

// Create BusinessType Bulk godoc
// @Description Create BusinessType
// @Tags		BusinessType
// @Param		BusinessType  body      []models.BusinessType  true  "BusinessType"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /organization/business-type/bulk [post]
func (h BusinessTypeHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	input := ctx.ReadInput()

	dataReq := []models.BusinessType{}
	err := json.Unmarshal([]byte(input), &dataReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	bulkResponse, err := h.svc.SaveInBatch(holdingCode, authUsername, dataReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	ctx.Response(
		http.StatusCreated,
		common.BulkResponse{
			Success:    true,
			BulkImport: bulkResponse,
		},
	)

	return nil
}
