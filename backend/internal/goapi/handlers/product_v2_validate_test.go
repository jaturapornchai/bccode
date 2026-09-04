package handlers

import (
	"strings"
	"testing"

	productmodels "smlcloudplatform/internal/product/product/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func codesOf(errs []productV2FieldError) string {
	codes := make([]string, 0, len(errs))
	for _, e := range errs {
		codes = append(codes, e.Field+":"+e.Code)
	}
	return strings.Join(codes, " ")
}

func dec(t *testing.T, s string) productmodels.ListingPrice {
	t.Helper()
	d, err := primitive.ParseDecimal128(s)
	if err != nil {
		t.Fatal(err)
	}
	return productmodels.ListingPrice{Decimal128: d}
}

func TestValidateListingTitle_E01(t *testing.T) {
	limits := productV2DefaultLimits
	cases := []struct{ title, want string }{
		{"abcd", "listing.title:TOO_SHORT"},
		{"abcde", ""},
		{strings.Repeat("ก", 120), ""},
		{strings.Repeat("ก", 121), "listing.title:TOO_LONG"},
		{"\aabcd", "listing.title:INVALID_CHAR"},
		{"  abc  ", "listing.title:TOO_SHORT"},
	}
	for _, tc := range cases {
		if got := codesOf(validateListingTitle(tc.title, limits)); got != tc.want {
			t.Errorf("title %q: got %q want %q", tc.title, got, tc.want)
		}
	}
}

func TestValidateDescriptions_E02(t *testing.T) {
	limits := productV2DefaultLimits
	if got := codesOf(validateListingDescription(strings.Repeat("ก", 2000), limits)); got != "" {
		t.Errorf("2000 ok, got %q", got)
	}
	if got := codesOf(validateListingDescription(strings.Repeat("ก", 2001), limits)); got != "listing.description:TOO_LONG" {
		t.Errorf("2001: got %q", got)
	}
	limits.ListingDescriptionMin = 10
	if got := codesOf(validateListingDescription("สั้น", limits)); got != "listing.description:TOO_SHORT" {
		t.Errorf("min: got %q", got)
	}
	if got := codesOf(validateProductDescription(strings.Repeat("ก", 1500), limits)); got != "" {
		t.Errorf("1500 ok, got %q", got)
	}
	if got := codesOf(validateProductDescription(strings.Repeat("ก", 1501), limits)); got != "description:TOO_LONG" {
		t.Errorf("1501: got %q", got)
	}
}

func TestValidatePackage_E05(t *testing.T) {
	one := ptrFloat(1)
	cases := []struct {
		name string
		pkg  productV2PackagePatch
		want string
	}{
		{"none", productV2PackagePatch{}, ""},
		{"all", productV2PackagePatch{one, one, one, one}, ""},
		{"three of four", productV2PackagePatch{one, one, one, nil}, "package:PACKAGE_INCOMPLETE"},
		{"zero", productV2PackagePatch{ptrFloat(0), one, one, one}, "package.weight:INVALID_VALUE"},
		{"negative", productV2PackagePatch{one, one, ptrFloat(-2), one}, "package.width:INVALID_VALUE"},
	}
	for _, tc := range cases {
		if got := codesOf(validatePackage(tc.pkg, productV2DefaultLimits)); got != tc.want {
			t.Errorf("%s: got %q want %q", tc.name, got, tc.want)
		}
	}
}

