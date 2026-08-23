package journalreport_test

import (
	"context"
	"fmt"
	chartofaccountModel "smlcloudplatform/internal/vfgl/chartofaccount/models"
	"smlcloudplatform/internal/vfgl/journalreport"
	"smlcloudplatform/internal/vfgl/journalreport/models"
	"testing"
	"time"

	mocktest "smlcloudplatform/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockJournalReportPgRepository struct {
	mock.Mock
}

func (m *MockJournalReportPgRepository) GetDataTrialBalance(holdingCode string, accountGroup string, includeCloseAccountMode bool, startDate time.Time, endDate time.Time) ([]models.TrialBalanceSheetAccountDetail, error) {
	ret := m.Called(holdingCode, accountGroup, includeCloseAccountMode, startDate, endDate)
	return ret.Get(0).([]models.TrialBalanceSheetAccountDetail), ret.Error(1)
}

func (m *MockJournalReportPgRepository) GetDataProfitAndLoss(holdingCode string, accountGroup string, includeCloseAccountMode bool, startDate time.Time, endDate time.Time) ([]models.ProfitAndLossSheetAccountDetail, error) {
	ret := m.Called(holdingCode, accountGroup, includeCloseAccountMode, startDate, endDate)
	return ret.Get(0).([]models.ProfitAndLossSheetAccountDetail), ret.Error(1)
}

func (m *MockJournalReportPgRepository) GetDataBalanceSheet(holdingCode string, accountGroup string, includeCloseAccountMode bool, endDate time.Time) ([]models.BalanceSheetAccountDetail, error) {
	ret := m.Called(holdingCode, accountGroup, includeCloseAccountMode, endDate)
	return ret.Get(0).([]models.BalanceSheetAccountDetail), ret.Error(1)
}

func (m *MockJournalReportPgRepository) GetDataLedgerAccount(holdingCode string, accountGroup string, creditorCode string, debtorCode string, consolidateAccountCode string, accountCodeRanges []models.LedgerAccountCodeRange, bookCode string, startDate time.Time, endDate time.Time) ([]models.LedgerAccountRaw, error) {
	ret := m.Called(holdingCode, accountGroup, creditorCode, debtorCode, consolidateAccountCode, accountCodeRanges, bookCode, startDate, endDate)
	return ret.Get(0).([]models.LedgerAccountRaw), ret.Error(1)
}

func (m *MockJournalReportPgRepository) GetMultiShopRevenue(holdingCodes []string, startDate time.Time, endDate time.Time) ([]models.MultiShopRevenueRaw, error) {
	ret := m.Called(holdingCodes, startDate, endDate)
	return ret.Get(0).([]models.MultiShopRevenueRaw), ret.Error(1)
}

type MockJournalReportMongoRepository struct {
	mock.Mock
}

func (m *MockJournalReportMongoRepository) FindCountDetailByDocs(ctx context.Context, holdingCode string, docs []string) ([]models.JournalSummary, error) {
	ret := m.Called(ctx, holdingCode, docs)
	return ret.Get(0).([]models.JournalSummary), ret.Error(1)
}
func (m *MockJournalReportMongoRepository) FindCountImageByDocs(ctx context.Context, holdingCode string, docs []string) ([]models.JournalImageSummary, error) {
	ret := m.Called(ctx, holdingCode, docs)
	return ret.Get(0).([]models.JournalImageSummary), ret.Error(1)
}
func (m *MockJournalReportMongoRepository) FindImageCountByShops(ctx context.Context, holdingCodes []string, startDate time.Time, endDate time.Time) ([]models.ShopImageCount, error) {
	ret := m.Called(ctx, holdingCodes, startDate, endDate)
	return ret.Get(0).([]models.ShopImageCount), ret.Error(1)
}

func (m *MockJournalReportMongoRepository) GetDocNosByShopsAndDateRange(ctx context.Context, holdingCodes []string, startDate time.Time, endDate time.Time) (map[string][]string, error) {
	ret := m.Called(ctx, holdingCodes, startDate, endDate)
	return ret.Get(0).(map[string][]string), ret.Error(1)
}

func (m *MockJournalReportMongoRepository) CountImagesByDocNos(ctx context.Context, shopDocNos map[string][]string) ([]models.ShopImageCount, error) {
	ret := m.Called(ctx, shopDocNos)
	return ret.Get(0).([]models.ShopImageCount), ret.Error(1)
}

func (m *MockJournalReportMongoRepository) CountAllImagesByShops(ctx context.Context, holdingCodes []string) ([]models.ShopImageCount, error) {
	ret := m.Called(ctx, holdingCodes)
	return ret.Get(0).([]models.ShopImageCount), ret.Error(1)
}

