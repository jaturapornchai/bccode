package branch

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	orgaccess "smlcloudplatform/internal/organization"
	"smlcloudplatform/internal/shop/branch/models"
	"smlcloudplatform/internal/shop/branch/repositories"
	"smlcloudplatform/internal/shop/branch/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/apperr"
	"smlcloudplatform/pkg/microservice"
)

type IBranchHttp interface{}

type BranchHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IBranchHttpService
}

func NewBranchHttp(ms *microservice.Microservice, cfg config.IConfig) BranchHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())

	repo := repositories.NewBranchRepository(pst)

	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewBranchHttpService(repo, masterSyncCacheRepo)

	return BranchHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h BranchHttp) RegisterHttp() {

	h.ms.GET("/holding/branch", h.SearchBranchPage)
	h.ms.GET("/shop/branch", h.SearchBranchPage)
	h.ms.GET("/holding/branch/list", h.SearchBranchStep)
	h.ms.GET("/shop/branch/list", h.SearchBranchStep)
	h.ms.POST("/holding/branch", h.CreateBranch)
	h.ms.POST("/shop/branch", h.CreateBranch)
	h.ms.GET("/holding/branch/:id", h.InfoBranch)
	h.ms.GET("/shop/branch/:id", h.InfoBranch)
	h.ms.PUT("/holding/branch/:id", h.UpdateBranch)
	h.ms.PUT("/shop/branch/:id", h.UpdateBranch)
	h.ms.DELETE("/holding/branch/:id", h.DeleteBranch)
	h.ms.DELETE("/shop/branch/:id", h.DeleteBranch)
	h.ms.DELETE("/holding/branch", h.DeleteBranchByGUIDs)
	h.ms.DELETE("/shop/branch", h.DeleteBranchByGUIDs)
}

// Create Branch godoc
// @Description Create Branch
// @Tags		Branch
// @Param		Branch  body      models.Branch  true  "Branch"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /shop/branch [post]
func (h BranchHttp) CreateBranch(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode
	if authErr := orgaccess.RequireHoldingAdmin(h.ms.MongoPersister(h.cfg.MongoPersisterConfig()), userInfo); authErr != nil {
		return apperr.Respond(ctx, authErr)
	}
	input := ctx.ReadInput()

	docReq := &models.Branch{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.CreateBranch(holdingCode, authUsername, *docReq)

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

// Update Branch godoc
// @Description Update Branch
// @Tags		Branch
// @Param		id  path      string  true  "Branch ID"
// @Param		Branch  body      models.Branch  true  "Branch"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /shop/branch/{id} [put]
func (h BranchHttp) UpdateBranch(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.Branch{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateBranch(holdingCode, id, authUsername, *docReq)

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

// Delete Branch godoc
// @Description Delete Branch
// @Tags		Branch
// @Param		id  path      string  true  "Branch ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /shop/branch/{id} [delete]
func (h BranchHttp) DeleteBranch(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteBranch(holdingCode, id, authUsername)

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

// Delete Branch godoc
// @Description Delete Branch
// @Tags		Branch
// @Param		Branch  body      []string  true  "Branch GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /shop/branch [delete]
func (h BranchHttp) DeleteBranchByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeleteBranchByGUIDs(holdingCode, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get Branch godoc
// @Description get struct array by ID
// @Tags		Branch
// @Param		id  path      string  true  "Branch ID"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /shop/branch/{id} [get]
func (h BranchHttp) InfoBranch(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get Branch %v", id)
	doc, err := h.svc.InfoBranch(holdingCode, id)

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

// List Branch godoc
// @Description get struct array by ID
// @Tags		Branch
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "page"
// @Param		limit	query	integer		false  "limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /shop/branch [get]
func (h BranchHttp) SearchBranchPage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	pageable := utils.GetPageable(ctx.QueryParam)

	docList, pagination, err := h.svc.SearchBranch(holdingCode, map[string]interface{}{}, pageable)

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

// List Branch godoc
// @Description search limit offset
// @Tags		Branch
// @Param		q		query	string		false  "Search Value"
// @Param		offset	query	integer		false  "offset"
// @Param		limit	query	integer		false  "limit"
// @Param		lang	query	string		false  "lang"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /shop/branch/list [get]
func (h BranchHttp) SearchBranchStep(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	pageableStep := utils.GetPageableStep(ctx.QueryParam)

	lang := ctx.QueryParam("lang")

	docList, total, err := h.svc.SearchBranchStep(holdingCode, lang, pageableStep)

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
