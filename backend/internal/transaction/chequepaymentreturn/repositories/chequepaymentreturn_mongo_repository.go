package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/chequepaymentreturn/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IChequePaymentReturnRepository interface {
	Count(ctx context.Context, holdingCode string) (int, error)
	Create(ctx context.Context, doc models.ChequePaymentReturnDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.ChequePaymentReturnDoc) error
	Update(ctx context.Context, holdingCode string, guid string, doc models.ChequePaymentReturnDoc) error
	DeleteByGuidfixed(ctx context.Context, holdingCode string, guid string, username string) error
	Delete(ctx context.Context, holdingCode string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable micromodels.Pageable) ([]models.ChequePaymentReturnInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, holdingCode string, guid string) (models.ChequePaymentReturnDoc, error)
	FindByGuids(ctx context.Context, holdingCode string, guids []string) ([]models.ChequePaymentReturnDoc, error)

	FindInItemGuid(ctx context.Context, holdingCode string, columnName string, itemGuidList []string) ([]models.ChequePaymentReturnItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, holdingCode string, indentityField string, indentityValue interface{}) (models.ChequePaymentReturnDoc, error)
	FindPageFilter(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.ChequePaymentReturnInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.ChequePaymentReturnInfo, int, error)

	FindDeletedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ChequePaymentReturnDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ChequePaymentReturnActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ChequePaymentReturnDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ChequePaymentReturnActivity, error)

	FindLastDocNo(ctx context.Context, holdingCode string, prefixDocNo string) (models.ChequePaymentReturnDoc, error)
}

type ChequePaymentReturnRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.ChequePaymentReturnDoc]
	repositories.SearchRepository[models.ChequePaymentReturnInfo]
	repositories.GuidRepository[models.ChequePaymentReturnItemGuid]
	repositories.ActivityRepository[models.ChequePaymentReturnActivity, models.ChequePaymentReturnDeleteActivity]
}

func NewChequePaymentReturnRepository(pst microservice.IPersisterMongo) *ChequePaymentReturnRepository {

	insRepo := &ChequePaymentReturnRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.ChequePaymentReturnDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.ChequePaymentReturnInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.ChequePaymentReturnItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.ChequePaymentReturnActivity, models.ChequePaymentReturnDeleteActivity](pst)

	return insRepo
}
func (repo ChequePaymentReturnRepository) FindLastDocNo(ctx context.Context, holdingCode string, prefixDocNo string) (models.ChequePaymentReturnDoc, error) {
	filters := bson.M{
		"holdingcode": holdingCode,
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

	doc := models.ChequePaymentReturnDoc{}
	err := repo.pst.FindOne(ctx, models.ChequePaymentReturnDoc{}, filters, &doc, &optSort)

	if err != nil {
		return doc, err
	}

	return doc, nil
}
