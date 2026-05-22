package journalreport

import (
	"fmt"
	"net/http"
	"smlcloudplatform/internal/config"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/shop"
	"smlcloudplatform/internal/vfgl/journalreport/models"
	"smlcloudplatform/pkg/microservice"
	"strings"
	"time"
)

type JournalReportHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc IJournalReportService
}

func NewJournalReportHttp(ms *microservice.Microservice, cfg config.IConfig) JournalReportHttp {

	pstConfig := config.NewPersisterConfig()
	pst := microservice.NewPersister(pstConfig)
	repoPg := NewJournalReportPgRepository(pst)

	pstMongo := microservice.NewPersisterMongo(config.NewMongoPersisterConfig())
	repoMongo := NewJournalMongoRepository(pstMongo)

	// เพิ่ม shop repositories
	shopUserRepo := shop.NewShopUserRepository(pstMongo)
	shopRepo := shop.NewShopRepository(pstMongo)

	jouralReportService := NewJournalReportService(
		repoPg,
		repoMongo,
		shopUserRepo,
		shopRepo,
	)
	return JournalReportHttp{
		ms:  ms,
		cfg: cfg,
		svc: jouralReportService,
	}
}

func (h JournalReportHttp) RegisterHttp() {
	h.ms.GET("/gl/report/trialbalancesheet", h.ProcessReportTrialBalanceSheet)
	h.ms.GET("/gl/report/balancesheet", h.ProcessBalanceSheetReport)
	h.ms.GET("/gl/report/profitandloss", h.ProcessProfitAndLossReport)
	h.ms.GET("/gl/report/ledgeraccount", h.ProcessReportLedgerAccount)
	h.ms.GET("/gl/dashboard/multi-shop-summary", h.ProcessMultiShopDashboard)
}

// Fetch TrialBalanceSheet godoc
// @Summary		ดึงข้อมูลงบทดลอง
// @Description ดึงข้อมูลงบทดลอง
// @Tags		GLReport
// @Param		accountgroup query string true "กลุ่มบัญชี"
// @Param		startdate query string true "จากวันที่ (Date Format: YYYY-MM-DD)"
// @Param		enddate query string true "ถึงวันที่ (Date Format: YYYY-MM-DD)"
// @Param		ica query int true "รวมรายการปิดปัญชี"
// @Param		timezone query string false "TimeZone"
// @Accept		json
// @Success		200 {object} models.TrialBalanceSheetReportResponse
// @Failure		400 {object} common.AuthResponseFailed
// @Failure		401 {object} common.AuthResponseFailed
// @Security	AccessToken
// @Router		/gl/report/trialbalancesheet [get]
func (r JournalReportHttp) ProcessReportTrialBalanceSheet(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	layout := "2006-01-02 -0700" //
	accountGroup := ctx.QueryParam("accountgroup")
	startDateStr := ctx.QueryParam("start_date")
	endDateStr := ctx.QueryParam("end_date")
	timeZone := ctx.QueryParam("timezone") // +07

	if timeZone != "" {
		// phase with timezone
		startDateStr = startDateStr + " " + timeZone + "00"
		endDateStr = endDateStr + " " + timeZone + "00"
	} else {
		startDateStr = startDateStr + " +0000"
		endDateStr = endDateStr + " +0000"
	}

	includeCloseAccountMode := ctx.QueryParam("ica") == "1"

	if len(startDateStr) < 1 || len(endDateStr) < 1 {
		ctx.ResponseError(400, "Invalid Payload.")
		return nil
	}

	startDate, err := time.Parse(layout, startDateStr)
	if err != nil {
		ctx.ResponseError(400, "startdate format invalid.")
		return err
	}

	endDate, err := time.Parse(layout, endDateStr)
	if err != nil {
		ctx.ResponseError(400, "enddate format invalid.")
		return err
	}
	endDate = endDate.AddDate(0, 0, 1).Add(time.Second * -1)

	r.ms.Logger.Debugf("Start Process TrialBalanceSheet %v:%v, includecloseaccount: %v", startDate, endDate, includeCloseAccountMode)
	reportData, err := r.svc.ProcessTrialBalanceSheetReport(shopID, accountGroup, includeCloseAccountMode, startDate.UTC(), endDate.UTC())
	if err != nil {
		ctx.ResponseError(500, fmt.Sprintf("Failed on Process Report : %v.", err.Error()))
		return err
	}

	ctx.Response(
		http.StatusOK,
		models.TrialBalanceSheetReportResponse{
			Success: true,
			Data:    *reportData,
		})
	return nil
}

