package handlers

// ตรวจสิทธิ์ "/product/listing" ฝั่ง server สำหรับ Product Listing API v2
//
// ทำซ้ำตรรกะของ InfoMyRolePermission (internal/organization/rolepermission) เพราะฟังก์ชันเหล่านั้นเป็น unexported
// และใช้ IPersisterMongo ส่วน goapi มีแค่ *mongo.Database — กติกาเดียวกัน:
//   สมาชิก (shopusers) ต้องยังใช้งานได้ → รวมสิทธิ์ของ role + ชุดสิทธิ์ที่เลือก → "*" หรือ key ตรงกัน = ผ่าน
//   ADMIN/OWNER ที่ไม่มี record ของ role ตัวเอง = ผ่านทุกจอ; USER ไม่มี record = ไม่ผ่าน (fail-closed)

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	msmodels "smlcloudplatform/pkg/microservice/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	productV2RoleUser   uint8 = 0
	productV2RoleAdmin  uint8 = 1
	productV2RoleOwner  uint8 = 2
	productV2RoleSystem uint8 = 255
)

var productV2RoleCodes = map[uint8]string{
	productV2RoleUser:  "USER",
	productV2RoleAdmin: "ADMIN",
	productV2RoleOwner: "OWNER",
}

type productV2Membership struct {
	Role             uint8     `bson:"role"`
	PermissionSets   []string  `bson:"permissionsets"`
	IsAccessDisabled bool      `bson:"isaccessdisabled"`
	AccessExpiryDate time.Time `bson:"accessexpirydate"`
}

type productV2PermissionRecord struct {
	RoleCode    string   `bson:"rolecode"`
	Permissions []string `bson:"permissions"`
}

// productV2CheckPermission — จุดเดียวที่ handler เรียก; เป็นตัวแปรเพื่อให้ unit test แทนที่ได้ (ไม่ต้องมี Mongo)
var productV2CheckPermission = loadProductV2Permission

func loadProductV2Permission(ctx context.Context, userInfo msmodels.UserInfo, key string) (bool, error) {
	if userInfo.Role == productV2RoleSystem {
		return true, nil
	}
	_, db := GetAtlasConnection()
	if db == nil {
		return false, errors.New("mongo database not ready")
	}
	holdingCode := strings.TrimSpace(userInfo.HoldingCode)
	userUID := strings.TrimSpace(userInfo.UID)
	if holdingCode == "" || userUID == "" {
		return false, nil
	}

	var membership productV2Membership
	err := db.Collection("shopusers").FindOne(ctx, bson.M{
		"holdingcode": holdingCode,
		"useruid":     userUID,
		"isdeleted":   false,
	}).Decode(&membership)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, nil
		}
		return false, fmt.Errorf("load membership: %w", err)
	}
	now := time.Now().UTC()
	if membership.IsAccessDisabled || (!membership.AccessExpiryDate.IsZero() && !now.Before(membership.AccessExpiryDate)) {
		return false, nil
	}
	if membership.Role == productV2RoleSystem {
		return true, nil
	}
	roleCode, ok := productV2RoleCodes[membership.Role]
	if !ok {
		return false, nil
	}

	setCodes := append([]string{roleCode}, membership.PermissionSets...)
	cursor, err := db.Collection("role_permission").Find(ctx, bson.M{
		"holdingcode": holdingCode,
		"rolecode":    bson.M{"$in": setCodes},
		"isactive":    true,
		"isdeleted":   false,
	}, options.Find().SetProjection(bson.M{"rolecode": 1, "permissions": 1}).SetLimit(int64(len(setCodes)+1)))
	if err != nil {
		return false, fmt.Errorf("load role permission: %w", err)
	}
	var records []productV2PermissionRecord
	if err := cursor.All(ctx, &records); err != nil {
		return false, fmt.Errorf("decode role permission: %w", err)
	}
	return hasProductV2Permission(records, roleCode, key), nil
}

// hasProductV2Permission — ตรรกะล้วน (ทดสอบได้โดยไม่ต้องมี Mongo)
func hasProductV2Permission(records []productV2PermissionRecord, roleCode, key string) bool {
	hasOwnRecord := false
	for _, record := range records {
		if record.RoleCode == roleCode {
			hasOwnRecord = true
		}
		for _, entry := range record.Permissions {
			if entry == "*" || entry == key {
				return true
			}
		}
	}
	if (roleCode == "ADMIN" || roleCode == "OWNER") && !hasOwnRecord {
		return true
	}
	return false
}

// productV2CheckAnyPermission — ผ่านถ้ามีสิทธิ์ใดสิทธิ์หนึ่งในรายการ (ใช้กับสิทธิ์ที่มีได้หลายชื่อ)
func productV2CheckAnyPermission(ctx context.Context, userInfo msmodels.UserInfo, keys []string) (bool, error) {
	for _, key := range keys {
		allowed, err := productV2CheckPermission(ctx, userInfo, key)
		if err != nil {
			return false, err
		}
		if allowed {
			return true, nil
		}
	}
	return false, nil
}
