package authentication

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"smlcloudplatform/internal/authentication/models"
	"smlcloudplatform/internal/authentication/repositories"
	"smlcloudplatform/internal/authentication/services"
	"smlcloudplatform/internal/centraldb"
	"smlcloudplatform/internal/config"
	"smlcloudplatform/internal/demo"
	"smlcloudplatform/internal/goapi/language"
	"smlcloudplatform/internal/logger"
	common "smlcloudplatform/internal/models"
	orgaccess "smlcloudplatform/internal/organization"
	companyModels "smlcloudplatform/internal/organization/company/models"
	"smlcloudplatform/internal/shop"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/apperr"
	"smlcloudplatform/pkg/microservice"
	msModels "smlcloudplatform/pkg/microservice/models"
	"strconv"
	"strings"
	"time"
)

const devLoginSecretHeader = "X-BC-Dev-Login-Secret"

type devLoginConfig struct {
	userUID string
	secret  string
}

type IAuthenticationHttp interface {
	Login(ctx microservice.IContext) error
	Poslogin(ctx microservice.IContext) error
	Register(ctx microservice.IContext) error
	Logout(ctx microservice.IContext) error
	Profile(ctx microservice.IContext) error
	DisableUser(ctx microservice.IContext) error
	// DeleteUser(ctx microservice.IContext) error
	RegisterHttp()
}
type AuthenticationHttp struct {
	ms                    *microservice.Microservice
	cfg                   config.IConfig
	db                    *sql.DB
	authService           *microservice.AuthService
	authenticationService services.IAuthenticationService
	shopService           shop.IShopService
	shopUserService       shop.IShopUserService
}

func NewAuthenticationHttp(ms *microservice.Microservice, cfg config.IConfig) IAuthenticationHttp {

	cache := ms.Cacher()
	db, err := centraldb.Open()
	if err != nil {
		// PostgreSQL is the only user store; without it no request can be served.
		logger.GetLogger().Fatalf("authentication: %v", err)
	}
	// db also drives live authorization: workspace selection re-checks membership and scopes.
	authService := microservice.NewAuthService(cache, 24*3*time.Hour, 24*30*time.Hour, db)
	authRepo := repositories.NewAuthenticationPostgresRepository(db)
	shopRepo := shop.NewShopPostgresRepository(db)
	shopUserRepo := shop.NewShopUserPostgresRepository(db)
	shopUserAccessLogRepo := shop.NewShopUserAccessLogPostgresRepository(db)
	smsRepo := repositories.NewAuthenticationSMSRepository(cache)
	authenticationService := services.NewAuthenticationService(
		authRepo,
		shopUserRepo,
		shopUserAccessLogRepo,
		smsRepo,
		authService,
		utils.RandStringBytesMaskImprSrcUnsafe,
		utils.RandNumber,
		utils.NewGUID,
		utils.HashPassword,
		utils.CheckHashPassword,
		ms.TimeNow)

	shopService := shop.NewShopService(shopRepo, shopUserRepo, ms.TimeNow)
	shopUserService := shop.NewShopUserService(shopUserRepo)
	return AuthenticationHttp{
		ms:                    ms,
		cfg:                   cfg,
		db:                    db,
		authService:           authService,
		authenticationService: authenticationService,
		shopUserService:       shopUserService,
		shopService:           shopService,
	}
}

func currentDevLoginConfig() (devLoginConfig, bool) {
	rawEnvironment := strings.TrimSpace(os.Getenv("BC_ENV"))
	return devLoginConfigFor(
		strings.ToLower(rawEnvironment),
		rawEnvironment != "",
		os.Getenv("BCAI_DEV_LOGIN_ENABLED"),
		os.Getenv("BCAI_DEV_LOGIN_USER_UID"),
		os.Getenv("BCAI_DEV_LOGIN_SECRET"),
	)
}

func devLoginConfigFor(dataEnvironment string, environmentConfigured bool, enabled string, userUID string, secret string) (devLoginConfig, bool) {
	userUID = strings.TrimSpace(userUID)
	if !environmentConfigured || dataEnvironment != config.DataEnvironmentDev ||
		!strings.EqualFold(strings.TrimSpace(enabled), "true") || userUID == "" || len(secret) < 32 {
		return devLoginConfig{}, false
	}
	return devLoginConfig{userUID: userUID, secret: secret}, true
}

func devLoginSecretMatches(expected string, provided string) bool {
	return subtle.ConstantTimeCompare([]byte(expected), []byte(provided)) == 1
}