func TestProcessBalanceSheetReport(t *testing.T) {

	var balances = journalreport.MockBalanceSheetDetailReport()
	endDate := time.Date(2022, 05, 31, 0, 0, 0, 0, time.UTC)

	repo := new(MockJournalReportPgRepository)
	repo.On("GetDataBalanceSheet", "TESTSHOP", "01", false, endDate).Return(balances, nil)

	var liabilities []models.BalanceSheetAccountDetail

	fixReportDate := time.Now()
	want := &models.BalanceSheetReport{
		ReportDate:   fixReportDate,
		EndDate:      endDate,
		AccountGroup: "01",
		Assets: &[]models.BalanceSheetAccountDetail{
			{
				ChartOfAccountPG: chartofaccountModel.ChartOfAccountPG{
					AccountCode:     "12101",
					AccountName:     "เงินฝากธนาคาร บัญชี 1 (เงินล้าน)",
					AccountCategory: 1,
				},
				Amount: 10000,
			},
			{
				ChartOfAccountPG: chartofaccountModel.ChartOfAccountPG{
					AccountCode:     "13010",
					AccountName:     "ลูกหนี้เงินกู้ - บัญชี 1 (เงินล้าน)",
					AccountCategory: 1,
				},
				Amount: 20000,
			},
			{
				ChartOfAccountPG: chartofaccountModel.ChartOfAccountPG{
					AccountCode:     "11010",
					AccountName:     "เงินสด - บัญชี 1",
					AccountCategory: 1,
				},
				Amount: 20,
			},
		},
		Liabilities: &liabilities,
		OwnesEquities: &[]models.BalanceSheetAccountDetail{
			{
				ChartOfAccountPG: chartofaccountModel.ChartOfAccountPG{
					AccountCode:     "32010",
					AccountName:     "ทุน - เงินล้าน",
					AccountCategory: 3,
				},
				Amount: 30000,
			},
			{
				ChartOfAccountPG: chartofaccountModel.ChartOfAccountPG{
					AccountName:     "กำไร (ขาดทุน) สุทธิ",
					AccountCategory: 3,
				},
				Amount: 20,
			},
		},
		TotalAssetAmount:                    30020,
		TotalOwnersEquityAmount:             30020,
		TotalLiabilityAndOwnersEquityAmount: 30020,
	}

	repoMongo := new(MockJournalReportMongoRepository)
	repoMongo.On("FindCountDetailByDocs", mock.Anything, "TESTSHOP", []string{"", "DOC001", "DOC002", "DOC003"}).Return([]models.JournalSummary{}, nil)
	repoMongo.On("FindCountImageByDocs", mock.Anything, "TESTSHOP", []string{"", "DOC001", "DOC002", "DOC003"}).Return([]models.JournalImageSummary{}, nil)

	service := journalreport.NewJournalReportService(repo, repoMongo, nil, nil)
	get, err := service.ProcessBalanceSheetReport("TESTSHOP", "01", false, endDate)
	get.ReportDate = fixReportDate
	assert.Nil(t, err, "Error should be nil")

	assert.Equal(t, get, want, "Process BalanceSheet Report Not Match")
}

