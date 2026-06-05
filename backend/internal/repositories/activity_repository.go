package repositories

import (
	"context"
	"smlcloudplatform/pkg/microservice"
	"strings"
	"time"

	micromodels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IActivityRepository[TCU any, TDEL any] interface {
	// InitialActivityRepository(pst microservice.IPersisterMongo)
	FindDeletedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageable micromodels.Pageable) ([]TDEL, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageable micromodels.Pageable) ([]TCU, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageableStep micromodels.PageableStep) ([]TDEL, error)
	FindCreatedOrUpdatedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageableStep micromodels.PageableStep) ([]TCU, error)
}
type ActivityRepository[TCU any, TDEL any] struct {
	pst microservice.IPersisterMongo
}

func NewActivityRepository[TCU any, TDEL any](pst microservice.IPersisterMongo) ActivityRepository[TCU, TDEL] {
	return ActivityRepository[TCU, TDEL]{
		pst: pst,
	}
}

func (repo *ActivityRepository[TCU, TDEL]) InitialActivityRepository(pst microservice.IPersisterMongo) {
	repo.pst = pst
}

func (repo ActivityRepository[TCU, TDEL]) FindDeletedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageable micromodels.Pageable) ([]TDEL, mongopagination.PaginationData, error) {

	filterQueries := bson.M{
		"holdingcode": holdingCode,
		"deletedat":   bson.M{"$gte": lastUpdatedDate},
	}

	extraFilterQueries := repo.generateExtraFilters(extraFilters)
	if len(extraFilterQueries) > 0 {
		filterQueries["$and"] = extraFilterQueries
	}

	docList := []TDEL{}
	pagination, err := repo.pst.FindPage(ctx, new(TDEL), filterQueries, pageable, &docList)

	if err != nil {
		return []TDEL{}, mongopagination.PaginationData{}, err
	}

	return docList, pagination, nil
}

func (repo ActivityRepository[TCU, TDEL]) FindCreatedOrUpdatedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageable micromodels.Pageable) ([]TCU, mongopagination.PaginationData, error) {

	filterQueries := bson.M{
		"holdingcode": holdingCode,
		"deletedat":   bson.M{"$not": bson.M{"$gte": lastUpdatedDate}},
		"$or": []interface{}{
			bson.M{"createdat": bson.M{"$gte": lastUpdatedDate}},
			bson.M{"updatedat": bson.M{"$gte": lastUpdatedDate}},
		},
	}

	extraFilterQueries := repo.generateExtraFilters(extraFilters)
	if len(extraFilterQueries) > 0 {
		filterQueries["$and"] = extraFilterQueries
	}

	docList := []TCU{}
	pagination, err := repo.pst.FindPage(ctx, new(TCU), filterQueries, pageable, &docList)

	if err != nil {
		return []TCU{}, mongopagination.PaginationData{}, err
	}

	return docList, pagination, nil
}

func (repo ActivityRepository[TCU, TDEL]) FindDeletedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageableStep micromodels.PageableStep) ([]TDEL, error) {

	docList := []TDEL{}

	// Handle multiple holding Codes separated by comma
	var holdingCodeFilter interface{}
	if strings.Contains(holdingCode, ",") {
		holdingCodes := strings.Split(holdingCode, ",")
		for i, id := range holdingCodes {
			holdingCodes[i] = strings.TrimSpace(id)
		}
		holdingCodeFilter = bson.M{"$in": holdingCodes}
	} else {
		holdingCodeFilter = holdingCode
	}

	filterQueries := bson.M{
		"holdingcode": holdingCodeFilter,
		"deletedat":   bson.M{"$gte": lastUpdatedDate},
	}

	extraFilterQueries := repo.generateExtraFilters(extraFilters)
	if len(extraFilterQueries) > 0 {
		filterQueries["$and"] = extraFilterQueries
	}

	tempOptions := &options.FindOptions{}
	tempOptions.SetSkip(int64(pageableStep.Skip))
	tempOptions.SetLimit(int64(pageableStep.Limit))

	err := repo.pst.Find(ctx, new(TDEL), filterQueries, &docList, tempOptions)

	if err != nil {
		return []TDEL{}, err
	}

	return docList, nil
}

func (repo ActivityRepository[TCU, TDEL]) FindCreatedOrUpdatedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageableStep micromodels.PageableStep) ([]TCU, error) {

	docList := []TCU{}

	// Handle multiple holding Codes separated by comma
	var holdingCodeFilter interface{}
	if strings.Contains(holdingCode, ",") {
		holdingCodes := strings.Split(holdingCode, ",")
		for i, id := range holdingCodes {
			holdingCodes[i] = strings.TrimSpace(id)
		}
		holdingCodeFilter = bson.M{"$in": holdingCodes}
	} else {
		holdingCodeFilter = holdingCode
	}

	filterQueries := bson.M{
		"holdingcode": holdingCodeFilter,
		"deletedat":   bson.M{"$not": bson.M{"$gte": lastUpdatedDate}},
		"$or": []interface{}{
			bson.M{"createdat": bson.M{"$gte": lastUpdatedDate}},
			bson.M{"updatedat": bson.M{"$gte": lastUpdatedDate}},
		},
	}

	extraFilterQueries := repo.generateExtraFilters(extraFilters)
	if len(extraFilterQueries) > 0 {
		filterQueries["$and"] = extraFilterQueries
	}

	tempOptions := &options.FindOptions{}
	tempOptions.SetSkip(int64(pageableStep.Skip))
	tempOptions.SetLimit(int64(pageableStep.Limit))

	err := repo.pst.Find(ctx, new(TCU), filterQueries, &docList, tempOptions)

	if err != nil {
		return []TCU{}, err
	}

	return docList, nil
}

func (repo ActivityRepository[TCU, TDEL]) generateExtraFilters(extraFilters map[string]interface{}) []interface{} {
	if len(extraFilters) > 0 {
		tempExtraFilters := []interface{}{}
		for key, value := range extraFilters {
			tempExtraFilters = append(tempExtraFilters, bson.M{key: value})
		}

		return tempExtraFilters
	}

	return []interface{}{}
}
