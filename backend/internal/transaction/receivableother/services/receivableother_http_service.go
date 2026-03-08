package services

import (
	"context"
	"errors"
	"fmt"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/services"
	"smlcloudplatform/internal/transaction/receivableother/models"
	"smlcloudplatform/internal/transaction/receivableother/repositories"
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

type IReceivableOtherHttpService interface {
	CreateReceivableOther(shopID string, authUsername string, doc models.ReceivableOther) (string, string, error)
	UpdateReceivableOther(shopID string, guid string, authUsername string, doc models.ReceivableOther) error
	DeleteReceivableOther(shopID string, guid string, authUsername string) error
	DeleteReceivableOtherByGUIDs(shopID string, authUsername string, GUIDs []string) error
	InfoReceivableOther(shopID string, guid string) (models.ReceivableOtherInfo, error)
	InfoReceivableOtherByCode(shopID string, code string) (models.ReceivableOtherInfo, error)
	SearchReceivableOther(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ReceivableOtherInfo, mongopagination.PaginationData, error)
	SearchReceivableOtherStep(shopID string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ReceivableOtherInfo, int, error)
	SaveInBatch(shopID string, authUsername string, dataList []models.ReceivableOther) (common.BulkImport, error)

	GetModuleName() string
}

const (
	MODULE_NAME = "AOB"
	TRANS_FLAG  = 99
)

type ReceivableOtherHttpService struct {
	repo             repositories.IReceivableOtherRepository
	repoCache        trancache.ICacheRepository
	repoMq           repositories.IReceivableOtherMessageQueueRepository
	cacheExpireDocNo time.Duration
	syncCacheRepo    mastersync.IMasterSyncCacheRepository
	services.ActivityService[models.ReceivableOtherActivity, models.ReceivableOtherDeleteActivity]
	contextTimeout time.Duration
}

func NewReceivableOtherHttpService(repo repositories.IReceivableOtherRepository, repoMq repositories.IReceivableOtherMessageQueueRepository, repoCache trancache.ICacheRepository, syncCacheRepo mastersync.IMasterSyncCacheRepository) *ReceivableOtherHttpService {

	contextTimeout := time.Duration(15) * time.Second

	insSvc := &ReceivableOtherHttpService{
		repo:             repo,
		repoCache:        repoCache,
		repoMq:           repoMq,
		syncCacheRepo:    syncCacheRepo,
		cacheExpireDocNo: time.Hour * 24,
		contextTimeout:   contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.ReceivableOtherActivity, models.ReceivableOtherDeleteActivity](repo)

	return insSvc
}

func (svc ReceivableOtherHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc ReceivableOtherHttpService) getDocNoPrefix(docDate time.Time) string {
	docDateStr := docDate.Format("20060102")
	return fmt.Sprintf("%s%s", MODULE_NAME, docDateStr)
}

func (svc ReceivableOtherHttpService) generateNewDocNo(ctx context.Context, shopID, prefixDocNo string, docNumber int) (string, int, error) {
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

func (svc ReceivableOtherHttpService) CreateReceivableOther(shopID string, authUsername string, doc models.ReceivableOther) (string, string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	docDate := doc.DocDatetime
	prefixDocNo := svc.getDocNoPrefix(docDate)

	newDocNo, newDocNumber, err := svc.generateNewDocNo(ctx, shopID, prefixDocNo, 1)

	if err != nil {
		return "", "", err
	}

	newGuidFixed := utils.NewGUID()

	docData := models.ReceivableOtherDoc{}
	docData.ShopID = shopID
	docData.GuidFixed = newGuidFixed
	docData.ReceivableOther = doc

	docData.DocNo = newDocNo
	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return "", "", err
	}

	go svc.repoCache.Save(shopID, prefixDocNo, newDocNumber, svc.cacheExpireDocNo)

	svc.saveMasterSync(shopID)

	go func() {
		svc.repoMq.Create(docData)
	}()

	return newGuidFixed, newDocNo, nil
}

func (svc ReceivableOtherHttpService) UpdateReceivableOther(shopID string, guid string, authUsername string, doc models.ReceivableOther) error {

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
	dataDoc.ReceivableOther = doc

	dataDoc.DocNo = findDoc.DocNo
	dataDoc.UpdatedBy = authUsername
	dataDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, shopID, guid, dataDoc)

	if err != nil {
		return err
	}

	svc.saveMasterSync(shopID)

	go func() {
		svc.repoMq.Update(findDoc)
	}()

	return nil
}

