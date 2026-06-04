package smstransaction

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	common "smlcloudplatform/internal/models"
	smspatternsRepo "smlcloudplatform/internal/smsreceive/smspatterns/repositories"
	smssetingsRepo "smlcloudplatform/internal/smsreceive/smspaymentsettings/repositories"
	"smlcloudplatform/internal/smsreceive/smstransaction/models"
	"smlcloudplatform/internal/smsreceive/smstransaction/repositories"
	"smlcloudplatform/internal/smsreceive/smstransaction/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
)

type ISmsTransactionHttp interface{}

type SmsTransactionHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.ISmsTransactionHttpService
}

func NewSmsTransactionHttp(ms *microservice.Microservice, cfg config.IConfig) SmsTransactionHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())

	repo := repositories.NewSmsTransactionRepository(pst)
	smsPatternRepo := smspatternsRepo.NewSmsPatternsRepository(pst)
	smsSetingsRepo := smssetingsRepo.NewSmsPaymentSettingsRepository(pst)
	svc := services.NewSmsTransactionHttpService(repo, smsPatternRepo, smsSetingsRepo, utils.NewGUID, ms.TimeNow)

	return SmsTransactionHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h SmsTransactionHttp) RegisterHttp() {
	h.ms.GET("/smstransaction", h.SearchSmsTransaction)
	h.ms.POST("/smstransaction", h.CreateSmsTransaction)
	h.ms.GET("/smstransaction/:id", h.InfoSmsTransaction)
	h.ms.PUT("/smstransaction/:id", h.UpdateSmsTransaction)
	h.ms.DELETE("/smstransaction/:id", h.DeleteSmsTransaction)
	h.ms.POST("/smstransaction/checksms/:storefront", h.CheckSMS)
	h.ms.POST("/smstransaction/confirm/:id", h.ConfirmSmsTransaction)
}

// Create SMS Transaction godoc
// @Summary		รับข้อมูล sms
// @Description รับข้อมูล sms
// @Tags		SMS
// @Param		SMSTransaction  body      models.SmsTransaction  true  "sms data"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /smstransaction [post]
func (h SmsTransactionHttp) CreateSmsTransaction(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	holdingCode := ctx.UserInfo().HoldingCode
	input := ctx.ReadInput()

	docReq := &models.SmsTransaction{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.CreateSmsTransaction(holdingCode, authUsername, *docReq)

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

// Update SMS Transaction godoc
// @Summary		รับข้อมูล sms
// @Description รับข้อมูล sms
// @Tags		SMS
// @Param		id  path      string  true  "GIUDFIXED"
// @Param		Journal  body      models.SmsTransaction  true  "sms data"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /smstransaction/{id} [put]
func (h SmsTransactionHttp) UpdateSmsTransaction(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.SmsTransaction{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateSmsTransaction(holdingCode, id, authUsername, *docReq)

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

// Delete SMS Transaction godoc
// @Summary		รับข้อมูล sms
// @Description รับข้อมูล sms
// @Tags		SMS
// @Param		id  path      string  true  "GIUDFIXED"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /smstransaction/{id} [delete]
func (h SmsTransactionHttp) DeleteSmsTransaction(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteSmsTransaction(holdingCode, id, authUsername)

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

// GET SMS Transaction godoc
// @Summary		รับข้อมูล sms
// @Description รับข้อมูล sms
// @Tags		SMS
// @Param		id  path      string  true  "GIUDFIXED"
// @Accept 		json
// @Success		200	{object}	models.SmsTransactionInfoResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /smstransaction/{id} [get]
func (h SmsTransactionHttp) InfoSmsTransaction(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get SmsTransaction %v", id)
	doc, err := h.svc.InfoSmsTransaction(holdingCode, id)

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

// LIST SMS Transaction godoc
// @Summary		รับข้อมูล sms
// @Description รับข้อมูล sms
// @Tags		SMS
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Size"
// @Accept 		json
// @Success		200	{object}	models.SmsTransactionPageResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /smstransaction [get]
func (h SmsTransactionHttp) SearchSmsTransaction(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	pageable := utils.GetPageable(ctx.QueryParam)
	docList, pagination, err := h.svc.SearchSmsTransaction(holdingCode, pageable)

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

// Check SMS Transaction godoc
// @Summary		รับข้อมูล sms
// @Description รับข้อมูล sms
// @Tags		SMS
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Size"
// @Accept 		json
// @Success		200	{object}	models.SmsTransactionPageResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /checksms [get]
func (h SmsTransactionHttp) CheckSMS(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	storefrontCode := ctx.Param("storefront")

	input := ctx.ReadInput()

	docReq := &models.SmsTransactionAmount{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	// mockTime, _ := time.Parse(time.RFC3339, "2022-08-25T03:09:57.335+00:00")

	checkTime := h.ms.TimeNow()

	amount, err := h.svc.CheckSMS(holdingCode, storefrontCode, docReq.Amount, checkTime)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    amount,
	})
	return nil
}

// Check SMS Transaction godoc
// @Summary		รับข้อมูล sms
// @Description รับข้อมูล sms
// @Tags		SMS
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Size"
// @Accept 		json
// @Success		200	{object}	models.SmsTransactionPageResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /smstransaction/confirm/{id} [get]
func (h SmsTransactionHttp) ConfirmSmsTransaction(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	// mockTime, _ := time.Parse(time.RFC3339, "2022-08-25T03:09:57.335+00:00")

	// checkTime := h.ms.TimeNow()

	id := ctx.Param("id")

	err := h.svc.ConfirmSmsTransaction(holdingCode, id)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})
	return nil
}
