package processstock

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/myglobal"
	"strconv"
	"strings"
	"time"
)

// IncrementalConfig - Configuration for incremental calculation
type IncrementalConfig struct {
	Enabled    bool // Enable incremental calculation (skip unchanged items)
	MinimalLog bool // Use UPSERT instead of DELETE+INSERT to reduce WAL
}

// DefaultIncrementalConfig - Default configuration
var DefaultIncrementalConfig = IncrementalConfig{
	Enabled:    true,
	MinimalLog: true,
}

// CalculateItemChecksum - Calculate MD5 checksum for an item's docdetail records
// The checksum is based on all relevant fields that affect cost calculation
func CalculateItemChecksum(ctx context.Context, db *sql.DB, shopId, itemCode string) (string, error) {
	// Build query with transflags
	transFlagStrings := make([]string, len(myglobal.TransFlagsToProcess))
	for i, flag := range myglobal.TransFlagsToProcess {
		transFlagStrings[i] = strconv.Itoa(flag)
	}
	transFlagForQuery := strings.Join(transFlagStrings, ",")

	// Query to get all relevant data for checksum calculation
	// Order by docdatetime, linenumber to ensure consistent ordering
	query := fmt.Sprintf(`
		SELECT
			COALESCE(docdatetime::text, ''),
			COALESCE(docno, ''),
			COALESCE(linenumber::text, '0'),
			COALESCE(transflag::text, '0'),
			COALESCE(totalqty::text, '0'),
			COALESCE(priceexcludevat::text, '0'),
			COALESCE(sumamount::text, '0'),
			COALESCE(unitstand::text, '1'),
			COALESCE(unitdivide::text, '1')
		FROM docdetail
		WHERE itemcode = $1 AND transflag IN (%s)
		ORDER BY docdatetime, linenumber, docno
	`, transFlagForQuery)

	rows, err := db.QueryContext(ctx, query, itemCode)
	if err != nil {
		return "", fmt.Errorf("query docdetail for checksum: %w", err)
	}
	defer rows.Close()

	// Build a string from all values
	var builder strings.Builder
	rowCount := 0

	for rows.Next() {
		var docdatetime, docno, linenumber, transflag, totalqty, price, sumamount, unitstand, unitdivide string
		if err := rows.Scan(&docdatetime, &docno, &linenumber, &transflag, &totalqty, &price, &sumamount, &unitstand, &unitdivide); err != nil {
			return "", fmt.Errorf("scan docdetail row: %w", err)
		}

		// Concatenate all values with a delimiter
		builder.WriteString(docdatetime)
		builder.WriteString("|")
		builder.WriteString(docno)
		builder.WriteString("|")
		builder.WriteString(linenumber)
		builder.WriteString("|")
		builder.WriteString(transflag)
		builder.WriteString("|")
		builder.WriteString(totalqty)
		builder.WriteString("|")
		builder.WriteString(price)
		builder.WriteString("|")
		builder.WriteString(sumamount)
		builder.WriteString("|")
		builder.WriteString(unitstand)
		builder.WriteString("|")
		builder.WriteString(unitdivide)
		builder.WriteString("\n")
		rowCount++
	}

	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("iterate docdetail rows: %w", err)
	}

	// If no rows, return empty checksum (item has no transactions)
	if rowCount == 0 {
		return "", nil
	}

	// Calculate MD5 hash
	hash := md5.Sum([]byte(builder.String()))
	return hex.EncodeToString(hash[:]), nil
}

