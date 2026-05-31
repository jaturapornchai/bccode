package services

import (
	"context"
	"reflect"
	"smlcloudplatform/internal/product/bom/models"
	product_models "smlcloudplatform/internal/product/productbarcode/models"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"testing"
	"time"

	"github.com/smlsoft/mongopagination"
)

type captureBOMRepository struct {
	searchStepFields []string
}

func (repo *captureBOMRepository) Count(ctx context.Context, shopID string) (int, error) {
	return 0, nil
}

func (repo *captureBOMRepository) Create(ctx context.Context, doc models.ProductBarcodeBOMViewDoc) (string, error) {
	return "", nil
}

func (repo *captureBOMRepository) CreateInBatch(ctx context.Context, docList []models.ProductBarcodeBOMViewDoc) error {
	return nil
}

func (repo *captureBOMRepository) Update(ctx context.Context, shopID string, guid string, doc models.ProductBarcodeBOMViewDoc) error {
	return nil
}

func (repo *captureBOMRepository) DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error {
	return nil
}

func (repo *captureBOMRepository) Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error {
	return nil
}

func (repo *captureBOMRepository) FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.ProductBarcodeBOMViewInfo, mongopagination.PaginationData, error) {
	return []models.ProductBarcodeBOMViewInfo{}, mongopagination.PaginationData{}, nil
}

func (repo *captureBOMRepository) FindByGuid(ctx context.Context, shopID string, guid string) (models.ProductBarcodeBOMViewDoc, error) {
	return models.ProductBarcodeBOMViewDoc{}, nil
}

func (repo *captureBOMRepository) FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.ProductBarcodeBOMViewGuid, error) {
	return []models.ProductBarcodeBOMViewGuid{}, nil
}

func (repo *captureBOMRepository) FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.ProductBarcodeBOMViewDoc, error) {
	return models.ProductBarcodeBOMViewDoc{}, nil
}

func (repo *captureBOMRepository) FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, selectFields map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ProductBarcodeBOMViewInfo, int, error) {
	repo.searchStepFields = append([]string(nil), searchInFields...)
	return []models.ProductBarcodeBOMViewInfo{}, 0, nil
}

func (repo *captureBOMRepository) FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageable micromodels.Pageable) ([]models.ProductBarcodeBOMViewDeleteActivity, mongopagination.PaginationData, error) {
	return []models.ProductBarcodeBOMViewDeleteActivity{}, mongopagination.PaginationData{}, nil
}

func (repo *captureBOMRepository) FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageable micromodels.Pageable) ([]models.ProductBarcodeBOMViewActivity, mongopagination.PaginationData, error) {
	return []models.ProductBarcodeBOMViewActivity{}, mongopagination.PaginationData{}, nil
}

func (repo *captureBOMRepository) FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ProductBarcodeBOMViewDeleteActivity, error) {
	return []models.ProductBarcodeBOMViewDeleteActivity{}, nil
}

func (repo *captureBOMRepository) FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ProductBarcodeBOMViewActivity, error) {
	return []models.ProductBarcodeBOMViewActivity{}, nil
}

func (repo *captureBOMRepository) FindUseBOMByBarcode(ctx context.Context, shopID string, barcode string) (models.ProductBarcodeBOMViewDoc, error) {
	return models.ProductBarcodeBOMViewDoc{}, nil
}

func (repo *captureBOMRepository) ClearUseBOMByBarcode(ctx context.Context, shopID string, barcode string) error {
	return nil
}

func TestSearchBOMStepSearchesBarcodeAndNames(t *testing.T) {
	repo := &captureBOMRepository{}
	svc := BOMHttpService{
		repo:           repo,
		contextTimeout: time.Second,
	}

	_, _, err := svc.SearchBOMStep("shop-01", "th", map[string]interface{}{}, micromodels.PageableStep{Query: "RECIPE-001", Limit: 20})
	if err != nil {
		t.Fatalf("SearchBOMStep returned error: %v", err)
	}

	want := []string{"barcode", "names.name"}
	if !reflect.DeepEqual(repo.searchStepFields, want) {
		t.Fatalf("SearchBOMStep search fields = %#v, want %#v", repo.searchStepFields, want)
	}
}

func TestRecipeComponentMaterialTypeAllowsOnlyIngredientClassifications(t *testing.T) {
	cases := []struct {
		name         string
		materialType int8
		wantAllowed  bool
	}{
		{name: "general is rejected", materialType: product_models.MaterialTypeGeneral, wantAllowed: false},
		{name: "material is allowed", materialType: product_models.MaterialTypeMaterial, wantAllowed: true},
		{name: "semi finished is allowed", materialType: product_models.MaterialTypeSemiFinished, wantAllowed: true},
		{name: "set is rejected", materialType: product_models.MaterialTypeSet, wantAllowed: false},
		{name: "agricultural is allowed", materialType: product_models.MaterialTypeAgricultural, wantAllowed: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isAllowedRecipeComponentMaterialType(tc.materialType); got != tc.wantAllowed {
				t.Fatalf("isAllowedRecipeComponentMaterialType(%d) = %v, want %v", tc.materialType, got, tc.wantAllowed)
			}
		})
	}
}
