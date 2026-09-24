package rolepermission

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"smlcloudplatform/internal/centraldb"
	"smlcloudplatform/internal/config"
	common "smlcloudplatform/internal/models"
	rolemodels "smlcloudplatform/internal/organization/rolepermission/models"
	"smlcloudplatform/pkg/microservice"
	"smlcloudplatform/pkg/textguard"

	"github.com/lib/pq"
)

const rolePermissionTimeout = 15 * time.Second

type RolePermissionHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
}

func NewRolePermissionHttp(ms *microservice.Microservice, cfg config.IConfig) RolePermissionHttp {
	return RolePermissionHttp{ms: ms, cfg: cfg}
}

func (h RolePermissionHttp) RegisterHttp() {
	h.ms.GET("/organization/role-permission/me", h.InfoMyRolePermission)
	h.ms.GET("/organization/role-permission", h.SearchRolePermissions)
	h.ms.POST("/organization/role-permission", h.CreateRolePermission)
	h.ms.GET("/organization/role-permission/:id", h.InfoRolePermission)
	h.ms.PUT("/organization/role-permission/:id", h.UpdateRolePermission)
	h.ms.DELETE("/organization/role-permission/:id", h.DeleteRolePermission)
}

type RolePermissionItem struct {
	ID          string                     `json:"_id"`
	HoldingCode string                     `json:"holdingcode"`
	RoleCode    string                     `json:"rolecode"`
	Names       []rolemodels.LocalizedName `json:"names"`
	Permissions []string                   `json:"permissions"`
	IsActive    bool                       `json:"isactive"`
	CreatedAt   time.Time                  `json:"createdat,omitempty"`
	UpdatedAt   time.Time                  `json:"updatedat,omitempty"`
	IsDeleted   bool                       `json:"isdeleted"`
	Version     int64                      `json:"__v"`
}

func (h RolePermissionHttp) searchRolePermissionsPostgres(ctx microservice.IContext, db *sql.DB, holdingCode string, q string, offset int, limit int) error {
	if !authorizePostgresRoleManager(ctx, db, holdingCode) {
		return nil
	}
	holdingCode = strings.TrimSpace(holdingCode)
	if holdingCode == "" {
		ctx.Response(http.StatusOK, common.ApiResponse{Success: true, Data: []RolePermissionItem{}, Total: 0})
		return nil
	}
	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	q = strings.TrimSpace(q)
	countQuery := `SELECT COUNT(*) FROM role_permissions WHERE LOWER(holding_code) = LOWER($1) AND is_active = true`
	dataQuery := `SELECT COALESCE(id, role_code), holding_code, role_code, names, permissions, is_active, created_at, updated_at, version
	              FROM role_permissions WHERE LOWER(holding_code) = LOWER($1) AND is_active = true`
	args := []interface{}{holdingCode}

	if q != "" {
		countQuery += ` AND (role_code ILIKE $2 OR names::text ILIKE $2)`
		dataQuery += ` AND (role_code ILIKE $2 OR names::text ILIKE $2)`
		args = append(args, "%"+q+"%")
	}

	qCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var total int64
	if err := db.QueryRowContext(qCtx, countQuery, args...).Scan(&total); err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	dataQuery += fmt.Sprintf(` ORDER BY role_code LIMIT %d OFFSET %d`, limit, offset)
	rows, err := db.QueryContext(qCtx, dataQuery, args...)
	if err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}
	defer rows.Close()

	list := make([]RolePermissionItem, 0)
	for rows.Next() {
		var (
			id             string
			hCode          string
			roleCode       string
			rawNames       []byte
			rawPermissions []byte
			isActive       bool
			createdAt      time.Time
			updatedAt      time.Time
			version        int64
		)
		if err := rows.Scan(&id, &hCode, &roleCode, &rawNames, &rawPermissions, &isActive, &createdAt, &updatedAt, &version); err != nil {
			continue
		}
		var namesList []rolemodels.LocalizedName
		if len(rawNames) > 0 {
			_ = json.Unmarshal(rawNames, &namesList)
		}
		var permList []string
		if len(rawPermissions) > 0 {
			_ = json.Unmarshal(rawPermissions, &permList)
		}
		list = append(list, RolePermissionItem{
			ID:          id,
			HoldingCode: hCode,
			RoleCode:    roleCode,
			Names:       namesList,
			Permissions: permList,
			IsActive:    isActive,
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
			IsDeleted:   false,
			Version:     version,
		})
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    list,
		Total:   total,
	})
	return nil
}

