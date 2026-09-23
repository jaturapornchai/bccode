package shop

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"smlcloudplatform/internal/authentication/models"
	"smlcloudplatform/internal/centraldb"
	"smlcloudplatform/internal/config"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/apperr"
	"smlcloudplatform/pkg/microservice"
	"strings"
	"time"
)

type ShopMemberHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
}

func NewShopMemberHttp(ms *microservice.Microservice, cfg config.IConfig) *ShopMemberHttp {
	return &ShopMemberHttp{
		ms:  ms,
		cfg: cfg,
	}
}

// service binds the membership service to the central database.
func (h ShopMemberHttp) service() (IShopUserService, error) {
	db, err := centraldb.Open()
	if err != nil {
		return nil, err
	}
	return NewShopUserService(NewShopUserPostgresRepository(db)), nil
}

func (h *ShopMemberHttp) RegisterHttp() {
	h.ms.GET("/holding/users", h.ListUserInShop)
	h.ms.GET("/shop/users", h.ListUserInShop)

	h.ms.GET("/holding/permission/:username", h.InfoShopUser)
	h.ms.GET("/shop/permission/:username", h.InfoShopUser)

	// Save user profile + scope (system-settings "ผู้ใช้งาน" screen) and delete.
	// The screen writes via PUT /holding/permission (id in body, not path).
	h.ms.PUT("/holding/permission", h.SaveUserPermissionShop)
	h.ms.POST("/holding/permission", h.SaveUserPermissionShop)
	h.ms.DELETE("/holding/permission/:username", h.DeleteUserPermissionShop)

	// Holding admin management by email (holdingcode comes from the request; the caller's role
	// is resolved per-holding so it works from the holding-selection screen, no select required).
	h.ms.GET("/holding-member/list", h.ListHoldingMembers)
	// Read-only here: adding/removing members is done in Settings & Access Control › Login Accounts via
	// PUT/DELETE /holding/permission (decision 2026-09-23).
}

// List Shop User godoc
// @Description get shopuser
// @Tags		ShopUser
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /shop/users [get]
func (h ShopMemberHttp) ListUserInShop(ctx microservice.IContext) error {
	svc, svcErr := h.service()
	if svcErr != nil {
		return apperr.Respond(ctx, apperr.ErrInternal.WithWrap(svcErr))
	}
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username

	pageable := utils.GetPageable(ctx.QueryParam)

	// Role is resolved per-holding inside the service (requireHoldingManager); do NOT gate on
	// the JWT-selected userInfo.Role here — an owner/admin of this holding must always pass.
	docList, pagination, err := svc.ListUserInShop(holdingCode, authUsername, pageable)

	if err != nil {
		if err.Error() == "permission denied" {
			ctx.Response(http.StatusForbidden, &common.ApiResponse{Success: false, Message: "permission denied"})
			return err
		}
		ctx.ResponseError(400, "find failed")
		h.ms.Logger.Error("HTTP:: SearchShopUser " + err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success:    true,
		Pagination: pagination,
		Data:       docList,
	})
	return nil
}

// Get Shop User godoc
// @Description get shopuser info by username
// @Tags		ShopUser
// @Param		username	path     string  true  "username"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /shop/permission/{username} [get]
func (h ShopMemberHttp) InfoShopUser(ctx microservice.IContext) error {
	svc, svcErr := h.service()
	if svcErr != nil {
		return apperr.Respond(ctx, apperr.ErrInternal.WithWrap(svcErr))
	}
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	if err := svc.EnsureHoldingManager(holdingCode, userInfo.Username); err != nil {
		ctx.Response(http.StatusForbidden, &common.ApiResponse{Success: false, Message: "permission denied"})
		return err
	}

	username := strings.TrimSpace(decodePathUsername(ctx.Param("username")))

	if len(username) < 1 {
		ctx.ResponseError(400, "username invalid")
		return nil
	}

	doc, err := svc.InfoShopByUser(holdingCode, username)

	if err != nil {
		ctx.ResponseError(400, "find failed")
		h.ms.Logger.Error("HTTP:: SearchShopUser " + err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    doc,
	})
	return nil
}

// ListHoldingMembers godoc — list members of a holding for its owner/admin (holdingcode via query).
// @Tags ShopUser
// @Security AccessToken
// @Router /holding-member/list [get]
func (h ShopMemberHttp) ListHoldingMembers(ctx microservice.IContext) error {
	svc, svcErr := h.service()
	if svcErr != nil {
		return apperr.Respond(ctx, apperr.ErrInternal.WithWrap(svcErr))
	}
	authUsername := ctx.UserInfo().Username

	holdingCode, err := utils.NormalizeHoldingCode(ctx.QueryParam("holdingcode"))
	if err != nil || holdingCode == "" {
		ctx.ResponseError(400, "holdingcode invalid")
		return nil
	}

	pageable := utils.GetPageable(ctx.QueryParam)

	docList, pagination, err := svc.ListHoldingMembersByAdmin(holdingCode, authUsername, pageable)
	if err != nil {
		ctx.Response(http.StatusOK, &common.ApiResponse{Success: false, Message: err.Error()})
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{Success: true, Pagination: pagination, Data: docList})
	return nil
}

