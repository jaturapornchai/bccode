//go:build integration

package repositories_test

import (
	"encoding/json"
	"os"
	"smlcloudplatform/internal/config"
	"smlcloudplatform/internal/logger"
	"smlcloudplatform/internal/vfgl/journal/models"
	"smlcloudplatform/internal/vfgl/journal/repositories"
	"smlcloudplatform/mock"
	"smlcloudplatform/pkg/microservice"
	"testing"

	"github.com/stretchr/testify/assert"
)

var journal_json string = `{
	"id": "000000000000000000000000",
	"holdingcode": "27dcEdktOoaSBYFmnN6G6ett4Jb",
	"guidfixed": "2ABh7CJyA7RbeZ1WmdwXWvs0GQa",
	"parid": "0000000",
	"batchId": "",
	"docno": "JO-202206067CFB22",
	"docdate": "2022-06-06T04:11:28.56Z",
	"accountperiod": 1,
	"accountyear": 2022,
	"accountgroup": "1",
	"journaldetail": [
		{
			"accountcode": "11010",
			"accountname": "เงินสด - บัญชี 1 (เงินล้าน) ",
			"debitamount": 1000,
			"creditamount": 0
		},
		{
			"accountcode": "11",
			"accountname": "11",
			"debitamount": 0,
			"creditamount": 1000
		}
	],
	"amount": 1000,
	"accountdescription": "",
	"bookcode": ""
}`

func newRealDBRepository(t *testing.T) repositories.JournalPgRepository {
	t.Helper()
	if os.Getenv("BC_REAL_DB_TESTS") != "1" {
		t.Skip("set BC_REAL_DB_TESTS=1 to run PostgreSQL integration tests")
	}

	logger.NewAppLogger(config.NewLoggerConfig())

	persisterConfig := mock.NewPersisterPostgresqlConfig()
	pst := microservice.NewPersister(persisterConfig)
	repo := repositories.NewJournalPgRepository(pst)
	pst.AutoMigrate(
		models.JournalPg{},
		models.JournalDetailPg{},
		models.JournalVatPg{},
		models.JournalTaxPg{},
	)
	return repo
}

func TestJournalRepositoryRealDBCreate(t *testing.T) {
	repo := newRealDBRepository(t)

	assert.NotNil(t, repo, "Failed to Init Repo")
	doc := models.JournalPg{}
	err := json.Unmarshal([]byte(journal_json), &doc)

	assert.Nil(t, err, "Failed Unmarshal Json to JournalDoc Create1")
	assert.Equal(t, doc.DocNo, "JO-202206067CFB22", "Failed Doc No Not Match")

	err = repo.Create(doc)
	assert.Nil(t, err, "Failed Unmarshal Json to JournalDoc Create1")
}

func TestJournalRepositoryRealDBGetAssertErrNotFound(t *testing.T) {
	repo := newRealDBRepository(t)

	assert.NotNil(t, repo, "Failed to Init Repo")

	_, err := repo.Get("TESTSHOP", "DOC01")
	assert.Error(t, err, "Failed Get Journal Doc NotFound But No Error")
}

func TestCreateAndDeleteJournal(t *testing.T) {
	repo := newRealDBRepository(t)

	json_str := `{
		"id": "62cdc14ca3f6ef3ca30543e8",
		"holdingcode": "2BYWCndV194TYXVEO7NlRLuJYWY",
		"guidfixed": "2Br5noZ5LmgRuQLwYrpreUl2J9a",
		"batchId": "",
		"docno": "JO-20220713014532831-1",
		"docdate": "2022-07-12T18:45:32.068Z",
		"documentref": "",
		"accountperiod": 7,
		"accountyear": 2565,
		"accountgroup": "2",
		"amount": 200,
		"accountdescription": "ชำระค่าธรรมเนียมสมัครสมาชิก#6500000049",
		"bookcode": "2",
		"vats": [],
		"taxes": [],
		"parid": "",
		"journaldetail": [
		  {
			"accountcode": "11020",
			"accountname": "เงินสด - บัญชี 2",
			"debitamount": 200,
			"creditamount": 0
		  },
		  {
			"accountcode": "43010",
			"accountname": "รายได้ - ค่าธรรมเนียม-แรกเข้า",
			"debitamount": 0,
			"creditamount": 200
		  }
		]
	  }
	`
	assert.NotNil(t, repo, "Failed to Init Repo")

	doc := models.JournalPg{}
	err := json.Unmarshal([]byte(json_str), &doc)
	assert.Nil(t, err, "Failed Unmarshal Json to JournalDoc Create1")

	err = repo.Create(doc)
	assert.Nil(t, err, "Failed Unmarshal Json to JournalDoc Create1")

	err = repo.Delete(doc.HoldingCode, doc.DocNo)
	assert.Nil(t, err, "Failed Unmarshal Json to JournalDoc Create1")

}

