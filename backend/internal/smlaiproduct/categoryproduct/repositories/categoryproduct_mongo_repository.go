package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	categorymodels "smlcloudplatform/internal/smlaiproduct/categoryproduct/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
)

type ICategoryProductRepository interface {
	Count(ctx context.Context, holdingCode string) (int, error)
	Create(ctx context.Context, doc categorymodels.CategoryProductDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []categorymodels.CategoryProductDoc) error
	Update(ctx context.Context, holdingCode string, guid string, doc categorymodels.CategoryProductDoc) error
	DeleteByGuidfixed(ctx context.Context, holdingCode string, guid string, username string) error
	Delete(ctx context.Context, holdingCode string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable micromodels.Pageable) ([]categorymodels.CategoryProductInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, holdingCode string, guid string) (categorymodels.CategoryProductDoc, error)
	FindByGuids(ctx context.Context, holdingCode string, guids []string) ([]categorymodels.CategoryProductDoc, error)
	FindByCodes(ctx context.Context, holdingCode string, codes []string) ([]categorymodels.CategoryProductInfo, error)

	FindInItemGuid(ctx context.Context, holdingCode string, columnName string, itemGuidList []string) ([]categorymodels.CategoryProductItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, holdingCode string, indentityField string, indentityValue interface{}) (categorymodels.CategoryProductDoc, error)
	FindPageFilter(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]categorymodels.CategoryProductInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]categorymodels.CategoryProductInfo, int, error)

	FindDeletedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]categorymodels.CategoryProductDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]categorymodels.CategoryProductActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]categorymodels.CategoryProductDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]categorymodels.CategoryProductActivity, error)
}

type CategoryProductRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[categorymodels.CategoryProductDoc]
	repositories.SearchRepository[categorymodels.CategoryProductInfo]
	repositories.GuidRepository[categorymodels.CategoryProductItemGuid]
	repositories.ActivityRepository[categorymodels.CategoryProductActivity, categorymodels.CategoryProductDeleteActivity]
}

func NewCategoryProductRepository(pst microservice.IPersisterMongo) *CategoryProductRepository {

	insRepo := &CategoryProductRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[categorymodels.CategoryProductDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[categorymodels.CategoryProductInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[categorymodels.CategoryProductItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[categorymodels.CategoryProductActivity, categorymodels.CategoryProductDeleteActivity](pst)

	return insRepo
}

// FindByCodes finds CategoryProduct documents by multiple codes
func (repo *CategoryProductRepository) FindByCodes(ctx context.Context, holdingCode string, codes []string) ([]categorymodels.CategoryProductInfo, error) {
	if len(codes) == 0 {
		return []categorymodels.CategoryProductInfo{}, nil
	}

	filters := map[string]interface{}{
		"code": map[string]interface{}{
			"$in": codes,
		},
	}

	docs, _, err := repo.FindPageFilter(ctx, holdingCode, filters, []string{}, micromodels.Pageable{
		Page:  1,
		Limit: len(codes),
	})

	return docs, err
}