func (h RolePermissionHttp) infoRolePermissionPostgres(ctx microservice.IContext, db *sql.DB, holdingCode string, id string) error {
	if !authorizePostgresRoleManager(ctx, db, holdingCode) {
		return nil
	}
	holdingCode = strings.TrimSpace(holdingCode)
	id = strings.TrimSpace(id)

	query := `SELECT COALESCE(id, role_code), holding_code, role_code, names, permissions, is_active, created_at, updated_at, version
	          FROM role_permissions
	          WHERE LOWER(holding_code) = LOWER($1) AND (id = $2 OR LOWER(role_code) = LOWER($2)) AND is_active = true
	          LIMIT 1`
	var (
		recID          string
		hCode          string
		roleCode       string
		rawNames       []byte
		rawPermissions []byte
		isActive       bool
		createdAt      time.Time
		updatedAt      time.Time
		version        int64
	)
	err := db.QueryRowContext(context.Background(), query, holdingCode, id).Scan(
		&recID, &hCode, &roleCode, &rawNames, &rawPermissions, &isActive, &createdAt, &updatedAt, &version,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.ResponseError(http.StatusNotFound, "ไม่พบรายการสิทธิ์ตามบทบาท")
			return err
		}
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	var namesList []rolemodels.LocalizedName
	if len(rawNames) > 0 {
		_ = json.Unmarshal(rawNames, &namesList)
	}
	var permList []string
	if len(rawPermissions) > 0 {
		_ = json.Unmarshal(rawPermissions, &permList)
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data: RolePermissionItem{
			ID:          recID,
			HoldingCode: hCode,
			RoleCode:    roleCode,
			Names:       namesList,
			Permissions: permList,
			IsActive:    isActive,
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
			IsDeleted:   false,
			Version:     version,
		},
	})
	return nil
}

func (h RolePermissionHttp) infoMyRolePermissionPostgres(ctx microservice.IContext, db *sql.DB, holdingCode string, username string) error {
	holdingCode = strings.TrimSpace(holdingCode)
	username = strings.TrimSpace(username)

	roleCode, rawSets, err := postgresMemberRole(ctx, db, holdingCode)
	if err != nil {
		respondMemberRoleError(ctx, err, "ปฏิเสธสิทธิ์: ไม่พบสมาชิกที่ใช้งานได้ใน Holding")
		return nil
	}
	var permissionSets []string
	if len(rawSets) > 0 && json.Unmarshal(rawSets, &permissionSets) != nil {
		ctx.ResponseError(http.StatusForbidden, "ปฏิเสธสิทธิ์: ข้อมูลชุดสิทธิ์ไม่ถูกต้อง")
		return nil
	}
	setCodes := append([]string{roleCode}, permissionSets...)

	// Codes are matched ignoring case: older rows may be stored as "admin" while members list "ADMIN".
	for i, code := range setCodes {
		setCodes[i] = strings.ToUpper(strings.TrimSpace(code))
	}
	rows, err := db.QueryContext(context.Background(),
		`SELECT permissions FROM role_permissions WHERE LOWER(holding_code) = LOWER($1) AND UPPER(role_code) = ANY($2) AND is_active = true`,
		holdingCode, pq.Array(setCodes),
	)
	allPermissions := make([]string, 0)
	hasWildcard := false
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var rawPerms []byte
			if err := rows.Scan(&rawPerms); err == nil && len(rawPerms) > 0 {
				var pList []string
				if err := json.Unmarshal(rawPerms, &pList); err == nil {
					for _, p := range pList {
						if p == "*" {
							hasWildcard = true
						}
						allPermissions = append(allPermissions, p)
					}
				}
			}
		}
	}

	var finalPerms []string
	if roleCode == "ADMIN" || roleCode == "OWNER" || hasWildcard {
		finalPerms = []string{"*"}
	} else {
		seen := make(map[string]bool)
		for _, p := range allPermissions {
			if !seen[p] {
				seen[p] = true
				finalPerms = append(finalPerms, p)
			}
		}
		sort.Strings(finalPerms)
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data: map[string]interface{}{
			"holdingcode":    holdingCode,
			"rolecode":       roleCode,
			"permissionsets": permissionSets,
			"permissions":    finalPerms,
			"isactive":       true,
		},
	})
	return nil
}

