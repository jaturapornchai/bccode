package temp

import (
	"net/http"
	"smlcloudplatform/internal/config"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/pos/temp/repositories"
	"smlcloudplatform/internal/pos/temp/services"
	"strings"

	"smlcloudplatform/pkg/microservice"
)

type IPOSTempHttp interface{}

type POSTempHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IPOSTempService
}

func NewPOSTempHttp(ms *microservice.Microservice, cfg config.IConfig) POSTempHttp {
	cache := ms.Cacher(cfg.CacherConfig())

	repo := repositories.NewCacheRepository(cache)

	svc := services.NewPOSTempService(repo)

	return POSTempHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h POSTempHttp) RegisterHttp() {

	h.ms.GET("/pos/temp", h.InfoPOSTemp)
	h.ms.POST("/pos/temp", h.CreatePOSTemp)
	h.ms.DELETE("/pos/temp", h.DeletePOSTemp)
}

// Create POSTemp godoc
// @Description Create POSTemp
// @Tags		POSTemp
// @Param		POSTemp  body      string  true  "pos temp data"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /pos/temp [post]
func (h POSTempHttp) CreatePOSTemp(ctx microservice.IContext) error {

	shopID := ctx.UserInfo().ShopID
	branchCode := strings.Trim(ctx.QueryParam("branch-code"), " ")

	if branchCode == "" {
		ctx.ResponseError(http.StatusBadRequest, "branch-code is required")
		return nil
	}

	input := ctx.ReadInput()

	err := h.svc.SaveTemp(shopID, branchCode, input)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
	})
	return nil
}

// Get POSTemp godoc
// @Description Get POSTemp
// @Tags		POSTemp
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /pos/temp [get]
func (h POSTempHttp) InfoPOSTemp(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	branchCode := strings.Trim(ctx.QueryParam("branch-code"), " ")

	if branchCode == "" {
		ctx.ResponseError(http.StatusBadRequest, "branch-code is required")
		return nil
	}

	result, err := h.svc.InfoTemp(shopID, branchCode)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
		Data:    result,
	})

	return nil
}

// Delete POSTemp godoc
// @Description Delete POSTemp
// @Tags		POSTemp
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /pos/temp [delete]
func (h POSTempHttp) DeletePOSTemp(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	branchCode := strings.Trim(ctx.QueryParam("branch-code"), " ")

	if branchCode == "" {
		ctx.ResponseError(http.StatusBadRequest, "branch-code is required")
		return nil
	}

	err := h.svc.DeleteTemp(shopID, branchCode)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}
