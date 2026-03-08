package warehouse

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/warehouse/models"
	"smlcloudplatform/internal/warehouse/repositories"
	"smlcloudplatform/internal/warehouse/services"
	"smlcloudplatform/pkg/microservice"
)

type IWarehouseHttp interface{}

type WarehouseHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IWarehouseHttpService
}

func NewWarehouseHttp(ms *microservice.Microservice, cfg config.IConfig) WarehouseHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())
	producer := ms.Producer(cfg.MQConfig())

	repoMq := repositories.NewWarehouseMessageQueueRepository(producer)
	repo := repositories.NewWarehouseRepository(pst)

	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	svc := services.NewWarehouseHttpService(repo, repoMq, masterSyncCacheRepo)

	return WarehouseHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h WarehouseHttp) RegisterHttp() {

	h.ms.POST("/warehouse/bulk", h.SaveBulk)

	h.ms.GET("/warehouse", h.SearchWarehousePage)
	h.ms.GET("/warehouse/list", h.SearchWarehouseStep)
	h.ms.GET("/warehouse/location", h.SearchWarehouseLocationPage)
	h.ms.GET("/warehouse/location/shelf", h.SearchWarehouseLocationShelfPage)
	h.ms.POST("/warehouse", h.CreateWarehouse)
	h.ms.GET("/warehouse/:id", h.InfoWarehouse)
	h.ms.GET("/warehouse/code/:code", h.InfoWarehouseByCode)
	h.ms.PUT("/warehouse/:id", h.UpdateWarehouse)
	h.ms.DELETE("/warehouse/:id", h.DeleteWarehouse)
	h.ms.DELETE("/warehouse", h.DeleteWarehouseByGUIDs)

	h.ms.GET("/warehouse/:warehouseCode/location/:locationCode", h.InfoLocation)
	h.ms.POST("/warehouse/:warehouseCode/location", h.CreateLocation)
	h.ms.PUT("/warehouse/:warehouseCode/location/:locationCode", h.UpdateLocation)
	h.ms.DELETE("/warehouse/:warehouseCode/location", h.DeleteLocation)

	h.ms.GET("/warehouse/:warehouseCode/location/:locationCode/shelf/:shelfCode", h.InfoShelf)
	h.ms.POST("/warehouse/:warehouseCode/location/:locationCode/shelf", h.CreateShelf)
	h.ms.PUT("/warehouse/:warehouseCode/location/:locationCode/shelf/:shelfCode", h.UpdateShelf)
	h.ms.DELETE("/warehouse/:warehouseCode/location/:locationCode/shelf", h.DeleteShelf)

	// ProductBarcode in Shelf endpoints
	h.ms.GET("/warehouse/:warehouseCode/location/:locationCode/shelf/:shelfCode/products", h.GetShelfProducts)
	h.ms.POST("/warehouse/:warehouseCode/location/:locationCode/shelf/:shelfCode/products", h.AddProductToShelf)
	h.ms.DELETE("/warehouse/:warehouseCode/location/:locationCode/shelf/:shelfCode/products/:productGuidFixed", h.RemoveProductFromShelf)

	// Bulk ProductBarcode operations
	h.ms.POST("/warehouse/:warehouseCode/location/:locationCode/shelf/:shelfCode/products/bulk", h.BulkAddProductsToShelf)
	h.ms.DELETE("/warehouse/:warehouseCode/location/:locationCode/shelf/:shelfCode/products/bulk", h.BulkRemoveProductsFromShelf)
}

