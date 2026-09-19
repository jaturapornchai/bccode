package rolepermission

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"
	authmodels "smlcloudplatform/internal/authentication/models"
	"smlcloudplatform/internal/config"
	mypg "smlcloudplatform/internal/goapi/mypg"
	common "smlcloudplatform/internal/models"
	orgpolicy "smlcloudplatform/internal/organization/access"
	rolemodels "smlcloudplatform/internal/organization/rolepermission/models"
	"smlcloudplatform/pkg/microservice"
	msmodels "smlcloudplatform/pkg/microservice/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const rolePermissionTimeout = 15 * time.Second

var allowedRoleCodes = map[string]uint8{
	"USER":  authmodels.ROLE_USER,
	"ADMIN": authmodels.ROLE_ADMIN,
	"OWNER": authmodels.ROLE_OWNER,
}

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
	dataQuery := `SELECT COALESCE(id, role_code), holding_code, role_code, names, permissions, is_active, created_at, updated_at
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
		)
		if err := rows.Scan(&id, &hCode, &roleCode, &rawNames, &rawPermissions, &isActive, &createdAt, &updatedAt); err != nil {
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
			Version:     0,
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
	holdingCode = strings.TrimSpace(holdingCode)
	id = strings.TrimSpace(id)

	query := `SELECT COALESCE(id, role_code), holding_code, role_code, names, permissions, is_active, created_at, updated_at
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
	)
	err := db.QueryRowContext(context.Background(), query, holdingCode, id).Scan(
		&recID, &hCode, &roleCode, &rawNames, &rawPermissions, &isActive, &createdAt, &updatedAt,
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
			Version:     0,
		},
	})
	return nil
}

func (h RolePermissionHttp) infoMyRolePermissionPostgres(ctx microservice.IContext, db *sql.DB, holdingCode string, username string) error {
	holdingCode = strings.TrimSpace(holdingCode)
	username = strings.TrimSpace(username)

	var (
		roleNum        int
		rawSets        []byte
		permissionSets []string
	)

	userQuery := `SELECT COALESCE(hm.role, u.role, 1), COALESCE(hm.permission_sets, '[]'::jsonb)
	              FROM users u
	              LEFT JOIN holding_members hm ON LOWER(hm.holding_code) = LOWER(u.holding_code) AND hm.user_id = u.id
	              WHERE LOWER(u.username) = LOWER($1) AND (u.holding_code IS NULL OR LOWER(u.holding_code) = LOWER($2))
	              LIMIT 1`
	err := db.QueryRowContext(context.Background(), userQuery, username, holdingCode).Scan(&roleNum, &rawSets)
	if err != nil {
		roleNum = int(authmodels.ROLE_OWNER)
	}
	if len(rawSets) > 0 {
		_ = json.Unmarshal(rawSets, &permissionSets)
	}

	roleCode := "OWNER"
	if r, ok := roleCodeFromRole(uint8(roleNum)); ok {
		roleCode = r
	}

	setCodes := append([]string{roleCode}, permissionSets...)

	rows, err := db.QueryContext(context.Background(),
		`SELECT permissions FROM role_permissions WHERE LOWER(holding_code) = LOWER($1) AND role_code = ANY($2) AND is_active = true`,
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

	newID := "rp-" + strings.ToLower(roleCode)
	query := `INSERT INTO role_permissions (id, holding_code, role_code, names, permissions, is_active, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6, now(), now())
	          ON CONFLICT (holding_code, role_code) DO UPDATE SET
	            names = EXCLUDED.names,
	            permissions = EXCLUDED.permissions,
	            is_active = EXCLUDED.is_active,
	            updated_at = now()`
	_, err := db.ExecContext(context.Background(), query, newID, holdingCode, roleCode, rawNames, rawPerms, isActive)
	if err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
		ID:      newID,
		Data: RolePermissionItem{
			ID:          newID,
			HoldingCode: holdingCode,
			RoleCode:    roleCode,
			Names:       req.Names,
			Permissions: req.Permissions,
			IsActive:    isActive,
		},
	})
	return nil
}

