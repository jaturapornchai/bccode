package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"smlcloudplatform/internal/logger"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	productbarcode_repositories "smlcloudplatform/internal/product/productbarcode/repositories"
	"smlcloudplatform/internal/product/unit/models"

	"smlcloudplatform/internal/product/unit/repositories"
	"smlcloudplatform/internal/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/importdata"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/samber/lo"
	"github.com/smlsoft/mongopagination"
	"github.com/xuri/excelize/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IUnitHttpService interface {
	CreateUnit(holdingCode string, authUsername string, doc models.Unit) (string, error)
	UpdateUnit(holdingCode string, guid string, authUsername string, doc models.Unit) error
	UpdateFieldUnit(holdingCode string, guid string, authUsername string, doc models.Unit) error
	DeleteUnit(holdingCode string, guid string, authUsername string) error
	DeleteUnitByGUIDs(holdingCode string, authUsername string, GUIDs []string) error
	InfoUnit(holdingCode string, guid string) (models.UnitInfo, error)
	InfoUnitWTFArray(holdingCode string, unitCodes []string) ([]interface{}, error)
	InfoWTFArrayMaster(codes []string) ([]interface{}, error)
	SearchUnit(holdingCode string, companyGuid string, codeFilters []string, pageable micromodels.Pageable) ([]models.UnitInfo, mongopagination.PaginationData, error)
	SearchUnitMultiShops(shopsID []string, codeFilters []string, pageable micromodels.Pageable) ([]models.UnitInfo, mongopagination.PaginationData, error)
	SearchUnitLimit(holdingCode string, companyGuid string, langCode string, codeFilters []string, pageableStep micromodels.PageableStep) ([]models.UnitInfo, int, error)
	SaveInBatch(holdingCode string, authUsername string, dataList []models.Unit) (common.BulkImport, error)
	GetModuleName() string
	ImportUnitsFromFile(file []byte, holdingCode string, authUsername string) (string, error)
}

type UnitHttpService struct {
	repo               repositories.IUnitRepository
	repoProductBarcode productbarcode_repositories.IProductBarcodeRepository
	syncCacheRepo      mastersync.IMasterSyncCacheRepository
	repoMessageQueue   repositories.IUnitMessageQueueRepository

	services.ActivityService[models.UnitActivity, models.UnitDeleteActivity]
	contextTimeout time.Duration
}

func NewUnitHttpService(
	repo repositories.IUnitRepository,
	repoProductBarcode productbarcode_repositories.IProductBarcodeRepository,
	repoMessageQueue repositories.IUnitMessageQueueRepository,
	syncCacheRepo mastersync.IMasterSyncCacheRepository,
) *UnitHttpService {

	contextTimeout := time.Duration(15) * time.Second

	insSvc := &UnitHttpService{
		repo:               repo,
		repoProductBarcode: repoProductBarcode,
		repoMessageQueue:   repoMessageQueue,
		syncCacheRepo:      syncCacheRepo,
		contextTimeout:     contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.UnitActivity, models.UnitDeleteActivity](repo)
	return insSvc
}

func (svc UnitHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc UnitHttpService) getContextTimeoutForImport() (context.Context, context.CancelFunc) {
	// Use longer timeout for file import operations
	importTimeout := time.Duration(300) * time.Second // 5 minutes
	return context.WithTimeout(context.Background(), importTimeout)
}

func unitDocFound(doc models.UnitDoc) bool {
	return doc.ID != primitive.NilObjectID
}

