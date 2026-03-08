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
	"smlcloudplatform/internal/transaction/stocktransfer/models"
	"smlcloudplatform/internal/transaction/stocktransfer/repositories"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/importdata"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"strconv"
	"strings"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type IStockTransferService interface {
	CreateStockTransfer(shopID string, authUsername string, doc models.StockTransfer) (string, string, error)
	UpdateStockTransfer(shopID string, guid string, authUsername string, doc models.StockTransfer) error
	DeleteStockTransfer(shopID string, guid string, authUsername string) error
	DeleteStockTransferByGUIDs(shopID string, authUsername string, GUIDs []string) error
	InfoStockTransfer(shopID string, guid string) (models.StockTransferInfo, error)
	InfoStockTransferByCode(shopID string, code string) (models.StockTransferInfo, error)
	SearchStockTransfer(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.StockTransferInfo, mongopagination.PaginationData, error)
	SearchStockTransferStep(shopID string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.StockTransferInfo, int, error)
	SaveInBatch(shopID string, authUsername string, dataList []models.StockTransfer) (common.BulkImport, error)

	GetModuleName() string
}

type IStockTransferParser interface {
	ParseProductBarcode(detail trans_models.Detail, productBarcodeInfo productbarcode_models.ProductBarcodeInfo) trans_models.Detail
}

const (
	MODULE_NAME = "TF"
)

type StockTransferService struct {
	repoMq             repositories.IStockTransferMessageQueueRepository
	repo               repositories.IStockTransferRepository
	repoCache          trancache.ICacheRepository
	productbarcodeRepo productbarcode_repositories.IProductBarcodeRepository
	cacheExpireDocNo   time.Duration
	syncCacheRepo      mastersync.IMasterSyncCacheRepository
	services.ActivityService[models.StockTransferActivity, models.StockTransferDeleteActivity]
	parser         IStockTransferParser
	contextTimeout time.Duration
}

func NewStockTransferService(
	repo repositories.IStockTransferRepository,
	repoCache trancache.ICacheRepository,
	productbarcodeRepo productbarcode_repositories.IProductBarcodeRepository,
	repoMq repositories.IStockTransferMessageQueueRepository,
	syncCacheRepo mastersync.IMasterSyncCacheRepository,
	parser IStockTransferParser,
) *StockTransferService {

	contextTimeout := time.Duration(15) * time.Second

	insSvc := &StockTransferService{
		repoMq:             repoMq,
		repo:               repo,
		repoCache:          repoCache,
		productbarcodeRepo: productbarcodeRepo,
		syncCacheRepo:      syncCacheRepo,
		parser:             parser,
		cacheExpireDocNo:   time.Hour * 24,
		contextTimeout:     contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.StockTransferActivity, models.StockTransferDeleteActivity](repo)

	return insSvc
}

func (svc StockTransferService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc StockTransferService) getDocNoPrefix(docDate time.Time) string {
	docDateStr := docDate.Format("20060102")
	return fmt.Sprintf("%s%s", MODULE_NAME, docDateStr)
}

func (svc StockTransferService) generateNewDocNo(ctx context.Context, shopID, prefixDocNo string, docNumber int) (string, int, error) {
	prevoiusDocNumber, err := svc.repoCache.Get(shopID, prefixDocNo)

	if prevoiusDocNumber == 0 || err != nil {
		lastDoc, err := svc.repo.FindLastDocNo(ctx, shopID, prefixDocNo)

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

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "docno", newDocNo)

	if err != nil {
		return "", 0, err
	}

	if len(findDoc.GuidFixed) > 0 {
		return "", 0, errors.New("DocNo is exists")
	}

	return newDocNo, newDocNumber, nil
}

func (svc StockTransferService) CreateStockTransfer(shopID string, authUsername string, doc models.StockTransfer) (string, string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	docDate := doc.DocDatetime
	prefixDocNo := svc.getDocNoPrefix(docDate)

	newDocNo, newDocNumber, err := svc.generateNewDocNo(ctx, shopID, prefixDocNo, 1)

	if err != nil {
		return "", "", err
	}

	newGuidFixed := utils.NewGUID()

	dataDoc := models.StockTransferDoc{}
	dataDoc.ShopID = shopID
	dataDoc.GuidFixed = newGuidFixed
	dataDoc.StockTransfer = doc

	productBarcodes, err := svc.GetDetailProductBarcodes(ctx, shopID, *doc.Details)
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

	go func() {
		svc.repoMq.Create(dataDoc)
		svc.repoCache.Save(shopID, prefixDocNo, newDocNumber, svc.cacheExpireDocNo)
		svc.saveMasterSync(shopID)
	}()

	return newGuidFixed, newDocNo, nil
}

