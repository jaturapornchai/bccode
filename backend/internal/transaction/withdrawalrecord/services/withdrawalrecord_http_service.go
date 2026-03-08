package services

import (
	"context"
	"errors"
	"fmt"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/services"
	trancache "smlcloudplatform/internal/transaction/repositories"
	"smlcloudplatform/internal/transaction/withdrawalrecord/models"
	"smlcloudplatform/internal/transaction/withdrawalrecord/repositories"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/importdata"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"strconv"
	"strings"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type IWithdrawalRecordHttpService interface {
	CreateWithdrawalRecord(shopID string, authUsername string, doc models.WithdrawalRecord) (string, string, error)
	UpdateWithdrawalRecord(shopID string, guid string, authUsername string, doc models.WithdrawalRecord) error
	DeleteWithdrawalRecord(shopID string, guid string, authUsername string) error
	DeleteWithdrawalRecordByGUIDs(shopID string, authUsername string, GUIDs []string) error
	InfoWithdrawalRecord(shopID string, guid string) (models.WithdrawalRecordInfo, error)
	InfoWithdrawalRecordByCode(shopID string, code string) (models.WithdrawalRecordInfo, error)
	SearchWithdrawalRecord(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.WithdrawalRecordInfo, mongopagination.PaginationData, error)
	SearchWithdrawalRecordStep(shopID string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.WithdrawalRecordInfo, int, error)
	SaveInBatch(shopID string, authUsername string, dataList []models.WithdrawalRecord) (common.BulkImport, error)

	GetModuleName() string
}

const (
	MODULE_NAME = "TM"
)

type WithdrawalRecordHttpService struct {
	repoMq           repositories.IWithdrawalRecordMessageQueueRepository
	repo             repositories.IWithdrawalRecordRepository
	repoCache        trancache.ICacheRepository
	cacheExpireDocNo time.Duration
	syncCacheRepo    mastersync.IMasterSyncCacheRepository
	services.ActivityService[models.WithdrawalRecordActivity, models.WithdrawalRecordDeleteActivity]
	contextTimeout time.Duration
}

func NewWithdrawalRecordHttpService(
	repo repositories.IWithdrawalRecordRepository,
	repoCache trancache.ICacheRepository,
	repoMq repositories.IWithdrawalRecordMessageQueueRepository,
	syncCacheRepo mastersync.IMasterSyncCacheRepository,
) *WithdrawalRecordHttpService {

	contextTimeout := time.Duration(15) * time.Second

	insSvc := &WithdrawalRecordHttpService{
		repo:             repo,
		repoMq:           repoMq,
		repoCache:        repoCache,
		syncCacheRepo:    syncCacheRepo,
		cacheExpireDocNo: time.Hour * 24,
		contextTimeout:   contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.WithdrawalRecordActivity, models.WithdrawalRecordDeleteActivity](repo)

	return insSvc
}

func (svc WithdrawalRecordHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc WithdrawalRecordHttpService) getDocNoPrefix(docDate time.Time) string {
	docDateStr := docDate.Format("20060102")
	return fmt.Sprintf("%s%s", MODULE_NAME, docDateStr)
}

func (svc WithdrawalRecordHttpService) generateNewDocNo(ctx context.Context, shopID, prefixDocNo string, docNumber int) (string, int, error) {
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

func (svc WithdrawalRecordHttpService) CreateWithdrawalRecord(shopID string, authUsername string, doc models.WithdrawalRecord) (string, string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	docDate := doc.DocDatetime
	prefixDocNo := svc.getDocNoPrefix(docDate)

	newDocNo, newDocNumber, err := svc.generateNewDocNo(ctx, shopID, prefixDocNo, 1)

	if err != nil {
		return "", "", err
	}

	newGuidFixed := utils.NewGUID()

	docData := models.WithdrawalRecordDoc{}
	docData.ShopID = shopID
	docData.GuidFixed = newGuidFixed
	docData.WithdrawalRecord = doc

	docData.DocNo = newDocNo
	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return "", "", err
	}

	go svc.repoCache.Save(shopID, prefixDocNo, newDocNumber, svc.cacheExpireDocNo)

	go func() {
		svc.repoMq.Create(docData)
		svc.repoCache.Save(shopID, prefixDocNo, newDocNumber, svc.cacheExpireDocNo)
		svc.saveMasterSync(shopID)
	}()

	return newGuidFixed, newDocNo, nil
}

func (svc WithdrawalRecordHttpService) UpdateWithdrawalRecord(shopID string, guid string, authUsername string, doc models.WithdrawalRecord) error {

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
	docData.WithdrawalRecord = doc

	docData.DocNo = findDoc.DocNo
	docData.UpdatedBy = authUsername
	docData.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, shopID, guid, docData)

	if err != nil {
		return err
	}

	func() {
		svc.repoMq.Update(docData)
		svc.saveMasterSync(shopID)
	}()

	return nil
}

func (svc WithdrawalRecordHttpService) DeleteWithdrawalRecord(shopID string, guid string, authUsername string) error {

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

func (svc WithdrawalRecordHttpService) DeleteWithdrawalRecordByGUIDs(shopID string, authUsername string, GUIDs []string) error {

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

func (svc WithdrawalRecordHttpService) InfoWithdrawalRecord(shopID string, guid string) (models.WithdrawalRecordInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return models.WithdrawalRecordInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.WithdrawalRecordInfo{}, errors.New("document not found")
	}

	return findDoc.WithdrawalRecordInfo, nil
}

func (svc WithdrawalRecordHttpService) InfoWithdrawalRecordByCode(shopID string, code string) (models.WithdrawalRecordInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "docno", code)

	if err != nil {
		return models.WithdrawalRecordInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.WithdrawalRecordInfo{}, errors.New("document not found")
	}

	return findDoc.WithdrawalRecordInfo, nil
}

func (svc WithdrawalRecordHttpService) SearchWithdrawalRecord(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.WithdrawalRecordInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"docno",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, shopID, filters, searchInFields, pageable)

	if err != nil {
		return []models.WithdrawalRecordInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc WithdrawalRecordHttpService) SearchWithdrawalRecordStep(shopID string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.WithdrawalRecordInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"docno",
	}

	selectFields := map[string]interface{}{}

	docList, total, err := svc.repo.FindStep(ctx, shopID, filters, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.WithdrawalRecordInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc WithdrawalRecordHttpService) SaveInBatch(shopID string, authUsername string, dataList []models.WithdrawalRecord) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.WithdrawalRecord](dataList, svc.getDocIDKey)

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

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.WithdrawalRecord, models.WithdrawalRecordDoc](
		shopID,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(shopID string, authUsername string, doc models.WithdrawalRecord) models.WithdrawalRecordDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.WithdrawalRecordDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.ShopID = shopID
			dataDoc.WithdrawalRecord = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.WithdrawalRecord, models.WithdrawalRecordDoc](
		shopID,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(shopID string, guid string) (models.WithdrawalRecordDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, shopID, "docno", guid)
		},
		func(doc models.WithdrawalRecordDoc) bool {
			return doc.DocNo != ""
		},
		func(shopID string, authUsername string, data models.WithdrawalRecord, doc models.WithdrawalRecordDoc) error {

			doc.WithdrawalRecord = data
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

func (svc WithdrawalRecordHttpService) getDocIDKey(doc models.WithdrawalRecord) string {
	return doc.DocNo
}

func (svc WithdrawalRecordHttpService) saveMasterSync(shopID string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(shopID, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc WithdrawalRecordHttpService) GetModuleName() string {
	return "withdrawalrecord"
}
