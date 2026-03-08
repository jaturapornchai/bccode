package ocr

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/pkg/microservice"
)

type IOcrHttp interface{}

type OcrHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc IOcrService
}

func NewOcrHttp(ms *microservice.Microservice, cfg config.IConfig) OcrHttp {

	svc := NewOcrService()
	return OcrHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h OcrHttp) RegisterHttp() {
	h.ms.POST("/ocr/upload", h.OcrUpload)
	h.ms.POST("/ocr/result", h.OcrResault)
}

// Upload Ocr godoc
// @Summary		Upload Ocr
// @Description Upload Ocr
// @Tags		OCR
// @Param		OcrRequest body      OcrRequest  true  "Ocr Request"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /ocr/upload [post]
func (h OcrHttp) OcrUpload(ctx microservice.IContext) error {
	input := ctx.ReadInput()

	docReq := &OcrRequest{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	result, err := h.svc.UploadOcr(docReq.ResourceKey, docReq.UrlResources)

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

// Result Ocr godoc
// @Summary		Result Ocr
// @Description Result Ocr
// @Tags		OCR
// @Param		OcrRequest body      OcrRequest  true  "Ocr Request"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /ocr/result [post]
func (h OcrHttp) OcrResault(ctx microservice.IContext) error {

	input := ctx.ReadInput()

	docReq := &OcrRequest{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	result, err := h.svc.ResultOcr(docReq.ResourceKey, docReq.UrlResources)

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
