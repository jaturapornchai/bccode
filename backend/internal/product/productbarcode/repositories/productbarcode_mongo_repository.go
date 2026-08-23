package repositories

import (
	"context"
	"errors"
	"fmt"
	"os"
	"smlcloudplatform/internal/product/productbarcode/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IProductBarcodeRepository interface {
	Count(ctx context.Context, holdingCode string) (int, error)
	CountByRefKey(ctx context.Context, holdingCode string, itemCode string, refBarcode string) (int, error)
	CountByRefGuids(ctx context.Context, holdingCode string, GUIDs []string) (int, error)
	CountByBOMKey(ctx context.Context, holdingCode string, itemCode string, bomBarcode string) (int, error)
	CountByBOMGuids(ctx context.Context, holdingCode string, GUIDs []string) (int, error)
	CountByUnitCodes(ctx context.Context, holdingCode string, unitCodes []string) (int, error)
	CountByGroupGUIDs(ctx context.Context, holdingCode string, GUIDs []string) (int, error)
	CountByGroupCodes(ctx context.Context, holdingCode string, unitCodes []string) (int, error)
	CountByOrderTypes(ctx context.Context, holdingCode string, GUIDs []string) (int, error)
	CountByProductTypes(ctx context.Context, holdingCode string, GUIDs []string) (int, error)
	CountByBrandProducts(ctx context.Context, holdingCode string, GUIDs []string) (int, error)
	CountByDesignProducts(ctx context.Context, holdingCode string, GUIDs []string) (int, error)
	CountByModelProducts(ctx context.Context, holdingCode string, GUIDs []string) (int, error)
	CountByGroupProducts(ctx context.Context, holdingCode string, GUIDs []string) (int, error)
	CountByPatternProducts(ctx context.Context, holdingCode string, GUIDs []string) (int, error)
	CountByGradeProducts(ctx context.Context, holdingCode string, GUIDs []string) (int, error)
	CountByCategoryProducts(ctx context.Context, holdingCode string, GUIDs []string) (int, error)
	CountByClassProducts(ctx context.Context, holdingCode string, GUIDs []string) (int, error)
	CountByGroupsuboneProducts(ctx context.Context, holdingCode string, GUIDs []string) (int, error)
	CountByGroupsubtwoProducts(ctx context.Context, holdingCode string, GUIDs []string) (int, error)
	Create(ctx context.Context, doc models.ProductBarcodeDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.ProductBarcodeDoc) error
	Update(ctx context.Context, holdingCode string, guid string, doc models.ProductBarcodeDoc) error
	DeleteByGuidfixed(ctx context.Context, holdingCode string, guid string, username string) error
	Delete(ctx context.Context, holdingCode string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable micromodels.Pageable) ([]models.ProductBarcodeInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, holdingCode string, guid string) (models.ProductBarcodeDoc, error)
	FindByGuids(ctx context.Context, holdingCode string, guids []string) ([]models.ProductBarcodeDoc, error)
	FindPageFilter(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.ProductBarcodeInfo, mongopagination.PaginationData, error)
	FindPageFilterNoHoldingCode(ctx context.Context, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.ProductBarcodeInfo, mongopagination.PaginationData, error)

	FindInItemGuid(ctx context.Context, holdingCode string, columnName string, itemGuidList []string) ([]models.ProductBarcodeItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, holdingCode string, indentityField string, indentityValue interface{}) (models.ProductBarcodeDoc, error)
	FindByDocIndentityGuids(ctx context.Context, holdingCode string, indentityField string, indentityValues interface{}) ([]models.ProductBarcodeDoc, error)

	FindStep(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, selectFields map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ProductBarcodeInfo, int, error)
	FindStepNoHoldingCode(ctx context.Context, filters map[string]interface{}, searchInFields []string, selectFields map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ProductBarcodeInfo, int, error)

	FindDeletedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageable micromodels.Pageable) ([]models.ProductBarcodeDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageable micromodels.Pageable) ([]models.ProductBarcodeActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ProductBarcodeDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ProductBarcodeActivity, error)
	FindByItemCode(ctx context.Context, holdingCode string, itemCode string) ([]models.ProductBarcodeDoc, error)
	FindByItemCodeInCompany(ctx context.Context, holdingCode, businessCode, itemCode string) ([]models.ProductBarcodeDoc, error)
	FindMasterInCodes(ctx context.Context, codes []string) ([]models.ProductBarcodeInfo, error)
	UpdateParentGuidByGuids(ctx context.Context, holdingCode string, parentGUID string, guids []string) error
	Transaction(ctx context.Context, fnc func(ctx context.Context) error) error
	FindByRefBarcode(ctx context.Context, holdingCode string, barcode string) ([]models.ProductBarcodeDoc, error)
	FindByBOMBarcode(ctx context.Context, holdingCode string, barcode string) ([]models.ProductBarcodeDoc, error)
	FindByRefKey(ctx context.Context, holdingCode string, itemCode string, barcode string) ([]models.ProductBarcodeDoc, error)
	FindByBOMKey(ctx context.Context, holdingCode string, itemCode string, barcode string) ([]models.ProductBarcodeDoc, error)

	Find(ctx context.Context, holdingCode string, filters interface{}, opts ...*options.FindOptions) ([]models.ProductBarcodeDoc, error)
	FindOne(ctx context.Context, holdingCode string, filters interface{}) (models.ProductBarcodeDoc, error)
	FindByBarcode(ctx context.Context, holdingCode string, barcode string) (models.ProductBarcodeDoc, error)
	FindByBusinessKey(ctx context.Context, holdingCode string, itemCode string, barcode string) (models.ProductBarcodeDoc, error)
	FindByItemCodeAndBarcode(ctx context.Context, holdingCode string, itemCode string, barcode string) (models.ProductBarcodeDoc, error)
	FindByBarcodes(ctx context.Context, holdingCode string, barcodes []string) ([]models.ProductBarcodeInfo, error)
	FindByBarcodesInCompany(ctx context.Context, holdingCode, businessCode string, barcodes []string) ([]models.ProductBarcodeInfo, error)
	FindPageByUnits(ctx context.Context, holdingCode string, unitCodes []string, pageable micromodels.Pageable) ([]models.ProductBarcodeInfo, mongopagination.PaginationData, error)
	FindPageByGroups(ctx context.Context, holdingCode string, groupCodes []string, pageable micromodels.Pageable) ([]models.ProductBarcodeInfo, mongopagination.PaginationData, error)
	EnsureIndexes(ctx context.Context) error
	FindByGuidInCompany(ctx context.Context, holdingCode, businessCode, guid string) (models.ProductBarcodeDoc, error)
	FindByBarcodeInCompany(ctx context.Context, holdingCode, businessCode, barcode string) (models.ProductBarcodeDoc, error)
	FindByBusinessKeyInCompany(ctx context.Context, holdingCode, businessCode, itemCode, barcode string) (models.ProductBarcodeDoc, error)
	FindPageFilterInCompany(ctx context.Context, holdingCode, businessCode string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.ProductBarcodeInfo, mongopagination.PaginationData, error)
	UpdateInCompany(ctx context.Context, holdingCode, businessCode, guid string, doc models.ProductBarcodeDoc) error
	UpdateUnitSnapshotsInCompany(ctx context.Context, holdingCode, businessCode string, docs []models.ProductBarcodeDoc) error
	DeleteByGuidfixedInCompany(ctx context.Context, holdingCode, businessCode, guid, username string) error
	CountByRefKeyInCompany(ctx context.Context, holdingCode, businessCode, itemCode, refBarcode string) (int, error)
	CountByBOMKeyInCompany(ctx context.Context, holdingCode, businessCode, itemCode, bomBarcode string) (int, error)
	FindByRefKeyInCompany(ctx context.Context, holdingCode, businessCode, itemCode, barcode string) ([]models.ProductBarcodeDoc, error)
	FindByBOMKeyInCompany(ctx context.Context, holdingCode, businessCode, itemCode, barcode string) ([]models.ProductBarcodeDoc, error)

	UpdateRefBarcodeByKey(ctx context.Context, holdingCode string, itemCode string, barcode string, refBarcode models.RefProductBarcode) error
	UpdateAllProductTypeByGUID(ctx context.Context, holdingCode string, guid string, doc models.ProductType) error
	UpdateAllProductGroup(ctx context.Context, holdingCode string, doc models.ProductGroup) error
	UpdateAllProductUnitByCode(ctx context.Context, holdingCode string, doc models.ProductUnit) error
	UpdateAllProductOrderTypeByGUID(ctx context.Context, holdingCode string, guid string, doc models.ProductOrderType) error

	UpdateBranch(ctx context.Context, holdingCode string, branch models.ProductBarcodeBranch, productBarcodeGUIDFixedes []string) error
	UpdateBusinessType(ctx context.Context, holdingCode string, businessType models.ProductBarcodeBusinessType, productBarcodeGUIDFixedes []string) error
	FindByBarcodesMap(holdingCode string, barcodes []string) (map[string]models.ProductBarcodeInfo, error)
	FindByBarcodesMapInCompany(holdingCode, businessCode string, barcodes []string) (map[string]models.ProductBarcodeInfo, error)
	UpdateByID(id string, updateData bson.M) error
	UpdateByIDInCompany(holdingCode, businessCode, id string, updateData bson.M) error
}

