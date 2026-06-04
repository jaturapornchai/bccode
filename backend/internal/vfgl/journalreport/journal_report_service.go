package journalreport

import (
	"context"
	"fmt"
	authModels "smlcloudplatform/internal/authentication/models"
	"smlcloudplatform/internal/shop"
	chartofaccountModel "smlcloudplatform/internal/vfgl/chartofaccount/models"
	"smlcloudplatform/internal/vfgl/journalreport/models"
	"smlcloudplatform/internal/vfgl/journalreport/usecase"
	"time"

	"github.com/shopspring/decimal"
)

type IJournalReportService interface {
	ProcessTrialBalanceSheetReport(holdingCode string, accountGroup string, includeCloseAccountMode bool, startDate time.Time, endDate time.Time) (*models.TrialBalanceSheetReport, error)
	ProcessProfitAndLossSheetReport(holdingCode string, accountGroup string, includeCloseAccountMode bool, startDate time.Time, endDate time.Time) (*models.ProfitAndLossSheetReport, error)
	ProcessBalanceSheetReport(holdingCode string, accountGroup string, includeCloseAccountMode bool, endDate time.Time) (*models.BalanceSheetReport, error)
	ProcessLedgerAccount(holdingCode string, accountGroup string, creditorCode string, debtorCode string, consolidateAccountCode string, accountRanges []models.LedgerAccountCodeRange, bookCode string, startDate time.Time, endDate time.Time) ([]models.LedgerAccount, error)
	ProcessMultiShopDashboard(username string, holdingCodes []string, startDate time.Time, endDate time.Time) (*models.MultiShopDashboardResponse, error)
}

type JournalReportService struct {
	repoPg         IJournalReportPgRepository
	repoMongo      IJournalReportMongoRepository
	shopUserRepo   shop.IShopUserRepository
	shopRepo       shop.IShopRepository
	usecase        usecase.ITrialBalanceSheetReportUsecase
	contextTimeout time.Duration
}

func NewJournalReportService(
	repoPg IJournalReportPgRepository,
	repoMongo IJournalReportMongoRepository,
	shopUserRepo shop.IShopUserRepository,
	shopRepo shop.IShopRepository,
) JournalReportService {

	contextTimeout := time.Duration(15) * time.Second

	usecase := &usecase.TrialBalanceSheetReportUsecase{}

	return JournalReportService{
		repoPg:         repoPg,
		repoMongo:      repoMongo,
		shopUserRepo:   shopUserRepo,
		shopRepo:       shopRepo,
		usecase:        usecase,
		contextTimeout: contextTimeout,
	}
}