func TestValidateWholesale_E07(t *testing.T) {
	w := func(min, max int, price string) productmodels.ProductListingWholesale {
		return productmodels.ProductListingWholesale{MinCount: min, MaxCount: max, UnitPrice: dec(t, price)}
	}
	cases := []struct {
		name   string
		tiers  []productmodels.ProductListingWholesale
		normal float64
		want   string
	}{
		{"valid", []productmodels.ProductListingWholesale{w(10, 49, "179.00"), w(50, 99, "169.00")}, 199, ""},
		{"min>=max", []productmodels.ProductListingWholesale{w(10, 10, "179.00")}, 199, "listing.wholesale[0]:INVALID_VALUE"},
		{"overlap", []productmodels.ProductListingWholesale{w(10, 49, "179.00"), w(40, 99, "169.00")}, 199, "listing.wholesale[1]:RANGE_OVERLAP"},
		{"overlap unsorted input", []productmodels.ProductListingWholesale{w(40, 99, "169.00"), w(10, 49, "179.00")}, 199, "listing.wholesale[0]:RANGE_OVERLAP"},
		{"above normal", []productmodels.ProductListingWholesale{w(10, 49, "199.01")}, 199, "listing.wholesale[0].unitprice:OUT_OF_RANGE"},
		{"equal normal ok", []productmodels.ProductListingWholesale{w(10, 49, "199.00")}, 199, ""},
		{"no normal price skips compare", []productmodels.ProductListingWholesale{w(10, 49, "999.00")}, 0, ""},
		{"negative", []productmodels.ProductListingWholesale{w(10, 49, "-1")}, 199, "listing.wholesale[0].unitprice:INVALID_VALUE"},
	}
	for _, tc := range cases {
		if got := codesOf(validateWholesale(tc.tiers, tc.normal)); got != tc.want {
			t.Errorf("%s: got %q want %q", tc.name, got, tc.want)
		}
	}
	if productV2NormalPrice([]float64{250, 199, 0, 300}) != 199 {
		t.Errorf("normal price should be lowest positive model price")
	}
}

func TestValidatePurchaseLimitAndPreorder_E08(t *testing.T) {
	pl := func(min, max int) *productmodels.ProductListingPurchaseLimit {
		return &productmodels.ProductListingPurchaseLimit{Min: min, Max: max}
	}
	po := func(is bool, days int) *productmodels.ProductListingPreorder {
		return &productmodels.ProductListingPreorder{IsPreorder: is, DaysToShip: days}
	}
	cases := []struct {
		name  string
		limit *productmodels.ProductListingPurchaseLimit
		pre   *productmodels.ProductListingPreorder
		want  string
	}{
		{"ok", pl(1, 10), po(true, 7), ""},
		{"min>max", pl(11, 10), nil, "listing.purchaselimit:OUT_OF_RANGE"},
		{"max 0 = unlimited", pl(5, 0), nil, ""},
		{"negative", pl(-1, 5), nil, "listing.purchaselimit:INVALID_VALUE"},
		{"days 0", nil, po(true, 0), "listing.preorder.daystoship:INVALID_VALUE"},
		{"days 1", nil, po(true, 1), ""},
		{"days 31 (no category limits yet)", nil, po(true, 31), ""},
		{"not preorder ignores days", nil, po(false, 99), ""},
	}
	for _, tc := range cases {
		if got := codesOf(validatePurchaseLimitAndPreorder(tc.limit, tc.pre)); got != tc.want {
			t.Errorf("%s: got %q want %q", tc.name, got, tc.want)
		}
	}
}

func TestValidateCondition_E20(t *testing.T) {
	for _, ok := range []string{"", "NEW", "USED"} {
		if got := validateCondition(ok); len(got) != 0 {
			t.Errorf("%q should pass: %v", ok, got)
		}
	}
	for _, bad := range []string{"new", "Used", "REFURBISHED"} {
		if got := codesOf(validateCondition(bad)); got != "listing.condition:INVALID_VALUE" {
			t.Errorf("%q: got %q", bad, got)
		}
	}
}