type ProductBarcodeRepository struct {
	pst   microservice.IPersisterMongo
	cache microservice.ICacher
	repositories.CrudRepository[models.ProductBarcodeDoc]
	repositories.SearchRepository[models.ProductBarcodeInfo]
	repositories.GuidRepository[models.ProductBarcodeItemGuid]
	repositories.ActivityRepository[models.ProductBarcodeActivity, models.ProductBarcodeDeleteActivity]
}

func NewProductBarcodeRepository(pst microservice.IPersisterMongo, cache microservice.ICacher) *ProductBarcodeRepository {

	insRepo := &ProductBarcodeRepository{
		pst:   pst,
		cache: cache,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.ProductBarcodeDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.ProductBarcodeInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.ProductBarcodeItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.ProductBarcodeActivity, models.ProductBarcodeDeleteActivity](pst)

	return insRepo
}

func (repo ProductBarcodeRepository) EnsureIndexes(ctx context.Context) error {
	if _, err := repo.pst.CreateIndex(
		ctx,
		models.ProductBarcodeDoc{},
		"uniq_productbarcodes_holdingcode_guidfixed",
		bson.D{{Key: "holdingcode", Value: 1}, {Key: "guidfixed", Value: 1}},
	); err != nil {
		return fmt.Errorf("ensure product barcode guidfixed index: %w", err)
	}

	if _, err := repo.pst.CreatePartialUniqueIndex(
		ctx,
		models.ProductBarcodeDoc{},
		"uniq_productbarcodes_active_holdingcode_itemcode_barcode",
		bson.D{
			{Key: "holdingcode", Value: 1},
			{Key: "itemcode", Value: 1},
			{Key: "barcode", Value: 1},
		},
		bson.M{
			"deletedat": nil,
			"barcode":   bson.M{"$type": "string", "$gt": ""},
		},
	); err != nil {
		return fmt.Errorf("ensure active product barcode key index: %w", err)
	}

	// Stock still resolves barcode by Holding. Keep this guard until the full
	// stock spine includes BusinessCode in every key, cache and projection.
	if _, err := repo.pst.CreatePartialUniqueIndex(
		ctx,
		models.ProductBarcodeDoc{},
		"uniq_productbarcodes_active_holdingcode_barcode_stock_guard",
		bson.D{
			{Key: "holdingcode", Value: 1},
			{Key: "barcode", Value: 1},
		},
		bson.M{
			"deletedat": nil,
			"barcode":   bson.M{"$type": "string", "$gt": ""},
		},
	); err != nil {
		return fmt.Errorf("ensure temporary holding barcode stock guard: %w", err)
	}

	_, err := repo.pst.CreatePartialUniqueIndex(
		ctx,
		models.ProductBarcodeDoc{},
		"uniq_productbarcodes_active_holdingcode_businesscode_barcode",
		bson.D{
			{Key: "holdingcode", Value: 1},
			{Key: "businesscode", Value: 1},
			{Key: "barcode", Value: 1},
		},
		bson.M{
			"deletedat":    nil,
			"businesscode": bson.M{"$type": "string", "$gt": ""},
			"barcode":      bson.M{"$type": "string", "$gt": ""},
		},
	)
	if err != nil {
		return fmt.Errorf("ensure active company product barcode key index: %w", err)
	}
	return nil
}

