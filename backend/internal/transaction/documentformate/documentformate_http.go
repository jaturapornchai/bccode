package documentformate

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/documentformate/models"
	"smlcloudplatform/internal/transaction/documentformate/repositories"
	"smlcloudplatform/internal/transaction/documentformate/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/requestfilter"
	"smlcloudplatform/pkg/microservice"
)

type IDocumentFormateHttp interface{}

type DocumentFormateHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IDocumentFormateHttpService
}

func NewDocumentFormateHttp(ms *microservice.Microservice, cfg config.IConfig) DocumentFormateHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())

	repo := repositories.NewDocumentFormateRepository(pst)

	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewDocumentFormateHttpService(repo, masterSyncCacheRepo)

	return DocumentFormateHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h DocumentFormateHttp) RegisterHttp() {

	h.ms.POST("/transaction/document-formate/bulk", h.SaveBulk)

	h.ms.GET("/transaction/document-formate", h.SearchDocumentFormatePage)
	h.ms.GET("/transaction/document-formate/list", h.SearchDocumentFormateStep)
	h.ms.POST("/transaction/document-formate", h.CreateDocumentFormate)
	h.ms.GET("/transaction/document-formate/:id", h.InfoDocumentFormate)
	h.ms.GET("/transaction/document-formate/code/:code", h.InfoDocumentFormateByCode)
	h.ms.PUT("/transaction/document-formate/:id", h.UpdateDocumentFormate)
	h.ms.DELETE("/transaction/document-formate/:id", h.DeleteDocumentFormate)
	h.ms.DELETE("/transaction/document-formate", h.DeleteDocumentFormateByGUIDs)
	h.ms.GET("/transaction/document-formate/default", h.InfoDocumentFormateDefault)
}

// Create DocumentFormate godoc
// @Description Create DocumentFormate
// @Tags		DocumentFormate
// @Param		DocumentFormate  body      models.DocumentFormate  true  "DocumentFormate"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/document-formate [post]
func (h DocumentFormateHttp) CreateDocumentFormate(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	holdingCode := ctx.UserInfo().HoldingCode
	input := ctx.ReadInput()

	docReq := &models.DocumentFormate{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.CreateDocumentFormate(holdingCode, authUsername, *docReq)

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

// Update DocumentFormate godoc
// @Description Update DocumentFormate
// @Tags		DocumentFormate
// @Param		id  path      string  true  "DocumentFormate ID"
// @Param		DocumentFormate  body      models.DocumentFormate  true  "DocumentFormate"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/document-formate/{id} [put]
func (h DocumentFormateHttp) UpdateDocumentFormate(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.DocumentFormate{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateDocumentFormate(holdingCode, id, authUsername, *docReq)

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

// Delete DocumentFormate godoc
// @Description Delete DocumentFormate
// @Tags		DocumentFormate
// @Param		id  path      string  true  "DocumentFormate ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/document-formate/{id} [delete]
func (h DocumentFormateHttp) DeleteDocumentFormate(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteDocumentFormate(holdingCode, id, authUsername)

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

// Delete DocumentFormate godoc
// @Description Delete DocumentFormate
// @Tags		DocumentFormate
// @Param		DocumentFormate  body      []string  true  "DocumentFormate GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/document-formate [delete]
func (h DocumentFormateHttp) DeleteDocumentFormateByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeleteDocumentFormateByGUIDs(holdingCode, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get DocumentFormate godoc
// @Description get DocumentFormate info by guidfixed
// @Tags		DocumentFormate
// @Param		id  path      string  true  "DocumentFormate guidfixed"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/document-formate/{id} [get]
func (h DocumentFormateHttp) InfoDocumentFormate(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get DocumentFormate %v", id)
	doc, err := h.svc.InfoDocumentFormate(holdingCode, id)

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

// Get DocumentFormate Default Module godoc
// @Description get DocumentFormate Default Module info by guidfixed
// @Tags		DocumentFormate
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/document-formate/default [get]
func (h DocumentFormateHttp) InfoDocumentFormateDefault(ctx microservice.IContext) error {

	doc, err := h.svc.GetModuleDefault()

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

// Get DocumentFormate By Code godoc
// @Description get DocumentFormate info by Code
// @Tags		DocumentFormate
// @Param		code  path      string  true  "DocumentFormate Code"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/document-formate/code/{code} [get]
func (h DocumentFormateHttp) InfoDocumentFormateByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	code := ctx.Param("code")

	doc, err := h.svc.InfoDocumentFormateByCode(holdingCode, code)

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

// List DocumentFormate step godoc
// @Description get list step
// @Tags		DocumentFormate
// @Param		q		query	string		false  "Search Value"
// @Param		module		query	string		false  "module ex.TF,RF,GL"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/document-formate [get]
func (h DocumentFormateHttp) SearchDocumentFormatePage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	pageable := utils.GetPageable(ctx.QueryParam)

	filters := requestfilter.GenerateFilters(ctx.QueryParam, []requestfilter.FilterRequest{
		{
			Param: "module",
			Type:  requestfilter.FieldTypeString,
		},
	})

	docList, pagination, err := h.svc.SearchDocumentFormate(holdingCode, filters, pageable)

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

// List DocumentFormate godoc
// @Description search limit offset
// @Tags		DocumentFormate
// @Param		q		query	string		false  "Search Value"
// @Param		module		query	string		false  "module ex.TF,RF,GL"
// @Param		offset	query	integer		false  "offset"
// @Param		limit	query	integer		false  "limit"
// @Param		lang	query	string		false  "lang"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/document-formate/list [get]
func (h DocumentFormateHttp) SearchDocumentFormateStep(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	pageableStep := utils.GetPageableStep(ctx.QueryParam)

	filters := requestfilter.GenerateFilters(ctx.QueryParam, []requestfilter.FilterRequest{
		{
			Param: "module",
			Type:  requestfilter.FieldTypeString,
		},
	})

	lang := ctx.QueryParam("lang")

	docList, total, err := h.svc.SearchDocumentFormateStep(holdingCode, lang, filters, pageableStep)

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

// Create DocumentFormate Bulk godoc
// @Description Create DocumentFormate
// @Tags		DocumentFormate
// @Param		DocumentFormate  body      []models.DocumentFormate  true  "DocumentFormate"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /transaction/document-formate/bulk [post]
func (h DocumentFormateHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	input := ctx.ReadInput()

	dataReq := []models.DocumentFormate{}
	err := json.Unmarshal([]byte(input), &dataReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	bulkResponse, err := h.svc.SaveInBatch(holdingCode, authUsername, dataReq)

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
