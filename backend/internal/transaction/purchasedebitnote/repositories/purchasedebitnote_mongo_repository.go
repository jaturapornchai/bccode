package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/purchasedebitnote/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IPurchaseDebitNoteRepository interface {
	Count(ctx context.Context, holdingCode string) (int, error)
	Create(ctx context.Context, doc models.PurchaseDebitNoteDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.PurchaseDebitNoteDoc) error
	Update(ctx context.Context, holdingCode string, guid string, doc models.PurchaseDebitNoteDoc) error
	DeleteByGuidfixed(ctx context.Context, holdingCode string, guid string, username string) error
	Delete(ctx context.Context, holdingCode string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable micromodels.Pageable) ([]models.PurchaseDebitNoteInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, holdingCode string, guid string) (models.PurchaseDebitNoteDoc, error)
	FindByGuids(ctx context.Context, holdingCode string, guids []string) ([]models.PurchaseDebitNoteDoc, error)

	FindInItemGuid(ctx context.Context, holdingCode string, columnName string, itemGuidList []string) ([]models.PurchaseDebitNoteItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, holdingCode string, indentityField string, indentityValue interface{}) (models.PurchaseDebitNoteDoc, error)
	FindPageFilter(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.PurchaseDebitNoteInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.PurchaseDebitNoteInfo, int, error)

	FindDeletedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.PurchaseDebitNoteDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.PurchaseDebitNoteActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.PurchaseDebitNoteDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.PurchaseDebitNoteActivity, error)

	FindLastDocNo(ctx context.Context, holdingCode string, prefixDocNo string) (models.PurchaseDebitNoteDoc, error)
}

type PurchaseDebitNoteRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.PurchaseDebitNoteDoc]
	repositories.SearchRepository[models.PurchaseDebitNoteInfo]
	repositories.GuidRepository[models.PurchaseDebitNoteItemGuid]
	repositories.ActivityRepository[models.PurchaseDebitNoteActivity, models.PurchaseDebitNoteDeleteActivity]
}

func NewPurchaseDebitNoteRepository(pst microservice.IPersisterMongo) *PurchaseDebitNoteRepository {

	insRepo := &PurchaseDebitNoteRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.PurchaseDebitNoteDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.PurchaseDebitNoteInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.PurchaseDebitNoteItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.PurchaseDebitNoteActivity, models.PurchaseDebitNoteDeleteActivity](pst)

	return insRepo
}
func (repo PurchaseDebitNoteRepository) FindLastDocNo(ctx context.Context, holdingCode string, prefixDocNo string) (models.PurchaseDebitNoteDoc, error) {
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

	doc := models.PurchaseDebitNoteDoc{}
	err := repo.pst.FindOne(ctx, models.PurchaseDebitNoteDoc{}, filters, &doc, &optSort)

	if err != nil {
		return doc, err
	}

	return doc, nil
}
