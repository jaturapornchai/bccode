package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/chequedisqualified/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IChequeDisqualifiedRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.ChequeDisqualifiedDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.ChequeDisqualifiedDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.ChequeDisqualifiedDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.ChequeDisqualifiedInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.ChequeDisqualifiedDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]models.ChequeDisqualifiedDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.ChequeDisqualifiedItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.ChequeDisqualifiedDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.ChequeDisqualifiedInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.ChequeDisqualifiedInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ChequeDisqualifiedDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ChequeDisqualifiedActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ChequeDisqualifiedDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ChequeDisqualifiedActivity, error)

	FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.ChequeDisqualifiedDoc, error)
}

type ChequeDisqualifiedRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.ChequeDisqualifiedDoc]
	repositories.SearchRepository[models.ChequeDisqualifiedInfo]
	repositories.GuidRepository[models.ChequeDisqualifiedItemGuid]
	repositories.ActivityRepository[models.ChequeDisqualifiedActivity, models.ChequeDisqualifiedDeleteActivity]
}

func NewChequeDisqualifiedRepository(pst microservice.IPersisterMongo) *ChequeDisqualifiedRepository {

	insRepo := &ChequeDisqualifiedRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.ChequeDisqualifiedDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.ChequeDisqualifiedInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.ChequeDisqualifiedItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.ChequeDisqualifiedActivity, models.ChequeDisqualifiedDeleteActivity](pst)

	return insRepo
}
func (repo ChequeDisqualifiedRepository) FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.ChequeDisqualifiedDoc, error) {
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

	doc := models.ChequeDisqualifiedDoc{}
	err := repo.pst.FindOne(ctx, models.ChequeDisqualifiedDoc{}, filters, &doc, &optSort)

	if err != nil {
		return doc, err
	}

	return doc, nil
}
