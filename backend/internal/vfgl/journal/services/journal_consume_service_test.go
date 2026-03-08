package services_test

import (
	"encoding/json"
	"fmt"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/vfgl/journal/models"
	"smlcloudplatform/internal/vfgl/journal/repositories"
	"smlcloudplatform/internal/vfgl/journal/services"
	"smlcloudplatform/pkg/microservice"
	"testing"
	"time"

	msmock "smlcloudplatform/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type MockJournalRepsitory struct {
	mock.Mock
	ms *microservice.Microservice
}

func (m *MockJournalRepsitory) CreateInBatch(docList []models.JournalPg) error {
	ret := m.Called(docList)
	return ret.Error(0)
}

func (m *MockJournalRepsitory) Create(doc models.JournalPg) error {
	ret := m.Called(doc)
	return ret.Error(0)
}

func (m *MockJournalRepsitory) Update(shopID string, docNo string, doc models.JournalPg) error {
	ret := m.Called(shopID, docNo, doc)
	return ret.Error(0)
}

func (m *MockJournalRepsitory) Delete(shopID string, docNo string) error {
	ret := m.Called(shopID, docNo)
	return ret.Error(0)
}

func (m *MockJournalRepsitory) Get(shopID string, docNo string) (*models.JournalPg, error) {
	ret := m.Called(shopID, docNo)
	return ret.Get(0).(*models.JournalPg), ret.Error(1)
}

func TestJournalConsumeServiceCreated(ms *microservice.Microservice, t *testing.T) {

	get := models.JournalPg{
		ShopIdentity: common.ShopIdentity{
			ShopID: "27dcEdktOoaSBYFmnN6G6ett4Jb",
		},
		PartitionIdentity: common.PartitionIdentity{
			ParID: "0000000",
		},
		JournalBody: models.JournalBody{

			DocNo:              "JO-202206067CFB22",
			DocDate:            time.Date(2022, 6, 6, 4, 11, 28, 56, time.UTC),
			Amount:             1000,
			AccountDescription: "",
			AccountGroup:       "1",
			AccountYear:        2022,
			AccountPeriod:      1,
			BatchID:            "",
		},
		AccountBook: &[]models.JournalDetailPg{
			{
				AccountCode:  "11010",
				AccountName:  "เงินสด - บัญชี 1 (เงินล้าน) ",
				DebitAmount:  1000,
				CreditAmount: 0,
			},
			{
				AccountCode:  "11",
				AccountName:  "11",
				DebitAmount:  0,
				CreditAmount: 1000,
			},
		},
	}

	give := models.JournalDoc{
		JournalData: models.JournalData{
			ShopIdentity: common.ShopIdentity{
				ShopID: "27dcEdktOoaSBYFmnN6G6ett4Jb",
			},
			JournalInfo: models.JournalInfo{
				Journal: models.Journal{
					PartitionIdentity: common.PartitionIdentity{
						ParID: "0000000",
					},
					JournalBody: models.JournalBody{
						DocNo:              "JO-202206067CFB22",
						DocDate:            time.Date(2022, 6, 6, 4, 11, 28, 56, time.UTC),
						Amount:             1000,
						AccountDescription: "",
						AccountGroup:       "1",
						AccountYear:        2022,
						AccountPeriod:      1,
						BatchID:            "",
					},
					AccountBook: &[]models.JournalDetail{
						{
							AccountCode:  "11010",
							AccountName:  "เงินสด - บัญชี 1 (เงินล้าน) ",
							DebitAmount:  1000,
							CreditAmount: 0,
						},
						{
							AccountCode:  "11",
							AccountName:  "11",
							DebitAmount:  0,
							CreditAmount: 1000,
						},
					},
				},
			},
		},
	}

	mockRepo := new(MockJournalRepsitory)
	mockRepo.On("Create", get).Return(nil)

	journalService := services.NewJournalConsumeService(mockRepo)
	_, err := journalService.UpSert(give.ShopID, give.DocNo, give)
	assert.Nil(t, err, "Error should be nil")
}

func TestJournalConsumeServiceUpdate(ms *microservice.Microservice, t *testing.T) {

	get := models.JournalPg{
		JournalBody: models.JournalBody{
			DocNo: "0001",
		},
	}

	give := models.JournalDoc{
		JournalData: models.JournalData{
			JournalInfo: models.JournalInfo{
				Journal: models.Journal{
					JournalBody: models.JournalBody{
						DocNo: "0001",
					},
				},
			},
		},
	}

	mockRepo := new(MockJournalRepsitory)
	mockRepo.On("Update", "SHOPID", "0001", get).Return(nil)

	journalService := services.NewJournalConsumeService(mockRepo)
	_, err := journalService.UpSert("SHOPID", "0001", give)
	assert.Nil(t, err, "Error should be nil")
}

