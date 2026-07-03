package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/warehouse/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IWarehouseLocationRepository interface {
	Create(ctx context.Context, doc models.WarehouseLocationDoc) (string, error)
	Update(ctx context.Context, holdingCode string, guid string, doc models.WarehouseLocationDoc) error
	DeleteByGuidfixed(ctx context.Context, holdingCode string, guid string, username string) error
	FindByGuid(ctx context.Context, holdingCode string, guid string) (models.WarehouseLocationDoc, error)
	FindByWarehouseAndCode(ctx context.Context, holdingCode string, warehouseGuid string, code string) (models.WarehouseLocationDoc, error)
	FindPageByWarehouse(ctx context.Context, holdingCode string, warehouseGuid string, searchInFields []string, pageable micromodels.Pageable) ([]models.WarehouseLocationInfo, mongopagination.PaginationData, error)
	FindByWarehouseGuids(ctx context.Context, holdingCode string, warehouseGuids []string) ([]models.WarehouseLocationInfo, error)

	EnsureIndexes(ctx context.Context) error
}

type WarehouseLocationRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.WarehouseLocationDoc]
	repositories.SearchRepository[models.WarehouseLocationInfo]
}

func NewWarehouseLocationRepository(pst microservice.IPersisterMongo) *WarehouseLocationRepository {
	insRepo := &WarehouseLocationRepository{
		pst: pst,
	}
	insRepo.CrudRepository = repositories.NewCrudRepository[models.WarehouseLocationDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.WarehouseLocationInfo](pst)
	return insRepo
}

func (repo WarehouseLocationRepository) FindByWarehouseAndCode(ctx context.Context, holdingCode string, warehouseGuid string, code string) (models.WarehouseLocationDoc, error) {
	filters := bson.M{
		"holdingcode":   holdingCode,
		"warehouseguid": warehouseGuid,
		"code":          code,
		"deletedat":     bson.M{"$exists": false},
	}

	doc := models.WarehouseLocationDoc{}
	err := repo.pst.FindOne(ctx, models.WarehouseLocationDoc{}, filters, &doc)
	if err != nil {
		return models.WarehouseLocationDoc{}, err
	}
	return doc, nil
}

func (repo WarehouseLocationRepository) FindPageByWarehouse(ctx context.Context, holdingCode string, warehouseGuid string, searchInFields []string, pageable micromodels.Pageable) ([]models.WarehouseLocationInfo, mongopagination.PaginationData, error) {
	return repo.SearchRepository.FindPageFilter(ctx, holdingCode, map[string]interface{}{"warehouseguid": warehouseGuid}, searchInFields, pageable)
}

func (repo WarehouseLocationRepository) FindByWarehouseGuids(ctx context.Context, holdingCode string, warehouseGuids []string) ([]models.WarehouseLocationInfo, error) {
	filters := bson.M{
		"holdingcode":   holdingCode,
		"warehouseguid": bson.M{"$in": warehouseGuids},
		"deletedat":     bson.M{"$exists": false},
	}

	docList := []models.WarehouseLocationInfo{}
	err := repo.pst.Find(ctx, models.WarehouseLocationDoc{}, filters, &docList)
	if err != nil {
		return []models.WarehouseLocationInfo{}, err
	}
	return docList, nil
}

// EnsureIndexes creates the unique (holdingcode, warehouseguid, code) index once on first use, per
// the Backend-Owned Schema Rule (backend code owns index creation, never manual createIndex).
func (repo WarehouseLocationRepository) EnsureIndexes(ctx context.Context) error {
	collection, err := repo.pst.Exec(ctx, models.WarehouseLocationDoc{})
	if err != nil {
		return err
	}

	_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "holdingcode", Value: 1},
			{Key: "warehouseguid", Value: 1},
			{Key: "code", Value: 1},
		},
		Options: options.Index().SetName("idxwarehouselocationcode").SetUnique(true),
	})
	return err
}
