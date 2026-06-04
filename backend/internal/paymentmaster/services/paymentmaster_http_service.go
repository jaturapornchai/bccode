package services

import (
	"context"
	"errors"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/paymentmaster/models"
	"smlcloudplatform/internal/paymentmaster/repositories"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/importdata"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IPaymentMasterHttpService interface {
	CreatePaymentMaster(holdingCode string, authUsername string, doc models.PaymentMaster) (string, error)
	UpdatePaymentMaster(guid string, holdingCode string, authUsername string, doc models.PaymentMaster) error
	DeletePaymentMaster(guid string, holdingCode string, authUsername string) error
	InfoPaymentMaster(guid string, holdingCode string) (models.PaymentMasterInfo, error)
	SearchPaymentMaster(holdingCode string, q string) ([]models.PaymentMasterInfo, error)
	SaveInBatch(holdingCode string, authUsername string, dataList []models.PaymentMaster) (common.BulkImport, error)
}

type PaymentMasterHttpService struct {
	repo           repositories.PaymentMasterRepository
	contextTimeout time.Duration
}

func NewPaymentMasterHttpService(repo repositories.PaymentMasterRepository) PaymentMasterHttpService {

	contextTimeout := time.Duration(15) * time.Second

	return PaymentMasterHttpService{
		repo:           repo,
		contextTimeout: contextTimeout,
	}
}

func (svc PaymentMasterHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc PaymentMasterHttpService) CreatePaymentMaster(holdingCode string, authUsername string, doc models.PaymentMaster) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "paymentcode", doc.PaymentCode)

	if err != nil {
		return "", err
	}

	if findDoc.PaymentCode != "" {
		return "", errors.New("PaymentCode is exists")
	}

	newGuidFixed := utils.NewGUID()

	docData := models.PaymentMasterDoc{}
	docData.HoldingCode = holdingCode
	docData.GuidFixed = newGuidFixed
	docData.PaymentMaster = doc

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	return newGuidFixed, nil
}

func (svc PaymentMasterHttpService) UpdatePaymentMaster(guid string, holdingCode string, authUsername string, doc models.PaymentMaster) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	findDoc.PaymentMaster = doc

	findDoc.UpdatedBy = authUsername
	findDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, holdingCode, guid, findDoc)

	if err != nil {
		return err
	}

	return nil
}

func (svc PaymentMasterHttpService) DeletePaymentMaster(guid string, holdingCode string, authUsername string) error {

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

	return nil
}

func (svc PaymentMasterHttpService) InfoPaymentMaster(guid string, holdingCode string) (models.PaymentMasterInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return models.PaymentMasterInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.PaymentMasterInfo{}, errors.New("document not found")
	}

	return findDoc.PaymentMasterInfo, nil

}

func (svc PaymentMasterHttpService) SearchPaymentMaster(holdingCode string, q string) ([]models.PaymentMasterInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"guid_fixed",
		"paymentcode",
	}

	docList, err := svc.repo.Find(ctx, holdingCode, searchInFields, q)

	if err != nil {
		return []models.PaymentMasterInfo{}, err
	}

	return docList, nil
}

func (svc PaymentMasterHttpService) SaveInBatch(holdingCode string, authUsername string, dataList []models.PaymentMaster) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	createDataList := []models.PaymentMasterDoc{}
	duplicateDataList := []models.PaymentMaster{}

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.PaymentMaster](dataList, svc.getDocIDKey)

	itemCodeGuidList := []string{}
	for _, doc := range payloadList {
		itemCodeGuidList = append(itemCodeGuidList, doc.PaymentCode)
	}

	findItemGuid, err := svc.repo.FindInItemGuid(ctx, holdingCode, "paymentcode", itemCodeGuidList)

	if err != nil {
		return common.BulkImport{}, err
	}

	foundItemGuidList := []string{}
	for _, doc := range findItemGuid {
		foundItemGuidList = append(foundItemGuidList, doc.PaymentCode)
	}

	duplicateDataList, createDataList = importdata.PreparePayloadData[models.PaymentMaster, models.PaymentMasterDoc](
		holdingCode,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(holdingCode string, authUsername string, doc models.PaymentMaster) models.PaymentMasterDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.PaymentMasterDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.HoldingCode = holdingCode
			dataDoc.PaymentMaster = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.PaymentMaster, models.PaymentMasterDoc](
		holdingCode,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(holdingCode string, guid string) (models.PaymentMasterDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "paymentcode", guid)
		},
		func(doc models.PaymentMasterDoc) bool {
			if doc.PaymentCode != "" {
				return true
			}
			return false
		},
		func(holdingCode string, authUsername string, data models.PaymentMaster, doc models.PaymentMasterDoc) error {

			doc.PaymentMaster = data
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
		createDataKey = append(createDataKey, doc.PaymentCode)
	}

	payloadDuplicateDataKey := []string{}
	for _, doc := range payloadDuplicateList {
		payloadDuplicateDataKey = append(payloadDuplicateDataKey, doc.PaymentCode)
	}

	updateDataKey := []string{}
	for _, doc := range updateSuccessDataList {
		updateDataKey = append(updateDataKey, doc.PaymentCode)
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

func (svc PaymentMasterHttpService) getDocIDKey(doc models.PaymentMaster) string {
	return doc.PaymentCode
}
