package repositories

import (
	"context"
	"smlcloudplatform/internal/dimension/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
)

type IDimensionRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.DimensionDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.DimensionDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.DimensionDoc) error
	DeleteByGuidfixed(sctx context.Context, hopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.DimensionInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.DimensionDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]models.DimensionDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.DimensionItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.DimensionDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.DimensionInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.DimensionInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.DimensionDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(sctx context.Context, hopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.DimensionActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.DimensionDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.DimensionActivity, error)
}

type DimensionRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.DimensionDoc]
	repositories.SearchRepository[models.DimensionInfo]
	repositories.GuidRepository[models.DimensionItemGuid]
	repositories.ActivityRepository[models.DimensionActivity, models.DimensionDeleteActivity]
}

func NewDimensionRepository(pst microservice.IPersisterMongo) *DimensionRepository {

	insRepo := &DimensionRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.DimensionDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.DimensionInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.DimensionItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.DimensionActivity, models.DimensionDeleteActivity](pst)

	return insRepo
}