func (h AuthenticationHttp) RegisterHttp() {

	h.ms.POST("/login", h.Login)
	h.ms.POST("/googlelogin", h.GoogleLogin)
	h.ms.POST("/logout", h.Logout)
	h.ms.POST("/refresh", h.RefreshToken)
	if _, enabled := currentDevLoginConfig(); enabled {
		h.ms.POST("/dev-login", h.DevLogin)
	}
	if demo.Enabled() {
		h.ms.POST("/demo-login", h.DemoLogin)
	}

	h.ms.GET("/verify-token", h.VerifyToken)
	h.ms.GET("/sessions/active-count", h.SessionsActiveCount)
	// ไม่อยู่ใน exceptShopPath: ต้องเลือก Holding แล้ว middleware จึงให้บริษัท/สาขาของ session มาครบ
	h.ms.GET("/session/selection", h.SessionSelection)

	h.ms.GET("/profile", h.Profile)
	h.ms.PUT("/profile/disable-user", h.DisableUser)
	// h.ms.DELETE("/profile/delete-user", h.DeleteUser)
	h.ms.GET("/profileshop", h.ProfileShop)

	h.ms.PUT("/profile", h.Update)
	h.ms.PUT("/profile/password", h.UpdatePassword)
	h.ms.PUT("/profile/link-line", h.LinkLine)
	h.ms.DELETE("/profile/link-line", h.UnlinkLine)

	middlewareShop := h.authService.MWFuncWithShop(h.ms.Cacher())
	h.ms.GET("/list-holding", h.ListShopCanAccess, middlewareShop)
	h.ms.GET("/list-shop", h.ListShopCanAccess, middlewareShop)
	h.ms.POST("/select-holding", h.SelectShop, middlewareShop)
	h.ms.POST("/select-shop", h.SelectShop, middlewareShop)
	h.ms.PUT("/favorite-holding", h.UpdateShopFavorite, middlewareShop)
	h.ms.PUT("/favorite-shop", h.UpdateShopFavorite, middlewareShop)

	shopHttp := shop.NewShopHttp(h.ms, h.cfg)
	h.ms.POST("/create-holding", shopHttp.CreateShop, middlewareShop)
	h.ms.POST("/create-shop", shopHttp.CreateShop, middlewareShop)
}

func (h AuthenticationHttp) DevLogin(ctx microservice.IContext) error {
	devConfig, enabled := currentDevLoginConfig()
	if !enabled || !devLoginSecretMatches(devConfig.secret, ctx.Header(devLoginSecretHeader)) {
		return apperr.Respond(ctx, apperr.ErrUnauthorized.WithMessage("dev login failed"))
	}

	result, err := h.authenticationService.DevLoginByUID(devConfig.userUID, models.AuthenticationContext{Ip: ctx.RealIp()})
	if err != nil {
		if errors.Is(err, &models.UserDisableLoginError{}) {
			return apperr.Respond(ctx, apperr.ErrDisabled.WithMessage("user is disabled"))
		}
		if appErr := apperr.FromError(err); appErr != nil {
			return apperr.Respond(ctx, appErr)
		}
		return apperr.Respond(ctx, apperr.ErrUnauthorized.WithMessage("dev login failed"))
	}

	ctx.Response(http.StatusOK, map[string]interface{}{
		"success": true,
		"token":   result.Token,
		"refresh": result.Refresh,
	})
	return nil
}

// Login with phone number
// @Description Login with phone number
// @Tags		Authentication
// @Param		UserLoginPhoneNumberRequest  body      models.UserLoginPhoneNumberRequest  true  "User Login PhoneNumber Request"
// @Accept 		json
// @Success		200	{object}	common.AuthResponse
// @Failure		400 {object}	common.AuthResponseFailed
// @Router /login [post]
func (h AuthenticationHttp) LoginWithPhoneNumber(ctx microservice.IContext) error {

	input := ctx.ReadInput()

	userReq := &models.UserLoginPhoneNumberRequest{}
	err := json.Unmarshal([]byte(input), &userReq)

	if err != nil {
		return apperr.Respond(ctx, apperr.ErrBadRequest.WithMessage("user payload invalid"))
	}

	if err = ctx.Validate(userReq); err != nil {
		return apperr.Respond(ctx, apperr.ErrValidation.WithWrap(err))
	}

	authContext := models.AuthenticationContext{
		Ip: ctx.RealIp(),
	}

	result, err := h.authenticationService.LoginWithPhoneNumber(userReq, authContext)

	if err != nil {
		return apperr.Respond(ctx, apperr.ErrUnauthorized.WithWrap(err).WithMessage("login failed."))
	}

	ctx.Response(http.StatusOK, map[string]interface{}{
		"success": true,
		"token":   result.Token,
		"refresh": result.Refresh,
	})

	return nil
}

