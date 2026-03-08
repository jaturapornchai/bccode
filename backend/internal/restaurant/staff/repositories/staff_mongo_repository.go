package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/restaurant/staff/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
)

type IStaffRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.StaffDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.StaffDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.StaffDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.StaffInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.StaffDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.StaffItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.StaffDoc, error)

	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, selectFields map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.StaffInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageable micromodels.Pageable) ([]models.StaffDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageable micromodels.Pageable) ([]models.StaffActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.StaffDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.StaffActivity, error)
}

type StaffRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.StaffDoc]
	repositories.SearchRepository[models.StaffInfo]
	repositories.GuidRepository[models.StaffItemGuid]
	repositories.ActivityRepository[models.StaffActivity, models.StaffDeleteActivity]
}

func NewStaffRepository(pst microservice.IPersisterMongo) *StaffRepository {

	insRepo := &StaffRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.StaffDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.StaffInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.StaffItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.StaffActivity, models.StaffDeleteActivity](pst)

	return insRepo
}
