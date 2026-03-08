package services

import (
	"context"
	"errors"
	"fmt"
	custRepositories "smlcloudplatform/internal/debtaccount/debtor/repositories"
	master_sync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	productbarcode_models "smlcloudplatform/internal/product/productbarcode/models"
	productbarcode_repositories "smlcloudplatform/internal/product/productbarcode/repositories"
	"smlcloudplatform/internal/services"
	trans_models "smlcloudplatform/internal/transaction/models"
	trans_cache "smlcloudplatform/internal/transaction/repositories"
	"smlcloudplatform/internal/transaction/saleinvoicereturn/models"
	"smlcloudplatform/internal/transaction/saleinvoicereturn/repositories"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/importdata"
	micro_models "smlcloudplatform/pkg/microservice/models"
	"strconv"
	"strings"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type ISaleInvoiceReturnService interface {
	CreateSaleInvoiceReturn(shopID string, authUsername string, doc models.SaleInvoiceReturn) (string, string, error)
	UpdateSaleInvoiceReturn(shopID string, guid string, authUsername string, doc models.SaleInvoiceReturn) error
	UpdateSlip(shopID string, authUsername string, docNo string, mode uint8, machineCode string, zoneGroupNumber string, imageUrl string) error

	DeleteSaleInvoiceReturn(shopID string, guid string, authUsername string) error
	DeleteSaleInvoiceReturnByGUIDs(shopID string, authUsername string, GUIDs []string) error
	InfoSaleInvoiceReturn(shopID string, guid string) (models.SaleInvoiceReturnInfo, error)
	InfoSaleInvoiceReturnByCode(shopID string, code string) (models.SaleInvoiceReturnInfo, error)
	SearchSaleInvoiceReturn(shopID string, filters map[string]interface{}, pageable micro_models.Pageable) ([]models.SaleInvoiceReturnInfo, mongopagination.PaginationData, error)
	SearchSaleInvoiceReturnStep(shopID string, langCode string, filters map[string]interface{}, pageableStep micro_models.PageableStep) ([]models.SaleInvoiceReturnInfo, int, error)
	SaveInBatch(shopID string, authUsername string, dataList []models.SaleInvoiceReturn) (common.BulkImport, error)
	GetLastPOSDocNo(shopID, posID, maxDocNo string) (string, error)

	GetModuleName() string
}

type ISaleInvocieReturnParser interface {
	ParseProductBarcode(detail trans_models.Detail, productBarcode productbarcode_models.ProductBarcodeInfo) trans_models.Detail
}

const (
	MODULE_NAME = "ST"
)

type SaleInvoiceReturnService struct {
	repoMq               repositories.ISaleInvoiceReturnMessageQueueRepository
	repo                 repositories.ISaleInvoiceReturnRepository
	repoCache            trans_cache.ICacheRepository
	productbarcodeRepo   productbarcode_repositories.IProductBarcodeRepository
	pointTransactionRepo custRepositories.IPointTransactionRepository
	repoCust             custRepositories.IDebtorRepository
	cacheExpireDocNo     time.Duration
	syncCacheRepo        master_sync.IMasterSyncCacheRepository
	services.ActivityService[models.SaleInvoiceReturnActivity, models.SaleInvoiceReturnDeleteActivity]
	parser         ISaleInvocieReturnParser
	contextTimeout time.Duration
}

func NewSaleInvoiceReturnService(
	repo repositories.ISaleInvoiceReturnRepository,
	repoCache trans_cache.ICacheRepository,
	productbarcodeRepo productbarcode_repositories.IProductBarcodeRepository,
	repoMq repositories.ISaleInvoiceReturnMessageQueueRepository,
	syncCacheRepo master_sync.IMasterSyncCacheRepository,
	parser ISaleInvocieReturnParser,
	pointTransactionRepo custRepositories.IPointTransactionRepository,
	repoCust custRepositories.IDebtorRepository,
) *SaleInvoiceReturnService {

	contextTimeout := time.Duration(15) * time.Second

	insSvc := &SaleInvoiceReturnService{
		repo:                 repo,
		repoMq:               repoMq,
		repoCache:            repoCache,
		productbarcodeRepo:   productbarcodeRepo,
		pointTransactionRepo: pointTransactionRepo,
		repoCust:             repoCust,
		syncCacheRepo:        syncCacheRepo,
		parser:               parser,
		cacheExpireDocNo:     time.Hour * 24,
		contextTimeout:       contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.SaleInvoiceReturnActivity, models.SaleInvoiceReturnDeleteActivity](repo)

	return insSvc
}

func (svc SaleInvoiceReturnService) processPointReturnTransactions(ctx context.Context, shopID string, authUsername string, custCode string, pointsCode string, docNo string, docDate time.Time, getPoint float64, usePoint float64) error {
	// SaleInvoiceReturn ไม่มีการคืน point
	// การคืน point จะเกิดขึ้นเฉพาะเมื่อ SaleInvoice มีการ set iscancel = true เท่านั้น
	return nil
}

// processPointReturnCancelTransactions handles point refund when cancelling a sale return
func (svc SaleInvoiceReturnService) processPointReturnCancelTransactions(ctx context.Context, shopID string, authUsername string, custCode string, pointsCode string, docNo string, docDate time.Time, getPoint float64, usePoint float64) error {
	// SaleInvoiceReturn ไม่มีการคืน point
	// การคืน point จะเกิดขึ้นเฉพาะเมื่อ SaleInvoice มีการ set iscancel = true เท่านั้น
	return nil
}

func (svc SaleInvoiceReturnService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc SaleInvoiceReturnService) getDocNoPrefix(docDate time.Time) string {
	docDateStr := docDate.Format("20060102")
	return fmt.Sprintf("%s%s", MODULE_NAME, docDateStr)
}

func (svc SaleInvoiceReturnService) generateNewDocNo(ctx context.Context, shopID, prefixDocNo string, docNumber int) (string, int, error) {
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
func (svc SaleInvoiceReturnService) CreateSaleInvoiceReturn(shopID string, authUsername string, doc models.SaleInvoiceReturn) (string, string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	isGenerateDocNo := !doc.IsPOS

	prefixDocNo, docNo, newDocNumber := "", "", 0

	if isGenerateDocNo {
		docDate := doc.DocDatetime
		prefixDocNo := svc.getDocNoPrefix(docDate)

		tempNewDocNo, tempNewDocNumber, err := svc.generateNewDocNo(ctx, shopID, prefixDocNo, 1)

		if err != nil {
			return "", "", err
		}

		docNo = tempNewDocNo
		newDocNumber = tempNewDocNumber
	} else {
		if doc.DocNo == "" {
			return "", "", errors.New("docno is required")
		}

		docNo = doc.DocNo
	}

	if isGenerateDocNo {
		findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "docno", docNo)

		if err != nil {
			return "", "", err
		}

		if len(findDoc.GuidFixed) > 0 {
			return "", "", errors.New("DocNo is exists")
		}
	}

	newGuidFixed := utils.NewGUID()

	dataDoc := models.SaleInvoiceReturnDoc{}
	dataDoc.ShopID = shopID
	dataDoc.GuidFixed = newGuidFixed
	dataDoc.SaleInvoiceReturn = doc

	dataDoc.DocNo = docNo
	if isGenerateDocNo && doc.TaxDocNo == "" {
		dataDoc.TaxDocNo = docNo
	}

	productBarcodes, err := svc.GetDetailProductBarcodes(ctx, shopID, *doc.Details)
	if err != nil {
		return "", "", err
	}

	details := svc.PrepareDetail(*doc.Details, productBarcodes)
	dataDoc.Details = &details

	dataDoc.CreatedBy = authUsername
	dataDoc.CreatedAt = time.Now()

	_, err = svc.repo.Create(ctx, dataDoc)

	if err != nil {
		return "", "", err
	}

	// SaleInvoiceReturn ไม่มีการคืน point
	// การคืน point จะเกิดขึ้นเฉพาะเมื่อ SaleInvoice มีการ set iscancel = true เท่านั้น

	go func() {
		svc.repoMq.Create(dataDoc)
		svc.saveMasterSync(shopID)
		if isGenerateDocNo {
			svc.repoCache.Save(shopID, prefixDocNo, newDocNumber, svc.cacheExpireDocNo)
		}
	}()

	return newGuidFixed, docNo, nil
}
func (svc SaleInvoiceReturnService) GetDetailProductBarcodes(ctx context.Context, shopID string, details []trans_models.Detail) ([]productbarcode_models.ProductBarcodeInfo, error) {
	var tempBarcodes []string
	for _, doc := range details {
		tempBarcodes = append(tempBarcodes, doc.Barcode)
	}
	return svc.productbarcodeRepo.FindByBarcodes(ctx, shopID, tempBarcodes)
}

func (svc SaleInvoiceReturnService) PrepareDetail(details []trans_models.Detail, productBarcodes []productbarcode_models.ProductBarcodeInfo) []trans_models.Detail {

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

func (svc SaleInvoiceReturnService) UpdateSaleInvoiceReturn(shopID string, guid string, authUsername string, doc models.SaleInvoiceReturn) error {

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
	dataDoc.SaleInvoiceReturn = doc

	productBarcodes, err := svc.GetDetailProductBarcodes(ctx, shopID, *doc.Details)
	if err != nil {
		return err
	}

	details := svc.PrepareDetail(*doc.Details, productBarcodes)
	dataDoc.Details = &details

	dataDoc.SlipQrUrl = findDoc.SlipQrUrl
	dataDoc.SlipQrUrlHistories = findDoc.SlipQrUrlHistories
	dataDoc.SlipUrl = findDoc.SlipUrl
	dataDoc.SlipUrlHistories = findDoc.SlipUrlHistories

	dataDoc.DocNo = findDoc.DocNo
	dataDoc.UpdatedBy = authUsername
	dataDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, shopID, guid, dataDoc)

	if err != nil {
		return err
	}

	// SaleInvoiceReturn ไม่มีการคืน point
	// การคืน point จะเกิดขึ้นเฉพาะเมื่อ SaleInvoice มีการ set iscancel = true เท่านั้น

	func() {
		svc.repoMq.Update(dataDoc)
		svc.saveMasterSync(shopID)
	}()

	return nil
}

func (svc SaleInvoiceReturnService) UpdateSlip(shopID string, authUsername string, docNo string, mode uint8, machineCode string, zoneGroupNumber string, imageUrl string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "docno", docNo)

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	dataDoc := findDoc

	if mode == 1 {
		if findDoc.SlipQrUrl != "" {
			dataDoc.SlipQrUrlHistories = append(dataDoc.SlipQrUrlHistories, findDoc.SlipQrUrl)
		}

		dataDoc.SlipQrUrl = imageUrl
	} else {
		if findDoc.SlipUrl != "" {
			dataDoc.SlipUrlHistories = append(dataDoc.SlipUrlHistories, findDoc.SlipUrl)
		}

		dataDoc.SlipUrl = imageUrl
	}

	dataDoc.MachineCode = machineCode
	dataDoc.ZoneGroupNumber = zoneGroupNumber

	dataDoc.UpdatedBy = authUsername
	dataDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, shopID, findDoc.GuidFixed, dataDoc)

	if err != nil {
		return err
	}

	go func() {
		err := svc.repoMq.Update(dataDoc)
		if err != nil {
			fmt.Printf("create mq error :: %s", err.Error())
		}
		svc.saveMasterSync(shopID)
	}()

	return nil
}

