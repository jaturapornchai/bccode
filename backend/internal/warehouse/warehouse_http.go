package warehouse

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/utils"
	warehouseModels "smlcloudplatform/internal/warehouse/models"
	"smlcloudplatform/internal/warehouse/repositories"
	"smlcloudplatform/internal/warehouse/services"

	"smlcloudplatform/internal/config"
	"smlcloudplatform/pkg/microservice"
)

type WarehouseHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IWarehouseHttpService
}

func NewWarehouseHttp(ms *microservice.Microservice, cfg config.IConfig) WarehouseHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	producer := ms.Producer(cfg.MQConfig())
	cache := ms.Cacher(cfg.CacherConfig())

	repo := repositories.NewWarehouseRepository(pst)
	repoMq := repositories.NewWarehouseMessageQueueRepository(producer)
	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)

	svc := services.NewWarehouseHttpService(repo, repoMq, masterSyncCacheRepo)

	return WarehouseHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h WarehouseHttp) RegisterHttp() {
	h.ms.POST("/warehouse", h.CreateWarehouse)
	h.ms.GET("/warehouse", h.SearchWarehouse)
	h.ms.GET("/warehouse/:id", h.InfoWarehouse)
	h.ms.PUT("/warehouse/:id", h.UpdateWarehouse)
	h.ms.DELETE("/warehouse/:id", h.DeleteWarehouse)

	// Zone endpoints
	h.ms.POST("/warehouse/:warehouseGuid/zone", h.CreateZone)
	h.ms.GET("/warehouse/:warehouseGuid/zone", h.SearchZone)
	h.ms.GET("/warehouse/:warehouseGuid/zone/:id", h.InfoZone)
	h.ms.PUT("/warehouse/:warehouseGuid/zone/:id", h.UpdateZone)
	h.ms.DELETE("/warehouse/:warehouseGuid/zone/:id", h.DeleteZone)

	// Shelf endpoints
	h.ms.POST("/warehouse/:warehouseGuid/zone/:zoneGuid/shelf", h.CreateShelf)
	h.ms.GET("/warehouse/:warehouseGuid/zone/:zoneGuid/shelf", h.SearchShelf)
	h.ms.GET("/warehouse/:warehouseGuid/zone/:zoneGuid/shelf/:id", h.InfoShelf)
	h.ms.PUT("/warehouse/:warehouseGuid/zone/:zoneGuid/shelf/:id", h.UpdateShelf)
	h.ms.DELETE("/warehouse/:warehouseGuid/zone/:zoneGuid/shelf/:id", h.DeleteShelf)
}

// warehouseUpsertRequest is the wire shape for POST/PUT /warehouse. Location and CompanyGuids are
// pointers so we can tell "field absent" (nil, keep existing) apart from "field present, even empty"
// (non-nil, replace). This is the load-bearing property for partial updates from the frontend: a
// name-only edit omits "location" entirely and must not wipe existing zones/shelves.
type warehouseUpsertRequest struct {
	Code         string                      `json:"code"`
	Names        *[]common.NameX             `json:"names"`
	Location     *[]warehouseModels.Location `json:"location"`
	Latitude     float64                     `json:"latitude"`
	Longitude    float64                     `json:"longitude"`
	CompanyGuids *[]string                   `json:"companyguids"`
}

// validateLocationCompanyScope enforces that every restricted Location's CompanyGuids is a subset
// of the warehouse-level CompanyGuids. An empty warehouseCompanyGuids means "all companies" — any
// location list is a valid subset of that, so no check is needed. A location with its own empty
// CompanyGuids means "inherit all the warehouse allows" — also always valid.
func validateLocationCompanyScope(locations []warehouseModels.Location, warehouseCompanyGuids []string) error {
	if len(warehouseCompanyGuids) == 0 {
		return nil
	}
	allowed := make(map[string]bool, len(warehouseCompanyGuids))
	for _, g := range warehouseCompanyGuids {
		allowed[g] = true
	}
	for _, loc := range locations {
		for _, g := range loc.CompanyGuids {
			if !allowed[g] {
				return fmt.Errorf("โซน %s ใช้ได้เฉพาะบริษัทที่คลังอนุญาตเท่านั้น (บริษัท %s ไม่อยู่ในสิทธิ์ของคลัง)", loc.Code, g)
			}
		}
	}
	return nil
}

