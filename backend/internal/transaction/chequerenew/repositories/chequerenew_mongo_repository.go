package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/chequerenew/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IChequeRenewRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.ChequeRenewDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.ChequeRenewDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.ChequeRenewDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.ChequeRenewInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.ChequeRenewDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]models.ChequeRenewDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.ChequeRenewItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.ChequeRenewDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.ChequeRenewInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.ChequeRenewInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ChequeRenewDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ChequeRenewActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ChequeRenewDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ChequeRenewActivity, error)

	FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.ChequeRenewDoc, error)
}

type ChequeRenewRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.ChequeRenewDoc]
	repositories.SearchRepository[models.ChequeRenewInfo]
	repositories.GuidRepository[models.ChequeRenewItemGuid]
	repositories.ActivityRepository[models.ChequeRenewActivity, models.ChequeRenewDeleteActivity]
}

func NewChequeRenewRepository(pst microservice.IPersisterMongo) *ChequeRenewRepository {

	insRepo := &ChequeRenewRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.ChequeRenewDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.ChequeRenewInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.ChequeRenewItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.ChequeRenewActivity, models.ChequeRenewDeleteActivity](pst)

	return insRepo
}
func (repo ChequeRenewRepository) FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.ChequeRenewDoc, error) {
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

	doc := models.ChequeRenewDoc{}
	err := repo.pst.FindOne(ctx, models.ChequeRenewDoc{}, filters, &doc, &optSort)

	if err != nil {
		return doc, err
	}

	return doc, nil
}
