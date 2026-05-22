package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/accrualreceive/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IAccrualreceiveRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.AccrualreceiveDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.AccrualreceiveDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.AccrualreceiveDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.AccrualreceiveInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.AccrualreceiveDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]models.AccrualreceiveDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.AccrualreceiveItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.AccrualreceiveDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.AccrualreceiveInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.AccrualreceiveInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.AccrualreceiveDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.AccrualreceiveActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.AccrualreceiveDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.AccrualreceiveActivity, error)

	FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.AccrualreceiveDoc, error)
}

type AccrualreceiveRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.AccrualreceiveDoc]
	repositories.SearchRepository[models.AccrualreceiveInfo]
	repositories.GuidRepository[models.AccrualreceiveItemGuid]
	repositories.ActivityRepository[models.AccrualreceiveActivity, models.AccrualreceiveDeleteActivity]
}

func NewAccrualreceiveRepository(pst microservice.IPersisterMongo) *AccrualreceiveRepository {

	insRepo := &AccrualreceiveRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.AccrualreceiveDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.AccrualreceiveInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.AccrualreceiveItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.AccrualreceiveActivity, models.AccrualreceiveDeleteActivity](pst)

	return insRepo
}
func (repo AccrualreceiveRepository) FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.AccrualreceiveDoc, error) {
	filters := bson.M{
		"shopid": shopID,
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

	doc := models.AccrualreceiveDoc{}
	err := repo.pst.FindOne(ctx, models.AccrualreceiveDoc{}, filters, &doc, &optSort)

	if err != nil {
		return doc, err
	}

	return doc, nil
}
