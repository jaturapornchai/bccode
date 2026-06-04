package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"smlcloudplatform/internal/product/productcategory/models"
	"smlcloudplatform/internal/product/productcategory/repositories"
	"smlcloudplatform/internal/services"
	"smlcloudplatform/internal/utils"
	"time"

	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	productbarcodeRepo "smlcloudplatform/internal/product/productbarcode/repositories"
	micromodels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IProductCategoryHttpService interface {
	CreateProductCategory(holdingCode string, authUsername string, doc models.ProductCategory) (string, error)
	UpdateProductCategory(holdingCode string, guid string, authUsername string, doc models.ProductCategory) error
	DeleteProductCategory(holdingCode string, guid string, authUsername string) error
	DeleteProductCategoryByGUIDs(holdingCode string, authUsername string, GUIDs []string) error
	InfoProductCategory(holdingCode string, guid string) (models.ProductCategoryInfo, error)
	SearchProductCategory(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ProductCategoryInfo, mongopagination.PaginationData, error)
	SearchProductCategoryStep(holdingCode string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ProductCategoryInfo, int, error)
	SaveInBatch(holdingCode string, authUsername string, dataList []models.ProductCategory) error
	XSortsSave(holdingCode string, authUsername string, xsorts []common.XSortModifyReqesut) error
	XBarcodesSave(holdingCode string, authUsername string, xsorts []common.XSortModifyReqesut) error
	UpdateBarcode(holdingCode string, codeXSort models.CodeXSort) error

	GetModuleName() string
}

type ProductCategoryHttpService struct {
	repo               repositories.IProductCategoryRepository
	syncCacheRepo      mastersync.IMasterSyncCacheRepository
	productBarcodeRepo productbarcodeRepo.IProductBarcodeRepository

	services.ActivityService[models.ProductCategoryActivity, models.ProductCategoryDeleteActivity]
	contextTimeout time.Duration
}

func NewProductCategoryHttpService(repo repositories.IProductCategoryRepository, syncCacheRepo mastersync.IMasterSyncCacheRepository, productBarcodeRepo productbarcodeRepo.IProductBarcodeRepository) *ProductCategoryHttpService {

	contextTimeout := time.Duration(15) * time.Second

	insSvc := &ProductCategoryHttpService{
		repo:               repo,
		syncCacheRepo:      syncCacheRepo,
		productBarcodeRepo: productBarcodeRepo,
		contextTimeout:     contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.ProductCategoryActivity, models.ProductCategoryDeleteActivity](repo)

	return insSvc
}

func (svc ProductCategoryHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc ProductCategoryHttpService) buildDefaultAllProductsCategory(ctx context.Context, holdingCode string, groupNumber int) (models.ProductCategoryInfo, error) {
	// Set multi-language names
	thName := "สินค้าทั้งหมด"
	enName := "All"
	cnName := "全部"
	krName := "모두"
	jpName := "すべて"

	names := []common.NameX{
		{Code: &[]string{"th"}[0], Name: &thName, IsAuto: false, IsDelete: false},
		{Code: &[]string{"en"}[0], Name: &enName, IsAuto: false, IsDelete: false},
		{Code: &[]string{"cn"}[0], Name: &cnName, IsAuto: false, IsDelete: false},
		{Code: &[]string{"kr"}[0], Name: &krName, IsAuto: false, IsDelete: false},
		{Code: &[]string{"jp"}[0], Name: &jpName, IsAuto: false, IsDelete: false},
	}

	// Fetch all products where materialtype != 1
	filters := bson.M{
		"holding_code": holdingCode,
		"deleted_at":   bson.M{"$exists": false},
		"materialtype": bson.M{"$ne": 1},
	}

	findOpts := options.Find()
	findOpts.SetProjection(bson.M{
		"barcode":          1,
		"itemcode":         1,
		"names":            1,
		"item_unit_code":   1,
		"itemunitnames":    1,
		"manufacturerguid": 1,
	})
	findOpts.SetSort(bson.M{"barcode": 1})

	products, err := svc.productBarcodeRepo.Find(ctx, holdingCode, filters, findOpts)
	if err != nil {
		return models.ProductCategoryInfo{}, err
	}

	// Build CodeList from products
	codeList := []models.CodeXSort{}
	for i, product := range products {
		codeList = append(codeList, models.CodeXSort{
			Code:             product.ItemCode,
			XOrder:           uint(i),
			Barcode:          product.Barcode,
			UnitCode:         product.ItemUnitCode,
			UnitNames:        product.ItemUnitNames,
			Names:            product.Names,
			ManufacturerGUID: product.ManufacturerGUID,
		})
	}

	// Initialize empty slices for required fields
	xsorts := []common.XSort{}

	xsorts = append(xsorts, common.XSort{
		Code:   "X",
		XOrder: 0,
	})
	timeForSales := []models.ProductCategoryTimeForSale{}

	guidFixed := "00000000000000000"

	// Return ProductCategoryInfo
	categoryInfo := models.ProductCategoryInfo{
		ProductCategory: models.ProductCategory{
			ChildCount:      0,
			ParentGUID:      "",
			ParentGUIDAll:   "",
			ImageUri:        "",
			Names:           &names,
			XSorts:          &xsorts,
			CodeList:        &codeList,
			UseImageOrColor: false,
			ColorSelect:     "",
			ColorSelectHex:  "",
			IsDisabled:      false,
			CoverURI:        "",
			GroupNumber:     groupNumber,
			TimeForSales:    &timeForSales,
		},
	}
	categoryInfo.GuidFixed = guidFixed

	return categoryInfo, nil
}