func productBarcodeCompanyFilters(businessCode string, filters map[string]interface{}) map[string]interface{} {
	companyFilters := make(map[string]interface{}, len(filters)+1)
	for key, value := range filters {
		if key == "holdingcode" || key == "businesscode" {
			continue
		}
		companyFilters[key] = value
	}
	companyFilters["businesscode"] = businessCode
	return companyFilters
}

func (repo ProductBarcodeRepository) findInCompany(ctx context.Context, holdingCode, businessCode string, filters bson.M, opts ...*options.FindOptions) ([]models.ProductBarcodeDoc, error) {
	companyFilters := bson.M{
		"holdingcode":  holdingCode,
		"businesscode": businessCode,
		"deletedat":    bson.M{"$exists": false},
	}
	for key, value := range filters {
		if key == "holdingcode" || key == "businesscode" || key == "deletedat" {
			continue
		}
		companyFilters[key] = value
	}

	docs := []models.ProductBarcodeDoc{}
	if err := repo.pst.Find(ctx, models.ProductBarcodeDoc{}, companyFilters, &docs, opts...); err != nil {
		return nil, err
	}
	return docs, nil
}

func (repo ProductBarcodeRepository) FindByGuidInCompany(ctx context.Context, holdingCode, businessCode, guid string) (models.ProductBarcodeDoc, error) {
	doc := models.ProductBarcodeDoc{}
	err := repo.pst.FindOne(ctx, models.ProductBarcodeDoc{}, bson.M{
		"holdingcode":  holdingCode,
		"businesscode": businessCode,
		"guidfixed":    guid,
		"deletedat":    bson.M{"$exists": false},
	}, &doc)
	return doc, err
}

func (repo ProductBarcodeRepository) FindByBarcodeInCompany(ctx context.Context, holdingCode, businessCode, barcode string) (models.ProductBarcodeDoc, error) {
	docs, err := repo.findInCompany(ctx, holdingCode, businessCode, bson.M{"barcode": barcode}, options.Find().SetLimit(2))
	if err != nil {
		return models.ProductBarcodeDoc{}, err
	}
	if len(docs) == 0 {
		return models.ProductBarcodeDoc{}, nil
	}
	if len(docs) > 1 {
		return models.ProductBarcodeDoc{}, fmt.Errorf("barcode %s matches multiple products; itemcode is required", barcode)
	}
	return docs[0], nil
}

func (repo ProductBarcodeRepository) FindByBusinessKeyInCompany(ctx context.Context, holdingCode, businessCode, itemCode, barcode string) (models.ProductBarcodeDoc, error) {
	doc := models.ProductBarcodeDoc{}
	err := repo.pst.FindOne(ctx, models.ProductBarcodeDoc{}, bson.M{
		"holdingcode":  holdingCode,
		"businesscode": businessCode,
		"itemcode":     itemCode,
		"barcode":      barcode,
		"deletedat":    bson.M{"$exists": false},
	}, &doc)
	return doc, err
}

func (repo ProductBarcodeRepository) FindPageFilterInCompany(ctx context.Context, holdingCode, businessCode string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.ProductBarcodeInfo, mongopagination.PaginationData, error) {
	return repo.SearchRepository.FindPageFilter(ctx, holdingCode, productBarcodeCompanyFilters(businessCode, filters), searchInFields, pageable)
}

