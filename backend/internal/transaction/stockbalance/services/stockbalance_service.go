package services

import (
	"context"
	"errors"
	"fmt"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	productbarcode_models "smlcloudplatform/internal/product/productbarcode/models"
	"smlcloudplatform/internal/services"
	trans_models "smlcloudplatform/internal/transaction/models"
	trancache "smlcloudplatform/internal/transaction/repositories"
	"smlcloudplatform/internal/transaction/stockbalance/models"
	"smlcloudplatform/internal/transaction/stockbalance/repositories"
	stockbalancedetail_services "smlcloudplatform/internal/transaction/stockbalancedetail/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/importdata"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"strconv"
	"strings"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type IStockBalanceHttpService interface {
	CreateStockBalance(shopID string, authUsername string, doc models.StockBalance) (*models.StockBalanceDoc, string, string, error)
	UpdateStockBalance(shopID string, guid string, authUsername string, doc models.StockBalance) error
	DeleteStockBalance(shopID string, guid string, authUsername string) error
	DeleteStockBalanceByGUIDs(shopID string, authUsername string, GUIDs []string) error
	InfoStockBalance(shopID string, guid string) (models.StockBalanceInfo, error)
	InfoStockBalanceByCode(shopID string, code string) (models.StockBalanceInfo, error)
	SearchStockBalance(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.StockBalanceInfo, mongopagination.PaginationData, error)
	SearchStockBalanceStep(shopID string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.StockBalanceInfo, int, error)
	SaveInBatch(shopID string, authUsername string, dataList []models.StockBalance) (common.BulkImport, error)

	GetModuleName() string

	ProduceCreateStockBalance(shopID string, doc models.StockBalanceMessage) error
}

type IStockBalanceParser interface {
	ParseProductBarcode(detail trans_models.Detail, productBarcodeInfo productbarcode_models.ProductBarcodeInfo) trans_models.Detail
}

const (
	MODULE_NAME = "IB"
)

type StockBalanceHttpService struct {
	svcStockBalanceDetail stockbalancedetail_services.IStockBalanceDetailService
	repoMq                repositories.IStockBalanceMessageQueueRepository
	repo                  repositories.IStockBalanceRepository
	repoCache             trancache.ICacheRepository
	cacheExpireDocNo      time.Duration
	syncCacheRepo         mastersync.IMasterSyncCacheRepository
	services.ActivityService[models.StockBalanceActivity, models.StockBalanceDeleteActivity]
	contextTimeout time.Duration
}

func NewStockBalanceHttpService(
	svcStockBalanceDetail stockbalancedetail_services.IStockBalanceDetailService,
	repo repositories.IStockBalanceRepository,
	repoCache trancache.ICacheRepository,
	repoMq repositories.IStockBalanceMessageQueueRepository,
	syncCacheRepo mastersync.IMasterSyncCacheRepository,
) *StockBalanceHttpService {

	contextTimeout := time.Duration(15) * time.Second

	insSvc := &StockBalanceHttpService{
		svcStockBalanceDetail: svcStockBalanceDetail,
		repoMq:                repoMq,
		repo:                  repo,
		repoCache:             repoCache,
		syncCacheRepo:         syncCacheRepo,
		cacheExpireDocNo:      time.Hour * 24,
		contextTimeout:        contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.StockBalanceActivity, models.StockBalanceDeleteActivity](repo)

	return insSvc
}

func (svc StockBalanceHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc StockBalanceHttpService) getDocNoPrefix(docDate time.Time) string {
	docDateStr := docDate.Format("20060102")
	return fmt.Sprintf("%s%s", MODULE_NAME, docDateStr)
}

func (svc StockBalanceHttpService) generateNewDocNo(ctx context.Context, shopID, prefixDocNo string, docNumber int) (string, int, error) {
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
func (svc StockBalanceHttpService) CreateStockBalance(shopID string, authUsername string, doc models.StockBalance) (*models.StockBalanceDoc, string, string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	docDate := doc.DocDatetime
	prefixDocNo := svc.getDocNoPrefix(docDate)

	newDocNo, _, err := svc.generateNewDocNo(ctx, shopID, prefixDocNo, 1)

	if err != nil {
		return nil, "", "", err
	}

	newGuidFixed := utils.NewGUID()

	docData := models.StockBalanceDoc{}
	docData.ShopID = shopID
	docData.GuidFixed = newGuidFixed
	docData.StockBalance = doc

	docData.DocNo = newDocNo
	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return nil, "", "", err
	}

	// go func() {
	// 	stockBalanceDocMessage := models.StockBalanceMessage{}
	// 	stockBalanceDocMessage.StockBalance = doc
	// 	svc.repoMq.Create(stockBalanceDocMessage)

	// 	svc.repoCache.Save(shopID, prefixDocNo, newDocNumber, svc.cacheExpireDocNo)
	// 	svc.saveMasterSync(shopID)
	// }()

	return &docData, newGuidFixed, newDocNo, nil
}

func (svc StockBalanceHttpService) UpdateStockBalance(shopID string, guid string, authUsername string, doc models.StockBalance) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	docData := findDoc
	docData.StockBalance = doc

	docData.DocNo = findDoc.DocNo
	docData.UpdatedBy = authUsername
	docData.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, shopID, guid, docData)

	if err != nil {
		return err
	}

	func() {
		stockBalanceDocMessage := models.StockBalanceMessage{}
		stockBalanceDocMessage.StockBalance = findDoc.StockBalance
		svc.repoMq.Update(stockBalanceDocMessage)
		svc.saveMasterSync(shopID)
	}()

	return nil
}

