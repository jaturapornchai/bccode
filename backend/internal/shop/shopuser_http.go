package shop

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"smlcloudplatform/internal/authentication/models"
	"smlcloudplatform/internal/config"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
	"strings"
)

type IShopMemberHttp interface{}

type ShopMemberHttp struct {
	ms  *microservice.Microservice
	svc IShopUserService
}

func NewShopMemberHttp(ms *microservice.Microservice, cfg config.IConfig) *ShopMemberHttp {

	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	repo := NewShopUserRepository(pst)
	svc := NewShopUserService(repo)
	return &ShopMemberHttp{
		svc: svc,
		ms:  ms,
	}
}

func (h *ShopMemberHttp) RegisterHttp() {
	h.ms.GET("/user/permissions", h.ListShopUser)
	h.ms.GET("/holding/users", h.ListUserInShop)
	h.ms.GET("/shop/users", h.ListUserInShop)

	h.ms.GET("/holding/permission/:username", h.InfoShopUser)
	h.ms.GET("/shop/permission/:username", h.InfoShopUser)

	// Bulk import users into the holding from an uploaded .csv/.xlsx (base64 in JSON body).

	// Holding admin management by email (holdingcode comes from the request; the caller's role
	// is resolved per-holding so it works from the holding-selection screen, no select required).
	h.ms.GET("/holding-member/list", h.ListHoldingMembers)
	// Adding and removing members must go through the invitation lifecycle.
	// Legacy direct-grant routes stay unregistered to prevent bypassing acceptance and audit.

	// Public endpoint สำหรับ sync LINE data จาก lineoa-liff (LIFF callback)
	h.ms.POST("/line-sync", h.SyncLineData)

	// Endpoint สำหรับให้ผู้ใช้อัปเดต LINE data ของตัวเอง (ใช้จาก Flutter)
	h.ms.PUT("/profile/my-line", h.SaveMyLineData)
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
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username

	pageable := utils.GetPageable(ctx.QueryParam)

	// Role is resolved per-holding inside the service (requireHoldingManager); do NOT gate on
	// the JWT-selected userInfo.Role here — an owner/admin of this holding must always pass.
	docList, pagination, err := h.svc.ListUserInShop(holdingCode, authUsername, pageable)

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
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	if err := h.svc.EnsureHoldingManager(holdingCode, userInfo.Username); err != nil {
		ctx.Response(http.StatusForbidden, &common.ApiResponse{Success: false, Message: "permission denied"})
		return err
	}

	username := strings.TrimSpace(decodePathUsername(ctx.Param("username")))

	if len(username) < 1 {
		ctx.ResponseError(400, "username invalid")
		return nil
	}

	doc, err := h.svc.InfoShopByUser(holdingCode, username)

	if err != nil {
		ctx.ResponseError(400, "find failed")
		h.ms.Logger.Error("HTTP:: SearchShopUser " + err.Error())
		return err
	}

	// Debug log
	h.ms.Logger.Debug("InfoShopUser - username: " + username)
	h.ms.Logger.Debug("InfoShopUser - LineUserID: " + doc.LineUserID)
	h.ms.Logger.Debug("InfoShopUser - LineDisplayName: " + doc.LineDisplayName)

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    doc,
	})
	return nil
}

// List Shop By User godoc
// @Description get shopuser
// @Tags		ShopUser
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /user/permissions [get]
func (h ShopMemberHttp) ListShopUser(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username

	if userInfo.Role != models.ROLE_OWNER && userInfo.Role != models.ROLE_ADMIN {
		ctx.Response(http.StatusForbidden, &common.ApiResponse{
			Success: false,
			Message: "permission denied",
		})

		return errors.New("permission denied")
	}

	pageable := utils.GetPageable(ctx.QueryParam)

	docList, pagination, err := h.svc.ListShopByUser(authUsername, userInfo.UID, pageable)

	if err != nil {
		ctx.ResponseError(400, "find failed")
		h.ms.Logger.Error("HTTP:: SearchShopUser " + err.Error())
		return err
	}

	ctx.Response(http.StatusOK,
		common.ApiResponse{
			Success:    true,
			Pagination: pagination,
			Data:       docList,
		})
	return nil
}

type holdingMemberRequest struct {
	HoldingCode string `json:"holdingcode"`
	Email       string `json:"email"`
}

