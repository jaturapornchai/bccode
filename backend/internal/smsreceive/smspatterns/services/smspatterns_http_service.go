package services

import (
	"context"
	"errors"
	"regexp"
	"smlcloudplatform/internal/smsreceive/smspatterns/models"
	"smlcloudplatform/internal/smsreceive/smspatterns/repositories"
	"smlcloudplatform/internal/utils"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"strconv"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ISmsPatternsHttpService interface {
	CreateSmsPatterns(authUsername string, doc models.SmsPatterns) (string, error)
	UpdateSmsPatterns(guid string, authUsername string, doc models.SmsPatterns) error
	DeleteSmsPatterns(guid string) error
	InfoSmsPatterns(guid string) (models.SmsPatternsInfo, error)
	SearchSmsPatterns(pageable micromodels.Pageable) ([]models.SmsPatternsInfo, mongopagination.PaginationData, error)
}

type SmsPatternsHttpService struct {
	repo           repositories.ISmsPatternsRepository
	contextTimeout time.Duration
}

func NewSmsPatternsHttpService(repo repositories.ISmsPatternsRepository) *SmsPatternsHttpService {

	contextTimeout := time.Duration(15) * time.Second

	return &SmsPatternsHttpService{
		repo:           repo,
		contextTimeout: contextTimeout,
	}
}

func (svc SmsPatternsHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc SmsPatternsHttpService) CreateSmsPatterns(authUsername string, doc models.SmsPatterns) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	newGuidFixed := utils.NewGUID()

	docData := models.SmsPatternsDoc{}
	docData.GuidFixed = newGuidFixed
	docData.SmsPatterns = doc

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err := svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	return newGuidFixed, nil
}

func (svc SmsPatternsHttpService) UpdateSmsPatterns(guid string, authUsername string, doc models.SmsPatterns) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	tempCode := findDoc.Code

	findDoc.SmsPatterns = doc
	findDoc.Code = tempCode

	findDoc.UpdatedBy = authUsername
	findDoc.UpdatedAt = time.Now()

	err = svc.repo.UpdateByGuid(ctx, guid, findDoc)

	if err != nil {
		return err
	}

	return nil
}

func (svc SmsPatternsHttpService) DeleteSmsPatterns(guid string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	err = svc.repo.DeleteByGuid(ctx, guid)
	if err != nil {
		return err
	}

	return nil
}

func (svc SmsPatternsHttpService) InfoSmsPatterns(guid string) (models.SmsPatternsInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, guid)

	if err != nil {
		return models.SmsPatternsInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.SmsPatternsInfo{}, errors.New("document not found")
	}

	return findDoc.SmsPatternsInfo, nil

}

func (svc SmsPatternsHttpService) SearchSmsPatterns(pageable micromodels.Pageable) ([]models.SmsPatternsInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"name",
		"address",
	}

	docList, pagination, err := svc.repo.FindPage(ctx, searchInFields, pageable)

	if err != nil {
		return []models.SmsPatternsInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func GetAmountFromPattern(pattern string, message string) (float64, error) {
	re := regexp.MustCompile(pattern)

	reVal := re.FindStringSubmatch(message)

	if len(reVal) > 1 {

		return strconv.ParseFloat(reVal[1], 64)

	}

	return 0.0, errors.New("message not match")
}