func (svc SaleInvoiceReturnService) DeleteSaleInvoiceReturn(shopID string, guid string, authUsername string) error {

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

func (svc SaleInvoiceReturnService) DeleteSaleInvoiceReturnByGUIDs(shopID string, authUsername string, GUIDs []string) error {

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

func (svc SaleInvoiceReturnService) InfoSaleInvoiceReturn(shopID string, guid string) (models.SaleInvoiceReturnInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return models.SaleInvoiceReturnInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.SaleInvoiceReturnInfo{}, errors.New("document not found")
	}

	return findDoc.SaleInvoiceReturnInfo, nil
}

func (svc SaleInvoiceReturnService) InfoSaleInvoiceReturnByCode(shopID string, code string) (models.SaleInvoiceReturnInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "docno", code)

	if err != nil {
		return models.SaleInvoiceReturnInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.SaleInvoiceReturnInfo{}, errors.New("document not found")
	}

	return findDoc.SaleInvoiceReturnInfo, nil
}

func (svc SaleInvoiceReturnService) GetLastPOSDocNo(shopID, posID, maxDocNo string) (string, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	lastDocNo, err := svc.repo.FindLastPOSDocNo(ctx, shopID, posID, maxDocNo)

	if err != nil {
		return "", err
	}

	return lastDocNo, nil
}

