package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/billingnote/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IBillingNoteRepository interface {
	Count(ctx context.Context, holdingCode string) (int, error)
	Create(ctx context.Context, doc models.BillingNoteDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.BillingNoteDoc) error
	Update(ctx context.Context, holdingCode string, guid string, doc models.BillingNoteDoc) error
	DeleteByGuidfixed(ctx context.Context, holdingCode string, guid string, username string) error
	Delete(ctx context.Context, holdingCode string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable micromodels.Pageable) ([]models.BillingNoteInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, holdingCode string, guid string) (models.BillingNoteDoc, error)

	FindInItemGuid(ctx context.Context, holdingCode string, columnName string, itemGuidList []string) ([]models.BillingNoteItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, holdingCode string, indentityField string, indentityValue interface{}) (models.BillingNoteDoc, error)
	FindPageFilter(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.BillingNoteInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.BillingNoteInfo, int, error)

	FindDeletedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.BillingNoteDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.BillingNoteActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.BillingNoteDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.BillingNoteActivity, error)

	FindLastDocNo(ctx context.Context, holdingCode string, prefixDocNo string) (models.BillingNoteDoc, error)
}

type BillingNoteRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.BillingNoteDoc]
	repositories.SearchRepository[models.BillingNoteInfo]
	repositories.GuidRepository[models.BillingNoteItemGuid]
	repositories.ActivityRepository[models.BillingNoteActivity, models.BillingNoteDeleteActivity]
}

func NewBillingNoteRepository(pst microservice.IPersisterMongo) *BillingNoteRepository {

	insRepo := &BillingNoteRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.BillingNoteDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.BillingNoteInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.BillingNoteItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.BillingNoteActivity, models.BillingNoteDeleteActivity](pst)

	return insRepo
}

func (repo BillingNoteRepository) FindLastDocNo(ctx context.Context, holdingCode string, prefixDocNo string) (models.BillingNoteDoc, error) {

	filter := bson.M{
		"holdingcode": holdingCode,
		"docno":       bson.M{"$regex": "^" + prefixDocNo},
	}

	findOptions := options.FindOne()
	findOptions.SetSort(bson.D{{Key: "docno", Value: -1}})

	doc := models.BillingNoteDoc{}
	err := repo.pst.FindOne(ctx, models.BillingNoteDoc{}, filter, &doc, findOptions)

	if err != nil {
		return models.BillingNoteDoc{}, err
	}

	return doc, nil
}
