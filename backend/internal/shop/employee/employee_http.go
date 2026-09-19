package employee

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
	"smlcloudplatform/internal/shop/employee/models"
	"smlcloudplatform/internal/shop/employee/repositories"
	"smlcloudplatform/internal/shop/employee/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
)

type IEmployeeHttp interface{}

type EmployeeHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IEmployeeHttpService
}

func NewEmployeeHttp(ms *microservice.Microservice, cfg config.IConfig) EmployeeHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())

	repo := repositories.NewEmployeeRepository(pst)

	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewEmployeeHttpService(repo, masterSyncCacheRepo, utils.HashPassword)

	return EmployeeHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h EmployeeHttp) RegisterHttp() {

	h.ms.GET("/holding/employee", h.SearchEmployeePage)
	h.ms.GET("/shop/employee", h.SearchEmployeePage)
	h.ms.GET("/holding/employee/list", h.SearchEmployeeStep)
	h.ms.GET("/shop/employee/list", h.SearchEmployeeStep)
	h.ms.POST("/holding/employee", h.CreateEmployee)
	h.ms.POST("/shop/employee", h.CreateEmployee)
	h.ms.GET("/holding/employee/:id", h.InfoEmployee)
	h.ms.GET("/shop/employee/:id", h.InfoEmployee)
	h.ms.GET("/holding/employee/code/:code", h.InfoEmployeeByCode)
	h.ms.GET("/shop/employee/code/:code", h.InfoEmployeeByCode)
	h.ms.GET("/holding/employee/email/:email", h.InfoEmployeeByEmail)
	h.ms.GET("/shop/employee/email/:email", h.InfoEmployeeByEmail)
	h.ms.PUT("/holding/employee/:id", h.UpdateEmployee)
	h.ms.PUT("/shop/employee/:id", h.UpdateEmployee)
	h.ms.PUT("/holding/employee/password", h.UpdatePassword)
	h.ms.PUT("/shop/employee/password", h.UpdatePassword)
	h.ms.DELETE("/holding/employee/:id", h.DeleteEmployee)
	h.ms.DELETE("/shop/employee/:id", h.DeleteEmployee)
	h.ms.DELETE("/holding/employee", h.DeleteEmployeeByGUIDs)
	h.ms.DELETE("/shop/employee", h.DeleteEmployeeByGUIDs)
}

