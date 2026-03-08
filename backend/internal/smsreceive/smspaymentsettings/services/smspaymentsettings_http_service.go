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
	SaveSmsPaymentSettings(shopID string, authUsername string, storefrontGUID string, doc models.SmsPaymentSettings) error
	InfoSmsPaymentSettings(shopID string, storefrontGUID string) (models.SmsPaymentSettingsInfo, error)
	SearchSmsPaymentSettings(shopID string, pageable micromodels.Pageable) ([]models.SmsPaymentSettingsInfo, mongopagination.PaginationData, error)
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

func (svc SmsPaymentSettingsHttpService) SaveSmsPaymentSettings(shopID string, authUsername string, storefrontGUID string, doc models.SmsPaymentSettings) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findPattern, err := svc.repoPattern.FindByCode(ctx, doc.PatternCode)

	if err != nil {
		return err
	}

	if len(findPattern.Code) < 1 {
		return errors.New("pattern code not found")
	}

	findDoc, err := svc.repo.FindOne(ctx, shopID, bson.M{})

	if err != nil {
		return err
	}

	isExitsSetting, err := svc.isExistsPaymentSettings(storefrontGUID, findDoc)

	if err != nil {
		return err
	}

	if isExitsSetting {
		return svc.updateSmsPaymentSettings(shopID, findDoc.GuidFixed, authUsername, doc)
	} else {
		return svc.createSmsPaymentSettings(shopID, authUsername, doc)
	}

}

func (svc SmsPaymentSettingsHttpService) isExistsPaymentSettings(storefrontGUID string, findDoc models.SmsPaymentSettingsDoc) (bool, error) {

	if len(findDoc.ShopID) > 0 && findDoc.StorefrontGUID == storefrontGUID {
		return true, nil
	}

	return false, nil
}

func (svc SmsPaymentSettingsHttpService) createSmsPaymentSettings(shopID string, authUsername string, doc models.SmsPaymentSettings) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	newGuidFixed := utils.NewGUID()

	docData := models.SmsPaymentSettingsDoc{}
	docData.SmsPaymentSettings = doc

	docData.ShopID = shopID
	docData.GuidFixed = newGuidFixed

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err := svc.repo.Create(ctx, docData)

	if err != nil {
		return err
	}

	return nil
}

func (svc SmsPaymentSettingsHttpService) updateSmsPaymentSettings(shopID string, guid string, authUsername string, doc models.SmsPaymentSettings) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	findDoc.SmsPaymentSettings = doc

	findDoc.UpdatedBy = authUsername
	findDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, shopID, guid, findDoc)

	if err != nil {
		return err
	}

	return nil
}

func (svc SmsPaymentSettingsHttpService) InfoSmsPaymentSettings(shopID string, storefrontGUID string) (models.SmsPaymentSettingsInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindOne(ctx, shopID, bson.M{"storefrontguid": storefrontGUID})

	if err != nil {
		return models.SmsPaymentSettingsInfo{}, err
	}

	return findDoc.SmsPaymentSettingsInfo, nil

}

func (svc SmsPaymentSettingsHttpService) SearchSmsPaymentSettings(shopID string, pageable micromodels.Pageable) ([]models.SmsPaymentSettingsInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	docList, pagination, err := svc.repo.FindPage(ctx, shopID, []string{}, pageable)

	if err != nil {
		return []models.SmsPaymentSettingsInfo{}, mongopagination.PaginationData{}, err
	}

	return docList, pagination, nil

}