// Login login
// @Description get struct array by ID
// @Tags		Authentication
// @Param		User  body      models.UserLoginRequest  true  "User Account"
// @Accept 		json
// @Success		200	{object}	common.AuthResponse
// @Failure		400 {object}	common.AuthResponseFailed
// @Router /login [post]
func (h AuthenticationHttp) Login(ctx microservice.IContext) error {

	// Never log the login body: it contains the password (removed 2026-09-24).
	input := ctx.ReadInput()

	userReq := &models.UserLoginRequest{}
	err := json.Unmarshal([]byte(input), &userReq)

	if err != nil {
		// Error type only: a JSON syntax error quotes a character of the body, which may be the password.
		logger.GetLogger().Errorf("login payload unmarshal failed: %T", err)
		return apperr.Respond(ctx, apperr.ErrBadRequest.WithMessage("user payload invalid"))
	}

	if err = ctx.Validate(userReq); err != nil {
		return apperr.Respond(ctx, apperr.ErrValidation.WithWrap(err))
	}

	authContext := models.AuthenticationContext{
		Ip: ctx.RealIp(),
	}

	result, err := h.authenticationService.Login(userReq, authContext)

	if err != nil {

		if errors.Is(err, &models.UserDisableLoginError{}) {
			return apperr.Respond(ctx, apperr.ErrDisabled.WithMessage("user is disabled"))
		}
		return apperr.Respond(ctx, apperr.ErrUnauthorized.WithWrap(err).WithMessage("login failed."))
	}

	ctx.Response(http.StatusOK, map[string]interface{}{
		"success": true,
		"token":   result.Token,
		"refresh": result.Refresh,
	})

	return nil
}

// Login poslogin
// @Description get struct array by ID
// @Tags		Authentication
// @Param		User  body      models.PosLoginRequest  true  "User Account"
// @Accept 		json
// @Success		200	{object}	common.AuthResponse
// @Failure		400 {object}	common.AuthResponseFailed
func (h AuthenticationHttp) Poslogin(ctx microservice.IContext) error {

	input := ctx.ReadInput()

	userReq := &models.PosLoginRequest{}
	err := json.Unmarshal([]byte(input), &userReq)

	if err != nil {
		return apperr.Respond(ctx, apperr.ErrBadRequest.WithMessage("user payload invalid"))
	}

	if err = ctx.Validate(userReq); err != nil {
		return apperr.Respond(ctx, apperr.ErrValidation.WithWrap(err))
	}

	authContext := models.AuthenticationContext{
		Ip: ctx.RealIp(),
	}

	result, err := h.authenticationService.Poslogin(userReq, authContext)

	if err != nil {
		return apperr.Respond(ctx, apperr.ErrUnauthorized.WithWrap(err).WithMessage("login failed."))
	}

	ctx.Response(http.StatusOK, map[string]interface{}{
		"success": true,
		"token":   result.Token,
		"refresh": result.Refresh,
	})

	return nil
}

// Login refresh
// @Description refresh token
// @Tags		Authentication
// @Param		TokenLoginRequest  body      models.TokenLoginRequest  true  "Reresh Token"
// @Accept 		json
// @Success		200	{object}	common.AuthResponse
// @Failure		400 {object}	common.AuthResponseFailed
// @Router /refresh [post]
func (h AuthenticationHttp) RefreshToken(ctx microservice.IContext) error {

	input := ctx.ReadInput()

	reqBody := models.TokenLoginRequest{}
	err := json.Unmarshal([]byte(input), &reqBody)

	if err != nil {
		return apperr.Respond(ctx, apperr.ErrBadRequest.WithMessage("payload invalid"))
	}

	if err = ctx.Validate(reqBody); err != nil {
		return apperr.Respond(ctx, apperr.ErrValidation.WithWrap(err))
	}

	result, err := h.authenticationService.RefreshToken(reqBody)

	if err != nil {
		return apperr.Respond(ctx, apperr.ErrUnauthorized.WithWrap(err).WithMessage("login failed."))
	}

	ctx.Response(http.StatusOK, map[string]interface{}{
		"success": true,
		"token":   result.Token,
		"refresh": result.Refresh,
	})

	return nil
}

// Google Login สำหรับ mobile (Android/iOS)
// @Description Login ด้วย Google email จาก Google OAuth บน mobile
// @Tags		Authentication
// @Param		GoogleLoginRequest  body  models.GoogleLoginRequest  true  "Google Login Data"
// @Accept		json
// @Success		200	{object}	common.AuthResponse
// @Failure		400	{object}	common.AuthResponseFailed
// @Router /googlelogin [post]
func (h AuthenticationHttp) GoogleLogin(ctx microservice.IContext) error {
	input := ctx.ReadInput()

	req := &models.GoogleLoginRequest{}
	err := json.Unmarshal([]byte(input), req)
	if err != nil {
		return apperr.Respond(ctx, apperr.ErrBadRequest.WithMessage("payload invalid"))
	}

	// SECURITY (2026-06-21): require + verify a real Google ID token. Never trust a
	// caller-supplied email — derive the trusted email from the verified token claims.
	// This closes the account-takeover hole where posting any email minted a token.
	if req.Credential == "" {
		return apperr.Respond(ctx, apperr.ErrUnauthorized.WithMessage("google credential required"))
	}
	claims, err := verifyGoogleIDToken(req.Credential, h.cfg.GoogleClientId())
	if err != nil {
		return apperr.Respond(ctx, apperr.ErrUnauthorized.WithWrap(err).WithMessage("google token verification failed"))
	}

	result, err := h.authenticationService.LoginWithGoogleIdentity(claims.Iss, claims.Sub, claims.Email, claims.emailVerified(), claims.Name)
	if err != nil {
		if errors.Is(err, &models.UserDisableLoginError{}) {
			return apperr.Respond(ctx, apperr.ErrDisabled.WithMessage("user is disabled"))
		}
		if errors.Is(err, repositories.ErrGooglePrecreatedAmbiguous) {
			logger.GetLogger().Warnf("google login refused: %v", err) // account ids for the admin; no email in the log
			return apperr.Respond(ctx, apperr.ErrConflict.WithMessage(language.Text("auth_err_google_precreated_ambiguous", authRequestLanguage(ctx))).
				WithThaiMessage(language.Text("auth_err_google_precreated_ambiguous", "th")))
		}
		return apperr.Respond(ctx, apperr.ErrUnauthorized.WithWrap(err).WithMessage("login failed."))
	}

	ctx.Response(http.StatusOK, map[string]interface{}{
		"success": true,
		"token":   result.Token,
		"refresh": result.Refresh,
	})

	return nil
}

