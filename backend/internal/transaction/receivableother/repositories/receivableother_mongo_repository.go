package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/receivableother/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IReceivableOtherRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.ReceivableOtherDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.ReceivableOtherDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.ReceivableOtherDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.ReceivableOtherInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.ReceivableOtherDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.ReceivableOtherItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.ReceivableOtherDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.ReceivableOtherInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.ReceivableOtherInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ReceivableOtherDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ReceivableOtherActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ReceivableOtherDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ReceivableOtherActivity, error)

	FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.ReceivableOtherDoc, error)
}

type ReceivableOtherRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.ReceivableOtherDoc]
	repositories.SearchRepository[models.ReceivableOtherInfo]
	repositories.GuidRepository[models.ReceivableOtherItemGuid]
	repositories.ActivityRepository[models.ReceivableOtherActivity, models.ReceivableOtherDeleteActivity]
}

func NewReceivableOtherRepository(pst microservice.IPersisterMongo) *ReceivableOtherRepository {

	insRepo := &ReceivableOtherRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.ReceivableOtherDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.ReceivableOtherInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.ReceivableOtherItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.ReceivableOtherActivity, models.ReceivableOtherDeleteActivity](pst)

	return insRepo
}

func (repo ReceivableOtherRepository) FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.ReceivableOtherDoc, error) {
	filters := bson.M{
		"shopid": shopID,
		"deletedat": bson.M{
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

	doc := models.ReceivableOtherDoc{}
	err := repo.pst.FindOne(ctx, models.ReceivableOtherDoc{}, filters, &doc, &optSort)

	if err != nil {
		return doc, err
	}

	return doc, nil
}