func (svc JournalReportService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc JournalReportService) ProcessTrialBalanceSheetReport(holdingCode string, accountGroup string, includeCloseAccountMode bool, startDate time.Time, endDate time.Time) (*models.TrialBalanceSheetReport, error) {
	// mock := MockTrialBalanceSheetReport(holdingCode, accountGroup, startDate, endDate)
	// return mock, nil
	details, err := svc.repoPg.GetDataTrialBalance(holdingCode, accountGroup, includeCloseAccountMode, startDate, endDate)

	var totalBalanceDebit float64
	var totalBalanceCredit float64
	var totalAmountDebit float64
	var totalAmountCredit float64
	var totalNextBalanceDebit float64
	var totalnextBalanceCredit float64

	for index, v := range details {

		// is lower than zero
		isBalanceDebit := svc.usecase.IsAmountDebitSide(v.AccountCategory, v.BalanceAmount)
		if isBalanceDebit {
			details[index].BalanceDebitAmount = svc.usecase.DisplayAmount(v.BalanceAmount)
		} else {
			details[index].BalanceCreditAmount = svc.usecase.DisplayAmount(v.BalanceAmount)
		}

		isDebit := svc.usecase.IsAmountDebitSide(v.AccountCategory, v.Amount)
		if isDebit {
			details[index].DebitAmount = svc.usecase.DisplayAmount(v.Amount)
		} else {
			details[index].CreditAmount = svc.usecase.DisplayAmount(v.Amount)
		}

		isNextDebit := svc.usecase.IsAmountDebitSide(v.AccountCategory, v.NextBalanceAmount)
		if isNextDebit {
			details[index].NextBalanceDebitAmount = svc.usecase.DisplayAmount(v.NextBalanceAmount)
		} else {
			details[index].NextBalanceCreditAmount = svc.usecase.DisplayAmount(v.NextBalanceAmount)
		}

		totalBalanceDebit += details[index].BalanceDebitAmount
		totalBalanceCredit += details[index].BalanceCreditAmount
		totalAmountDebit += details[index].DebitAmount
		totalAmountCredit += details[index].CreditAmount
		totalNextBalanceDebit += details[index].NextBalanceDebitAmount
		totalnextBalanceCredit += details[index].NextBalanceCreditAmount
	}

	result := &models.TrialBalanceSheetReport{
		ReportDate:             time.Now(),
		StartDate:              startDate,
		EndDate:                endDate,
		AccountGroup:           accountGroup,
		AccountDetails:         &details,
		TotalBalanceDebit:      totalBalanceDebit,
		TotalBalanceCredit:     totalBalanceCredit,
		TotalAmountDebit:       totalAmountDebit,
		TotalAmountCredit:      totalAmountCredit,
		TotalNextBalanceDebit:  totalNextBalanceDebit,
		TotalNextBalanceCredit: totalnextBalanceCredit,
	}
	return result, err
}

