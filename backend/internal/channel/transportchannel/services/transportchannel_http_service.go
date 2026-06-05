package services

import (
	"context"
	"errors"
	"fmt"
	"smlcloudplatform/internal/channel/transportchannel/models"
	"smlcloudplatform/internal/channel/transportchannel/repositories"
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

type ITransportChannelHttpService interface {
	CreateTransportChannel(holdingCode string, authUsername string, doc models.TransportChannel) (string, error)
	UpdateTransportChannel(holdingCode string, guid string, authUsername string, doc models.TransportChannel) error
	DeleteTransportChannel(holdingCode string, guid string, authUsername string) error
	DeleteTransportChannelByGUIDs(holdingCode string, authUsername string, GUIDs []string) error
	InfoTransportChannel(holdingCode string, guid string) (models.TransportChannelInfo, error)
	InfoTransportChannelByCode(holdingCode string, code string) (models.TransportChannelInfo, error)
	SearchTransportChannel(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.TransportChannelInfo, mongopagination.PaginationData, error)
	SearchTransportChannelStep(holdingCode string, langCode string, pageableStep micromodels.PageableStep) ([]models.TransportChannelInfo, int, error)
	SaveInBatch(holdingCode string, authUsername string, dataList []models.TransportChannel) (common.BulkImport, error)

	GetModuleName() string
}

type TransportChannelHttpService struct {
	repo repositories.ITransportChannelRepository

	syncCacheRepo mastersync.IMasterSyncCacheRepository
	services.ActivityService[models.TransportChannelActivity, models.TransportChannelDeleteActivity]
	contextTimeout time.Duration
}

func NewTransportChannelHttpService(repo repositories.ITransportChannelRepository, syncCacheRepo mastersync.IMasterSyncCacheRepository) *TransportChannelHttpService {
	contextTimeout := time.Duration(15) * time.Second
	insSvc := &TransportChannelHttpService{
		repo:           repo,
		syncCacheRepo:  syncCacheRepo,
		contextTimeout: contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.TransportChannelActivity, models.TransportChannelDeleteActivity](repo)

	return insSvc
}

func (svc TransportChannelHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc TransportChannelHttpService) CreateTransportChannel(holdingCode string, authUsername string, doc models.TransportChannel) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "code", doc.Code)

	if err != nil {
		return "", err
	}

	if len(findDoc.GuidFixed) > 0 {
		return "", errors.New("Code is exists")
	}

	newGuidFixed := utils.NewGUID()

	docData := models.TransportChannelDoc{}
	docData.HoldingCode = holdingCode
	docData.GuidFixed = newGuidFixed
	docData.TransportChannel = doc

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	svc.saveMasterSync(holdingCode)

	return newGuidFixed, nil
}

func (svc TransportChannelHttpService) UpdateTransportChannel(holdingCode string, guid string, authUsername string, doc models.TransportChannel) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	findDoc.TransportChannel = doc

	findDoc.UpdatedBy = authUsername
	findDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, holdingCode, guid, findDoc)

	if err != nil {
		return err
	}

	svc.saveMasterSync(holdingCode)

	return nil
}

func (svc TransportChannelHttpService) DeleteTransportChannel(holdingCode string, guid string, authUsername string) error {

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

func (svc TransportChannelHttpService) DeleteTransportChannelByGUIDs(holdingCode string, authUsername string, GUIDs []string) error {

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

func (svc TransportChannelHttpService) InfoTransportChannel(holdingCode string, guid string) (models.TransportChannelInfo, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return models.TransportChannelInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.TransportChannelInfo{}, errors.New("document not found")
	}

	return findDoc.TransportChannelInfo, nil
}

func (svc TransportChannelHttpService) InfoTransportChannelByCode(holdingCode string, code string) (models.TransportChannelInfo, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "code", code)

	if err != nil {
		return models.TransportChannelInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.TransportChannelInfo{}, errors.New("document not found")
	}

	return findDoc.TransportChannelInfo, nil
}

func (svc TransportChannelHttpService) SearchTransportChannel(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.TransportChannelInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
		"names.name",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, holdingCode, filters, searchInFields, pageable)

	if err != nil {
		return []models.TransportChannelInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc TransportChannelHttpService) SearchTransportChannelStep(holdingCode string, langCode string, pageableStep micromodels.PageableStep) ([]models.TransportChannelInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
		"names.name",
	}

	selectFields := map[string]interface{}{}

	docList, total, err := svc.repo.FindStep(ctx, holdingCode, map[string]interface{}{}, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.TransportChannelInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc TransportChannelHttpService) SaveInBatch(holdingCode string, authUsername string, dataList []models.TransportChannel) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.TransportChannel](dataList, svc.getDocIDKey)

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

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.TransportChannel, models.TransportChannelDoc](
		holdingCode,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(holdingCode string, authUsername string, doc models.TransportChannel) models.TransportChannelDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.TransportChannelDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.HoldingCode = holdingCode
			dataDoc.TransportChannel = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.TransportChannel, models.TransportChannelDoc](
		holdingCode,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(holdingCode string, guid string) (models.TransportChannelDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "code", guid)
		},
		func(doc models.TransportChannelDoc) bool {
			return doc.Code != ""
		},
		func(holdingCode string, authUsername string, data models.TransportChannel, doc models.TransportChannelDoc) error {

			doc.TransportChannel = data
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

func (svc TransportChannelHttpService) getDocIDKey(doc models.TransportChannel) string {
	return doc.Code
}

func (svc TransportChannelHttpService) saveMasterSync(holdingCode string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(holdingCode, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc TransportChannelHttpService) GetModuleName() string {
	return "transportChannel"
}
