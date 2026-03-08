package services

import (
	"context"
	"errors"
	"smlcloudplatform/internal/shopcoupon/models"
	"smlcloudplatform/internal/shopcoupon/repositories"
	"smlcloudplatform/internal/utils"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IShopCouponHttpService interface {
	CreateShopCoupon(shopID string, authUsername string, doc models.ShopCoupon) (string, error)
	UpdateShopCoupon(shopID string, guid string, authUsername string, doc models.ShopCoupon) error
	DeleteShopCoupon(shopID string, guid string, authUsername string) error
	InfoShopCoupon(shopID string, guid string) (models.ShopCouponInfo, error)
	SearchShopCoupon(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ShopCouponInfo, mongopagination.PaginationData, error)
}

type ShopCouponHttpService struct {
	repo           repositories.IShopCouponRepository
	contextTimeout time.Duration
}

func NewShopCouponHttpService(repo repositories.IShopCouponRepository) *ShopCouponHttpService {

	contextTimeout := time.Duration(15) * time.Second

	return &ShopCouponHttpService{
		repo:           repo,
		contextTimeout: contextTimeout,
	}
}

func (svc ShopCouponHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc ShopCouponHttpService) CreateShopCoupon(shopID string, authUsername string, doc models.ShopCoupon) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	newGuidFixed := utils.NewGUID()

	docData := models.ShopCouponDoc{}
	docData.ShopID = shopID
	docData.GuidFixed = newGuidFixed
	docData.ShopCoupon = doc

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err := svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	return newGuidFixed, nil
}

func (svc ShopCouponHttpService) UpdateShopCoupon(shopID string, guid string, authUsername string, doc models.ShopCoupon) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	findDoc.ShopCoupon = doc

	findDoc.UpdatedBy = authUsername
	findDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, shopID, guid, findDoc)

	if err != nil {
		return err
	}

	return nil
}

func (svc ShopCouponHttpService) DeleteShopCoupon(shopID string, guid string, authUsername string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	err = svc.repo.DeleteByGuidfixed(ctx, shopID, guid, authUsername)
	if err != nil {
		return err
	}

	return nil
}

func (svc ShopCouponHttpService) InfoShopCoupon(shopID string, guid string) (models.ShopCouponInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return models.ShopCouponInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.ShopCouponInfo{}, errors.New("document not found")
	}

	return findDoc.ShopCouponInfo, nil

}

func (svc ShopCouponHttpService) SearchShopCoupon(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ShopCouponInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"name1",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, shopID, filters, searchInFields, pageable)

	if err != nil {
		return []models.ShopCouponInfo{}, pagination, err
	}

	return docList, pagination, nil
}
