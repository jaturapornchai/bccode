package warehouse

import (
	"context"
	"encoding/json"
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
	ms           *microservice.Microservice
	cfg          config.IConfig
	svc          services.IWarehouseHttpService
	svcLocation  services.IWarehouseLocationHttpService
	svcBin       services.IWarehouseBinHttpService
	repo         repositories.IWarehouseRepository
	repoLocation repositories.IWarehouseLocationRepository
	repoBin      repositories.IWarehouseBinRepository
}

func NewWarehouseHttp(ms *microservice.Microservice, cfg config.IConfig) WarehouseHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	producer := ms.Producer(cfg.MQConfig())
	cache := ms.Cacher(cfg.CacherConfig())

	repo := repositories.NewWarehouseRepository(pst)
	repoLocation := repositories.NewWarehouseLocationRepository(pst)
	repoBin := repositories.NewWarehouseBinRepository(pst)
	repoMq := repositories.NewWarehouseMessageQueueRepository(producer)
	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)

	svc := services.NewWarehouseHttpService(repo, repoMq, repoLocation, masterSyncCacheRepo)
	svcLocation := services.NewWarehouseLocationHttpService(repoLocation, repo, masterSyncCacheRepo)
	svcBin := services.NewWarehouseBinHttpService(repoBin, repoLocation, masterSyncCacheRepo)

	// Backend-Owned Schema Rule: indexes are created here, on service construction (effectively
	// "first use" at process start), never by hand-run createIndex/mongosh.
	go func() {
		bgCtx := context.Background()
		_ = repo.EnsureIndexes(bgCtx)
		_ = repoLocation.EnsureIndexes(bgCtx)
		_ = repoBin.EnsureIndexes(bgCtx)
	}()

	return WarehouseHttp{
		ms:           ms,
		cfg:          cfg,
		svc:          svc,
		svcLocation:  svcLocation,
		svcBin:       svcBin,
		repo:         repo,
		repoLocation: repoLocation,
		repoBin:      repoBin,
	}
}

func (h WarehouseHttp) RegisterHttp() {
	h.ms.POST("/warehouse", h.CreateWarehouse)
	h.ms.GET("/warehouse", h.SearchWarehouse)
	h.ms.GET("/warehouse/tree", h.WarehouseTree)
	h.ms.GET("/warehouse/:id", h.InfoWarehouse)
	h.ms.PUT("/warehouse/:id", h.UpdateWarehouse)
	h.ms.DELETE("/warehouse/:id", h.DeleteWarehouse)

	// Location ("ที่เก็บสินค้า") endpoints — master data, own collection, references warehouse by guid.
	h.ms.POST("/warehouse/:warehouseguid/location", h.CreateLocation)
	h.ms.GET("/warehouse/:warehouseguid/location", h.SearchLocation)
	h.ms.GET("/warehouse/:warehouseguid/location/:locationguid", h.InfoLocation)
	h.ms.PUT("/warehouse/:warehouseguid/location/:locationguid", h.UpdateLocation)
	h.ms.DELETE("/warehouse/:warehouseguid/location/:locationguid", h.DeleteLocation)

	// Bin ("ที่วางสินค้า") endpoints — master data, own collection, references location by guid.
	h.ms.POST("/warehouse/:warehouseguid/location/:locationguid/bin", h.CreateBin)
	h.ms.GET("/warehouse/:warehouseguid/location/:locationguid/bin", h.SearchBin)
	h.ms.GET("/warehouse/:warehouseguid/location/:locationguid/bin/:binguid", h.InfoBin)
	h.ms.PUT("/warehouse/:warehouseguid/location/:locationguid/bin/:binguid", h.UpdateBin)
	h.ms.DELETE("/warehouse/:warehouseguid/location/:locationguid/bin/:binguid", h.DeleteBin)
}