func (h RolePermissionHttp) updateRolePermissionPostgres(ctx microservice.IContext, db *sql.DB, holdingCode string, id string, req rolemodels.RolePermissionRequest) error {
	holdingCode = strings.TrimSpace(holdingCode)
	id = strings.TrimSpace(id)
	roleCode := strings.ToUpper(strings.TrimSpace(req.RoleCode))

	rawNames, _ := json.Marshal(req.Names)
	rawPerms, _ := json.Marshal(req.Permissions)
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	query := `UPDATE role_permissions SET
	            role_code = COALESCE(NULLIF($1, ''), role_code),
	            names = $2,
	            permissions = $3,
	            is_active = $4,
	            updated_at = now()
	          WHERE LOWER(holding_code) = LOWER($5) AND (id = $6 OR LOWER(role_code) = LOWER($6))`
	_, err := db.ExecContext(context.Background(), query, roleCode, rawNames, rawPerms, isActive, holdingCode, id)
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

func (h RolePermissionHttp) deleteRolePermissionPostgres(ctx microservice.IContext, db *sql.DB, holdingCode string, id string) error {
	holdingCode = strings.TrimSpace(holdingCode)
	id = strings.TrimSpace(id)

	query := `UPDATE role_permissions SET is_active = false, updated_at = now()
	          WHERE LOWER(holding_code) = LOWER($1) AND (id = $2 OR LOWER(role_code) = LOWER($2))`
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

func (h RolePermissionHttp) SearchRolePermissions(ctx microservice.IContext) error {
	holdingCode := strings.TrimSpace(ctx.UserInfo().HoldingCode)
	if db, err := mypg.PgSqlFastConnect("bcai_projection"); err == nil && db != nil {
		limit := boundedQueryInt(ctx.QueryParam("limit"), 100, 1, 1000)
		offset := boundedQueryInt(ctx.QueryParam("offset"), 0, 0, 1_000_000)
		q := ctx.QueryParam("q")
		return h.searchRolePermissionsPostgres(ctx, db, holdingCode, q, offset, limit)
	}

	mongoCtx, cancel := context.WithTimeout(context.Background(), rolePermissionTimeout)
	defer cancel()
	pst, holdingCode, err := h.managerStore(mongoCtx, ctx)
	if err != nil {
		return respondAccessError(ctx, err)
	}

	filter := activeRolePermissionFilter(holdingCode)
	if query := strings.TrimSpace(ctx.QueryParam("q")); query != "" {
		pattern := primitive.Regex{Pattern: regexp.QuoteMeta(query), Options: "i"}
		filter["$or"] = bson.A{
			bson.M{"rolecode": pattern},
			bson.M{"names.name": pattern},
		}
	}

	limit := boundedQueryInt(ctx.QueryParam("limit"), 100, 1, 1000)
	offset := boundedQueryInt(ctx.QueryParam("offset"), 0, 0, 1_000_000)
	findOptions := options.Find().SetSort(bson.D{{Key: "rolecode", Value: 1}}).SetSkip(int64(offset)).SetLimit(int64(limit))
	var records []rolemodels.RolePermissionDoc
	if err := pst.Find(mongoCtx, rolemodels.RolePermissionDoc{}, filter, &records, findOptions); err != nil {
		return respondInternalError(ctx, err, "โหลดรายการสิทธิ์ตามบทบาทไม่สำเร็จ")
	}
	total, err := pst.Count(mongoCtx, rolemodels.RolePermissionDoc{}, filter)
	if err != nil {
		return respondInternalError(ctx, err, "นับรายการสิทธิ์ตามบทบาทไม่สำเร็จ")
	}

	ctx.Response(http.StatusOK, common.ApiResponse{Success: true, Data: records, Total: total})
	return nil
}

func (h RolePermissionHttp) InfoRolePermission(ctx microservice.IContext) error {
	holdingCode := strings.TrimSpace(ctx.UserInfo().HoldingCode)
	id := strings.TrimSpace(ctx.Param("id"))
	if db, err := mypg.PgSqlFastConnect("bcai_projection"); err == nil && db != nil {
		return h.infoRolePermissionPostgres(ctx, db, holdingCode, id)
	}

	mongoCtx, cancel := context.WithTimeout(context.Background(), rolePermissionTimeout)
	defer cancel()
	pst, holdingCode, err := h.managerStore(mongoCtx, ctx)
	if err != nil {
		return respondAccessError(ctx, err)
	}
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, "รหัสรายการสิทธิ์ไม่ถูกต้อง")
		return err
	}
	record, found, err := findRolePermission(mongoCtx, pst, bson.M{"_id": objID, "holdingcode": holdingCode, "isdeleted": false})
	if err != nil {
		return respondInternalError(ctx, err, "โหลดข้อมูลสิทธิ์ตามบทบาทไม่สำเร็จ")
	}
	if !found {
		err = errors.New("role permission not found")
		ctx.ResponseError(http.StatusNotFound, "ไม่พบรายการสิทธิ์ตามบทบาท")
		return err
	}
	ctx.Response(http.StatusOK, common.ApiResponse{Success: true, Data: record})
	return nil
}