// Fetch Data Balance Sheet godoc
// @Summary		ดึงข้อมูลงบดุล
// @Description ดึงข้อมูลงบดุล
// @Tags		GLReport
// @Param		accountgroup query string true "กลุ่มบัญชี"
// @Param		enddate query string true "ณ วันที่ (Date Format: YYYY-MM-DD)"
// @Param		ica query int true "รวมรายการปิดปัญชี"
// @Param		timezone query string false "TimeZone"
// @Accept		json
// @Success		200 {object} models.BalanceSheetReportResponse
// @Failure		400 {object} common.AuthResponseFailed
// @Failure		401 {object} common.AuthResponseFailed
// @Security	AccessToken
// @Router		/gl/report/balancesheet [get]
func (r JournalReportHttp) ProcessBalanceSheetReport(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	layout := "2006-01-02 -0700" //
	endDateStr := ctx.QueryParam("end_date")
	accountGroup := ctx.QueryParam("accountgroup")
	includeCloseAccountMode := ctx.QueryParam("ica") == "1"

	timeZone := ctx.QueryParam("timezone") // +07

	if timeZone != "" {
		// phase with timezone
		endDateStr = endDateStr + " " + timeZone + "00"
	} else {
		endDateStr = endDateStr + " +0000"
	}

	if len(endDateStr) < 1 {
		ctx.ResponseError(400, "enddate format invalid.")
		return nil
	}

	endDate, err := time.Parse(layout, endDateStr)
	if err != nil {
		ctx.ResponseError(400, "enddate format invalid.")
		return err
	}
	endDate = endDate.AddDate(0, 0, 1).Add(time.Second * -1)

	r.ms.Logger.Debugf("Start Process BalanceSheet at %v", endDate)
	reportData, err := r.svc.ProcessBalanceSheetReport(shopID, accountGroup, includeCloseAccountMode, endDate.UTC())
	if err != nil {
		ctx.ResponseError(500, fmt.Sprintf("Failed on Process Report : %v.", err.Error()))
		return err
	}

	ctx.Response(
		http.StatusOK,
		common.ApiResponse{
			Success: true,
			Data:    reportData,
		})
	return nil
}

