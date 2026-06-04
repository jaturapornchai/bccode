package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/chequepass/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IChequePassRepository interface {
	Count(ctx context.Context, holdingCode string) (int, error)
	Create(ctx context.Context, doc models.ChequePassDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.ChequePassDoc) error
	Update(ctx context.Context, holdingCode string, guid string, doc models.ChequePassDoc) error
	DeleteByGuidfixed(ctx context.Context, holdingCode string, guid string, username string) error
	Delete(ctx context.Context, holdingCode string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable micromodels.Pageable) ([]models.ChequePassInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, holdingCode string, guid string) (models.ChequePassDoc, error)
	FindByGuids(ctx context.Context, holdingCode string, guids []string) ([]models.ChequePassDoc, error)

	FindInItemGuid(ctx context.Context, holdingCode string, columnName string, itemGuidList []string) ([]models.ChequePassItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, holdingCode string, indentityField string, indentityValue interface{}) (models.ChequePassDoc, error)
	FindPageFilter(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.ChequePassInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.ChequePassInfo, int, error)

	FindDeletedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ChequePassDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ChequePassActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ChequePassDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ChequePassActivity, error)

	FindLastDocNo(ctx context.Context, holdingCode string, prefixDocNo string) (models.ChequePassDoc, error)
}

type ChequePassRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.ChequePassDoc]
	repositories.SearchRepository[models.ChequePassInfo]
	repositories.GuidRepository[models.ChequePassItemGuid]
	repositories.ActivityRepository[models.ChequePassActivity, models.ChequePassDeleteActivity]
}

func NewChequePassRepository(pst microservice.IPersisterMongo) *ChequePassRepository {

	insRepo := &ChequePassRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.ChequePassDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.ChequePassInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.ChequePassItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.ChequePassActivity, models.ChequePassDeleteActivity](pst)

	return insRepo
}
func (repo ChequePassRepository) FindLastDocNo(ctx context.Context, holdingCode string, prefixDocNo string) (models.ChequePassDoc, error) {
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

	doc := models.ChequePassDoc{}
	err := repo.pst.FindOne(ctx, models.ChequePassDoc{}, filters, &doc, &optSort)

	if err != nil {
		return doc, err
	}

	return doc, nil
}
