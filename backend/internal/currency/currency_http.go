package currency

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	"smlcloudplatform/internal/currency/models"
	"smlcloudplatform/internal/currency/repositories"
	"smlcloudplatform/internal/currency/services"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/requestfilter"
	"smlcloudplatform/pkg/microservice"
	"time"
)

type ICurrencyHttp interface{}

type CurrencyHttp struct {
	ms                     *microservice.Microservice
	cfg                    config.IConfig
	currencySvc            services.ICurrencyHttpService
	exchangeRateHistorySvc services.IExchangeRateHistorySeparateService // ใช้ separate collection
}

func NewCurrencyHttp(ms *microservice.Microservice, cfg config.IConfig) CurrencyHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())

	currencyRepo := repositories.NewCurrencyRepository(pst)
	exchangeRateHistoryRepo := repositories.NewExchangeRateHistoryRepository(pst) // separate collection

	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	currencySvc := services.NewCurrencyHttpService(currencyRepo, masterSyncCacheRepo, 15*time.Second)
	exchangeRateHistorySvc := services.NewExchangeRateHistorySeparateService(exchangeRateHistoryRepo, masterSyncCacheRepo, 15*time.Second)

	return CurrencyHttp{
		ms:                     ms,
		cfg:                    cfg,
		currencySvc:            currencySvc,
		exchangeRateHistorySvc: exchangeRateHistorySvc,
	}
}

func (h CurrencyHttp) RegisterHttp() {
	// Currency endpoints
	h.ms.GET("/currency", h.SearchCurrencyPage)
	h.ms.GET("/currency/list", h.SearchCurrencyStep)
	h.ms.POST("/currency", h.CreateCurrency)
	h.ms.GET("/currency/:id", h.InfoCurrency)
	h.ms.PUT("/currency/:id", h.UpdateCurrency)
	h.ms.DELETE("/currency/:id", h.DeleteCurrency)
	h.ms.DELETE("/currency", h.DeleteCurrencyByGUIDs)

	// Exchange Rate History endpoints
	h.ms.GET("/exchange-rate-history", h.SearchExchangeRateHistoryPage)
	h.ms.GET("/exchange-rate-history/list", h.SearchExchangeRateHistoryStep)
	h.ms.POST("/exchange-rate-history", h.CreateExchangeRateHistory)
	h.ms.GET("/exchange-rate-history/:id", h.InfoExchangeRateHistory)
	h.ms.PUT("/exchange-rate-history/:id", h.UpdateExchangeRateHistory)
	h.ms.DELETE("/exchange-rate-history/:id", h.DeleteExchangeRateHistory)
	h.ms.DELETE("/exchange-rate-history", h.DeleteExchangeRateHistoryByGUIDs)

	// Get latest exchange rate (for PO screen)
	h.ms.GET("/exchange-rate-history/latest", h.GetLatestExchangeRate)
}

// ============ Currency CRUD ============

// Create Currency godoc
// @Description Create Currency
// @Tags		Currency
// @Param		Currency  body      models.Currency  true  "Currency"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /currency [post]
func (h CurrencyHttp) CreateCurrency(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	holdingCode := ctx.UserInfo().HoldingCode
	input := ctx.ReadInput()

	docReq := &models.Currency{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.currencySvc.CreateCurrency(holdingCode, authUsername, *docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
		ID:      idx,
	})
	return nil
}

// Update Currency godoc
// @Description Update Currency
// @Tags		Currency
// @Param		id  path      string  true  "Currency ID"
// @Param		Currency  body      models.Currency  true  "Currency"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /currency/{id} [put]
func (h CurrencyHttp) UpdateCurrency(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.Currency{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.currencySvc.UpdateCurrency(holdingCode, id, authUsername, *docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
		ID:      id,
	})

	return nil
}

// Delete Currency godoc
// @Description Delete Currency
// @Tags		Currency
// @Param		id  path      string  true  "Currency ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /currency/{id} [delete]
func (h CurrencyHttp) DeleteCurrency(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.currencySvc.DeleteCurrency(holdingCode, id, authUsername)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		ID:      id,
	})

	return nil
}

