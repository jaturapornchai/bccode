package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/saleinvoicereturn/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ISaleInvoiceReturnRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.SaleInvoiceReturnDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.SaleInvoiceReturnDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.SaleInvoiceReturnDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.SaleInvoiceReturnInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.SaleInvoiceReturnDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]models.SaleInvoiceReturnDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.SaleInvoiceReturnItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.SaleInvoiceReturnDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.SaleInvoiceReturnInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.SaleInvoiceReturnInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.SaleInvoiceReturnDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.SaleInvoiceReturnActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.SaleInvoiceReturnDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.SaleInvoiceReturnActivity, error)

	FindLastPOSDocNo(ctx context.Context, shopID string, posID string, maxDocNo string) (string, error)
	FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.SaleInvoiceReturnDoc, error)
}

type SaleInvoiceReturnRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.SaleInvoiceReturnDoc]
	repositories.SearchRepository[models.SaleInvoiceReturnInfo]
	repositories.GuidRepository[models.SaleInvoiceReturnItemGuid]
	repositories.ActivityRepository[models.SaleInvoiceReturnActivity, models.SaleInvoiceReturnDeleteActivity]
}

func NewSaleInvoiceReturnRepository(pst microservice.IPersisterMongo) *SaleInvoiceReturnRepository {

	insRepo := &SaleInvoiceReturnRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.SaleInvoiceReturnDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.SaleInvoiceReturnInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.SaleInvoiceReturnItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.SaleInvoiceReturnActivity, models.SaleInvoiceReturnDeleteActivity](pst)

	return insRepo
}

func (repo SaleInvoiceReturnRepository) FindLastPOSDocNo(ctx context.Context, shopID string, posID string, maxDocNo string) (string, error) {

	opts := options.FindOneOptions{}
	opts.SetSort(bson.M{"docno": -1})

	filters := bson.M{
		"shopid": shopID,
		"posid":  posID,
		"docno": bson.M{
			"$lte": maxDocNo,
		},
	}

	doc := models.SaleInvoiceReturnDoc{}
	err := repo.pst.FindOne(ctx, models.SaleInvoiceReturnDoc{}, filters, &doc, &opts)

	if err != nil {
		return "", err
	}

	return doc.DocNo, nil
}

func (repo SaleInvoiceReturnRepository) FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.SaleInvoiceReturnDoc, error) {
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

	doc := models.SaleInvoiceReturnDoc{}
	err := repo.pst.FindOne(ctx, models.SaleInvoiceReturnDoc{}, filters, &doc, &optSort)

	if err != nil {
		return doc, err
	}

	return doc, nil
}
