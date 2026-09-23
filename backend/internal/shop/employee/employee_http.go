package employee

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"smlcloudplatform/internal/centraldb"
	"smlcloudplatform/internal/config"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/shop/employee/models"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/apperr"
	"smlcloudplatform/pkg/microservice"
)

const employeeQueryTimeout = 5 * time.Second

type EmployeeHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
}

func NewEmployeeHttp(ms *microservice.Microservice, cfg config.IConfig) EmployeeHttp {
	return EmployeeHttp{ms: ms, cfg: cfg}
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
	h.ms.GET("/holding/employee/code/:code", h.InfoEmployee)
	h.ms.GET("/shop/employee/code/:code", h.InfoEmployee)
	h.ms.GET("/holding/employee/email/:email", h.InfoEmployee)
	h.ms.GET("/shop/employee/email/:email", h.InfoEmployee)
	h.ms.PUT("/holding/employee/:id", h.UpdateEmployee)
	h.ms.PUT("/shop/employee/:id", h.UpdateEmployee)
	h.ms.DELETE("/holding/employee/:id", h.DeleteEmployee)
	h.ms.DELETE("/shop/employee/:id", h.DeleteEmployee)
	h.ms.DELETE("/holding/employee", h.DeleteEmployeeByGUIDs)
	h.ms.DELETE("/shop/employee", h.DeleteEmployeeByGUIDs)
}

// withEmployeeDB opens the central database for the caller's Holding.
func withEmployeeDB(ctx microservice.IContext, fn func(context.Context, *sql.DB, string) error) error {
	holdingCode := strings.ToLower(strings.TrimSpace(ctx.UserInfo().HoldingCode))
	if holdingCode == "" {
		return apperr.Respond(ctx, apperr.ErrForbidden.WithMessage("holding is required").WithThaiMessage("กรุณาเลือกกลุ่มกิจการก่อน"))
	}
	db, err := centraldb.Open()
	if err != nil {
		return apperr.Respond(ctx, apperr.ErrInternal.WithWrap(err))
	}
	reqCtx, cancel := context.WithTimeout(context.Background(), employeeQueryTimeout)
	defer cancel()
	return fn(reqCtx, db, holdingCode)
}

func readEmployee(ctx microservice.IContext) (models.Employee, error) {
	doc := models.Employee{}
	if err := json.Unmarshal([]byte(ctx.ReadInput()), &doc); err != nil {
		return doc, apperr.ErrBadRequest.WithMessage("payload invalid").WithWrap(err)
	}
	if err := ctx.Validate(doc); err != nil {
		return doc, apperr.ErrBadRequest.WithMessage(err.Error()).WithWrap(err)
	}
	doc.Code = strings.TrimSpace(doc.Code)
	doc.Name = strings.TrimSpace(doc.Name)
	doc.Email = strings.TrimSpace(doc.Email)
	return doc, nil
}

type employeeJSON struct {
	roles, contact, branches, scopes []byte
}

func marshalEmployeeJSON(doc models.Employee) (employeeJSON, error) {
	var out employeeJSON
	var err error
	roles := []string{}
	if doc.Roles != nil {
		roles = *doc.Roles
	}
	branches := []models.EmployeeBranch{}
	if doc.Branches != nil {
		branches = *doc.Branches
	}
	scopes := doc.AccessScopes
	if scopes == nil {
		scopes = []models.EmployeeAccessScope{}
	}
	if out.roles, err = json.Marshal(roles); err != nil {
		return out, err
	}
	if out.contact, err = json.Marshal(doc.Contact); err != nil {
		return out, err
	}
	if out.branches, err = json.Marshal(branches); err != nil {
		return out, err
	}
	out.scopes, err = json.Marshal(scopes)
	return out, err
}

const employeeColumns = `id, code, name, email, roles, is_enabled, is_use_pos, pin_code, contact, branches, access_scopes, profile_picture, profile_picture_thumb`

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func scanEmployee(row rowScanner) (models.EmployeeInfo, error) {
	var (
		info                                     models.EmployeeInfo
		rawRoles, rawContact, rawBranches, rawSc []byte
	)
	emp := &info.Employee
	if err := row.Scan(&info.GuidFixed, &emp.Code, &emp.Name, &emp.Email, &rawRoles, &emp.IsEnabled, &emp.IsUsePOS, &emp.PinCode,
		&rawContact, &rawBranches, &rawSc, &emp.ProfilePicture, &emp.ProfilePictureThumb); err != nil {
		return info, err
	}
	roles := []string{}
	branches := []models.EmployeeBranch{}
	for _, part := range []struct {
		raw    []byte
		target interface{}
	}{{rawRoles, &roles}, {rawContact, &emp.Contact}, {rawBranches, &branches}, {rawSc, &emp.AccessScopes}} {
		if len(part.raw) > 0 {
			if err := json.Unmarshal(part.raw, part.target); err != nil {
				return info, err
			}
		}
	}
	emp.Roles = &roles
	emp.Branches = &branches
	return info, nil
}

