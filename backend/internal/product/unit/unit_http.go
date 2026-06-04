package unit

import (
	"encoding/json"
	"net/http"
	"net/url"
	"smlcloudplatform/internal/config"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	productbarcode_repositories "smlcloudplatform/internal/product/productbarcode/repositories"
	"smlcloudplatform/internal/product/unit/models"
	"smlcloudplatform/internal/product/unit/repositories"
	"smlcloudplatform/internal/product/unit/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
	"strings"
)

type IUnitHttp interface{}

type UnitHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IUnitHttpService
}

func NewUnitHttp(ms *microservice.Microservice, cfg config.IConfig) UnitHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())
	prod := ms.Producer(cfg.MQConfig())

	repo := repositories.NewUnitRepository(pst)
	repoProductBarcode := productbarcode_repositories.NewProductBarcodeRepository(pst, cache)

	repoMessageQueue := repositories.NewUnitMessageQueueRepository(prod)

	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewUnitHttpService(repo, repoProductBarcode, repoMessageQueue, masterSyncCacheRepo)

	return UnitHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h UnitHttp) RegisterHttp() {

	h.ms.POST("/unit/bulk", h.SaveBulk)
	h.ms.GET("/unit", h.SearchUnit)
	h.ms.GET("/unit/list", h.SearchUnitLimit)
	h.ms.POST("/unit", h.CreateUnit)
	h.ms.POST("/unit/uploadfile", h.UploadFile)
	h.ms.GET("/unit/:id", h.InfoUnit)
	h.ms.GET("/unit/by-code", h.InfoArray)
	h.ms.GET("/unit/master", h.InfoArrayMaster)
	h.ms.PUT("/unit/:id", h.UpdateUnit)
	h.ms.PATCH("/unit/:id", h.UpdateFieldUnit)
	h.ms.DELETE("/unit/:id", h.DeleteUnit)
	h.ms.DELETE("/unit", h.DeleteByGUIDs)
}

