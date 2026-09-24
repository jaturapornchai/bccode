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
	"smlcloudplatform/internal/goapi/language"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/apperr"
	"smlcloudplatform/pkg/microservice"
	"smlcloudplatform/pkg/textguard"
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
		h.ms.Logger.Error("HTTP:: SearchShopUser " + err.Error())
		return apperr.Respond(ctx, shopUserAppError(err, requestLanguage(ctx)))
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
		return apperr.Respond(ctx, shopUserAppError(err, requestLanguage(ctx)))
	}

	username := strings.TrimSpace(decodePathUsername(ctx.Param("username")))

	if len(username) < 1 {
		return apperr.Respond(ctx, localizedAppError(apperr.ErrValidation, "ss_err_invalid_data", requestLanguage(ctx)).WithField("username"))
	}

	doc, err := svc.InfoShopByUser(holdingCode, username)

	if err != nil {
		h.ms.Logger.Error("HTTP:: SearchShopUser " + err.Error())
		return apperr.Respond(ctx, shopUserAppError(err, requestLanguage(ctx)))
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
		return apperr.Respond(ctx, localizedAppError(apperr.ErrValidation, "ss_err_invalid_data", requestLanguage(ctx)).WithField("holdingcode"))
	}

	pageable := utils.GetPageable(ctx.QueryParam)

	docList, pagination, err := svc.ListHoldingMembersByAdmin(holdingCode, authUsername, pageable)
	if err != nil {
		h.ms.Logger.Error("HTTP:: ListHoldingMembers " + err.Error())
		return apperr.Respond(ctx, shopUserAppError(err, requestLanguage(ctx)))
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
		return apperr.Respond(ctx, shopUserAppError(errors.New("permission denied"), requestLanguage(ctx)))
	}

	input := ctx.ReadInput()

	userRoleReq := &models.UserRoleRequest{}
	err := json.Unmarshal([]byte(input), &userRoleReq)

	if err != nil {
		// The decoder text names Go struct fields: log it, never show it.
		h.ms.Logger.Error("SaveUserPermissionShop invalid body: " + err.Error())
		if errors.Is(err, models.ErrInvalidAccessExpiryDate) {
			return apperr.Respond(ctx, localizedAppError(apperr.ErrValidation, "ss_err_access_expiry_date_invalid", requestLanguage(ctx)).
				WithField("accessexpirydate").WithWrap(err))
		}
		return apperr.Respond(ctx, localizedAppError(apperr.ErrBadRequest, "ss_err_invalid_data", requestLanguage(ctx)).WithWrap(err))
	}

	if appErr := memberNULError(userRoleReq, requestLanguage(ctx)); appErr != nil {
		return apperr.Respond(ctx, appErr)
	}

	// ใช้ SaveUserFullProfile เพื่อบันทึกข้อมูลทั้งหมด (รวม position, department, LINE, approval)
	if err := h.hydrateAccessScopes(holdingCode, userRoleReq); err != nil {
		h.ms.Logger.Error("SaveUserPermissionShop access scopes: " + err.Error())
		return apperr.Respond(ctx, shopUserAppError(err, requestLanguage(ctx)))
	}
	err = svc.SaveUserFullProfile(holdingCode, authUsername, userRoleReq)
	if err != nil {
		h.ms.Logger.Error("SaveUserPermissionShop failed: " + err.Error())
		return apperr.Respond(ctx, shopUserAppError(err, requestLanguage(ctx)))
	}
	ctx.Response(http.StatusOK,
		common.ApiResponse{
			Success: true,
		})

	return nil
}

// memberNULError is the 400 for a member save whose text carries U+0000 (usually pasted from a
// PDF), or nil. PostgreSQL TEXT cannot store it, so the save used to fail as a 500 "try again".
func memberNULError(req *models.UserRoleRequest, lang string) *apperr.AppError {
	path := textguard.NULField(req)
	if path == "" {
		return nil
	}
	message := func(lang string) string {
		return textguard.Message(path, memberNULLabels, func(key string) string { return language.Text(key, lang) })
	}
	return apperr.ErrValidation.WithMessage(message(lang)).WithThaiMessage(message("th")).WithField(path)
}

// memberNULLabels names the member form fields (languages.tsv keys) in the NUL message.
var memberNULLabels = map[string]string{
	"username": "user_code", "editusername": "user_code", "email": "email", "userprofilename": "user_name",
	"position": "user_position", "department": "department", "avatar": "image", "avatarthumb": "image",
	"lineuserid": "lineuserid", "permissionsets": "ss_f_permission_set_code",
}

// requestLanguage is the caller's language for user-facing messages (query lang, then
// Accept-Language).
func requestLanguage(ctx microservice.IContext) string {
	if lang := strings.TrimSpace(ctx.QueryParam("lang")); lang != "" {
		return lang
	}
	return ctx.Header("Accept-Language")
}

// localizedAppError is base with the languages.tsv row key rendered in lang as its message
// (Thai in message_th), so the settings screen can show it as-is.
func localizedAppError(base *apperr.AppError, key string, lang string) *apperr.AppError {
	return base.WithMessage(language.Text(key, lang)).WithThaiMessage(language.Text(key, "th"))
}