func (svc JournalReportService) ProcessProfitAndLossSheetReport(holdingCode string, accountGroup string, includeCloseAccountMode bool, startDate time.Time, endDate time.Time) (*models.ProfitAndLossSheetReport, error) {
	// mock := MockProfitAndLossSheetReport(holdingCode, accountGroup, startDate, endDate)
	// return mock, nil
	details, err := svc.repoPg.GetDataProfitAndLoss(holdingCode, accountGroup, includeCloseAccountMode, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// var totalData = len(details)
	//fmt.Printf("rows: %v", rows)
	//fmt.Printf("details: %+v\n", details)
	var incomeAmount float64 = 0
	var expenseAmount float64 = 0
	var profitAndLossAmount float64 = 0

	var incomes []models.ProfitAndLossSheetAccountDetail
	var expenses []models.ProfitAndLossSheetAccountDetail

	for _, v := range details {
		if v.AccountCategory == 4 {
			incomes = append(incomes, v)
			incomeAmount += v.Amount
		} else {
			expenses = append(expenses, v)
			expenseAmount += v.Amount
		}
	}

	profitAndLossAmount = incomeAmount - expenseAmount

	result := &models.ProfitAndLossSheetReport{
		ReportDate:          time.Now(),
		StartDate:           startDate,
		EndDate:             endDate,
		AccountGroup:        accountGroup,
		Incomes:             &incomes,
		Expenses:            &expenses,
		TotalIncomeAmount:   incomeAmount,
		TotalExpenseAmount:  expenseAmount,
		ProfitAndLossAmount: profitAndLossAmount,
	}

	return result, nil
}

func (svc JournalReportService) ProcessBalanceSheetReport(holdingCode string, accountGroup string, includeCloseAccountMode bool, endDate time.Time) (*models.BalanceSheetReport, error) {
	// mock := MockBalanceSheetReport(holdingCode, accountGroup, endDate)
	// return mock, nil
	details, err := svc.repoPg.GetDataBalanceSheet(holdingCode, accountGroup, includeCloseAccountMode, endDate)
	if err != nil {
		return nil, err
	}

	// var totalData = len(details)
	fmt.Printf("rows: %v", len(details))
	//fmt.Printf("details: %+v\n", details)
	var totalAssetAmount float64 = 0
	var totalLiabilityAmount float64 = 0
	var totalOwnersEquityAmount float64 = 0
	var totalLiabilityAndOwnersEquityAmount float64 = 0

	var totalIncome float64 = 0
	var totalExpense float64 = 0
	var totalProfitAndLoss float64 = 0

	var assets []models.BalanceSheetAccountDetail
	var liabilities []models.BalanceSheetAccountDetail
	var ownesEquities []models.BalanceSheetAccountDetail

	for _, v := range details {

		// fmt.Printf("%+v\n", v)

		if v.AccountCategory <= 3 {
			if v.AccountCategory == 1 {
				assets = append(assets, v)
				totalAssetAmount += v.Amount
			} else if v.AccountCategory == 2 {
				liabilities = append(liabilities, v)
				totalLiabilityAmount += v.Amount
			} else {
				totalOwnersEquityAmount += v.Amount
				ownesEquities = append(ownesEquities, v)
			}
		} else {
			if v.AccountCategory == 4 {
				totalIncome += v.Amount
			} else {
				totalExpense += v.Amount
			}
		}
	}

	totalProfitAndLoss = totalIncome - totalExpense
	if totalProfitAndLoss != 0 {
		totalOwnersEquityAmount += totalProfitAndLoss
		ownesEquities = append(ownesEquities, models.BalanceSheetAccountDetail{
			ChartOfAccountPG: chartofaccountModel.ChartOfAccountPG{
				AccountName:     "กำไร (ขาดทุน) สุทธิ",
				AccountCategory: 3,
			},
			Amount: totalProfitAndLoss,
		})
	}
	totalLiabilityAndOwnersEquityAmount = totalLiabilityAmount + totalOwnersEquityAmount

	result := &models.BalanceSheetReport{
		ReportDate:                          time.Now(),
		EndDate:                             endDate,
		AccountGroup:                        accountGroup,
		Assets:                              &assets,
		Liabilities:                         &liabilities,
		OwnesEquities:                       &ownesEquities,
		TotalAssetAmount:                    totalAssetAmount,
		TotalLiabilityAmount:                totalLiabilityAmount,
		TotalOwnersEquityAmount:             totalOwnersEquityAmount,
		TotalLiabilityAndOwnersEquityAmount: totalLiabilityAndOwnersEquityAmount,
	}

	return result, nil
}

func (svc JournalReportService) ProcessLedgerAccount(holdingCode string, accountGroup string, creditorCode string, debtorCode string, consolidateAccountCode string, accountRanges []models.LedgerAccountCodeRange, bookCode string, startDate time.Time, endDate time.Time) ([]models.LedgerAccount, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	rawDocList, err := svc.repoPg.GetDataLedgerAccount(holdingCode, accountGroup, creditorCode, debtorCode, consolidateAccountCode, accountRanges, bookCode, startDate, endDate)

	if err != nil {
		return nil, err
	}

	docList := []models.LedgerAccount{}

	lastAccountCode := ""
	lastAmount := decimal.NewFromFloat(0.0)
	tempDoc := models.LedgerAccount{}

	docNoList := map[string]struct{}{}

	currentIndexAccount := -1
	for _, doc := range rawDocList {

		if lastAccountCode != doc.AccountCode && doc.RowMode == -1 {
			currentIndexAccount++
			tempDoc = models.LedgerAccount{}
			tempDoc.Details = &[]models.LedgerAccountDetail{}
			tempDoc.AccountCode = doc.AccountCode
			tempDoc.AccountName = doc.AccountName
			tempDoc.AccountGroup = doc.AccountGroup
			tempDoc.ConsolidateAccountCode = doc.ConsolidateAccountCode

			lastAmount = decimal.NewFromFloat(doc.Amount)
			tempDoc.Balance, _ = lastAmount.Float64()
			tempDoc.NextBalance, _ = lastAmount.Float64()

			docList = append(docList, tempDoc)
		}

		if doc.RowMode == 0 && currentIndexAccount != -1 {
			debDecimal := decimal.NewFromFloat(doc.DebitAmount)
			credDecimal := decimal.NewFromFloat(doc.CreditAmount)

			lastAmount = lastAmount.Add(debDecimal).Sub(credDecimal)
			tempLastAmount, _ := lastAmount.Float64()

			docList[currentIndexAccount].NextBalance = tempLastAmount

			detail := models.LedgerAccountDetail{
				DocNo:              doc.DocNo,
				AccountDescription: doc.AccountDescription,
				DocDate:            doc.DocDate,
				Debit:              doc.DebitAmount,
				Credit:             doc.CreditAmount,
				Amount:             tempLastAmount,
			}
			*tempDoc.Details = append(*tempDoc.Details, detail)

			docNoList[doc.DocNo] = struct{}{}
		}

		lastAccountCode = doc.AccountCode
	}

	if len(docNoList) > 0 {
		tempDocNoList := []string{}
		for k := range docNoList {
			tempDocNoList = append(tempDocNoList, k)
		}

		journalSummaryList, err := svc.repoMongo.FindCountDetailByDocs(ctx, holdingCode, tempDocNoList)

		if err != nil {
			return nil, err
		}

		tempMapJournalSummary := map[string]models.JournalSummary{}

		for _, v := range journalSummaryList {
			tempMapJournalSummary[v.DocNo] = v
		}

		for _, doc := range docList {
			for i, detail := range *doc.Details {
				if v, ok := tempMapJournalSummary[detail.DocNo]; ok {
					(*doc.Details)[i].CountVat = v.CountVat
					(*doc.Details)[i].CountTax = v.CountTax
				}
			}
		}

		journalImageSummaryList, err := svc.repoMongo.FindCountImageByDocs(ctx, holdingCode, tempDocNoList)

		if err != nil {
			return nil, err
		}

		tempMapJournalImageSummary := map[string]models.JournalImageSummary{}

		for _, v := range journalImageSummaryList {
			tempMapJournalImageSummary[v.DocNo] = v
		}

		for _, doc := range docList {
			for i, detail := range *doc.Details {
				if v, ok := tempMapJournalImageSummary[detail.DocNo]; ok {
					(*doc.Details)[i].CountImage = v.CountImage
				}
			}
		}

	}

	return docList, nil
}

func (svc JournalReportService) ProcessMultiShopDashboard(
	username string,
	holdingCodes []string,
	startDate time.Time,
	endDate time.Time,
) (*models.MultiShopDashboardResponse, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	// 1. Get user's accessible shops
	accessibleShops, err := svc.shopUserRepo.FindByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user shops: %w", err)
	}

	// 2. Filter target shops
	targetHoldingCodes := svc.filterAccessibleShops(accessibleShops, holdingCodes)

	if len(targetHoldingCodes) == 0 {
		return &models.MultiShopDashboardResponse{
			Success: true,
			Data:    []models.ShopDashboardSummary{},
			Period:  svc.calculatePeriodInfo(startDate, endDate),
		}, nil
	}

	// 3. Get shop names
	shopDetails := svc.getShopDetails(ctx, targetHoldingCodes)

	// 4. Get revenue/expense data (PostgreSQL)
	revenueData, err := svc.repoPg.GetMultiShopRevenue(targetHoldingCodes, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get revenue data: %w", err)
	}

	// 5. Get image counts (MongoDB) - count all images per shop
	imageCounts, err := svc.repoMongo.CountAllImagesByShops(ctx, targetHoldingCodes)
	if err != nil {
		return nil, fmt.Errorf("failed to get image counts: %w", err)
	}

	// 6. Aggregate and calculate
	summaries := svc.aggregateShopData(
		targetHoldingCodes,
		shopDetails,
		revenueData,
		imageCounts,
		startDate,
		endDate,
	)

	return &models.MultiShopDashboardResponse{
		Success: true,
		Data:    summaries,
		Period:  svc.calculatePeriodInfo(startDate, endDate),
	}, nil
}

