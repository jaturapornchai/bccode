package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/pickandpack/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IPickandpackRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.PickandpackDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.PickandpackDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.PickandpackDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.PickandpackInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.PickandpackDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]models.PickandpackDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.PickandpackItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.PickandpackDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.PickandpackInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.PickandpackInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.PickandpackDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.PickandpackActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.PickandpackDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.PickandpackActivity, error)

	FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.PickandpackDoc, error)
	AggregateWarehouseDashboard(ctx context.Context, shopID string, whcodes []string, locationcodes []string, fromDate, toDate string) ([]models.PickandpackWarehouseDashboard, error)
}

type PickandpackRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.PickandpackDoc]
	repositories.SearchRepository[models.PickandpackInfo]
	repositories.GuidRepository[models.PickandpackItemGuid]
	repositories.ActivityRepository[models.PickandpackActivity, models.PickandpackDeleteActivity]
}

func NewPickandpackRepository(pst microservice.IPersisterMongo) *PickandpackRepository {

	insRepo := &PickandpackRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.PickandpackDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.PickandpackInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.PickandpackItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.PickandpackActivity, models.PickandpackDeleteActivity](pst)

	return insRepo
}
func (repo PickandpackRepository) FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.PickandpackDoc, error) {
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

	doc := models.PickandpackDoc{}
	err := repo.pst.FindOne(ctx, models.PickandpackDoc{}, filters, &doc, &optSort)

	if err != nil {
		return doc, err
	}

	return doc, nil
}

// AggregateWarehouseDashboard aggregates warehouse dashboard data
func (repo PickandpackRepository) AggregateWarehouseDashboard(ctx context.Context, shopID string, whcodes []string, locationcodes []string, fromDate, toDate string) ([]models.PickandpackWarehouseDashboard, error) {
	// Build filter query
	filterQuery := bson.M{
		"shopid":   shopID,
		"iscancel": false,
	}

	// Add date range filter if provided
	if fromDate != "" && toDate != "" {
		fromDateTime, err1 := time.Parse("2006-01-02", fromDate)
		toDateTime, err2 := time.Parse("2006-01-02", toDate)

		if err1 == nil && err2 == nil {
			// Set time to beginning of fromDate and end of toDate to include full days
			fromDateTimeUTC := time.Date(fromDateTime.Year(), fromDateTime.Month(), fromDateTime.Day(), 0, 0, 0, 0, time.UTC)
			toDateTimeUTC := time.Date(toDateTime.Year(), toDateTime.Month(), toDateTime.Day(), 23, 59, 59, 999999999, time.UTC)

			filterQuery["docdatetime"] = bson.M{
				"$gte": fromDateTimeUTC,
				"$lte": toDateTimeUTC,
			}
		}
	}

	// Add warehouse codes filter if provided
	if len(whcodes) > 0 {
		filterQuery["whcode"] = bson.M{"$in": whcodes}
	}

	// Add location codes filter if provided
	if len(locationcodes) > 0 {
		filterQuery["locationcode"] = bson.M{"$in": locationcodes}
	}

	// Aggregation pipeline to group by warehouse and location
	pipeline := []interface{}{
		bson.M{"$match": filterQuery},
		bson.M{
			"$group": bson.M{
				"_id": bson.M{
					"whcode":       "$whcode",
					"locationcode": "$locationcode",
					"packstatus":   "$packstatus",
				},
				"count":         bson.M{"$sum": 1},
				"whnames":       bson.M{"$first": "$whnames"},
				"locationnames": bson.M{"$first": "$locationnames"},
			},
		},
		bson.M{
			"$group": bson.M{
				"_id": bson.M{
					"whcode":       "$_id.whcode",
					"locationcode": "$_id.locationcode",
				},
				"whnames":       bson.M{"$first": "$whnames"},
				"locationnames": bson.M{"$first": "$locationnames"},
				"statusCounts": bson.M{
					"$push": bson.M{
						"packstatus": "$_id.packstatus",
						"count":      "$count",
					},
				},
				"totalCount": bson.M{"$sum": "$count"},
			},
		},
		bson.M{
			"$sort": bson.M{
				"_id.whcode":       1,
				"_id.locationcode": 1,
			},
		},
	}

	var results []models.PickandpackWarehouseDashboard
	err := repo.pst.Aggregate(ctx, &models.PickandpackWarehouseDashboard{}, pipeline, &results)
	if err != nil {
		return nil, err
	}

	return results, nil
}