func TestLedgerAccount(t *testing.T) {
	repo := new(MockJournalReportPgRepository)
	repo.On("GetDataLedgerAccount", "TESTSHOP", "accGroup", "", "", "conAcc", []models.LedgerAccountCodeRange{
		{
			Start: "100000",
			End:   "150000",
		},
	}, "B01", mocktest.MockTime(), mocktest.MockTime()).Return([]models.LedgerAccountRaw{
		{
			RowMode:      -1,
			DocNo:        "",
			DocDate:      mocktest.MockTime(),
			AccountCode:  "AC001",
			AccountName:  "AC Name 1",
			DebitAmount:  0,
			CreditAmount: 0,
			Amount:       75,
		},
		{
			RowMode:      0,
			DocNo:        "DOC001",
			DocDate:      mocktest.MockTime(),
			AccountCode:  "AC001",
			AccountName:  "AC Name 1",
			DebitAmount:  50,
			CreditAmount: 0,
			Amount:       0,
		},
		{
			RowMode:      0,
			DocNo:        "DOC002",
			DocDate:      mocktest.MockTime(),
			AccountCode:  "AC001",
			AccountName:  "AC Name 1",
			DebitAmount:  50,
			CreditAmount: 0,
			Amount:       0,
		},
		{
			RowMode:      -1,
			DocNo:        "",
			DocDate:      mocktest.MockTime(),
			AccountCode:  "AC002",
			AccountName:  "AC Name 2",
			DebitAmount:  0,
			CreditAmount: 0,
			Amount:       200,
		},
		{
			RowMode:      0,
			DocNo:        "DOC003",
			DocDate:      mocktest.MockTime(),
			AccountCode:  "AC002",
			AccountName:  "AC Name 2",
			DebitAmount:  0,
			CreditAmount: 250,
			Amount:       0,
		},
		{
			RowMode:      -1,
			DocNo:        "",
			DocDate:      mocktest.MockTime(),
			AccountCode:  "AC003",
			AccountName:  "AC Name 3",
			DebitAmount:  0,
			CreditAmount: 0,
			Amount:       -50,
		},
		{
			RowMode:      0,
			DocNo:        "",
			DocDate:      mocktest.MockTime(),
			AccountCode:  "AC003",
			AccountName:  "AC Name 3",
			DebitAmount:  100,
			CreditAmount: 0,
			Amount:       0,
		},
		{
			RowMode:      -1,
			DocNo:        "",
			DocDate:      mocktest.MockTime(),
			AccountCode:  "AC004",
			AccountName:  "AC Name 4",
			DebitAmount:  0,
			CreditAmount: 0,
			Amount:       -50,
		},
		{
			RowMode:      0,
			DocNo:        "",
			DocDate:      mocktest.MockTime(),
			AccountCode:  "AC004",
			AccountName:  "AC Name 4",
			DebitAmount:  0,
			CreditAmount: 100,
			Amount:       0,
		},
		{
			RowMode:      -1,
			DocNo:        "",
			DocDate:      mocktest.MockTime(),
			AccountCode:  "AC005",
			AccountName:  "AC Name 5",
			DebitAmount:  0,
			CreditAmount: 0,
			Amount:       0,
		},
		{
			RowMode:      0,
			DocNo:        "",
			DocDate:      mocktest.MockTime(),
			AccountCode:  "AC005",
			AccountName:  "AC Name 5",
			DebitAmount:  100.35,
			CreditAmount: 0,
			Amount:       0,
		},
		{
			RowMode:      0,
			DocNo:        "",
			DocDate:      mocktest.MockTime(),
			AccountCode:  "AC005",
			AccountName:  "AC Name 5",
			DebitAmount:  0,
			CreditAmount: 100.35,
			Amount:       0,
		},
	}, nil)

	repoMongo := new(MockJournalReportMongoRepository)
	repoMongo.On("FindCountDetailByDocs", mock.Anything, "TESTSHOP", []string{"", "DOC001", "DOC002", "DOC003"}).Return([]models.JournalSummary{}, nil)
	repoMongo.On("FindCountImageByDocs", mock.Anything, "TESTSHOP", []string{"", "DOC001", "DOC002", "DOC003"}).Return([]models.JournalImageSummary{}, nil)

	service := journalreport.NewJournalReportService(repo, repoMongo, nil, nil)
	docList, err := service.ProcessLedgerAccount("TESTSHOP", "accGroup", "", "", "conAcc", []models.LedgerAccountCodeRange{
		{
			Start: "100000",
			End:   "150000",
		},
	},
		"B01",
		mocktest.MockTime(), mocktest.MockTime())

	assert.Nil(t, err)
	assert.Equal(t, 5, len(docList))

	assert.Equal(t, 175.0, docList[0].NextBalance)
	assert.Equal(t, -50.0, docList[1].NextBalance)
	assert.Equal(t, 50.0, docList[2].NextBalance)
	assert.Equal(t, -150.0, docList[3].NextBalance)

	for _, detail := range *docList[4].Details {
		fmt.Printf("%f \n", detail.Amount)
	}
}