// Create Warehouse godoc
// @Description Create Warehouse
// @Tags		Warehouse
// @Param		Warehouse  body      models.Warehouse  true  "Warehouse"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /warehouse [post]
func (h WarehouseHttp) CreateWarehouse(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := &models.Warehouse{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.CreateWarehouse(shopID, authUsername, *docReq)

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

// Update Warehouse godoc
// @Description Update Warehouse
// @Tags		Warehouse
// @Param		id  path      string  true  "Warehouse ID"
// @Param		Warehouse  body      models.Warehouse  true  "Warehouse"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /warehouse/{id} [put]
func (h WarehouseHttp) UpdateWarehouse(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.Warehouse{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateWarehouse(shopID, id, authUsername, *docReq)

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

// Create Warehouse Location godoc
// @Description Create Warehouse Location
// @Tags		Warehouse
// @Param		warehouseCode  path      string  true  "Warehouse Code"
// @Param		LocationRequest  body      models.LocationRequest  true  "Location Request"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /warehouse/{warehouseCode}/location [post]
func (h WarehouseHttp) CreateLocation(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	warehouseCode := ctx.Param("warehouseCode")

	docReq := &models.LocationRequest{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.CreateLocation(shopID, authUsername, warehouseCode, *docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
	})
	return nil
}

// Update Warehouse Location godoc
// @Description Update Warehouse Location
// @Tags		Warehouse
// @Param		warehouseCode  path      string  true  "Warehouse Code"
// @Param		locationCode  path      string  true  "location Code"
// @Param		LocationRequest  body      models.LocationRequest  true  "Location Request"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /warehouse/{warehouseCode}/location/{locationCode} [put]
func (h WarehouseHttp) UpdateLocation(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	warehouseCode := ctx.Param("warehouseCode")
	locationCode := ctx.Param("locationCode")

	input := ctx.ReadInput()

	docReq := &models.LocationRequest{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateLocation(shopID, authUsername, warehouseCode, locationCode, *docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Delete Warehouse Location godoc
// @Description Delete Warehouse Location
// @Tags		Warehouse
// @Param		warehouseCode  path      string  true  "Warehouse Code"
// @Param		LocationCode  body      []string  true  "Location Code"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /warehouse/{warehouseCode}/location/ [delete]
func (h WarehouseHttp) DeleteLocation(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	input := ctx.ReadInput()

	warehouseCode := ctx.Param("warehouseCode")

	docReq := []string{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.DeleteLocationByCodes(shopID, authUsername, warehouseCode, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Create Warehouse Shelf godoc
// @Description Create Warehouse Shelf
// @Tags		Warehouse
// @Param		warehouseCode  path      string  true  "Warehouse Code"
// @Param		locationCode  path      string  true  "Location Code"
// @Param		ShelfRequest  body      models.ShelfRequest  true  "Shelf Request"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /warehouse/{warehouseCode}/location/{locationCode}/shelf [post]
func (h WarehouseHttp) CreateShelf(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	warehouseCode := ctx.Param("warehouseCode")
	locationCode := ctx.Param("locationCode")

	docReq := &models.ShelfRequest{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.CreateShelf(shopID, authUsername, warehouseCode, locationCode, *docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
	})
	return nil
}

// Update Warehouse Shelf godoc
// @Description Update Warehouse Shelf
// @Tags		Warehouse
// @Param		warehouseCode  path      string  true  "Warehouse Code"
// @Param		locationCode  path      string  true  "location Code"
// @Param		shelfCode  path      string  true  "shelf Code"
// @Param		ShelfRequest  body      models.ShelfRequest  true  "Shelf Request"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /warehouse/{warehouseCode}/location/{locationCode}/shelf/{shelfCode} [put]
func (h WarehouseHttp) UpdateShelf(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	warehouseCode := ctx.Param("warehouseCode")
	locationCode := ctx.Param("locationCode")
	shelfCode := ctx.Param("shelfCode")

	input := ctx.ReadInput()

	docReq := &models.ShelfRequest{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateShelf(shopID, authUsername, warehouseCode, locationCode, shelfCode, *docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Delete Warehouse Shelf godoc
// @Description Delete Warehouse Shelf
// @Tags		Warehouse
// @Param		warehouseCode  path      string  true  "Warehouse Code"
// @Param		locationCode  path      string  true  "location Code"
// @Param		ShelfCode  body      []string  true  "Shelf Code"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /warehouse/{warehouseCode}/location/{locationCode}/shelf [delete]
func (h WarehouseHttp) DeleteShelf(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	input := ctx.ReadInput()

	warehouseCode := ctx.Param("warehouseCode")
	locationCode := ctx.Param("locationCode")

	docReq := []string{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.DeleteShelfByCodes(shopID, authUsername, warehouseCode, locationCode, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Delete Warehouse godoc
// @Description Delete Warehouse
// @Tags		Warehouse
// @Param		id  path      string  true  "Warehouse ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /warehouse/{id} [delete]
func (h WarehouseHttp) DeleteWarehouse(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteWarehouse(shopID, id, authUsername)

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

// Delete Warehouse godoc
// @Description Delete Warehouse
// @Tags		Warehouse
// @Param		Warehouse  body      []string  true  "Warehouse GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /warehouse [delete]
func (h WarehouseHttp) DeleteWarehouseByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeleteWarehouseByGUIDs(shopID, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get Warehouse godoc
// @Description get struct array by ID
// @Tags		Warehouse
// @Param		id  path      string  true  "Warehouse ID"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /warehouse/{id} [get]
func (h WarehouseHttp) InfoWarehouse(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get Warehouse %v", id)
	doc, err := h.svc.InfoWarehouse(shopID, id)

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

// Get Warehouse By Code godoc
// @Description get Warehouse by code
// @Tags		Warehouse
// @Param		id  path      string  true  "Warehouse ID"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /warehouse/code/{code} [get]
func (h WarehouseHttp) InfoWarehouseByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	code := ctx.Param("code")

	doc, err := h.svc.InfoWarehouseByCode(shopID, code)

	if err != nil {
		h.ms.Logger.Errorf("Error getting document by code %v: %v", code, err)
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    doc,
	})
	return nil
}

// List Warehouse godoc
// @Description List Warehouse
// @Tags		Warehouse
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "page"
// @Param		limit	query	integer		false  "limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /warehouse [get]
func (h WarehouseHttp) SearchWarehousePage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)

	docList, pagination, err := h.svc.SearchWarehouse(shopID, map[string]interface{}{}, pageable)

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

// Get List Location By Code godoc
// @Description get Location by code
// @Tags		Warehouse
// @Param		warehouseCode  path      string  true  "Warehouse Code"
// @Param		locationCode  path      string  true  "Location Code"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /warehouse/{warehouseCode}/location/{locationCode} [get]
func (h WarehouseHttp) InfoLocation(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	warehouseCode := ctx.Param("warehouseCode")
	locationCode := ctx.Param("locationCode")

	doc, err := h.svc.InfoLocation(shopID, warehouseCode, locationCode)

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

// List Warehouse Location godoc
// @Description get data warehouse location list
// @Tags		Warehouse
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "page"
// @Param		limit	query	integer		false  "limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /warehouse/location [get]
func (h WarehouseHttp) SearchWarehouseLocationPage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)

	docList, pagination, err := h.svc.SearchLocation(shopID, pageable)

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

// Get Shelf By Code godoc
// @Description get Shelf by code
// @Tags		Warehouse
// @Param		warehouseCode  path      string  true  "Warehouse Code"
// @Param		locationCode  path      string  true  "Location Code"
// @Param		shelfCode  path      string  true  "Shelf Code"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /warehouse/{warehouseCode}/location/{locationCode}/shelf/{shelfCode} [get]
func (h WarehouseHttp) InfoShelf(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	warehouseCode := ctx.Param("warehouseCode")
	locationCode := ctx.Param("locationCode")
	shelfCode := ctx.Param("shelfCode")

	doc, err := h.svc.InfoShelf(shopID, warehouseCode, locationCode, shelfCode)

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

// List Warehouse Location Shelf godoc
// @Description get data warehouse location shelf list
// @Tags		Warehouse
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "page"
// @Param		limit	query	integer		false  "limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /warehouse/location/shelf [get]
func (h WarehouseHttp) SearchWarehouseLocationShelfPage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)

	docList, pagination, err := h.svc.SearchShelf(shopID, pageable)

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

// List Warehouse godoc
// @Description search limit offset
// @Tags		Warehouse
// @Param		q		query	string		false  "Search Value"
// @Param		offset	query	integer		false  "offset"
// @Param		limit	query	integer		false  "limit"
// @Param		lang	query	string		false  "lang"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /warehouse/list [get]
func (h WarehouseHttp) SearchWarehouseStep(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageableStep := utils.GetPageableStep(ctx.QueryParam)

	lang := ctx.QueryParam("lang")

	docList, total, err := h.svc.SearchWarehouseStep(shopID, lang, pageableStep)

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

// Create Warehouse Bulk godoc
// @Description Create Warehouse
// @Tags		Warehouse
// @Param		Warehouse  body      []models.Warehouse  true  "Warehouse"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /warehouse/bulk [post]
func (h WarehouseHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	input := ctx.ReadInput()

	dataReq := []models.Warehouse{}
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

	ctx.Response(
		http.StatusCreated,
		common.BulkResponse{
			Success:    true,
			BulkImport: bulkResponse,
		},
	)

	return nil
}

// ProductBarcode in Shelf Management Endpoints

// Get Shelf Products godoc
// @Description Get all products in a specific shelf
// @Tags		Warehouse
// @Param		warehouseCode	path	string	true	"Warehouse Code"
// @Param		locationCode	path	string	true	"Location Code"
// @Param		shelfCode		path	string	true	"Shelf Code"
// @Success		200 {object} models.ApiResponse{data=[]models.ShelfProductBarcode}
// @Failure		400 {object} models.ApiResponse
// @Router /warehouse/{warehouseCode}/location/{locationCode}/shelf/{shelfCode}/products [get]
func (h WarehouseHttp) GetShelfProducts(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	warehouseCode := ctx.Param("warehouseCode")
	locationCode := ctx.Param("locationCode")
	shelfCode := ctx.Param("shelfCode")

	products, err := h.svc.GetShelfProductBarcodes(shopID, warehouseCode, locationCode, shelfCode)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    products,
	})
	return nil
}

// Add Product to Shelf godoc
// @Description Add a product barcode to a specific shelf
// @Tags		Warehouse
// @Param		warehouseCode	path	string	true	"Warehouse Code"
// @Param		locationCode	path	string	true	"Location Code"
// @Param		shelfCode		path	string	true	"Shelf Code"
// @Param		request			body	models.ShelfProductBarcode	true	"Product Barcode Data"
// @Success		201 {object} models.ApiResponse
// @Failure		400 {object} models.ApiResponse
// @Router /warehouse/{warehouseCode}/location/{locationCode}/shelf/{shelfCode}/products [post]
func (h WarehouseHttp) AddProductToShelf(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	warehouseCode := ctx.Param("warehouseCode")
	locationCode := ctx.Param("locationCode")
	shelfCode := ctx.Param("shelfCode")

	input := ctx.ReadInput()
	var productBarcode models.ShelfProductBarcode
	if err := json.Unmarshal([]byte(input), &productBarcode); err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	if err := ctx.Validate(productBarcode); err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	err := h.svc.AddProductBarcodeToShelf(shopID, authUsername, warehouseCode, locationCode, shelfCode, productBarcode)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
		Message: "Product added to shelf successfully",
	})
	return nil
}

// Remove Product from Shelf godoc
// @Description Remove a product barcode from a specific shelf
// @Tags		Warehouse
// @Param		warehouseCode		path	string	true	"Warehouse Code"
// @Param		locationCode		path	string	true	"Location Code"
// @Param		shelfCode			path	string	true	"Shelf Code"
// @Param		productGuidFixed	path	string	true	"Product GUID Fixed"
// @Success		200 {object} models.ApiResponse
// @Failure		400 {object} models.ApiResponse
// @Router /warehouse/{warehouseCode}/location/{locationCode}/shelf/{shelfCode}/products/{productGuidFixed} [delete]
func (h WarehouseHttp) RemoveProductFromShelf(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	warehouseCode := ctx.Param("warehouseCode")
	locationCode := ctx.Param("locationCode")
	shelfCode := ctx.Param("shelfCode")
	productGuidFixed := ctx.Param("productGuidFixed")

	err := h.svc.RemoveProductBarcodeFromShelf(shopID, authUsername, warehouseCode, locationCode, shelfCode, productGuidFixed)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Message: "Product removed from shelf successfully",
	})
	return nil
}

// Bulk Add Products to Shelf godoc
// @Description Add multiple product barcodes to a specific shelf
// @Tags		Warehouse
// @Param		warehouseCode	path	string	true	"Warehouse Code"
// @Param		locationCode	path	string	true	"Location Code"
// @Param		shelfCode		path	string	true	"Shelf Code"
// @Param		request			body	models.BulkShelfProductRequest	true	"Bulk Product Data"
// @Success		200 {object} models.BulkProductOperationResponse
// @Failure		400 {object} models.ApiResponse
// @Router /warehouse/{warehouseCode}/location/{locationCode}/shelf/{shelfCode}/products/bulk [post]
func (h WarehouseHttp) BulkAddProductsToShelf(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	warehouseCode := ctx.Param("warehouseCode")
	locationCode := ctx.Param("locationCode")
	shelfCode := ctx.Param("shelfCode")

	input := ctx.ReadInput()
	var request models.BulkShelfProductRequest
	if err := json.Unmarshal([]byte(input), &request); err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	if err := ctx.Validate(request); err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	response := h.svc.BulkAddProductBarcodesToShelf(shopID, authUsername, warehouseCode, locationCode, shelfCode, request.Products)

	statusCode := http.StatusOK
	if !response.Success {
		statusCode = http.StatusBadRequest
	}

	ctx.Response(statusCode, response)
	return nil
}

// Bulk Remove Products from Shelf godoc
// @Description Remove multiple product barcodes from a specific shelf
// @Tags		Warehouse
// @Param		warehouseCode	path	string	true	"Warehouse Code"
// @Param		locationCode	path	string	true	"Location Code"
// @Param		shelfCode		path	string	true	"Shelf Code"
// @Param		request			body	models.BulkRemoveProductRequest	true	"Bulk Remove Product Data"
// @Success		200 {object} models.BulkProductOperationResponse
// @Failure		400 {object} models.ApiResponse
// @Router /warehouse/{warehouseCode}/location/{locationCode}/shelf/{shelfCode}/products/bulk [delete]
func (h WarehouseHttp) BulkRemoveProductsFromShelf(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	warehouseCode := ctx.Param("warehouseCode")
	locationCode := ctx.Param("locationCode")
	shelfCode := ctx.Param("shelfCode")

	input := ctx.ReadInput()
	var request models.BulkRemoveProductRequest
	if err := json.Unmarshal([]byte(input), &request); err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	if err := ctx.Validate(request); err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	response := h.svc.BulkRemoveProductBarcodesFromShelf(shopID, authUsername, warehouseCode, locationCode, shelfCode, request.ProductGuidFixedList)

	statusCode := http.StatusOK
	if !response.Success {
		statusCode = http.StatusBadRequest
	}

	ctx.Response(statusCode, response)
	return nil
}
