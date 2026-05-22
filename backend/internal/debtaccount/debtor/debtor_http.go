package debtor

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	"smlcloudplatform/internal/debtaccount/debtor/models"
	"smlcloudplatform/internal/debtaccount/debtor/repositories"
	"smlcloudplatform/internal/debtaccount/debtor/services"
	groupRepositories "smlcloudplatform/internal/debtaccount/debtorgroup/repositories"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/requestfilter"
	"smlcloudplatform/pkg/microservice"
)

type IDebtorHttp interface{}

type DebtorHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IDebtorHttpService
}

func NewDebtorHttp(ms *microservice.Microservice, cfg config.IConfig) DebtorHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())
	prod := ms.Producer(cfg.MQConfig())

	repo := repositories.NewDebtorRepository(pst)
	repoMq := repositories.NewDebtorMessageQueueRepository(prod)
	repoGroup := groupRepositories.NewDebtorGroupRepository(pst)
	pointTransRepo := repositories.NewPointTransactionRepository(pst)

	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewDebtorHttpService(
		repo,
		repoMq,
		repoGroup,
		pointTransRepo,
		masterSyncCacheRepo,
		utils.HashPassword,
		utils.CheckHashPassword,
	)

	return DebtorHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h DebtorHttp) RegisterHttp() {

	h.ms.POST("/debtaccount/debtor/bulk", h.SaveBulk)

	h.ms.POST("/debtaccount/debtor/auth", h.AuthDebtor)
	h.ms.GET("/debtaccount/debtor", h.SearchDebtorPage)
	h.ms.GET("/debtaccount/debtor/list", h.SearchDebtorStep)
	h.ms.POST("/debtaccount/debtor", h.CreateDebtor)
	h.ms.GET("/debtaccount/debtor/:id", h.InfoDebtor)
	h.ms.GET("/debtaccount/debtor/code/:code", h.InfoDebtorByCode)
	h.ms.GET("/debtaccount/debtor/line/:code", h.InfoDebtorByLine)
	h.ms.GET("/debtaccount/debtor/code/:code/pointtransactions", h.SearchPointTransactions)
	h.ms.POST("/debtaccount/debtor/pointscode/:pointscode/recalpoint", h.RecalPointByPointsCode)
	h.ms.POST("/debtaccount/debtor/pointscode/:pointscode/addpoint", h.AddPointManually)
	h.ms.POST("/debtaccount/debtor/points/bulk-add", h.BulkAddPoints)
	h.ms.DELETE("/debtaccount/debtor/pointscode/:pointscode/manual-transaction/:docno", h.DeleteManualPointTransaction)
	h.ms.PUT("/debtaccount/debtor/:id", h.UpdateDebtor)
	h.ms.DELETE("/debtaccount/debtor/:id", h.DeleteDebtor)
	h.ms.DELETE("/debtaccount/debtor", h.DeleteDebtorByGUIDs)
}

