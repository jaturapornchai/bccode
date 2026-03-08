// groupsuboneproduct_http.go
package groupsuboneproduct

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	productbarcode_repositories "smlcloudplatform/internal/product/productbarcode/repositories"
	groupRepositories "smlcloudplatform/internal/smlaiproduct/groupproduct/repositories"
	"smlcloudplatform/internal/smlaiproduct/groupsuboneproduct/models"
	groupsubonerepo "smlcloudplatform/internal/smlaiproduct/groupsuboneproduct/repositories"
	groupsubonesvc "smlcloudplatform/internal/smlaiproduct/groupsuboneproduct/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
	"time"
)

type IGroupsuboneProductHttp interface{}

type GroupsuboneProductHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc groupsubonesvc.IGroupsuboneProductHttpService
}

func NewGroupsuboneProductHttp(ms *microservice.Microservice, cfg config.IConfig) GroupsuboneProductHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())

	repo := groupsubonerepo.NewGroupsuboneProductRepository(pst)
	groupRepo := groupRepositories.NewGroupProductRepository(pst)
	repoProductBarcode := productbarcode_repositories.NewProductBarcodeRepository(pst, cache)
	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)

	svc := groupsubonesvc.NewGroupsuboneProductHttpService(repo, groupRepo, repoProductBarcode, masterSyncCacheRepo, 15*time.Second)

	return GroupsuboneProductHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h GroupsuboneProductHttp) RegisterHttp() {
	h.ms.POST("/aicloud/groupsubone/bulk", h.SaveBulk)
	h.ms.GET("/aicloud/groupsubone", h.SearchGroupsuboneProductPage)
	h.ms.GET("/aicloud/groupsubone/list", h.SearchGroupsuboneProductStep)
	h.ms.POST("/aicloud/groupsubone", h.CreateGroupsuboneProduct)
	h.ms.GET("/aicloud/groupsubone/:id", h.InfoGroupsuboneProduct)
	h.ms.GET("/aicloud/groupsubone/code/:code", h.InfoGroupsuboneProductByCode)
	h.ms.PUT("/aicloud/groupsubone/:id", h.UpdateGroupsuboneProduct)
	h.ms.DELETE("/aicloud/groupsubone/:id", h.DeleteGroupsuboneProduct)
	h.ms.DELETE("/aicloud/groupsubone", h.DeleteGroupsuboneProductByGUIDs)
}