// Save Permission Shop User godoc
// @Description save shopuser permission and profile (position, department, LINE, approval)
// @Tags		ShopUser
// @Param		UserRoleRequest  body      models.UserRoleRequest  true  "UserRoleRequest"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /shop/permission [put]
func (h ShopMemberHttp) SaveUserPermissionShop(ctx microservice.IContext) error {
	svc, svcErr := h.service()
	if svcErr != nil {
		return apperr.Respond(ctx, apperr.ErrInternal.WithWrap(svcErr))
	}
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	if userInfo.Role != models.ROLE_OWNER && userInfo.Role != models.ROLE_ADMIN {
		ctx.Response(http.StatusForbidden, &common.ApiResponse{
			Success: false,
			Message: "permission denied",
		})

		return errors.New("permission denied")
	}

	input := ctx.ReadInput()

	userRoleReq := &models.UserRoleRequest{}
	err := json.Unmarshal([]byte(input), &userRoleReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	// ใช้ SaveUserFullProfile เพื่อบันทึกข้อมูลทั้งหมด (รวม position, department, LINE, approval)
	if err := h.hydrateAccessScopes(holdingCode, userRoleReq); err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}
	err = svc.SaveUserFullProfile(holdingCode, authUsername, userRoleReq)
	if err != nil {
		h.ms.Logger.Error("SaveUserPermissionShop failed: " + err.Error())
		ctx.ResponseError(400, err.Error())
		return err
	}
	ctx.Response(http.StatusOK,
		common.ApiResponse{
			Success: true,
		})

	return nil
}

// hydrateAccessScopes resolves immutable ids for saved scope rules and expands
// a holding-wide scope into per-company scopes (allow-list semantics,
// docs/organization.md): ScopeType "company"/"branch" must carry CompanyUID /
// BranchUID or scope enforcement cannot see them; ScopeType "holding" is a UI
// shorthand for "every company in this holding".
func (h ShopMemberHttp) hydrateAccessScopes(holdingCode string, req *models.UserRoleRequest) error {
	if req == nil || len(req.AccessScopes) == 0 {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	db, err := centraldb.Open()
	if err != nil {
		return err
	}
	expanded, err := hydrateAccessScopes(ctx, db, holdingCode, req.AccessScopes)
	if err != nil {
		return err
	}
	req.AccessScopes = expanded
	return nil
}

// Delete Shop User godoc
// @Description get shopuser info by username
// @Tags		ShopUser
// @Param		username	path     string  true  "username"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /shop/permission/{username} [delete]
func (h ShopMemberHttp) DeleteUserPermissionShop(ctx microservice.IContext) error {
	svc, svcErr := h.service()
	if svcErr != nil {
		return apperr.Respond(ctx, apperr.ErrInternal.WithWrap(svcErr))
	}
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	// Debug log
	h.ms.Logger.Debug("DeleteUserPermissionShop - authUsername: " + authUsername)
	h.ms.Logger.Debug("DeleteUserPermissionShop - holdingCode: " + holdingCode)

	if userInfo.Role != models.ROLE_OWNER && userInfo.Role != models.ROLE_ADMIN {
		ctx.Response(http.StatusForbidden, &common.ApiResponse{
			Success: false,
			Message: "permission denied",
		})

		return errors.New("permission denied")
	}

	username := strings.TrimSpace(decodePathUsername(ctx.Param("username")))

	// Debug log
	h.ms.Logger.Debug("DeleteUserPermissionShop - target username: " + username)

	if len(username) < 1 {
		h.ms.Logger.Error("DeleteUserPermissionShop - username is empty!")
		ctx.ResponseError(400, "username invalid")
		return nil
	}

	err := svc.DeleteUserPermissionShop(holdingCode, authUsername, username)

	if err != nil {
		h.ms.Logger.Error("DeleteUserPermissionShop - error: " + err.Error())
		ctx.ResponseError(400, err.Error())
		return err
	}

	h.ms.Logger.Debug("DeleteUserPermissionShop - success for: " + username)
	ctx.Response(http.StatusOK,
		common.ApiResponse{
			Success: true,
		})

	return nil
}

func decodePathUsername(value string) string {
	decoded, err := url.PathUnescape(value)
	if err != nil {
		return value
	}
	return decoded
}
