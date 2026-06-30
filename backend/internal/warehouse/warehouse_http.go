package warehouse

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/utils"
	warehouseModels "smlcloudplatform/internal/warehouse/models"
	"smlcloudplatform/pkg/microservice"
	"sync"
	"time"

	"gorm.io/gorm"
)

type WarehouseHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
}

func NewWarehouseHttp(ms *microservice.Microservice, cfg config.IConfig) WarehouseHttp {
	return WarehouseHttp{
		ms:  ms,
		cfg: cfg,
	}
}

// warehouseTenantMigrated guards the per-tenant AutoMigrate so it runs once per holdingcode.
var warehouseTenantMigrated sync.Map

// ensureTenantSchema guarantees the warehouse PG tables exist in this tenant's database.
// The framework's tenantModels auto-migration is supposed to create them, but make the handler
// self-sufficient (Backend-Owned Schema Rule): create the correct structure on first access per
// tenant so a fresh/empty tenant DB never 500s with `relation "warehouse" does not exist`.
func (h WarehouseHttp) ensureTenantSchema(holdingCode string) {
	if holdingCode == "" {
		return
	}
	if _, done := warehouseTenantMigrated.Load(holdingCode); done {
		return
	}
	pst := h.ms.PersisterTenant(h.cfg.PersisterConfig(), holdingCode)
	// Migrate in dependency order (referenced tables first) — AutoMigrate(WarehousePg) alone fails
	// when its Zones/Shelves relation tries to FK to tables that don't exist yet. Per-model so a
	// single failure does not abort the rest.
	models := []interface{}{
		warehouseModels.ShelfPg{},
		warehouseModels.ZonePg{},
		warehouseModels.WarehousePg{},
		warehouseModels.CompanyWarehousePg{},
	}
	ok := true
	for _, m := range models {
		if err := pst.AutoMigrate(m); err != nil {
			h.ms.Logger.Errorf("warehouse tenant AutoMigrate failed for %s: %v", holdingCode, err)
			ok = false
		}
	}
	if ok {
		warehouseTenantMigrated.Store(holdingCode, true)
	}
}

type WarehouseResponse struct {
	warehouseModels.WarehousePg
	Companies []string `json:"companyguids"`
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

func (h WarehouseHttp) CreateWarehouse(ctx microservice.IContext) error {
	holdingCode := ctx.UserInfo().HoldingCode
	h.ensureTenantSchema(holdingCode)
	input := ctx.ReadInput()

	type CreateWarehouseRequest struct {
		warehouseModels.WarehousePg
		CompanyGuids []string `json:"companyguids"`
	}

	var req CreateWarehouseRequest
	if err := json.Unmarshal([]byte(input), &req); err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	req.HoldingCode = holdingCode
	if req.GuidFixed == "" {
		req.GuidFixed = utils.NewGUID()
	}
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()
	req.IsActive = true

	pst := h.ms.PersisterTenant(h.cfg.PersisterConfig(), holdingCode)
	db := pst.DBClient()

	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&req.WarehousePg).Error; err != nil {
			return err
		}

		for _, compGuid := range req.CompanyGuids {
			cw := warehouseModels.CompanyWarehousePg{
				CompanyGuid:   compGuid,
				WarehouseGuid: req.GuidFixed,
			}
			if err := tx.Create(&cw).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
		ID:      req.GuidFixed,
	})
	return nil
}

