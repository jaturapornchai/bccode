package repositories

import (
	"context"
	"smlcloudplatform/internal/organization/businesstype/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type IBusinessTypeRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.BusinessTypeDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.BusinessTypeDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.BusinessTypeDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.BusinessTypeInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.BusinessTypeDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.BusinessTypeItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.BusinessTypeDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.BusinessTypeInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.BusinessTypeInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.BusinessTypeDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.BusinessTypeActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.BusinessTypeDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.BusinessTypeActivity, error)

	ClearDefault(ctx context.Context, shopID string) error
	GetDefault(ctx context.Context, shopID string) (models.BusinessTypeDoc, error)

	Transaction(ctx context.Context, transaction func(ctx context.Context) error) error
}

type BusinessTypeRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.BusinessTypeDoc]
	repositories.SearchRepository[models.BusinessTypeInfo]
	repositories.GuidRepository[models.BusinessTypeItemGuid]
	repositories.ActivityRepository[models.BusinessTypeActivity, models.BusinessTypeDeleteActivity]
}

func NewBusinessTypeRepository(pst microservice.IPersisterMongo) *BusinessTypeRepository {

	insRepo := &BusinessTypeRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.BusinessTypeDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.BusinessTypeInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.BusinessTypeItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.BusinessTypeActivity, models.BusinessTypeDeleteActivity](pst)

	return insRepo
}

func (repo BusinessTypeRepository) ClearDefault(ctx context.Context, shopID string) error {

	filter := bson.M{
		"shopid": shopID,
		"deletedat": bson.M{
			"$exists": false,
		},
		"isdefault": true,
	}

	data := bson.M{
		"$set": bson.M{
			"isdefault": false,
		},
	}

	return repo.pst.Update(ctx, models.BusinessTypeDoc{}, filter, data)
}

func (repo BusinessTypeRepository) GetDefault(ctx context.Context, shopID string) (models.BusinessTypeDoc, error) {

	filter := bson.M{
		"shopid": shopID,
		"deletedat": bson.M{
			"$exists": false,
		},
		"isdefault": true,
	}

	doc := models.BusinessTypeDoc{}
	err := repo.pst.FindOne(ctx, models.BusinessTypeDoc{}, filter, &doc, nil)

	return doc, err
}

func (repo BusinessTypeRepository) Transaction(ctx context.Context, transaction func(ctx context.Context) error) error {
	return repo.pst.Transaction(ctx, transaction)
}
