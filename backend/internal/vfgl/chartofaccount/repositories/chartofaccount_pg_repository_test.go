package repositories_test

import (
	"database/sql"
	"regexp"
	"smlcloudplatform/internal/mocktest"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/vfgl/chartofaccount/models"
	"smlcloudplatform/internal/vfgl/chartofaccount/repositories"
	"smlcloudplatform/pkg/microservice"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"gorm.io/gorm"
)

// var repo repositories.ChartOfAccountPgRepository

// func init() {
// 	persisterConfig := mock.NewPersisterPostgresqlConfig()
// 	pst := microservice.NewPersister(persisterConfig)
// 	repo = repositories.NewChartOfAccountPgRepository(pst)
// }

// func TestChartOfAccountRepositoryCreate(t *testing.T) {

// 	assert := assert.New(t)
// 	assert.NotNil(repo)

// 	give := &models.ChartOfAccountPG{
// 		ShopIdentity: models.ShopIdentity{
// 			ShopID: "SHOPTEST",
// 		},
// 		AccountCode: "10000",
// 		AccountName: "เงินสด",
// 	}

// 	err := repo.Create(*give)
// 	assert.Nil(err)

// 	get, err := repo.Get("10000")
// 	assert.NotNil(get)

// }

type chartOfAccountRepositoryTestSuite struct {
	db             *gorm.DB
	repo           repositories.ChartOfAccountPgRepository
	mock           sqlmock.Sqlmock
	chartofaccount models.ChartOfAccountPG
}

func TestChartOfAccountRepositoryCreate(t *testing.T) {
	s := &chartOfAccountRepositoryTestSuite{}

	var (
		db  *sql.DB
		err error
	)

	// move to /pgk/mocktest/postgreqldb.go
	// db, s.mock, err = sqlmock.New()
	// if err != nil {
	// 	t.Errorf("Failed to open mock sql db, got error: %v", err)
	// }

	// if db == nil {
	// 	t.Error("mock db is null")
	// }

	// dialector := postgres.New(postgres.Config{
	// 	DSN:                  "sqlmock_db_0",
	// 	DriverName:           "postgres",
	// 	Conn:                 db,
	// 	PreferSimpleProtocol: true,
	// })

	// s.db, err = gorm.Open(dialector, &gorm.Config{})
	// if err != nil {
	// 	t.Errorf("Failed to open gorm v2 db, got error: %v", err)
	// }

	// if s.db == nil {
	// 	t.Error("gorm db is null")
	// }

	db, s.db, s.mock, err = mocktest.MockPostgreSQL()
	if err != nil {
		t.Errorf("Failed to open mock sql db, got error: %v", err)
	}
	defer db.Close()

	s.repo = repositories.NewChartOfAccountPgRepository(microservice.NewPersisterWithDB(s.db))

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

	if err = s.repo.Create(s.chartofaccount); err != nil {
		t.Errorf("Failed to insert to gorm db, got error: %v", err)
		t.FailNow()
	}

	err = s.mock.ExpectationsWereMet()
	if err != nil {
		t.Errorf("Failed to meet expectations, got error: %v", err)
	}
}

func TestChartOfAccountRepositoryGetByShopIDAndAccountCode(t *testing.T) {
	s := &chartOfAccountRepositoryTestSuite{}

	var (
		db  *sql.DB
		err error
	)
	db, s.db, s.mock, err = mocktest.MockPostgreSQL()
	if err != nil {
		t.Errorf("Failed to open mock sql db, got error: %v", err)
	}
	defer db.Close()

	s.repo = repositories.NewChartOfAccountPgRepository(microservice.NewPersisterWithDB(s.db))

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

	colums := []string{"shopid", "parid", "accountcode", "accountname", "accountcategory", "accountbalancetype", "accountgroup", "accountlevel", "consolidateaccountcode"}
	rows := sqlmock.NewRows(colums).
		AddRow(
			s.chartofaccount.ShopIdentity.ShopID,
			s.chartofaccount.ParID,
			s.chartofaccount.AccountCode,
			s.chartofaccount.AccountName,
			s.chartofaccount.AccountCategory,
			s.chartofaccount.AccountBalanceType,
			s.chartofaccount.AccountGroup,
			s.chartofaccount.AccountLevel,
			s.chartofaccount.ConsolidateAccountCode,
		)

	s.mock.MatchExpectationsInOrder(false)
	//s.mock.ExpectBegin()

	s.mock.ExpectQuery(
		regexp.QuoteMeta(`SELECT * FROM "chartofaccounts" WHERE shopid=$1 AND accountcode=$2`)).
		WithArgs(s.chartofaccount.ShopID, s.chartofaccount.AccountCode).
		WillReturnRows(rows)

	get, err := s.repo.Get(s.chartofaccount.ShopID, s.chartofaccount.AccountCode)
	if err != nil {
		t.Errorf("Failed to insert to gorm db, got error: %v", err)
		t.FailNow()
	}

	assert.Equal(t, get, &s.chartofaccount, "Failed get data should be equal")

	err = s.mock.ExpectationsWereMet()
	if err != nil {
		t.Errorf("Failed to meet expectations, got error: %v", err)
	}
}

func TestChartOfAccountRepositoryGetByShopIDAndAccountCodeAssertNotFoundData(t *testing.T) {
	s := &chartOfAccountRepositoryTestSuite{}

	var (
		db  *sql.DB
		err error
	)
	db, s.db, s.mock, err = mocktest.MockPostgreSQL()
	if err != nil {
		t.Errorf("Failed to open mock sql db, got error: %v", err)
	}
	defer db.Close()

	s.repo = repositories.NewChartOfAccountPgRepository(microservice.NewPersisterWithDB(s.db))

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

	// colums := []string{"shopid", "parid", "accountcode", "accountname", "accountcategory", "accountbalancetype", "accountgroup", "accountlevel", "consolidateaccountcode"}
	// rows := sqlmock.NewRows(colums)

	s.mock.MatchExpectationsInOrder(false)
	//s.mock.ExpectBegin()

	s.mock.ExpectQuery(
		regexp.QuoteMeta(`SELECT * FROM "chartofaccounts" WHERE shopid=$1 AND accountcode=$2`)).
		WithArgs(s.chartofaccount.ShopID, s.chartofaccount.AccountCode).
		WillReturnError(gorm.ErrRecordNotFound)
		//WillReturnRows(rows)

	get, err := s.repo.Get(s.chartofaccount.ShopID, s.chartofaccount.AccountCode)
	assert.NotNil(t, err, "Failed to assert not found record")
	if get != nil {
		t.Errorf("Failed on Get Blank Data from gorm db, got Data: %v", get)
		t.FailNow()
	}

	err = s.mock.ExpectationsWereMet()
	if err != nil {
		t.Errorf("Failed to meet expectations, got error: %v", err)
	}
}