// shopUserAppError maps a membership service error to the response the settings screen shows:
// known rejections get a plain-language message that says what to do next; anything else is
// INTERNAL and the cause stays in the server log, never in the response.
func shopUserAppError(err error, lang string) *apperr.AppError {
	var scopeErr *accessScopeError
	if !errors.As(err, &scopeErr) {
		if appErr := apperr.FromError(err); appErr != nil {
			return appErr
		}
	}
	base, key, field := shopUserErrorRow(err)
	appErr := localizedAppError(base, key, lang)
	if field != "" {
		appErr = appErr.WithField(field)
	}
	// The cause is for the server log only (ToResponse never includes it).
	return appErr.WithWrap(err)
}

// errLoginExistsBase is its own errorcode: the settings screen opens the "attach the existing
// login account" dialog only for this answer (offersExistingLogin), not for DUPLICATE.
var errLoginExistsBase = apperr.New("LOGIN_EXISTS", http.StatusConflict, "login account exists", "มีบัญชีเข้าระบบของรหัสผู้ใช้นี้อยู่แล้ว")

// memberSaveRejections are the save errors that name the form field to fix.
var memberSaveRejections = []struct {
	err   error
	base  *apperr.AppError
	key   string
	field string
}{
	{errMemberAlreadyExists, apperr.ErrDuplicate, "ss_err_user_already_member", "username"},
	{errUsernameTaken, apperr.ErrDuplicate, "ss_err_user_code_taken", "username"},
	{errLoginExists, errLoginExistsBase, "ss_err_user_login_exists", "username"},
	{errUserCodeLocked, apperr.ErrConflict, "ss_err_user_code_locked", "username"},
	{errUserEmailLocked, apperr.ErrConflict, "ss_err_user_email_locked", "email"},
	{errCreatorAccessExpiry, apperr.ErrConflict, "ss_err_creator_access_expiry", "accessexpirydate"},
	{models.ErrInvalidAccessExpiryDate, apperr.ErrValidation, "ss_err_access_expiry_date_invalid", "accessexpirydate"},
	{errSaveTargetNotFound, apperr.ErrNotFound, "ss_err_not_found", ""},
}

// shopUserErrorRow is the apperr status, languages.tsv row and form field for a membership
// service error; an unexpected error is INTERNAL / "try again".
func shopUserErrorRow(err error) (base *apperr.AppError, key string, field string) {
	var scopeErr *accessScopeError
	if errors.As(err, &scopeErr) {
		return scopeErr.status, scopeErr.key, "accessscopes"
	}
	for _, rejection := range memberSaveRejections {
		if errors.Is(err, rejection.err) {
			return rejection.base, rejection.key, rejection.field
		}
	}
	if errors.Is(err, errMemberRequestInvalid) {
		return apperr.ErrValidation, "ss_err_invalid_data", ""
	}
	if errors.Is(err, ErrShopUserNotFound) {
		// Only the caller's own membership lookup (requireHoldingManager) surfaces this.
		return apperr.ErrForbidden, "ss_err_no_permission", ""
	}
	switch err.Error() {
	case "permission denied":
		return apperr.ErrForbidden, "ss_err_no_permission", ""
	case "user not found":
		return apperr.ErrNotFound, "ss_err_not_found", ""
	case "can not edit self permission":
		return apperr.ErrConflict, "ss_err_cannot_change_own_role", "role"
	case "can't delete your permission":
		return apperr.ErrConflict, "ss_err_cannot_delete_self", ""
	case "creator_cannot_delete":
		return apperr.ErrConflict, err.Error(), ""
	case "creator_access_cannot_be_disabled":
		return apperr.ErrConflict, err.Error(), "isaccessdisabled"
	}
	return apperr.ErrInternal, "ss_err_try_again", ""
}

// hydrateAccessScopes validates the saved scope rules against this Holding and resolves
// their immutable ids (see hydrateAccessScopes in scopes_postgres.go): "company"/"branch"
// rules must name a company/branch of this Holding; a "holding" rule is stored as-is
// and covers every active company, including companies created later.
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
		return apperr.Respond(ctx, shopUserAppError(errors.New("permission denied"), requestLanguage(ctx)))
	}

	username := strings.TrimSpace(decodePathUsername(ctx.Param("username")))

	// Debug log
	h.ms.Logger.Debug("DeleteUserPermissionShop - target username: " + username)

	if len(username) < 1 {
		h.ms.Logger.Error("DeleteUserPermissionShop - username is empty!")
		return apperr.Respond(ctx, localizedAppError(apperr.ErrValidation, "ss_err_invalid_data", requestLanguage(ctx)).WithField("username"))
	}

	err := svc.DeleteUserPermissionShop(holdingCode, authUsername, username)

	if err != nil {
		h.ms.Logger.Error("DeleteUserPermissionShop - error: " + err.Error())
		return apperr.Respond(ctx, shopUserAppError(err, requestLanguage(ctx)))
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