func (h EmployeeHttp) searchEmployeeStepPostgres(ctx microservice.IContext, db *sql.DB, holdingCode string, offset int, limit int, q string) error {
	holdingCode = strings.TrimSpace(holdingCode)
	if holdingCode == "" {
		ctx.Response(http.StatusOK, common.ApiResponse{Success: true, Data: []models.EmployeeInfo{}, Total: 0})
		return nil
	}
	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	q = strings.TrimSpace(q)
	countQuery := `SELECT COUNT(*) FROM employees WHERE LOWER(holding_code) = LOWER($1) AND is_enabled = true`
	dataQuery := `SELECT id, code, name, email, COALESCE(roles, '[]'::jsonb), is_enabled, is_use_pos, COALESCE(pin_code, ''), COALESCE(contact, '{}'::jsonb), COALESCE(branches, '[]'::jsonb), COALESCE(access_scopes, '[]'::jsonb)
	              FROM employees WHERE LOWER(holding_code) = LOWER($1) AND is_enabled = true`
	args := []interface{}{holdingCode}

	if q != "" {
		countQuery += ` AND (code ILIKE $2 OR name ILIKE $2 OR email ILIKE $2)`
		dataQuery += ` AND (code ILIKE $2 OR name ILIKE $2 OR email ILIKE $2)`
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

	list := make([]models.EmployeeInfo, 0)
	for rows.Next() {
		var (
			id          string
			code        string
			name        string
			email       string
			rawRoles    []byte
			isEnabled   bool
			isUsePOS    bool
			pinCode     string
			rawContact  []byte
			rawBranches []byte
			rawScopes   []byte
		)
		if err := rows.Scan(&id, &code, &name, &email, &rawRoles, &isEnabled, &isUsePOS, &pinCode, &rawContact, &rawBranches, &rawScopes); err != nil {
			continue
		}
		var rolesList []string
		if len(rawRoles) > 0 {
			_ = json.Unmarshal(rawRoles, &rolesList)
		}
		var contact models.EmployeeContact
		if len(rawContact) > 0 {
			_ = json.Unmarshal(rawContact, &contact)
		}
		var branchesList []models.EmployeeBranch
		if len(rawBranches) > 0 {
			_ = json.Unmarshal(rawBranches, &branchesList)
		}
		var scopesList []models.EmployeeAccessScope
		if len(rawScopes) > 0 {
			_ = json.Unmarshal(rawScopes, &scopesList)
		}

		list = append(list, models.EmployeeInfo{
			DocIdentity: common.DocIdentity{
				GuidFixed: id,
			},
			Employee: models.Employee{
				Code:         code,
				Name:         name,
				Email:        email,
				Roles:        &rolesList,
				IsEnabled:    isEnabled,
				IsUsePOS:     isUsePOS,
				PinCode:      pinCode,
				Contact:      contact,
				Branches:     &branchesList,
				AccessScopes: scopesList,
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

func (h EmployeeHttp) infoEmployeePostgres(ctx microservice.IContext, db *sql.DB, holdingCode string, id string) error {
	holdingCode = strings.TrimSpace(holdingCode)
	id = strings.TrimSpace(id)

	query := `SELECT id, code, name, email, COALESCE(roles, '[]'::jsonb), is_enabled, is_use_pos, COALESCE(pin_code, ''), COALESCE(contact, '{}'::jsonb), COALESCE(branches, '[]'::jsonb), COALESCE(access_scopes, '[]'::jsonb)
	          FROM employees
	          WHERE LOWER(holding_code) = LOWER($1) AND (id = $2 OR LOWER(code) = LOWER($2) OR LOWER(email) = LOWER($2)) AND is_enabled = true
	          LIMIT 1`
	var (
		recID       string
		code        string
		name        string
		email       string
		rawRoles    []byte
		isEnabled   bool
		isUsePOS    bool
		pinCode     string
		rawContact  []byte
		rawBranches []byte
		rawScopes   []byte
	)
	err := db.QueryRowContext(context.Background(), query, holdingCode, id).Scan(
		&recID, &code, &name, &email, &rawRoles, &isEnabled, &isUsePOS, &pinCode, &rawContact, &rawBranches, &rawScopes,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.ResponseError(http.StatusNotFound, "ไม่พบข้อมูลพนักงาน")
			return err
		}
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	var rolesList []string
	if len(rawRoles) > 0 {
		_ = json.Unmarshal(rawRoles, &rolesList)
	}
	var contact models.EmployeeContact
	if len(rawContact) > 0 {
		_ = json.Unmarshal(rawContact, &contact)
	}
	var branchesList []models.EmployeeBranch
	if len(rawBranches) > 0 {
		_ = json.Unmarshal(rawBranches, &branchesList)
	}
	var scopesList []models.EmployeeAccessScope
	if len(rawScopes) > 0 {
		_ = json.Unmarshal(rawScopes, &scopesList)
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data: models.EmployeeInfo{
			DocIdentity: common.DocIdentity{
				GuidFixed: recID,
			},
			Employee: models.Employee{
				Code:         code,
				Name:         name,
				Email:        email,
				Roles:        &rolesList,
				IsEnabled:    isEnabled,
				IsUsePOS:     isUsePOS,
				PinCode:      pinCode,
				Contact:      contact,
				Branches:     &branchesList,
				AccessScopes: scopesList,
			},
		},
	})
	return nil
}

func (h EmployeeHttp) createEmployeePostgres(ctx microservice.IContext, db *sql.DB, holdingCode string, docReq models.EmployeeRequestRegister) error {
	holdingCode = strings.TrimSpace(holdingCode)
	code := strings.TrimSpace(docReq.Code)
	if code == "" {
		ctx.ResponseError(http.StatusBadRequest, "code is required")
		return nil
	}

	rawRoles, _ := json.Marshal(docReq.Roles)
	rawContact, _ := json.Marshal(docReq.Contact)
	rawBranches, _ := json.Marshal(docReq.Branches)
	rawScopes, _ := json.Marshal(docReq.AccessScopes)

	newID := uuid.New().String()
	query := `INSERT INTO employees (id, holding_code, code, name, email, roles, is_enabled, is_use_pos, pin_code, contact, branches, access_scopes, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, now(), now())
	          ON CONFLICT (holding_code, code) DO UPDATE SET
	            name = EXCLUDED.name,
	            email = EXCLUDED.email,
	            roles = EXCLUDED.roles,
	            is_enabled = EXCLUDED.is_enabled,
	            is_use_pos = EXCLUDED.is_use_pos,
	            pin_code = EXCLUDED.pin_code,
	            contact = EXCLUDED.contact,
	            branches = EXCLUDED.branches,
	            access_scopes = EXCLUDED.access_scopes,
	            updated_at = now()`
	_, err := db.ExecContext(context.Background(), query,
		newID, holdingCode, code, docReq.Name, docReq.Email, rawRoles, docReq.IsEnabled, docReq.IsUsePOS, docReq.PinCode, rawContact, rawBranches, rawScopes,
	)
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

func (h EmployeeHttp) updateEmployeePostgres(ctx microservice.IContext, db *sql.DB, holdingCode string, id string, docReq models.EmployeeRequestUpdate) error {
	holdingCode = strings.TrimSpace(holdingCode)
	id = strings.TrimSpace(id)

	rawRoles, _ := json.Marshal(docReq.Roles)
	rawContact, _ := json.Marshal(docReq.Contact)
	rawBranches, _ := json.Marshal(docReq.Branches)
	rawScopes, _ := json.Marshal(docReq.AccessScopes)

	query := `UPDATE employees SET
	            name = $1,
	            email = $2,
	            roles = $3,
	            is_enabled = $4,
	            is_use_pos = $5,
	            pin_code = $6,
	            contact = $7,
	            branches = $8,
	            access_scopes = $9,
	            updated_at = now()
	          WHERE LOWER(holding_code) = LOWER($10) AND (id = $11 OR LOWER(code) = LOWER($11))`
	_, err := db.ExecContext(context.Background(), query,
		docReq.Name, docReq.Email, rawRoles, docReq.IsEnabled, docReq.IsUsePOS, docReq.PinCode, rawContact, rawBranches, rawScopes, holdingCode, id,
	)
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

func (h EmployeeHttp) deleteEmployeePostgres(ctx microservice.IContext, db *sql.DB, holdingCode string, id string) error {
	holdingCode = strings.TrimSpace(holdingCode)
	id = strings.TrimSpace(id)

	query := `UPDATE employees SET is_enabled = false, updated_at = now()
	          WHERE LOWER(holding_code) = LOWER($1) AND (id = $2 OR LOWER(code) = LOWER($2))`
	_, err := db.ExecContext(context.Background(), query, holdingCode, id)
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

func (h EmployeeHttp) deleteEmployeeByGUIDsPostgres(ctx microservice.IContext, db *sql.DB, holdingCode string, ids []string) error {
	holdingCode = strings.TrimSpace(holdingCode)
	if len(ids) == 0 {
		ctx.Response(http.StatusOK, common.ApiResponse{Success: true})
		return nil
	}
	query := `UPDATE employees SET is_enabled = false, updated_at = now()
	          WHERE LOWER(holding_code) = LOWER($1) AND (id = ANY($2) OR code = ANY($2))`
	_, err := db.ExecContext(context.Background(), query, holdingCode, pq.Array(ids))
	if err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}
	ctx.Response(http.StatusOK, common.ApiResponse{Success: true})
	return nil
}

// Create Employee godoc
// @Description Create Employee
// @Tags		Employee
// @Param		Employee  body      models.EmployeeRequestRegister  true  "Employee"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /shop/employee [post]
func (h EmployeeHttp) CreateEmployee(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	holdingCode := ctx.UserInfo().HoldingCode
	input := ctx.ReadInput()

	docReq := &models.EmployeeRequestRegister{}
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
		return h.createEmployeePostgres(ctx, db, holdingCode, *docReq)
	}

	idx, err := h.svc.CreateEmployee(holdingCode, authUsername, *docReq)

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

// Update Employee godoc
// @Description Update Employee
// @Tags		Employee
// @Param		id  path      string  true  "Employee ID"
// @Param		Employee  body      models.EmployeeRequestUpdate  true  "Employee"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /shop/employee/{id} [put]
func (h EmployeeHttp) UpdateEmployee(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.EmployeeRequestUpdate{}
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
		return h.updateEmployeePostgres(ctx, db, holdingCode, id, *docReq)
	}

	err = h.svc.UpdateEmployee(holdingCode, id, authUsername, *docReq)

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

// Delete Employee godoc
// @Description Delete Employee
// @Tags		Employee
// @Param		id  path      string  true  "Employee ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /shop/employee/{id} [delete]
func (h EmployeeHttp) DeleteEmployee(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username

	id := ctx.Param("id")

	if db, err := mypg.PgSqlFastConnect("bcai_projection"); err == nil && db != nil {
		return h.deleteEmployeePostgres(ctx, db, holdingCode, id)
	}

	err := h.svc.DeleteEmployee(holdingCode, id, authUsername)

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

// Delete Employee godoc
// @Description Delete Employee
// @Tags		Employee
// @Param		Employee  body      []string  true  "Employee GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /shop/employee [delete]
func (h EmployeeHttp) DeleteEmployeeByGUIDs(ctx microservice.IContext) error {
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
		return h.deleteEmployeeByGUIDsPostgres(ctx, db, holdingCode, docReq)
	}

	err = h.svc.DeleteEmployeeByGUIDs(holdingCode, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get Employee godoc
// @Description get struct array by ID
// @Tags		Employee
// @Param		id  path      string  true  "Employee ID"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /shop/employee/{id} [get]
func (h EmployeeHttp) InfoEmployee(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get Employee %v", id)

	if db, err := mypg.PgSqlFastConnect("bcai_projection"); err == nil && db != nil {
		return h.infoEmployeePostgres(ctx, db, holdingCode, id)
	}

	doc, err := h.svc.InfoEmployee(holdingCode, id)

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

// Get Employee By Code godoc
// @Description get employee by code
// @Tags		Employee
// @Param		code  path      string  true  "Employee code"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /shop/employee/code/{code} [get]
func (h EmployeeHttp) InfoEmployeeByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	code := ctx.Param("code")

	if db, err := mypg.PgSqlFastConnect("bcai_projection"); err == nil && db != nil {
		return h.infoEmployeePostgres(ctx, db, holdingCode, code)
	}

	doc, err := h.svc.InfoEmployeeByCode(holdingCode, code)

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

// Get Employee By Email godoc
// @Description get employee by email
// @Tags		Employee
// @Param		email  path      string  true  "Employee email"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /shop/employee/email/{email} [get]
func (h EmployeeHttp) InfoEmployeeByEmail(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	email := ctx.Param("email")

	if db, err := mypg.PgSqlFastConnect("bcai_projection"); err == nil && db != nil {
		return h.infoEmployeePostgres(ctx, db, holdingCode, email)
	}

	doc, err := h.svc.InfoEmployeeByEmail(holdingCode, email)

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

// List Employee godoc
// @Description get struct array by ID
// @Tags		Employee
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "page"
// @Param		limit	query	integer		false  "limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /shop/employee [get]
func (h EmployeeHttp) SearchEmployeePage(ctx microservice.IContext) error {
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
		return h.searchEmployeeStepPostgres(ctx, db, holdingCode, offset, limit, q)
	}

	docList, pagination, err := h.svc.SearchEmployee(holdingCode, map[string]interface{}{}, pageable)

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

// List Employee godoc
// @Description search limit offset
// @Tags		Employee
// @Param		q		query	string		false  "Search Value"
// @Param		offset	query	integer		false  "offset"
// @Param		limit	query	integer		false  "limit"
// @Param		lang	query	string		false  "lang"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /shop/employee/list [get]
func (h EmployeeHttp) SearchEmployeeStep(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	pageableStep := utils.GetPageableStep(ctx.QueryParam)

	// Set default limit to 1000 if not provided
	if ctx.QueryParam("limit") == "" {
		pageableStep.Limit = 1000
	}

	if db, err := mypg.PgSqlFastConnect("bcai_projection"); err == nil && db != nil {
		limit := 1000
		offset := pageableStep.Skip
		if pageableStep.Limit > 0 {
			limit = pageableStep.Limit
		}
		q := ctx.QueryParam("q")
		if pageableStep.Query != "" {
			q = pageableStep.Query
		}
		return h.searchEmployeeStepPostgres(ctx, db, holdingCode, offset, limit, q)
	}

	lang := ctx.QueryParam("lang")

	docList, total, err := h.svc.SearchEmployeeStep(holdingCode, lang, pageableStep)

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

// Update Password Employee godoc
// @Summary		Update Password Employee
// @Description	Update Password Employee
// @Tags		Employee
// @Param		id  path      string  true  "Employee ID"
// @Param		Employee  body      models.EmployeeRequestPassword  true  "Register Employee"
// @Success		200	{object}	models.ResponseSuccess
// @Failure		400 {object}	models.AuthResponseFailed
// @Accept 		json
// @Security     AccessToken
// @Router		/employee/password [put]
func (h EmployeeHttp) UpdatePassword(ctx microservice.IContext) error {
	userAuthInfo := ctx.UserInfo()
	authUsername := userAuthInfo.Username
	holdingCode := userAuthInfo.HoldingCode

	input := ctx.ReadInput()

	userPwdReq := models.EmployeeRequestPassword{}
	err := json.Unmarshal([]byte(input), &userPwdReq)

	if err != nil {
		ctx.ResponseError(400, "user payload invalid")
		return err
	}

	err = h.svc.UpdatePassword(holdingCode, authUsername, userPwdReq)

	if err != nil {
		ctx.Response(http.StatusBadRequest, common.ApiResponse{
			Success: false,
			Message: err.Error(),
		})
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
	})

	return nil
}
