package services

import (
	"context"
	"errors"
	"fmt"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/organization/department/models"
	"smlcloudplatform/internal/organization/department/repositories"
	"smlcloudplatform/internal/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/importdata"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type IDepartmentHttpService interface {
	CreateDepartment(holdingCode string, authUsername string, doc models.Department) (string, error)
	UpdateDepartment(holdingCode string, guid string, authUsername string, doc models.Department) error
	DeleteDepartment(holdingCode string, guid string, authUsername string) error
	DeleteDepartmentByGUIDs(holdingCode string, authUsername string, GUIDs []string) error
	InfoDepartment(holdingCode string, guid string) (models.DepartmentInfo, error)
	InfoDepartmentByCode(holdingCode, branchCode, departmentCode string) (models.DepartmentInfo, error)
	SearchDepartment(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.DepartmentInfo, mongopagination.PaginationData, error)
	SearchDepartmentStep(holdingCode string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.DepartmentInfo, int, error)
	SaveInBatch(holdingCode string, authUsername string, dataList []models.Department) (common.BulkImport, error)

	GetModuleName() string
}

type DepartmentHttpService struct {
	repo repositories.IDepartmentRepository

	syncCacheRepo mastersync.IMasterSyncCacheRepository
	services.ActivityService[models.DepartmentActivity, models.DepartmentDeleteActivity]
	contextTimeout time.Duration
}

func NewDepartmentHttpService(repo repositories.IDepartmentRepository, syncCacheRepo mastersync.IMasterSyncCacheRepository) *DepartmentHttpService {

	contextTimeout := time.Duration(15) * time.Second

	insSvc := &DepartmentHttpService{
		repo:           repo,
		syncCacheRepo:  syncCacheRepo,
		contextTimeout: contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.DepartmentActivity, models.DepartmentDeleteActivity](repo)

	return insSvc
}

func (svc DepartmentHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc DepartmentHttpService) CreateDepartment(holdingCode string, authUsername string, doc models.Department) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindOneFilter(ctx, holdingCode, departmentBranchCodeFilter(doc))

	if err != nil {
		return "", err
	}

	if findDoc.Code != "" {
		return "", errors.New("Code is exists")
	}

	newGuidFixed := utils.NewGUID()

	docData := models.DepartmentDoc{}
	docData.HoldingCode = holdingCode
	docData.GuidFixed = newGuidFixed
	docData.Department = doc

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	svc.saveMasterSync(holdingCode)

	return newGuidFixed, nil
}

func (svc DepartmentHttpService) UpdateDepartment(holdingCode string, guid string, authUsername string, doc models.Department) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	findDoc.Department = doc

	findDoc.UpdatedBy = authUsername
	findDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, holdingCode, guid, findDoc)

	if err != nil {
		return err
	}

	svc.saveMasterSync(holdingCode)

	return nil
}

func (svc DepartmentHttpService) DeleteDepartment(holdingCode string, guid string, authUsername string) error {

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

func (svc DepartmentHttpService) DeleteDepartmentByGUIDs(holdingCode string, authUsername string, GUIDs []string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	deleteFilterQuery := map[string]interface{}{
		"guid_fixed": bson.M{"$in": GUIDs},
	}

	err := svc.repo.Delete(ctx, holdingCode, authUsername, deleteFilterQuery)
	if err != nil {
		return err
	}

	return nil
}

func (svc DepartmentHttpService) InfoDepartment(holdingCode string, guid string) (models.DepartmentInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return models.DepartmentInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.DepartmentInfo{}, errors.New("document not found")
	}

	return findDoc.DepartmentInfo, nil
}

func (svc DepartmentHttpService) InfoDepartmentByCode(holdingCode, branchCode, departmentCode string) (models.DepartmentInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindOneByCode(ctx, holdingCode, branchCode, departmentCode)

	if err != nil {
		return models.DepartmentInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.DepartmentInfo{}, errors.New("document not found")
	}

	return findDoc.DepartmentInfo, nil
}

func (svc DepartmentHttpService) SearchDepartment(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.DepartmentInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, holdingCode, filters, searchInFields, pageable)

	if err != nil {
		return []models.DepartmentInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc DepartmentHttpService) SearchDepartmentStep(holdingCode string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.DepartmentInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
	}

	selectFields := map[string]interface{}{}

	docList, total, err := svc.repo.FindStep(ctx, holdingCode, filters, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.DepartmentInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc DepartmentHttpService) SaveInBatch(holdingCode string, authUsername string, dataList []models.Department) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.Department](dataList, svc.getDocIDKey)

	itemCodeGuidList := []string{}
	for _, doc := range payloadList {
		itemCodeGuidList = append(itemCodeGuidList, doc.Code)
	}

	findItemGuid, err := svc.repo.FindInItemGuid(ctx, holdingCode, "code", itemCodeGuidList)

	if err != nil {
		return common.BulkImport{}, err
	}

	foundItemGuidList := []string{}
	for _, doc := range findItemGuid {
		foundItemGuidList = append(foundItemGuidList, doc.Code)
	}

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.Department, models.DepartmentDoc](
		holdingCode,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(holdingCode string, authUsername string, doc models.Department) models.DepartmentDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.DepartmentDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.HoldingCode = holdingCode
			dataDoc.Department = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.Department, models.DepartmentDoc](
		holdingCode,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(holdingCode string, guid string) (models.DepartmentDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "code", guid)
		},
		func(doc models.DepartmentDoc) bool {
			return doc.Code != ""
		},
		func(holdingCode string, authUsername string, data models.Department, doc models.DepartmentDoc) error {

			doc.Department = data
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
		createDataKey = append(createDataKey, doc.Code)
	}

	payloadDuplicateDataKey := []string{}
	for _, doc := range payloadDuplicateList {
		payloadDuplicateDataKey = append(payloadDuplicateDataKey, doc.Code)
	}

	updateDataKey := []string{}
	for _, doc := range updateSuccessDataList {

		updateDataKey = append(updateDataKey, doc.Code)
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

func (svc DepartmentHttpService) getDocIDKey(doc models.Department) string {
	return doc.Code
}

func (svc DepartmentHttpService) saveMasterSync(holdingCode string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(holdingCode, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc DepartmentHttpService) GetModuleName() string {
	return "department"
}

func departmentBranchCodeFilter(doc models.Department) map[string]interface{} {
	filter := map[string]interface{}{
		"code": doc.Code,
	}
	if doc.BranchCode != "" {
		filter["branchcode"] = doc.BranchCode
		return filter
	}
	if doc.BranchGuid != "" {
		filter["branchguid"] = doc.BranchGuid
		return filter
	}
	if doc.BranchKey != "" {
		filter["branch_key"] = doc.BranchKey
	}
	return filter
}
