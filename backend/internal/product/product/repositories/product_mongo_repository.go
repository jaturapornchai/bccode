package repositories

import (
	"context"
	"fmt"
	"smlcloudplatform/internal/product/product/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type IProductRepository interface {
	Count(ctx context.Context, holdingCode string) (int, error)
	Create(ctx context.Context, doc models.ProductDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.ProductDoc) error
	Update(ctx context.Context, holdingCode string, guid string, doc models.ProductDoc) error
	DeleteByGuidfixed(ctx context.Context, holdingCode string, guid string, username string) error
	Delete(ctx context.Context, holdingCode string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable micromodels.Pageable) ([]models.ProductInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, holdingCode string, guid string) (models.ProductDoc, error)

	FindInItemGuid(ctx context.Context, holdingCode string, columnName string, itemGuidList []string) ([]models.ProductItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, holdingCode string, indentityField string, indentityValue interface{}) (models.ProductDoc, error)
	FindPageFilter(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.ProductInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.ProductInfo, int, error)

	FindDeletedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ProductDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ProductActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ProductDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ProductActivity, error)

	FindOneByCode(ctx context.Context, holdingCode, code string) (models.ProductDoc, error)
	FindFilter(ctx context.Context, holdingCode string, filters map[string]interface{}) ([]models.ProductDoc, error)
	EnsureIndexes(ctx context.Context) error
}

type ProductRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.ProductDoc]
	repositories.SearchRepository[models.ProductInfo]
	repositories.GuidRepository[models.ProductItemGuid]
	repositories.ActivityRepository[models.ProductActivity, models.ProductDeleteActivity]
}

func NewProductRepository(pst microservice.IPersisterMongo) *ProductRepository {

	insRepo := &ProductRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.ProductDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.ProductInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.ProductItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.ProductActivity, models.ProductDeleteActivity](pst)

	return insRepo
}

func (repo ProductRepository) EnsureIndexes(ctx context.Context) error {
	if _, err := repo.pst.CreateIndex(
		ctx,
		models.ProductDoc{},
		"uniq_products_holdingcode_guidfixed",
		bson.D{{Key: "holdingcode", Value: 1}, {Key: "guidfixed", Value: 1}},
	); err != nil {
		return fmt.Errorf("ensure product guidfixed index: %w", err)
	}

	_, err := repo.pst.CreatePartialUniqueIndex(
		ctx,
		models.ProductDoc{},
		"uniq_products_active_holdingcode_code",
		bson.D{
			{Key: "holdingcode", Value: 1},
			{Key: "code", Value: 1},
		},
		bson.M{
			"deletedat": nil,
			"code":      bson.M{"$type": "string", "$gt": ""},
		},
	)
	if err != nil {
		return fmt.Errorf("ensure active product code index: %w", err)
	}
	return nil
}

func (repo ProductRepository) FindOneByCode(ctx context.Context, holdingCode string, code string) (models.ProductDoc, error) {
	doc := models.ProductDoc{}
	err := repo.pst.FindOne(ctx,
		models.ProductDoc{},
		bson.M{
			"holdingcode": holdingCode,
			"deletedat":   bson.M{"$exists": false},
			"code":        code,
		}, &doc)

	if err != nil {
		return doc, err
	}

	return doc, nil
}