// assignGuids fills in a GuidFixed for any location/shelf that arrives without one, leaving
// existing GuidFixed values untouched.
func assignGuids(locations *[]warehouseModels.Location) {
	if locations == nil {
		return
	}
	for i := range *locations {
		if (*locations)[i].GuidFixed == "" {
			(*locations)[i].GuidFixed = utils.NewGUID()
		}
		shelves := (*locations)[i].Shelf
		for j := range shelves {
			if shelves[j].GuidFixed == "" {
				shelves[j].GuidFixed = utils.NewGUID()
			}
		}
	}
}

func (h WarehouseHttp) CreateWarehouse(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username
	input := ctx.ReadInput()

	var req warehouseUpsertRequest
	if err := json.Unmarshal([]byte(input), &req); err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	assignGuids(req.Location)

	doc := warehouseModels.Warehouse{
		Code:      req.Code,
		Names:     req.Names,
		Location:  req.Location,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
	}
	if req.CompanyGuids != nil {
		doc.CompanyGuids = *req.CompanyGuids
	}
	if req.Location != nil {
		if err := validateLocationCompanyScope(*req.Location, doc.CompanyGuids); err != nil {
			ctx.ResponseError(http.StatusBadRequest, err.Error())
			return err
		}
	}

	idx, err := h.svc.CreateWarehouse(holdingCode, authUsername, doc)
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

func (h WarehouseHttp) SearchWarehouse(ctx microservice.IContext) error {
	holdingCode := ctx.UserInfo().HoldingCode

	pageable := utils.GetPageable(ctx.QueryParam)

	filters := map[string]interface{}{}
	if companyGuid := ctx.QueryParam("companyguid"); companyGuid != "" {
		filters["companyguids"] = companyGuid
	}

	docList, pagination, err := h.svc.SearchWarehouse(holdingCode, filters, pageable)
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

func (h WarehouseHttp) InfoWarehouse(ctx microservice.IContext) error {
	holdingCode := ctx.UserInfo().HoldingCode
	id := ctx.Param("id")

	data, err := h.svc.InfoWarehouse(holdingCode, id)
	if err != nil {
		ctx.ResponseError(http.StatusNotFound, "Warehouse not found")
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    data,
	})
	return nil
}

func (h WarehouseHttp) UpdateWarehouse(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username
	id := ctx.Param("id")
	input := ctx.ReadInput()

	existing, err := h.svc.InfoWarehouse(holdingCode, id)
	if err != nil {
		ctx.ResponseError(http.StatusNotFound, "Warehouse not found")
		return err
	}

	var req warehouseUpsertRequest
	if err := json.Unmarshal([]byte(input), &req); err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	doc := existing.Warehouse
	doc.Code = req.Code
	doc.Names = req.Names
	doc.Latitude = req.Latitude
	doc.Longitude = req.Longitude

	// CRITICAL: only overwrite Location/CompanyGuids when the request actually sent them. A
	// name/lat/long-only edit (no "location" key at all) must leave existing zones/shelves intact.
	if req.CompanyGuids != nil {
		doc.CompanyGuids = *req.CompanyGuids
	}
	if req.Location != nil {
		assignGuids(req.Location)
		// Effective warehouse-level scope for this call: the request's CompanyGuids if it sent
		// one, otherwise whatever was already stored — validate BEFORE saving so a rejected
		// request never touches the document.
		if err := validateLocationCompanyScope(*req.Location, doc.CompanyGuids); err != nil {
			ctx.ResponseError(http.StatusBadRequest, err.Error())
			return err
		}
		doc.Location = req.Location
	}

	if err := h.svc.UpdateWarehouse(holdingCode, id, authUsername, doc); err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		ID:      id,
	})
	return nil
}

func (h WarehouseHttp) DeleteWarehouse(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username
	id := ctx.Param("id")

	if err := h.svc.DeleteWarehouse(holdingCode, id, authUsername); err != nil {
		ctx.ResponseError(http.StatusNotFound, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		ID:      id,
	})
	return nil
}

// Zone (aka Location) CRUD — Code-scoped sub-resource endpoints. Not currently called by the
// frontend (which does whole-warehouse replace via PUT /warehouse/:id), kept for API-surface parity.
func (h WarehouseHttp) CreateZone(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username
	warehouseGuid := ctx.Param("warehouseGuid")
	input := ctx.ReadInput()

	warehouseInfo, err := h.svc.InfoWarehouse(holdingCode, warehouseGuid)
	if err != nil {
		ctx.ResponseError(http.StatusNotFound, "Warehouse not found")
		return err
	}

	var req warehouseModels.LocationRequest
	if err := json.Unmarshal([]byte(input), &req); err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}
	req.WarehouseCode = warehouseInfo.Code

	if err := h.svc.CreateLocation(holdingCode, authUsername, warehouseInfo.Code, req); err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
		ID:      req.Code,
	})
	return nil
}

