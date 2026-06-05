package services

import (
	"context"
	"errors"
	"fmt"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/organization/branch/models"
	"smlcloudplatform/internal/organization/branch/repositories"
	businessTypeRepositories "smlcloudplatform/internal/organization/businesstype/repositories"
	deparmentRepositories "smlcloudplatform/internal/organization/department/repositories"
	"smlcloudplatform/internal/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/importdata"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type IBranchHttpService interface {
	CreateBranch(holdingCode string, authUsername string, doc models.Branch) (string, error)
	UpdateBranch(holdingCode string, guid string, authUsername string, doc models.Branch) error
	DeleteBranch(holdingCode string, guid string, authUsername string) error
	DeleteBranchByGUIDs(holdingCode string, authUsername string, GUIDs []string) error
	InfoBranch(holdingCode string, guid string) (models.BranchInfoResponse, error)
	InfoBranchByCode(holdingCode string, code string) (models.BranchInfoResponse, error)
	SearchBranch(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.BranchInfoResponse, mongopagination.PaginationData, error)
	SearchBranchStep(holdingCode string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.BranchInfoResponse, int, error)
	SaveInBatch(holdingCode string, authUsername string, dataList []models.Branch) (common.BulkImport, error)

	GetModuleName() string
}

type BranchHttpService struct {
	repo             repositories.IBranchRepository
	repoDepartment   deparmentRepositories.IDepartmentRepository
	repoBusinessType businessTypeRepositories.IBusinessTypeRepository

	syncCacheRepo mastersync.IMasterSyncCacheRepository
	services.ActivityService[models.BranchActivity, models.BranchDeleteActivity]
	contextTimeout time.Duration
}

func NewBranchHttpService(repo repositories.IBranchRepository, repoDepartment deparmentRepositories.IDepartmentRepository, repoBusinessType businessTypeRepositories.IBusinessTypeRepository, syncCacheRepo mastersync.IMasterSyncCacheRepository) *BranchHttpService {

	contextTimeout := time.Duration(15) * time.Second

	insSvc := &BranchHttpService{
		repo:             repo,
		repoDepartment:   repoDepartment,
		repoBusinessType: repoBusinessType,
		syncCacheRepo:    syncCacheRepo,
		contextTimeout:   contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.BranchActivity, models.BranchDeleteActivity](repo)

	return insSvc
}

func (svc BranchHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc BranchHttpService) CreateBranch(holdingCode string, authUsername string, doc models.Branch) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	normalizedCode, err := models.NormalizeThaiTaxBranchCode(doc.Code)
	if err != nil {
		return "", err
	}
	doc.Code = normalizedCode

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "code", doc.Code)
	if err != nil {
		return "", err
	}

	if len(findDoc.GuidFixed) > 0 {
		return "", errors.New("branch is exists")
	}

	tempMap := map[string]struct{}{}

	if doc.Departments == nil {
		doc.Departments = &[]models.Department{}
	}

	for _, department := range *doc.Departments {
		if _, ok := tempMap[department.Code]; ok {
			return "", fmt.Errorf("department code %s is duplicated", department.Code)
		}
		tempMap[department.Code] = struct{}{}
	}

	newGuidFixed := utils.NewGUID()

	docData := models.BranchDoc{}
	docData.HoldingCode = holdingCode
	docData.GuidFixed = newGuidFixed
	docData.Branch = doc

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	svc.saveMasterSync(holdingCode)

	return newGuidFixed, nil
}

func (svc BranchHttpService) UpdateBranch(holdingCode string, guid string, authUsername string, doc models.Branch) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	normalizedCode, err := models.NormalizeThaiTaxBranchCode(doc.Code)
	if err != nil {
		return err
	}
	doc.Code = normalizedCode

	existingCodeDoc, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "code", doc.Code)
	if err != nil {
		return err
	}
	if len(existingCodeDoc.GuidFixed) > 0 && existingCodeDoc.GuidFixed != guid {
		return errors.New("branch code is exists")
	}

	if doc.Departments == nil {
		doc.Departments = &[]models.Department{}
	}

	tempMap := map[string]struct{}{}
	for _, department := range *doc.Departments {
		if _, ok := tempMap[department.Code]; ok {
			return fmt.Errorf("department code %s is duplicated", department.Code)
		}
		tempMap[department.Code] = struct{}{}
	}

	docData := findDoc

	docData.Branch = doc

	docData.GuidFixed = findDoc.GuidFixed
	docData.UpdatedBy = authUsername
	docData.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, holdingCode, guid, docData)

	if err != nil {
		return err
	}

	svc.saveMasterSync(holdingCode)

	return nil
}