// Delete Currency godoc
// @Description Delete Currency
// @Tags		Currency
// @Param		Currency  body      []string  true  "Currency GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /currency [delete]
func (h CurrencyHttp) DeleteCurrencyByGUIDs(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username

	input := ctx.ReadInput()

	docReq := []string{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.currencySvc.DeleteCurrencyByGUIDs(holdingCode, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get Currency godoc
// @Description get Currency info by guidfixed
// @Tags		Currency
// @Param		id  path      string  true  "Currency guidfixed"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /currency/{id} [get]
func (h CurrencyHttp) InfoCurrency(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get Currency %v", id)
	doc, err := h.currencySvc.InfoCurrency(holdingCode, id)

	if err != nil {
		h.ms.Logger.Errorf("Error getting document %v: %v", id, err)
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    doc,
	})
	return nil
}

// List Currency page godoc
// @Description get list page
// @Tags		Currency
// @Param		disabled		query	boolean		false  "disabled"
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /currency [get]
func (h CurrencyHttp) SearchCurrencyPage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	pageable := utils.GetPageable(ctx.QueryParam)

	filters := requestfilter.GenerateFilters(ctx.QueryParam, []requestfilter.FilterRequest{
		{
			Param: "disabled",
			Field: "isdisabled",
			Type:  requestfilter.FieldTypeBoolean,
		},
	})

	docList, pagination, err := h.currencySvc.SearchCurrency(holdingCode, filters, pageable)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success:    true,
		Data:       docList,
		Pagination: pagination,
	})
	return nil
}

// List Currency step godoc
// @Description search limit offset
// @Tags		Currency
// @Param		disabled		query	boolean		false  "disabled"
// @Param		q		query	string		false  "Search Value"
// @Param		offset	query	integer		false  "offset"
// @Param		limit	query	integer		false  "limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /currency/list [get]
func (h CurrencyHttp) SearchCurrencyStep(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	pageableStep := utils.GetPageableStep(ctx.QueryParam)

	filters := requestfilter.GenerateFilters(ctx.QueryParam, []requestfilter.FilterRequest{
		{
			Param: "isdisabled",
			Field: "isdisabled",
			Type:  requestfilter.FieldTypeBoolean,
		},
	})

	docList, total, err := h.currencySvc.SearchCurrencyStep(holdingCode, filters, pageableStep)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    docList,
		Total:   total,
	})
	return nil
}

// ============ Exchange Rate History CRUD ============

// ExchangeRateRequest - Request for creating/updating exchange rate
type ExchangeRateRequest struct {
	Currency string  `json:"currency" validate:"required"`  // USD, EUR, JPY
	Date     string  `json:"date" validate:"required"`      // YYYY-MM-DD
	Rate     float64 `json:"rate" validate:"required,gt=0"` // Exchange rate to THB
}

// Create ExchangeRateHistory godoc
// @Description Create ExchangeRateHistory
// @Tags		ExchangeRateHistory
// @Param		ExchangeRateHistory  body      ExchangeRateRequest  true  "ExchangeRateHistory"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /exchange-rate-history [post]
func (h CurrencyHttp) CreateExchangeRateHistory(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	holdingCode := ctx.UserInfo().HoldingCode
	input := ctx.ReadInput()

	docReq := &ExchangeRateRequest{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.exchangeRateHistorySvc.CreateExchangeRateHistory(holdingCode, authUsername, docReq.Currency, docReq.Date, docReq.Rate)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
		ID:      idx,
	})
	return nil
}

// Update ExchangeRateHistory godoc
// @Description Update ExchangeRateHistory
// @Tags		ExchangeRateHistory
// @Param		id  path      string  true  "ExchangeRateHistory ID"
// @Param		ExchangeRateHistory  body      ExchangeRateRequest  true  "ExchangeRateHistory"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /exchange-rate-history/{id} [put]
func (h CurrencyHttp) UpdateExchangeRateHistory(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &ExchangeRateRequest{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.exchangeRateHistorySvc.UpdateExchangeRateHistory(holdingCode, id, authUsername, docReq.Date, docReq.Rate)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
		ID:      id,
	})

	return nil
}