func (svc ProductCategoryHttpService) CreateProductCategory(holdingCode string, authUsername string, doc models.ProductCategory) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	newGuidFixed := utils.NewGUID()

	docData := models.ProductCategoryDoc{}
	docData.HoldingCode = holdingCode
	docData.GuidFixed = newGuidFixed
	docData.ProductCategory = doc

	docData.EmptyOnNil()

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err := svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	svc.saveMasterSync(holdingCode)

	return newGuidFixed, nil
}

func (svc ProductCategoryHttpService) UpdateProductCategory(holdingCode string, guid string, authUsername string, doc models.ProductCategory) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	findDoc.ProductCategory = doc

	findDoc.UpdatedBy = authUsername
	findDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, holdingCode, guid, findDoc)

	if err != nil {
		return err
	}

	svc.saveMasterSync(holdingCode)

	return nil
}

func (svc ProductCategoryHttpService) UpdateBarcode(holdingCode string, codeXSort models.CodeXSort) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	return svc.repo.UpdateCodeList(ctx, holdingCode, codeXSort)
}

func (svc ProductCategoryHttpService) DeleteProductCategory(holdingCode string, guid string, authUsername string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	err = svc.repo.DeleteByGuidfixed(ctx, holdingCode, guid, authUsername)
	if err != nil {
		return err
	}

	svc.saveMasterSync(holdingCode)

	return nil
}

func (svc ProductCategoryHttpService) InfoProductCategory(holdingCode string, guid string) (models.ProductCategoryInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return models.ProductCategoryInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.ProductCategoryInfo{}, errors.New("document not found")
	}

	return findDoc.ProductCategoryInfo, nil

}

func (svc ProductCategoryHttpService) SearchProductCategory(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ProductCategoryInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	// Extract groupNumber from filters (default to 0)
	groupNumber := 0
	if groupNumFilter, exists := filters["group_number"]; exists {
		if gn, ok := groupNumFilter.(int); ok {
			groupNumber = gn
		}
	}

	// Build default "All Products" category
	defaultCategory, err := svc.buildDefaultAllProductsCategory(ctx, holdingCode, groupNumber)
	if err != nil {
		// Log error but continue without default category
		log.Printf("Failed to build default category: %v", err)
	}

	searchInFields := []string{
		"code",
		"names.name",
	}

	if len(pageable.Sorts) == 0 {
		pageable.Sorts = []micromodels.KeyInt{
			{Key: "code", Value: 1},
		}
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, holdingCode, filters, searchInFields, pageable)

	if err != nil {
		return []models.ProductCategoryInfo{}, pagination, err
	}

	// Create new slice with default category first
	var finalDocList []models.ProductCategoryInfo

	if defaultCategory.GuidFixed != "" { // Check if default category was built successfully
		finalDocList = append(finalDocList, defaultCategory)
		// Increment total count since we added default category
		pagination.Total = pagination.Total + 1
	}

	// Append regular categories
	finalDocList = append(finalDocList, docList...)

	for i := range finalDocList {
		if finalDocList[i].TimeForSales == nil {
			finalDocList[i].TimeForSales = &[]models.ProductCategoryTimeForSale{}
		}
	}

	return finalDocList, pagination, nil
}