func (svc *UnitHttpService) ImportUnitsFromFile(file []byte, holdingCode string, authUsername string) (string, error) {
	f, err := excelize.OpenReader(bytes.NewReader(file))
	if err != nil {
		return "", fmt.Errorf("failed to open Excel file: %v", err)
	}

	rows, err := f.GetRows(f.GetSheetName(0))
	if err != nil {
		return "", fmt.Errorf("failed to read rows: %v", err)
	}

	if len(rows) < 2 {
		return "", fmt.Errorf("the Excel file is empty or missing headers")
	}

	headers := rows[0]
	var existingUnitCodes []string
	var newUnits []models.UnitDoc

	// Use longer timeout for import operations
	ctx, cancel := svc.getContextTimeoutForImport()
	defer cancel()

	const batchSize = 100 // Process in batches to avoid timeout

	for _, row := range rows[1:] {
		entry := map[string]string{}
		for i, value := range row {
			if i < len(headers) {
				entry[headers[i]] = value
			}
		}

		unitCode, ok := entry["code"]
		if !ok || unitCode == "" {
			continue
		}

		findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "unitcode", unitCode)
		if err != nil {
			return "", fmt.Errorf("error checking existing unit code %s: %v", unitCode, err)
		}

		if findDoc.UnitCode != "" {
			existingUnitCodes = append(existingUnitCodes, unitCode)
			continue
		}

		var names []common.NameX
		for i, lang := range headers {
			if lang != "code" && i < len(row) {
				name := row[i]
				if name != "" {
					copiedLang := lang
					copiedName := name
					names = append(names, common.NameX{
						Code:     &copiedLang,
						Name:     &copiedName,
						IsAuto:   false,
						IsDelete: false,
					})
				}
			}
		}

		newUnit := models.UnitDoc{
			UnitData: models.UnitData{

				HoldingCodeentity: common.HoldingCodeentity{HoldingCode: holdingCode},
				UnitInfo: models.UnitInfo{
					DocIdentity: common.DocIdentity{GuidFixed: utils.NewGUID()},
					Unit: models.Unit{
						UnitCode: unitCode,
						Names:    &names,
						UnitName: common.UnitName{UnitName1: "Import by " + authUsername},
					},
				},
			},
			ActivityDoc: common.ActivityDoc{
				CreatedBy: "Import by " + authUsername,
				CreatedAt: time.Now(),
			},
		}
		newUnits = append(newUnits, newUnit)

		// Process in batches to prevent timeout
		if len(newUnits) >= batchSize {
			err = svc.repo.CreateInBatch(ctx, newUnits)
			if err != nil {
				return "", fmt.Errorf("failed to insert batch of units: %v", err)
			}
			newUnits = []models.UnitDoc{} // Reset batch
		}
	}

	// Insert remaining units
	if len(newUnits) > 0 {
		err = svc.repo.CreateInBatch(ctx, newUnits)
		if err != nil {
			return "", fmt.Errorf("failed to insert remaining units: %v", err)
		}
	}

	totalProcessed := len(rows) - 1 - len(existingUnitCodes) // Total rows minus header minus existing
	return fmt.Sprintf("Import completed. Processed: %d new units, Existing unit codes skipped: %v",
		totalProcessed, existingUnitCodes), nil
}

func (svc UnitHttpService) CreateUnit(holdingCode string, authUsername string, doc models.Unit) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "unitcode", doc.UnitCode)

	if err != nil {
		return "", err
	}

	if findDoc.UnitCode != "" {
		return "", errors.New("unit code is exists")
	}

	newGuidFixed := utils.NewGUID()

	docData := models.UnitDoc{}
	docData.HoldingCode = holdingCode
	docData.GuidFixed = newGuidFixed
	docData.Unit = doc
	svc.syncUnitNames(&docData.Unit)

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	go func() {
		svc.repoMessageQueue.Create(docData)
		svc.saveMasterSync(holdingCode)
	}()
	svc.saveMasterSync(holdingCode)

	return newGuidFixed, nil
}

func (svc UnitHttpService) UpdateUnit(holdingCode string, guid string, authUsername string, doc models.Unit) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return err
	}

	if !unitDocFound(findDoc) {
		return errors.New("document not found")
	}

	tempCode := findDoc.UnitCode

	findDoc.Unit = doc
	svc.syncUnitNames(&findDoc.Unit)

	//
	findDoc.UnitCode = tempCode

	findDoc.UpdatedBy = authUsername
	findDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, holdingCode, findDoc.GuidFixed, findDoc)

	if err != nil {
		return err
	}

	go func() {
		err := svc.repoMessageQueue.Update(findDoc)

		if err != nil {
			logger.GetLogger().Error(err)
		}

		svc.saveMasterSync(holdingCode)
	}()

	return nil
}

func (svc UnitHttpService) UpdateFieldUnit(holdingCode string, guid string, authUsername string, doc models.Unit) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return err
	}

	if !unitDocFound(findDoc) {
		return errors.New("document not found")
	}

	temp := map[string]common.NameX{}

	for _, v := range *findDoc.Names {
		temp[*v.Code] = v
	}

	for _, v := range *doc.Names {
		temp[*v.Code] = v
	}

	tempNames := []common.NameX{}

	for _, v := range temp {
		tempNames = append(tempNames, v)
	}

	lo.Filter[common.NameX](tempNames, func(n common.NameX, i int) bool {
		notDelete := !n.IsDelete
		return notDelete
	})

	findDoc.Unit.Names = &tempNames
	svc.syncUnitNames(&findDoc.Unit)

	findDoc.UpdatedBy = authUsername
	findDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, holdingCode, findDoc.GuidFixed, findDoc)

	if err != nil {
		return err
	}

	go func() {

		err := svc.repoMessageQueue.Update(findDoc)

		if err != nil {
			logger.GetLogger().Error(err)
		}

		svc.saveMasterSync(holdingCode)
	}()

	return nil
}

