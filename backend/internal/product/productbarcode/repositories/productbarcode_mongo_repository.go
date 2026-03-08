package repositories

import (
	"context"
	"errors"
	"os"
	"smlcloudplatform/internal/product/productbarcode/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IProductBarcodeRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	CountByRefBarcode(ctx context.Context, shopID string, refBarcode string) (int, error)
	CountByRefGuids(ctx context.Context, shopID string, GUIDs []string) (int, error)
	CountByBOM(ctx context.Context, shopID string, bomBarcode string) (int, error)
	CountByBOMGuids(ctx context.Context, shopID string, GUIDs []string) (int, error)
	CountByUnitCodes(ctx context.Context, shopID string, unitCodes []string) (int, error)
	CountByGroupCodes(ctx context.Context, shopID string, unitCodes []string) (int, error)
	CountByOrderTypes(ctx context.Context, shopID string, GUIDs []string) (int, error)
	CountByProductTypes(ctx context.Context, shopID string, GUIDs []string) (int, error)
	CountByBrandProducts(ctx context.Context, shopID string, GUIDs []string) (int, error)
	CountByDesignProducts(ctx context.Context, shopID string, GUIDs []string) (int, error)
	CountByModelProducts(ctx context.Context, shopID string, GUIDs []string) (int, error)
	CountByGroupProducts(ctx context.Context, shopID string, GUIDs []string) (int, error)
	CountByPatternProducts(ctx context.Context, shopID string, GUIDs []string) (int, error)
	CountByGradeProducts(ctx context.Context, shopID string, GUIDs []string) (int, error)
	CountByCategoryProducts(ctx context.Context, shopID string, GUIDs []string) (int, error)
	CountByClassProducts(ctx context.Context, shopID string, GUIDs []string) (int, error)
	CountByGroupsuboneProducts(ctx context.Context, shopID string, GUIDs []string) (int, error)
	CountByGroupsubtwoProducts(ctx context.Context, shopID string, GUIDs []string) (int, error)
	Create(ctx context.Context, doc models.ProductBarcodeDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.ProductBarcodeDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.ProductBarcodeDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.ProductBarcodeInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.ProductBarcodeDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]models.ProductBarcodeDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.ProductBarcodeInfo, mongopagination.PaginationData, error)
	FindPageFilterNoShopid(ctx context.Context, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.ProductBarcodeInfo, mongopagination.PaginationData, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.ProductBarcodeItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.ProductBarcodeDoc, error)
	FindByDocIndentityGuids(ctx context.Context, shopID string, indentityField string, indentityValues interface{}) ([]models.ProductBarcodeDoc, error)

	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, selectFields map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ProductBarcodeInfo, int, error)
	FindStepNoShopid(ctx context.Context, filters map[string]interface{}, searchInFields []string, selectFields map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ProductBarcodeInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageable micromodels.Pageable) ([]models.ProductBarcodeDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageable micromodels.Pageable) ([]models.ProductBarcodeActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ProductBarcodeDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ProductBarcodeActivity, error)
	FindByItemCode(ctx context.Context, shopID string, itemCode string) ([]models.ProductBarcodeDoc, error)
	FindMasterInCodes(ctx context.Context, codes []string) ([]models.ProductBarcodeInfo, error)
	UpdateParentGuidByGuids(ctx context.Context, shopID string, parentGUID string, guids []string) error
	Transaction(ctx context.Context, fnc func(ctx context.Context) error) error
	FindByRefBarcode(ctx context.Context, shopID string, barcode string) ([]models.ProductBarcodeDoc, error)
	FindByBOMBarcode(ctx context.Context, shopID string, barcode string) ([]models.ProductBarcodeDoc, error)

	Find(ctx context.Context, shopID string, filters interface{}, opts ...*options.FindOptions) ([]models.ProductBarcodeDoc, error)
	FindOne(ctx context.Context, shopID string, filters interface{}) (models.ProductBarcodeDoc, error)
	FindByBarcode(ctx context.Context, shopID string, barcode string) (models.ProductBarcodeDoc, error)
	FindByBarcodes(ctx context.Context, shopID string, barcodes []string) ([]models.ProductBarcodeInfo, error)
	FindPageByUnits(ctx context.Context, shopID string, unitCodes []string, pageable micromodels.Pageable) ([]models.ProductBarcodeInfo, mongopagination.PaginationData, error)
	FindPageByGroups(ctx context.Context, shopID string, groupCodes []string, pageable micromodels.Pageable) ([]models.ProductBarcodeInfo, mongopagination.PaginationData, error)

	UpdateRefBarcodeByGUID(ctx context.Context, shopID string, guid string, refBarcode models.RefProductBarcode) error
	UpdateAllProductTypeByGUID(ctx context.Context, shopID string, guid string, doc models.ProductType) error
	UpdateAllProductGroupByCode(ctx context.Context, shopID string, doc models.ProductGroup) error
	UpdateAllProductUnitByCode(ctx context.Context, shopID string, doc models.ProductUnit) error
	UpdateAllProductOrderTypeByGUID(ctx context.Context, shopID string, guid string, doc models.ProductOrderType) error

	UpdateBranch(ctx context.Context, shopID string, branch models.ProductBarcodeBranch, productBarcodeGUIDFixedes []string) error
	UpdateBusinessType(ctx context.Context, shopID string, businessType models.ProductBarcodeBusinessType, productBarcodeGUIDFixedes []string) error
	FindByBarcodesMap(shopID string, barcodes []string) (map[string]models.ProductBarcodeInfo, error)
	UpdateByID(id string, updateData bson.M) error
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

