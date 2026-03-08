package repositories

import (
	"context"
	"smlcloudplatform/internal/product/promotion/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
)

type IPromotionRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.PromotionDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.PromotionDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.PromotionDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.PromotionInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.PromotionDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.PromotionItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.PromotionDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.PromotionInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.PromotionInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.PromotionDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.PromotionActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.PromotionDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.PromotionActivity, error)
}

type PromotionRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.PromotionDoc]
	repositories.SearchRepository[models.PromotionInfo]
	repositories.GuidRepository[models.PromotionItemGuid]
	repositories.ActivityRepository[models.PromotionActivity, models.PromotionDeleteActivity]
}

func NewPromotionRepository(pst microservice.IPersisterMongo) *PromotionRepository {

	insRepo := &PromotionRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.PromotionDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.PromotionInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.PromotionItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.PromotionActivity, models.PromotionDeleteActivity](pst)

	return insRepo
}
