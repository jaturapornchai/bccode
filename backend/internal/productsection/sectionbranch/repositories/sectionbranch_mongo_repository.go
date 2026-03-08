package repositories

import (
	"context"
	"smlcloudplatform/internal/productsection/sectionbranch/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
)

type ISectionBranchRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.SectionBranchDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.SectionBranchDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.SectionBranchDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.SectionBranchInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.SectionBranchDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.SectionBranchItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.SectionBranchDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.SectionBranchInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.SectionBranchInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.SectionBranchDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.SectionBranchActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.SectionBranchDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.SectionBranchActivity, error)
}

type SectionBranchRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.SectionBranchDoc]
	repositories.SearchRepository[models.SectionBranchInfo]
	repositories.GuidRepository[models.SectionBranchItemGuid]
	repositories.ActivityRepository[models.SectionBranchActivity, models.SectionBranchDeleteActivity]
}

func NewSectionBranchRepository(pst microservice.IPersisterMongo) *SectionBranchRepository {

	insRepo := &SectionBranchRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.SectionBranchDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.SectionBranchInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.SectionBranchItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.SectionBranchActivity, models.SectionBranchDeleteActivity](pst)

	return insRepo
}
