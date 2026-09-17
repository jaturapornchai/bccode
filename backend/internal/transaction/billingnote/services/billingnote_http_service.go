package services

import (
	"context"
	"errors"
	"fmt"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/services"
	"smlcloudplatform/internal/transaction/billingnote/models"
	"smlcloudplatform/internal/transaction/billingnote/repositories"
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

type IBillingNoteHttpService interface {
	CreateBillingNote(holdingCode string, authUsername string, doc models.BillingNote) (string, string, error)
	UpdateBillingNote(holdingCode string, guid string, authUsername string, doc models.BillingNote) error
	DeleteBillingNote(holdingCode string, guid string, authUsername string) error
	DeleteBillingNoteByGUIDs(holdingCode string, authUsername string, GUIDs []string) error
	InfoBillingNote(holdingCode string, guid string) (models.BillingNoteInfo, error)
	InfoBillingNoteByCode(holdingCode string, code string) (models.BillingNoteInfo, error)
	SearchBillingNote(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.BillingNoteInfo, mongopagination.PaginationData, error)
	SearchBillingNoteStep(holdingCode string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.BillingNoteInfo, int, error)
	SaveInBatch(holdingCode string, authUsername string, dataList []models.BillingNote) (common.BulkImport, error)

	GetModuleName() string
}

const (
	MODULE_NAME = "BN"
	TRANS_FLAG  = 70
)

type BillingNoteHttpService struct {
	repo             repositories.IBillingNoteRepository
	repoCache        trancache.ICacheRepository
	repoMq           repositories.IBillingNoteMessageQueueRepository
	cacheExpireDocNo time.Duration
	syncCacheRepo    mastersync.IMasterSyncCacheRepository
	services.ActivityService[models.BillingNoteActivity, models.BillingNoteDeleteActivity]
	contextTimeout time.Duration
}

func NewBillingNoteHttpService(repo repositories.IBillingNoteRepository, repoMq repositories.IBillingNoteMessageQueueRepository, repoCache trancache.ICacheRepository, syncCacheRepo mastersync.IMasterSyncCacheRepository) *BillingNoteHttpService {

	contextTimeout := time.Duration(15) * time.Second

	insSvc := &BillingNoteHttpService{
		repo:             repo,
		repoCache:        repoCache,
		repoMq:           repoMq,
		syncCacheRepo:    syncCacheRepo,
		cacheExpireDocNo: time.Hour * 24,
		contextTimeout:   contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.BillingNoteActivity, models.BillingNoteDeleteActivity](repo)

	return insSvc
}

func (svc BillingNoteHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc BillingNoteHttpService) getDocNoPrefix(docDate time.Time) string {
	docDateStr := docDate.Format("20060102")
	return fmt.Sprintf("%s%s", MODULE_NAME, docDateStr)
}

func (svc BillingNoteHttpService) generateNewDocNo(ctx context.Context, holdingCode, prefixDocNo string, docNumber int) (string, int, error) {
	prevoiusDocNumber, err := svc.repoCache.Get(holdingCode, prefixDocNo)

	if prevoiusDocNumber == 0 || err != nil {
		lastDoc, err := svc.repo.FindLastDocNo(ctx, holdingCode, prefixDocNo)

		if err != nil {
			return "", 0, err
		}

		if len(lastDoc.DocNo) > 0 {
			rawNumber := strings.Replace(lastDoc.DocNo, prefixDocNo, "", -1)
			docNumber, err = strconv.Atoi(rawNumber)

			if err != nil {
				return "", 0, err
			}
			docNumber++
		}
		newDocNo := fmt.Sprintf("%s%04d", prefixDocNo, docNumber)
		return newDocNo, docNumber, nil
	}

	docNumber = prevoiusDocNumber + 1
	newDocNo := fmt.Sprintf("%s%04d", prefixDocNo, docNumber)

	return newDocNo, docNumber, nil
}

func (svc BillingNoteHttpService) CreateBillingNote(holdingCode string, authUsername string, doc models.BillingNote) (string, string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, docNo, err := svc.findDoc(ctx, holdingCode, doc)

	if err != nil {
		return "", "", err
	}

	if len(findDoc.GuidFixed) > 0 {
		return "", "", errors.New("DocNo is exists")
	}

	newGuidFixed := utils.NewGUID()

	docData := models.BillingNoteDoc{}
	docData.HoldingCode = holdingCode
	docData.GuidFixed = newGuidFixed
	docData.BillingNote = doc

	docData.DocNo = docNo
	if docData.TransFlag == 0 {
		docData.TransFlag = TRANS_FLAG
	}
	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return "", "", err
	}

	docDate := doc.DocDatetime
	prefixDocNo := svc.getDocNoPrefix(docDate)
	if strings.HasPrefix(docNo, prefixDocNo) {
		rawNumber := strings.Replace(docNo, prefixDocNo, "", -1)
		if num, parseErr := strconv.Atoi(rawNumber); parseErr == nil {
			go svc.repoCache.Save(holdingCode, prefixDocNo, num, svc.cacheExpireDocNo)
		}
	}

	svc.saveMasterSync(holdingCode)

	go func() {
		svc.repoMq.Create(docData)
	}()

	return newGuidFixed, docNo, nil
}

