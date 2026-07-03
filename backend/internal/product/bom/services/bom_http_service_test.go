package services

import (
	"context"
	"reflect"
	common_models "smlcloudplatform/internal/models"
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

func (repo *captureBOMRepository) Count(ctx context.Context, holdingCode string) (int, error) {
	return 0, nil
}

func (repo *captureBOMRepository) Create(ctx context.Context, doc models.ProductBarcodeBOMViewDoc) (string, error) {
	return "", nil
}

func (repo *captureBOMRepository) CreateInBatch(ctx context.Context, docList []models.ProductBarcodeBOMViewDoc) error {
	return nil
}

func (repo *captureBOMRepository) Update(ctx context.Context, holdingCode string, guid string, doc models.ProductBarcodeBOMViewDoc) error {
	return nil
}

func (repo *captureBOMRepository) DeleteByGuidfixed(ctx context.Context, holdingCode string, guid string, username string) error {
	return nil
}

func (repo *captureBOMRepository) Delete(ctx context.Context, holdingCode string, username string, filters map[string]interface{}) error {
	return nil
}

func (repo *captureBOMRepository) FindPageFilter(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.ProductBarcodeBOMViewInfo, mongopagination.PaginationData, error) {
	return []models.ProductBarcodeBOMViewInfo{}, mongopagination.PaginationData{}, nil
}

func (repo *captureBOMRepository) FindByGuid(ctx context.Context, holdingCode string, guid string) (models.ProductBarcodeBOMViewDoc, error) {
	return models.ProductBarcodeBOMViewDoc{}, nil
}

func (repo *captureBOMRepository) FindInItemGuid(ctx context.Context, holdingCode string, columnName string, itemGuidList []string) ([]models.ProductBarcodeBOMViewGuid, error) {
	return []models.ProductBarcodeBOMViewGuid{}, nil
}

func (repo *captureBOMRepository) FindByDocIndentityGuid(ctx context.Context, holdingCode string, indentityField string, indentityValue interface{}) (models.ProductBarcodeBOMViewDoc, error) {
	return models.ProductBarcodeBOMViewDoc{}, nil
}

func (repo *captureBOMRepository) FindStep(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, selectFields map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ProductBarcodeBOMViewInfo, int, error) {
	repo.searchStepFields = append([]string(nil), searchInFields...)
	return []models.ProductBarcodeBOMViewInfo{}, 0, nil
}

func (repo *captureBOMRepository) FindDeletedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageable micromodels.Pageable) ([]models.ProductBarcodeBOMViewDeleteActivity, mongopagination.PaginationData, error) {
	return []models.ProductBarcodeBOMViewDeleteActivity{}, mongopagination.PaginationData{}, nil
}

func (repo *captureBOMRepository) FindCreatedOrUpdatedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageable micromodels.Pageable) ([]models.ProductBarcodeBOMViewActivity, mongopagination.PaginationData, error) {
	return []models.ProductBarcodeBOMViewActivity{}, mongopagination.PaginationData{}, nil
}

func (repo *captureBOMRepository) FindDeletedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ProductBarcodeBOMViewDeleteActivity, error) {
	return []models.ProductBarcodeBOMViewDeleteActivity{}, nil
}

func (repo *captureBOMRepository) FindCreatedOrUpdatedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ProductBarcodeBOMViewActivity, error) {
	return []models.ProductBarcodeBOMViewActivity{}, nil
}

func (repo *captureBOMRepository) FindUseBOMByBarcode(ctx context.Context, holdingCode string, barcode string) (models.ProductBarcodeBOMViewDoc, error) {
	return models.ProductBarcodeBOMViewDoc{}, nil
}

