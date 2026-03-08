package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	classmodels "smlcloudplatform/internal/smlaiproduct/classproduct/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
)

type IClassProductRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc classmodels.ClassProductDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []classmodels.ClassProductDoc) error
	Update(ctx context.Context, shopID string, guid string, doc classmodels.ClassProductDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]classmodels.ClassProductInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (classmodels.ClassProductDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]classmodels.ClassProductDoc, error)
	FindByCodes(ctx context.Context, shopID string, codes []string) ([]classmodels.ClassProductInfo, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]classmodels.ClassProductItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (classmodels.ClassProductDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]classmodels.ClassProductInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]classmodels.ClassProductInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]classmodels.ClassProductDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]classmodels.ClassProductActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]classmodels.ClassProductDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]classmodels.ClassProductActivity, error)
}

type ClassProductRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[classmodels.ClassProductDoc]
	repositories.SearchRepository[classmodels.ClassProductInfo]
	repositories.GuidRepository[classmodels.ClassProductItemGuid]
	repositories.ActivityRepository[classmodels.ClassProductActivity, classmodels.ClassProductDeleteActivity]
}

func NewClassProductRepository(pst microservice.IPersisterMongo) *ClassProductRepository {

	insRepo := &ClassProductRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[classmodels.ClassProductDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[classmodels.ClassProductInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[classmodels.ClassProductItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[classmodels.ClassProductActivity, classmodels.ClassProductDeleteActivity](pst)

	return insRepo
}

// FindByCodes finds ClassProduct documents by multiple codes
func (repo *ClassProductRepository) FindByCodes(ctx context.Context, shopID string, codes []string) ([]classmodels.ClassProductInfo, error) {
	if len(codes) == 0 {
		return []classmodels.ClassProductInfo{}, nil
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