// Delete ExchangeRateHistory godoc
// @Description Delete ExchangeRateHistory
// @Tags		ExchangeRateHistory
// @Param		id  path      string  true  "ExchangeRateHistory ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /exchange-rate-history/{id} [delete]
func (h CurrencyHttp) DeleteExchangeRateHistory(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.exchangeRateHistorySvc.DeleteExchangeRateHistory(holdingCode, id, authUsername)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		ID:      id,
	})

	return nil
}

// Delete ExchangeRateHistory godoc
// @Description Delete ExchangeRateHistory
// @Tags		ExchangeRateHistory
// @Param		ExchangeRateHistory  body      []string  true  "ExchangeRateHistory GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /exchange-rate-history [delete]
func (h CurrencyHttp) DeleteExchangeRateHistoryByGUIDs(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username

	input := ctx.ReadInput()

	docReq := []string{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.exchangeRateHistorySvc.DeleteExchangeRateHistoryByGUIDs(holdingCode, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get ExchangeRateHistory godoc
// @Description get ExchangeRateHistory info by guidfixed
// @Tags		ExchangeRateHistory
// @Param		id  path      string  true  "ExchangeRateHistory guidfixed"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /exchange-rate-history/{id} [get]
func (h CurrencyHttp) InfoExchangeRateHistory(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get ExchangeRateHistory %v", id)
	doc, err := h.exchangeRateHistorySvc.InfoExchangeRateHistory(holdingCode, id)

	if err != nil {
		h.ms.Logger.Errorf("Error getting document %v: %v", id, err)
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    doc,
	})
	return nil
}

// List ExchangeRateHistory page godoc
// @Description get list page
// @Tags		ExchangeRateHistory
// @Param		currency		query	string		false  "Currency code"
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /exchange-rate-history [get]
func (h CurrencyHttp) SearchExchangeRateHistoryPage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	pageable := utils.GetPageable(ctx.QueryParam)

	filters := requestfilter.GenerateFilters(ctx.QueryParam, []requestfilter.FilterRequest{
		{
			Param: "currency",
			Field: "currency",
			Type:  requestfilter.FieldTypeString,
		},
	})

	docList, pagination, err := h.exchangeRateHistorySvc.SearchExchangeRateHistory(holdingCode, filters, pageable)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success:    true,
		Data:       docList,
		Pagination: pagination,
	})
	return nil
}

// List ExchangeRateHistory step godoc
// @Description search limit offset
// @Tags		ExchangeRateHistory
// @Param		currency		query	string		false  "Currency code"
// @Param		q		query	string		false  "Search Value"
// @Param		offset	query	integer		false  "offset"
// @Param		limit	query	integer		false  "limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /exchange-rate-history/list [get]
func (h CurrencyHttp) SearchExchangeRateHistoryStep(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	pageableStep := utils.GetPageableStep(ctx.QueryParam)

	filters := requestfilter.GenerateFilters(ctx.QueryParam, []requestfilter.FilterRequest{
		{
			Param: "currency",
			Field: "currency",
			Type:  requestfilter.FieldTypeString,
		},
	})

	docList, total, err := h.exchangeRateHistorySvc.SearchExchangeRateHistoryStep(holdingCode, filters, pageableStep)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    docList,
		Total:   total,
	})
	return nil
}

// Get Latest Exchange Rate godoc
// @Description Get latest exchange rate for a currency on or before a specific date
// @Tags		ExchangeRateHistory
// @Param		currency		query	string		true  "Currency code (e.g. USD, EUR, JPY)"
// @Param		date		query	string		true  "Date in YYYY-MM-DD format"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /exchange-rate-history/latest [get]
func (h CurrencyHttp) GetLatestExchangeRate(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	currency := ctx.QueryParam("currency")
	date := ctx.QueryParam("date")

	if currency == "" || date == "" {
		ctx.ResponseError(http.StatusBadRequest, "currency and date are required")
		return nil
	}

	h.ms.Logger.Debugf("Get latest exchange rate for %s on %s", currency, date)
	doc, err := h.exchangeRateHistorySvc.GetLatestExchangeRate(holdingCode, currency, date)

	if err != nil {
		h.ms.Logger.Errorf("Error getting exchange rate: %v", err)
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    doc,
	})
	return nil
}
