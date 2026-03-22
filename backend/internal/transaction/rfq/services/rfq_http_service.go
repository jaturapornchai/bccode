package services

import (
	"context"
	"errors"
	"fmt"
	"math"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/services"
	"smlcloudplatform/internal/transaction/rfq/models"
	"smlcloudplatform/internal/transaction/rfq/repositories"
	"smlcloudplatform/internal/transaction/rfq/validators"
	trancache "smlcloudplatform/internal/transaction/repositories"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/importdata"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"strconv"
	"strings"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type IRFQHttpService interface {
	CreateRFQ(shopID string, authUsername string, doc models.RFQ) (string, string, *validators.ValidationResult, error)
	UpdateRFQ(shopID string, guid string, authUsername string, doc models.RFQ) (*validators.ValidationResult, error)
	DeleteRFQ(shopID string, guid string, authUsername string) error
	DeleteRFQByGUIDs(shopID string, authUsername string, GUIDs []string) error
	InfoRFQ(shopID string, guid string) (models.RFQInfo, error)
	InfoRFQByCode(shopID string, code string) (models.RFQInfo, error)
	SearchRFQ(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.RFQInfo, mongopagination.PaginationData, error)
	SearchRFQStep(shopID string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.RFQInfo, int, error)
	SaveInBatch(shopID string, authUsername string, dataList []models.RFQ) (common.BulkImport, error)
	GetModuleName() string
}

const (
	MODULE_NAME = "RFQ"
)

type RFQHttpService struct {
	repoMq           repositories.IRFQMessageQueueRepository
	repo             repositories.IRFQRepository
	repoCache        trancache.ICacheRepository
	cacheExpireDocNo time.Duration
	syncCacheRepo    mastersync.IMasterSyncCacheRepository
	services.ActivityService[models.RFQActivity, models.RFQDeleteActivity]
	contextTimeout time.Duration
}

func NewRFQHttpService(
	repo repositories.IRFQRepository,
	repoCache trancache.ICacheRepository,
	repoMq repositories.IRFQMessageQueueRepository,
	syncCacheRepo mastersync.IMasterSyncCacheRepository,
) *RFQHttpService {
	contextTimeout := time.Duration(15) * time.Second
	insSvc := &RFQHttpService{
		repo:             repo,
		repoMq:           repoMq,
		repoCache:        repoCache,
		syncCacheRepo:    syncCacheRepo,
		cacheExpireDocNo: time.Hour * 24,
		contextTimeout:   contextTimeout,
	}
	insSvc.ActivityService = services.NewActivityService[models.RFQActivity, models.RFQDeleteActivity](repo)
	return insSvc
}

func (svc RFQHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func sanitizeExchangeRate(doc *models.RFQ) {
	rate := doc.ExchangeRate
	if rate <= 0 || math.IsNaN(rate) || math.IsInf(rate, 0) {
		doc.ExchangeRate = 1.0
	}
}

func (svc RFQHttpService) getDocNoPrefix(docDateLocalStr string) string {
	docDateStr := strings.ReplaceAll(docDateLocalStr, "-", "")
	if len(docDateStr) >= 8 {
		docDateStr = docDateStr[:8]
	}
	return fmt.Sprintf("%s%s", MODULE_NAME, docDateStr)
}

func (svc RFQHttpService) generateNewDocNo(ctx context.Context, shopID, prefixDocNo string, docNumber int) (string, int, error) {
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

func (svc RFQHttpService) CreateRFQ(shopID string, authUsername string, doc models.RFQ) (string, string, *validators.ValidationResult, error) {
	validationResult := validators.ValidateRFQ(&doc)
	if !validationResult.IsValid() {
		return "", "", validationResult, nil
	}

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	prefixDocNo := svc.getDocNoPrefix(doc.DocDateLocal)
	newDocNo, newDocNumber, err := svc.generateNewDocNo(ctx, shopID, prefixDocNo, 1)
	if err != nil {
		return "", "", nil, err
	}

	newGuidFixed := utils.NewGUID()
	docData := models.RFQDoc{}
	docData.ShopID = shopID
	docData.GuidFixed = newGuidFixed
	docData.RFQ = doc
	docData.DocNo = newDocNo
	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	sanitizeExchangeRate(&docData.RFQ)

	_, err = svc.repo.Create(ctx, docData)
	if err != nil {
		return "", "", nil, err
	}

	go svc.repoCache.Save(shopID, prefixDocNo, newDocNumber, svc.cacheExpireDocNo)

	go func() {
		err := svc.repoMq.Create(docData)
		if err != nil {
			fmt.Printf("[KAFKA-ERROR] Failed to publish RFQ create message: DocNo=%s, Error=%v\n", docData.DocNo, err)
		}
		svc.repoCache.Save(shopID, prefixDocNo, newDocNumber, svc.cacheExpireDocNo)
		svc.saveMasterSync(shopID)
	}()

	return newGuidFixed, newDocNo, nil, nil
}

func (svc RFQHttpService) UpdateRFQ(shopID string, guid string, authUsername string, doc models.RFQ) (*validators.ValidationResult, error) {
	validationResult := validators.ValidateRFQ(&doc)
	if !validationResult.IsValid() {
		return validationResult, nil
	}

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)
	if err != nil {
		return nil, err
	}
	if len(findDoc.GuidFixed) < 1 {
		return nil, errors.New("document not found")
	}

	docData := findDoc
	docData.RFQ = doc
	docData.DocNo = findDoc.DocNo
	docData.UpdatedBy = authUsername
	docData.UpdatedAt = time.Now()

	sanitizeExchangeRate(&docData.RFQ)

	err = svc.repo.Update(ctx, shopID, guid, docData)
	if err != nil {
		return nil, err
	}

	func() {
		err := svc.repoMq.Update(docData)
		if err != nil {
			fmt.Printf("[KAFKA-ERROR] Failed to publish RFQ update message: DocNo=%s, Error=%v\n", docData.DocNo, err)
		}
		svc.saveMasterSync(shopID)
	}()

	return nil, nil
}

func (svc RFQHttpService) DeleteRFQ(shopID string, guid string, authUsername string) error {
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
		err := svc.repoMq.Delete(findDoc)
		if err != nil {
			fmt.Printf("[KAFKA-ERROR] Failed to publish RFQ delete message: DocNo=%s, Error=%v\n", findDoc.DocNo, err)
		}
		svc.saveMasterSync(shopID)
	}()

	return nil
}

func (svc RFQHttpService) DeleteRFQByGUIDs(shopID string, authUsername string, GUIDs []string) error {
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

func (svc RFQHttpService) InfoRFQ(shopID string, guid string) (models.RFQInfo, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)
	if err != nil {
		return models.RFQInfo{}, err
	}
	if len(findDoc.GuidFixed) < 1 {
		return models.RFQInfo{}, errors.New("document not found")
	}
	return findDoc.RFQInfo, nil
}

func (svc RFQHttpService) InfoRFQByCode(shopID string, code string) (models.RFQInfo, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "docno", code)
	if err != nil {
		return models.RFQInfo{}, err
	}
	if len(findDoc.GuidFixed) < 1 {
		return models.RFQInfo{}, errors.New("document not found")
	}
	return findDoc.RFQInfo, nil
}

