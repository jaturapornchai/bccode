package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/smlaiproduct/brandproduct/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
)

type IBrandProductRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.BrandProductDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.BrandProductDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.BrandProductDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.BrandProductInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.BrandProductDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]models.BrandProductDoc, error)
	FindByCodes(ctx context.Context, shopID string, codes []string) ([]models.BrandProductInfo, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.BrandProductItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.BrandProductDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.BrandProductInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.BrandProductInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.BrandProductDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.BrandProductActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.BrandProductDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.BrandProductActivity, error)
}

type BrandProductRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.BrandProductDoc]
	repositories.SearchRepository[models.BrandProductInfo]
	repositories.GuidRepository[models.BrandProductItemGuid]
	repositories.ActivityRepository[models.BrandProductActivity, models.BrandProductDeleteActivity]
}

func NewBrandProductRepository(pst microservice.IPersisterMongo) *BrandProductRepository {

	insRepo := &BrandProductRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.BrandProductDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.BrandProductInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.BrandProductItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.BrandProductActivity, models.BrandProductDeleteActivity](pst)

	return insRepo
}

// FindByCodes finds BrandProduct documents by multiple codes
func (repo *BrandProductRepository) FindByCodes(ctx context.Context, shopID string, codes []string) ([]models.BrandProductInfo, error) {
	if len(codes) == 0 {
		return []models.BrandProductInfo{}, nil
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
