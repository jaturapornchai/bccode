package repositories

import (
	"context"
	"smlcloudplatform/internal/productsection/sectionbusinesstype/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
)

type ISectionBusinessTypeRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.SectionBusinessTypeDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.SectionBusinessTypeDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.SectionBusinessTypeDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.SectionBusinessTypeInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.SectionBusinessTypeDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.SectionBusinessTypeItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.SectionBusinessTypeDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.SectionBusinessTypeInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.SectionBusinessTypeInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.SectionBusinessTypeDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.SectionBusinessTypeActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.SectionBusinessTypeDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.SectionBusinessTypeActivity, error)
}

type SectionBusinessTypeRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.SectionBusinessTypeDoc]
	repositories.SearchRepository[models.SectionBusinessTypeInfo]
	repositories.GuidRepository[models.SectionBusinessTypeItemGuid]
	repositories.ActivityRepository[models.SectionBusinessTypeActivity, models.SectionBusinessTypeDeleteActivity]
}

func NewSectionBusinessTypeRepository(pst microservice.IPersisterMongo) *SectionBusinessTypeRepository {

	insRepo := &SectionBusinessTypeRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.SectionBusinessTypeDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.SectionBusinessTypeInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.SectionBusinessTypeItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.SectionBusinessTypeActivity, models.SectionBusinessTypeDeleteActivity](pst)

	return insRepo
}