func (h WarehouseHttp) SearchZone(ctx microservice.IContext) error {
	holdingCode := ctx.UserInfo().HoldingCode
	warehouseGuid := ctx.Param("warehouseGuid")

	warehouseInfo, err := h.svc.InfoWarehouse(holdingCode, warehouseGuid)
	if err != nil {
		ctx.ResponseError(http.StatusNotFound, "Warehouse not found")
		return err
	}

	list := []warehouseModels.Location{}
	if warehouseInfo.Location != nil {
		list = *warehouseInfo.Location
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    list,
	})
	return nil
}

func (h WarehouseHttp) InfoZone(ctx microservice.IContext) error {
	holdingCode := ctx.UserInfo().HoldingCode
	warehouseGuid := ctx.Param("warehouseGuid")
	id := ctx.Param("id")

	warehouseInfo, err := h.svc.InfoWarehouse(holdingCode, warehouseGuid)
	if err != nil {
		ctx.ResponseError(http.StatusNotFound, "Warehouse not found")
		return err
	}

	if warehouseInfo.Location != nil {
		for _, location := range *warehouseInfo.Location {
			if location.Code == id || location.GuidFixed == id {
				ctx.Response(http.StatusOK, common.ApiResponse{
					Success: true,
					Data:    location,
				})
				return nil
			}
		}
	}

	ctx.ResponseError(http.StatusNotFound, "Zone not found")
	return nil
}

func (h WarehouseHttp) UpdateZone(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username
	warehouseGuid := ctx.Param("warehouseGuid")
	id := ctx.Param("id")
	input := ctx.ReadInput()

	warehouseInfo, err := h.svc.InfoWarehouse(holdingCode, warehouseGuid)
	if err != nil {
		ctx.ResponseError(http.StatusNotFound, "Warehouse not found")
		return err
	}

	var req warehouseModels.LocationRequest
	if err := json.Unmarshal([]byte(input), &req); err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}
	// Default only when absent — a request body carrying a DIFFERENT warehousecode means
	// "move this zone to that warehouse"; overwriting it forced every move into the
	// same-warehouse branch (200 but nothing moved).
	if req.WarehouseCode == "" {
		req.WarehouseCode = warehouseInfo.Code
	}

	if err := h.svc.UpdateLocation(holdingCode, authUsername, warehouseInfo.Code, id, req); err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		ID:      id,
	})
	return nil
}

func (h WarehouseHttp) DeleteZone(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username
	warehouseGuid := ctx.Param("warehouseGuid")
	id := ctx.Param("id")

	warehouseInfo, err := h.svc.InfoWarehouse(holdingCode, warehouseGuid)
	if err != nil {
		ctx.ResponseError(http.StatusNotFound, "Warehouse not found")
		return err
	}

	if err := h.svc.DeleteLocationByCodes(holdingCode, authUsername, warehouseInfo.Code, []string{id}); err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		ID:      id,
	})
	return nil
}

// Shelf CRUD — Code-scoped sub-resource endpoints, same API-surface-parity rationale as Zone above.
func (h WarehouseHttp) CreateShelf(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username
	warehouseGuid := ctx.Param("warehouseGuid")
	zoneGuid := ctx.Param("zoneGuid")
	input := ctx.ReadInput()

	warehouseInfo, locationCode, err := h.findWarehouseAndLocationCode(holdingCode, warehouseGuid, zoneGuid)
	if err != nil {
		ctx.ResponseError(http.StatusNotFound, err.Error())
		return err
	}

	var req warehouseModels.ShelfRequest
	if err := json.Unmarshal([]byte(input), &req); err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}
	req.WarehouseCode = warehouseInfo.Code
	req.LocationCode = locationCode

	if err := h.svc.CreateShelf(holdingCode, authUsername, warehouseInfo.Code, locationCode, req); err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
		ID:      req.Code,
	})
	return nil
}

