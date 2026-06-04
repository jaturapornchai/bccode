package services

import (
	"context"
	"errors"
	"fmt"
	"smlcloudplatform/internal/debtaccount/customer/models"
	"smlcloudplatform/internal/debtaccount/customer/repositories"
	groupModels "smlcloudplatform/internal/debtaccount/customergroup/models"
	groupRepositories "smlcloudplatform/internal/debtaccount/customergroup/repositories"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/importdata"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/samber/lo"
	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ICustomerHttpService interface {
	CreateCustomer(holdingCode string, authUsername string, doc models.CustomerRequest) (string, error)
	UpdateCustomer(holdingCode string, guid string, authUsername string, doc models.CustomerRequest) error
	DeleteCustomer(holdingCode string, guid string, authUsername string) error
	DeleteCustomerByGUIDs(holdingCode string, authUsername string, GUIDs []string) error
	InfoCustomer(holdingCode string, guid string) (models.CustomerInfo, error)
	InfoCustomerByCode(holdingCode string, code string) (models.CustomerInfo, error)
	SearchCustomer(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.CustomerInfo, mongopagination.PaginationData, error)
	SearchCustomerStep(holdingCode string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.CustomerInfo, int, error)
	SaveInBatch(holdingCode string, authUsername string, dataList []models.CustomerRequest) (common.BulkImport, error)

	GetModuleName() string
}

type CustomerHttpService struct {
	repo          repositories.ICustomerRepository
	repoGroup     groupRepositories.ICustomerGroupRepository
	syncCacheRepo mastersync.IMasterSyncCacheRepository
	services.ActivityService[models.CustomerActivity, models.CustomerDeleteActivity]
	contextTimeout time.Duration
}

func NewCustomerHttpService(repo repositories.ICustomerRepository, repoGroup groupRepositories.ICustomerGroupRepository, syncCacheRepo mastersync.IMasterSyncCacheRepository) *CustomerHttpService {

	contextTimeout := time.Duration(15) * time.Second

	insSvc := &CustomerHttpService{
		repo:           repo,
		repoGroup:      repoGroup,
		syncCacheRepo:  syncCacheRepo,
		contextTimeout: contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.CustomerActivity, models.CustomerDeleteActivity](repo)

	return insSvc
}

func (svc CustomerHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc CustomerHttpService) CreateCustomer(holdingCode string, authUsername string, doc models.CustomerRequest) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "code", doc.Code)

	if err != nil {
		return "", err
	}

	if findDoc.Code != "" {
		return "", errors.New("code is exists")
	}

	newGuidFixed := utils.NewGUID()

	docData := models.CustomerDoc{}
	docData.HoldingCode = holdingCode
	docData.GuidFixed = newGuidFixed
	docData.Customer = doc.Customer
	docData.GroupGUIDs = &doc.Groups

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	if docData.GroupGUIDs == nil {
		docData.GroupGUIDs = &[]string{}
	}

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	svc.saveMasterSync(holdingCode)

	return newGuidFixed, nil
}

func (svc CustomerHttpService) UpdateCustomer(holdingCode string, guid string, authUsername string, doc models.CustomerRequest) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	findDoc.Customer = doc.Customer
	findDoc.GroupGUIDs = &doc.Groups

	findDoc.UpdatedBy = authUsername
	findDoc.UpdatedAt = time.Now()

	if findDoc.GroupGUIDs == nil {
		findDoc.GroupGUIDs = &[]string{}
	}

	err = svc.repo.Update(ctx, holdingCode, guid, findDoc)

	if err != nil {
		return err
	}

	svc.saveMasterSync(holdingCode)

	return nil
}

func (svc CustomerHttpService) DeleteCustomer(holdingCode string, guid string, authUsername string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	err = svc.repo.DeleteByGuidfixed(ctx, holdingCode, guid, authUsername)
	if err != nil {
		return err
	}

	svc.saveMasterSync(holdingCode)

	return nil
}

func (svc CustomerHttpService) DeleteCustomerByGUIDs(holdingCode string, authUsername string, GUIDs []string) error {

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

func (svc CustomerHttpService) InfoCustomer(holdingCode string, guid string) (models.CustomerInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return models.CustomerInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.CustomerInfo{}, errors.New("document not found")
	}

	if findDoc.GroupGUIDs != nil && len(*findDoc.GroupGUIDs) > 0 {
		findGroups, err := svc.repoGroup.FindByGuids(ctx, holdingCode, *findDoc.GroupGUIDs)

		if err != nil {
			return models.CustomerInfo{}, err
		}

		groupInfo := lo.Map[groupModels.CustomerGroupDoc, groupModels.CustomerGroupInfo](
			findGroups,
			func(docGroup groupModels.CustomerGroupDoc, idx int) groupModels.CustomerGroupInfo {
				return docGroup.CustomerGroupInfo
			})

		findDoc.CustomerInfo.Groups = &groupInfo
	}

	return findDoc.CustomerInfo, nil
}