func (h RolePermissionHttp) InfoMyRolePermission(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	if db, err := mypg.PgSqlFastConnect("bcai_projection"); err == nil && db != nil {
		return h.infoMyRolePermissionPostgres(ctx, db, userInfo.HoldingCode, userInfo.Username)
	}

	mongoCtx, cancel := context.WithTimeout(context.Background(), rolePermissionTimeout)
	defer cancel()
	pst := h.ms.MongoPersister(h.cfg.MongoPersisterConfig())
	membership, err := orgpolicy.FindActiveMembership(mongoCtx, pst, userInfo, time.Now().UTC())
	if err != nil {
		return respondAccessError(ctx, err)
	}
	if err := orgpolicy.RequireActiveHolding(mongoCtx, pst, userInfo.HoldingCode); err != nil {
		return respondAccessError(ctx, err)
	}
	roleCode, ok := roleCodeFromRole(membership.Role)
	if !ok {
		err = errors.New("unsupported membership role")
		ctx.ResponseError(http.StatusForbidden, "บทบาทผู้ใช้งานไม่รองรับ")
		return err
	}
	// Effective permissions = union of the role's built-in set and every
	// permission set the member picked (docs/organization.md).
	setCodes := append([]string{roleCode}, membership.PermissionSets...)
	var records []rolemodels.RolePermissionDoc
	err = pst.Find(mongoCtx, rolemodels.RolePermissionDoc{}, bson.M{
		"holdingcode": userInfo.HoldingCode,
		"rolecode":    bson.M{"$in": setCodes},
		"isactive":    true,
		"isdeleted":   false,
	}, &records, options.Find().SetLimit(int64(len(setCodes)+1)))
	if err != nil {
		return respondInternalError(ctx, err, "โหลดสิทธิ์ของผู้ใช้งานไม่สำเร็จ")
	}
	h.ms.Logger.Debug(fmt.Sprintf("InfoMyRolePermission: holding=%s role=%d sets=%v records=%d", userInfo.HoldingCode, membership.Role, setCodes, len(records)))
	permissions := unionPermissions(records)
	// ADMIN/OWNER default to full screen access when nothing is assigned for
	// the role itself (product rule: an admin must reach every screen by
	// default, whatever extra sets they picked). The frontend expands the "*"
	// wildcard to every menu item. USER without any assignment stays fail-closed.
	if (roleCode == "ADMIN" || roleCode == "OWNER") && !hasRoleRecord(records, roleCode) {
		permissions = []string{"*"}
	}
	if len(permissions) == 0 {
		ctx.Response(http.StatusOK, common.ApiResponse{Success: true})
		return nil
	}
	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data: map[string]interface{}{
			"holdingcode":    userInfo.HoldingCode,
			"rolecode":       roleCode,
			"permissionsets": membership.PermissionSets,
			"permissions":    permissions,
			"isactive":       true,
		},
	})
	return nil
}

func hasRoleRecord(records []rolemodels.RolePermissionDoc, roleCode string) bool {
	for _, record := range records {
		if record.RoleCode == roleCode {
			return true
		}
	}
	return false
}

// unionPermissions merges the permission entries of every set, sorted and
// deduplicated; a "*" anywhere collapses the result to ["*"].
func unionPermissions(records []rolemodels.RolePermissionDoc) []string {
	seen := map[string]struct{}{}
	for _, record := range records {
		for _, entry := range record.Permissions {
			if entry == "*" {
				return []string{"*"}
			}
			seen[entry] = struct{}{}
		}
	}
	out := make([]string, 0, len(seen))
	for entry := range seen {
		out = append(out, entry)
	}
	sort.Strings(out)
	return out
}

