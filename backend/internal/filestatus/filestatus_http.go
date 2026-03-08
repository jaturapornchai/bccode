package filestatus

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	"smlcloudplatform/internal/filestatus/models"
	"smlcloudplatform/internal/filestatus/repositories"
	"smlcloudplatform/internal/filestatus/services"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/requestfilter"
	"smlcloudplatform/pkg/microservice"
	"time"
)

type IFileStatusHttp interface{}

type FileStatusHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IFileStatusHttpService
}

func NewFileStatusHttp(ms *microservice.Microservice, cfg config.IConfig) FileStatusHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	repo := repositories.NewFileStatusRepository(pst)

	svc := services.NewFileStatusHttpService(repo, 15*time.Second)

	return FileStatusHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h FileStatusHttp) RegisterHttp() {

	h.ms.GET("/file-status", h.SearchFileStatusPage)
	h.ms.GET("/file-status/list", h.SearchFileStatusStep)
	h.ms.POST("/file-status", h.CreateFileStatus)
	h.ms.PUT("/file-status/:id", h.UpdateFileStatus)
	h.ms.GET("/file-status/:id", h.InfoFileStatus)
	h.ms.DELETE("/file-status/:id", h.DeleteFileStatus)
	h.ms.DELETE("/file-status/menu/:menu", h.DeleteFileStatusByMenu)
	h.ms.DELETE("/file-status", h.DeleteFileStatusByGUIDs)
}

// Create FileStatus godoc
// @Description Create FileStatus
// @Tags		FileStatus
// @Param		FileStatus  body      models.FileStatus  true  "FileStatus"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /file-status [post]
func (h FileStatusHttp) CreateFileStatus(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := &models.FileStatus{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.CreateFileStatus(shopID, authUsername, *docReq)

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

// Update FileStatus godoc
// @Description Update FileStatus
// @Tags		FileStatus
// @Param		id  path      string  true  "FileStatus ID"
// @Param		FileStatus  body      models.FileStatus  true  "FileStatus"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /file-status/{id} [put]
func (h FileStatusHttp) UpdateFileStatus(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.FileStatus{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateFileStatus(shopID, id, authUsername, *docReq)

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

// Delete FileStatus godoc
// @Description Delete FileStatus
// @Tags		FileStatus
// @Param		id  path      string  true  "FileStatus ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /file-status/{id} [delete]
func (h FileStatusHttp) DeleteFileStatus(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteFileStatus(shopID, id, authUsername)

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

// Delete FileStatus Menu godoc
// @Description Delete FileStatus Menu
// @Tags		FileStatus
// @Param		menu  path      string  true  "Menu ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /file-status/menu/{menu} [delete]
func (h FileStatusHttp) DeleteFileStatusByMenu(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	menu := ctx.Param("menu")

	err := h.svc.DeleteFileStatusByMenu(shopID, authUsername, menu)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Delete FileStatus godoc
// @Description Delete FileStatus
// @Tags		FileStatus
// @Param		FileStatus  body      []string  true  "FileStatus GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /file-status [delete]
func (h FileStatusHttp) DeleteFileStatusByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeleteFileStatusByGUIDs(shopID, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get FileStatus godoc
// @Description get FileStatus info by guidfixed
// @Tags		FileStatus
// @Param		id  path      string  true  "FileStatus guidfixed"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /file-status/{id} [get]
func (h FileStatusHttp) InfoFileStatus(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get FileStatus %v", id)
	doc, err := h.svc.InfoFileStatus(shopID, id)

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

// List FileStatus step godoc
// @Description get list step
// @Tags		FileStatus
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /file-status [get]
func (h FileStatusHttp) SearchFileStatusPage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	pageable := utils.GetPageable(ctx.QueryParam)

	filters := requestfilter.GenerateFilters(ctx.QueryParam, []requestfilter.FilterRequest{
		{
			Param: "menu",
			Field: "menu",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "jobid",
			Field: "jobid",
			Type:  requestfilter.FieldTypeString,
		},
	})

	filters["username"] = authUsername

	docList, pagination, err := h.svc.SearchFileStatus(shopID, filters, pageable)

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

// List FileStatus godoc
// @Description search limit offset
// @Tags		FileStatus
// @Param		q		query	string		false  "Search Value"
// @Param		offset	query	integer		false  "offset"
// @Param		limit	query	integer		false  "limit"
// @Param		lang	query	string		false  "lang"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /file-status/list [get]
func (h FileStatusHttp) SearchFileStatusStep(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	pageableStep := utils.GetPageableStep(ctx.QueryParam)

	lang := ctx.QueryParam("lang")

	filters := requestfilter.GenerateFilters(ctx.QueryParam, []requestfilter.FilterRequest{
		{
			Param: "menu",
			Field: "menu",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "jobid",
			Field: "jobid",
			Type:  requestfilter.FieldTypeString,
		},
	})

	filters["username"] = authUsername

	docList, total, err := h.svc.SearchFileStatusStep(shopID, lang, filters, pageableStep)

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