func (svc CustomerHttpService) InfoCustomerByCode(holdingCode string, code string) (models.CustomerInfo, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "code", code)

	if err != nil {
		return models.CustomerInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.CustomerInfo{}, errors.New("document not found")
	}

	if findDoc.GroupGUIDs != nil && len(*findDoc.GroupGUIDs) > 0 {
		findGroups, err := svc.repoGroup.FindByGuids(ctx, holdingCode, *findDoc.GroupGUIDs)

		if err != nil {
			return models.CustomerInfo{}, err
		}

		groupInfo := lo.Map[groupModels.CustomerGroupDoc, groupModels.CustomerGroupInfo](
			findGroups,
			func(docGroup groupModels.CustomerGroupDoc, idx int) groupModels.CustomerGroupInfo {
				return docGroup.CustomerGroupInfo
			})

		findDoc.CustomerInfo.Groups = &groupInfo
	}

	return findDoc.CustomerInfo, nil
}

func (svc CustomerHttpService) SearchCustomer(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.CustomerInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
		"names.name",
		"addressforbilling.phoneprimary",
		"addressforbilling.phonesecondary",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, holdingCode, filters, searchInFields, pageable)

	if err != nil {
		return []models.CustomerInfo{}, pagination, err
	}

	for idx, doc := range docList {
		if doc.GroupGUIDs != nil {
			findCustGroups, err := svc.repoGroup.FindByGuids(ctx, holdingCode, *doc.GroupGUIDs)
			if err != nil {
				return []models.CustomerInfo{}, pagination, err
			}

			custGroupInfo := lo.Map[groupModels.CustomerGroupDoc, groupModels.CustomerGroupInfo](
				findCustGroups,
				func(docGroup groupModels.CustomerGroupDoc, idx int) groupModels.CustomerGroupInfo {
					return docGroup.CustomerGroupInfo
				})

			docList[idx].Groups = &custGroupInfo
		}
	}

	return docList, pagination, nil
}

func (svc CustomerHttpService) SearchCustomerStep(holdingCode string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.CustomerInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
		"names.name",
		"addressforbilling.phoneprimary",
		"addressforbilling.phonesecondary",
	}

	selectFields := map[string]interface{}{}

	docList, total, err := svc.repo.FindStep(ctx, holdingCode, filters, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.CustomerInfo{}, 0, err
	}

	for idx, doc := range docList {
		if doc.GroupGUIDs != nil {
			findCustGroups, err := svc.repoGroup.FindByGuids(ctx, holdingCode, *doc.GroupGUIDs)
			if err != nil {
				return []models.CustomerInfo{}, 0, err
			}

			custGroupInfo := lo.Map[groupModels.CustomerGroupDoc, groupModels.CustomerGroupInfo](
				findCustGroups,
				func(docGroup groupModels.CustomerGroupDoc, idx int) groupModels.CustomerGroupInfo {
					return docGroup.CustomerGroupInfo
				})

			docList[idx].Groups = &custGroupInfo
		}
	}

	return docList, total, nil
}

func (svc CustomerHttpService) SaveInBatch(holdingCode string, authUsername string, dataListParam []models.CustomerRequest) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	dataList := []models.Customer{}
	for _, doc := range dataListParam {
		doc.GroupGUIDs = &doc.Groups
		dataList = append(dataList, doc.Customer)
	}

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.Customer](dataList, svc.getDocIDKey)

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

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.Customer, models.CustomerDoc](
		holdingCode,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(holdingCode string, authUsername string, doc models.Customer) models.CustomerDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.CustomerDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.HoldingCode = holdingCode
			dataDoc.Customer = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.Customer, models.CustomerDoc](
		holdingCode,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(holdingCode string, guid string) (models.CustomerDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "code", guid)
		},
		func(doc models.CustomerDoc) bool {
			return doc.Code != ""
		},
		func(holdingCode string, authUsername string, data models.Customer, doc models.CustomerDoc) error {

			doc.Customer = data
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

func (svc CustomerHttpService) getDocIDKey(doc models.Customer) string {
	return doc.Code
}

func (svc CustomerHttpService) saveMasterSync(holdingCode string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(holdingCode, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc CustomerHttpService) GetModuleName() string {
	return "customer"
}
