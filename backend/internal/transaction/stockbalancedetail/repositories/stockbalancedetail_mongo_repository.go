package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/stockbalancedetail/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
)

type IStockBalanceDetailRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.StockBalanceDetailDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.StockBalanceDetailDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.StockBalanceDetailDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.StockBalanceDetailInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.StockBalanceDetailDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]models.StockBalanceDetailDoc, error)
	FindByDocIndentityGuids(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) ([]models.StockBalanceDetailDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.StockBalanceDetailItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.StockBalanceDetailDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.StockBalanceDetailInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.StockBalanceDetailInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.StockBalanceDetailDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.StockBalanceDetailActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.StockBalanceDetailDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.StockBalanceDetailActivity, error)
}

type StockBalanceDetailRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.StockBalanceDetailDoc]
	repositories.SearchRepository[models.StockBalanceDetailInfo]
	repositories.GuidRepository[models.StockBalanceDetailItemGuid]
	repositories.ActivityRepository[models.StockBalanceDetailActivity, models.StockBalanceDetailDeleteActivity]
}

func NewStockBalanceDetailRepository(pst microservice.IPersisterMongo) *StockBalanceDetailRepository {

	insRepo := &StockBalanceDetailRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.StockBalanceDetailDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.StockBalanceDetailInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.StockBalanceDetailItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.StockBalanceDetailActivity, models.StockBalanceDetailDeleteActivity](pst)

	return insRepo
}