// Create GroupsuboneProduct godoc
// @Description Create GroupsuboneProduct
// @Tags		AICloudProductGroupsubone
// @Param		GroupsuboneProduct	body	models.GroupsuboneProduct	true	"GroupsuboneProduct"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401	{object}	common.AuthResponseFailed
// @Security		AccessToken
// @Router /aicloud/groupsubone [post]
func (h GroupsuboneProductHttp) CreateGroupsuboneProduct(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()
	docReq := &models.GroupsuboneProduct{}
	if err := json.Unmarshal([]byte(input), docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err := ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.CreateGroupsuboneProduct(shopID, authUsername, *docReq)
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

// Update GroupsuboneProduct godoc
// @Description Update GroupsuboneProduct
// @Tags		AICloudProductGroupsubone
// @Param		id	path	string	true	"GroupsuboneProduct ID"
// @Param		GroupsuboneProduct	body	models.GroupsuboneProduct	true	"GroupsuboneProduct"
// @Accept		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401	{object}	common.AuthResponseFailed
// @Security		AccessToken
// @Router /aicloud/groupsubone/{id} [put]
func (h GroupsuboneProductHttp) UpdateGroupsuboneProduct(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()
	docReq := &models.GroupsuboneProduct{}
	if err := json.Unmarshal([]byte(input), docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err := ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err := h.svc.UpdateGroupsuboneProduct(shopID, id, authUsername, *docReq)
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

// Delete GroupsuboneProduct godoc
// @Description Delete GroupsuboneProduct
// @Tags		AICloudProductGroupsubone
// @Param		id  path      string  true  "GroupsuboneProduct ID"
// @Accept 		json
// @Success		200	{object} common.ResponseSuccessWithID
// @Failure		401 {object} common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/groupsubone/{id} [delete]
func (h GroupsuboneProductHttp) DeleteGroupsuboneProduct(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteGroupsuboneProduct(shopID, id, authUsername)
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

// Delete GroupsuboneProduct godoc
// @Description Delete GroupsuboneProduct
// @Tags		AICloudProductGroupsubone
// @Param		GroupsuboneProduct  body      []string  true  "GroupsuboneProduct GUIDs"
// @Accept 		json
// @Success		200	{object} common.ResponseSuccessWithID
// @Failure		401 {object} common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/groupsubone [delete]
func (h GroupsuboneProductHttp) DeleteGroupsuboneProductByGUIDs(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	input := ctx.ReadInput()
	docReq := []string{}
	if err := json.Unmarshal([]byte(input), &docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err := h.svc.DeleteGroupsuboneProductByGUIDs(shopID, authUsername, docReq)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})
	return nil
}

// Get GroupsuboneProduct godoc
// @Description get GroupsuboneProduct info by guidfixed
// @Tags		AICloudProductGroupsubone
// @Param		id  path      string  true  "GroupsuboneProduct guidfixed"
// @Accept 		json
// @Success		200	{object} common.ApiResponse
// @Failure		401 {object} common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/groupsubone/{id} [get]
func (h GroupsuboneProductHttp) InfoGroupsuboneProduct(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get GroupsuboneProduct %v", id)
	doc, err := h.svc.InfoGroupsuboneProduct(shopID, id)
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

// Get GroupsuboneProduct By Code godoc
// @Description get GroupsuboneProduct info by Code
// @Tags		AICloudProductGroupsubone
// @Param		code  path      string  true  "GroupsuboneProduct Code"
// @Accept 		json
// @Success		200	{object} common.ApiResponse
// @Failure		401 {object} common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/groupsubone/code/{code} [get]
func (h GroupsuboneProductHttp) InfoGroupsuboneProductByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	code := ctx.Param("code")
	doc, err := h.svc.InfoGroupsuboneProductByCode(shopID, code)
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

// List GroupsuboneProduct step godoc
// @Description get list step
// @Tags		AICloudProductGroupsubone
// @Param		q	query	string	false  "Search Value"
// @Param		page	query	integer	false  "Page"
// @Param		limit	query	integer	false  "Limit"
// @Param		GroupMainGuid	query	string	false  "Filter by Group Main GUID"
// @Accept 		json
// @Success		200	{array} common.ApiResponse
// @Failure		401 {object} common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/groupsubone [get]
func (h GroupsuboneProductHttp) SearchGroupsuboneProductPage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	// สร้าง filter จาก query parameters
	filter := make(map[string]interface{})

	// เพิ่ม filter สำหรับ GroupMainGuid
	if groupMainGuid := ctx.QueryParam("GroupMainGuid"); groupMainGuid != "" {
		filter["groupMainGuid"] = groupMainGuid
	}

	pageable := utils.GetPageable(ctx.QueryParam)
	docList, pagination, err := h.svc.SearchGroupsuboneProduct(shopID, filter, pageable)
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

// List GroupsuboneProduct godoc
// @Description search limit offset
// @Tags		AICloudProductGroupsubone
// @Param		q	query	string	false  "Search Value"
// @Param		offset	query	integer	false  "offset"
// @Param		limit	query	integer	false  "limit"
// @Param		lang	query	string	false  "lang"
// @Accept 		json
// @Success		200	{array} common.ApiResponse
// @Failure		401 {object} common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/groupsubone/list [get]
func (h GroupsuboneProductHttp) SearchGroupsuboneProductStep(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageableStep := utils.GetPageableStep(ctx.QueryParam)
	lang := ctx.QueryParam("lang")
	docList, total, err := h.svc.SearchGroupsuboneProductStep(shopID, lang, pageableStep)
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

// Create GroupsuboneProduct Bulk godoc
// @Description Create GroupsuboneProduct
// @Tags		AICloudProductGroupsubone
// @Param		GroupsuboneProduct  body      []models.GroupsuboneProduct  true  "GroupsuboneProduct"
// @Accept 		json
// @Success		201	{object} common.BulkResponse
// @Failure		401 {object} common.AuthResponseFailed
// @Security     AccessToken
// @Router /aicloud/groupsubone/bulk [post]
func (h GroupsuboneProductHttp) SaveBulk(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	input := ctx.ReadInput()
	dataReq := []models.GroupsuboneProduct{}
	if err := json.Unmarshal([]byte(input), &dataReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	bulkResponse, err := h.svc.SaveInBatch(shopID, authUsername, dataReq)
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