func (svc BranchHttpService) DeleteBranch(holdingCode string, guid string, authUsername string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	if models.IsThaiHeadOfficeBranchCode(findDoc.Code) {
		return errors.New("head office branch cannot be deleted")
	}

	branchCount, err := svc.repo.Count(ctx, holdingCode)
	if err != nil {
		return err
	}
	if branchCount <= 1 {
		return errors.New("company must have at least one branch")
	}

	err = svc.repo.DeleteByGuidfixed(ctx, holdingCode, guid, authUsername)
	if err != nil {
		return err
	}

	svc.saveMasterSync(holdingCode)

	return nil
}

func (svc BranchHttpService) DeleteBranchByGUIDs(holdingCode string, authUsername string, GUIDs []string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	branchCount, err := svc.repo.Count(ctx, holdingCode)
	if err != nil {
		return err
	}
	foundCount := 0
	for _, guid := range GUIDs {
		findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)
		if err != nil {
			return err
		}
		if len(findDoc.GuidFixed) < 1 {
			continue
		}
		foundCount++
		if models.IsThaiHeadOfficeBranchCode(findDoc.Code) {
			return errors.New("head office branch cannot be deleted")
		}
	}
	if branchCount <= foundCount {
		return errors.New("company must have at least one branch")
	}

	deleteFilterQuery := map[string]interface{}{
		"guidfixed": bson.M{"$in": GUIDs},
	}

	err = svc.repo.Delete(ctx, holdingCode, authUsername, deleteFilterQuery)
	if err != nil {
		return err
	}

	return nil
}

func (svc BranchHttpService) InfoBranch(holdingCode string, guid string) (models.BranchInfoResponse, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return models.BranchInfoResponse{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.BranchInfoResponse{}, errors.New("document not found")
	}

	resultDoc, err := svc.mapBranchInfo(findDoc.BranchInfo, holdingCode)

	if err != nil {
		return models.BranchInfoResponse{}, err
	}

	return resultDoc, nil
}

func (svc BranchHttpService) mapBranchInfo(findInfo models.BranchInfo, holdingCode string) (models.BranchInfoResponse, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	if findInfo.Branch.Departments == nil {
		findInfo.Branch.Departments = &[]models.Department{}
	}

	if findInfo.Branch.BusinessTypes == nil {
		findInfo.Branch.BusinessTypes = &[]string{}
	}

	businesstypes := []models.BusinessType{}
	for _, businesstypeGUID := range *findInfo.Branch.BusinessTypes {
		findBusinessTypeDoc, err := svc.repoBusinessType.FindByGuid(ctx, holdingCode, businesstypeGUID)

		if err != nil {
			return models.BranchInfoResponse{}, err
		}

		if len(findBusinessTypeDoc.GuidFixed) > 0 {
			businesstypes = append(businesstypes, models.BusinessType{
				GuidFixed: findBusinessTypeDoc.GuidFixed,
				Code:      findBusinessTypeDoc.BusinessType.Code,
				Names:     *findBusinessTypeDoc.BusinessType.Names,
			})
		}
	}

	// departments, err := svc.mapDepartmentToBranch(holdingCode, *findInfo.Branch.Departments)
	// if err != nil {
	// 	return models.BranchInfoResponse{}, err
	// }

	resultDoc := models.BranchInfoResponse{}
	resultDoc.BranchInfo = findInfo
	resultDoc.BusinessTypes = businesstypes
	// resultDoc.Departments = departments

	return resultDoc, nil
}

func (svc BranchHttpService) mapDepartmentToBranch(holdingCode string, departmentGUIDs []string) ([]models.Department, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	departments := []models.Department{}
	for _, departmentGUID := range departmentGUIDs {
		findDepartmentDoc, err := svc.repoDepartment.FindByGuid(ctx, holdingCode, departmentGUID)

		if err != nil {
			return nil, err
		}

		if len(findDepartmentDoc.GuidFixed) > 0 {
			departments = append(departments, models.Department{
				// GuidFixed: findDepartmentDoc.GuidFixed,
				Code:  findDepartmentDoc.Department.Code,
				Names: *findDepartmentDoc.Department.Names,
			})
		}
	}

	return departments, nil
}

func (svc BranchHttpService) InfoBranchByCode(holdingCode string, code string) (models.BranchInfoResponse, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	normalizedCode, err := models.NormalizeThaiTaxBranchCode(code)
	if err != nil {
		return models.BranchInfoResponse{}, err
	}

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "code", normalizedCode)

	if err != nil {
		return models.BranchInfoResponse{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.BranchInfoResponse{}, errors.New("document not found")
	}

	resultDoc, err := svc.mapBranchInfo(findDoc.BranchInfo, holdingCode)

	if err != nil {
		return models.BranchInfoResponse{}, err
	}

	return resultDoc, nil
}