func likePattern(q string) string {
	return "%" + strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(q) + "%"
}

func (h EmployeeHttp) searchEmployees(ctx microservice.IContext, offset int, limit int, q string) error {
	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return withEmployeeDB(ctx, func(reqCtx context.Context, db *sql.DB, holdingCode string) error {
		where := ` FROM employees WHERE holding_code = $1 AND is_enabled = true`
		args := []interface{}{holdingCode}
		if q = strings.TrimSpace(q); q != "" {
			where += ` AND (code ILIKE $2 OR name ILIKE $2 OR email ILIKE $2)`
			args = append(args, likePattern(q))
		}
		var total int64
		if err := db.QueryRowContext(reqCtx, `SELECT COUNT(*)`+where, args...).Scan(&total); err != nil {
			return apperr.Respond(ctx, apperr.ErrInternal.WithWrap(err))
		}
		limitArg, offsetArg := len(args)+1, len(args)+2
		args = append(args, limit, offset)
		rows, err := db.QueryContext(reqCtx, `SELECT `+employeeColumns+where+
			` ORDER BY code LIMIT $`+strconv.Itoa(limitArg)+` OFFSET $`+strconv.Itoa(offsetArg), args...)
		if err != nil {
			return apperr.Respond(ctx, apperr.ErrInternal.WithWrap(err))
		}
		defer rows.Close()
		list := make([]models.EmployeeInfo, 0)
		for rows.Next() {
			info, err := scanEmployee(rows)
			if err != nil {
				return apperr.Respond(ctx, apperr.ErrInternal.WithWrap(err))
			}
			list = append(list, info)
		}
		if err := rows.Err(); err != nil {
			return apperr.Respond(ctx, apperr.ErrInternal.WithWrap(err))
		}
		ctx.Response(http.StatusOK, common.ApiResponse{Success: true, Data: list, Total: total})
		return nil
	})
}

func (h EmployeeHttp) SearchEmployeePage(ctx microservice.IContext) error {
	pageable := utils.GetPageable(ctx.QueryParam)
	return h.searchEmployees(ctx, pageable.GetOffest(), pageable.Limit, pageable.Query)
}

func (h EmployeeHttp) SearchEmployeeStep(ctx microservice.IContext) error {
	pageableStep := utils.GetPageableStep(ctx.QueryParam)
	limit := pageableStep.Limit
	if ctx.QueryParam("limit") == "" {
		limit = 1000
	}
	return h.searchEmployees(ctx, pageableStep.Skip, limit, pageableStep.Query)
}

// InfoEmployee serves /:id, /code/:code and /email/:email — one lookup by id, code or email.
func (h EmployeeHttp) InfoEmployee(ctx microservice.IContext) error {
	key := strings.TrimSpace(firstNonBlank(ctx.Param("id"), ctx.Param("code"), ctx.Param("email")))
	return withEmployeeDB(ctx, func(reqCtx context.Context, db *sql.DB, holdingCode string) error {
		info, err := scanEmployee(db.QueryRowContext(reqCtx, `SELECT `+employeeColumns+` FROM employees
			WHERE holding_code = $1 AND is_enabled = true AND (id = $2 OR LOWER(code) = LOWER($2) OR (email <> '' AND LOWER(email) = LOWER($2)))
			ORDER BY (id = $2) DESC, (LOWER(code) = LOWER($2)) DESC, code LIMIT 1`, holdingCode, key))
		if errors.Is(err, sql.ErrNoRows) {
			return apperr.Respond(ctx, apperr.NotFound("employee").WithThaiMessage("ไม่พบข้อมูลพนักงาน"))
		}
		if err != nil {
			return apperr.Respond(ctx, apperr.ErrInternal.WithWrap(err))
		}
		ctx.Response(http.StatusOK, common.ApiResponse{Success: true, Data: info})
		return nil
	})
}

