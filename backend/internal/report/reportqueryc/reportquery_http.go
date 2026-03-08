package reportqueryc

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/report/reportqueryc/models"
	"smlcloudplatform/internal/report/reportqueryc/repositories"
	"smlcloudplatform/internal/report/reportqueryc/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
)

type IReportQueryHttp interface{}

type ReportQueryHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IReportQueryHttpService
}

func NewReportQueryHttp(ms *microservice.Microservice, cfg config.IConfig) ReportQueryHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	pstClickHouse := ms.ClickHousePersister(cfg.ClickHouseConfig())

	repo := repositories.NewReportQueryRepository(pst)
	repoClickHouse := repositories.NewReportQueryClickHouseRepository(pstClickHouse)

	svc := services.NewReportQueryHttpService(repo, repoClickHouse)

	return ReportQueryHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h ReportQueryHttp) RegisterHttp() {

	h.ms.POST("/report/playground", h.PlaygroundReportQuery)
	h.ms.POST("/report/execute", h.ExecuteReportQuery)
	h.ms.GET("/report/query", h.SearchReportQueryPage)
	h.ms.GET("/report/query/list", h.SearchReportQueryStep)
	h.ms.POST("/report/query", h.CreateReportQuery)
	h.ms.GET("/report/query/:id", h.InfoReportQuery)
	h.ms.GET("/report/query/code/:code", h.InfoReportQueryByCode)
	h.ms.PUT("/report/query/:id", h.UpdateReportQuery)
	h.ms.DELETE("/report/query/:id", h.DeleteReportQuery)
	h.ms.DELETE("/report/query", h.DeleteReportQueryByGUIDs)
}

// Playground ReportQuery godoc
// @Description Playground ReportQuery
// @Tags		ReportQuery
// @Param		ReportQuery  body      models.ReportQuery  true  "ReportQuery"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /report/playground [post]
func (h ReportQueryHttp) PlaygroundReportQuery(ctx microservice.IContext) error {
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	reqBody := &models.Query{}
	err := json.Unmarshal([]byte(input), &reqBody)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(reqBody); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	result, err := h.svc.PlaygroundReportQuery(shopID, *reqBody)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
		Data:    result,
	})
	return nil
}

// Execute ReportQuery godoc
// @Description Execute ReportQuery
// @Tags		ReportQuery
// @Param		ReportQuery  body      models.ReportQuery  true  "ReportQuery"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /report/execute [post]
func (h ReportQueryHttp) ExecuteReportQuery(ctx microservice.IContext) error {
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	reportCode := ctx.QueryParam("code")
	pageable := utils.GetPageable(ctx.QueryParam)

	reqBody := &[]models.QueryParamRequest{}
	err := json.Unmarshal([]byte(input), &reqBody)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(reqBody); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	result, pagination, err := h.svc.ExecuteReportQuery(shopID, reportCode, *reqBody, pageable)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success:    true,
		Data:       result,
		Pagination: pagination,
	})
	return nil
}

// Create ReportQuery godoc
// @Description Create ReportQuery
// @Tags		ReportQuery
// @Param		ReportQuery  body      models.ReportQuery  true  "ReportQuery"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /report/query [post]
func (h ReportQueryHttp) CreateReportQuery(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := &models.ReportQuery{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.CreateReportQuery(shopID, authUsername, *docReq)

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

// Update ReportQuery godoc
// @Description Update ReportQuery
// @Tags		ReportQuery
// @Param		id  path      string  true  "ReportQuery ID"
// @Param		ReportQuery  body      models.ReportQuery  true  "ReportQuery"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /report/query/{id} [put]
func (h ReportQueryHttp) UpdateReportQuery(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.ReportQuery{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateReportQuery(shopID, id, authUsername, *docReq)

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

// Delete ReportQuery godoc
// @Description Delete ReportQuery
// @Tags		ReportQuery
// @Param		id  path      string  true  "ReportQuery ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /report/query/{id} [delete]
func (h ReportQueryHttp) DeleteReportQuery(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteReportQuery(shopID, id, authUsername)

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

// Delete ReportQuery godoc
// @Description Delete ReportQuery
// @Tags		ReportQuery
// @Param		ReportQuery  body      []string  true  "ReportQuery GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /report/query [delete]
func (h ReportQueryHttp) DeleteReportQueryByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeleteReportQueryByGUIDs(shopID, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get ReportQuery godoc
// @Description get ReportQuery info by guidfixed
// @Tags		ReportQuery
// @Param		id  path      string  true  "ReportQuery guidfixed"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /report/query/{id} [get]
func (h ReportQueryHttp) InfoReportQuery(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get ReportQuery %v", id)
	doc, err := h.svc.InfoReportQuery(shopID, id)

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

// Get ReportQuery By Code godoc
// @Description get ReportQuery info by Code
// @Tags		ReportQuery
// @Param		code  path      string  true  "ReportQuery Code"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /report/query/code/{code} [get]
func (h ReportQueryHttp) InfoReportQueryByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	code := ctx.Param("code")

	doc, err := h.svc.InfoReportQueryByCode(shopID, code)

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

// List ReportQuery step godoc
// @Description get list step
// @Tags		ReportQuery
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /report/query [get]
func (h ReportQueryHttp) SearchReportQueryPage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)

	docList, pagination, err := h.svc.SearchReportQuery(shopID, map[string]interface{}{}, pageable)

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

// List ReportQuery godoc
// @Description search limit offset
// @Tags		ReportQuery
// @Param		q		query	string		false  "Search Value"
// @Param		offset	query	integer		false  "offset"
// @Param		limit	query	integer		false  "limit"
// @Param		lang	query	string		false  "lang"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /report/query/list [get]
func (h ReportQueryHttp) SearchReportQueryStep(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageableStep := utils.GetPageableStep(ctx.QueryParam)

	lang := ctx.QueryParam("lang")

	docList, total, err := h.svc.SearchReportQueryStep(shopID, lang, pageableStep)

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