// Helper: Filter shops by user permission
func (svc JournalReportService) filterAccessibleShops(
	accessibleShops *[]authModels.ShopUser,
	filterHoldingCodes []string,
) []string {

	accessMap := make(map[string]bool)
	for _, shop := range *accessibleShops {
		accessMap[shop.HoldingCode] = true
	}

	if len(filterHoldingCodes) > 0 {
		result := []string{}
		for _, sid := range filterHoldingCodes {
			if accessMap[sid] {
				result = append(result, sid)
			}
		}
		return result
	}

	result := []string{}
	for holdingCode := range accessMap {
		result = append(result, holdingCode)
	}
	return result
}

// Helper: Get shop details (names)
func (svc JournalReportService) getShopDetails(
	ctx context.Context,
	holdingCodes []string,
) map[string]string {

	shopDetails := make(map[string]string)
	for _, holdingCode := range holdingCodes {
		shop, err := svc.shopRepo.FindByGuid(ctx, holdingCode)
		if err == nil {
			shopDetails[holdingCode] = shop.Name1
		} else {
			shopDetails[holdingCode] = holdingCode
		}
	}
	return shopDetails
}

// Helper: Aggregate all data
func (svc JournalReportService) aggregateShopData(
	holdingCodes []string,
	shopDetails map[string]string,
	revenueData []models.MultiShopRevenueRaw,
	imageCounts []models.ShopImageCount,
	startDate time.Time,
	endDate time.Time,
) []models.ShopDashboardSummary {

	shopRevenue := make(map[string]float64)
	shopExpense := make(map[string]float64)
	shopImages := make(map[string]int)

	for _, data := range revenueData {
		if data.AccountCategory == 4 {
			shopRevenue[data.HoldingCode] = data.TotalAmount
		} else if data.AccountCategory == 5 {
			shopExpense[data.HoldingCode] = data.TotalAmount
		}
	}

	for _, img := range imageCounts {
		shopImages[img.HoldingCode] = img.ImageCount
	}

	totalDays := svc.calculateTotalDays(startDate, endDate)
	totalMonths := svc.calculateMonthsBetween(startDate, endDate)
	totalYears := svc.calculateYearsBetween(startDate, endDate)

	results := []models.ShopDashboardSummary{}
	for _, holdingCode := range holdingCodes {
		revenue := shopRevenue[holdingCode]
		expense := shopExpense[holdingCode]
		profit := revenue - expense

		summary := models.ShopDashboardSummary{
			HoldingCode:    holdingCode,
			ShopName:       shopDetails[holdingCode],
			TotalRevenue:   revenue,
			TotalProfit:    profit,
			ImageCount:     shopImages[holdingCode],
			DailyAverage:   revenue / float64(totalDays),
			MonthlyAverage: revenue / float64(totalMonths),
			YearlyAverage:  revenue / float64(totalYears),
		}

		results = append(results, summary)
	}

	return results
}

// Helper: Calculate period info
func (svc JournalReportService) calculatePeriodInfo(startDate, endDate time.Time) models.PeriodInfo {
	totalDays := int(endDate.Sub(startDate).Hours()/24) + 1
	return models.PeriodInfo{
		StartDate: startDate,
		EndDate:   endDate,
		TotalDays: totalDays,
	}
}

func (svc JournalReportService) calculateTotalDays(start, end time.Time) int {
	days := int(end.Sub(start).Hours()/24) + 1
	if days <= 0 {
		return 1
	}
	return days
}

func (svc JournalReportService) calculateMonthsBetween(start, end time.Time) int {
	months := (end.Year()-start.Year())*12 + int(end.Month()-start.Month()) + 1
	if months <= 0 {
		return 1
	}
	return months
}

func (svc JournalReportService) calculateYearsBetween(start, end time.Time) int {
	years := end.Year() - start.Year() + 1
	if years <= 0 {
		return 1
	}
	return years
}
