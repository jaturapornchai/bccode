package department

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/organization/department/models"
	"smlcloudplatform/internal/organization/department/repositories"
	"smlcloudplatform/internal/organization/department/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
)

type IDepartmentHttp interface{}

type DepartmentHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IDepartmentHttpService
}

func NewDepartmentHttp(ms *microservice.Microservice, cfg config.IConfig) DepartmentHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())

	repo := repositories.NewDepartmentRepository(pst)

	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewDepartmentHttpService(repo, masterSyncCacheRepo)

	return DepartmentHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h DepartmentHttp) RegisterHttp() {

	h.ms.POST("/organization/department/bulk", h.SaveBulk)

	h.ms.GET("/organization/department", h.SearchDepartmentPage)
	h.ms.GET("/organization/department/list", h.SearchDepartmentStep)
	h.ms.POST("/organization/department", h.CreateDepartment)
	h.ms.GET("/organization/department/:id", h.InfoDepartment)
	h.ms.GET("/organization/department/:departmentCode/branch/:branchCode", h.InfoDepartmentByCode)
	h.ms.PUT("/organization/department/:id", h.UpdateDepartment)
	h.ms.DELETE("/organization/department/:id", h.DeleteDepartment)
	h.ms.DELETE("/organization/department", h.DeleteDepartmentByGUIDs)
}

// Create Department godoc
// @Description Create Department
// @Tags		Department
// @Param		Department  body      models.Department  true  "Department"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /organization/department [post]
func (h DepartmentHttp) CreateDepartment(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := &models.Department{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.CreateDepartment(shopID, authUsername, *docReq)

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

// Update Department godoc
// @Description Update Department
// @Tags		Department
// @Param		id  path      string  true  "Department ID"
// @Param		Department  body      models.Department  true  "Department"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /organization/department/{id} [put]
func (h DepartmentHttp) UpdateDepartment(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.Department{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateDepartment(shopID, id, authUsername, *docReq)

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

// Delete Department godoc
// @Description Delete Department
// @Tags		Department
// @Param		id  path      string  true  "Department ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /organization/department/{id} [delete]
func (h DepartmentHttp) DeleteDepartment(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteDepartment(shopID, id, authUsername)

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

// Delete Department godoc
// @Description Delete Department
// @Tags		Department
// @Param		Department  body      []string  true  "Department GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /organization/department [delete]
func (h DepartmentHttp) DeleteDepartmentByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeleteDepartmentByGUIDs(shopID, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get Department godoc
// @Description get Department info by guidfixed
// @Tags		Department
// @Param		id  path      string  true  "Department guidfixed"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /organization/department/{id} [get]
func (h DepartmentHttp) InfoDepartment(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get Department %v", id)
	doc, err := h.svc.InfoDepartment(shopID, id)

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

// Get Department By Code godoc
// @Description get Department info by Code
// @Tags		Department
// @Param		departmentCode  path      string  true  "Department Code"
// @Param		branchCode  path      string  true  "Branch Code"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /organization/department/{departmentCode}/branch/{branchCode} [get]
func (h DepartmentHttp) InfoDepartmentByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	branchCode := ctx.Param("branchCode")
	departmentCode := ctx.Param("departmentCode")

	doc, err := h.svc.InfoDepartmentByCode(shopID, branchCode, departmentCode)

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

// List Department step godoc
// @Description get list step
// @Tags		Department
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /organization/department [get]
func (h DepartmentHttp) SearchDepartmentPage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)

	docList, pagination, err := h.svc.SearchDepartment(shopID, map[string]interface{}{}, pageable)

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

// List Department godoc
// @Description search limit offset
// @Tags		Department
// @Param		q		query	string		false  "Search Value"
// @Param		offset	query	integer		false  "offset"
// @Param		limit	query	integer		false  "limit"
// @Param		lang	query	string		false  "lang"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /organization/department/list [get]
func (h DepartmentHttp) SearchDepartmentStep(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageableStep := utils.GetPageableStep(ctx.QueryParam)

	lang := ctx.QueryParam("lang")

	docList, total, err := h.svc.SearchDepartmentStep(shopID, lang, pageableStep)

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

// Create Department Bulk godoc
// @Description Create Department
// @Tags		Department
// @Param		Department  body      []models.Department  true  "Department"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /organization/department/bulk [post]
func (h DepartmentHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	input := ctx.ReadInput()

	dataReq := []models.Department{}
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
