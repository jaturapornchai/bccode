package repositories

import (
	"context"
	"smlcloudplatform/internal/purchasetype/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
)

type IPurchaseTypeRepository interface {
	Count(ctx context.Context, holdingCode string) (int, error)
	Create(ctx context.Context, doc models.PurchaseTypeDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.PurchaseTypeDoc) error
	Update(ctx context.Context, holdingCode string, guid string, doc models.PurchaseTypeDoc) error
	DeleteByGuidfixed(sctx context.Context, hopID string, guid string, username string) error
	Delete(ctx context.Context, holdingCode string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable micromodels.Pageable) ([]models.PurchaseTypeInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, holdingCode string, guid string) (models.PurchaseTypeDoc, error)
	FindByGuids(ctx context.Context, holdingCode string, guids []string) ([]models.PurchaseTypeDoc, error)

	FindInItemGuid(ctx context.Context, holdingCode string, columnName string, itemGuidList []string) ([]models.PurchaseTypeItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, holdingCode string, indentityField string, indentityValue interface{}) (models.PurchaseTypeDoc, error)
	FindPageFilter(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.PurchaseTypeInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.PurchaseTypeInfo, int, error)

	FindDeletedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.PurchaseTypeDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(sctx context.Context, hopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.PurchaseTypeActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.PurchaseTypeDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.PurchaseTypeActivity, error)
}

type PurchaseTypeRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.PurchaseTypeDoc]
	repositories.SearchRepository[models.PurchaseTypeInfo]
	repositories.GuidRepository[models.PurchaseTypeItemGuid]
	repositories.ActivityRepository[models.PurchaseTypeActivity, models.PurchaseTypeDeleteActivity]
}

func NewPurchaseTypeRepository(pst microservice.IPersisterMongo) *PurchaseTypeRepository {

	insRepo := &PurchaseTypeRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.PurchaseTypeDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.PurchaseTypeInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.PurchaseTypeItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.PurchaseTypeActivity, models.PurchaseTypeDeleteActivity](pst)

	return insRepo
}
