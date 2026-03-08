package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/chequedeposit/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IChequeDepositRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.ChequeDepositDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.ChequeDepositDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.ChequeDepositDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.ChequeDepositInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.ChequeDepositDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]models.ChequeDepositDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.ChequeDepositItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.ChequeDepositDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.ChequeDepositInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.ChequeDepositInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ChequeDepositDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ChequeDepositActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ChequeDepositDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ChequeDepositActivity, error)

	FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.ChequeDepositDoc, error)
}

type ChequeDepositRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.ChequeDepositDoc]
	repositories.SearchRepository[models.ChequeDepositInfo]
	repositories.GuidRepository[models.ChequeDepositItemGuid]
	repositories.ActivityRepository[models.ChequeDepositActivity, models.ChequeDepositDeleteActivity]
}

func NewChequeDepositRepository(pst microservice.IPersisterMongo) *ChequeDepositRepository {

	insRepo := &ChequeDepositRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.ChequeDepositDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.ChequeDepositInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.ChequeDepositItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.ChequeDepositActivity, models.ChequeDepositDeleteActivity](pst)

	return insRepo
}
func (repo ChequeDepositRepository) FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.ChequeDepositDoc, error) {
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

	doc := models.ChequeDepositDoc{}
	err := repo.pst.FindOne(ctx, models.ChequeDepositDoc{}, filters, &doc, &optSort)

	if err != nil {
		return doc, err
	}

	return doc, nil
}
