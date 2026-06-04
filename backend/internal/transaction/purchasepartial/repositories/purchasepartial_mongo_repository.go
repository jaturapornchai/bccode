package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/purchasepartial/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IPurchasepartialRepository interface {
	Count(ctx context.Context, holdingCode string) (int, error)
	Create(ctx context.Context, doc models.PurchasepartialDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.PurchasepartialDoc) error
	Update(ctx context.Context, holdingCode string, guid string, doc models.PurchasepartialDoc) error
	DeleteByGuidfixed(ctx context.Context, holdingCode string, guid string, username string) error
	Delete(ctx context.Context, holdingCode string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable micromodels.Pageable) ([]models.PurchasepartialInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, holdingCode string, guid string) (models.PurchasepartialDoc, error)
	FindByGuids(ctx context.Context, holdingCode string, guids []string) ([]models.PurchasepartialDoc, error)

	FindInItemGuid(ctx context.Context, holdingCode string, columnName string, itemGuidList []string) ([]models.PurchasepartialItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, holdingCode string, indentityField string, indentityValue interface{}) (models.PurchasepartialDoc, error)
	FindPageFilter(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.PurchasepartialInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.PurchasepartialInfo, int, error)

	FindDeletedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.PurchasepartialDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.PurchasepartialActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.PurchasepartialDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.PurchasepartialActivity, error)

	FindLastDocNo(ctx context.Context, holdingCode string, prefixDocNo string) (models.PurchasepartialDoc, error)
}

type PurchasepartialRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.PurchasepartialDoc]
	repositories.SearchRepository[models.PurchasepartialInfo]
	repositories.GuidRepository[models.PurchasepartialItemGuid]
	repositories.ActivityRepository[models.PurchasepartialActivity, models.PurchasepartialDeleteActivity]
}

func NewPurchasepartialRepository(pst microservice.IPersisterMongo) *PurchasepartialRepository {

	insRepo := &PurchasepartialRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.PurchasepartialDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.PurchasepartialInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.PurchasepartialItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.PurchasepartialActivity, models.PurchasepartialDeleteActivity](pst)

	return insRepo
}
func (repo PurchasepartialRepository) FindLastDocNo(ctx context.Context, holdingCode string, prefixDocNo string) (models.PurchasepartialDoc, error) {
	filters := bson.M{
		"holding_code": holdingCode,
		"deleted_at": bson.M{
			"$exists": false,
		},
		"docno": bson.M{
			"$regex": "^" + prefixDocNo + ".*$",
		},
	}

	optSort := options.FindOneOptions{}
	optSort.SetSort(bson.M{
		"docno": -1,
	})

	doc := models.PurchasepartialDoc{}
	err := repo.pst.FindOne(ctx, models.PurchasepartialDoc{}, filters, &doc, &optSort)

	if err != nil {
		return doc, err
	}

	return doc, nil
}
