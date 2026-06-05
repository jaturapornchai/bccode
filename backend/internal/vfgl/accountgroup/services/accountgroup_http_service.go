package services

import (
	"context"
	"errors"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/importdata"
	"smlcloudplatform/internal/vfgl/accountgroup/models"
	"smlcloudplatform/internal/vfgl/accountgroup/repositories"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IAccountGroupHttpService interface {
	Create(holdingCode string, authUsername string, doc models.AccountGroup) (string, error)
	Update(guid string, holdingCode string, authUsername string, doc models.AccountGroup) error
	Delete(guid string, holdingCode string, authUsername string) error
	Info(guid string, holdingCode string) (models.AccountGroupInfo, error)
	Search(holdingCode string, pageable micromodels.Pageable) ([]models.AccountGroupInfo, mongopagination.PaginationData, error)
	SaveInBatch(holdingCode string, authUsername string, dataList []models.AccountGroup) (common.BulkImport, error)
}

type AccountGroupHttpService struct {
	repo           repositories.AccountGroupMongoRepository
	mqRepo         repositories.AccountGroupMqRepository
	contextTimeout time.Duration
}

func NewAccountGroupHttpService(repo repositories.AccountGroupMongoRepository, mqRepo repositories.AccountGroupMqRepository) AccountGroupHttpService {

	contextTimeout := time.Duration(15) * time.Second

	return AccountGroupHttpService{
		repo:           repo,
		mqRepo:         mqRepo,
		contextTimeout: contextTimeout,
	}
}

func (svc AccountGroupHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc AccountGroupHttpService) Create(holdingCode string, authUsername string, doc models.AccountGroup) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindOne(ctx, holdingCode, bson.M{"code": doc.Code})

	if err != nil {
		return "", err
	}

	if len(findDoc.Code) > 0 {
		return "", errors.New("code is exists")
	}

	newGuidFixed := utils.NewGUID()

	docData := models.AccountGroupDoc{}
	docData.HoldingCode = holdingCode
	docData.GuidFixed = newGuidFixed
	docData.AccountGroup = doc

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	svc.mqRepo.Create(docData)

	if err != nil {
		return "", err
	}
	return newGuidFixed, nil
}

func (svc AccountGroupHttpService) Update(guid string, holdingCode string, authUsername string, doc models.AccountGroup) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	findDocCode, err := svc.repo.FindOne(ctx, holdingCode, bson.M{"code": doc.Code})

	if err != nil {
		return err
	}

	if findDoc.Code != doc.Code && len(findDocCode.Code) > 0 {
		return errors.New("code is exists")
	}

	findDoc.AccountGroup = doc

	findDoc.UpdatedBy = authUsername
	findDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, holdingCode, guid, findDoc)

	if err != nil {
		return err
	}
	return nil
}

func (svc AccountGroupHttpService) Delete(guid string, holdingCode string, authUsername string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	err := svc.repo.DeleteByGuidfixed(ctx, holdingCode, guid, authUsername)

	if err != nil {
		return err
	}
	return nil
}

func (svc AccountGroupHttpService) Info(guid string, holdingCode string) (models.AccountGroupInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return models.AccountGroupInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.AccountGroupInfo{}, errors.New("document not found")
	}

	return findDoc.AccountGroupInfo, nil

}

func (svc AccountGroupHttpService) Search(holdingCode string, pageable micromodels.Pageable) ([]models.AccountGroupInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"guidfixed",
		"code",
	}

	docList, pagination, err := svc.repo.FindPage(ctx, holdingCode, searchInFields, pageable)

	if err != nil {
		return []models.AccountGroupInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc AccountGroupHttpService) SaveInBatch(holdingCode string, authUsername string, dataList []models.AccountGroup) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	createDataList := []models.AccountGroupDoc{}
	duplicateDataList := []models.AccountGroup{}

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.AccountGroup](dataList, svc.getDocIDKey)

	itemCodeGuidList := []string{}
	for _, doc := range payloadList {
		itemCodeGuidList = append(itemCodeGuidList, doc.Code)
	}

	findItemGuid, err := svc.repo.FindInItemGuid(ctx, holdingCode, "docno", itemCodeGuidList)

	if err != nil {
		return common.BulkImport{}, err
	}

	foundItemGuidList := []string{}
	for _, doc := range findItemGuid {
		foundItemGuidList = append(foundItemGuidList, doc.DocNo)
	}

	duplicateDataList, createDataList = importdata.PreparePayloadData[models.AccountGroup, models.AccountGroupDoc](
		holdingCode,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(holdingCode string, authUsername string, doc models.AccountGroup) models.AccountGroupDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.AccountGroupDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.HoldingCode = holdingCode
			dataDoc.AccountGroup = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.AccountGroup, models.AccountGroupDoc](
		holdingCode,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(holdingCode string, guid string) (models.AccountGroupDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "code", guid)
		},
		func(doc models.AccountGroupDoc) bool {
			if doc.Code != "" {
				return true
			}
			return false
		},
		func(holdingCode string, authUsername string, data models.AccountGroup, doc models.AccountGroupDoc) error {

			doc.AccountGroup = data
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

		svc.mqRepo.CreateInBatch(createDataList)

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
		svc.mqRepo.Update(doc)
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

func (svc AccountGroupHttpService) getDocIDKey(doc models.AccountGroup) string {
	return doc.Code
}