func (repo ProductBarcodeRepository) UpdateInCompany(ctx context.Context, holdingCode, businessCode, guid string, doc models.ProductBarcodeDoc) error {
	return repo.pst.UpdateOne(ctx, models.ProductBarcodeDoc{}, bson.M{
		"holdingcode":  holdingCode,
		"businesscode": businessCode,
		"guidfixed":    guid,
	}, doc)
}

func (repo ProductBarcodeRepository) UpdateUnitSnapshotsInCompany(ctx context.Context, holdingCode, businessCode string, docs []models.ProductBarcodeDoc) error {
	if len(docs) == 0 {
		return nil
	}
	collection, err := repo.pst.Exec(ctx, models.ProductBarcodeDoc{})
	if err != nil {
		return err
	}
	writes := make([]mongo.WriteModel, 0, len(docs))
	for _, doc := range docs {
		writes = append(writes, mongo.NewUpdateOneModel().
			SetFilter(bson.M{
				"holdingcode":  holdingCode,
				"businesscode": businessCode,
				"guidfixed":    doc.GuidFixed,
				"deletedat":    bson.M{"$exists": false},
			}).
			SetUpdate(bson.M{"$set": bson.M{
				"itemunitguid":  doc.ItemUnitGuid,
				"itemunitcode":  doc.ItemUnitCode,
				"itemunitnames": doc.ItemUnitNames,
				"condition":     doc.Condition,
				"dividevalue":   doc.DivideValue,
				"standvalue":    doc.StandValue,
				"updatedat":     doc.UpdatedAt,
				"updatedby":     doc.UpdatedBy,
			}}))
	}
	result, err := collection.BulkWrite(ctx, writes)
	if err != nil {
		return err
	}
	if result.MatchedCount != int64(len(docs)) {
		return fmt.Errorf("update product barcode unit snapshots: matched %d of %d", result.MatchedCount, len(docs))
	}
	return nil
}

func (repo ProductBarcodeRepository) DeleteByGuidfixedInCompany(ctx context.Context, holdingCode, businessCode, guid, username string) error {
	return repo.pst.SoftDelete(ctx, models.ProductBarcodeDoc{}, username, bson.M{
		"holdingcode":  holdingCode,
		"businesscode": businessCode,
		"guidfixed":    guid,
	})
}

func (repo ProductBarcodeRepository) CountByRefKeyInCompany(ctx context.Context, holdingCode, businessCode, itemCode, refBarcode string) (int, error) {
	return repo.pst.Count(ctx, models.ProductBarcodeDoc{}, bson.M{
		"holdingcode":  holdingCode,
		"businesscode": businessCode,
		"deletedat":    bson.M{"$exists": false},
		"refbarcodes":  bson.M{"$elemMatch": bson.M{"itemcode": itemCode, "barcode": refBarcode}},
	})
}

func (repo ProductBarcodeRepository) CountByBOMKeyInCompany(ctx context.Context, holdingCode, businessCode, itemCode, bomBarcode string) (int, error) {
	return repo.pst.Count(ctx, models.ProductBarcodeDoc{}, bson.M{
		"holdingcode":  holdingCode,
		"businesscode": businessCode,
		"deletedat":    bson.M{"$exists": false},
		"$or": []bson.M{
			{"bom": bson.M{"$elemMatch": bson.M{"itemcode": itemCode, "barcode": bomBarcode}}},
			{"boms": bson.M{"$elemMatch": bson.M{
				"bom": bson.M{"$elemMatch": bson.M{"itemcode": itemCode, "barcode": bomBarcode}},
			}}},
		},
	})
}

func (repo ProductBarcodeRepository) FindByRefKeyInCompany(ctx context.Context, holdingCode, businessCode, itemCode, barcode string) ([]models.ProductBarcodeDoc, error) {
	return repo.findInCompany(ctx, holdingCode, businessCode, bson.M{
		"refbarcodes": bson.M{"$elemMatch": bson.M{"itemcode": itemCode, "barcode": barcode}},
		"itemtype":    bson.M{"$ne": 2},
	})
}

func (repo ProductBarcodeRepository) FindByBOMKeyInCompany(ctx context.Context, holdingCode, businessCode, itemCode, barcode string) ([]models.ProductBarcodeDoc, error) {
	return repo.findInCompany(ctx, holdingCode, businessCode, bson.M{
		"$or": []bson.M{
			{"bom": bson.M{"$elemMatch": bson.M{"itemcode": itemCode, "barcode": barcode}}},
			{"boms": bson.M{"$elemMatch": bson.M{
				"bom": bson.M{"$elemMatch": bson.M{"itemcode": itemCode, "barcode": barcode}},
			}}},
		},
	})
}

func (repo ProductBarcodeRepository) CountByRefKey(ctx context.Context, holdingCode string, itemCode string, refBarcode string) (int, error) {
	return repo.pst.Count(ctx, models.ProductBarcodeDoc{}, bson.M{
		"holdingcode": holdingCode,
		"deletedat":   bson.M{"$exists": false},
		"refbarcodes": bson.M{"$elemMatch": bson.M{"itemcode": itemCode, "barcode": refBarcode}},
	})
}

