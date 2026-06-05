package mypg

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"smlcloudplatform/internal/goapi/logger"
)

// QueryResult represents a row in query_results table
type QueryResult struct {
	ID          int64           `json:"id"`
	GUID        string          `json:"guid"`
	HoldingCode string          `json:"holdingcode"`
	DocDatetime time.Time       `json:"docdatetime"`
	LineNumber  int             `json:"linenumber"`
	DataJSON    json.RawMessage `json:"datajson"`
	CreatedAt   time.Time       `json:"createdat"`
}

// CreateResultTableIfNotExists สร้าง query_results table ถ้ายังไม่มี
func CreateResultTableIfNotExists(db *sql.DB) error {
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS public.query_results (
			id SERIAL PRIMARY KEY,
			guid TEXT NOT NULL,
			holdingcode TEXT NOT NULL,
			docdatetime TIMESTAMPTZ DEFAULT NOW(),
			linenumber INTEGER NOT NULL,
			datajson JSONB NOT NULL,
			createdat TIMESTAMPTZ DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_query_results_guid
		ON public.query_results(guid);

		CREATE INDEX IF NOT EXISTS idx_query_results_holdingcode_guid
		ON public.query_results(holdingcode, guid);
	`

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, createTableQuery)
	if err != nil {
		return fmt.Errorf("failed to create query_results table: %w", err)
	}

	logger.Info("query_results table created/verified successfully")
	return nil
}

// InsertQueryResult inserts a single query result row
func InsertQueryResult(db *sql.DB, guid, holdingCode string, lineNumber int, data map[string]interface{}) error {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	query := `
		INSERT INTO public.query_results (guid, holdingcode, linenumber, datajson)
		VALUES ($1, $2, $3, $4)
	`

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = db.ExecContext(ctx, query, guid, holdingCode, lineNumber, dataJSON)
	if err != nil {
		return fmt.Errorf("failed to insert query result: %w", err)
	}

	return nil
}

// InsertQueryResultsBatch inserts multiple query results in a batch
func InsertQueryResultsBatch(db *sql.DB, guid, holdingCode string, results []map[string]interface{}) (int, error) {
	if len(results) == 0 {
		return 0, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO public.query_results (guid, holdingcode, linenumber, datajson)
		VALUES ($1, $2, $3, $4)
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for i, data := range results {
		dataJSON, err := json.Marshal(data)
		if err != nil {
			return i, fmt.Errorf("failed to marshal data at row %d: %w", i, err)
		}

		_, err = stmt.ExecContext(ctx, guid, holdingCode, i+1, dataJSON)
		if err != nil {
			return i, fmt.Errorf("failed to insert row %d: %w", i, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return len(results), nil
}

// GetQueryResults retrieves query results by GUID with pagination
func GetQueryResults(db *sql.DB, holdingCode, guid string, limit, offset int) ([]map[string]interface{}, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM public.query_results WHERE holdingcode = $1 AND guid = $2`
	err := db.QueryRowContext(ctx, countQuery, holdingCode, guid).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get count: %w", err)
	}

	// Get data with pagination
	query := `
		SELECT datajson FROM public.query_results
		WHERE holdingcode = $1 AND guid = $2
		ORDER BY linenumber
		LIMIT $3 OFFSET $4
	`

	rows, err := db.QueryContext(ctx, query, holdingCode, guid, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query results: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var dataJSON []byte
		if err := rows.Scan(&dataJSON); err != nil {
			return nil, 0, fmt.Errorf("failed to scan row: %w", err)
		}

		var data map[string]interface{}
		if err := json.Unmarshal(dataJSON, &data); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal data: %w", err)
		}

		results = append(results, data)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("row iteration error: %w", err)
	}

	return results, total, nil
}

// DeleteQueryResults deletes query results by GUID
func DeleteQueryResults(db *sql.DB, holdingCode, guid string) error {
	query := `DELETE FROM public.query_results WHERE holdingcode = $1 AND guid = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, query, holdingCode, guid)
	if err != nil {
		return fmt.Errorf("failed to delete query results: %w", err)
	}

	return nil
}
