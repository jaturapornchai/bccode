package shop

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	authmodels "smlcloudplatform/internal/authentication/models"
	"smlcloudplatform/internal/config"
	"smlcloudplatform/internal/logger"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	orgaccess "smlcloudplatform/internal/organization"
	orgpolicy "smlcloudplatform/internal/organization/access"
	branch_model "smlcloudplatform/internal/organization/branch/models"
	branch_repositories "smlcloudplatform/internal/organization/branch/repositories"
	branch_services "smlcloudplatform/internal/organization/branch/services"
	businesstype_models "smlcloudplatform/internal/organization/businesstype/models"
	businesstype_repositories "smlcloudplatform/internal/organization/businesstype/repositories"
	businesstype_services "smlcloudplatform/internal/organization/businesstype/services"
	company_model "smlcloudplatform/internal/organization/company/models"
	deparment_repositories "smlcloudplatform/internal/organization/department/repositories"
	"smlcloudplatform/internal/shop/models"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/apperr"
	"smlcloudplatform/pkg/microservice"
	"strings"
	"time"

	warehouse_models "smlcloudplatform/internal/warehouse/models"
	warehouse_repositories "smlcloudplatform/internal/warehouse/repositories"
	warehouse_services "smlcloudplatform/internal/warehouse/services"

	"go.mongodb.org/mongo-driver/bson"
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

	authService := microservice.NewAuthService(ms.Cacher(cfg.CacherConfig()), 24*3*time.Hour, 24*30*time.Hour, pst)

	repoBrach := branch_repositories.NewBranchRepository(pst)

	repoDepartment := deparment_repositories.NewDepartmentRepository(pst)
	repoBusinessType := businesstype_repositories.NewBusinessTypeRepository(pst)

	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)
	serviceBranch := branch_services.NewBranchHttpService(repoBrach, repoDepartment, repoBusinessType, masterSyncCacheRepo)

	serviceBusinessType := businesstype_services.NewBusinessTypeHttpService(repoBusinessType, masterSyncCacheRepo)

	repoWarehouse := warehouse_repositories.NewWarehouseRepository(pst)
	repoWarehouseMq := warehouse_repositories.NewWarehouseMessageQueueRepository(producer)
	repoWarehouseLocation := warehouse_repositories.NewWarehouseLocationRepository(pst)
	svcWarehouse := warehouse_services.NewWarehouseHttpService(repoWarehouse, repoWarehouseMq, repoWarehouseLocation, masterSyncCacheRepo)

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
	h.ms.GET("/holding/:id", h.InfoShop)
	h.ms.GET("/shop/:id", h.InfoShop)
}

