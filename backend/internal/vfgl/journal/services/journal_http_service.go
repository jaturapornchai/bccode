package services

import (
	"context"
	"errors"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/importdata"
	"smlcloudplatform/internal/vfgl/journal/models"
	"smlcloudplatform/internal/vfgl/journal/repositories"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"strings"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IJournalHttpService interface {
	CreateJournal(holdingCode string, actor micromodels.UserInfo, doc models.Journal) (string, error)
	RebuildPgJournal(holdingCode string, authUsername string) ([]models.JournalDoc, error)
	UpdateJournal(guid string, holdingCode string, actor micromodels.UserInfo, doc models.Journal) (oldDocNo string, newDocNo string, err error)
	DeleteJournal(guid string, holdingCode string, actor micromodels.UserInfo) error
	DeleteJournalByGUIDs(holdingCode string, actor micromodels.UserInfo, GUIDs []string) error
	DeleteJournalByBatchID(holdingCode string, actor micromodels.UserInfo, batchID string) error
	InfoJournal(holdingCode string, guid string) (models.JournalInfo, error)
	InfoJournalByDocNo(holdingCode string, docNo string) (models.JournalInfo, error)
	InfoJournalByDocumentRef(holdingCode string, documentRef string) (models.JournalInfo, error)
	SearchJournal(holdingCode string, pagable micromodels.Pageable, searchFilters map[string]interface{}, startDate time.Time, endDate time.Time, accountGroup string) ([]models.JournalInfo, mongopagination.PaginationData, error)
	SaveInBatch(holdingCode string, actor micromodels.UserInfo, dataList []models.Journal) (common.BulkImport, error)
	GetDuplicateDocNos(holdingCode string) ([]models.DuplicateDocNo, error)
	CheckVatDocNoExists(holdingCode string, debtType int, code string, vatDocNo string) (models.VatDocNoCheckResult, error)
	CheckTaxDocNoExists(holdingCode string, debtType int, code string, taxDocNo string) (models.VatDocNoCheckResult, error)

	FindLastDocnoFromFormat(holdingCode string, docFormat string) (string, error)
	ReGenerateGuidEmpty() error
}

type JournalHttpService struct {
	repo           repositories.JournalRepository
	mqRepo         repositories.JournalMqRepository
	contextTimeout time.Duration
}

func NewJournalHttpService(repo repositories.JournalRepository, mqRepo repositories.JournalMqRepository) JournalHttpService {

	contextTimeout := time.Duration(15) * time.Second

	return JournalHttpService{
		repo:           repo,
		mqRepo:         mqRepo,
		contextTimeout: contextTimeout,
	}
}

func (svc JournalHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func journalActorName(actor micromodels.UserInfo) string {
	if name := strings.TrimSpace(actor.Name); name != "" {
		return name
	}
	return actor.Username
}

func applyJournalDeletedActor(doc *models.JournalDoc, actor micromodels.UserInfo, deletedAt time.Time) {
	actorName := journalActorName(actor)
	doc.ActivityDoc.DeletedBy = actor.Username
	doc.ActivityDoc.DeletedByName = actorName
	doc.ActivityDoc.DeletedAt = deletedAt
	doc.JournalInfo.DeletedBy = actor.Username
	doc.JournalInfo.DeletedByName = actorName
	doc.JournalInfo.DeletedAt = deletedAt
}

func (svc JournalHttpService) CreateJournal(holdingCode string, actor micromodels.UserInfo, doc models.Journal) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "docno", doc.DocNo)

	if err != nil {
		return "", err
	}

	if findDoc.DocNo != "" {
		return "", errors.New("docno is exists")
	}

	newGuidFixed := utils.NewGUID()

	docData := models.JournalDoc{}
	docData.HoldingCode = holdingCode
	docData.GuidFixed = newGuidFixed
	docData.Journal = doc

	// docDate := doc.DocDate.Format("2006-01-02")
	docData.DocDate = time.Date(doc.DocDate.Year(), doc.DocDate.Month(), doc.DocDate.Day(), 0, 0, 0, 0, time.UTC)

	docData.CreatedBy = actor.Username
	docData.CreatedByName = journalActorName(actor)
	docData.JournalInfo.CreatedBy = actor.Username
	docData.JournalInfo.CreatedByName = journalActorName(actor)
	docData.CreatedAt = time.Now()
	docData.JournalInfo.CreatedAt = docData.CreatedAt

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	err = svc.mqRepo.Create(docData)
	if err != nil {
		return "", err
	}
	return newGuidFixed, nil
}
func (svc JournalHttpService) RebuildPgJournal(holdingCode string, authUsername string) ([]models.JournalDoc, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.Find(ctx, holdingCode, []string{}, "")

	if err != nil {
		return []models.JournalDoc{}, err
	}
	docDatax := []models.JournalDoc{}
	// loop for  findDoc
	for _, doc := range findDoc {

		docData := models.JournalDoc{}
		docData.HoldingCode = holdingCode
		docData.JournalInfo = doc
		docData.ActivityDoc = common.ActivityDoc{
			CreatedBy:     doc.CreatedBy,
			CreatedByName: doc.CreatedByName,
			CreatedAt:     doc.CreatedAt,
			UpdatedBy:     doc.UpdatedBy,
			UpdatedByName: doc.UpdatedByName,
			UpdatedAt:     doc.UpdatedAt,
			DeletedBy:     doc.DeletedBy,
			DeletedByName: doc.DeletedByName,
			DeletedAt:     doc.DeletedAt,
		}
		docDatax = append(docDatax, docData)
		err = svc.mqRepo.Create(docData)
		if err != nil {
			return []models.JournalDoc{}, err
		}
	}
	return docDatax, nil
}

