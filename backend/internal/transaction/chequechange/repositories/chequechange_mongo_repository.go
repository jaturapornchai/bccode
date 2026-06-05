package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/chequechange/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IChequeChangeRepository interface {
	Count(ctx context.Context, holdingCode string) (int, error)
	Create(ctx context.Context, doc models.ChequeChangeDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.ChequeChangeDoc) error
	Update(ctx context.Context, holdingCode string, guid string, doc models.ChequeChangeDoc) error
	DeleteByGuidfixed(ctx context.Context, holdingCode string, guid string, username string) error
	Delete(ctx context.Context, holdingCode string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable micromodels.Pageable) ([]models.ChequeChangeInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, holdingCode string, guid string) (models.ChequeChangeDoc, error)
	FindByGuids(ctx context.Context, holdingCode string, guids []string) ([]models.ChequeChangeDoc, error)

	FindInItemGuid(ctx context.Context, holdingCode string, columnName string, itemGuidList []string) ([]models.ChequeChangeItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, holdingCode string, indentityField string, indentityValue interface{}) (models.ChequeChangeDoc, error)
	FindPageFilter(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.ChequeChangeInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.ChequeChangeInfo, int, error)

	FindDeletedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ChequeChangeDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ChequeChangeActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ChequeChangeDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ChequeChangeActivity, error)

	FindLastDocNo(ctx context.Context, holdingCode string, prefixDocNo string) (models.ChequeChangeDoc, error)
}

type ChequeChangeRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.ChequeChangeDoc]
	repositories.SearchRepository[models.ChequeChangeInfo]
	repositories.GuidRepository[models.ChequeChangeItemGuid]
	repositories.ActivityRepository[models.ChequeChangeActivity, models.ChequeChangeDeleteActivity]
}

func NewChequeChangeRepository(pst microservice.IPersisterMongo) *ChequeChangeRepository {

	insRepo := &ChequeChangeRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.ChequeChangeDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.ChequeChangeInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.ChequeChangeItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.ChequeChangeActivity, models.ChequeChangeDeleteActivity](pst)

	return insRepo
}
func (repo ChequeChangeRepository) FindLastDocNo(ctx context.Context, holdingCode string, prefixDocNo string) (models.ChequeChangeDoc, error) {
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

	doc := models.ChequeChangeDoc{}
	err := repo.pst.FindOne(ctx, models.ChequeChangeDoc{}, filters, &doc, &optSort)

	if err != nil {
		return doc, err
	}

	return doc, nil
}