func (svc RFQHttpService) SearchRFQ(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.RFQInfo, mongopagination.PaginationData, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{"docno", "refprdocno", "selectedvendor", "selectionreason"}
	docList, pagination, err := svc.repo.FindPageFilter(ctx, shopID, filters, searchInFields, pageable)
	if err != nil {
		return []models.RFQInfo{}, pagination, err
	}
	return docList, pagination, nil
}

func (svc RFQHttpService) SearchRFQStep(shopID string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.RFQInfo, int, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{"docno", "refprdocno", "selectedvendor", "selectionreason"}
	selectFields := map[string]interface{}{}
	docList, total, err := svc.repo.FindStep(ctx, shopID, filters, searchInFields, selectFields, pageableStep)
	if err != nil {
		return []models.RFQInfo{}, 0, err
	}
	return docList, total, nil
}

func (svc RFQHttpService) SaveInBatch(shopID string, authUsername string, dataList []models.RFQ) (common.BulkImport, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.RFQ](dataList, svc.getDocIDKey)

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

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.RFQ, models.RFQDoc](
		shopID, authUsername, foundItemGuidList, payloadList, svc.getDocIDKey,
		func(shopID string, authUsername string, doc models.RFQ) models.RFQDoc {
			newGuid := utils.NewGUID()
			dataDoc := models.RFQDoc{}
			dataDoc.GuidFixed = newGuid
			dataDoc.ShopID = shopID
			dataDoc.RFQ = doc
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = time.Now()
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.RFQ, models.RFQDoc](
		shopID, authUsername, duplicateDataList, svc.getDocIDKey,
		func(shopID string, guid string) (models.RFQDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, shopID, "docno", guid)
		},
		func(doc models.RFQDoc) bool {
			return doc.DocNo != ""
		},
		func(shopID string, authUsername string, data models.RFQ, doc models.RFQDoc) error {
			doc.RFQ = data
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

func (svc RFQHttpService) getDocIDKey(doc models.RFQ) string {
	return doc.DocNo
}

func (svc RFQHttpService) saveMasterSync(shopID string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(shopID, svc.GetModuleName())
		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc RFQHttpService) GetModuleName() string {
	return "rfq"
}