func (h RolePermissionHttp) CreateRolePermission(ctx microservice.IContext) error {
	holdingCode := strings.TrimSpace(ctx.UserInfo().HoldingCode)
	req, err := readRolePermissionRequest(ctx, true)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}
	if db, err := mypg.PgSqlFastConnect("bcai_projection"); err == nil && db != nil {
		return h.createRolePermissionPostgres(ctx, db, holdingCode, req)
	}

	mongoCtx, cancel := context.WithTimeout(context.Background(), rolePermissionTimeout)
	defer cancel()
	pst, holdingCode, err := h.managerStore(mongoCtx, ctx)
	if err != nil {
		return respondAccessError(ctx, err)
	}
	if err := requireRolePermissionWrite(mongoCtx, pst, ctx, req.RoleCode); err != nil {
		return respondAccessError(ctx, err)
	}
	if duplicate, err := rolePermissionCodeExists(mongoCtx, pst, holdingCode, req.RoleCode, primitive.NilObjectID); err != nil {
		return respondInternalError(ctx, err, "ตรวจสอบรหัสบทบาทซ้ำไม่สำเร็จ")
	} else if duplicate {
		err = fmt.Errorf("rolecode %s already exists", req.RoleCode)
		ctx.ResponseError(http.StatusConflict, "บทบาทนี้มีรายการสิทธิ์อยู่แล้ว")
		return err
	}

	now := time.Now().UTC()
	record := rolemodels.RolePermissionDoc{
		HoldingCode: holdingCode,
		RoleCode:    req.RoleCode,
		Names:       req.Names,
		Permissions: req.Permissions,
		IsActive:    *req.IsActive,
		CreatedAt:   now,
		CreatedBy:   actorID(ctx.UserInfo()),
		IsDeleted:   false,
		Version:     0,
	}
	id, err := pst.Create(mongoCtx, rolemodels.RolePermissionDoc{}, record)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			ctx.ResponseError(http.StatusConflict, "บทบาทนี้มีรายการสิทธิ์อยู่แล้ว")
		} else {
			ctx.ResponseError(http.StatusInternalServerError, "สร้างรายการสิทธิ์ตามบทบาทไม่สำเร็จ")
		}
		return err
	}
	record.ID = id
	ctx.Response(http.StatusCreated, common.ApiResponse{Success: true, ID: id.Hex(), Data: record})
	return nil
}

func (h RolePermissionHttp) UpdateRolePermission(ctx microservice.IContext) error {
	holdingCode := strings.TrimSpace(ctx.UserInfo().HoldingCode)
	id := strings.TrimSpace(ctx.Param("id"))
	req, err := readRolePermissionRequest(ctx, false)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}
	if db, err := mypg.PgSqlFastConnect("bcai_projection"); err == nil && db != nil {
		return h.updateRolePermissionPostgres(ctx, db, holdingCode, id, req)
	}

	mongoCtx, cancel := context.WithTimeout(context.Background(), rolePermissionTimeout)
	defer cancel()
	pst, holdingCode, err := h.managerStore(mongoCtx, ctx)
	if err != nil {
		return respondAccessError(ctx, err)
	}
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, "รหัสรายการสิทธิ์ไม่ถูกต้อง")
		return err
	}
	existing, found, err := findRolePermission(mongoCtx, pst, bson.M{"_id": objID, "holdingcode": holdingCode, "isdeleted": false})
	if err != nil {
		return respondInternalError(ctx, err, "โหลดข้อมูลสิทธิ์ตามบทบาทไม่สำเร็จ")
	}
	if !found {
		err = errors.New("role permission not found")
		ctx.ResponseError(http.StatusNotFound, "ไม่พบรายการสิทธิ์ตามบทบาท")
		return err
	}
	if err := requireRolePermissionWrite(mongoCtx, pst, ctx, existing.RoleCode, req.RoleCode); err != nil {
		return respondAccessError(ctx, err)
	}
	if duplicate, err := rolePermissionCodeExists(mongoCtx, pst, holdingCode, req.RoleCode, objID); err != nil {
		return respondInternalError(ctx, err, "ตรวจสอบรหัสบทบาทซ้ำไม่สำเร็จ")
	} else if duplicate {
		err = fmt.Errorf("rolecode %s already exists", req.RoleCode)
		ctx.ResponseError(http.StatusConflict, "บทบาทนี้มีรายการสิทธิ์อยู่แล้ว")
		return err
	}

	collection, err := pst.Exec(mongoCtx, rolemodels.RolePermissionDoc{})
	if err != nil {
		return respondInternalError(ctx, err, "เปิดแหล่งข้อมูลสิทธิ์ตามบทบาทไม่สำเร็จ")
	}
	now := time.Now().UTC()
	result, err := collection.UpdateOne(mongoCtx, bson.M{
		"_id":         objID,
		"holdingcode": holdingCode,
		"isdeleted":   false,
		"__v":         *req.Version,
	}, bson.M{
		"$set": bson.M{
			"rolecode":    req.RoleCode,
			"names":       req.Names,
			"permissions": req.Permissions,
			"isactive":    *req.IsActive,
			"updatedat":   now,
			"updatedby":   actorID(ctx.UserInfo()),
		},
		"$inc": bson.M{"__v": 1},
	})
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			ctx.ResponseError(http.StatusConflict, "บทบาทนี้มีรายการสิทธิ์อยู่แล้ว")
		} else {
			ctx.ResponseError(http.StatusInternalServerError, "บันทึกรายการสิทธิ์ตามบทบาทไม่สำเร็จ")
		}
		return err
	}
	if result.MatchedCount != 1 {
		err = errors.New("role permission was changed by another request")
		ctx.ResponseError(http.StatusConflict, "ข้อมูลถูกแก้ไขจากหน้าจออื่น กรุณาโหลดใหม่")
		return err
	}
	record, _, err := findRolePermission(mongoCtx, pst, bson.M{"_id": objID, "holdingcode": holdingCode, "isdeleted": false})
	if err != nil {
		return respondInternalError(ctx, err, "โหลดข้อมูลสิทธิ์หลังบันทึกไม่สำเร็จ")
	}
	ctx.Response(http.StatusOK, common.ApiResponse{Success: true, ID: objID.Hex(), Data: record})
	return nil
}

