package option

import (
	"context"
	"errors"
	"smlcloudplatform/internal/product/option/models"
	"smlcloudplatform/internal/utils"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IOptionService interface {
	CreateOption(shopID string, authUsername string, invOpt models.InventoryOptionMain) (string, error)
	UpdateOption(shopID string, guid string, authUsername string, invOpt models.InventoryOptionMain) error
	DeleteOption(shopID string, guid string, username string) error
	InfoOption(shopID string, guid string) (models.InventoryOptionMainInfo, error)
	InfoWTFArray(shopID string, codes []string) ([]interface{}, error)
	SearchOption(shopID string, pageable micromodels.Pageable) ([]models.InventoryOptionMainInfo, mongopagination.PaginationData, error)
}

type OptionService struct {
	repo           IOptionRepository
	contextTimeout time.Duration
}

func NewOptionService(inventoryOptionRepository IOptionRepository) OptionService {

	contextTimeout := time.Duration(15) * time.Second

	return OptionService{
		repo:           inventoryOptionRepository,
		contextTimeout: contextTimeout,
	}
}

func (svc OptionService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc OptionService) CreateOption(shopID string, authUsername string, invOpt models.InventoryOptionMain) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	newGuidFixed := utils.NewGUID()

	invOptDoc := models.InventoryOptionMainDoc{}
	invOptDoc.ShopID = shopID
	invOptDoc.GuidFixed = newGuidFixed

	invOptDoc.InventoryOptionMain = invOpt

	invOptDoc.CreatedBy = authUsername
	invOptDoc.CreatedAt = time.Now()

	if invOptDoc.InventoryOptionMain.Choices == nil {
		invOptDoc.InventoryOptionMain.Choices = &[]models.Choice{}
	}

	_, err := svc.repo.Create(ctx, invOptDoc)

	if err != nil {
		return "", err
	}

	return newGuidFixed, nil
}

func (svc OptionService) UpdateOption(shopID string, guid string, authUsername string, invOpt models.InventoryOptionMain) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	findDoc.InventoryOptionMain = invOpt
	findDoc.UpdatedBy = authUsername
	findDoc.UpdatedAt = time.Now()

	if findDoc.InventoryOptionMain.Choices == nil {
		findDoc.InventoryOptionMain.Choices = &[]models.Choice{}
	}

	svc.repo.Update(ctx, shopID, guid, findDoc)

	if err != nil {
		return err
	}

	return nil
}

func (svc OptionService) DeleteOption(shopID string, guid string, username string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	err := svc.repo.Delete(ctx, shopID, guid, username)

	if err != nil {
		return err
	}

	return nil
}

func (svc OptionService) InfoOption(shopID string, guid string) (models.InventoryOptionMainInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return models.InventoryOptionMainInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.InventoryOptionMainInfo{}, errors.New("document not found")
	}

	return findDoc.InventoryOptionMainInfo, nil
}

func (svc OptionService) InfoWTFArray(shopID string, codes []string) ([]interface{}, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	docList := []interface{}{}

	for _, code := range codes {
		findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", code)
		if err != nil || findDoc.ID == primitive.NilObjectID {
			// add item empty
			docList = append(docList, nil)
		} else {
			docList = append(docList, findDoc.InventoryOptionMainInfo)
		}
	}

	return docList, nil
}

func (svc OptionService) SearchOption(shopID string, pageable micromodels.Pageable) ([]models.InventoryOptionMainInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	docList, pagination, err := svc.repo.FindPage(ctx, shopID, pageable)

	if err != nil {
		return []models.InventoryOptionMainInfo{}, pagination, err
	}

	return docList, pagination, nil
}
