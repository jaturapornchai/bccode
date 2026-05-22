package repositories

import (
	"context"
	"smlcloudplatform/internal/debtaccount/debtor/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type IDebtorRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.DebtorDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.DebtorDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.DebtorDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.DebtorInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.DebtorDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]models.DebtorDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.DebtorItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.DebtorDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.DebtorInfo, mongopagination.PaginationData, error)
	FindPageFilterNoShopid(ctx context.Context, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.DebtorInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.DebtorInfo, int, error)
	FindStepNoShopid(ctx context.Context, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.DebtorInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.DebtorDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.DebtorActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.DebtorDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.DebtorActivity, error)

	FindAuthByUsername(ctx context.Context, shopID string, username string) (models.DebtorDoc, error)
}

type DebtorRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.DebtorDoc]
	repositories.SearchRepository[models.DebtorInfo]
	repositories.GuidRepository[models.DebtorItemGuid]
	repositories.ActivityRepository[models.DebtorActivity, models.DebtorDeleteActivity]
}

func NewDebtorRepository(pst microservice.IPersisterMongo) *DebtorRepository {

	insRepo := &DebtorRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.DebtorDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.DebtorInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.DebtorItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.DebtorActivity, models.DebtorDeleteActivity](pst)

	return insRepo
}

func (repo DebtorRepository) FindAuthByUsername(ctx context.Context, shopID string, username string) (models.DebtorDoc, error) {
	var doc models.DebtorDoc

	filter := bson.M{
		"shopid":        shopID,
		"deleted_at":     bson.M{"$exists": false},
		"auth.username": username,
	}

	err := repo.pst.FindOne(ctx, models.DebtorDoc{}, filter, &doc)

	if err != nil {
		return doc, err
	}

	return doc, err
}
