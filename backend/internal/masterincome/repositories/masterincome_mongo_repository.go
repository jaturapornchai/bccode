package repositories

import (
	"context"
	"smlcloudplatform/internal/masterincome/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
)

type IMasterIncomeRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.MasterIncomeDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.MasterIncomeDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.MasterIncomeDoc) error
	DeleteByGuidfixed(sctx context.Context, hopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.MasterIncomeInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.MasterIncomeDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]models.MasterIncomeDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.MasterIncomeItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.MasterIncomeDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.MasterIncomeInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.MasterIncomeInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.MasterIncomeDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(sctx context.Context, hopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.MasterIncomeActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.MasterIncomeDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.MasterIncomeActivity, error)
}

type MasterIncomeRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.MasterIncomeDoc]
	repositories.SearchRepository[models.MasterIncomeInfo]
	repositories.GuidRepository[models.MasterIncomeItemGuid]
	repositories.ActivityRepository[models.MasterIncomeActivity, models.MasterIncomeDeleteActivity]
}

func NewMasterIncomeRepository(pst microservice.IPersisterMongo) *MasterIncomeRepository {

	insRepo := &MasterIncomeRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.MasterIncomeDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.MasterIncomeInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.MasterIncomeItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.MasterIncomeActivity, models.MasterIncomeDeleteActivity](pst)

	return insRepo
}
