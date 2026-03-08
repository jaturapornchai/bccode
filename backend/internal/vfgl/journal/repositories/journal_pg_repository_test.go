package repositories_test

import (
	"database/sql"
	"os"
	"smlcloudplatform/internal/mocktest"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/vfgl/journal/models"
	"smlcloudplatform/internal/vfgl/journal/repositories"
	"smlcloudplatform/pkg/microservice"
	"testing"
	"time"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"gorm.io/gorm"
)

type journalRepositoryTestSuite struct {
	db      *gorm.DB
	repo    repositories.JournalPgRepository
	mock    sqlmock.Sqlmock
	journal models.JournalPg
}

func TestJournalRepositoryCreate(t *testing.T) {

	if os.Getenv("SERVERLESS") == "serverless" {
		t.Skip()
	}

	s := &journalRepositoryTestSuite{}
	var (
		db  *sql.DB
		err error
	)

	db, s.db, s.mock, err = mocktest.MockPostgreSQL()
	if err != nil {
		t.Errorf("Failed to open mock sql db, got error: %v", err)
	}
	defer db.Close()

	s.repo = repositories.NewJournalPgRepository(microservice.NewPersisterWithDB(s.db))
	s.journal = models.JournalPg{
		ShopIdentity: common.ShopIdentity{
			ShopID: "SHOPTEST",
		},
		JournalBody: models.JournalBody{
			DocNo:   "TESTDOCNO",
			DocDate: time.Date(2022, 05, 01, 0, 0, 0, 0, time.UTC),
		},
		AccountBook: &[]models.JournalDetailPg{
			{
				Docno:       "TESTDOCNO",
				ShopID:      "SHOPTEST",
				AccountCode: "1000",
				DebitAmount: 100,
			},
		},
	}

	// journal_mock_cols := []string{"shopid", "parid", "docno", "batchid", "docdate", "accountperiod", "accountyear", "accountgroup", "amount", "accountdescription"}
	// rows := sqlmock.NewRows(journal_mock_cols).
	// 	AddRow(
	// 		s.journal.ShopID,
	// 		s.journal.ParID,
	// 		s.journal.Docno,
	// 		s.journal.BatchID,
	// 		s.journal.DocDate,
	// 		s.journal.AccountPeriod,
	// 		s.journal.AccountYear,
	// 		s.journal.AccountGroup,
	// 		s.journal.Amount,
	// 		s.journal.AccountDescription,
	// 	)
	s.mock.ExpectBegin()
	s.mock.ExpectExec(`INSERT INTO "journals"`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// s.mock.ExpectExec(`INSERT INTO "journals_detail"`).
	// 	WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	if err = s.repo.Create(s.journal); err != nil {
		t.Errorf("Failed to insert to gorm db, got error: %v", err)
		t.FailNow()
	}

	err = s.mock.ExpectationsWereMet()
	if err != nil {
		t.Errorf("Failed to meet expectations, got error: %v", err)
	}
}
