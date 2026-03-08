package zone

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/restaurant/zone/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
)

type IZoneRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, category models.ZoneDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.ZoneDoc) error
	Update(ctx context.Context, shopID string, guid string, category models.ZoneDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, authUsername string, filters map[string]interface{}) error
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.ZoneInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.ZoneDoc, error)
	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.ZoneItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, columnName string, filters interface{}) (models.ZoneDoc, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.ZoneInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageable micromodels.Pageable) ([]models.ZoneDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageable micromodels.Pageable) ([]models.ZoneActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ZoneDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ZoneActivity, error)
}

type ZoneRepository struct {
	pst microservice.IPersisterMongo

	repositories.CrudRepository[models.ZoneDoc]
	repositories.SearchRepository[models.ZoneInfo]
	repositories.GuidRepository[models.ZoneItemGuid]
	repositories.ActivityRepository[models.ZoneActivity, models.ZoneDeleteActivity]
}

func NewZoneRepository(pst microservice.IPersisterMongo) ZoneRepository {
	tempRepo := ZoneRepository{
		pst: pst,
	}

	tempRepo.CrudRepository = repositories.NewCrudRepository[models.ZoneDoc](pst)
	tempRepo.SearchRepository = repositories.NewSearchRepository[models.ZoneInfo](pst)
	tempRepo.GuidRepository = repositories.NewGuidRepository[models.ZoneItemGuid](pst)
	tempRepo.ActivityRepository = repositories.NewActivityRepository[models.ZoneActivity, models.ZoneDeleteActivity](pst)

	return tempRepo
}
