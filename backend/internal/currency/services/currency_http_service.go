package services

import (
	"context"
	"errors"
	"fmt"
	"smlcloudplatform/internal/currency/models"
	"smlcloudplatform/internal/currency/repositories"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	"smlcloudplatform/internal/services"
	"smlcloudplatform/internal/utils"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"strings"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type ICurrencyHttpService interface {
	CreateCurrency(holdingCode string, authUsername string, doc models.Currency) (string, error)
	UpdateCurrency(holdingCode string, guid string, authUsername string, doc models.Currency) error
	DeleteCurrency(holdingCode string, guid string, authUsername string) error
	DeleteCurrencyByGUIDs(holdingCode string, authUsername string, GUIDs []string) error
	InfoCurrency(holdingCode string, guid string) (models.CurrencyInfo, error)
	SearchCurrency(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.CurrencyInfo, mongopagination.PaginationData, error)
	SearchCurrencyStep(holdingCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.CurrencyInfo, int, error)

	GetModuleName() string
}

type CurrencyHttpService struct {
	repo repositories.ICurrencyRepository

	syncCacheRepo mastersync.IMasterSyncCacheRepository
	services.ActivityService[models.CurrencyActivity, models.CurrencyDeleteActivity]
	contextTimeout time.Duration
}

func NewCurrencyHttpService(
	repo repositories.ICurrencyRepository,
	syncCacheRepo mastersync.IMasterSyncCacheRepository,
	contextTimeout time.Duration,
) *CurrencyHttpService {

	insSvc := &CurrencyHttpService{
		repo:           repo,
		syncCacheRepo:  syncCacheRepo,
		contextTimeout: contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.CurrencyActivity, models.CurrencyDeleteActivity](repo)

	return insSvc
}

func (svc CurrencyHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc CurrencyHttpService) CreateCurrency(holdingCode string, authUsername string, doc models.Currency) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	// Validate: code ต้องไม่ซ้ำ
	doc.Code = strings.ToUpper(strings.TrimSpace(doc.Code))
	existingDoc, err := svc.repo.FindByCode(ctx, holdingCode, doc.Code)
	if err == nil && len(existingDoc.GuidFixed) > 0 {
		return "", errors.New("currency code already exists")
	}

	newGuidFixed := utils.NewGUID()

	dataDoc := models.CurrencyDoc{}
	dataDoc.HoldingCode = holdingCode
	dataDoc.GuidFixed = newGuidFixed
	dataDoc.Currency = doc

	dataDoc.CreatedBy = authUsername
	dataDoc.CreatedAt = time.Now()

	_, err = svc.repo.Create(ctx, dataDoc)

	if err != nil {
		return "", err
	}

	go func() {
		svc.saveMasterSync(holdingCode)
	}()

	return newGuidFixed, nil
}

func (svc CurrencyHttpService) UpdateCurrency(holdingCode string, guid string, authUsername string, doc models.Currency) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	// Validate: ถ้าเปลี่ยน code ต้องไม่ซ้ำกับรายการอื่น
	doc.Code = strings.ToUpper(strings.TrimSpace(doc.Code))
	if findDoc.Code != doc.Code {
		existingDoc, err := svc.repo.FindByCode(ctx, holdingCode, doc.Code)
		if err == nil && len(existingDoc.GuidFixed) > 0 && existingDoc.GuidFixed != guid {
			return errors.New("currency code already exists")
		}
	}

	dataDoc := findDoc
	dataDoc.Currency = doc

	dataDoc.GuidFixed = findDoc.GuidFixed
	dataDoc.UpdatedBy = authUsername
	dataDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, holdingCode, guid, dataDoc)

	if err != nil {
		return err
	}

	go func() {
		svc.saveMasterSync(holdingCode)
	}()

	return nil
}

func (svc CurrencyHttpService) DeleteCurrency(holdingCode string, guid string, authUsername string) error {

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

	go func() {
		svc.saveMasterSync(holdingCode)
	}()

	return nil
}

func (svc CurrencyHttpService) DeleteCurrencyByGUIDs(holdingCode string, authUsername string, GUIDs []string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	deleteFilterQuery := map[string]interface{}{
		"guidfixed": bson.M{"$in": GUIDs},
	}

	err := svc.repo.Delete(ctx, holdingCode, authUsername, deleteFilterQuery)
	if err != nil {
		return err
	}

	go func() {
		svc.saveMasterSync(holdingCode)
	}()

	return nil
}

func (svc CurrencyHttpService) InfoCurrency(holdingCode string, guid string) (models.CurrencyInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return models.CurrencyInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.CurrencyInfo{}, errors.New("document not found")
	}

	return findDoc.CurrencyInfo, nil
}

func (svc CurrencyHttpService) SearchCurrency(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.CurrencyInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"guidfixed",
		"code",
		"name",
		"symbol",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, holdingCode, filters, searchInFields, pageable)

	if err != nil {
		return []models.CurrencyInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc CurrencyHttpService) SearchCurrencyStep(holdingCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.CurrencyInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"guidfixed",
		"code",
		"name",
		"symbol",
	}

	selectFields := map[string]interface{}{}

	docList, total, err := svc.repo.FindStep(ctx, holdingCode, filters, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.CurrencyInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc CurrencyHttpService) saveMasterSync(holdingCode string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(holdingCode, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc CurrencyHttpService) GetModuleName() string {
	return "currency"
}
