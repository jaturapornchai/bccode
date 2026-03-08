package repositories

import (
	"context"
	"smlcloudplatform/internal/debtaccount/creditor/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
)

type ICreditorRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.CreditorDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.CreditorDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.CreditorDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.CreditorInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.CreditorDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]models.CreditorDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.CreditorItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.CreditorDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.CreditorInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.CreditorInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.CreditorDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.CreditorActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.CreditorDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.CreditorActivity, error)
}

type CreditorRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.CreditorDoc]
	repositories.SearchRepository[models.CreditorInfo]
	repositories.GuidRepository[models.CreditorItemGuid]
	repositories.ActivityRepository[models.CreditorActivity, models.CreditorDeleteActivity]
}

func NewCreditorRepository(pst microservice.IPersisterMongo) *CreditorRepository {

	insRepo := &CreditorRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.CreditorDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.CreditorInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.CreditorItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.CreditorActivity, models.CreditorDeleteActivity](pst)

	return insRepo
}