func (h WarehouseHttp) SearchWarehouse(ctx microservice.IContext) error {
	holdingCode := ctx.UserInfo().HoldingCode
	companyGuid := ctx.QueryParam("companyguid")

	h.ensureTenantSchema(holdingCode)
	pst := h.ms.PersisterTenant(h.cfg.PersisterConfig(), holdingCode)
	db := pst.DBClient()

	var list []warehouseModels.WarehousePg
	if companyGuid != "" {
		// Get warehouses shared with this company
		var whGuids []string
		if err := db.Table("company_warehouses").Where("companyguid = ?", companyGuid).Pluck("warehouse_guid", &whGuids).Error; err != nil {
			ctx.ResponseError(http.StatusInternalServerError, err.Error())
			return err
		}
		if len(whGuids) == 0 {
			ctx.Response(http.StatusOK, common.ApiResponse{
				Success: true,
				Data:    []WarehouseResponse{},
			})
			return nil
		}
		if err := db.Where("holdingcode = ? AND guidfixed IN ?", holdingCode, whGuids).Preload("Zones.Shelves").Find(&list).Error; err != nil {
			ctx.ResponseError(http.StatusInternalServerError, err.Error())
			return err
		}
	} else {
		if err := db.Where("holdingcode = ?", holdingCode).Preload("Zones.Shelves").Find(&list).Error; err != nil {
			ctx.ResponseError(http.StatusInternalServerError, err.Error())
			return err
		}
	}

	var responseList []WarehouseResponse
	for _, wh := range list {
		var compGuids []string
		db.Table("company_warehouses").Where("warehouse_guid = ?", wh.GuidFixed).Pluck("companyguid", &compGuids)
		responseList = append(responseList, WarehouseResponse{
			WarehousePg: wh,
			Companies:   compGuids,
		})
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    responseList,
	})
	return nil
}

func (h WarehouseHttp) InfoWarehouse(ctx microservice.IContext) error {
	holdingCode := ctx.UserInfo().HoldingCode
	id := ctx.Param("id")

	pst := h.ms.PersisterTenant(h.cfg.PersisterConfig(), holdingCode)
	db := pst.DBClient()
	var data warehouseModels.WarehousePg
	if err := db.Where("holdingcode = ? AND guidfixed = ?", holdingCode, id).Preload("Zones.Shelves").First(&data).Error; err != nil {
		ctx.ResponseError(http.StatusNotFound, "Warehouse not found")
		return err
	}

	var compGuids []string
	db.Table("company_warehouses").Where("warehouse_guid = ?", data.GuidFixed).Pluck("companyguid", &compGuids)

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data: WarehouseResponse{
			WarehousePg: data,
			Companies:   compGuids,
		},
	})
	return nil
}

func (h WarehouseHttp) UpdateWarehouse(ctx microservice.IContext) error {
	holdingCode := ctx.UserInfo().HoldingCode
	id := ctx.Param("id")
	input := ctx.ReadInput()

	pst := h.ms.PersisterTenant(h.cfg.PersisterConfig(), holdingCode)
	db := pst.DBClient()
	var existing warehouseModels.WarehousePg
	if err := db.Where("holdingcode = ? AND guidfixed = ?", holdingCode, id).First(&existing).Error; err != nil {
		ctx.ResponseError(http.StatusNotFound, "Warehouse not found")
		return err
	}

	type UpdateWarehouseRequest struct {
		warehouseModels.WarehousePg
		CompanyGuids []string `json:"companyguids"`
	}

	var req UpdateWarehouseRequest
	if err := json.Unmarshal([]byte(input), &req); err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	existing.Names = req.Names
	existing.Code = req.Code
	existing.Latitude = req.Latitude
	existing.Longitude = req.Longitude
	existing.IsActive = req.IsActive
	existing.UpdatedAt = time.Now()

	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&existing).Error; err != nil {
			return err
		}

		// Delete existing junction rows
		if err := tx.Table("company_warehouses").Where("warehouse_guid = ?", id).Delete(nil).Error; err != nil {
			return err
		}

		// Insert new junction rows
		for _, compGuid := range req.CompanyGuids {
			cw := warehouseModels.CompanyWarehousePg{
				CompanyGuid:   compGuid,
				WarehouseGuid: id,
			}
			if err := tx.Create(&cw).Error; err != nil {
				return err
			}
		}

		// Delete existing shelves of zones belonging to this warehouse
		var zoneGuids []string
		if err := tx.Table("warehouse_zones").Where("warehouse_guid = ?", id).Pluck("guidfixed", &zoneGuids).Error; err != nil {
			return err
		}
		if len(zoneGuids) > 0 {
			if err := tx.Unscoped().Where("zone_guid IN ?", zoneGuids).Delete(&warehouseModels.ShelfPg{}).Error; err != nil {
				return err
			}
		}

		// Delete existing zones
		if err := tx.Unscoped().Where("warehouse_guid = ?", id).Delete(&warehouseModels.ZonePg{}).Error; err != nil {
			return err
		}

		// Insert new zones and shelves
		for _, zone := range req.Zones {
			zone.WarehouseGuid = id
			zone.HoldingCode = holdingCode
			if zone.GuidFixed == "" {
				zone.GuidFixed = utils.NewGUID()
			}
			zone.CreatedAt = time.Now()
			zone.UpdatedAt = time.Now()
			zone.IsActive = true

			// Prevent GORM from auto-saving association with incomplete fields (e.g. empty GUID)
			shelvesToCreate := zone.Shelves
			zone.Shelves = nil

			if err := tx.Create(&zone).Error; err != nil {
				return err
			}

			for _, shelf := range shelvesToCreate {
				shelf.ZoneGuid = zone.GuidFixed
				shelf.HoldingCode = holdingCode
				if shelf.GuidFixed == "" {
					shelf.GuidFixed = utils.NewGUID()
				}
				shelf.CreatedAt = time.Now()
				shelf.UpdatedAt = time.Now()
				shelf.IsActive = true

				if err := tx.Create(&shelf).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
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
	holdingCode := ctx.UserInfo().HoldingCode
	id := ctx.Param("id")

	pst := h.ms.PersisterTenant(h.cfg.PersisterConfig(), holdingCode)
	db := pst.DBClient()
	var data warehouseModels.WarehousePg
	if err := db.Where("holdingcode = ? AND guidfixed = ?", holdingCode, id).First(&data).Error; err != nil {
		ctx.ResponseError(http.StatusNotFound, "Warehouse not found")
		return err
	}

	// Soft delete
	if err := db.Delete(&data).Error; err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		ID:      id,
	})
	return nil
}

