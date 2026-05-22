package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/saledebitnote/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ISaleDebitNoteRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.SaleDebitNoteDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.SaleDebitNoteDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.SaleDebitNoteDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.SaleDebitNoteInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.SaleDebitNoteDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]models.SaleDebitNoteDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.SaleDebitNoteItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.SaleDebitNoteDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.SaleDebitNoteInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.SaleDebitNoteInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.SaleDebitNoteDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.SaleDebitNoteActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.SaleDebitNoteDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.SaleDebitNoteActivity, error)

	FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.SaleDebitNoteDoc, error)
}

type SaleDebitNoteRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.SaleDebitNoteDoc]
	repositories.SearchRepository[models.SaleDebitNoteInfo]
	repositories.GuidRepository[models.SaleDebitNoteItemGuid]
	repositories.ActivityRepository[models.SaleDebitNoteActivity, models.SaleDebitNoteDeleteActivity]
}

func NewSaleDebitNoteRepository(pst microservice.IPersisterMongo) *SaleDebitNoteRepository {

	insRepo := &SaleDebitNoteRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.SaleDebitNoteDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.SaleDebitNoteInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.SaleDebitNoteItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.SaleDebitNoteActivity, models.SaleDebitNoteDeleteActivity](pst)

	return insRepo
}
func (repo SaleDebitNoteRepository) FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.SaleDebitNoteDoc, error) {
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

	doc := models.SaleDebitNoteDoc{}
	err := repo.pst.FindOne(ctx, models.SaleDebitNoteDoc{}, filters, &doc, &optSort)

	if err != nil {
		return doc, err
	}

	return doc, nil
}
