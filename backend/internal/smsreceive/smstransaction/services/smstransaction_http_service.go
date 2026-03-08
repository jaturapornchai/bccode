package services

import (
	"context"
	"errors"
	"regexp"
	smspatternsRepo "smlcloudplatform/internal/smsreceive/smspatterns/repositories"
	smssetingsRepo "smlcloudplatform/internal/smsreceive/smspaymentsettings/repositories"
	"smlcloudplatform/internal/smsreceive/smstransaction/models"
	"smlcloudplatform/internal/smsreceive/smstransaction/repositories"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"strconv"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ISmsTransactionHttpService interface {
	CreateSmsTransaction(shopID string, authUsername string, doc models.SmsTransaction) (string, error)
	UpdateSmsTransaction(guid string, shopID string, authUsername string, doc models.SmsTransaction) error
	DeleteSmsTransaction(guid string, shopID string, authUsername string) error
	InfoSmsTransaction(guid string, shopID string) (models.SmsTransactionInfo, error)
	SearchSmsTransaction(shopID string, pageable micromodels.Pageable) ([]models.SmsTransactionInfo, mongopagination.PaginationData, error)
	CheckSMS(shopID string, storefrontGUID string, amountCheck float64, checkTime time.Time) (models.SmsTransactionCheck, error)
	ConfirmSmsTransaction(shopID string, smsTransactionGUIDFixed string) error
}

type SmsTransactionHttpService struct {
	repo           repositories.ISmsTransactionRepository
	smsPatternRepo smspatternsRepo.ISmsPatternsRepository
	smsSetingsRepo smssetingsRepo.ISmsPaymentSettingsRepository
	genGUID        func() string
	timeNow        func() time.Time
	contextTimeout time.Duration
}

func NewSmsTransactionHttpService(
	repo repositories.ISmsTransactionRepository,
	smsPatternRepo smspatternsRepo.ISmsPatternsRepository,
	smsSetingsRepo smssetingsRepo.ISmsPaymentSettingsRepository,
	genGUID func() string,
	timeNow func() time.Time,
) *SmsTransactionHttpService {

	contextTimeout := time.Duration(15) * time.Second

	return &SmsTransactionHttpService{
		repo:           repo,
		smsPatternRepo: smsPatternRepo,
		smsSetingsRepo: smsSetingsRepo,
		genGUID:        genGUID,
		timeNow:        timeNow,
		contextTimeout: contextTimeout,
	}
}

func (svc SmsTransactionHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc SmsTransactionHttpService) CreateSmsTransaction(shopID string, authUsername string, doc models.SmsTransaction) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "transid", doc.TransId)

	if err != nil {
		return "", err
	}

	if findDoc.TransId != "" {
		return "", errors.New("TransId is exists")
	}

	newGuidFixed := svc.genGUID()

	docData := models.SmsTransactionDoc{}
	docData.ShopID = shopID
	docData.SmsTransaction = doc

	docData.GuidFixed = newGuidFixed
	docData.TransId = svc.genGUID()

	docData.CreatedBy = authUsername
	docData.CreatedAt = svc.timeNow()

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	return newGuidFixed, nil
}

func (svc SmsTransactionHttpService) UpdateSmsTransaction(guid string, shopID string, authUsername string, doc models.SmsTransaction) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	tempTransId := findDoc.TransId

	findDoc.SmsTransaction = doc
	findDoc.TransId = tempTransId

	findDoc.UpdatedBy = authUsername
	findDoc.UpdatedAt = svc.timeNow()

	err = svc.repo.Update(ctx, shopID, guid, findDoc)

	if err != nil {
		return err
	}

	return nil
}

func (svc SmsTransactionHttpService) DeleteSmsTransaction(guid string, shopID string, authUsername string) error {

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

func (svc SmsTransactionHttpService) InfoSmsTransaction(guid string, shopID string) (models.SmsTransactionInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return models.SmsTransactionInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.SmsTransactionInfo{}, errors.New("document not found")
	}

	return findDoc.SmsTransactionInfo, nil

}

func (svc SmsTransactionHttpService) SearchSmsTransaction(shopID string, pageable micromodels.Pageable) ([]models.SmsTransactionInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"guidfixed",
		"transid",
	}

	docList, pagination, err := svc.repo.FindPage(ctx, shopID, searchInFields, pageable)

	if err != nil {
		return []models.SmsTransactionInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc SmsTransactionHttpService) CheckSMS(shopID string, storefrontGUID string, amountCheck float64, checkTime time.Time) (models.SmsTransactionCheck, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	storefrontSmsPaymentSettingDoc, err := svc.smsSetingsRepo.FindOne(ctx, shopID, bson.M{"storefrontguid": storefrontGUID})

	if err != nil {
		return models.SmsTransactionCheck{
			Pass:        false,
			Amount:      0,
			AmountCheck: amountCheck,
		}, err
	}

	smsPatternDoc, err := svc.smsPatternRepo.FindByCode(ctx, storefrontSmsPaymentSettingDoc.PatternCode)

	if err != nil {
		return models.SmsTransactionCheck{
			Pass:        false,
			Amount:      0,
			AmountCheck: amountCheck,
		}, err
	}

	// timeNowUnix := time.Unix(svc.timeNow().Unix(), 0)

	startTime := checkTime.Add(time.Duration(-(storefrontSmsPaymentSettingDoc.TimeMinuteBefore)) * time.Minute)
	endTime := checkTime.Add(time.Duration(storefrontSmsPaymentSettingDoc.TimeMinuteAfter) * time.Minute)

	addressKey := smsPatternDoc.Address

	smsList, err := svc.repo.FindFilterSms(ctx, shopID, storefrontGUID, addressKey, startTime, endTime)
	if err != nil {
		return models.SmsTransactionCheck{
			Pass:        false,
			Amount:      0,
			AmountCheck: amountCheck,
		}, err
	}

	var amountVal float64 = 0

	tempSmsTrans := models.SmsTransactionInfo{}
	for _, smsMessage := range smsList {
		amountVal, _ = GetAmountFromPattern(smsPatternDoc.Pattern, smsMessage.Body)
	}

	if amountVal != amountCheck {
		return models.SmsTransactionCheck{
			Pass:        false,
			Amount:      0,
			AmountCheck: amountCheck,
		}, err
	}

	return models.SmsTransactionCheck{
		SmsTransactionGUIDFixed: tempSmsTrans.GuidFixed,
		Pass:                    true,
		Amount:                  amountVal,
		AmountCheck:             amountCheck,
	}, nil
}

func (svc SmsTransactionHttpService) ConfirmSmsTransaction(shopID string, smsTransactionGUIDFixed string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, smsTransactionGUIDFixed)

	if err != nil {
		return err
	}

	findDoc.Status = 1

	return svc.repo.Update(ctx, shopID, findDoc.GuidFixed, findDoc)
}

func GetAmountFromPattern(pattern string, message string) (float64, error) {
	re := regexp.MustCompile(pattern)

	reVal := re.FindStringSubmatch(message)

	if len(reVal) > 1 {

		return strconv.ParseFloat(reVal[1], 64)

	}

	return 0.0, errors.New("message not match")
}
