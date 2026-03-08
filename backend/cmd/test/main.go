package main

import (
	"encoding/json"
	"fmt"
	"smlcloudplatform/internal/vfgl/journal/models"
	"smlcloudplatform/internal/vfgl/journal/repositories"
	"smlcloudplatform/internal/vfgl/journal/services"
	"smlcloudplatform/mock"
	"smlcloudplatform/pkg/microservice"
)

var repo repositories.JournalPgRepository
var ms *microservice.Microservice

func init() {
	persisterConfig := mock.NewPersisterPostgresqlConfig()
	pst := microservice.NewPersister(persisterConfig)
	repo = repositories.NewJournalPgRepository(pst)
}

func main() {
	testUpdate()
}

func testUpdate() {

	var journal_json string = `
	{
		"id": "628c4ee982bcbf8133668cf6",
		"shopid": "27dcEdktOoaSBYFmnN6G6ett4Jb",
		"guidfixed": "29asDMDazTOCwD7Qm7FDi9GDMu0",
		"parid": "0000000",
		"batchId": "1124541",
		"docno": "JO-202205246DBB8C",
		"docdate": "2022-05-24T03:19:33.545Z",
		"accountperiod": 3,
		"accountyear": 2565,
		"accountgroup": "112",
		"amount": 10,
		"accountdescription": "ทดสอบ",
		"bookcode": "",
		"journaldetail": [
			{
				"accountcode": "10000",
				"accountname": "**สินทรัพย์",
				"debitamount": 5,
				"creditamount": 0
			},
			{
				"accountcode": "11000",
				"accountname": "**สินทรัพย์หมุนเวียน",
				"debitamount": 0,
				"creditamount": 5
			},
			{
				"accountcode": "11010",
				"accountname": "เงินสด - บัญชี 1",
				"debitamount": 5,
				"creditamount": 0
			},
			{
				"accountcode": "10000",
				"accountname": "**สินทรัพย์",
				"debitamount": 0,
				"creditamount": 5
			}
		]
	}`

	doc := models.JournalDoc{}
	err := json.Unmarshal([]byte(journal_json), &doc)

	if err != nil {
		fmt.Errorf("error %v", err)
	}

	journalService := services.NewJournalConsumeService(repo)
	resp, err := journalService.UpSert(doc.ShopID, doc.DocNo, doc)
	if err != nil {
		fmt.Printf("error %s", err.Error())
	}

	fmt.Printf("%v", resp)
}

func testGetJournal() {
	data, err := repo.Get("27dcEdktOoaSBYFmnN6G6ett4Jb", "JO-202206067CFB22")
	if err != nil {
		fmt.Printf("error %v", err)
	}

	fmt.Printf("%v", data)
}

func testCreateJournal() {

	var journal_json string = `{
		"id": "000000000000000000000000",
		"shopid": "27dcEdktOoaSBYFmnN6G6ett4Jb",
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

	doc := models.JournalPg{}
	err := json.Unmarshal([]byte(journal_json), &doc)

	if err != nil {
		fmt.Errorf("error %v", err)
	}

	err = repo.Create(doc)
	if err != nil {
		fmt.Errorf("error %v", err)
	}
}
