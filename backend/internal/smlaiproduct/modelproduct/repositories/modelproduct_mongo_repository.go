package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	modelmodels "smlcloudplatform/internal/smlaiproduct/modelproduct/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
)

type IModelProductRepository interface {
	Count(ctx context.Context, holdingCode string) (int, error)
	Create(ctx context.Context, doc modelmodels.ModelProductDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []modelmodels.ModelProductDoc) error
	Update(ctx context.Context, holdingCode string, guid string, doc modelmodels.ModelProductDoc) error
	DeleteByGuidfixed(ctx context.Context, holdingCode string, guid string, username string) error
	Delete(ctx context.Context, holdingCode string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable micromodels.Pageable) ([]modelmodels.ModelProductInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, holdingCode string, guid string) (modelmodels.ModelProductDoc, error)
	FindByGuids(ctx context.Context, holdingCode string, guids []string) ([]modelmodels.ModelProductDoc, error)
	FindByCodes(ctx context.Context, holdingCode string, codes []string) ([]modelmodels.ModelProductInfo, error)

	FindInItemGuid(ctx context.Context, holdingCode string, columnName string, itemGuidList []string) ([]modelmodels.ModelProductItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, holdingCode string, indentityField string, indentityValue interface{}) (modelmodels.ModelProductDoc, error)
	FindPageFilter(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]modelmodels.ModelProductInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]modelmodels.ModelProductInfo, int, error)

	FindDeletedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]modelmodels.ModelProductDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]modelmodels.ModelProductActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]modelmodels.ModelProductDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]modelmodels.ModelProductActivity, error)
}

type ModelProductRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[modelmodels.ModelProductDoc]
	repositories.SearchRepository[modelmodels.ModelProductInfo]
	repositories.GuidRepository[modelmodels.ModelProductItemGuid]
	repositories.ActivityRepository[modelmodels.ModelProductActivity, modelmodels.ModelProductDeleteActivity]
}

func NewModelProductRepository(pst microservice.IPersisterMongo) *ModelProductRepository {

	insRepo := &ModelProductRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[modelmodels.ModelProductDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[modelmodels.ModelProductInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[modelmodels.ModelProductItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[modelmodels.ModelProductActivity, modelmodels.ModelProductDeleteActivity](pst)

	return insRepo
}

// FindByCodes finds ModelProduct documents by multiple codes
func (repo *ModelProductRepository) FindByCodes(ctx context.Context, holdingCode string, codes []string) ([]modelmodels.ModelProductInfo, error) {
	if len(codes) == 0 {
		return []modelmodels.ModelProductInfo{}, nil
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
