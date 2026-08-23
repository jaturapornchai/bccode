package shop

import (
	"context"
	"errors"
	auth_model "smlcloudplatform/internal/authentication/models"
	common "smlcloudplatform/internal/models"
	organization "smlcloudplatform/internal/organization"
	"smlcloudplatform/internal/shop/models"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/apperr"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"strings"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type IShopService interface {
	CreateShop(userUID string, username string, shop models.Shop) (string, error)
	UpdateShop(guid string, username string, shop models.Shop) error
	DeleteShop(guid string, username string) error
	InfoShop(guid string) (models.ShopInfo, error)
	SearchShop(pageable micromodels.Pageable) ([]models.ShopInfo, mongopagination.PaginationData, error)
}

type ShopService struct {
	shopRepo     IShopRepository
	shopUserRepo IShopUserRepository
	newGUID      func() string
	timeNow      func() time.Time
}

var ErrHoldingCodeRenameUnavailable = errors.New("holdingcode change requires atomic rename workflow")

func NewShopService(shopRepo IShopRepository, shopUserRepo IShopUserRepository, newGUID func() string, timeNow func() time.Time) ShopService {
	return ShopService{
		shopRepo:     shopRepo,
		shopUserRepo: shopUserRepo,
		newGUID:      newGUID,
		timeNow:      timeNow,
	}
}

func (svc ShopService) CreateShop(userUID string, username string, doc models.Shop) (string, error) {

	dataDoc := models.ShopDoc{}
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
	if existing, findErr := svc.shopRepo.FindByHoldingCode(context.Background(), holdingCode); findErr == nil && existing.GuidFixed != "" {
		return "", apperr.DuplicateCode("holdingcode", holdingCode)
	}
	holdingUID := strings.TrimSpace(svc.newGUID())
	if holdingUID == "" || strings.TrimSpace(userUID) == "" {
		return "", errors.New("stable Holding and User identity are required")
	}
	dataDoc.GuidFixed = holdingUID
	dataDoc.HoldingUID = holdingUID
	dataDoc.Version = 0
	dataDoc.IsDeleted = false
	dataDoc.CreatedBy = userUID
	now := svc.timeNow().UTC()
	dataDoc.CreatedAt = now
	dataDoc.Shop = doc
	dataDoc.HoldingCode = holdingCode
	dataDoc.IsActive = true

	if dataDoc.Shop.MainHoldingCode != "" {
		dataDoc.IsMainShop = false
		dataDoc.MainHoldingCode = doc.MainHoldingCode
	}
	dataDoc.DebtorCenterType = doc.DebtorCenterType
	dataDoc.ProductCenterType = doc.ProductCenterType
	dataDoc.PosProductCenterType = doc.PosProductCenterType

	if err := svc.shopRepo.EnsureBootstrapIndexes(context.Background()); err != nil {
		return "", err
	}
	auditUID := primitive.NewObjectID().Hex()
	eventUID := primitive.NewObjectID().Hex()
	err = svc.shopRepo.Transaction(context.Background(), func(transactionContext context.Context) error {
		if err := svc.shopRepo.CreateCodeClaim(transactionContext, organization.OrganizationCodeClaim{
			EntityType: "holding", ScopeUID: "global", NormalizedCode: holdingCode,
			EntityUID: holdingUID, ClaimedAt: now, ClaimedBy: userUID,
		}); err != nil {
			return err
		}
		if _, err := svc.shopRepo.Create(transactionContext, dataDoc); err != nil {
			return err
		}
		if err := svc.shopUserRepo.SaveStable(transactionContext, holdingCode, holdingUID, userUID, username, auth_model.ROLE_OWNER, now); err != nil {
			return err
		}
		if err := svc.shopRepo.CreateAudit(transactionContext, organization.OrganizationAudit{
			AuditUID: auditUID, ActorUID: userUID, Action: "holding.created",
			TargetType: "holding", TargetUID: holdingUID, HoldingUID: holdingUID,
			After:      bson.M{"holdinguid": holdingUID, "holdingcode": holdingCode, "isactive": true},
			OccurredAt: now,
		}); err != nil {
			return err
		}
		return svc.shopRepo.CreateOutbox(transactionContext, organization.OrganizationOutboxEvent{
			EventUID: eventUID, AggregateType: "holding", AggregateUID: holdingUID,
			Version: 0, EventType: "holding.created",
			Payload: bson.M{"audituid": auditUID, "holdinguid": holdingUID, "holdingcode": holdingCode},
			Status:  "PENDING", Attempts: 0, OccurredAt: now,
		})
	})
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return "", apperr.DuplicateCode("holdingcode", holdingCode).WithWrap(err)
		}
		return "", err
	}

	return holdingUID, nil
}

func hasValidHoldingName(names []common.NameX) bool {
	if len(names) == 0 {
		return false
	}
	for _, name := range names {
		if name.Code != nil && name.Name != nil && strings.TrimSpace(*name.Code) != "" && strings.TrimSpace(*name.Name) != "" {
			return true
		}
	}
	return false
}

func (svc ShopService) UpdateShop(guid string, username string, shop models.Shop) error {

	findShop, err := svc.shopRepo.FindByGuid(context.Background(), guid)

	if err != nil {
		return err
	}

	if findShop.ID == primitive.NilObjectID {
		return errors.New("shop not found")
	}

	dataDoc := findShop
	holdingCodeInput := strings.TrimSpace(shop.HoldingCode)
	if holdingCodeInput == "" {
		shop.HoldingCode = findShop.HoldingCode
	} else {
		holdingCode, err := utils.NormalizeHoldingCode(holdingCodeInput)
		if err != nil {
			return err
		}
		if holdingCode != findShop.HoldingCode {
			return ErrHoldingCodeRenameUnavailable
		}
		shop.HoldingCode = holdingCode
	}

	expectedVersion := findShop.Version
	expectedActive := findShop.IsActive
	shop.IsActive = expectedActive
	dataDoc.Shop = shop

	dataDoc.UpdatedBy = username
	dataDoc.UpdatedAt = time.Now()
	dataDoc.GuidFixed = findShop.GuidFixed

	err = svc.shopRepo.Update(context.Background(), findShop.GuidFixed, expectedVersion, expectedActive, dataDoc)

	if err != nil {
		return err
	}

	return nil
}

func (svc ShopService) DeleteShop(guid string, username string) error {

	err := svc.shopRepo.Delete(context.Background(), guid, username)

	if err != nil {
		return err
	}
	return nil
}

func (svc ShopService) InfoShop(guid string) (models.ShopInfo, error) {
	findShop, err := svc.shopRepo.FindByGuid(context.Background(), guid)

	if err != nil {
		return models.ShopInfo{}, err
	}

	return findShop.ShopInfo, nil
}

func (svc ShopService) SearchShop(pageable micromodels.Pageable) ([]models.ShopInfo, mongopagination.PaginationData, error) {
	shopList, pagination, err := svc.shopRepo.FindPage(context.Background(), pageable)

	if err != nil {
		return shopList, pagination, err
	}

	return shopList, pagination, nil
}
