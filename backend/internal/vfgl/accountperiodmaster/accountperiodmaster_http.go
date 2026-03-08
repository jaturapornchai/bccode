package accountperiodmaster

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/vfgl/accountperiodmaster/models"
	"smlcloudplatform/internal/vfgl/accountperiodmaster/repositories"
	"smlcloudplatform/internal/vfgl/accountperiodmaster/services"
	"smlcloudplatform/pkg/microservice"
	"strings"
	"time"
)

type IAccountPeriodMasterHttp interface{}

type AccountPeriodMasterHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IAccountPeriodMasterHttpService
}

func NewAccountPeriodMasterHttp(ms *microservice.Microservice, cfg config.IConfig) AccountPeriodMasterHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())

	repo := repositories.NewAccountPeriodMasterRepository(pst)

	svc := services.NewAccountPeriodMasterHttpService(repo)

	return AccountPeriodMasterHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h AccountPeriodMasterHttp) RegisterHttp() {

	h.ms.GET("/gl/accountperiodmaster", h.SearchAccountPeriodMasterPage)
	h.ms.GET("/gl/accountperiodmaster/list", h.SearchAccountPeriodMasterLimit)
	h.ms.POST("/gl/accountperiodmaster", h.CreateAccountPeriodMaster)
	h.ms.POST("/gl/accountperiodmaster/bulk", h.SaveBulkAccountPeriodMaster)
	h.ms.GET("/gl/accountperiodmaster/:id", h.InfoAccountPeriodMaster)
	h.ms.GET("/gl/accountperiodmaster/by-date", h.InfoAccountPeriodMasterByDate)
	h.ms.PUT("/gl/accountperiodmaster/:id", h.UpdateAccountPeriodMaster)
	h.ms.DELETE("/gl/accountperiodmaster/:id", h.DeleteAccountPeriodMaster)
	h.ms.DELETE("/gl/accountperiodmaster", h.DeleteAccountPeriodMasterByGUIDs)
}

// Create AccountPeriodMaster godoc
// @Description Create AccountPeriodMaster
// @Tags		AccountPeriodMaster
// @Param		AccountPeriodMaster  body      models.AccountPeriodMaster true  "AccountPeriodMaster"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /gl/accountperiodmaster [post]
func (h AccountPeriodMasterHttp) CreateAccountPeriodMaster(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := &models.AccountPeriodMasterRequest{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.CreateAccountPeriodMaster(shopID, authUsername, docReq.ToAccountPeriodMaster())

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

// Update AccountPeriodMaster godoc
// @Description Update AccountPeriodMaster
// @Tags		AccountPeriodMaster
// @Param		id  path      string  true  "AccountPeriodMaster ID"
// @Param		AccountPeriodMaster  body      models.AccountPeriodMaster true  "AccountPeriodMaster"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /gl/accountperiodmaster/{id} [put]
func (h AccountPeriodMasterHttp) UpdateAccountPeriodMaster(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.AccountPeriodMasterRequest{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateAccountPeriodMaster(shopID, id, authUsername, docReq.ToAccountPeriodMaster())

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

// Delete AccountPeriodMaster godoc
// @Description Delete AccountPeriodMaster
// @Tags		AccountPeriodMaster
// @Param		id  path      string  true  "AccountPeriodMaster ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /gl/accountperiodmaster/{id} [delete]
func (h AccountPeriodMasterHttp) DeleteAccountPeriodMaster(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteAccountPeriodMaster(shopID, id, authUsername)

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

// Delete AccountPeriodMaster godoc
// @Description Delete AccountPeriodMaster
// @Tags		AccountPeriodMaster
// @Param		AccountPeriodMaster  body      []string  true  "AccountPeriodMaster GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /gl/accountperiodmaster [delete]
func (h AccountPeriodMasterHttp) DeleteAccountPeriodMasterByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeleteAccountPeriodMasterByGUIDs(shopID, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get AccountPeriodMaster godoc
// @Description get struct array by ID
// @Tags		AccountPeriodMaster
// @Param		id  path      string  true  "AccountPeriodMaster ID"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /gl/accountperiodmaster/{id} [get]
func (h AccountPeriodMasterHttp) InfoAccountPeriodMaster(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get AccountPeriodMaster %v", id)
	doc, err := h.svc.InfoAccountPeriodMaster(shopID, id)

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

// Get AccountPeriodMaster godoc
// @Description Get AccountPeriodMaster by date
// @Tags		AccountPeriodMaster
// @Param		date-list		query	string		false  "date-list"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /gl/accountperiodmaster/by-date [get]
func (h AccountPeriodMasterHttp) InfoAccountPeriodMasterByDate(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	layout := "2006-01-02"
	reqDateList := ctx.QueryParam("date-list")

	dateListStr := strings.Split(reqDateList, ",")

	dateList := []time.Time{}

	for _, dateStr := range dateListStr {

		dateStr = strings.Trim(dateStr, " ")
		if len(dateStr) < 1 {
			continue
		}

		tempDate, err := time.Parse(layout, dateStr)
		if err != nil {
			continue
		}

		dateList = append(dateList, tempDate)
	}

	doc, err := h.svc.InfoAccountPeriodMasterByDateList(shopID, dateList)

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

// List AccountPeriodMaster godoc
// @Description get struct array by ID
// @Tags		AccountPeriodMaster
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /gl/accountperiodmaster [get]
func (h AccountPeriodMasterHttp) SearchAccountPeriodMasterPage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)
	docList, pagination, err := h.svc.SearchAccountPeriodMaster(shopID, pageable)

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

// List AccountPeriodMaster godoc
// @Description search limit offset
// @Tags		AccountPeriodMaster
// @Param		q		query	string		false  "Search Value"
// @Param		offset	query	integer		false  "offset"
// @Param		limit	query	integer		false  "limit"
// @Param		lang	query	string		false  "lang"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /gl/accountperiodmaster/list [get]
func (h AccountPeriodMasterHttp) SearchAccountPeriodMasterLimit(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageableStep := utils.GetPageableStep(ctx.QueryParam)

	lang := ctx.QueryParam("lang")

	docList, total, err := h.svc.SearchAccountPeriodMasterStep(shopID, lang, pageableStep)

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

// Bulk Create AccountPeriodMaster godoc
// @Description Bulk Create AccountPeriodMaster
// @Tags		AccountPeriodMaster
// @Param		AccountPeriodMaster  body      []models.AccountPeriodMaster true  "Array AccountPeriodMaster"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /gl/accountperiodmaster/bulk [post]
func (h AccountPeriodMasterHttp) SaveBulkAccountPeriodMaster(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := &[]models.AccountPeriodMasterRequest{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	tempDocListReq := []models.AccountPeriodMaster{}
	for _, doc := range *docReq {
		tempDocListReq = append(tempDocListReq, doc.ToAccountPeriodMaster())
	}

	err = h.svc.SaveInBatch(shopID, authUsername, tempDocListReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
	})
	return nil
}
