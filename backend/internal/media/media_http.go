package media

import (
	"net/http"
	"smlcloudplatform/internal/config"
	"smlcloudplatform/pkg/microservice"

	common "smlcloudplatform/internal/models"
)

type IMediaHTTP interface {
	RegisterHttp()
	UploadVideo(ctx microservice.IContext) error
	UploadImage(ctx microservice.IContext) error
}

type MediaHttp struct {
	ms      *microservice.Microservice
	cfg     config.IConfig
	service IMediaService
}

func InitMediaUploadHttp(ms *microservice.Microservice, cfg config.IConfig) IMediaHTTP {

	return NewMediaUploadHttp(ms, cfg, InitMediaService())
}

func NewMediaUploadHttp(ms *microservice.Microservice, cfg config.IConfig, service IMediaService) IMediaHTTP {
	return &MediaHttp{
		ms:      ms,
		cfg:     cfg,
		service: service,
	}
}

func (svc MediaHttp) RegisterHttp() {
	svc.ms.POST("/media/upload/video", svc.UploadVideo)
	svc.ms.POST("/media/upload/image", svc.UploadVideo)

}

func (svc MediaHttp) UploadVideo(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	shopId := userInfo.ShopID
	fileHeader, _ := ctx.FormFile("file")

	media, err := svc.service.UploadVideo(shopId, fileHeader)

	if err != nil {
		ctx.Response(http.StatusBadRequest, &common.ApiResponse{
			Success: false,
			Message: err.Error(),
		})
		return nil
	}

	ctx.Response(http.StatusOK, &common.ApiResponse{
		Success: true,
		Data:    media,
	})
	return nil
}

func (svc MediaHttp) UploadImage(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	shopId := userInfo.ShopID
	fileHeader, _ := ctx.FormFile("file")

	media, err := svc.service.UploadImage(shopId, fileHeader)

	if err != nil {
		ctx.Response(http.StatusBadRequest, &common.ApiResponse{
			Success: false,
			Message: err.Error(),
		})
		return nil
	}

	ctx.Response(http.StatusOK, &common.ApiResponse{
		Success: true,
		Data:    media,
	})
	return nil
}