// googleTokenInfo holds the verified claims returned by Google's tokeninfo endpoint.
type googleTokenInfo struct {
	Aud           string `json:"aud"`
	Iss           string `json:"iss"`
	Exp           string `json:"exp"`
	Email         string `json:"email"`
	EmailVerified string `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	Sub           string `json:"sub"`
}

// verifyGoogleIDToken validates a Google ID token (JWT) via Google's tokeninfo endpoint
// (which verifies the signature), then enforces audience, issuer, expiry and a verified
// email. Returns the verified claims, or an error if the token is not trustworthy.
func verifyGoogleIDToken(credential string, clientID string) (*googleTokenInfo, error) {
	if clientID == "" {
		return nil, errors.New("google client id not configured")
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://oauth2.googleapis.com/tokeninfo?id_token=" + url.QueryEscape(credential))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tokeninfo status %d", resp.StatusCode)
	}
	var info googleTokenInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}
	if info.Aud != clientID {
		return nil, errors.New("google token audience mismatch")
	}
	if info.Iss != "accounts.google.com" && info.Iss != "https://accounts.google.com" {
		return nil, errors.New("invalid google token issuer")
	}
	expUnix, err := strconv.ParseInt(info.Exp, 10, 64)
	if err != nil || time.Now().Unix() >= expUnix {
		return nil, errors.New("google token expired")
	}
	if info.Email == "" || !info.emailVerified() {
		return nil, errors.New("google email not verified")
	}
	return &info, nil
}

// emailVerified is Google's email_verified claim ("true"/"1" from tokeninfo).
func (info googleTokenInfo) emailVerified() bool {
	return info.EmailVerified == "true" || info.EmailVerified == "1"
}

// Register Member godoc
// @Summary		Register An Account
// @Description	For User Register Application
// @Tags		Authentication
// @Param		RegisterEmailRequest  body      models.RegisterEmailRequest  true  "Register account"
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		400 {object}	common.AuthResponseFailed
// @Accept 		json
func (h AuthenticationHttp) Register(ctx microservice.IContext) error {
	h.ms.Logger.Debug("Receive Register Data")
	input := ctx.ReadInput()

	userReq := models.RegisterEmailRequest{}
	err := json.Unmarshal([]byte(input), &userReq)

	if err != nil {
		return apperr.Respond(ctx, apperr.ErrBadRequest.WithMessage("user payload invalid"))
	}

	if err = ctx.Validate(userReq); err != nil {
		return apperr.Respond(ctx, apperr.ErrValidation.WithWrap(err))
	}

	idx, err := h.authenticationService.Register(userReq)

	if err != nil {
		return apperr.RespondErr(ctx, err)
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
		ID:      idx,
	})

	return nil
}

// RegisterByUsername — สมัครสมาชิกด้วยรหัสพนักงาน + รหัสผ่าน (ไม่ต้องมี email)
// @Summary		Register By Username
// @Description	สมัครสมาชิกด้วยรหัสพนักงาน + รหัสผ่าน
// @Tags		Authentication
// @Param		RegisterUsernameRequest  body      models.RegisterUsernameRequest  true  "Register by username"
// @Success		200	{object}	common.ApiResponse
// @Failure		400 {object}	common.AuthResponseFailed
// @Accept 		json
func (h AuthenticationHttp) RegisterByUsername(ctx microservice.IContext) error {
	h.ms.Logger.Debug("สมัครสมาชิกด้วยรหัสพนักงาน")
	input := ctx.ReadInput()

	userReq := models.RegisterUsernameRequest{}
	err := json.Unmarshal([]byte(input), &userReq)

	if err != nil {
		return apperr.Respond(ctx, apperr.ErrBadRequest.WithMessage("user payload invalid"))
	}

	if err = ctx.Validate(userReq); err != nil {
		return apperr.Respond(ctx, apperr.ErrValidation.WithWrap(err))
	}

	idx, err := h.authenticationService.RegisterByUsername(userReq)

	if err != nil {
		return apperr.RespondErr(ctx, err)
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
		ID:      idx,
	})

	return nil
}

// Send Phonenumber OTP godoc
// @Summary		Send Phonenumber OTP
// @Description	For User Send Phonenumber OTP
// @Tags		Authentication
// @Param		OTPRequest  body      models.OTPRequest  true  "OTP Request"
// @Success		200	{object}	common.ApiResponse
// @Failure		400 {object}	common.AuthResponseFailed
// @Accept 		json
func (h AuthenticationHttp) SendPhoneNumberOTP(ctx microservice.IContext) error {

	input := ctx.ReadInput()

	payload := models.OTPRequest{}
	err := json.Unmarshal([]byte(input), &payload)

	if err != nil {
		return apperr.Respond(ctx, apperr.ErrBadRequest.WithMessage("user payload invalid"))
	}

	if err = ctx.Validate(payload); err != nil {
		return apperr.Respond(ctx, apperr.ErrValidation.WithWrap(err))
	}

	result, err := h.authenticationService.SendPhonenumberOTP(payload)

	if err != nil {
		return apperr.RespondErr(ctx, err)
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
		Data:    result,
	})

	return nil
}

// Register By Phonenumber  godoc
// @Summary		Register By Phonenumber
// @Description	For User Register Phonenumber
// @Tags		Authentication
// @Param		OTPRequest  body      models.OTPRequest  true  "OTP Request"
// @Success		200	{object}	common.ApiResponse
// @Failure		400 {object}	common.AuthResponseFailed
// @Accept 		json
func (h AuthenticationHttp) RegisterByPhoneNumber(ctx microservice.IContext) error {

	input := ctx.ReadInput()

	payload := models.RegisterPhoneNumberRequest{}
	err := json.Unmarshal([]byte(input), &payload)

	if err != nil {
		return apperr.Respond(ctx, apperr.ErrBadRequest.WithMessage("user payload invalid"))
	}

	if err = ctx.Validate(payload); err != nil {
		return apperr.Respond(ctx, apperr.ErrValidation.WithWrap(err))
	}

	idx, err := h.authenticationService.RegisterByPhonenumber(payload)

	if err != nil {
		return apperr.RespondErr(ctx, err)
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
		ID:      idx,
	})

	return nil
}

// Forgot Password By Phonenumber  godoc
// @Summary		Forgot Password By Phonenumber
// @Description	For User Forgot Password Phonenumber
// @Tags		Authentication
// @Param		ForgotPasswordPhoneNumberRequest  body      models.ForgotPasswordPhoneNumberRequest  true  "Forgot Password PhoneNumber Request"
// @Success		200	{object}	common.ApiResponse
// @Failure		400 {object}	common.AuthResponseFailed
// @Accept 		json
func (h AuthenticationHttp) ForgotPasswordByPhoneNumber(ctx microservice.IContext) error {
	input := ctx.ReadInput()

	payload := models.ForgotPasswordPhoneNumberRequest{}
	err := json.Unmarshal([]byte(input), &payload)

	if err != nil {
		return apperr.Respond(ctx, apperr.ErrBadRequest.WithMessage("user payload invalid"))
	}

	if err = ctx.Validate(payload); err != nil {
		return apperr.Respond(ctx, apperr.ErrValidation.WithWrap(err))
	}

	err = h.authenticationService.ForgotPasswordByPhonenumber(payload)

	if err != nil {
		return apperr.RespondErr(ctx, err)
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Register Check Exists Username godoc
// @Summary		Register Check Exists Username
// @Description	Check Exists Username
// @Tags		Authentication
// @Param		Username  body      models.UsernameField  true  "Username"
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		400 {object}	common.AuthResponseFailed
// @Accept 		json
func (h AuthenticationHttp) RegisterCheckExistUsername(ctx microservice.IContext) error {

	input := ctx.ReadInput()

	payload := models.UsernameField{}
	err := json.Unmarshal([]byte(input), &payload)

	if err != nil {
		return apperr.Respond(ctx, apperr.ErrBadRequest.WithMessage("user payload invalid"))
	}

	if err = ctx.Validate(payload); err != nil {
		return apperr.Respond(ctx, apperr.ErrValidation.WithWrap(err))
	}

	isExists, err := h.authenticationService.CheckExistsUsername(payload.Username)

	if err != nil {
		return apperr.RespondErr(ctx, err)
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
		Data:    isExists,
	})

	return nil
}

// Register Check Exists Phone Number godoc
// @Summary		Register Check Exists Phone Number
// @Description	Check Exists Phone Number
// @Tags		Authentication
// @Param		PhoneNumber  body      models.PhoneNumberField  true  "Username"
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		400 {object}	common.AuthResponseFailed
// @Accept 		json
func (h AuthenticationHttp) RegisterCheckExistPhonenumber(ctx microservice.IContext) error {

	input := ctx.ReadInput()

	payload := models.PhoneNumberField{}
	err := json.Unmarshal([]byte(input), &payload)

	if err != nil {
		return apperr.Respond(ctx, apperr.ErrBadRequest.WithMessage("user payload invalid"))
	}

	if err = ctx.Validate(payload); err != nil {
		return apperr.Respond(ctx, apperr.ErrValidation.WithWrap(err))
	}

	isExists, err := h.authenticationService.CheckExistsPhonenumber(payload.PhoneNumber)

	if err != nil {
		return apperr.RespondErr(ctx, err)
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
		Data:    isExists,
	})

	return nil
}

// Update User Profile godoc
// @Summary		Update profile
// @Description	For User Update Profile
// @Tags		Authentication
// @Param		UserProfileRequest  body      models.UserProfileRequest  true  "Update account"
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		400 {object}	common.AuthResponseFailed
// @Accept 		json
// @Router		/profile [put]
func (h AuthenticationHttp) Update(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	input := ctx.ReadInput()

	userReq := models.UserProfileRequest{}
	err := json.Unmarshal([]byte(input), &userReq)

	if err != nil {
		return apperr.Respond(ctx, apperr.ErrBadRequest.WithMessage("user payload invalid"))
	}

	if err = ctx.Validate(userReq); err != nil {
		return apperr.Respond(ctx, apperr.ErrValidation.WithWrap(err))
	}

	err = h.authenticationService.Update(userInfo.UID, userReq)

	if err != nil {
		return apperr.RespondErr(ctx, err)
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Update User Profile Password godoc
// @Summary		Update profile Password
// @Description	For User Update Profile Password
// @Tags		Authentication
// @Param		UserPasswordRequest  body      models.UserPasswordRequest  true  "Update account password"
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		400 {object}	common.AuthResponseFailed
// @Accept 		json
// @Router		/profile/password [put]
func (h AuthenticationHttp) UpdatePassword(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	input := ctx.ReadInput()

	userPwdReq := models.UserPasswordRequest{}
	err := json.Unmarshal([]byte(input), &userPwdReq)

	if err != nil {
		return apperr.Respond(ctx, apperr.ErrBadRequest.WithMessage("user payload invalid"))
	}

	if err = ctx.Validate(userPwdReq); err != nil {
		return apperr.Respond(ctx, apperr.ErrValidation.WithWrap(err))
	}

	err = h.authenticationService.UpdatePassword(authUsername, userPwdReq.CurrentPassword, userPwdReq.NewPassword)

	if err != nil {
		return apperr.RespondErr(ctx, err)
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
	})

	return nil
}

// ResetPasswordToDefault is retained for interface compatibility but is not registered as an HTTP route.
func (h AuthenticationHttp) ResetPasswordToDefault(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	targetUsername := ctx.Param("username")

	if len(targetUsername) < 1 {
		return apperr.Respond(ctx, apperr.ErrValidation.WithField("username").WithMessage("username invalid"))
	}

	err := h.authenticationService.ResetPasswordToDefault(userInfo.HoldingCode, userInfo.Username, targetUsername)
	if err != nil {
		return apperr.RespondErr(ctx, err)
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Logout
// @Description Logout Current Profile
// @Tags		Authentication
// @Accept 		json
// @Success		200	{array}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /logout [post]
func (h AuthenticationHttp) Logout(ctx microservice.IContext) error {

	authorizationHeader := ctx.Header("Authorization")

	err := h.authenticationService.Logout(authorizationHeader)

	if err != nil {
		return apperr.RespondErr(ctx, err)
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// SessionSelection — Holding/บริษัท/สาขาที่ session นี้เลือกอยู่
// BFF อ่านก่อนสลับ Holding ชั่วคราวเพื่อโหลดรายชื่อบริษัท/สาขา แล้วคืนค่าเดิมครบทั้งบริษัทและสาขา
// (เดิมจอตั้งค่าที่ไม่ส่ง businesscode ทำให้ session เหลือแต่ Holding ทุกจอบัญชีขึ้น "กรุณาเลือกบริษัท" — UAT S7 2026-09-24)
// @Summary		Current workspace selection
// @Tags		Authentication
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /session/selection [get]
func (h AuthenticationHttp) SessionSelection(ctx microservice.IContext) error {
	ctx.Response(http.StatusOK, common.ApiResponse{Success: true, Data: sessionSelectionPayload(ctx.UserInfo())})
	return nil
}

func sessionSelectionPayload(user msModels.UserInfo) map[string]string {
	return map[string]string{"holdingcode": user.HoldingCode, "businesscode": user.BusinessCode, "branchuid": user.BranchUID}
}

// Get Current Profile
// VerifyToken — ตรวจสอบ token ว่า valid หรือไม่ (สำหรับ app ภายนอก)
// @Summary		Verify Token
// @Description	ตรวจสอบว่า Bearer token ยัง valid อยู่หรือไม่ และดึงข้อมูล user
// @Tags		Authentication
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /verify-token [get]
func (h AuthenticationHttp) VerifyToken(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	ctx.Response(http.StatusOK, map[string]interface{}{
		"success":  true,
		"username": userInfo.Username,
		"uid":      userInfo.UID,
		"name":     userInfo.Name,
	})
	return nil
}

// SessionsActiveCount — จำนวนเซสชันที่กำลังใช้งานระบบ (อ่านจากตาราง cache_entries ใน PostgreSQL)
// @Description จำนวนเซสชัน login ทั้งหมด และที่ active ใน 30 นาทีหลัง แยกตามกลุ่มกิจการ
// @Tags		Authentication
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /sessions/active-count [get]
func (h AuthenticationHttp) SessionsActiveCount(ctx microservice.IContext) error {
	stats, err := h.authService.ActiveSessionStats()
	if err != nil {
		return apperr.RespondErr(ctx, err)
	}
	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    stats,
	})
	return nil
}

// @Description Get Current Profile
// @Tags		Authentication
// @Accept 		json
// @Success		200	{array}	models.UserProfileReponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /profile [get]
func (h AuthenticationHttp) Profile(ctx microservice.IContext) error {

	// stime := time.Now()
	userInfo := ctx.UserInfo()
	userProfile, err := h.authenticationService.Profile(userInfo.Username, userInfo.UID)

	if err != nil {
		return apperr.RespondErr(ctx, err)
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    userProfile,
	})

	// fmt.Println("Time Profile", time.Since(stime))
	return nil
}

// Get Current Profile
// @Description Get Current Profile
// @Tags		Authentication
// @Accept 		json
// @Success		200	{array}	models.UserProfileReponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /profileshop [get]
func (h AuthenticationHttp) ProfileShop(ctx microservice.IContext) error {

	userProfile, err := h.shopService.InfoShop(ctx.UserInfo().HoldingCode)

	if err != nil {
		return apperr.RespondErr(ctx, err)
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    userProfile,
	})
	return nil
}

// Access Shop godoc
// @Description Access to Shop
// @Tags		Authentication
// @Param		User  body      models.ShopSelectRequest  true  "Shop"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.ApiResponse
// @Security     AccessToken
// @Router /select-shop [post]
func (h AuthenticationHttp) SelectShop(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	authorizationHeader := ctx.Header("Authorization")

	input := ctx.ReadInput()

	shopSelectReq := &models.ShopSelectRequest{}
	err := json.Unmarshal([]byte(input), &shopSelectReq)

	if err != nil {
		return apperr.Respond(ctx, apperr.ErrBadRequest.WithWrap(err))
	}
	holdingCode, normalizeErr := utils.NormalizeHoldingCode(shopSelectReq.HoldingCode)
	if normalizeErr != nil || holdingCode == "" {
		return apperr.Respond(ctx, apperr.ErrValidation.WithField("holdingcode").WithMessage("holdingcode invalid"))
	}
	shopSelectReq.HoldingCode = holdingCode
	shopSelectReq.BusinessCode = companyModels.NormalizeCompanyCode(shopSelectReq.BusinessCode)
	shopSelectReq.BranchUID = strings.TrimSpace(shopSelectReq.BranchUID)
	if shopSelectReq.BusinessCode != "" {
		exists := false
		err := h.db.QueryRowContext(context.Background(), `SELECT EXISTS (SELECT 1 FROM companies
			WHERE holding_code = $1 AND UPPER(code) = UPPER($2) AND is_active = true)`,
			shopSelectReq.HoldingCode, shopSelectReq.BusinessCode).Scan(&exists)
		if err != nil {
			return apperr.Respond(ctx, apperr.ErrInternal.WithWrap(err))
		}
		if !exists {
			return apperr.Respond(ctx, apperr.ErrForbidden.WithMessage("company not found in Holding").WithThaiMessage("ไม่พบบริษัทนี้ใน Holding ที่เลือก"))
		}
	}

	authContext := models.AuthenticationContext{
		Ip: ctx.RealIp(),
	}

	err = h.authenticationService.AccessShop(shopSelectReq.HoldingCode, shopSelectReq.BusinessCode, shopSelectReq.BranchUID, authUsername, userInfo.UID, authorizationHeader, authContext)

	if err != nil {
		// Expired/disabled membership: say why, in the caller's language.
		if services.ShopAccessDeniedKey(err) != "" {
			return apperr.Respond(ctx, services.ShopAccessDenied(err, authRequestLanguage(ctx)))
		}
		return apperr.RespondErr(ctx, err)
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// List Shop godoc
// @Description List Merchant In My Account
// @Tags		Authentication
// @Accept 		json
// @Success		200	{array}	models.ShopUserInfo
// @Failure		401 {object}	common.ApiResponse
// @Security     AccessToken
// @Router /list-shop [get]
func (h AuthenticationHttp) ListShopCanAccess(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username

	pageable := utils.GetPageable(ctx.QueryParam)

	docList, pagination, err := h.shopUserService.ListShopByUser(authUsername, userInfo.UID, pageable)

	if err != nil {
		return apperr.RespondErr(ctx, err)
	}
	canCreateHolding := orgaccess.RequireEmailedAccount(context.Background(), h.db, userInfo) == nil

	ctx.Response(http.StatusOK,
		map[string]interface{}{
			"success":          true,
			"data":             docList,
			"pagination":       pagination,
			"cancreateholding": canCreateHolding,
		},
	)

	return nil
}

// Favorite Shop godoc
// @Description Favorite Shop In Account
// @Tags		Authentication
// @Accept 		json
// @Param		ShopFavoriteRequest  body      models.ShopFavoriteRequest  true  "Shop Favorite Request"
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.ApiResponse
// @Security     AccessToken
// @Router /favorite-shop [put]
func (h AuthenticationHttp) UpdateShopFavorite(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username

	input := ctx.ReadInput()

	reqBody := models.ShopFavoriteRequest{}
	err := json.Unmarshal([]byte(input), &reqBody)

	if err != nil {
		return apperr.Respond(ctx, apperr.ErrBadRequest.WithMessage("request payload invalid"))
	}

	err = h.authenticationService.UpdateFavoriteShop(reqBody.HoldingCode, authUsername, userInfo.UID, reqBody.IsFavorite)

	if err != nil {
		return apperr.RespondErr(ctx, err)
	}

	ctx.Response(http.StatusOK,
		common.ApiResponse{
			Success: true,
		},
	)

	return nil
}

// Link LINE to user profile
// @Description เชื่อมต่อ LINE กับบัญชีผู้ใช้ (ระดับ user ใช้ร่วมทุก shop)
// @Tags		Authentication
// @Param		LinkLineRequest  body      models.LinkLineRequest  true  "LINE User Info"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		400 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router		/profile/link-line [put]
func (h AuthenticationHttp) LinkLine(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	input := ctx.ReadInput()

	req := models.LinkLineRequest{}
	err := json.Unmarshal([]byte(input), &req)

	if err != nil {
		return apperr.Respond(ctx, apperr.ErrBadRequest.WithMessage("user payload invalid"))
	}

	if err = ctx.Validate(req); err != nil {
		return apperr.Respond(ctx, apperr.ErrValidation.WithWrap(err))
	}

	err = h.authenticationService.LinkLine(authUsername, req)

	if err != nil {
		return apperr.RespondErr(ctx, err)
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Unlink LINE from user profile
// @Description ยกเลิกการเชื่อมต่อ LINE จากบัญชีผู้ใช้
// @Tags		Authentication
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		400 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router		/profile/link-line [delete]
func (h AuthenticationHttp) UnlinkLine(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username

	err := h.authenticationService.UnlinkLine(authUsername)

	if err != nil {
		return apperr.RespondErr(ctx, err)
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Disable User Profile godoc
// @Summary		Disable profile
// @Description	For User Disable Profile
// @Tags		Authentication
// @Success		200	{object}	common.ApiResponse
// @Failure		400 {object}	common.AuthResponseFailed
// @Accept 		json
// @Security     AccessToken
// @Router		/profile/disable-user [put]
func (h AuthenticationHttp) DisableUser(ctx microservice.IContext) error {
	userUID := ctx.UserInfo().UID

	err := h.authenticationService.DisableUser(userUID)

	if err != nil {
		return apperr.RespondErr(ctx, err)
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// // Delete User godoc
// // @Summary		Delete User
// // @Description	For Delete User
// // @Tags		Authentication
// // @Success		200	{object}	common.ApiResponse
// // @Failure		400 {object}	common.AuthResponseFailed
// // @Accept 		json
// // @Security     AccessToken
// // @Router		/profile/delete-user [delete]
// func (h AuthenticationHttp) DeleteUser(ctx microservice.IContext) error {
// 	authUsername := ctx.UserInfo().Username

// 	err := h.authenticationService.DeleteUser(authUsername)

// 	if err != nil {
// 		ctx.ResponseError(400, err.Error())
// 		return err
// 	}

// 	ctx.Response(http.StatusOK, common.ApiResponse{
// 		Success: true,
// 	})

// 	return nil
// }

// authRequestLanguage is the caller's language for user-facing messages (query lang, then
// Accept-Language; language.Text falls back to Thai).
func authRequestLanguage(ctx microservice.IContext) string {
	if lang := strings.TrimSpace(ctx.QueryParam("lang")); lang != "" {
		return lang
	}
	return ctx.Header("Accept-Language")
}