func (h RolePermissionHttp) createRolePermissionPostgres(ctx microservice.IContext, db *sql.DB, holdingCode string, req rolemodels.RolePermissionRequest) error {
	if !authorizePostgresRoleManager(ctx, db, holdingCode, req.RoleCode) {
		return nil
	}
	if path := textguard.NULField(req); path != "" {
		return respondPermissionSetNUL(ctx, path)
	}
	holdingCode = strings.TrimSpace(holdingCode)
	roleCode := strings.ToUpper(strings.TrimSpace(req.RoleCode))
	if roleCode == "" {
		ctx.ResponseError(http.StatusBadRequest, "rolecode is required")
		return nil
	}

	rawNames, _ := json.Marshal(req.Names)
	rawPerms, _ := json.Marshal(req.Permissions)
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	// Never overwrite an existing set: a code already used (in any spelling) answers 409.
	saved, err := createPermissionSet(ctx.Request().Context(), db, holdingCode, roleCode,
		permissionSetValues{names: rawNames, permissions: rawPerms, isActive: isActive})
	if err != nil {
		return respondPermissionSetError(ctx, err)
	}

	// Timestamps and __v are what the row stored (was zero time "0001-01-01" and no version).
	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
		ID:      saved.id,
		Data: RolePermissionItem{
			ID:          saved.id,
			HoldingCode: holdingCode,
			RoleCode:    saved.roleCode,
			Names:       req.Names,
			Permissions: req.Permissions,
			IsActive:    isActive,
			CreatedAt:   saved.createdAt,
			UpdatedAt:   saved.updatedAt,
			Version:     saved.version,
		},
	})
	return nil
}

func (h RolePermissionHttp) updateRolePermissionPostgres(ctx microservice.IContext, db *sql.DB, holdingCode string, id string, req rolemodels.RolePermissionRequest) error {
	if !authorizePostgresRoleManager(ctx, db, holdingCode) {
		return nil
	}
	oldRole, found := postgresTargetRole(ctx, db, holdingCode, id)
	if !found || !authorizePostgresRoleManager(ctx, db, holdingCode, oldRole, req.RoleCode) {
		return nil
	}
	if path := textguard.NULField(req); path != "" {
		return respondPermissionSetNUL(ctx, path)
	}
	holdingCode = strings.TrimSpace(holdingCode)
	id = strings.TrimSpace(id)
	roleCode := strings.ToUpper(strings.TrimSpace(req.RoleCode))

	rawNames, _ := json.Marshal(req.Names)
	rawPerms, _ := json.Marshal(req.Permissions)
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	if req.Version == nil { // readRolePermissionRequest refuses this for HTTP; guards direct callers
		return respondPermissionSetError(ctx, errPermissionSetChanged)
	}
	version, err := updatePermissionSet(ctx.Request().Context(), db, holdingCode, oldRole, roleCode, *req.Version,
		permissionSetValues{names: rawNames, permissions: rawPerms, isActive: isActive})
	if err != nil {
		return respondPermissionSetError(ctx, err)
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		ID:      id,
		Data:    map[string]int64{"__v": version},
	})
	return nil
}

func (h RolePermissionHttp) deleteRolePermissionPostgres(ctx microservice.IContext, db *sql.DB, holdingCode string, id string, version int64) error {
	if !authorizePostgresRoleManager(ctx, db, holdingCode) {
		return nil
	}
	oldRole, found := postgresTargetRole(ctx, db, holdingCode, id)
	if !found || !authorizePostgresRoleManager(ctx, db, holdingCode, oldRole) {
		return nil
	}
	holdingCode = strings.TrimSpace(holdingCode)
	id = strings.TrimSpace(id)

	if err := deletePermissionSet(ctx.Request().Context(), db, holdingCode, oldRole, version); err != nil {
		return respondPermissionSetError(ctx, err)
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		ID:      id,
	})
	return nil
}

func withCentralDB(ctx microservice.IContext, fn func(db *sql.DB) error) error {
	db, err := centraldb.Open()
	if err != nil {
		ctx.ResponseError(http.StatusInternalServerError, "ไม่สามารถเชื่อมต่อฐานข้อมูลได้")
		return err
	}
	return fn(db)
}