func (svc JournalHttpService) UpdateJournal(guid string, holdingCode string, actor micromodels.UserInfo, doc models.Journal) (oldDocNo string, newDocNo string, err error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return "", "", err
	}

	if findDoc.ID == primitive.NilObjectID {
		return "", "", errors.New("document not found")
	}

	oldDocNo = findDoc.DocNo
	newDocNo = doc.DocNo

	findDoc.Journal = doc
	findDoc.UpdatedBy = actor.Username
	findDoc.UpdatedByName = journalActorName(actor)
	findDoc.JournalInfo.UpdatedBy = actor.Username
	findDoc.JournalInfo.UpdatedByName = journalActorName(actor)
	findDoc.UpdatedAt = time.Now()
	findDoc.JournalInfo.UpdatedAt = findDoc.UpdatedAt

	err = svc.repo.Update(ctx, holdingCode, guid, findDoc)

	if err != nil {
		return "", "", err
	}
	err = svc.mqRepo.Update(findDoc)
	if err != nil {
		return "", "", err
	}
	return oldDocNo, newDocNo, nil
}

func (svc JournalHttpService) DeleteJournal(guid string, holdingCode string, actor micromodels.UserInfo) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	deletedAt := time.Now().UTC()
	err = svc.repo.DeleteWithActor(ctx, holdingCode, map[string]interface{}{"guidfixed": guid}, actor, deletedAt)
	if err != nil {
		return err
	}
	applyJournalDeletedActor(&findDoc, actor, deletedAt)
	err = svc.mqRepo.Delete(findDoc)
	if err != nil {
		return err
	}
	return nil
}

func (svc JournalHttpService) DeleteJournalByGUIDs(holdingCode string, actor micromodels.UserInfo, GUIDs []string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	docs, err := svc.repo.FindByGuids(ctx, holdingCode, GUIDs)
	if err != nil {
		return err
	}

	deleteFilterQuery := map[string]interface{}{
		"guidfixed": bson.M{"$in": GUIDs},
	}

	deletedAt := time.Now().UTC()
	err = svc.repo.DeleteWithActor(ctx, holdingCode, deleteFilterQuery, actor, deletedAt)
	if err != nil {
		return err
	}

	for idx := range docs {
		applyJournalDeletedActor(&docs[idx], actor, deletedAt)
	}
	return svc.mqRepo.DeleteInBatch(docs)
}

func (svc JournalHttpService) DeleteJournalByBatchID(holdingCode string, actor micromodels.UserInfo, batchID string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDocs, err := svc.repo.FindFilter(ctx, holdingCode, bson.M{"batchid": batchID})

	if err != nil {
		return err
	}

	if len(findDocs) == 0 {
		return errors.New("document not found")
	}

	deletedAt := time.Now().UTC()
	err = svc.repo.DeleteWithActor(ctx, holdingCode, map[string]interface{}{"batchid": batchID}, actor, deletedAt)
	if err != nil {
		return err
	}

	for idx := range findDocs {
		applyJournalDeletedActor(&findDocs[idx], actor, deletedAt)
	}
	err = svc.mqRepo.DeleteInBatch(findDocs)

	if err != nil {
		return err
	}
	return nil
}