func (repo ProductBarcodeRepository) CountByBOMKey(ctx context.Context, holdingCode string, itemCode string, bomBarcode string) (int, error) {
	return repo.pst.Count(ctx, models.ProductBarcodeDoc{}, bson.M{
		"holdingcode": holdingCode,
		"deletedat":   bson.M{"$exists": false},
		"$or": []bson.M{
			{"bom": bson.M{"$elemMatch": bson.M{"itemcode": itemCode, "barcode": bomBarcode}}},
			{"boms": bson.M{"$elemMatch": bson.M{
				"bom": bson.M{"$elemMatch": bson.M{"itemcode": itemCode, "barcode": bomBarcode}},
			}}},
		},
	})
}

func (repo ProductBarcodeRepository) CountByRefGuids(ctx context.Context, holdingCode string, GUIDs []string) (int, error) {
	return repo.CountByInKeys(ctx, holdingCode, "refbarcodes.guidfixed", GUIDs)
}

func (repo ProductBarcodeRepository) CountByBOMGuids(ctx context.Context, holdingCode string, GUIDs []string) (int, error) {
	return repo.CountByInKeys(ctx, holdingCode, "bom.guidfixed", GUIDs)
}

func (repo ProductBarcodeRepository) CountByUnitCodes(ctx context.Context, holdingCode string, unitCodes []string) (int, error) {
	return repo.CountByInKeys(ctx, holdingCode, "itemunitcode", unitCodes)
}

func (repo ProductBarcodeRepository) CountByGroupGUIDs(ctx context.Context, holdingCode string, GUIDs []string) (int, error) {
	return repo.CountByInKeys(ctx, holdingCode, "groupguid", GUIDs)
}

func (repo ProductBarcodeRepository) CountByGroupCodes(ctx context.Context, holdingCode string, unitCodes []string) (int, error) {
	return repo.CountByInKeys(ctx, holdingCode, "groupcode", unitCodes)
}

func (repo ProductBarcodeRepository) CountByOrderTypes(ctx context.Context, holdingCode string, GUIDs []string) (int, error) {
	return repo.CountByInKeys(ctx, holdingCode, "ordertypes.guidfixed", GUIDs)
}

func (repo ProductBarcodeRepository) FindByItemCode(ctx context.Context, holdingCode string, itemcode string) ([]models.ProductBarcodeDoc, error) {

	filters := bson.M{
		"holdingcode": holdingCode,
		"deletedat":   bson.M{"$exists": false},
		"itemcode":    itemcode,
	}

	result, err := repo.Find(ctx, holdingCode, filters)

	if err != nil {
		return []models.ProductBarcodeDoc{}, err
	}

	return result, nil
}

func (repo ProductBarcodeRepository) FindByItemCodeInCompany(ctx context.Context, holdingCode, businessCode, itemCode string) ([]models.ProductBarcodeDoc, error) {
	return repo.findInCompany(ctx, holdingCode, businessCode, bson.M{
		"itemcode": utils.NormalizeBusinessCode(itemCode),
	})
}

func (repo ProductBarcodeRepository) CountByProductTypes(ctx context.Context, holdingCode string, GUIDs []string) (int, error) {
	return repo.CountByInKeys(ctx, holdingCode, "producttype.guidfixed", GUIDs)
}

func (repo ProductBarcodeRepository) CountByBrandProducts(ctx context.Context, holdingCode string, GUIDs []string) (int, error) {
	return repo.CountByInKeys(ctx, holdingCode, "brandproduct.guidfixed", GUIDs)
}

func (repo ProductBarcodeRepository) CountByDesignProducts(ctx context.Context, holdingCode string, GUIDs []string) (int, error) {
	return repo.CountByInKeys(ctx, holdingCode, "designproduct.guidfixed", GUIDs)
}
func (repo ProductBarcodeRepository) CountByModelProducts(ctx context.Context, holdingCode string, GUIDs []string) (int, error) {
	return repo.CountByInKeys(ctx, holdingCode, "modelproduct.guidfixed", GUIDs)
}
func (repo ProductBarcodeRepository) CountByGroupProducts(ctx context.Context, holdingCode string, GUIDs []string) (int, error) {
	return repo.CountByInKeys(ctx, holdingCode, "groupproduct.guidfixed", GUIDs)
}
func (repo ProductBarcodeRepository) CountByPatternProducts(ctx context.Context, holdingCode string, GUIDs []string) (int, error) {
	return repo.CountByInKeys(ctx, holdingCode, "patternproduct.guidfixed", GUIDs)
}
func (repo ProductBarcodeRepository) CountByGradeProducts(ctx context.Context, holdingCode string, GUIDs []string) (int, error) {
	return repo.CountByInKeys(ctx, holdingCode, "gradeproduct.guidfixed", GUIDs)
}
func (repo ProductBarcodeRepository) CountByCategoryProducts(ctx context.Context, holdingCode string, GUIDs []string) (int, error) {
	return repo.CountByInKeys(ctx, holdingCode, "categoryproduct.guidfixed", GUIDs)
}
func (repo ProductBarcodeRepository) CountByClassProducts(ctx context.Context, holdingCode string, GUIDs []string) (int, error) {
	return repo.CountByInKeys(ctx, holdingCode, "classproduct.guidfixed", GUIDs)
}
func (repo ProductBarcodeRepository) CountByGroupsuboneProducts(ctx context.Context, holdingCode string, GUIDs []string) (int, error) {
	return repo.CountByInKeys(ctx, holdingCode, "groupsuboneproduct.guidfixed", GUIDs)
}
func (repo ProductBarcodeRepository) CountByGroupsubtwoProducts(ctx context.Context, holdingCode string, GUIDs []string) (int, error) {
	return repo.CountByInKeys(ctx, holdingCode, "groupsubtwoproduct.guidfixed", GUIDs)
}

