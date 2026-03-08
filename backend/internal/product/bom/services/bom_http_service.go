package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"smlcloudplatform/internal/logger"
	"smlcloudplatform/internal/product/bom/models"
	"smlcloudplatform/internal/product/bom/repositories"
	"smlcloudplatform/internal/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/checksum"
	"time"

	micromodels "smlcloudplatform/pkg/microservice/models"

	product_models "smlcloudplatform/internal/product/productbarcode/models"
	product_repositories "smlcloudplatform/internal/product/productbarcode/repositories"
	product_services "smlcloudplatform/internal/product/productbarcode/services"
	saleinvoicebom_services "smlcloudplatform/internal/transaction/saleinvoicebomprice/services"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IBOMHttpService interface {
	UpsertBOM(shopID string, authUsername string, dcoNo string, barcode string) (string, error)
	DeleteBOM(shopID string, guid string, authUsername string) error
	InfoBOM(shopID string, guid string) (models.ProductBarcodeBOMViewInfo, error)
	SearchBOM(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ProductBarcodeBOMViewInfo, mongopagination.PaginationData, error)
	SearchBOMStep(shopID string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ProductBarcodeBOMViewInfo, int, error)
}

type BOMHttpService struct {
	repo              repositories.IBomRepository
	repoMq            repositories.IBomMessageQueueRepository
	saleInvoiceBomSvc saleinvoicebom_services.ISaleInvoiceBomPriceService
	productRepo       product_repositories.IProductBarcodeRepository
	services.ActivityService[models.ProductBarcodeBOMViewActivity, models.ProductBarcodeBOMViewDeleteActivity]
	contextTimeout time.Duration
}

func NewBOMHttpService(repo repositories.IBomRepository, repoMq repositories.IBomMessageQueueRepository, productRepo product_repositories.IProductBarcodeRepository, saleInvoiceBomSvc saleinvoicebom_services.ISaleInvoiceBomPriceService) *BOMHttpService {

	contextTimeout := time.Duration(15) * time.Second

	insSvc := &BOMHttpService{
		repo:              repo,
		repoMq:            repoMq,
		productRepo:       productRepo,
		saleInvoiceBomSvc: saleInvoiceBomSvc,
		contextTimeout:    contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.ProductBarcodeBOMViewActivity, models.ProductBarcodeBOMViewDeleteActivity](repo)

	return insSvc
}

func (svc BOMHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc BOMHttpService) UpsertBOM(shopID string, authUsername string, docNo string, barcode string) (string, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	productBarcodeDict := map[string]product_models.ProductBarcodeDoc{}
	bomViewDict := map[string]*product_models.ProductBarcodeBOMView{}
	bomView := product_models.ProductBarcodeBOMView{}

	// product_services.BuildBOMViewCache(
	// 	ctx,
	// 	svc.productRepo.FindByBarcode,
	// 	0,
	// 	&productBarcodeDict,
	// 	&bomViewDict,
	// 	shopID,
	// 	barcode,
	// 	[]product_models.BOMProductBarcode{},
	// 	&bomView,
	// )

	doc, err := svc.productRepo.FindByBarcode(ctx, shopID, barcode)

	if err != nil {
		return "", err
	}

	if len(doc.ProductBarcode.Barcode) == 0 {
		return "", fmt.Errorf("barcode not found")
	}

	bomView.FromProductBarcode(doc.ProductBarcodeData)

	if _, ok := bomViewDict[doc.Barcode]; !ok {
		bomViewDict[doc.Barcode] = &bomView
	}

	bomView.Level = 1

	if doc.BOM != nil && len(*doc.BOM) > 0 {
		productBarcodeDict := map[string]product_models.ProductBarcodeDoc{}
		err = product_services.BuildBOMView(ctx, svc.productRepo.FindByBarcode, bomView.Level, &productBarcodeDict, &bomViewDict, shopID, doc.BOM, &bomView.BOM)
		if err != nil {
			return "", err
		}
	}

	findDoc, err := svc.repo.FindUseBOMByBarcode(ctx, shopID, barcode)
	if err != nil {
		return "", nil
	}

	isCreate := false

	bomBytes, err := json.Marshal(bomView)

	if err != nil {
		return "", err
	}

	docBomView := models.ProductBarcodeBOMView{}

	err = json.Unmarshal(bomBytes, &docBomView)

	if err != nil {
		return "", err
	}

	checkSumStr, err := checksum.Sum(docBomView)
	if err != nil {
		return "", err
	}

	// Create
	if findDoc.GuidFixed == "" {
		isCreate = true
	} else {

		isEqual := findDoc.CheckSum == checkSumStr

		if !isEqual {
			isCreate = true
		}
	}

	bomBarcodes := []string{}

	for tempBarcode := range productBarcodeDict {
		bomBarcodes = append(bomBarcodes, tempBarcode)
	}

	if isCreate {

		err := svc.clearUseBOMByBarcode(ctx, shopID, barcode)

		if err != nil {
			return "", err
		}

		newGUID, err := svc.create(ctx, shopID, authUsername, checkSumStr, docBomView)

		if err != nil {
			return newGUID, err
		}

		_, err = svc.saleInvoiceBomSvc.CreateSaleInvoiceBomPrice(shopID, authUsername, docNo, newGUID, bomBarcodes)

		if err != nil {
			return newGUID, err
		}

		return newGUID, nil
	}

	_, err = svc.saleInvoiceBomSvc.CreateSaleInvoiceBomPrice(shopID, authUsername, docNo, findDoc.GuidFixed, bomBarcodes)

	if err != nil {
		return findDoc.GuidFixed, err
	}

	return findDoc.GuidFixed, nil
}

