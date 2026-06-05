package rfq

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	"strings"

	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	trancache "smlcloudplatform/internal/transaction/repositories"
	"smlcloudplatform/internal/transaction/rfq/models"
	"smlcloudplatform/internal/transaction/rfq/repositories"
	"smlcloudplatform/internal/transaction/rfq/services"
	"smlcloudplatform/internal/transaction/rfq/validators"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/requestfilter"
	"smlcloudplatform/pkg/microservice"
)

type IRFQHttp interface{}

type RFQHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IRFQHttpService
}

func NewRFQHttp(ms *microservice.Microservice, cfg config.IConfig) RFQHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())
	producer := ms.Producer(cfg.MQConfig())

	repo := repositories.NewRFQRepository(pst)
	repoMq := repositories.NewRFQMessageQueueRepository(producer)
	transRepo := trancache.NewCacheRepository(cache)
	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewRFQHttpService(repo, transRepo, repoMq, masterSyncCacheRepo)

	return RFQHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func getRequestLanguage(ctx microservice.IContext) string {
	if lang := ctx.QueryParam("lang"); lang != "" {
		return lang
	}
	if lang := ctx.Header("Accept-Language"); lang != "" {
		return lang
	}
	return "en"
}

func sanitizeEmptyTimeFields(input string) string {
	timeFields := []string{
		"docdatetime", "docrefdate", "taxdocdate",
		"createdat", "modified_at", "docrefdatetime",
	}
	result := input
	for _, field := range timeFields {
		result = strings.ReplaceAll(result,
			`"`+field+`":""`,
			`"`+field+`":"0001-01-01T00:00:00Z"`)
	}
	return result
}

func (h RFQHttp) RegisterHttp() {
	h.ms.POST("/transaction/rfq/bulk", h.SaveBulk)
	h.ms.GET("/transaction/rfq", h.SearchRFQPage)
	h.ms.GET("/transaction/rfq/list", h.SearchRFQStep)
	h.ms.POST("/transaction/rfq", h.CreateRFQ)
	h.ms.GET("/transaction/rfq/:id", h.InfoRFQ)
	h.ms.GET("/transaction/rfq/code/:code", h.InfoRFQByCode)
	h.ms.PUT("/transaction/rfq/:id", h.UpdateRFQ)
	h.ms.DELETE("/transaction/rfq/:id", h.DeleteRFQ)
	h.ms.DELETE("/transaction/rfq", h.DeleteRFQByGUIDs)
}

func (h RFQHttp) CreateRFQ(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	holdingCode := ctx.UserInfo().HoldingCode
	input := ctx.ReadInput()
	lang := getRequestLanguage(ctx)

	docReq := &models.RFQ{}
	err := json.Unmarshal([]byte(sanitizeEmptyTimeFields(input)), &docReq)
	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}
	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, docNo, validationResult, err := h.svc.CreateRFQ(holdingCode, authUsername, *docReq)
	if validationResult != nil && !validationResult.IsValid() {
		ctx.Response(http.StatusBadRequest, validationResult.ToErrorResponse(lang))
		return nil
	}
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, validators.NewCreateSuccessResponse(idx, docNo, lang))
	return nil
}

func (h RFQHttp) UpdateRFQ(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode
	id := ctx.Param("id")
	input := ctx.ReadInput()
	lang := getRequestLanguage(ctx)

	docReq := &models.RFQ{}
	err := json.Unmarshal([]byte(sanitizeEmptyTimeFields(input)), &docReq)
	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}
	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	validationResult, err := h.svc.UpdateRFQ(holdingCode, id, authUsername, *docReq)
	if validationResult != nil && !validationResult.IsValid() {
		ctx.Response(http.StatusBadRequest, validationResult.ToUpdateErrorResponse(lang))
		return nil
	}
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, validators.NewUpdateSuccessResponse(id, lang))
	return nil
}

func (h RFQHttp) DeleteRFQ(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username
	id := ctx.Param("id")

	err := h.svc.DeleteRFQ(holdingCode, id, authUsername)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}
	ctx.Response(http.StatusOK, common.ApiResponse{Success: true, ID: id})
	return nil
}

func (h RFQHttp) DeleteRFQByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeleteRFQByGUIDs(holdingCode, authUsername, docReq)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}
	ctx.Response(http.StatusOK, common.ApiResponse{Success: true})
	return nil
}

func (h RFQHttp) InfoRFQ(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	id := ctx.Param("id")

	doc, err := h.svc.InfoRFQ(holdingCode, id)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}
	ctx.Response(http.StatusOK, common.ApiResponse{Success: true, Data: doc})
	return nil
}

func (h RFQHttp) InfoRFQByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	code := ctx.Param("code")

	doc, err := h.svc.InfoRFQByCode(holdingCode, code)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}
	ctx.Response(http.StatusOK, common.ApiResponse{Success: true, Data: doc})
	return nil
}

func (h RFQHttp) SearchRFQPage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	pageable := utils.GetPageable(ctx.QueryParam)

	filters := requestfilter.GenerateFilters(ctx.QueryParam, []requestfilter.FilterRequest{
		{Param: "refprdocno", Type: requestfilter.FieldTypeString},
		{Param: "selectedvendor", Type: requestfilter.FieldTypeString},
		{Param: "-", Field: "docdatetime", Type: requestfilter.FieldTypeRangeDate},
		{Param: "branchcode", Field: "branch.code", Type: requestfilter.FieldTypeString},
	})

	docList, pagination, err := h.svc.SearchRFQ(holdingCode, filters, pageable)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}
	ctx.Response(http.StatusOK, common.ApiResponse{Success: true, Data: docList, Pagination: pagination})
	return nil
}

func (h RFQHttp) SearchRFQStep(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	pageableStep := utils.GetPageableStep(ctx.QueryParam)
	lang := ctx.QueryParam("lang")

	filters := requestfilter.GenerateFilters(ctx.QueryParam, []requestfilter.FilterRequest{
		{Param: "refprdocno", Type: requestfilter.FieldTypeString},
		{Param: "selectedvendor", Type: requestfilter.FieldTypeString},
		{Param: "-", Field: "docdatetime", Type: requestfilter.FieldTypeRangeDate},
		{Param: "branchcode", Field: "branch.code", Type: requestfilter.FieldTypeString},
	})

	docList, total, err := h.svc.SearchRFQStep(holdingCode, lang, filters, pageableStep)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}
	ctx.Response(http.StatusOK, common.ApiResponse{Success: true, Data: docList, Total: total})
	return nil
}

func (h RFQHttp) SaveBulk(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode
	input := ctx.ReadInput()

	dataReq := []models.RFQ{}
	err := json.Unmarshal([]byte(input), &dataReq)
	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	bulkResponse, err := h.svc.SaveInBatch(holdingCode, authUsername, dataReq)
	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}
	ctx.Response(http.StatusCreated, common.BulkResponse{Success: true, BulkImport: bulkResponse})
	return nil
}
