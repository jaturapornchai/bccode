package organization

import (
	"context"
	"net/mail"
	"strings"
	"time"

	authmodels "smlcloudplatform/internal/authentication/models"
	"smlcloudplatform/pkg/apperr"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
		return apperr.ErrInternal.WithWrap(err)
	}
	if user.ID == primitive.NilObjectID {
		return apperr.ErrUnauthorized.WithMessage("authenticated user no longer exists").WithThaiMessage("ไม่พบบัญชีผู้ใช้ที่เข้าสู่ระบบ")
	}
	if policyErr := validateCreatorPolicy(user.Email, user.DisabledAt, authmodels.ROLE_USER, false); policyErr != nil {
		return policyErr
	}
	if !requireAdmin {
		return nil
	}

	holdingCode := strings.TrimSpace(userInfo.HoldingCode)
	if holdingCode == "" {
		return apperr.ErrForbidden.WithMessage("a Holding must be selected").WithThaiMessage("กรุณาเลือก Holding ก่อนสร้างข้อมูลโครงสร้างองค์กร")
	}

	identityFilter := bson.M{"username": user.Username}
	if uid := strings.TrimSpace(user.UID); uid != "" {
		identityFilter = bson.M{"$or": bson.A{bson.M{"useruid": uid}, bson.M{"username": user.Username}}}
	}
	membershipFilter := bson.M{
		"$and": bson.A{
			bson.M{"holdingcode": holdingCode},
			identityFilter,
			bson.M{"role": bson.M{"$in": bson.A{authmodels.ROLE_ADMIN, authmodels.ROLE_OWNER}}},
			bson.M{"isaccessdisabled": bson.M{"$ne": true}},
		},
	}
	membership := &authmodels.ShopUser{}
	if err := pst.FindOne(ctx, &authmodels.ShopUser{}, membershipFilter, membership); err != nil {
		return apperr.ErrInternal.WithWrap(err)
	}
	if membership.ID == primitive.NilObjectID || (!membership.AccessExpiryDate.IsZero() && time.Now().After(membership.AccessExpiryDate)) {
		return apperr.ErrForbidden.WithMessage("Holding OWNER or ADMIN permission is required").WithThaiMessage("เฉพาะ OWNER หรือ ADMIN ของ Holding เท่านั้นที่สร้างได้")
	}
	return validateCreatorPolicy(user.Email, user.DisabledAt, membership.Role, true)
}

func validateCreatorPolicy(email string, disabledAt time.Time, role authmodels.UserRole, requireAdmin bool) *apperr.AppError {
	if !disabledAt.IsZero() {
		return apperr.ErrForbidden.WithMessage("account is disabled").WithThaiMessage("บัญชีผู้ใช้ถูกปิดใช้งาน")
	}
	email = strings.TrimSpace(email)
	address, err := mail.ParseAddress(email)
	if err != nil || !strings.EqualFold(address.Address, email) {
		return apperr.New("EMAIL_REQUIRED", 403, "a valid linked email is required", "ต้องเชื่อมอีเมลที่ถูกต้องก่อนสร้าง Holding, Company หรือ Branch")
	}
	if requireAdmin && role != authmodels.ROLE_ADMIN && role != authmodels.ROLE_OWNER {
		return apperr.ErrForbidden.WithMessage("Holding OWNER or ADMIN permission is required").WithThaiMessage("เฉพาะ OWNER หรือ ADMIN ของ Holding เท่านั้นที่สร้างได้")
	}
	return nil
}
