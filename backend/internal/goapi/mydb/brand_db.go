package mydb

import (
	"database/sql"
	"fmt"
	"smlcloudplatform/internal/goapi/config"
	"smlcloudplatform/internal/goapi/logger"
	"sync"

	_ "github.com/lib/pq"
)

var (
	brandDB     *sql.DB
	brandDBOnce sync.Once
	brandDBErr  error
)

// GetBrandDB returns a singleton connection to brand_mappings_db
func GetBrandDB() (*sql.DB, error) {
	brandDBOnce.Do(func() {
		cfg := config.NewServiceConfig()

		// Build connection string for brand_mappings_db
		connStr := fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
			cfg.PostgresHost(),
			cfg.PostgresPort(),
			cfg.PostgresUser(),
			cfg.PostgresPassword(),
			cfg.PostgresDatabase(), // This should be "brand_mappings_db" from .env
			cfg.PostgresSSLMode(),
		)

		db, err := sql.Open("postgres", connStr)
		if err != nil {
			brandDBErr = fmt.Errorf("failed to open brand database: %w", err)
			logger.Error("Failed to open brand database: %v", err)
			return
		}

		// Test connection
		if err := db.Ping(); err != nil {
			brandDBErr = fmt.Errorf("failed to ping brand database: %w", err)
			logger.Error("Failed to ping brand database: %v", err)
			return
		}

		// Set connection pool settings (ลดลงครึ่งหนึ่งเพื่อไม่กระทบระบบอื่น)
		db.SetMaxOpenConns(5)
		db.SetMaxIdleConns(2)

		brandDB = db
		logger.Info("Successfully connected to brand_mappings_db")
	})

	return brandDB, brandDBErr
}
