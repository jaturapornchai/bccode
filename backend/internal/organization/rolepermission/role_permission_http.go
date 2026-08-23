package rolepermission

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	authmodels "smlcloudplatform/internal/authentication/models"
	"smlcloudplatform/internal/config"
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

func (h RolePermissionHttp) SearchRolePermissions(ctx microservice.IContext) error {
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
	mongoCtx, cancel := context.WithTimeout(context.Background(), rolePermissionTimeout)
	defer cancel()
	pst, holdingCode, err := h.managerStore(mongoCtx, ctx)
	if err != nil {
		return respondAccessError(ctx, err)
	}
	id, err := primitive.ObjectIDFromHex(strings.TrimSpace(ctx.Param("id")))
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, "รหัสรายการสิทธิ์ไม่ถูกต้อง")
		return err
	}
	record, found, err := findRolePermission(mongoCtx, pst, bson.M{"_id": id, "holdingcode": holdingCode, "isdeleted": false})
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
	mongoCtx, cancel := context.WithTimeout(context.Background(), rolePermissionTimeout)
	defer cancel()
	pst := h.ms.MongoPersister(h.cfg.MongoPersisterConfig())
	userInfo := ctx.UserInfo()
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
	var records []rolemodels.RolePermissionDoc
	err = pst.Find(mongoCtx, rolemodels.RolePermissionDoc{}, bson.M{
		"holdingcode": userInfo.HoldingCode,
		"rolecode":    roleCode,
		"isactive":    true,
		"isdeleted":   false,
	}, &records, options.Find().SetLimit(2))
	if err != nil {
		return respondInternalError(ctx, err, "โหลดสิทธิ์ของผู้ใช้งานไม่สำเร็จ")
	}
	if len(records) == 0 {
		ctx.Response(http.StatusOK, common.ApiResponse{Success: true})
		return nil
	}
	if len(records) > 1 {
		err = errors.New("duplicate active role permissions")
		ctx.ResponseError(http.StatusInternalServerError, "พบรายการสิทธิ์ซ้ำสำหรับบทบาทนี้ กรุณาให้ผู้ดูแลแก้ข้อมูล")
		return err
	}
	ctx.Response(http.StatusOK, common.ApiResponse{Success: true, Data: records[0]})
	return nil
}

func (h RolePermissionHttp) CreateRolePermission(ctx microservice.IContext) error {
	mongoCtx, cancel := context.WithTimeout(context.Background(), rolePermissionTimeout)
	defer cancel()
	pst, holdingCode, err := h.managerStore(mongoCtx, ctx)
	if err != nil {
		return respondAccessError(ctx, err)
	}
	req, err := readRolePermissionRequest(ctx, true)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
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
	mongoCtx, cancel := context.WithTimeout(context.Background(), rolePermissionTimeout)
	defer cancel()
	pst, holdingCode, err := h.managerStore(mongoCtx, ctx)
	if err != nil {
		return respondAccessError(ctx, err)
	}
	id, err := primitive.ObjectIDFromHex(strings.TrimSpace(ctx.Param("id")))
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, "รหัสรายการสิทธิ์ไม่ถูกต้อง")
		return err
	}
	req, err := readRolePermissionRequest(ctx, false)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}
	existing, found, err := findRolePermission(mongoCtx, pst, bson.M{"_id": id, "holdingcode": holdingCode, "isdeleted": false})
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
	if duplicate, err := rolePermissionCodeExists(mongoCtx, pst, holdingCode, req.RoleCode, id); err != nil {
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
		"_id":         id,
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
	record, _, err := findRolePermission(mongoCtx, pst, bson.M{"_id": id, "holdingcode": holdingCode, "isdeleted": false})
	if err != nil {
		return respondInternalError(ctx, err, "โหลดข้อมูลสิทธิ์หลังบันทึกไม่สำเร็จ")
	}
	ctx.Response(http.StatusOK, common.ApiResponse{Success: true, ID: id.Hex(), Data: record})
	return nil
}

func (h RolePermissionHttp) DeleteRolePermission(ctx microservice.IContext) error {
	mongoCtx, cancel := context.WithTimeout(context.Background(), rolePermissionTimeout)
	defer cancel()
	pst, holdingCode, err := h.managerStore(mongoCtx, ctx)
	if err != nil {
		return respondAccessError(ctx, err)
	}
	id, err := primitive.ObjectIDFromHex(strings.TrimSpace(ctx.Param("id")))
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
	record, found, err := findRolePermission(mongoCtx, pst, bson.M{"_id": id, "holdingcode": holdingCode, "isdeleted": false})
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
		"_id":         id,
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
	ctx.Response(http.StatusOK, common.ApiResponse{Success: true, ID: id.Hex()})
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
