package services

import (
	"context"
	"errors"
	"fmt"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	productbarcode_models "smlcloudplatform/internal/product/productbarcode/models"
	productbarcode_repositories "smlcloudplatform/internal/product/productbarcode/repositories"
	"smlcloudplatform/internal/services"
	trans_models "smlcloudplatform/internal/transaction/models"
	trancache "smlcloudplatform/internal/transaction/repositories"
	"smlcloudplatform/internal/transaction/stockadjustment/models"
	"smlcloudplatform/internal/transaction/stockadjustment/repositories"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/importdata"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"strconv"
	"strings"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type IStockAdjustmentService interface {
	CreateStockAdjustment(holdingCode string, authUsername string, doc models.StockAdjustment) (string, string, error)
	UpdateStockAdjustment(holdingCode string, guid string, authUsername string, doc models.StockAdjustment) error
	DeleteStockAdjustment(holdingCode string, guid string, authUsername string) error
	DeleteStockAdjustmentByGUIDs(holdingCode string, authUsername string, GUIDs []string) error
	InfoStockAdjustment(holdingCode string, guid string) (models.StockAdjustmentInfo, error)
	InfoStockAdjustmentByCode(holdingCode string, code string) (models.StockAdjustmentInfo, error)
	SearchStockAdjustment(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.StockAdjustmentInfo, mongopagination.PaginationData, error)
	SearchStockAdjustmentStep(holdingCode string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.StockAdjustmentInfo, int, error)
	SaveInBatch(holdingCode string, authUsername string, dataList []models.StockAdjustment) (common.BulkImport, error)

	GetModuleName() string
}

type IStockAdjustmenParser interface {
	ParseProductBarcode(detail trans_models.Detail, productBarcodeInfo productbarcode_models.ProductBarcodeInfo) trans_models.Detail
}

const (
	MODULE_NAME = "AJ"
)

type StockAdjustmentService struct {
	repoMq             repositories.IStockAdjustmentMessageQueueRepository
	repo               repositories.IStockAdjustmentRepository
	repoCache          trancache.ICacheRepository
	productbarcodeRepo productbarcode_repositories.IProductBarcodeRepository
	cacheExpireDocNo   time.Duration
	syncCacheRepo      mastersync.IMasterSyncCacheRepository
	services.ActivityService[models.StockAdjustmentActivity, models.StockAdjustmentDeleteActivity]
	parser         IStockAdjustmenParser
	contextTimeout time.Duration
}

func NewStockAdjustmentService(
	repo repositories.IStockAdjustmentRepository,
	repoCache trancache.ICacheRepository,
	productbarcodeRepo productbarcode_repositories.IProductBarcodeRepository,
	repoMq repositories.IStockAdjustmentMessageQueueRepository,
	syncCacheRepo mastersync.IMasterSyncCacheRepository,
	parser IStockAdjustmenParser,
) *StockAdjustmentService {

	contextTimeout := time.Duration(15) * time.Second

	insSvc := &StockAdjustmentService{
		repo:               repo,
		repoMq:             repoMq,
		repoCache:          repoCache,
		productbarcodeRepo: productbarcodeRepo,
		syncCacheRepo:      syncCacheRepo,
		parser:             parser,
		cacheExpireDocNo:   time.Hour * 24,
		contextTimeout:     contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.StockAdjustmentActivity, models.StockAdjustmentDeleteActivity](repo)

	return insSvc
}

func (svc StockAdjustmentService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc StockAdjustmentService) getDocNoPrefix(docDate time.Time) string {
	docDateStr := docDate.Format("20060102")
	return fmt.Sprintf("%s%s", MODULE_NAME, docDateStr)
}

func (svc StockAdjustmentService) generateNewDocNo(ctx context.Context, holdingCode, prefixDocNo string, docNumber int) (string, int, error) {
	prevoiusDocNumber, err := svc.repoCache.Get(holdingCode, prefixDocNo)

	if prevoiusDocNumber == 0 || err != nil {
		lastDoc, err := svc.repo.FindLastDocNo(ctx, holdingCode, prefixDocNo)

		if err != nil {
			return "", 0, err
		}

		if len(lastDoc.DocNo) > 0 {
			rawNumber := strings.Replace(lastDoc.DocNo, prefixDocNo, "", -1)
			prevoiusDocNumber, err = strconv.Atoi(rawNumber)

			if err != nil {
				prevoiusDocNumber = 0
			}
		}

	}

	newDocNumber := prevoiusDocNumber + 1
	newDocNo := fmt.Sprintf("%s%05d", prefixDocNo, newDocNumber)

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "docno", newDocNo)

	if err != nil {
		return "", 0, err
	}

	if len(findDoc.GuidFixed) > 0 {
		return "", 0, errors.New("DocNo is exists")
	}

	return newDocNo, newDocNumber, nil
}

