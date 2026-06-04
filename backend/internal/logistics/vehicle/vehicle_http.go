package vehicle

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	"smlcloudplatform/internal/logistics/vehicle/models"
	"smlcloudplatform/internal/logistics/vehicle/repositories"
	"smlcloudplatform/internal/logistics/vehicle/services"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/requestfilter"
	"smlcloudplatform/pkg/microservice"
)

type IVehicleHttp interface{}

type VehicleHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IVehicleHttpService
}

func NewVehicleHttp(ms *microservice.Microservice, cfg config.IConfig) VehicleHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())

	repo := repositories.NewVehicleRepository(pst)

	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewVehicleHttpService(repo, masterSyncCacheRepo)

	return VehicleHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h VehicleHttp) RegisterHttp() {

	h.ms.POST("/logistics/vehicle/bulk", h.SaveBulk)

	h.ms.GET("/logistics/vehicle", h.SearchVehiclePage)
	h.ms.GET("/logistics/vehicle/list", h.SearchVehicleStep)
	h.ms.POST("/logistics/vehicle", h.CreateVehicle)
	h.ms.GET("/logistics/vehicle/:id", h.InfoVehicle)
	h.ms.GET("/logistics/vehicle/code/:code", h.InfoVehicleByCode)
	h.ms.PUT("/logistics/vehicle/:id", h.UpdateVehicle)
	h.ms.DELETE("/logistics/vehicle/:id", h.DeleteVehicle)
	h.ms.DELETE("/logistics/vehicle", h.DeleteVehicleByGUIDs)
}

// Create Vehicle godoc
// @Description สร้างข้อมูลยานพาหนะ
// @Tags		Vehicle
// @Param		Vehicle  body      models.Vehicle  true  "Vehicle"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /logistics/vehicle [post]
func (h VehicleHttp) CreateVehicle(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	holdingCode := ctx.UserInfo().HoldingCode
	input := ctx.ReadInput()

	docReq := &models.Vehicle{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.CreateVehicle(holdingCode, authUsername, *docReq)

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

// Update Vehicle godoc
// @Description อัพเดทข้อมูลยานพาหนะ
// @Tags		Vehicle
// @Param		id  path      string  true  "Vehicle ID"
// @Param		Vehicle  body      models.Vehicle  true  "Vehicle"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /logistics/vehicle/{id} [put]
func (h VehicleHttp) UpdateVehicle(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.Vehicle{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateVehicle(holdingCode, id, authUsername, *docReq)

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

// Delete Vehicle godoc
// @Description ลบข้อมูลยานพาหนะ
// @Tags		Vehicle
// @Param		id  path      string  true  "Vehicle ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /logistics/vehicle/{id} [delete]
func (h VehicleHttp) DeleteVehicle(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteVehicle(holdingCode, id, authUsername)

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

// Delete Vehicle By GUIDs godoc
// @Description ลบข้อมูลยานพาหนะแบบหลายรายการ
// @Tags		Vehicle
// @Param		Vehicle  body      []string  true  "Vehicle GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /logistics/vehicle [delete]
func (h VehicleHttp) DeleteVehicleByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeleteVehicleByGUIDs(holdingCode, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get Vehicle godoc
// @Description ดึงข้อมูลยานพาหนะตาม guidfixed
// @Tags		Vehicle
// @Param		id  path      string  true  "Vehicle guidfixed"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /logistics/vehicle/{id} [get]
func (h VehicleHttp) InfoVehicle(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")

	h.ms.Logger.Debugf("ดึงข้อมูลยานพาหนะ %v", id)
	doc, err := h.svc.InfoVehicle(holdingCode, id)

	if err != nil {
		h.ms.Logger.Errorf("ดึงข้อมูลยานพาหนะผิดพลาด %v: %v", id, err)
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    doc,
	})
	return nil
}

// Get Vehicle By Code godoc
// @Description ดึงข้อมูลยานพาหนะตามรหัส
// @Tags		Vehicle
// @Param		code  path      string  true  "Vehicle Code"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /logistics/vehicle/code/{code} [get]
func (h VehicleHttp) InfoVehicleByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	code := ctx.Param("code")

	doc, err := h.svc.InfoVehicleByCode(holdingCode, code)

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

// List Vehicle page godoc
// @Description ค้นหายานพาหนะแบบแบ่งหน้า
// @Tags		Vehicle
// @Param		q		query	string		false  "Search Value"
// @Param		vehicletype		query	string		false  "Vehicle Type"
// @Param		status		query	integer		false  "Status"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /logistics/vehicle [get]
func (h VehicleHttp) SearchVehiclePage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	pageable := utils.GetPageable(ctx.QueryParam)

	filters := requestfilter.GenerateFilters(ctx.QueryParam, []requestfilter.FilterRequest{
		{
			Param: "vehicletype",
			Field: "vehicletype",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "status",
			Field: "status",
			Type:  requestfilter.FieldTypeInt,
		},
	})

	docList, pagination, err := h.svc.SearchVehicle(holdingCode, filters, pageable)

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

// List Vehicle step godoc
// @Description ค้นหายานพาหนะแบบ offset/limit
// @Tags		Vehicle
// @Param		q		query	string		false  "Search Value"
// @Param		vehicletype		query	string		false  "Vehicle Type"
// @Param		status		query	integer		false  "Status"
// @Param		offset	query	integer		false  "offset"
// @Param		limit	query	integer		false  "limit"
// @Param		lang	query	string		false  "lang"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /logistics/vehicle/list [get]
func (h VehicleHttp) SearchVehicleStep(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	pageableStep := utils.GetPageableStep(ctx.QueryParam)

	lang := ctx.QueryParam("lang")

	filters := requestfilter.GenerateFilters(ctx.QueryParam, []requestfilter.FilterRequest{
		{
			Param: "vehicletype",
			Field: "vehicletype",
			Type:  requestfilter.FieldTypeString,
		},
		{
			Param: "status",
			Field: "status",
			Type:  requestfilter.FieldTypeInt,
		},
	})

	docList, total, err := h.svc.SearchVehicleStep(holdingCode, lang, filters, pageableStep)

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

// Create Vehicle Bulk godoc
// @Description นำเข้ายานพาหนะแบบหลายรายการ
// @Tags		Vehicle
// @Param		Vehicle  body      []models.Vehicle  true  "Vehicle"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /logistics/vehicle/bulk [post]
func (h VehicleHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	input := ctx.ReadInput()

	dataReq := []models.Vehicle{}
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
