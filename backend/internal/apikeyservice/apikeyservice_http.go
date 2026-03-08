package apikeyservice

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/apikeyservice/models"
	"smlcloudplatform/internal/apikeyservice/services"
	"smlcloudplatform/internal/config"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/pkg/microservice"
	"time"
)

type IApiKeyServiceHttp interface{}

type ApiKeyServiceHttp struct {
	ms          *microservice.Microservice
	cfg         config.IConfig
	svc         services.IApiKeyServiceHttpService
	authService microservice.IAuthService
}

func NewApiKeyServiceHttp(ms *microservice.Microservice, cfg config.IConfig) ApiKeyServiceHttp {

	authService := microservice.NewAuthService(ms.Cacher(cfg.CacherConfig()), 24*3*time.Hour, 24*30*time.Hour)
	svc := services.NewApiKeyServiceHttpService(authService)

	return ApiKeyServiceHttp{
		ms:          ms,
		cfg:         cfg,
		svc:         svc,
		authService: authService,
	}
}

func (h ApiKeyServiceHttp) RegisterHttp() {

	h.ms.POST("/apikeyservice", h.CreateApiKey)
	h.ms.DELETE("/apikeyservice", h.RemoveApiKey)
}

// X Api key generate
// @Description generate x-api-key
// @Tags		XApiKey
// @Accept 		json
// @Success		200	{object}	common.AuthResponse
// @Failure		400 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /apikeyservice [post]
func (h ApiKeyServiceHttp) CreateApiKey(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()

	token, err := h.svc.CreateApiKeyService(userInfo, time.Duration(24*180)*time.Hour)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.AuthResponse{
		Success: true,
		Token:   token,
	})
	return nil
}

// X Api key generate
// @Description delete x-api-key
// @Tags		XApiKey
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		400 {object}	common.AuthResponseFailed
// @Router /apikeyservice [delete]
func (h ApiKeyServiceHttp) RemoveApiKey(ctx microservice.IContext) error {

	input := ctx.ReadInput()

	req := &models.ApiKeyRequest{}
	err := json.Unmarshal([]byte(input), &req)

	if err != nil {
		ctx.ResponseError(400, "user payload invalid")
		return err
	}

	err = h.svc.RemoveApiKey(req.ApiKey)

	if err != nil {
		ctx.Response(http.StatusBadRequest, common.ApiResponse{
			Success: false,
			Message: err.Error(),
		})
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}