// Zone CRUD
func (h WarehouseHttp) CreateZone(ctx microservice.IContext) error {
	holdingCode := ctx.UserInfo().HoldingCode
	warehouseGuid := ctx.Param("warehouseGuid")
	input := ctx.ReadInput()

	var req warehouseModels.ZonePg
	if err := json.Unmarshal([]byte(input), &req); err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	req.HoldingCode = holdingCode
	req.WarehouseGuid = warehouseGuid
	if req.GuidFixed == "" {
		req.GuidFixed = utils.NewGUID()
	}
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()
	req.IsActive = true

	pst := h.ms.PersisterTenant(h.cfg.PersisterConfig(), holdingCode)
	db := pst.DBClient()
	if err := db.Create(&req).Error; err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
		ID:      req.GuidFixed,
	})
	return nil
}

func (h WarehouseHttp) SearchZone(ctx microservice.IContext) error {
	holdingCode := ctx.UserInfo().HoldingCode
	warehouseGuid := ctx.Param("warehouseGuid")

	pst := h.ms.PersisterTenant(h.cfg.PersisterConfig(), holdingCode)
	db := pst.DBClient()
	var list []warehouseModels.ZonePg
	if err := db.Where("holdingcode = ? AND warehouse_guid = ?", holdingCode, warehouseGuid).Find(&list).Error; err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    list,
	})
	return nil
}

func (h WarehouseHttp) InfoZone(ctx microservice.IContext) error {
	holdingCode := ctx.UserInfo().HoldingCode
	id := ctx.Param("id")

	pst := h.ms.PersisterTenant(h.cfg.PersisterConfig(), holdingCode)
	db := pst.DBClient()
	var data warehouseModels.ZonePg
	if err := db.Where("holdingcode = ? AND guidfixed = ?", holdingCode, id).First(&data).Error; err != nil {
		ctx.ResponseError(http.StatusNotFound, "Zone not found")
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    data,
	})
	return nil
}

