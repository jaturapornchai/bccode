package organization

import (
	"context"
	"errors"
	"strings"
	"time"

	authmodels "smlcloudplatform/internal/authentication/models"
	"smlcloudplatform/pkg/apperr"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// RequireEmailedAccount authorizes the root Holding bootstrap. A Holding has no
// scoped role until after it exists; its creator becomes OWNER in ShopService.
func RequireEmailedAccount(pst microservice.IPersisterMongo, userInfo micromodels.UserInfo) *apperr.AppError {
	return requireOrganizationCreator(pst, userInfo, false)
}

// RequireHoldingAdmin authorizes Company/Branch creation from the current Holding.
func RequireHoldingAdmin(pst microservice.IPersisterMongo, userInfo micromodels.UserInfo) *apperr.AppError {
	return requireOrganizationCreator(pst, userInfo, true)
}

func requireOrganizationCreator(pst microservice.IPersisterMongo, userInfo micromodels.UserInfo, requireAdmin bool) *apperr.AppError {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user := &authmodels.UserDoc{}
	userFilter := bson.M{"username": strings.TrimSpace(userInfo.Username)}
	if uid := strings.TrimSpace(userInfo.UID); uid != "" {
		userFilter = bson.M{"uid": uid}
	}
	if err := pst.FindOne(ctx, &authmodels.UserDoc{}, userFilter, user); err != nil {
		return creatorFindError(err, apperr.ErrUnauthorized.WithMessage("authenticated user no longer exists").WithThaiMessage("ไม่พบบัญชีผู้ใช้ที่เข้าสู่ระบบ"))
	}
	if user.ID == primitive.NilObjectID {
		return apperr.ErrUnauthorized.WithMessage("authenticated user no longer exists").WithThaiMessage("ไม่พบบัญชีผู้ใช้ที่เข้าสู่ระบบ")
	}
	if policyErr := validateCreatorPolicy(user.DisabledAt, authmodels.ROLE_USER, false); policyErr != nil {
		return policyErr
	}
	identity := &authmodels.GoogleIdentity{}
	if err := pst.FindOne(ctx, identity, bson.M{"useruid": user.UID, "isactive": true, "revokedat": nil}, identity); err != nil {
		return creatorFindError(err, apperr.ErrForbidden.WithMessage("an active verified Google identity is required").WithThaiMessage("ต้องเชื่อม Google Identity ที่ยืนยันแล้วก่อนสร้างองค์กร"))
	}
	if !isActiveGoogleIdentity(*identity) {
		return apperr.ErrForbidden.WithMessage("an active verified Google identity is required").WithThaiMessage("ต้องเชื่อม Google Identity ที่ยืนยันแล้วก่อนสร้างองค์กร")
	}
	if !requireAdmin {
		return nil
	}

	holdingCode := strings.TrimSpace(userInfo.HoldingCode)
	if holdingCode == "" {
		return apperr.ErrForbidden.WithMessage("a Holding must be selected").WithThaiMessage("กรุณาเลือก Holding ก่อนสร้างข้อมูลโครงสร้างองค์กร")
	}

	userUID := strings.TrimSpace(user.UID)
	if userUID == "" {
		return apperr.ErrForbidden.WithMessage("a stable user identity is required").WithThaiMessage("ไม่พบรหัสผู้ใช้ถาวรสำหรับตรวจสิทธิ์ Holding")
	}
	membershipFilter := holdingAdminMembershipFilter(holdingCode, userUID)
	membership := &authmodels.ShopUser{}
	if err := pst.FindOne(ctx, &authmodels.ShopUser{}, membershipFilter, membership); err != nil {
		return creatorFindError(err, apperr.ErrForbidden.WithMessage("Holding OWNER or ADMIN permission is required").WithThaiMessage("เฉพาะ OWNER หรือ ADMIN ของ Holding เท่านั้นที่สร้างได้"))
	}
	if membership.ID == primitive.NilObjectID || (!membership.AccessExpiryDate.IsZero() && time.Now().After(membership.AccessExpiryDate)) {
		return apperr.ErrForbidden.WithMessage("Holding OWNER or ADMIN permission is required").WithThaiMessage("เฉพาะ OWNER หรือ ADMIN ของ Holding เท่านั้นที่สร้างได้")
	}
	return validateCreatorPolicy(user.DisabledAt, membership.Role, true)
}

func holdingAdminMembershipFilter(holdingCode, userUID string) bson.M {
	return bson.M{
		"holdingcode":      holdingCode,
		"useruid":          userUID,
		"role":             bson.M{"$in": bson.A{authmodels.ROLE_ADMIN, authmodels.ROLE_OWNER}},
		"isaccessdisabled": bson.M{"$ne": true},
		"isdeleted":        false,
	}
}

func isActiveGoogleIdentity(identity authmodels.GoogleIdentity) bool {
	return identity.ID != primitive.NilObjectID &&
		identity.IsActive &&
		identity.RevokedAt == nil &&
		strings.TrimSpace(identity.Issuer) != "" &&
		strings.TrimSpace(identity.Subject) != ""
}

func creatorFindError(err error, notFound *apperr.AppError) *apperr.AppError {
	if errors.Is(err, mongo.ErrNoDocuments) {
		return notFound
	}
	return apperr.ErrInternal.WithWrap(err)
}

func validateCreatorPolicy(disabledAt time.Time, role authmodels.UserRole, requireAdmin bool) *apperr.AppError {
	if !disabledAt.IsZero() {
		return apperr.ErrForbidden.WithMessage("account is disabled").WithThaiMessage("บัญชีผู้ใช้ถูกปิดใช้งาน")
	}
	if requireAdmin && role != authmodels.ROLE_ADMIN && role != authmodels.ROLE_OWNER {
		return apperr.ErrForbidden.WithMessage("Holding OWNER or ADMIN permission is required").WithThaiMessage("เฉพาะ OWNER หรือ ADMIN ของ Holding เท่านั้นที่สร้างได้")
	}
	return nil
}
