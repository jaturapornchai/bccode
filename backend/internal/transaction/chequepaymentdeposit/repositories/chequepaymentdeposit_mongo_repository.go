package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/chequepaymentdeposit/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IChequePaymentDepositRepository interface {
	Count(ctx context.Context, holdingCode string) (int, error)
	Create(ctx context.Context, doc models.ChequePaymentDepositDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.ChequePaymentDepositDoc) error
	Update(ctx context.Context, holdingCode string, guid string, doc models.ChequePaymentDepositDoc) error
	DeleteByGuidfixed(ctx context.Context, holdingCode string, guid string, username string) error
	Delete(ctx context.Context, holdingCode string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable micromodels.Pageable) ([]models.ChequePaymentDepositInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, holdingCode string, guid string) (models.ChequePaymentDepositDoc, error)
	FindByGuids(ctx context.Context, holdingCode string, guids []string) ([]models.ChequePaymentDepositDoc, error)

	FindInItemGuid(ctx context.Context, holdingCode string, columnName string, itemGuidList []string) ([]models.ChequePaymentDepositItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, holdingCode string, indentityField string, indentityValue interface{}) (models.ChequePaymentDepositDoc, error)
	FindPageFilter(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.ChequePaymentDepositInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.ChequePaymentDepositInfo, int, error)

	FindDeletedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ChequePaymentDepositDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ChequePaymentDepositActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ChequePaymentDepositDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ChequePaymentDepositActivity, error)

	FindLastDocNo(ctx context.Context, holdingCode string, prefixDocNo string) (models.ChequePaymentDepositDoc, error)
}

type ChequePaymentDepositRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.ChequePaymentDepositDoc]
	repositories.SearchRepository[models.ChequePaymentDepositInfo]
	repositories.GuidRepository[models.ChequePaymentDepositItemGuid]
	repositories.ActivityRepository[models.ChequePaymentDepositActivity, models.ChequePaymentDepositDeleteActivity]
}

func NewChequePaymentDepositRepository(pst microservice.IPersisterMongo) *ChequePaymentDepositRepository {

	insRepo := &ChequePaymentDepositRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.ChequePaymentDepositDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.ChequePaymentDepositInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.ChequePaymentDepositItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.ChequePaymentDepositActivity, models.ChequePaymentDepositDeleteActivity](pst)

	return insRepo
}
func (repo ChequePaymentDepositRepository) FindLastDocNo(ctx context.Context, holdingCode string, prefixDocNo string) (models.ChequePaymentDepositDoc, error) {
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

	doc := models.ChequePaymentDepositDoc{}
	err := repo.pst.FindOne(ctx, models.ChequePaymentDepositDoc{}, filters, &doc, &optSort)

	if err != nil {
		return doc, err
	}

	return doc, nil
}
