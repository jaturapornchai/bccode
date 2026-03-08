package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/restaurant/device/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
)

type IDeviceRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.DeviceDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.DeviceDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.DeviceDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.DeviceInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.DeviceDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.DeviceItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.DeviceDoc, error)

	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, selectFields map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.DeviceInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageable micromodels.Pageable) ([]models.DeviceDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageable micromodels.Pageable) ([]models.DeviceActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.DeviceDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.DeviceActivity, error)
}

type DeviceRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.DeviceDoc]
	repositories.SearchRepository[models.DeviceInfo]
	repositories.GuidRepository[models.DeviceItemGuid]
	repositories.ActivityRepository[models.DeviceActivity, models.DeviceDeleteActivity]
}

func NewDeviceRepository(pst microservice.IPersisterMongo) *DeviceRepository {

	insRepo := &DeviceRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.DeviceDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.DeviceInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.DeviceItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.DeviceActivity, models.DeviceDeleteActivity](pst)

	return insRepo
}
