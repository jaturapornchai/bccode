// groupsubtwoproduct_http.go
package groupsubtwoproduct

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	productbarcode_repositories "smlcloudplatform/internal/product/productbarcode/repositories"
	groupRepositories "smlcloudplatform/internal/smlaiproduct/groupproduct/repositories"
	groupsuboneRepositories "smlcloudplatform/internal/smlaiproduct/groupsuboneproduct/repositories"
	"smlcloudplatform/internal/smlaiproduct/groupsubtwoproduct/models"
	groupsubtworepo "smlcloudplatform/internal/smlaiproduct/groupsubtwoproduct/repositories"
	groupsubtwosvc "smlcloudplatform/internal/smlaiproduct/groupsubtwoproduct/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
	"time"
)

type IGroupsubtwoProductHttp interface{}

type GroupsubtwoProductHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc groupsubtwosvc.IGroupsubtwoProductHttpService
}

func NewGroupsubtwoProductHttp(ms *microservice.Microservice, cfg config.IConfig) GroupsubtwoProductHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())

	repo := groupsubtworepo.NewGroupsubtwoProductRepository(pst)
	groupRepo := groupRepositories.NewGroupProductRepository(pst)
	groupSuboneRepo := groupsuboneRepositories.NewGroupsuboneProductRepository(pst)
	repoProductBarcode := productbarcode_repositories.NewProductBarcodeRepository(pst, cache)
	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)

	svc := groupsubtwosvc.NewGroupsubtwoProductHttpService(repo, groupRepo, groupSuboneRepo, repoProductBarcode, masterSyncCacheRepo, 15*time.Second)

	return GroupsubtwoProductHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h GroupsubtwoProductHttp) RegisterHttp() {
	h.ms.POST("/aicloud/groupsubtwo/bulk", h.SaveBulk)
	h.ms.GET("/aicloud/groupsubtwo", h.SearchGroupsubtwoProductPage)
	h.ms.GET("/aicloud/groupsubtwo/list", h.SearchGroupsubtwoProductStep)
	h.ms.POST("/aicloud/groupsubtwo", h.CreateGroupsubtwoProduct)
	h.ms.GET("/aicloud/groupsubtwo/:id", h.InfoGroupsubtwoProduct)
	h.ms.GET("/aicloud/groupsubtwo/code/:code", h.InfoGroupsubtwoProductByCode)
	h.ms.PUT("/aicloud/groupsubtwo/:id", h.UpdateGroupsubtwoProduct)
	h.ms.DELETE("/aicloud/groupsubtwo/:id", h.DeleteGroupsubtwoProduct)
	h.ms.DELETE("/aicloud/groupsubtwo", h.DeleteGroupsubtwoProductByGUIDs)
}

