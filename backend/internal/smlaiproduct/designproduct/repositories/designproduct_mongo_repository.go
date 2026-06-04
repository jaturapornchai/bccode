package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	designmodels "smlcloudplatform/internal/smlaiproduct/designproduct/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
)

type IDesignProductRepository interface {
	Count(ctx context.Context, holdingCode string) (int, error)
	Create(ctx context.Context, doc designmodels.DesignProductDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []designmodels.DesignProductDoc) error
	Update(ctx context.Context, holdingCode string, guid string, doc designmodels.DesignProductDoc) error
	DeleteByGuidfixed(ctx context.Context, holdingCode string, guid string, username string) error
	Delete(ctx context.Context, holdingCode string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable micromodels.Pageable) ([]designmodels.DesignProductInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, holdingCode string, guid string) (designmodels.DesignProductDoc, error)
	FindByGuids(ctx context.Context, holdingCode string, guids []string) ([]designmodels.DesignProductDoc, error)
	FindByCodes(ctx context.Context, holdingCode string, codes []string) ([]designmodels.DesignProductInfo, error)

	FindInItemGuid(ctx context.Context, holdingCode string, columnName string, itemGuidList []string) ([]designmodels.DesignProductItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, holdingCode string, indentityField string, indentityValue interface{}) (designmodels.DesignProductDoc, error)
	FindPageFilter(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]designmodels.DesignProductInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]designmodels.DesignProductInfo, int, error)

	FindDeletedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]designmodels.DesignProductDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]designmodels.DesignProductActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]designmodels.DesignProductDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]designmodels.DesignProductActivity, error)
}

type DesignProductRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[designmodels.DesignProductDoc]
	repositories.SearchRepository[designmodels.DesignProductInfo]
	repositories.GuidRepository[designmodels.DesignProductItemGuid]
	repositories.ActivityRepository[designmodels.DesignProductActivity, designmodels.DesignProductDeleteActivity]
}

func NewDesignProductRepository(pst microservice.IPersisterMongo) *DesignProductRepository {

	insRepo := &DesignProductRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[designmodels.DesignProductDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[designmodels.DesignProductInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[designmodels.DesignProductItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[designmodels.DesignProductActivity, designmodels.DesignProductDeleteActivity](pst)

	return insRepo
}

// FindByCodes finds DesignProduct documents by multiple codes
func (repo *DesignProductRepository) FindByCodes(ctx context.Context, holdingCode string, codes []string) ([]designmodels.DesignProductInfo, error) {
	if len(codes) == 0 {
		return []designmodels.DesignProductInfo{}, nil
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
