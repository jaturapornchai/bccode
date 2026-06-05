package services

import (
	"context"
	"errors"
	"fmt"
	"smlcloudplatform/internal/channel/salechannel/models"
	"smlcloudplatform/internal/channel/salechannel/repositories"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/importdata"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type ISaleChannelHttpService interface {
	CreateSaleChannel(holdingCode string, authUsername string, doc models.SaleChannel) (string, error)
	UpdateSaleChannel(holdingCode string, guid string, authUsername string, doc models.SaleChannel) error
	DeleteSaleChannel(holdingCode string, guid string, authUsername string) error
	DeleteSaleChannelByGUIDs(holdingCode string, authUsername string, GUIDs []string) error
	InfoSaleChannel(holdingCode string, guid string) (models.SaleChannelInfo, error)
	InfoSaleChannelByCode(holdingCode string, code string) (models.SaleChannelInfo, error)
	SearchSaleChannel(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.SaleChannelInfo, mongopagination.PaginationData, error)
	SearchSaleChannelStep(holdingCode string, langCode string, pageableStep micromodels.PageableStep) ([]models.SaleChannelInfo, int, error)
	SaveInBatch(holdingCode string, authUsername string, dataList []models.SaleChannel) (common.BulkImport, error)

	GetModuleName() string
}

type SaleChannelHttpService struct {
	repo repositories.ISaleChannelRepository

	syncCacheRepo mastersync.IMasterSyncCacheRepository
	services.ActivityService[models.SaleChannelActivity, models.SaleChannelDeleteActivity]
}

func NewSaleChannelHttpService(repo repositories.ISaleChannelRepository, syncCacheRepo mastersync.IMasterSyncCacheRepository) *SaleChannelHttpService {

	insSvc := &SaleChannelHttpService{
		repo:          repo,
		syncCacheRepo: syncCacheRepo,
	}

	insSvc.ActivityService = services.NewActivityService[models.SaleChannelActivity, models.SaleChannelDeleteActivity](repo)

	return insSvc
}

func (svc SaleChannelHttpService) CreateSaleChannel(holdingCode string, authUsername string, doc models.SaleChannel) (string, error) {

	findDoc, err := svc.repo.FindByDocIndentityGuid(context.Background(), holdingCode, "code", doc.Code)

	if err != nil {
		return "", err
	}

	if len(findDoc.GuidFixed) > 0 {
		return "", errors.New("Code is exists")
	}

	newGuidFixed := utils.NewGUID()

	docData := models.SaleChannelDoc{}
	docData.HoldingCode = holdingCode
	docData.GuidFixed = newGuidFixed
	docData.SaleChannel = doc

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err = svc.repo.Create(context.Background(), docData)

	if err != nil {
		return "", err
	}

	svc.saveMasterSync(holdingCode)

	return newGuidFixed, nil
}

func (svc SaleChannelHttpService) UpdateSaleChannel(holdingCode string, guid string, authUsername string, doc models.SaleChannel) error {

	findDoc, err := svc.repo.FindByGuid(context.Background(), holdingCode, guid)

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	findDoc.SaleChannel = doc

	findDoc.UpdatedBy = authUsername
	findDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(context.Background(), holdingCode, guid, findDoc)

	if err != nil {
		return err
	}

	svc.saveMasterSync(holdingCode)

	return nil
}

func (svc SaleChannelHttpService) DeleteSaleChannel(holdingCode string, guid string, authUsername string) error {

	findDoc, err := svc.repo.FindByGuid(context.Background(), holdingCode, guid)

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	err = svc.repo.DeleteByGuidfixed(context.Background(), holdingCode, guid, authUsername)
	if err != nil {
		return err
	}

	svc.saveMasterSync(holdingCode)

	return nil
}

func (svc SaleChannelHttpService) DeleteSaleChannelByGUIDs(holdingCode string, authUsername string, GUIDs []string) error {

	deleteFilterQuery := map[string]interface{}{
		"guidfixed": bson.M{"$in": GUIDs},
	}

	err := svc.repo.Delete(context.Background(), holdingCode, authUsername, deleteFilterQuery)
	if err != nil {
		return err
	}

	return nil
}

