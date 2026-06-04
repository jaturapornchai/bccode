package repositories

import (
	"context"
	"smlcloudplatform/internal/logistics/vehicle/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
)

type IVehicleRepository interface {
	Count(ctx context.Context, holdingCode string) (int, error)
	Create(ctx context.Context, doc models.VehicleDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.VehicleDoc) error
	Update(ctx context.Context, holdingCode string, guid string, doc models.VehicleDoc) error
	DeleteByGuidfixed(ctx context.Context, holdingCode string, guid string, username string) error
	Delete(ctx context.Context, holdingCode string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable micromodels.Pageable) ([]models.VehicleInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, holdingCode string, guid string) (models.VehicleDoc, error)

	FindInItemGuid(ctx context.Context, holdingCode string, columnName string, itemGuidList []string) ([]models.VehicleItemGuid, error)
	FindInItemGuids(ctx context.Context, holdingCode string, columnName string, itemGuidList []interface{}) ([]models.VehicleItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, holdingCode string, indentityField string, indentityValue interface{}) (models.VehicleDoc, error)
	FindPageFilter(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.VehicleInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.VehicleInfo, int, error)
	FindOneFilter(ctx context.Context, holdingCode string, filters map[string]interface{}) (models.VehicleDoc, error)

	FindDeletedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.VehicleDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.VehicleActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.VehicleDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.VehicleActivity, error)
}

type VehicleRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.VehicleDoc]
	repositories.SearchRepository[models.VehicleInfo]
	repositories.GuidRepository[models.VehicleItemGuid]
	repositories.ActivityRepository[models.VehicleActivity, models.VehicleDeleteActivity]
}

func NewVehicleRepository(pst microservice.IPersisterMongo) *VehicleRepository {

	insRepo := &VehicleRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.VehicleDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.VehicleInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.VehicleItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.VehicleActivity, models.VehicleDeleteActivity](pst)

	return insRepo
}