func TestValidateImages(t *testing.T) {
	limits := productV2DefaultLimits
	images := make([]productmodels.ProductImage, 10)
	for i := range images {
		images[i] = productmodels.ProductImage{XOrder: i, URI: "/x", URIThumb: "/x_t"}
	}
	if got := codesOf(validateImages(images, limits)); got != "images:TOO_LONG" {
		t.Errorf("10 images: got %q", got)
	}
	dup := []productmodels.ProductImage{{XOrder: 0, URI: "/a", URIThumb: "/a_t"}, {XOrder: 0, URI: "/b", URIThumb: "/b_t"}}
	if got := codesOf(validateImages(dup, limits)); got != "images[1].xorder:DUPLICATE" {
		t.Errorf("dup: got %q", got)
	}
	noThumb := []productmodels.ProductImage{{XOrder: 0, URI: "/a"}}
	if got := codesOf(validateImages(noThumb, limits)); got != "images[0].urithumb:REQUIRED" {
		t.Errorf("no thumb: got %q", got)
	}
	badURI := []productmodels.ProductImage{{XOrder: 0, URI: "javascript:alert(1)", URIThumb: "data:image/png;base64,AA"}}
	if got := codesOf(validateImages(badURI, limits)); got != "images[0].uri:INVALID_VALUE images[0].urithumb:INVALID_VALUE" {
		t.Errorf("bad uri: got %q", got)
	}
}

func TestValidateURIAndVideos(t *testing.T) {
	limits := productV2DefaultLimits
	cases := []struct{ uri, want string }{
		{"", ""},
		{"/goapi/s3/file/a.jpg", ""},
		{"https://cdn.example.com/a.jpg", ""},
		{"HTTP://cdn.example.com/a.jpg", ""},
		{"ftp://x/a.jpg", "f:INVALID_VALUE"},
		{"/a b.jpg", "f:INVALID_CHAR"},
		{"/" + strings.Repeat("a", limits.URIMax), "f:TOO_LONG"},
	}
	for _, tc := range cases {
		if got := codesOf(validateURI("f", tc.uri, false, limits)); got != tc.want {
			t.Errorf("%q: got %q want %q", tc.uri, got, tc.want)
		}
	}
	if got := codesOf(validateURI("f", "", true, limits)); got != "f:REQUIRED" {
		t.Errorf("required: got %q", got)
	}
	videos := []productmodels.ProductVideo{{XOrder: 0, URI: "/v1"}, {XOrder: 0, URI: "", DurationSec: -1}}
	if got := codesOf(validateVideos(videos, limits)); got != "videos:TOO_LONG videos[1].xorder:DUPLICATE videos[1].uri:REQUIRED videos[1]:INVALID_VALUE" {
		t.Errorf("videos: got %q", got)
	}
	req := productV2UpdateListingRequest{
		Listing:       &productV2ListingPatch{Wholesale: &[]productmodels.ProductListingWholesale{}, SizeChart: &productmodels.ProductListingSizeChart{URI: "data:x"}},
		ImageURI:      ptrString("javascript:x"),
		ImageURIThumb: ptrString("/ok_t"),
	}
	many := make([]productmodels.ProductListingWholesale, limits.WholesaleCountMax+1)
	for i := range many {
		many[i] = productmodels.ProductListingWholesale{MinCount: i*10 + 1, MaxCount: i*10 + 10, UnitPrice: dec(t, "1.00")}
	}
	req.Listing.Wholesale = &many
	got := codesOf(validateProductV2Update(req, nil, limits))
	if got != "listing.wholesale:TOO_LONG listing.sizechart.uri:INVALID_VALUE imageuri:INVALID_VALUE" {
		t.Errorf("aggregate: got %q", got)
	}
}

func TestValidateProductV2Update_Aggregates(t *testing.T) {
	req := productV2UpdateListingRequest{
		Listing: &productV2ListingPatch{Title: ptrString("abc"), Condition: ptrString("new")},
		Package: &productV2PackagePatch{Weight: ptrFloat(1)},
	}
	got := codesOf(validateProductV2Update(req, nil, productV2DefaultLimits))
	if got != "listing.title:TOO_SHORT listing.condition:INVALID_VALUE package:PACKAGE_INCOMPLETE" {
		t.Fatalf("got %q", got)
	}
	huge := productV2PackagePatch{Weight: ptrFloat(1e308), Length: ptrFloat(1001), Width: ptrFloat(1), Height: ptrFloat(1)}
	if got := codesOf(validatePackage(huge, productV2DefaultLimits)); got != "package.weight:OUT_OF_RANGE package.length:OUT_OF_RANGE" {
		t.Fatalf("package upper bound: got %q", got)
	}
}
