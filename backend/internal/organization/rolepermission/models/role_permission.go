package models

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const collectionName = "role_permission"

type LocalizedName struct {
	Code string `json:"code" bson:"code"`
	Name string `json:"name" bson:"name"`
}

type RolePermissionDoc struct {
	ID          primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	HoldingCode string             `json:"holdingcode" bson:"holdingcode"`
	RoleCode    string             `json:"rolecode" bson:"rolecode"`
	Names       []LocalizedName    `json:"names" bson:"names"`
	Permissions []string           `json:"permissions" bson:"permissions"`
	IsActive    bool               `json:"isactive" bson:"isactive"`
	CreatedAt   time.Time          `json:"createdat,omitempty" bson:"createdat,omitempty"`
	CreatedBy   string             `json:"createdby,omitempty" bson:"createdby,omitempty"`
	UpdatedAt   time.Time          `json:"updatedat,omitempty" bson:"updatedat,omitempty"`
	UpdatedBy   string             `json:"updatedby,omitempty" bson:"updatedby,omitempty"`
	IsDeleted   bool               `json:"isdeleted" bson:"isdeleted"`
	DeletedAt   *time.Time         `json:"deletedat,omitempty" bson:"deletedat,omitempty"`
	DeletedBy   string             `json:"deletedby,omitempty" bson:"deletedby,omitempty"`
	Version     int64              `json:"__v" bson:"__v"`
}

func (RolePermissionDoc) CollectionName() string { return collectionName }

type RolePermissionRequest struct {
	RoleCode    string          `json:"rolecode"`
	Names       []LocalizedName `json:"names"`
	Permissions []string        `json:"permissions"`
	IsActive    *bool           `json:"isactive"`
	Version     *int64          `json:"__v"`
}

func NormalizeRequest(req *RolePermissionRequest) error {
	req.RoleCode = strings.ToUpper(strings.TrimSpace(req.RoleCode))
	switch req.RoleCode {
	case "USER", "ADMIN", "OWNER":
	default:
		return errors.New("rolecode ต้องเป็น USER, ADMIN หรือ OWNER")
	}

	seenNames := make(map[string]struct{}, len(req.Names))
	names := make([]LocalizedName, 0, len(req.Names))
	for _, item := range req.Names {
		code := strings.ToLower(strings.TrimSpace(item.Code))
		name := strings.TrimSpace(item.Name)
		if code == "" || name == "" {
			continue
		}
		if _, duplicate := seenNames[code]; duplicate {
			return fmt.Errorf("ชื่อภาษารหัส %s ซ้ำ", code)
		}
		seenNames[code] = struct{}{}
		names = append(names, LocalizedName{Code: code, Name: name})
	}
	if len(names) == 0 {
		return errors.New("ต้องระบุชื่อบทบาทอย่างน้อย 1 ภาษา")
	}
	req.Names = names

	seenPermissions := make(map[string]struct{}, len(req.Permissions))
	permissions := make([]string, 0, len(req.Permissions))
	for _, permission := range req.Permissions {
		permission = strings.TrimSpace(permission)
		if permission == "" {
			continue
		}
		if !isValidPermissionEntry(permission) {
			return fmt.Errorf("รูปแบบสิทธิ์ %q ไม่ถูกต้อง (ต้องเป็น รหัสจอ หรือ รหัสจอ:create|update|delete)", permission)
		}
		if _, duplicate := seenPermissions[permission]; duplicate {
			continue
		}
		seenPermissions[permission] = struct{}{}
		permissions = append(permissions, permission)
	}
	sort.Strings(permissions)
	req.Permissions = permissions
	return nil
}

// Permission entries: "<screen>" grants entry (เข้า) to a screen; "<screen>:create",
// "<screen>:update", "<screen>:delete" grant the matching action. "*" = everything.
var permissionActions = map[string]struct{}{"create": {}, "update": {}, "delete": {}}

func isValidPermissionEntry(entry string) bool {
	if entry == "*" {
		return true
	}
	screen, action, hasAction := strings.Cut(entry, ":")
	if screen == "" || strings.ContainsAny(screen, " 	") {
		return false
	}
	if !hasAction {
		return true
	}
	_, ok := permissionActions[action]
	return ok
}
