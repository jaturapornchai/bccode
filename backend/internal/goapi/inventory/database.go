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
		{"product_costing_config", ddlProductCostingConfig},
		{"inventory_cost_layers", ddlInventoryCostLayers},
		{"inventory_stock_balances", ddlInventoryStockBalances},
		{"marketplace_stock_balances", ddlMarketplaceStockBalances},
		{"marketplace_dimension_prices", ddlMarketplaceDimensionPrices},
		{"inventory_cost_transactions", ddlInventoryCostTransactions},
		{"inventory_variances", ddlInventoryVariances},
		{"inventory_landed_costs", ddlInventoryLandedCosts},
		{"inventory_landed_cost_allocations", ddlInventoryLandedCostAllocations},
		{"inventory_accounting_periods", ddlInventoryAccountingPeriods},
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
CREATE TABLE IF NOT EXISTS product_costing_config (
    id SERIAL PRIMARY KEY,
    holdingcode TEXT NOT NULL,
    itemcode TEXT NOT NULL,
    costingmethod VARCHAR(20) NOT NULL DEFAULT 'moving_average',
    lottrackingenabled BOOLEAN NOT NULL DEFAULT false,
    serialtrackingenabled BOOLEAN NOT NULL DEFAULT false,
    expirytrackingenabled BOOLEAN NOT NULL DEFAULT false,
    expiryalertdays INTEGER DEFAULT 30,
    autoblockexpired BOOLEAN DEFAULT false,
    autowriteoffexpired BOOLEAN DEFAULT false,
    minimumshelflifedays INTEGER DEFAULT 0,
    allownegativestock BOOLEAN NOT NULL DEFAULT false,
    cost_by_warehouse BOOLEAN NOT NULL DEFAULT false,
    standardcost NUMERIC(18,4) DEFAULT 0,
    standardcosteffectivedate DATE,
    createdat TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updatedat TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (holdingcode, itemcode)
)`

const ddlInventoryCostLayers = `
CREATE TABLE IF NOT EXISTS inventory_cost_layers (
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
    CONSTRAINT chk_remaining_qty CHECK (remainingqty >= 0)
)`

const ddlInventoryStockBalances = `
CREATE TABLE IF NOT EXISTS inventory_stock_balances (
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
CREATE TABLE IF NOT EXISTS marketplace_stock_balances (
    id BIGSERIAL PRIMARY KEY,
    holdingcode TEXT NOT NULL,
    item_code TEXT NOT NULL,
    dimension_key TEXT NOT NULL DEFAULT '',
    dimension_values JSONB NOT NULL DEFAULT '{}'::jsonb,
    current_qty NUMERIC(18,4) NOT NULL DEFAULT 0,
    reserved_qty NUMERIC(18,4) NOT NULL DEFAULT 0,
    available_qty NUMERIC(18,4) NOT NULL DEFAULT 0,
    source TEXT NOT NULL DEFAULT 'accounting',
    updatedat TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (holdingcode, item_code, dimension_key)
)`

const ddlMarketplaceDimensionPrices = `
CREATE TABLE IF NOT EXISTS marketplace_dimension_prices (
    id BIGSERIAL PRIMARY KEY,
    holdingcode TEXT NOT NULL,
    item_code TEXT NOT NULL,
    barcode TEXT NOT NULL DEFAULT '',
    dimension_key TEXT NOT NULL DEFAULT '',
    dimension_values JSONB NOT NULL DEFAULT '{}'::jsonb,
    price_level TEXT NOT NULL DEFAULT '',
    marketplace TEXT NOT NULL DEFAULT '',
    currency VARCHAR(3) NOT NULL DEFAULT 'THB',
    price NUMERIC(18,4) NOT NULL DEFAULT 0,
    sale_price NUMERIC(18,4) NOT NULL DEFAULT 0,
    compare_at_price NUMERIC(18,4) NOT NULL DEFAULT 0,
    effective_from TIMESTAMPTZ,
    effective_to TIMESTAMPTZ,
    isactive BOOLEAN NOT NULL DEFAULT true,
    updatedat TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (holdingcode, item_code, barcode, dimension_key, price_level, marketplace, currency)
)`

const ddlInventoryCostTransactions = `
CREATE TABLE IF NOT EXISTS inventory_cost_transactions (
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
CREATE TABLE IF NOT EXISTS inventory_variances (
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
CREATE TABLE IF NOT EXISTS inventory_landed_costs (
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
CREATE TABLE IF NOT EXISTS inventory_landed_cost_allocations (
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
CREATE TABLE IF NOT EXISTS inventory_accounting_periods (
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
	`CREATE INDEX IF NOT EXISTS idx_cl_fifo ON inventory_cost_layers (holdingcode, itemcode, whcode, locationcode, receiveddate ASC, id ASC) WHERE remainingqty > 0`,
	// Cost Layers — LIFO sort
	`CREATE INDEX IF NOT EXISTS idx_cl_lifo ON inventory_cost_layers (holdingcode, itemcode, whcode, locationcode, receiveddate DESC, id DESC) WHERE remainingqty > 0`,
	// Cost Layers — FEFO sort
	`CREATE INDEX IF NOT EXISTS idx_cl_fefo ON inventory_cost_layers (holdingcode, itemcode, whcode, locationcode, expirydate ASC, receiveddate ASC) WHERE remainingqty > 0`,
	// Cost Layers — Lot lookup
	`CREATE INDEX IF NOT EXISTS idx_cl_lot ON inventory_cost_layers (holdingcode, itemcode, whcode, lotnumber) WHERE remainingqty > 0`,
	// Cost Transactions — product + date
	`CREATE INDEX IF NOT EXISTS idx_ct_product ON inventory_cost_transactions (holdingcode, itemcode, whcode, transactiondate)`,
	// Cost Transactions — period
	`CREATE INDEX IF NOT EXISTS idx_ct_period ON inventory_cost_transactions (holdingcode, accountingperiod, itemcode)`,
	// Cost Transactions — doc reference
	`CREATE INDEX IF NOT EXISTS idx_ct_ref ON inventory_cost_transactions (holdingcode, refdoctype, refdocno)`,
	// Variances — product
	`CREATE INDEX IF NOT EXISTS idx_var_product ON inventory_variances (holdingcode, itemcode, transactiondate)`,
	// Stock Balances — shop
	`CREATE INDEX IF NOT EXISTS idx_sb_shop ON inventory_stock_balances (holdingcode, itemcode)`,
	`CREATE INDEX IF NOT EXISTS idx_marketplace_stock_dimension ON marketplace_stock_balances (holdingcode, item_code, dimension_key)`,
	`CREATE INDEX IF NOT EXISTS idx_marketplace_price_dimension ON marketplace_dimension_prices (holdingcode, item_code, dimension_key, marketplace, price_level)`,
}
