package services

import (
	"context"
	"errors"
	"fmt"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	productbarcode_models "smlcloudplatform/internal/product/productbarcode/models"
	productbarcode_repositories "smlcloudplatform/internal/product/productbarcode/repositories"
	"smlcloudplatform/internal/services"
	trans_models "smlcloudplatform/internal/transaction/models"
	trancache "smlcloudplatform/internal/transaction/repositories"
	"smlcloudplatform/internal/transaction/stockbalancedetail/models"
	"smlcloudplatform/internal/transaction/stockbalancedetail/repositories"
	"smlcloudplatform/internal/utils"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type IStockBalanceDetailService interface {
	CreateStockBalanceDetail(shopID string, authUsername string, doc []models.StockBalanceDetail) error
	UpdateStockBalanceDetail(shopID string, guid string, authUsername string, doc models.StockBalanceDetail) error
	DeleteStockBalanceDetail(shopID string, guid string, authUsername string) error
	DeleteStockBalanceDetailByGUIDs(shopID string, authUsername string, GUIDs []string) error
	DeleteStockBalanceDetailByDocNo(shopID string, authUsername string, docNo string) error
	InfoStockBalanceDetail(shopID string, guid string) (models.StockBalanceDetailInfo, error)
	InfoStockBalanceDetailByCode(shopID string, code string) (models.StockBalanceDetailInfo, error)
	SearchStockBalanceDetail(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.StockBalanceDetailInfo, mongopagination.PaginationData, error)
	SearchStockBalanceDetailStep(shopID string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.StockBalanceDetailInfo, int, error)

	GetModuleName() string
}

type IStockBalanceDetailParser interface {
	ParseProductBarcode(detail trans_models.Detail, productBarcodeInfo productbarcode_models.ProductBarcodeInfo) trans_models.Detail
}

type StockBalanceDetailService struct {
	repoMq             repositories.IStockBalanceDetailMessageQueueRepository
	repo               repositories.IStockBalanceDetailRepository
	repoCache          trancache.ICacheRepository
	productBarcodeRepo productbarcode_repositories.IProductBarcodeRepository
	cacheExpireDocNo   time.Duration
	syncCacheRepo      mastersync.IMasterSyncCacheRepository
	services.ActivityService[models.StockBalanceDetailActivity, models.StockBalanceDetailDeleteActivity]
	parser         IStockBalanceDetailParser
	contextTimeout time.Duration
}

func NewStockBalanceDetailService(
	repo repositories.IStockBalanceDetailRepository,
	repoCache trancache.ICacheRepository,
	productBarcodeRepo productbarcode_repositories.IProductBarcodeRepository,
	repoMq repositories.IStockBalanceDetailMessageQueueRepository,
	syncCacheRepo mastersync.IMasterSyncCacheRepository,
	parser IStockBalanceDetailParser,
) *StockBalanceDetailService {

	contextTimeout := time.Duration(15) * time.Second

	insSvc := &StockBalanceDetailService{
		repoMq:             repoMq,
		repo:               repo,
		productBarcodeRepo: productBarcodeRepo,
		parser:             parser,
		repoCache:          repoCache,
		syncCacheRepo:      syncCacheRepo,
		cacheExpireDocNo:   time.Hour * 24,
		contextTimeout:     contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.StockBalanceDetailActivity, models.StockBalanceDetailDeleteActivity](repo)

	return insSvc
}

func (svc StockBalanceDetailService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc StockBalanceDetailService) CreateStockBalanceDetail(shopID string, authUsername string, docs []models.StockBalanceDetail) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	prepareDocs := []models.StockBalanceDetailDoc{}

	for _, doc := range docs {
		newGuidFixed := utils.NewGUID()

		dataDoc := models.StockBalanceDetailDoc{}
		dataDoc.ShopID = shopID
		dataDoc.GuidFixed = newGuidFixed
		dataDoc.StockBalanceDetail = doc

		dataDoc.CreatedBy = authUsername
		dataDoc.CreatedAt = time.Now()

		productBarcode, err := svc.GetDetailProductBarcode(ctx, shopID, doc.Barcode)
		if err != nil {
			return err
		}

		dataDoc.Detail = svc.PrepareDetail(doc.Detail, productBarcode)

		prepareDocs = append(prepareDocs, dataDoc)
	}

	err := svc.repo.CreateInBatch(ctx, prepareDocs)

	if err != nil {
		return err
	}

	go func() {
		svc.repoMq.CreateInBatch(prepareDocs)
		svc.saveMasterSync(shopID)
	}()

	return nil
}

