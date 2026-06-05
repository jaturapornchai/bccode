package inventory

import (
	"database/sql"
	"fmt"

	"smlcloudplatform/internal/goapi/logger"
)

// CreateInventoryCostingTables — สร้างตารางทั้งหมดสำหรับระบบ Inventory Costing
// เรียกตอน migration หรือสร้าง database ใหม่
func CreateInventoryCostingTables(db *sql.DB) error {
	tables := []struct {
		name string
		ddl  string
	}{
		{"productcostingconfig", ddlProductCostingConfig},
		{"inventorycostlayers", ddlInventoryCostLayers},
		{"inventorystockbalances", ddlInventoryStockBalances},
		{"marketplacestockbalances", ddlMarketplaceStockBalances},
		{"marketplacedimensionprices", ddlMarketplaceDimensionPrices},
		{"inventorycosttransactions", ddlInventoryCostTransactions},
		{"inventoryvariances", ddlInventoryVariances},
		{"inventorylandedcosts", ddlInventoryLandedCosts},
		{"inventorylandedcostallocations", ddlInventoryLandedCostAllocations},
		{"inventoryaccountingperiods", ddlInventoryAccountingPeriods},
	}

	for _, t := range tables {
		_, err := db.Exec(t.ddl)
		if err != nil {
			return fmt.Errorf("สร้างตาราง %s ไม่สำเร็จ: %w", t.name, err)
		}
		logger.Info(fmt.Sprintf("สร้างตาราง %s สำเร็จ", t.name))
	}

	// สร้าง indexes
	for _, idx := range inventoryIndexes {
		_, err := db.Exec(idx)
		if err != nil {
			// index อาจมีอยู่แล้ว → ไม่ต้อง error
			logger.Warn(fmt.Sprintf("สร้าง index: %v", err))
		}
	}

	return nil
}

// === DDL Statements ===

const ddlProductCostingConfig = `
CREATE TABLE IF NOT EXISTS productcostingconfig (
    id SERIAL PRIMARY KEY,
    holdingcode TEXT NOT NULL,
    itemcode TEXT NOT NULL,
    costingmethod VARCHAR(20) NOT NULL DEFAULT 'movingaverage',
    lottrackingenabled BOOLEAN NOT NULL DEFAULT false,
    serialtrackingenabled BOOLEAN NOT NULL DEFAULT false,
    expirytrackingenabled BOOLEAN NOT NULL DEFAULT false,
    expiryalertdays INTEGER DEFAULT 30,
    autoblockexpired BOOLEAN DEFAULT false,
    autowriteoffexpired BOOLEAN DEFAULT false,
    minimumshelflifedays INTEGER DEFAULT 0,
    allownegativestock BOOLEAN NOT NULL DEFAULT false,
    costbywarehouse BOOLEAN NOT NULL DEFAULT false,
    standardcost NUMERIC(18,4) DEFAULT 0,
    standardcosteffectivedate DATE,
    createdat TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updatedat TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (holdingcode, itemcode)
)`

const ddlInventoryCostLayers = `
CREATE TABLE IF NOT EXISTS inventorycostlayers (
    id BIGSERIAL PRIMARY KEY,
    holdingcode TEXT NOT NULL,
    itemcode TEXT NOT NULL,
    barcode TEXT,
    whcode TEXT NOT NULL,
    locationcode TEXT NOT NULL DEFAULT '',
    layertype VARCHAR(20) NOT NULL,
    refdoctype VARCHAR(50),
    refdocno VARCHAR(50),
    originalqty NUMERIC(18,4) NOT NULL,
    remainingqty NUMERIC(18,4) NOT NULL,
    unitcost NUMERIC(18,4) NOT NULL,
    landedcostperunit NUMERIC(18,4) NOT NULL DEFAULT 0,
    totalunitcost NUMERIC(18,4) NOT NULL DEFAULT 0,
    lotnumber VARCHAR(100),
    supplierlotnumber VARCHAR(100),
    manufacturingdate DATE,
    expirydate DATE,
    qualitystatus VARCHAR(20) DEFAULT 'approved',
    receiveddate DATE NOT NULL,
    createdat TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updatedat TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chkremainingqty CHECK (remainingqty >= 0)
)`

