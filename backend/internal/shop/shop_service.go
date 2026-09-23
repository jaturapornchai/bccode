package shop

import (
	"context"
	"errors"
	"strings"
	"time"

	auth_model "smlcloudplatform/internal/authentication/models"
	"smlcloudplatform/internal/centraldb"
	common "smlcloudplatform/internal/models"
	orgaccess "smlcloudplatform/internal/organization"
	"smlcloudplatform/internal/shop/models"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/apperr"
)

type IShopService interface {
	CreateShop(userUID string, username string, shop models.Shop) (string, error)
	UpdateShop(holdingCode string, username string, shop models.Shop) error
	ChangeShopStatus(holdingCode string, actorUID string, username string, expectedActive bool, active bool, reason string) error
	InfoShop(holdingCode string) (models.ShopInfo, error)
}

type ShopService struct {
	shopRepo     IShopRepository
	shopUserRepo IShopUserRepository
	timeNow      func() time.Time
}

var ErrHoldingCodeRenameUnavailable = errors.New("holdingcode change requires atomic rename workflow")

func NewShopService(shopRepo IShopRepository, shopUserRepo IShopUserRepository, timeNow func() time.Time) ShopService {
	return ShopService{
		shopRepo:     shopRepo,
		shopUserRepo: shopUserRepo,
		timeNow:      timeNow,
	}
}

// CreateShop creates a Holding and makes the creator its OWNER in one transaction.
// In PostgreSQL the Holding identity is its code, which is also the returned UID.
func (svc ShopService) CreateShop(userUID string, username string, doc models.Shop) (string, error) {
	holdingCode, err := utils.NormalizeHoldingCode(doc.HoldingCode)
	if err != nil {
		return "", err
	}
	if holdingCode == "" {
		return "", errors.New("holdingcode invalid")
	}
	if !hasValidHoldingName(doc.Names) {
		return "", errors.New("holding name is required")
	}
	userUID = strings.TrimSpace(userUID)
	if userUID == "" {
		return "", errors.New("stable User identity is required")
	}
	if _, findErr := svc.shopRepo.FindByHoldingCode(context.Background(), holdingCode); findErr == nil {
		return "", apperr.DuplicateCode("holdingcode", holdingCode)
	}

	now := svc.timeNow().UTC()
	dataDoc := models.ShopDoc{HoldingUID: holdingCode}
	dataDoc.Shop = doc
	dataDoc.HoldingCode = holdingCode
	dataDoc.IsActive = true
	dataDoc.GuidFixed = holdingCode
	dataDoc.CreatedBy = userUID
	dataDoc.CreatedAt = now

	err = svc.shopRepo.Transaction(context.Background(), func(txCtx context.Context) error {
		if err := svc.shopRepo.Create(txCtx, dataDoc); err != nil {
			return err
		}
		if err := svc.shopUserRepo.SaveStable(txCtx, holdingCode, userUID, auth_model.ROLE_OWNER); err != nil {
			return err
		}
		return svc.shopRepo.RecordAudit(txCtx, orgaccess.Audit{
			HoldingCode: holdingCode, Action: "holding.created", TargetType: "holding", TargetCode: holdingCode,
			ActorUID: userUID, After: map[string]interface{}{"holdingcode": holdingCode, "isactive": true}, OccurredAt: now,
		})
	})
	if err != nil {
		if centraldb.IsUniqueViolation(err) {
			return "", apperr.DuplicateCode("holdingcode", holdingCode).WithWrap(err)
		}
		return "", err
	}
	return holdingCode, nil
}

func hasValidHoldingName(names []common.NameX) bool {
	for _, name := range names {
		if name.Code != nil && name.Name != nil && strings.TrimSpace(*name.Code) != "" && strings.TrimSpace(*name.Name) != "" {
			return true
		}
	}
	return false
}

// UpdateShop rewrites the Holding profile. Status changes go through ChangeShopStatus.
func (svc ShopService) UpdateShop(holdingCode string, username string, shop models.Shop) error {
	findShop, err := svc.shopRepo.FindByHoldingCode(context.Background(), holdingCode)
	if err != nil {
		return err
	}
	if requested := strings.TrimSpace(shop.HoldingCode); requested != "" {
		normalized, err := utils.NormalizeHoldingCode(requested)
		if err != nil {
			return err
		}
		if normalized != findShop.HoldingCode {
			return ErrHoldingCodeRenameUnavailable
		}
	}
	shop.HoldingCode = findShop.HoldingCode
	shop.IsActive = findShop.IsActive

	dataDoc := findShop
	dataDoc.Shop = shop
	dataDoc.UpdatedBy = username
	dataDoc.UpdatedAt = svc.timeNow().UTC()
	return svc.shopRepo.Update(context.Background(), findShop.HoldingCode, findShop.IsActive, dataDoc)
}

// ChangeShopStatus opens/closes a Holding with an audited reason; it fails with
// organization.ErrStatusChangeConflict when the status changed concurrently.
func (svc ShopService) ChangeShopStatus(holdingCode string, actorUID string, username string, expectedActive bool, active bool, reason string) error {
	now := svc.timeNow().UTC()
	return svc.shopRepo.Transaction(context.Background(), func(txCtx context.Context) error {
		if err := svc.shopRepo.UpdateStatus(txCtx, holdingCode, expectedActive, active, username, now); err != nil {
			return err
		}
		return svc.shopRepo.RecordAudit(txCtx, orgaccess.Audit{
			HoldingCode: holdingCode, Action: "organization.status.changed", TargetType: "holding", TargetCode: holdingCode,
			ActorUID: actorUID, Reason: reason,
			Before: map[string]bool{"isactive": expectedActive}, After: map[string]bool{"isactive": active},
			OccurredAt: now,
		})
	})
}

func (svc ShopService) InfoShop(holdingCode string) (models.ShopInfo, error) {
	findShop, err := svc.shopRepo.FindByHoldingCode(context.Background(), holdingCode)
	if errors.Is(err, errHoldingNotFound) {
		return models.ShopInfo{}, apperr.ErrNotFound.WithWrap(err)
	}
	if err != nil {
		return models.ShopInfo{}, err
	}
	return findShop.ShopInfo, nil
}
