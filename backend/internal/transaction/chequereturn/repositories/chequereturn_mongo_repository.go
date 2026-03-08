package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/chequereturn/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IChequeReturnRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.ChequeReturnDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.ChequeReturnDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.ChequeReturnDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.ChequeReturnInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.ChequeReturnDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]models.ChequeReturnDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.ChequeReturnItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.ChequeReturnDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.ChequeReturnInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.ChequeReturnInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ChequeReturnDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ChequeReturnActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ChequeReturnDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ChequeReturnActivity, error)

	FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.ChequeReturnDoc, error)
}

type ChequeReturnRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.ChequeReturnDoc]
	repositories.SearchRepository[models.ChequeReturnInfo]
	repositories.GuidRepository[models.ChequeReturnItemGuid]
	repositories.ActivityRepository[models.ChequeReturnActivity, models.ChequeReturnDeleteActivity]
}

func NewChequeReturnRepository(pst microservice.IPersisterMongo) *ChequeReturnRepository {

	insRepo := &ChequeReturnRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.ChequeReturnDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.ChequeReturnInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.ChequeReturnItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.ChequeReturnActivity, models.ChequeReturnDeleteActivity](pst)

	return insRepo
}
func (repo ChequeReturnRepository) FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.ChequeReturnDoc, error) {
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

	doc := models.ChequeReturnDoc{}
	err := repo.pst.FindOne(ctx, models.ChequeReturnDoc{}, filters, &doc, &optSort)

	if err != nil {
		return doc, err
	}

	return doc, nil
}