const ddlInventoryStockBalances = `
CREATE TABLE IF NOT EXISTS inventorystockbalances (
    id BIGSERIAL PRIMARY KEY,
    holdingcode TEXT NOT NULL,
    itemcode TEXT NOT NULL,
    barcode TEXT,
    whcode TEXT NOT NULL,
    locationcode TEXT NOT NULL DEFAULT '',
    currentqty NUMERIC(18,4) NOT NULL DEFAULT 0,
    reservedqty NUMERIC(18,4) NOT NULL DEFAULT 0,
    currentavgcost NUMERIC(18,4) NOT NULL DEFAULT 0,
    currenttotalvalue NUMERIC(18,4) NOT NULL DEFAULT 0,
    lastpurchasecost NUMERIC(18,4) DEFAULT 0,
    lastpurchasedate DATE,
    updatedat TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (holdingcode, itemcode, whcode, locationcode)
)`

const ddlMarketplaceStockBalances = `
CREATE TABLE IF NOT EXISTS marketplacestockbalances (
    id BIGSERIAL PRIMARY KEY,
    holdingcode TEXT NOT NULL,
    itemcode TEXT NOT NULL,
    dimensionkey TEXT NOT NULL DEFAULT '',
    dimensionvalues JSONB NOT NULL DEFAULT '{}'::jsonb,
    currentqty NUMERIC(18,4) NOT NULL DEFAULT 0,
    reservedqty NUMERIC(18,4) NOT NULL DEFAULT 0,
    availableqty NUMERIC(18,4) NOT NULL DEFAULT 0,
    source TEXT NOT NULL DEFAULT 'accounting',
    updatedat TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (holdingcode, itemcode, dimensionkey)
)`

const ddlMarketplaceDimensionPrices = `
CREATE TABLE IF NOT EXISTS marketplacedimensionprices (
    id BIGSERIAL PRIMARY KEY,
    holdingcode TEXT NOT NULL,
    itemcode TEXT NOT NULL,
    barcode TEXT NOT NULL DEFAULT '',
    dimensionkey TEXT NOT NULL DEFAULT '',
    dimensionvalues JSONB NOT NULL DEFAULT '{}'::jsonb,
    pricelevel TEXT NOT NULL DEFAULT '',
    marketplace TEXT NOT NULL DEFAULT '',
    currency VARCHAR(3) NOT NULL DEFAULT 'THB',
    price NUMERIC(18,4) NOT NULL DEFAULT 0,
    saleprice NUMERIC(18,4) NOT NULL DEFAULT 0,
    compareatprice NUMERIC(18,4) NOT NULL DEFAULT 0,
    effectivefrom TIMESTAMPTZ,
    effectiveto TIMESTAMPTZ,
    isactive BOOLEAN NOT NULL DEFAULT true,
    updatedat TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (holdingcode, itemcode, barcode, dimensionkey, pricelevel, marketplace, currency)
)`

const ddlInventoryCostTransactions = `
CREATE TABLE IF NOT EXISTS inventorycosttransactions (
    id BIGSERIAL PRIMARY KEY,
    holdingcode TEXT NOT NULL,
    itemcode TEXT NOT NULL,
    barcode TEXT,
    whcode TEXT NOT NULL,
    locationcode TEXT NOT NULL DEFAULT '',
    transactiontype VARCHAR(30) NOT NULL,
    transflag INT,
    refdoctype VARCHAR(50),
    refdocno VARCHAR(50),
    qty NUMERIC(18,4) NOT NULL,
    unitcost NUMERIC(18,4) NOT NULL,
    totalcost NUMERIC(18,4) NOT NULL,
    landedcost NUMERIC(18,4) DEFAULT 0,
    lotnumber VARCHAR(100),
    expirydate DATE,
    costlayerid BIGINT,
    balanceqty NUMERIC(18,4) NOT NULL,
    balanceavgcost NUMERIC(18,4),
    balancetotalvalue NUMERIC(18,4),
    costingmethodused VARCHAR(20) NOT NULL,
    transactiondate DATE NOT NULL,
    accountingperiod VARCHAR(7),
    createdby VARCHAR(50),
    createdat TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    isreversed BOOLEAN DEFAULT false,
    reversedbyid BIGINT
)`

const ddlInventoryVariances = `
CREATE TABLE IF NOT EXISTS inventoryvariances (
    id BIGSERIAL PRIMARY KEY,
    holdingcode TEXT NOT NULL,
    itemcode TEXT NOT NULL,
    whcode TEXT NOT NULL,
    variancetype VARCHAR(30) NOT NULL,
    refdoctype VARCHAR(50),
    refdocno VARCHAR(50),
    standardcost NUMERIC(18,4) NOT NULL,
    actualcost NUMERIC(18,4) NOT NULL,
    qty NUMERIC(18,4) NOT NULL,
    varianceamount NUMERIC(18,4) NOT NULL,
    transactiondate DATE NOT NULL,
    accountingperiod VARCHAR(7),
    createdat TIMESTAMPTZ NOT NULL DEFAULT NOW()
)`