func (svc ReceivableOtherHttpService) DeleteReceivableOther(shopID string, guid string, authUsername string) error {

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

	svc.saveMasterSync(shopID)

	go func() {
		svc.repoMq.Delete(findDoc)
	}()

	return nil
}

func (svc ReceivableOtherHttpService) DeleteReceivableOtherByGUIDs(shopID string, authUsername string, GUIDs []string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	deleteFilterQuery := map[string]interface{}{
		"guidfixed": bson.M{"$in": GUIDs},
	}

	err := svc.repo.Delete(ctx, shopID, authUsername, deleteFilterQuery)
	if err != nil {
		return err
	}

	return nil
}

func (svc ReceivableOtherHttpService) InfoReceivableOther(shopID string, guid string) (models.ReceivableOtherInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return models.ReceivableOtherInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.ReceivableOtherInfo{}, errors.New("document not found")
	}

	return findDoc.ReceivableOtherInfo, nil
}

func (svc ReceivableOtherHttpService) InfoReceivableOtherByCode(shopID string, code string) (models.ReceivableOtherInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "docno", code)

	if err != nil {
		return models.ReceivableOtherInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.ReceivableOtherInfo{}, errors.New("document not found")
	}

	return findDoc.ReceivableOtherInfo, nil
}

func (svc ReceivableOtherHttpService) SearchReceivableOther(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ReceivableOtherInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"docno",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, shopID, filters, searchInFields, pageable)

	if err != nil {
		return []models.ReceivableOtherInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc ReceivableOtherHttpService) SearchReceivableOtherStep(shopID string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ReceivableOtherInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"docno",
	}

	selectFields := map[string]interface{}{}

	docList, total, err := svc.repo.FindStep(ctx, shopID, filters, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.ReceivableOtherInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc ReceivableOtherHttpService) SaveInBatch(shopID string, authUsername string, dataList []models.ReceivableOther) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.ReceivableOther](dataList, svc.getDocIDKey)

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

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.ReceivableOther, models.ReceivableOtherDoc](
		shopID,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(shopID string, authUsername string, doc models.ReceivableOther) models.ReceivableOtherDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.ReceivableOtherDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.ShopID = shopID
			dataDoc.ReceivableOther = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.ReceivableOther, models.ReceivableOtherDoc](
		shopID,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(shopID string, guid string) (models.ReceivableOtherDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, shopID, "docno", guid)
		},
		func(doc models.ReceivableOtherDoc) bool {
			return doc.DocNo != ""
		},
		func(shopID string, authUsername string, data models.ReceivableOther, doc models.ReceivableOtherDoc) error {

			doc.ReceivableOther = data
			doc.UpdatedBy = authUsername
			doc.UpdatedAt = time.Now()

			err = svc.repo.Update(ctx, shopID, doc.GuidFixed, doc)
			if err != nil {
				return nil
			}

			go func() {
				svc.repoMq.Update(doc)
			}()

			return nil
		},
	)

	if len(createDataList) > 0 {
		err = svc.repo.CreateInBatch(ctx, createDataList)

		if err != nil {
			return common.BulkImport{}, err
		}

		go func() {
			svc.repoMq.CreateInBatch(createDataList)
		}()

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

func (svc ReceivableOtherHttpService) getDocIDKey(doc models.ReceivableOther) string {
	return doc.DocNo
}

func (svc ReceivableOtherHttpService) saveMasterSync(shopID string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(shopID, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc ReceivableOtherHttpService) GetModuleName() string {
	return "receivableother"
}
