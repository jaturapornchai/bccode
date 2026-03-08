package kitchen

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/restaurant/kitchen/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type IKitchenRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, category models.KitchenDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.KitchenDoc) error
	Update(ctx context.Context, shopID string, guid string, category models.KitchenDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, authUsername string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.KitchenInfo, mongopagination.PaginationData, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.KitchenInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.KitchenDoc, error)
	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.KitchenItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, columnName string, filters interface{}) (models.KitchenDoc, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.KitchenInfo, int, error)
	Find(ctx context.Context, shopID string, filters map[string]interface{}) ([]models.KitchenInfo, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageable micromodels.Pageable) ([]models.KitchenDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageable micromodels.Pageable) ([]models.KitchenActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.KitchenDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.KitchenActivity, error)
}

type KitchenRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.KitchenDoc]
	repositories.SearchRepository[models.KitchenInfo]
	repositories.GuidRepository[models.KitchenItemGuid]
	repositories.ActivityRepository[models.KitchenActivity, models.KitchenDeleteActivity]
}

func NewKitchenRepository(pst microservice.IPersisterMongo) KitchenRepository {

	insRepo := KitchenRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.KitchenDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.KitchenInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.KitchenItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.KitchenActivity, models.KitchenDeleteActivity](pst)

	return insRepo
}

func (repo KitchenRepository) Find(ctx context.Context, shopID string, filters map[string]interface{}) ([]models.KitchenInfo, error) {
	result := []models.KitchenInfo{}

	matchFilterList := []interface{}{}

	for key, value := range filters {
		matchFilterList = append(matchFilterList, bson.M{key: value})
	}

	queryFilters := bson.M{
		"shopid":    shopID,
		"deletedat": bson.M{"$exists": false},
	}

	if len(matchFilterList) > 0 {
		queryFilters["$and"] = matchFilterList
	}

	err := repo.pst.Find(ctx, models.KitchenInfo{}, queryFilters, &result)

	if err != nil {
		return nil, err
	}

	return result, nil
}