func (h RolePermissionHttp) SearchRolePermissions(ctx microservice.IContext) error {
	holdingCode := strings.TrimSpace(ctx.UserInfo().HoldingCode)
	limit := boundedQueryInt(ctx.QueryParam("limit"), 100, 1, 1000)
	offset := boundedQueryInt(ctx.QueryParam("offset"), 0, 0, 1_000_000)
	return withCentralDB(ctx, func(db *sql.DB) error {
		return h.searchRolePermissionsPostgres(ctx, db, holdingCode, ctx.QueryParam("q"), offset, limit)
	})
}

func (h RolePermissionHttp) InfoRolePermission(ctx microservice.IContext) error {
	holdingCode := strings.TrimSpace(ctx.UserInfo().HoldingCode)
	id := strings.TrimSpace(ctx.Param("id"))
	return withCentralDB(ctx, func(db *sql.DB) error {
		return h.infoRolePermissionPostgres(ctx, db, holdingCode, id)
	})
}

func (h RolePermissionHttp) InfoMyRolePermission(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	return withCentralDB(ctx, func(db *sql.DB) error {
		return h.infoMyRolePermissionPostgres(ctx, db, userInfo.HoldingCode, userInfo.Username)
	})
}

func (h RolePermissionHttp) CreateRolePermission(ctx microservice.IContext) error {
	holdingCode := strings.TrimSpace(ctx.UserInfo().HoldingCode)
	req, err := readRolePermissionRequest(ctx, true)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}
	return withCentralDB(ctx, func(db *sql.DB) error {
		return h.createRolePermissionPostgres(ctx, db, holdingCode, req)
	})
}

func (h RolePermissionHttp) UpdateRolePermission(ctx microservice.IContext) error {
	holdingCode := strings.TrimSpace(ctx.UserInfo().HoldingCode)
	id := strings.TrimSpace(ctx.Param("id"))
	req, err := readRolePermissionRequest(ctx, false)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}
	return withCentralDB(ctx, func(db *sql.DB) error {
		return h.updateRolePermissionPostgres(ctx, db, holdingCode, id, req)
	})
}

func (h RolePermissionHttp) DeleteRolePermission(ctx microservice.IContext) error {
	holdingCode := strings.TrimSpace(ctx.UserInfo().HoldingCode)
	id := strings.TrimSpace(ctx.Param("id"))
	version, err := deleteVersion(ctx.QueryParam("__v"))
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}
	return withCentralDB(ctx, func(db *sql.DB) error {
		return h.deleteRolePermissionPostgres(ctx, db, holdingCode, id, version)
	})
}

// deleteVersion reads the __v query of a delete (the version the screen loaded; the settings
// screen sends it) — required like the update's __v so a stale delete cannot switch a set off.
func deleteVersion(raw string) (int64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, errors.New("ต้องระบุ __v เพื่อป้องกันการลบทับข้อมูลใหม่ — รีเฟรชรายการแล้วลบอีกครั้ง")
	}
	version, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || version < 0 {
		return 0, errors.New("__v ต้องเป็นเลขจำนวนเต็มไม่น้อยกว่า 0 — รีเฟรชรายการแล้วลบอีกครั้ง")
	}
	return version, nil
}

func readRolePermissionRequest(ctx microservice.IContext, creating bool) (rolemodels.RolePermissionRequest, error) {
	var req rolemodels.RolePermissionRequest
	if err := json.Unmarshal([]byte(ctx.ReadInput()), &req); err != nil {
		return req, errors.New("รูปแบบข้อมูลไม่ถูกต้อง")
	}
	if creating && req.IsActive == nil {
		defaultActive := true
		req.IsActive = &defaultActive
	}
	if !creating && req.Version == nil {
		return req, errors.New("ต้องระบุ __v เพื่อป้องกันการบันทึกทับข้อมูลใหม่")
	}
	if req.Version != nil && *req.Version < 0 {
		return req, errors.New("__v ต้องไม่น้อยกว่า 0")
	}
	if req.IsActive == nil {
		return req, errors.New("ต้องระบุสถานะเปิดใช้งาน")
	}
	if err := rolemodels.NormalizeRequest(&req); err != nil {
		return req, err
	}
	return req, nil
}

func boundedQueryInt(raw string, fallback, minimum, maximum int) int {
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return fallback
	}
	if value < minimum {
		return minimum
	}
	if value > maximum {
		return maximum
	}
	return value
}
