package handlers

import (
	"net/http"
	"strings"
	"testing"
	"time"

	productmodels "smlcloudplatform/internal/product/product/models"

	"go.mongodb.org/mongo-driver/bson"
)

func ptrString(s string) *string  { return &s }
func ptrFloat(f float64) *float64 { return &f }
func setKeys(set bson.M) []string {
	keys := make([]string, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	return keys
}

func TestBuildProductV2Set_OnlyWhitelistedKeys(t *testing.T) {
	allowed := map[string]struct{}{
		"listing": {}, "packageweight": {}, "packagelength": {}, "packagewidth": {}, "packageheight": {},
		"description": {}, "imageuri": {}, "imageurithumb": {}, "images": {}, "videos": {}, "updatedat": {}, "updatedby": {},
	}
	req := productV2UpdateListingRequest{
		Listing:       &productV2ListingPatch{Title: ptrString("เสื้อยืดคอกลม")},
		Package:       &productV2PackagePatch{Weight: ptrFloat(0.25), Length: ptrFloat(30), Width: ptrFloat(20), Height: ptrFloat(3)},
		Description:   ptrString("d"),
		ImageURI:      ptrString("/a"),
		ImageURIThumb: ptrString("/a_t"),
		Images:        &[]productmodels.ProductImage{{XOrder: 0, URI: "/a", URIThumb: "/a_t"}},
		Videos:        &[]productmodels.ProductVideo{{XOrder: 0, URI: "/v", DurationSec: 10, SizeBytes: 100}},
	}
	set := buildProductV2Set(productmodels.ProductDoc{}, req, "tester", time.Now())
	for _, key := range setKeys(set) {
		if _, ok := allowed[key]; !ok {
			t.Errorf("unexpected key %q in $set", key)
		}
		if strings.HasPrefix(key, "listing.") || strings.Contains(key, "$") {
			t.Errorf("dotted/operator key %q in $set", key)
		}
	}
	if len(set) != len(allowed) {
		t.Fatalf("expected %d keys got %v", len(allowed), setKeys(set))
	}
	if set["updatedby"] != "tester" {
		t.Fatalf("updatedby = %v", set["updatedby"])
	}
	if _, ok := set["images"].([]productmodels.ProductImage); !ok {
		t.Fatalf("images should be typed slice, got %T", set["images"])
	}
	if _, ok := set["listing"].(productmodels.ProductListing); !ok {
		t.Fatalf("listing should be whole struct, got %T", set["listing"])
	}
}

func TestBuildProductV2Set_PackageHandling(t *testing.T) {
	set := buildProductV2Set(productmodels.ProductDoc{}, productV2UpdateListingRequest{Description: ptrString("x")}, "u", time.Now())
	for _, key := range []string{"packageweight", "packagelength", "packagewidth", "packageheight", "listing", "images"} {
		if _, ok := set[key]; ok {
			t.Errorf("key %q must be omitted when not sent", key)
		}
	}
	set = buildProductV2Set(productmodels.ProductDoc{}, productV2UpdateListingRequest{Package: &productV2PackagePatch{}}, "u", time.Now())
	for _, key := range []string{"packageweight", "packagelength", "packagewidth", "packageheight"} {
		if set[key] != float64(0) {
			t.Errorf("empty package should clear %q to 0, got %v", key, set[key])
		}
	}
}

func TestMergeListing_KeepsUntouchedFields(t *testing.T) {
	current := &productmodels.ProductListing{Title: "เดิม", Description: "รายละเอียดเดิม", Condition: "USED", Preorder: &productmodels.ProductListingPreorder{IsPreorder: true, DaysToShip: 10}}
	merged := mergeListing(current, &productV2ListingPatch{Condition: ptrString("NEW")})
	if merged.Title != "เดิม" || merged.Description != "รายละเอียดเดิม" || merged.Condition != "NEW" || merged.Preorder == nil || merged.Preorder.DaysToShip != 10 {
		t.Fatalf("merged = %+v", merged)
	}
	if merged.Wholesale == nil || merged.Tiers == nil || merged.PurchaseLimit == nil {
		t.Fatalf("arrays must be non-nil: %+v", merged)
	}
	merged = mergeListing(nil, &productV2ListingPatch{Preorder: &productmodels.ProductListingPreorder{IsPreorder: false, DaysToShip: 15}})
	if merged.Preorder.DaysToShip != 0 {
		t.Fatalf("daystoship must be zeroed when ispreorder=false: %+v", merged.Preorder)
	}
	empty := mergeListing(nil, nil)
	if empty.Condition != productmodels.ListingConditionNew || empty.Preorder == nil || empty.PurchaseLimit == nil {
		t.Fatalf("empty listing must default condition NEW and carry preorder/purchaselimit objects: %+v", empty)
	}
	doc, err := bson.Marshal(merged)
	if err != nil {
		t.Fatal(err)
	}
	var back bson.M
	if err := bson.Unmarshal(doc, &back); err != nil {
		t.Fatal(err)
	}
	if _, ok := back["wholesale"]; !ok {
		t.Fatalf("bson listing missing wholesale: %v", back)
	}
}

func TestProductV2VersionFilter(t *testing.T) {
	zero := productV2VersionFilter(0)
	or, ok := zero["$or"].(bson.A)
	if !ok || len(or) != 2 {
		t.Fatalf("version 0 filter = %v", zero)
	}
	exists := or[1].(bson.M)["__v"].(bson.M)["$exists"]
	if exists != false {
		t.Fatalf("expected $exists:false, got %v", or[1])
	}
	if productV2VersionFilter(7)["__v"] != int64(7) {
		t.Fatalf("version 7 filter = %v", productV2VersionFilter(7))
	}
}

func TestClassifyProductV2UpdateResult(t *testing.T) {
	if status, code := classifyProductV2UpdateResult(0, false); status != http.StatusNotFound || code != "ITEM_NOT_FOUND" {
		t.Fatalf("got %d %s", status, code)
	}
	if status, code := classifyProductV2UpdateResult(0, true); status != http.StatusConflict || code != "VERSION_CONFLICT" {
		t.Fatalf("got %d %s", status, code)
	}
	if status, code := classifyProductV2UpdateResult(1, true); status != http.StatusOK || code != "" {
		t.Fatalf("got %d %s", status, code)
	}
}

func TestProductV2ListingReadiness(t *testing.T) {
	doc := productmodels.ProductDoc{}
	result := productV2ListingReadiness(doc)
	if result["ready"] != false || len(result["missing"].([]map[string]string)) != 4 {
		t.Fatalf("empty doc readiness = %v", result)
	}
	doc.Listing = &productmodels.ProductListing{Title: "เสื้อยืด"}
	doc.PackageWeight, doc.PackageLength, doc.PackageWidth, doc.PackageHeight = 0.2, 1, 1, 1
	doc.Images = &[]productmodels.ProductImage{{URI: "/a"}}
	if result := productV2ListingReadiness(doc); result["ready"] != true {
		t.Fatalf("complete doc readiness = %v", result)
	}
}
