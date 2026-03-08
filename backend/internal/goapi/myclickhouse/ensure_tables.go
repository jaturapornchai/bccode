package myclickhouse

import (
	"context"
	"fmt"
	"smlcloudplatform/internal/goapi/logger"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
)

// coreTableNames รายชื่อ table หลักที่ต้องมีใน ClickHouse
// ใช้ตรวจสอบว่า database มี table ครบหรือยัง
var coreTableNames = []string{
	"doc",
	"docdetail",
	"docdetail_updated",
	"docpayment",
	"docref",
	"productbarcode",
	"productbarcodeprocess",
	"productbarcoderef",
	"productbarcodeimport",
	"processstockcost",
	"processstockdetail",
	"processstocklot",
	"processstock",
	"stockwaitprocess",
	"cartorder",
	"cartorderdetail",
	"creditors",
	"debtors",
	"warehouses",
	"locations",
	"salechannel",
	"shop",
	"token",
	"userlogin",
	"ic_inventory",
	"result",
	"resultfordashboard",
	"task_status",
	"temp_barcode_map",
}

// getCreateTableDDLs คืน DDL statements สำหรับสร้าง tables ทั้งหมด
// ใช้ {DB} เป็น placeholder แทนชื่อ database
func getCreateTableDDLs() []string {
	return []string{
		// cartorder
		`CREATE TABLE IF NOT EXISTS {DB}.cartorder (
			` + "`usercode`" + ` String,
			` + "`shopid`" + ` String,
			` + "`cartnumber`" + ` String,
			` + "`totalamount`" + ` Float64,
			` + "`createdatetime`" + ` DateTime,
			` + "`custcode`" + ` String,
			` + "`remark`" + ` String
		) ENGINE = MergeTree PARTITION BY shopid ORDER BY (shopid, cartnumber) SETTINGS index_granularity = 8192`,

		// cartorderdetail
		`CREATE TABLE IF NOT EXISTS {DB}.cartorderdetail (
			` + "`usercode`" + ` String,
			` + "`shopid`" + ` String,
			` + "`barcode`" + ` String,
			` + "`price`" + ` Float64,
			` + "`pricemember`" + ` Float64,
			` + "`qty`" + ` Float64,
			` + "`totalamount`" + ` Float64,
			` + "`totalamountmember`" + ` Float64,
			` + "`discountamount`" + ` Float64,
			` + "`discountword`" + ` String,
			` + "`unitcode`" + ` String,
			` + "`remark`" + ` String,
			` + "`createdatetime`" + ` DateTime,
			` + "`cartnumber`" + ` String,
			` + "`name`" + ` String,
			` + "`unitname`" + ` String
		) ENGINE = MergeTree PARTITION BY shopid ORDER BY (shopid, barcode) SETTINGS index_granularity = 8192`,

		// creditors
		`CREATE TABLE IF NOT EXISTS {DB}.creditors (
			` + "`shopid`" + ` String,
			` + "`guidfixed`" + ` String,
			` + "`code`" + ` String,
			` + "`name1`" + ` String,
			` + "`name2`" + ` String
		) ENGINE = MergeTree PARTITION BY shopid ORDER BY shopid SETTINGS index_granularity = 8192`,

		// debtors
		`CREATE TABLE IF NOT EXISTS {DB}.debtors (
			` + "`shopid`" + ` String,
			` + "`guidfixed`" + ` String,
			` + "`code`" + ` String,
			` + "`name1`" + ` String,
			` + "`name2`" + ` String
		) ENGINE = MergeTree PARTITION BY shopid ORDER BY shopid SETTINGS index_granularity = 8192`,

		// doc
		`CREATE TABLE IF NOT EXISTS {DB}.doc (
			` + "`shopid`" + ` String,
			` + "`docno`" + ` String,
			` + "`docdatetime`" + ` DateTime,
			` + "`perioddatetime`" + ` DateTime,
			` + "`taxdocno`" + ` String DEFAULT '',
			` + "`totalamount`" + ` Float64,
			` + "`roundamount`" + ` Float64 DEFAULT 0,
			` + "`paytype`" + ` Int16 DEFAULT 0,
			` + "`paycashamount`" + ` Float64 DEFAULT 0,
			` + "`paycashchange`" + ` Float64 DEFAULT 0,
			` + "`paycashbalance`" + ` Float64 DEFAULT 0,
			` + "`deliverycode`" + ` String DEFAULT '',
			` + "`checksum`" + ` String,
			` + "`branchid`" + ` String DEFAULT '',
			` + "`slipurl`" + ` String DEFAULT '',
			` + "`salechannelcode`" + ` String,
			` + "`deliveryamount`" + ` Float64,
			` + "`iscancel`" + ` Bool DEFAULT false,
			` + "`cancelreason`" + ` String DEFAULT '',
			` + "`guidpos`" + ` String DEFAULT '',
			` + "`guidbranch`" + ` String DEFAULT '',
			` + "`guidfixed`" + ` String DEFAULT '',
			` + "`transflag`" + ` Int16 DEFAULT 0,
			` + "`isdelete`" + ` Bool DEFAULT 0,
			` + "`currency`" + ` String DEFAULT '',
			` + "`currency_symbol`" + ` String DEFAULT '',
			` + "`doc_currency`" + ` String DEFAULT '',
			` + "`doc_currency_symbol`" + ` String DEFAULT '',
			` + "`exchange_rate`" + ` Float64 DEFAULT 1,
			` + "`totalamount_doc`" + ` Float64 DEFAULT 0,
			` + "`approval_status`" + ` String DEFAULT '',
			` + "`isclosedmanual`" + ` Bool DEFAULT false,
			` + "`closedmanual_by_code`" + ` String DEFAULT '',
			` + "`closedmanual_by_name`" + ` String DEFAULT '',
			` + "`closedmanual_at`" + ` DateTime DEFAULT '1970-01-01 00:00:00',
			` + "`closedmanual_reason`" + ` String DEFAULT '',
			INDEX idx_shopid shopid TYPE minmax GRANULARITY 4,
			INDEX idx_branchid branchid TYPE minmax GRANULARITY 4,
			INDEX idx_perioddatetime perioddatetime TYPE minmax GRANULARITY 4,
			INDEX idx_shopid_transflag (shopid, transflag) TYPE minmax GRANULARITY 4
		) ENGINE = ReplacingMergeTree PARTITION BY shopid ORDER BY (guidfixed, shopid, docno) SETTINGS index_granularity = 8192`,

		// docdetail
		`CREATE TABLE IF NOT EXISTS {DB}.docdetail (
			` + "`shopid`" + ` String,
			` + "`docno`" + ` String,
			` + "`docdatetime`" + ` DateTime,
			` + "`perioddatetime`" + ` DateTime,
			` + "`line_number`" + ` UInt32,
			` + "`barcode`" + ` String,
			` + "`unitcode`" + ` String,
			` + "`qty`" + ` Float64,
			` + "`price`" + ` Float64,
			` + "`discount`" + ` String,
			` + "`whcode`" + ` String,
			` + "`locationcode`" + ` String,
			` + "`sumofcost`" + ` Float64,
			` + "`discountamount`" + ` Float64,
			` + "`sumamount`" + ` Float64,
			` + "`branchid`" + ` String,
			` + "`itemnames`" + ` String,
			` + "`refguid`" + ` String,
			` + "`sumamountchoice`" + ` Float64,
			` + "`ischoice`" + ` Int32,
			` + "`guidfixed`" + ` String DEFAULT '',
			` + "`guidpos`" + ` String DEFAULT '',
			` + "`guidbranch`" + ` String DEFAULT '',
			` + "`transflag`" + ` Int16 DEFAULT 0,
			` + "`isupdated`" + ` Bool DEFAULT false,
			` + "`itemcode`" + ` Nullable(String) DEFAULT NULL,
			` + "`unitstand`" + ` Float64 DEFAULT 1.,
			` + "`unitdivide`" + ` Float64 DEFAULT 1.,
			` + "`itemname`" + ` String,
			` + "`calcflag`" + ` Int8,
			` + "`calcseq`" + ` Int32,
			` + "`barcodemain`" + ` String,
			` + "`iscalcstock`" + ` Int8 DEFAULT 0,
			` + "`price_doc`" + ` Float64 DEFAULT 0,
			` + "`sumamount_doc`" + ` Float64 DEFAULT 0,
			` + "`discountamount_doc`" + ` Float64 DEFAULT 0,
			` + "`priceexcludevat_doc`" + ` Float64 DEFAULT 0,
			` + "`sumamountexcludevat_doc`" + ` Float64 DEFAULT 0,
			` + "`totalvaluevat_doc`" + ` Float64 DEFAULT 0,
			INDEX idx_barcode barcode TYPE bloom_filter(0.01) GRANULARITY 8192,
			INDEX idx_shopid shopid TYPE minmax GRANULARITY 4,
			INDEX idx_branchid branchid TYPE minmax GRANULARITY 4,
			INDEX idx_perioddatetime perioddatetime TYPE minmax GRANULARITY 4,
			INDEX idx_shopid_transflag (shopid, transflag) TYPE minmax GRANULARITY 4,
			INDEX idx_shopid_isupdated (shopid, isupdated) TYPE set(100) GRANULARITY 1,
			INDEX idx_shop_item (shopid, itemcode) TYPE minmax GRANULARITY 1
		) ENGINE = MergeTree PARTITION BY shopid ORDER BY (shopid, docno, line_number) SETTINGS index_granularity = 8192`,

		// docdetail_updated
		`CREATE TABLE IF NOT EXISTS {DB}.docdetail_updated (
			` + "`shopid`" + ` String,
			` + "`docno`" + ` String,
			` + "`docdatetime`" + ` DateTime,
			` + "`perioddatetime`" + ` DateTime,
			` + "`line_number`" + ` UInt32,
			` + "`barcode`" + ` String,
			` + "`unitcode`" + ` String,
			` + "`qty`" + ` Float64,
			` + "`price`" + ` Float64,
			` + "`discount`" + ` String,
			` + "`whcode`" + ` String,
			` + "`locationcode`" + ` String,
			` + "`sumofcost`" + ` Float64,
			` + "`discountamount`" + ` Float64,
			` + "`sumamount`" + ` Float64,
			` + "`branchid`" + ` String,
			` + "`itemnames`" + ` String,
			` + "`refguid`" + ` String,
			` + "`sumamountchoice`" + ` Float64,
			` + "`ischoice`" + ` Int32,
			` + "`guidfixed`" + ` String DEFAULT '',
			` + "`guidpos`" + ` String DEFAULT '',
			` + "`guidbranch`" + ` String DEFAULT '',
			` + "`transflag`" + ` Int16 DEFAULT 0,
			` + "`isupdated`" + ` Bool DEFAULT false,
			` + "`itemcode`" + ` Nullable(String) DEFAULT NULL,
			` + "`unitstand`" + ` Float64 DEFAULT 1.,
			` + "`unitdivide`" + ` Float64 DEFAULT 1.,
			INDEX idx_barcode barcode TYPE bloom_filter(0.01) GRANULARITY 8192,
			INDEX idx_shopid shopid TYPE minmax GRANULARITY 4,
			INDEX idx_branchid branchid TYPE minmax GRANULARITY 4,
			INDEX idx_perioddatetime perioddatetime TYPE minmax GRANULARITY 4,
			INDEX idx_shopid_transflag (shopid, transflag) TYPE minmax GRANULARITY 4,
			INDEX idx_shopid_isupdated (shopid, isupdated) TYPE set(100) GRANULARITY 1,
			INDEX idx_shop_item (shopid, itemcode) TYPE minmax GRANULARITY 1
		) ENGINE = ReplacingMergeTree(line_number) PARTITION BY shopid ORDER BY (guidfixed, shopid, docno, line_number) SETTINGS index_granularity = 8192`,

		// docpayment
		`CREATE TABLE IF NOT EXISTS {DB}.docpayment (
			` + "`shopid`" + ` String,
			` + "`branchid`" + ` String,
			` + "`docdatetime`" + ` DateTime,
			` + "`perioddatetime`" + ` DateTime,
			` + "`amount`" + ` Decimal(18, 6),
			` + "`description`" + ` String,
			` + "`docno`" + ` String,
			` + "`trans_flag`" + ` Int32,
			` + "`guidfixed`" + ` String DEFAULT '',
			` + "`guidbranch`" + ` String DEFAULT ''
		) ENGINE = ReplacingMergeTree PARTITION BY shopid ORDER BY (guidfixed, shopid, docno) SETTINGS index_granularity = 8192`,

		// docref
		`CREATE TABLE IF NOT EXISTS {DB}.docref (
			` + "`shopid`" + ` String,
			` + "`docno`" + ` String,
			` + "`docnotransflag`" + ` Int32,
			` + "`docnoref`" + ` String,
			` + "`docnoreftransflag`" + ` Int32
		) ENGINE = MergeTree ORDER BY (shopid, docno, docnotransflag) SETTINGS index_granularity = 8192`,

		// ic_inventory
		`CREATE TABLE IF NOT EXISTS {DB}.ic_inventory (
			` + "`code`" + ` String,
			` + "`name_1`" + ` String,
			` + "`unit_cost`" + ` String,
			` + "`unit_standard`" + ` String,
			` + "`item_type`" + ` Int32,
			` + "`shopid`" + ` String
		) ENGINE = MergeTree ORDER BY code SETTINGS index_granularity = 8192`,

		// locations
		`CREATE TABLE IF NOT EXISTS {DB}.locations (
			` + "`shopid`" + ` String,
			` + "`guidfixed`" + ` String,
			` + "`locationcode`" + ` String,
			` + "`warehousecode`" + ` String,
			` + "`name1`" + ` String,
			` + "`name2`" + ` String
		) ENGINE = MergeTree PARTITION BY shopid ORDER BY shopid SETTINGS index_granularity = 8192`,

		// processstock
		`CREATE TABLE IF NOT EXISTS {DB}.processstock (
			` + "`shopid`" + ` String,
			` + "`docno`" + ` String,
			` + "`docdatetime`" + ` DateTime,
			` + "`perioddatetime`" + ` DateTime,
			` + "`taxdocno`" + ` String DEFAULT '',
			` + "`totalamount`" + ` Float64,
			` + "`roundamount`" + ` Float64 DEFAULT 0.,
			` + "`paytype`" + ` Int16 DEFAULT 0,
			` + "`paycashamount`" + ` Float64 DEFAULT 0.,
			` + "`paycashchange`" + ` Float64 DEFAULT 0.,
			` + "`paycashbalance`" + ` Float64 DEFAULT 0.,
			` + "`deliverycode`" + ` String DEFAULT '',
			` + "`checksum`" + ` String,
			` + "`branchid`" + ` String DEFAULT '',
			` + "`slipurl`" + ` String DEFAULT '',
			` + "`salechannelcode`" + ` String,
			` + "`deliveryamount`" + ` Float64,
			` + "`iscancel`" + ` Bool DEFAULT false,
			` + "`cancelreason`" + ` String DEFAULT '',
			` + "`guidpos`" + ` String DEFAULT '',
			` + "`guidbranch`" + ` String DEFAULT '',
			` + "`guidfixed`" + ` String DEFAULT '',
			` + "`transflag`" + ` UInt16,
			INDEX idx_shopid shopid TYPE minmax GRANULARITY 4,
			INDEX idx_branchid branchid TYPE minmax GRANULARITY 4,
			INDEX idx_perioddatetime perioddatetime TYPE minmax GRANULARITY 4
		) ENGINE = ReplacingMergeTree PARTITION BY shopid ORDER BY (guidfixed, shopid, docno) SETTINGS index_granularity = 8192`,

		// processstockcost
		`CREATE TABLE IF NOT EXISTS {DB}.processstockcost (
			` + "`shopid`" + ` LowCardinality(String) CODEC(ZSTD(1)),
			` + "`itemcode`" + ` LowCardinality(String) CODEC(ZSTD(1)),
			` + "`docdatetime`" + ` DateTime CODEC(Delta(4), ZSTD(1)),
			` + "`docno`" + ` String CODEC(ZSTD(1)),
			` + "`linenumber`" + ` UInt16 CODEC(ZSTD(1)),
			` + "`transflag`" + ` UInt8 CODEC(ZSTD(1)),
			` + "`barcodemain`" + ` String CODEC(ZSTD(1)),
			` + "`barcode`" + ` String CODEC(ZSTD(1)),
			` + "`unitcode`" + ` LowCardinality(String) CODEC(ZSTD(1)),
			` + "`whcode`" + ` LowCardinality(String) CODEC(ZSTD(1)),
			` + "`locationcode`" + ` LowCardinality(String) CODEC(ZSTD(1)),
			` + "`totalqty`" + ` Float64 CODEC(ZSTD(1)),
			` + "`unitstand`" + ` Float64 CODEC(ZSTD(1)),
			` + "`unitdivide`" + ` Float64 CODEC(ZSTD(1)),
			` + "`price`" + ` Float64 CODEC(ZSTD(1)),
			` + "`averagecost`" + ` Float64 CODEC(ZSTD(1)),
			` + "`calcamount`" + ` Float64 CODEC(ZSTD(1)),
			` + "`balanceqty`" + ` Float64 CODEC(ZSTD(1)),
			` + "`balanceamount`" + ` Float64 CODEC(ZSTD(1)),
			` + "`guid`" + ` String CODEC(ZSTD(1)),
			` + "`unitcost`" + ` Float64 CODEC(ZSTD(1)),
			` + "`docref`" + ` String CODEC(ZSTD(1)),
			` + "`originalqty`" + ` Int32,
			` + "`qty`" + ` Int32,
			INDEX idx_docno docno TYPE bloom_filter(0.001) GRANULARITY 4,
			INDEX idx_barcode barcode TYPE bloom_filter(0.001) GRANULARITY 4,
			INDEX idx_docdatetime docdatetime TYPE minmax GRANULARITY 4
		) ENGINE = MergeTree PARTITION BY (shopid, itemcode) ORDER BY (shopid, itemcode, docdatetime, whcode, locationcode) SETTINGS index_granularity = 16384, parts_to_throw_insert = 1000, parts_to_delay_insert = 500`,

		// processstockdetail
		`CREATE TABLE IF NOT EXISTS {DB}.processstockdetail (
			` + "`shopid`" + ` LowCardinality(String) CODEC(ZSTD(1)),
			` + "`docdatetime`" + ` DateTime,
			` + "`docno`" + ` String CODEC(ZSTD(1)),
			` + "`linenumber`" + ` UInt16,
			` + "`transflag`" + ` UInt16,
			` + "`calcflag`" + ` UInt16,
			` + "`calcseq`" + ` UInt16,
			` + "`barcodemain`" + ` String CODEC(ZSTD(1)),
			` + "`barcode`" + ` String CODEC(ZSTD(1)),
			` + "`unitcode`" + ` LowCardinality(String) CODEC(ZSTD(1)),
			` + "`whcode`" + ` LowCardinality(String) CODEC(ZSTD(1)),
			` + "`locationcode`" + ` LowCardinality(String) CODEC(ZSTD(1)),
			` + "`totalqty`" + ` Float64,
			` + "`unitstand`" + ` Float64,
			` + "`unitdivide`" + ` Float64,
			` + "`price`" + ` Float64,
			` + "`priceexcludevat`" + ` Float64,
			` + "`docref`" + ` String,
			INDEX idx_barcodemain (shopid, barcodemain) TYPE minmax GRANULARITY 1
		) ENGINE = MergeTree PARTITION BY shopid ORDER BY (shopid, barcode) SETTINGS index_granularity = 8192, index_granularity_bytes = 20971520, min_bytes_for_wide_part = 10485760, write_final_mark = 1`,

		// processstocklot
		`CREATE TABLE IF NOT EXISTS {DB}.processstocklot (
			` + "`shopid`" + ` LowCardinality(String) CODEC(ZSTD(1)),
			` + "`docdatetime`" + ` DateTime,
			` + "`lotnumber`" + ` String CODEC(ZSTD(1)),
			` + "`docno`" + ` String CODEC(ZSTD(1)),
			` + "`transflag`" + ` UInt16,
			` + "`barcodemain`" + ` String CODEC(ZSTD(1)),
			` + "`unitcode`" + ` LowCardinality(String) CODEC(ZSTD(1)),
			` + "`whcode`" + ` LowCardinality(String) CODEC(ZSTD(1)),
			` + "`locationcode`" + ` LowCardinality(String) CODEC(ZSTD(1)),
			` + "`qty`" + ` Float64,
			` + "`unitstand`" + ` Float64,
			` + "`unitdivide`" + ` Float64,
			` + "`price`" + ` Float64,
			` + "`cost`" + ` Float64,
			` + "`balanceqty`" + ` Float64,
			` + "`balanceamount`" + ` Float64,
			` + "`guidref`" + ` String CODEC(ZSTD(1)),
			INDEX idx_barcodemain (shopid, barcodemain) TYPE minmax GRANULARITY 1,
			INDEX idx_barcodemain_guid (shopid, barcodemain, guidref) TYPE minmax GRANULARITY 1
		) ENGINE = MergeTree PARTITION BY shopid ORDER BY (shopid, barcodemain) SETTINGS index_granularity = 8192, index_granularity_bytes = 20971520, min_bytes_for_wide_part = 10485760, write_final_mark = 1`,

		// productbarcode
		`CREATE TABLE IF NOT EXISTS {DB}.productbarcode (
			` + "`shopid`" + ` String,
			` + "`itemcode`" + ` String,
			` + "`barcode`" + ` String,
			` + "`name0`" + ` String,
			` + "`name1`" + ` String,
			` + "`name2`" + ` String,
			` + "`name3`" + ` String,
			` + "`name4`" + ` String,
			` + "`name5`" + ` String,
			` + "`checksum`" + ` String,
			` + "`groupcode`" + ` String DEFAULT '',
			` + "`groupnames`" + ` String DEFAULT '',
			` + "`unitcode`" + ` String DEFAULT '',
			` + "`unitname`" + ` String DEFAULT '',
			` + "`price1`" + ` Float64 DEFAULT 0,
			` + "`unitstand`" + ` Float64 DEFAULT 1.,
			` + "`unitdivide`" + ` Float64 DEFAULT 1.,
			` + "`unitstandanddivideisupdated`" + ` Bool DEFAULT false,
			` + "`isupdated`" + ` Bool DEFAULT false,
			` + "`barcoderef`" + ` String DEFAULT '',
			` + "`imageuri`" + ` String,
			` + "`price_retail`" + ` Float64 DEFAULT 0,
			INDEX idx_barcode barcode TYPE bloom_filter(0.01) GRANULARITY 8192,
			INDEX idx_shopid shopid TYPE minmax GRANULARITY 4
		) ENGINE = ReplacingMergeTree PARTITION BY shopid ORDER BY (shopid, barcode) SETTINGS index_granularity = 8192, parts_to_throw_insert = 3000, parts_to_delay_insert = 1500`,

		// productbarcodeimport
		`CREATE TABLE IF NOT EXISTS {DB}.productbarcodeimport (
			` + "`guidfixed`" + ` String,
			` + "`shopid`" + ` String,
			` + "`taskid`" + ` String,
			` + "`rownumber`" + ` Float64,
			` + "`barcode`" + ` String,
			` + "`code`" + ` String,
			` + "`name`" + ` String,
			` + "`unitcode`" + ` String,
			` + "`price`" + ` Float64,
			` + "`pricemember`" + ` Float64,
			` + "`isduplicate`" + ` Bool,
			` + "`isexist`" + ` Bool,
			` + "`createdby`" + ` String,
			` + "`createdat`" + ` DateTime,
			` + "`isunitnotexist`" + ` Bool,
			` + "`pricedelivery`" + ` Float64 DEFAULT 0,
			` + "`priceone`" + ` Float64,
			` + "`pricetwo`" + ` Float64,
			` + "`pricethree`" + ` Float64,
			` + "`pricefour`" + ` Float64,
			` + "`pricefive`" + ` Float64,
			` + "`pricesix`" + ` Float64,
			` + "`priseseven`" + ` Float64,
			` + "`priceeight`" + ` Float64,
			` + "`pricenine`" + ` Float64,
			` + "`groupcode`" + ` String,
			` + "`groupsubonecode`" + ` String,
			` + "`groupsubtwocode`" + ` String,
			` + "`brandcode`" + ` String,
			` + "`designcode`" + ` String,
			` + "`modelcode`" + ` String,
			` + "`patterncode`" + ` String,
			` + "`gradecode`" + ` String,
			` + "`categorycode`" + ` String,
			` + "`classcode`" + ` String,
			` + "`barcoderef`" + ` String,
			` + "`standvalue`" + ` Float64 DEFAULT 1,
			` + "`dividevalue`" + ` Float64 DEFAULT 1,
			` + "`issumpoint`" + ` Bool DEFAULT false
		) ENGINE = MergeTree ORDER BY (taskid, rownumber) SETTINGS index_granularity = 8192`,

		// productbarcodeprocess
		`CREATE TABLE IF NOT EXISTS {DB}.productbarcodeprocess (
			` + "`shopid`" + ` LowCardinality(String) CODEC(ZSTD(1)),
			` + "`barcode`" + ` String CODEC(ZSTD(1)),
			` + "`name0`" + ` String CODEC(ZSTD(3)),
			` + "`unitname`" + ` LowCardinality(String) CODEC(ZSTD(1)),
			` + "`unitcode`" + ` LowCardinality(String) CODEC(ZSTD(1)),
			` + "`itemtype`" + ` UInt8,
			` + "`barcoderef`" + ` String CODEC(ZSTD(1)),
			` + "`barcoderefunitstand`" + ` Float64 CODEC(Delta(8), ZSTD(1)),
			` + "`barcoderefunitdivide`" + ` Float64 CODEC(Delta(8), ZSTD(1)),
			` + "`isstock`" + ` UInt8,
			` + "`itemcode`" + ` String CODEC(ZSTD(1)),
			INDEX idx_itemcode (shopid, itemcode) TYPE minmax GRANULARITY 1,
			INDEX idx_barcode_ref (shopid, barcoderef) TYPE minmax GRANULARITY 1
		) ENGINE = MergeTree ORDER BY (shopid, barcode) SETTINGS index_granularity = 8192, index_granularity_bytes = 20971520, min_bytes_for_wide_part = 10485760, write_final_mark = 1`,

		// productbarcoderef
		`CREATE TABLE IF NOT EXISTS {DB}.productbarcoderef (
			` + "`id`" + ` UInt64,
			` + "`shopid`" + ` String,
			` + "`barcode`" + ` String,
			` + "`barcoderef`" + ` String,
			` + "`itemcode`" + ` String,
			` + "`unitcode`" + ` String DEFAULT '',
			` + "`standvalue`" + ` Float64 DEFAULT 1,
			` + "`dividevalue`" + ` Float64 DEFAULT 1,
			INDEX idx_barcode barcode TYPE bloom_filter GRANULARITY 1,
			INDEX idx_barcoderef barcoderef TYPE bloom_filter GRANULARITY 1,
			INDEX idx_itemcode itemcode TYPE bloom_filter GRANULARITY 1
		) ENGINE = MergeTree PARTITION BY shopid ORDER BY (shopid, barcode, id) SETTINGS index_granularity = 8192`,

		// result
		`CREATE TABLE IF NOT EXISTS {DB}.result (
			` + "`shopid`" + ` LowCardinality(String) CODEC(ZSTD(1)),
			` + "`guid`" + ` String CODEC(ZSTD(1)),
			` + "`docdatetime`" + ` DateTime,
			` + "`linenumber`" + ` UInt32,
			` + "`datajson`" + ` String CODEC(ZSTD(1))
		) ENGINE = MergeTree PARTITION BY shopid ORDER BY (shopid, guid, docdatetime, linenumber) SETTINGS index_granularity = 8192, index_granularity_bytes = 20971520, min_bytes_for_wide_part = 10485760, write_final_mark = 1`,

		// resultfordashboard
		`CREATE TABLE IF NOT EXISTS {DB}.resultfordashboard (
			` + "`shopid`" + ` LowCardinality(String) CODEC(ZSTD(1)),
			` + "`branchid`" + ` LowCardinality(String) CODEC(ZSTD(1)),
			` + "`guid`" + ` String CODEC(ZSTD(1)),
			` + "`docdatetime`" + ` DateTime,
			` + "`linenumber`" + ` UInt32,
			` + "`datajson`" + ` String CODEC(ZSTD(1))
		) ENGINE = MergeTree PARTITION BY shopid ORDER BY (shopid, branchid, guid, docdatetime, linenumber) SETTINGS index_granularity = 8192, index_granularity_bytes = 20971520, min_bytes_for_wide_part = 10485760, write_final_mark = 1`,

		// salechannel
		`CREATE TABLE IF NOT EXISTS {DB}.salechannel (
			` + "`shopid`" + ` String,
			` + "`guidfixed`" + ` String,
			` + "`code`" + ` String,
			` + "`name`" + ` String
		) ENGINE = MergeTree PARTITION BY shopid ORDER BY shopid SETTINGS index_granularity = 8192`,

		// shop
		`CREATE TABLE IF NOT EXISTS {DB}.shop (
			` + "`shopid`" + ` String,
			` + "`name`" + ` String,
			` + "`checksum`" + ` String,
			` + "`mongodbname`" + ` String,
			INDEX idx_shopid shopid TYPE minmax GRANULARITY 4,
			INDEX idx_mongodbname mongodbname TYPE minmax GRANULARITY 4
		) ENGINE = ReplacingMergeTree PARTITION BY mongodbname ORDER BY (mongodbname, shopid) SETTINGS index_granularity = 8192`,

		// stockwaitprocess
		`CREATE TABLE IF NOT EXISTS {DB}.stockwaitprocess (
			` + "`shopid`" + ` LowCardinality(String) CODEC(ZSTD(1)),
			` + "`docdatetime`" + ` DateTime,
			` + "`barcodemain`" + ` String CODEC(ZSTD(1))
		) ENGINE = MergeTree PARTITION BY shopid ORDER BY (shopid, docdatetime) SETTINGS index_granularity = 8192, index_granularity_bytes = 20971520, min_bytes_for_wide_part = 10485760, write_final_mark = 1`,

		// task_status
		`CREATE TABLE IF NOT EXISTS {DB}.task_status (
			` + "`task_id`" + ` String,
			` + "`shop_id`" + ` String,
			` + "`status`" + ` String,
			` + "`error_message`" + ` String,
			` + "`progress`" + ` Int32,
			` + "`created_at`" + ` DateTime('UTC'),
			` + "`updated_at`" + ` DateTime('UTC'),
			` + "`completed_at`" + ` DateTime('UTC')
		) ENGINE = MergeTree ORDER BY (shop_id, task_id, updated_at) SETTINGS index_granularity = 8192`,

		// temp_barcode_map
		`CREATE TABLE IF NOT EXISTS {DB}.temp_barcode_map (
			` + "`barcode`" + ` String,
			` + "`itemcode`" + ` String,
			` + "`unitstand`" + ` Float64,
			` + "`unitdivide`" + ` Float64
		) ENGINE = Memory`,

		// token
		`CREATE TABLE IF NOT EXISTS {DB}.token (
			` + "`tokenid`" + ` String,
			` + "`shopid`" + ` String,
			` + "`userid`" + ` String,
			` + "`active`" + ` UInt8 DEFAULT 1
		) ENGINE = MergeTree ORDER BY (tokenid, shopid) SETTINGS index_granularity = 8192`,

		// userlogin
		`CREATE TABLE IF NOT EXISTS {DB}.userlogin (
			` + "`email`" + ` String,
			` + "`jsonshoplist`" + ` String
		) ENGINE = MergeTree ORDER BY email SETTINGS index_granularity = 8192`,

		// warehouses
		`CREATE TABLE IF NOT EXISTS {DB}.warehouses (
			` + "`shopid`" + ` String,
			` + "`guidfixed`" + ` String,
			` + "`code`" + ` String,
			` + "`name1`" + ` String,
			` + "`name2`" + ` String
		) ENGINE = MergeTree PARTITION BY shopid ORDER BY shopid SETTINGS index_granularity = 8192`,
	}
}