func (svc ProductCategoryHttpService) SearchProductCategoryStep(holdingCode string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ProductCategoryInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
		"names.name",
	}

	if len(pageableStep.Sorts) == 0 {
		pageableStep.Sorts = []micromodels.KeyInt{
			{Key: "code", Value: 1},
		}
	}

	selectFields := map[string]interface{}{}

	docList, total, err := svc.repo.FindStep(ctx, holdingCode, filters, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.ProductCategoryInfo{}, 0, err
	}

	for i := range docList {
		if docList[i].TimeForSales == nil {
			docList[i].TimeForSales = &[]models.ProductCategoryTimeForSale{}
		}
	}

	return docList, total, nil
}

func (svc ProductCategoryHttpService) SaveInBatch(holdingCode string, authUsername string, dataList []models.ProductCategory) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	createDataList := []models.ProductCategoryDoc{}

	createdAt := time.Now()
	for _, doc := range dataList {

		newGuidFixed := utils.NewGUID()

		docData := models.ProductCategoryDoc{}
		docData.HoldingCode = holdingCode
		docData.GuidFixed = newGuidFixed
		docData.ProductCategory = doc

		docData.EmptyOnNil()

		docData.CreatedBy = authUsername
		docData.CreatedAt = createdAt

		createDataList = append(createDataList, docData)
	}

	if len(dataList) > 0 {
		err := svc.repo.CreateInBatch(ctx, createDataList)

		if err != nil {
			return err
		}

	}

	svc.saveMasterSync(holdingCode)

	return nil
}

func (svc ProductCategoryHttpService) XSortsSave(holdingCode string, authUsername string, xsorts []common.XSortModifyReqesut) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	for _, xsort := range xsorts {
		if len(xsort.GUIDFixed) < 1 {
			continue
		}
		findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, xsort.GUIDFixed)

		if err != nil {
			return err
		}

		if len(findDoc.GuidFixed) < 1 {
			continue
		}

		if findDoc.XSorts == nil {
			findDoc.XSorts = &[]common.XSort{}
		}

		dictXSorts := map[string]common.XSort{}

		for _, tempXSort := range *findDoc.XSorts {
			dictXSorts[tempXSort.Code] = tempXSort
		}

		dictXSorts[xsort.Code] = common.XSort{
			Code:   xsort.Code,
			XOrder: xsort.XOrder,
		}

		tempXSorts := []common.XSort{}

		for _, tempXSort := range dictXSorts {
			tempXSorts = append(tempXSorts, tempXSort)
		}

		findDoc.XSorts = &tempXSorts

		err = svc.repo.UpdateXSorts(ctx, holdingCode, findDoc.GuidFixed, tempXSorts, authUsername, time.Now())

		if err != nil {
			return err
		}
	}

	svc.saveMasterSync(holdingCode)

	return nil

}

func (svc ProductCategoryHttpService) XBarcodesSave(holdingCode string, authUsername string, xsorts []common.XSortModifyReqesut) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	for _, xsort := range xsorts {
		if len(xsort.GUIDFixed) < 1 {
			continue
		}
		findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, xsort.GUIDFixed)

		if err != nil {
			return err
		}

		if len(findDoc.GuidFixed) < 1 {
			continue
		}

		if findDoc.CodeList == nil {
			findDoc.CodeList = &[]models.CodeXSort{}
		}

		dictXSorts := map[string]models.CodeXSort{}

		for _, tempXSort := range *findDoc.CodeList {
			dictXSorts[tempXSort.Code] = tempXSort
		}

		dictXSorts[xsort.Code] = models.CodeXSort{
			Code:   xsort.Code,
			XOrder: xsort.XOrder,
		}

		tempXSorts := []models.CodeXSort{}

		for _, tempXSort := range dictXSorts {
			tempXSorts = append(tempXSorts, tempXSort)
		}

		findDoc.CodeList = &tempXSorts

		findDoc.UpdatedBy = authUsername
		findDoc.UpdatedAt = time.Now()

		err = svc.repo.Update(ctx, holdingCode, findDoc.GuidFixed, findDoc)

		if err != nil {
			return err
		}

	}

	svc.saveMasterSync(holdingCode)

	return nil

}

func (svc ProductCategoryHttpService) DeleteProductCategoryByGUIDs(holdingCode string, authUsername string, GUIDs []string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	deleteFilterQuery := map[string]interface{}{
		"guid_fixed": bson.M{"$in": GUIDs},
	}

	err := svc.repo.Delete(ctx, holdingCode, authUsername, deleteFilterQuery)
	if err != nil {
		return err
	}

	svc.saveMasterSync(holdingCode)

	return nil
}

func (svc ProductCategoryHttpService) saveMasterSync(holdingCode string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(holdingCode, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc ProductCategoryHttpService) GetModuleName() string {
	return "productcategory"
}
