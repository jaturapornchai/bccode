package smspaymentsettings

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	common "smlcloudplatform/internal/models"
	smspatternsrepo "smlcloudplatform/internal/smsreceive/smspatterns/repositories"
	"smlcloudplatform/internal/smsreceive/smspaymentsettings/models"
	"smlcloudplatform/internal/smsreceive/smspaymentsettings/repositories"
	"smlcloudplatform/internal/smsreceive/smspaymentsettings/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
)

type ISmsPaymentSettingsHttp interface{}

type SmsPaymentSettingsHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.ISmsPaymentSettingsHttpService
}

func NewSmsPaymentSettingsHttp(ms *microservice.Microservice, cfg config.IConfig) SmsPaymentSettingsHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())

	repo := repositories.NewSmsPaymentSettingsRepository(pst)

	smspatternRepo := smspatternsrepo.NewSmsPatternsRepository(pst)

	svc := services.NewSmsPaymentSettingsHttpService(repo, smspatternRepo)

	return SmsPaymentSettingsHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h SmsPaymentSettingsHttp) RegisterHttp() {

	h.ms.PUT("/smspaymentsettings/:storefrontguid", h.CreateSmsPaymentSettings)
	h.ms.GET("/smspaymentsettings/:storefrontguid", h.InfoSmsPaymentSettings)
	h.ms.GET("/smspaymentsettings", h.SearchSmsPaymentSettings)
}

// Save SMS Payment Settings godoc
// @Summary		sms payment settings
// @Description sms payment received settings service
// @Tags		SmsPayment
// @Param		storefrontguid  path      string  true  "storefront guidfixed"
// @Param		SmsPaymentSettings  body      models.SmsPaymentSettings  true  "sms payment settings"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /smspaymentsettings/{storefrontguid} [put]
func (h SmsPaymentSettingsHttp) CreateSmsPaymentSettings(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	storefrontGUID := ctx.Param("storefrontguid")

	docReq := &models.SmsPaymentSettings{}
	err := json.Unmarshal([]byte(input), &docReq)

	docReq.StorefrontGUID = storefrontGUID

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.SaveSmsPaymentSettings(shopID, authUsername, storefrontGUID, *docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
	})
	return nil
}

// Get SMS Payment Settings godoc
// @Summary		sms payment settings
// @Description sms payment received settings service
// @Tags		SmsPayment
// @Param		storefrontguid  path      string  true  "storefront guidfixed"
// @Accept 		json
// @Success		200	{object}	models.JournalInfoResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /smspaymentsettings/{storefrontguid} [get]
func (h SmsPaymentSettingsHttp) InfoSmsPaymentSettings(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	storefrontGUID := ctx.Param("storefrontguid")

	doc, err := h.svc.InfoSmsPaymentSettings(shopID, storefrontGUID)

	if err != nil {
		h.ms.Logger.Errorf("Error getting document %v", err)
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    doc,
	})
	return nil
}

// List SMS Payment Settings godoc
// @Summary		sms payment settings
// @Description sms payment received settings service
// @Tags		SmsPayment
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Size"
// @Accept 		json
// @Success		200	{object}	models.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /smspaymentsettings [get]
func (h SmsPaymentSettingsHttp) SearchSmsPaymentSettings(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)
	docList, pagination, err := h.svc.SearchSmsPaymentSettings(shopID, pageable)

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
