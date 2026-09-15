package kafka

import (
	"testing"
	"time"

	"smlcloudplatform/internal/goapi/models"
)

// ใบโอนคลังหนึ่งบรรทัดต้องกลายเป็นสองบรรทัดเสมอ ขาออกจากคลังต้นทางและขาเข้าคลังปลายทาง
//
// เคยสร้างแค่บรรทัดเดียว (คลังต้นทาง) ของจึงหายจากบริษัททุกครั้งที่โอนคลัง
func TestStockTransferCreatesBothLegs(t *testing.T) {
	docDate := time.Date(2026, 3, 10, 9, 0, 0, 0, time.UTC)
	processData := models.ProcessMongoTransModel{
		DocNo:       "TF001",
		DocDateTime: docDate,
		Details: []models.ProcessMongoTransDetailModel{{
			ItemCode:       "ITEM01",
			Barcode:        "BC01",
			UnitCode:       "PCS",
			WhCode:         "WH01",
			LocationCode:   "L01",
			ToWhCode:       "WH02",
			ToLocationCode: "L02",
			Qty:            4,
		}},
	}

	lines := MapStockTransferToDocDetailStructs(processData, "HOLD01")
	if len(lines) != 2 {
		t.Fatalf("lines = %d, want 2 (one out of the source warehouse, one into the destination)", len(lines))
	}

	out, in := lines[0], lines[1]
	if out.WhCode != "WH01" || out.LocationCode != "L01" {
		t.Errorf("source leg warehouse = %s/%s, want WH01/L01", out.WhCode, out.LocationCode)
	}
	if out.CalcFlag != -1 || out.TotalQty != -4 {
		t.Errorf("source leg calcflag/qty = %v/%v, want -1/-4", out.CalcFlag, out.TotalQty)
	}
	if in.WhCode != "WH02" || in.LocationCode != "L02" {
		t.Errorf("destination leg warehouse = %s/%s, want WH02/L02", in.WhCode, in.LocationCode)
	}
	if in.CalcFlag != 1 || in.TotalQty != 4 {
		t.Errorf("destination leg calcflag/qty = %v/%v, want 1/4", in.CalcFlag, in.TotalQty)
	}
	if out.LineNumber == in.LineNumber {
		t.Errorf("both legs share line number %d; the ledger key would collide", out.LineNumber)
	}
	if !out.DocDateTime.Equal(docDate) || !in.DocDateTime.Equal(docDate) {
		t.Error("both legs must carry the document date")
	}
}

// ข้อมูลคลังปลายทางต้องรอดจากการแปลงข้อความ Kafka มาเป็นโครงสร้างกลาง
// ถ้าตกหล่นตรงนี้ ขาเข้าจะไม่มีคลังปลายทางให้ลง
func TestStockTransferConversionKeepsDestinationWarehouse(t *testing.T) {
	source := models.StockTransferStruct{
		DocNo: "TF002",
		Details: []models.StockTransferDetailStruct{{
			ItemCode:       "ITEM01",
			WhCode:         "WH01",
			LocationCode:   "L01",
			ToWhCode:       "WH02",
			ToLocationCode: "L02",
			Qty:            2,
		}},
	}

	converted := ConvertStockTransferMongoDocToProcessModel(source)
	if len(converted.Details) != 1 {
		t.Fatalf("details = %d, want 1", len(converted.Details))
	}
	if converted.Details[0].ToWhCode != "WH02" || converted.Details[0].ToLocationCode != "L02" {
		t.Errorf("destination = %s/%s, want WH02/L02",
			converted.Details[0].ToWhCode, converted.Details[0].ToLocationCode)
	}
}
