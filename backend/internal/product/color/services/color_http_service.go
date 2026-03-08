package services

import (
	"context"
	"errors"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/product/color/models"
	"smlcloudplatform/internal/product/color/repositories"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/importdata"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IColorHttpService interface {
	CreateColor(shopID string, authUsername string, doc models.Color) (string, error)
	UpdateColor(shopID string, guid string, authUsername string, doc models.Color) error
	DeleteColor(shopID string, guid string, authUsername string) error
	InfoColor(shopID string, guid string) (models.ColorInfo, error)
	InfoWTFArray(shopID string, codes []string) ([]interface{}, error)
	SearchColor(shopID string, pageable micromodels.Pageable) ([]models.ColorInfo, mongopagination.PaginationData, error)
	SearchColorStep(shopID string, langCode string, pageableStep micromodels.PageableStep) ([]models.ColorInfo, int, error)
	SaveInBatch(shopID string, authUsername string, dataList []models.Color) (common.BulkImport, error)
}

type ColorHttpService struct {
	repo           repositories.IColorRepository
	contextTimeout time.Duration
}

func NewColorHttpService(repo repositories.IColorRepository) *ColorHttpService {

	contextTimeout := time.Duration(15) * time.Second

	return &ColorHttpService{
		repo:           repo,
		contextTimeout: contextTimeout,
	}
}

func (svc ColorHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc ColorHttpService) CreateColor(shopID string, authUsername string, doc models.Color) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", doc.Code)

	if err != nil {
		return "", err
	}

	if findDoc.Code != "" {
		return "", errors.New("Code is exists")
	}

	newGuidFixed := utils.NewGUID()

	docData := models.ColorDoc{}
	docData.ShopID = shopID
	docData.GuidFixed = newGuidFixed
	docData.Color = doc

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	return newGuidFixed, nil
}

func (svc ColorHttpService) UpdateColor(shopID string, guid string, authUsername string, doc models.Color) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	findDoc.Color = doc

	findDoc.UpdatedBy = authUsername
	findDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, shopID, guid, findDoc)

	if err != nil {
		return err
	}

	return nil
}

func (svc ColorHttpService) DeleteColor(shopID string, guid string, authUsername string) error {

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

func (svc ColorHttpService) InfoColor(shopID string, guid string) (models.ColorInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return models.ColorInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.ColorInfo{}, errors.New("document not found")
	}

	return findDoc.ColorInfo, nil

}

func (svc ColorHttpService) InfoWTFArray(shopID string, codes []string) ([]interface{}, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	docList := []interface{}{}

	for _, code := range codes {
		findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", code)
		if err != nil || findDoc.ID == primitive.NilObjectID {
			// add item empty
			docList = append(docList, nil)
		} else {
			docList = append(docList, findDoc.ColorInfo)
		}
	}

	return docList, nil
}

func (svc ColorHttpService) SearchColor(shopID string, pageable micromodels.Pageable) ([]models.ColorInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
		"names.name",
	}

	docList, pagination, err := svc.repo.FindPage(ctx, shopID, searchInFields, pageable)

	if err != nil {
		return []models.ColorInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc ColorHttpService) SearchColorStep(shopID string, langCode string, pageableStep micromodels.PageableStep) ([]models.ColorInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
		"names.name",
	}

	selectFields := map[string]interface{}{
		"guidfixed":      1,
		"code":           1,
		"colorselect":    1,
		"colorsystem":    1,
		"colorhex":       1,
		"colorselecthex": 1,
		"colorsystemhex": 1,
	}

	if langCode != "" {
		selectFields["names"] = bson.M{"$elemMatch": bson.M{"code": langCode}}
	} else {
		selectFields["names"] = 1
	}

	docList, total, err := svc.repo.FindStep(ctx, shopID, map[string]interface{}{}, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.ColorInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc ColorHttpService) SaveInBatch(shopID string, authUsername string, dataList []models.Color) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.Color](dataList, svc.getDocIDKey)

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

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.Color, models.ColorDoc](
		shopID,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(shopID string, authUsername string, doc models.Color) models.ColorDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.ColorDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.ShopID = shopID
			dataDoc.Color = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.Color, models.ColorDoc](
		shopID,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(shopID string, guid string) (models.ColorDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", guid)
		},
		func(doc models.ColorDoc) bool {
			return doc.Code != ""
		},
		func(shopID string, authUsername string, data models.Color, doc models.ColorDoc) error {

			doc.Color = data
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

	return common.BulkImport{
		Created:          createDataKey,
		Updated:          updateDataKey,
		UpdateFailed:     updateFailDataKey,
		PayloadDuplicate: payloadDuplicateDataKey,
	}, nil
}

func (svc ColorHttpService) getDocIDKey(doc models.Color) string {
	return doc.Code
}