// CheckItemChanged - Check if an item's data has changed since last calculation
// Returns true if the item needs recalculation, false if it can be skipped
func CheckItemChanged(ctx context.Context, db *sql.DB, shopId, itemCode string) (bool, string, error) {
	// Calculate current checksum
	currentChecksum, err := CalculateItemChecksum(ctx, db, shopId, itemCode)
	if err != nil {
		// If we can't calculate checksum, assume changed for safety
		return true, "", fmt.Errorf("calculate checksum: %w", err)
	}

	// If no data (empty checksum), check if we have existing calculation
	if currentChecksum == "" {
		// Check if there's existing data in processstockcost
		var count int
		err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM processstockcost WHERE itemcode = $1 LIMIT 1", itemCode).Scan(&count)
		if err != nil {
			return true, currentChecksum, nil
		}
		// If we have existing data but no source data, need to recalculate (delete old)
		return count > 0, currentChecksum, nil
	}

	// Get last checksum from stock_calculation_state
	var lastChecksum sql.NullString
	err = db.QueryRowContext(ctx,
		"SELECT last_checksum FROM stock_calculation_state WHERE shop_id = $1 AND item_code = $2",
		shopId, itemCode,
	).Scan(&lastChecksum)

	if err != nil {
		if err == sql.ErrNoRows {
			// No previous calculation, need to calculate
			return true, currentChecksum, nil
		}
		// Other error, assume changed for safety
		return true, currentChecksum, fmt.Errorf("query stock_calculation_state: %w", err)
	}

	// Compare checksums
	if !lastChecksum.Valid || lastChecksum.String != currentChecksum {
		return true, currentChecksum, nil
	}

	// Checksums match, no change
	return false, currentChecksum, nil
}

// UpdateItemChecksum - Update the checksum after successful calculation
func UpdateItemChecksum(ctx context.Context, db *sql.DB, shopId, itemCode, checksum string) error {
	query := `
		INSERT INTO stock_calculation_state (shop_id, item_code, last_checksum, last_calc_time, version, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), 1, NOW(), NOW())
		ON CONFLICT (shop_id, item_code) DO UPDATE
		SET last_checksum = EXCLUDED.last_checksum,
			last_calc_time = EXCLUDED.last_calc_time,
			version = stock_calculation_state.version + 1,
			updated_at = NOW()
	`

	_, err := db.ExecContext(ctx, query, shopId, itemCode, checksum)
	if err != nil {
		return fmt.Errorf("upsert stock_calculation_state: %w", err)
	}

	return nil
}

// DeleteItemChecksum - Delete the checksum when item data is cleared
func DeleteItemChecksum(ctx context.Context, db *sql.DB, shopId, itemCode string) error {
	_, err := db.ExecContext(ctx,
		"DELETE FROM stock_calculation_state WHERE shop_id = $1 AND item_code = $2",
		shopId, itemCode,
	)
	if err != nil {
		return fmt.Errorf("delete stock_calculation_state: %w", err)
	}
	return nil
}

// IncrementalStats - Statistics for incremental calculation
type IncrementalStats struct {
	TotalItems int           `json:"total_items"`
	SkippedItems int           `json:"skipped_items"`
	ProcessedItems int         `json:"processed_items"`
	Duration time.Duration `json:"duration"`
	WALSavedPercent float64    `json:"wal_saved_percent"`
}

// LogIncrementalStats - Log statistics for incremental calculation
func LogIncrementalStats(shopId string, stats IncrementalStats) {
	if stats.TotalItems == 0 {
		return
	}

	skipPercent := float64(stats.SkippedItems) / float64(stats.TotalItems) * 100
	logger.Info("Incremental calculation stats | shop=%s total=%d skipped=%d (%.1f%%) processed=%d duration=%v",
		shopId,
		stats.TotalItems,
		stats.SkippedItems,
		skipPercent,
		stats.ProcessedItems,
		stats.Duration.Round(time.Millisecond),
	)
}

// EnsureStockCalculationStateTable - Create table if not exists (for auto-migration)
func EnsureStockCalculationStateTable(ctx context.Context, db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS stock_calculation_state (
			shop_id VARCHAR(100) NOT NULL,
			item_code VARCHAR(100) NOT NULL,
			last_checksum CHAR(32),
			last_calc_time TIMESTAMPTZ DEFAULT NOW(),
			version INTEGER DEFAULT 0,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW(),
			PRIMARY KEY (shop_id, item_code)
		)
	`

	_, err := db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("create stock_calculation_state table: %w", err)
	}

	// Create indexes
	indexQueries := []string{
		"CREATE INDEX IF NOT EXISTS idx_stock_calc_state_shop_item ON stock_calculation_state(shop_id, item_code)",
		"CREATE INDEX IF NOT EXISTS idx_stock_calc_state_last_calc_time ON stock_calculation_state(last_calc_time)",
	}

	for _, q := range indexQueries {
		if _, err := db.ExecContext(ctx, q); err != nil {
			logger.Warn("Failed to create index: %v", err)
		}
	}

	return nil
}