// Auth Debtor godoc
// @Description Auth Debtor
// @Tags		Debtor
// @Param		DebtorAuth  body      models.DebtorAuth  true  "Debtor Auth"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/debtor/auth [post]
func (h DebtorHttp) AuthDebtor(ctx microservice.IContext) error {

	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	payload := &models.DebtorAuth{}
	err := json.Unmarshal([]byte(input), &payload)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(payload); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.InfoAuthDebtor(shopID, payload.Username, payload.Password)

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

// Create Debtor godoc
// @Description Create Debtor
// @Tags		Debtor
// @Param		Debtor  body      models.DebtorRequest  true  "Debtor"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/debtor [post]
func (h DebtorHttp) CreateDebtor(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := &models.DebtorRequest{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.CreateDebtor(shopID, authUsername, *docReq)

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

// Update Debtor godoc
// @Description Update Debtor
// @Tags		Debtor
// @Param		id  path      string  true  "Debtor ID"
// @Param		Debtor  body      models.DebtorRequest  true  "Debtor"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/debtor/{id} [put]
func (h DebtorHttp) UpdateDebtor(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.DebtorRequest{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateDebtor(shopID, id, authUsername, *docReq)

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

// Delete Debtor godoc
// @Description Delete Debtor
// @Tags		Debtor
// @Param		id  path      string  true  "Debtor ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/debtor/{id} [delete]
func (h DebtorHttp) DeleteDebtor(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteDebtor(shopID, id, authUsername)

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

// Delete Debtor godoc
// @Description Delete Debtor
// @Tags		Debtor
// @Param		Debtor  body      []string  true  "Debtor GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/debtor [delete]
func (h DebtorHttp) DeleteDebtorByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeleteDebtorByGUIDs(shopID, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get Debtor godoc
// @Description get struct array by ID
// @Tags		Debtor
// @Param		id  path      string  true  "Debtor ID"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/debtor/{id} [get]
func (h DebtorHttp) InfoDebtor(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get Debtor %v", id)
	doc, err := h.svc.InfoDebtor(shopID, id)

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

// Get Debtor By Code godoc
// @Description get debtor by code
// @Tags		Debtor
// @Param		code  path      string  true  "Debtor Code"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/debtor/code/{code} [get]
func (h DebtorHttp) InfoDebtorByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	code := ctx.Param("code")

	doc, err := h.svc.InfoDebtorByCode(shopID, code)

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

// Get Debtor By Line godoc
// @Description get debtor by code
// @Tags		Debtor
// @Param		code  path      string  true  "Debtor Line"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/debtor/line/{code} [get]
func (h DebtorHttp) InfoDebtorByLine(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	code := ctx.Param("code")

	doc, err := h.svc.InfoDebtorByLine(shopID, code)

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

// List Debtor godoc
// @Description get struct array by ID
// @Tags		Debtor
// @Param		q		query	string		false  "Search Value"
// @Param		groups		query	string		false  "groups guidfixed"
// @Param		shopsid		query	string		false  "shopsid ex. s001,s002"
// @Param		page	query	integer		false  "page"
// @Param		limit	query	integer		false  "limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/debtor [get]
func (h DebtorHttp) SearchDebtorPage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)

	filters := requestfilter.GenerateFilters(ctx.QueryParam, []requestfilter.FilterRequest{
		{
			Param: "groups",
			Field: "groups",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "shopsid",
			Field: "shopid",
			Type:  requestfilter.FieldTypeString,
		},
	})
	docList, pagination, err := h.svc.SearchDebtor(shopID, filters, pageable)

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

// List Debtor godoc
// @Description search limit offset
// @Tags		Debtor
// @Param		q		query	string		false  "Search Value"
// @Param		groups		query	string		false  "groups guidfixed"
// @Param		offset	query	integer		false  "offset"
// @Param		limit	query	integer		false  "limit"
// @Param		lang	query	string		false  "lang"
// @Param 		shopsid query 	string 		false  "shopsid ex. s001,s002"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/debtor/list [get]
func (h DebtorHttp) SearchDebtorStep(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageableStep := utils.GetPageableStep(ctx.QueryParam)

	lang := ctx.QueryParam("lang")

	filters := requestfilter.GenerateFilters(ctx.QueryParam, []requestfilter.FilterRequest{
		{
			Param: "groups",
			Field: "groups",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "shopsid",
			Field: "shopid",
			Type:  requestfilter.FieldTypeString,
		},
	})

	docList, total, err := h.svc.SearchDebtorStep(shopID, lang, filters, pageableStep)

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

// Create Debtor Bulk godoc
// @Description Create Debtor
// @Tags		Debtor
// @Param		Debtor  body      []models.DebtorRequest  true  "Debtor"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/debtor/bulk [post]
func (h DebtorHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	input := ctx.ReadInput()

	dataReq := []models.DebtorRequest{}
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

// Search Point Transactions godoc
// @Description get point transaction movements for debtor
// @Tags		Debtor
// @Param		code  path      string  true  "Point Code"
// @Param		offset	query	integer		false  "offset"
// @Param		limit	query	integer		false  "limit"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/debtor/code/{code}/pointtransactions [get]
func (h DebtorHttp) SearchPointTransactions(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	code := ctx.Param("code")
	pageableStep := utils.GetPageableStep(ctx.QueryParam)

	docList, total, err := h.svc.SearchPointTransactions(shopID, code, pageableStep)

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

// Recalculate Point Balance godoc
// @Description recalculate point balance for a pointscode based on all point transactions
// @Tags		Debtor
// @Param		pointscode  path      string  true  "Points Code"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/debtor/pointscode/{pointscode}/recalpoint [post]
func (h DebtorHttp) RecalPointByPointsCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	pointsCode := ctx.Param("points_code")

	if pointsCode == "" {
		ctx.ResponseError(http.StatusBadRequest, "pointscode is required")
		return nil
	}

	err := h.svc.RecalPointByPointsCode(shopID, pointsCode, authUsername)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Message: "Point balance recalculated successfully",
	})
	return nil
}

// Add Point Manually godoc
// @Description Add points manually to a customer (can be used for opening balance, adjustments, promotions, etc.)
// @Tags		Debtor
// @Param		pointscode  path      string  true  "Points Code"
// @Param		AddPointRequest  body      models.OpeningBalancePointRequest  true  "Add Point Request"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		400 {object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/debtor/pointscode/{pointscode}/addpoint [post]
func (h DebtorHttp) AddPointManually(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	pointsCode := ctx.Param("points_code")

	if pointsCode == "" {
		ctx.ResponseError(http.StatusBadRequest, "pointscode is required")
		return nil
	}

	// Parse request body
	var req models.OpeningBalancePointRequest
	if err := json.NewDecoder(ctx.Request().Body).Decode(&req); err != nil {
		ctx.ResponseError(http.StatusBadRequest, "invalid request body")
		return err
	}

	// Override pointsCode from path parameter
	req.PointsCode = pointsCode

	// Validate
	if req.PointAmount <= 0 {
		ctx.ResponseError(http.StatusBadRequest, "pointAmount must be greater than 0")
		return nil
	}

	err := h.svc.AddPointManually(shopID, req.PointsCode, req.PointAmount, req.Description, authUsername)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Message: "Points added successfully",
	})
	return nil
}

// Bulk Add Points godoc
// @Description Bulk add points for multiple customers (can be used for opening balance, adjustments, promotions, etc.)
// @Tags		Debtor
// @Param		BulkAddPointsRequest  body      []models.OpeningBalancePointRequest  true  "Bulk Add Points Requests"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse{data=models.BulkImportPointResult}
// @Failure		400 {object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/debtor/points/bulk-add [post]
func (h DebtorHttp) BulkAddPoints(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	// Parse request body
	var pointsList []models.OpeningBalancePointRequest
	if err := json.NewDecoder(ctx.Request().Body).Decode(&pointsList); err != nil {
		ctx.ResponseError(http.StatusBadRequest, "invalid request body")
		return err
	}

	if len(pointsList) == 0 {
		ctx.ResponseError(http.StatusBadRequest, "points list cannot be empty")
		return nil
	}

	result, err := h.svc.BulkAddPoints(shopID, pointsList, authUsername)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Message: "Bulk add points completed",
		Data:    result,
	})
	return nil
}

// DeleteManualPointTransaction godoc
// @Description Delete manual point transaction by docNo (only TransactionType 3 can be deleted)
// @Tags		Debtor
// @Param		pointscode	path	string	true	"Points Code"
// @Param		docno		path	string	true	"Transaction Document Number"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		400 {object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /debtaccount/debtor/pointscode/{pointscode}/manual-transaction/{docno} [delete]
func (h DebtorHttp) DeleteManualPointTransaction(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	// Get parameters from URL
	pointsCode := ctx.Param("points_code")
	docNo := ctx.Param("docno")

	// Validate inputs
	if pointsCode == "" {
		ctx.ResponseError(http.StatusBadRequest, "pointsCode is required")
		return nil
	}

	if docNo == "" {
		ctx.ResponseError(http.StatusBadRequest, "docNo is required")
		return nil
	}

	// Delete manual point transaction
	err := h.svc.DeleteManualPointTransaction(shopID, pointsCode, docNo, authUsername)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Message: "Manual point transaction deleted successfully",
	})
	return nil
}