func (svc StockAdjustmentService) CreateStockAdjustment(holdingCode string, authUsername string, doc models.StockAdjustment) (string, string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	docDate := doc.DocDatetime
	prefixDocNo := svc.getDocNoPrefix(docDate)

	newDocNo, newDocNumber, err := svc.generateNewDocNo(ctx, holdingCode, prefixDocNo, 1)

	if err != nil {
		return "", "", err
	}

	newGuidFixed := utils.NewGUID()

	dataDoc := models.StockAdjustmentDoc{}
	dataDoc.HoldingCode = holdingCode
	dataDoc.GuidFixed = newGuidFixed
	dataDoc.StockAdjustment = doc

	productBarcodes, err := svc.GetDetailProductBarcodes(ctx, holdingCode, *doc.Details)
	if err != nil {
		return "", "", err
	}

	details := svc.PrepareDetail(*doc.Details, productBarcodes)
	dataDoc.Details = &details

	dataDoc.DocNo = newDocNo
	dataDoc.CreatedBy = authUsername
	dataDoc.CreatedAt = time.Now()

	_, err = svc.repo.Create(ctx, dataDoc)

	if err != nil {
		return "", "", err
	}

	go svc.repoCache.Save(holdingCode, prefixDocNo, newDocNumber, svc.cacheExpireDocNo)

	go func() {
		svc.repoMq.Create(dataDoc)
		svc.repoCache.Save(holdingCode, prefixDocNo, newDocNumber, svc.cacheExpireDocNo)
		svc.saveMasterSync(holdingCode)
	}()

	return newGuidFixed, newDocNo, nil
}

func (svc StockAdjustmentService) GetDetailProductBarcodes(ctx context.Context, holdingCode string, details []trans_models.Detail) ([]productbarcode_models.ProductBarcodeInfo, error) {
	var tempBarcodes []string
	for _, doc := range details {
		tempBarcodes = append(tempBarcodes, doc.Barcode)
	}
	return svc.productbarcodeRepo.FindByBarcodes(ctx, holdingCode, tempBarcodes)
}

func (svc StockAdjustmentService) PrepareDetail(details []trans_models.Detail, productBarcodes []productbarcode_models.ProductBarcodeInfo) []trans_models.Detail {

	productBarcodeDict := map[string]productbarcode_models.ProductBarcodeInfo{}
	for _, doc := range productBarcodes {
		productBarcodeDict[doc.Barcode] = doc
	}

	for i := 0; i < len(details); i++ {
		tempDetail := (details)[i]
		tempProduct := productBarcodeDict[tempDetail.Barcode]
		if _, ok := productBarcodeDict[tempDetail.Barcode]; ok {
			tempDetail = svc.parser.ParseProductBarcode(tempDetail, tempProduct)
		}
		tempDetail.Discount = (details)[i].Discount
		tempDetail.StandValue = (details)[i].StandValue
		tempDetail.DivideValue = (details)[i].DivideValue
		(details)[i] = tempDetail
	}

	return details
}

func (svc StockAdjustmentService) UpdateStockAdjustment(holdingCode string, guid string, authUsername string, doc models.StockAdjustment) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	dataDoc := findDoc
	dataDoc.StockAdjustment = doc

	productBarcodes, err := svc.GetDetailProductBarcodes(ctx, holdingCode, *doc.Details)
	if err != nil {
		return err
	}

	details := svc.PrepareDetail(*doc.Details, productBarcodes)
	dataDoc.Details = &details

	dataDoc.DocNo = findDoc.DocNo
	dataDoc.UpdatedBy = authUsername
	dataDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, holdingCode, guid, dataDoc)

	if err != nil {
		return err
	}

	func() {
		svc.repoMq.Update(dataDoc)
		svc.saveMasterSync(holdingCode)
	}()

	return nil
}

func (svc StockAdjustmentService) DeleteStockAdjustment(holdingCode string, guid string, authUsername string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	err = svc.repo.DeleteByGuidfixed(ctx, holdingCode, guid, authUsername)
	if err != nil {
		return err
	}

	func() {
		svc.repoMq.Delete(findDoc)
		svc.saveMasterSync(holdingCode)
	}()

	return nil
}

