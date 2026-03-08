package repositories

import (
	"context"
	"smlcloudplatform/internal/product/productcategory/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type IProductCategoryRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.ProductCategoryDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.ProductCategoryDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.ProductCategoryDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.ProductCategoryInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.ProductCategoryDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.ProductCategoryItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.ProductCategoryDoc, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, selectFields map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ProductCategoryInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageable micromodels.Pageable) ([]models.ProductCategoryDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageable micromodels.Pageable) ([]models.ProductCategoryActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ProductCategoryDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ProductCategoryActivity, error)

	UpdateCodeList(ctx context.Context, shopID string, codeXSort models.CodeXSort) error
}

type ProductCategoryRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.ProductCategoryDoc]
	repositories.SearchRepository[models.ProductCategoryInfo]
	repositories.GuidRepository[models.ProductCategoryItemGuid]
	repositories.ActivityRepository[models.ProductCategoryActivity, models.ProductCategoryDeleteActivity]
}

func NewProductCategoryRepository(pst microservice.IPersisterMongo) *ProductCategoryRepository {

	insRepo := &ProductCategoryRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.ProductCategoryDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.ProductCategoryInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.ProductCategoryItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.ProductCategoryActivity, models.ProductCategoryDeleteActivity](pst)

	return insRepo
}

func (repo ProductCategoryRepository) UpdateCodeList(ctx context.Context, shopID string, codeXSort models.CodeXSort) error {

	filters := bson.M{
		"shopid":           shopID,
		"codelist.barcode": codeXSort.Barcode,
	}

	doc := bson.M{
		"$set": bson.M{
			"codelist.$.code":      codeXSort.Code,
			"codelist.$.names":     codeXSort.Names,
			"codelist.$.unitcode":  codeXSort.UnitCode,
			"codelist.$.unitnames": codeXSort.UnitNames,
		},
	}

	return repo.pst.Update(
		ctx,
		models.ProductCategoryDoc{},
		filters,
		doc,
	)
}
