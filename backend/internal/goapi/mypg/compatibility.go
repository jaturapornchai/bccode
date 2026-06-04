package mypg

import (
	"context"
	"database/sql"

	"smlcloudplatform/internal/goapi/mydb"
)

// PgSqlFastConnectV2 - เชื่อมต่อ PostgreSQL โดยใช้ unified database manager
func PgSqlFastConnectV2(holdingCode string) (*sql.DB, error) {
	return mydb.GetGlobalConnectionFromPool(holdingCode)
}

// GetDatabaseManager - ดึง DatabaseManager สำหรับ shop
// ใช้สำหรับ advanced features เช่น QueryPostgreSQL, circuit breaker stats
func GetDatabaseManager(holdingCode string) (*mydb.DatabaseManager, error) {
	return mydb.GetGlobalManagerFromPool(holdingCode)
}

// QueryPostgreSQL - รัน query บน PostgreSQL พร้อม circuit breaker (แบบใหม่)
func QueryPostgreSQL(ctx context.Context, holdingCode, query string, args ...any) ([]map[string]any, error) {
	manager, err := GetDatabaseManager(holdingCode)
	if err != nil {
		return nil, err
	}

	return manager.QueryPostgreSQL(ctx, query, args...)
}

// ExecPostgreSQL - รัน exec บน PostgreSQL พร้อม circuit breaker (แบบใหม่)
func ExecPostgreSQL(ctx context.Context, holdingCode, query string, args ...any) (sql.Result, error) {
	manager, err := GetDatabaseManager(holdingCode)
	if err != nil {
		return nil, err
	}

	return manager.ExecPostgreSQL(ctx, query, args...)
}

// GetCircuitBreakerStats - ดึงสถิติ circuit breaker สำหรับ shop
func GetCircuitBreakerStats(holdingCode string) (map[string]interface{}, error) {
	manager, err := GetDatabaseManager(holdingCode)
	if err != nil {
		return nil, err
	}

	return manager.GetCircuitBreakerStats(), nil
}

// ConnectAdminDatabase - เชื่อมต่อกับ admin database (postgres) โดยตรง
func ConnectAdminDatabase() (*sql.DB, error) {
	return ConnectOptimized("postgres")
}
