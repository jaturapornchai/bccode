package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/creditcardwithdrawal/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ICreditCardWithdrawalRepository interface {
	Count(ctx context.Context, holdingCode string) (int, error)
	Create(ctx context.Context, doc models.CreditCardWithdrawalDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.CreditCardWithdrawalDoc) error
	Update(ctx context.Context, holdingCode string, guid string, doc models.CreditCardWithdrawalDoc) error
	DeleteByGuidfixed(ctx context.Context, holdingCode string, guid string, username string) error
	Delete(ctx context.Context, holdingCode string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable micromodels.Pageable) ([]models.CreditCardWithdrawalInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, holdingCode string, guid string) (models.CreditCardWithdrawalDoc, error)
	FindByGuids(ctx context.Context, holdingCode string, guids []string) ([]models.CreditCardWithdrawalDoc, error)

	FindInItemGuid(ctx context.Context, holdingCode string, columnName string, itemGuidList []string) ([]models.CreditCardWithdrawalItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, holdingCode string, indentityField string, indentityValue interface{}) (models.CreditCardWithdrawalDoc, error)
	FindPageFilter(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.CreditCardWithdrawalInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.CreditCardWithdrawalInfo, int, error)

	FindDeletedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.CreditCardWithdrawalDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.CreditCardWithdrawalActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.CreditCardWithdrawalDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.CreditCardWithdrawalActivity, error)

	FindLastDocNo(ctx context.Context, holdingCode string, prefixDocNo string) (models.CreditCardWithdrawalDoc, error)
}

type CreditCardWithdrawalRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.CreditCardWithdrawalDoc]
	repositories.SearchRepository[models.CreditCardWithdrawalInfo]
	repositories.GuidRepository[models.CreditCardWithdrawalItemGuid]
	repositories.ActivityRepository[models.CreditCardWithdrawalActivity, models.CreditCardWithdrawalDeleteActivity]
}

func NewCreditCardWithdrawalRepository(pst microservice.IPersisterMongo) *CreditCardWithdrawalRepository {

	insRepo := &CreditCardWithdrawalRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.CreditCardWithdrawalDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.CreditCardWithdrawalInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.CreditCardWithdrawalItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.CreditCardWithdrawalActivity, models.CreditCardWithdrawalDeleteActivity](pst)

	return insRepo
}
func (repo CreditCardWithdrawalRepository) FindLastDocNo(ctx context.Context, holdingCode string, prefixDocNo string) (models.CreditCardWithdrawalDoc, error) {
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

	doc := models.CreditCardWithdrawalDoc{}
	err := repo.pst.FindOne(ctx, models.CreditCardWithdrawalDoc{}, filters, &doc, &optSort)

	if err != nil {
		return doc, err
	}

	return doc, nil
}
