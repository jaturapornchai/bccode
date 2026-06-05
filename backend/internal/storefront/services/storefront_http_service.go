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
	CreateStorefront(holdingCode string, authUsername string, doc models.Storefront) (string, error)
	UpdateStorefront(holdingCode string, guid string, authUsername string, doc models.Storefront) error
	DeleteStorefront(holdingCode string, guid string, authUsername string) error
	InfoStorefront(holdingCode string, guid string) (models.StorefrontInfo, error)
	SearchStorefront(holdingCode string, pageable micromodels.Pageable) ([]models.StorefrontInfo, mongopagination.PaginationData, error)
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

func (svc StorefrontHttpService) CreateStorefront(holdingCode string, authUsername string, doc models.Storefront) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	newGuidFixed := utils.NewGUID()

	docData := models.StorefrontDoc{}
	docData.HoldingCode = holdingCode
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

func (svc StorefrontHttpService) UpdateStorefront(holdingCode string, guid string, authUsername string, doc models.Storefront) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	findDoc.Storefront = doc

	findDoc.UpdatedBy = authUsername
	findDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, holdingCode, guid, findDoc)

	if err != nil {
		return err
	}

	return nil
}

func (svc StorefrontHttpService) DeleteStorefront(holdingCode string, guid string, authUsername string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	err = svc.repo.DeleteByGuidfixed(ctx, holdingCode, guid, authUsername)
	if err != nil {
		return err
	}

	return nil
}

func (svc StorefrontHttpService) InfoStorefront(holdingCode string, guid string) (models.StorefrontInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return models.StorefrontInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.StorefrontInfo{}, errors.New("document not found")
	}

	return findDoc.StorefrontInfo, nil

}

func (svc StorefrontHttpService) SearchStorefront(holdingCode string, pageable micromodels.Pageable) ([]models.StorefrontInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"guidfixed",
		"code",
	}

	docList, pagination, err := svc.repo.FindPage(ctx, holdingCode, searchInFields, pageable)

	if err != nil {
		return []models.StorefrontInfo{}, pagination, err
	}

	return docList, pagination, nil
}
