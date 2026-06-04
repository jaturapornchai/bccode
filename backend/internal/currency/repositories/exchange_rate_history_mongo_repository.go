package repositories

import (
	"context"
	"smlcloudplatform/internal/currency/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type IExchangeRateHistoryRepository interface {
	Count(ctx context.Context, holdingCode string) (int, error)
	Create(ctx context.Context, doc models.ExchangeRateHistoryDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.ExchangeRateHistoryDoc) error
	Update(ctx context.Context, holdingCode string, guid string, doc models.ExchangeRateHistoryDoc) error
	DeleteByGuidfixed(ctx context.Context, holdingCode string, guid string, username string) error
	Delete(ctx context.Context, holdingCode string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable micromodels.Pageable) ([]models.ExchangeRateHistoryInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, holdingCode string, guid string) (models.ExchangeRateHistoryDoc, error)
	FindByGuids(ctx context.Context, holdingCode string, guids []string) ([]models.ExchangeRateHistoryDoc, error)
	FindPageFilter(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.ExchangeRateHistoryInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.ExchangeRateHistoryInfo, int, error)

	FindLatestRate(ctx context.Context, holdingCode string, currency string, date string) (models.ExchangeRateHistoryDoc, error)

	FindDeletedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ExchangeRateHistoryDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ExchangeRateHistoryActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ExchangeRateHistoryDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ExchangeRateHistoryActivity, error)
}

type ExchangeRateHistoryRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.ExchangeRateHistoryDoc]
	repositories.SearchRepository[models.ExchangeRateHistoryInfo]
	repositories.ActivityRepository[models.ExchangeRateHistoryActivity, models.ExchangeRateHistoryDeleteActivity]
}

func NewExchangeRateHistoryRepository(pst microservice.IPersisterMongo) *ExchangeRateHistoryRepository {
	insRepo := &ExchangeRateHistoryRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.ExchangeRateHistoryDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.ExchangeRateHistoryInfo](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.ExchangeRateHistoryActivity, models.ExchangeRateHistoryDeleteActivity](pst)

	return insRepo
}

// FindLatestRate - หาอัตราแลกเปลี่ยนล่าสุดที่ <= วันที่ที่ระบุ
func (r *ExchangeRateHistoryRepository) FindLatestRate(ctx context.Context, holdingCode string, currency string, date string) (models.ExchangeRateHistoryDoc, error) {
	// Query: currency = ? AND date <= ? ORDER BY date DESC LIMIT 1
	filters := map[string]interface{}{
		"currency": currency,
		"date":     bson.M{"$lte": date},
	}

	searchInFields := []string{}
	pageable := micromodels.Pageable{
		Page:  1,
		Limit: 1,
		Sorts: []micromodels.KeyInt{
			{Key: "date", Value: -1}, // -1 = descending (latest first)
		},
	}

	docList, _, err := r.SearchRepository.FindPageFilter(ctx, holdingCode, filters, searchInFields, pageable)
	if err != nil {
		return models.ExchangeRateHistoryDoc{}, err
	}

	if len(docList) == 0 {
		return models.ExchangeRateHistoryDoc{}, nil
	}

	// Convert ExchangeRateHistoryInfo to ExchangeRateHistoryDoc
	doc := models.ExchangeRateHistoryDoc{
		ExchangeRateHistoryData: models.ExchangeRateHistoryData{
			ExchangeRateHistoryInfo: docList[0],
		},
	}

	return doc, nil
}
