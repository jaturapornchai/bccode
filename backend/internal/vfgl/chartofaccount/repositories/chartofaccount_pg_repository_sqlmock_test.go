package repositories_test

import (
	"database/sql"
	"regexp"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/vfgl/chartofaccount/models"
	"testing"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type v2Suite struct {
	db             *gorm.DB
	mock           sqlmock.Sqlmock
	chartofaccount models.ChartOfAccountPG
}

func TestCreateChartOfAccount(t *testing.T) {
	s := &v2Suite{}
	var (
		db  *sql.DB
		err error
	)

	db, s.mock, err = sqlmock.New()
	if err != nil {
		t.Errorf("Failed to open mock sql db, got error: %v", err)
	}

	if db == nil {
		t.Error("mock db is null")
	}

	dialector := postgres.New(postgres.Config{
		DSN:                  "sqlmock_db_0",
		DriverName:           "postgres",
		Conn:                 db,
		PreferSimpleProtocol: true,
	})
	s.db, err = gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		t.Errorf("Failed to open gorm v2 db, got error: %v", err)
	}

	if s.db == nil {
		t.Error("gorm db is null")
	}

	s.chartofaccount = models.ChartOfAccountPG{
		ShopIdentity: common.ShopIdentity{
			ShopID: "TESTSHOP",
		},
		PartitionIdentity: common.PartitionIdentity{
			ParID: "",
		},
		AccountCode:            "0001",
		AccountName:            "TEST",
		AccountCategory:        0,
		AccountBalanceType:     0,
		AccountGroup:           "",
		AccountLevel:           0,
		ConsolidateAccountCode: "",
	}

	defer db.Close()

	s.mock.MatchExpectationsInOrder(false)
	s.mock.ExpectBegin()

	s.mock.ExpectExec(
		regexp.QuoteMeta(`INSERT INTO "chartofaccounts" ("shopid","parid","accountcode","accountname","accountcategory","accountbalancetype","accountgroup","accountlevel","consolidateaccountcode")
	                VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`)).
		WithArgs(
			s.chartofaccount.ShopID, s.chartofaccount.ParID, s.chartofaccount.AccountCode, s.chartofaccount.AccountName,
			s.chartofaccount.AccountCategory, s.chartofaccount.AccountBalanceType, s.chartofaccount.AccountGroup, s.chartofaccount.AccountLevel,
			s.chartofaccount.ConsolidateAccountCode).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// s.mock.ExpectQuery(regexp.QuoteMeta(
	// 	`INSERT INTO "chartofaccounts" ("shopid", "accountcode","accountname")
	// 						VALUES ($1,$2,$3) RETURNING "chartofaccounts"."accountcode"`)).
	// 	WithArgs(s.chartofaccount.ShopID, s.chartofaccount.ParID, s.chartofaccount.AccountCode, s.chartofaccount.AccountName, s.chartofaccount.AccountCategory, s.chartofaccount.AccountBalanceType, s.chartofaccount.AccountGroup, s.chartofaccount.AccountLevel, s.chartofaccount.ConsolidateAccountCode).
	// 	WillReturnRows(sqlmock.NewRows([]string{"0001"}).
	// 		AddRow(s.chartofaccount.AccountCode))

	s.mock.ExpectCommit()

	if err = s.db.Create(&s.chartofaccount).Error; err != nil {
		t.Errorf("Failed to insert to gorm db, got error: %v", err)
		t.FailNow()
	}

	err = s.mock.ExpectationsWereMet()
	if err != nil {
		t.Errorf("Failed to meet expectations, got error: %v", err)
	}
}