func (h RolePermissionHttp) DeleteRolePermission(ctx microservice.IContext) error {
	holdingCode := strings.TrimSpace(ctx.UserInfo().HoldingCode)
	id := strings.TrimSpace(ctx.Param("id"))
	if db, err := mypg.PgSqlFastConnect("bcai_projection"); err == nil && db != nil {
		return h.deleteRolePermissionPostgres(ctx, db, holdingCode, id)
	}

	mongoCtx, cancel := context.WithTimeout(context.Background(), rolePermissionTimeout)
	defer cancel()
	pst, holdingCode, err := h.managerStore(mongoCtx, ctx)
	if err != nil {
		return respondAccessError(ctx, err)
	}
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, "รหัสรายการสิทธิ์ไม่ถูกต้อง")
		return err
	}
	expectedVersion, err := strconv.ParseInt(strings.TrimSpace(ctx.QueryParam("__v")), 10, 64)
	if err != nil || expectedVersion < 0 {
		err = errors.New("valid __v is required")
		ctx.ResponseError(http.StatusBadRequest, "ต้องระบุ __v ที่ถูกต้องเพื่อลบข้อมูล")
		return err
	}
	record, found, err := findRolePermission(mongoCtx, pst, bson.M{"_id": objID, "holdingcode": holdingCode, "isdeleted": false})
	if err != nil {
		return respondInternalError(ctx, err, "โหลดข้อมูลสิทธิ์ตามบทบาทไม่สำเร็จ")
	}
	if !found {
		err = errors.New("role permission not found")
		ctx.ResponseError(http.StatusNotFound, "ไม่พบรายการสิทธิ์ตามบทบาท")
		return err
	}
	if err := requireRolePermissionWrite(mongoCtx, pst, ctx, record.RoleCode); err != nil {
		return respondAccessError(ctx, err)
	}

	collection, err := pst.Exec(mongoCtx, rolemodels.RolePermissionDoc{})
	if err != nil {
		return respondInternalError(ctx, err, "เปิดแหล่งข้อมูลสิทธิ์ตามบทบาทไม่สำเร็จ")
	}
	now := time.Now().UTC()
	result, err := collection.UpdateOne(mongoCtx, bson.M{
		"_id":         objID,
		"holdingcode": holdingCode,
		"isdeleted":   false,
		"__v":         expectedVersion,
	}, bson.M{
		"$set": bson.M{
			"isdeleted": true,
			"deletedat": now,
			"deletedby": actorID(ctx.UserInfo()),
			"updatedat": now,
			"updatedby": actorID(ctx.UserInfo()),
		},
		"$inc": bson.M{"__v": 1},
	})
	if err != nil {
		return respondInternalError(ctx, err, "ลบรายการสิทธิ์ตามบทบาทไม่สำเร็จ")
	}
	if result.MatchedCount != 1 {
		err = errors.New("role permission was changed by another request")
		ctx.ResponseError(http.StatusConflict, "ข้อมูลถูกแก้ไขจากหน้าจออื่น กรุณาโหลดใหม่")
		return err
	}
	ctx.Response(http.StatusOK, common.ApiResponse{Success: true, ID: objID.Hex()})
	return nil
}

