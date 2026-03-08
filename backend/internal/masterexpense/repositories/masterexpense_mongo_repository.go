package repositories

import (
	"context"
	"smlcloudplatform/internal/masterexpense/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
)

type IMasterExpenseRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.MasterExpenseDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.MasterExpenseDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.MasterExpenseDoc) error
	DeleteByGuidfixed(sctx context.Context, hopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.MasterExpenseInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.MasterExpenseDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]models.MasterExpenseDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.MasterExpenseItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.MasterExpenseDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.MasterExpenseInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.MasterExpenseInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.MasterExpenseDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(sctx context.Context, hopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.MasterExpenseActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.MasterExpenseDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.MasterExpenseActivity, error)
}

type MasterExpenseRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.MasterExpenseDoc]
	repositories.SearchRepository[models.MasterExpenseInfo]
	repositories.GuidRepository[models.MasterExpenseItemGuid]
	repositories.ActivityRepository[models.MasterExpenseActivity, models.MasterExpenseDeleteActivity]
}

func NewMasterExpenseRepository(pst microservice.IPersisterMongo) *MasterExpenseRepository {

	insRepo := &MasterExpenseRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.MasterExpenseDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.MasterExpenseInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.MasterExpenseItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.MasterExpenseActivity, models.MasterExpenseDeleteActivity](pst)

	return insRepo
}