func (h WarehouseHttp) CreateWarehouse(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username
	input := ctx.ReadInput()

	var doc warehouseModels.Warehouse
	if err := json.Unmarshal([]byte(input), &doc); err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
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

	// Unmarshal onto the existing values (not a zero-valued struct) so a field omitted from the
	// request JSON keeps its current value instead of being wiped to Go's zero value —
	// encoding/json only overwrites fields actually present in the JSON body.
	doc := existing.Warehouse
	if err := json.Unmarshal([]byte(input), &doc); err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	if err := h.svc.UpdateWarehouse(holdingCode, id, authUsername, doc); err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
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

// CreateLocation — POST /warehouse/:warehouseguid/location
func (h WarehouseHttp) CreateLocation(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username
	warehouseGuid := ctx.Param("warehouseguid")
	input := ctx.ReadInput()

	var doc warehouseModels.WarehouseLocation
	if err := json.Unmarshal([]byte(input), &doc); err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}
	doc.WarehouseGuid = warehouseGuid

	idx, err := h.svcLocation.CreateLocation(holdingCode, authUsername, doc)
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

// SearchLocation — GET /warehouse/:warehouseguid/location
func (h WarehouseHttp) SearchLocation(ctx microservice.IContext) error {
	holdingCode := ctx.UserInfo().HoldingCode
	warehouseGuid := ctx.Param("warehouseguid")
	pageable := utils.GetPageable(ctx.QueryParam)

	docList, pagination, err := h.svcLocation.SearchLocation(holdingCode, warehouseGuid, pageable)
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

// InfoLocation — GET /warehouse/:warehouseguid/location/:locationguid
func (h WarehouseHttp) InfoLocation(ctx microservice.IContext) error {
	holdingCode := ctx.UserInfo().HoldingCode
	locationGuid := ctx.Param("locationguid")

	data, err := h.svcLocation.InfoLocation(holdingCode, locationGuid)
	if err != nil {
		ctx.ResponseError(http.StatusNotFound, "Location not found")
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    data,
	})
	return nil
}

// UpdateLocation — PUT /warehouse/:warehouseguid/location/:locationguid
func (h WarehouseHttp) UpdateLocation(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username
	warehouseGuid := ctx.Param("warehouseguid")
	locationGuid := ctx.Param("locationguid")
	input := ctx.ReadInput()

	existing, err := h.svcLocation.InfoLocation(holdingCode, locationGuid)
	if err != nil {
		ctx.ResponseError(http.StatusNotFound, "Location not found")
		return err
	}

	// Same fetch-then-unmarshal-onto-existing preservation as UpdateWarehouse (see that handler's
	// comment) — a field omitted from the request JSON keeps its current value.
	doc := existing.WarehouseLocation
	if err := json.Unmarshal([]byte(input), &doc); err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}
	doc.WarehouseGuid = warehouseGuid

	if err := h.svcLocation.UpdateLocation(holdingCode, locationGuid, authUsername, doc); err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		ID:      locationGuid,
	})
	return nil
}

// DeleteLocation — DELETE /warehouse/:warehouseguid/location/:locationguid
func (h WarehouseHttp) DeleteLocation(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username
	locationGuid := ctx.Param("locationguid")

	if err := h.svcLocation.DeleteLocation(holdingCode, locationGuid, authUsername); err != nil {
		ctx.ResponseError(http.StatusNotFound, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		ID:      locationGuid,
	})
	return nil
}

// CreateBin — POST /warehouse/:warehouseguid/location/:locationguid/bin
func (h WarehouseHttp) CreateBin(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username
	warehouseGuid := ctx.Param("warehouseguid")
	locationGuid := ctx.Param("locationguid")
	input := ctx.ReadInput()

	var doc warehouseModels.WarehouseBin
	if err := json.Unmarshal([]byte(input), &doc); err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}
	doc.WarehouseGuid = warehouseGuid
	doc.LocationGuid = locationGuid

	idx, err := h.svcBin.CreateBin(holdingCode, authUsername, doc)
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

// SearchBin — GET /warehouse/:warehouseguid/location/:locationguid/bin
func (h WarehouseHttp) SearchBin(ctx microservice.IContext) error {
	holdingCode := ctx.UserInfo().HoldingCode
	locationGuid := ctx.Param("locationguid")
	pageable := utils.GetPageable(ctx.QueryParam)

	docList, pagination, err := h.svcBin.SearchBin(holdingCode, locationGuid, pageable)
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

// InfoBin — GET /warehouse/:warehouseguid/location/:locationguid/bin/:binguid
func (h WarehouseHttp) InfoBin(ctx microservice.IContext) error {
	holdingCode := ctx.UserInfo().HoldingCode
	binGuid := ctx.Param("binguid")

	data, err := h.svcBin.InfoBin(holdingCode, binGuid)
	if err != nil {
		ctx.ResponseError(http.StatusNotFound, "Bin not found")
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    data,
	})
	return nil
}