func (repo ProductBarcodeRepository) CountByRefBarcode(ctx context.Context, shopID string, refBarcode string) (int, error) {
	return repo.CountByKey(ctx, shopID, "refbarcodes.barcode", refBarcode)
}

func (repo ProductBarcodeRepository) CountByBOM(ctx context.Context, shopID string, bomBarcode string) (int, error) {
	return repo.CountByKey(ctx, shopID, "bom.barcode", bomBarcode)
}

func (repo ProductBarcodeRepository) CountByRefGuids(ctx context.Context, shopID string, GUIDs []string) (int, error) {
	return repo.CountByInKeys(ctx, shopID, "refbarcodes.guidfixed", GUIDs)
}

func (repo ProductBarcodeRepository) CountByBOMGuids(ctx context.Context, shopID string, GUIDs []string) (int, error) {
	return repo.CountByInKeys(ctx, shopID, "bom.guidfixed", GUIDs)
}

func (repo ProductBarcodeRepository) CountByUnitCodes(ctx context.Context, shopID string, unitCodes []string) (int, error) {
	return repo.CountByInKeys(ctx, shopID, "itemunitcode", unitCodes)
}

func (repo ProductBarcodeRepository) CountByGroupCodes(ctx context.Context, shopID string, unitCodes []string) (int, error) {
	return repo.CountByInKeys(ctx, shopID, "groupcode", unitCodes)
}

func (repo ProductBarcodeRepository) CountByOrderTypes(ctx context.Context, shopID string, GUIDs []string) (int, error) {
	return repo.CountByInKeys(ctx, shopID, "ordertypes.guidfixed", GUIDs)
}

func (repo ProductBarcodeRepository) FindByItemCode(ctx context.Context, shopID string, itemcode string) ([]models.ProductBarcodeDoc, error) {

	filters := bson.M{
		"shopid":    shopID,
		"deletedat": bson.M{"$exists": false},
		"itemcode":  itemcode,
	}

	result, err := repo.Find(ctx, shopID, filters)

	if err != nil {
		return []models.ProductBarcodeDoc{}, err
	}

	return result, nil
}

func (repo ProductBarcodeRepository) CountByProductTypes(ctx context.Context, shopID string, GUIDs []string) (int, error) {
	return repo.CountByInKeys(ctx, shopID, "producttype.guidfixed", GUIDs)
}

func (repo ProductBarcodeRepository) CountByBrandProducts(ctx context.Context, shopID string, GUIDs []string) (int, error) {
	return repo.CountByInKeys(ctx, shopID, "brandproduct.guidfixed", GUIDs)
}