func TestJournalRepositoryRealDBCreateWithTaxAndVat(t *testing.T) {
	repo := newRealDBRepository(t)

	assert.NotNil(t, repo, "Failed to Init Repo")

	jsonStr := `{
	"id": "694cb8ec3586a1b47bb12de9",
	"holdingcode": "2V5zu2gmRgd7sgWj3g6gu7mxYk0",
	"guidfixed": "37Jz2mblpBhQzJFfOMRrmjgaUPs",
	"batchid": "",
	"docno": "JO-20251225CB8F0D",
	"docdate": "2025-12-25T21:00:00Z",
	"documentref": "",
	"accountperiod": 24,
	"accountyear": 2568,
	"accountgroup": "",
	"amount": 1500,
	"accountdescription": "",
	"bookcode": "03",
	"vats": [
		{
			"vatdocno": "JO-20251225CB8F0D",
			"vatdate": "2025-12-26T08:07:55.078Z",
			"vattype": 0,
			"vatmode": 1,
			"vatperiod": 12,
			"vatyear": 2568,
			"vatbase": 1000,
			"vatrate": 70,
			"vatamount": 700,
			"exceptvat": 0,
			"vatsubmit": false,
			"remark": "",
			"custtaxid": "15399002684487",
			"cust_name": "โก้ เทพ",
			"custtype": 0,
			"organization": 0,
			"branchcode": "00000",
			"address": ""
		}
	],
	"taxes": [
		{
			"taxdocno": "TAX1111",
			"taxdate": "2025-12-25T21:00:00Z",
			"taxtype": 1,
			"taxamount": 0,
			"custtaxid": "032156498231231",
			"cust_name": "โก้ๆ",
			"custtype": 0,
			"organization": 0,
			"branchcode": "",
			"address": "",
			"details": [
				{
					"taxbase": 1000,
					"taxrate": 3,
					"taxamount": 30,
					"description": "บริการ"
				},
				{
					"taxbase": 0,
					"taxrate": 0,
					"taxamount": 0,
					"description": ""
				}
			]
		}
	],
	"journaltype": 0,
	"exdocrefno": "",
	"exdocrefdate": "0001-01-01T00:00:00Z",
	"docformat": "",
	"appname": "",
	"debtaccounttype": 0,
	"creditor": {
		"guidfixed": "",
		"code": "",
		"personal_type": 0,
		"customertype": 0,
		"branch_number": "",
		"taxid": "",
		"names": null,
		"addressforbilling": {
			"guid": "",
			"address": null,
			"country_code": "",
			"province_code": "",
			"district_code": "",
			"sub_district_code": "",
			"zip_code": "",
			"contactnames": null,
			"phone_primary": "",
			"phone_secondary": "",
			"latitude": 0,
			"longitude": 0
		}
	},
	"debtor": {
		"guidfixed": "",
		"code": "",
		"personal_type": 0,
		"customertype": 0,
		"branch_number": "",
		"taxid": "",
		"names": null,
		"addressforbilling": {
			"guid": "",
			"address": null,
			"country_code": "",
			"province_code": "",
			"district_code": "",
			"sub_district_code": "",
			"zip_code": "",
			"contactnames": null,
			"phone_primary": "",
			"phone_secondary": "",
			"latitude": 0,
			"longitude": 0
		}
	},
	"jobguidfixed": "",
	"journaldetail": [
		{
			"accountcode": "111110",
			"accountname": "เงินสดในมือ",
			"debitamount": 1500,
			"creditamount": 0
		},
		{
			"accountcode": "410010",
			"accountname": "รายได้จากการขายสินค้า",
			"debitamount": 0,
			"creditamount": 1500
		}
	],
	"createdby": "",
	"createdat": "0001-01-01T00:00:00Z"
}`

	doc := models.JournalPg{}
	err := json.Unmarshal([]byte(jsonStr), &doc)
	assert.Nil(t, err, "Failed Unmarshal Json to JournalDoc")

	err = repo.Create(doc)
	assert.Nil(t, err, "Failed to Create Journal with Tax and Vat")

	// Clean up
	// err = repo.Delete(doc.HoldingCode, doc.DocNo)
	// assert.Nil(t, err, "Failed to Delete Journal")
}
