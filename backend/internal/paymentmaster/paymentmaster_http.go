package paymentmaster

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/paymentmaster/models"
	"smlcloudplatform/internal/paymentmaster/repositories"
	"smlcloudplatform/internal/paymentmaster/services"
	"smlcloudplatform/pkg/microservice"
)

type IPaymentMasterHttp interface{}

type PaymentMasterHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IPaymentMasterHttpService
}

func NewPaymentMasterHttp(ms *microservice.Microservice, cfg config.IConfig) PaymentMasterHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())

	repo := repositories.NewPaymentMasterRepository(pst)
	svc := services.NewPaymentMasterHttpService(repo)

	return PaymentMasterHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h PaymentMasterHttp) RegisterHttp() {

	h.ms.POST("/paymentmaster/bulk", h.SaveBulk)

	h.ms.GET("/paymentmaster", h.SearchPaymentMaster)
	h.ms.POST("/paymentmaster", h.CreatePaymentMaster)
	h.ms.GET("/paymentmaster/:id", h.InfoPaymentMaster)
	h.ms.PUT("/paymentmaster/:id", h.UpdatePaymentMaster)
	h.ms.DELETE("/paymentmaster/:id", h.DeletePaymentMaster)

	h.ms.GET("/paymentmaster-type", h.InfoPaymentMasterType)
}

// Create Payment Master godoc
// @Summary		สร้าง payment master
// @Description สร้าง payment master
// @Tags		PaymentMaster
// @Param		PaymentMaster  body      models.PaymentMaster  true  "paymentmaster"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /paymentmaster [post]
func (h PaymentMasterHttp) CreatePaymentMaster(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := &models.PaymentMaster{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.CreatePaymentMaster(shopID, authUsername, *docReq)

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

// Update Payment Master godoc
// @Summary		แก้ไข payment master
// @Description แก้ไข payment master
// @Tags		PaymentMaster
// @Param		id  path      string  true  "Payment Master ID"
// @Param		PaymentMaster  body      models.PaymentMaster  true  "paymentmaster"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /paymentmaster/{id} [put]
func (h PaymentMasterHttp) UpdatePaymentMaster(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.PaymentMaster{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdatePaymentMaster(id, shopID, authUsername, *docReq)

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

// Delete Payment Master godoc
// @Summary		ลบ payment master
// @Description ลบ payment master
// @Tags		PaymentMaster
// @Param		id  path      string  true  "Journal ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /paymentmaster/{id} [delete]
func (h PaymentMasterHttp) DeletePaymentMaster(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeletePaymentMaster(id, shopID, authUsername)

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

// Get Payment Master godoc
// @Summary		แสดงรายละเอียด payment master
// @Description แสดงรายละเอียด payment master
// @Tags		PaymentMaster
// @Param		id  path      string  true  "Journal Id"
// @Accept 		json
// @Success		200	{object}	models.PaymentMasterInfoResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /paymentmaster/{id} [get]
func (h PaymentMasterHttp) InfoPaymentMaster(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get PaymentMaster %v", id)
	doc, err := h.svc.InfoPaymentMaster(id, shopID)

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

// List Payment Master godoc
// @Summary		แสดงรายการ payment master
// @Description แสดงรายการ payment master
// @Tags		PaymentMaster
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Size"
// @Accept 		json
// @Success		200	{object}	models.PaymentMasterPageResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /paymentmaster [get]
func (h PaymentMasterHttp) SearchPaymentMaster(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	q := ctx.QueryParam("q")
	docList, err := h.svc.SearchPaymentMaster(shopID, q)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    docList,
	})
	return nil
}

// Create Payment Master godoc
// @Summary		นำเข้าข้อมูล payment master
// @Description นำเข้าข้อมูล payment master
// @Tags		PaymentMaster
// @Param		Journal  body      []models.Journal  true  "paymentmaster"
// @Accept 		json
// @Success		201	{object}	common.BulkInsertResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /paymentmaster/bulk [post]
func (h PaymentMasterHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	input := ctx.ReadInput()

	dataReq := []models.PaymentMaster{}
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

// Get Payment Master Type godoc
// @Summary		แสดงรายละเอียด payment master type
// @Description แสดงรายละเอียด payment master type
// @Tags		PaymentMaster
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /paymentmaster-type [get]
func (h PaymentMasterHttp) InfoPaymentMasterType(ctx microservice.IContext) error {

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data: map[int8]string{
			0: "bank",
			1: "qr payment",
		},
	})
	return nil
}