// getDictionaryDDL คืน DDL สำหรับสร้าง dictionary
// แยกออกมาเพราะ dictionary ต้องสร้างหลัง table productbarcode
func getDictionaryDDL() string {
	return `CREATE DICTIONARY IF NOT EXISTS {DB}.productbarcode_dict (
		` + "`barcode`" + ` String,
		` + "`itemcode`" + ` String,
		` + "`unitstand`" + ` Float64,
		` + "`unitdivide`" + ` Float64
	) PRIMARY KEY barcode
	SOURCE(CLICKHOUSE(QUERY 'SELECT DISTINCT barcode, any(itemcode) as itemcode, any(unitstand) as unitstand, any(unitdivide) as unitdivide FROM {DB}.productbarcode GROUP BY barcode'))
	LIFETIME(MIN 0 MAX 0)
	LAYOUT(COMPLEX_KEY_HASHED())`
}

// ensureTables ตรวจสอบและสร้าง tables ใน ClickHouse ถ้ายังไม่มี
// เรียกใช้หลังจาก ensureDatabase() สร้าง database เสร็จแล้ว
func ensureTables(conn clickhouse.Conn, dbName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// ตรวจสอบจำนวน table ที่มีอยู่ใน database
	countQuery := fmt.Sprintf("SELECT count() as cnt FROM system.tables WHERE database = '%s'", dbName)
	rows, err := conn.Query(ctx, countQuery)
	if err != nil {
		return fmt.Errorf("ตรวจสอบ tables ล้มเหลว: %w", err)
	}
	defer rows.Close()

	var tableCount uint64
	if rows.Next() {
		if err := rows.Scan(&tableCount); err != nil {
			return fmt.Errorf("อ่านจำนวน tables ล้มเหลว: %w", err)
		}
	}

	// ถ้ามี table อยู่แล้วเพียงพอ ไม่ต้องสร้างใหม่
	// ใช้ CREATE TABLE IF NOT EXISTS ดังนั้น table ที่มีอยู่แล้วจะไม่ถูกเขียนทับ
	if tableCount >= uint64(len(coreTableNames)) {
		logger.Info("[ClickHouse] database '%s' มี %d tables แล้ว — ข้ามการสร้าง", dbName, tableCount)
		return nil
	}

	logger.Info("[ClickHouse] database '%s' มี %d tables — กำลังสร้าง tables ที่ขาด...", dbName, tableCount)

	// สร้าง tables ทั้งหมด (ใช้ IF NOT EXISTS ปลอดภัย)
	ddls := getCreateTableDDLs()
	successCount := 0
	for _, ddl := range ddls {
		sql := strings.ReplaceAll(ddl, "{DB}", dbName)
		if err := conn.Exec(ctx, sql); err != nil {
			logger.Error("[ClickHouse] สร้าง table ล้มเหลว: %v", err)
			// ไม่ return error — พยายามสร้าง tables อื่นต่อไป
			continue
		}
		successCount++
	}

	// สร้าง dictionary (ต้องสร้างหลัง productbarcode table)
	dictDDL := strings.ReplaceAll(getDictionaryDDL(), "{DB}", dbName)
	if err := conn.Exec(ctx, dictDDL); err != nil {
		logger.Error("[ClickHouse] สร้าง dictionary ล้มเหลว: %v — ไม่ร้ายแรง จะสร้างภายหลังได้", err)
	} else {
		successCount++
	}

	logger.Info("[ClickHouse] สร้าง tables สำเร็จ %d/%d รายการ ใน database '%s'", successCount, len(ddls)+1, dbName)
	return nil
}