// UpdateBin — PUT /warehouse/:warehouseguid/location/:locationguid/bin/:binguid
func (h WarehouseHttp) UpdateBin(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username
	warehouseGuid := ctx.Param("warehouseguid")
	locationGuid := ctx.Param("locationguid")
	binGuid := ctx.Param("binguid")
	input := ctx.ReadInput()

	existing, err := h.svcBin.InfoBin(holdingCode, binGuid)
	if err != nil {
		ctx.ResponseError(http.StatusNotFound, "Bin not found")
		return err
	}

	// Same fetch-then-unmarshal-onto-existing preservation as UpdateWarehouse (see that handler's
	// comment) — a field omitted from the request JSON keeps its current value. This is what closes
	// the maxvolumecm3-gets-zeroed-on-every-edit bug found in review: any bin field the frontend
	// form doesn't send now survives instead of being wiped.
	doc := existing.WarehouseBin
	if err := json.Unmarshal([]byte(input), &doc); err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}
	doc.WarehouseGuid = warehouseGuid
	doc.LocationGuid = locationGuid

	if err := h.svcBin.UpdateBin(holdingCode, binGuid, authUsername, doc); err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		ID:      binGuid,
	})
	return nil
}

// DeleteBin — DELETE /warehouse/:warehouseguid/location/:locationguid/bin/:binguid
func (h WarehouseHttp) DeleteBin(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username
	binGuid := ctx.Param("binguid")

	if err := h.svcBin.DeleteBin(holdingCode, binGuid, authUsername); err != nil {
		ctx.ResponseError(http.StatusNotFound, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		ID:      binGuid,
	})
	return nil
}

// warehouseTreeNode is the GET /warehouse/tree response shape: every warehouse in the holding,
// each with its locations, each location with its bins — assembled from 3 flat queries in Go
// (not a Mongo aggregation pipeline), per the task spec.
type warehouseTreeLocationNode struct {
	warehouseModels.WarehouseLocationInfo
	Bins []warehouseModels.WarehouseBinInfo `json:"bins"`
}

type warehouseTreeNode struct {
	warehouseModels.WarehouseInfo
	Locations []warehouseTreeLocationNode `json:"locations"`
}

// WarehouseTree — GET /warehouse/tree
func (h WarehouseHttp) WarehouseTree(ctx microservice.IContext) error {
	holdingCode := ctx.UserInfo().HoldingCode
	bgCtx := context.Background()

	warehouseList, err := h.repo.Find(bgCtx, holdingCode, []string{"code"}, "")
	if err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	warehouseGuids := make([]string, 0, len(warehouseList))
	for _, wh := range warehouseList {
		warehouseGuids = append(warehouseGuids, wh.GuidFixed)
	}

	locationList := []warehouseModels.WarehouseLocationInfo{}
	if len(warehouseGuids) > 0 {
		locationList, err = h.repoLocation.FindByWarehouseGuids(bgCtx, holdingCode, warehouseGuids)
		if err != nil {
			ctx.ResponseError(http.StatusInternalServerError, err.Error())
			return err
		}
	}

	locationGuids := make([]string, 0, len(locationList))
	for _, loc := range locationList {
		locationGuids = append(locationGuids, loc.GuidFixed)
	}

	binList := []warehouseModels.WarehouseBinInfo{}
	if len(locationGuids) > 0 {
		binList, err = h.repoBin.FindByLocationGuids(bgCtx, holdingCode, locationGuids)
		if err != nil {
			ctx.ResponseError(http.StatusInternalServerError, err.Error())
			return err
		}
	}

	binsByLocation := map[string][]warehouseModels.WarehouseBinInfo{}
	for _, bin := range binList {
		binsByLocation[bin.LocationGuid] = append(binsByLocation[bin.LocationGuid], bin)
	}

	locationsByWarehouse := map[string][]warehouseTreeLocationNode{}
	for _, loc := range locationList {
		locationsByWarehouse[loc.WarehouseGuid] = append(locationsByWarehouse[loc.WarehouseGuid], warehouseTreeLocationNode{
			WarehouseLocationInfo: loc,
			Bins:                  binsByLocation[loc.GuidFixed],
		})
	}

	tree := make([]warehouseTreeNode, 0, len(warehouseList))
	for _, wh := range warehouseList {
		tree = append(tree, warehouseTreeNode{
			WarehouseInfo: wh,
			Locations:     locationsByWarehouse[wh.GuidFixed],
		})
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    tree,
	})
	return nil
}
