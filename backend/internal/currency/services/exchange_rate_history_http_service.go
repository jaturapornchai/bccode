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
)

type IExchangeRateHistoryHttpService interface {
	CreateExchangeRateHistory(holdingCode string, authUsername string, currency string, date string, rate float64) (string, error)
	UpdateExchangeRateHistory(holdingCode string, guid string, authUsername string, date string, rate float64) error
	DeleteExchangeRateHistory(holdingCode string, guid string, authUsername string) error
	DeleteExchangeRateHistoryByGUIDs(holdingCode string, authUsername string, GUIDs []string) error
	InfoExchangeRateHistory(holdingCode string, guid string) (models.ExchangeRateEntry, error)
	SearchExchangeRateHistory(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ExchangeRateEntry, mongopagination.PaginationData, error)
	SearchExchangeRateHistoryStep(holdingCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ExchangeRateEntry, int, error)

	GetLatestExchangeRate(holdingCode string, currency string, date string) (models.ExchangeRateEntry, error)

	GetModuleName() string
}

type ExchangeRateHistoryHttpService struct {
	currencyRepo repositories.ICurrencyRepository

	syncCacheRepo mastersync.IMasterSyncCacheRepository
	services.ActivityService[models.CurrencyActivity, models.CurrencyDeleteActivity]
	contextTimeout time.Duration
}

func NewExchangeRateHistoryHttpService(
	currencyRepo repositories.ICurrencyRepository,
	syncCacheRepo mastersync.IMasterSyncCacheRepository,
	contextTimeout time.Duration,
) *ExchangeRateHistoryHttpService {

	insSvc := &ExchangeRateHistoryHttpService{
		currencyRepo:   currencyRepo,
		syncCacheRepo:  syncCacheRepo,
		contextTimeout: contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.CurrencyActivity, models.CurrencyDeleteActivity](currencyRepo)

	return insSvc
}

func (svc ExchangeRateHistoryHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc ExchangeRateHistoryHttpService) CreateExchangeRateHistory(holdingCode string, authUsername string, currency string, date string, rate float64) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	// Normalize currency code
	currency = strings.ToUpper(strings.TrimSpace(currency))

	// Find currency document
	currencyDoc, err := svc.currencyRepo.FindByCode(ctx, holdingCode, currency)
	if err != nil {
		return "", err
	}
	if len(currencyDoc.GuidFixed) < 1 {
		return "", errors.New("currency not found")
	}

	// Create new exchange rate entry
	newGuidFixed := utils.NewGUID()
	newEntry := models.ExchangeRateEntry{
		GuidFixed: newGuidFixed,
		Date:      date,
		Rate:      rate,
	}

	// Add to ExchangeRates array
	if currencyDoc.ExchangeRates == nil {
		currencyDoc.ExchangeRates = []models.ExchangeRateEntry{}
	}
	currencyDoc.ExchangeRates = append(currencyDoc.ExchangeRates, newEntry)

	// Update document
	currencyDoc.UpdatedBy = authUsername
	currencyDoc.UpdatedAt = time.Now()

	err = svc.currencyRepo.Update(ctx, holdingCode, currencyDoc.GuidFixed, currencyDoc)
	if err != nil {
		return "", err
	}

	go func() {
		svc.saveMasterSync(holdingCode)
	}()

	return newGuidFixed, nil
}

