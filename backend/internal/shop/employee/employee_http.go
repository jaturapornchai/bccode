package employee

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/shop/employee/models"
	"smlcloudplatform/internal/shop/employee/repositories"
	"smlcloudplatform/internal/shop/employee/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
)

type IEmployeeHttp interface{}

type EmployeeHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IEmployeeHttpService
}

func NewEmployeeHttp(ms *microservice.Microservice, cfg config.IConfig) EmployeeHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())

	repo := repositories.NewEmployeeRepository(pst)

	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewEmployeeHttpService(repo, masterSyncCacheRepo, utils.HashPassword)

	return EmployeeHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h EmployeeHttp) RegisterHttp() {

	h.ms.GET("/holding/employee", h.SearchEmployeePage)
	h.ms.GET("/shop/employee", h.SearchEmployeePage)
	h.ms.GET("/holding/employee/list", h.SearchEmployeeStep)
	h.ms.GET("/shop/employee/list", h.SearchEmployeeStep)
	h.ms.POST("/holding/employee", h.CreateEmployee)
	h.ms.POST("/shop/employee", h.CreateEmployee)
	h.ms.GET("/holding/employee/:id", h.InfoEmployee)
	h.ms.GET("/shop/employee/:id", h.InfoEmployee)
	h.ms.GET("/holding/employee/code/:code", h.InfoEmployeeByCode)
	h.ms.GET("/shop/employee/code/:code", h.InfoEmployeeByCode)
	h.ms.GET("/holding/employee/email/:email", h.InfoEmployeeByEmail)
	h.ms.GET("/shop/employee/email/:email", h.InfoEmployeeByEmail)
	h.ms.PUT("/holding/employee/:id", h.UpdateEmployee)
	h.ms.PUT("/shop/employee/:id", h.UpdateEmployee)
	h.ms.PUT("/holding/employee/password", h.UpdatePassword)
	h.ms.PUT("/shop/employee/password", h.UpdatePassword)
	h.ms.DELETE("/holding/employee/:id", h.DeleteEmployee)
	h.ms.DELETE("/shop/employee/:id", h.DeleteEmployee)
	h.ms.DELETE("/holding/employee", h.DeleteEmployeeByGUIDs)
	h.ms.DELETE("/shop/employee", h.DeleteEmployeeByGUIDs)
}

// Create Employee godoc
// @Description Create Employee
// @Tags		Employee
// @Param		Employee  body      models.EmployeeRequestRegister  true  "Employee"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /shop/employee [post]
func (h EmployeeHttp) CreateEmployee(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	holdingCode := ctx.UserInfo().HoldingCode
	input := ctx.ReadInput()

	docReq := &models.EmployeeRequestRegister{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.CreateEmployee(holdingCode, authUsername, *docReq)

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

// Update Employee godoc
// @Description Update Employee
// @Tags		Employee
// @Param		id  path      string  true  "Employee ID"
// @Param		Employee  body      models.EmployeeRequestUpdate  true  "Employee"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /shop/employee/{id} [put]
func (h EmployeeHttp) UpdateEmployee(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.EmployeeRequestUpdate{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateEmployee(holdingCode, id, authUsername, *docReq)

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

// Delete Employee godoc
// @Description Delete Employee
// @Tags		Employee
// @Param		id  path      string  true  "Employee ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /shop/employee/{id} [delete]
func (h EmployeeHttp) DeleteEmployee(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteEmployee(holdingCode, id, authUsername)

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

// Delete Employee godoc
// @Description Delete Employee
// @Tags		Employee
// @Param		Employee  body      []string  true  "Employee GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /shop/employee [delete]
func (h EmployeeHttp) DeleteEmployeeByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeleteEmployeeByGUIDs(holdingCode, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get Employee godoc
// @Description get struct array by ID
// @Tags		Employee
// @Param		id  path      string  true  "Employee ID"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /shop/employee/{id} [get]
func (h EmployeeHttp) InfoEmployee(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get Employee %v", id)
	doc, err := h.svc.InfoEmployee(holdingCode, id)

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

// Get Employee By Code godoc
// @Description get employee by code
// @Tags		Employee
// @Param		code  path      string  true  "Employee code"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /shop/employee/code/{code} [get]
func (h EmployeeHttp) InfoEmployeeByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	code := ctx.Param("code")

	doc, err := h.svc.InfoEmployeeByCode(holdingCode, code)

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

// Get Employee By Email godoc
// @Description get employee by email
// @Tags		Employee
// @Param		email  path      string  true  "Employee email"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /shop/employee/email/{email} [get]
func (h EmployeeHttp) InfoEmployeeByEmail(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	email := ctx.Param("email")

	doc, err := h.svc.InfoEmployeeByEmail(holdingCode, email)

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

// List Employee godoc
// @Description get struct array by ID
// @Tags		Employee
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "page"
// @Param		limit	query	integer		false  "limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /shop/employee [get]
func (h EmployeeHttp) SearchEmployeePage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	pageable := utils.GetPageable(ctx.QueryParam)

	docList, pagination, err := h.svc.SearchEmployee(holdingCode, map[string]interface{}{}, pageable)

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

// List Employee godoc
// @Description search limit offset
// @Tags		Employee
// @Param		q		query	string		false  "Search Value"
// @Param		offset	query	integer		false  "offset"
// @Param		limit	query	integer		false  "limit"
// @Param		lang	query	string		false  "lang"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /shop/employee/list [get]
func (h EmployeeHttp) SearchEmployeeStep(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	pageableStep := utils.GetPageableStep(ctx.QueryParam)

	// Set default limit to 1000 if not provided
	if ctx.QueryParam("limit") == "" {
		pageableStep.Limit = 1000
	}

	lang := ctx.QueryParam("lang")

	docList, total, err := h.svc.SearchEmployeeStep(holdingCode, lang, pageableStep)

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

// Update Password Employee godoc
// @Summary		Update Password Employee
// @Description	Update Password Employee
// @Tags		Employee
// @Param		id  path      string  true  "Employee ID"
// @Param		Employee  body      models.EmployeeRequestPassword  true  "Register Employee"
// @Success		200	{object}	models.ResponseSuccess
// @Failure		400 {object}	models.AuthResponseFailed
// @Accept 		json
// @Security     AccessToken
// @Router		/employee/password [put]
func (h EmployeeHttp) UpdatePassword(ctx microservice.IContext) error {
	userAuthInfo := ctx.UserInfo()
	authUsername := userAuthInfo.Username
	holdingCode := userAuthInfo.HoldingCode

	input := ctx.ReadInput()

	userPwdReq := models.EmployeeRequestPassword{}
	err := json.Unmarshal([]byte(input), &userPwdReq)

	if err != nil {
		ctx.ResponseError(400, "user payload invalid")
		return err
	}

	err = h.svc.UpdatePassword(holdingCode, authUsername, userPwdReq)

	if err != nil {
		ctx.Response(http.StatusBadRequest, common.ApiResponse{
			Success: false,
			Message: err.Error(),
		})
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
	})

	return nil
}