// Create GroupsubtwoProduct godoc
// @Description Create GroupsubtwoProduct
// @Tags		AICloudProductGroupsubtwo
// @Param		GroupsubtwoProduct	body	models.GroupsubtwoProduct	true	"GroupsubtwoProduct"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401	{object}	common.AuthResponseFailed
// @Security		AccessToken
// @Router /aicloud/groupsubtwo [post]
func (h GroupsubtwoProductHttp) CreateGroupsubtwoProduct(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	holdingCode := ctx.UserInfo().HoldingCode
	input := ctx.ReadInput()
	docReq := &models.GroupsubtwoProduct{}
	if err := json.Unmarshal([]byte(input), docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err := ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.CreateGroupsubtwoProduct(holdingCode, authUsername, *docReq)
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

// Update GroupsubtwoProduct godoc
// @Description Update GroupsubtwoProduct
// @Tags		AICloudProductGroupsubtwo
// @Param		id	path	string	true	"GroupsubtwoProduct ID"
// @Param		GroupsubtwoProduct	body	models.GroupsubtwoProduct	true	"GroupsubtwoProduct"
// @Accept		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401	{object}	common.AuthResponseFailed
// @Security		AccessToken
// @Router /aicloud/groupsubtwo/{id} [put]
func (h GroupsubtwoProductHttp) UpdateGroupsubtwoProduct(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")
	input := ctx.ReadInput()
	docReq := &models.GroupsubtwoProduct{}
	if err := json.Unmarshal([]byte(input), docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err := ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err := h.svc.UpdateGroupsubtwoProduct(holdingCode, id, authUsername, *docReq)
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

// Delete GroupsubtwoProduct godoc
// @Description Delete GroupsubtwoProduct
// @Tags		AICloudProductGroupsubtwo
// @Param		id  path      string  true  "GroupsubtwoProduct ID"
// @Accept 		json
// @Success		200	{object} common.ResponseSuccessWithID
// @Failure		401 {object} common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/groupsubtwo/{id} [delete]
func (h GroupsubtwoProductHttp) DeleteGroupsubtwoProduct(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteGroupsubtwoProduct(holdingCode, id, authUsername)
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

// Delete GroupsubtwoProduct godoc
// @Description Delete GroupsubtwoProduct
// @Tags		AICloudProductGroupsubtwo
// @Param		GroupsubtwoProduct  body      []string  true  "GroupsubtwoProduct GUIDs"
// @Accept 		json
// @Success		200	{object} common.ResponseSuccessWithID
// @Failure		401 {object} common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/groupsubtwo [delete]
func (h GroupsubtwoProductHttp) DeleteGroupsubtwoProductByGUIDs(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username

	input := ctx.ReadInput()
	docReq := []string{}
	if err := json.Unmarshal([]byte(input), &docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err := h.svc.DeleteGroupsubtwoProductByGUIDs(holdingCode, authUsername, docReq)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})
	return nil
}

// Get GroupsubtwoProduct godoc
// @Description get GroupsubtwoProduct info by guidfixed
// @Tags		AICloudProductGroupsubtwo
// @Param		id  path      string  true  "GroupsubtwoProduct guidfixed"
// @Accept 		json
// @Success		200	{object} common.ApiResponse
// @Failure		401 {object} common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/groupsubtwo/{id} [get]
func (h GroupsubtwoProductHttp) InfoGroupsubtwoProduct(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get GroupsubtwoProduct %v", id)
	doc, err := h.svc.InfoGroupsubtwoProduct(holdingCode, id)
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

// Get GroupsubtwoProduct By Code godoc
// @Description get GroupsubtwoProduct info by Code
// @Tags		AICloudProductGroupsubtwo
// @Param		code  path      string  true  "GroupsubtwoProduct Code"
// @Accept 		json
// @Success		200	{object} common.ApiResponse
// @Failure		401 {object} common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/groupsubtwo/code/{code} [get]
func (h GroupsubtwoProductHttp) InfoGroupsubtwoProductByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	code := ctx.Param("code")
	doc, err := h.svc.InfoGroupsubtwoProductByCode(holdingCode, code)
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

// List GroupsubtwoProduct step godoc
// @Description get list step
// @Tags		AICloudProductGroupsubtwo
// @Param		q	query	string	false  "Search Value"
// @Param		page	query	integer	false  "Page"
// @Param		limit	query	integer	false  "Limit"
// @Param		GroupMainGuid	query	string	false  "Filter by Group Main GUID"
// @Param		GroupSubGuid	query	string	false  "Filter by Group Sub GUID"
// @Accept 		json
// @Success		200	{array} common.ApiResponse
// @Failure		401 {object} common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/groupsubtwo [get]
func (h GroupsubtwoProductHttp) SearchGroupsubtwoProductPage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	// สร้าง filter จาก query parameters
	filter := make(map[string]interface{})

	// เพิ่ม filter สำหรับ GroupMainGuid
	if groupMainGuid := ctx.QueryParam("GroupMainGuid"); groupMainGuid != "" {
		filter["groupMainGuid"] = groupMainGuid
	}

	// เพิ่ม filter สำหรับ GroupSubGuid
	if groupSubGuid := ctx.QueryParam("GroupSubGuid"); groupSubGuid != "" {
		filter["groupSubGuid"] = groupSubGuid
	}

	pageable := utils.GetPageable(ctx.QueryParam)
	docList, pagination, err := h.svc.SearchGroupsubtwoProduct(holdingCode, filter, pageable)
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

// List GroupsubtwoProduct godoc
// @Description search limit offset
// @Tags		AICloudProductGroupsubtwo
// @Param		q	query	string	false  "Search Value"
// @Param		offset	query	integer	false  "offset"
// @Param		limit	query	integer	false  "limit"
// @Param		lang	query	string	false  "lang"
// @Accept 		json
// @Success		200	{array} common.ApiResponse
// @Failure		401 {object} common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/groupsubtwo/list [get]
func (h GroupsubtwoProductHttp) SearchGroupsubtwoProductStep(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	pageableStep := utils.GetPageableStep(ctx.QueryParam)
	lang := ctx.QueryParam("lang")
	docList, total, err := h.svc.SearchGroupsubtwoProductStep(holdingCode, lang, pageableStep)
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

// Create GroupsubtwoProduct Bulk godoc
// @Description Create GroupsubtwoProduct
// @Tags		AICloudProductGroupsubtwo
// @Param		GroupsubtwoProduct  body      []models.GroupsubtwoProduct  true  "GroupsubtwoProduct"
// @Accept 		json
// @Success		201	{object} common.BulkResponse
// @Failure		401 {object} common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/groupsubtwo/bulk [post]
func (h GroupsubtwoProductHttp) SaveBulk(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	input := ctx.ReadInput()
	dataReq := []models.GroupsubtwoProduct{}
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
