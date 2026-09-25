//go:build integration

package generalledger_test

import (
	"database/sql"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/lib/pq"
	gl "smlcloudplatform/internal/generalledger"
)

// uatCentralFixtureSQL เตรียมตารางกลางด้าน tenancy ของการติดตั้งใหม่
// โครงตารางตรงกับ backend/internal/database/fresh_provision_all.sql หมวดหมวด 1
// โดยเจาะจงเฉพาะข้อมูลที่ชุดทดสอบ UAT ใช้จริง (Holding THAI_HOLDING บริษัท 01/02)
const uatCentralFixtureSQL = `
CREATE TABLE IF NOT EXISTS holdings (
    code TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    tax_id TEXT DEFAULT '',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS companies (
    holding_code TEXT NOT NULL REFERENCES holdings(code) ON DELETE CASCADE,
    code TEXT NOT NULL,
    name TEXT NOT NULL,
    tax_id TEXT DEFAULT '',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    -- ที่อยู่สำหรับภาษี (สำนักงานใหญ่) ตาม key หัวแบบ rdform + โทรศัพท์ (internal/taxaddress)
    addr_building TEXT NOT NULL DEFAULT '',
    addr_room TEXT NOT NULL DEFAULT '',
    addr_floor TEXT NOT NULL DEFAULT '',
    addr_village TEXT NOT NULL DEFAULT '',
    addr_no TEXT NOT NULL DEFAULT '',
    addr_moo TEXT NOT NULL DEFAULT '',
    addr_soi TEXT NOT NULL DEFAULT '',
    addr_junction TEXT NOT NULL DEFAULT '',
    addr_road TEXT NOT NULL DEFAULT '',
    addr_subdistrict TEXT NOT NULL DEFAULT '',
    addr_district TEXT NOT NULL DEFAULT '',
    addr_province TEXT NOT NULL DEFAULT '',
    addr_postcode TEXT NOT NULL DEFAULT '',
    phone TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (holding_code, code)
);
CREATE TABLE IF NOT EXISTS branches (
    holding_code TEXT NOT NULL,
    company_code TEXT NOT NULL,
    code TEXT NOT NULL,
    name TEXT NOT NULL,
    is_headquarters BOOLEAN NOT NULL DEFAULT false,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (holding_code, company_code, code),
    FOREIGN KEY (holding_code, company_code) REFERENCES companies(holding_code, code) ON DELETE CASCADE
);
INSERT INTO holdings (code, name, tax_id, is_active) VALUES
    ('THAI_HOLDING', 'บริษัท สยามพาณิชย์ กรุ๊ป จำกัด (มหาชน)', '0107558000123', true)
ON CONFLICT (code) DO NOTHING;
INSERT INTO companies (holding_code, code, name, tax_id, is_active) VALUES
    ('THAI_HOLDING', '01', 'บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด', '0107558000123', true),
    ('THAI_HOLDING', '02', 'บริษัท รุ่งเรืองขนส่งและโลจิสติกส์ จำกัด', '0107558000456', true)
ON CONFLICT (holding_code, code) DO NOTHING;
INSERT INTO branches (holding_code, company_code, code, name, is_headquarters, is_active) VALUES
    ('THAI_HOLDING', '01', '00000', 'สำนักงานใหญ่', true, true),
    ('THAI_HOLDING', '02', '00000', 'สำนักงานใหญ่', true, true),
    ('THAI_HOLDING', '02', '00001', 'สาขาลาดหลุมแก้ว', false, true)
ON CONFLICT (holding_code, company_code, code) DO NOTHING;
`

// uatFreshInstallDB เตรียมฐานติดตั้งใหม่สำหรับ UAT: สร้าง schema ชั่วคราว โหลด
// GL runtime schema (schema.sql + subledger.sql) และตารางกลาง tenancy แล้วคืน
// projection ที่ชี้ฐานนั้น ทุกอย่างถูกลบทิ้งเมื่อจบการทดสอบ (t.Cleanup)
func uatFreshInstallDB(t *testing.T) (*gl.Postgres, *sql.DB) {
	t.Helper()
	dsn := os.Getenv("BC_GL_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("set BC_GL_TEST_POSTGRES_DSN to isolated PostgreSQL")
	}
	admin, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("เชื่อมต่อ PostgreSQL ทดสอบไม่ได้: %v", err)
	}
	schema := "bc_gl_uat_" + strings.ReplaceAll(uuid.NewString(), "-", "")[:20]
	if _, err := admin.Exec(`CREATE SCHEMA ` + pq.QuoteIdentifier(schema)); err != nil {
		admin.Close()
		t.Fatalf("สร้าง schema ทดสอบไม่ได้: %v", err)
	}
	t.Cleanup(func() {
		_, err := admin.Exec(`DROP SCHEMA ` + pq.QuoteIdentifier(schema) + ` CASCADE`)
		if err != nil {
			t.Errorf("ลบ schema ทดสอบไม่สำเร็จ: %v", err)
		}
		admin.Close()
	})
	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	query := u.Query()
	query.Set("search_path", schema)
	u.RawQuery = query.Encode()
	db, err := sql.Open("postgres", u.String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	for _, file := range []string{"schema.sql", "subledger.sql", "budget.sql"} {
		raw, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("อ่าน %s ไม่ได้ (คาดหวังรันจาก package dir): %v", file, err)
		}
		if _, err := db.Exec(string(raw)); err != nil {
			t.Fatalf("โหลด %s ไม่สำเร็จ: %v", file, err)
		}
	}
	if _, err := db.Exec(uatCentralFixtureSQL); err != nil {
		t.Fatalf("โหลด fixture ตารางกลางไม่สำเร็จ: %v", err)
	}
	projection := gl.NewPostgres(func(holding string) (*sql.DB, error) {
		return db, nil
	})
	return projection, db
}
