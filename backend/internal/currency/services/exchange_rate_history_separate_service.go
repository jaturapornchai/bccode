package services

import (
	"context"
	"errors"
	"smlcloudplatform/internal/currency/models"
	"smlcloudplatform/internal/currency/repositories"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	commonmodels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/services"
	"smlcloudplatform/internal/utils"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"strings"
	"time"

	"github.com/smlsoft/mongopagination"
)

// IExchangeRateHistorySeparateService - ใช้ separate collection (exchangeRateHistory)
type IExchangeRateHistorySeparateService interface {
	CreateExchangeRateHistory(holdingCode string, authUsername string, currency string, date string, rate float64) (string, error)
	UpdateExchangeRateHistory(holdingCode string, guid string, authUsername string, date string, rate float64) error
	DeleteExchangeRateHistory(holdingCode string, guid string, authUsername string) error
	DeleteExchangeRateHistoryByGUIDs(holdingCode string, authUsername string, GUIDs []string) error
	InfoExchangeRateHistory(holdingCode string, guid string) (models.ExchangeRateHistoryDoc, error)
	SearchExchangeRateHistory(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ExchangeRateHistoryInfo, mongopagination.PaginationData, error)
	SearchExchangeRateHistoryStep(holdingCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ExchangeRateHistoryInfo, int, error)
	GetLatestExchangeRate(holdingCode string, currency string, date string) (models.ExchangeRateHistoryDoc, error)
	GetModuleName() string
}

type ExchangeRateHistorySeparateService struct {
	repo          repositories.IExchangeRateHistoryRepository
	syncCacheRepo mastersync.IMasterSyncCacheRepository
	services.ActivityService[models.ExchangeRateHistoryActivity, models.ExchangeRateHistoryDeleteActivity]
	contextTimeout time.Duration
}

func NewExchangeRateHistorySeparateService(
	repo repositories.IExchangeRateHistoryRepository,
	syncCacheRepo mastersync.IMasterSyncCacheRepository,
	contextTimeout time.Duration,
) *ExchangeRateHistorySeparateService {
	insSvc := &ExchangeRateHistorySeparateService{
		repo:           repo,
		syncCacheRepo:  syncCacheRepo,
		contextTimeout: contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.ExchangeRateHistoryActivity, models.ExchangeRateHistoryDeleteActivity](repo)

	return insSvc
}

func (svc ExchangeRateHistorySeparateService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc ExchangeRateHistorySeparateService) CreateExchangeRateHistory(holdingCode string, authUsername string, currency string, date string, rate float64) (string, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	// Normalize currency code
	currency = strings.ToUpper(strings.TrimSpace(currency))

	// Validate
	if currency == "" {
		return "", errors.New("currency is required")
	}
	if date == "" {
		return "", errors.New("date is required")
	}
	if rate <= 0 {
		return "", errors.New("rate must be greater than 0")
	}

	// Create new document
	newGUID := utils.NewGUID()
	now := time.Now()

	doc := models.ExchangeRateHistoryDoc{
		ExchangeRateHistoryData: models.ExchangeRateHistoryData{
			HoldingCodeentity: commonmodels.HoldingCodeentity{HoldingCode: holdingCode},
			ExchangeRateHistoryInfo: models.ExchangeRateHistoryInfo{
				DocIdentity: commonmodels.DocIdentity{GuidFixed: newGUID},
				ExchangeRateHistory: models.ExchangeRateHistory{
					Currency: currency,
					Date:     date,
					Rate:     rate,
					Source:   "manual",
				},
			},
		},
		ActivityDoc: commonmodels.ActivityDoc{
			CreatedBy: authUsername,
			CreatedAt: now,
			UpdatedBy: authUsername,
			UpdatedAt: now,
		},
	}

	guid, err := svc.repo.Create(ctx, doc)
	if err != nil {
		return "", err
	}

	go svc.saveMasterSync(holdingCode)

	return guid, nil
}

func (svc ExchangeRateHistorySeparateService) UpdateExchangeRateHistory(holdingCode string, guid string, authUsername string, date string, rate float64) error {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	// Find existing document
	doc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)
	if err != nil {
		return err
	}
	if doc.GuidFixed == "" {
		return errors.New("exchange rate not found")
	}

	// Update fields
	doc.Date = date
	doc.Rate = rate
	doc.UpdatedBy = authUsername
	doc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, holdingCode, guid, doc)
	if err != nil {
		return err
	}

	go svc.saveMasterSync(holdingCode)

	return nil
}

func (svc ExchangeRateHistorySeparateService) DeleteExchangeRateHistory(holdingCode string, guid string, authUsername string) error {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	err := svc.repo.DeleteByGuidfixed(ctx, holdingCode, guid, authUsername)
	if err != nil {
		return err
	}

	go svc.saveMasterSync(holdingCode)

	return nil
}

func (svc ExchangeRateHistorySeparateService) DeleteExchangeRateHistoryByGUIDs(holdingCode string, authUsername string, GUIDs []string) error {
	for _, guid := range GUIDs {
		err := svc.DeleteExchangeRateHistory(holdingCode, guid, authUsername)
		if err != nil {
			continue // Continue with other deletes
		}
	}
	return nil
}

func (svc ExchangeRateHistorySeparateService) InfoExchangeRateHistory(holdingCode string, guid string) (models.ExchangeRateHistoryDoc, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	doc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)
	if err != nil {
		return models.ExchangeRateHistoryDoc{}, err
	}
	if doc.GuidFixed == "" {
		return models.ExchangeRateHistoryDoc{}, errors.New("exchange rate not found")
	}

	return doc, nil
}

func (svc ExchangeRateHistorySeparateService) SearchExchangeRateHistory(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ExchangeRateHistoryInfo, mongopagination.PaginationData, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{"currency", "date"}

	// Add default sort by date descending
	if len(pageable.Sorts) == 0 {
		pageable.Sorts = []micromodels.KeyInt{
			{Key: "date", Value: -1},
		}
	}

	return svc.repo.FindPageFilter(ctx, holdingCode, filters, searchInFields, pageable)
}

func (svc ExchangeRateHistorySeparateService) SearchExchangeRateHistoryStep(holdingCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ExchangeRateHistoryInfo, int, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{"currency", "date"}

	return svc.repo.FindStep(ctx, holdingCode, filters, searchInFields, nil, pageableStep)
}

func (svc ExchangeRateHistorySeparateService) GetLatestExchangeRate(holdingCode string, currency string, date string) (models.ExchangeRateHistoryDoc, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	currency = strings.ToUpper(strings.TrimSpace(currency))

	doc, err := svc.repo.FindLatestRate(ctx, holdingCode, currency, date)
	if err != nil {
		return models.ExchangeRateHistoryDoc{}, err
	}
	if doc.GuidFixed == "" {
		return models.ExchangeRateHistoryDoc{}, errors.New("exchange rate not found for date")
	}

	return doc, nil
}

func (svc ExchangeRateHistorySeparateService) saveMasterSync(holdingCode string) {
	if svc.syncCacheRepo != nil {
		svc.syncCacheRepo.Save(holdingCode, svc.GetModuleName())
	}
}

func (svc ExchangeRateHistorySeparateService) GetModuleName() string {
	return "exchange_rate_history"
}
