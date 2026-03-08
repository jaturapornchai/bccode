package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	groupsubtwomodels "smlcloudplatform/internal/smlaiproduct/groupsubtwoproduct/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
)

type IGroupsubtwoProductRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc groupsubtwomodels.GroupsubtwoProductDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []groupsubtwomodels.GroupsubtwoProductDoc) error
	Update(ctx context.Context, shopID string, guid string, doc groupsubtwomodels.GroupsubtwoProductDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]groupsubtwomodels.GroupsubtwoProductInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (groupsubtwomodels.GroupsubtwoProductDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]groupsubtwomodels.GroupsubtwoProductDoc, error)
	FindByCodes(ctx context.Context, shopID string, codes []string) ([]groupsubtwomodels.GroupsubtwoProductInfo, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]groupsubtwomodels.GroupsubtwoProductItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (groupsubtwomodels.GroupsubtwoProductDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]groupsubtwomodels.GroupsubtwoProductInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]groupsubtwomodels.GroupsubtwoProductInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]groupsubtwomodels.GroupsubtwoProductDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]groupsubtwomodels.GroupsubtwoProductActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]groupsubtwomodels.GroupsubtwoProductDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]groupsubtwomodels.GroupsubtwoProductActivity, error)
}

type GroupsubtwoProductRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[groupsubtwomodels.GroupsubtwoProductDoc]
	repositories.SearchRepository[groupsubtwomodels.GroupsubtwoProductInfo]
	repositories.GuidRepository[groupsubtwomodels.GroupsubtwoProductItemGuid]
	repositories.ActivityRepository[groupsubtwomodels.GroupsubtwoProductActivity, groupsubtwomodels.GroupsubtwoProductDeleteActivity]
}

func NewGroupsubtwoProductRepository(pst microservice.IPersisterMongo) *GroupsubtwoProductRepository {

	insRepo := &GroupsubtwoProductRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[groupsubtwomodels.GroupsubtwoProductDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[groupsubtwomodels.GroupsubtwoProductInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[groupsubtwomodels.GroupsubtwoProductItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[groupsubtwomodels.GroupsubtwoProductActivity, groupsubtwomodels.GroupsubtwoProductDeleteActivity](pst)

	return insRepo
}

// FindByCodes finds GroupsubtwoProduct documents by multiple codes
func (repo *GroupsubtwoProductRepository) FindByCodes(ctx context.Context, shopID string, codes []string) ([]groupsubtwomodels.GroupsubtwoProductInfo, error) {
	if len(codes) == 0 {
		return []groupsubtwomodels.GroupsubtwoProductInfo{}, nil
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