func (repo *captureBOMRepository) ClearUseBOMByBarcode(ctx context.Context, holdingCode string, barcode string) error {
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

// graphBOMRepository is an in-memory fake keyed by barcode, letting tests wire up arbitrary
// recipe-reference graphs (including deliberate cycles that bypass save-time prevention, to prove
// resolveRecipeTree never hangs on them) without a real MongoDB.
type graphBOMRepository struct {
	captureBOMRepository
	byBarcode map[string]models.ProductBarcodeBOMViewDoc
}

func newGraphBOMRepository() *graphBOMRepository {
	return &graphBOMRepository{byBarcode: map[string]models.ProductBarcodeBOMViewDoc{}}
}

// put registers a recipe barcode with the given sub-items (reftype "recipe" or "product").
func (repo *graphBOMRepository) put(barcode string, outputQty float64, items ...models.ProductBarcodeBOMView) {
	itemsCopy := append([]models.ProductBarcodeBOMView(nil), items...)
	doc := models.ProductBarcodeBOMViewDoc{}
	doc.GuidFixed = "guid-" + barcode
	doc.Barcode = barcode
	doc.Qty = outputQty
	doc.BOM = &itemsCopy
	repo.byBarcode[barcode] = doc
}

func recipeItem(barcode string, qty float64) models.ProductBarcodeBOMView {
	item := models.ProductBarcodeBOMView{}
	item.Barcode = barcode
	item.RefType = "recipe"
	item.Qty = qty
	item.YieldPercent = 100
	return item
}

func productItem(barcode string, qty float64, cost float64) models.ProductBarcodeBOMView {
	item := models.ProductBarcodeBOMView{}
	item.Barcode = barcode
	item.RefType = "product"
	item.Qty = qty
	item.YieldPercent = 100
	item.AverageCost = cost
	return item
}

func (repo *graphBOMRepository) FindUseBOMByBarcode(ctx context.Context, holdingCode string, barcode string) (models.ProductBarcodeBOMViewDoc, error) {
	doc, ok := repo.byBarcode[barcode]
	if !ok {
		return models.ProductBarcodeBOMViewDoc{}, nil
	}
	return doc, nil
}

func newTestBOMService(repo *graphBOMRepository) BOMHttpService {
	return BOMHttpService{repo: repo, contextTimeout: 5 * time.Second}
}

func TestCheckNoCycleAllowsDiamondDependency(t *testing.T) {
	// A -> B, A -> C, B -> D, C -> D is NOT a cycle (D reachable via two paths) — must be allowed.
	repo := newGraphBOMRepository()
	repo.put("D", 1, productItem("SALT", 1, 5))
	repo.put("B", 1, recipeItem("D", 1))
	repo.put("C", 1, recipeItem("D", 1))
	svc := newTestBOMService(repo)

	ctx := context.Background()
	if err := svc.checkNoCycle(ctx, "test", "B", "A", map[string]bool{"A": true}, 0); err != nil {
		t.Fatalf("checkNoCycle rejected a legitimate diamond dependency (B): %v", err)
	}
	if err := svc.checkNoCycle(ctx, "test", "C", "A", map[string]bool{"A": true}, 0); err != nil {
		t.Fatalf("checkNoCycle rejected a legitimate diamond dependency (C): %v", err)
	}
}

func TestCheckNoCycleRejectsDirectCycle(t *testing.T) {
	// Saving A with B as an ingredient, where B already references A -- direct 2-node cycle.
	repo := newGraphBOMRepository()
	repo.put("B", 1, recipeItem("A", 1))
	svc := newTestBOMService(repo)

	err := svc.checkNoCycle(context.Background(), "test", "B", "A", map[string]bool{"A": true}, 0)
	if err == nil {
		t.Fatal("checkNoCycle did not reject a direct A<->B cycle")
	}
}

func TestCheckNoCycleRejectsTransitiveCycle(t *testing.T) {
	// Saving A with B as an ingredient, where B -> C -> A -- a 3-node transitive cycle, the case
	// a naive "does barcode == recipeCode" self-reference check alone would miss.
	repo := newGraphBOMRepository()
	repo.put("C", 1, recipeItem("A", 1))
	repo.put("B", 1, recipeItem("C", 1))
	svc := newTestBOMService(repo)

	err := svc.checkNoCycle(context.Background(), "test", "B", "A", map[string]bool{"A": true}, 0)
	if err == nil {
		t.Fatal("checkNoCycle did not reject a transitive A->B->C->A cycle")
	}
}

func TestCheckNoCycleTerminatesOnPreexistingCycleWithoutHanging(t *testing.T) {
	// Defense-in-depth: if a cycle somehow already exists in stored data (e.g. the documented
	// TOCTOU race between two concurrent saves), checkNoCycle must still terminate quickly with an
	// error instead of recursing forever, when a THIRD recipe is saved referencing into that cycle.
	repo := newGraphBOMRepository()
	repo.put("X", 1, recipeItem("Y", 1))
	repo.put("Y", 1, recipeItem("X", 1)) // X <-> Y already cyclic, bypassing save-time checks

	svc := newTestBOMService(repo)
	done := make(chan error, 1)
	go func() {
		done <- svc.checkNoCycle(context.Background(), "test", "X", "Z", map[string]bool{"Z": true}, 0)
	}()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("checkNoCycle returned nil walking into a pre-existing cycle; expected a depth/cycle error")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("checkNoCycle hung (>3s) walking into a pre-existing cycle instead of terminating via the depth cap")
	}
}

