package formtemplate

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	"smlcloudplatform/internal/form/formtemplate/models"
	"smlcloudplatform/internal/form/formtemplate/repositories"
	"smlcloudplatform/internal/form/formtemplate/services"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
)

type FormTemplateHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IFormTemplateHttpService
}

func NewFormTemplateHttp(ms *microservice.Microservice, cfg config.IConfig) FormTemplateHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())

	repo := repositories.NewFormTemplateRepository(pst)
	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewFormTemplateHttpService(repo, masterSyncCacheRepo)

	return FormTemplateHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h FormTemplateHttp) RegisterHttp() {
	h.ms.POST("/form/template/bulk", h.SaveBulk)
	h.ms.GET("/form/template", h.SearchFormTemplatePage)
	h.ms.GET("/form/template/list", h.SearchFormTemplateStep)
	h.ms.POST("/form/template", h.CreateFormTemplate)
	h.ms.POST("/form/template/save", h.SaveFormTemplate)
	h.ms.GET("/form/template/:id", h.InfoFormTemplate)
	h.ms.PUT("/form/template/:id", h.UpdateFormTemplate)
	h.ms.DELETE("/form/template/:id", h.DeleteFormTemplate)
	h.ms.DELETE("/form/template", h.DeleteFormTemplateByGUIDs)
}

func (h FormTemplateHttp) CreateFormTemplate(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	holdingCode := ctx.UserInfo().HoldingCode
	input := ctx.ReadInput()

	docReq := &models.FormTemplate{}
	err := json.Unmarshal([]byte(input), &docReq)
	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.CreateFormTemplate(holdingCode, authUsername, *docReq)
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

func (h FormTemplateHttp) SaveFormTemplate(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	holdingCode := ctx.UserInfo().HoldingCode
	input := ctx.ReadInput()

	docReq := &models.FormTemplate{}
	err := json.Unmarshal([]byte(input), &docReq)
	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.SaveFormTemplate(holdingCode, authUsername, *docReq)
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

func (h FormTemplateHttp) UpdateFormTemplate(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.FormTemplate{}
	err := json.Unmarshal([]byte(input), &docReq)
	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateFormTemplate(holdingCode, id, authUsername, *docReq)
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

func (h FormTemplateHttp) DeleteFormTemplate(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteFormTemplate(holdingCode, id, authUsername)
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

func (h FormTemplateHttp) DeleteFormTemplateByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeleteFormTemplateByGUIDs(holdingCode, authUsername, docReq)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})
	return nil
}

func (h FormTemplateHttp) InfoFormTemplate(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")

	doc, err := h.svc.InfoFormTemplate(holdingCode, id)
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

func (h FormTemplateHttp) SearchFormTemplatePage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	pageable := utils.GetPageable(ctx.QueryParam)

	// Support filtering by doctype
	filters := map[string]interface{}{}
	docType := ctx.QueryParam("doc_type")
	if docType != "" {
		filters["doc_type"] = docType
	}

	docList, pagination, err := h.svc.SearchFormTemplate(holdingCode, filters, pageable)
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

func (h FormTemplateHttp) SearchFormTemplateStep(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	pageableStep := utils.GetPageableStep(ctx.QueryParam)
	lang := ctx.QueryParam("lang")

	docList, total, err := h.svc.SearchFormTemplateStep(holdingCode, lang, pageableStep)
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

func (h FormTemplateHttp) SaveBulk(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	input := ctx.ReadInput()

	dataReq := []models.FormTemplate{}
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

	ctx.Response(http.StatusCreated, common.BulkResponse{
		Success:    true,
		BulkImport: bulkResponse,
	})
	return nil
}
