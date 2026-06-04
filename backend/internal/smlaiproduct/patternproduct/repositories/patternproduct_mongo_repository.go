package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	patternmodels "smlcloudplatform/internal/smlaiproduct/patternproduct/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
)

type IPatternProductRepository interface {
	Count(ctx context.Context, holdingCode string) (int, error)
	Create(ctx context.Context, doc patternmodels.PatternProductDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []patternmodels.PatternProductDoc) error
	Update(ctx context.Context, holdingCode string, guid string, doc patternmodels.PatternProductDoc) error
	DeleteByGuidfixed(ctx context.Context, holdingCode string, guid string, username string) error
	Delete(ctx context.Context, holdingCode string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable micromodels.Pageable) ([]patternmodels.PatternProductInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, holdingCode string, guid string) (patternmodels.PatternProductDoc, error)
	FindByGuids(ctx context.Context, holdingCode string, guids []string) ([]patternmodels.PatternProductDoc, error)
	FindByCodes(ctx context.Context, holdingCode string, codes []string) ([]patternmodels.PatternProductInfo, error)

	FindInItemGuid(ctx context.Context, holdingCode string, columnName string, itemGuidList []string) ([]patternmodels.PatternProductItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, holdingCode string, indentityField string, indentityValue interface{}) (patternmodels.PatternProductDoc, error)
	FindPageFilter(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]patternmodels.PatternProductInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]patternmodels.PatternProductInfo, int, error)

	FindDeletedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]patternmodels.PatternProductDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]patternmodels.PatternProductActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]patternmodels.PatternProductDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]patternmodels.PatternProductActivity, error)
}

type PatternProductRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[patternmodels.PatternProductDoc]
	repositories.SearchRepository[patternmodels.PatternProductInfo]
	repositories.GuidRepository[patternmodels.PatternProductItemGuid]
	repositories.ActivityRepository[patternmodels.PatternProductActivity, patternmodels.PatternProductDeleteActivity]
}

func NewPatternProductRepository(pst microservice.IPersisterMongo) *PatternProductRepository {

	insRepo := &PatternProductRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[patternmodels.PatternProductDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[patternmodels.PatternProductInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[patternmodels.PatternProductItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[patternmodels.PatternProductActivity, patternmodels.PatternProductDeleteActivity](pst)

	return insRepo
}

// FindByCodes finds PatternProduct documents by multiple codes
func (repo *PatternProductRepository) FindByCodes(ctx context.Context, holdingCode string, codes []string) ([]patternmodels.PatternProductInfo, error) {
	if len(codes) == 0 {
		return []patternmodels.PatternProductInfo{}, nil
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
