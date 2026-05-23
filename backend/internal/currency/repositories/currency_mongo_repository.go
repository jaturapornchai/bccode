package repositories

import (
	"context"
	"smlcloudplatform/internal/currency/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
)

type ICurrencyRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.CurrencyDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.CurrencyDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.CurrencyDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.CurrencyInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.CurrencyDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]models.CurrencyDoc, error)
	FindByCode(ctx context.Context, shopID string, code string) (models.CurrencyDoc, error)
	FindByExchangeRateGuid(ctx context.Context, shopID string, rateGuid string) (models.CurrencyDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.CurrencyInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.CurrencyInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.CurrencyDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.CurrencyActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.CurrencyDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.CurrencyActivity, error)
}

type CurrencyRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.CurrencyDoc]
	repositories.SearchRepository[models.CurrencyInfo]
	repositories.ActivityRepository[models.CurrencyActivity, models.CurrencyDeleteActivity]
}

func NewCurrencyRepository(pst microservice.IPersisterMongo) *CurrencyRepository {
	insRepo := &CurrencyRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.CurrencyDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.CurrencyInfo](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.CurrencyActivity, models.CurrencyDeleteActivity](pst)

	return insRepo
}

// FindByCode - หาสกุลเงินจาก code (USD, EUR, JPY, etc.)
func (r *CurrencyRepository) FindByCode(ctx context.Context, shopID string, code string) (models.CurrencyDoc, error) {
	filters := map[string]interface{}{
		"code": code,
	}

	doc := models.CurrencyDoc{}
	err := r.pst.FindOne(ctx, shopID, filters, &doc)

	return doc, err
}

// FindByExchangeRateGuid - หาสกุลเงินที่มี exchange rate guid ที่ระบุ
func (r *CurrencyRepository) FindByExchangeRateGuid(ctx context.Context, shopID string, rateGuid string) (models.CurrencyDoc, error) {
	filters := map[string]interface{}{
		"exchange_rates.guid_fixed": rateGuid,
	}

	doc := models.CurrencyDoc{}
	err := r.pst.FindOne(ctx, shopID, filters, &doc)

	return doc, err
}
