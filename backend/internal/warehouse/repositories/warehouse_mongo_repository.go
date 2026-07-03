package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/warehouse/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IWarehouseRepository interface {
	Count(ctx context.Context, holdingCode string) (int, error)
	Create(ctx context.Context, doc models.WarehouseDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.WarehouseDoc) error
	Update(ctx context.Context, holdingCode string, guid string, doc models.WarehouseDoc) error
	DeleteByGuidfixed(ctx context.Context, holdingCode string, guid string, username string) error
	Delete(ctx context.Context, holdingCode string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable micromodels.Pageable) ([]models.WarehouseInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, holdingCode string, guid string) (models.WarehouseDoc, error)
	// Find returns every non-deleted warehouse for the holding, unpaginated — used by GET
	// /warehouse/tree to assemble the full tree in Go (see warehouse_http.go WarehouseTree).
	Find(ctx context.Context, holdingCode string, searchInFields []string, q string) ([]models.WarehouseInfo, error)

	FindInItemGuid(ctx context.Context, holdingCode string, columnName string, itemGuidList []string) ([]models.WarehouseItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, holdingCode string, indentityField string, indentityValue interface{}) (models.WarehouseDoc, error)
	FindPageFilter(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.WarehouseInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.WarehouseInfo, int, error)

	FindDeletedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.WarehouseDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.WarehouseActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.WarehouseDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.WarehouseActivity, error)

	Transaction(ctx context.Context, queryFunc func(ctx context.Context) error) error

	EnsureIndexes(ctx context.Context) error
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

// EnsureIndexes creates the unique (holdingcode, code) index once on first use, per the
// Backend-Owned Schema Rule (backend code owns index creation, never manual createIndex).
func (repo WarehouseRepository) EnsureIndexes(ctx context.Context) error {
	collection, err := repo.pst.Exec(ctx, models.WarehouseDoc{})
	if err != nil {
		return err
	}

	_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "holdingcode", Value: 1},
			{Key: "code", Value: 1},
		},
		Options: options.Index().SetName("idxwarehousecode").SetUnique(true),
	})
	return err
}
