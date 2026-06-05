package services

import (
	"context"
	"errors"
	"fmt"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/productsection/sectionbranch/models"
	"smlcloudplatform/internal/productsection/sectionbranch/repositories"
	"smlcloudplatform/internal/services"
	"smlcloudplatform/internal/utils/importdata"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type ISectionBranchHttpService interface {
	SaveSectionBranch(holdingCode string, authUsername string, doc models.SectionBranch) (string, error)
	DeleteSectionBranch(holdingCode string, guid string, authUsername string) error
	DeleteSectionBranchByGUIDs(holdingCode string, authUsername string, GUIDs []string) error
	InfoSectionBranch(holdingCode string, guid string) (models.SectionBranchInfo, error)
	InfoSectionBranchByBranchCode(holdingCode string, branchcode string) (models.SectionBranchInfo, error)
	SearchSectionBranch(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.SectionBranchInfo, mongopagination.PaginationData, error)
	SearchSectionBranchStep(holdingCode string, langBranchCode string, pageableStep micromodels.PageableStep) ([]models.SectionBranchInfo, int, error)
	SaveInBatch(holdingCode string, authUsername string, dataList []models.SectionBranch) (common.BulkImport, error)

	GetModuleName() string
}

type SectionBranchHttpService struct {
	repo repositories.ISectionBranchRepository

	syncCacheRepo mastersync.IMasterSyncCacheRepository
	services.ActivityService[models.SectionBranchActivity, models.SectionBranchDeleteActivity]
	contextTimeout time.Duration
	generateGUID   func() string
}

func NewSectionBranchHttpService(repo repositories.ISectionBranchRepository, generateGUID func() string, syncCacheRepo mastersync.IMasterSyncCacheRepository) *SectionBranchHttpService {

	contextTimeout := time.Duration(15) * time.Second

	insSvc := &SectionBranchHttpService{
		repo:           repo,
		syncCacheRepo:  syncCacheRepo,
		generateGUID:   generateGUID,
		contextTimeout: contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.SectionBranchActivity, models.SectionBranchDeleteActivity](repo)

	return insSvc
}

func (svc SectionBranchHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc SectionBranchHttpService) SaveSectionBranch(holdingCode string, authUsername string, doc models.SectionBranch) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "branchcode", doc.BranchCode)

	if err != nil {
		return "", err
	}

	guidFixed := ""
	if len(findDoc.GuidFixed) < 1 {
		guidFixed, err = svc.create(findDoc, holdingCode, authUsername, doc)
	} else {
		err = svc.update(findDoc, holdingCode, authUsername, doc)
		guidFixed = findDoc.GuidFixed
	}

	if err != nil {
		return "", err
	}

	return guidFixed, nil
}

func (svc SectionBranchHttpService) create(findDoc models.SectionBranchDoc, holdingCode string, authUsername string, doc models.SectionBranch) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	newGuidFixed := svc.generateGUID()

	docData := models.SectionBranchDoc{}
	docData.HoldingCode = holdingCode
	docData.GuidFixed = newGuidFixed
	docData.SectionBranch = doc

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err := svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	svc.saveMasterSync(holdingCode)

	return newGuidFixed, nil
}

func (svc SectionBranchHttpService) update(findDoc models.SectionBranchDoc, holdingCode string, authUsername string, doc models.SectionBranch) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	findDoc.SectionBranch = doc

	findDoc.UpdatedBy = authUsername
	findDoc.UpdatedAt = time.Now()

	err := svc.repo.Update(ctx, holdingCode, findDoc.GuidFixed, findDoc)

	if err != nil {
		return err
	}

	svc.saveMasterSync(holdingCode)

	return nil
}

func (svc SectionBranchHttpService) DeleteSectionBranch(holdingCode string, guid string, authUsername string) error {

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

	return nil
}

func (svc SectionBranchHttpService) DeleteSectionBranchByGUIDs(holdingCode string, authUsername string, GUIDs []string) error {

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

func (svc SectionBranchHttpService) InfoSectionBranch(holdingCode string, guid string) (models.SectionBranchInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return models.SectionBranchInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.SectionBranchInfo{}, errors.New("document not found")
	}

	return findDoc.SectionBranchInfo, nil
}

func (svc SectionBranchHttpService) InfoSectionBranchByBranchCode(holdingCode string, branchcode string) (models.SectionBranchInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "branchcode", branchcode)

	if err != nil {
		return models.SectionBranchInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.SectionBranchInfo{}, errors.New("document not found")
	}

	return findDoc.SectionBranchInfo, nil
}

func (svc SectionBranchHttpService) SearchSectionBranch(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.SectionBranchInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"branchcode",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, holdingCode, filters, searchInFields, pageable)

	if err != nil {
		return []models.SectionBranchInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc SectionBranchHttpService) SearchSectionBranchStep(holdingCode string, langBranchCode string, pageableStep micromodels.PageableStep) ([]models.SectionBranchInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"branchcode",
	}

	selectFields := map[string]interface{}{}

	if langBranchCode != "" {
		selectFields["names"] = bson.M{"$elemMatch": bson.M{"branchcode": langBranchCode}}
	} else {
		selectFields["names"] = 1
	}

	docList, total, err := svc.repo.FindStep(ctx, holdingCode, map[string]interface{}{}, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.SectionBranchInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc SectionBranchHttpService) SaveInBatch(holdingCode string, authUsername string, dataList []models.SectionBranch) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.SectionBranch](dataList, svc.getDocIDKey)

	itemBranchCodeGuidList := []string{}
	for _, doc := range payloadList {
		itemBranchCodeGuidList = append(itemBranchCodeGuidList, doc.BranchCode)
	}

	findItemGuid, err := svc.repo.FindInItemGuid(ctx, holdingCode, "branchcode", itemBranchCodeGuidList)

	if err != nil {
		return common.BulkImport{}, err
	}

	foundItemGuidList := []string{}
	for _, doc := range findItemGuid {
		foundItemGuidList = append(foundItemGuidList, doc.BranchCode)
	}

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.SectionBranch, models.SectionBranchDoc](
		holdingCode,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(holdingCode string, authUsername string, doc models.SectionBranch) models.SectionBranchDoc {
			newGuid := svc.generateGUID()

			dataDoc := models.SectionBranchDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.HoldingCode = holdingCode
			dataDoc.SectionBranch = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.SectionBranch, models.SectionBranchDoc](
		holdingCode,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(holdingCode string, guid string) (models.SectionBranchDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "branchcode", guid)
		},
		func(doc models.SectionBranchDoc) bool {
			return doc.BranchCode != ""
		},
		func(holdingCode string, authUsername string, data models.SectionBranch, doc models.SectionBranchDoc) error {

			doc.SectionBranch = data
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
		createDataKey = append(createDataKey, doc.BranchCode)
	}

	payloadDuplicateDataKey := []string{}
	for _, doc := range payloadDuplicateList {
		payloadDuplicateDataKey = append(payloadDuplicateDataKey, doc.BranchCode)
	}

	updateDataKey := []string{}
	for _, doc := range updateSuccessDataList {

		updateDataKey = append(updateDataKey, doc.BranchCode)
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

func (svc SectionBranchHttpService) getDocIDKey(doc models.SectionBranch) string {
	return doc.BranchCode
}

func (svc SectionBranchHttpService) saveMasterSync(holdingCode string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(holdingCode, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc SectionBranchHttpService) GetModuleName() string {
	return "productsectionbranch"
}