func (h WarehouseHttp) UpdateZone(ctx microservice.IContext) error {
	holdingCode := ctx.UserInfo().HoldingCode
	id := ctx.Param("id")
	input := ctx.ReadInput()

	pst := h.ms.PersisterTenant(h.cfg.PersisterConfig(), holdingCode)
	db := pst.DBClient()
	var existing warehouseModels.ZonePg
	if err := db.Where("holdingcode = ? AND guidfixed = ?", holdingCode, id).First(&existing).Error; err != nil {
		ctx.ResponseError(http.StatusNotFound, "Zone not found")
		return err
	}

	var req warehouseModels.ZonePg
	if err := json.Unmarshal([]byte(input), &req); err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	existing.Names = req.Names
	existing.Code = req.Code
	existing.SuitableProductTypes = req.SuitableProductTypes
	existing.IsActive = req.IsActive
	existing.UpdatedAt = time.Now()

	if err := db.Save(&existing).Error; err != nil {
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
	holdingCode := ctx.UserInfo().HoldingCode
	id := ctx.Param("id")

	pst := h.ms.PersisterTenant(h.cfg.PersisterConfig(), holdingCode)
	db := pst.DBClient()
	var data warehouseModels.ZonePg
	if err := db.Where("holdingcode = ? AND guidfixed = ?", holdingCode, id).First(&data).Error; err != nil {
		ctx.ResponseError(http.StatusNotFound, "Zone not found")
		return err
	}

	// Soft delete
	if err := db.Delete(&data).Error; err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		ID:      id,
	})
	return nil
}

// Shelf CRUD
func (h WarehouseHttp) CreateShelf(ctx microservice.IContext) error {
	holdingCode := ctx.UserInfo().HoldingCode
	zoneGuid := ctx.Param("zoneGuid")
	input := ctx.ReadInput()

	var req warehouseModels.ShelfPg
	if err := json.Unmarshal([]byte(input), &req); err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	req.HoldingCode = holdingCode
	req.ZoneGuid = zoneGuid
	if req.GuidFixed == "" {
		req.GuidFixed = utils.NewGUID()
	}
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()
	req.IsActive = true

	pst := h.ms.PersisterTenant(h.cfg.PersisterConfig(), holdingCode)
	db := pst.DBClient()
	if err := db.Create(&req).Error; err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
		ID:      req.GuidFixed,
	})
	return nil
}

func (h WarehouseHttp) SearchShelf(ctx microservice.IContext) error {
	holdingCode := ctx.UserInfo().HoldingCode
	zoneGuid := ctx.Param("zoneGuid")

	pst := h.ms.PersisterTenant(h.cfg.PersisterConfig(), holdingCode)
	db := pst.DBClient()
	var list []warehouseModels.ShelfPg
	if err := db.Where("holdingcode = ? AND zone_guid = ?", holdingCode, zoneGuid).Find(&list).Error; err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    list,
	})
	return nil
}

func (h WarehouseHttp) InfoShelf(ctx microservice.IContext) error {
	holdingCode := ctx.UserInfo().HoldingCode
	id := ctx.Param("id")

	pst := h.ms.PersisterTenant(h.cfg.PersisterConfig(), holdingCode)
	db := pst.DBClient()
	var data warehouseModels.ShelfPg
	if err := db.Where("holdingcode = ? AND guidfixed = ?", holdingCode, id).First(&data).Error; err != nil {
		ctx.ResponseError(http.StatusNotFound, "Shelf not found")
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    data,
	})
	return nil
}

func (h WarehouseHttp) UpdateShelf(ctx microservice.IContext) error {
	holdingCode := ctx.UserInfo().HoldingCode
	id := ctx.Param("id")
	input := ctx.ReadInput()

	pst := h.ms.PersisterTenant(h.cfg.PersisterConfig(), holdingCode)
	db := pst.DBClient()
	var existing warehouseModels.ShelfPg
	if err := db.Where("holdingcode = ? AND guidfixed = ?", holdingCode, id).First(&existing).Error; err != nil {
		ctx.ResponseError(http.StatusNotFound, "Shelf not found")
		return err
	}

	var req warehouseModels.ShelfPg
	if err := json.Unmarshal([]byte(input), &req); err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	existing.Name = req.Name
	existing.Code = req.Code
	existing.IsActive = req.IsActive
	existing.UpdatedAt = time.Now()

	if err := db.Save(&existing).Error; err != nil {
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
	holdingCode := ctx.UserInfo().HoldingCode
	id := ctx.Param("id")

	pst := h.ms.PersisterTenant(h.cfg.PersisterConfig(), holdingCode)
	db := pst.DBClient()
	var data warehouseModels.ShelfPg
	if err := db.Where("holdingcode = ? AND guidfixed = ?", holdingCode, id).First(&data).Error; err != nil {
		ctx.ResponseError(http.StatusNotFound, "Shelf not found")
		return err
	}

	// Soft delete
	if err := db.Delete(&data).Error; err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		ID:      id,
	})
	return nil
}
