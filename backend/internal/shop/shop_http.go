package shop

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	authmodels "smlcloudplatform/internal/authentication/models"
	"smlcloudplatform/internal/centraldb"
	"smlcloudplatform/internal/config"
	common "smlcloudplatform/internal/models"
	orgaccess "smlcloudplatform/internal/organization"
	orgpolicy "smlcloudplatform/internal/organization/access"
	"smlcloudplatform/internal/shop/models"
	"smlcloudplatform/pkg/apperr"
	"smlcloudplatform/pkg/microservice"
)

type ShopHttp struct {
	ms          *microservice.Microservice
	cfg         config.IConfig
	authService *microservice.AuthService
}

func NewShopHttp(ms *microservice.Microservice, cfg config.IConfig) ShopHttp {
	return ShopHttp{
		ms:          ms,
		cfg:         cfg,
		authService: microservice.NewAuthService(ms.Cacher(), 24*3*time.Hour, 24*30*time.Hour),
	}
}

func (h ShopHttp) RegisterHttp() {
	h.ms.GET("/holding/:id", h.InfoShop)
	h.ms.GET("/shop/:id", h.InfoShop)

	h.ms.POST("/holding", h.CreateShop, h.authService.MWFuncWithShop(h.ms.Cacher()))
	h.ms.POST("/shop", h.CreateShop, h.authService.MWFuncWithShop(h.ms.Cacher()))
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
	reqCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	db, err := centraldb.Open()
	if err != nil {
		return apperr.Respond(ctx, apperr.ErrInternal.WithWrap(err))
	}
	if authErr := orgaccess.RequireEmailedAccount(reqCtx, db, userInfo); authErr != nil {
		return apperr.Respond(ctx, authErr)
	}

	input := ctx.ReadInput()

	shopPayload := &models.ShopRequest{}
	err = json.Unmarshal([]byte(input), &shopPayload)

	if err != nil {
		return apperr.Respond(ctx, apperr.ErrBadRequest.WithMessage("shop payload invalid"))
	}

	service := NewShopService(NewShopPostgresRepository(db), NewShopUserPostgresRepository(db), h.ms.TimeNow)
	holdingUID, err := service.CreateShop(userInfo.UID, authUsername, shopPayload.Shop)
	if err != nil {
		return apperr.RespondErr(ctx, err)
	}

	ctx.Response(http.StatusOK, &common.ApiResponse{
		Success: true,
		ID:      holdingUID,
	})

	return nil
}

// UpdateShop godoc
// @Description Update the selected Holding profile; changing isactive needs OWNER and a reason.
// @Tags		Shop
// @Param		id	path     string  true  "Holding Code"
// @Router /shop/{id} [put]
func (h ShopHttp) UpdateShop(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	id := strings.TrimSpace(ctx.Param("id"))
	input := ctx.ReadInput()

	reqCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	db, err := centraldb.Open()
	if err != nil {
		return apperr.Respond(ctx, apperr.ErrInternal.WithWrap(err))
	}
	membership, err := orgpolicy.FindActiveHoldingManager(reqCtx, db, userInfo, time.Now())
	if err != nil {
		return respondHoldingAccessError(ctx, err)
	}
	service := NewShopService(NewShopPostgresRepository(db), NewShopUserPostgresRepository(db), h.ms.TimeNow)
	existing, err := service.InfoShop(id)
	if err != nil {
		return apperr.RespondErr(ctx, err)
	}
	if !strings.EqualFold(strings.TrimSpace(existing.HoldingCode), strings.TrimSpace(userInfo.HoldingCode)) {
		return apperr.Respond(ctx, apperr.ErrForbidden.WithMessage("Holding access denied"))
	}

	shopRequest := models.Shop{}
	if err := json.Unmarshal([]byte(input), &shopRequest); err != nil {
		return apperr.Respond(ctx, apperr.ErrBadRequest.WithWrap(err))
	}
	requestedStatus, statusChanged, err := orgaccess.ResolveRequestedActiveStatus(input, existing.IsActive)
	if err != nil {
		return apperr.Respond(ctx, apperr.ErrBadRequest.WithWrap(err))
	}
	if statusChanged {
		if membership.Role != authmodels.ROLE_OWNER {
			return respondHoldingAccessError(ctx, orgpolicy.ErrHoldingOwnerRequired)
		}
		reason, reasonErr := orgaccess.StatusChangeReason(input)
		if reasonErr != nil {
			return apperr.Respond(ctx, apperr.ErrBadRequest.WithWrap(reasonErr))
		}
		err = service.ChangeShopStatus(existing.HoldingCode, userInfo.UID, userInfo.Username, existing.IsActive, requestedStatus, reason)
	} else {
		err = service.UpdateShop(existing.HoldingCode, userInfo.Username, shopRequest)
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

func respondHoldingAccessError(ctx microservice.IContext, err error) error {
	if expired := orgaccess.AccessExpiredError(err, requestLanguage(ctx)); expired != nil {
		return apperr.Respond(ctx, expired)
	}
	if errors.Is(err, orgpolicy.ErrActiveMembershipRequired) ||
		errors.Is(err, orgpolicy.ErrHoldingManagerRequired) ||
		errors.Is(err, orgpolicy.ErrHoldingOwnerRequired) {
		return apperr.Respond(ctx, apperr.ErrForbidden.WithMessage(err.Error()))
	}
	return apperr.Respond(ctx, apperr.ErrInternal.WithWrap(err))
}

// InfoShop godoc
// @Description Holding profile
// @Tags		Shop
// @Param		id	path     string  true  "Holding Code"
// @Router /shop/{id} [get]
func (h ShopHttp) InfoShop(ctx microservice.IContext) error {
	holdingCode := strings.TrimSpace(ctx.Param("id"))
	reqCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	db, err := centraldb.Open()
	if err != nil {
		return apperr.Respond(ctx, apperr.ErrInternal.WithWrap(err))
	}
	// Any member (also before the Holding is selected) may read its profile.
	member, err := NewShopUserPostgresRepository(db).FindByHoldingCodeAndUserUID(reqCtx, holdingCode, ctx.UserInfo().UID)
	if err != nil || member.IsAccessDisabled {
		return apperr.Respond(ctx, apperr.ErrForbidden.WithMessage("Holding access denied").WithThaiMessage("ไม่มีสิทธิ์ใช้งานกลุ่มกิจการนี้"))
	}
	if authmodels.AccessExpired(member.AccessExpiryDate, time.Now()) {
		return apperr.Respond(ctx, orgaccess.AccessExpiredError(orgpolicy.ErrAccessExpired, requestLanguage(ctx)))
	}
	service := NewShopService(NewShopPostgresRepository(db), NewShopUserPostgresRepository(db), h.ms.TimeNow)
	shopInfo, err := service.InfoShop(holdingCode)
	if err != nil {
		return apperr.RespondErr(ctx, err)
	}
	ctx.Response(http.StatusOK, &common.ApiResponse{
		Success: true,
		Data:    shopInfo,
	})
	return nil
}
