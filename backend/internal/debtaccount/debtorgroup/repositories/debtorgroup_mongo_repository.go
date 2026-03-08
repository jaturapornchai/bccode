package repositories

import (
	"context"
	"smlcloudplatform/internal/debtaccount/debtorgroup/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
)

type IDebtorGroupRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.DebtorGroupDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.DebtorGroupDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.DebtorGroupDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.DebtorGroupInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.DebtorGroupDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]models.DebtorGroupDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.DebtorGroupItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.DebtorGroupDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.DebtorGroupInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.DebtorGroupInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.DebtorGroupDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.DebtorGroupActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.DebtorGroupDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.DebtorGroupActivity, error)
}

type DebtorGroupRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.DebtorGroupDoc]
	repositories.SearchRepository[models.DebtorGroupInfo]
	repositories.GuidRepository[models.DebtorGroupItemGuid]
	repositories.ActivityRepository[models.DebtorGroupActivity, models.DebtorGroupDeleteActivity]
}

func NewDebtorGroupRepository(pst microservice.IPersisterMongo) *DebtorGroupRepository {

	insRepo := &DebtorGroupRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.DebtorGroupDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.DebtorGroupInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.DebtorGroupItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.DebtorGroupActivity, models.DebtorGroupDeleteActivity](pst)

	return insRepo
}
