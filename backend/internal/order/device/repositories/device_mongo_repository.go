package repositories

import (
	"context"
	"smlcloudplatform/internal/order/device/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
)

type IDeviceRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.OrderDeviceDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.OrderDeviceDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.OrderDeviceDoc) error
	DeleteByGuidfixed(sctx context.Context, hopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.OrderDeviceInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.OrderDeviceDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]models.OrderDeviceDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.OrderDeviceItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.OrderDeviceDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.OrderDeviceInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.OrderDeviceInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.OrderDeviceDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(sctx context.Context, hopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.OrderDeviceActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.OrderDeviceDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.OrderDeviceActivity, error)
}

type DeviceRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.OrderDeviceDoc]
	repositories.SearchRepository[models.OrderDeviceInfo]
	repositories.GuidRepository[models.OrderDeviceItemGuid]
	repositories.ActivityRepository[models.OrderDeviceActivity, models.OrderDeviceDeleteActivity]
}

func NewDeviceRepository(pst microservice.IPersisterMongo) *DeviceRepository {

	insRepo := &DeviceRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.OrderDeviceDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.OrderDeviceInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.OrderDeviceItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.OrderDeviceActivity, models.OrderDeviceDeleteActivity](pst)

	return insRepo
}
