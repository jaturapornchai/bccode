//go:build integration

package company

import (
	"context"
	"testing"
	"time"

	"smlcloudplatform/internal/centraldb/centraldbtest"
	common "smlcloudplatform/internal/models"
	orgaccess "smlcloudplatform/internal/organization"
	companyModels "smlcloudplatform/internal/organization/company/models"
	"smlcloudplatform/internal/taxaddress"
)

// ที่อยู่สำหรับภาษีบนตาราง companies จริง: INSERT/SELECT/UPDATE ชุดเดียวกับ handler — PUT ไม่ส่งที่อยู่ต้องคงค่าเดิม, ส่งมาแทนทั้งชุด
func TestCompanyTaxAddressAgainstCentralSchema(t *testing.T) {
	db := centraldbtest.New(t)
	ctx := context.Background()
	centraldbtest.Exec(t, db, `INSERT INTO holdings (code, name) VALUES ('rungrueng','กลุ่มกิจการรุ่งเรืองกรุ๊ป')`)

	th, name := "th", "บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด"
	names, err := orgaccess.EncodeNames(common.JSONB{{Code: &th, Name: &name}})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	created := taxaddress.Address{No: "88/12", Road: "ลาดหลุมแก้ว", Subdistrict: "คูบางหลวง", District: "ลาดหลุมแก้ว", Province: "ปทุมธานี", Postcode: "12140"}
	args := append([]interface{}{"rungrueng", "01", name, names, "0105558012349", "", "owner@example.com", now},
		companyTaxAddressArgs(created, "02-123-4567")...)
	if _, err := db.ExecContext(ctx, insertCompanySQL, args...); err != nil {
		t.Fatalf("insert: %v", err)
	}
	// บริษัทที่สร้างก่อนมีคอลัมน์ที่อยู่ (ไม่ระบุคอลัมน์) ต้องได้ค่าว่าง ไม่ใช่ NULL
	centraldbtest.Exec(t, db, `INSERT INTO companies (holding_code, code, name) VALUES ('rungrueng','02','บริษัท หอมกรุ่น คอฟฟี่ แอนด์ เบเกอรี่ จำกัด')`)
	blank, err := findCompany(ctx, db, "rungrueng", "02", false)
	if err != nil || blank.Address == nil || !blank.Address.IsBlank() || blank.Phone == nil || *blank.Phone != "" {
		t.Fatalf("legacy row: %+v err=%v", blank, err)
	}

	existing, err := findCompany(ctx, db, "rungrueng", "01", false)
	if err != nil {
		t.Fatal(err)
	}
	if existing.Address == nil || *existing.Address != created || existing.Phone == nil || *existing.Phone != "02-123-4567" {
		t.Fatalf("scan: %+v phone %v", existing.Address, existing.Phone)
	}

	update := func(req companyModels.CompanyDoc) companyModels.CompanyDoc {
		t.Helper()
		current, err := findCompany(ctx, db, "rungrueng", "01", false)
		if err != nil {
			t.Fatal(err)
		}
		address, phone := mergeCompanyTaxAddress(current, req)
		args := append([]interface{}{"rungrueng", "01", name, names, "0105558012349", "", true, "owner@example.com", time.Now().UTC(), true},
			companyTaxAddressArgs(address, phone)...)
		if result, err := db.ExecContext(ctx, updateCompanySQL, args...); err != nil {
			t.Fatalf("update: %v", err)
		} else if n, _ := result.RowsAffected(); n != 1 {
			t.Fatalf("update rows %d", n)
		}
		after, err := findCompany(ctx, db, "rungrueng", "01", false)
		if err != nil {
			t.Fatal(err)
		}
		return after
	}

	// client เก่าไม่ส่ง address/phone → คงค่าเดิม
	if after := update(companyModels.CompanyDoc{}); *after.Address != created || *after.Phone != "02-123-4567" {
		t.Fatalf("nil address must keep existing: %+v %q", after.Address, *after.Phone)
	}
	// ส่งมา → แทนทั้งชุด (ช่องที่ไม่ส่งกลายเป็นว่าง)
	replaced := taxaddress.Address{Building: "สาทรซิตี้ทาวเวอร์", No: "175", Road: "สาทรใต้", Province: "กรุงเทพมหานคร", Postcode: "10120"}
	emptyPhone := ""
	req := companyModels.CompanyDoc{}
	req.Address, req.Phone = &replaced, &emptyPhone
	if after := update(req); *after.Address != replaced || *after.Phone != "" {
		t.Fatalf("sent address must replace: %+v %q", after.Address, *after.Phone)
	}
}