func (svc StockTransferService) GetDetailProductBarcodes(ctx context.Context, shopID string, details []trans_models.Detail) ([]productbarcode_models.ProductBarcodeInfo, error) {
	var tempBarcodes []string
	for _, doc := range details {
		tempBarcodes = append(tempBarcodes, doc.Barcode)
	}
	return svc.productbarcodeRepo.FindByBarcodes(ctx, shopID, tempBarcodes)
}

func (svc StockTransferService) PrepareDetail(details []trans_models.Detail, productBarcodes []productbarcode_models.ProductBarcodeInfo) []trans_models.Detail {

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

func (svc StockTransferService) UpdateStockTransfer(shopID string, guid string, authUsername string, doc models.StockTransfer) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	findExists, err := svc.repo.FindDocOne(ctx, shopID, doc.DocNo, doc.TransFlag)

	if err != nil {
		return err
	}

	if findExists.DocNo != findDoc.DocNo && findExists.TransFlag != findDoc.TransFlag && len(findExists.GuidFixed) > 0 {
		return errors.New("docno and trans flag is exists")
	}

	dataDoc := findDoc
	dataDoc.StockTransfer = doc

	productBarcodes, err := svc.GetDetailProductBarcodes(ctx, shopID, *doc.Details)
	if err != nil {
		return err
	}

	details := svc.PrepareDetail(*doc.Details, productBarcodes)
	dataDoc.Details = &details

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

func (svc StockTransferService) DeleteStockTransfer(shopID string, guid string, authUsername string) error {

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

func (svc StockTransferService) DeleteStockTransferByGUIDs(shopID string, authUsername string, GUIDs []string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	deleteFilterQuery := map[string]interface{}{
		"guidfixed": bson.M{"$in": GUIDs},
	}

	err := svc.repo.Delete(ctx, shopID, authUsername, deleteFilterQuery)
	if err != nil {
		return err
	}

	func() {
		docs, _ := svc.repo.FindByGuids(ctx, shopID, GUIDs)
		svc.repoMq.DeleteInBatch(docs)
		svc.saveMasterSync(shopID)
	}()

	return nil
}

func (svc StockTransferService) InfoStockTransfer(shopID string, guid string) (models.StockTransferInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return models.StockTransferInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.StockTransferInfo{}, errors.New("document not found")
	}

	return findDoc.StockTransferInfo, nil
}

func (svc StockTransferService) InfoStockTransferByCode(shopID string, code string) (models.StockTransferInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "docno", code)

	if err != nil {
		return models.StockTransferInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.StockTransferInfo{}, errors.New("document not found")
	}

	return findDoc.StockTransferInfo, nil
}

func (svc StockTransferService) SearchStockTransfer(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.StockTransferInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"docno",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, shopID, filters, searchInFields, pageable)

	if err != nil {
		return []models.StockTransferInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc StockTransferService) SearchStockTransferStep(shopID string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.StockTransferInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"docno",
	}

	selectFields := map[string]interface{}{}

	docList, total, err := svc.repo.FindStep(ctx, shopID, filters, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.StockTransferInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc StockTransferService) SaveInBatch(shopID string, authUsername string, dataList []models.StockTransfer) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.StockTransfer](dataList, svc.getDocIDKey)

	itemCodeGuidList := []string{}
	for _, doc := range payloadList {
		itemCodeGuidList = append(itemCodeGuidList, doc.DocNo)
	}

	findItemGuid, err := svc.repo.FindInItemGuid(ctx, shopID, "docno", itemCodeGuidList)

	if err != nil {
		return common.BulkImport{}, err
	}

	foundItemGuidList := []string{}
	for _, doc := range findItemGuid {
		foundItemGuidList = append(foundItemGuidList, doc.DocNo)
	}

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.StockTransfer, models.StockTransferDoc](
		shopID,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(shopID string, authUsername string, doc models.StockTransfer) models.StockTransferDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.StockTransferDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.ShopID = shopID
			dataDoc.StockTransfer = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.StockTransfer, models.StockTransferDoc](
		shopID,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(shopID string, guid string) (models.StockTransferDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, shopID, "docno", guid)
		},
		func(doc models.StockTransferDoc) bool {
			return doc.DocNo != ""
		},
		func(shopID string, authUsername string, data models.StockTransfer, doc models.StockTransferDoc) error {

			doc.StockTransfer = data
			doc.UpdatedBy = authUsername
			doc.UpdatedAt = time.Now()

			err = svc.repo.Update(ctx, shopID, doc.GuidFixed, doc)
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

	svc.saveMasterSync(shopID)

	return common.BulkImport{
		Created:          createDataKey,
		Updated:          updateDataKey,
		UpdateFailed:     updateFailDataKey,
		PayloadDuplicate: payloadDuplicateDataKey,
	}, nil
}

func (svc StockTransferService) getDocIDKey(doc models.StockTransfer) string {
	return doc.DocNo
}

func (svc StockTransferService) saveMasterSync(shopID string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(shopID, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc StockTransferService) GetModuleName() string {
	return "stockTransfer"
}
