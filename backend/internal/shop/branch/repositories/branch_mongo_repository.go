package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/shop/branch/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
)

type IBranchRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.BranchDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.BranchDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.BranchDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.BranchInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.BranchDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.BranchItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.BranchDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.BranchInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.BranchInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageable micromodels.Pageable) ([]models.BranchDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageable micromodels.Pageable) ([]models.BranchActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.BranchDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.BranchActivity, error)
}

type BranchRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.BranchDoc]
	repositories.SearchRepository[models.BranchInfo]
	repositories.GuidRepository[models.BranchItemGuid]
	repositories.ActivityRepository[models.BranchActivity, models.BranchDeleteActivity]
}

func NewBranchRepository(pst microservice.IPersisterMongo) *BranchRepository {

	insRepo := &BranchRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.BranchDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.BranchInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.BranchItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.BranchActivity, models.BranchDeleteActivity](pst)

	return insRepo
}