func (svc BOMHttpService) clearUseBOMByBarcode(ctx context.Context, shopID string, barcode string) error {
	return svc.repo.ClearUseBOMByBarcode(ctx, shopID, barcode)
}

func (svc BOMHttpService) create(ctx context.Context, shopID string, authUsername string, checkSum string, doc models.ProductBarcodeBOMView) (string, error) {

	currentDate := time.Now()
	newGuidFixed := utils.NewGUID()

	docData := models.ProductBarcodeBOMViewDoc{}
	docData.ShopID = shopID
	docData.GuidFixed = newGuidFixed
	docData.ProductBarcodeBOMView = doc
	docData.CheckSum = checkSum
	docData.IsCurrentUse = true
	docData.UseInDate = currentDate

	docData.EmptyOnNil()

	docData.CreatedBy = authUsername
	docData.CreatedAt = currentDate

	_, err := svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	go func() {
		err := svc.repoMq.Create(docData)

		if err != nil {
			logger.GetLogger().Error(err)
		}
	}()

	return newGuidFixed, nil
}

func (svc BOMHttpService) DeleteBOM(shopID string, guid string, authUsername string) error {

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

	go func() {
		err := svc.repoMq.Delete(findDoc)

		if err != nil {
			logger.GetLogger().Error(err)
		}
	}()

	return nil
}

func (svc BOMHttpService) InfoBOM(shopID string, guid string) (models.ProductBarcodeBOMViewInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return models.ProductBarcodeBOMViewInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.ProductBarcodeBOMViewInfo{}, errors.New("document not found")
	}

	return findDoc.ProductBarcodeBOMViewInfo, nil

}

func (svc BOMHttpService) SearchBOM(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ProductBarcodeBOMViewInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"barcode",
		"names.name",
	}

	if len(pageable.Sorts) == 0 {
		pageable.Sorts = []micromodels.KeyInt{
			{Key: "barcode", Value: 1},
			{Key: "level", Value: 1},
		}
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, shopID, filters, searchInFields, pageable)

	if err != nil {
		return []models.ProductBarcodeBOMViewInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc BOMHttpService) SearchBOMStep(shopID string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ProductBarcodeBOMViewInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
		"names.name",
	}

	if len(pageableStep.Sorts) == 0 {
		pageableStep.Sorts = []micromodels.KeyInt{
			{Key: "code", Value: 1},
		}
	}

	selectFields := map[string]interface{}{}

	docList, total, err := svc.repo.FindStep(ctx, shopID, filters, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.ProductBarcodeBOMViewInfo{}, 0, err
	}

	return docList, total, nil
}
