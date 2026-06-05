package aichat

import (
	"context"
	"fmt"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/mydb"
)

// GetCustomerData retrieves customer/debtor information from legacy PostgreSQL projections.
// MongoDB remains the operational source of truth for debtor/creditor CRUD.
func GetCustomerData(ctx context.Context, holdingCode string) ([]StockData, error) {
	db, err := mydb.GetGlobalConnectionFromPool(holdingCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	query := `
		SELECT
			code as productcode,
			name0 as productname,
			COALESCE(phone, '') as barcodelist,
			COALESCE(address, '') as unitstructure,
			COALESCE(creditlimit::text, '0') as stockqty
		FROM customer
		WHERE code IS NOT NULL

		UNION ALL

		SELECT
			code as productcode,
			name0 as productname,
			COALESCE(phone, '') as barcodelist,
			COALESCE(address, '') as unitstructure,
			COALESCE(creditlimit::text, '0') as stockqty
		FROM debtor
		WHERE code IS NOT NULL

		ORDER BY productcode
	`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query customer data: %w", err)
	}
	defer rows.Close()

	var customerData []StockData
	for rows.Next() {
		var item StockData
		if err := rows.Scan(
			&item.ProductCode,
			&item.ProductName,
			&item.BarcodeList,
			&item.UnitStruct,
			&item.StockQty,
		); err != nil {
			logger.Warn("Failed to scan customer row: %v", err)
			continue
		}
		customerData = append(customerData, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating customer rows: %w", err)
	}

	logger.Info("Retrieved %d customer/debtor items for shop %s", len(customerData), holdingCode)
	return customerData, nil
}
