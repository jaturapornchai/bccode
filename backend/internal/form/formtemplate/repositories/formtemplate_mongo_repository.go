package repositories

import (
	"context"
	"smlcloudplatform/internal/form/formtemplate/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
)

type IFormTemplateRepository interface {
	Count(ctx context.Context, holdingCode string) (int, error)
	Create(ctx context.Context, doc models.FormTemplateDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.FormTemplateDoc) error
	Update(ctx context.Context, holdingCode string, guid string, doc models.FormTemplateDoc) error
	DeleteByGuidfixed(ctx context.Context, holdingCode string, guid string, username string) error
	Delete(ctx context.Context, holdingCode string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable micromodels.Pageable) ([]models.FormTemplateInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, holdingCode string, guid string) (models.FormTemplateDoc, error)
	FindByGuids(ctx context.Context, holdingCode string, guids []string) ([]models.FormTemplateDoc, error)
	FindInItemGuid(ctx context.Context, holdingCode string, columnName string, itemGuidList []string) ([]models.FormTemplateItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, holdingCode string, indentityField string, indentityValue interface{}) (models.FormTemplateDoc, error)
	FindPageFilter(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.FormTemplateInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.FormTemplateInfo, int, error)

	FindDeletedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageable micromodels.Pageable) ([]models.FormTemplateDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageable micromodels.Pageable) ([]models.FormTemplateActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.FormTemplateDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.FormTemplateActivity, error)
}

type FormTemplateRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.FormTemplateDoc]
	repositories.SearchRepository[models.FormTemplateInfo]
	repositories.GuidRepository[models.FormTemplateItemGuid]
	repositories.ActivityRepository[models.FormTemplateActivity, models.FormTemplateDeleteActivity]
}

func NewFormTemplateRepository(pst microservice.IPersisterMongo) *FormTemplateRepository {
	insRepo := &FormTemplateRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.FormTemplateDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.FormTemplateInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.FormTemplateItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.FormTemplateActivity, models.FormTemplateDeleteActivity](pst)

	return insRepo
}