func TestResolveRecipeTreeLiveCascadeAcrossTwoParents(t *testing.T) {
	// The actual feature this task is about: editing a shared sub-recipe (SAUCE) must be reflected
	// live in every parent that references it, without resaving those parents.
	repo := newGraphBOMRepository()
	repo.put("SAUCE", 1, productItem("LIME", 0.2, 60))
	repo.put("PADTHAI", 4, recipeItem("SAUCE", 0.05))
	repo.put("SOMTAM", 2, recipeItem("SAUCE", 0.04))
	svc := newTestBOMService(repo)
	ctx := context.Background()

	resolveParent := func(barcode string) []models.ProductBarcodeBOMView {
		doc := repo.byBarcode[barcode]
		resolved, err := svc.resolveRecipeTree(ctx, "test", *doc.BOM, map[string]bool{barcode: true}, 0)
		if err != nil {
			t.Fatalf("resolveRecipeTree(%s) error: %v", barcode, err)
		}
		return resolved
	}

	before := resolveParent("PADTHAI")
	if got := (*before[0].BOM)[0].AverageCost; got != 60 {
		t.Fatalf("PADTHAI pre-edit resolved lime cost = %v, want 60", got)
	}
	if got := before[0].SubRecipeOutputQty; got != 1 {
		t.Fatalf("PADTHAI resolved subrecipeoutputqty = %v, want 1", got)
	}

	// Edit ONLY the shared sub-recipe -- neither parent is touched.
	repo.put("SAUCE", 1, productItem("LIME", 0.35, 75))

	for _, parent := range []string{"PADTHAI", "SOMTAM"} {
		resolved := resolveParent(parent)
		got := (*resolved[0].BOM)[0].AverageCost
		if got != 75 {
			t.Fatalf("%s post-edit resolved lime cost = %v, want 75 (live cascade did not propagate)", parent, got)
		}
	}
}

func TestResolveRecipeTreeTerminatesOnPreexistingCycleWithoutHanging(t *testing.T) {
	// Same defense-in-depth property as the checkNoCycle test above, but for the READ path: if a
	// cycle ever exists in stored data, InfoBOM must return an error, never hang.
	repo := newGraphBOMRepository()
	repo.put("X", 1, recipeItem("Y", 1))
	repo.put("Y", 1, recipeItem("X", 1))
	svc := newTestBOMService(repo)

	done := make(chan error, 1)
	go func() {
		_, err := svc.resolveRecipeTree(context.Background(), "test", *repo.byBarcode["X"].BOM, map[string]bool{"X": true}, 0)
		done <- err
	}()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("resolveRecipeTree returned nil resolving a pre-existing cycle; expected a depth/cycle error")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("resolveRecipeTree hung (>3s) resolving a pre-existing cycle instead of terminating via the depth cap")
	}
}

func TestSaveRecipeBOMRejectsScrapPercentAtOrAbove100(t *testing.T) {
	// ScrapPercent >= 100 would divide-by-zero or go negative in the cost-per-unit formula
	// (cost / (1 - scrap/100)) — must be rejected before it ever reaches storage.
	repo := newGraphBOMRepository()
	svc := newTestBOMService(repo)

	nameCode := "th"
	nameVal := "ทดสอบ"
	req := models.ProductBarcodeBOMSaveRequest{
		Barcode:      "RCPTEST",
		Names:        &[]common_models.NameX{{Code: &nameCode, Name: &nameVal}},
		OutputQty:    2,
		ScrapPercent: 100,
	}

	_, err := svc.SaveRecipeBOM("test", "tester", "", req)
	if err == nil {
		t.Fatal("SaveRecipeBOM accepted scrappercent=100; expected a validation error")
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
