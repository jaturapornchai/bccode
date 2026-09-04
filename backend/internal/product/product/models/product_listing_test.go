package models

import (
	"encoding/json"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func bsonKeys(t *testing.T, v any) map[string]bool {
	t.Helper()
	raw, err := bson.Marshal(v)
	if err != nil {
		t.Fatalf("bson.Marshal: %v", err)
	}
	var m bson.M
	if err := bson.Unmarshal(raw, &m); err != nil {
		t.Fatalf("bson.Unmarshal: %v", err)
	}
	keys := map[string]bool{}
	for k := range m {
		keys[k] = true
	}
	return keys
}

// ฟิลด์ชั้นลงขายต้องหายจาก bson เมื่อว่าง — กันการบันทึกฝั่งบัญชี ($set ทั้ง struct) ลบค่าทิ้ง
func TestProductBsonOmitsListingWhenNil(t *testing.T) {
	keys := bsonKeys(t, Product{})
	for _, k := range []string{"listing", "imageurithumb"} {
		if keys[k] {
			t.Errorf("Product{} bson must omit %q", k)
		}
	}
	if bsonKeys(t, ProductImage{})["urithumb"] {
		t.Error("ProductImage{} bson must omit urithumb")
	}
	vk := bsonKeys(t, ProductVideo{})
	if vk["durationsec"] || vk["sizebytes"] {
		t.Error("ProductVideo{} bson must omit durationsec/sizebytes")
	}
}

func TestWholesaleUnitPriceJSONRoundTrip(t *testing.T) {
	d, err := primitive.ParseDecimal128("179.00")
	if err != nil {
		t.Fatal(err)
	}
	out, err := json.Marshal(ProductListingWholesale{MinCount: 1, MaxCount: 9, UnitPrice: ListingPrice{Decimal128: d}})
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != `{"mincount":1,"maxcount":9,"unitprice":"179.00"}` {
		t.Fatalf("unexpected json: %s", out)
	}
	var back ProductListingWholesale
	if err := json.Unmarshal(out, &back); err != nil {
		t.Fatal(err)
	}
	if back.UnitPrice.String() != "179.00" {
		t.Fatalf("round trip got %q", back.UnitPrice.String())
	}
}

func TestListingPriceAcceptsNumberAndString(t *testing.T) {
	for _, in := range []string{`{"unitprice":199.5}`, `{"unitprice":"199.5"}`, `{"unitprice":" 199.5 "}`} {
		var w ProductListingWholesale
		if err := json.Unmarshal([]byte(in), &w); err != nil {
			t.Fatalf("%s: %v", in, err)
		}
		if w.UnitPrice.String() != "199.5" {
			t.Fatalf("%s: got %q", in, w.UnitPrice.String())
		}
	}
	var bad ProductListingWholesale
	if err := json.Unmarshal([]byte(`{"unitprice":"abc"}`), &bad); err == nil {
		t.Fatal("expected error for non-numeric price")
	}
}

func TestListingPriceBSONIsDecimal128(t *testing.T) {
	d, _ := primitive.ParseDecimal128("179.00")
	raw, err := bson.Marshal(ProductListingWholesale{MinCount: 1, MaxCount: 9, UnitPrice: ListingPrice{Decimal128: d}})
	if err != nil {
		t.Fatal(err)
	}
	var m bson.M
	if err := bson.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	if v, ok := m["unitprice"].(primitive.Decimal128); !ok || v.String() != "179.00" {
		t.Fatalf("unitprice stored as %T %v, want Decimal128 179.00", m["unitprice"], m["unitprice"])
	}
	var back ProductListingWholesale
	if err := bson.Unmarshal(raw, &back); err != nil || back.UnitPrice.String() != "179.00" {
		t.Fatalf("bson round trip: %v %q", err, back.UnitPrice.String())
	}
}