// Fetch ProfitAndLossSheet godoc
// @Summary		ดึงข้อมูลงบกำไรขาดทุน
// @Description ดึงข้อมูลงบกำไรขาดทุน
// @Tags		GLReport
// @Param		accountgroup query string true "กลุ่มบัญชี"
// @Param		startdate query string true "จากวันที่ (Date Format: YYYY-MM-DD)"
// @Param		enddate query string true "ถึงวันที่ (Date Format: YYYY-MM-DD)"
// @Param		ica query int true "รวมรายการปิดปัญชี"
// @Param		timezone query string false "TimeZone"
// @Accept		json
// @Success		200 {object} models.LostAndProfitSheetReportResponse
// @Failure		400 {object} common.AuthResponseFailed
// @Failure		401 {object} common.AuthResponseFailed
// @Security	AccessToken
// @Router		/gl/report/profitandloss [get]
func (r JournalReportHttp) ProcessProfitAndLossReport(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	layout := "2006-01-02 -0700" //
	accountGroup := ctx.QueryParam("accountgroup")
	startDateStr := ctx.QueryParam("start_date")
	endDateStr := ctx.QueryParam("end_date")
	includeCloseAccountMode := ctx.QueryParam("ica") == "1"

	timeZone := ctx.QueryParam("timezone") // +07

	if timeZone != "" {
		// phase with timezone
		startDateStr = startDateStr + " " + timeZone + "00"
		endDateStr = endDateStr + " " + timeZone + "00"
	} else {
		startDateStr = startDateStr + " +0000"
		endDateStr = endDateStr + " +0000"
	}

	if len(startDateStr) < 1 || len(endDateStr) < 1 {
		ctx.ResponseError(400, "date format invalid.")
		return nil
	}

	startDate, err := time.Parse(layout, startDateStr)
	if err != nil {
		ctx.ResponseError(400, "startdate format invalid.")
		return err
	}

	endDate, err := time.Parse(layout, endDateStr)
	if err != nil {
		ctx.ResponseError(400, "enddate format invalid.")
		return err
	}
	endDate = endDate.AddDate(0, 0, 1).Add(time.Second * -1)

	r.ms.Logger.Debugf("Start Process ProfitAndLoss %v:%v", startDate, endDate)
	reportData, err := r.svc.ProcessProfitAndLossSheetReport(shopID, accountGroup, includeCloseAccountMode, startDate.UTC(), endDate.UTC())
	if err != nil {
		ctx.ResponseError(500, fmt.Sprintf("Failed on Process Report : %v.", err.Error()))
		return err
	}

	ctx.Response(
		http.StatusOK,
		common.ApiResponse{
			Success: true,
			Data:    reportData,
		})
	return nil
}

// Fetch Ledger Account godoc
// @Summary		บัญชีแยกประเภท
// @Description บัญชีแยกประเภท
// @Tags		GLReport
// @Param		startdate query string true "จากวันที่ (Date Format: YYYY-MM-DD)"
// @Param		enddate query string true "ถึงวันที่ (Date Format: YYYY-MM-DD)"
// @Param		accountcode query string true "Account Code"
// @Param		creditorcode query string true "Creditor Code"
// @Param		debtorcode query string true "Debtor Code"
// @Param		timezone query string false "TimeZone"
// @Accept		json
// @Success		200 {object} models.TrialBalanceSheetReportResponse
// @Failure		400 {object} common.AuthResponseFailed
// @Failure		401 {object} common.AuthResponseFailed
// @Security	AccessToken
// @Router		/gl/report/ledgeraccount [get]
func (r JournalReportHttp) ProcessReportLedgerAccount(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	layout := "2006-01-02 -0700" //
	startDateStr := ctx.QueryParam("start_date")
	endDateStr := ctx.QueryParam("end_date")
	timeZone := ctx.QueryParam("timezone") // +07
	bookCode := ctx.QueryParam("bookcode")

	accountGroup := ctx.QueryParam("accountgroup")
	creditorCode := ctx.QueryParam("creditorcode")
	debtorCode := ctx.QueryParam("debtorcode")
	consolidateAccountCode := ctx.QueryParam("consolidateaccountcode")

	if timeZone != "" {
		// phase with timezone
		startDateStr = startDateStr + " " + timeZone + "00"
		endDateStr = endDateStr + " " + timeZone + "00"
	} else {
		startDateStr = startDateStr + " +0000"
		endDateStr = endDateStr + " +0000"
	}

	startDate, err := time.Parse(layout, startDateStr)
	if err != nil {
		ctx.ResponseError(400, "startdate format invalid.")
		return err
	}

	endDate, err := time.Parse(layout, endDateStr)
	if err != nil {
		ctx.ResponseError(400, "enddate format invalid.")
		return err
	}
	endDate = endDate.AddDate(0, 0, 1).Add(time.Second * -1)

	accountCodeRangeRaw := strings.TrimSpace(ctx.QueryParam("accountcode"))

	accRanges := []models.LedgerAccountCodeRange{}
	if len(accountCodeRangeRaw) > 0 {
		accSplitRaw := strings.Split(accountCodeRangeRaw, ",")
		for _, accRaw := range accSplitRaw {
			if len(accRaw) > 0 {
				accRangeRaw := strings.Split(accRaw, ":")
				if len(accRangeRaw) == 2 {

					accRanges = append(accRanges, models.LedgerAccountCodeRange{
						Start: accRangeRaw[0],
						End:   accRangeRaw[1],
					})
				}
			}

		}

	}

	r.ms.Logger.Debugf("Start Process Ledger Account %v:%v", startDate, endDate)
	reportData, err := r.svc.ProcessLedgerAccount(shopID, accountGroup, creditorCode, debtorCode, consolidateAccountCode, accRanges, bookCode, startDate.UTC(), endDate.UTC())
	if err != nil {
		ctx.ResponseError(500, fmt.Sprintf("Failed on Process Report : %v.", err.Error()))
		return err
	}

	ctx.Response(
		http.StatusOK,
		common.ApiResponse{
			Success: true,
			Data:    reportData,
		})
	return nil
}

