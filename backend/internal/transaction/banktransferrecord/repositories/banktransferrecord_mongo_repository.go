package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/banktransferrecord/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IBankTransferRecordRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.BankTransferRecordDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.BankTransferRecordDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.BankTransferRecordDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.BankTransferRecordInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.BankTransferRecordDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]models.BankTransferRecordDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.BankTransferRecordItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.BankTransferRecordDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.BankTransferRecordInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.BankTransferRecordInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.BankTransferRecordDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.BankTransferRecordActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.BankTransferRecordDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.BankTransferRecordActivity, error)

	FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.BankTransferRecordDoc, error)
}

type BankTransferRecordRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.BankTransferRecordDoc]
	repositories.SearchRepository[models.BankTransferRecordInfo]
	repositories.GuidRepository[models.BankTransferRecordItemGuid]
	repositories.ActivityRepository[models.BankTransferRecordActivity, models.BankTransferRecordDeleteActivity]
}

func NewBankTransferRecordRepository(pst microservice.IPersisterMongo) *BankTransferRecordRepository {

	insRepo := &BankTransferRecordRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.BankTransferRecordDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.BankTransferRecordInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.BankTransferRecordItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.BankTransferRecordActivity, models.BankTransferRecordDeleteActivity](pst)

	return insRepo
}
func (repo BankTransferRecordRepository) FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.BankTransferRecordDoc, error) {
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

	doc := models.BankTransferRecordDoc{}
	err := repo.pst.FindOne(ctx, models.BankTransferRecordDoc{}, filters, &doc, &optSort)

	if err != nil {
		return doc, err
	}

	return doc, nil
}
