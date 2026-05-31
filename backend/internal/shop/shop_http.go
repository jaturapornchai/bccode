package shop

import (
	"encoding/json"
	"errors"
	"net/http"
	auth_model "smlcloudplatform/internal/authentication/models"
	"smlcloudplatform/internal/config"
	"smlcloudplatform/internal/logger"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	branch_model "smlcloudplatform/internal/organization/branch/models"
	branch_repositories "smlcloudplatform/internal/organization/branch/repositories"
	branch_services "smlcloudplatform/internal/organization/branch/services"
	company_model "smlcloudplatform/internal/organization/company/models"
	businesstype_models "smlcloudplatform/internal/organization/businesstype/models"
	businesstype_repositories "smlcloudplatform/internal/organization/businesstype/repositories"
	businesstype_services "smlcloudplatform/internal/organization/businesstype/services"
	deparment_repositories "smlcloudplatform/internal/organization/department/repositories"
	"smlcloudplatform/internal/shop/models"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
	"time"

	warehouse_models "smlcloudplatform/internal/warehouse/models"
	warehouse_repositories "smlcloudplatform/internal/warehouse/repositories"
	warehouse_services "smlcloudplatform/internal/warehouse/services"
)

type IShopHttp interface {
	RegisterHttp()
	CreateShop(ctx microservice.IContext) error
	UpdateShop(ctx microservice.IContext) error
	DeleteShop(ctx microservice.IContext) error
	InfoShop(ctx microservice.IContext) error
	SearchShop(ctx microservice.IContext) error
}

type ShopHttp struct {
	ms                  *microservice.Microservice
	cfg                 config.IConfig
	service             IShopService
	serviceBranch       branch_services.IBranchHttpService
	serviceWarehouse    warehouse_services.IWarehouseHttpService
	servicebusinessType businesstype_services.IBusinessTypeHttpService
	authService         *microservice.AuthService
}

func NewShopHttp(ms *microservice.Microservice, cfg config.IConfig) ShopHttp {

	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	repo := NewShopRepository(pst)
	cache := ms.Cacher(cfg.CacherConfig())
	producer := ms.Producer(cfg.MQConfig())

	shopUserRepo := NewShopUserRepository(pst)
	service := NewShopService(repo, shopUserRepo, utils.NewGUID, ms.TimeNow)

	authService := microservice.NewAuthService(ms.Cacher(cfg.CacherConfig()), 24*3*time.Hour, 24*30*time.Hour)

	repoBrach := branch_repositories.NewBranchRepository(pst)

	repoDepartment := deparment_repositories.NewDepartmentRepository(pst)
	repoBusinessType := businesstype_repositories.NewBusinessTypeRepository(pst)

	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	serviceBranch := branch_services.NewBranchHttpService(repoBrach, repoDepartment, repoBusinessType, masterSyncCacheRepo)

	serviceBusinessType := businesstype_services.NewBusinessTypeHttpService(repoBusinessType, masterSyncCacheRepo)

	repoWarehouse := warehouse_repositories.NewWarehouseRepository(pst)
	repoWarehouseMq := warehouse_repositories.NewWarehouseMessageQueueRepository(producer)
	svcWarehouse := warehouse_services.NewWarehouseHttpService(repoWarehouse, repoWarehouseMq, masterSyncCacheRepo)

	return ShopHttp{
		ms:                  ms,
		cfg:                 cfg,
		service:             service,
		serviceBranch:       serviceBranch,
		serviceWarehouse:    svcWarehouse,
		servicebusinessType: serviceBusinessType,
		authService:         authService,
	}
}

func (h ShopHttp) RegisterHttpMember() {
	h.ms.GET("/shop/:id", h.InfoShop)
}

