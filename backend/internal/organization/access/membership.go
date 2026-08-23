package access

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	authmodels "smlcloudplatform/internal/authentication/models"
	micromodels "smlcloudplatform/pkg/microservice/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrActiveMembershipRequired = errors.New("active membership required")
	ErrHoldingManagerRequired   = errors.New("Holding OWNER or ADMIN required")
	ErrHoldingOwnerRequired     = errors.New("Holding OWNER required")
)

type MembershipFinder interface {
	FindOne(context.Context, interface{}, interface{}, interface{}, ...*options.FindOneOptions) error
}

func FindActiveMembership(ctx context.Context, finder MembershipFinder, userInfo micromodels.UserInfo, now time.Time) (authmodels.ShopUser, error) {
	holdingCode := strings.TrimSpace(userInfo.HoldingCode)
	userUID := strings.TrimSpace(userInfo.UID)
	if holdingCode == "" || userUID == "" {
		return authmodels.ShopUser{}, ErrActiveMembershipRequired
	}

	var membership authmodels.ShopUser
	if err := findMembership(ctx, finder, bson.M{"holdingcode": holdingCode, "useruid": userUID, "isdeleted": false}, &membership); err != nil {
		return authmodels.ShopUser{}, err
	}
	if membership.ID == primitive.NilObjectID || membership.IsDeleted || membership.IsAccessDisabled ||
		(!membership.AccessExpiryDate.IsZero() && !now.Before(membership.AccessExpiryDate)) {
		return authmodels.ShopUser{}, ErrActiveMembershipRequired
	}
	return membership, nil
}

func FindActiveHoldingManager(ctx context.Context, finder MembershipFinder, userInfo micromodels.UserInfo, now time.Time) (authmodels.ShopUser, error) {
	membership, err := FindActiveMembership(ctx, finder, userInfo, now)
	if err != nil {
		return authmodels.ShopUser{}, err
	}
	if membership.Role != authmodels.ROLE_OWNER && membership.Role != authmodels.ROLE_ADMIN {
		return authmodels.ShopUser{}, ErrHoldingManagerRequired
	}
	return membership, nil
}

func FindActiveHoldingOwner(ctx context.Context, finder MembershipFinder, userInfo micromodels.UserInfo, now time.Time) (authmodels.ShopUser, error) {
	membership, err := FindActiveMembership(ctx, finder, userInfo, now)
	if err != nil {
		return authmodels.ShopUser{}, err
	}
	if membership.Role != authmodels.ROLE_OWNER {
		return authmodels.ShopUser{}, ErrHoldingOwnerRequired
	}
	return membership, nil
}

func findMembership(ctx context.Context, finder MembershipFinder, filter bson.M, membership *authmodels.ShopUser) error {
	if err := finder.FindOne(ctx, &authmodels.ShopUser{}, filter, membership); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil
		}
		return fmt.Errorf("load membership: %w", err)
	}
	return nil
}
