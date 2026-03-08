package services

import (
	"context"
	"errors"
	"fmt"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/productsection/sectiondepartment/models"
	"smlcloudplatform/internal/productsection/sectiondepartment/repositories"
	"smlcloudplatform/internal/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/importdata"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type ISectionDepartmentHttpService interface {
	SaveSectionDepartment(shopID string, authUsername string, doc models.SectionDepartment) (string, error)
	DeleteSectionDepartment(shopID string, guid string, authUsername string) error
	DeleteSectionDepartmentByGUIDs(shopID string, authUsername string, GUIDs []string) error
	InfoSectionDepartment(shopID string, guid string) (models.SectionDepartmentInfo, error)
	InfoSectionDepartmentByCode(shopID, branchCode, departmentCode string) (models.SectionDepartmentInfo, error)
	SearchSectionDepartment(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.SectionDepartmentInfo, mongopagination.PaginationData, error)
	SearchSectionDepartmentStep(shopID string, langCode string, pageableStep micromodels.PageableStep) ([]models.SectionDepartmentInfo, int, error)
	SaveInBatch(shopID string, authUsername string, dataList []models.SectionDepartment) (common.BulkImport, error)

	GetModuleName() string
}

type SectionDepartmentHttpService struct {
	repo repositories.ISectionDepartmentRepository

	syncCacheRepo mastersync.IMasterSyncCacheRepository
	services.ActivityService[models.SectionDepartmentActivity, models.SectionDepartmentDeleteActivity]
	contextTimeout time.Duration
}

func NewSectionDepartmentHttpService(repo repositories.ISectionDepartmentRepository, syncCacheRepo mastersync.IMasterSyncCacheRepository) *SectionDepartmentHttpService {

	contextTimeout := time.Duration(15) * time.Second

	insSvc := &SectionDepartmentHttpService{
		repo:           repo,
		syncCacheRepo:  syncCacheRepo,
		contextTimeout: contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.SectionDepartmentActivity, models.SectionDepartmentDeleteActivity](repo)

	return insSvc
}

func (svc SectionDepartmentHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc SectionDepartmentHttpService) SaveSectionDepartment(shopID string, authUsername string, doc models.SectionDepartment) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindOneFilter(
		ctx,
		shopID,
		map[string]interface{}{
			"branchcode":     doc.BranchCode,
			"departmentcode": doc.DepartmentCode,
		},
	)

	if err != nil {
		return "", err
	}

	guidFixed := ""
	if len(findDoc.GuidFixed) < 1 {
		guidFixed, err = svc.create(ctx, findDoc, shopID, authUsername, doc)
	} else {
		err = svc.update(ctx, findDoc, shopID, authUsername, doc)
		guidFixed = findDoc.GuidFixed
	}

	if err != nil {
		return "", err
	}

	return guidFixed, nil
}

func (svc SectionDepartmentHttpService) create(ctx context.Context, findDoc models.SectionDepartmentDoc, shopID, authUsername string, doc models.SectionDepartment) (string, error) {

	newGuidFixed := utils.NewGUID()

	docData := models.SectionDepartmentDoc{}
	docData.ShopID = shopID
	docData.GuidFixed = newGuidFixed
	docData.SectionDepartment = doc

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err := svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	svc.saveMasterSync(shopID)

	return newGuidFixed, nil
}

func (svc SectionDepartmentHttpService) update(ctx context.Context, findDoc models.SectionDepartmentDoc, shopID, authUsername string, doc models.SectionDepartment) error {

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	findDoc.SectionDepartment = doc

	findDoc.UpdatedBy = authUsername
	findDoc.UpdatedAt = time.Now()

	err := svc.repo.Update(ctx, shopID, findDoc.GuidFixed, findDoc)

	if err != nil {
		return err
	}

	svc.saveMasterSync(shopID)

	return nil
}

