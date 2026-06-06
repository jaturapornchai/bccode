package shop

import (
	"context"
	"errors"
	auth_model "smlcloudplatform/internal/authentication/models"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/shop/models"
	"smlcloudplatform/internal/utils"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"strings"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IShopService interface {
	CreateShop(username string, shop models.Shop) (string, error)
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

func NewShopService(shopRepo IShopRepository, shopUserRepo IShopUserRepository, newGUID func() string, timeNow func() time.Time) ShopService {
	return ShopService{
		shopRepo:     shopRepo,
		shopUserRepo: shopUserRepo,
		newGUID:      newGUID,
		timeNow:      timeNow,
	}
}

func (svc ShopService) CreateShop(username string, doc models.Shop) (string, error) {

	dataDoc := models.ShopDoc{}
	holdingCode, err := utils.NormalizeHoldingCode(doc.HoldingCode)
	if err != nil {
		return "", err
	}
	if holdingCode == "" {
		return "", errors.New("holdingcode invalid")
	}
	if existing, findErr := svc.shopRepo.FindByHoldingCode(context.Background(), holdingCode); findErr == nil && existing.GuidFixed != "" {
		return "", errors.New("holdingcode is exists")
	}
	dataDoc.GuidFixed = holdingCode
	dataDoc.CreatedBy = username
	dataDoc.CreatedAt = svc.timeNow()
	dataDoc.Shop = doc
	dataDoc.HoldingCode = holdingCode

	if dataDoc.Shop.MainHoldingCode != "" {
		dataDoc.IsMainShop = false
		dataDoc.MainHoldingCode = doc.MainHoldingCode
	}
	dataDoc.DebtorCenterType = doc.DebtorCenterType
	dataDoc.ProductCenterType = doc.ProductCenterType
	dataDoc.PosProductCenterType = doc.PosProductCenterType

	if doc.Names == nil {
		dataDoc.Names = []common.NameX{}
	}

	_, err = svc.shopRepo.Create(context.Background(), dataDoc)

	if err != nil {
		return "", err
	}

	err = svc.shopUserRepo.Save(context.Background(), holdingCode, username, auth_model.ROLE_OWNER)

	if err != nil {
		return "", err
	}

	return holdingCode, nil
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
			if existing, findErr := svc.shopRepo.FindByHoldingCode(context.Background(), holdingCode); findErr == nil && existing.GuidFixed != "" && existing.GuidFixed != guid {
				return errors.New("holdingcode is exists")
			}
		}
		shop.HoldingCode = holdingCode
	}

	dataDoc.Shop = shop

	dataDoc.UpdatedBy = username
	dataDoc.UpdatedAt = time.Now()
	dataDoc.GuidFixed = findShop.GuidFixed

	err = svc.shopRepo.Update(context.Background(), guid, dataDoc)

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