func (svc StockBalanceHttpService) DeleteStockBalance(shopID string, guid string, authUsername string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	err = svc.repo.Transaction(ctx, func(ctx context.Context) error {
		err = svc.repo.DeleteByGuidfixed(ctx, shopID, guid, authUsername)
		if err != nil {
			return err
		}

		err = svc.svcStockBalanceDetail.DeleteStockBalanceDetailByDocNo(shopID, authUsername, findDoc.DocNo)

		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	func() {
		stockBalanceDocMessage := models.StockBalanceMessage{}
		stockBalanceDocMessage.StockBalance = findDoc.StockBalance
		svc.repoMq.Delete(stockBalanceDocMessage)
		svc.saveMasterSync(shopID)
	}()

	return nil
}

func (svc StockBalanceHttpService) DeleteStockBalanceByGUIDs(shopID string, authUsername string, GUIDs []string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	// prepare items for message queue
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

	for _, doc := range docs {
		err = svc.svcStockBalanceDetail.DeleteStockBalanceDetailByDocNo(shopID, authUsername, doc.DocNo)

		if err != nil {
			return err
		}
	}

	func() {

		stockBalanceDocMessages := []models.StockBalanceMessage{}

		for _, doc := range docs {
			stockBalanceDocMessage := models.StockBalanceMessage{}
			stockBalanceDocMessage.StockBalance = doc.StockBalance
			stockBalanceDocMessages = append(stockBalanceDocMessages, stockBalanceDocMessage)
		}

		svc.repoMq.DeleteInBatch(stockBalanceDocMessages)
		svc.saveMasterSync(shopID)
	}()

	return nil
}

func (svc StockBalanceHttpService) InfoStockBalance(shopID string, guid string) (models.StockBalanceInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return models.StockBalanceInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.StockBalanceInfo{}, errors.New("document not found")
	}

	return findDoc.StockBalanceInfo, nil
}

func (svc StockBalanceHttpService) InfoStockBalanceByCode(shopID string, code string) (models.StockBalanceInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "docno", code)

	if err != nil {
		return models.StockBalanceInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.StockBalanceInfo{}, errors.New("document not found")
	}

	return findDoc.StockBalanceInfo, nil
}

func (svc StockBalanceHttpService) SearchStockBalance(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.StockBalanceInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"docno",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, shopID, filters, searchInFields, pageable)

	if err != nil {
		return []models.StockBalanceInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc StockBalanceHttpService) SearchStockBalanceStep(shopID string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.StockBalanceInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"docno",
	}

	selectFields := map[string]interface{}{}

	docList, total, err := svc.repo.FindStep(ctx, shopID, filters, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.StockBalanceInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc StockBalanceHttpService) SaveInBatch(shopID string, authUsername string, dataList []models.StockBalance) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.StockBalance](dataList, svc.getDocIDKey)

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

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.StockBalance, models.StockBalanceDoc](
		shopID,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(shopID string, authUsername string, doc models.StockBalance) models.StockBalanceDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.StockBalanceDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.ShopID = shopID
			dataDoc.StockBalance = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.StockBalance, models.StockBalanceDoc](
		shopID,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(shopID string, guid string) (models.StockBalanceDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, shopID, "docno", guid)
		},
		func(doc models.StockBalanceDoc) bool {
			return doc.DocNo != ""
		},
		func(shopID string, authUsername string, data models.StockBalance, doc models.StockBalanceDoc) error {

			doc.StockBalance = data
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

func (svc StockBalanceHttpService) getDocIDKey(doc models.StockBalance) string {
	return doc.DocNo
}

func (svc StockBalanceHttpService) saveMasterSync(shopID string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(shopID, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc StockBalanceHttpService) GetModuleName() string {
	return "stockBalance"
}

func (svc StockBalanceHttpService) ProduceCreateStockBalance(shopID string, doc models.StockBalanceMessage) error {
	svc.repoMq.Create(doc)
	return nil
}