func (svc BranchHttpService) SearchBranch(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.BranchInfoResponse, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
		"names.name",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, holdingCode, filters, searchInFields, pageable)

	if err != nil {
		return []models.BranchInfoResponse{}, pagination, err
	}

	resultDocs := []models.BranchInfoResponse{}
	for _, docInfo := range docList {

		resultDoc, err := svc.mapBranchInfo(docInfo, holdingCode)

		if err != nil {
			return []models.BranchInfoResponse{}, pagination, err
		}

		resultDocs = append(resultDocs, resultDoc)
	}

	return resultDocs, pagination, nil
}

func (svc BranchHttpService) SearchBranchStep(holdingCode string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.BranchInfoResponse, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
		"names.name",
	}

	selectFields := map[string]interface{}{}

	docList, total, err := svc.repo.FindStep(ctx, holdingCode, filters, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.BranchInfoResponse{}, 0, err
	}

	resultDocs := []models.BranchInfoResponse{}
	for _, docInfo := range docList {

		resultDoc, err := svc.mapBranchInfo(docInfo, holdingCode)

		if err != nil {
			return []models.BranchInfoResponse{}, 0, err
		}

		resultDocs = append(resultDocs, resultDoc)
	}

	return resultDocs, total, nil
}

func (svc BranchHttpService) SaveInBatch(holdingCode string, authUsername string, dataList []models.Branch) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	for i := range dataList {
		normalizedCode, err := models.NormalizeThaiTaxBranchCode(dataList[i].Code)
		if err != nil {
			return common.BulkImport{}, err
		}
		dataList[i].Code = normalizedCode
	}

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.Branch](dataList, svc.getDocIDKey)

	itemCodeGuidList := []interface{}{}
	for _, doc := range payloadList {
		itemCodeGuidList = append(itemCodeGuidList, doc.Code)
	}

	findItemGuid, err := svc.repo.FindInItemGuids(ctx, holdingCode, "code", itemCodeGuidList)

	if err != nil {
		return common.BulkImport{}, err
	}

	foundItemGuidList := []string{}
	for _, doc := range findItemGuid {
		foundItemGuidList = append(foundItemGuidList, doc.Code)
	}

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.Branch, models.BranchDoc](
		holdingCode,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(holdingCode string, authUsername string, doc models.Branch) models.BranchDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.BranchDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.HoldingCode = holdingCode
			dataDoc.Branch = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.Branch, models.BranchDoc](
		holdingCode,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(holdingCode string, guid string) (models.BranchDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "code", guid)
		},
		func(doc models.BranchDoc) bool {
			return doc.Code != ""
		},
		func(holdingCode string, authUsername string, data models.Branch, doc models.BranchDoc) error {

			doc.Branch = data
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

	createDataKey := []interface{}{}

	for _, doc := range createDataList {
		createDataKey = append(createDataKey, doc.Code)
	}

	payloadDuplicateDataKey := []interface{}{}
	for _, doc := range payloadDuplicateList {
		payloadDuplicateDataKey = append(payloadDuplicateDataKey, doc.Code)
	}

	updateDataKey := []interface{}{}
	for _, doc := range updateSuccessDataList {

		updateDataKey = append(updateDataKey, doc.Code)
	}

	updateFailDataKey := []interface{}{}
	for _, doc := range updateFailDataList {
		updateFailDataKey = append(updateFailDataKey, svc.getDocIDKey(doc))
	}

	svc.saveMasterSync(holdingCode)

	tempCreateDataKey := svc.toSliceString(createDataKey)
	tempUpdateDataKey := svc.toSliceString(updateDataKey)
	tempUpdateFailDataKey := svc.toSliceString(updateFailDataKey)
	tempPayloadDuplicateDataKey := svc.toSliceString(payloadDuplicateDataKey)

	return common.BulkImport{
		Created:          tempCreateDataKey,
		Updated:          tempUpdateDataKey,
		UpdateFailed:     tempUpdateFailDataKey,
		PayloadDuplicate: tempPayloadDuplicateDataKey,
	}, nil
}

func (svc BranchHttpService) toSliceString(data []interface{}) []string {
	tempData := make([]string, len(data))
	for i, v := range data {
		tempData[i] = v.(string)
	}

	return tempData
}

func (svc BranchHttpService) getDocIDKey(doc models.Branch) string {
	return doc.Code
}

func (svc BranchHttpService) saveMasterSync(holdingCode string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(holdingCode, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc BranchHttpService) GetModuleName() string {
	return "branch"
}
