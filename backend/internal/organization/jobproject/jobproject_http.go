package jobproject

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/organization/jobproject/models"
	"smlcloudplatform/internal/organization/jobproject/repositories"
	"smlcloudplatform/internal/organization/jobproject/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
)

type IJobProjectHttp interface{}

type JobProjectHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IJobProjectHttpService
}

func NewJobProjectHttp(ms *microservice.Microservice, cfg config.IConfig) JobProjectHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())
	prod := ms.Producer(cfg.MQConfig())

	repo := repositories.NewJobProjectRepository(pst)
	repoMessageQueue := repositories.NewJobProjectMessageQueueRepository(prod)

	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewJobProjectHttpService(repo, repoMessageQueue, masterSyncCacheRepo)

	return JobProjectHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h JobProjectHttp) RegisterHttp() {

	h.ms.POST("/organization/jobproject/bulk", h.SaveBulk)

	h.ms.GET("/organization/jobproject", h.SearchJobProjectPage)
	h.ms.GET("/organization/jobproject/list", h.SearchJobProjectStep)
	h.ms.POST("/organization/jobproject", h.CreateJobProject)
	h.ms.GET("/organization/jobproject/:id", h.InfoJobProject)
	h.ms.GET("/organization/jobproject/:jobProjectCode/branch/:branchCode", h.InfoJobProjectByCode)
	h.ms.PUT("/organization/jobproject/:id", h.UpdateJobProject)
	h.ms.DELETE("/organization/jobproject/:id", h.DeleteJobProject)
	h.ms.DELETE("/organization/jobproject", h.DeleteJobProjectByGUIDs)
}

// Create JobProject godoc
// @Description Create JobProject
// @Tags		JobProject
// @Param		JobProject  body      models.JobProject  true  "JobProject"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /organization/jobproject [post]
func (h JobProjectHttp) CreateJobProject(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := &models.JobProject{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.CreateJobProject(shopID, authUsername, *docReq)

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

// Update JobProject godoc
// @Description Update JobProject
// @Tags		JobProject
// @Param		id  path      string  true  "JobProject ID"
// @Param		JobProject  body      models.JobProject  true  "JobProject"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /organization/jobproject/{id} [put]
func (h JobProjectHttp) UpdateJobProject(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.JobProject{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateJobProject(shopID, id, authUsername, *docReq)

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

// Delete JobProject godoc
// @Description Delete JobProject
// @Tags		JobProject
// @Param		id  path      string  true  "JobProject ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /organization/jobproject/{id} [delete]
func (h JobProjectHttp) DeleteJobProject(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteJobProject(shopID, id, authUsername)

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

// Delete JobProject godoc
// @Description Delete JobProject
// @Tags		JobProject
// @Param		JobProject  body      []string  true  "JobProject GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /organization/jobproject [delete]
func (h JobProjectHttp) DeleteJobProjectByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeleteJobProjectByGUIDs(shopID, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get JobProject godoc
// @Description get JobProject info by guidfixed
// @Tags		JobProject
// @Param		id  path      string  true  "JobProject guidfixed"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /organization/jobproject/{id} [get]
func (h JobProjectHttp) InfoJobProject(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get JobProject %v", id)
	doc, err := h.svc.InfoJobProject(shopID, id)

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

// Get JobProject By Code godoc
// @Description get JobProject info by Code
// @Tags		JobProject
// @Param		jobProjectCode  path      string  true  "JobProject Code"
// @Param		branchCode  path      string  true  "Branch Code"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /organization/jobproject/{jobProjectCode}/branch/{branchCode} [get]
func (h JobProjectHttp) InfoJobProjectByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	branchCode := ctx.Param("branchCode")
	jobProjectCode := ctx.Param("jobProjectCode")

	doc, err := h.svc.InfoJobProjectByCode(shopID, branchCode, jobProjectCode)

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

// List JobProject step godoc
// @Description get list step
// @Tags		JobProject
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /organization/jobproject [get]
func (h JobProjectHttp) SearchJobProjectPage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)

	docList, pagination, err := h.svc.SearchJobProject(shopID, map[string]interface{}{}, pageable)

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

// List JobProject godoc
// @Description search limit offset
// @Tags		JobProject
// @Param		q		query	string		false  "Search Value"
// @Param		offset	query	integer		false  "offset"
// @Param		limit	query	integer		false  "limit"
// @Param		lang	query	string		false  "lang"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /organization/jobproject/list [get]
func (h JobProjectHttp) SearchJobProjectStep(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageableStep := utils.GetPageableStep(ctx.QueryParam)

	lang := ctx.QueryParam("lang")

	docList, total, err := h.svc.SearchJobProjectStep(shopID, lang, pageableStep)

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

// Create JobProject Bulk godoc
// @Description Create JobProject
// @Tags		JobProject
// @Param		JobProject  body      []models.JobProject  true  "JobProject"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /organization/jobproject/bulk [post]
func (h JobProjectHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	input := ctx.ReadInput()

	dataReq := []models.JobProject{}
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
