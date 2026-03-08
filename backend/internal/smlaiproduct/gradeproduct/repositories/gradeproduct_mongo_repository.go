package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	grademodels "smlcloudplatform/internal/smlaiproduct/gradeproduct/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
)

type IGradeProductRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc grademodels.GradeProductDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []grademodels.GradeProductDoc) error
	Update(ctx context.Context, shopID string, guid string, doc grademodels.GradeProductDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]grademodels.GradeProductInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (grademodels.GradeProductDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]grademodels.GradeProductDoc, error)
	FindByCodes(ctx context.Context, shopID string, codes []string) ([]grademodels.GradeProductInfo, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]grademodels.GradeProductItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (grademodels.GradeProductDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]grademodels.GradeProductInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]grademodels.GradeProductInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]grademodels.GradeProductDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]grademodels.GradeProductActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]grademodels.GradeProductDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]grademodels.GradeProductActivity, error)
}

type GradeProductRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[grademodels.GradeProductDoc]
	repositories.SearchRepository[grademodels.GradeProductInfo]
	repositories.GuidRepository[grademodels.GradeProductItemGuid]
	repositories.ActivityRepository[grademodels.GradeProductActivity, grademodels.GradeProductDeleteActivity]
}

func NewGradeProductRepository(pst microservice.IPersisterMongo) *GradeProductRepository {

	insRepo := &GradeProductRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[grademodels.GradeProductDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[grademodels.GradeProductInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[grademodels.GradeProductItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[grademodels.GradeProductActivity, grademodels.GradeProductDeleteActivity](pst)

	return insRepo
}

// FindByCodes finds GradeProduct documents by multiple codes
func (repo *GradeProductRepository) FindByCodes(ctx context.Context, shopID string, codes []string) ([]grademodels.GradeProductInfo, error) {
	if len(codes) == 0 {
		return []grademodels.GradeProductInfo{}, nil
	}

	filters := map[string]interface{}{
		"code": map[string]interface{}{
			"$in": codes,
		},
	}

	docs, _, err := repo.FindPageFilter(ctx, shopID, filters, []string{}, micromodels.Pageable{
		Page:  1,
		Limit: len(codes),
	})

	return docs, err
}
