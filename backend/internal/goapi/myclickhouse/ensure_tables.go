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
	"docdetailupdated",
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
	"icinventory",
	"result",
	"resultfordashboard",
	"taskstatus",
	"tempbarcodemap",
}

// getCreateTableDDLs คืน DDL statements สำหรับสร้าง tables ทั้งหมด
// ใช้ {DB} เป็น placeholder แทนชื่อ database
func getCreateTableDDLs() []string {
	return []string{
		// cartorder
		`CREATE TABLE IF NOT EXISTS {DB}.cartorder (
			` + "`usercode`" + ` String,
			` + "`holdingcode`" + ` String,
			` + "`cartnumber`" + ` String,
			` + "`totalamount`" + ` Float64,
			` + "`createdatetime`" + ` DateTime,
			` + "`custcode`" + ` String,
			` + "`remark`" + ` String
		) ENGINE = MergeTree PARTITION BY holdingcode ORDER BY (holdingcode, cartnumber) SETTINGS index_granularity = 8192`,

		// cartorderdetail
		`CREATE TABLE IF NOT EXISTS {DB}.cartorderdetail (
			` + "`usercode`" + ` String,
			` + "`holdingcode`" + ` String,
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
		) ENGINE = MergeTree PARTITION BY holdingcode ORDER BY (holdingcode, barcode) SETTINGS index_granularity = 8192`,

		// creditors
		`CREATE TABLE IF NOT EXISTS {DB}.creditors (
			` + "`holdingcode`" + ` String,
			` + "`guidfixed`" + ` String,
			` + "`code`" + ` String,
			` + "`name1`" + ` String,
			` + "`name2`" + ` String
		) ENGINE = MergeTree PARTITION BY holdingcode ORDER BY holdingcode SETTINGS index_granularity = 8192`,

		// debtors
		`CREATE TABLE IF NOT EXISTS {DB}.debtors (
			` + "`holdingcode`" + ` String,
			` + "`guidfixed`" + ` String,
			` + "`code`" + ` String,
			` + "`name1`" + ` String,
			` + "`name2`" + ` String
		) ENGINE = MergeTree PARTITION BY holdingcode ORDER BY holdingcode SETTINGS index_granularity = 8192`,

		// doc
		`CREATE TABLE IF NOT EXISTS {DB}.doc (
			` + "`holdingcode`" + ` String,
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
			` + "`currencysymbol`" + ` String DEFAULT '',
			` + "`doccurrency`" + ` String DEFAULT '',
			` + "`doccurrencysymbol`" + ` String DEFAULT '',
			` + "`exchangerate`" + ` Float64 DEFAULT 1,
			` + "`totalamountdoc`" + ` Float64 DEFAULT 0,
			` + "`approvalstatus`" + ` String DEFAULT '',
			` + "`custcode`" + ` String DEFAULT '',
			` + "`isclosedmanual`" + ` Bool DEFAULT false,
			` + "`closedmanualbycode`" + ` String DEFAULT '',
			` + "`closedmanualbyname`" + ` String DEFAULT '',
			` + "`closedmanualat`" + ` DateTime DEFAULT '1970-01-01 00:00:00',
			` + "`closedmanualreason`" + ` String DEFAULT '',
			INDEX idxholdingcode holdingcode TYPE minmax GRANULARITY 4,
			INDEX idxbranchid branchid TYPE minmax GRANULARITY 4,
			INDEX idxperioddatetime perioddatetime TYPE minmax GRANULARITY 4,
			INDEX idxholdingcodetransflag (holdingcode, transflag) TYPE minmax GRANULARITY 4
		) ENGINE = ReplacingMergeTree PARTITION BY holdingcode ORDER BY (guidfixed, holdingcode, docno) SETTINGS index_granularity = 8192`,

		// docdetail
		`CREATE TABLE IF NOT EXISTS {DB}.docdetail (
			` + "`holdingcode`" + ` String,
			` + "`docno`" + ` String,
			` + "`docdatetime`" + ` DateTime,
			` + "`perioddatetime`" + ` DateTime,
			` + "`linenumber`" + ` UInt32,
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
			` + "`pricedoc`" + ` Float64 DEFAULT 0,
			` + "`sumamountdoc`" + ` Float64 DEFAULT 0,
			` + "`discountamountdoc`" + ` Float64 DEFAULT 0,
			` + "`priceexcludevatdoc`" + ` Float64 DEFAULT 0,
			` + "`sumamountexcludevatdoc`" + ` Float64 DEFAULT 0,
			` + "`totalvaluevatdoc`" + ` Float64 DEFAULT 0,
			INDEX idxbarcode barcode TYPE bloom_filter(0.01) GRANULARITY 8192,
			INDEX idxholdingcode holdingcode TYPE minmax GRANULARITY 4,
			INDEX idxbranchid branchid TYPE minmax GRANULARITY 4,
			INDEX idxperioddatetime perioddatetime TYPE minmax GRANULARITY 4,
			INDEX idxholdingcodetransflag (holdingcode, transflag) TYPE minmax GRANULARITY 4,
			INDEX idxholdingcodeisupdated (holdingcode, isupdated) TYPE set(100) GRANULARITY 1,
			INDEX idxshopitem (holdingcode, itemcode) TYPE minmax GRANULARITY 1
		) ENGINE = MergeTree PARTITION BY holdingcode ORDER BY (holdingcode, docno, linenumber) SETTINGS index_granularity = 8192`,

		// docdetailupdated
		`CREATE TABLE IF NOT EXISTS {DB}.docdetailupdated (
			` + "`holdingcode`" + ` String,
			` + "`docno`" + ` String,
			` + "`docdatetime`" + ` DateTime,
			` + "`perioddatetime`" + ` DateTime,
			` + "`linenumber`" + ` UInt32,
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
			INDEX idxbarcode barcode TYPE bloom_filter(0.01) GRANULARITY 8192,
			INDEX idxholdingcode holdingcode TYPE minmax GRANULARITY 4,
			INDEX idxbranchid branchid TYPE minmax GRANULARITY 4,
			INDEX idxperioddatetime perioddatetime TYPE minmax GRANULARITY 4,
			INDEX idxholdingcodetransflag (holdingcode, transflag) TYPE minmax GRANULARITY 4,
			INDEX idxholdingcodeisupdated (holdingcode, isupdated) TYPE set(100) GRANULARITY 1,
			INDEX idxshopitem (holdingcode, itemcode) TYPE minmax GRANULARITY 1
		) ENGINE = ReplacingMergeTree(linenumber) PARTITION BY holdingcode ORDER BY (guidfixed, holdingcode, docno, linenumber) SETTINGS index_granularity = 8192`,

		// docpayment
		`CREATE TABLE IF NOT EXISTS {DB}.docpayment (
			` + "`holdingcode`" + ` String,
			` + "`branchid`" + ` String,
			` + "`docdatetime`" + ` DateTime,
			` + "`perioddatetime`" + ` DateTime,
			` + "`amount`" + ` Decimal(18, 6),
			` + "`description`" + ` String,
			` + "`docno`" + ` String,
			` + "`transflag`" + ` Int32,
			` + "`guidfixed`" + ` String DEFAULT '',
			` + "`guidbranch`" + ` String DEFAULT ''
		) ENGINE = ReplacingMergeTree PARTITION BY holdingcode ORDER BY (guidfixed, holdingcode, docno) SETTINGS index_granularity = 8192`,

		// docref
		`CREATE TABLE IF NOT EXISTS {DB}.docref (
			` + "`holdingcode`" + ` String,
			` + "`docno`" + ` String,
			` + "`docnotransflag`" + ` Int32,
			` + "`docnoref`" + ` String,
			` + "`docnoreftransflag`" + ` Int32
		) ENGINE = MergeTree ORDER BY (holdingcode, docno, docnotransflag) SETTINGS index_granularity = 8192`,

		// icinventory
		`CREATE TABLE IF NOT EXISTS {DB}.icinventory (
			` + "`code`" + ` String,
			` + "`name1`" + ` String,
			` + "`unitcost`" + ` String,
			` + "`unitstandard`" + ` String,
			` + "`itemtype`" + ` Int32,
			` + "`holdingcode`" + ` String
		) ENGINE = MergeTree ORDER BY code SETTINGS index_granularity = 8192`,

		// locations
		`CREATE TABLE IF NOT EXISTS {DB}.locations (
			` + "`holdingcode`" + ` String,
			` + "`guidfixed`" + ` String,
			` + "`locationcode`" + ` String,
			` + "`warehousecode`" + ` String,
			` + "`name1`" + ` String,
			` + "`name2`" + ` String
		) ENGINE = MergeTree PARTITION BY holdingcode ORDER BY holdingcode SETTINGS index_granularity = 8192`,

		// processstock
		`CREATE TABLE IF NOT EXISTS {DB}.processstock (
			` + "`holdingcode`" + ` String,
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
			INDEX idxholdingcode holdingcode TYPE minmax GRANULARITY 4,
			INDEX idxbranchid branchid TYPE minmax GRANULARITY 4,
			INDEX idxperioddatetime perioddatetime TYPE minmax GRANULARITY 4
		) ENGINE = ReplacingMergeTree PARTITION BY holdingcode ORDER BY (guidfixed, holdingcode, docno) SETTINGS index_granularity = 8192`,

		// processstockcost
		`CREATE TABLE IF NOT EXISTS {DB}.processstockcost (
			` + "`holdingcode`" + ` LowCardinality(String) CODEC(ZSTD(1)),
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
			INDEX idxdocno docno TYPE bloom_filter(0.001) GRANULARITY 4,
			INDEX idxbarcode barcode TYPE bloom_filter(0.001) GRANULARITY 4,
			INDEX idxdocdatetime docdatetime TYPE minmax GRANULARITY 4
		) ENGINE = MergeTree PARTITION BY (holdingcode, itemcode) ORDER BY (holdingcode, itemcode, docdatetime, whcode, locationcode) SETTINGS index_granularity = 16384, parts_to_throw_insert = 1000, parts_to_delay_insert = 500`,

		// processstockdetail
		`CREATE TABLE IF NOT EXISTS {DB}.processstockdetail (
			` + "`holdingcode`" + ` LowCardinality(String) CODEC(ZSTD(1)),
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
			INDEX idxbarcodemain (holdingcode, barcodemain) TYPE minmax GRANULARITY 1
		) ENGINE = MergeTree PARTITION BY holdingcode ORDER BY (holdingcode, barcode) SETTINGS index_granularity = 8192, index_granularity_bytes = 20971520, min_bytes_for_wide_part = 10485760, write_final_mark = 1`,

		// processstocklot
		`CREATE TABLE IF NOT EXISTS {DB}.processstocklot (
			` + "`holdingcode`" + ` LowCardinality(String) CODEC(ZSTD(1)),
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
			INDEX idxbarcodemain (holdingcode, barcodemain) TYPE minmax GRANULARITY 1,
			INDEX idxbarcodemainguid (holdingcode, barcodemain, guidref) TYPE minmax GRANULARITY 1
		) ENGINE = MergeTree PARTITION BY holdingcode ORDER BY (holdingcode, barcodemain) SETTINGS index_granularity = 8192, index_granularity_bytes = 20971520, min_bytes_for_wide_part = 10485760, write_final_mark = 1`,

		// productbarcode
		`CREATE TABLE IF NOT EXISTS {DB}.productbarcode (
			` + "`holdingcode`" + ` String,
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
			` + "`priceretail`" + ` Float64 DEFAULT 0,
			INDEX idxbarcode barcode TYPE bloom_filter(0.01) GRANULARITY 8192,
			INDEX idxholdingcode holdingcode TYPE minmax GRANULARITY 4
		) ENGINE = ReplacingMergeTree PARTITION BY holdingcode ORDER BY (holdingcode, barcode) SETTINGS index_granularity = 8192, parts_to_throw_insert = 3000, parts_to_delay_insert = 1500`,

		// productbarcodeimport
		`CREATE TABLE IF NOT EXISTS {DB}.productbarcodeimport (
			` + "`guidfixed`" + ` String,
			` + "`holdingcode`" + ` String,
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
			` + "`holdingcode`" + ` LowCardinality(String) CODEC(ZSTD(1)),
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
			INDEX idxitemcode (holdingcode, itemcode) TYPE minmax GRANULARITY 1,
			INDEX idxbarcoderef (holdingcode, barcoderef) TYPE minmax GRANULARITY 1
		) ENGINE = MergeTree ORDER BY (holdingcode, barcode) SETTINGS index_granularity = 8192, index_granularity_bytes = 20971520, min_bytes_for_wide_part = 10485760, write_final_mark = 1`,

		// productbarcoderef
		`CREATE TABLE IF NOT EXISTS {DB}.productbarcoderef (
			` + "`id`" + ` UInt64,
			` + "`holdingcode`" + ` String,
			` + "`barcode`" + ` String,
			` + "`barcoderef`" + ` String,
			` + "`itemcode`" + ` String,
			` + "`unitcode`" + ` String DEFAULT '',
			` + "`standvalue`" + ` Float64 DEFAULT 1,
			` + "`dividevalue`" + ` Float64 DEFAULT 1,
			INDEX idxbarcode barcode TYPE bloom_filter GRANULARITY 1,
			INDEX idxbarcoderef barcoderef TYPE bloom_filter GRANULARITY 1,
			INDEX idxitemcode itemcode TYPE bloom_filter GRANULARITY 1
		) ENGINE = MergeTree PARTITION BY holdingcode ORDER BY (holdingcode, barcode, id) SETTINGS index_granularity = 8192`,

		// result
		`CREATE TABLE IF NOT EXISTS {DB}.result (
			` + "`holdingcode`" + ` LowCardinality(String) CODEC(ZSTD(1)),
			` + "`guid`" + ` String CODEC(ZSTD(1)),
			` + "`docdatetime`" + ` DateTime,
			` + "`linenumber`" + ` UInt32,
			` + "`datajson`" + ` String CODEC(ZSTD(1))
		) ENGINE = MergeTree PARTITION BY holdingcode ORDER BY (holdingcode, guid, docdatetime, linenumber) SETTINGS index_granularity = 8192, index_granularity_bytes = 20971520, min_bytes_for_wide_part = 10485760, write_final_mark = 1`,

		// resultfordashboard
		`CREATE TABLE IF NOT EXISTS {DB}.resultfordashboard (
			` + "`holdingcode`" + ` LowCardinality(String) CODEC(ZSTD(1)),
			` + "`branchid`" + ` LowCardinality(String) CODEC(ZSTD(1)),
			` + "`guid`" + ` String CODEC(ZSTD(1)),
			` + "`docdatetime`" + ` DateTime,
			` + "`linenumber`" + ` UInt32,
			` + "`datajson`" + ` String CODEC(ZSTD(1))
		) ENGINE = MergeTree PARTITION BY holdingcode ORDER BY (holdingcode, branchid, guid, docdatetime, linenumber) SETTINGS index_granularity = 8192, index_granularity_bytes = 20971520, min_bytes_for_wide_part = 10485760, write_final_mark = 1`,

		// salechannel
		`CREATE TABLE IF NOT EXISTS {DB}.salechannel (
			` + "`holdingcode`" + ` String,
			` + "`guidfixed`" + ` String,
			` + "`code`" + ` String,
			` + "`name`" + ` String
		) ENGINE = MergeTree PARTITION BY holdingcode ORDER BY holdingcode SETTINGS index_granularity = 8192`,

		// shop
		`CREATE TABLE IF NOT EXISTS {DB}.shop (
			` + "`holdingcode`" + ` String,
			` + "`name`" + ` String,
			` + "`checksum`" + ` String,
			` + "`mongodbname`" + ` String,
			INDEX idxholdingcode holdingcode TYPE minmax GRANULARITY 4,
			INDEX idxmongodbname mongodbname TYPE minmax GRANULARITY 4
		) ENGINE = ReplacingMergeTree PARTITION BY mongodbname ORDER BY (mongodbname, holdingcode) SETTINGS index_granularity = 8192`,

		// stockwaitprocess
		`CREATE TABLE IF NOT EXISTS {DB}.stockwaitprocess (
			` + "`holdingcode`" + ` LowCardinality(String) CODEC(ZSTD(1)),
			` + "`docdatetime`" + ` DateTime,
			` + "`barcodemain`" + ` String CODEC(ZSTD(1))
		) ENGINE = MergeTree PARTITION BY holdingcode ORDER BY (holdingcode, docdatetime) SETTINGS index_granularity = 8192, index_granularity_bytes = 20971520, min_bytes_for_wide_part = 10485760, write_final_mark = 1`,

		// taskstatus
		`CREATE TABLE IF NOT EXISTS {DB}.taskstatus (
			` + "`taskid`" + ` String,
			` + "`holdingcode`" + ` String,
			` + "`status`" + ` String,
			` + "`errormessage`" + ` String,
			` + "`progress`" + ` Int32,
			` + "`createdat`" + ` DateTime('UTC'),
			` + "`updatedat`" + ` DateTime('UTC'),
			` + "`completedat`" + ` DateTime('UTC')
		) ENGINE = MergeTree ORDER BY (holdingcode, taskid, updatedat) SETTINGS index_granularity = 8192`,

		// tempbarcodemap
		`CREATE TABLE IF NOT EXISTS {DB}.tempbarcodemap (
			` + "`barcode`" + ` String,
			` + "`itemcode`" + ` String,
			` + "`unitstand`" + ` Float64,
			` + "`unitdivide`" + ` Float64
		) ENGINE = Memory`,

		// token
		`CREATE TABLE IF NOT EXISTS {DB}.token (
			` + "`tokenid`" + ` String,
			` + "`holdingcode`" + ` String,
			` + "`userid`" + ` String,
			` + "`active`" + ` UInt8 DEFAULT 1
		) ENGINE = MergeTree ORDER BY (tokenid, holdingcode) SETTINGS index_granularity = 8192`,

		// userlogin
		`CREATE TABLE IF NOT EXISTS {DB}.userlogin (
			` + "`email`" + ` String,
			` + "`jsonshoplist`" + ` String
		) ENGINE = MergeTree ORDER BY email SETTINGS index_granularity = 8192`,

		// warehouses
		`CREATE TABLE IF NOT EXISTS {DB}.warehouses (
			` + "`holdingcode`" + ` String,
			` + "`guidfixed`" + ` String,
			` + "`code`" + ` String,
			` + "`name1`" + ` String,
			` + "`name2`" + ` String
		) ENGINE = MergeTree PARTITION BY holdingcode ORDER BY holdingcode SETTINGS index_granularity = 8192`,
	}
}

// getDictionaryDDL คืน DDL สำหรับสร้าง dictionary
// แยกออกมาเพราะ dictionary ต้องสร้างหลัง table productbarcode
func getDictionaryDDL() string {
	return `CREATE DICTIONARY IF NOT EXISTS {DB}.productbarcodedict (
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