// ProcessMultiShopDashboard godoc
// @Summary      ดึงข้อมูล Dashboard หลาย Shops
// @Description  ดึงข้อมูลสรุปรายได้ กำไร และจำนวนรูปภาพสำหรับหลาย shops
// @Tags         GLDashboard
// @Param        startdate query string true "จากวันที่ (YYYY-MM-DD)"
// @Param        enddate query string true "ถึงวันที่ (YYYY-MM-DD)"
// @Param        timezone query string false "TimeZone (default: +00)"
// @Param        shopids query string false "Shop IDs comma-separated"
// @Accept       json
// @Success      200 {object} models.MultiShopDashboardResponse
// @Failure      400 {object} common.AuthResponseFailed
// @Failure      500 {object} common.AuthResponseFailed
// @Security     AccessToken
// @Router       /gl/dashboard/multi-shop-summary [get]
func (r JournalReportHttp) ProcessMultiShopDashboard(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	username := userInfo.Username

	layout := "2006-01-02 -0700"
	startDateStr := ctx.QueryParam("start_date")
	endDateStr := ctx.QueryParam("end_date")
	timeZone := ctx.QueryParam("timezone")
	shopIDsParam := ctx.QueryParam("shopids")

	// Validate
	if len(startDateStr) < 1 || len(endDateStr) < 1 {
		ctx.ResponseError(400, "startdate and enddate are required")
		return nil
	}

	// Apply timezone
	if timeZone != "" {
		startDateStr = startDateStr + " " + timeZone + "00"
		endDateStr = endDateStr + " " + timeZone + "00"
	} else {
		startDateStr = startDateStr + " +0000"
		endDateStr = endDateStr + " +0000"
	}

	// Parse dates
	startDate, err := time.Parse(layout, startDateStr)
	if err != nil {
		ctx.ResponseError(400, "startdate format invalid")
		return err
	}

	endDate, err := time.Parse(layout, endDateStr)
	if err != nil {
		ctx.ResponseError(400, "enddate format invalid")
		return err
	}
	endDate = endDate.AddDate(0, 0, 1).Add(time.Second * -1)

	// Parse shop IDs
	var shopIDs []string
	if len(shopIDsParam) > 0 {
		shopIDs = strings.Split(shopIDsParam, ",")
		for i, id := range shopIDs {
			shopIDs[i] = strings.TrimSpace(id)
		}
	}

	// Process
	r.ms.Logger.Debugf("Multi-Shop Dashboard: user=%s, period=%v to %v", username, startDate, endDate)

	result, err := r.svc.ProcessMultiShopDashboard(username, shopIDs, startDate.UTC(), endDate.UTC())
	if err != nil {
		ctx.ResponseError(500, fmt.Sprintf("Failed to process dashboard: %v", err.Error()))
		return err
	}

	ctx.Response(http.StatusOK, result)
	return nil
}