func TestJournalConsumeServiceInsertWhenGetDataNotFound(ms *microservice.Microservice, t *testing.T) {

	giveJournalMongoDB := models.JournalDoc{
		JournalData: models.JournalData{
			ShopIdentity: common.ShopIdentity{
				ShopID: "SHOPID",
			},
			JournalInfo: models.JournalInfo{
				Journal: models.Journal{
					JournalBody: models.JournalBody{
						DocNo:   "0001",
						BatchID: "",
						DocDate: time.Date(2022, 05, 01, 0, 0, 0, 0, time.UTC),
					},
				},
			},
		},
	}

	giveJournalPG := models.JournalPg{
		ShopIdentity: common.ShopIdentity{
			ShopID: "SHOPID",
		},
		JournalBody: giveJournalMongoDB.JournalBody,
	}

	mockRepo := new(MockJournalRepsitory)
	mockRepo.On("Get", "SHOPID", "0001").Return(&models.JournalPg{}, gorm.ErrRecordNotFound)
	mockRepo.On("Create", giveJournalPG).Return(nil)

	journalService := services.NewJournalConsumeService(mockRepo)
	get, err := journalService.UpSert("SHOPID", "0001", giveJournalMongoDB)

	assert.Nil(t, err, "Failed Upsert Journal Comsume")
	assert.NotNil(t, get, "Failed Upsert Data is Nil")
	assert.Equal(t, get, &giveJournalPG, "Failed After Upsert Consume Data")
}

func TestJournalConsumeServiceUpdateWhenFoundOldData(ms *microservice.Microservice, t *testing.T) {
	giveJournalMongoDB := models.JournalDoc{
		JournalData: models.JournalData{
			ShopIdentity: common.ShopIdentity{
				ShopID: "SHOPID",
			},
			JournalInfo: models.JournalInfo{
				Journal: models.Journal{
					JournalBody: models.JournalBody{
						DocNo:   "DOC0001",
						BatchID: "",
						DocDate: time.Date(2022, 05, 01, 0, 0, 0, 0, time.UTC),
					},

					AccountBook: &[]models.JournalDetail{
						{
							AccountCode: "1000",
							DebitAmount: 1200,
						},
						{
							AccountCode: "1200",
							DebitAmount: 200,
						},
						{
							AccountCode:  "4000",
							CreditAmount: 1400,
						},
					},
				},
			},
		},
	}

	giveJournalPG := models.JournalPg{
		ShopIdentity: common.ShopIdentity{
			ShopID: "SHOPID",
		},
		JournalBody: giveJournalMongoDB.JournalBody,
		AccountBook: &[]models.JournalDetailPg{
			{
				ID:          1,
				AccountCode: "1000",
				DebitAmount: 1000,
			},
			{
				ID:           2,
				AccountCode:  "4000",
				CreditAmount: 1000,
			},
		},
	}

	want := models.JournalPg{
		ShopIdentity: common.ShopIdentity{
			ShopID: "SHOPID",
		},
		JournalBody: giveJournalMongoDB.JournalBody,
		AccountBook: &[]models.JournalDetailPg{
			{
				ID:          1,
				AccountCode: "1000",
				DebitAmount: 1200,
			},
			{
				AccountCode: "1200",
				DebitAmount: 200,
			},
			{
				AccountCode:  "4000",
				CreditAmount: 1400,
			},
		},
	}

	mockRepo := new(MockJournalRepsitory)
	mockRepo.On("Get", giveJournalMongoDB.ShopID, giveJournalMongoDB.DocNo).Return(&giveJournalPG, nil)
	mockRepo.On("Update", giveJournalMongoDB.ShopID, giveJournalMongoDB.DocNo, want).Return(nil)

	journalService := services.NewJournalConsumeService(mockRepo)
	get, err := journalService.UpSert("SHOPID", "DOC0001", giveJournalMongoDB)

	assert.Nil(t, err, "Failed Upsert Journal Comsume")
	assert.NotNil(t, get, "Failed Upsert Data is Nil")
	//assert.Equal(t, &get, want, "Failed After Upsert Consume Data")
}