func TestLedgerAccount2(t *testing.T) {
	repo := new(MockJournalReportPgRepository)
	repo.On("GetDataLedgerAccount", "TESTSHOP", "accGroup", "", "", "conAcc", []models.LedgerAccountCodeRange{
		{
			Start: "100000",
			End:   "150000",
		},
	}, "B01",
		mocktest.MockTime(), mocktest.MockTime()).Return(
		[]models.LedgerAccountRaw{
			{
				RowMode:      -1,
				DocNo:        "",
				DocDate:      mocktest.MockTime(),
				AccountCode:  "AC005",
				AccountName:  "AC Name 5",
				DebitAmount:  0,
				CreditAmount: 0,
				Amount:       0,
			},
			{
				RowMode:      0,
				DocNo:        "",
				DocDate:      mocktest.MockTime(),
				AccountCode:  "AC005",
				AccountName:  "AC Name 5",
				DebitAmount:  0,
				CreditAmount: 511.67,
				Amount:       0,
			},
			{
				RowMode:      0,
				DocNo:        "",
				DocDate:      mocktest.MockTime(),
				AccountCode:  "AC005",
				AccountName:  "AC Name 5",
				DebitAmount:  2391.0,
				CreditAmount: 0,
				Amount:       0,
			},
			{
				RowMode:      0,
				DocNo:        "",
				DocDate:      mocktest.MockTime(),
				AccountCode:  "AC005",
				AccountName:  "AC Name 5",
				DebitAmount:  0,
				CreditAmount: 1879.33,
				Amount:       0,
			},
			{
				RowMode:      0,
				DocNo:        "",
				DocDate:      mocktest.MockTime(),
				AccountCode:  "AC005",
				AccountName:  "AC Name 5",
				DebitAmount:  0,
				CreditAmount: 151.30,
				Amount:       0,
			},
			{
				RowMode:      0,
				DocNo:        "",
				DocDate:      mocktest.MockTime(),
				AccountCode:  "AC005",
				AccountName:  "AC Name 5",
				DebitAmount:  707.0,
				CreditAmount: 0,
				Amount:       0,
			},
			{
				RowMode:      0,
				DocNo:        "",
				DocDate:      mocktest.MockTime(),
				AccountCode:  "AC005",
				AccountName:  "AC Name 5",
				DebitAmount:  0,
				CreditAmount: 555.70,
				Amount:       0,
			},
			{
				RowMode:      0,
				DocNo:        "",
				DocDate:      mocktest.MockTime(),
				AccountCode:  "AC005",
				AccountName:  "AC Name 5",
				DebitAmount:  961.0,
				CreditAmount: 0,
				Amount:       0,
			},
			{
				RowMode:      0,
				DocNo:        "",
				DocDate:      mocktest.MockTime(),
				AccountCode:  "AC005",
				AccountName:  "AC Name 5",
				DebitAmount:  0,
				CreditAmount: 104.0,
				Amount:       0,
			},
			{
				RowMode:      0,
				DocNo:        "",
				DocDate:      mocktest.MockTime(),
				AccountCode:  "AC005",
				AccountName:  "AC Name 5",
				DebitAmount:  0,
				CreditAmount: 205.65,
				Amount:       0,
			},
			{
				RowMode:      0,
				DocNo:        "",
				DocDate:      mocktest.MockTime(),
				AccountCode:  "AC005",
				AccountName:  "AC Name 5",
				DebitAmount:  486.0,
				CreditAmount: 0,
				Amount:       0,
			},
			{
				RowMode:      0,
				DocNo:        "",
				DocDate:      mocktest.MockTime(),
				AccountCode:  "AC005",
				AccountName:  "AC Name 5",
				DebitAmount:  0,
				CreditAmount: 755.35,
				Amount:       0,
			},
			{
				RowMode:      0,
				DocNo:        "",
				DocDate:      mocktest.MockTime(),
				AccountCode:  "AC005",
				AccountName:  "AC Name 5",
				DebitAmount:  0,
				CreditAmount: 382.0,
				Amount:       0,
			},
			{
				RowMode:      0,
				DocNo:        "",
				DocDate:      mocktest.MockTime(),
				AccountCode:  "AC005",
				AccountName:  "AC Name 5",
				DebitAmount:  787.0,
				CreditAmount: 0,
				Amount:       0,
			},
		}, nil)

	repoMongo := new(MockJournalReportMongoRepository)
	repoMongo.On("FindCountDetailByDocs", mock.Anything, "TESTSHOP", []string{""}).Return([]models.JournalSummary{}, nil)
	repoMongo.On("FindCountImageByDocs", mock.Anything, "TESTSHOP", []string{""}).Return([]models.JournalImageSummary{}, nil)

	service := journalreport.NewJournalReportService(repo, repoMongo, nil, nil)
	docList, err := service.ProcessLedgerAccount("TESTSHOP", "accGroup", "", "", "conAcc", []models.LedgerAccountCodeRange{
		{
			Start: "100000",
			End:   "150000",
		},
	},
		"B01",
		mocktest.MockTime(), mocktest.MockTime())

	assert.Nil(t, err)
	assert.Equal(t, 1, len(docList))

	// assert.Equal(t, 0.0, docList[0].NextBalance)

	for _, detail := range *docList[0].Details {
		fmt.Printf("%f \n", detail.Amount)
	}
}
