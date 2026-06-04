// gradeproduct_http.go
package gradeproduct

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	productbarcode_repositories "smlcloudplatform/internal/product/productbarcode/repositories"
	"smlcloudplatform/internal/smlaiproduct/gradeproduct/models"
	graderepo "smlcloudplatform/internal/smlaiproduct/gradeproduct/repositories"
	gradesvc "smlcloudplatform/internal/smlaiproduct/gradeproduct/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
	"time"
)

type IGradeProductHttp interface{}

type GradeProductHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc gradesvc.IGradeProductHttpService
}

func NewGradeProductHttp(ms *microservice.Microservice, cfg config.IConfig) GradeProductHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())

	repo := graderepo.NewGradeProductRepository(pst)
	repoProductBarcode := productbarcode_repositories.NewProductBarcodeRepository(pst, cache)
	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)

	svc := gradesvc.NewGradeProductHttpService(repo, repoProductBarcode, masterSyncCacheRepo, 15*time.Second)

	return GradeProductHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h GradeProductHttp) RegisterHttp() {
	h.ms.POST("/aicloud/grade/bulk", h.SaveBulk)
	h.ms.GET("/aicloud/grade", h.SearchGradeProductPage)
	h.ms.GET("/aicloud/grade/list", h.SearchGradeProductStep)
	h.ms.POST("/aicloud/grade", h.CreateGradeProduct)
	h.ms.GET("/aicloud/grade/:id", h.InfoGradeProduct)
	h.ms.GET("/aicloud/grade/code/:code", h.InfoGradeProductByCode)
	h.ms.PUT("/aicloud/grade/:id", h.UpdateGradeProduct)
	h.ms.DELETE("/aicloud/grade/:id", h.DeleteGradeProduct)
	h.ms.DELETE("/aicloud/grade", h.DeleteGradeProductByGUIDs)
}

// Create GradeProduct godoc
// @Description Create GradeProduct
// @Tags		AICloudProductGrade
// @Param		GradeProduct	body	models.GradeProduct	true	"GradeProduct"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401	{object}	common.AuthResponseFailed
// @Security		AccessToken
// @Router /aicloud/grade [post]
func (h GradeProductHttp) CreateGradeProduct(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	holdingCode := ctx.UserInfo().HoldingCode
	input := ctx.ReadInput()
	docReq := &models.GradeProduct{}
	if err := json.Unmarshal([]byte(input), docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err := ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.CreateGradeProduct(holdingCode, authUsername, *docReq)
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

// Update GradeProduct godoc
// @Description Update GradeProduct
// @Tags		AICloudProductGrade
// @Param		id	path	string	true	"GradeProduct ID"
// @Param		GradeProduct	body	models.GradeProduct	true	"GradeProduct"
// @Accept		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401	{object}	common.AuthResponseFailed
// @Security		AccessToken
// @Router /aicloud/grade/{id} [put]
func (h GradeProductHttp) UpdateGradeProduct(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")
	input := ctx.ReadInput()
	docReq := &models.GradeProduct{}
	if err := json.Unmarshal([]byte(input), docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err := ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err := h.svc.UpdateGradeProduct(holdingCode, id, authUsername, *docReq)
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

// Delete GradeProduct godoc
// @Description Delete GradeProduct
// @Tags		AICloudProductGrade
// @Param		id  path      string  true  "GradeProduct ID"
// @Accept 		json
// @Success		200	{object} common.ResponseSuccessWithID
// @Failure		401 {object} common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/grade/{id} [delete]
func (h GradeProductHttp) DeleteGradeProduct(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteGradeProduct(holdingCode, id, authUsername)
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

// Delete GradeProduct godoc
// @Description Delete GradeProduct
// @Tags		AICloudProductGrade
// @Param		GradeProduct  body      []string  true  "GradeProduct GUIDs"
// @Accept 		json
// @Success		200	{object} common.ResponseSuccessWithID
// @Failure		401 {object} common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/grade [delete]
func (h GradeProductHttp) DeleteGradeProductByGUIDs(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username

	input := ctx.ReadInput()
	docReq := []string{}
	if err := json.Unmarshal([]byte(input), &docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err := h.svc.DeleteGradeProductByGUIDs(holdingCode, authUsername, docReq)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})
	return nil
}

// Get GradeProduct godoc
// @Description get GradeProduct info by guidfixed
// @Tags		AICloudProductGrade
// @Param		id  path      string  true  "GradeProduct guidfixed"
// @Accept 		json
// @Success		200	{object} common.ApiResponse
// @Failure		401 {object} common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/grade/{id} [get]
func (h GradeProductHttp) InfoGradeProduct(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get GradeProduct %v", id)
	doc, err := h.svc.InfoGradeProduct(holdingCode, id)
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

// Get GradeProduct By Code godoc
// @Description get GradeProduct info by Code
// @Tags		AICloudProductGrade
// @Param		code  path      string  true  "GradeProduct Code"
// @Accept 		json
// @Success		200	{object} common.ApiResponse
// @Failure		401 {object} common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/grade/code/{code} [get]
func (h GradeProductHttp) InfoGradeProductByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	code := ctx.Param("code")
	doc, err := h.svc.InfoGradeProductByCode(holdingCode, code)
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

// List GradeProduct step godoc
// @Description get list step
// @Tags		AICloudProductGrade
// @Param		q	query	string	false  "Search Value"
// @Param		page	query	integer	false  "Page"
// @Param		limit	query	integer	false  "Limit"
// @Accept 		json
// @Success		200	{array} common.ApiResponse
// @Failure		401 {object} common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/grade [get]
func (h GradeProductHttp) SearchGradeProductPage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	pageable := utils.GetPageable(ctx.QueryParam)
	docList, pagination, err := h.svc.SearchGradeProduct(holdingCode, map[string]interface{}{}, pageable)
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

// List GradeProduct godoc
// @Description search limit offset
// @Tags		AICloudProductGrade
// @Param		q	query	string	false  "Search Value"
// @Param		offset	query	integer	false  "offset"
// @Param		limit	query	integer	false  "limit"
// @Param		lang	query	string	false  "lang"
// @Accept 		json
// @Success		200	{array} common.ApiResponse
// @Failure		401 {object} common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/grade/list [get]
func (h GradeProductHttp) SearchGradeProductStep(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	pageableStep := utils.GetPageableStep(ctx.QueryParam)
	lang := ctx.QueryParam("lang")
	docList, total, err := h.svc.SearchGradeProductStep(holdingCode, lang, pageableStep)
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

// Create GradeProduct Bulk godoc
// @Description Create GradeProduct
// @Tags		AICloudProductGrade
// @Param		GradeProduct  body      []models.GradeProduct  true  "GradeProduct"
// @Accept 		json
// @Success		201	{object} common.BulkResponse
// @Failure		401 {object} common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/grade/bulk [post]
func (h GradeProductHttp) SaveBulk(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	input := ctx.ReadInput()
	dataReq := []models.GradeProduct{}
	if err := json.Unmarshal([]byte(input), &dataReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	bulkResponse, err := h.svc.SaveInBatch(holdingCode, authUsername, dataReq)
	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.BulkResponse{
		Success:    true,
		BulkImport: bulkResponse,
	})
	return nil
}