func (svc UnitHttpService) existsUnitRefInProduct(holdingCode, unitCode string) (bool, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	docCount, err := svc.repoProductBarcode.CountByUnitCodes(ctx, holdingCode, []string{unitCode})

	if err != nil {
		return true, err
	}

	if docCount > 0 {
		return true, fmt.Errorf("unit code %s is referenced by product", unitCode)
	}

	return false, nil
}

func (svc UnitHttpService) deleteByUnitCode(holdingCode, guid, authUsername string) (models.UnitDoc, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return findDoc, err
	}

	if !unitDocFound(findDoc) {
		return findDoc, nil
	}

	existsInProduct, err := svc.existsUnitRefInProduct(holdingCode, findDoc.UnitCode)

	if existsInProduct {
		return findDoc, err
	}

	err = svc.repo.DeleteByGuidfixed(ctx, holdingCode, findDoc.GuidFixed, authUsername)
	if err != nil {
		return findDoc, err
	}

	return findDoc, nil
}

func (svc UnitHttpService) DeleteUnit(holdingCode, guid, authUsername string) error {

	doc, err := svc.deleteByUnitCode(holdingCode, guid, authUsername)

	if err != nil {
		return err
	}

	go func() {
		svc.repoMessageQueue.Delete(doc)
		svc.saveMasterSync(holdingCode)
	}()

	return nil
}

func (svc UnitHttpService) DeleteUnitByGUIDs(holdingCode, authUsername string, GUIDs []string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDocs, err := svc.repo.FindByGuids(ctx, holdingCode, GUIDs)

	if err != nil {
		return err
	}

	if len(findDocs) == 0 {
		return nil
	}

	unitCodes := []string{}
	for _, v := range findDocs {
		unitCodes = append(unitCodes, v.UnitCode)
	}

	docCount, err := svc.repoProductBarcode.CountByUnitCodes(ctx, holdingCode, unitCodes)

	if err != nil {
		return err
	}

	if docCount > 0 {
		return fmt.Errorf("unit code is referenced by product")
	}

	deleteFilterQuery := map[string]interface{}{
		"guidfixed": bson.M{"$in": GUIDs},
	}

	err = svc.repo.Delete(ctx, holdingCode, authUsername, deleteFilterQuery)
	if err != nil {
		return err
	}

	go func() {
		err := svc.repoMessageQueue.DeleteInBatch(findDocs)

		if err != nil {
			logger.GetLogger().Error(err)
		}

		svc.saveMasterSync(holdingCode)
	}()

	return nil
}

func (svc UnitHttpService) InfoUnit(holdingCode string, guid string) (models.UnitInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return models.UnitInfo{}, err
	}

	if !unitDocFound(findDoc) {
		return models.UnitInfo{}, errors.New("document not found")
	}

	return findDoc.UnitInfo, nil

}

func (svc UnitHttpService) InfoUnitWTFArray(holdingCode string, unitCodes []string) ([]interface{}, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	docList := []interface{}{}

	for _, unitCode := range unitCodes {
		findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "unitcode", unitCode)
		if err != nil || findDoc.ID == primitive.NilObjectID {
			// add item empty
			emptyDoc := models.UnitInfo{}
			emptyDoc.UnitCode = unitCode
			docList = append(docList, nil)
		} else {
			docList = append(docList, findDoc.UnitInfo)
		}
	}

	return docList, nil
}

func (svc UnitHttpService) InfoWTFArrayMaster(codes []string) ([]interface{}, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	docList := []interface{}{}

	findDocList, err := svc.repo.FindMasterInCodes(ctx, codes)

	if err != nil {
		return []interface{}{}, err
	}

	for _, code := range codes {
		findDoc, ok := lo.Find(findDocList, func(item models.UnitInfo) bool {
			return item.UnitCode == code
		})
		if !ok {
			// add item empty
			docList = append(docList, nil)
		} else {
			docList = append(docList, findDoc)
		}
	}

	return docList, nil
}

