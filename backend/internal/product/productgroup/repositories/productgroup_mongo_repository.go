package repositories

import (
	"context"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/product/productgroup/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type IProductGroupRepository interface {
	Count(ctx context.Context, holdingCode string) (int, error)
	Create(ctx context.Context, doc models.ProductGroupDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.ProductGroupDoc) error
	Update(ctx context.Context, holdingCode string, guid string, doc models.ProductGroupDoc) error
	DeleteByGuidfixed(ctx context.Context, holdingCode string, guid string, username string) error
	Delete(ctx context.Context, holdingCode string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable micromodels.Pageable) ([]models.ProductGroupInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, holdingCode string, guid string) (models.ProductGroupDoc, error)
	FindByGuids(ctx context.Context, holdingCode string, guids []string) ([]models.ProductGroupDoc, error)

	FindInItemGuid(ctx context.Context, holdingCode string, columnName string, itemGuidList []string) ([]models.ProductGroupItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, holdingCode string, indentityField string, indentityValue interface{}) (models.ProductGroupDoc, error)
	FindPageFilter(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.ProductGroupInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.ProductGroupInfo, int, error)

	FindDeletedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageable micromodels.Pageable) ([]models.ProductGroupDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageable micromodels.Pageable) ([]models.ProductGroupActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ProductGroupDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ProductGroupActivity, error)

	UpdateXSorts(ctx context.Context, holdingCode string, guid string, xsorts []common.XSort, username string, updatedAt time.Time) error
}

type ProductGroupRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.ProductGroupDoc]
	repositories.SearchRepository[models.ProductGroupInfo]
	repositories.GuidRepository[models.ProductGroupItemGuid]
	repositories.ActivityRepository[models.ProductGroupActivity, models.ProductGroupDeleteActivity]
}

func NewProductGroupRepository(pst microservice.IPersisterMongo) *ProductGroupRepository {

	insRepo := &ProductGroupRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.ProductGroupDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.ProductGroupInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.ProductGroupItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.ProductGroupActivity, models.ProductGroupDeleteActivity](pst)

	return insRepo
}

// UpdateXSorts persists the merged sibling-ordering list (xsorts) of a node.
func (repo ProductGroupRepository) UpdateXSorts(ctx context.Context, holdingCode string, guid string, xsorts []common.XSort, username string, updatedAt time.Time) error {
	filters := bson.M{
		"holdingcode": holdingCode,
		"guidfixed":   guid,
		"deletedat":   bson.M{"$exists": false},
	}

	doc := bson.M{
		"$set": bson.M{
			"xsorts":    xsorts,
			"updatedby": username,
			"updatedat": updatedAt,
		},
	}

	return repo.pst.Update(
		ctx,
		models.ProductGroupDoc{},
		filters,
		doc,
	)
}
