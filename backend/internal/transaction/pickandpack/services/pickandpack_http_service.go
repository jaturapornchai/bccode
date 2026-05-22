package services

import (
	"context"
	"errors"
	"fmt"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/services"
	"smlcloudplatform/internal/transaction/pickandpack/models"
	"smlcloudplatform/internal/transaction/pickandpack/repositories"
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

type IPickandpackHttpService interface {
	CreatePickandpack(shopID string, authUsername string, doc models.Pickandpack) (string, string, error)
	UpdatePickandpack(shopID string, guid string, authUsername string, doc models.Pickandpack) error
	DeletePickandpack(shopID string, guid string, authUsername string) error
	DeletePickandpackByGUIDs(shopID string, authUsername string, GUIDs []string) error
	InfoPickandpack(shopID string, guid string) (models.PickandpackInfo, error)
	InfoPickandpackByCode(shopID string, code string) (models.PickandpackInfo, error)
	SearchPickandpack(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.PickandpackInfo, mongopagination.PaginationData, error)
	SearchPickandpackStep(shopID string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.PickandpackInfo, int, error)
	SaveInBatch(shopID string, authUsername string, dataList []models.Pickandpack) (common.BulkImport, error)

	// New update methods
	UpdatePrint(shopID string, docNo string, authUsername string) error
	ConfirmPickandpack(shopID string, docNo string, authUsername string) error
	CancelPickandpack(shopID string, docNo string, authUsername string) error

	// Dashboard methods
	GetWarehouseDashboard(shopID string, whcodes []string, locationcodes []string, fromDate, toDate string) (interface{}, error)

	GetModuleName() string
}

const (
	MODULE_NAME = "PP"
)

type PickandpackHttpService struct {
	repoMq           repositories.IPickandpackMessageQueueRepository
	repo             repositories.IPickandpackRepository
	repoCache        trancache.ICacheRepository
	cacheExpireDocNo time.Duration
	syncCacheRepo    mastersync.IMasterSyncCacheRepository
	services.ActivityService[models.PickandpackActivity, models.PickandpackDeleteActivity]
	contextTimeout time.Duration
}

func NewPickandpackHttpService(
	repo repositories.IPickandpackRepository,
	repoCache trancache.ICacheRepository,
	repoMq repositories.IPickandpackMessageQueueRepository,
	syncCacheRepo mastersync.IMasterSyncCacheRepository,
) *PickandpackHttpService {

	contextTimeout := time.Duration(15) * time.Second

	insSvc := &PickandpackHttpService{
		repo:             repo,
		repoMq:           repoMq,
		repoCache:        repoCache,
		syncCacheRepo:    syncCacheRepo,
		cacheExpireDocNo: time.Hour * 24,
		contextTimeout:   contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.PickandpackActivity, models.PickandpackDeleteActivity](repo)

	return insSvc
}

func (svc PickandpackHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc PickandpackHttpService) getDocNoPrefix(docDate time.Time) string {
	docDateStr := docDate.Format("20060102")
	return fmt.Sprintf("%s%s", MODULE_NAME, docDateStr)
}

func (svc PickandpackHttpService) generateNewDocNo(ctx context.Context, shopID, prefixDocNo string, docNumber int) (string, int, error) {
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

func (svc PickandpackHttpService) CreatePickandpack(shopID string, authUsername string, doc models.Pickandpack) (string, string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	docDate := doc.DocDatetime
	prefixDocNo := svc.getDocNoPrefix(docDate)

	newDocNo, newDocNumber, err := svc.generateNewDocNo(ctx, shopID, prefixDocNo, 1)

	if err != nil {
		return "", "", err
	}

	newGuidFixed := utils.NewGUID()

	docData := models.PickandpackDoc{}
	docData.ShopID = shopID
	docData.GuidFixed = newGuidFixed
	docData.Pickandpack = doc

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

func (svc PickandpackHttpService) UpdatePickandpack(shopID string, guid string, authUsername string, doc models.Pickandpack) error {

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
	docData.Pickandpack = doc

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

func (svc PickandpackHttpService) DeletePickandpack(shopID string, guid string, authUsername string) error {

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

func (svc PickandpackHttpService) DeletePickandpackByGUIDs(shopID string, authUsername string, GUIDs []string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	deleteFilterQuery := map[string]interface{}{
		"guid_fixed": bson.M{"$in": GUIDs},
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

func (svc PickandpackHttpService) InfoPickandpack(shopID string, guid string) (models.PickandpackInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return models.PickandpackInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.PickandpackInfo{}, errors.New("document not found")
	}

	return findDoc.PickandpackInfo, nil
}

func (svc PickandpackHttpService) InfoPickandpackByCode(shopID string, code string) (models.PickandpackInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "docno", code)

	if err != nil {
		return models.PickandpackInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.PickandpackInfo{}, errors.New("document not found")
	}

	return findDoc.PickandpackInfo, nil
}

func (svc PickandpackHttpService) SearchPickandpack(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.PickandpackInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"docno",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, shopID, filters, searchInFields, pageable)

	if err != nil {
		return []models.PickandpackInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc PickandpackHttpService) SearchPickandpackStep(shopID string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.PickandpackInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"docno",
	}

	selectFields := map[string]interface{}{}

	docList, total, err := svc.repo.FindStep(ctx, shopID, filters, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.PickandpackInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc PickandpackHttpService) SaveInBatch(shopID string, authUsername string, dataList []models.Pickandpack) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.Pickandpack](dataList, svc.getDocIDKey)

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

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.Pickandpack, models.PickandpackDoc](
		shopID,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(shopID string, authUsername string, doc models.Pickandpack) models.PickandpackDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.PickandpackDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.ShopID = shopID
			dataDoc.Pickandpack = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.Pickandpack, models.PickandpackDoc](
		shopID,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(shopID string, guid string) (models.PickandpackDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, shopID, "docno", guid)
		},
		func(doc models.PickandpackDoc) bool {
			return doc.DocNo != ""
		},
		func(shopID string, authUsername string, data models.Pickandpack, doc models.PickandpackDoc) error {

			doc.Pickandpack = data
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

func (svc PickandpackHttpService) getDocIDKey(doc models.Pickandpack) string {
	return doc.DocNo
}

func (svc PickandpackHttpService) saveMasterSync(shopID string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(shopID, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

// UpdatePrint updates pickandpack print status
func (svc PickandpackHttpService) UpdatePrint(shopID string, docNo string, authUsername string) error {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	// Find document by DocNo
	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "docno", docNo)
	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	// Update print fields and pack status
	now := time.Now().UTC()
	findDoc.IsPrint = true
	findDoc.PrintAt = &now
	findDoc.PrintBy = authUsername
	findDoc.PackStatus = 1

	findDoc.UpdatedBy = authUsername
	findDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, shopID, findDoc.GuidFixed, findDoc)
	if err != nil {
		return err
	}

	go func() {
		err := svc.repoMq.Update(findDoc)
		if err != nil {
			fmt.Printf("update mq error :: %s", err.Error())
		}
		svc.saveMasterSync(shopID)
	}()

	return nil
}

// ConfirmPickandpack updates pickandpack confirm status
func (svc PickandpackHttpService) ConfirmPickandpack(shopID string, docNo string, authUsername string) error {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	// Find document by DocNo
	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "docno", docNo)
	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	// Update confirm fields and pack status
	now := time.Now().UTC()
	findDoc.IsConfirm = true
	findDoc.ConfirmAt = &now
	findDoc.ConfirmBy = authUsername
	findDoc.PackStatus = 2

	findDoc.UpdatedBy = authUsername
	findDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, shopID, findDoc.GuidFixed, findDoc)
	if err != nil {
		return err
	}

	go func() {
		err := svc.repoMq.Update(findDoc)
		if err != nil {
			fmt.Printf("update mq error :: %s", err.Error())
		}
		svc.saveMasterSync(shopID)
	}()

	return nil
}

// CancelPickandpack cancels pickandpack document
func (svc PickandpackHttpService) CancelPickandpack(shopID string, docNo string, authUsername string) error {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	// Find document by DocNo
	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "docno", docNo)
	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	// Update cancel status and pack status
	findDoc.IsCancel = true
	findDoc.PackStatus = 3

	findDoc.UpdatedBy = authUsername
	findDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, shopID, findDoc.GuidFixed, findDoc)
	if err != nil {
		return err
	}

	go func() {
		err := svc.repoMq.Update(findDoc)
		if err != nil {
			fmt.Printf("update mq error :: %s", err.Error())
		}
		svc.saveMasterSync(shopID)
	}()

	return nil
}