func (h ShopHttp) RegisterHttp() {
	h.ms.GET("/holding/:id", h.InfoShop)
	h.ms.GET("/shop/:id", h.InfoShop)
	// h.ms.GET("/shop", h.SearchShop)

	h.ms.POST("/holding", h.CreateShop, h.authService.MWFuncWithShop(h.ms.Cacher(h.cfg.CacherConfig())))
	h.ms.POST("/shop", h.CreateShop, h.authService.MWFuncWithShop(h.ms.Cacher(h.cfg.CacherConfig())))
	h.ms.PUT("/holding/:id", h.UpdateShop)
	h.ms.PUT("/shop/:id", h.UpdateShop)
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
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	if strings.TrimSpace(userInfo.UID) == "" {
		return apperr.Respond(ctx, apperr.ErrUnauthorized.WithMessage("user authentication invalid"))
	}
	if authErr := orgaccess.RequireEmailedAccount(h.ms.MongoPersister(h.cfg.MongoPersisterConfig()), userInfo); authErr != nil {
		return apperr.Respond(ctx, authErr)
	}

	input := ctx.ReadInput()

	shopPayload := &models.ShopRequest{}
	err := json.Unmarshal([]byte(input), &shopPayload)

	if err != nil {
		return apperr.Respond(ctx, apperr.ErrBadRequest.WithMessage("shop payload invalid"))
	}

	shopTemp := shopPayload.Shop

	holdingUID, err := h.service.CreateShop(userInfo.UID, authUsername, shopTemp)

	if err != nil {
		return apperr.RespondErr(ctx, err)
	}

	ctx.Response(http.StatusOK, &common.ApiResponse{
		Success: true,
		ID:      holdingUID,
	})

	return nil
}
func (h ShopHttp) initialShop(holdingCode string, authUsername string, shopReq models.ShopRequest) (err error) {

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

	businessTypeGUIDFixed, err := h.servicebusinessType.CreateBusinessType(holdingCode, authUsername, businessTypeDefault)

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

	if shopReq.Shop.MainHoldingCode != "" {
		branchDefault.IsMainShop = false
		branchDefault.MainHoldingCode = shopReq.Shop.MainHoldingCode
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

	branchGUIDFixed, err := h.serviceBranch.CreateBranch(holdingCode, authUsername, branchDefault)

	if err != nil {
		err = h.servicebusinessType.DeleteBusinessType(holdingCode, businessTypeGUIDFixed, authUsername)
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

	_, err = h.serviceWarehouse.CreateWarehouse(holdingCode, authUsername, warehouseDefault)

	if err != nil {

		err = h.serviceBranch.DeleteBranch(holdingCode, branchGUIDFixed, authUsername)

		if err != nil {
			logger.GetLogger().Error("HTTP:: Error Rollback Branch " + err.Error())
		}

		err = h.servicebusinessType.DeleteBusinessType(holdingCode, businessTypeGUIDFixed, authUsername)

		if err != nil {
			logger.GetLogger().Error("HTTP:: Error Rollback BusinessType " + err.Error())
		}

		return err
	}

	// Insert default company, branch, and warehouse into PostgreSQL tenant DB
	pst := h.ms.PersisterTenant(h.cfg.PersisterConfig(), holdingCode)
	dbPg := pst.DBClient()

	companyNames := common.JSONB{}
	if branchDefault.CompanyNames != nil {
		companyNames = common.JSONB(*branchDefault.CompanyNames)
	} else if len(shopReq.Names) > 0 {
		companyNames = common.JSONB(shopReq.Names)
	}

	companyGUIDFixed := utils.NewGUID()
	companyPg := company_model.CompanyPg{
		HoldingCode: holdingCode,
		GuidFixed:   companyGUIDFixed,
		Code:        "00000",
		Names:       companyNames,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
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
		HoldingCode: holdingCode,
		GuidFixed:   company2GUIDFixed,
		Code:        "00001",
		Names:       company2Names,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := dbPg.Create(&company2Pg).Error; err != nil {
		logger.GetLogger().Error("HTTP:: Error creating default second company in PostgreSQL: " + err.Error())
	}

	branchNames := common.JSONB{}
	if branchDefault.Names != nil {
		branchNames = common.JSONB(*branchDefault.Names)
	}

	branchPg := branch_model.BranchPg{
		HoldingCode: holdingCode,
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
		HoldingCode: holdingCode,
		GuidFixed:   warehouseGUIDFixed,
		Code:        "00000",
		Names:       whNames,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
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
// @Param		id	path     string  true  "Holding Code"
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
	mongoCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pst := h.ms.MongoPersister(h.cfg.MongoPersisterConfig())
	membership, err := orgpolicy.FindActiveHoldingManager(mongoCtx, pst, userInfo, time.Now())
	if err != nil {
		return respondHoldingAccessError(ctx, err)
	}
	existing, err := h.service.InfoShop(id)
	if err != nil {
		return apperr.RespondErr(ctx, err)
	}
	if !strings.EqualFold(strings.TrimSpace(existing.HoldingCode), strings.TrimSpace(userInfo.HoldingCode)) {
		return apperr.Respond(ctx, apperr.ErrForbidden.WithMessage("Holding access denied"))
	}

	shopRequest := &models.Shop{}
	err = json.Unmarshal([]byte(input), &shopRequest)

	if err != nil {
		return apperr.Respond(ctx, apperr.ErrBadRequest.WithWrap(err))
	}

	requestedStatus, statusChanged, err := orgaccess.ResolveRequestedActiveStatus(input, existing.IsActive)
	if err != nil {
		return apperr.Respond(ctx, apperr.ErrBadRequest.WithWrap(err))
	}
	shopRequest.IsActive = requestedStatus
	if statusChanged && membership.Role != authmodels.ROLE_OWNER {
		return respondHoldingAccessError(ctx, orgpolicy.ErrHoldingOwnerRequired)
	}

	if statusChanged {
		reason, reasonErr := orgaccess.StatusChangeReason(input)
		if reasonErr != nil {
			return apperr.Respond(ctx, apperr.ErrBadRequest.WithWrap(reasonErr))
		}
		var existingDoc models.ShopDoc
		if err := pst.FindOne(mongoCtx, &models.ShopDoc{}, bson.M{
			"$or":       bson.A{bson.M{"guidfixed": id}, bson.M{"holdingcode": id}},
			"deletedat": bson.M{"$exists": false},
		}, &existingDoc); err != nil {
			return apperr.RespondErr(ctx, err)
		}
		holdingUID := strings.TrimSpace(membership.HoldingUID)
		if holdingUID == "" {
			holdingUID = strings.TrimSpace(existingDoc.GuidFixed)
		}
		if holdingUID == "" {
			return apperr.Respond(ctx, apperr.ErrConflict.WithMessage("stable Holding identity is required"))
		}
		now := time.Now().UTC()
		err = orgaccess.ApplyStatusChange(mongoCtx, pst, orgaccess.StatusChange{
			TargetModel: &models.ShopDoc{},
			TargetFilter: bson.M{
				"$or":       bson.A{bson.M{"guidfixed": id}, bson.M{"holdingcode": id}},
				"deletedat": bson.M{"$exists": false},
			},
			MembershipFilter: orgaccess.HoldingMembershipStatusFilter(existingDoc.HoldingCode),
			TargetType:       "holding",
			TargetUID:        holdingUID,
			HoldingUID:       holdingUID,
			ActorUID:         userInfo.UID,
			Reason:           reason,
			Before:           existing.IsActive,
			After:            requestedStatus,
			Version:          existingDoc.Version,
			Set:              bson.M{"updatedat": now, "updatedby": authUsername},
			OccurredAt:       now,
		})
	} else {
		err = h.service.UpdateShop(id, authUsername, *shopRequest)
	}

	if err != nil {
		if errors.Is(err, orgaccess.ErrStatusChangeConflict) || errors.Is(err, ErrHoldingCodeRenameUnavailable) {
			return apperr.Respond(ctx, apperr.ErrConflict.WithWrap(err))
		}
		return apperr.RespondErr(ctx, err)
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
// @Param		id	path     string  true  "Holding Code"
// @Success		200	{object}		models.Shop
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /shop/{id} [delete]
func (h ShopHttp) DeleteShop(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username

	id := ctx.Param("id")
	mongoCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if _, err := orgpolicy.FindActiveHoldingManager(mongoCtx, h.ms.MongoPersister(h.cfg.MongoPersisterConfig()), userInfo, time.Now()); err != nil {
		return respondHoldingAccessError(ctx, err)
	}
	existing, err := h.service.InfoShop(id)
	if err != nil {
		return apperr.RespondErr(ctx, err)
	}
	if !strings.EqualFold(strings.TrimSpace(existing.HoldingCode), strings.TrimSpace(userInfo.HoldingCode)) {
		return apperr.Respond(ctx, apperr.ErrForbidden.WithMessage("Holding access denied"))
	}

	err = h.service.DeleteShop(id, authUsername)

	if err != nil {
		return apperr.RespondErr(ctx, err)
	}
	ctx.Response(http.StatusOK, &common.ApiResponse{
		Success: true,
		ID:      id,
	})
	return nil
}

func respondHoldingAccessError(ctx microservice.IContext, err error) error {
	if errors.Is(err, orgpolicy.ErrActiveMembershipRequired) ||
		errors.Is(err, orgpolicy.ErrHoldingManagerRequired) ||
		errors.Is(err, orgpolicy.ErrHoldingOwnerRequired) {
		return apperr.Respond(ctx, apperr.ErrForbidden.WithMessage(err.Error()))
	}
	return apperr.Respond(ctx, apperr.ErrInternal.WithWrap(err))
}

// Info Shop godoc
// @Description Infomation Shop Profile
// @Tags		Shop
// @Accept 		json
// @Param		id	path     string  true  "Holding Code"
// @Success		200	{array}	models.ShopInfo
// @Failure		401 {object}	common.ApiResponse
// @Security     AccessToken
// @Router /shop/{id} [get]
func (h ShopHttp) InfoShop(ctx microservice.IContext) error {
	id := ctx.Param("id")

	shopInfo, err := h.service.InfoShop(id)

	if err != nil {
		return apperr.RespondErr(ctx, err)
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
		h.ms.Logger.Error("HTTP:: SearchShop " + err.Error())
		return apperr.Respond(ctx, apperr.ErrInternal.WithWrap(err).WithMessage("database error"))
	}

	ctx.Response(http.StatusOK, map[string]interface{}{"success": true, "pagination": pagination, "data": shopList})
	return nil
}
