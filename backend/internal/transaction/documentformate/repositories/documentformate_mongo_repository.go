package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/documentformate/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
)

type IDocumentFormateRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.DocumentFormateDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.DocumentFormateDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.DocumentFormateDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.DocumentFormateInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.DocumentFormateDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.DocumentFormateItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.DocumentFormateDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.DocumentFormateInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.DocumentFormateInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.DocumentFormateDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.DocumentFormateActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.DocumentFormateDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.DocumentFormateActivity, error)
}

type DocumentFormateRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.DocumentFormateDoc]
	repositories.SearchRepository[models.DocumentFormateInfo]
	repositories.GuidRepository[models.DocumentFormateItemGuid]
	repositories.ActivityRepository[models.DocumentFormateActivity, models.DocumentFormateDeleteActivity]
}

func NewDocumentFormateRepository(pst microservice.IPersisterMongo) *DocumentFormateRepository {

	insRepo := &DocumentFormateRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.DocumentFormateDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.DocumentFormateInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.DocumentFormateItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.DocumentFormateActivity, models.DocumentFormateDeleteActivity](pst)

	return insRepo
}