func (repo ProductBarcodeRepository) UpdateParentGuidByGuids(ctx context.Context, holdingCode string, parentGUID string, guids []string) error {

	filters := bson.M{
		"holdingcode": holdingCode,
		"deletedat":   bson.M{"$exists": false},
		"guidfixed":   bson.M{"$in": guids},
	}

	return repo.pst.Update(ctx, models.ProductBarcodeDoc{}, filters, bson.M{"$set": bson.M{"parentguid": parentGUID}})
}

func (repo ProductBarcodeRepository) Find(ctx context.Context, holdingCode string, filters interface{}, opts ...*options.FindOptions) ([]models.ProductBarcodeDoc, error) {

	var filterQuery interface{}

	switch filterType := filters.(type) {
	case bson.M:
		tempQuery := filterType
		tempQuery["holdingcode"] = holdingCode
		tempQuery["deletedat"] = bson.M{"$exists": false}
		filterQuery = tempQuery
	default:
		return nil, errors.New("invalid query filter type")
	}

	docs := []models.ProductBarcodeDoc{}
	err := repo.pst.Find(ctx, models.ProductBarcodeDoc{}, filterQuery, &docs, opts...)

	if err != nil {
		return nil, err
	}

	return docs, nil
}

func (repo ProductBarcodeRepository) FindMasterInCodes(ctx context.Context, codes []string) ([]models.ProductBarcodeInfo, error) {

	masterHoldingCode := os.Getenv("MASTER_HOLDING_CODE")

	if len(masterHoldingCode) == 0 {
		return []models.ProductBarcodeInfo{}, errors.New("master holding code is empty")
	}

	docList := []models.ProductBarcodeInfo{}

	filters := bson.M{
		"holdingcode": masterHoldingCode,
		"barcode": bson.M{
			"$in": codes,
		},
	}

	err := repo.pst.Find(ctx, models.ProductBarcodeInfo{}, filters, &docList)

	if err != nil {
		return nil, err
	}

	return docList, nil
}

func (repo ProductBarcodeRepository) FindByRefBarcode(ctx context.Context, holdingCode string, barcode string) ([]models.ProductBarcodeDoc, error) {

	docList := []models.ProductBarcodeDoc{}

	filters := bson.M{
		"holdingcode":         holdingCode,
		"refbarcodes.barcode": barcode,
		"itemtype":            bson.M{"$ne": 2},
		"deletedat":           bson.M{"$exists": false},
	}

	err := repo.pst.Find(ctx, models.ProductBarcodeDoc{}, filters, &docList)

	if err != nil {
		return nil, err
	}

	return docList, nil
}

func (repo ProductBarcodeRepository) FindByBOMBarcode(ctx context.Context, holdingCode string, barcode string) ([]models.ProductBarcodeDoc, error) {

	docList := []models.ProductBarcodeDoc{}

	filters := bson.M{
		"holdingcode": holdingCode,
		"bom.barcode": barcode,
		"itemtype":    bson.M{"$ne": 2},
		"deletedat":   bson.M{"$exists": false},
	}

	err := repo.pst.Find(ctx, models.ProductBarcodeDoc{}, filters, &docList)

	if err != nil {
		return nil, err
	}

	return docList, nil
}

func (repo ProductBarcodeRepository) FindByRefKey(ctx context.Context, holdingCode string, itemCode string, barcode string) ([]models.ProductBarcodeDoc, error) {
	return repo.Find(ctx, holdingCode, bson.M{
		"refbarcodes": bson.M{"$elemMatch": bson.M{"itemcode": itemCode, "barcode": barcode}},
		"itemtype":    bson.M{"$ne": 2},
	})
}

func (repo ProductBarcodeRepository) FindByBOMKey(ctx context.Context, holdingCode string, itemCode string, barcode string) ([]models.ProductBarcodeDoc, error) {
	return repo.Find(ctx, holdingCode, bson.M{
		"$or": []bson.M{
			{"bom": bson.M{"$elemMatch": bson.M{"itemcode": itemCode, "barcode": barcode}}},
			{"boms": bson.M{"$elemMatch": bson.M{
				"bom": bson.M{"$elemMatch": bson.M{"itemcode": itemCode, "barcode": barcode}},
			}}},
		},
	})
}

func (repo ProductBarcodeRepository) Transaction(ctx context.Context, fnc func(ctx context.Context) error) error {
	return repo.pst.Transaction(ctx, fnc)
}

func (repo ProductBarcodeRepository) FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable micromodels.Pageable) ([]models.ProductBarcodeInfo, mongopagination.PaginationData, error) {

	results, pagination, err := repo.SearchRepository.FindPage(ctx, holdingCode, searchInFields, pageable)

	if err != nil {
		return nil, mongopagination.PaginationData{}, err
	}

	return results, pagination, nil
}

func (repo ProductBarcodeRepository) FindByBarcode(ctx context.Context, holdingCode string, barcode string) (models.ProductBarcodeDoc, error) {

	filters := bson.M{
		"holdingcode": holdingCode,
		"deletedat":   bson.M{"$exists": false},
		"barcode":     barcode,
	}

	results, err := repo.Find(ctx, holdingCode, filters, options.Find().SetLimit(2))
	if err != nil {
		return models.ProductBarcodeDoc{}, err
	}
	if len(results) == 0 {
		return models.ProductBarcodeDoc{}, nil
	}
	if len(results) > 1 {
		return models.ProductBarcodeDoc{}, fmt.Errorf("barcode %s matches multiple products; itemcode is required", barcode)
	}
	return results[0], nil
}

