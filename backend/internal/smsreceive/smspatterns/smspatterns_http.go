package smspatterns

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/smsreceive/smspatterns/models"
	"smlcloudplatform/internal/smsreceive/smspatterns/repositories"
	"smlcloudplatform/internal/smsreceive/smspatterns/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
)

type ISmsPatternsHttp interface{}

type SmsPatternsHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.ISmsPatternsHttpService
}

func NewSmsPatternsHttp(ms *microservice.Microservice, cfg config.IConfig) SmsPatternsHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())

	repo := repositories.NewSmsPatternsRepository(pst)

	svc := services.NewSmsPatternsHttpService(repo)

	return SmsPatternsHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h SmsPatternsHttp) RegisterHttp() {

	h.ms.GET("/smspatterns", h.SearchSmsPatterns)
	h.ms.POST("/smspatterns", h.CreateSmsPatterns)
	h.ms.GET("/smspatterns/:id", h.InfoSmsPatterns)
	h.ms.PUT("/smspatterns/:id", h.UpdateSmsPatterns)
	h.ms.DELETE("/smspatterns/:id", h.DeleteSmsPatterns)
}

func (h SmsPatternsHttp) CreateSmsPatterns(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	input := ctx.ReadInput()

	docReq := &models.SmsPatterns{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.CreateSmsPatterns(authUsername, *docReq)

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

func (h SmsPatternsHttp) UpdateSmsPatterns(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.SmsPatterns{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateSmsPatterns(id, authUsername, *docReq)

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

func (h SmsPatternsHttp) DeleteSmsPatterns(ctx microservice.IContext) error {

	id := ctx.Param("id")

	err := h.svc.DeleteSmsPatterns(id)

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

func (h SmsPatternsHttp) InfoSmsPatterns(ctx microservice.IContext) error {

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get SmsPatterns %v", id)
	doc, err := h.svc.InfoSmsPatterns(id)

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

func (h SmsPatternsHttp) SearchSmsPatterns(ctx microservice.IContext) error {

	pageable := utils.GetPageable(ctx.QueryParam)
	docList, pagination, err := h.svc.SearchSmsPatterns(pageable)

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
