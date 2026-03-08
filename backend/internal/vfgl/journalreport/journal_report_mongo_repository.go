package journalreport

import (
	"context"
	"smlcloudplatform/internal/vfgl/journalreport/models"
	"smlcloudplatform/pkg/microservice"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type IJournalReportMongoRepository interface {
	FindCountDetailByDocs(ctx context.Context, shopID string, docs []string) ([]models.JournalSummary, error)
	FindCountImageByDocs(ctx context.Context, shopID string, docs []string) ([]models.JournalImageSummary, error)
	FindImageCountByShops(ctx context.Context, shopIDs []string, startDate time.Time, endDate time.Time) ([]models.ShopImageCount, error)
	GetDocNosByShopsAndDateRange(ctx context.Context, shopIDs []string, startDate time.Time, endDate time.Time) (map[string][]string, error)
	CountImagesByDocNos(ctx context.Context, shopDocNos map[string][]string) ([]models.ShopImageCount, error)
	CountAllImagesByShops(ctx context.Context, shopIDs []string) ([]models.ShopImageCount, error)
}

type JournalMongoRepository struct {
	pst microservice.IPersisterMongo
}

func NewJournalMongoRepository(pst microservice.IPersisterMongo) *JournalMongoRepository {

	insRepo := &JournalMongoRepository{
		pst: pst,
	}

	return insRepo
}

func (repo *JournalMongoRepository) FindCountDetailByDocs(ctx context.Context, shopID string, docs []string) ([]models.JournalSummary, error) {

	matchQuery := bson.M{
		"shopid": shopID,
		"docno":  bson.M{"$in": docs},
		"vats":   bson.M{"$exists": true, "$type": "array"},
		"taxes":  bson.M{"$exists": true, "$type": "array"},
	}

	projectQuery := bson.M{
		"docno":    1,
		"countvat": bson.M{"$size": "$vats"},
		"counttax": bson.M{"$size": "$taxes"},
	}

	query := []interface{}{
		bson.M{"$match": matchQuery},
		bson.M{"$project": projectQuery},
	}

	docList := []models.JournalSummary{}
	err := repo.pst.Aggregate(ctx, &models.JournalSummary{}, query, &docList)

	if err != nil {
		return []models.JournalSummary{}, err
	}

	return docList, nil
}

func (repo *JournalMongoRepository) FindCountImageByDocs(ctx context.Context, shopID string, docs []string) ([]models.JournalImageSummary, error) {

	matchQuery := bson.M{
		"shopid":            shopID,
		"imagereferences":   bson.M{"$exists": true},
		"references.module": "GL",
		"references.docno":  bson.M{"$in": docs},
	}

	projectQuery := bson.M{
		"docno":      bson.M{"$first": "$references.docno"},
		"countimage": bson.M{"$size": "$imagereferences"},
	}

	query := []interface{}{
		bson.M{"$match": matchQuery},
		bson.M{"$project": projectQuery},
	}

	docList := []models.JournalImageSummary{}
	err := repo.pst.Aggregate(ctx, &models.JournalImageSummary{}, query, &docList)

	if err != nil {
		return []models.JournalImageSummary{}, err
	}

	return docList, nil
}

func (repo *JournalMongoRepository) FindImageCountByShops(
	ctx context.Context,
	shopIDs []string,
	startDate time.Time,
	endDate time.Time,
) ([]models.ShopImageCount, error) {

	// First, we need to get all journals in the date range to get docno list
	// Then count images that reference those docnos
	// This is a two-step aggregation:
	// 1. Unwind references array
	// 2. Match by shopid, module=GL, and references in the docno list
	// 3. Group by shopid and count imagereferences

	// Match stage - find all document image groups with GL module references
	matchStage := bson.M{
		"shopid":            bson.M{"$in": shopIDs},
		"references.module": "GL",
		"imagereferences":   bson.M{"$exists": true},
	}

	// Unwind references to access individual reference items
	unwindStage := bson.M{
		"path":                       "$references",
		"preserveNullAndEmptyArrays": false,
	}

	// Match only GL module references
	matchGLStage := bson.M{
		"references.module": "GL",
	}

	// Group stage - sum image counts per shop
	groupStage := bson.M{
		"_id": bson.M{
			"shopid": "$shopid",
			"docno":  "$references.docno",
		},
		"imagecount": bson.M{
			"$sum": bson.M{"$size": "$imagereferences"},
		},
	}

	// Group again by shopid to sum all images
	groupByShopStage := bson.M{
		"_id": "$_id.shopid",
		"imagecount": bson.M{
			"$sum": "$imagecount",
		},
	}

	// Project stage
	projectStage := bson.M{
		"shopid":     "$_id",
		"imagecount": 1,
		"_id":        0,
	}

	pipeline := []interface{}{
		bson.M{"$match": matchStage},
		bson.M{"$unwind": unwindStage},
		bson.M{"$match": matchGLStage},
		bson.M{"$group": groupStage},
		bson.M{"$group": groupByShopStage},
		bson.M{"$project": projectStage},
	}

	results := []models.ShopImageCount{}
	err := repo.pst.Aggregate(ctx, &models.ShopImageCount{}, pipeline, &results)

	if err != nil {
		return []models.ShopImageCount{}, err
	}

	return results, nil
}

// GetDocNosByShopsAndDateRange retrieves all journal docnos for given shops within date range
func (repo *JournalMongoRepository) GetDocNosByShopsAndDateRange(
	ctx context.Context,
	shopIDs []string,
	startDate time.Time,
	endDate time.Time,
) (map[string][]string, error) {

	matchStage := bson.M{
		"shopid": bson.M{"$in": shopIDs},
		"docdate": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
		"journaltype": 0, // Only regular journals (not closing entries)
	}

	projectStage := bson.M{
		"shopid": 1,
		"docno":  1,
		"_id":    0,
	}

	pipeline := []interface{}{
		bson.M{"$match": matchStage},
		bson.M{"$project": projectStage},
	}

	results := []models.JournalDocNoByShop{}
	err := repo.pst.Aggregate(ctx, &models.JournalDocNoByShop{}, pipeline, &results)

	if err != nil {
		return nil, err
	}

	// Group by shopID
	shopDocNos := make(map[string][]string)
	for _, item := range results {
		shopDocNos[item.ShopID] = append(shopDocNos[item.ShopID], item.DocNo)
	}

	return shopDocNos, nil
}

// CountImagesByDocNos counts images for each shop based on their docno lists
func (repo *JournalMongoRepository) CountImagesByDocNos(
	ctx context.Context,
	shopDocNos map[string][]string,
) ([]models.ShopImageCount, error) {

	results := []models.ShopImageCount{}

	// Process each shop separately
	for shopID, docNos := range shopDocNos {
		if len(docNos) == 0 {
			continue
		}

		// Match documentImageGroups that reference these docnos for this shop
		matchStage := bson.M{
			"shopid":            shopID,
			"references.module": "GL",
			"references.docno":  bson.M{"$in": docNos},
			"imagereferences":   bson.M{"$exists": true},
		}

		groupStage := bson.M{
			"_id": "$shopid",
			"imagecount": bson.M{
				"$sum": bson.M{"$size": "$imagereferences"},
			},
		}

		projectStage := bson.M{
			"shopid":     "$_id",
			"imagecount": 1,
			"_id":        0,
		}

		pipeline := []interface{}{
			bson.M{"$match": matchStage},
			bson.M{"$group": groupStage},
			bson.M{"$project": projectStage},
		}

		shopResults := []models.ShopImageCount{}
		err := repo.pst.Aggregate(ctx, &models.ShopImageCount{}, pipeline, &shopResults)

		if err != nil {
			return nil, err
		}

		results = append(results, shopResults...)
	}

	return results, nil
}

// CountAllImagesByShops counts all images for each shop without any filtering by docno or module
func (repo *JournalMongoRepository) CountAllImagesByShops(
	ctx context.Context,
	shopIDs []string,
) ([]models.ShopImageCount, error) {

	// Match all documentImageGroups for the given shops that have imagereferences
	matchStage := bson.M{
		"shopid":          bson.M{"$in": shopIDs},
		"imagereferences": bson.M{"$exists": true},
	}

	// Group by shopid and sum all imagereferences sizes
	groupStage := bson.M{
		"_id": "$shopid",
		"imagecount": bson.M{
			"$sum": bson.M{"$size": "$imagereferences"},
		},
	}

	// Project to match our model structure
	projectStage := bson.M{
		"shopid":     "$_id",
		"imagecount": 1,
		"_id":        0,
	}

	pipeline := []interface{}{
		bson.M{"$match": matchStage},
		bson.M{"$group": groupStage},
		bson.M{"$project": projectStage},
	}

	results := []models.ShopImageCount{}
	err := repo.pst.Aggregate(ctx, &models.ShopImageCount{}, pipeline, &results)

	if err != nil {
		return []models.ShopImageCount{}, err
	}

	return results, nil
}
