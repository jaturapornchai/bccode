package services

import (
	"context"
	"errors"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/product/optionpattern/models"
	"smlcloudplatform/internal/product/optionpattern/repositories"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/importdata"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IOptionPatternHttpService interface {
	CreateOptionPattern(shopID string, authUsername string, doc models.OptionPattern) (string, error)
	UpdateOptionPattern(shopID string, guid string, authUsername string, doc models.OptionPattern) error
	DeleteOptionPattern(shopID string, guid string, authUsername string) error
	InfoOptionPattern(shopID string, guid string) (models.OptionPatternInfo, error)
	SearchOptionPattern(shopID string, pageable micromodels.Pageable) ([]models.OptionPatternInfo, mongopagination.PaginationData, error)
	SaveInBatch(shopID string, authUsername string, dataList []models.OptionPattern) (common.BulkImport, error)
}

type OptionPatternHttpService struct {
	repo           repositories.IOptionPatternRepository
	contextTimeout time.Duration
}

func NewOptionPatternHttpService(repo repositories.IOptionPatternRepository) *OptionPatternHttpService {

	contextTimeout := time.Duration(15) * time.Second

	return &OptionPatternHttpService{
		repo:           repo,
		contextTimeout: contextTimeout,
	}
}

func (svc OptionPatternHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc OptionPatternHttpService) CreateOptionPattern(shopID string, authUsername string, doc models.OptionPattern) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "patterncode", doc.PatternCode)

	if err != nil {
		return "", err
	}

	if findDoc.PatternCode != "" {
		return "", errors.New("PatternCode is exists")
	}

	newGuidFixed := utils.NewGUID()

	docData := models.OptionPatternDoc{}
	docData.ShopID = shopID
	docData.GuidFixed = newGuidFixed
	docData.OptionPattern = doc

	for _, detail := range *docData.OptionPatternDetails {
		detail.Option = nil
	}

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	return newGuidFixed, nil
}

func (svc OptionPatternHttpService) UpdateOptionPattern(shopID string, guid string, authUsername string, doc models.OptionPattern) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	findDoc.OptionPattern = doc

	findDoc.UpdatedBy = authUsername
	findDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, shopID, guid, findDoc)

	if err != nil {
		return err
	}

	return nil
}

func (svc OptionPatternHttpService) DeleteOptionPattern(shopID string, guid string, authUsername string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	err = svc.repo.DeleteByGuidfixed(ctx, shopID, guid, authUsername)
	if err != nil {
		return err
	}

	return nil
}

func (svc OptionPatternHttpService) InfoOptionPattern(shopID string, guid string) (models.OptionPatternInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return models.OptionPatternInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.OptionPatternInfo{}, errors.New("document not found")
	}

	return findDoc.OptionPatternInfo, nil

}

func (svc OptionPatternHttpService) SearchOptionPattern(shopID string, pageable micromodels.Pageable) ([]models.OptionPatternInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"guidfixed",
		"patterncode",
	}

	docList, pagination, err := svc.repo.FindPage(ctx, shopID, searchInFields, pageable)

	if err != nil {
		return []models.OptionPatternInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc OptionPatternHttpService) SaveInBatch(shopID string, authUsername string, dataList []models.OptionPattern) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.OptionPattern](dataList, svc.getDocIDKey)

	itemCodeGuidList := []string{}
	for _, doc := range payloadList {
		itemCodeGuidList = append(itemCodeGuidList, doc.PatternCode)
	}

	findItemGuid, err := svc.repo.FindInItemGuid(ctx, shopID, "patterncode", itemCodeGuidList)

	if err != nil {
		return common.BulkImport{}, err
	}

	foundItemGuidList := []string{}
	for _, doc := range findItemGuid {
		foundItemGuidList = append(foundItemGuidList, doc.PatternCode)
	}

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.OptionPattern, models.OptionPatternDoc](
		shopID,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(shopID string, authUsername string, doc models.OptionPattern) models.OptionPatternDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.OptionPatternDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.ShopID = shopID
			dataDoc.OptionPattern = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.OptionPattern, models.OptionPatternDoc](
		shopID,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(shopID string, guid string) (models.OptionPatternDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, shopID, "patterncode", guid)
		},
		func(doc models.OptionPatternDoc) bool {
			return doc.PatternCode != ""
		},
		func(shopID string, authUsername string, data models.OptionPattern, doc models.OptionPatternDoc) error {

			doc.OptionPattern = data
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
		createDataKey = append(createDataKey, doc.PatternCode)
	}

	payloadDuplicateDataKey := []string{}
	for _, doc := range payloadDuplicateList {
		payloadDuplicateDataKey = append(payloadDuplicateDataKey, doc.PatternCode)
	}

	updateDataKey := []string{}
	for _, doc := range updateSuccessDataList {

		updateDataKey = append(updateDataKey, doc.PatternCode)
	}

	updateFailDataKey := []string{}
	for _, doc := range updateFailDataList {
		updateFailDataKey = append(updateFailDataKey, svc.getDocIDKey(doc))
	}

	return common.BulkImport{
		Created:          createDataKey,
		Updated:          updateDataKey,
		UpdateFailed:     updateFailDataKey,
		PayloadDuplicate: payloadDuplicateDataKey,
	}, nil
}

func (svc OptionPatternHttpService) getDocIDKey(doc models.OptionPattern) string {
	return doc.PatternCode
}