func (svc JournalHttpService) InfoJournal(holdingCode string, guid string) (models.JournalInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return models.JournalInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.JournalInfo{}, errors.New("document not found")
	}

	findDoc.JournalInfo.CreatedBy = findDoc.ActivityDoc.CreatedBy
	findDoc.JournalInfo.CreatedByName = findDoc.ActivityDoc.CreatedByName
	findDoc.JournalInfo.CreatedAt = findDoc.ActivityDoc.CreatedAt
	findDoc.JournalInfo.UpdatedBy = findDoc.ActivityDoc.UpdatedBy
	findDoc.JournalInfo.UpdatedByName = findDoc.ActivityDoc.UpdatedByName
	findDoc.JournalInfo.UpdatedAt = findDoc.ActivityDoc.UpdatedAt

	return findDoc.JournalInfo, nil

}

func (svc JournalHttpService) InfoJournalByDocNo(holdingCode string, docNo string) (models.JournalInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	filters := bson.M{"docno": docNo}

	findDoc, err := svc.repo.FindOne(ctx, holdingCode, filters)

	if err != nil {
		return models.JournalInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.JournalInfo{}, errors.New("document not found")
	}

	return findDoc.JournalInfo, nil

}

func (svc JournalHttpService) InfoJournalByDocumentRef(holdingCode string, documentRef string) (models.JournalInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	filters := bson.M{
		"documentref": documentRef,
	}

	findDoc, err := svc.repo.FindOne(ctx, holdingCode, filters)

	if err != nil {
		return models.JournalInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.JournalInfo{}, errors.New("document not found")
	}

	return findDoc.JournalInfo, nil

}

func (svc JournalHttpService) SearchJournal(holdingCode string, pageable micromodels.Pageable, searchFilters map[string]interface{}, startDate time.Time, endDate time.Time, accountGroup string) ([]models.JournalInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"docno",
	}

	filters := map[string]interface{}{}

	if !startDate.IsZero() && !endDate.IsZero() {
		filters["docdate"] = bson.M{"$gte": startDate, "$lt": endDate}
	} else if !startDate.IsZero() {
		filters["docdate"] = bson.M{"$gte": startDate}
	} else if !endDate.IsZero() {
		filters["docdate"] = bson.M{"$lt": endDate}
	}

	if accountGroup != "" {
		filters["accountgroup"] = accountGroup
	}

	for key, value := range searchFilters {

		// Handle special fields that require custom query structure
		if key == "debtorname" || key == "creditorname" {
			if strVal, ok := value.(string); ok {
				// Use $elemMatch to search within array of names
				arrayField := "debtor.names"
				if key == "creditorname" {
					arrayField = "creditor.names"
				}

				filters[arrayField] = bson.M{
					"$elemMatch": bson.M{
						"name": bson.M{
							"$regex": primitive.Regex{
								Pattern: ".*" + strVal + ".*",
								Options: "i",
							},
						},
					},
				}
			}
			continue
		}

		// Handle regular fields
		switch tempVal := value.(type) {
		case string:
			filters[key] = bson.M{"$regex": primitive.Regex{
				Pattern: ".*" + tempVal + ".*",
				Options: "i",
			}}
		case int, int16, int32, float64, bool:
			filters[key] = value
		case time.Time:
			filters[key] = value
		}
	}

	// Map user-friendly sort fields to actual MongoDB paths for debt account names
	for i, sort := range pageable.Sorts {
		switch sort.Key {
		case "debtorname":
			// Sort by first name in debtor.names array
			pageable.Sorts[i].Key = "debtor.names.0.name"
		case "creditorname":
			// Sort by first name in creditor.names array
			pageable.Sorts[i].Key = "creditor.names.0.name"
		}
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, holdingCode, filters, searchInFields, pageable)

	if err != nil {
		return []models.JournalInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc JournalHttpService) ReGenerateGuidEmpty() error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	docs, err := svc.repo.FindGUIDEmptyAll()

	if err != nil {
		return err
	}

	for _, doc := range docs {
		newGuid := utils.NewGUID()
		err = svc.repo.UpdateGuidEmpty(ctx, doc.ID, newGuid)
		if err != nil {
			return err
		}
	}

	return nil
}

