package repositories

import (
	"context"
	"smlcloudplatform/internal/payment/bookbank/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
)

type IBookBankRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.BookBankDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.BookBankDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.BookBankDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.BookBankInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.BookBankDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.BookBankItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.BookBankDoc, error)

	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, selectFields map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.BookBankInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageable micromodels.Pageable) ([]models.BookBankDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageable micromodels.Pageable) ([]models.BookBankActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.BookBankDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.BookBankActivity, error)
}

type BookBankRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.BookBankDoc]
	repositories.SearchRepository[models.BookBankInfo]
	repositories.GuidRepository[models.BookBankItemGuid]
	repositories.ActivityRepository[models.BookBankActivity, models.BookBankDeleteActivity]
}

func NewBookBankRepository(pst microservice.IPersisterMongo) *BookBankRepository {

	insRepo := &BookBankRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.BookBankDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.BookBankInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.BookBankItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.BookBankActivity, models.BookBankDeleteActivity](pst)

	return insRepo
}