func (repo ProductBarcodeRepository) FindByItemCodeAndBarcode(ctx context.Context, holdingCode string, itemCode string, barcode string) (models.ProductBarcodeDoc, error) {
	return repo.FindOne(ctx, holdingCode, bson.M{
		"itemcode": itemCode,
		"barcode":  barcode,
	})
}

func (repo ProductBarcodeRepository) FindByBusinessKey(ctx context.Context, holdingCode string, itemCode string, barcode string) (models.ProductBarcodeDoc, error) {
	return repo.FindByItemCodeAndBarcode(ctx, holdingCode, itemCode, barcode)
}

func (repo ProductBarcodeRepository) FindByBarcodes(ctx context.Context, holdingCode string, barcodes []string) ([]models.ProductBarcodeInfo, error) {

	filters := bson.M{
		"holdingcode": holdingCode,
		"deletedat":   bson.M{"$exists": false},
		"barcode":     bson.M{"$in": barcodes},
	}

	var results []models.ProductBarcodeInfo
	err := repo.pst.Find(ctx, models.ProductBarcodeInfo{}, filters, &results)

	if err != nil {
		return nil, err
	}

	return results, nil
}

func (repo ProductBarcodeRepository) FindByBarcodesInCompany(ctx context.Context, holdingCode, businessCode string, barcodes []string) ([]models.ProductBarcodeInfo, error) {
	filters := bson.M{
		"holdingcode":  holdingCode,
		"businesscode": businessCode,
		"deletedat":    bson.M{"$exists": false},
		"barcode":      bson.M{"$in": barcodes},
	}

	var results []models.ProductBarcodeInfo
	if err := repo.pst.Find(ctx, models.ProductBarcodeInfo{}, filters, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func (repo ProductBarcodeRepository) FindPageByUnits(ctx context.Context, holdingCode string, unitCodes []string, pageable micromodels.Pageable) ([]models.ProductBarcodeInfo, mongopagination.PaginationData, error) {

	filters := bson.M{
		"holdingcode": holdingCode,
		"deletedat": bson.M{
			"$exists": false,
		},
		"itemunitcode": bson.M{
			"$in": unitCodes,
		},
	}

	results := []models.ProductBarcodeInfo{}
	pagination, err := repo.pst.FindPage(ctx, models.ProductBarcodeInfo{}, filters, pageable, &results)

	if err != nil {
		return nil, mongopagination.PaginationData{}, err
	}

	return results, pagination, nil
}

func (repo ProductBarcodeRepository) FindPageByGroups(ctx context.Context, holdingCode string, groupCodes []string, pageable micromodels.Pageable) ([]models.ProductBarcodeInfo, mongopagination.PaginationData, error) {

	filters := bson.M{
		"holdingcode": holdingCode,
		"deletedat": bson.M{
			"$exists": false,
		},
		"groupcode": bson.M{
			"$in": groupCodes,
		},
	}

	results := []models.ProductBarcodeInfo{}
	pagination, err := repo.pst.FindPage(ctx, models.ProductBarcodeInfo{}, filters, pageable, &results)

	if err != nil {
		return nil, mongopagination.PaginationData{}, err
	}

	return results, pagination, nil
}

func (repo ProductBarcodeRepository) UpdateRefBarcodeByKey(ctx context.Context, holdingCode string, itemCode string, barcode string, refBarcode models.RefProductBarcode) error {

	filters := bson.M{
		"holdingcode": holdingCode,
		"deletedat":   bson.M{"$exists": false},
		"refbarcodes": bson.M{"$elemMatch": bson.M{"itemcode": itemCode, "barcode": barcode}},
	}

	update := bson.M{
		"$set": bson.M{
			"refbarcodes.$.names":         refBarcode.Names,
			"refbarcodes.$.itemunitcode":  refBarcode.ItemUnitCode,
			"refbarcodes.$.itemunitnames": refBarcode.ItemUnitNames,
			"refbarcodes.$.condition":     refBarcode.Condition,
			"refbarcodes.$.dividevalue":   refBarcode.DivideValue,
			"refbarcodes.$.standvalue":    refBarcode.StandValue,
			"refbarcodes.$.qty":           refBarcode.Qty,
		},
	}

	return repo.pst.Update(ctx, models.ProductBarcodeDoc{}, filters, update)
}

func (repo ProductBarcodeRepository) UpdateAllProductTypeByGUID(ctx context.Context, holdingCode string, guid string, doc models.ProductType) error {
	filters := bson.M{
		"holdingcode":           holdingCode,
		"deletedat":             bson.M{"$exists": false},
		"producttype.guidfixed": guid,
	}

	update := bson.M{
		"$set": bson.M{
			"producttype.code":  doc.Code,
			"producttype.names": doc.Names,
		},
	}

	return repo.pst.Update(ctx, models.ProductBarcodeDoc{}, filters, update)
}

func (repo ProductBarcodeRepository) UpdateAllProductGroup(ctx context.Context, holdingCode string, doc models.ProductGroup) error {
	groupCode := utils.NormalizeBusinessCode(doc.Code)
	if groupCode == "" {
		return nil
	}
	filters := bson.M{
		"holdingcode": holdingCode,
		"deletedat":   bson.M{"$exists": false},
		"groupcode":   groupCode,
	}

	updateFields := bson.M{
		"groupcode":  groupCode,
		"groupnames": doc.Names,
	}

	return repo.pst.Update(ctx, models.ProductBarcodeDoc{}, filters, bson.M{"$set": updateFields})
}

func (repo ProductBarcodeRepository) UpdateAllProductUnitByCode(ctx context.Context, holdingCode string, doc models.ProductUnit) error {
	filters := bson.M{
		"holdingcode":  holdingCode,
		"deletedat":    bson.M{"$exists": false},
		"itemunitcode": doc.UnitCode,
	}

	update := bson.M{
		"$set": bson.M{
			"itemunitcode":  doc.UnitCode,
			"itemunitnames": doc.Names,
		},
	}

	return repo.pst.Update(ctx, models.ProductBarcodeDoc{}, filters, update)
}

func (repo ProductBarcodeRepository) UpdateAllProductOrderTypeByGUID(ctx context.Context, holdingCode string, guid string, doc models.ProductOrderType) error {
	filters := bson.M{
		"holdingcode":          holdingCode,
		"deletedat":            bson.M{"$exists": false},
		"ordertypes.guidfixed": guid,
	}

	update := bson.M{
		"$set": bson.M{
			"ordertypes.$.code":  doc.Code,
			"ordertypes.$.names": doc.Names,
			"ordertypes.$.price": doc.Price,
		},
	}

	return repo.pst.Update(ctx, models.ProductBarcodeDoc{}, filters, update)
}

func (repo ProductBarcodeRepository) UpdateBranch(ctx context.Context, holdingCode string, branch models.ProductBarcodeBranch, productBarcodeGUIDFixedes []string) error {
	filters := bson.M{
		"holdingcode": holdingCode,
		"deletedat":   bson.M{"$exists": false},
		"guidfixed":   bson.M{"$in": productBarcodeGUIDFixedes},
		"branches.code": bson.M{
			"$ne": branch.Code,
		},
	}

	update := bson.M{
		"$push": bson.M{
			"branches": branch,
		},
	}

	return repo.pst.Update(ctx, models.ProductBarcodeDoc{}, filters, update)
}

func (repo ProductBarcodeRepository) UpdateBusinessType(ctx context.Context, holdingCode string, businessType models.ProductBarcodeBusinessType, productBarcodeGUIDFixedes []string) error {
	filters := bson.M{
		"holdingcode": holdingCode,
		"deletedat":   bson.M{"$exists": false},
		"guidfixed":   bson.M{"$in": productBarcodeGUIDFixedes},
		"branches.code": bson.M{
			"$ne": businessType.Code,
		},
	}

	update := bson.M{
		"$push": bson.M{
			"businesstypes": businessType,
		},
	}

	return repo.pst.Update(ctx, models.ProductBarcodeDoc{}, filters, update)
}

func (repo ProductBarcodeRepository) FindByBarcodesMap(holdingCode string, barcodes []string) (map[string]models.ProductBarcodeInfo, error) {
	filter := bson.M{
		"holdingcode": holdingCode,
		"barcode":     bson.M{"$in": barcodes},
		"deletedat":   bson.M{"$exists": false},
	}

	var docs []models.ProductBarcodeInfo
	err := repo.pst.Find(context.Background(), models.ProductBarcodeInfo{}, filter, &docs)
	if err != nil {
		return nil, err
	}

	productMap := make(map[string]models.ProductBarcodeInfo)
	for _, product := range docs {
		if existing, exists := productMap[product.Barcode]; exists && existing.ItemCode != product.ItemCode {
			return nil, fmt.Errorf("barcode %s matches multiple products; itemcode is required", product.Barcode)
		}
		productMap[product.Barcode] = product
	}

	return productMap, nil
}

func (repo ProductBarcodeRepository) FindByBarcodesMapInCompany(holdingCode, businessCode string, barcodes []string) (map[string]models.ProductBarcodeInfo, error) {
	filter := bson.M{
		"holdingcode":  holdingCode,
		"businesscode": businessCode,
		"barcode":      bson.M{"$in": barcodes},
		"deletedat":    bson.M{"$exists": false},
	}

	var docs []models.ProductBarcodeInfo
	if err := repo.pst.Find(context.Background(), models.ProductBarcodeInfo{}, filter, &docs); err != nil {
		return nil, err
	}

	productMap := make(map[string]models.ProductBarcodeInfo, len(docs))
	for _, product := range docs {
		productMap[product.Barcode] = product
	}
	return productMap, nil
}

func (repo ProductBarcodeRepository) UpdateByID(id string, updateData bson.M) error {
	filter := bson.M{
		"guidfixed": id,
		"deletedat": bson.M{"$exists": false},
	}
	return repo.pst.Update(context.Background(), models.ProductBarcodeDoc{}, filter, updateData)
}

func (repo ProductBarcodeRepository) UpdateByIDInCompany(holdingCode, businessCode, id string, updateData bson.M) error {
	filter := bson.M{
		"holdingcode":  holdingCode,
		"businesscode": businessCode,
		"guidfixed":    id,
		"deletedat":    bson.M{"$exists": false},
	}
	return repo.pst.Update(context.Background(), models.ProductBarcodeDoc{}, filter, updateData)
}
