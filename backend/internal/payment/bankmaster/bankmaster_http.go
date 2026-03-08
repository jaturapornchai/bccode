package bankmaster

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/payment/bankmaster/models"
	"smlcloudplatform/internal/payment/bankmaster/repositories"
	"smlcloudplatform/internal/payment/bankmaster/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
)

type IBankMasterHttp interface{}

type BankMasterHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IBankMasterHttpService
}

func NewBankMasterHttp(ms *microservice.Microservice, cfg config.IConfig) BankMasterHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())

	repo := repositories.NewBankMasterRepository(pst)

	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewBankMasterHttpService(repo, masterSyncCacheRepo)

	return BankMasterHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h BankMasterHttp) RegisterHttp() {

	h.ms.POST("/payment/bankmaster/bulk", h.SaveBulk)

	h.ms.GET("/payment/bankmaster", h.SearchBankMasterPage)
	h.ms.GET("/payment/bankmaster/list", h.SearchBankMasterLimit)
	h.ms.POST("/payment/bankmaster", h.CreateBankMaster)
	h.ms.GET("/payment/bankmaster/:id", h.InfoBankMaster)
	h.ms.PUT("/payment/bankmaster/:id", h.UpdateBankMaster)
	h.ms.DELETE("/payment/bankmaster/:id", h.DeleteBankMaster)
	h.ms.DELETE("/payment/bankmaster", h.DeleteBankMasterByGUIDs)
}

// Create BankMaster godoc
// @Description Create BankMaster
// @Tags		BankMaster
// @Param		BankMaster  body      models.BankMaster  true  "BankMaster"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /payment/bankmaster [post]
func (h BankMasterHttp) CreateBankMaster(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := &models.BankMaster{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.CreateBankMaster(shopID, authUsername, *docReq)

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

// Update BankMaster godoc
// @Description Update BankMaster
// @Tags		BankMaster
// @Param		id  path      string  true  "BankMaster ID"
// @Param		BankMaster  body      models.BankMaster  true  "BankMaster"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /payment/bankmaster/{id} [put]
func (h BankMasterHttp) UpdateBankMaster(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.BankMaster{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateBankMaster(shopID, id, authUsername, *docReq)

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

// Delete BankMaster godoc
// @Description Delete BankMaster
// @Tags		BankMaster
// @Param		id  path      string  true  "BankMaster ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /payment/bankmaster/{id} [delete]
func (h BankMasterHttp) DeleteBankMaster(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteBankMaster(shopID, id, authUsername)

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

// Delete BankMaster godoc
// @Description Delete BankMaster
// @Tags		BankMaster
// @Param		BankMaster  body      []string  true  "BankMaster GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /payment/bankmaster [delete]
func (h BankMasterHttp) DeleteBankMasterByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeleteBankMasterByGUIDs(shopID, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get BankMaster godoc
// @Description get struct array by ID
// @Tags		BankMaster
// @Param		id  path      string  true  "BankMaster ID"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /payment/bankmaster/{id} [get]
func (h BankMasterHttp) InfoBankMaster(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get BankMaster %v", id)
	doc, err := h.svc.InfoBankMaster(shopID, id)

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

// List BankMaster godoc
// @Description get struct array by ID
// @Tags		BankMaster
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /payment/bankmaster [get]
func (h BankMasterHttp) SearchBankMasterPage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)
	docList, pagination, err := h.svc.SearchBankMaster(shopID, pageable)

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

// List BankMaster godoc
// @Description search limit offset
// @Tags		BankMaster
// @Param		q		query	string		false  "Search Value"
// @Param		offset	query	integer		false  "offset"
// @Param		limit	query	integer		false  "limit"
// @Param		lang	query	string		false  "lang"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /payment/bankmaster/list [get]
func (h BankMasterHttp) SearchBankMasterLimit(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageStep := utils.GetPageableStep(ctx.QueryParam)

	lang := ctx.QueryParam("lang")

	docList, total, err := h.svc.SearchBankMasterStep(shopID, lang, pageStep)

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

// Create BankMaster Bulk godoc
// @Description Create BankMaster
// @Tags		BankMaster
// @Param		BankMaster  body      []models.BankMaster  true  "BankMaster"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /payment/bankmaster/bulk [post]
func (h BankMasterHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	input := ctx.ReadInput()

	dataReq := []models.BankMaster{}
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
