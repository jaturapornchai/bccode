package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/saleinvoice/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ISaleInvoiceRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.SaleInvoiceDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.SaleInvoiceDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.SaleInvoiceDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.SaleInvoiceInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.SaleInvoiceDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]models.SaleInvoiceDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.SaleInvoiceItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.SaleInvoiceDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.SaleInvoiceInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.SaleInvoiceInfo, int, error)

	Find(ctx context.Context, shopID string, searchInFields []string, q string) ([]models.SaleInvoiceInfo, error)
	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.SaleInvoiceDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.SaleInvoiceActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.SaleInvoiceDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.SaleInvoiceActivity, error)

	FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.SaleInvoiceDoc, error)
	FindLastPOSDocNo(ctx context.Context, shopID string, posID string, maxDocNo string) (string, error)
}

type SaleInvoiceRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.SaleInvoiceDoc]
	repositories.SearchRepository[models.SaleInvoiceInfo]
	repositories.GuidRepository[models.SaleInvoiceItemGuid]
	repositories.ActivityRepository[models.SaleInvoiceActivity, models.SaleInvoiceDeleteActivity]
}

func NewSaleInvoiceRepository(pst microservice.IPersisterMongo) *SaleInvoiceRepository {

	insRepo := &SaleInvoiceRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.SaleInvoiceDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.SaleInvoiceInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.SaleInvoiceItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.SaleInvoiceActivity, models.SaleInvoiceDeleteActivity](pst)

	return insRepo
}

func (repo SaleInvoiceRepository) FindLastPOSDocNo(ctx context.Context, shopID string, posID string, maxDocNo string) (string, error) {

	opts := options.FindOneOptions{}
	opts.SetSort(bson.M{"docno": -1})

	filters := bson.M{
		"shopid": shopID,
		"posid":  posID,
		"docno": bson.M{
			"$lte": maxDocNo,
		},
	}

	doc := models.SaleInvoiceDoc{}
	err := repo.pst.FindOne(ctx, models.SaleInvoiceDoc{}, filters, &doc, &opts)

	if err != nil {
		return "", err
	}

	return doc.DocNo, nil
}

func (repo SaleInvoiceRepository) FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.SaleInvoiceDoc, error) {
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

	doc := models.SaleInvoiceDoc{}
	err := repo.pst.FindOne(ctx, models.SaleInvoiceDoc{}, filters, &doc, &optSort)
	if err != nil {
		return doc, err
	}
	return doc, nil
}