func (repo ProductBarcodeRepository) CountByDesignProducts(ctx context.Context, shopID string, GUIDs []string) (int, error) {
	return repo.CountByInKeys(ctx, shopID, "designproduct.guidfixed", GUIDs)
}
func (repo ProductBarcodeRepository) CountByModelProducts(ctx context.Context, shopID string, GUIDs []string) (int, error) {
	return repo.CountByInKeys(ctx, shopID, "modelproduct.guidfixed", GUIDs)
}
func (repo ProductBarcodeRepository) CountByGroupProducts(ctx context.Context, shopID string, GUIDs []string) (int, error) {
	return repo.CountByInKeys(ctx, shopID, "groupproduct.guidfixed", GUIDs)
}
func (repo ProductBarcodeRepository) CountByPatternProducts(ctx context.Context, shopID string, GUIDs []string) (int, error) {
	return repo.CountByInKeys(ctx, shopID, "patternproduct.guidfixed", GUIDs)
}
func (repo ProductBarcodeRepository) CountByGradeProducts(ctx context.Context, shopID string, GUIDs []string) (int, error) {
	return repo.CountByInKeys(ctx, shopID, "gradeproduct.guidfixed", GUIDs)
}
func (repo ProductBarcodeRepository) CountByCategoryProducts(ctx context.Context, shopID string, GUIDs []string) (int, error) {
	return repo.CountByInKeys(ctx, shopID, "categoryproduct.guidfixed", GUIDs)
}
func (repo ProductBarcodeRepository) CountByClassProducts(ctx context.Context, shopID string, GUIDs []string) (int, error) {
	return repo.CountByInKeys(ctx, shopID, "classproduct.guidfixed", GUIDs)
}
func (repo ProductBarcodeRepository) CountByGroupsuboneProducts(ctx context.Context, shopID string, GUIDs []string) (int, error) {
	return repo.CountByInKeys(ctx, shopID, "groupsuboneproduct.guidfixed", GUIDs)
}
func (repo ProductBarcodeRepository) CountByGroupsubtwoProducts(ctx context.Context, shopID string, GUIDs []string) (int, error) {
	return repo.CountByInKeys(ctx, shopID, "groupsubtwoproduct.guidfixed", GUIDs)
}

func (repo ProductBarcodeRepository) UpdateParentGuidByGuids(ctx context.Context, shopID string, parentGUID string, guids []string) error {

	filters := bson.M{
		"shopid":    shopID,
		"deletedat": bson.M{"$exists": false},
		"guidfixed": bson.M{"$in": guids},
	}

	return repo.pst.Update(ctx, models.ProductBarcodeDoc{}, filters, bson.M{"$set": bson.M{"parentguid": parentGUID}})
}

