package services

import (
	"context"
	"errors"
	"fmt"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/pos/media/models"
	"smlcloudplatform/internal/pos/media/repositories"
	"smlcloudplatform/internal/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/importdata"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type IMediaHttpService interface {
	CreateMedia(shopID string, authUsername string, doc models.Media) (string, error)
	UpdateMedia(shopID string, guid string, authUsername string, doc models.Media) error
	DeleteMedia(shopID string, guid string, authUsername string) error
	DeleteMediaByGUIDs(shopID string, authUsername string, GUIDs []string) error
	InfoMedia(shopID string, guid string) (models.MediaInfo, error)
	InfoMediaByCode(shopID string, code string) (models.MediaInfo, error)
	SearchMedia(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.MediaInfo, mongopagination.PaginationData, error)
	SearchMediaStep(shopID string, langCode string, pageableStep micromodels.PageableStep) ([]models.MediaInfo, int, error)
	SaveInBatch(shopID string, authUsername string, dataList []models.Media) (common.BulkImport, error)

	GetModuleName() string
}

type MediaHttpService struct {
	repo repositories.IMediaRepository

	syncCacheRepo mastersync.IMasterSyncCacheRepository
	services.ActivityService[models.MediaActivity, models.MediaDeleteActivity]
	contextTimeout time.Duration
}

func NewMediaHttpService(
	repo repositories.IMediaRepository,
	syncCacheRepo mastersync.IMasterSyncCacheRepository,

	contextTimeout time.Duration,
) *MediaHttpService {

	insSvc := &MediaHttpService{
		repo:          repo,
		syncCacheRepo: syncCacheRepo,

		contextTimeout: contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.MediaActivity, models.MediaDeleteActivity](repo)

	return insSvc
}

func (svc MediaHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc MediaHttpService) CreateMedia(shopID string, authUsername string, doc models.Media) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", doc.Code)

	if err != nil {
		return "", err
	}

	if len(findDoc.GuidFixed) > 0 {
		return "", errors.New("Code is exists")
	}

	newGuidFixed := utils.NewGUID()

	docData := models.MediaDoc{}
	docData.ShopID = shopID
	docData.GuidFixed = newGuidFixed
	docData.Media = doc

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	go func() {
		svc.saveMasterSync(shopID)
	}()

	return newGuidFixed, nil
}

func (svc MediaHttpService) UpdateMedia(shopID string, guid string, authUsername string, doc models.Media) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	findDoc.Media = doc

	findDoc.UpdatedBy = authUsername
	findDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, shopID, guid, findDoc)

	if err != nil {
		return err
	}

	go func() {
		svc.saveMasterSync(shopID)
	}()

	return nil
}

func (svc MediaHttpService) DeleteMedia(shopID string, guid string, authUsername string) error {

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

	go func() {
		svc.saveMasterSync(shopID)
	}()

	return nil
}

func (svc MediaHttpService) DeleteMediaByGUIDs(shopID string, authUsername string, GUIDs []string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	deleteFilterQuery := map[string]interface{}{
		"guidfixed": bson.M{"$in": GUIDs},
	}

	err := svc.repo.Delete(ctx, shopID, authUsername, deleteFilterQuery)
	if err != nil {
		return err
	}

	go func() {
		svc.saveMasterSync(shopID)
	}()

	return nil
}

func (svc MediaHttpService) InfoMedia(shopID string, guid string) (models.MediaInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return models.MediaInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.MediaInfo{}, errors.New("document not found")
	}

	return findDoc.MediaInfo, nil
}

func (svc MediaHttpService) InfoMediaByCode(shopID string, code string) (models.MediaInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", code)

	if err != nil {
		return models.MediaInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.MediaInfo{}, errors.New("document not found")
	}

	return findDoc.MediaInfo, nil
}

func (svc MediaHttpService) SearchMedia(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.MediaInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, shopID, filters, searchInFields, pageable)

	if err != nil {
		return []models.MediaInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc MediaHttpService) SearchMediaStep(shopID string, langCode string, pageableStep micromodels.PageableStep) ([]models.MediaInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
	}

	selectFields := map[string]interface{}{}

	/*
		if langCode != "" {
			selectFields["names"] = bson.M{"$elemMatch": bson.M{"code": langCode}}
		} else {
			selectFields["names"] = 1
		}
	*/

	docList, total, err := svc.repo.FindStep(ctx, shopID, map[string]interface{}{}, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.MediaInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc MediaHttpService) SaveInBatch(shopID string, authUsername string, dataList []models.Media) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.Media](dataList, svc.getDocIDKey)

	itemCodeGuidList := []string{}
	for _, doc := range payloadList {
		itemCodeGuidList = append(itemCodeGuidList, doc.Code)
	}

	findItemGuid, err := svc.repo.FindInItemGuid(ctx, shopID, "code", itemCodeGuidList)

	if err != nil {
		return common.BulkImport{}, err
	}

	foundItemGuidList := []string{}
	for _, doc := range findItemGuid {
		foundItemGuidList = append(foundItemGuidList, doc.Code)
	}

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.Media, models.MediaDoc](
		shopID,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(shopID string, authUsername string, doc models.Media) models.MediaDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.MediaDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.ShopID = shopID
			dataDoc.Media = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.Media, models.MediaDoc](
		shopID,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(shopID string, guid string) (models.MediaDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", guid)
		},
		func(doc models.MediaDoc) bool {
			return doc.Code != ""
		},
		func(shopID string, authUsername string, data models.Media, doc models.MediaDoc) error {

			doc.Media = data
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

	svc.saveMasterSync(shopID)

	return common.BulkImport{
		Created:          createDataKey,
		Updated:          updateDataKey,
		UpdateFailed:     updateFailDataKey,
		PayloadDuplicate: payloadDuplicateDataKey,
	}, nil
}

func (svc MediaHttpService) getDocIDKey(doc models.Media) string {
	return doc.Code
}

func (svc MediaHttpService) saveMasterSync(shopID string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(shopID, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc MediaHttpService) GetModuleName() string {
	return "media"
}