func (svc JournalHttpService) SaveInBatch(holdingCode string, actor micromodels.UserInfo, dataList []models.Journal) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.Journal](dataList, svc.getDocIDKey)

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

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.Journal, models.JournalDoc](
		holdingCode,
		actor.Username,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(holdingCode string, authUsername string, doc models.Journal) models.JournalDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.JournalDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.HoldingCode = holdingCode
			dataDoc.Journal = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedByName = journalActorName(actor)
			dataDoc.JournalInfo.CreatedBy = authUsername
			dataDoc.JournalInfo.CreatedByName = journalActorName(actor)
			dataDoc.CreatedAt = currentTime
			dataDoc.JournalInfo.CreatedAt = currentTime
			return dataDoc
		},
	)

	updatedDocs := []models.JournalDoc{}
	_, updateFailDataList := importdata.UpdateOnDuplicate[models.Journal, models.JournalDoc](
		holdingCode,
		actor.Username,
		duplicateDataList,
		svc.getDocIDKey,
		func(holdingCode string, guid string) (models.JournalDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "docno", guid)
		},
		func(doc models.JournalDoc) bool {
			if doc.DocNo != "" {
				return true
			}
			return false
		},
		func(holdingCode string, authUsername string, data models.Journal, doc models.JournalDoc) error {

			doc.Journal = data
			doc.UpdatedBy = authUsername
			doc.UpdatedByName = journalActorName(actor)
			doc.JournalInfo.UpdatedBy = authUsername
			doc.JournalInfo.UpdatedByName = journalActorName(actor)
			doc.UpdatedAt = time.Now()
			doc.JournalInfo.UpdatedAt = doc.UpdatedAt

			if err := svc.repo.Update(ctx, holdingCode, doc.GuidFixed, doc); err != nil {
				return err
			}
			updatedDocs = append(updatedDocs, doc)
			return nil
		},
	)

	if len(createDataList) > 0 {
		err = svc.repo.CreateInBatch(ctx, createDataList)

		if err != nil {
			return common.BulkImport{}, err
		}

		err = svc.mqRepo.CreateInBatch(createDataList)
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
	for _, doc := range updatedDocs {
		if err = svc.mqRepo.Update(doc); err != nil {
			return common.BulkImport{}, err
		}
		updateDataKey = append(updateDataKey, doc.DocNo)
	}

	updateFailDataKey := []string{}
	for _, doc := range updateFailDataList {
		updateFailDataKey = append(updateFailDataKey, svc.getDocIDKey(doc))
	}

	return common.BulkImport{
		Created:          createDataKey,
		Updated:          updateDataKey,
		UpdateFailed:     updateFailDataKey,
		PayloadDuplicate: payloadDuplicateDataKey,
	}, nil
}

func (svc JournalHttpService) getDocIDKey(doc models.Journal) string {
	return doc.DocNo
}

func (svc JournalHttpService) FindLastDocnoFromFormat(holdingCode string, docFormat string) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	lastDocNo, err := svc.repo.FindLastDocno(ctx, holdingCode, docFormat)

	if err != nil {
		return "", err
	}

	return lastDocNo, nil

}

func (svc JournalHttpService) GetDuplicateDocNos(holdingCode string) ([]models.DuplicateDocNo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	duplicates, err := svc.repo.GetDuplicateDocNos(ctx, holdingCode)

	if err != nil {
		return []models.DuplicateDocNo{}, err
	}

	return duplicates, nil
}

func (svc JournalHttpService) CheckVatDocNoExists(holdingCode string, debtType int, code string, vatDocNo string) (models.VatDocNoCheckResult, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	exists, err := svc.repo.CheckVatDocNoExists(ctx, holdingCode, debtType, code, vatDocNo)
	if err != nil {
		return models.VatDocNoCheckResult{}, err
	}

	if exists {
		return models.VatDocNoCheckResult{
			Available: false,
			Message:   "เลขที่ภาษีนี้ถูกใช้งานแล้ว",
		}, nil
	}

	return models.VatDocNoCheckResult{
		Available: true,
		Message:   "สามารถใช้งานได้",
	}, nil
}

func (svc JournalHttpService) CheckTaxDocNoExists(holdingCode string, debtType int, code string, taxDocNo string) (models.VatDocNoCheckResult, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	exists, err := svc.repo.CheckTaxDocNoExists(ctx, holdingCode, debtType, code, taxDocNo)
	if err != nil {
		return models.VatDocNoCheckResult{}, err
	}

	if exists {
		return models.VatDocNoCheckResult{
			Available: false,
			Message:   "เลขที่ภาษีหัก ณ ที่จ่ายนี้ถูกใช้งานแล้ว",
		}, nil
	}

	return models.VatDocNoCheckResult{
		Available: true,
		Message:   "สามารถใช้งานได้",
	}, nil
}
