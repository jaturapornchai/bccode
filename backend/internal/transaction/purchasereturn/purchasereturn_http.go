package purchasereturn

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	productbarcode_repositories "smlcloudplatform/internal/product/productbarcode/repositories"
	"smlcloudplatform/internal/transaction/purchasereturn/models"
	"smlcloudplatform/internal/transaction/purchasereturn/repositories"
	"smlcloudplatform/internal/transaction/purchasereturn/services"
	trancache "smlcloudplatform/internal/transaction/repositories"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/requestfilter"
	"smlcloudplatform/pkg/microservice"
)

type IPurchaseReturnHttp interface{}

type PurchaseReturnHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IPurchaseReturnService
}

func NewPurchaseReturnHttp(ms *microservice.Microservice, cfg config.IConfig) PurchaseReturnHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())
	producer := ms.Producer(cfg.MQConfig())

	repo := repositories.NewPurchaseReturnRepository(pst)
	repoMq := repositories.NewPurchaseReturnMessageQueueRepository(producer)

	productBarcodeRepo := productbarcode_repositories.NewProductBarcodeRepository(pst, cache)

	transCacheRepo := trancache.NewCacheRepository(cache)
	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewPurchaseReturnService(repo, transCacheRepo, productBarcodeRepo, repoMq, masterSyncCacheRepo, services.PurchaseReturnParser{})

	return PurchaseReturnHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h PurchaseReturnHttp) RegisterHttp() {

	h.ms.POST("/transaction/purchase-return/bulk", h.SaveBulk)

	h.ms.GET("/transaction/purchase-return", h.SearchPurchaseReturnPage)
	h.ms.GET("/transaction/purchase-return/list", h.SearchPurchaseReturnStep)
	h.ms.POST("/transaction/purchase-return", h.CreatePurchaseReturn)
	h.ms.GET("/transaction/purchase-return/:id", h.InfoPurchaseReturn)
	h.ms.GET("/transaction/purchase-return/code/:code", h.InfoPurchaseReturnByCode)
	h.ms.PUT("/transaction/purchase-return/:id", h.UpdatePurchaseReturn)
	h.ms.DELETE("/transaction/purchase-return/:id", h.DeletePurchaseReturn)
	h.ms.DELETE("/transaction/purchase-return", h.DeletePurchaseReturnByGUIDs)
}

// Create PurchaseReturn godoc
// @Description Create PurchaseReturn
// @Tags		PurchaseReturn
// @Param		PurchaseReturn  body      models.PurchaseReturn  true  "PurchaseReturn"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/purchase-return [post]
func (h PurchaseReturnHttp) CreatePurchaseReturn(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := &models.PurchaseReturn{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, docNo, err := h.svc.CreatePurchaseReturn(shopID, authUsername, *docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
		ID:      idx,
		Data:    docNo,
	})
	return nil
}

// Update PurchaseReturn godoc
// @Description Update PurchaseReturn
// @Tags		PurchaseReturn
// @Param		id  path      string  true  "PurchaseReturn ID"
// @Param		PurchaseReturn  body      models.PurchaseReturn  true  "PurchaseReturn"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/purchase-return/{id} [put]
func (h PurchaseReturnHttp) UpdatePurchaseReturn(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.PurchaseReturn{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdatePurchaseReturn(shopID, id, authUsername, *docReq)

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

// Delete PurchaseReturn godoc
// @Description Delete PurchaseReturn
// @Tags		PurchaseReturn
// @Param		id  path      string  true  "PurchaseReturn ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/purchase-return/{id} [delete]
func (h PurchaseReturnHttp) DeletePurchaseReturn(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeletePurchaseReturn(shopID, id, authUsername)

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

// Delete PurchaseReturn godoc
// @Description Delete PurchaseReturn
// @Tags		PurchaseReturn
// @Param		PurchaseReturn  body      []string  true  "PurchaseReturn GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/purchase-return [delete]
func (h PurchaseReturnHttp) DeletePurchaseReturnByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeletePurchaseReturnByGUIDs(shopID, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get PurchaseReturn godoc
// @Description get PurchaseReturn info by guidfixed
// @Tags		PurchaseReturn
// @Param		id  path      string  true  "PurchaseReturn guidfixed"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/purchase-return/{id} [get]
func (h PurchaseReturnHttp) InfoPurchaseReturn(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get PurchaseReturn %v", id)
	doc, err := h.svc.InfoPurchaseReturn(shopID, id)

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

// Get PurchaseReturn By Code godoc
// @Description get PurchaseReturn info by Code
// @Tags		PurchaseReturn
// @Param		code  path      string  true  "PurchaseReturn Code"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/purchase-return/code/{code} [get]
func (h PurchaseReturnHttp) InfoPurchaseReturnByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	code := ctx.Param("code")

	doc, err := h.svc.InfoPurchaseReturnByCode(shopID, code)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    doc,
	})
	return nil
}

// List PurchaseReturn step godoc
// @Description get list step
// @Tags		PurchaseReturn
// @Param		q		query	string		false  "Search Value"
// @Param		custcode	query	string		false  "cust code"
// @Param		branchcode	query	string		false  "branch code"
// @Param		fromdate	query	string		false  "from date"
// @Param		todate	query	string		false  "to date"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/purchase-return [get]
func (h PurchaseReturnHttp) SearchPurchaseReturnPage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)

	filters := requestfilter.GenerateFilters(ctx.QueryParam, []requestfilter.FilterRequest{
		{
			Param: "custcode",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "-",
			Field: "docdatetime",
			Type:  requestfilter.FieldTypeRangeDate,
		},
		{
			Param: "branchcode",
			Field: "branch.code",
			Type:  requestfilter.FieldTypeString,
		},
	})

	docList, pagination, err := h.svc.SearchPurchaseReturn(shopID, filters, pageable)

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

// List PurchaseReturn godoc
// @Description search limit offset
// @Tags		PurchaseReturn
// @Param		q		query	string		false  "Search Value"
// @Param		custcode	query	string		false  "cust code"
// @Param		branchcode	query	string		false  "branch code"
// @Param		fromdate	query	string		false  "from date"
// @Param		todate	query	string		false  "to date"
// @Param		offset	query	integer		false  "offset"
// @Param		limit	query	integer		false  "limit"
// @Param		lang	query	string		false  "lang"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/purchase-return/list [get]
func (h PurchaseReturnHttp) SearchPurchaseReturnStep(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageableStep := utils.GetPageableStep(ctx.QueryParam)

	lang := ctx.QueryParam("lang")

	filters := requestfilter.GenerateFilters(ctx.QueryParam, []requestfilter.FilterRequest{
		{
			Param: "custcode",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "-",
			Field: "docdatetime",
			Type:  requestfilter.FieldTypeRangeDate,
		},
		{
			Param: "branchcode",
			Field: "branch.code",
			Type:  requestfilter.FieldTypeString,
		},
	})

	docList, total, err := h.svc.SearchPurchaseReturnStep(shopID, lang, filters, pageableStep)

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

// Create PurchaseReturn Bulk godoc
// @Description Create PurchaseReturn
// @Tags		PurchaseReturn
// @Param		PurchaseReturn  body      []models.PurchaseReturn  true  "PurchaseReturn"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/purchase-return/bulk [post]
func (h PurchaseReturnHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	input := ctx.ReadInput()

	dataReq := []models.PurchaseReturn{}
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
