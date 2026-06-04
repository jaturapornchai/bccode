package services

import (
	"context"
	"errors"
	"fmt"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/services"
	"smlcloudplatform/internal/transaction/creditcardwithdrawal/models"
	"smlcloudplatform/internal/transaction/creditcardwithdrawal/repositories"
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

type ICreditCardWithdrawalHttpService interface {
	CreateCreditCardWithdrawal(holdingCode string, authUsername string, doc models.CreditCardWithdrawal) (string, string, error)
	UpdateCreditCardWithdrawal(holdingCode string, guid string, authUsername string, doc models.CreditCardWithdrawal) error
	DeleteCreditCardWithdrawal(holdingCode string, guid string, authUsername string) error
	DeleteCreditCardWithdrawalByGUIDs(holdingCode string, authUsername string, GUIDs []string) error
	InfoCreditCardWithdrawal(holdingCode string, guid string) (models.CreditCardWithdrawalInfo, error)
	InfoCreditCardWithdrawalByCode(holdingCode string, code string) (models.CreditCardWithdrawalInfo, error)
	SearchCreditCardWithdrawal(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.CreditCardWithdrawalInfo, mongopagination.PaginationData, error)
	SearchCreditCardWithdrawalStep(holdingCode string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.CreditCardWithdrawalInfo, int, error)
	SaveInBatch(holdingCode string, authUsername string, dataList []models.CreditCardWithdrawal) (common.BulkImport, error)

	GetModuleName() string
}

const (
	MODULE_NAME = "CRD"
)

type CreditCardWithdrawalHttpService struct {
	repoMq           repositories.ICreditCardWithdrawalMessageQueueRepository
	repo             repositories.ICreditCardWithdrawalRepository
	repoCache        trancache.ICacheRepository
	cacheExpireDocNo time.Duration
	syncCacheRepo    mastersync.IMasterSyncCacheRepository
	services.ActivityService[models.CreditCardWithdrawalActivity, models.CreditCardWithdrawalDeleteActivity]
	contextTimeout time.Duration
}

func NewCreditCardWithdrawalHttpService(
	repo repositories.ICreditCardWithdrawalRepository,
	repoCache trancache.ICacheRepository,
	repoMq repositories.ICreditCardWithdrawalMessageQueueRepository,
	syncCacheRepo mastersync.IMasterSyncCacheRepository,
) *CreditCardWithdrawalHttpService {

	contextTimeout := time.Duration(15) * time.Second

	insSvc := &CreditCardWithdrawalHttpService{
		repo:             repo,
		repoMq:           repoMq,
		repoCache:        repoCache,
		syncCacheRepo:    syncCacheRepo,
		cacheExpireDocNo: time.Hour * 24,
		contextTimeout:   contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.CreditCardWithdrawalActivity, models.CreditCardWithdrawalDeleteActivity](repo)

	return insSvc
}

func (svc CreditCardWithdrawalHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc CreditCardWithdrawalHttpService) getDocNoPrefix(docDate time.Time) string {
	docDateStr := docDate.Format("20060102")
	return fmt.Sprintf("%s%s", MODULE_NAME, docDateStr)
}

func (svc CreditCardWithdrawalHttpService) generateNewDocNo(ctx context.Context, holdingCode, prefixDocNo string, docNumber int) (string, int, error) {
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

func (svc CreditCardWithdrawalHttpService) CreateCreditCardWithdrawal(holdingCode string, authUsername string, doc models.CreditCardWithdrawal) (string, string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	docDate := doc.DocDatetime
	prefixDocNo := svc.getDocNoPrefix(docDate)

	newDocNo, newDocNumber, err := svc.generateNewDocNo(ctx, holdingCode, prefixDocNo, 1)

	if err != nil {
		return "", "", err
	}

	newGuidFixed := utils.NewGUID()

	docData := models.CreditCardWithdrawalDoc{}
	docData.HoldingCode = holdingCode
	docData.GuidFixed = newGuidFixed
	docData.CreditCardWithdrawal = doc

	docData.DocNo = newDocNo
	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return "", "", err
	}

	go svc.repoCache.Save(holdingCode, prefixDocNo, newDocNumber, svc.cacheExpireDocNo)

	go func() {
		svc.repoMq.Create(docData)
		svc.repoCache.Save(holdingCode, prefixDocNo, newDocNumber, svc.cacheExpireDocNo)
		svc.saveMasterSync(holdingCode)
	}()

	return newGuidFixed, newDocNo, nil
}

func (svc CreditCardWithdrawalHttpService) UpdateCreditCardWithdrawal(holdingCode string, guid string, authUsername string, doc models.CreditCardWithdrawal) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	docData := findDoc
	docData.CreditCardWithdrawal = doc

	docData.DocNo = findDoc.DocNo
	docData.UpdatedBy = authUsername
	docData.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, holdingCode, guid, docData)

	if err != nil {
		return err
	}

	func() {
		svc.repoMq.Update(docData)
		svc.saveMasterSync(holdingCode)
	}()

	return nil
}

