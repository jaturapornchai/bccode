package purchaserequisition

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	"strings"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/purchaserequisition/models"
	"smlcloudplatform/internal/transaction/purchaserequisition/repositories"
	"smlcloudplatform/internal/transaction/purchaserequisition/services"
	"smlcloudplatform/internal/transaction/purchaserequisition/validators"
	trancache "smlcloudplatform/internal/transaction/repositories"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/requestfilter"
	"smlcloudplatform/pkg/microservice"
)

type IPurchaseRequisitionHttp interface{}

type PurchaseRequisitionHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IPurchaseRequisitionHttpService
}

func NewPurchaseRequisitionHttp(ms *microservice.Microservice, cfg config.IConfig) PurchaseRequisitionHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())
	producer := ms.Producer(cfg.MQConfig())

	repo := repositories.NewPurchaseRequisitionRepository(pst)
	repoMq := repositories.NewPurchaseRequisitionMessageQueueRepository(producer)
	transRepo := trancache.NewCacheRepository(cache)
	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewPurchaseRequisitionHttpService(repo, transRepo, repoMq, masterSyncCacheRepo)

	return PurchaseRequisitionHttp{
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
		"created_at", "modified_at", "docrefdatetime",
	}
	result := input
	for _, field := range timeFields {
		result = strings.ReplaceAll(result,
			`"`+field+`":""`,
			`"`+field+`":"0001-01-01T00:00:00Z"`)
	}
	return result
}

func (h PurchaseRequisitionHttp) RegisterHttp() {
	h.ms.POST("/transaction/purchase-requisition/bulk", h.SaveBulk)
	h.ms.GET("/transaction/purchase-requisition", h.SearchPurchaseRequisitionPage)
	h.ms.GET("/transaction/purchase-requisition/list", h.SearchPurchaseRequisitionStep)
	h.ms.POST("/transaction/purchase-requisition", h.CreatePurchaseRequisition)
	h.ms.GET("/transaction/purchase-requisition/:id", h.InfoPurchaseRequisition)
	h.ms.GET("/transaction/purchase-requisition/code/:code", h.InfoPurchaseRequisitionByCode)
	h.ms.PUT("/transaction/purchase-requisition/:id", h.UpdatePurchaseRequisition)
	h.ms.DELETE("/transaction/purchase-requisition/:id", h.DeletePurchaseRequisition)
	h.ms.DELETE("/transaction/purchase-requisition", h.DeletePurchaseRequisitionByGUIDs)
}

func (h PurchaseRequisitionHttp) CreatePurchaseRequisition(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()
	lang := getRequestLanguage(ctx)

	docReq := &models.PurchaseRequisition{}
	err := json.Unmarshal([]byte(sanitizeEmptyTimeFields(input)), &docReq)
	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}
	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, docNo, validationResult, err := h.svc.CreatePurchaseRequisition(shopID, authUsername, *docReq)
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

func (h PurchaseRequisitionHttp) UpdatePurchaseRequisition(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID
	id := ctx.Param("id")
	input := ctx.ReadInput()
	lang := getRequestLanguage(ctx)

	docReq := &models.PurchaseRequisition{}
	err := json.Unmarshal([]byte(sanitizeEmptyTimeFields(input)), &docReq)
	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}
	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	validationResult, err := h.svc.UpdatePurchaseRequisition(shopID, id, authUsername, *docReq)
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

func (h PurchaseRequisitionHttp) DeletePurchaseRequisition(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username
	id := ctx.Param("id")

	err := h.svc.DeletePurchaseRequisition(shopID, id, authUsername)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}
	ctx.Response(http.StatusOK, common.ApiResponse{Success: true, ID: id})
	return nil
}

func (h PurchaseRequisitionHttp) DeletePurchaseRequisitionByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeletePurchaseRequisitionByGUIDs(shopID, authUsername, docReq)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}
	ctx.Response(http.StatusOK, common.ApiResponse{Success: true})
	return nil
}

func (h PurchaseRequisitionHttp) InfoPurchaseRequisition(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	id := ctx.Param("id")

	doc, err := h.svc.InfoPurchaseRequisition(shopID, id)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}
	ctx.Response(http.StatusOK, common.ApiResponse{Success: true, Data: doc})
	return nil
}

func (h PurchaseRequisitionHttp) InfoPurchaseRequisitionByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	code := ctx.Param("code")

	doc, err := h.svc.InfoPurchaseRequisitionByCode(shopID, code)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}
	ctx.Response(http.StatusOK, common.ApiResponse{Success: true, Data: doc})
	return nil
}

func (h PurchaseRequisitionHttp) SearchPurchaseRequisitionPage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	pageable := utils.GetPageable(ctx.QueryParam)

	filters := requestfilter.GenerateFilters(ctx.QueryParam, []requestfilter.FilterRequest{
		{Param: "requestercode", Type: requestfilter.FieldTypeString},
		{Param: "departmentcode", Type: requestfilter.FieldTypeString},
		{Param: "-", Field: "docdatetime", Type: requestfilter.FieldTypeRangeDate},
		{Param: "branchcode", Field: "branch.code", Type: requestfilter.FieldTypeString},
	})

	docList, pagination, err := h.svc.SearchPurchaseRequisition(shopID, filters, pageable)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}
	ctx.Response(http.StatusOK, common.ApiResponse{Success: true, Data: docList, Pagination: pagination})
	return nil
}

func (h PurchaseRequisitionHttp) SearchPurchaseRequisitionStep(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	pageableStep := utils.GetPageableStep(ctx.QueryParam)
	lang := ctx.QueryParam("lang")

	filters := requestfilter.GenerateFilters(ctx.QueryParam, []requestfilter.FilterRequest{
		{Param: "requestercode", Type: requestfilter.FieldTypeString},
		{Param: "departmentcode", Type: requestfilter.FieldTypeString},
		{Param: "-", Field: "docdatetime", Type: requestfilter.FieldTypeRangeDate},
		{Param: "branchcode", Field: "branch.code", Type: requestfilter.FieldTypeString},
	})

	docList, total, err := h.svc.SearchPurchaseRequisitionStep(shopID, lang, filters, pageableStep)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}
	ctx.Response(http.StatusOK, common.ApiResponse{Success: true, Data: docList, Total: total})
	return nil
}

func (h PurchaseRequisitionHttp) SaveBulk(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID
	input := ctx.ReadInput()

	dataReq := []models.PurchaseRequisition{}
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
	ctx.Response(http.StatusCreated, common.BulkResponse{Success: true, BulkImport: bulkResponse})
	return nil
}
