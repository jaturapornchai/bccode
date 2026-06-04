package repositories

import (
	"context"
	"smlcloudplatform/internal/pickandpack/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
)

type IDeviceRepository interface {
	Count(ctx context.Context, holdingCode string) (int, error)
	Create(ctx context.Context, doc models.PickandpackDeviceDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.PickandpackDeviceDoc) error
	Update(ctx context.Context, holdingCode string, guid string, doc models.PickandpackDeviceDoc) error
	DeleteByGuidfixed(sctx context.Context, hopID string, guid string, username string) error
	Delete(ctx context.Context, holdingCode string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable micromodels.Pageable) ([]models.PickandpackDeviceInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, holdingCode string, guid string) (models.PickandpackDeviceDoc, error)
	FindByGuids(ctx context.Context, holdingCode string, guids []string) ([]models.PickandpackDeviceDoc, error)

	FindInItemGuid(ctx context.Context, holdingCode string, columnName string, itemGuidList []string) ([]models.PickandpackDeviceItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, holdingCode string, indentityField string, indentityValue interface{}) (models.PickandpackDeviceDoc, error)
	FindPageFilter(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.PickandpackDeviceInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.PickandpackDeviceInfo, int, error)

	FindDeletedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.PickandpackDeviceDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(sctx context.Context, hopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.PickandpackDeviceActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.PickandpackDeviceDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.PickandpackDeviceActivity, error)
}

type DeviceRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.PickandpackDeviceDoc]
	repositories.SearchRepository[models.PickandpackDeviceInfo]
	repositories.GuidRepository[models.PickandpackDeviceItemGuid]
	repositories.ActivityRepository[models.PickandpackDeviceActivity, models.PickandpackDeviceDeleteActivity]
}

func NewDeviceRepository(pst microservice.IPersisterMongo) *DeviceRepository {

	insRepo := &DeviceRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.PickandpackDeviceDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.PickandpackDeviceInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.PickandpackDeviceItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.PickandpackDeviceActivity, models.PickandpackDeviceDeleteActivity](pst)

	return insRepo
}
