package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	groupsubonemodels "smlcloudplatform/internal/smlaiproduct/groupsuboneproduct/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
)

type IGroupsuboneProductRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc groupsubonemodels.GroupsuboneProductDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []groupsubonemodels.GroupsuboneProductDoc) error
	Update(ctx context.Context, shopID string, guid string, doc groupsubonemodels.GroupsuboneProductDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]groupsubonemodels.GroupsuboneProductInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (groupsubonemodels.GroupsuboneProductDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]groupsubonemodels.GroupsuboneProductDoc, error)
	FindByCodes(ctx context.Context, shopID string, codes []string) ([]groupsubonemodels.GroupsuboneProductInfo, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]groupsubonemodels.GroupsuboneProductItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (groupsubonemodels.GroupsuboneProductDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]groupsubonemodels.GroupsuboneProductInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]groupsubonemodels.GroupsuboneProductInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]groupsubonemodels.GroupsuboneProductDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]groupsubonemodels.GroupsuboneProductActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]groupsubonemodels.GroupsuboneProductDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]groupsubonemodels.GroupsuboneProductActivity, error)
}

type GroupsuboneProductRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[groupsubonemodels.GroupsuboneProductDoc]
	repositories.SearchRepository[groupsubonemodels.GroupsuboneProductInfo]
	repositories.GuidRepository[groupsubonemodels.GroupsuboneProductItemGuid]
	repositories.ActivityRepository[groupsubonemodels.GroupsuboneProductActivity, groupsubonemodels.GroupsuboneProductDeleteActivity]
}

func NewGroupsuboneProductRepository(pst microservice.IPersisterMongo) *GroupsuboneProductRepository {

	insRepo := &GroupsuboneProductRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[groupsubonemodels.GroupsuboneProductDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[groupsubonemodels.GroupsuboneProductInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[groupsubonemodels.GroupsuboneProductItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[groupsubonemodels.GroupsuboneProductActivity, groupsubonemodels.GroupsuboneProductDeleteActivity](pst)

	return insRepo
}

// FindByCodes finds GroupsuboneProduct documents by multiple codes
func (repo *GroupsuboneProductRepository) FindByCodes(ctx context.Context, shopID string, codes []string) ([]groupsubonemodels.GroupsuboneProductInfo, error) {
	if len(codes) == 0 {
		return []groupsubonemodels.GroupsuboneProductInfo{}, nil
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
