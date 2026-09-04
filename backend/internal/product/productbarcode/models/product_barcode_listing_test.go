package models

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

// ฟิลด์ชั้นลงขายต้องหายจาก bson เมื่อว่าง — กันการบันทึกฝั่งบัญชี ($set ทั้ง struct) ลบค่าทิ้ง
func TestProductBarcodeBaseBsonOmitsListingWhenNil(t *testing.T) {
	raw, err := bson.Marshal(ProductBarcodeBase{})
	if err != nil {
		t.Fatal(err)
	}
	var m bson.M
	if err := bson.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"listing", "imageurithumb"} {
		if _, ok := m[k]; ok {
			t.Errorf("ProductBarcodeBase{} bson must omit %q", k)
		}
	}
	if _, ok := bsonMap(t, ProductImage{})["urithumb"]; ok {
		t.Error("ProductImage{} bson must omit urithumb")
	}
}

func bsonMap(t *testing.T, v any) bson.M {
	t.Helper()
	raw, err := bson.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	var m bson.M
	if err := bson.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	return m
}
