package mocktest

import (
	"database/sql"
	"errors"
	"fmt"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func MockPostgreSQL() (*sql.DB, *gorm.DB, sqlmock.Sqlmock, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, nil, nil, errors.New(fmt.Sprintf("Failed to open mock sql db, got error: %v", err))
	}

	if db == nil {
		return nil, nil, nil, errors.New("mock db is null")
	}

	dialector := postgres.New(postgres.Config{
		DSN:                  "sqlmock_db_0",
		DriverName:           "postgres",
		Conn:                 db,
		PreferSimpleProtocol: true,
	})

	gormdb, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, nil, nil, errors.New(fmt.Sprintf("Failed to open gorm v2 db, got error: %v", err))
	}

	if gormdb == nil {
		return nil, nil, nil, errors.New("gorm db is null")
	}
	return db, gormdb, mock, nil
}
