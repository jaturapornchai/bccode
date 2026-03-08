package services

import (
	"context"
	"errors"
	"fmt"
	"smlcloudplatform/internal/transaction/saleinvoicebomprice/models"
	"smlcloudplatform/internal/transaction/saleinvoicebomprice/repositories"
	"smlcloudplatform/internal/utils"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
)

type ISaleInvoiceBomPriceService interface {
	CreateSaleInvoiceBomPrice(shopID string, authUsername string, docNo string, bomGUID string, bomBarcodes []string) (string, error)
	DeleteSaleInvoiceBomPrice(shopID string, guid string, authUsername string) error
	InfoSaleInvoiceBomPrice(shopID string, guid string) (models.SaleInvoiceBomPriceInfo, error)
	InfoSaleInvoiceBomPriceByCode(shopID string, code string) (models.SaleInvoiceBomPriceInfo, error)
	InfoSaleInvoiceBomPriceByDocNo(shopID string, docNo string) ([]models.SaleInvoiceBomPriceInfo, error)
	SearchSaleInvoiceBomPrice(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.SaleInvoiceBomPriceInfo, mongopagination.PaginationData, error)
	SearchSaleInvoiceBomPriceStep(shopID string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.SaleInvoiceBomPriceInfo, int, error)
}

type SaleInvoiceBomPriceService struct {
	repoMq           repositories.ISaleInvoiceBOMPriceMessageQueueRepository
	repo             repositories.ISaleInvoiceBomPriceRepository
	cacheExpireDocNo time.Duration
	contextTimeout   time.Duration
}

func NewSaleInvoiceBomPriceService(
	repoMq repositories.ISaleInvoiceBOMPriceMessageQueueRepository,
	repo repositories.ISaleInvoiceBomPriceRepository,
) *SaleInvoiceBomPriceService {

	contextTimeout := time.Duration(15) * time.Second

	insSvc := &SaleInvoiceBomPriceService{
		repo:             repo,
		repoMq:           repoMq,
		cacheExpireDocNo: time.Hour * 24,
		contextTimeout:   contextTimeout,
	}

	return insSvc
}

func (svc SaleInvoiceBomPriceService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc SaleInvoiceBomPriceService) CreateSaleInvoiceBomPrice(shopID string, authUsername string, docNo string, bomGUID string, bomBarcodes []string) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	dataDoc := models.SaleInvoiceBomPriceDoc{}

	newGuidFixed := utils.NewGUID()

	dataDoc.ShopID = shopID
	dataDoc.GuidFixed = newGuidFixed
	dataDoc.BOMGuid = bomGUID
	dataDoc.DocNo = docNo

	// implement bom price
	dataDoc.Prices = []models.SaleInvoicePrice{}
	for _, barcode := range bomBarcodes {
		price := models.SaleInvoicePrice{}
		price.Barcode = barcode
		dataDoc.Prices = append(dataDoc.Prices, price)
	}

	dataDoc.CreatedBy = authUsername
	dataDoc.CreatedAt = time.Now()

	_, err := svc.repo.Create(ctx, dataDoc)

	if err != nil {
		return "", err
	}

	go func() {
		err := svc.repoMq.Create(dataDoc)
		if err != nil {
			fmt.Printf("create mq error :: %s", err.Error())
		}

	}()

	return newGuidFixed, nil
}

func (svc SaleInvoiceBomPriceService) DeleteSaleInvoiceBomPrice(shopID string, guid string, authUsername string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	err = svc.repo.DeleteByGuidfixed(ctx, shopID, guid, authUsername)
	if err != nil {
		return err
	}

	return nil
}

func (svc SaleInvoiceBomPriceService) InfoSaleInvoiceBomPrice(shopID string, guid string) (models.SaleInvoiceBomPriceInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return models.SaleInvoiceBomPriceInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.SaleInvoiceBomPriceInfo{}, errors.New("document not found")
	}

	return findDoc.SaleInvoiceBomPriceInfo, nil
}

func (svc SaleInvoiceBomPriceService) InfoSaleInvoiceBomPriceByCode(shopID string, code string) (models.SaleInvoiceBomPriceInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "docno", code)

	if err != nil {
		return models.SaleInvoiceBomPriceInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.SaleInvoiceBomPriceInfo{}, errors.New("document not found")
	}

	return findDoc.SaleInvoiceBomPriceInfo, nil
}

func (svc SaleInvoiceBomPriceService) InfoSaleInvoiceBomPriceByDocNo(shopID string, docNo string) ([]models.SaleInvoiceBomPriceInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDocs, err := svc.repo.FindByDocNo(ctx, shopID, docNo)

	if err != nil {
		return []models.SaleInvoiceBomPriceInfo{}, err
	}

	if len(findDocs) < 1 {
		return []models.SaleInvoiceBomPriceInfo{}, errors.New("document not found")
	}

	return findDocs, nil
}

func (svc SaleInvoiceBomPriceService) SearchSaleInvoiceBomPrice(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.SaleInvoiceBomPriceInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"docno",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, shopID, filters, searchInFields, pageable)

	if err != nil {
		return []models.SaleInvoiceBomPriceInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc SaleInvoiceBomPriceService) SearchSaleInvoiceBomPriceStep(shopID string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.SaleInvoiceBomPriceInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"docno",
	}

	selectFields := map[string]interface{}{}

	docList, total, err := svc.repo.FindStep(ctx, shopID, filters, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.SaleInvoiceBomPriceInfo{}, 0, err
	}

	return docList, total, nil
}
