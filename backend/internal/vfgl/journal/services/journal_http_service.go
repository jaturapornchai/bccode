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
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IJournalHttpService interface {
	CreateJournal(shopID string, authUsername string, doc models.Journal) (string, error)
	RebuildPgJournal(shopID string, authUsername string) ([]models.JournalDoc, error)
	UpdateJournal(guid string, shopID string, authUsername string, doc models.Journal) (oldDocNo string, newDocNo string, err error)
	DeleteJournal(guid string, shopID string, authUsername string) error
	DeleteJournalByGUIDs(shopID string, authUsername string, GUIDs []string) error
	DeleteJournalByBatchID(shopID string, authUsername string, batchID string) error
	InfoJournal(shopID string, guid string) (models.JournalInfo, error)
	InfoJournalByDocNo(shopID string, docNo string) (models.JournalInfo, error)
	InfoJournalByDocumentRef(shopID string, documentRef string) (models.JournalInfo, error)
	SearchJournal(shopID string, pagable micromodels.Pageable, searchFilters map[string]interface{}, startDate time.Time, endDate time.Time, accountGroup string) ([]models.JournalInfo, mongopagination.PaginationData, error)
	SaveInBatch(shopID string, authUsername string, dataList []models.Journal) (common.BulkImport, error)
	GetDuplicateDocNos(shopID string) ([]models.DuplicateDocNo, error)
	CheckVatDocNoExists(shopID string, debtType int, code string, vatDocNo string) (models.VatDocNoCheckResult, error)
	CheckTaxDocNoExists(shopID string, debtType int, code string, taxDocNo string) (models.VatDocNoCheckResult, error)

	FindLastDocnoFromFormat(shopID string, docFormat string) (string, error)
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

func (svc JournalHttpService) CreateJournal(shopID string, authUsername string, doc models.Journal) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "docno", doc.DocNo)

	if err != nil {
		return "", err
	}

	if findDoc.DocNo != "" {
		return "", errors.New("docno is exists")
	}

	newGuidFixed := utils.NewGUID()

	docData := models.JournalDoc{}
	docData.ShopID = shopID
	docData.GuidFixed = newGuidFixed
	docData.Journal = doc

	// docDate := doc.DocDate.Format("2006-01-02")
	docData.DocDate = time.Date(doc.DocDate.Year(), doc.DocDate.Month(), doc.DocDate.Day(), 0, 0, 0, 0, time.UTC)

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	svc.mqRepo.Create(docData)

	if err != nil {
		return "", err
	}
	return newGuidFixed, nil
}
func (svc JournalHttpService) RebuildPgJournal(shopID string, authUsername string) ([]models.JournalDoc, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.Find(ctx, shopID, []string{}, "")

	if err != nil {
		return []models.JournalDoc{}, err
	}
	docDatax := []models.JournalDoc{}
	// loop for  findDoc
	for _, doc := range findDoc {

		docData := models.JournalDoc{}
		docData.ShopID = shopID
		docData.GuidFixed = doc.GuidFixed
		// docData.Journal = doc.Journal
		docData.AccountBook = doc.AccountBook
		docData.AccountGroup = doc.AccountGroup
		docData.DocNo = doc.DocNo
		docData.DocumentRef = doc.DocumentRef
		docData.BatchID = doc.BatchID
		docData.DocDate = doc.DocDate
		docData.CreatedAt = doc.CreatedAt
		docData.CreatedBy = doc.CreatedBy
		docData.AccountPeriod = doc.AccountPeriod
		docData.AccountYear = doc.AccountYear
		docData.Amount = doc.Amount
		docData.AccountDescription = doc.AccountDescription
		docData.BookCode = doc.BookCode
		docData.Vats = doc.Vats
		docData.Taxes = doc.Taxes
		docData.JournalType = doc.JournalType
		docData.ExDocRefNo = doc.ExDocRefNo
		docData.ExDocRefDate = doc.ExDocRefDate
		docData.DocFormat = doc.DocFormat
		docData.AppName = doc.AppName
		docData.DebtAccountType = doc.DebtAccountType
		docData.Creditor = doc.Creditor
		docData.Debtor = doc.Debtor
		docDatax = append(docDatax, docData)
		svc.mqRepo.Create(docData)

		if err != nil {
			return []models.JournalDoc{}, err
		}
	}
	return docDatax, nil
}

func (svc JournalHttpService) UpdateJournal(guid string, shopID string, authUsername string, doc models.Journal) (oldDocNo string, newDocNo string, err error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return "", "", err
	}

	if findDoc.ID == primitive.NilObjectID {
		return "", "", errors.New("document not found")
	}

	oldDocNo = findDoc.DocNo
	newDocNo = doc.DocNo

	findDoc.Journal = doc
	findDoc.UpdatedBy = authUsername
	findDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, shopID, guid, findDoc)

	if err != nil {
		return "", "", err
	}
	svc.mqRepo.Update(findDoc)

	if err != nil {
		return "", "", err
	}
	return oldDocNo, newDocNo, nil
}

