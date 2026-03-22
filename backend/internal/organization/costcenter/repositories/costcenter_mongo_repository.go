package repositories

import (
	"context"
	"smlcloudplatform/internal/organization/costcenter/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type ICostCenterRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.CostCenterDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.CostCenterDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.CostCenterDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.CostCenterInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.CostCenterDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.CostCenterItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.CostCenterDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.CostCenterInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.CostCenterInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.CostCenterDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.CostCenterActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.CostCenterDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.CostCenterActivity, error)

	FindOneByCode(ctx context.Context, shopID, branchCode, costCenterCode string) (models.CostCenterDoc, error)
}

type CostCenterRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.CostCenterDoc]
	repositories.SearchRepository[models.CostCenterInfo]
	repositories.GuidRepository[models.CostCenterItemGuid]
	repositories.ActivityRepository[models.CostCenterActivity, models.CostCenterDeleteActivity]
}

func NewCostCenterRepository(pst microservice.IPersisterMongo) *CostCenterRepository {

	insRepo := &CostCenterRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.CostCenterDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.CostCenterInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.CostCenterItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.CostCenterActivity, models.CostCenterDeleteActivity](pst)

	return insRepo
}

func (repo CostCenterRepository) FindOneByCode(ctx context.Context, shopID string, branchCode, costCenterCode string) (models.CostCenterDoc, error) {
	doc := models.CostCenterDoc{}
	err := repo.pst.FindOne(ctx,
		models.CostCenterDoc{},
		bson.M{
			"shopid":         shopID,
			"deletedat":      bson.M{"$exists": false},
			"branchcode":     branchCode,
			"costcentercode": costCenterCode,
		}, &doc)

	if err != nil {
		return doc, err
	}

	return doc, nil
}
