package qrpayment

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/payment/qrpayment/models"
	"smlcloudplatform/internal/payment/qrpayment/repositories"
	"smlcloudplatform/internal/payment/qrpayment/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
)

type IQrPaymentHttp interface{}

type QrPaymentHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IQrPaymentHttpService
}

func NewQrPaymentHttp(ms *microservice.Microservice, cfg config.IConfig) QrPaymentHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())

	repo := repositories.NewQrPaymentRepository(pst)

	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewQrPaymentHttpService(repo, masterSyncCacheRepo)

	return QrPaymentHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h QrPaymentHttp) RegisterHttp() {

	h.ms.POST("/payment/qrpayment/bulk", h.SaveBulk)

	h.ms.GET("/payment/qrpayment", h.SearchQrPaymentPage)
	h.ms.GET("/payment/qrpayment/list", h.SearchQrPaymentLimit)
	h.ms.POST("/payment/qrpayment", h.CreateQrPayment)
	h.ms.GET("/payment/qrpayment/:id", h.InfoQrPayment)
	h.ms.PUT("/payment/qrpayment/:id", h.UpdateQrPayment)
	h.ms.DELETE("/payment/qrpayment/:id", h.DeleteQrPayment)
	h.ms.DELETE("/payment/qrpayment", h.DeleteQrPaymentByGUIDs)
}

// Create QrPayment godoc
// @Description Create QrPayment
// @Tags		QrPayment
// @Param		QrPayment  body      models.QrPayment  true  "QrPayment"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /payment/qrpayment [post]
func (h QrPaymentHttp) CreateQrPayment(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := &models.QrPayment{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.CreateQrPayment(shopID, authUsername, *docReq)

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

// Update QrPayment godoc
// @Description Update QrPayment
// @Tags		QrPayment
// @Param		id  path      string  true  "QrPayment ID"
// @Param		QrPayment  body      models.QrPayment  true  "QrPayment"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /payment/qrpayment/{id} [put]
func (h QrPaymentHttp) UpdateQrPayment(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.QrPayment{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateQrPayment(shopID, id, authUsername, *docReq)

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

// Delete QrPayment godoc
// @Description Delete QrPayment
// @Tags		QrPayment
// @Param		id  path      string  true  "QrPayment ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /payment/qrpayment/{id} [delete]
func (h QrPaymentHttp) DeleteQrPayment(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteQrPayment(shopID, id, authUsername)

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

// Delete QrPayment godoc
// @Description Delete QrPayment
// @Tags		QrPayment
// @Param		QrPayment  body      []string  true  "QrPayment GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /payment/qrpayment [delete]
func (h QrPaymentHttp) DeleteQrPaymentByGUIDs(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	input := ctx.ReadInput()

	docReq := []string{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.DeleteQrPaymentByGUIDs(shopID, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get QrPayment godoc
// @Description get struct array by ID
// @Tags		QrPayment
// @Param		id  path      string  true  "QrPayment ID"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /payment/qrpayment/{id} [get]
func (h QrPaymentHttp) InfoQrPayment(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get QrPayment %v", id)
	doc, err := h.svc.InfoQrPayment(shopID, id)

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

// List QrPayment godoc
// @Description get struct array by ID
// @Tags		QrPayment
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /payment/qrpayment [get]
func (h QrPaymentHttp) SearchQrPaymentPage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)

	docList, pagination, err := h.svc.SearchQrPayment(shopID, pageable)

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

// List QrPayment godoc
// @Description search limit offset
// @Tags		QrPayment
// @Param		q		query	string		false  "Search Value"
// @Param		offset	query	integer		false  "offset"
// @Param		limit	query	integer		false  "limit"
// @Param		lang	query	string		false  "lang"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /payment/qrpayment/list [get]
func (h QrPaymentHttp) SearchQrPaymentLimit(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageableStep := utils.GetPageableStep(ctx.QueryParam)

	lang := ctx.QueryParam("lang")

	docList, total, err := h.svc.SearchQrPaymentStep(shopID, lang, pageableStep)

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

// Create QrPayment Bulk godoc
// @Description Create QrPayment
// @Tags		QrPayment
// @Param		QrPayment  body      []models.QrPayment  true  "QrPayment"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /payment/qrpayment/bulk [post]
func (h QrPaymentHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	input := ctx.ReadInput()

	dataReq := []models.QrPayment{}
	err := json.Unmarshal([]byte(input), &dataReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	bulkResponse, err := h.svc.SaveInBatch(shopID, authUsername, dataReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	ctx.Response(
		http.StatusCreated,
		common.BulkResponse{
			Success:    true,
			BulkImport: bulkResponse,
		},
	)

	return nil
}