func (svc JournalHttpService) DeleteJournal(guid string, shopID string, authUsername string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	err = svc.repo.DeleteByGuidfixed(ctx, shopID, guid, authUsername)
	if err != nil {
		return err
	}
	svc.mqRepo.Delete(findDoc)

	if err != nil {
		return err
	}
	return nil
}

func (svc JournalHttpService) DeleteJournalByGUIDs(shopID string, authUsername string, GUIDs []string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	docs, _ := svc.repo.FindByGuids(ctx, shopID, GUIDs)

	deleteFilterQuery := map[string]interface{}{
		"guidfixed": bson.M{"$in": GUIDs},
	}

	err := svc.repo.Delete(ctx, shopID, authUsername, deleteFilterQuery)
	if err != nil {
		return err
	}

	func() {

		svc.mqRepo.DeleteInBatch(docs)
	}()

	return nil
}

func (svc JournalHttpService) DeleteJournalByBatchID(shopID string, authUsername string, batchID string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDocs, err := svc.repo.FindFilter(ctx, shopID, bson.M{"batchid": batchID})

	if err != nil {
		return err
	}

	if len(findDocs) == 0 {
		return errors.New("document not found")
	}

	err = svc.repo.Delete(ctx, shopID, authUsername, map[string]interface{}{"batchid": batchID})
	if err != nil {
		return err
	}

	err = svc.mqRepo.DeleteInBatch(findDocs)

	if err != nil {
		return err
	}
	return nil
}

func (svc JournalHttpService) InfoJournal(shopID string, guid string) (models.JournalInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return models.JournalInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.JournalInfo{}, errors.New("document not found")
	}

	findDoc.JournalInfo.CreatedBy = findDoc.ActivityDoc.CreatedBy
	findDoc.JournalInfo.CreatedAt = findDoc.ActivityDoc.CreatedAt

	return findDoc.JournalInfo, nil

}

func (svc JournalHttpService) InfoJournalByDocNo(shopID string, docNo string) (models.JournalInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	filters := bson.M{"docno": docNo}

	findDoc, err := svc.repo.FindOne(ctx, shopID, filters)

	if err != nil {
		return models.JournalInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.JournalInfo{}, errors.New("document not found")
	}

	return findDoc.JournalInfo, nil

}

func (svc JournalHttpService) InfoJournalByDocumentRef(shopID string, documentRef string) (models.JournalInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	filters := bson.M{
		"documentref": documentRef,
	}

	findDoc, err := svc.repo.FindOne(ctx, shopID, filters)

	if err != nil {
		return models.JournalInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.JournalInfo{}, errors.New("document not found")
	}

	return findDoc.JournalInfo, nil

}

func (svc JournalHttpService) SearchJournal(shopID string, pageable micromodels.Pageable, searchFilters map[string]interface{}, startDate time.Time, endDate time.Time, accountGroup string) ([]models.JournalInfo, mongopagination.PaginationData, error) {

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

	docList, pagination, err := svc.repo.FindPageFilter(ctx, shopID, filters, searchInFields, pageable)

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

func (svc JournalHttpService) SaveInBatch(shopID string, authUsername string, dataList []models.Journal) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.Journal](dataList, svc.getDocIDKey)

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

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.Journal, models.JournalDoc](
		shopID,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(shopID string, authUsername string, doc models.Journal) models.JournalDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.JournalDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.ShopID = shopID
			dataDoc.Journal = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.Journal, models.JournalDoc](
		shopID,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(shopID string, guid string) (models.JournalDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, shopID, "docno", guid)
		},
		func(doc models.JournalDoc) bool {
			if doc.DocNo != "" {
				return true
			}
			return false
		},
		func(shopID string, authUsername string, data models.Journal, doc models.JournalDoc) error {

			doc.Journal = data
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

		svc.mqRepo.CreateInBatch(createDataList)

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
		svc.mqRepo.Update(doc)
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

func (svc JournalHttpService) FindLastDocnoFromFormat(shopID string, docFormat string) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	lastDocNo, err := svc.repo.FindLastDocno(ctx, shopID, docFormat)

	if err != nil {
		return "", err
	}

	return lastDocNo, nil

}

func (svc JournalHttpService) GetDuplicateDocNos(shopID string) ([]models.DuplicateDocNo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	duplicates, err := svc.repo.GetDuplicateDocNos(ctx, shopID)

	if err != nil {
		return []models.DuplicateDocNo{}, err
	}

	return duplicates, nil
}

func (svc JournalHttpService) CheckVatDocNoExists(shopID string, debtType int, code string, vatDocNo string) (models.VatDocNoCheckResult, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	exists, err := svc.repo.CheckVatDocNoExists(ctx, shopID, debtType, code, vatDocNo)
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

func (svc JournalHttpService) CheckTaxDocNoExists(shopID string, debtType int, code string, taxDocNo string) (models.VatDocNoCheckResult, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	exists, err := svc.repo.CheckTaxDocNoExists(ctx, shopID, debtType, code, taxDocNo)
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
