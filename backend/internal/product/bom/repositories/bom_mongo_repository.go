package repositories

import (
	"context"
	"smlcloudplatform/internal/product/bom/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type IBomRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.ProductBarcodeBOMViewDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.ProductBarcodeBOMViewDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.ProductBarcodeBOMViewDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.ProductBarcodeBOMViewInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.ProductBarcodeBOMViewDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.ProductBarcodeBOMViewGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.ProductBarcodeBOMViewDoc, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, selectFields map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ProductBarcodeBOMViewInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageable micromodels.Pageable) ([]models.ProductBarcodeBOMViewDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageable micromodels.Pageable) ([]models.ProductBarcodeBOMViewActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ProductBarcodeBOMViewDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ProductBarcodeBOMViewActivity, error)

	FindUseBOMByBarcode(ctx context.Context, shopID string, barcode string) (models.ProductBarcodeBOMViewDoc, error)
	ClearUseBOMByBarcode(ctx context.Context, shopID string, barcode string) error
}

type BomRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.ProductBarcodeBOMViewDoc]
	repositories.SearchRepository[models.ProductBarcodeBOMViewInfo]
	repositories.GuidRepository[models.ProductBarcodeBOMViewGuid]
	repositories.ActivityRepository[models.ProductBarcodeBOMViewActivity, models.ProductBarcodeBOMViewDeleteActivity]
}

func NewBomRepository(pst microservice.IPersisterMongo) *BomRepository {

	insRepo := &BomRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.ProductBarcodeBOMViewDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.ProductBarcodeBOMViewInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.ProductBarcodeBOMViewGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.ProductBarcodeBOMViewActivity, models.ProductBarcodeBOMViewDeleteActivity](pst)

	return insRepo
}

func (repo BomRepository) FindUseBOMByBarcode(ctx context.Context, shopID string, barcode string) (models.ProductBarcodeBOMViewDoc, error) {

	filters := bson.M{
		"shopid":       shopID,
		"iscurrentuse": true,
		"barcode":      barcode,
		"deletedat":    bson.M{"$exists": false},
	}

	doc := models.ProductBarcodeBOMViewDoc{}

	err := repo.pst.FindOne(
		ctx,
		models.ProductBarcodeBOMViewDoc{},
		filters,
		&doc,
	)

	if err != nil {
		return models.ProductBarcodeBOMViewDoc{}, err
	}

	return doc, nil
}

func (repo BomRepository) ClearUseBOMByBarcode(ctx context.Context, shopID string, barcode string) error {

	filters := bson.M{
		"shopid":       shopID,
		"barcode":      barcode,
		"iscurrentuse": true,
		"deletedat":    bson.M{"$exists": false},
	}

	err := repo.pst.Update(ctx, models.ProductBarcodeBOMViewDoc{}, filters, bson.M{"$set": bson.M{"iscurrentuse": false}})
	if err != nil {
		return err
	}

	return nil

}