func (svc SaleInvoiceReturnService) SearchSaleInvoiceReturn(shopID string, filters map[string]interface{}, pageable micro_models.Pageable) ([]models.SaleInvoiceReturnInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"docno",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, shopID, filters, searchInFields, pageable)

	if err != nil {
		return []models.SaleInvoiceReturnInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc SaleInvoiceReturnService) SearchSaleInvoiceReturnStep(shopID string, langCode string, filters map[string]interface{}, pageableStep micro_models.PageableStep) ([]models.SaleInvoiceReturnInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"docno",
	}

	selectFields := map[string]interface{}{}

	docList, total, err := svc.repo.FindStep(ctx, shopID, filters, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.SaleInvoiceReturnInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc SaleInvoiceReturnService) SaveInBatch(shopID string, authUsername string, dataList []models.SaleInvoiceReturn) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.SaleInvoiceReturn](dataList, svc.getDocIDKey)

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

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.SaleInvoiceReturn, models.SaleInvoiceReturnDoc](
		shopID,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(shopID string, authUsername string, doc models.SaleInvoiceReturn) models.SaleInvoiceReturnDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.SaleInvoiceReturnDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.ShopID = shopID
			dataDoc.SaleInvoiceReturn = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.SaleInvoiceReturn, models.SaleInvoiceReturnDoc](
		shopID,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(shopID string, guid string) (models.SaleInvoiceReturnDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, shopID, "docno", guid)
		},
		func(doc models.SaleInvoiceReturnDoc) bool {
			return doc.DocNo != ""
		},
		func(shopID string, authUsername string, data models.SaleInvoiceReturn, doc models.SaleInvoiceReturnDoc) error {

			doc.SaleInvoiceReturn = data
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

func (svc SaleInvoiceReturnService) getDocIDKey(doc models.SaleInvoiceReturn) string {
	return doc.DocNo
}

func (svc SaleInvoiceReturnService) saveMasterSync(shopID string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(shopID, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc SaleInvoiceReturnService) GetModuleName() string {
	return "saleInvoiceReturn"
}
