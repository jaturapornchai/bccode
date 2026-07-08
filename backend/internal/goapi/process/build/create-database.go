package build

import (
	"context"
	"database/sql"
	"fmt"
	"smlcloudplatform/internal/goapi/handlers/approval"
	"smlcloudplatform/internal/goapi/logger"

	"slices"
	"sync"

	"smlcloudplatform/internal/goapi/inventory"
	processdoc "smlcloudplatform/internal/goapi/process/process-doc"
	processstock "smlcloudplatform/internal/goapi/process/process-stock"

	"smlcloudplatform/internal/goapi/myclickhouse"
	"smlcloudplatform/internal/goapi/myglobal"
	"smlcloudplatform/internal/goapi/mypg"
	"smlcloudplatform/internal/goapi/process"
)

// checkedShops — cache holdingcode ที่ผ่าน DatabaseChecker แล้วในรอบ process นี้
// ponytail: process-lifetime cache, กัน DatabaseRebuildAll ทำงานซ้ำทุก Kafka message (hot path)
var checkedShops sync.Map

func TableProductCreate(db *sql.DB) error {
	logger.Info("Creating product table")

	// สร้างทุกอย่างใน transaction เดียว เพื่อความเร็วสูงสุด
	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. สร้าง table ก่อน
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS product (
			id SERIAL PRIMARY KEY,
			itemcode TEXT,		
			name0 TEXT NOT NULL,
			unitcode TEXT NOT NULL,
			unitname TEXT NOT NULL
		)`

	if _, err := tx.ExecContext(context.Background(), createTableQuery); err != nil {
		return fmt.Errorf("create product table: %w", err)
	}

	// 1.5. เพิ่ม columns ใหม่ถ้ายังไม่มี (สำหรับ table เก่า)
	alterProductQuery := `
		ALTER TABLE product ADD COLUMN IF NOT EXISTS balanceqty NUMERIC(18,4) DEFAULT 0;
		ALTER TABLE product ADD COLUMN IF NOT EXISTS balanceqtyword TEXT DEFAULT '';
		ALTER TABLE product ADD COLUMN IF NOT EXISTS pendingrecvqty NUMERIC(18,4) DEFAULT 0;
		ALTER TABLE product ADD COLUMN IF NOT EXISTS pendingrecvqtyword TEXT DEFAULT '';
		ALTER TABLE product ADD COLUMN IF NOT EXISTS pendingsendqty NUMERIC(18,4) DEFAULT 0;
		ALTER TABLE product ADD COLUMN IF NOT EXISTS pendingsendqtyword TEXT DEFAULT '';
	`
	if _, err := tx.ExecContext(context.Background(), alterProductQuery); err != nil {
		return fmt.Errorf("add product balance columns: %w", err)
	}

	// 2. สร้าง comments และ indexes ใน query เดียว
	commentsAndIndexesQuery := `
		-- สร้าง table comment
		COMMENT ON TABLE product IS 'ตารางเก็บข้อมูลสินค้า';

		-- สร้าง column comments
		COMMENT ON COLUMN product.id IS 'รหัสอัตโนมัติ';
		COMMENT ON COLUMN product.itemcode IS 'รหัสสินค้า';
		COMMENT ON COLUMN product.name0 IS 'ชื่อสินค้า';
		COMMENT ON COLUMN product.unitcode IS 'รหัสหน่วยนับ';
		COMMENT ON COLUMN product.unitname IS 'ชื่อหน่วยนับ';
		COMMENT ON COLUMN product.balanceqty IS 'ยอดคงเหลือที่แปลงแล้ว 1:1 (ใช้รหัสสินค้า)';
		COMMENT ON COLUMN product.balanceqtyword IS 'ยอดคงเหลือเป็นข้อความ เช่น 10 โหลx3 ชิ้น';
		COMMENT ON COLUMN product.pendingrecvqty IS 'ยอดค้างรับ (PO สั่งซื้อ - รับแล้ว)';
		COMMENT ON COLUMN product.pendingrecvqtyword IS 'ยอดค้างรับเป็นข้อความ';
		COMMENT ON COLUMN product.pendingsendqty IS 'ยอดค้างส่ง (SO สั่งขาย - ส่งแล้ว)';
		COMMENT ON COLUMN product.pendingsendqtyword IS 'ยอดค้างส่งเป็นข้อความ';

		-- สร้าง indexes และ constraints
		CREATE INDEX IF NOT EXISTS idx_product_itemcode ON product (itemcode)`

	if _, err := tx.ExecContext(context.Background(), commentsAndIndexesQuery); err != nil {
		return fmt.Errorf("create product comments and indexes: %w", err)
	}

	// 3. สร้าง unique constraint แยกต่างหาก (เพราะ PostgreSQL ไม่รองรับ IF NOT EXISTS)
	_, err = tx.ExecContext(context.Background(), `
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint 
				WHERE conname = 'product_itemcode_unique'
			) THEN
				ALTER TABLE product ADD CONSTRAINT product_itemcode_unique UNIQUE (itemcode);
			END IF;
		END
		$$`)
	if err != nil {
		return fmt.Errorf("create product unique constraint: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	logger.Success("created product table with all indexes and comments")
	return nil
}

func TableProductBarcodeCreate(db *sql.DB) error {
	logger.Info("Creating productbarcode table")

	// สร้างทุกอย่างใน transaction เดียว เพื่อความเร็วสูงสุด
	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. สร้าง table ก่อน
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS productbarcode (
			id SERIAL PRIMARY KEY,
			barcode TEXT NOT NULL,
			barcoderef TEXT,
			itemcode TEXT,
			name0 TEXT NOT NULL,
			unitcode TEXT NOT NULL,
			unitname TEXT NOT NULL,
			groupcode TEXT,
			groupnames TEXT,
			price1 NUMERIC(18,2),
			price_retail NUMERIC(18,2) DEFAULT 0,
			barcoderefunitstand NUMERIC(18,8),
			barcoderefunitdivide NUMERIC(18,8),
			isstock INT,
			itemtype INT,
			checksum CHAR(32) NOT NULL,
			isupdated BOOLEAN DEFAULT FALSE,
			unitstandanddivideisupdated BOOLEAN DEFAULT FALSE
		)
	`

	if _, err := tx.ExecContext(context.Background(), createTableQuery); err != nil {
		return fmt.Errorf("create productbarcode table: %w", err)
	}

	// 1.5. เพิ่ม columns ใหม่ถ้ายังไม่มี (สำหรับ table เก่า)
	alterTableQuery := `
		ALTER TABLE productbarcode
		ADD COLUMN IF NOT EXISTS price_retail NUMERIC(18,2) DEFAULT 0;

		ALTER TABLE productbarcode ADD COLUMN IF NOT EXISTS guidfixed TEXT DEFAULT '';
		ALTER TABLE productbarcode ADD COLUMN IF NOT EXISTS holding_code TEXT DEFAULT '';
		ALTER TABLE productbarcode ADD COLUMN IF NOT EXISTS brandcode TEXT DEFAULT '';
		ALTER TABLE productbarcode ADD COLUMN IF NOT EXISTS brandnames TEXT DEFAULT '';
		ALTER TABLE productbarcode ADD COLUMN IF NOT EXISTS categorycode TEXT DEFAULT '';
		ALTER TABLE productbarcode ADD COLUMN IF NOT EXISTS categorynames TEXT DEFAULT '';
		ALTER TABLE productbarcode ADD COLUMN IF NOT EXISTS classcode TEXT DEFAULT '';
		ALTER TABLE productbarcode ADD COLUMN IF NOT EXISTS classnames TEXT DEFAULT '';
		ALTER TABLE productbarcode ADD COLUMN IF NOT EXISTS designcode TEXT DEFAULT '';
		ALTER TABLE productbarcode ADD COLUMN IF NOT EXISTS designnames TEXT DEFAULT '';
		ALTER TABLE productbarcode ADD COLUMN IF NOT EXISTS gradecode TEXT DEFAULT '';
		ALTER TABLE productbarcode ADD COLUMN IF NOT EXISTS gradenames TEXT DEFAULT '';
		ALTER TABLE productbarcode ADD COLUMN IF NOT EXISTS modelcode TEXT DEFAULT '';
		ALTER TABLE productbarcode ADD COLUMN IF NOT EXISTS modelnames TEXT DEFAULT '';
		ALTER TABLE productbarcode ADD COLUMN IF NOT EXISTS patterncode TEXT DEFAULT '';
		ALTER TABLE productbarcode ADD COLUMN IF NOT EXISTS patternnames TEXT DEFAULT '';
		ALTER TABLE productbarcode ADD COLUMN IF NOT EXISTS imageuri TEXT DEFAULT '';
		ALTER TABLE productbarcode ADD COLUMN IF NOT EXISTS isusesubbarcodes BOOLEAN DEFAULT FALSE;
		ALTER TABLE productbarcode ADD COLUMN IF NOT EXISTS groupsubonecode TEXT DEFAULT '';
		ALTER TABLE productbarcode ADD COLUMN IF NOT EXISTS groupsubonenames TEXT DEFAULT '';
		ALTER TABLE productbarcode ADD COLUMN IF NOT EXISTS groupsubtwocode TEXT DEFAULT '';
		ALTER TABLE productbarcode ADD COLUMN IF NOT EXISTS groupsubtwonames TEXT DEFAULT '';
	`
	if _, err := tx.ExecContext(context.Background(), alterTableQuery); err != nil {
		return fmt.Errorf("add new columns: %w", err)
	}

	// 2. สร้าง comments และ indexes ใน query เดียว
	commentsAndIndexesQuery := `
		-- สร้าง table comment
		COMMENT ON TABLE productbarcode IS 'ตารางเก็บข้อมูลบาร์โค้ดสินค้า';

		-- สร้าง column comments
		COMMENT ON COLUMN productbarcode.id IS 'รหัสอัตโนมัติ';
		COMMENT ON COLUMN productbarcode.barcode IS 'บาร์โค้ด';
		COMMENT ON COLUMN productbarcode.barcoderef IS 'บาร์โค้ดอ้างอิง (ถ้ามี)';
		COMMENT ON COLUMN productbarcode.itemcode IS 'รหัสสินค้า';
		COMMENT ON COLUMN productbarcode.name0 IS 'ชื่อสินค้า';
		COMMENT ON COLUMN productbarcode.unitcode IS 'รหัสหน่วยนับ';
		COMMENT ON COLUMN productbarcode.unitname IS 'ชื่อหน่วยนับ';
		COMMENT ON COLUMN productbarcode.groupcode IS 'รหัสกลุ่มสินค้า';
		COMMENT ON COLUMN productbarcode.groupnames IS 'ชื่อกลุ่มสินค้า';
		COMMENT ON COLUMN productbarcode.price1 IS 'ราคาขาย';
		COMMENT ON COLUMN productbarcode.price_retail IS 'ราคาขายปลีก (keynumber=1 จาก MongoDB)';
		COMMENT ON COLUMN productbarcode.barcoderefunitstand IS 'ค่ามาตรฐานของบาร์โค้ดอ้างอิง';
		COMMENT ON COLUMN productbarcode.barcoderefunitdivide IS 'ค่าหารของบาร์โค้ดอ้างอิง';
		COMMENT ON COLUMN productbarcode.isstock IS 'สถานะสต็อก (1=มีสต็อก มีการอ้างอิงไป barcode อื่น, 0=ไม่มีสต็อก ไม่มีการอ้างอิงไป barcode อื่น)';
		COMMENT ON COLUMN productbarcode.itemtype IS 'ประเภทสินค้า (1=สินค้าปกติ, 2=บริการ, 3=ชุดสินค้า)';
		COMMENT ON COLUMN productbarcode.checksum IS 'ค่าเช็คซัมสำหรับตรวจสอบ mongodb มีการเปลี่ยนแปลงข้อมูลหรือไม่';
		COMMENT ON COLUMN productbarcode.isupdated IS 'สถานะการอัพเดต (TRUE=มีการเปลี่ยนแปลงข้อมูล, FALSE=ไม่มีการเปลี่ยนแปลงข้อมูล)';
		COMMENT ON COLUMN productbarcode.unitstandanddivideisupdated IS 'สถานะการอัพเดตค่ามาตรฐานและค่าหาร (TRUE=มีการเปลี่ยนแปลงข้อมูล, FALSE=ไม่มีการเปลี่ยนแปลงข้อมูล)';

		-- สร้าง indexes
		CREATE INDEX IF NOT EXISTS idx_productbarcode_itemcode ON productbarcode (itemcode);
		CREATE INDEX IF NOT EXISTS idx_productbarcode_itemcode_unitcode ON productbarcode (itemcode,unitcode);
		CREATE INDEX IF NOT EXISTS idx_productbarcode_barcode ON productbarcode (barcode);
		CREATE INDEX IF NOT EXISTS idx_productbarcode_barcoderef ON productbarcode (barcoderef);
		CREATE INDEX IF NOT EXISTS idx_productbarcode_isupdated ON productbarcode (isupdated);
		CREATE INDEX IF NOT EXISTS idx_productbarcode_unitstandanddivideisupdated ON productbarcode (unitstandanddivideisupdated);
		CREATE INDEX IF NOT EXISTS idx_productbarcode_holding_code ON productbarcode (holding_code);
		CREATE INDEX IF NOT EXISTS idx_productbarcode_guidfixed ON productbarcode (guidfixed);
		CREATE INDEX IF NOT EXISTS idx_productbarcode_groupcode ON productbarcode (groupcode);
		CREATE INDEX IF NOT EXISTS idx_productbarcode_brandcode ON productbarcode (brandcode);
		CREATE INDEX IF NOT EXISTS idx_productbarcode_categorycode ON productbarcode (categorycode);

		CREATE INDEX IF NOT EXISTS idx_productbarcode_name0_trgm ON productbarcode USING gin (name0 gin_trgm_ops);
		CREATE INDEX IF NOT EXISTS idx_productbarcode_barcode_trgm ON productbarcode USING gin (barcode gin_trgm_ops);
		CREATE INDEX IF NOT EXISTS idx_productbarcode_itemcode_trgm ON productbarcode USING gin (itemcode gin_trgm_ops);
		`

	if _, err := tx.ExecContext(context.Background(), commentsAndIndexesQuery); err != nil {
		return fmt.Errorf("create productbarcode comments and indexes: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	// pgvector column + HNSW index — แยก connection/transaction ต่างหาก เพราะถ้า pgvector
	// ไม่ได้ติดตั้ง statement นี้ fail แล้วจะทำให้ tx ทั้งก้อนด้านบน abort ไปด้วย (ทั้งที่ commit ไปแล้วก็ไม่กระทบ)
	if _, err := db.ExecContext(context.Background(), `ALTER TABLE productbarcode ADD COLUMN IF NOT EXISTS name_embedding vector(768)`); err != nil {
		logger.Info("productbarcode.name_embedding column skipped (pgvector extension not installed): %v", err)
	} else if _, err := db.ExecContext(context.Background(), `CREATE INDEX IF NOT EXISTS idx_productbarcode_embedding_hnsw ON productbarcode USING hnsw (name_embedding vector_cosine_ops) WITH (m = 16, ef_construction = 64)`); err != nil {
		logger.Info("productbarcode embedding HNSW index skipped: %v", err)
	}

	logger.Success("created productbarcode table with all indexes and comments")
	return nil
}

func TableDocDetailCreate(db *sql.DB) error {
	logger.Info("Creating docdetail table")

	// สร้างทุกอย่างใน transaction เดียว เพื่อความเร็วสูงสุด
	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. สร้าง table ก่อน
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS docdetail (
			id SERIAL PRIMARY KEY,
			docdatetime TIMESTAMPTZ,
			docno TEXT,
			docref TEXT,
			description TEXT,
			linenumber INT,
			transflag INT,
			calcflag INT,
			calcseq INT,
			iscancel BOOLEAN DEFAULT FALSE,
			itemcode TEXT,
			barcodemain TEXT,
			barcode TEXT,
			unitcode TEXT,
			whcode TEXT,
			locationcode TEXT,
			totalqty NUMERIC(18,8),
			unitstand NUMERIC(18,8),
			unitdivide NUMERIC(18,8),
			price NUMERIC(18,2),
			priceexcludevat NUMERIC(18,2),
			sumamount NUMERIC(18,2),
			isupdated BOOLEAN DEFAULT FALSE,
			iscalcstock INT DEFAULT 0,
			price_doc NUMERIC(18,2) DEFAULT 0,
			sumamount_doc NUMERIC(18,2) DEFAULT 0,
			discountamount_doc NUMERIC(18,2) DEFAULT 0,
			priceexcludevat_doc NUMERIC(18,2) DEFAULT 0,
			sumamountexcludevat_doc NUMERIC(18,2) DEFAULT 0,
			totalvaluevat_doc NUMERIC(18,2) DEFAULT 0
		)`

	if _, err := tx.ExecContext(context.Background(), createTableQuery); err != nil {
		return fmt.Errorf("create docdetail table: %w", err)
	}

	// 2. สร้าง comments และ indexes ใน query เดียว
	commentsAndIndexesQuery := `
		-- สร้าง table comment
		COMMENT ON TABLE docdetail IS 'ตารางเก็บข้อมูลรายละเอียดสินค้าในเอกสารทุกอย่าง';

		-- สร้าง column comments
		COMMENT ON COLUMN docdetail.id IS 'รหัสอัตโนมัติ';
		COMMENT ON COLUMN docdetail.docdatetime IS 'วันที่และเวลาของเอกสาร';
		COMMENT ON COLUMN docdetail.docno IS 'เลขที่เอกสาร';
		COMMENT ON COLUMN docdetail.docref IS 'เลขที่อ้างอิงเอกสาร';
		COMMENT ON COLUMN docdetail.description IS 'รายละเอียดสินค้า';
		COMMENT ON COLUMN docdetail.linenumber IS 'เลขที่บรรทัด';
		COMMENT ON COLUMN docdetail.transflag IS 'ประเภทเอกสาร';
		COMMENT ON COLUMN docdetail.calcflag IS 'สถานะการคำนวณสต็อก (0=ยังไม่คำนวณ, 1=คำนวณแล้ว)';
		COMMENT ON COLUMN docdetail.calcseq IS 'ลำดับการคำนวณสต็อก';
		COMMENT ON COLUMN docdetail.iscancel IS 'สถานะยกเลิก (TRUE=ยกเลิก, FALSE=ไม่ยกเลิก)';
		COMMENT ON COLUMN docdetail.itemcode IS 'รหัสสินค้า';
		COMMENT ON COLUMN docdetail.barcodemain IS 'บาร์โค้ดหลักของสินค้า';
		COMMENT ON COLUMN docdetail.barcode IS 'บาร์โค้ดย่อยของสินค้า';
		COMMENT ON COLUMN docdetail.unitcode IS 'รหัสหน่วยนับ';
		COMMENT ON COLUMN docdetail.whcode IS 'รหัสคลังสินค้า';
		COMMENT ON COLUMN docdetail.locationcode IS 'รหัสที่เก็บสินค้า';
		COMMENT ON COLUMN docdetail.totalqty IS 'จำนวนสินค้า';
		COMMENT ON COLUMN docdetail.unitstand IS 'ค่ามาตรฐานของหน่วยนับ';
		COMMENT ON COLUMN docdetail.unitdivide IS 'ค่าหารของหน่วยนับ';
		COMMENT ON COLUMN docdetail.price IS 'ราคาขาย';
		COMMENT ON COLUMN docdetail.priceexcludevat IS 'ราคาขายไม่รวมภาษีมูลค่าเพิ่ม';
		COMMENT ON COLUMN docdetail.sumamount IS 'จำนวนเงินรวม';
		COMMENT ON COLUMN docdetail.iscalcstock IS 'มีการคำนวณสต็อกหรือไม่ (0=ไม่คำนวณ, 1=คำนวณ เช่น ซื้อ/ขาย)';

		-- สร้าง indexes
		CREATE INDEX IF NOT EXISTS idx_docdetail_barcodemain ON docdetail (barcodemain);
		CREATE INDEX IF NOT EXISTS idx_docdetail_itemcode ON docdetail (itemcode);
		CREATE INDEX IF NOT EXISTS idx_docdetail_itemcode_tr ON docdetail (itemcode,transflag);
		CREATE INDEX IF NOT EXISTS idx_docdetail_docno ON docdetail (docno);
		CREATE INDEX IF NOT EXISTS idx_docdetail_docref ON docdetail (docref);
		CREATE INDEX IF NOT EXISTS idx_docdetail_docdatetime ON docdetail (docdatetime);
		CREATE INDEX IF NOT EXISTS idx_docdetail_barcode ON docdetail (barcode);
		CREATE INDEX IF NOT EXISTS idx_docdetail_isupdated ON docdetail (isupdated);
		`

	if _, err := tx.ExecContext(context.Background(), commentsAndIndexesQuery); err != nil {
		return fmt.Errorf("create docdetail comments and indexes: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	logger.Success("created docdetail table with all indexes and comments")
	return nil
}

func TableDocCreate(db *sql.DB) error {
	logger.Info("Creating doc table")

	// สร้างทุกอย่างใน transaction เดียว เพื่อความเร็วสูงสุด
	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. สร้าง table ก่อน
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS doc (
			id SERIAL PRIMARY KEY,
			transflag INT,
			docno TEXT,
			custcode TEXT,
			docdatetime TIMESTAMPTZ,
			perioddatetime TIMESTAMPTZ,
			taxdocno TEXT,
			totalamount NUMERIC(18,8),
			roundamount NUMERIC(18,8),
			paytype INT,
			paycashamount NUMERIC(18,8),
			paycashchange NUMERIC(18,8),
			paycashbalance NUMERIC(18,8),
			deliverycode TEXT,
			checksum CHAR(32) NOT NULL,
			branchid TEXT,
			slipurl TEXT,
			salechannelcode TEXT,
			deliveryamount NUMERIC(18,8),
			iscancel BOOLEAN DEFAULT FALSE,
			cancelreason TEXT,
			guidpos TEXT,
			guidbranch TEXT,
			guidfixed TEXT,
			isref BOOLEAN DEFAULT FALSE,
			islocked BOOLEAN DEFAULT FALSE,
			isclosed BOOLEAN DEFAULT FALSE,
			iscomparedsuccess INT DEFAULT 0,
			currency TEXT DEFAULT 'THB',
			currency_symbol TEXT DEFAULT '฿',
			doc_currency TEXT DEFAULT 'THB',
			doc_currency_symbol TEXT DEFAULT '฿',
			exchange_rate NUMERIC(18,8) DEFAULT 1,
			totalvalue_doc NUMERIC(18,8) DEFAULT 0,
			totaldiscount_doc NUMERIC(18,8) DEFAULT 0,
			totalvatvalue_doc NUMERIC(18,8) DEFAULT 0,
			totalbeforevat_doc NUMERIC(18,8) DEFAULT 0,
			totalaftervat_doc NUMERIC(18,8) DEFAULT 0,
			totalamount_doc NUMERIC(18,8) DEFAULT 0,
			creator_code TEXT DEFAULT '',
			creator_name TEXT DEFAULT '',
			created_at TIMESTAMPTZ DEFAULT NOW(),
			modifier_code TEXT DEFAULT '',
			modifier_name TEXT DEFAULT '',
			modified_at TIMESTAMPTZ DEFAULT NOW(),
			isdelete BOOLEAN DEFAULT FALSE,
			approval_status TEXT DEFAULT '',
			isclosedmanual BOOLEAN DEFAULT FALSE,
			closedmanual_by_code TEXT DEFAULT '',
			closedmanual_by_name TEXT DEFAULT '',
			closedmanual_at TIMESTAMPTZ DEFAULT '1970-01-01',
			closedmanual_reason TEXT DEFAULT ''
		)`

	if _, err := tx.ExecContext(context.Background(), createTableQuery); err != nil {
		return fmt.Errorf("create doc table: %w", err)
	}

	// 2. สร้าง comments และ indexes ใน query เดียว
	commentsAndIndexesQuery := `
		-- สร้าง table comment
		COMMENT ON TABLE doc IS 'ตารางเก็บข้อมูลเอกสารหลักทุกประเภท';

		-- สร้าง column comments
		COMMENT ON COLUMN doc.id IS 'รหัสอัตโนมัติ';
		COMMENT ON COLUMN doc.transflag IS 'ประเภทเอกสาร';
		COMMENT ON COLUMN doc.docno IS 'เลขที่เอกสาร';
		COMMENT ON COLUMN doc.custcode IS 'รหัสลูกค้า';
		COMMENT ON COLUMN doc.docdatetime IS 'วันที่และเวลาของเอกสาร';
		COMMENT ON COLUMN doc.perioddatetime IS 'วันที่งวดบัญชี';
		COMMENT ON COLUMN doc.taxdocno IS 'เลขที่เอกสารภาษี';
		COMMENT ON COLUMN doc.totalamount IS 'จำนวนเงินรวมทั้งหมด';
		COMMENT ON COLUMN doc.roundamount IS 'จำนวนเงินหลังปัดเศษ';
		COMMENT ON COLUMN doc.paytype IS 'ประเภทการชำระเงิน';
		COMMENT ON COLUMN doc.paycashamount IS 'จำนวนเงินสดที่จ่าย';
		COMMENT ON COLUMN doc.paycashchange IS 'เงินทอน';
		COMMENT ON COLUMN doc.paycashbalance IS 'ยอดเงินคงเหลือ';
		COMMENT ON COLUMN doc.deliverycode IS 'รหัสการส่งของ';
		COMMENT ON COLUMN doc.checksum IS 'ค่าเช็คซัมสำหรับตรวจสอบการเปลี่ยนแปลงข้อมูล จาก mongodb';
		COMMENT ON COLUMN doc.branchid IS 'รหัสสาขา';
		COMMENT ON COLUMN doc.slipurl IS 'URL ของสลิป';
		COMMENT ON COLUMN doc.salechannelcode IS 'รหัสช่องทางการขาย';
		COMMENT ON COLUMN doc.deliveryamount IS 'ค่าส่งของ';
		COMMENT ON COLUMN doc.guidpos IS 'GUID ของ POS';
		COMMENT ON COLUMN doc.guidbranch IS 'GUID ของสาขา';
		COMMENT ON COLUMN doc.guidfixed IS 'GUID คงที่';
		COMMENT ON COLUMN doc.isref IS 'สถานะเอกสาร (TRUE=มีการอ้างอิงแล้ว ห้ามแก้ไข, FALSE=สามารถแก้ไขได้)';
		COMMENT ON COLUMN doc.iscancel IS 'สถานะยกเลิก (TRUE=ยกเลิก, FALSE=ไม่ยกเลิก)';
		COMMENT ON COLUMN doc.cancelreason IS 'เหตุผลการยกเลิก';
		COMMENT ON COLUMN doc.islocked IS 'สถานะล็อกเอกสาร (TRUE=ล็อก ห้ามแก้ไข, FALSE=ไม่ล็อก สามารถแก้ไขได้)';
		COMMENT ON COLUMN doc.isclosed IS 'สถานะปิดเอกสาร (TRUE=ปิดเอกสาร ห้ามแก้ไข, FALSE=ไม่ปิด สามารถแก้ไขได้)';
		COMMENT ON COLUMN doc.iscomparedsuccess IS 'สถานะการเปรียบเทียบ (0=ยังไม่ไม่,1=ครบแล้ว,2=ขาด,3=เกิน)';
		COMMENT ON COLUMN doc.currency IS 'สกุลเงินหลักสำหรับลงบัญชี (THB, USD, JPY, EUR, ...)';
		COMMENT ON COLUMN doc.currency_symbol IS 'สัญลักษณ์สกุลเงินหลัก (฿, $, ¥, €, ...)';
		COMMENT ON COLUMN doc.doc_currency IS 'สกุลเงินเอกสาร (THB, USD, JPY, EUR, ...) - ใช้สำหรับแสดงผลและบันทึกธุรกรรม';
		COMMENT ON COLUMN doc.doc_currency_symbol IS 'สัญลักษณ์สกุลเงินเอกสาร (฿, $, ¥, €, ...)';
		COMMENT ON COLUMN doc.exchange_rate IS 'อัตราแลกเปลี่ยน (Base Currency = Doc Amount x Exchange Rate)';
		COMMENT ON COLUMN doc.totalvalue_doc IS 'มูลค่ารวมสินค้า (Document Currency)';
		COMMENT ON COLUMN doc.totaldiscount_doc IS 'ส่วนลดท้ายบิล (Document Currency)';
		COMMENT ON COLUMN doc.totalvatvalue_doc IS 'มูลค่าภาษี (Document Currency)';
		COMMENT ON COLUMN doc.totalbeforevat_doc IS 'มูลค่าก่อนหักภาษี (Document Currency)';
		COMMENT ON COLUMN doc.totalaftervat_doc IS 'มูลค่าหลังหักภาษี (Document Currency)';
		COMMENT ON COLUMN doc.totalamount_doc IS 'มูลค่ารวมทั้งหมด (Document Currency)';
		COMMENT ON COLUMN doc.creator_code IS 'รหัสผู้สร้างเอกสาร';
		COMMENT ON COLUMN doc.creator_name IS 'ชื่อผู้สร้างเอกสาร';
		COMMENT ON COLUMN doc.created_at IS 'วันเวลาที่สร้างเอกสาร';
		COMMENT ON COLUMN doc.modifier_code IS 'รหัสผู้แก้ไขเอกสารล่าสุด';
		COMMENT ON COLUMN doc.modifier_name IS 'ชื่อผู้แก้ไขเอกสารล่าสุด';
		COMMENT ON COLUMN doc.modified_at IS 'วันเวลาที่แก้ไขเอกสารล่าสุด';

		COMMENT ON COLUMN doc.isdelete IS 'สถานะลบเอกสาร (TRUE=ลบแล้ว, FALSE=ยังใช้งาน)';
		COMMENT ON COLUMN doc.approval_status IS 'สถานะการอนุมัติ (draft, pending, approved, rejected, auto_approved, cancelled)';

		-- สร้าง indexes
		CREATE INDEX IF NOT EXISTS idx_doc_transflag_docno_docdatetime ON doc (transflag, docno, docdatetime);
		CREATE INDEX IF NOT EXISTS idx_doc_transflag_docno_perioddatetime ON doc (transflag, docno, perioddatetime);
		CREATE INDEX IF NOT EXISTS idx_doc_creator_code ON doc (creator_code);
		CREATE INDEX IF NOT EXISTS idx_doc_created_at ON doc (created_at);
		CREATE INDEX IF NOT EXISTS idx_doc_approval_status ON doc (approval_status) WHERE approval_status != ''`

	if _, err := tx.ExecContext(context.Background(), commentsAndIndexesQuery); err != nil {
		return fmt.Errorf("create doc comments and indexes: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	logger.Success("created doc table with all indexes and comments")
	return nil
}

func TableResultCreate(db *sql.DB) error {
	logger.Info("Creating result table")

	// สร้างทุกอย่างใน transaction เดียว เพื่อความเร็วสูงสุด
	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. สร้าง table ก่อน
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS result (
			id SERIAL PRIMARY KEY,
			guid TEXT,
			querynumber INTEGER default 0,
			level INTEGER default 0,
			typejson INTEGER default 0,
			docdatetime TIMESTAMPTZ,
			linenumber INTEGER,
			datajson JSONB
		)`

	if _, err := tx.ExecContext(context.Background(), createTableQuery); err != nil {
		return fmt.Errorf("create result table: %w", err)
	}

	// 2. สร้าง comments และ indexes ใน query เดียว
	commentsAndIndexesQuery := `
		-- สร้าง table comment
		COMMENT ON TABLE result IS 'ตารางเก็บผลลัพธ์การประมวลผลข้อมูล';

		-- สร้าง column comments
		COMMENT ON COLUMN result.id IS 'รหัสอัตโนมัติ';
		COMMENT ON COLUMN result.guid IS 'รหัส GUID สำหรับอ้างอิง';
		COMMENT ON COLUMN result.querynumber IS 'ประเภทของผลลัพธ์';
		COMMENT ON COLUMN result.level IS 'ระดับของข้อมูล (0=หัวข้อหลัก, 1=รายละเอียด, 2=รายละเอียดย่อย ,ฯลฯ)';
		COMMENT ON COLUMN result.typejson IS 'ชนิดของข้อมูล JSON (สำรองไว้ในกรณีที่ต้องการแยกประเภทข้อมูล เช่น ข้อมูล,ยอดรวม)';
		COMMENT ON COLUMN result.docdatetime IS 'วันที่และเวลาของเอกสาร';
		COMMENT ON COLUMN result.linenumber IS 'เลขที่บรรทัด';
		COMMENT ON COLUMN result.datajson IS 'ข้อมูล JSONB ของผลลัพธ์ (รองรับการ query และ index)';

		-- สร้าง indexes
		CREATE INDEX IF NOT EXISTS idx_result_docdatetime ON result (docdatetime);
		CREATE INDEX IF NOT EXISTS idx_result_guid_querynumber_linenumber ON result (guid,querynumber,linenumber);
		`

	if _, err := tx.ExecContext(context.Background(), commentsAndIndexesQuery); err != nil {
		return fmt.Errorf("create result comments and indexes: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	logger.Success("created result table with all indexes and comments")
	return nil
}

func TableStockWaitProcessCreate(db *sql.DB) error {
	logger.Info("Creating stockwaitprocess table")

	// สร้างทุกอย่างใน transaction เดียว เพื่อความเร็วสูงสุด
	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. สร้าง table ก่อน
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS stockwaitprocess (
			id SERIAL PRIMARY KEY,
			itemcode TEXT
		)`

	if _, err := tx.ExecContext(context.Background(), createTableQuery); err != nil {
		return fmt.Errorf("create stockwaitprocess table: %w", err)
	}

	// 2. สร้าง comments และ indexes ใน query เดียว
	commentsAndIndexesQuery := `
		-- สร้าง table comment
		COMMENT ON TABLE stockwaitprocess IS 'ตารางเก็บข้อมูลสต็อกที่รอการประมวลผล';

		-- สร้าง column comments
		COMMENT ON COLUMN stockwaitprocess.id IS 'รหัสอัตโนมัติ';
		COMMENT ON COLUMN stockwaitprocess.itemcode IS 'รหัสสินค้า';

		-- สร้าง indexes
		CREATE INDEX IF NOT EXISTS idx_stockwaitprocess_itemcode ON stockwaitprocess (itemcode)`

	if _, err := tx.ExecContext(context.Background(), commentsAndIndexesQuery); err != nil {
		return fmt.Errorf("create stockwaitprocess comments and indexes: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	logger.Success("created stockwaitprocess table with all indexes and comments")
	return nil
}

func TableDocWaitProcessCreate(db *sql.DB) error {
	logger.Info("Creating docwaitprocess table")

	// สร้างทุกอย่างใน transaction เดียว เพื่อความเร็วสูงสุด
	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. สร้าง table ก่อน
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS docwaitprocess (
			id SERIAL PRIMARY KEY,
			transflag INT,
			docno TEXT
		)`

	if _, err := tx.ExecContext(context.Background(), createTableQuery); err != nil {
		return fmt.Errorf("create docwaitprocess table: %w", err)
	}

	// 2. สร้าง comments และ indexes ใน query เดียว
	commentsAndIndexesQuery := `
		-- สร้าง table comment
		COMMENT ON TABLE docwaitprocess IS 'ตารางเก็บข้อมูลเอกสารที่รอการประมวลผล';

		-- สร้าง column comments
		COMMENT ON COLUMN docwaitprocess.id IS 'รหัสอัตโนมัติ';
		COMMENT ON COLUMN docwaitprocess.transflag IS 'ประเภทเอกสาร';
		COMMENT ON COLUMN docwaitprocess.docno IS 'เลขที่เอกสาร';

		-- สร้าง indexes
		CREATE INDEX IF NOT EXISTS idx_docwaitprocess_transflag_docno ON docwaitprocess (transflag, docno)`

	if _, err := tx.ExecContext(context.Background(), commentsAndIndexesQuery); err != nil {
		return fmt.Errorf("create docwaitprocess comments and indexes: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	logger.Success("created docwaitprocess table with all indexes and comments")
	return nil
}

func TableProcessStockCostCreate(db *sql.DB) error {
	logger.Info("Creating processstockcost table")

	// สร้างทุกอย่างใน transaction เดียว เพื่อความเร็วสูงสุด
	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. สร้าง table ก่อน
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS processstockcost (
			id SERIAL PRIMARY KEY,
			docdatetime TIMESTAMPTZ,
			docno TEXT,
			docref TEXT,
			linenumber INT,
			transflag INT,
			itemcode TEXT,
			barcode TEXT,
			unitcode TEXT,
			whcode TEXT,
			locationcode TEXT,
			totalqty NUMERIC(18,8),
			unitstand NUMERIC(18,8),
			unitdivide NUMERIC(18,8),
			price NUMERIC(18,2),
			averagecost NUMERIC(18,2),
			calcamount NUMERIC(18,2),
			balanceqty NUMERIC(18,8),
			balanceamount NUMERIC(18,2),
			unitcost NUMERIC(18,2),			
			guid TEXT
		)`

	if _, err := tx.ExecContext(context.Background(), createTableQuery); err != nil {
		return fmt.Errorf("create processstockcost table: %w", err)
	}

	// 2. สร้าง comments และ indexes ใน query เดียว
	commentsAndIndexesQuery := `
		-- สร้าง table comment
		COMMENT ON TABLE processstockcost IS 'ตารางเก็บข้อมูลการประมวลผลต้นทุนสต็อก';

		-- สร้าง column comments
		COMMENT ON COLUMN processstockcost.id IS 'รหัสอัตโนมัติ';
		COMMENT ON COLUMN processstockcost.docdatetime IS 'วันที่และเวลาของเอกสาร';
		COMMENT ON COLUMN processstockcost.docno IS 'เลขที่เอกสาร';
		COMMENT ON COLUMN processstockcost.docref IS 'เลขที่อ้างอิงเอกสาร';
		COMMENT ON COLUMN processstockcost.linenumber IS 'เลขที่บรรทัด';
		COMMENT ON COLUMN processstockcost.transflag IS 'ประเภทเอกสาร';
		COMMENT ON COLUMN processstockcost.itemcode IS 'รหัสสินค้าหลัก';
		COMMENT ON COLUMN processstockcost.barcode IS 'บาร์โค้ดของสินค้า';
		COMMENT ON COLUMN processstockcost.unitcode IS 'รหัสหน่วยนับมาตรฐาน';
		COMMENT ON COLUMN processstockcost.whcode IS 'รหัสคลังสินค้า';
		COMMENT ON COLUMN processstockcost.locationcode IS 'รหัสที่เก็บสินค้า';
		COMMENT ON COLUMN processstockcost.totalqty IS 'จำนวนสินค้าทั้งหมด';
		COMMENT ON COLUMN processstockcost.unitstand IS 'ค่ามาตรฐานของหน่วยนับ';
		COMMENT ON COLUMN processstockcost.unitdivide IS 'ค่าหารของหน่วยนับ';
		COMMENT ON COLUMN processstockcost.price IS 'ราคาขาย';
		COMMENT ON COLUMN processstockcost.averagecost IS 'ต้นทุนเฉลี่ย';
		COMMENT ON COLUMN processstockcost.calcamount IS 'จำนวนเงินที่คำนวณ';
		COMMENT ON COLUMN processstockcost.balanceqty IS 'จำนวนคงเหลือ';
		COMMENT ON COLUMN processstockcost.balanceamount IS 'มูลค่าคงเหลือ';
		COMMENT ON COLUMN processstockcost.unitcost IS 'ต้นทุนต่อหน่วย';
		COMMENT ON COLUMN processstockcost.guid IS 'รหัส GUID สำหรับอ้างอิง';

		-- สร้าง indexes
		CREATE INDEX IF NOT EXISTS idx_processstockcost_itemcode ON processstockcost (itemcode);
		CREATE INDEX IF NOT EXISTS idx_processstockcost_itemcode_guid ON processstockcost (itemcode, guid);
		CREATE INDEX IF NOT EXISTS idx_processstockcost_docno ON processstockcost (docno);
		CREATE INDEX IF NOT EXISTS idx_processstockcost_docdatetime ON processstockcost (docdatetime);
		CREATE INDEX IF NOT EXISTS idx_processstockcost_guid ON processstockcost (guid)`

	if _, err := tx.ExecContext(context.Background(), commentsAndIndexesQuery); err != nil {
		return fmt.Errorf("create processstockcost comments and indexes: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	logger.Success("created processstockcost table with all indexes and comments")
	return nil
}

func TableProcessStockLotCreate(db *sql.DB) error {
	logger.Info("Creating processstocklot table")

	// สร้างทุกอย่างใน transaction เดียว เพื่อความเร็วสูงสุด
	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. สร้าง table ก่อน
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS processstocklot (
			id SERIAL PRIMARY KEY,
			docdatetime TIMESTAMPTZ,
			lotnumber TEXT,
			docno TEXT,
			transflag INT,
			itemcode TEXT,
			unitcode TEXT,
			whcode TEXT,
			locationcode TEXT,
			qty NUMERIC(18,8),
			price NUMERIC(18,2),
			unitstand NUMERIC(18,8),
			unitdivide NUMERIC(18,8),
			cost NUMERIC(18,2),
			balanceqty NUMERIC(18,8),
			balanceamount NUMERIC(18,2),
			guidref TEXT
		)`

	if _, err := tx.ExecContext(context.Background(), createTableQuery); err != nil {
		return fmt.Errorf("create processstocklot table: %w", err)
	}

	// 2. สร้าง comments และ indexes ใน query เดียว
	commentsAndIndexesQuery := `
		-- สร้าง table comment
		COMMENT ON TABLE processstocklot IS 'ตารางเก็บข้อมูลการประมวลผลสต็อกแบบ Lot';

		-- สร้าง column comments
		COMMENT ON COLUMN processstocklot.id IS 'รหัสอัตโนมัติ';
		COMMENT ON COLUMN processstocklot.docdatetime IS 'วันที่และเวลาของเอกสาร';
		COMMENT ON COLUMN processstocklot.lotnumber IS 'หมายเลข Lot';
		COMMENT ON COLUMN processstocklot.docno IS 'เลขที่เอกสาร';
		COMMENT ON COLUMN processstocklot.transflag IS 'ประเภทเอกสาร';
		COMMENT ON COLUMN processstocklot.itemcode IS 'รหัสสินค้าหลัก';
		COMMENT ON COLUMN processstocklot.unitcode IS 'รหัสหน่วยนับ';
		COMMENT ON COLUMN processstocklot.whcode IS 'รหัสคลังสินค้า';
		COMMENT ON COLUMN processstocklot.locationcode IS 'รหัสที่เก็บสินค้า';
		COMMENT ON COLUMN processstocklot.qty IS 'จำนวนสินค้า';
		COMMENT ON COLUMN processstocklot.price IS 'ราคาขาย';
		COMMENT ON COLUMN processstocklot.unitstand IS 'ค่ามาตรฐานของหน่วยนับ';
		COMMENT ON COLUMN processstocklot.unitdivide IS 'ค่าหารของหน่วยนับ';
		COMMENT ON COLUMN processstocklot.cost IS 'ต้นทุน';
		COMMENT ON COLUMN processstocklot.balanceqty IS 'จำนวนคงเหลือ';
		COMMENT ON COLUMN processstocklot.balanceamount IS 'มูลค่าคงเหลือ';
		COMMENT ON COLUMN processstocklot.guidref IS 'รหัส GUID อ้างอิง';

		-- สร้าง indexes
		CREATE INDEX IF NOT EXISTS idx_processstocklot_itemcode ON processstocklot (itemcode);
		CREATE INDEX IF NOT EXISTS idx_processstocklot_itemcode_guidref ON processstocklot (itemcode, guidref);
		CREATE INDEX IF NOT EXISTS idx_processstocklot_lotnumber ON processstocklot (lotnumber);
		CREATE INDEX IF NOT EXISTS idx_processstocklot_docno ON processstocklot (docno);
		CREATE INDEX IF NOT EXISTS idx_processstocklot_docdatetime ON processstocklot (docdatetime);
		CREATE INDEX IF NOT EXISTS idx_processstocklot_guidref ON processstocklot (guidref)`

	if _, err := tx.ExecContext(context.Background(), commentsAndIndexesQuery); err != nil {
		return fmt.Errorf("create processstocklot comments and indexes: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	logger.Success("created processstocklot table with all indexes and comments")
	return nil
}

func TableOrderCartCreate(db *sql.DB) error {
	logger.Info("Creating order_cart table")

	// สร้างทุกอย่างใน transaction เดียว เพื่อความเร็วสูงสุด
	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. สร้าง table ก่อน
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS order_cart (
			ignore_sync integer DEFAULT 0,
			is_lock_record integer DEFAULT 0,
			roworder serial,
			user_owner character varying(255) NOT NULL DEFAULT ''::character varying,
			cart_number character varying(255) NOT NULL DEFAULT ''::character varying,
			po_number character varying(255) NOT NULL DEFAULT ''::character varying,
			cust_code character varying(255) NOT NULL DEFAULT ''::character varying,
			status INT DEFAULT 0,
			amount numeric DEFAULT 0.0,
			trans_type INT DEFAULT 0,
			bill_type INT DEFAULT 0,
			tax_type INT DEFAULT 0,
			branch_code character varying(255) DEFAULT ''::character varying,
			create_date_time_now timestamp without time zone NOT NULL DEFAULT ('now'::text)::timestamp without time zone,
			remark character varying(255) DEFAULT ''::character varying,
			CONSTRAINT order_cart_pk_order_cart_user_number PRIMARY KEY (user_owner, cart_number)
		)`

	if _, err := tx.ExecContext(context.Background(), createTableQuery); err != nil {
		return fmt.Errorf("create order_cart table: %w", err)
	}

	// 2. สร้าง comments และ indexes ใน query เดียว
	commentsAndIndexesQuery := `
		-- สร้าง table comment
		COMMENT ON TABLE order_cart IS 'ตารางเก็บข้อมูลตะกร้าสินค้าของผู้ใช้';

		-- สร้าง column comments
		COMMENT ON COLUMN order_cart.ignore_sync IS 'ไม่นำไป sync (0=sync, 1=ไม่sync)';
		COMMENT ON COLUMN order_cart.is_lock_record IS 'สถานะล็อคเรคอร์ด (0=ปกติ, 1=ล็อค)';
		COMMENT ON COLUMN order_cart.roworder IS 'ลำดับแถว (auto increment)';
		COMMENT ON COLUMN order_cart.user_owner IS 'รหัสผู้ใช้เจ้าของตะกร้า';
		COMMENT ON COLUMN order_cart.cart_number IS 'หมายเลขตะกร้า';
		COMMENT ON COLUMN order_cart.cust_code IS 'รหัสลูกค้า';
		COMMENT ON COLUMN order_cart.status IS 'สถานะตะกร้า (0=ใช้งาน, 1=สั่งซื้อแล้ว, 2=ยกเลิก)';
		COMMENT ON COLUMN order_cart.amount IS 'ยอดรวมในตะกร้า';
		COMMENT ON COLUMN order_cart.trans_type IS 'ประเภทธุรกรรม';
		COMMENT ON COLUMN order_cart.bill_type IS 'ประเภทใบเสร็จ';
		COMMENT ON COLUMN order_cart.tax_type IS 'ประเภทภาษี';
		COMMENT ON COLUMN order_cart.branch_code IS 'รหัสสาขา';
		COMMENT ON COLUMN order_cart.create_date_time_now IS 'วันที่และเวลาสร้างตะกร้า';
		COMMENT ON COLUMN order_cart.remark IS 'หมายเหตุ';

		-- สร้าง indexes
		CREATE INDEX IF NOT EXISTS idx_order_cart_user_owner ON order_cart (user_owner);
		CREATE INDEX IF NOT EXISTS idx_order_cart_cust_code ON order_cart (cust_code);
		CREATE INDEX IF NOT EXISTS idx_order_cart_status ON order_cart (status);
		CREATE INDEX IF NOT EXISTS idx_order_cart_create_date_time_now ON order_cart (create_date_time_now);
		CREATE INDEX IF NOT EXISTS idx_order_cart_branch_code ON order_cart (branch_code)`

	if _, err := tx.ExecContext(context.Background(), commentsAndIndexesQuery); err != nil {
		return fmt.Errorf("create order_cart comments and indexes: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	logger.Success("created order_cart table with all indexes and comments")
	return nil
}

func TableOrderItemCreate(db *sql.DB) error {
	logger.Info("Creating order_item table")

	// สร้างทุกอย่างใน transaction เดียว เพื่อความเร็วสูงสุด
	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. สร้าง table ก่อน
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS order_item (
			ignore_sync integer DEFAULT 0,
			is_lock_record integer DEFAULT 0,
			roworder serial,
			user_owner character varying(255) NOT NULL DEFAULT ''::character varying,
			cart_number character varying(255) NOT NULL DEFAULT ''::character varying,
			item_code character varying(255) NOT NULL DEFAULT ''::character varying,
			unit_code character varying(255) NOT NULL DEFAULT ''::character varying,
			qty numeric DEFAULT 0.0,
			price numeric DEFAULT 0.0,
			discount_word character varying(255) DEFAULT ''::character varying,
			amount numeric DEFAULT 0.0,
			remark character varying(100) DEFAULT ''::character varying,
			order_date timestamp without time zone,
			order_time character varying(10) DEFAULT ''::character varying,
			wh_code character varying(255) DEFAULT ''::character varying,
			shelf_code character varying(255) DEFAULT ''::character varying,
			barcode character varying(255) DEFAULT ''::character varying,
			branch_code character varying(255) DEFAULT ''::character varying,
			create_date_time_now timestamp without time zone NOT NULL DEFAULT ('now'::text)::timestamp without time zone,
			lot_number character varying(255) DEFAULT ''::character varying,
			CONSTRAINT order_item_pkey PRIMARY KEY (roworder)
		)`

	if _, err := tx.ExecContext(context.Background(), createTableQuery); err != nil {
		return fmt.Errorf("create order_item table: %w", err)
	}

	// 2. สร้าง comments และ indexes ใน query เดียว
	commentsAndIndexesQuery := `
		-- สร้าง table comment
		COMMENT ON TABLE order_item IS 'ตารางเก็บรายการสินค้าในตะกร้าสั่งซื้อ';

		-- สร้าง column comments
		COMMENT ON COLUMN order_item.ignore_sync IS 'ไม่นำไป sync (0=sync, 1=ไม่sync)';
		COMMENT ON COLUMN order_item.is_lock_record IS 'สถานะล็อคเรคอร์ด (0=ปกติ, 1=ล็อค)';
		COMMENT ON COLUMN order_item.roworder IS 'ลำดับแถว (auto increment primary key)';
		COMMENT ON COLUMN order_item.user_owner IS 'รหัสผู้ใช้เจ้าของตะกร้า';
		COMMENT ON COLUMN order_item.cart_number IS 'หมายเลขตะกร้า';
		COMMENT ON COLUMN order_item.item_code IS 'รหัสสินค้า';
		COMMENT ON COLUMN order_item.unit_code IS 'รหัสหน่วยนับ';
		COMMENT ON COLUMN order_item.qty IS 'จำนวนสินค้า';
		COMMENT ON COLUMN order_item.price IS 'ราคาต่อหน่วย';
		COMMENT ON COLUMN order_item.discount_word IS 'คำอธิบายส่วนลด';
		COMMENT ON COLUMN order_item.amount IS 'ยอดรวมของรายการ';
		COMMENT ON COLUMN order_item.remark IS 'หมายเหตุ';
		COMMENT ON COLUMN order_item.order_date IS 'วันที่สั่งซื้อ';
		COMMENT ON COLUMN order_item.order_time IS 'เวลาสั่งซื้อ';
		COMMENT ON COLUMN order_item.wh_code IS 'รหัสคลังสินค้า';
		COMMENT ON COLUMN order_item.shelf_code IS 'รหัสชั้นวางสินค้า';
		COMMENT ON COLUMN order_item.barcode IS 'บาร์โค้ดสินค้า';
		COMMENT ON COLUMN order_item.branch_code IS 'รหัสสาขา';
		COMMENT ON COLUMN order_item.create_date_time_now IS 'วันที่และเวลาสร้างรายการ';
		COMMENT ON COLUMN order_item.lot_number IS 'หมายเลข Lot ของสินค้า';

		-- สร้าง indexes
		CREATE INDEX IF NOT EXISTS idx_order_item_user_owner_cart_number ON order_item (user_owner, cart_number);
		CREATE INDEX IF NOT EXISTS idx_order_item_item_code ON order_item (item_code);
		CREATE INDEX IF NOT EXISTS idx_order_item_barcode ON order_item (barcode);
		CREATE INDEX IF NOT EXISTS idx_order_item_order_date ON order_item (order_date);
		CREATE INDEX IF NOT EXISTS idx_order_item_create_date_time_now ON order_item (create_date_time_now);
		CREATE INDEX IF NOT EXISTS idx_order_item_branch_code ON order_item (branch_code);
		CREATE INDEX IF NOT EXISTS idx_order_item_wh_code ON order_item (wh_code)`

	if _, err := tx.ExecContext(context.Background(), commentsAndIndexesQuery); err != nil {
		return fmt.Errorf("create order_item comments and indexes: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	logger.Success("created order_item table with all indexes and comments")
	return nil
}

func TableIcWarehouseCreate(db *sql.DB) error {
	logger.Info("Creating ic_warehouse table")

	// สร้างทุกอย่างใน transaction เดียว เพื่อความเร็วสูงสุด
	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. สร้าง table ก่อน (extension ถูกสร้างไว้แล้วตอนต้น)
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS ic_warehouse (
			ignore_sync integer DEFAULT 0,
			is_lock_record integer DEFAULT 0,
			roworder serial,
			code character varying(255) NOT NULL DEFAULT ''::character varying,
			name_1 character varying(100) DEFAULT ''::character varying,
			name_2 character varying(100) DEFAULT ''::character varying,
			address character varying(255) DEFAULT ''::character varying,
			telephone character varying(250) DEFAULT ''::character varying,
			fax character varying(250) DEFAULT ''::character varying,
			user_group character varying(100) DEFAULT ''::character varying,
			wh_manager character varying(255) DEFAULT ''::character varying,
			status INT DEFAULT 0,
			guid_code character varying(100) DEFAULT ''::character varying,
			branch_code character varying(255) DEFAULT ''::character varying,
			branch_use character varying(500) DEFAULT ''::character varying,
			create_date_time_now timestamp without time zone NOT NULL DEFAULT ('now'::text)::timestamp without time zone,
			guid uuid DEFAULT gen_random_uuid(),
			latitude numeric DEFAULT 0.0,
			longitude numeric DEFAULT 0.0,
			CONSTRAINT ic_warehouse_pk_code PRIMARY KEY (code)
		)`

	if _, err := tx.ExecContext(context.Background(), createTableQuery); err != nil {
		return fmt.Errorf("create ic_warehouse table: %w", err)
	}

	// 2. สร้าง comments และ indexes ใน query เดียว
	commentsAndIndexesQuery := `
		-- สร้าง table comment
		COMMENT ON TABLE ic_warehouse IS 'ตารางเก็บข้อมูลคลังสินค้า';

		-- สร้าง column comments
		COMMENT ON COLUMN ic_warehouse.ignore_sync IS 'ไม่นำไป sync (0=sync, 1=ไม่sync)';
		COMMENT ON COLUMN ic_warehouse.is_lock_record IS 'สถานะล็อคเรคอร์ด (0=ปกติ, 1=ล็อค)';
		COMMENT ON COLUMN ic_warehouse.roworder IS 'ลำดับแถว (auto increment)';
		COMMENT ON COLUMN ic_warehouse.code IS 'รหัสคลังสินค้า';
		COMMENT ON COLUMN ic_warehouse.name_1 IS 'ชื่อคลังสินค้า (ภาษาที่ 1)';
		COMMENT ON COLUMN ic_warehouse.name_2 IS 'ชื่อคลังสินค้า (ภาษาที่ 2)';
		COMMENT ON COLUMN ic_warehouse.address IS 'ที่อยู่คลังสินค้า';
		COMMENT ON COLUMN ic_warehouse.telephone IS 'หมายเลขโทรศัพท์';
		COMMENT ON COLUMN ic_warehouse.fax IS 'หมายเลขโทรสาร';
		COMMENT ON COLUMN ic_warehouse.user_group IS 'กลุ่มผู้ใช้งาน';
		COMMENT ON COLUMN ic_warehouse.wh_manager IS 'ผู้จัดการคลังสินค้า';
		COMMENT ON COLUMN ic_warehouse.status IS 'สถานะ (0=ใช้งาน, 1=ไม่ใช้งาน)';
		COMMENT ON COLUMN ic_warehouse.guid_code IS 'รหัส GUID (ข้อความ)';
		COMMENT ON COLUMN ic_warehouse.branch_code IS 'รหัสสาขา';
		COMMENT ON COLUMN ic_warehouse.branch_use IS 'สาขาที่ใช้งาน';
		COMMENT ON COLUMN ic_warehouse.create_date_time_now IS 'วันที่และเวลาสร้าง';
		COMMENT ON COLUMN ic_warehouse.guid IS 'รหัส GUID (UUID)';
		COMMENT ON COLUMN ic_warehouse.latitude IS 'พิกัดละติจูด';
		COMMENT ON COLUMN ic_warehouse.longitude IS 'พิกัดลองจิจูด';

		-- สร้าง indexes
		CREATE INDEX IF NOT EXISTS idx_ic_warehouse_name_1 ON ic_warehouse (name_1);
		CREATE INDEX IF NOT EXISTS idx_ic_warehouse_status ON ic_warehouse (status);
		CREATE INDEX IF NOT EXISTS idx_ic_warehouse_branch_code ON ic_warehouse (branch_code);
		CREATE INDEX IF NOT EXISTS idx_ic_warehouse_wh_manager ON ic_warehouse (wh_manager);
		CREATE INDEX IF NOT EXISTS idx_ic_warehouse_user_group ON ic_warehouse (user_group);
		CREATE INDEX IF NOT EXISTS idx_ic_warehouse_guid_code ON ic_warehouse (guid_code);
		CREATE INDEX IF NOT EXISTS idx_ic_warehouse_create_date_time_now ON ic_warehouse (create_date_time_now)`

	if _, err := tx.ExecContext(context.Background(), commentsAndIndexesQuery); err != nil {
		return fmt.Errorf("create ic_warehouse comments and indexes: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	logger.Success("created ic_warehouse table with all indexes and comments")
	return nil
}

func TableIcShelfCreate(db *sql.DB) error {
	logger.Info("Creating ic_shelf table")

	// สร้างทุกอย่างใน transaction เดียว เพื่อความเร็วสูงสุด
	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. สร้าง table ก่อน (extension ถูกสร้างไว้แล้วตอนต้น)
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS ic_shelf (
			ignore_sync integer DEFAULT 0,
			is_lock_record integer DEFAULT 0,
			roworder serial,
			code character varying(25) NOT NULL DEFAULT ''::character varying,
			name_1 character varying(100) DEFAULT ''::character varying,
			name_2 character varying(100) DEFAULT ''::character varying,
			whcode character varying(10) NOT NULL DEFAULT ''::character varying,
			width character varying(10) DEFAULT ''::character varying,
			weight character varying(10) DEFAULT ''::character varying,
			height character varying(10) DEFAULT ''::character varying,
			depth character varying(10) DEFAULT ''::character varying,
			status INT DEFAULT 0,
			guid_code character varying(35) DEFAULT ''::character varying,
			stock_control INT DEFAULT 0,
			remark character varying(255) DEFAULT ''::character varying,
			create_date_time_now timestamp without time zone NOT NULL DEFAULT ('now'::text)::timestamp without time zone,
			guid uuid DEFAULT gen_random_uuid(),
			CONSTRAINT ic_shelf_pk_code PRIMARY KEY (whcode, code)
		)`

	if _, err := tx.ExecContext(context.Background(), createTableQuery); err != nil {
		return fmt.Errorf("create ic_shelf table: %w", err)
	}

	// 2. สร้าง comments และ indexes ใน query เดียว
	commentsAndIndexesQuery := `
		-- สร้าง table comment
		COMMENT ON TABLE ic_shelf IS 'ตารางเก็บข้อมูลชั้นวางสินค้าในคลัง';

		-- สร้าง column comments
		COMMENT ON COLUMN ic_shelf.ignore_sync IS 'ไม่นำไป sync (0=sync, 1=ไม่sync)';
		COMMENT ON COLUMN ic_shelf.is_lock_record IS 'สถานะล็อคเรคอร์ด (0=ปกติ, 1=ล็อค)';
		COMMENT ON COLUMN ic_shelf.roworder IS 'ลำดับแถว (auto increment)';
		COMMENT ON COLUMN ic_shelf.code IS 'รหัสชั้นวางสินค้า';
		COMMENT ON COLUMN ic_shelf.name_1 IS 'ชื่อชั้นวางสินค้า (ภาษาที่ 1)';
		COMMENT ON COLUMN ic_shelf.name_2 IS 'ชื่อชั้นวางสินค้า (ภาษาที่ 2)';
		COMMENT ON COLUMN ic_shelf.whcode IS 'รหัสคลังสินค้า';
		COMMENT ON COLUMN ic_shelf.width IS 'ความกว้าง';
		COMMENT ON COLUMN ic_shelf.weight IS 'น้ำหนักที่รองรับ';
		COMMENT ON COLUMN ic_shelf.height IS 'ความสูง';
		COMMENT ON COLUMN ic_shelf.depth IS 'ความลึก';
		COMMENT ON COLUMN ic_shelf.status IS 'สถานะ (0=ใช้งาน, 1=ไม่ใช้งาน)';
		COMMENT ON COLUMN ic_shelf.guid_code IS 'รหัส GUID (ข้อความ)';
		COMMENT ON COLUMN ic_shelf.stock_control IS 'ควบคุมสต็อก (0=ไม่ควบคุม, 1=ควบคุม)';
		COMMENT ON COLUMN ic_shelf.remark IS 'หมายเหตุ';
		COMMENT ON COLUMN ic_shelf.create_date_time_now IS 'วันที่และเวลาสร้าง';
		COMMENT ON COLUMN ic_shelf.guid IS 'รหัส GUID (UUID)';

		-- สร้าง indexes
		CREATE INDEX IF NOT EXISTS idx_ic_shelf_whcode ON ic_shelf (whcode);
		CREATE INDEX IF NOT EXISTS idx_ic_shelf_code ON ic_shelf (code);
		CREATE INDEX IF NOT EXISTS idx_ic_shelf_name_1 ON ic_shelf (name_1);
		CREATE INDEX IF NOT EXISTS idx_ic_shelf_status ON ic_shelf (status);
		CREATE INDEX IF NOT EXISTS idx_ic_shelf_stock_control ON ic_shelf (stock_control);
		CREATE INDEX IF NOT EXISTS idx_ic_shelf_guid_code ON ic_shelf (guid_code);
		CREATE INDEX IF NOT EXISTS idx_ic_shelf_create_date_time_now ON ic_shelf (create_date_time_now)`

	if _, err := tx.ExecContext(context.Background(), commentsAndIndexesQuery); err != nil {
		return fmt.Errorf("create ic_shelf comments and indexes: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	logger.Success("created ic_shelf table with all indexes and comments")
	return nil
}

func TableErpUserCreate(db *sql.DB) error {
	logger.Info("Creating erp_user table")

	// สร้างทุกอย่างใน transaction เดียว เพื่อความเร็วสูงสุด
	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. สร้าง table ก่อน
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS erp_user (
			roworder serial PRIMARY KEY,
			code character varying(50) NOT NULL UNIQUE,
			name character varying(100) NOT NULL,
			password character varying(255) NOT NULL
		)`

	if _, err := tx.ExecContext(context.Background(), createTableQuery); err != nil {
		return fmt.Errorf("create erp_user table: %w", err)
	}

	// 2. สร้าง comments และ indexes ใน query เดียว
	commentsAndIndexesQuery := `
		-- สร้าง table comment
		COMMENT ON TABLE erp_user IS 'ตารางเก็บข้อมูลผู้ใช้งานระบบ ERP';

		-- สร้าง column comments
		COMMENT ON COLUMN erp_user.roworder IS 'ลำดับแถว (auto increment primary key)';
		COMMENT ON COLUMN erp_user.code IS 'รหัสผู้ใช้ (unique)';
		COMMENT ON COLUMN erp_user.name IS 'ชื่อผู้ใช้';
		COMMENT ON COLUMN erp_user.password IS 'รหัสผ่าน (ควรเข้ารหัส)';

		-- สร้าง indexes
		CREATE INDEX IF NOT EXISTS idx_erp_user_code ON erp_user (code);
		CREATE INDEX IF NOT EXISTS idx_erp_user_name ON erp_user (name)`

	if _, err := tx.ExecContext(context.Background(), commentsAndIndexesQuery); err != nil {
		return fmt.Errorf("create erp_user comments and indexes: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	logger.Success("created erp_user table with all indexes and comments")
	return nil
}

func TableDocRefCreate(db *sql.DB) error {
	logger.Info("Creating docref table")

	// สร้างทุกอย่างใน transaction เดียว เพื่อความเร็วสูงสุด
	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. สร้าง table ก่อน
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS docref (
			id SERIAL PRIMARY KEY,
			docno TEXT,
			docnotransflag INT,
			docnoref TEXT,
			docnoreftransflag INT
		)`

	if _, err := tx.ExecContext(context.Background(), createTableQuery); err != nil {
		return fmt.Errorf("create docref table: %w", err)
	}

	// 2. สร้าง comments และ indexes ใน query เดียว
	commentsAndIndexesQuery := `
		-- สร้าง table comment
		COMMENT ON TABLE docref IS 'ตารางเก็บข้อมูลการอ้างอิงเอกสาร';

		-- สร้าง column comments
		COMMENT ON COLUMN docref.id IS 'รหัสอัตโนมัติ';
		COMMENT ON COLUMN docref.docno IS 'เอกสาร';
		COMMENT ON COLUMN docref.docnotransflag IS 'ธงสถานะการทำรายการ';
		COMMENT ON COLUMN docref.docnoref IS 'เอกสารที่อ้างอิง';
		COMMENT ON COLUMN docref.docnoreftransflag IS 'สถานะการทำรายการของเอกสารที่อ้างอิง';

		-- สร้าง indexes
		CREATE INDEX IF NOT EXISTS idx_docref_docno ON docref (docno);
		CREATE INDEX IF NOT EXISTS idx_docref_docnoref ON docref (docnoref)`

	if _, err := tx.ExecContext(context.Background(), commentsAndIndexesQuery); err != nil {
		return fmt.Errorf("create docref comments and indexes: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	logger.Success("created docref table with all indexes and comments")
	return nil
}

func TableCustomerCreate(db *sql.DB) error {
	logger.Info("Creating customer table")

	// สร้างทุกอย่างใน transaction เดียว เพื่อความเร็วสูงสุด
	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. สร้าง table ก่อน
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS customer (
			id SERIAL PRIMARY KEY,
			code TEXT,
			personaltype INT,
			customertype INT,
			name0 TEXT,
			taxid TEXT,
			telephonelist TEXT
		)`

	if _, err := tx.ExecContext(context.Background(), createTableQuery); err != nil {
		return fmt.Errorf("create customer table: %w", err)
	}

	// 2. สร้าง comments และ indexes ใน query เดียว
	commentsAndIndexesQuery := `
		-- สร้าง table comment
		COMMENT ON TABLE customer IS 'ตารางเก็บข้อมูลลูกค้า';

		-- สร้าง column comments
		COMMENT ON COLUMN customer.id IS 'รหัสอัตโนมัติ';
		COMMENT ON COLUMN customer.code IS 'รหัสลูกค้า';
		COMMENT ON COLUMN customer.personaltype IS 'ประเภทบุคคล';
		COMMENT ON COLUMN customer.customertype IS 'ประเภทลูกค้า';
		COMMENT ON COLUMN customer.name0 IS 'ชื่อ';
		COMMENT ON COLUMN customer.taxid IS 'หมายเลขประจำตัวผู้เสียภาษี';
		COMMENT ON COLUMN customer.telephonelist IS 'หมายเลขโทรศัพท์ เพื่อค้นหา';

		-- สร้าง indexes
		CREATE INDEX IF NOT EXISTS idx_customer_code ON customer (code);
		CREATE INDEX IF NOT EXISTS idx_customer_name0 ON customer (name0);
		CREATE INDEX IF NOT EXISTS idx_customer_taxid ON customer (taxid)`

	if _, err := tx.ExecContext(context.Background(), commentsAndIndexesQuery); err != nil {
		return fmt.Errorf("create customer comments and indexes: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	// pgvector column + index — แยก connection/transaction ต่างหาก เพราะถ้า pgvector
	// ไม่ได้ติดตั้ง statement นี้ fail แล้วจะทำให้ tx ทั้งก้อนด้านบน abort ไปด้วย (ทั้งที่ commit ไปแล้วก็ไม่กระทบ)
	if _, err := db.ExecContext(context.Background(), `ALTER TABLE customer ADD COLUMN IF NOT EXISTS name_embedding vector(768)`); err != nil {
		logger.Info("customer.name_embedding column skipped (pgvector extension not installed): %v", err)
	} else if _, err := db.ExecContext(context.Background(), `CREATE INDEX IF NOT EXISTS idx_customer_embedding_hnsw ON customer USING hnsw (name_embedding vector_cosine_ops) WITH (m = 16, ef_construction = 64)`); err != nil {
		logger.Info("customer embedding HNSW index skipped: %v", err)
	}

	logger.Success("created customer table with all indexes and comments")
	return nil
}

func TableDebtorCreate(db *sql.DB) error {
	// เจ้าหนี้
	logger.Info("Creating debtor table")

	// สร้างทุกอย่างใน transaction เดียว เพื่อความเร็วสูงสุด
	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. สร้าง table (ตรงกับ GORM DebtorPG model)
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS debtor (
			holding_code TEXT DEFAULT '',
			guidfixed TEXT NOT NULL DEFAULT '',
			parid TEXT DEFAULT '',
			code TEXT NOT NULL DEFAULT '',
			names JSONB DEFAULT '[]',
			taxid TEXT DEFAULT '',
			personaltype SMALLINT DEFAULT 0,
			customertype INTEGER DEFAULT 0,
			branchnumber TEXT DEFAULT '',
			fundcode TEXT DEFAULT '',
			creditday INTEGER DEFAULT 0,
			phoneprimary TEXT DEFAULT '',
			phonesecondary TEXT DEFAULT '',
			debtorbalanceamount DOUBLE PRECISION DEFAULT 0,
			email TEXT DEFAULT '',
			addressforbilling JSONB,
			PRIMARY KEY (guidfixed, code)
		)`

	if _, err := tx.ExecContext(context.Background(), createTableQuery); err != nil {
		return fmt.Errorf("create debtor table: %w", err)
	}

	// 2. สร้าง indexes + pg_trgm
	commentsAndIndexesQuery := `
		CREATE EXTENSION IF NOT EXISTS pg_trgm;
		CREATE INDEX IF NOT EXISTS idx_debtor_shop ON debtor (holding_code);
		CREATE INDEX IF NOT EXISTS idx_debtor_code_trgm ON debtor USING gin (code gin_trgm_ops);
		CREATE INDEX IF NOT EXISTS idx_debtor_taxid_trgm ON debtor USING gin (taxid gin_trgm_ops);
		CREATE INDEX IF NOT EXISTS idx_debtor_email_trgm ON debtor USING gin (email gin_trgm_ops);
		CREATE INDEX IF NOT EXISTS idx_debtor_phone_trgm ON debtor USING gin (phoneprimary gin_trgm_ops);
		CREATE INDEX IF NOT EXISTS idx_debtor_fundcode_trgm ON debtor USING gin (fundcode gin_trgm_ops);
		CREATE INDEX IF NOT EXISTS idx_debtor_names_text_trgm ON debtor USING gin ((names::text) gin_trgm_ops);
		CREATE INDEX IF NOT EXISTS idx_debtor_addr_text_trgm ON debtor USING gin ((COALESCE(addressforbilling::text, '')) gin_trgm_ops)`

	if _, err := tx.ExecContext(context.Background(), commentsAndIndexesQuery); err != nil {
		return fmt.Errorf("create debtor comments and indexes: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	// pgvector column + index — แยก connection/transaction ต่างหาก เพราะถ้า pgvector
	// ไม่ได้ติดตั้ง statement นี้ fail แล้วจะทำให้ tx ทั้งก้อนด้านบน abort ไปด้วย (ทั้งที่ commit ไปแล้วก็ไม่กระทบ)
	if _, err := db.ExecContext(context.Background(), `ALTER TABLE debtor ADD COLUMN IF NOT EXISTS name_embedding vector(768)`); err != nil {
		logger.Info("debtor.name_embedding column skipped (pgvector extension not installed): %v", err)
	} else if _, err := db.ExecContext(context.Background(), `CREATE INDEX IF NOT EXISTS idx_debtor_embedding_hnsw ON debtor USING hnsw (name_embedding vector_cosine_ops) WITH (m = 16, ef_construction = 64)`); err != nil {
		logger.Info("debtor embedding HNSW index skipped: %v", err)
	}

	logger.Success("created debtor table with all indexes and comments")
	return nil
}

func TableCreditorCreate(db *sql.DB) error {
	// เจ้าหนี้
	logger.Info("Creating creditor table")

	// สร้างทุกอย่างใน transaction เดียว เพื่อความเร็วสูงสุด
	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. สร้าง table (ตรงกับ GORM CreditorPG model)
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS creditor (
			holding_code TEXT DEFAULT '',
			guidfixed TEXT NOT NULL DEFAULT '',
			parid TEXT DEFAULT '',
			code TEXT NOT NULL DEFAULT '',
			names JSONB DEFAULT '[]',
			taxid TEXT DEFAULT '',
			personaltype SMALLINT DEFAULT 0,
			customertype INTEGER DEFAULT 0,
			branchnumber TEXT DEFAULT '',
			fundcode TEXT DEFAULT '',
			creditday INTEGER DEFAULT 0,
			phoneprimary TEXT DEFAULT '',
			phonesecondary TEXT DEFAULT '',
			creditorbalanceamount DOUBLE PRECISION DEFAULT 0,
			email TEXT DEFAULT '',
			addressforbilling JSONB,
			PRIMARY KEY (guidfixed, code)
		)`

	if _, err := tx.ExecContext(context.Background(), createTableQuery); err != nil {
		return fmt.Errorf("create creditor table: %w", err)
	}

	// 2. สร้าง indexes + pg_trgm
	commentsAndIndexesQuery := `
		CREATE EXTENSION IF NOT EXISTS pg_trgm;
		CREATE INDEX IF NOT EXISTS idx_creditor_shop ON creditor (holding_code);
		CREATE INDEX IF NOT EXISTS idx_creditor_code_trgm ON creditor USING gin (code gin_trgm_ops);
		CREATE INDEX IF NOT EXISTS idx_creditor_taxid_trgm ON creditor USING gin (taxid gin_trgm_ops);
		CREATE INDEX IF NOT EXISTS idx_creditor_email_trgm ON creditor USING gin (email gin_trgm_ops);
		CREATE INDEX IF NOT EXISTS idx_creditor_phone_trgm ON creditor USING gin (phoneprimary gin_trgm_ops);
		CREATE INDEX IF NOT EXISTS idx_creditor_fundcode_trgm ON creditor USING gin (fundcode gin_trgm_ops);
		CREATE INDEX IF NOT EXISTS idx_creditor_names_text_trgm ON creditor USING gin ((names::text) gin_trgm_ops);
		CREATE INDEX IF NOT EXISTS idx_creditor_addr_text_trgm ON creditor USING gin ((COALESCE(addressforbilling::text, '')) gin_trgm_ops)`

	if _, err := tx.ExecContext(context.Background(), commentsAndIndexesQuery); err != nil {
		return fmt.Errorf("create creditor indexes: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	// pgvector column + index — แยก connection/transaction ต่างหาก เพราะถ้า pgvector
	// ไม่ได้ติดตั้ง statement นี้ fail แล้วจะทำให้ tx ทั้งก้อนด้านบน abort ไปด้วย (ทั้งที่ commit ไปแล้วก็ไม่กระทบ)
	if _, err := db.ExecContext(context.Background(), `ALTER TABLE creditor ADD COLUMN IF NOT EXISTS name_embedding vector(768)`); err != nil {
		logger.Info("creditor.name_embedding column skipped (pgvector extension not installed): %v", err)
	} else if _, err := db.ExecContext(context.Background(), `CREATE INDEX IF NOT EXISTS idx_creditor_embedding_hnsw ON creditor USING hnsw (name_embedding vector_cosine_ops) WITH (m = 16, ef_construction = 64)`); err != nil {
		logger.Info("creditor embedding HNSW index skipped: %v", err)
	}

	logger.Success("created creditor table with all indexes and comments")
	return nil
}

func TableDocPaymentCreate(db *sql.DB) error {
	logger.Info("Creating docpayment table")

	// สร้างทุกอย่างใน transaction เดียว เพื่อความเร็วสูงสุด
	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. สร้าง table ก่อน
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS docpayment (
			id SERIAL PRIMARY KEY,
			branchid TEXT,
			docdatetime TIMESTAMPTZ,
			perioddatetime TIMESTAMPTZ,
			providername TEXT,
			amount NUMERIC(18,6),
			description TEXT,
			docno TEXT,
			transflag INTEGER,
			guidfixed TEXT DEFAULT '',
			guidbranch TEXT DEFAULT ''
		)`

	if _, err := tx.ExecContext(context.Background(), createTableQuery); err != nil {
		return fmt.Errorf("create docpayment table: %w", err)
	}

	// 2. สร้าง comments และ indexes ใน query เดียว
	commentsAndIndexesQuery := `
		-- สร้าง table comment
		COMMENT ON TABLE docpayment IS 'ตารางเก็บข้อมูลการชำระเงินของเอกสาร';

		-- สร้าง column comments
		COMMENT ON COLUMN docpayment.id IS 'รหัสอัตโนมัติ';
		COMMENT ON COLUMN docpayment.branchid IS 'รหัสสาขา';
		COMMENT ON COLUMN docpayment.docdatetime IS 'วันที่และเวลาของเอกสาร';
		COMMENT ON COLUMN docpayment.perioddatetime IS 'วันที่และเวลาของรอบบัญชี';
		COMMENT ON COLUMN docpayment.providername IS 'ชื่อผู้ให้บริการชำระเงิน';
		COMMENT ON COLUMN docpayment.amount IS 'จำนวนเงิน';
		COMMENT ON COLUMN docpayment.description IS 'รายละเอียดการชำระเงิน';
		COMMENT ON COLUMN docpayment.docno IS 'เลขที่เอกสาร';
		COMMENT ON COLUMN docpayment.transflag IS 'ประเภทธุรกรรม';
		COMMENT ON COLUMN docpayment.guidfixed IS 'รหัส GUID ที่แน่นอน';
		COMMENT ON COLUMN docpayment.guidbranch IS 'รหัส GUID ของสาขา';

		-- สร้าง indexes
		CREATE INDEX IF NOT EXISTS idx_docpayment_branchid ON docpayment (branchid);
		CREATE INDEX IF NOT EXISTS idx_docpayment_docdatetime ON docpayment (docdatetime);
		CREATE INDEX IF NOT EXISTS idx_docpayment_docno ON docpayment (docno);
		CREATE INDEX IF NOT EXISTS idx_docpayment_transflag ON docpayment (transflag);
		CREATE INDEX IF NOT EXISTS idx_docpayment_guidfixed ON docpayment (guidfixed);
		CREATE INDEX IF NOT EXISTS idx_docpayment_guidbranch ON docpayment (guidbranch)`

	if _, err := tx.ExecContext(context.Background(), commentsAndIndexesQuery); err != nil {
		return fmt.Errorf("create docpayment comments and indexes: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	logger.Success("created docpayment table with all indexes and comments")
	return nil
}

func getTableNames(db *sql.DB) ([]string, error) {
	ctx := context.Background()
	rows, err := db.QueryContext(ctx, `
        SELECT tablename
        FROM pg_catalog.pg_tables
        WHERE schemaname = 'public'
    `)
	if err != nil {
		return nil, fmt.Errorf("query table names: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, fmt.Errorf("scan table name: %w", err)
		}
		tables = append(tables, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return tables, nil
}

func containsTable(tables []string, name string) bool {
	return slices.Contains(tables, name)
}

func TableQueuesCreate(db *sql.DB) error {
	logger.Info("Creating queues table")

	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	createTableQuery := `
		CREATE TABLE IF NOT EXISTS queues (
			id BIGSERIAL PRIMARY KEY,
			holding_code VARCHAR(100) NOT NULL,
			doc_no VARCHAR(100) NOT NULL,
			trans_flag VARCHAR(10) NOT NULL,
			retry_count INTEGER NOT NULL DEFAULT 0,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			status VARCHAR(20) NOT NULL DEFAULT 'pending',
			processed_at TIMESTAMPTZ,
			error_message TEXT
		)`

	if _, err := tx.ExecContext(context.Background(), createTableQuery); err != nil {
		return fmt.Errorf("create queues table: %w", err)
	}

	indexesQuery := `
		CREATE INDEX IF NOT EXISTS idx_queues_shop_status ON queues (holding_code, status);
		CREATE INDEX IF NOT EXISTS idx_queues_created_at ON queues (created_at);
		CREATE INDEX IF NOT EXISTS idx_queues_status ON queues (status);
		CREATE INDEX IF NOT EXISTS idx_queues_pop ON queues (holding_code, created_at) WHERE status = 'pending';
		CREATE INDEX IF NOT EXISTS idx_queues_active_shops ON queues (holding_code) WHERE status = 'pending';
		CREATE INDEX IF NOT EXISTS idx_queues_processing ON queues (holding_code, processed_at) WHERE status = 'processing';
	`

	if _, err := tx.ExecContext(context.Background(), indexesQuery); err != nil {
		return fmt.Errorf("create queues indexes: %w", err)
	}

	updateFunc := `
		CREATE OR REPLACE FUNCTION set_queues_updated_at()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = NOW();
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;
	`

	if _, err := tx.ExecContext(context.Background(), updateFunc); err != nil {
		return fmt.Errorf("create queues updated_at trigger function: %w", err)
	}

	triggerQuery := `
		DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'queues_set_updated_at') THEN
				CREATE TRIGGER queues_set_updated_at
				BEFORE UPDATE ON queues
				FOR EACH ROW EXECUTE FUNCTION set_queues_updated_at();
			END IF;
		END;
		$$;
	`

	if _, err := tx.ExecContext(context.Background(), triggerQuery); err != nil {
		return fmt.Errorf("create queues updated_at trigger: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit queues creation: %w", err)
	}

	logger.Success("created queues table with indexes and trigger")
	return nil
}

func TableDeadLetterQueueCreate(db *sql.DB) error {
	logger.Info("Creating dead_letter_queue table")

	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	createTableQuery := `
		CREATE TABLE IF NOT EXISTS dead_letter_queue (
			id BIGSERIAL PRIMARY KEY,
			holding_code VARCHAR(100) NOT NULL,
			doc_no VARCHAR(100) NOT NULL,
			trans_flag VARCHAR(10) NOT NULL,
			retry_count INTEGER NOT NULL DEFAULT 0,
			created_at TIMESTAMPTZ NOT NULL,
			failed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			error_message TEXT NOT NULL
		)`

	if _, err := tx.ExecContext(context.Background(), createTableQuery); err != nil {
		return fmt.Errorf("create dead_letter_queue table: %w", err)
	}

	indexesQuery := `
		CREATE INDEX IF NOT EXISTS idx_dead_letter_queue_shop ON dead_letter_queue (holding_code);
		CREATE INDEX IF NOT EXISTS idx_dead_letter_queue_failed_at ON dead_letter_queue (failed_at);
	`

	if _, err := tx.ExecContext(context.Background(), indexesQuery); err != nil {
		return fmt.Errorf("create dead_letter_queue indexes: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit dead_letter_queue creation: %w", err)
	}

	logger.Success("created dead_letter_queue table with indexes")
	return nil
}

func TableDistributedLocksCreate(db *sql.DB) error {
	logger.Info("Creating distributed_locks table")

	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	createTableQuery := `
		CREATE TABLE IF NOT EXISTS distributed_locks (
			lock_key VARCHAR(500) PRIMARY KEY,
			owner VARCHAR(100) NOT NULL,
			acquired_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			expires_at TIMESTAMPTZ NOT NULL
		)`

	if _, err := tx.ExecContext(context.Background(), createTableQuery); err != nil {
		return fmt.Errorf("create distributed_locks table: %w", err)
	}

	indexesQuery := `
		CREATE INDEX IF NOT EXISTS idx_distributed_locks_expires_at ON distributed_locks (expires_at);
	`

	if _, err := tx.ExecContext(context.Background(), indexesQuery); err != nil {
		return fmt.Errorf("create distributed_locks indexes: %w", err)
	}

	cleanupFunc := `
		CREATE OR REPLACE FUNCTION cleanup_expired_locks()
		RETURNS INTEGER AS $$
		DECLARE
			deleted_count INTEGER;
		BEGIN
			DELETE FROM distributed_locks WHERE expires_at < NOW();
			GET DIAGNOSTICS deleted_count = ROW_COUNT;
			RETURN deleted_count;
		END;
		$$ LANGUAGE plpgsql;
	`

	if _, err := tx.ExecContext(context.Background(), cleanupFunc); err != nil {
		return fmt.Errorf("create cleanup_expired_locks function: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit distributed_locks creation: %w", err)
	}

	logger.Success("created distributed_locks table with cleanup function")
	return nil
}

func TableSearchAliasesCreate(db *sql.DB) error {
	logger.Info("Creating search_aliases table")

	createTableQuery := `
		CREATE TABLE IF NOT EXISTS search_aliases (
			id SERIAL PRIMARY KEY,
			alias TEXT NOT NULL,
			target TEXT NOT NULL,
			alias_type TEXT DEFAULT 'brand',
			created_at TIMESTAMPTZ DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_search_aliases_alias ON search_aliases (LOWER(alias));
		CREATE INDEX IF NOT EXISTS idx_search_aliases_alias_trgm ON search_aliases USING gin (alias gin_trgm_ops);
	`

	if _, err := db.ExecContext(context.Background(), createTableQuery); err != nil {
		return fmt.Errorf("create search_aliases table: %w", err)
	}

	logger.Success("created search_aliases table with indexes")
	return nil
}

func DatabaseRebuildAll(holdingCode string) {
	logger.Info("* Starting DatabaseRebuildAll for shop %s", holdingCode)

	// connect to admin/default DB (no target DB) — ใช้ database 'postgres' เป็น admin DB
	adminDB, err := mypg.ConnectAdminDatabase()
	if err != nil {
		logger.Info("Failed to connect to Postgres admin DB: %v", err)
		return
	}
	defer func() {
		_ = adminDB.Close()
	}()

	// check existence
	var exists bool
	err = adminDB.QueryRowContext(context.Background(), "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", holdingCode).Scan(&exists)
	if err != nil {
		logger.Info("Failed to check existing Postgres databases: %v", err)
		return
	}

	if !exists {
		createQuery := fmt.Sprintf(`CREATE DATABASE "%s"`, holdingCode)
		if _, err := adminDB.ExecContext(context.Background(), createQuery); err != nil {
			logger.Info("Failed to create Postgres database %s: %v", holdingCode, err)
			return
		}
		logger.Info("Created Postgres database %s", holdingCode)
	} else {
		logger.Info("Postgres database %s already exists", holdingCode)
	}

	// connect to the newly created database (connect directly to holdingCode)
	targetDB, err := mypg.PgSqlFastConnect(holdingCode)
	if err != nil {
		logger.Info("Failed to connect to Postgres database %s: %v", holdingCode, err)
		return
	}
	defer func() {
		_ = targetDB.Close()
	}()

	// ดึงชื่อ table
	tables, err := getTableNames(targetDB)
	if err != nil {
		logger.Info("Failed to get table names: %v", err)
		return
	}

	// ⭐ สร้าง extensions ทั้งหมดล่วงหน้า (เพื่อป้องกัน race condition จาก concurrent table creation)
	extensionsQuery := `
		CREATE EXTENSION IF NOT EXISTS "pgcrypto";
		CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
		CREATE EXTENSION IF NOT EXISTS "pg_trgm";
	`
	if _, err := targetDB.ExecContext(context.Background(), extensionsQuery); err != nil {
		logger.Info("Failed to create extensions: %v", err)
		return
	}
	// pgvector — สำหรับ semantic search (ถ้า extension ลงแล้ว)
	if _, err := targetDB.ExecContext(context.Background(), `CREATE EXTENSION IF NOT EXISTS "vector";`); err != nil {
		logger.Info("pgvector extension not available (install manually): %v", err)
	} else {
		logger.Info("pgvector extension ready")
	}
	logger.Info("PostgreSQL extensions ready (pgcrypto, uuid-ossp, pg_trgm)")

	// ⭐ สร้าง tables แบบ concurrent เพื่อความเร็ว
	type tableTask struct {
		name    string
		creator func(*sql.DB) error
	}

	// กลุ่มที่ 1: Tables ที่ไม่มี dependencies (สร้างพร้อมกันได้)
	independentTables := []tableTask{
		{"queues", TableQueuesCreate},
		{"dead_letter_queue", TableDeadLetterQueueCreate},
		{"distributed_locks", TableDistributedLocksCreate},
		{"product", TableProductCreate},
		{"productbarcode", TableProductBarcodeCreate},
		{"result", TableResultCreate},
		{"order_cart", TableOrderCartCreate},
		{"order_item", TableOrderItemCreate},
		{"ic_warehouse", TableIcWarehouseCreate},
		{"ic_shelf", TableIcShelfCreate},
		{"erp_user", TableErpUserCreate},
		{"customer", TableCustomerCreate},
		{"debtor", TableDebtorCreate},
		{"creditor", TableCreditorCreate},
	}

	// กลุ่มที่ 2: Tables ที่ต้องสร้างหลังจากกลุ่ม 1 เสร็จ (มี dependencies)
	dependentTables := []tableTask{
		{"doc", TableDocCreate},
		{"docdetail", TableDocDetailCreate},
		{"stockwaitprocess", TableStockWaitProcessCreate},
		{"docwaitprocess", TableDocWaitProcessCreate},
		{"processstockcost", TableProcessStockCostCreate},
		{"processstocklot", TableProcessStockLotCreate},
		{"docref", TableDocRefCreate},
		{"docpayment", TableDocPaymentCreate},
	}

	// สร้าง independent tables แบบ concurrent
	type taskResult struct {
		tableName string
		err       error
	}
	resultChan := make(chan taskResult, len(independentTables))

	for _, task := range independentTables {
		if containsTable(tables, task.name) {
			logger.Info("%s table already exists", task.name)
			resultChan <- taskResult{task.name, nil}
			continue
		}

		// สร้างแบบ concurrent
		go func(t tableTask) {
			err := t.creator(targetDB)
			resultChan <- taskResult{t.name, err}
		}(task)
	}

	// รอให้ independent tables สร้างเสร็จทั้งหมด
	independentSuccess := true
	for i := 0; i < len(independentTables); i++ {
		result := <-resultChan
		if result.err != nil {
			logger.Info("Failed to create %s table: %v", result.tableName, result.err)
			independentSuccess = false
		}
	}
	close(resultChan)

	if !independentSuccess {
		logger.Info("Some independent tables failed to create, aborting...")
		return
	}

	// สร้าง dependent tables แบบ sequential (เพราะมี dependencies)
	for _, task := range dependentTables {
		if containsTable(tables, task.name) {
			logger.Info("%s table already exists", task.name)
			continue
		}

		if err := task.creator(targetDB); err != nil {
			logger.Info("Failed to create %s table: %v", task.name, err)
			return
		}
	}

	// Inventory Costing tables (รวม inventorystockbalances ที่ /product screen ต้องใช้)
	// DDL เป็น CREATE TABLE IF NOT EXISTS อยู่แล้ว — เรียกซ้ำได้ปลอดภัย
	if err := inventory.CreateInventoryCostingTables(targetDB); err != nil {
		logger.Info("Failed to create inventory costing tables: %v", err)
	}

	logger.Info("* DatabaseRebuildAll done for %s", holdingCode)
}

func DatabaseNameIsExists(holdingCode string) bool {
	// connect to admin/default DB (no target DB) — ใช้ database 'postgres' เป็น admin DB
	adminDB, err := mypg.PgSqlFastConnect("postgres")
	if err != nil {
		return false
	}
	defer func() {
		_ = adminDB.Close()
	}()

	// check existence
	var exists bool
	err = adminDB.QueryRowContext(context.Background(), "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", holdingCode).Scan(&exists)
	if err != nil {
		return false
	}

	return exists
}

func DatabaseChecker(holdingCode string, recheckData bool) {
	// เช็คแค่ครั้งแรกต่อ shop ต่อ process (DatabaseRebuildAll เอง idempotent อยู่แล้ว
	// แต่ยังต้องเปิด admin connection + query table list ทุกครั้ง ไม่คุ้มทำซ้ำทุก Kafka message)
	if _, alreadyChecked := checkedShops.Load(holdingCode); !alreadyChecked {
		logger.Info("* Starting DatabaseChecker for shop %s", holdingCode)
		// DatabaseRebuildAll ครอบคลุมทั้ง "shop ใหม่ยังไม่มี database" และ
		// "database มีอยู่แล้วแต่ตารางบางตัวยังไม่ถูกสร้าง" (skip ตารางที่มีอยู่แล้วเอง)
		DatabaseRebuildAll(holdingCode)
		checkedShops.Store(holdingCode, true)
	}

	// recheckData = true → rebuild ข้อมูลทั้งหมดจาก MongoDB (ใช้ตอน rebuild stock เป็นต้น)
	if recheckData {
		DatabaseRebuild(holdingCode)
	}
}

func DatabaseRebuild(holdingCode string) {
	logger.Info("Database is not ready for shop %s", holdingCode)
	// Master
	DatabaseRebuildAll(holdingCode)
	ProcessErpUserRebuildAll(holdingCode)
	ProcessWarehouseRebuildAll(holdingCode)
	ProcessBarcodeRebuildAll(holdingCode)
	processProductByBarcodeBuild(holdingCode)
	ProcessCustomerRebuildAll(holdingCode)
	ProcessDebtorRebuildAll(holdingCode)
	ProcessCreditorRebuildAll(holdingCode)
	process.TruncateProcessTables(holdingCode)
	DocRebuildAllFromMongo(holdingCode, true) // สร้างทั้ง Doc และ DocDetail พร้อมกัน

	// TransFlag 54 ต้องประมวลผลแยก เพราะโครงสร้างต่างจากปกติ
	mongoClient := myglobal.SafeMongoConnectFast()
	postgresDB, err := mypg.PgSqlFastConnect(holdingCode)
	if err != nil {
		logger.Error("Failed to connect to PostgreSQL: %v", err)
	} else {
		DocDetailTransFlag54RebuildFromMongo(mongoClient, postgresDB, holdingCode)
	}

	go processstock.ProcessStockCostAll(holdingCode)

	// ประมวลผลสถานะเอกสารหลัง rebuild เสร็จ (ใช้ BATCH MODE - เร็วกว่า!)
	logger.Info("Processing document status after rebuild for shop %s (BATCH MODE)", holdingCode)
	processdoc.ProcessDocumentStatusByShopBatch(holdingCode)

	// Sync สถานะการอนุมัติจาก MongoDB po_approval_status ไปยัง PostgreSQL
	SyncApprovalStatusToPostgreSQL(holdingCode)
}

// SyncApprovalStatusToPostgreSQL — ดึง approval status จาก MongoDB แล้ว UPDATE กลับไปที่ doc table ใน PostgreSQL
func SyncApprovalStatusToPostgreSQL(holdingCode string) {
	logger.Info("[Rebuild] เริ่ม sync approval status สำหรับ shop %s", holdingCode)

	// ดึง approval status จาก MongoDB po_approval_status collection
	statusMap, err := approval.GetApprovalStatusMapByShop(holdingCode)
	if err != nil {
		logger.Warn("[Rebuild] ไม่สามารถดึง approval status: %v", err)
		return
	}

	if len(statusMap) == 0 {
		logger.Info("[Rebuild] ไม่มี approval status สำหรับ shop %s", holdingCode)
		return
	}

	// เชื่อมต่อ PostgreSQL
	db, err := mypg.PgSqlFastConnect(holdingCode)
	if err != nil {
		logger.Error("[Rebuild] เชื่อมต่อ PostgreSQL ล้มเหลว: %v", err)
		return
	}
	defer db.Close()

	// UPDATE doc table ทีละ docno
	updatedCount := 0
	for docNo, status := range statusMap {
		result, err := db.ExecContext(context.Background(),
			"UPDATE doc SET approval_status = $1 WHERE docno = $2 AND approval_status != $1",
			status, docNo)
		if err != nil {
			logger.Warn("[Rebuild] UPDATE approval_status ล้มเหลว สำหรับ %s: %v", docNo, err)
			continue
		}
		if rowsAffected, _ := result.RowsAffected(); rowsAffected > 0 {
			updatedCount++
		}
	}

	logger.Success("[Rebuild] sync approval status เสร็จ: อัพเดท %d/%d เอกสาร", updatedCount, len(statusMap))
}

// RebuildProductsOnly - Rebuild เฉพาะสินค้า (PostgreSQL, ClickHouse)
func RebuildProductsOnly(holdingCode string) error {
	logger.Info("=== Starting Products Only Rebuild for shop %s ===", holdingCode)

	// 0. สร้าง database ถ้ายังไม่มี
	logger.Info("[0/3] Ensuring database exists...")
	adminDB, err := mypg.ConnectAdminDatabase()
	if err != nil {
		return fmt.Errorf("failed to connect to admin database: %v", err)
	}

	// ตรวจสอบว่า database มีอยู่หรือไม่
	var exists bool
	err = adminDB.QueryRowContext(context.Background(), "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", holdingCode).Scan(&exists)
	if err != nil {
		adminDB.Close()
		return fmt.Errorf("failed to check database existence: %v", err)
	}

	if !exists {
		// สร้าง database
		createQuery := fmt.Sprintf(`CREATE DATABASE "%s"`, holdingCode)
		if _, err := adminDB.ExecContext(context.Background(), createQuery); err != nil {
			adminDB.Close()
			return fmt.Errorf("failed to create database %s: %v", holdingCode, err)
		}
		logger.Info("✓ Created database %s", holdingCode)
	} else {
		logger.Info("✓ Database %s already exists", holdingCode)
	}
	adminDB.Close()

	// 1. เชื่อมต่อ PostgreSQL
	db, err := mypg.PgSqlFastConnect(holdingCode)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %v", err)
	}

	// สร้าง productbarcode table ถ้ายังไม่มี
	logger.Info("Ensuring productbarcode table exists...")
	err = TableProductBarcodeCreate(db)
	if err != nil {
		logger.Error("Failed to create productbarcode table: %v", err)
	}

	// สร้าง search_aliases table สำหรับ alias expansion (ทีโอเอ→TOA)
	logger.Info("Ensuring search_aliases table exists...")
	err = TableSearchAliasesCreate(db)
	if err != nil {
		logger.Error("Failed to create search_aliases table: %v", err)
	}

	// เพิ่ม column price_retail ถ้ายังไม่มี (สำหรับ table เก่าที่สร้างก่อนมี field นี้)
	logger.Info("Ensuring price_retail column exists in PostgreSQL...")
	alterPgQuery := "ALTER TABLE productbarcode ADD COLUMN IF NOT EXISTS price_retail NUMERIC(18,2) DEFAULT 0"
	_, err = db.Exec(alterPgQuery)
	if err != nil {
		logger.Error("Failed to add price_retail column to PostgreSQL: %v", err)
	} else {
		logger.Info("✓ PostgreSQL price_retail column checked/added")
	}

	// 2. Rebuild Barcode และ Product ใน PostgreSQL
	logger.Info("[1/3] Rebuilding products in PostgreSQL...")
	ProcessBarcodeRebuildAll(holdingCode)
	processProductByBarcodeBuild(holdingCode)
	logger.Info("✓ PostgreSQL products rebuild completed")

	// 3. Rebuild ProductBarcode ใน ClickHouse
	logger.Info("[2/3] Rebuilding products in ClickHouse...")
	clickHouseDB, err := myclickhouse.ClickHouseFastConnect()
	if err != nil {
		logger.Error("Failed to connect to ClickHouse: %v", err)
	} else {
		// เพิ่ม column price_retail ถ้ายังไม่มี (สำหรับ table เก่าที่สร้างก่อนมี field นี้)
		alterQuery := fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS price_retail Float64 DEFAULT 0", myclickhouse.TableName("productbarcode"))
		logger.Info("Ensuring price_retail column exists...")
		err = myclickhouse.ExecuteCommand(context.Background(), clickHouseDB, alterQuery)
		if err != nil {
			logger.Error("Failed to add price_retail column: %v", err)
		} else {
			logger.Info("✓ price_retail column checked/added")
		}

		// ลบข้อมูลเก่าก่อน
		deleteQuery := fmt.Sprintf("ALTER TABLE %s DELETE WHERE holding_code = '%s'", myclickhouse.TableName("productbarcode"), holdingCode)
		logger.Info("Executing: %s", deleteQuery)
		err = myclickhouse.ExecuteCommand(context.Background(), clickHouseDB, deleteQuery)
		if err != nil {
			logger.Error("Failed to delete old ClickHouse data: %v", err)
		}

		// Insert ข้อมูลใหม่จาก PostgreSQL
		err = RebuildClickHouseProductBarcode(db, holdingCode)
		if err != nil {
			logger.Error("Failed to rebuild ClickHouse products: %v", err)
		} else {
			logger.Info("✓ ClickHouse products rebuild completed")
		}
	}

	logger.Info("=== Products Only Rebuild Completed for shop %s ===", holdingCode)
	return nil
}

// RebuildClickHouseProductBarcode - Rebuild productbarcode table ใน ClickHouse
func RebuildClickHouseProductBarcode(db *sql.DB, holdingCode string) error {
	ctx := context.Background()

	// ดึงข้อมูลจาก PostgreSQL productbarcode table
	query := `
		SELECT
			barcode,
			itemcode,
			name0,
			unitcode,
			unitname,
			COALESCE(groupcode, '') as groupcode,
			COALESCE(price1, 0) as price1,
			COALESCE(price_retail, 0) as price_retail
		FROM productbarcode
		WHERE barcode IS NOT NULL AND barcode != ''
		ORDER BY itemcode
	`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("error querying products: %v", err)
	}
	defer rows.Close()

	clickHouseDB, err := myclickhouse.ClickHouseFastConnect()
	if err != nil {
		return fmt.Errorf("failed to connect to ClickHouse: %v", err)
	}

	// Insert ข้อมูลทีละ batch
	batchSize := 1000
	values := make([]string, 0, batchSize)
	count := 0

	for rows.Next() {
		var barcode, itemcode, name0, unitcode, unitname, groupcode string
		var price1, priceRetail float64

		err := rows.Scan(&barcode, &itemcode, &name0, &unitcode, &unitname, &groupcode, &price1, &priceRetail)
		if err != nil {
			logger.Error("Error scanning product: %v", err)
			continue
		}

		// เตรียม value string สำหรับ INSERT
		valueStr := fmt.Sprintf("('%s', '%s', '%s', '%s', '%s', '%s', '%s', %.2f, %.2f)",
			holdingCode, barcode, itemcode, name0, unitcode, unitname, groupcode, price1, priceRetail)
		values = append(values, valueStr)
		count++

		// เมื่อครบ batch ให้ insert
		if len(values) >= batchSize {
			insertQuery := fmt.Sprintf(
				"INSERT INTO %s (holding_code, barcode, itemcode, name0, unitcode, unitname, groupcode, price1, price_retail) VALUES %s", myclickhouse.TableName("productbarcode"),
				fmt.Sprintf("%s", values[0]))

			for i := 1; i < len(values); i++ {
				insertQuery += ", " + values[i]
			}

			err = myclickhouse.ExecuteCommand(ctx, clickHouseDB, insertQuery)
			if err != nil {
				logger.Error("Failed to insert batch to ClickHouse: %v", err)
			} else {
				logger.Info("Inserted %d products to ClickHouse", len(values))
			}

			values = make([]string, 0, batchSize)
		}
	}

	// Insert batch สุดท้าย
	if len(values) > 0 {
		insertQuery := fmt.Sprintf(
			"INSERT INTO %s (holding_code, barcode, itemcode, name0, unitcode, unitname, groupcode, price1, price_retail) VALUES %s", myclickhouse.TableName("productbarcode"),
			fmt.Sprintf("%s", values[0]))

		for i := 1; i < len(values); i++ {
			insertQuery += ", " + values[i]
		}

		err = myclickhouse.ExecuteCommand(ctx, clickHouseDB, insertQuery)
		if err != nil {
			logger.Error("Failed to insert final batch to ClickHouse: %v", err)
		} else {
			logger.Info("Inserted final %d products to ClickHouse", len(values))
		}
	}

	logger.Info("Total %d products inserted to ClickHouse", count)
	return nil
}

func PgSqlDropDatabaseAndReProcess(holdingCode string) {
	// ถ้า dropTable = true ให้ลบ database เดิมทิ้ง แล้วสร้างใหม่ ถ้าไม่ใช่ ให้ truncate
	logger.Info("* Starting PgSqlDatabaseRebuildAndReProcess for shop %s", holdingCode)

	db, err := mypg.ConnectAdminDatabase()
	if err != nil {
		logger.Error("connecting to PostgreSQL: %v", err)
		return
	}
	defer db.Close()

	// ✅ เพิ่ม double quotes รอบ holdingCode
	dropDatabaseSQL := fmt.Sprintf(`DROP DATABASE IF EXISTS "%s" WITH (FORCE)`, holdingCode)
	logger.Info("Executing: %s", dropDatabaseSQL)

	if _, err := db.Exec(dropDatabaseSQL); err != nil {
		logger.Error("dropping database %s: %v", holdingCode, err)
		return
	}
	logger.Info("✓ Dropped database: %s (forced)", holdingCode)
	logger.Info("* PgSqlDatabaseDrop completed for shop %s", holdingCode)

	// ลบบน ClickHouse ด้วย (ลบด้วย holdingCode)
	tableNameOnClickHouse := []string{"doc", "docdetail", "docref", "docpayment", "productbarcode", "ic_inventory"}
	// เชื่อมต่อ ClickHouse ตรง
	clickHouseDB, err := myclickhouse.ClickHouseFastConnect()
	if err != nil {
		logger.Error("Failed to connect to ClickHouse: %v", err)
		return
	}
	// ใช้ connection pool ไม่ต้อง close

	for _, tableName := range tableNameOnClickHouse {
		deleteByHoldingCodeQuery := fmt.Sprintf("ALTER TABLE %s DELETE WHERE holding_code = '%s'", myclickhouse.TableName(tableName), holdingCode)
		logger.Info("Executing: %s", deleteByHoldingCodeQuery)
		err = myclickhouse.ExecuteCommand(context.Background(), clickHouseDB, deleteByHoldingCodeQuery)
		if err != nil {
			logger.Error("Failed to execute ClickHouse query %s: %v", deleteByHoldingCodeQuery, err)
		}
	}

	// เรียก DatabaseRebuild
	DatabaseRebuild(holdingCode)
}

// CalcStockCostAll - คำนวณ stock cost และ document status ทั้งหมด (ไม่ drop database)
func CalcStockCostAll(holdingCode string) {
	logger.Info("=== Starting CalcStockCostAll for shop %s ===", holdingCode)

	// 1. คำนวณต้นทุนสต็อกทั้งหมด
	logger.Info("Step 1: Processing stock cost for all items...")
	processstock.ProcessStockCostAll(holdingCode)

	// 2. ประมวลผลสถานะเอกสาร
	logger.Info("Step 2: Processing document status...")
	processdoc.ProcessDocumentStatusByShopBatch(holdingCode)

	logger.Success("=== CalcStockCostAll completed for shop %s ===", holdingCode)
}

// CalcStockCostForItems - คำนวณ stock cost เฉพาะ item codes ที่ระบุ
func CalcStockCostForItems(holdingCode string, itemCodes []string) {
	logger.Info("=== Starting CalcStockCostForItems for shop %s (%d items) ===", holdingCode, len(itemCodes))

	db, err := mypg.PgSqlFastConnect(holdingCode)
	if err != nil {
		logger.Error("Failed to connect to PostgreSQL: %v", err)
		return
	}

	pointQty := myglobal.ConfigSystem.StockQtyPoint
	pointAmount := myglobal.ConfigSystem.StockAmountPoint
	pointCost := myglobal.ConfigSystem.StockCostPoint

	// คำนวณต้นทุนทีละ item
	for i, itemCode := range itemCodes {
		logger.Info("Processing item %d/%d: %s", i+1, len(itemCodes), itemCode)
		processstock.ProductCalcCost(db, holdingCode, itemCode, pointQty, pointAmount, pointCost, true)
	}

	// ประมวลผลสถานะเอกสาร
	logger.Info("Processing document status for shop %s", holdingCode)
	processdoc.ProcessDocumentStatusByShopBatch(holdingCode)

	logger.Success("=== CalcStockCostForItems completed for shop %s (%d items) ===", holdingCode, len(itemCodes))
}

// =====================================================================
// WithProgress versions — เหมือน functions เดิม แต่ส่ง progress events
// =====================================================================

// PgSqlDropDatabaseAndReProcessWithProgress — Full rebuild พร้อม progress tracking
func PgSqlDropDatabaseAndReProcessWithProgress(holdingCode string, job *RebuildJob) {
	totalSteps := 16 // 2 (drop PG + drop CH) + 14 (rebuild)

	// Step 1: Drop PostgreSQL database
	job.SendProgress(1, totalSteps, "ลบฐานข้อมูล PostgreSQL", "running")
	logger.Info("* Starting PgSqlDatabaseRebuildAndReProcess for shop %s", holdingCode)

	db, err := mypg.ConnectAdminDatabase()
	if err != nil {
		logger.Error("connecting to PostgreSQL: %v", err)
		job.SendError(1, totalSteps, "ลบฐานข้อมูล PostgreSQL", err)
		return
	}
	defer db.Close()

	dropDatabaseSQL := fmt.Sprintf(`DROP DATABASE IF EXISTS "%s" WITH (FORCE)`, holdingCode)
	logger.Info("Executing: %s", dropDatabaseSQL)

	if _, err := db.Exec(dropDatabaseSQL); err != nil {
		logger.Error("dropping database %s: %v", holdingCode, err)
		job.SendError(1, totalSteps, "ลบฐานข้อมูล PostgreSQL", err)
		return
	}
	logger.Info("✓ Dropped database: %s (forced)", holdingCode)

	// Step 2: Delete ClickHouse data
	job.SendProgress(2, totalSteps, "ลบข้อมูล ClickHouse", "running")
	tableNameOnClickHouse := []string{"doc", "docdetail", "docref", "docpayment", "productbarcode", "ic_inventory"}
	clickHouseDB, err := myclickhouse.ClickHouseFastConnect()
	if err != nil {
		logger.Error("Failed to connect to ClickHouse: %v", err)
		job.SendError(2, totalSteps, "ลบข้อมูล ClickHouse", err)
		return
	}

	for _, tableName := range tableNameOnClickHouse {
		deleteByHoldingCodeQuery := fmt.Sprintf("ALTER TABLE %s DELETE WHERE holding_code = '%s'", myclickhouse.TableName(tableName), holdingCode)
		logger.Info("Executing: %s", deleteByHoldingCodeQuery)
		err = myclickhouse.ExecuteCommand(context.Background(), clickHouseDB, deleteByHoldingCodeQuery)
		if err != nil {
			logger.Error("Failed to execute ClickHouse query %s: %v", deleteByHoldingCodeQuery, err)
		}
	}

	// Step 3-15: DatabaseRebuild steps
	DatabaseRebuildWithProgress(holdingCode, job, 2, totalSteps)
}

// DatabaseRebuildWithProgress — Rebuild ทั้งหมดพร้อม progress tracking
// stepOffset = จำนวน steps ก่อนหน้า (เพื่อนับ step ต่อจากที่ drop เสร็จ)
func DatabaseRebuildWithProgress(holdingCode string, job *RebuildJob, stepOffset int, totalSteps int) {
	logger.Info("* Starting DatabaseRebuild for shop %s", holdingCode)

	// Step: สร้างโครงสร้างฐานข้อมูล
	job.SendProgress(stepOffset+1, totalSteps, "สร้างโครงสร้างฐานข้อมูล", "running")
	DatabaseRebuildAll(holdingCode)

	// Step: นำเข้าข้อมูลผู้ใช้ ERP
	job.SendProgress(stepOffset+2, totalSteps, "นำเข้าข้อมูลผู้ใช้ ERP", "running")
	ProcessErpUserRebuildAll(holdingCode)

	// Step: นำเข้าข้อมูลคลังสินค้า
	job.SendProgress(stepOffset+3, totalSteps, "นำเข้าข้อมูลคลังสินค้า", "running")
	ProcessWarehouseRebuildAll(holdingCode)

	// Step: นำเข้าข้อมูลบาร์โค้ด
	job.SendProgress(stepOffset+4, totalSteps, "นำเข้าข้อมูลบาร์โค้ด", "running")
	ProcessBarcodeRebuildAll(holdingCode)

	// Step: สร้างข้อมูลสินค้า
	job.SendProgress(stepOffset+5, totalSteps, "สร้างข้อมูลสินค้า", "running")
	processProductByBarcodeBuild(holdingCode)

	// Step: นำเข้าข้อมูลลูกค้า
	job.SendProgress(stepOffset+6, totalSteps, "นำเข้าข้อมูลลูกค้า", "running")
	ProcessCustomerRebuildAll(holdingCode)

	// Step: นำเข้าข้อมูลลูกหนี้
	job.SendProgress(stepOffset+7, totalSteps, "นำเข้าข้อมูลลูกหนี้", "running")
	debtorCount := ProcessDebtorRebuildAll(holdingCode)
	job.SendProgress(stepOffset+7, totalSteps, fmt.Sprintf("นำเข้าข้อมูลลูกหนี้ (%d รายการ → PG + ClickHouse)", debtorCount), "running")

	// Step: นำเข้าข้อมูลเจ้าหนี้
	job.SendProgress(stepOffset+8, totalSteps, "นำเข้าข้อมูลเจ้าหนี้", "running")
	creditorCount := ProcessCreditorRebuildAll(holdingCode)
	job.SendProgress(stepOffset+8, totalSteps, fmt.Sprintf("นำเข้าข้อมูลเจ้าหนี้ (%d รายการ → PG + ClickHouse)", creditorCount), "running")

	// Step: ล้างตารางประมวลผลชั่วคราว
	job.SendProgress(stepOffset+9, totalSteps, "ล้างตารางประมวลผลชั่วคราว", "running")
	process.TruncateProcessTables(holdingCode)

	// Step: นำเข้าเอกสารทั้งหมดจาก MongoDB (พร้อม sub-progress แสดงประเภทเอกสาร)
	job.SendProgress(stepOffset+10, totalSteps, "นำเข้าเอกสารทั้งหมดจาก MongoDB", "running")
	DocRebuildAllFromMongoWithCallback(holdingCode, true, func(current, total int, transName string) {
		detail := fmt.Sprintf("%d/%d ประเภท — %s", current, total, transName)
		job.SendProgressDetail(stepOffset+10, totalSteps, "นำเข้าเอกสารทั้งหมดจาก MongoDB", detail)
	})

	// Step: นำเข้าเอกสาร TransFlag 54
	job.SendProgress(stepOffset+11, totalSteps, "นำเข้าเอกสาร TransFlag 54", "running")
	mongoClient := myglobal.SafeMongoConnectFast()
	postgresDB, err := mypg.PgSqlFastConnect(holdingCode)
	if err != nil {
		logger.Error("Failed to connect to PostgreSQL: %v", err)
	} else {
		DocDetailTransFlag54RebuildFromMongo(mongoClient, postgresDB, holdingCode)
	}

	// Step: คำนวณต้นทุนสต็อก (พร้อม sub-progress แสดงจำนวนรายการ)
	job.SendProgress(stepOffset+12, totalSteps, "คำนวณต้นทุนสต็อก", "running")
	processstock.ProcessStockCostAllWithCallback(holdingCode, func(processed, total int) {
		detail := fmt.Sprintf("%d/%d รายการ", processed, total)
		job.SendProgressDetail(stepOffset+12, totalSteps, "คำนวณต้นทุนสต็อก", detail)
	})

	// Step: ประมวลผลสถานะเอกสาร
	job.SendProgress(stepOffset+13, totalSteps, "ประมวลผลสถานะเอกสาร", "running")
	processdoc.ProcessDocumentStatusByShopBatch(holdingCode)

	// Step: Sync สถานะการอนุมัติจาก MongoDB
	job.SendProgress(stepOffset+14, totalSteps, "Sync สถานะการอนุมัติ", "running")
	SyncApprovalStatusToPostgreSQL(holdingCode)

	// เสร็จสิ้น
	job.SendProgress(totalSteps, totalSteps, "เสร็จสิ้น", "completed")
	logger.Success("* DatabaseRebuild completed for shop %s", holdingCode)
}

// CalcStockCostAllWithProgress — คำนวณ stock cost ทั้งหมดพร้อม progress
func CalcStockCostAllWithProgress(holdingCode string, job *RebuildJob) {
	totalSteps := 2

	job.SendProgress(1, totalSteps, "คำนวณต้นทุนสต็อกทั้งหมด", "running")
	processstock.ProcessStockCostAllWithCallback(holdingCode, func(processed, total int) {
		detail := fmt.Sprintf("%d/%d รายการ", processed, total)
		job.SendProgressDetail(1, totalSteps, "คำนวณต้นทุนสต็อกทั้งหมด", detail)
	})

	job.SendProgress(2, totalSteps, "ประมวลผลสถานะเอกสาร", "running")
	processdoc.ProcessDocumentStatusByShopBatch(holdingCode)

	job.SendProgress(2, totalSteps, "เสร็จสิ้น", "completed")
}

// CalcStockCostForItemsWithProgress — คำนวณ stock cost เฉพาะ items พร้อม progress
func CalcStockCostForItemsWithProgress(holdingCode string, itemCodes []string, job *RebuildJob) {
	totalSteps := len(itemCodes) + 1 // items + document status

	db, err := mypg.PgSqlFastConnect(holdingCode)
	if err != nil {
		logger.Error("Failed to connect to PostgreSQL: %v", err)
		job.SendError(1, totalSteps, "เชื่อมต่อฐานข้อมูล", err)
		return
	}

	pointQty := myglobal.ConfigSystem.StockQtyPoint
	pointAmount := myglobal.ConfigSystem.StockAmountPoint
	pointCost := myglobal.ConfigSystem.StockCostPoint

	for i, itemCode := range itemCodes {
		stepName := fmt.Sprintf("คำนวณสินค้า %s (%d/%d)", itemCode, i+1, len(itemCodes))
		job.SendProgress(i+1, totalSteps, stepName, "running")
		processstock.ProductCalcCost(db, holdingCode, itemCode, pointQty, pointAmount, pointCost, true)
	}

	job.SendProgress(totalSteps, totalSteps, "ประมวลผลสถานะเอกสาร", "running")
	processdoc.ProcessDocumentStatusByShopBatch(holdingCode)

	job.SendProgress(totalSteps, totalSteps, "เสร็จสิ้น", "completed")
}

// RebuildDocumentFlowWithProgress — คำนวณ flow เอกสาร (isref, iscomparedsuccess, isclosed) พร้อม progress
// ใช้สำหรับคำนวณสถานะเอกสารใบสั่งซื้อว่ามีการอ้างอิง (รับสินค้า) หรือยัง
// แสดงผลทีละเอกสาร
func RebuildDocumentFlowWithProgress(holdingCode string, job *RebuildJob) {
	// Step 1: คำนวณด้วย MEGA-QUERY (เร็ว)
	job.SendProgress(1, 3, "คำนวณ flow เอกสารใบสั่งซื้อ", "running")
	processdoc.ProcessDocumentStatusByShopBatch(holdingCode)

	// Step 2: ดึงผลลัพธ์ทีละเอกสารมาแสดง
	job.SendProgress(2, 3, "กำลังดึงผลลัพธ์ทีละเอกสาร", "running")

	db, err := mypg.PgSqlFastConnect(holdingCode)
	if err != nil {
		logger.Error("เชื่อมต่อ PostgreSQL ไม่ได้: %v", err)
		job.SendProgress(3, 3, "เสร็จสิ้น (ไม่สามารถดึงรายละเอียดได้)", "completed")
		return
	}

	rows, err := db.Query("SELECT docno, COALESCE(isref, false), COALESCE(iscomparedsuccess, 0), COALESCE(isclosed, false), COALESCE(isclosedmanual, false) FROM doc WHERE transflag = 6 ORDER BY docdatetime DESC")
	if err != nil {
		logger.Error("ดึงข้อมูลเอกสาร PO ไม่ได้: %v", err)
		job.SendProgress(3, 3, "เสร็จสิ้น (ไม่สามารถดึงรายละเอียดได้)", "completed")
		return
	}
	defer rows.Close()

	type docResult struct {
		DocNo             string
		IsRef             bool
		IsComparedSuccess int
		IsClosed          bool
		IsClosedManual    bool
	}

	var results []docResult
	for rows.Next() {
		var r docResult
		if err := rows.Scan(&r.DocNo, &r.IsRef, &r.IsComparedSuccess, &r.IsClosed, &r.IsClosedManual); err != nil {
			continue
		}
		results = append(results, r)
	}

	totalDocs := len(results)
	if totalDocs == 0 {
		job.SendProgress(3, 3, "ไม่พบเอกสารใบสั่งซื้อ", "completed")
		return
	}

	// แสดงผลทีละเอกสาร
	totalSteps := totalDocs + 2 // +2 สำหรับ step คำนวณ + step เสร็จสิ้น
	for i, r := range results {
		// สร้างข้อความสถานะ
		status := "ยังไม่มีการอ้างอิง"
		switch {
		case r.IsClosedManual:
			status = "ปิดด้วยมือ"
		case r.IsComparedSuccess == 1:
			status = "รับครบแล้ว"
		case r.IsComparedSuccess == 2:
			status = "รับบางส่วน"
		case r.IsComparedSuccess == 3:
			status = "รับเกินจำนวน"
		case r.IsRef:
			status = "มีการอ้างอิงแล้ว"
		}

		stepName := fmt.Sprintf("%s — %s", r.DocNo, status)
		job.SendProgress(i+2, totalSteps, stepName, "running")
	}

	job.SendProgress(totalSteps, totalSteps, fmt.Sprintf("เสร็จสิ้น (%d เอกสาร)", totalDocs), "completed")
}
