package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/chequepaymentchange/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IChequePaymentChangeRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.ChequePaymentChangeDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.ChequePaymentChangeDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.ChequePaymentChangeDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.ChequePaymentChangeInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.ChequePaymentChangeDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]models.ChequePaymentChangeDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.ChequePaymentChangeItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.ChequePaymentChangeDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.ChequePaymentChangeInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.ChequePaymentChangeInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ChequePaymentChangeDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ChequePaymentChangeActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ChequePaymentChangeDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ChequePaymentChangeActivity, error)

	FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.ChequePaymentChangeDoc, error)
}

type ChequePaymentChangeRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.ChequePaymentChangeDoc]
	repositories.SearchRepository[models.ChequePaymentChangeInfo]
	repositories.GuidRepository[models.ChequePaymentChangeItemGuid]
	repositories.ActivityRepository[models.ChequePaymentChangeActivity, models.ChequePaymentChangeDeleteActivity]
}

func NewChequePaymentChangeRepository(pst microservice.IPersisterMongo) *ChequePaymentChangeRepository {

	insRepo := &ChequePaymentChangeRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.ChequePaymentChangeDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.ChequePaymentChangeInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.ChequePaymentChangeItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.ChequePaymentChangeActivity, models.ChequePaymentChangeDeleteActivity](pst)

	return insRepo
}
func (repo ChequePaymentChangeRepository) FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.ChequePaymentChangeDoc, error) {
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

	doc := models.ChequePaymentChangeDoc{}
	err := repo.pst.FindOne(ctx, models.ChequePaymentChangeDoc{}, filters, &doc, &optSort)

	if err != nil {
		return doc, err
	}

	return doc, nil
}