const ddlInventoryLandedCosts = `
CREATE TABLE IF NOT EXISTS inventorylandedcosts (
    id BIGSERIAL PRIMARY KEY,
    holdingcode TEXT NOT NULL,
    refdoctype VARCHAR(50) NOT NULL,
    refdocno VARCHAR(50) NOT NULL,
    costtype VARCHAR(50) NOT NULL,
    description VARCHAR(255),
    totalamount NUMERIC(18,4) NOT NULL,
    currency VARCHAR(3) DEFAULT 'THB',
    exchangerate NUMERIC(18,8) DEFAULT 1,
    amountthb NUMERIC(18,4),
    allocationmethod VARCHAR(20) NOT NULL,
    isallocated BOOLEAN DEFAULT false,
    allocatedat TIMESTAMPTZ,
    createdat TIMESTAMPTZ NOT NULL DEFAULT NOW()
)`

const ddlInventoryLandedCostAllocations = `
CREATE TABLE IF NOT EXISTS inventorylandedcostallocations (
    id BIGSERIAL PRIMARY KEY,
    holdingcode TEXT NOT NULL,
    landedcostid BIGINT NOT NULL,
    itemcode TEXT NOT NULL,
    whcode TEXT NOT NULL,
    costlayerid BIGINT,
    allocatedamount NUMERIC(18,4) NOT NULL,
    allocatedperunit NUMERIC(18,4) NOT NULL,
    qty NUMERIC(18,4) NOT NULL,
    createdat TIMESTAMPTZ NOT NULL DEFAULT NOW()
)`

const ddlInventoryAccountingPeriods = `
CREATE TABLE IF NOT EXISTS inventoryaccountingperiods (
    id BIGSERIAL PRIMARY KEY,
    holdingcode TEXT NOT NULL,
    periodcode VARCHAR(7) NOT NULL,
    startdate DATE NOT NULL,
    enddate DATE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'open',
    periodicavgcalculated BOOLEAN DEFAULT false,
    closedby VARCHAR(50),
    closedat TIMESTAMPTZ,
    createdat TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (holdingcode, periodcode)
)`

// inventoryIndexes — indexes สำหรับ performance
var inventoryIndexes = []string{
	// Cost Layers — FIFO sort
	`CREATE INDEX IF NOT EXISTS idxclfifo ON inventorycostlayers (holdingcode, itemcode, whcode, locationcode, receiveddate ASC, id ASC) WHERE remainingqty > 0`,
	// Cost Layers — LIFO sort
	`CREATE INDEX IF NOT EXISTS idxcllifo ON inventorycostlayers (holdingcode, itemcode, whcode, locationcode, receiveddate DESC, id DESC) WHERE remainingqty > 0`,
	// Cost Layers — FEFO sort
	`CREATE INDEX IF NOT EXISTS idxclfefo ON inventorycostlayers (holdingcode, itemcode, whcode, locationcode, expirydate ASC, receiveddate ASC) WHERE remainingqty > 0`,
	// Cost Layers — Lot lookup
	`CREATE INDEX IF NOT EXISTS idxcllot ON inventorycostlayers (holdingcode, itemcode, whcode, lotnumber) WHERE remainingqty > 0`,
	// Cost Transactions — product + date
	`CREATE INDEX IF NOT EXISTS idxctproduct ON inventorycosttransactions (holdingcode, itemcode, whcode, transactiondate)`,
	// Cost Transactions — period
	`CREATE INDEX IF NOT EXISTS idxctperiod ON inventorycosttransactions (holdingcode, accountingperiod, itemcode)`,
	// Cost Transactions — doc reference
	`CREATE INDEX IF NOT EXISTS idxctref ON inventorycosttransactions (holdingcode, refdoctype, refdocno)`,
	// Variances — product
	`CREATE INDEX IF NOT EXISTS idxvarproduct ON inventoryvariances (holdingcode, itemcode, transactiondate)`,
	// Stock Balances — shop
	`CREATE INDEX IF NOT EXISTS idxsbshop ON inventorystockbalances (holdingcode, itemcode)`,
	`CREATE INDEX IF NOT EXISTS idxmarketplacestockdimension ON marketplacestockbalances (holdingcode, itemcode, dimensionkey)`,
	`CREATE INDEX IF NOT EXISTS idxmarketplacepricedimension ON marketplacedimensionprices (holdingcode, itemcode, dimensionkey, marketplace, pricelevel)`,
}