func (svc SectionDepartmentHttpService) DeleteSectionDepartment(shopID string, guid string, authUsername string) error {

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

	return nil
}

func (svc SectionDepartmentHttpService) DeleteSectionDepartmentByGUIDs(shopID string, authUsername string, GUIDs []string) error {

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

func (svc SectionDepartmentHttpService) InfoSectionDepartment(shopID string, guid string) (models.SectionDepartmentInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return models.SectionDepartmentInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.SectionDepartmentInfo{}, errors.New("document not found")
	}

	return findDoc.SectionDepartmentInfo, nil
}

func (svc SectionDepartmentHttpService) InfoSectionDepartmentByCode(shopID, branchCode, departmentCode string) (models.SectionDepartmentInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindOneByCode(ctx, shopID, branchCode, departmentCode)

	if err != nil {
		return models.SectionDepartmentInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.SectionDepartmentInfo{}, errors.New("document not found")
	}

	return findDoc.SectionDepartmentInfo, nil
}

func (svc SectionDepartmentHttpService) SearchSectionDepartment(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.SectionDepartmentInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"departmentcode",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, shopID, filters, searchInFields, pageable)

	if err != nil {
		return []models.SectionDepartmentInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc SectionDepartmentHttpService) SearchSectionDepartmentStep(shopID string, langCode string, pageableStep micromodels.PageableStep) ([]models.SectionDepartmentInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"departmentcode",
	}

	selectFields := map[string]interface{}{}

	if langCode != "" {
		selectFields["names"] = bson.M{"$elemMatch": bson.M{"code": langCode}}
	} else {
		selectFields["names"] = 1
	}

	docList, total, err := svc.repo.FindStep(ctx, shopID, map[string]interface{}{}, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.SectionDepartmentInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc SectionDepartmentHttpService) SaveInBatch(shopID string, authUsername string, dataList []models.SectionDepartment) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.SectionDepartment](dataList, svc.getDocIDKey)

	itemCodeGuidList := []string{}
	for _, doc := range payloadList {
		itemCodeGuidList = append(itemCodeGuidList, doc.DepartmentCode)
	}

	findItemGuid, err := svc.repo.FindInItemGuid(ctx, shopID, "departmentcode", itemCodeGuidList)

	if err != nil {
		return common.BulkImport{}, err
	}

	foundItemGuidList := []string{}
	for _, doc := range findItemGuid {
		foundItemGuidList = append(foundItemGuidList, doc.DepartmentCode)
	}

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.SectionDepartment, models.SectionDepartmentDoc](
		shopID,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(shopID string, authUsername string, doc models.SectionDepartment) models.SectionDepartmentDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.SectionDepartmentDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.ShopID = shopID
			dataDoc.SectionDepartment = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.SectionDepartment, models.SectionDepartmentDoc](
		shopID,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(shopID string, guid string) (models.SectionDepartmentDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, shopID, "departmentcode", guid)
		},
		func(doc models.SectionDepartmentDoc) bool {
			return doc.DepartmentCode != ""
		},
		func(shopID string, authUsername string, data models.SectionDepartment, doc models.SectionDepartmentDoc) error {

			doc.SectionDepartment = data
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
		createDataKey = append(createDataKey, doc.DepartmentCode)
	}

	payloadDuplicateDataKey := []string{}
	for _, doc := range payloadDuplicateList {
		payloadDuplicateDataKey = append(payloadDuplicateDataKey, doc.DepartmentCode)
	}

	updateDataKey := []string{}
	for _, doc := range updateSuccessDataList {

		updateDataKey = append(updateDataKey, doc.DepartmentCode)
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

func (svc SectionDepartmentHttpService) getDocIDKey(doc models.SectionDepartment) string {
	return doc.DepartmentCode
}

func (svc SectionDepartmentHttpService) saveMasterSync(shopID string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(shopID, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc SectionDepartmentHttpService) GetModuleName() string {
	return "sectionDepartment"
}