func (h RolePermissionHttp) managerStore(ctx context.Context, requestContext microservice.IContext) (microservice.IPersisterMongo, string, error) {
	pst := h.ms.MongoPersister(h.cfg.MongoPersisterConfig())
	holdingCode := strings.TrimSpace(requestContext.UserInfo().HoldingCode)
	if _, err := orgpolicy.FindActiveHoldingManager(ctx, pst, requestContext.UserInfo(), time.Now().UTC()); err != nil {
		return pst, holdingCode, err
	}
	if err := orgpolicy.RequireActiveHolding(ctx, pst, holdingCode); err != nil {
		return pst, holdingCode, err
	}
	return pst, holdingCode, nil
}

func requireRolePermissionWrite(
	ctx context.Context,
	pst microservice.IPersisterMongo,
	requestContext microservice.IContext,
	roleCodes ...string,
) error {
	for _, roleCode := range roleCodes {
		if strings.EqualFold(strings.TrimSpace(roleCode), "USER") {
			continue
		}
		_, err := orgpolicy.FindActiveHoldingOwner(ctx, pst, requestContext.UserInfo(), time.Now().UTC())
		return err
	}
	return nil
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

func activeRolePermissionFilter(holdingCode string) bson.M {
	return bson.M{"holdingcode": strings.TrimSpace(holdingCode), "isdeleted": false}
}

func rolePermissionCodeExists(ctx context.Context, pst microservice.IPersisterMongo, holdingCode, roleCode string, exceptID primitive.ObjectID) (bool, error) {
	filter := activeRolePermissionFilter(holdingCode)
	filter["rolecode"] = roleCode
	if exceptID != primitive.NilObjectID {
		filter["_id"] = bson.M{"$ne": exceptID}
	}
	count, err := pst.Count(ctx, rolemodels.RolePermissionDoc{}, filter)
	return count > 0, err
}

func findRolePermission(ctx context.Context, pst microservice.IPersisterMongo, filter bson.M) (rolemodels.RolePermissionDoc, bool, error) {
	var record rolemodels.RolePermissionDoc
	if err := pst.FindOne(ctx, rolemodels.RolePermissionDoc{}, filter, &record); err != nil {
		return record, false, err
	}
	return record, record.ID != primitive.NilObjectID, nil
}

func roleCodeFromRole(role uint8) (string, bool) {
	for code, value := range allowedRoleCodes {
		if value == role {
			return code, true
		}
	}
	return "", false
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

func actorID(userInfo msmodels.UserInfo) string {
	if uid := strings.TrimSpace(userInfo.UID); uid != "" {
		return uid
	}
	return strings.TrimSpace(userInfo.Username)
}

func respondAccessError(ctx microservice.IContext, err error) error {
	status := http.StatusInternalServerError
	message := "ไม่สามารถตรวจสอบสิทธิ์ได้"
	if errors.Is(err, orgpolicy.ErrHoldingOwnerRequired) {
		status = http.StatusForbidden
		message = "ต้องเป็น OWNER ของ Holding ที่กำลังใช้งาน"
	} else if errors.Is(err, orgpolicy.ErrActiveMembershipRequired) ||
		errors.Is(err, orgpolicy.ErrHoldingManagerRequired) ||
		errors.Is(err, orgpolicy.ErrActiveHoldingRequired) {
		status = http.StatusForbidden
		message = "ต้องเป็น OWNER หรือ ADMIN ของ Holding ที่กำลังใช้งาน"
	}
	ctx.ResponseError(status, message)
	return err
}

func respondInternalError(ctx microservice.IContext, err error, message string) error {
	ctx.ResponseError(http.StatusInternalServerError, message)
	return err
}