func (svc StockAdjustmentService) DeleteStockAdjustmentByGUIDs(holdingCode string, authUsername string, GUIDs []string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	deleteFilterQuery := map[string]interface{}{
		"guidfixed": bson.M{"$in": GUIDs},
	}

	err := svc.repo.Delete(ctx, holdingCode, authUsername, deleteFilterQuery)
	if err != nil {
		return err
	}

	func() {
		docs, _ := svc.repo.FindByGuids(ctx, holdingCode, GUIDs)
		svc.repoMq.DeleteInBatch(docs)
		svc.saveMasterSync(holdingCode)
	}()

	return nil
}

func (svc StockAdjustmentService) InfoStockAdjustment(holdingCode string, guid string) (models.StockAdjustmentInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return models.StockAdjustmentInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.StockAdjustmentInfo{}, errors.New("document not found")
	}

	return findDoc.StockAdjustmentInfo, nil
}

func (svc StockAdjustmentService) InfoStockAdjustmentByCode(holdingCode string, code string) (models.StockAdjustmentInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "docno", code)

	if err != nil {
		return models.StockAdjustmentInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.StockAdjustmentInfo{}, errors.New("document not found")
	}

	return findDoc.StockAdjustmentInfo, nil
}

func (svc StockAdjustmentService) SearchStockAdjustment(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.StockAdjustmentInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"docno",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, holdingCode, filters, searchInFields, pageable)

	if err != nil {
		return []models.StockAdjustmentInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc StockAdjustmentService) SearchStockAdjustmentStep(holdingCode string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.StockAdjustmentInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"docno",
	}

	selectFields := map[string]interface{}{}

	docList, total, err := svc.repo.FindStep(ctx, holdingCode, filters, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.StockAdjustmentInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc StockAdjustmentService) SaveInBatch(holdingCode string, authUsername string, dataList []models.StockAdjustment) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.StockAdjustment](dataList, svc.getDocIDKey)

	itemCodeGuidList := []string{}
	for _, doc := range payloadList {
		itemCodeGuidList = append(itemCodeGuidList, doc.DocNo)
	}

	findItemGuid, err := svc.repo.FindInItemGuid(ctx, holdingCode, "docno", itemCodeGuidList)

	if err != nil {
		return common.BulkImport{}, err
	}

	foundItemGuidList := []string{}
	for _, doc := range findItemGuid {
		foundItemGuidList = append(foundItemGuidList, doc.DocNo)
	}

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.StockAdjustment, models.StockAdjustmentDoc](
		holdingCode,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(holdingCode string, authUsername string, doc models.StockAdjustment) models.StockAdjustmentDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.StockAdjustmentDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.HoldingCode = holdingCode
			dataDoc.StockAdjustment = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.StockAdjustment, models.StockAdjustmentDoc](
		holdingCode,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(holdingCode string, guid string) (models.StockAdjustmentDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "docno", guid)
		},
		func(doc models.StockAdjustmentDoc) bool {
			return doc.DocNo != ""
		},
		func(holdingCode string, authUsername string, data models.StockAdjustment, doc models.StockAdjustmentDoc) error {

			doc.StockAdjustment = data
			doc.UpdatedBy = authUsername
			doc.UpdatedAt = time.Now()

			err = svc.repo.Update(ctx, holdingCode, doc.GuidFixed, doc)
			if err != nil {
				return nil
			}
			return nil
		},
	)

	if len(createDataList) > 0 {
		err = svc.repo.CreateInBatch(ctx, createDataList)

		if err != nil {
			return common.BulkImport{}, err
		}

	}

	createDataKey := []string{}

	for _, doc := range createDataList {
		createDataKey = append(createDataKey, doc.DocNo)
	}

	payloadDuplicateDataKey := []string{}
	for _, doc := range payloadDuplicateList {
		payloadDuplicateDataKey = append(payloadDuplicateDataKey, doc.DocNo)
	}

	updateDataKey := []string{}
	for _, doc := range updateSuccessDataList {

		updateDataKey = append(updateDataKey, doc.DocNo)
	}

	updateFailDataKey := []string{}
	for _, doc := range updateFailDataList {
		updateFailDataKey = append(updateFailDataKey, svc.getDocIDKey(doc))
	}

	svc.saveMasterSync(holdingCode)

	return common.BulkImport{
		Created:          createDataKey,
		Updated:          updateDataKey,
		UpdateFailed:     updateFailDataKey,
		PayloadDuplicate: payloadDuplicateDataKey,
	}, nil
}

func (svc StockAdjustmentService) getDocIDKey(doc models.StockAdjustment) string {
	return doc.DocNo
}

func (svc StockAdjustmentService) saveMasterSync(holdingCode string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(holdingCode, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc StockAdjustmentService) GetModuleName() string {
	return "stockadjustment"
}