func (svc CreditCardWithdrawalHttpService) DeleteCreditCardWithdrawal(holdingCode string, guid string, authUsername string) error {

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

func (svc CreditCardWithdrawalHttpService) DeleteCreditCardWithdrawalByGUIDs(holdingCode string, authUsername string, GUIDs []string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	deleteFilterQuery := map[string]interface{}{
		"guid_fixed": bson.M{"$in": GUIDs},
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

func (svc CreditCardWithdrawalHttpService) InfoCreditCardWithdrawal(holdingCode string, guid string) (models.CreditCardWithdrawalInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return models.CreditCardWithdrawalInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.CreditCardWithdrawalInfo{}, errors.New("document not found")
	}

	return findDoc.CreditCardWithdrawalInfo, nil
}

func (svc CreditCardWithdrawalHttpService) InfoCreditCardWithdrawalByCode(holdingCode string, code string) (models.CreditCardWithdrawalInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "docno", code)

	if err != nil {
		return models.CreditCardWithdrawalInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.CreditCardWithdrawalInfo{}, errors.New("document not found")
	}

	return findDoc.CreditCardWithdrawalInfo, nil
}

func (svc CreditCardWithdrawalHttpService) SearchCreditCardWithdrawal(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.CreditCardWithdrawalInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"docno",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, holdingCode, filters, searchInFields, pageable)

	if err != nil {
		return []models.CreditCardWithdrawalInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc CreditCardWithdrawalHttpService) SearchCreditCardWithdrawalStep(holdingCode string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.CreditCardWithdrawalInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"docno",
	}

	selectFields := map[string]interface{}{}

	docList, total, err := svc.repo.FindStep(ctx, holdingCode, filters, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.CreditCardWithdrawalInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc CreditCardWithdrawalHttpService) SaveInBatch(holdingCode string, authUsername string, dataList []models.CreditCardWithdrawal) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.CreditCardWithdrawal](dataList, svc.getDocIDKey)

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

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.CreditCardWithdrawal, models.CreditCardWithdrawalDoc](
		holdingCode,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(holdingCode string, authUsername string, doc models.CreditCardWithdrawal) models.CreditCardWithdrawalDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.CreditCardWithdrawalDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.HoldingCode = holdingCode
			dataDoc.CreditCardWithdrawal = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.CreditCardWithdrawal, models.CreditCardWithdrawalDoc](
		holdingCode,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(holdingCode string, guid string) (models.CreditCardWithdrawalDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "docno", guid)
		},
		func(doc models.CreditCardWithdrawalDoc) bool {
			return doc.DocNo != ""
		},
		func(holdingCode string, authUsername string, data models.CreditCardWithdrawal, doc models.CreditCardWithdrawalDoc) error {

			doc.CreditCardWithdrawal = data
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

func (svc CreditCardWithdrawalHttpService) getDocIDKey(doc models.CreditCardWithdrawal) string {
	return doc.DocNo
}

func (svc CreditCardWithdrawalHttpService) saveMasterSync(holdingCode string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(holdingCode, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc CreditCardWithdrawalHttpService) GetModuleName() string {
	return "creditcardwithdrawal"
}