// GetWarehouseDashboard gets warehouse statistics for pickandpack documents
func (svc PickandpackHttpService) GetWarehouseDashboard(shopID string, whcodes []string, locationcodes []string, fromDate, toDate string) (interface{}, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	// Use repository method for aggregation
	results, err := svc.repo.AggregateWarehouseDashboard(ctx, shopID, whcodes, locationcodes, fromDate, toDate)
	if err != nil {
		return nil, err
	}

	// Process results to create dashboard structure
	warehouseStats := make(map[string]interface{})
	overallStats := map[string]int{
		"total":     0,
		"pending":   0, // packstatus = 0
		"printed":   0, // packstatus = 1
		"confirmed": 0, // packstatus = 2
		"cancelled": 0, // packstatus = 3
	}

	for _, result := range results {
		whcode := result.ID.WhCode
		locationcode := result.ID.LocationCode
		totalCount := int(result.TotalCount) // Convert int32 to int

		// Initialize warehouse if not exists
		if warehouseStats[whcode] == nil {
			warehouseStats[whcode] = map[string]interface{}{
				"whcode":    whcode,
				"whnames":   result.WhNames,
				"locations": make(map[string]interface{}),
				"totals": map[string]int{
					"total":     0,
					"pending":   0,
					"printed":   0,
					"confirmed": 0,
					"cancelled": 0,
				},
			}
		}

		// Location stats
		locationStats := map[string]interface{}{
			"locationcode":  locationcode,
			"locationnames": result.LocationNames,
			"total":         totalCount,
			"pending":       0,
			"printed":       0,
			"confirmed":     0,
			"cancelled":     0,
		}

		// Process status counts
		for _, statusCount := range result.StatusCounts {
			packstatus := statusCount.PackStatus
			count := int(statusCount.Count) // Convert int32 to int

			switch packstatus {
			case 0:
				locationStats["pending"] = count
				warehouseStats[whcode].(map[string]interface{})["totals"].(map[string]int)["pending"] += count
				overallStats["pending"] += count
			case 1:
				locationStats["printed"] = count
				warehouseStats[whcode].(map[string]interface{})["totals"].(map[string]int)["printed"] += count
				overallStats["printed"] += count
			case 2:
				locationStats["confirmed"] = count
				warehouseStats[whcode].(map[string]interface{})["totals"].(map[string]int)["confirmed"] += count
				overallStats["confirmed"] += count
			case 3:
				locationStats["cancelled"] = count
				warehouseStats[whcode].(map[string]interface{})["totals"].(map[string]int)["cancelled"] += count
				overallStats["cancelled"] += count
			}
		}

		// Update warehouse totals
		warehouseStats[whcode].(map[string]interface{})["totals"].(map[string]int)["total"] += totalCount
		overallStats["total"] += totalCount

		// Add location to warehouse
		warehouseStats[whcode].(map[string]interface{})["locations"].(map[string]interface{})[locationcode] = locationStats
	}

	// Convert warehouse stats to array
	warehouseArray := make([]interface{}, 0, len(warehouseStats))
	for _, wh := range warehouseStats {
		warehouse := wh.(map[string]interface{})
		locations := warehouse["locations"].(map[string]interface{})
		locationArray := make([]interface{}, 0, len(locations))
		for _, loc := range locations {
			locationArray = append(locationArray, loc)
		}
		warehouse["locations"] = locationArray
		warehouseArray = append(warehouseArray, warehouse)
	}

	return map[string]interface{}{
		"overallStats":   overallStats,
		"warehouseStats": warehouseArray,
		"filters": map[string]interface{}{
			"whcodes":       whcodes,
			"locationcodes": locationcodes,
			"fromDate":      fromDate,
			"toDate":        toDate,
		},
	}, nil
}

func (svc PickandpackHttpService) GetModuleName() string {
	return "pickandpack"
}
