package settings

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/restaurant/settings/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
)

type IRestaurantSettingsRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.RestaurantSettingsDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.RestaurantSettingsDoc) error
	Update(ctx context.Context, shopID string, guid string, category models.RestaurantSettingsDoc) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.RestaurantSettingsInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.RestaurantSettingsDoc, error)
	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.RestaurantSettingsItemGuid, error)
	FindOne(ctx context.Context, shopID string, filters interface{}) (models.RestaurantSettingsDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.RestaurantSettingsInfo, mongopagination.PaginationData, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.RestaurantSettingsDoc, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageable micromodels.Pageable) ([]models.RestaurantSettingsDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageable micromodels.Pageable) ([]models.RestaurantSettingsActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.RestaurantSettingsDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.RestaurantSettingsActivity, error)
}

type RestaurantSettingsRepository struct {
	pst microservice.IPersisterMongo

	repositories.CrudRepository[models.RestaurantSettingsDoc]
	repositories.SearchRepository[models.RestaurantSettingsInfo]
	repositories.GuidRepository[models.RestaurantSettingsItemGuid]
	repositories.ActivityRepository[models.RestaurantSettingsActivity, models.RestaurantSettingsDeleteActivity]
}

func NewRestaurantSettingsRepository(pst microservice.IPersisterMongo) RestaurantSettingsRepository {
	tempRepo := RestaurantSettingsRepository{
		pst: pst,
	}

	tempRepo.CrudRepository = repositories.NewCrudRepository[models.RestaurantSettingsDoc](pst)
	tempRepo.SearchRepository = repositories.NewSearchRepository[models.RestaurantSettingsInfo](pst)
	tempRepo.GuidRepository = repositories.NewGuidRepository[models.RestaurantSettingsItemGuid](pst)
	tempRepo.ActivityRepository = repositories.NewActivityRepository[models.RestaurantSettingsActivity, models.RestaurantSettingsDeleteActivity](pst)

	return tempRepo
}