// Upload File godoc
// @Description Upload an Excel file to import units
// @Tags        Unit
// @Param       excelfile  formData  file  true  "Excel file containing unit data"
// @Accept      multipart/form-data
// @Produce     json
// @Success     200 {string} string "Existing unit codes: [EA, PACK, BOX]"
// @Failure     400 {object} map[string]string "file is required"
// @Failure     500 {object} map[string]string "failed to read file or internal server error"
// @Security    AccessToken
// @Router      /unit/uploadfile [post]
func (h *UnitHttp) UploadFile(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	holdingCode := ctx.UserInfo().HoldingCode

	file, err := ctx.FormFile("excelfile")
	if err != nil {
		ctx.Response(http.StatusBadRequest, map[string]string{"error": "file is required"})
		return nil
	}

	src, err := file.Open()
	if err != nil {
		ctx.Response(http.StatusInternalServerError, map[string]string{"error": "failed to open file"})
		return nil
	}
	defer src.Close()

	// log.Printf("uploadDate: %s, effectDate: %s, dealers: %v", uploadDate, effectDate, dealers)

	// อ่านไฟล์เป็น byte
	fileBytes := make([]byte, file.Size)
	_, err = src.Read(fileBytes)
	if err != nil {
		ctx.Response(http.StatusInternalServerError, map[string]string{"error": "failed to read file"})
		return nil
	}

	// Process file และบันทึกข้อมูล
	result, err := h.svc.ImportUnitsFromFile(fileBytes, holdingCode, authUsername)
	if err != nil {
		ctx.Response(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return nil
	}

	ctx.Response(http.StatusOK, result)
	return nil
}

// Create Unit godoc
// @Description Create Unit
// @Tags		Unit
// @Param		Unit  body      models.Unit  true  "Unit"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /unit [post]
func (h UnitHttp) CreateUnit(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	holdingCode := ctx.UserInfo().HoldingCode
	input := ctx.ReadInput()

	docReq := &models.Unit{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.CreateUnit(holdingCode, authUsername, *docReq)

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

// Update Unit godoc
// @Description Update Unit
// @Tags		Unit
// @Param		id  path      string  true  "Unit ID"
// @Param		Unit  body      models.Unit  true  "Unit"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /unit/{id} [put]
func (h UnitHttp) UpdateUnit(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")
	if len(id) < 1 {
		ctx.ResponseError(http.StatusBadRequest, "guid is empty")
		return nil
	}

	input := ctx.ReadInput()

	docReq := &models.Unit{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateUnit(holdingCode, id, authUsername, *docReq)

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

// Update Field Unit godoc
// @Description Update Unit
// @Tags		Unit
// @Param		id  path      string  true  "Unit ID"
// @Param		Unit  body      models.Unit  true  "Unit"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /unit/{id} [patch]
func (h UnitHttp) UpdateFieldUnit(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")

	if len(id) < 1 {
		ctx.ResponseError(http.StatusBadRequest, "guid is empty")
		return nil
	}

	input := ctx.ReadInput()

	docReq := &models.Unit{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateFieldUnit(holdingCode, id, authUsername, *docReq)

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

// Get Unit godoc
// @Description get struct array by ID
// @Tags		Unit
// @Param		id  path      string  true  "Unit ID"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /unit/{id} [get]
func (h UnitHttp) InfoUnit(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")

	if len(id) < 1 {
		ctx.ResponseError(http.StatusBadRequest, "guid is empty")
		return nil
	}

	h.ms.Logger.Debugf("Get Unit %v", id)
	doc, err := h.svc.InfoUnit(holdingCode, id)

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

// Get Unit By unit code array godoc
// @Description get unit by unit code array
// @Tags		Unit
// @Param		codes	query	string		false  "Code filter, json array encode "
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /unit/by-code [get]
func (h UnitHttp) InfoArray(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	codesReq, err := url.QueryUnescape(ctx.QueryParam("codes"))

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	docReq := []string{}
	err = json.Unmarshal([]byte(codesReq), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	// where to filter array
	doc, err := h.svc.InfoUnitWTFArray(holdingCode, docReq)

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

// Get Master Unit By code array godoc
// @Description get master Unit by code array
// @Tags		Unit
// @Param		codes	query	string		false  "Code filter, json array encode "
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /unit/master [get]
func (h UnitHttp) InfoArrayMaster(ctx microservice.IContext) error {
	codesReq, err := url.QueryUnescape(ctx.QueryParam("codes"))

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	docReq := []string{}
	err = json.Unmarshal([]byte(codesReq), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	// where to filter array master
	doc, err := h.svc.InfoWTFArrayMaster(docReq)

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

// List Unit godoc
// @Description get struct array by ID
// @Tags		Unit
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "page "
// @Param		limit	query	integer		false  "liumit "
// @Param		unitcode	query	string		false  "unitcode filter ex. \"u001,u002,u003\""
// @Param		holding_code	query	string		false  "holding_code filter ex. \"holding1,holding2,holding3\""
// @Param		shopsid	query	string		false  "legacy holding_code filter"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /unit [get]
func (h UnitHttp) SearchUnit(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	pageable := utils.GetPageable(ctx.QueryParam)

	unitCode := ctx.QueryParam("unitcode")
	holdingCodeParam := strings.TrimSpace(ctx.QueryParam("holding_code"))
	shopsidParam := strings.TrimSpace(ctx.QueryParam("shopsid"))
	if holdingCodeParam != "" {
		shopsidParam = holdingCodeParam
	}
	companyGuid := ctx.QueryParam("company_guid")

	unitCodeFilters := []string{}
	if len(unitCode) > 0 {
		unitCodeFilters = strings.Split(unitCode, ",")
	}

	// Check if holding_code parameter is provided for multi-holding query
	if shopsidParam != "" {
		shopsidList := strings.Split(shopsidParam, ",")
		// Remove empty strings
		shopsID := []string{}
		for _, sid := range shopsidList {
			if strings.TrimSpace(sid) != "" {
				shopsID = append(shopsID, strings.TrimSpace(sid))
			}
		}

		if len(shopsID) > 0 {
			docList, pagination, err := h.svc.SearchUnitMultiShops(shopsID, unitCodeFilters, pageable)
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
	}

	// Default single shop query
	docList, pagination, err := h.svc.SearchUnit(holdingCode, companyGuid, unitCodeFilters, pageable)

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

// List Unit godoc
// @Description search limit offset
// @Tags		Unit
// @Param		q		query	string		false  "Search Value"
// @Param		offset	query	integer		false  "offset"
// @Param		limit	query	integer		false  "limit"
// @Param		lang	query	string		false  "lang ex. en,th"
// @Param		unitcode	query	string		false  "unitcode filter ex. \"u001,u002,u003\" "
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /unit/list [get]
func (h UnitHttp) SearchUnitLimit(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	pageableStep := utils.GetPageableStep(ctx.QueryParam)

	lang := ctx.QueryParam("lang")
	unitCode := ctx.QueryParam("unitcode")
	companyGuid := ctx.QueryParam("company_guid")

	unitCodeFilters := []string{}
	if len(unitCode) > 0 {
		unitCodeFilters = strings.Split(unitCode, ",")
	}

	docList, total, err := h.svc.SearchUnitLimit(holdingCode, companyGuid, lang, unitCodeFilters, pageableStep)

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

// Create Unit Bulk godoc
// @Description Create Unit
// @Tags		Unit
// @Param		Unit  body      []models.Unit  true  "Unit"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /unit/bulk [post]
func (h UnitHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	input := ctx.ReadInput()

	dataReq := []models.Unit{}
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

	ctx.Response(
		http.StatusCreated,
		common.BulkResponse{
			Success:    true,
			BulkImport: bulkResponse,
		},
	)

	return nil
}

// Delete Unit godoc
// @Description Delete Unit
// @Tags		Unit
// @Param		id  path      string  true  "Unit ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /unit/{id} [delete]
func (h UnitHttp) DeleteUnit(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username

	id := ctx.Param("id")

	if len(id) < 1 {
		ctx.ResponseError(http.StatusBadRequest, "unit guid is empty")
		return nil
	}

	err := h.svc.DeleteUnit(holdingCode, id, authUsername)

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

// Delete Unit By GUIDs godoc
// @Description Delete Unit
// @Tags		Unit
// @Param		Unit  body      []string  true  "Unit GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /unit [delete]
func (h UnitHttp) DeleteByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeleteUnitByGUIDs(holdingCode, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}