func (svc UnitHttpService) SearchUnit(holdingCode string, companyGuid string, codeFilters []string, pageable micromodels.Pageable) ([]models.UnitInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"unitcode",
		"names.name",
	}

	filters := map[string]interface{}{}
	if len(codeFilters) > 0 {
		filters["unitcode"] = bson.M{"$in": codeFilters}
	}

	if len(companyGuid) > 0 {
		filters["$or"] = []interface{}{
			bson.M{"companyguids": bson.M{"$exists": false}},
			bson.M{"companyguids": bson.M{"$size": 0}},
			bson.M{"companyguids": companyGuid},
		}
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, holdingCode, filters, searchInFields, pageable)

	if err != nil {
		return []models.UnitInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc UnitHttpService) SearchUnitMultiShops(shopsID []string, codeFilters []string, pageable micromodels.Pageable) ([]models.UnitInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"unitcode",
		"names.name",
	}

	filters := map[string]interface{}{}

	// Add holdingcode filter for multi-shop query
	if len(shopsID) > 0 {
		filters["holdingcode"] = bson.M{"$in": shopsID}
	}

	if len(codeFilters) > 0 {
		filters["unitcode"] = bson.M{"$in": codeFilters}
	}

	docList, pagination, err := svc.repo.FindPageFilterNoHoldingCode(ctx, filters, searchInFields, pageable)

	if err != nil {
		return []models.UnitInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc UnitHttpService) SearchUnitLimit(holdingCode string, companyGuid string, langCode string, codeFilters []string, pageableStep micromodels.PageableStep) ([]models.UnitInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"unitcode",
		"names.name",
	}

	selectFields := map[string]interface{}{
		"guidfixed":    1,
		"unitcode":     1,
		"companyguids": 1,
	}

	if langCode != "" {
		selectFields["names"] = bson.M{"$elemMatch": bson.M{"code": langCode}}
	} else {
		selectFields["names"] = 1
	}

	filters := map[string]interface{}{}
	if len(codeFilters) > 0 {
		filters["unitcode"] = bson.M{"$in": codeFilters}
	}

	if len(companyGuid) > 0 {
		filters["$or"] = []interface{}{
			bson.M{"companyguids": bson.M{"$exists": false}},
			bson.M{"companyguids": bson.M{"$size": 0}},
			bson.M{"companyguids": companyGuid},
		}
	}

	docList, total, err := svc.repo.FindStep(ctx, holdingCode, filters, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.UnitInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc UnitHttpService) SaveInBatch(holdingCode string, authUsername string, dataList []models.Unit) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.Unit](dataList, svc.getDocIDKey)

	itemCodeGuidList := []string{}
	for _, doc := range payloadList {
		itemCodeGuidList = append(itemCodeGuidList, doc.UnitCode)
	}

	findItemGuid, err := svc.repo.FindInItemGuid(ctx, holdingCode, "unitcode", itemCodeGuidList)

	if err != nil {
		return common.BulkImport{}, err
	}

	foundItemGuidList := []string{}
	for _, doc := range findItemGuid {
		foundItemGuidList = append(foundItemGuidList, doc.UnitCode)
	}

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.Unit, models.UnitDoc](
		holdingCode,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(holdingCode string, authUsername string, doc models.Unit) models.UnitDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.UnitDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.HoldingCode = holdingCode
			dataDoc.Unit = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.Unit, models.UnitDoc](
		holdingCode,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(holdingCode string, guid string) (models.UnitDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "unitcode", guid)
		},
		func(doc models.UnitDoc) bool {
			return doc.UnitCode != ""
		},
		func(holdingCode string, authUsername string, data models.Unit, doc models.UnitDoc) error {

			doc.Unit = data
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
		createDataKey = append(createDataKey, doc.UnitCode)
	}

	payloadDuplicateDataKey := []string{}
	for _, doc := range payloadDuplicateList {
		payloadDuplicateDataKey = append(payloadDuplicateDataKey, doc.UnitCode)
	}

	updateDataKey := []string{}
	for _, doc := range updateSuccessDataList {

		updateDataKey = append(updateDataKey, doc.UnitCode)
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

func (svc UnitHttpService) getDocIDKey(doc models.Unit) string {
	return doc.UnitCode
}

func (svc UnitHttpService) saveMasterSync(holdingCode string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(holdingCode, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc UnitHttpService) GetModuleName() string {
	return "productunit"
}

func (svc UnitHttpService) syncUnitNames(doc *models.Unit) {
	if doc.Names == nil {
		return
	}
	doc.UnitName1 = ""
	doc.UnitName2 = nil
	doc.UnitName3 = nil
	doc.UnitName4 = nil
	doc.UnitName5 = nil

	nameMap := make(map[string]string)
	for _, n := range *doc.Names {
		if n.Code != nil && n.Name != nil {
			nameMap[*n.Code] = *n.Name
		}
	}

	langs := []string{"th", "en", "lo", "my", "kh"}
	matchedNames := []string{}

	for _, lang := range langs {
		if val, ok := nameMap[lang]; ok {
			matchedNames = append(matchedNames, val)
			delete(nameMap, lang)
		}
	}

	for _, val := range nameMap {
		matchedNames = append(matchedNames, val)
	}

	if len(matchedNames) > 0 {
		doc.UnitName1 = matchedNames[0]
	}
	if len(matchedNames) > 1 {
		val := matchedNames[1]
		doc.UnitName2 = &val
	}
	if len(matchedNames) > 2 {
		val := matchedNames[2]
		doc.UnitName3 = &val
	}
	if len(matchedNames) > 3 {
		val := matchedNames[3]
		doc.UnitName4 = &val
	}
	if len(matchedNames) > 4 {
		val := matchedNames[4]
		doc.UnitName5 = &val
	}
}
