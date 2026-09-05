package microservice

import (
	"errors"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"testing"
)

func TestPersisterTransactionReturnsFailure(t *testing.T) {
	for _, mode := range []string{"callback", "commit", "begin"} {
		t.Run(mode, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			orm, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{DisableAutomaticPing: true})
			if err != nil {
				t.Fatal(err)
			}
			failure := errors.New("transaction rejected")
			begin := mock.ExpectBegin()
			switch mode {
			case "begin":
				begin.WillReturnError(failure)
			case "callback":
				mock.ExpectRollback()
			case "commit":
				mock.ExpectCommit().WillReturnError(failure)
			}
			err = NewPersisterWithDB(orm).Transaction(func(*Persister) error {
				if mode == "callback" {
					return failure
				}
				return nil
			})
			if !errors.Is(err, failure) {
				t.Fatalf("transaction failure was hidden: %v", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