// ListHoldingMembers godoc — list members of a holding for its owner/admin (holdingcode via query).
// @Tags ShopUser
// @Security AccessToken
// @Router /holding-member/list [get]
func (h ShopMemberHttp) ListHoldingMembers(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username

	holdingCode, err := utils.NormalizeHoldingCode(ctx.QueryParam("holdingcode"))
	if err != nil || holdingCode == "" {
		ctx.ResponseError(400, "holdingcode invalid")
		return nil
	}

	pageable := utils.GetPageable(ctx.QueryParam)

	docList, pagination, err := h.svc.ListHoldingMembersByAdmin(holdingCode, authUsername, pageable)
	if err != nil {
		ctx.Response(http.StatusOK, &common.ApiResponse{Success: false, Message: err.Error()})
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{Success: true, Pagination: pagination, Data: docList})
	return nil
}

// AddHoldingMemberAdmin godoc — grant ADMIN to a holding by email (owner/admin only, idempotent).
// @Tags ShopUser
// @Security AccessToken
// @Router /holding-member/add [post]
func (h ShopMemberHttp) AddHoldingMemberAdmin(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username

	req := &holdingMemberRequest{}
	if err := json.Unmarshal([]byte(ctx.ReadInput()), req); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	holdingCode, err := utils.NormalizeHoldingCode(req.HoldingCode)
	if err != nil || holdingCode == "" {
		ctx.ResponseError(400, "holdingcode invalid")
		return nil
	}

	if err := h.svc.AddHoldingAdminByEmail(holdingCode, authUsername, req.Email); err != nil {
		ctx.Response(http.StatusOK, &common.ApiResponse{Success: false, Message: err.Error()})
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{Success: true})
	return nil
}

// RemoveHoldingMemberAdmin godoc — remove a member by email (owner protected, owner/admin only).
// @Tags ShopUser
// @Security AccessToken
// @Router /holding-member/remove [post]
func (h ShopMemberHttp) RemoveHoldingMemberAdmin(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username

	req := &holdingMemberRequest{}
	if err := json.Unmarshal([]byte(ctx.ReadInput()), req); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	holdingCode, err := utils.NormalizeHoldingCode(req.HoldingCode)
	if err != nil || holdingCode == "" {
		ctx.ResponseError(400, "holdingcode invalid")
		return nil
	}

	if err := h.svc.RemoveHoldingMember(holdingCode, authUsername, req.Email); err != nil {
		ctx.Response(http.StatusOK, &common.ApiResponse{Success: false, Message: err.Error()})
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{Success: true})
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

	// Debug log
	h.ms.Logger.Debug("SaveUserPermissionShop - input: " + input)
	h.ms.Logger.Debug("SaveUserPermissionShop - LineUserID: " + userRoleReq.LineUserID)
	h.ms.Logger.Debug("SaveUserPermissionShop - LineDisplayName: " + userRoleReq.LineDisplayName)

	// ใช้ SaveUserFullProfile เพื่อบันทึกข้อมูลทั้งหมด (รวม position, department, LINE, approval)
	err = h.svc.SaveUserFullProfile(holdingCode, authUsername, userRoleReq)
	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}
	ctx.Response(http.StatusOK,
		common.ApiResponse{
			Success: true,
		})

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

	err := h.svc.DeleteUserPermissionShop(holdingCode, authUsername, username)

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

// CleanupEmptyUsers - ลบ users ที่ username ว่างออกจากระบบ
// @Description cleanup users with empty username
// @Tags		ShopUser
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security    AccessToken
// @Router /shop/users/cleanup [delete]
func (h ShopMemberHttp) CleanupEmptyUsers(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	// ต้องเป็น Owner เท่านั้น
	if userInfo.Role != models.ROLE_OWNER && userInfo.Role != models.ROLE_ADMIN {
		ctx.Response(http.StatusForbidden, &common.ApiResponse{
			Success: false,
			Message: "permission denied",
		})
		return nil
	}

	h.ms.Logger.Debug("CleanupEmptyUsers - holdingCode: " + holdingCode)

	// ลบ users ที่ username ว่าง
	deletedCount, err := h.svc.CleanupEmptyUsers(holdingCode)

	if err != nil {
		h.ms.Logger.Error("CleanupEmptyUsers - error: " + err.Error())
		ctx.ResponseError(400, err.Error())
		return err
	}

	h.ms.Logger.Debug("CleanupEmptyUsers - deleted count: " + fmt.Sprintf("%d", deletedCount))
	ctx.Response(http.StatusOK,
		common.ApiResponse{
			Success: true,
			Message: fmt.Sprintf("Deleted %d empty users", deletedCount),
			Data:    map[string]int64{"deleted_count": deletedCount},
		})

	return nil
}