func firstNonBlank(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func (h EmployeeHttp) CreateEmployee(ctx microservice.IContext) error {
	doc, err := readEmployee(ctx)
	if err != nil {
		return apperr.RespondErr(ctx, err)
	}
	if doc.Code == "" {
		return apperr.Respond(ctx, apperr.Validation("code", "required").WithThaiMessage("กรุณาระบุรหัสพนักงาน"))
	}
	raw, err := marshalEmployeeJSON(doc)
	if err != nil {
		return apperr.Respond(ctx, apperr.ErrBadRequest.WithWrap(err))
	}
	return withEmployeeDB(ctx, func(reqCtx context.Context, db *sql.DB, holdingCode string) error {
		newID := uuid.NewString()
		_, err := db.ExecContext(reqCtx, `INSERT INTO employees (id, holding_code, code, name, email, roles, is_enabled, is_use_pos, pin_code,
				contact, branches, access_scopes, profile_picture, profile_picture_thumb, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, now(), now())`,
			newID, holdingCode, doc.Code, doc.Name, doc.Email, raw.roles, doc.IsEnabled, doc.IsUsePOS, doc.PinCode,
			raw.contact, raw.branches, raw.scopes, doc.ProfilePicture, doc.ProfilePictureThumb)
		if centraldb.IsUniqueViolation(err) {
			return apperr.Respond(ctx, apperr.DuplicateCode("code", doc.Code).WithThaiMessage("รหัสพนักงานนี้มีอยู่แล้ว"))
		}
		if err != nil {
			return apperr.Respond(ctx, apperr.ErrInternal.WithWrap(err))
		}
		ctx.Response(http.StatusCreated, common.ApiResponse{Success: true, ID: newID})
		return nil
	})
}

func (h EmployeeHttp) UpdateEmployee(ctx microservice.IContext) error {
	id := strings.TrimSpace(ctx.Param("id"))
	doc, err := readEmployee(ctx)
	if err != nil {
		return apperr.RespondErr(ctx, err)
	}
	raw, err := marshalEmployeeJSON(doc)
	if err != nil {
		return apperr.Respond(ctx, apperr.ErrBadRequest.WithWrap(err))
	}
	return withEmployeeDB(ctx, func(reqCtx context.Context, db *sql.DB, holdingCode string) error {
		// The employee code is its business identity; it is not renamed through update.
		result, err := db.ExecContext(reqCtx, `UPDATE employees SET name = $3, email = $4, roles = $5, is_enabled = $6, is_use_pos = $7,
				pin_code = $8, contact = $9, branches = $10, access_scopes = $11, profile_picture = $12, profile_picture_thumb = $13, updated_at = now()
			WHERE holding_code = $1 AND (id = $2 OR LOWER(code) = LOWER($2))`,
			holdingCode, id, doc.Name, doc.Email, raw.roles, doc.IsEnabled, doc.IsUsePOS,
			doc.PinCode, raw.contact, raw.branches, raw.scopes, doc.ProfilePicture, doc.ProfilePictureThumb)
		if err != nil {
			return apperr.Respond(ctx, apperr.ErrInternal.WithWrap(err))
		}
		if affected, _ := result.RowsAffected(); affected == 0 {
			return apperr.Respond(ctx, apperr.NotFound("employee").WithThaiMessage("ไม่พบข้อมูลพนักงาน"))
		}
		ctx.Response(http.StatusOK, common.ApiResponse{Success: true, ID: id})
		return nil
	})
}

// DeleteEmployee disables the employee; rows stay for the audit trail.
func (h EmployeeHttp) DeleteEmployee(ctx microservice.IContext) error {
	id := strings.TrimSpace(ctx.Param("id"))
	return h.disableEmployees(ctx, []string{id}, id)
}

func (h EmployeeHttp) DeleteEmployeeByGUIDs(ctx microservice.IContext) error {
	ids := []string{}
	if err := json.Unmarshal([]byte(ctx.ReadInput()), &ids); err != nil {
		return apperr.Respond(ctx, apperr.ErrBadRequest.WithMessage("payload invalid").WithWrap(err))
	}
	return h.disableEmployees(ctx, ids, "")
}

func (h EmployeeHttp) disableEmployees(ctx microservice.IContext, ids []string, responseID string) error {
	cleaned := make([]string, 0, len(ids))
	for _, id := range ids {
		if id = strings.TrimSpace(id); id != "" {
			cleaned = append(cleaned, id)
		}
	}
	if len(cleaned) == 0 {
		ctx.Response(http.StatusOK, common.ApiResponse{Success: true, ID: responseID})
		return nil
	}
	return withEmployeeDB(ctx, func(reqCtx context.Context, db *sql.DB, holdingCode string) error {
		_, err := db.ExecContext(reqCtx, `UPDATE employees SET is_enabled = false, updated_at = now()
			WHERE holding_code = $1 AND (id = ANY($2) OR code = ANY($2))`, holdingCode, pq.Array(cleaned))
		if err != nil {
			return apperr.Respond(ctx, apperr.ErrInternal.WithWrap(err))
		}
		ctx.Response(http.StatusOK, common.ApiResponse{Success: true, ID: responseID})
		return nil
	})
}
