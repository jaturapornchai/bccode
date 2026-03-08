package services

import (
	"context"
	"errors"
	"fmt"
	"smlcloudplatform/internal/shopdesign/zonedesign/models"
	"smlcloudplatform/internal/shopdesign/zonedesign/repositories"
	"smlcloudplatform/internal/utils"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IZoneDesignService interface {
	CreateZoneDesign(shopID string, authUsername string, category models.ZoneDesign) (string, error)
	UpdateZoneDesign(shopID string, guid string, authUsername string, category models.ZoneDesign) error
	DeleteZoneDesign(shopID string, guid string, authUsername string) error
	InfoZoneDesign(shopID string, guid string) (models.ZoneDesignInfo, error)
	SearchZoneDesign(shopID string, pageable micromodels.Pageable) ([]models.ZoneDesignInfo, mongopagination.PaginationData, error)
}

type ZoneDesignService struct {
	repo           repositories.IZoneDesignRepository
	contextTimeout time.Duration
}

func NewZoneDesignService(repo repositories.IZoneDesignRepository) ZoneDesignService {

	contextTimeout := time.Duration(15) * time.Second

	return ZoneDesignService{
		repo:           repo,
		contextTimeout: contextTimeout,
	}
}

func (svc ZoneDesignService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc ZoneDesignService) CreateZoneDesign(shopID string, authUsername string, doc models.ZoneDesign) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	newGuidFixed := utils.NewGUID()

	docData := models.ZoneDesignDoc{}
	docData.ShopID = shopID
	docData.GuidFixed = newGuidFixed
	docData.ZoneDesign = doc

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err := svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}
	return newGuidFixed, nil
}

func (svc ZoneDesignService) UpdateZoneDesign(shopID string, guid string, authUsername string, category models.ZoneDesign) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	findDoc.ZoneDesign = category

	findDoc.UpdatedBy = authUsername
	findDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, shopID, guid, findDoc)

	if err != nil {
		return err
	}
	return nil
}

func (svc ZoneDesignService) DeleteZoneDesign(shopID string, guid string, authUsername string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	err := svc.repo.DeleteByGuidfixed(ctx, shopID, guid, authUsername)

	if err != nil {
		return err
	}
	return nil
}

func (svc ZoneDesignService) InfoZoneDesign(shopID string, guid string) (models.ZoneDesignInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return models.ZoneDesignInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.ZoneDesignInfo{}, errors.New("document not found")
	}

	return findDoc.ZoneDesignInfo, nil

}

func (svc ZoneDesignService) SearchZoneDesign(shopID string, pageable micromodels.Pageable) ([]models.ZoneDesignInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"guidfixed",
	}

	for i := range [5]bool{} {
		searchInFields = append(searchInFields, fmt.Sprintf("name%d", (i+1)))
	}

	docList, pagination, err := svc.repo.FindPage(ctx, shopID, searchInFields, pageable)

	if err != nil {
		return []models.ZoneDesignInfo{}, pagination, err
	}

	return docList, pagination, nil
}