func (h ShopHttp) RegisterHttp() {
	h.ms.GET("/shop/:id", h.InfoShop)
	// h.ms.GET("/shop", h.SearchShop)

	h.ms.POST("/shop", h.CreateShop, h.authService.MWFuncWithShop(h.ms.Cacher(h.cfg.CacherConfig())))
	h.ms.PUT("/shop/:id", h.UpdateShop)
	h.ms.DELETE("/shop/:id", h.DeleteShop)
}

// Create Shop On login  godoc
// @Description Create Shop on login
// @Tags		Authentication
// @Accept 		json
// @Param		Shop  body      models.Shop  true  "Add Shop"
// @Success		200	{object}		models.Shop
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /create-shop [post]
func Docs() {

}

// Create Shop godoc
// @Description Create Shop
// @Tags		Shop
// @Accept 		json
// @Param		ShopRequest  body      models.ShopRequest  true  "Add Shop"
// @Success		200	{object}		models.Shop
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /shop [post]
func (h ShopHttp) CreateShop(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	if len(authUsername) < 1 {
		ctx.ResponseError(400, "user authentication invalid")
		return nil
	}

	input := ctx.ReadInput()

	shopPayload := &models.ShopRequest{}
	err := json.Unmarshal([]byte(input), &shopPayload)

	if err != nil {
		ctx.ResponseError(400, "shop payload invalid")
		return err
	}

	shopTemp := shopPayload.Shop

	shopID, err := h.service.CreateShop(authUsername, shopTemp)

	if err != nil {
		ctx.Response(http.StatusBadRequest, &common.ApiResponse{
			Success: false,
			Message: err.Error(),
		})
		return err
	}

	err = h.initialShop(shopID, authUsername, *shopPayload)

	if err != nil {
		err2 := h.service.DeleteShop(shopID, authUsername)

		if err2 != nil {
			logger.GetLogger().Error("HTTP:: Error Rollback Shop " + err.Error())
		}

		ctx.Response(http.StatusBadRequest, &common.ApiResponse{
			Success: false,
			Message: err.Error(),
		})
		return err
	}

	ctx.Response(http.StatusOK, &common.ApiResponse{
		Success: true,
		ID:      shopID,
	})

	return nil
}
func (h ShopHttp) initialShop(shopID string, authUsername string, shopReq models.ShopRequest) (err error) {

	businessTypeDefault := businesstype_models.BusinessType{}

	businessTypeDefault.Code = shopReq.BusinessType.Code
	businessTypeDefault.Names = shopReq.BusinessType.Names
	businessTypeDefault.IsDefault = true

	if len(businessTypeDefault.Code) < 1 {

		businessTypeDefault.Code = "00000"

		businessTypeMainCodeTH := "th"
		businessTypeMainNameTH := "ธุรกิจหลัก"

		businessTypeMainCodeEN := "en"
		businessTypeMainNameEN := "Main Business"

		businessTypeDefault.Names = &[]common.NameX{
			{
				Code: &businessTypeMainCodeTH,
				Name: &businessTypeMainNameTH,
			},
			{
				Code: &businessTypeMainCodeEN,
				Name: &businessTypeMainNameEN,
			},
		}

	}

	businessTypeGUIDFixed, err := h.servicebusinessType.CreateBusinessType(shopID, authUsername, businessTypeDefault)

	if err != nil {
		return err
	}

	branchDefault := branch_model.Branch{}

	if len(shopReq.Settings.LanguageConfigs) > 0 {
		primaryLanguageConfigs := shopReq.Settings.LanguageConfigs[0]

		for _, langConf := range shopReq.Settings.LanguageConfigs {
			if langConf.IsDefault {
				primaryLanguageConfigs = langConf
				break
			}
		}

		for _, tempName := range shopReq.Names {
			if *tempName.Code == primaryLanguageConfigs.Code {
				branchDefault.CompanyNames = &[]common.NameX{
					{
						Code: tempName.Code,
						Name: tempName.Name,
					},
				}

				break
			}
		}

	}

	branchDefault.Code = "00000"

	branchMainCodeTH := "th"
	branchMainNameTH := "สำนักงานใหญ่"

	branchMainCodeEN := "en"
	branchMainNameEN := "Head Office"

	branchDefault.Names = &[]common.NameX{
		{
			Code: &branchMainCodeTH,
			Name: &branchMainNameTH,
		},
		{
			Code: &branchMainCodeEN,
			Name: &branchMainNameEN,
		},
	}

	if shopReq.Shop.MainShopId != "" {
		branchDefault.IsMainShop = false
		branchDefault.MainShopId = shopReq.Shop.MainShopId
	}
	branchDefault.BusinessType.GuidFixed = businessTypeGUIDFixed
	branchDefault.BusinessType.Code = businessTypeDefault.Code
	branchDefault.BusinessType.Names = businessTypeDefault.Names
	branchDefault.PaymentRounding = branch_model.PaymentRoundingSettings{
		Cash: branch_model.PaymentMethodRounding{
			Enabled: true,
			Rules: []branch_model.PaymentRoundingRule{
				{
					LowerBound: 0.01,
					UpperBound: 0.12,
					RoundTo:    0,
				},
				{
					LowerBound: 0.13,
					UpperBound: 0.37,
					RoundTo:    0.25,
				},
				{
					LowerBound: 0.38,
					UpperBound: 0.62,
					RoundTo:    0.5,
				},
				{
					LowerBound: 0.63,
					UpperBound: 0.87,
					RoundTo:    0.75,
				},
				{
					LowerBound: 0.88,
					UpperBound: 0.99,
					RoundTo:    1,
				},
			},
		},
		BankTransfer: branch_model.PaymentMethodRounding{
			Enabled: false,
			Rules: []branch_model.PaymentRoundingRule{
				{
					LowerBound: 0,
					UpperBound: 0,
					RoundTo:    0,
				},
			},
		},
		CreditCard: branch_model.PaymentMethodRounding{
			Enabled: false,
			Rules: []branch_model.PaymentRoundingRule{
				{
					LowerBound: 0,
					UpperBound: 0,
					RoundTo:    0,
				},
			},
		},
		Cheque: branch_model.PaymentMethodRounding{
			Enabled: false,
			Rules: []branch_model.PaymentRoundingRule{
				{
					LowerBound: 0,
					UpperBound: 0,
					RoundTo:    0,
				},
			},
		},
		Coupon: branch_model.PaymentMethodRounding{
			Enabled: false,
			Rules: []branch_model.PaymentRoundingRule{
				{
					LowerBound: 0,
					UpperBound: 0,
					RoundTo:    0,
				},
			},
		},
		QRCode: branch_model.PaymentMethodRounding{
			Enabled: false,
			Rules: []branch_model.PaymentRoundingRule{
				{
					LowerBound: 0,
					UpperBound: 0,
					RoundTo:    0,
				},
			},
		},
		Delivery: branch_model.PaymentMethodRounding{
			Enabled: false,
			Rules: []branch_model.PaymentRoundingRule{
				{
					LowerBound: 0,
					UpperBound: 0,
					RoundTo:    0,
				},
			},
		},
	}

	branchGUIDFixed, err := h.serviceBranch.CreateBranch(shopID, authUsername, branchDefault)

	if err != nil {
		err = h.servicebusinessType.DeleteBusinessType(shopID, businessTypeGUIDFixed, authUsername)
		if err != nil {
			logger.GetLogger().Error("HTTP:: Error Rollback BusinessType " + err.Error())
		}
		return err
	}

	warehouseDefault := warehouse_models.Warehouse{}
	warehouseDefault.Code = "00000"

	warehouseMainCodeTH := "th"
	warehouseMainNameTH := "สำนักงานใหญ่"

	warehouseMainCodeEN := "en"
	warehouseMainNameEN := "Head Office"

	warehouseDefault.Names = &[]common.NameX{
		{
			Code: &warehouseMainCodeTH,
			Name: &warehouseMainNameTH,
		},
		{
			Code: &warehouseMainCodeEN,
			Name: &warehouseMainNameEN,
		},
	}

	_, err = h.serviceWarehouse.CreateWarehouse(shopID, authUsername, warehouseDefault)

	if err != nil {

		err = h.serviceBranch.DeleteBranch(shopID, branchGUIDFixed, authUsername)

		if err != nil {
			logger.GetLogger().Error("HTTP:: Error Rollback Branch " + err.Error())
		}

		err = h.servicebusinessType.DeleteBusinessType(shopID, businessTypeGUIDFixed, authUsername)

		if err != nil {
			logger.GetLogger().Error("HTTP:: Error Rollback BusinessType " + err.Error())
		}

		return err
	}

	// Insert default company, branch, and warehouse into PostgreSQL tenant DB
	pst := h.ms.PersisterTenant(h.cfg.PersisterConfig(), shopID)
	dbPg := pst.DBClient()

	companyNames := common.JSONB{}
	if branchDefault.CompanyNames != nil {
		companyNames = common.JSONB(*branchDefault.CompanyNames)
	} else if len(shopReq.Names) > 0 {
		companyNames = common.JSONB(shopReq.Names)
	}

	companyGUIDFixed := utils.NewGUID()
	companyPg := company_model.CompanyPg{
		ShopID:    shopID,
		GuidFixed: companyGUIDFixed,
		Code:      "00000",
		Names:     companyNames,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := dbPg.Create(&companyPg).Error; err != nil {
		logger.GetLogger().Error("HTTP:: Error creating default company in PostgreSQL: " + err.Error())
	}

	company2Names := common.JSONB{}
	var namesList []common.NameX
	namesBytes, _ := json.Marshal(companyNames)
	if err := json.Unmarshal(namesBytes, &namesList); err == nil {
		for i := range namesList {
			suffix := " (สาขาย่อย)"
			if namesList[i].Code != nil && *namesList[i].Code == "en" {
				suffix = " (Branch)"
			}
			if namesList[i].Name != nil {
				newName := *namesList[i].Name + suffix
				namesList[i].Name = &newName
			}
		}
		company2Names = common.JSONB(namesList)
	} else {
		company2Names = companyNames
	}

	company2GUIDFixed := utils.NewGUID()
	company2Pg := company_model.CompanyPg{
		ShopID:    shopID,
		GuidFixed: company2GUIDFixed,
		Code:      "00001",
		Names:     company2Names,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := dbPg.Create(&company2Pg).Error; err != nil {
		logger.GetLogger().Error("HTTP:: Error creating default second company in PostgreSQL: " + err.Error())
	}

	branchNames := common.JSONB{}
	if branchDefault.Names != nil {
		branchNames = common.JSONB(*branchDefault.Names)
	}

	branchPg := branch_model.BranchPg{
		ShopID:      shopID,
		GuidFixed:   branchGUIDFixed,
		CompanyGuid: companyGUIDFixed,
		Code:        "00000",
		Names:       branchNames,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := dbPg.Create(&branchPg).Error; err != nil {
		logger.GetLogger().Error("HTTP:: Error creating default branch in PostgreSQL: " + err.Error())
	}

	whNames := common.JSONB{}
	if warehouseDefault.Names != nil {
		whNames = common.JSONB(*warehouseDefault.Names)
	}

	warehouseGUIDFixed := utils.NewGUID()
	warehousePg := warehouse_models.WarehousePg{
		ShopID:    shopID,
		GuidFixed: warehouseGUIDFixed,
		Code:      "00000",
		Names:     whNames,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := dbPg.Create(&warehousePg).Error; err != nil {
		logger.GetLogger().Error("HTTP:: Error creating default warehouse in PostgreSQL: " + err.Error())
	}

	// Link warehouse to company
	whLink := warehouse_models.CompanyWarehousePg{
		CompanyGuid:   companyGUIDFixed,
		WarehouseGuid: warehouseGUIDFixed,
	}
	if err := dbPg.Create(&whLink).Error; err != nil {
		logger.GetLogger().Error("HTTP:: Error creating default company_warehouse link in PostgreSQL: " + err.Error())
	}

	return nil
}

// Update Shop godoc
// @Description Update Shop
// @Tags		Shop
// @Accept 		json
// @Param		id	path     string  true  "Shop ID"
// @Param		Shop  body      models.Shop  true  "Shop Body"
// @Success		200	{object}		models.Shop
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /shop/{id} [put]
func (h ShopHttp) UpdateShop(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	id := ctx.Param("id")
	input := ctx.ReadInput()

	shopRequest := &models.Shop{}
	err := json.Unmarshal([]byte(input), &shopRequest)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if userInfo.Role != auth_model.ROLE_OWNER {
		ctx.Response(http.StatusOK, &common.ApiResponse{
			Success: false,
			Message: "permission denied",
		})

		return errors.New("permission denied")
	}

	err = h.service.UpdateShop(id, authUsername, *shopRequest)

	if err != nil {
		ctx.Response(http.StatusBadRequest, &common.ApiResponse{
			Success: false,
			Message: err.Error(),
		})
		return err
	}

	ctx.Response(http.StatusOK, &common.ApiResponse{
		Success: true,
		ID:      id,
	})
	return nil
}

// Delete Shop godoc
// @Description Delete Shop
// @Tags		Shop
// @Accept 		json
// @Param		id	path     string  true  "Shop ID"
// @Success		200	{object}		models.Shop
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /shop/{id} [delete]
func (h ShopHttp) DeleteShop(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username

	id := ctx.Param("id")

	if userInfo.Role != auth_model.ROLE_OWNER {
		ctx.Response(http.StatusOK, &common.ApiResponse{
			Success: false,
			Message: "permission denied",
		})

		return errors.New("permission denied")
	}

	err := h.service.DeleteShop(id, authUsername)

	if err != nil {
		ctx.Response(http.StatusBadRequest, &common.ApiResponse{
			Success: false,
			Message: err.Error(),
		})
		return err
	}
	ctx.Response(http.StatusOK, &common.ApiResponse{
		Success: true,
		ID:      id,
	})
	return nil
}

// Info Shop godoc
// @Description Infomation Shop Profile
// @Tags		Shop
// @Accept 		json
// @Param		id	path     string  true  "Shop ID"
// @Success		200	{array}	models.ShopInfo
// @Failure		401 {object}	common.ApiResponse
// @Security     AccessToken
// @Router /shop/{id} [get]
func (h ShopHttp) InfoShop(ctx microservice.IContext) error {
	id := ctx.Param("id")

	shopInfo, err := h.service.InfoShop(id)

	if err != nil {
		ctx.Response(http.StatusBadRequest, &common.ApiResponse{
			Success: false,
			Message: err.Error(),
		})
		return err
	}

	ctx.Response(http.StatusOK, &common.ApiResponse{
		Success: true,
		Data:    shopInfo,
	})
	return nil
}

// List Shop godoc
// @Description Access to Shop
// @Tags		Shop
// @Accept 		json
// @Success		200	{array}	models.ShopInfo
// @Failure		401 {object}	common.ApiResponse
// @Security     AccessToken
// @Router /shop [get]
func (h ShopHttp) SearchShop(ctx microservice.IContext) error {

	pageable := utils.GetPageable(ctx.QueryParam)

	shopList, pagination, err := h.service.SearchShop(pageable)

	if err != nil {
		ctx.ResponseError(400, "database error")
		h.ms.Logger.Error("HTTP:: SearchShop " + err.Error())
		return err
	}

	ctx.Response(http.StatusOK, map[string]interface{}{"success": true, "pagination": pagination, "data": shopList})
	return nil
}
