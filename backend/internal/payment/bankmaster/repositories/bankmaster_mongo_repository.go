package repositories

import (
	"context"
	"smlcloudplatform/internal/payment/bankmaster/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
)

type IBankMasterRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.BankMasterDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.BankMasterDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.BankMasterDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.BankMasterInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.BankMasterDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.BankMasterItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.BankMasterDoc, error)

	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, selectFields map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.BankMasterInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageable micromodels.Pageable) ([]models.BankMasterDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageable micromodels.Pageable) ([]models.BankMasterActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.BankMasterDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.BankMasterActivity, error)
}

type BankMasterRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.BankMasterDoc]
	repositories.SearchRepository[models.BankMasterInfo]
	repositories.GuidRepository[models.BankMasterItemGuid]
	repositories.ActivityRepository[models.BankMasterActivity, models.BankMasterDeleteActivity]
}

func NewBankMasterRepository(pst microservice.IPersisterMongo) *BankMasterRepository {

	insRepo := &BankMasterRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.BankMasterDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.BankMasterInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.BankMasterItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.BankMasterActivity, models.BankMasterDeleteActivity](pst)

	return insRepo
}