func (repo ProductBarcodeRepository) Find(ctx context.Context, shopID string, filters interface{}, opts ...*options.FindOptions) ([]models.ProductBarcodeDoc, error) {

	var filterQuery interface{}

	switch filterType := filters.(type) {
	case bson.M:
		tempQuery := filterType
		tempQuery["shopid"] = shopID
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

	masterShopID := os.Getenv("MASTER_SHOP_ID")

	if len(masterShopID) == 0 {
		return []models.ProductBarcodeInfo{}, errors.New("master shop id is empty")
	}

	docList := []models.ProductBarcodeInfo{}

	filters := bson.M{
		"shopid": masterShopID,
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

func (repo ProductBarcodeRepository) FindByRefBarcode(ctx context.Context, shopID string, barcode string) ([]models.ProductBarcodeDoc, error) {

	docList := []models.ProductBarcodeDoc{}

	filters := bson.M{
		"shopid":              shopID,
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

func (repo ProductBarcodeRepository) FindByBOMBarcode(ctx context.Context, shopID string, barcode string) ([]models.ProductBarcodeDoc, error) {

	docList := []models.ProductBarcodeDoc{}

	filters := bson.M{
		"shopid":      shopID,
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

func (repo ProductBarcodeRepository) Transaction(ctx context.Context, fnc func(ctx context.Context) error) error {
	return repo.pst.Transaction(ctx, fnc)
}

func (repo ProductBarcodeRepository) FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.ProductBarcodeInfo, mongopagination.PaginationData, error) {

	results, pagination, err := repo.SearchRepository.FindPage(ctx, shopID, searchInFields, pageable)

	if err != nil {
		return nil, mongopagination.PaginationData{}, err
	}

	return results, pagination, nil
}

func (repo ProductBarcodeRepository) FindByBarcode(ctx context.Context, shopID string, barcode string) (models.ProductBarcodeDoc, error) {

	filters := bson.M{
		"shopid":    shopID,
		"deletedat": bson.M{"$exists": false},
		"barcode":   barcode,
	}

	result, err := repo.FindOne(ctx, shopID, filters)

	if err != nil {
		return models.ProductBarcodeDoc{}, err
	}

	return result, nil
}

func (repo ProductBarcodeRepository) FindByBarcodes(ctx context.Context, shopID string, barcodes []string) ([]models.ProductBarcodeInfo, error) {

	filters := bson.M{
		"shopid":    shopID,
		"deletedat": bson.M{"$exists": false},
		"barcode":   bson.M{"$in": barcodes},
	}

	var results []models.ProductBarcodeInfo
	err := repo.pst.Find(ctx, models.ProductBarcodeInfo{}, filters, &results)

	if err != nil {
		return nil, err
	}

	return results, nil
}

func (repo ProductBarcodeRepository) FindPageByUnits(ctx context.Context, shopID string, unitCodes []string, pageable micromodels.Pageable) ([]models.ProductBarcodeInfo, mongopagination.PaginationData, error) {

	filters := bson.M{
		"shopid": shopID,
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

func (repo ProductBarcodeRepository) FindPageByGroups(ctx context.Context, shopID string, groupCodes []string, pageable micromodels.Pageable) ([]models.ProductBarcodeInfo, mongopagination.PaginationData, error) {

	filters := bson.M{
		"shopid": shopID,
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

func (repo ProductBarcodeRepository) UpdateRefBarcodeByGUID(ctx context.Context, shopID string, guid string, refBarcode models.RefProductBarcode) error {

	filters := bson.M{
		"shopid":                shopID,
		"deletedat":             bson.M{"$exists": false},
		"refbarcodes.guidfixed": guid,
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

func (repo ProductBarcodeRepository) UpdateAllProductTypeByGUID(ctx context.Context, shopID string, guid string, doc models.ProductType) error {
	filters := bson.M{
		"shopid":                shopID,
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

func (repo ProductBarcodeRepository) UpdateAllProductGroupByCode(ctx context.Context, shopID string, doc models.ProductGroup) error {
	filters := bson.M{
		"shopid":    shopID,
		"deletedat": bson.M{"$exists": false},
		"groupcode": doc.Code,
	}

	update := bson.M{
		"$set": bson.M{
			"groupcode":  doc.Code,
			"groupnames": doc.Names,
		},
	}

	return repo.pst.Update(ctx, models.ProductBarcodeDoc{}, filters, update)
}

func (repo ProductBarcodeRepository) UpdateAllProductUnitByCode(ctx context.Context, shopID string, doc models.ProductUnit) error {
	filters := bson.M{
		"shopid":       shopID,
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

func (repo ProductBarcodeRepository) UpdateAllProductOrderTypeByGUID(ctx context.Context, shopID string, guid string, doc models.ProductOrderType) error {
	filters := bson.M{
		"shopid":               shopID,
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

func (repo ProductBarcodeRepository) UpdateBranch(ctx context.Context, shopID string, branch models.ProductBarcodeBranch, productBarcodeGUIDFixedes []string) error {
	filters := bson.M{
		"shopid":    shopID,
		"deletedat": bson.M{"$exists": false},
		"guidfixed": bson.M{"$in": productBarcodeGUIDFixedes},
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

func (repo ProductBarcodeRepository) UpdateBusinessType(ctx context.Context, shopID string, businessType models.ProductBarcodeBusinessType, productBarcodeGUIDFixedes []string) error {
	filters := bson.M{
		"shopid":    shopID,
		"deletedat": bson.M{"$exists": false},
		"guidfixed": bson.M{"$in": productBarcodeGUIDFixedes},
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

func (repo ProductBarcodeRepository) FindByBarcodesMap(shopID string, barcodes []string) (map[string]models.ProductBarcodeInfo, error) {
	filter := bson.M{
		"shopid":    shopID,
		"barcode":   bson.M{"$in": barcodes},
		"deletedat": bson.M{"$exists": false},
	}

	var docs []models.ProductBarcodeInfo
	err := repo.pst.Find(context.Background(), models.ProductBarcodeInfo{}, filter, &docs)
	if err != nil {
		return nil, err
	}

	productMap := make(map[string]models.ProductBarcodeInfo)
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