// LineSyncRequest - request body สำหรับ sync LINE data
type LineSyncRequest struct {
	HoldingCode     string `json:"holdingcode"`
	Username        string `json:"username"`
	EmployeeCode    string `json:"employeecode"` // alias for username
	LineUserID      string `json:"lineuserid"`
	LineDisplayName string `json:"linedisplayname"`
	LinePictureURL  string `json:"linepictureurl"`
	APIKey          string `json:"apikey"` // สำหรับ authentication
}

// SyncLineData - Public endpoint สำหรับ sync LINE data จาก lineoa-liff
// @Description sync LINE data from LIFF callback
// @Tags		ShopUser
// @Param		LineSyncRequest  body      LineSyncRequest  true  "LineSyncRequest"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Router /line-sync [post]
func (h ShopMemberHttp) SyncLineData(ctx microservice.IContext) error {
	input := ctx.ReadInput()

	req := &LineSyncRequest{}
	err := json.Unmarshal([]byte(input), &req)

	if err != nil {
		ctx.ResponseError(400, "Invalid request body")
		return err
	}

	configuredAPIKey := strings.TrimSpace(os.Getenv("LINE_SYNC_API_KEY"))
	requestAPIKey := strings.TrimSpace(req.APIKey)
	if configuredAPIKey == "" {
		h.ms.Logger.Error("SyncLineData - LINE_SYNC_API_KEY is not configured")
		ctx.ResponseError(http.StatusUnauthorized, "LINE sync authentication is not configured")
		return errors.New("line sync api key is not configured")
	}
	if requestAPIKey == "" || subtle.ConstantTimeCompare([]byte(requestAPIKey), []byte(configuredAPIKey)) != 1 {
		h.ms.Logger.Error("SyncLineData - unauthorized request for shop: " + req.HoldingCode)
		ctx.ResponseError(http.StatusUnauthorized, "Unauthorized")
		return errors.New("line sync unauthorized")
	}

	// ใช้ employee_code หรือ username
	username := req.Username
	if username == "" {
		username = req.EmployeeCode
	}

	if req.HoldingCode == "" || username == "" {
		ctx.ResponseError(400, "holdingcode and username/employee_code are required")
		return errors.New("missing required fields")
	}

	h.ms.Logger.Debug("SyncLineData - authenticated request for: " + username + " in shop: " + req.HoldingCode)

	// Sync LINE data
	err = h.svc.SyncLineData(req.HoldingCode, username, req.LineUserID, req.LineDisplayName, req.LinePictureURL)

	if err != nil {
		h.ms.Logger.Error("SyncLineData - error: " + err.Error())
		ctx.ResponseError(400, err.Error())
		return err
	}

	h.ms.Logger.Debug("SyncLineData - success for: " + username + " in shop: " + req.HoldingCode)
	ctx.Response(http.StatusOK,
		common.ApiResponse{
			Success: true,
			Message: "LINE data synced successfully",
		})

	return nil
}

// MyLineDataRequest - request body สำหรับ user อัปเดต LINE data ของตัวเอง
type MyLineDataRequest struct {
	LineUserID      string `json:"lineuserid"`
	LineDisplayName string `json:"linedisplayname"`
	LinePictureURL  string `json:"linepictureurl"`
}

// SaveMyLineData - ให้ผู้ใช้อัปเดต LINE data ของตัวเอง (ใช้จาก Flutter หลัง LIFF linking)
// @Description save user's own LINE data
// @Tags		ShopUser
// @Param		MyLineDataRequest  body      MyLineDataRequest  true  "MyLineDataRequest"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security    AccessToken
// @Router /profile/my-line [put]
func (h ShopMemberHttp) SaveMyLineData(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	username := userInfo.Username

	input := ctx.ReadInput()

	req := &MyLineDataRequest{}
	err := json.Unmarshal([]byte(input), &req)

	if err != nil {
		ctx.ResponseError(400, "Invalid request body")
		return err
	}

	h.ms.Logger.Debug("SaveMyLineData - user: " + username + " in shop: " + holdingCode)
	h.ms.Logger.Debug("SaveMyLineData - LineUserID: " + req.LineUserID)
	h.ms.Logger.Debug("SaveMyLineData - LineDisplayName: " + req.LineDisplayName)

	// บันทึก LINE data ของตัวเอง
	err = h.svc.SaveMyLineData(holdingCode, username, req.LineUserID, req.LineDisplayName, req.LinePictureURL)

	if err != nil {
		h.ms.Logger.Error("SaveMyLineData - error: " + err.Error())
		ctx.ResponseError(400, err.Error())
		return err
	}

	h.ms.Logger.Debug("SaveMyLineData - success for: " + username)
	ctx.Response(http.StatusOK,
		common.ApiResponse{
			Success: true,
			Message: "LINE data saved successfully",
		})

	return nil
}
