package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/paidadvance/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IPaidAdvanceRepository interface {
	Count(ctx context.Context, holdingCode string) (int, error)
	Create(ctx context.Context, doc models.PaidAdvanceDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.PaidAdvanceDoc) error
	Update(ctx context.Context, holdingCode string, guid string, doc models.PaidAdvanceDoc) error
	DeleteByGuidfixed(ctx context.Context, holdingCode string, guid string, username string) error
	Delete(ctx context.Context, holdingCode string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable micromodels.Pageable) ([]models.PaidAdvanceInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, holdingCode string, guid string) (models.PaidAdvanceDoc, error)
	FindByGuids(ctx context.Context, holdingCode string, guids []string) ([]models.PaidAdvanceDoc, error)

	FindInItemGuid(ctx context.Context, holdingCode string, columnName string, itemGuidList []string) ([]models.PaidAdvanceItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, holdingCode string, indentityField string, indentityValue interface{}) (models.PaidAdvanceDoc, error)
	FindPageFilter(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.PaidAdvanceInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.PaidAdvanceInfo, int, error)

	FindDeletedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.PaidAdvanceDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.PaidAdvanceActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.PaidAdvanceDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.PaidAdvanceActivity, error)

	FindLastDocNo(ctx context.Context, holdingCode string, prefixDocNo string) (models.PaidAdvanceDoc, error)
}

type PaidAdvanceRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.PaidAdvanceDoc]
	repositories.SearchRepository[models.PaidAdvanceInfo]
	repositories.GuidRepository[models.PaidAdvanceItemGuid]
	repositories.ActivityRepository[models.PaidAdvanceActivity, models.PaidAdvanceDeleteActivity]
}

func NewPaidAdvanceRepository(pst microservice.IPersisterMongo) *PaidAdvanceRepository {

	insRepo := &PaidAdvanceRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.PaidAdvanceDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.PaidAdvanceInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.PaidAdvanceItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.PaidAdvanceActivity, models.PaidAdvanceDeleteActivity](pst)

	return insRepo
}
func (repo PaidAdvanceRepository) FindLastDocNo(ctx context.Context, holdingCode string, prefixDocNo string) (models.PaidAdvanceDoc, error) {
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

	doc := models.PaidAdvanceDoc{}
	err := repo.pst.FindOne(ctx, models.PaidAdvanceDoc{}, filters, &doc, &optSort)

	if err != nil {
		return doc, err
	}

	return doc, nil
}
