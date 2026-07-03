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

type IWarehouseBinRepository interface {
	Create(ctx context.Context, doc models.WarehouseBinDoc) (string, error)
	Update(ctx context.Context, holdingCode string, guid string, doc models.WarehouseBinDoc) error
	DeleteByGuidfixed(ctx context.Context, holdingCode string, guid string, username string) error
	FindByGuid(ctx context.Context, holdingCode string, guid string) (models.WarehouseBinDoc, error)
	FindByLocationAndCode(ctx context.Context, holdingCode string, locationGuid string, code string) (models.WarehouseBinDoc, error)
	FindByBarcode(ctx context.Context, holdingCode string, barcode string) (models.WarehouseBinDoc, error)
	FindPageByLocation(ctx context.Context, holdingCode string, locationGuid string, searchInFields []string, pageable micromodels.Pageable) ([]models.WarehouseBinInfo, mongopagination.PaginationData, error)
	FindByLocationGuids(ctx context.Context, holdingCode string, locationGuids []string) ([]models.WarehouseBinInfo, error)

	EnsureIndexes(ctx context.Context) error
}

type WarehouseBinRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.WarehouseBinDoc]
	repositories.SearchRepository[models.WarehouseBinInfo]
}

func NewWarehouseBinRepository(pst microservice.IPersisterMongo) *WarehouseBinRepository {
	insRepo := &WarehouseBinRepository{
		pst: pst,
	}
	insRepo.CrudRepository = repositories.NewCrudRepository[models.WarehouseBinDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.WarehouseBinInfo](pst)
	return insRepo
}

func (repo WarehouseBinRepository) FindByLocationAndCode(ctx context.Context, holdingCode string, locationGuid string, code string) (models.WarehouseBinDoc, error) {
	filters := bson.M{
		"holdingcode":  holdingCode,
		"locationguid": locationGuid,
		"code":         code,
		"deletedat":    bson.M{"$exists": false},
	}

	doc := models.WarehouseBinDoc{}
	err := repo.pst.FindOne(ctx, models.WarehouseBinDoc{}, filters, &doc)
	if err != nil {
		return models.WarehouseBinDoc{}, err
	}
	return doc, nil
}

func (repo WarehouseBinRepository) FindByBarcode(ctx context.Context, holdingCode string, barcode string) (models.WarehouseBinDoc, error) {
	filters := bson.M{
		"holdingcode": holdingCode,
		"barcode":     barcode,
		"deletedat":   bson.M{"$exists": false},
	}

	doc := models.WarehouseBinDoc{}
	err := repo.pst.FindOne(ctx, models.WarehouseBinDoc{}, filters, &doc)
	if err != nil {
		return models.WarehouseBinDoc{}, err
	}
	return doc, nil
}

func (repo WarehouseBinRepository) FindPageByLocation(ctx context.Context, holdingCode string, locationGuid string, searchInFields []string, pageable micromodels.Pageable) ([]models.WarehouseBinInfo, mongopagination.PaginationData, error) {
	return repo.SearchRepository.FindPageFilter(ctx, holdingCode, map[string]interface{}{"locationguid": locationGuid}, searchInFields, pageable)
}

func (repo WarehouseBinRepository) FindByLocationGuids(ctx context.Context, holdingCode string, locationGuids []string) ([]models.WarehouseBinInfo, error) {
	filters := bson.M{
		"holdingcode":  holdingCode,
		"locationguid": bson.M{"$in": locationGuids},
		"deletedat":    bson.M{"$exists": false},
	}

	docList := []models.WarehouseBinInfo{}
	err := repo.pst.Find(ctx, models.WarehouseBinDoc{}, filters, &docList)
	if err != nil {
		return []models.WarehouseBinInfo{}, err
	}
	return docList, nil
}

// EnsureIndexes creates the warehousebin unique indexes once on first use, per the Backend-Owned
// Schema Rule. barcode uniqueness is a PARTIAL unique index (only enforced when barcode is a
// non-empty string) so multiple bins with no barcode can coexist.
func (repo WarehouseBinRepository) EnsureIndexes(ctx context.Context) error {
	collection, err := repo.pst.Exec(ctx, models.WarehouseBinDoc{})
	if err != nil {
		return err
	}

	_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "holdingcode", Value: 1},
			{Key: "warehouseguid", Value: 1},
			{Key: "locationguid", Value: 1},
			{Key: "code", Value: 1},
		},
		Options: options.Index().SetName("idxwarehousebincode").SetUnique(true),
	})
	if err != nil {
		return err
	}

	_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "holdingcode", Value: 1},
			{Key: "barcode", Value: 1},
		},
		Options: options.Index().
			SetName("idxwarehousebinbarcode").
			SetUnique(true).
			SetPartialFilterExpression(bson.D{
				{Key: "barcode", Value: bson.D{{Key: "$type", Value: "string"}, {Key: "$gt", Value: ""}}},
			}),
	})
	return err
}