func (svc ExchangeRateHistoryHttpService) UpdateExchangeRateHistory(holdingCode string, guid string, authUsername string, date string, rate float64) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	// Find currency containing this exchange rate guid
	currencyDoc, err := svc.currencyRepo.FindByExchangeRateGuid(ctx, holdingCode, guid)
	if err != nil {
		return err
	}
	if len(currencyDoc.GuidFixed) < 1 {
		return errors.New("exchange rate not found")
	}

	// Update the specific rate entry
	found := false
	for i := range currencyDoc.ExchangeRates {
		if currencyDoc.ExchangeRates[i].GuidFixed == guid {
			currencyDoc.ExchangeRates[i].Date = date
			currencyDoc.ExchangeRates[i].Rate = rate
			found = true
			break
		}
	}

	if !found {
		return errors.New("exchange rate entry not found")
	}

	// Update document
	currencyDoc.UpdatedBy = authUsername
	currencyDoc.UpdatedAt = time.Now()

	err = svc.currencyRepo.Update(ctx, holdingCode, currencyDoc.GuidFixed, currencyDoc)
	if err != nil {
		return err
	}

	go func() {
		svc.saveMasterSync(holdingCode)
	}()

	return nil
}

func (svc ExchangeRateHistoryHttpService) DeleteExchangeRateHistory(holdingCode string, guid string, authUsername string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	// Find currency containing this exchange rate guid
	currencyDoc, err := svc.currencyRepo.FindByExchangeRateGuid(ctx, holdingCode, guid)
	if err != nil {
		return err
	}
	if len(currencyDoc.GuidFixed) < 1 {
		return errors.New("exchange rate not found")
	}

	// Remove the specific rate entry from array
	newRates := []models.ExchangeRateEntry{}
	for _, rate := range currencyDoc.ExchangeRates {
		if rate.GuidFixed != guid {
			newRates = append(newRates, rate)
		}
	}
	currencyDoc.ExchangeRates = newRates

	// Update document
	currencyDoc.UpdatedBy = authUsername
	currencyDoc.UpdatedAt = time.Now()

	err = svc.currencyRepo.Update(ctx, holdingCode, currencyDoc.GuidFixed, currencyDoc)
	if err != nil {
		return err
	}

	go func() {
		svc.saveMasterSync(holdingCode)
	}()

	return nil
}

func (svc ExchangeRateHistoryHttpService) DeleteExchangeRateHistoryByGUIDs(holdingCode string, authUsername string, GUIDs []string) error {

	// Delete each GUID individually
	for _, guid := range GUIDs {
		err := svc.DeleteExchangeRateHistory(holdingCode, guid, authUsername)
		if err != nil {
			// Continue with other deletes even if one fails
			continue
		}
	}

	return nil
}

func (svc ExchangeRateHistoryHttpService) InfoExchangeRateHistory(holdingCode string, guid string) (models.ExchangeRateEntry, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	// Find currency containing this exchange rate guid
	currencyDoc, err := svc.currencyRepo.FindByExchangeRateGuid(ctx, holdingCode, guid)
	if err != nil {
		return models.ExchangeRateEntry{}, err
	}
	if len(currencyDoc.GuidFixed) < 1 {
		return models.ExchangeRateEntry{}, errors.New("exchange rate not found")
	}

	// Find and return the specific exchange rate entry
	for _, rate := range currencyDoc.ExchangeRates {
		if rate.GuidFixed == guid {
			return rate, nil
		}
	}

	return models.ExchangeRateEntry{}, errors.New("exchange rate entry not found in array")
}

func (svc ExchangeRateHistoryHttpService) SearchExchangeRateHistory(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ExchangeRateEntry, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	// Get currency filter if exists
	currencyFilter := ""
	if cf, ok := filters["currency"]; ok {
		currencyFilter = strings.ToUpper(fmt.Sprintf("%v", cf))
	}

	// Find currencies (filter by currency code if specified)
	currencyFilters := map[string]interface{}{}
	if currencyFilter != "" {
		currencyFilters["code"] = currencyFilter
	}

	// Get all currencies matching filter
	currencies, _, err := svc.currencyRepo.FindPageFilter(ctx, holdingCode, currencyFilters, []string{"code"}, micromodels.Pageable{
		Page:  1,
		Limit: 1000, // Get all currencies
	})
	if err != nil {
		return []models.ExchangeRateEntry{}, mongopagination.PaginationData{}, err
	}

	// Flatten all exchange rates from all currencies
	allRates := []models.ExchangeRateEntry{}
	for _, currency := range currencies {
		// Get full currency doc to access exchangerates
		currencyDoc, err := svc.currencyRepo.FindByGuid(ctx, holdingCode, currency.GuidFixed)
		if err != nil {
			continue
		}
		allRates = append(allRates, currencyDoc.ExchangeRates...)
	}

	// Manual pagination
	start := (pageable.Page - 1) * pageable.Limit
	end := start + pageable.Limit
	if start > len(allRates) {
		start = len(allRates)
	}
	if end > len(allRates) {
		end = len(allRates)
	}

	paginatedRates := allRates[start:end]
	pagination := mongopagination.PaginationData{
		Total: int64(len(allRates)),
		Page:  int64(pageable.Page),
	}

	return paginatedRates, pagination, nil
}