func (svc SaleChannelHttpService) InfoSaleChannel(holdingCode string, guid string) (models.SaleChannelInfo, error) {

	findDoc, err := svc.repo.FindByGuid(context.Background(), holdingCode, guid)

	if err != nil {
		return models.SaleChannelInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.SaleChannelInfo{}, errors.New("document not found")
	}

	return findDoc.SaleChannelInfo, nil
}

func (svc SaleChannelHttpService) InfoSaleChannelByCode(holdingCode string, code string) (models.SaleChannelInfo, error) {

	findDoc, err := svc.repo.FindByDocIndentityGuid(context.Background(), holdingCode, "code", code)

	if err != nil {
		return models.SaleChannelInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.SaleChannelInfo{}, errors.New("document not found")
	}

	return findDoc.SaleChannelInfo, nil
}

func (svc SaleChannelHttpService) SearchSaleChannel(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.SaleChannelInfo, mongopagination.PaginationData, error) {
	searchInFields := []string{
		"code",
		"names.name",
	}

	docList, pagination, err := svc.repo.FindPageFilter(context.Background(), holdingCode, filters, searchInFields, pageable)

	if err != nil {
		return []models.SaleChannelInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc SaleChannelHttpService) SearchSaleChannelStep(holdingCode string, langCode string, pageableStep micromodels.PageableStep) ([]models.SaleChannelInfo, int, error) {
	searchInFields := []string{
		"code",
		"names.name",
	}

	selectFields := map[string]interface{}{}

	docList, total, err := svc.repo.FindStep(context.Background(), holdingCode, map[string]interface{}{}, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.SaleChannelInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc SaleChannelHttpService) SaveInBatch(holdingCode string, authUsername string, dataList []models.SaleChannel) (common.BulkImport, error) {

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.SaleChannel](dataList, svc.getDocIDKey)

	itemCodeGuidList := []string{}
	for _, doc := range payloadList {
		itemCodeGuidList = append(itemCodeGuidList, doc.Code)
	}

	findItemGuid, err := svc.repo.FindInItemGuid(context.Background(), holdingCode, "code", itemCodeGuidList)

	if err != nil {
		return common.BulkImport{}, err
	}

	foundItemGuidList := []string{}
	for _, doc := range findItemGuid {
		foundItemGuidList = append(foundItemGuidList, doc.Code)
	}

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.SaleChannel, models.SaleChannelDoc](
		holdingCode,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(holdingCode string, authUsername string, doc models.SaleChannel) models.SaleChannelDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.SaleChannelDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.HoldingCode = holdingCode
			dataDoc.SaleChannel = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.SaleChannel, models.SaleChannelDoc](
		holdingCode,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(holdingCode string, guid string) (models.SaleChannelDoc, error) {
			return svc.repo.FindByDocIndentityGuid(context.Background(), holdingCode, "code", guid)
		},
		func(doc models.SaleChannelDoc) bool {
			return doc.Code != ""
		},
		func(holdingCode string, authUsername string, data models.SaleChannel, doc models.SaleChannelDoc) error {

			doc.SaleChannel = data
			doc.UpdatedBy = authUsername
			doc.UpdatedAt = time.Now()

			err = svc.repo.Update(context.Background(), holdingCode, doc.GuidFixed, doc)
			if err != nil {
				return nil
			}
			return nil
		},
	)

	if len(createDataList) > 0 {
		err = svc.repo.CreateInBatch(context.Background(), createDataList)

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

func (svc SaleChannelHttpService) getDocIDKey(doc models.SaleChannel) string {
	return doc.Code
}

func (svc SaleChannelHttpService) saveMasterSync(holdingCode string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(holdingCode, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc SaleChannelHttpService) GetModuleName() string {
	return "saleChannel"
}
