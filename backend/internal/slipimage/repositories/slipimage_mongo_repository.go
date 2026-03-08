package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/slipimage/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type ISlipImageMongoRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.SlipImageDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.SlipImageDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.SlipImageDoc) error
	DeleteByGuidfixed(sctx context.Context, hopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.SlipImageInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.SlipImageDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]models.SlipImageDoc, error)
	FindOne(ctx context.Context, shopID string, filters interface{}) (models.SlipImageDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.SlipImageItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.SlipImageDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.SlipImageInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.SlipImageInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.SlipImageDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(sctx context.Context, hopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.SlipImageActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.SlipImageDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.SlipImageActivity, error)

	FindByDocNo(ctx context.Context, shopID string, mode uint8, docNo string) ([]models.SlipImageInfo, error)
}

type SlipImageMongoRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.SlipImageDoc]
	repositories.SearchRepository[models.SlipImageInfo]
	repositories.GuidRepository[models.SlipImageItemGuid]
	repositories.ActivityRepository[models.SlipImageActivity, models.SlipImageDeleteActivity]
}

func NewSlipImageMongoRepository(pst microservice.IPersisterMongo) *SlipImageMongoRepository {

	insRepo := &SlipImageMongoRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.SlipImageDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.SlipImageInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.SlipImageItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.SlipImageActivity, models.SlipImageDeleteActivity](pst)

	return insRepo
}

func (repo SlipImageMongoRepository) FindByDocNo(ctx context.Context, shopID string, mode uint8, docNo string) ([]models.SlipImageInfo, error) {

	filter := bson.M{
		"shopid": shopID,
		"mode":   mode,
		"docno":  docNo,
	}

	var docs []models.SlipImageInfo

	err := repo.pst.Find(ctx, &models.SlipImageInfo{}, filter, &docs)

	if err != nil {
		return []models.SlipImageInfo{}, err
	}

	return docs, nil
}