func TestConsumerServiceCreateDocFromJson(t *testing.T) {
	jsonStr := `
	{"id":"000000000000000000000000","shopid":"2E1NVOURRw9sxHxDFfdnmamPWXI","guidfixed":"2E4CzDZN07hcoq3mesUDWpI5upa","batchId":"","docno":"IV6506006","docdate":"2022-06-16T17:00:00Z","documentref":"","accountperiod":6,"accountyear":2565,"accountgroup":"0001","amount":8639.71,"accountdescription":"","bookcode":"01","vats":[],"taxes":[],"journaltype":0,"parid":"0000000","journaldetail":[{"accountcode":"115840","accountname":"ค่าภาษีเงินได้นิติบุคคลถูกหัก.-ณ.ที่จ่าย","debitamount":86.32,"creditamount":0},{"accountcode":"111110","accountname":"เงินสดในมือ","debitamount":8553.39,"creditamount":0},{"accountcode":"410010","accountname":"รายได้จากการขายสินค้า","debitamount":0,"creditamount":8521.21},{"accountcode":"410010","accountname":"รายได้จากการขายสินค้า","debitamount":0,"creditamount":29},{"accountcode":"410010","accountname":"รายได้จากการขายสินค้า","debitamount":0,"creditamount":81.75},{"accountcode":"215500","accountname":"ค่าภาษีขาย","debitamount":0,"creditamount":7.75}]}
	`

	fmt.Print(jsonStr)
}

func TestConsumerServiceCreateFromJsonFailed(ms *microservice.Microservice, t *testing.T) {

	persisterConfig := msmock.NewPersisterPostgresqlConfig()
	pst := microservice.NewPersister(persisterConfig)
	repo := repositories.NewJournalPgRepository(pst)
	pst.AutoMigrate(
		models.JournalPg{},
		models.JournalDetailPg{},
	)

	jsonStr := `{"id":"63106c165ae33da0f2f6a1a2","shopid":"2BYWCndV194TYXVEO7NlRLuJYWY","guidfixed":"2E9vUVOMvjkg2QCJOV7tDB5GxyL","batchId":"","docno":"JO-20220901F9CC9F","docdate":"2017-12-30T17:00:00Z","documentref":"","accountperiod":12,"accountyear":2560,"accountgroup":"1","amount":2314735,"accountdescription":"ยอดยกมาปี 2560","bookcode":"1","vats":[],"taxes":[],"parid":"0000000","journaldetail":[{"accountcode":"11010","accountname":"เงินสด - บัญชี 1 (เงินล้าน) ","debitamount":7084,"creditamount":0},{"accountcode":"12111","accountname":"เงินฝากธนาคาร บัญชี 1 (เงินล้าน) ธนาคารออมสิน","debitamount":1428252,"creditamount":0},{"accountcode":"13010","accountname":"ลูกหนี้เงินกู้ - บัญชี 1 (เงินล้าน)","debitamount":879399,"creditamount":0},{"accountcode":"32010","accountname":"ทุน - บัญชี 1 (เงินล้าน)","debitamount":0,"creditamount":1000000},{"accountcode":"32020","accountname":"ทุน - โครงการ 3A","debitamount":0,"creditamount":200000},{"accountcode":"32030","accountname":"ทุน - เงินเพิ่มทุนระยะ 2","debitamount":0,"creditamount":1000000},{"accountcode":"33104","accountname":"เงินประกันความเสี่ยง - กำไรที่จัดสรร - บัญชี 1 (เงินล้าน) ","debitamount":0,"creditamount":95400},{"accountcode":"33105","accountname":"เงินสมทบกองทุน - กำไรที่จัดสรร - บัญชี 1 (เงินล้าน) ","debitamount":0,"creditamount":11050},{"accountcode":"34010","accountname":"กำไรสะสม (ขาดทุน) สะสม บัญชี 1 (เงินล้าน) ","debitamount":0,"creditamount":0},{"accountcode":"35010","accountname":"กำไร ( ขาดทุน ) บัญชี 1 (เงินล้าน) ","debitamount":0,"creditamount":0},{"accountcode":"41010","accountname":"รายได้ - ดอกเบี้ยเงินกู้ - บัญชี 1 (เงินล้าน) ","debitamount":0,"creditamount":8200},{"accountcode":"45010","accountname":"รายได้ - ดอกเบี้ยเงินฝากธนาคาร -บัญชี 1 (เงินล้าน) ","debitamount":0,"creditamount":85}]}`

	var journalPg models.JournalDoc
	json.Unmarshal([]byte(jsonStr), &journalPg)

	assert.NotNil(t, journalPg, "Failed Upsert Data is Nil")
	journalService := services.NewJournalConsumeService(repo)
	journalPG, err := journalService.UpSert(journalPg.ShopID, journalPg.DocNo, journalPg)
	assert.Nil(t, err, "Failed Upsert Data is Nil")

	assert.NotNil(t, journalPG, "Failed Upsert Data is Nil")
}
