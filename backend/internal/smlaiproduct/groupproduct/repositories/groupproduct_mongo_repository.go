package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	groupmodels "smlcloudplatform/internal/smlaiproduct/groupproduct/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
)

type IGroupProductRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc groupmodels.GroupProductDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []groupmodels.GroupProductDoc) error
	Update(ctx context.Context, shopID string, guid string, doc groupmodels.GroupProductDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]groupmodels.GroupProductInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (groupmodels.GroupProductDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]groupmodels.GroupProductDoc, error)
	FindByCodes(ctx context.Context, shopID string, codes []string) ([]groupmodels.GroupProductInfo, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]groupmodels.GroupProductItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (groupmodels.GroupProductDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]groupmodels.GroupProductInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]groupmodels.GroupProductInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]groupmodels.GroupProductDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]groupmodels.GroupProductActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]groupmodels.GroupProductDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]groupmodels.GroupProductActivity, error)
}

type GroupProductRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[groupmodels.GroupProductDoc]
	repositories.SearchRepository[groupmodels.GroupProductInfo]
	repositories.GuidRepository[groupmodels.GroupProductItemGuid]
	repositories.ActivityRepository[groupmodels.GroupProductActivity, groupmodels.GroupProductDeleteActivity]
}

func NewGroupProductRepository(pst microservice.IPersisterMongo) *GroupProductRepository {

	insRepo := &GroupProductRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[groupmodels.GroupProductDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[groupmodels.GroupProductInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[groupmodels.GroupProductItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[groupmodels.GroupProductActivity, groupmodels.GroupProductDeleteActivity](pst)

	return insRepo
}

// FindByCodes finds GroupProduct documents by multiple codes
func (repo *GroupProductRepository) FindByCodes(ctx context.Context, shopID string, codes []string) ([]groupmodels.GroupProductInfo, error) {
	if len(codes) == 0 {
		return []groupmodels.GroupProductInfo{}, nil
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