func (svc StockBalanceDetailService) GetDetailProductBarcode(ctx context.Context, shopID string, barcode string) (productbarcode_models.ProductBarcodeInfo, error) {
	results, err := svc.productBarcodeRepo.FindByBarcodes(ctx, shopID, []string{barcode})

	if err != nil {
		return productbarcode_models.ProductBarcodeInfo{}, err
	}

	if len(results) < 1 {
		return productbarcode_models.ProductBarcodeInfo{}, errors.New("product barcode not found")
	}

	return results[0], nil
}

func (svc StockBalanceDetailService) PrepareDetail(detail trans_models.Detail, productBarcode productbarcode_models.ProductBarcodeInfo) trans_models.Detail {

	resultDetail := svc.parser.ParseProductBarcode(detail, productBarcode)

	return resultDetail
}

func (svc StockBalanceDetailService) UpdateStockBalanceDetail(shopID string, guid string, authUsername string, doc models.StockBalanceDetail) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	dataDoc := findDoc
	dataDoc.StockBalanceDetail = doc

	dataDoc.DocNo = findDoc.DocNo
	dataDoc.UpdatedBy = authUsername
	dataDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, shopID, guid, dataDoc)

	if err != nil {
		return err
	}

	func() {
		svc.repoMq.Update(dataDoc)
		svc.saveMasterSync(shopID)
	}()

	return nil
}

func (svc StockBalanceDetailService) DeleteStockBalanceDetail(shopID string, guid string, authUsername string) error {

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

	func() {
		svc.repoMq.Delete(findDoc)
		svc.saveMasterSync(shopID)
	}()

	return nil
}

func (svc StockBalanceDetailService) DeleteStockBalanceDetailByGUIDs(shopID string, authUsername string, GUIDs []string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	// prepare item for message queue
	docs, err := svc.repo.FindByGuids(ctx, shopID, GUIDs)
	if err != nil {
		return err
	}

	deleteFilterQuery := map[string]interface{}{
		"guidfixed": bson.M{"$in": GUIDs},
	}

	err = svc.repo.Delete(ctx, shopID, authUsername, deleteFilterQuery)
	if err != nil {
		return err
	}

	func() {
		svc.repoMq.DeleteInBatch(docs)
		svc.saveMasterSync(shopID)
	}()

	return nil
}

func (svc StockBalanceDetailService) DeleteStockBalanceDetailByDocNo(shopID string, authUsername string, docNo string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	docs, err := svc.repo.FindByDocIndentityGuids(ctx, shopID, "docno", docNo)
	if err != nil {
		return err
	}

	deleteFilterQuery := map[string]interface{}{
		"docno": docNo,
	}

	err = svc.repo.Delete(ctx, shopID, authUsername, deleteFilterQuery)
	if err != nil {
		return err
	}

	func() {
		svc.repoMq.DeleteInBatch(docs)
		svc.saveMasterSync(shopID)
	}()

	return nil
}

func (svc StockBalanceDetailService) InfoStockBalanceDetail(shopID string, guid string) (models.StockBalanceDetailInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return models.StockBalanceDetailInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.StockBalanceDetailInfo{}, errors.New("document not found")
	}

	return findDoc.StockBalanceDetailInfo, nil
}

func (svc StockBalanceDetailService) InfoStockBalanceDetailByCode(shopID string, code string) (models.StockBalanceDetailInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "docno", code)

	if err != nil {
		return models.StockBalanceDetailInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.StockBalanceDetailInfo{}, errors.New("document not found")
	}

	return findDoc.StockBalanceDetailInfo, nil
}

func (svc StockBalanceDetailService) SearchStockBalanceDetail(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.StockBalanceDetailInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"docno",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, shopID, filters, searchInFields, pageable)

	if err != nil {
		return []models.StockBalanceDetailInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc StockBalanceDetailService) SearchStockBalanceDetailStep(shopID string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.StockBalanceDetailInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"docno",
	}

	selectFields := map[string]interface{}{}

	docList, total, err := svc.repo.FindStep(ctx, shopID, filters, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.StockBalanceDetailInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc StockBalanceDetailService) saveMasterSync(shopID string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(shopID, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc StockBalanceDetailService) GetModuleName() string {
	return "stockBalanceDetail"
}
