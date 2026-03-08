package services_test

import (
	"smlcloudplatform/internal/smsreceive/smstransaction/services"
	"testing"

	"github.com/tj/assert"
)

/*
import (
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/smsreceive/smstransaction/models"
	"smlcloudplatform/internal/smsreceive/smstransaction/services"
	"testing"
	"time"

	utilmock "smlcloudplatform/mock"

	"github.com/smlsoft/mongopagination"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type SmsTransactionRepositoryMock struct {
	mock.Mock
}

func (m *SmsTransactionRepositoryMock) Count(ctx context.Context, shopID string) (int, error) {
	args := m.Called(shopID)
	return args.Int(0), args.Error(1)
}

func (m *SmsTransactionRepositoryMock) Create(ctx context.Context, doc models.SmsTransactionDoc) (string, error) {
	args := m.Called(doc)
	return args.String(0), args.Error(1)
}

func (m *SmsTransactionRepositoryMock) CreateInBatch(docList []models.SmsTransactionDoc) error {
	args := m.Called(docList)
	return args.Error(0)
}

func (m *SmsTransactionRepositoryMock) Update(shopID string, guid string, doc models.SmsTransactionDoc) error {
	args := m.Called(shopID, guid, doc)
	return args.Error(0)
}

func (m *SmsTransactionRepositoryMock) DeleteByGuidfixed(shopID string, guid string, username string) error {
	args := m.Called(shopID, guid, username)
	return args.Error(0)
}

func (m *SmsTransactionRepositoryMock) FindPage(shopID string, searchInFields []string, pageable micromodels.Pageable)([]models.SmsTransactionInfo, mongopagination.PaginationData, error) {
	args := m.Called(shopID, searchInFields, pageable)
	return args.Get(0).([]models.SmsTransactionInfo), args.Get(1).(mongopagination.PaginationData), args.Error(2)
}

func (m *SmsTransactionRepositoryMock) FindByGuid(shopID string, guid string) (models.SmsTransactionDoc, error) {
	args := m.Called(shopID, guid)
	return args.Get(0).(models.SmsTransactionDoc), args.Error(1)
}

func (m *SmsTransactionRepositoryMock) FindByDocIndentityGuid(shopID string, indentityField string, indentityValue interface{}) (models.SmsTransactionDoc, error) {
	args := m.Called(shopID, indentityField, indentityValue)
	return args.Get(0).(models.SmsTransactionDoc), args.Error(1)
}

func (m *SmsTransactionRepositoryMock) FindPage(shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.SmsTransactionInfo, mongopagination.PaginationData, error) {
	args := m.Called(shopID, searchInFields, pageable, sorts)
	return args.Get(0).([]models.SmsTransactionInfo), args.Get(1).(mongopagination.PaginationData), args.Error(2)
}

func (m *SmsTransactionRepositoryMock) FindFilterSms(shopID string, address string, startTime time.Time, endTime time.Time) ([]models.SmsTransactionInfo, error) {
	args := m.Called(shopID, address, startTime, endTime)
	return args.Get(0).([]models.SmsTransactionInfo), args.Error(1)
}

func TestFillterSms(t *testing.T) {

	repo := new(SmsTransactionRepositoryMock)

	mockTime, _ := time.Parse(time.RFC3339, "2022-08-25T03:09:57.335+00:00")

	startTime := mockTime.Add(time.Duration(-5) * time.Minute)
	endTime := mockTime.Add(time.Duration(5) * time.Minute)

	repo.On("FindFilterSms", "TESTSHOP", "kbank", startTime, endTime).Return([]models.SmsTransactionInfo{
		{
			DocIdentity: common.DocIdentity{
				GuidFixed: "GUID001",
			},
			SmsTransaction: models.SmsTransaction{
				TransId:  "001",
				Address:  "kbank",
				Body:     "12/04/63 09:25 บชX231148X รับโอนจากX815923X 1170.00บ คงเหลือ 2160.29บ",
				SendedAt: mockTime,
			},
		},
	}, nil)

	repo.On("FindFilterSms", "TESTSHOP", "kbank", utilmock.MockTime(), utilmock.MockTime()).Return([]models.SmsTransactionInfo{
		{
			DocIdentity: common.DocIdentity{
				GuidFixed: "GUID001",
			},
			SmsTransaction: models.SmsTransaction{
				TransId:  "001",
				Address:  "kbank",
				Body:     "test test test",
				SendedAt: mockTime,
			},
		},
	}, nil)

	type args struct {
		shopID    string
		amount    float64
		startTime time.Time
		endTime   time.Time
	}

	cases := []struct {
		name     string
		args     args
		wantErr  bool
		wantData float64
	}{
		{
			name:     "sms filter sms pass",
			wantErr:  false,
			wantData: 1170.00,
			args: args{
				shopID:    "TESTSHOP",
				amount:    1170.00,
				startTime: startTime,
				endTime:   endTime,
			},
		},
		{
			name:     "sms filter sms failure",
			wantErr:  true,
			wantData: 0.00,
			args: args{
				shopID:    "TESTSHOP",
				amount:    0,
				startTime: utilmock.MockTime(),
				endTime:   utilmock.MockTime(),
			},
		},
		{
			name:     "sms filter sms failure",
			wantErr:  true,
			wantData: 0.00,
			args: args{
				shopID:    "TESTSHOP",
				startTime: utilmock.MockTime(),
				endTime:   utilmock.MockTime(),
			},
		},
	}

	svc := services.NewSmsTransactionHttpService(repo, utilmock.MockGUID, utilmock.MockTime)
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			amount, err := svc.CheckSMS("TESTSHOP", tt.args.amount, tt.args.startTime)

			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.wantData, amount)
			}

		})
	}
}
*/

func TestGetAmountFromPatern(t *testing.T) {
	msg := "12/04/63 09:25 บชX231148X รับโอนจากX815923X 1170.01บ คงเหลือ 2160.29บ"
	pattern := `[0-9]{2}\/[0-9]{2}\/[0-9]{2} [0-9]{2}:[0-9]{2} บชX[0-9].*X (?P<Amount>[0-9].*)บ คงเหลือ [0-9].*บ`

	amount, err := services.GetAmountFromPattern(pattern, msg)

	amountExported := 1170.01

	assert.NoError(t, err)

	assert.Equal(t, amountExported, amount)

}
