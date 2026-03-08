package services

import (
	"context"
	"errors"
	"smlcloudplatform/internal/storefront/models"
	"smlcloudplatform/internal/storefront/repositories"
	"smlcloudplatform/internal/utils"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IStorefrontHttpService interface {
	CreateStorefront(shopID string, authUsername string, doc models.Storefront) (string, error)
	UpdateStorefront(shopID string, guid string, authUsername string, doc models.Storefront) error
	DeleteStorefront(shopID string, guid string, authUsername string) error
	InfoStorefront(shopID string, guid string) (models.StorefrontInfo, error)
	SearchStorefront(shopID string, pageable micromodels.Pageable) ([]models.StorefrontInfo, mongopagination.PaginationData, error)
}

type StorefrontHttpService struct {
	repo           repositories.IStorefrontRepository
	contextTimeout time.Duration
}

func NewStorefrontHttpService(repo repositories.IStorefrontRepository) *StorefrontHttpService {

	contextTimeout := time.Duration(15) * time.Second

	return &StorefrontHttpService{
		repo:           repo,
		contextTimeout: contextTimeout,
	}
}

func (svc StorefrontHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc StorefrontHttpService) CreateStorefront(shopID string, authUsername string, doc models.Storefront) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	newGuidFixed := utils.NewGUID()

	docData := models.StorefrontDoc{}
	docData.ShopID = shopID
	docData.GuidFixed = newGuidFixed
	docData.Storefront = doc

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err := svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	return newGuidFixed, nil
}

func (svc StorefrontHttpService) UpdateStorefront(shopID string, guid string, authUsername string, doc models.Storefront) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	findDoc.Storefront = doc

	findDoc.UpdatedBy = authUsername
	findDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, shopID, guid, findDoc)

	if err != nil {
		return err
	}

	return nil
}

func (svc StorefrontHttpService) DeleteStorefront(shopID string, guid string, authUsername string) error {

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

func (svc StorefrontHttpService) InfoStorefront(shopID string, guid string) (models.StorefrontInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return models.StorefrontInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.StorefrontInfo{}, errors.New("document not found")
	}

	return findDoc.StorefrontInfo, nil

}

func (svc StorefrontHttpService) SearchStorefront(shopID string, pageable micromodels.Pageable) ([]models.StorefrontInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"guidfixed",
		"code",
	}

	docList, pagination, err := svc.repo.FindPage(ctx, shopID, searchInFields, pageable)

	if err != nil {
		return []models.StorefrontInfo{}, pagination, err
	}

	return docList, pagination, nil
}
