package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/utils/mogoutil"
	"smlcloudplatform/internal/utils/search"
	"smlcloudplatform/internal/warehouse/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type IWarehouseRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.WarehouseDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.WarehouseDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.WarehouseDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.WarehouseInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.WarehouseDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.WarehouseItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.WarehouseDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.WarehouseInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.WarehouseInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.WarehouseDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.WarehouseActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.WarehouseDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.WarehouseActivity, error)

	FindLocationPage(ctx context.Context, shopID string, pageable micromodels.Pageable) ([]models.LocationInfo, mongopagination.PaginationData, error)
	FindShelfPage(ctx context.Context, shopID string, pageable micromodels.Pageable) ([]models.ShelfInfo, mongopagination.PaginationData, error)
	FindWarehouseByLocation(ctx context.Context, shopID, warehouseCode, locationCode string) (models.WarehouseDoc, error)
	FindWarehouseByShelf(ctx context.Context, shopID, warehouseCode, locationCode, shelfCode string) (models.WarehouseDoc, error)

	Transaction(ctx context.Context, queryFunc func(ctx context.Context) error) error
}

type WarehouseRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.WarehouseDoc]
	repositories.SearchRepository[models.WarehouseInfo]
	repositories.GuidRepository[models.WarehouseItemGuid]
	repositories.ActivityRepository[models.WarehouseActivity, models.WarehouseDeleteActivity]
}

func NewWarehouseRepository(pst microservice.IPersisterMongo) *WarehouseRepository {

	insRepo := &WarehouseRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.WarehouseDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.WarehouseInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.WarehouseItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.WarehouseActivity, models.WarehouseDeleteActivity](pst)

	return insRepo
}

func (repo WarehouseRepository) Transaction(ctx context.Context, queryFunc func(ctx context.Context) error) error {
	return repo.pst.Transaction(ctx, queryFunc)
}

func (repo WarehouseRepository) FindLocationPage(ctx context.Context, shopID string, pageable micromodels.Pageable) ([]models.LocationInfo, mongopagination.PaginationData, error) {

	criteria := []interface{}{}

	mainQuery := bson.M{
		"$match": bson.M{
			"shopid":    shopID,
			"deletedat": bson.M{"$exists": false},
		},
	}
	criteria = append(criteria, mainQuery)

	searchFilterQuery := search.CreateTextFilter([]string{"location.code", "location.code.names.name"}, pageable.Query)

	if len(searchFilterQuery) > 0 {
		searchQuery := bson.M{"$match": bson.M{"$or": searchFilterQuery}}
		criteria = append(criteria, searchQuery)
	}

	unwindQuery := bson.M{"$unwind": "$location"}
	criteria = append(criteria, unwindQuery)

	projectQuery := bson.M{"$project": bson.M{
		"guidfixed":      "$guidfixed",
		"warehousecode":  "$code",
		"warehousenames": "$names",
		"locationcode":   "$location.code",
		"locationnames":  "$location.names",
		"shelf":          1,
	}}
	criteria = append(criteria, projectQuery)

	aggData, err := repo.pst.AggregatePage(ctx, models.LocationInfo{}, pageable, criteria...)

	if err != nil {
		return []models.LocationInfo{}, mongopagination.PaginationData{}, err
	}

	docList, err := mogoutil.AggregatePageDecode[models.LocationInfo](aggData)

	if err != nil {
		return []models.LocationInfo{}, mongopagination.PaginationData{}, err
	}

	return docList, aggData.Pagination, nil
}

func (repo WarehouseRepository) FindShelfPage(ctx context.Context, shopID string, pageable micromodels.Pageable) ([]models.ShelfInfo, mongopagination.PaginationData, error) {

	criteria := []interface{}{}

	mainQuery := bson.M{
		"$match": bson.M{
			"shopid":    shopID,
			"deletedat": bson.M{"$exists": false},
		},
	}
	criteria = append(criteria, mainQuery)

	searchFilterQuery := search.CreateTextFilter([]string{"location.shelf.code", "location.shelf.name"}, pageable.Query)

	if len(searchFilterQuery) > 0 {
		searchQuery := bson.M{"$match": bson.M{"$or": searchFilterQuery}}
		criteria = append(criteria, searchQuery)
	}

	unwindQueryLevel1 := bson.M{"$unwind": "$location"}
	criteria = append(criteria, unwindQueryLevel1)

	unwindQueryLevel2 := bson.M{"$unwind": "$location.shelf"}
	criteria = append(criteria, unwindQueryLevel2)

	projectQuery := bson.M{"$project": bson.M{
		"guidfixed":      "$guidfixed",
		"warehousecode":  "$code",
		"warehousenames": "$names",
		"locationcode":   "$location.code",
		"locationnames":  "$location.names",
		"shelfcode":      "$location.shelf.code",
		"shelfname":      "$location.shelf.name",
	}}

	criteria = append(criteria, projectQuery)

	aggData, err := repo.pst.AggregatePage(ctx, models.ShelfInfo{}, pageable, criteria...)

	if err != nil {
		return []models.ShelfInfo{}, mongopagination.PaginationData{}, err
	}

	docList, err := mogoutil.AggregatePageDecode[models.ShelfInfo](aggData)

	if err != nil {
		return []models.ShelfInfo{}, mongopagination.PaginationData{}, err
	}

	return docList, aggData.Pagination, nil
}

func (repo WarehouseRepository) FindWarehouseByLocation(ctx context.Context, shopID, warehouseCode, locationCode string) (models.WarehouseDoc, error) {

	filters := bson.M{
		"shopid":        shopID,
		"code":          warehouseCode,
		"location.code": locationCode,
	}

	doc := models.WarehouseDoc{}
	err := repo.pst.FindOne(ctx, models.WarehouseDoc{}, filters, &doc)

	if err != nil {
		return models.WarehouseDoc{}, err
	}

	return doc, nil
}

func (repo WarehouseRepository) FindWarehouseByShelf(ctx context.Context, shopID, warehouseCode, locationCode, shelfCode string) (models.WarehouseDoc, error) {

	filters := bson.M{
		"shopid":              shopID,
		"code":                warehouseCode,
		"location.code":       locationCode,
		"location.shelf.code": shelfCode,
	}

	doc := models.WarehouseDoc{}
	err := repo.pst.FindOne(ctx, models.WarehouseDoc{}, filters, &doc)

	if err != nil {
		return models.WarehouseDoc{}, err
	}

	return doc, nil
}