func (svc ExchangeRateHistoryHttpService) SearchExchangeRateHistoryStep(holdingCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ExchangeRateEntry, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	// Get currency filter if exists
	currencyFilter := ""
	if cf, ok := filters["currency"]; ok {
		currencyFilter = strings.ToUpper(fmt.Sprintf("%v", cf))
	}

	// Find currencies (filter by currency code if specified)
	currencyFilters := map[string]interface{}{}
	if currencyFilter != "" {
		currencyFilters["code"] = currencyFilter
	}

	// Get all currencies matching filter
	currencies, _, err := svc.currencyRepo.FindStep(ctx, holdingCode, currencyFilters, []string{"code"}, map[string]interface{}{}, micromodels.PageableStep{
		Skip:  0,
		Limit: 1000, // Get all currencies
	})
	if err != nil {
		return []models.ExchangeRateEntry{}, 0, err
	}

	// Flatten all exchange rates from all currencies
	allRates := []models.ExchangeRateEntry{}
	for _, currency := range currencies {
		// Get full currency doc to access exchangerates
		currencyDoc, err := svc.currencyRepo.FindByGuid(ctx, holdingCode, currency.GuidFixed)
		if err != nil {
			continue
		}
		allRates = append(allRates, currencyDoc.ExchangeRates...)
	}

	// Manual pagination with skip/limit
	start := pageableStep.Skip
	end := start + pageableStep.Limit
	if start > len(allRates) {
		start = len(allRates)
	}
	if end > len(allRates) {
		end = len(allRates)
	}

	return allRates[start:end], len(allRates), nil
}

// GetLatestExchangeRate - หาอัตราแลกเปลี่ยนล่าสุดที่ <= วันที่ที่ระบุ
func (svc ExchangeRateHistoryHttpService) GetLatestExchangeRate(holdingCode string, currency string, date string) (models.ExchangeRateEntry, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	currency = strings.ToUpper(strings.TrimSpace(currency))

	// Find currency by code
	currencyDoc, err := svc.currencyRepo.FindByCode(ctx, holdingCode, currency)
	if err != nil {
		return models.ExchangeRateEntry{}, err
	}
	if len(currencyDoc.GuidFixed) < 1 {
		return models.ExchangeRateEntry{}, errors.New("currency not found")
	}

	// Find latest rate <= date
	var latestRate *models.ExchangeRateEntry
	for i := range currencyDoc.ExchangeRates {
		rate := &currencyDoc.ExchangeRates[i]
		// Only consider rates on or before the specified date
		if rate.Date <= date {
			if latestRate == nil || rate.Date > latestRate.Date {
				latestRate = rate
			}
		}
	}

	if latestRate == nil {
		return models.ExchangeRateEntry{}, errors.New("exchange rate not found for date")
	}

	return *latestRate, nil
}

func (svc ExchangeRateHistoryHttpService) saveMasterSync(holdingCode string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(holdingCode, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc ExchangeRateHistoryHttpService) GetModuleName() string {
	return "exchangeratehistory"
}