func (h WarehouseHttp) SearchShelf(ctx microservice.IContext) error {
	holdingCode := ctx.UserInfo().HoldingCode
	warehouseGuid := ctx.Param("warehouseGuid")
	zoneGuid := ctx.Param("zoneGuid")

	_, locationCode, location, err := h.findLocation(holdingCode, warehouseGuid, zoneGuid)
	if err != nil {
		ctx.ResponseError(http.StatusNotFound, err.Error())
		return err
	}
	_ = locationCode

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    location.Shelf,
	})
	return nil
}

func (h WarehouseHttp) InfoShelf(ctx microservice.IContext) error {
	holdingCode := ctx.UserInfo().HoldingCode
	warehouseGuid := ctx.Param("warehouseGuid")
	zoneGuid := ctx.Param("zoneGuid")
	id := ctx.Param("id")

	_, _, location, err := h.findLocation(holdingCode, warehouseGuid, zoneGuid)
	if err != nil {
		ctx.ResponseError(http.StatusNotFound, err.Error())
		return err
	}

	for _, shelf := range location.Shelf {
		if shelf.Code == id || shelf.GuidFixed == id {
			ctx.Response(http.StatusOK, common.ApiResponse{
				Success: true,
				Data:    shelf,
			})
			return nil
		}
	}

	ctx.ResponseError(http.StatusNotFound, "Shelf not found")
	return nil
}

func (h WarehouseHttp) UpdateShelf(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username
	warehouseGuid := ctx.Param("warehouseGuid")
	zoneGuid := ctx.Param("zoneGuid")
	id := ctx.Param("id")
	input := ctx.ReadInput()

	warehouseInfo, locationCode, err := h.findWarehouseAndLocationCode(holdingCode, warehouseGuid, zoneGuid)
	if err != nil {
		ctx.ResponseError(http.StatusNotFound, err.Error())
		return err
	}

	var req warehouseModels.ShelfRequest
	if err := json.Unmarshal([]byte(input), &req); err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}
	// Same default-only-when-absent rule as UpdateZone: a different warehousecode/locationcode
	// in the body means "move this shelf there" — do not clobber the move target.
	if req.WarehouseCode == "" {
		req.WarehouseCode = warehouseInfo.Code
	}
	if req.LocationCode == "" {
		req.LocationCode = locationCode
	}

	if err := h.svc.UpdateShelf(holdingCode, authUsername, warehouseInfo.Code, locationCode, id, req); err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		ID:      id,
	})
	return nil
}

func (h WarehouseHttp) DeleteShelf(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username
	warehouseGuid := ctx.Param("warehouseGuid")
	zoneGuid := ctx.Param("zoneGuid")
	id := ctx.Param("id")

	warehouseInfo, locationCode, err := h.findWarehouseAndLocationCode(holdingCode, warehouseGuid, zoneGuid)
	if err != nil {
		ctx.ResponseError(http.StatusNotFound, err.Error())
		return err
	}

	if err := h.svc.DeleteShelfByCodes(holdingCode, authUsername, warehouseInfo.Code, locationCode, []string{id}); err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		ID:      id,
	})
	return nil
}

// findLocation resolves a zone (location) by GuidFixed-or-Code under a warehouse GuidFixed-or-Code,
// since the sub-routes are keyed by :warehouseGuid/:zoneGuid but the underlying models still key
// locations by Code (the Code-scoped service methods predate the GuidFixed addition).
func (h WarehouseHttp) findLocation(holdingCode, warehouseGuid, zoneGuid string) (warehouseModels.WarehouseInfo, string, warehouseModels.Location, error) {
	warehouseInfo, err := h.svc.InfoWarehouse(holdingCode, warehouseGuid)
	if err != nil {
		return warehouseModels.WarehouseInfo{}, "", warehouseModels.Location{}, err
	}

	if warehouseInfo.Location != nil {
		for _, location := range *warehouseInfo.Location {
			if location.Code == zoneGuid || location.GuidFixed == zoneGuid {
				return warehouseInfo, location.Code, location, nil
			}
		}
	}

	return warehouseModels.WarehouseInfo{}, "", warehouseModels.Location{}, errors.New("zone not found")
}

func (h WarehouseHttp) findWarehouseAndLocationCode(holdingCode, warehouseGuid, zoneGuid string) (warehouseModels.WarehouseInfo, string, error) {
	warehouseInfo, locationCode, _, err := h.findLocation(holdingCode, warehouseGuid, zoneGuid)
	return warehouseInfo, locationCode, err
}
