package services

import (
	"context"
	"errors"
	"fmt"
	"smlcloudplatform/internal/product/productcategory/models"
	"smlcloudplatform/internal/product/productcategory/repositories"
	"smlcloudplatform/internal/services"
	"smlcloudplatform/internal/utils"
	"time"

	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	productRepo "smlcloudplatform/internal/product/product/repositories"
	micromodels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

	GetModuleName() string
}

type ProductCategoryHttpService struct {
	repo          repositories.IProductCategoryRepository
	syncCacheRepo mastersync.IMasterSyncCacheRepository
	productRepo   productRepo.IProductRepository

	services.ActivityService[models.ProductCategoryActivity, models.ProductCategoryDeleteActivity]
	contextTimeout time.Duration
}

func NewProductCategoryHttpService(repo repositories.IProductCategoryRepository, syncCacheRepo mastersync.IMasterSyncCacheRepository, productRepo productRepo.IProductRepository) *ProductCategoryHttpService {

	contextTimeout := time.Duration(15) * time.Second

	insSvc := &ProductCategoryHttpService{
		repo:           repo,
		syncCacheRepo:  syncCacheRepo,
		productRepo:    productRepo,
		contextTimeout: contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.ProductCategoryActivity, models.ProductCategoryDeleteActivity](repo)

	return insSvc
}

func (svc ProductCategoryHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc ProductCategoryHttpService) normalizeProductCodeList(ctx context.Context, holdingCode string, codeList *[]models.CodeXSort) (*[]models.CodeXSort, error) {
	if codeList == nil || len(*codeList) == 0 {
		empty := []models.CodeXSort{}
		return &empty, nil
	}

	codes := make([]string, 0, len(*codeList))
	seen := make(map[string]struct{}, len(*codeList))
	for _, item := range *codeList {
		code := utils.NormalizeBusinessCode(item.Code)
		if code == "" {
			return nil, errors.New("รหัสสินค้าในหมวดห้ามว่าง")
		}
		if _, exists := seen[code]; exists {
			return nil, fmt.Errorf("รหัสสินค้า %s ซ้ำในหมวด", code)
		}
		seen[code] = struct{}{}
		codes = append(codes, code)
	}

	products, err := svc.productRepo.FindFilter(ctx, holdingCode, map[string]interface{}{
		"code": bson.M{"$in": codes},
	})
	if err != nil {
		return nil, err
	}
	productsByCode := make(map[string]models.CodeXSort, len(products))
	for _, product := range products {
		productsByCode[product.Code] = models.CodeXSort{
			Code:  product.Code,
			Names: product.Names,
		}
	}

	normalized := make([]models.CodeXSort, 0, len(codes))
	for index, code := range codes {
		item, exists := productsByCode[code]
		if !exists {
			return nil, fmt.Errorf("ไม่พบสินค้า %s ใน Product master", code)
		}
		item.XOrder = uint(index)
		normalized = append(normalized, item)
	}

	return &normalized, nil
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

	products, _, err := svc.productRepo.FindStep(
		ctx,
		holdingCode,
		map[string]interface{}{"materialtype": bson.M{"$ne": 1}},
		[]string{},
		map[string]interface{}{"code": 1, "names": 1},
		micromodels.PageableStep{
			Limit: 0,
			Sorts: []micromodels.KeyInt{{Key: "code", Value: 1}},
		},
	)
	if err != nil {
		return models.ProductCategoryInfo{}, err
	}

	// Build CodeList from products
	codeList := []models.CodeXSort{}
	for i, product := range products {
		codeList = append(codeList, models.CodeXSort{
			Code:   product.Code,
			XOrder: uint(i),
			Names:  product.Names,
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

	codeList, err := svc.normalizeProductCodeList(ctx, holdingCode, doc.CodeList)
	if err != nil {
		return "", err
	}
	doc.CodeList = codeList

	newGuidFixed := utils.NewGUID()

	docData := models.ProductCategoryDoc{}
	docData.HoldingCode = holdingCode
	docData.GuidFixed = newGuidFixed
	docData.ProductCategory = doc

	docData.EmptyOnNil()

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	svc.saveMasterSync(holdingCode)

	return newGuidFixed, nil
}

func (svc ProductCategoryHttpService) UpdateProductCategory(holdingCode string, guid string, authUsername string, doc models.ProductCategory) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	codeList, err := svc.normalizeProductCodeList(ctx, holdingCode, doc.CodeList)
	if err != nil {
		return err
	}
	doc.CodeList = codeList

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
	if groupNumFilter, exists := filters["groupnumber"]; exists {
		if gn, ok := groupNumFilter.(int); ok {
			groupNumber = gn
		}
	}

	// Build default "All Products" category
	defaultCategory, err := svc.buildDefaultAllProductsCategory(ctx, holdingCode, groupNumber)
	if err != nil {
		return []models.ProductCategoryInfo{}, mongopagination.PaginationData{}, fmt.Errorf("build all-products category: %w", err)
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
		codeList, err := svc.normalizeProductCodeList(ctx, holdingCode, doc.CodeList)
		if err != nil {
			return err
		}
		doc.CodeList = codeList

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

func (svc ProductCategoryHttpService) DeleteProductCategoryByGUIDs(holdingCode string, authUsername string, GUIDs []string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	deleteFilterQuery := map[string]interface{}{
		"guidfixed": bson.M{"$in": GUIDs},
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
