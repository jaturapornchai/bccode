package aichat

import (
	"context"
	"fmt"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/mydb"
)

// GetProductData retrieves product information from legacy PostgreSQL projections.
// MongoDB remains the operational source of truth for product CRUD.
func GetProductData(ctx context.Context, holdingCode string) ([]StockData, error) {
	db, err := mydb.GetGlobalConnectionFromPool(holdingCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	query := `
		SELECT
			p.itemcode as productcode,
			p.name0 as productname,
			COALESCE(STRING_AGG(DISTINCT pb.barcode, ','), '') as barcodelist,
			COALESCE(p.unitname, '') as unitstructure,
			'0' as stockqty
		FROM product p
		LEFT JOIN productbarcode pb ON p.itemcode = pb.itemcode
		WHERE p.itemcode IS NOT NULL
		GROUP BY p.itemcode, p.name0, p.unitname
		ORDER BY p.itemcode
	`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query stock data: %w", err)
	}
	defer rows.Close()

	var stockData []StockData
	for rows.Next() {
		var item StockData
		if err := rows.Scan(
			&item.ProductCode,
			&item.ProductName,
			&item.BarcodeList,
			&item.UnitStruct,
			&item.StockQty,
		); err != nil {
			logger.Warn("Failed to scan row: %v", err)
			continue
		}
		stockData = append(stockData, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	logger.Info("Retrieved %d product items for shop %s", len(stockData), holdingCode)
	return stockData, nil
}
