package services

import (
	"context"
	"errors"
	smspatternsrepo "smlcloudplatform/internal/smsreceive/smspatterns/repositories"
	"smlcloudplatform/internal/smsreceive/smspaymentsettings/models"
	"smlcloudplatform/internal/smsreceive/smspaymentsettings/repositories"
	"smlcloudplatform/internal/utils"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ISmsPaymentSettingsHttpService interface {
	SaveSmsPaymentSettings(holdingCode string, authUsername string, storefrontGUID string, doc models.SmsPaymentSettings) error
	InfoSmsPaymentSettings(holdingCode string, storefrontGUID string) (models.SmsPaymentSettingsInfo, error)
	SearchSmsPaymentSettings(holdingCode string, pageable micromodels.Pageable) ([]models.SmsPaymentSettingsInfo, mongopagination.PaginationData, error)
}

type SmsPaymentSettingsHttpService struct {
	repo           repositories.SmsPaymentSettingsRepository
	repoPattern    smspatternsrepo.ISmsPatternsRepository
	contextTimeout time.Duration
}

func NewSmsPaymentSettingsHttpService(repo repositories.SmsPaymentSettingsRepository, repoPattern smspatternsrepo.ISmsPatternsRepository) SmsPaymentSettingsHttpService {

	contextTimeout := time.Duration(15) * time.Second

	return SmsPaymentSettingsHttpService{
		repo:           repo,
		repoPattern:    repoPattern,
		contextTimeout: contextTimeout,
	}
}

func (svc SmsPaymentSettingsHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc SmsPaymentSettingsHttpService) SaveSmsPaymentSettings(holdingCode string, authUsername string, storefrontGUID string, doc models.SmsPaymentSettings) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findPattern, err := svc.repoPattern.FindByCode(ctx, doc.PatternCode)

	if err != nil {
		return err
	}

	if len(findPattern.Code) < 1 {
		return errors.New("pattern code not found")
	}

	findDoc, err := svc.repo.FindOne(ctx, holdingCode, bson.M{})

	if err != nil {
		return err
	}

	isExitsSetting, err := svc.isExistsPaymentSettings(storefrontGUID, findDoc)

	if err != nil {
		return err
	}

	if isExitsSetting {
		return svc.updateSmsPaymentSettings(holdingCode, findDoc.GuidFixed, authUsername, doc)
	} else {
		return svc.createSmsPaymentSettings(holdingCode, authUsername, doc)
	}

}

func (svc SmsPaymentSettingsHttpService) isExistsPaymentSettings(storefrontGUID string, findDoc models.SmsPaymentSettingsDoc) (bool, error) {

	if len(findDoc.HoldingCode) > 0 && findDoc.StorefrontGUID == storefrontGUID {
		return true, nil
	}

	return false, nil
}

func (svc SmsPaymentSettingsHttpService) createSmsPaymentSettings(holdingCode string, authUsername string, doc models.SmsPaymentSettings) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	newGuidFixed := utils.NewGUID()

	docData := models.SmsPaymentSettingsDoc{}
	docData.SmsPaymentSettings = doc

	docData.HoldingCode = holdingCode
	docData.GuidFixed = newGuidFixed

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err := svc.repo.Create(ctx, docData)

	if err != nil {
		return err
	}

	return nil
}

func (svc SmsPaymentSettingsHttpService) updateSmsPaymentSettings(holdingCode string, guid string, authUsername string, doc models.SmsPaymentSettings) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	findDoc.SmsPaymentSettings = doc

	findDoc.UpdatedBy = authUsername
	findDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, holdingCode, guid, findDoc)

	if err != nil {
		return err
	}

	return nil
}

func (svc SmsPaymentSettingsHttpService) InfoSmsPaymentSettings(holdingCode string, storefrontGUID string) (models.SmsPaymentSettingsInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindOne(ctx, holdingCode, bson.M{"storefrontguid": storefrontGUID})

	if err != nil {
		return models.SmsPaymentSettingsInfo{}, err
	}

	return findDoc.SmsPaymentSettingsInfo, nil

}

func (svc SmsPaymentSettingsHttpService) SearchSmsPaymentSettings(holdingCode string, pageable micromodels.Pageable) ([]models.SmsPaymentSettingsInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	docList, pagination, err := svc.repo.FindPage(ctx, holdingCode, []string{}, pageable)

	if err != nil {
		return []models.SmsPaymentSettingsInfo{}, mongopagination.PaginationData{}, err
	}

	return docList, pagination, nil

}