func (svc BillingNoteHttpService) findDoc(ctx context.Context, holdingCode string, doc models.BillingNote) (models.BillingNoteDoc, string, error) {
	docNo := doc.DocNo

	if len(docNo) == 0 {
		docDate := doc.DocDatetime
		prefixDocNo := svc.getDocNoPrefix(docDate)

		newDocNo, _, err := svc.generateNewDocNo(ctx, holdingCode, prefixDocNo, 1)

		if err != nil {
			return models.BillingNoteDoc{}, "", err
		}
		docNo = newDocNo
	}

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "docno", docNo)

	if err != nil {
		return models.BillingNoteDoc{}, "", err
	}

	return findDoc, docNo, nil
}

func (svc BillingNoteHttpService) UpdateBillingNote(holdingCode string, guid string, authUsername string, doc models.BillingNote) error {

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
	dataDoc.BillingNote = doc

	dataDoc.DocNo = findDoc.DocNo
	dataDoc.UpdatedBy = authUsername
	dataDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, holdingCode, guid, dataDoc)

	if err != nil {
		return err
	}

	svc.saveMasterSync(holdingCode)

	go func() {
		svc.repoMq.Update(findDoc)
	}()

	return nil
}

func (svc BillingNoteHttpService) DeleteBillingNote(holdingCode string, guid string, authUsername string) error {

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

	svc.saveMasterSync(holdingCode)

	go func() {
		svc.repoMq.Delete(findDoc)
	}()

	return nil
}

func (svc BillingNoteHttpService) DeleteBillingNoteByGUIDs(holdingCode string, authUsername string, GUIDs []string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	deleteFilterQuery := map[string]interface{}{
		"guidfixed": bson.M{"$in": GUIDs},
	}

	err := svc.repo.Delete(ctx, holdingCode, authUsername, deleteFilterQuery)
	if err != nil {
		return err
	}

	return nil
}

func (svc BillingNoteHttpService) InfoBillingNote(holdingCode string, guid string) (models.BillingNoteInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return models.BillingNoteInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.BillingNoteInfo{}, errors.New("document not found")
	}

	return findDoc.BillingNoteInfo, nil
}

func (svc BillingNoteHttpService) InfoBillingNoteByCode(holdingCode string, code string) (models.BillingNoteInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "docno", code)

	if err != nil {
		return models.BillingNoteInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.BillingNoteInfo{}, errors.New("document not found")
	}

	return findDoc.BillingNoteInfo, nil
}

func (svc BillingNoteHttpService) SearchBillingNote(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.BillingNoteInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"docno",
		"custcode",
		"description",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, holdingCode, filters, searchInFields, pageable)

	if err != nil {
		return []models.BillingNoteInfo{}, mongopagination.PaginationData{}, err
	}

	return docList, pagination, nil
}

func (svc BillingNoteHttpService) SearchBillingNoteStep(holdingCode string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.BillingNoteInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"docno",
		"custcode",
		"description",
	}

	selectFields := map[string]interface{}{
		"guidfixed":   1,
		"docno":       1,
		"docdatetime": 1,
		"duedate":     1,
		"transflag":   1,
		"custcode":    1,
		"custnames":   1,
		"totalamount": 1,
		"description": 1,
	}

	docList, total, err := svc.repo.FindStep(ctx, holdingCode, filters, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.BillingNoteInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc BillingNoteHttpService) SaveInBatch(holdingCode string, authUsername string, dataList []models.BillingNote) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.BillingNote](dataList, svc.getDocIDKey)

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

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.BillingNote, models.BillingNoteDoc](
		holdingCode,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(holdingCode string, authUsername string, doc models.BillingNote) models.BillingNoteDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.BillingNoteDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.HoldingCode = holdingCode
			dataDoc.BillingNote = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.BillingNote, models.BillingNoteDoc](
		holdingCode,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(holdingCode string, guid string) (models.BillingNoteDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "docno", guid)
		},
		func(doc models.BillingNoteDoc) bool {
			return doc.DocNo != ""
		},
		func(holdingCode string, authUsername string, data models.BillingNote, doc models.BillingNoteDoc) error {

			doc.BillingNote = data
			doc.UpdatedBy = authUsername
			doc.UpdatedAt = time.Now()

			err = svc.repo.Update(ctx, holdingCode, doc.GuidFixed, doc)
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

	svc.saveMasterSync(holdingCode)

	return common.BulkImport{
		Created:          createDataKey,
		Updated:          updateDataKey,
		UpdateFailed:     updateFailDataKey,
		PayloadDuplicate: payloadDuplicateDataKey,
	}, nil
}

func (svc BillingNoteHttpService) getDocIDKey(doc models.BillingNote) string {
	return doc.DocNo
}

func (svc BillingNoteHttpService) saveMasterSync(holdingCode string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(holdingCode, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc BillingNoteHttpService) GetModuleName() string {
	return "billingnote"
}
