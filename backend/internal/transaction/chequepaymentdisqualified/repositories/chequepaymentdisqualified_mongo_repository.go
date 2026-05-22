package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/chequepaymentdisqualified/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IChequePaymentDisqualifiedRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.ChequePaymentDisqualifiedDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.ChequePaymentDisqualifiedDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.ChequePaymentDisqualifiedDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.ChequePaymentDisqualifiedInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.ChequePaymentDisqualifiedDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]models.ChequePaymentDisqualifiedDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.ChequePaymentDisqualifiedItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.ChequePaymentDisqualifiedDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.ChequePaymentDisqualifiedInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.ChequePaymentDisqualifiedInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ChequePaymentDisqualifiedDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ChequePaymentDisqualifiedActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ChequePaymentDisqualifiedDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ChequePaymentDisqualifiedActivity, error)

	FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.ChequePaymentDisqualifiedDoc, error)
}

type ChequePaymentDisqualifiedRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.ChequePaymentDisqualifiedDoc]
	repositories.SearchRepository[models.ChequePaymentDisqualifiedInfo]
	repositories.GuidRepository[models.ChequePaymentDisqualifiedItemGuid]
	repositories.ActivityRepository[models.ChequePaymentDisqualifiedActivity, models.ChequePaymentDisqualifiedDeleteActivity]
}

func NewChequePaymentDisqualifiedRepository(pst microservice.IPersisterMongo) *ChequePaymentDisqualifiedRepository {

	insRepo := &ChequePaymentDisqualifiedRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.ChequePaymentDisqualifiedDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.ChequePaymentDisqualifiedInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.ChequePaymentDisqualifiedItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.ChequePaymentDisqualifiedActivity, models.ChequePaymentDisqualifiedDeleteActivity](pst)

	return insRepo
}
func (repo ChequePaymentDisqualifiedRepository) FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.ChequePaymentDisqualifiedDoc, error) {
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

	doc := models.ChequePaymentDisqualifiedDoc{}
	err := repo.pst.FindOne(ctx, models.ChequePaymentDisqualifiedDoc{}, filters, &doc, &optSort)

	if err != nil {
		return doc, err
	}

	return doc, nil
}
