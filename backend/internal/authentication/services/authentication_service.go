package services

import (
	"context"
	"errors"
	"fmt"
	"smlcloudplatform/internal/authentication/models"
	auth_models "smlcloudplatform/internal/authentication/models"
	"smlcloudplatform/internal/authentication/repositories"
	"smlcloudplatform/internal/firebase"
	"smlcloudplatform/internal/line"
	"smlcloudplatform/internal/logger"
	"smlcloudplatform/internal/shop"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/apperr"
	"smlcloudplatform/pkg/microservice"
	"strings"
	"time"

	micromodel "smlcloudplatform/pkg/microservice/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const devLoginUserEmail = "jaturapornchai@gmail.com"

type IAuthenticationService interface {
	LoginWithPhoneNumber(userLoginReq *auth_models.UserLoginPhoneNumberRequest, authContext models.AuthenticationContext) (models.TokenLoginResponse, error)
	LoginWithPhoneNumberOTP(userLoginReq *auth_models.PhoneNumberOTPRequest, authContext models.AuthenticationContext) (models.TokenLoginResponse, error)
	Login(userReq *auth_models.UserLoginRequest, authContext models.AuthenticationContext) (models.TokenLoginResponse, error)
	DevLoginByUID(userUID string, authContext models.AuthenticationContext) (models.TokenLoginResponse, error)
	DemoLoginByUsername(username string, authContext models.AuthenticationContext) (models.TokenLoginResponse, error)
	Poslogin(userReq *auth_models.PosLoginRequest, authContext models.AuthenticationContext) (models.TokenLoginResponse, error)
	LoginEmail(userReq *auth_models.PosLoginRequest, authContext models.AuthenticationContext) (string, error)
	Register(userRequest auth_models.RegisterEmailRequest) (string, error)
	ForgotPasswordByPhonenumber(userRequest auth_models.ForgotPasswordPhoneNumberRequest) error
	Update(userUID string, userRequest auth_models.UserProfileRequest) error
	UpdatePassword(username string, currentPassword string, newPassword string) error
	ResetPasswordToDefault(holdingCode string, authUsername string, targetUsername string) error
	Logout(authorizationHeader string) error
	Profile(username string, userUID string) (auth_models.UserProfile, error)
	AccessShop(holdingCode string, businessCode string, branchUID string, username string, userUID string, authorizationHeader string, authContext models.AuthenticationContext) error
	UpdateFavoriteShop(holdingCode string, username string, userUID string, isFavorite bool) error
	LoginWithFirebaseToken(token string) (string, error)
	LoginWithLineToken(token string) (string, error)
	LoginWithLineUserID(lineUserID string, displayName string, pictureUrl string, email string) (string, string, error)
	LoginWithGoogleIdentity(issuer string, subject string, email string, displayName string) (models.TokenLoginResponse, error)
	RefreshToken(tokenRequest models.TokenLoginRequest) (models.TokenLoginResponse, error)

	LinkLine(username string, req auth_models.LinkLineRequest) error
	UnlinkLine(username string) error

	CheckExistsUsername(username string) (bool, error)
	CheckExistsPhonenumber(phoneNumber string) (bool, error)
	SendPhonenumberOTP(otpRequest auth_models.OTPRequest) (auth_models.OTPResponse, error)
	RegisterByPhonenumber(userRequest auth_models.RegisterPhoneNumberRequest) (string, error)
	RegisterByUsername(userRequest auth_models.RegisterUsernameRequest) (string, error)

	DisableUser(userUID string) error
	DeleteUser(username string) error
}

type AuthenticationService struct {
	authService           microservice.IAuthService
	authRepo              repositories.IAuthenticationMongoCacheRepository
	shopUserRepo          shop.IShopUserRepository
	shopUserAccessLogRepo shop.IShopUserAccessLogRepository
	smsRepo               repositories.IAuthenticationSMSRepository
	randdomString         func(int) string
	randdomNumber         func(int) string
	generateGUID          func() string
	passwordEncoder       func(string) (string, error)
	checkHashPassword     func(password string, hash string) bool
	timeNow               func() time.Time
	firebaseAdapter       firebase.IFirebaseAdapter
	lineAdapter           line.ILineAdapter
}

func NewAuthenticationService(
	authRepo repositories.IAuthenticationMongoCacheRepository,
	shopUserRepo shop.IShopUserRepository,
	shopUserAccessLogRepo shop.IShopUserAccessLogRepository,
	smsRepo repositories.IAuthenticationSMSRepository,
	authService microservice.IAuthService,
	randdomString func(int) string,
	randdomNumber func(int) string,
	generateGUID func() string,
	passwordEncoder func(string) (string, error),
	checkHashPassword func(password string, hash string) bool,
	timeNow func() time.Time,
	firebaseAdapter firebase.IFirebaseAdapter,
	lineAdapter line.ILineAdapter) IAuthenticationService {
	return AuthenticationService{
		authRepo:              authRepo,
		authService:           authService,
		shopUserRepo:          shopUserRepo,
		shopUserAccessLogRepo: shopUserAccessLogRepo,
		smsRepo:               smsRepo,
		randdomString:         randdomString,
		randdomNumber:         randdomNumber,
		generateGUID:          generateGUID,
		passwordEncoder:       passwordEncoder,
		checkHashPassword:     checkHashPassword,
		timeNow:               timeNow,
		firebaseAdapter:       firebaseAdapter,
		lineAdapter:           lineAdapter,
	}
}

func (svc AuthenticationService) ValidateOTP(refCode, OTP string) (bool, error) {
	return svc.smsRepo.VerifyOTP(refCode, OTP)
}

func (svc AuthenticationService) LoginWithPhoneNumberOTP(userLoginReq *auth_models.PhoneNumberOTPRequest, authContext models.AuthenticationContext) (models.TokenLoginResponse, error) {

	isOTPPassed, err := svc.ValidateOTP(userLoginReq.RefCode, userLoginReq.OTP)

	if err != nil {
		return models.TokenLoginResponse{}, errors.New("OTP invalid")
	}

	if !isOTPPassed {
		return models.TokenLoginResponse{}, errors.New("OTP invalid")
	}

	findUser, err := svc.authRepo.FindByIdentity(context.Background(), "phonenumber", userLoginReq.PhoneNumber)

	if err != nil && err.Error() != "mongo: no documents in result" {
		return models.TokenLoginResponse{}, errors.New("auth: database connect error")
	}

	if len(findUser.PhoneNumber) < 1 {
		return models.TokenLoginResponse{}, errors.New("username or password is invalid")
	}

	userInfo := svc.tokenUserInfo(*findUser)
	tokenString, refreshTokenString, err := svc.authService.CreateSession(userInfo)
	if err != nil {
		return models.TokenLoginResponse{}, errors.New("login failed")
	}

	return models.TokenLoginResponse{
		Token:   tokenString,
		Refresh: refreshTokenString,
	}, nil
}

func (svc AuthenticationService) LoginWithPhoneNumber(userLoginReq *auth_models.UserLoginPhoneNumberRequest, authContext models.AuthenticationContext) (models.TokenLoginResponse, error) {

	userLoginReq.HoldingCode = strings.TrimSpace(userLoginReq.HoldingCode)

	findUser, err := svc.authRepo.FindByPhonenumber(context.Background(), userLoginReq.PhoneNumberField)

	if err != nil && err.Error() != "mongo: no documents in result" {
		return models.TokenLoginResponse{}, errors.New("auth: database connect error")
	}

	if len(findUser.PhoneNumber) < 1 {
		return models.TokenLoginResponse{}, errors.New("username or password is invalid")
	}

	passwordInvalid := !svc.checkHashPassword(userLoginReq.Password, findUser.Password)

	if passwordInvalid {
		return models.TokenLoginResponse{}, errors.New("username or password is invalid")
	}

	holdingCode, err := svc.resolveLoginHoldingCode(context.Background(), userLoginReq.HoldingCode)
	if err != nil {
		return models.TokenLoginResponse{}, err
	}

	resultLogin, err := svc.processUserLogin(*findUser, holdingCode, authContext)

	if err != nil {
		return models.TokenLoginResponse{}, err
	}

	return resultLogin, nil
}

func (svc AuthenticationService) Login(userLoginReq *auth_models.UserLoginRequest, authContext models.AuthenticationContext) (models.TokenLoginResponse, error) {

	userLoginReq.Username = auth_models.NormalizeUsercode(userLoginReq.Username)
	userLoginReq.HoldingCode = strings.TrimSpace(userLoginReq.HoldingCode)
	if !auth_models.IsValidUsercode(userLoginReq.Username) || !auth_models.IsValidPasswordLength(userLoginReq.Password) {
		return models.TokenLoginResponse{}, errors.New("username or password is invalid")
	}

	findUser, err := svc.authRepo.FindUser(context.Background(), userLoginReq.Username)

	if err != nil && err.Error() != "mongo: no documents in result" {
		// svc.ms.Log("Authentication service", err.Error())
		return models.TokenLoginResponse{}, errors.New("auth: database connect error")
	}

	if len(findUser.Username) < 1 {
		return models.TokenLoginResponse{}, errors.New("username or password is invalid")
	}

	if !findUser.DisabledAt.IsZero() {
		return models.TokenLoginResponse{}, &auth_models.UserDisableLoginError{}
	}

	passwordInvalid := !svc.checkHashPassword(userLoginReq.Password, findUser.Password)

	if passwordInvalid {
		return models.TokenLoginResponse{}, errors.New("username or password is invalid")
	}

	holdingCode, err := svc.resolveLoginHoldingCode(context.Background(), userLoginReq.HoldingCode)
	if err != nil {
		return models.TokenLoginResponse{}, err
	}

	resultLogin, err := svc.processUserLogin(*findUser, holdingCode, authContext)

	if err != nil {
		return models.TokenLoginResponse{}, err
	}

	return resultLogin, nil
}

// DemoLoginByUsername signs in the public demo account (no password, no Google
// identity). Same session/audit pipeline as a normal login; audit action DEMO_LOGIN.
func (svc AuthenticationService) DemoLoginByUsername(username string, authContext models.AuthenticationContext) (models.TokenLoginResponse, error) {
	username = strings.ToLower(strings.TrimSpace(username))
	if username == "" {
		return models.TokenLoginResponse{}, apperr.ErrUnauthorized.WithMessage("demo login failed")
	}
	ctx := context.Background()
	user, err := svc.authRepo.FindUser(ctx, username)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return models.TokenLoginResponse{}, apperr.ErrInternal.WithWrap(err)
	}
	if err != nil || user == nil || strings.TrimSpace(user.UID) == "" {
		// First demo login in this environment: create the demo account (usercode
		// login with a random password nobody knows; sample data is seeded later).
		if _, registerErr := svc.RegisterByUsername(auth_models.RegisterUsernameRequest{
			UsernameField: auth_models.UsernameField{Username: username},
			UserPassword:  auth_models.UserPassword{Password: svc.generateGUID() + svc.generateGUID()},
			UserDetail:    auth_models.UserDetail{Name: "บัญชีทดลองใช้ (Demo)", RegisterType: "demo"},
		}); registerErr != nil {
			return models.TokenLoginResponse{}, apperr.ErrInternal.WithWrap(registerErr)
		}
		if user, err = svc.authRepo.FindUser(ctx, username); err != nil || user == nil {
			return models.TokenLoginResponse{}, apperr.ErrUnauthorized.WithMessage("demo login failed")
		}
	}
	if user.IsDeleted {
		return models.TokenLoginResponse{}, apperr.ErrUnauthorized.WithMessage("demo login failed")
	}
	if !user.DisabledAt.IsZero() {
		return models.TokenLoginResponse{}, &auth_models.UserDisableLoginError{}
	}
	result, err := svc.processUserLogin(*user, "", authContext)
	if err != nil {
		return models.TokenLoginResponse{}, err
	}
	audit := auth_models.AuthAudit{
		AuditUID:   svc.generateGUID(),
		UserUID:    user.UID,
		Action:     "DEMO_LOGIN",
		Outcome:    "SUCCESS",
		OccurredAt: svc.timeNow().UTC(),
	}
	if err := svc.authRepo.CreateAuthAudit(ctx, audit); err != nil {
		_ = svc.authService.RevokeSession("Bearer " + result.Token)
		return models.TokenLoginResponse{}, apperr.ErrInternal.WithWrap(err)
	}
	return result, nil
}

func (svc AuthenticationService) DevLoginByUID(userUID string, authContext models.AuthenticationContext) (models.TokenLoginResponse, error) {
	userUID = strings.TrimSpace(userUID)
	if userUID == "" {
		return models.TokenLoginResponse{}, apperr.ErrUnauthorized.WithMessage("dev login failed")
	}

	ctx := context.Background()
	user, err := svc.authRepo.FindUserByUID(ctx, userUID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return models.TokenLoginResponse{}, apperr.ErrUnauthorized.WithMessage("dev login failed")
		}
		return models.TokenLoginResponse{}, apperr.ErrInternal.WithWrap(err)
	}
	if user == nil || strings.TrimSpace(user.UID) == "" || user.IsDeleted {
		return models.TokenLoginResponse{}, apperr.ErrUnauthorized.WithMessage("dev login failed")
	}
	if !strings.EqualFold(strings.TrimSpace(user.Email), devLoginUserEmail) {
		return models.TokenLoginResponse{}, apperr.ErrUnauthorized.WithMessage("dev login failed")
	}
	if !user.DisabledAt.IsZero() {
		return models.TokenLoginResponse{}, &auth_models.UserDisableLoginError{}
	}

	result, err := svc.processUserLogin(*user, "", authContext)
	if err != nil {
		return models.TokenLoginResponse{}, err
	}
	audit := auth_models.AuthAudit{
		AuditUID:   svc.generateGUID(),
		UserUID:    user.UID,
		Action:     "DEV_LOGIN",
		Outcome:    "SUCCESS",
		OccurredAt: svc.timeNow().UTC(),
	}
	if err := svc.authRepo.CreateAuthAudit(ctx, audit); err != nil {
		_ = svc.authService.RevokeSession("Bearer " + result.Token)
		return models.TokenLoginResponse{}, apperr.ErrInternal.WithWrap(err)
	}

	return result, nil
}

func (svc AuthenticationService) Poslogin(userLoginReq *auth_models.PosLoginRequest, authContext models.AuthenticationContext) (models.TokenLoginResponse, error) {

	userLoginReq.Username = utils.NormalizeUsername(userLoginReq.Username)

	userLoginReq.Username = strings.TrimSpace(userLoginReq.Username)
	userLoginReq.HoldingCode = strings.TrimSpace(userLoginReq.HoldingCode)

	findUser, err := svc.authRepo.FindUser(context.Background(), userLoginReq.Username)

	if err != nil && err.Error() != "mongo: no documents in result" {
		// svc.ms.Log("Authentication service", err.Error())
		return models.TokenLoginResponse{}, errors.New("auth: database connect error")
	}

	if len(findUser.Username) < 1 {
		return models.TokenLoginResponse{}, errors.New("username or password is invalid")
	}

	// passwordInvalid := !svc.checkHashPassword(userLoginReq.Password, findUser.Password)

	// if passwordInvalid {
	// 	return models.TokenLoginResponse{}, errors.New("username or password is invalid")
	// }

	holdingCode, err := svc.resolveLoginHoldingCode(context.Background(), userLoginReq.HoldingCode)
	if err != nil {
		return models.TokenLoginResponse{}, err
	}

	resultLogin, err := svc.processUserLogin(*findUser, holdingCode, authContext)

	if err != nil {
		return models.TokenLoginResponse{}, err
	}

	return resultLogin, nil
}

func (svc AuthenticationService) LoginEmail(userLoginReq *auth_models.PosLoginRequest, authContext models.AuthenticationContext) (string, error) {

	userLoginReq.Username = utils.NormalizeUsername(userLoginReq.Username)

	userLoginReq.Username = strings.TrimSpace(userLoginReq.Username)
	userLoginReq.HoldingCode = strings.TrimSpace(userLoginReq.HoldingCode)

	findUser, err := svc.authRepo.FindUser(context.Background(), userLoginReq.Username)

	if err != nil && err.Error() != "mongo: no documents in result" {
		return "", errors.New("auth: database connect error")
	}

	if len(findUser.Username) == 0 {
		// Register user if not found
		user := auth_models.UserDoc{}
		user.UID = svc.generateGUID()
		user.Username = userLoginReq.Username
		user.Email = userLoginReq.Username
		user.Password = ""
		user.UserDetail.Name = userLoginReq.Username
		user.CreatedAt = svc.timeNow()

		_, err := svc.authRepo.CreateUser(context.Background(), user)
		if err != nil {
			return "", err
		}

		findUser, err = svc.authRepo.FindUser(context.Background(), userLoginReq.Username)
		if err != nil && err.Error() != "mongo: no documents in result" {
			return "", err
		}
	}

	tokenString, err := svc.authService.GenerateTokenWithRedis(microservice.AUTHTYPE_BEARER, svc.tokenUserInfo(*findUser))

	if err != nil {
		return "", errors.New("generate token error")
	}
	return tokenString, nil
}

func (svc *AuthenticationService) processUserLogin(findUser auth_models.UserDoc, holdingCode string, authContext models.AuthenticationContext) (result models.TokenLoginResponse, resultErr error) {
	userInfo := svc.tokenUserInfo(findUser)
	tokenString, refreshTokenString, err := svc.authService.CreateSession(userInfo)
	if err != nil {
		return models.TokenLoginResponse{}, errors.New("login failed")
	}
	sessionAccepted := false
	defer func() {
		if !sessionAccepted {
			_ = svc.authService.RevokeSession("Bearer " + tokenString)
		}
	}()

	if len(holdingCode) > 0 {
		if strings.TrimSpace(findUser.UID) == "" {
			return models.TokenLoginResponse{}, errors.New("user identity invalid")
		}
		shopUser, err := svc.shopUserRepo.FindByHoldingCodeAndUserUID(context.Background(), holdingCode, findUser.UID)

		if err != nil {
			return models.TokenLoginResponse{}, err
		}

		if shopUser.ID == primitive.NilObjectID {
			return models.TokenLoginResponse{}, errors.New("holdingcode invalid")
		}

		if err = svc.ensureShopAccessAllowed(shopUser); err != nil {
			return models.TokenLoginResponse{}, err
		}

		err = svc.authService.SelectShop(microservice.AUTHTYPE_BEARER, tokenString, holdingCode, "", "", shopUser.Role)

		if err != nil {
			return models.TokenLoginResponse{}, errors.New("failed shop select")
		}

		lastAccessedAt := svc.timeNow()

		err = svc.shopUserRepo.UpdateLastAccess(context.Background(), holdingCode, shopUser.UserUID, lastAccessedAt)
		if err != nil {
			logger.GetLogger().Error(err.Error())
		}

		err = svc.shopUserAccessLogRepo.Create(context.Background(), auth_models.ShopUserAccessLog{
			HoldingCode:    holdingCode,
			Username:       findUser.Username,
			Ip:             authContext.Ip,
			LastAccessedAt: lastAccessedAt,
		})

		if err != nil {
			logger.GetLogger().Error(err.Error())
		}
	}

	sessionAccepted = true
	return models.TokenLoginResponse{Token: tokenString, Refresh: refreshTokenString}, nil
}

func (svc AuthenticationService) tokenUserInfo(user auth_models.UserDoc) micromodel.UserInfo {
	return micromodel.UserInfo{
		Username: user.Username,
		Name:     user.Name,
		UID:      user.UID,
	}
}

func (svc *AuthenticationService) resolveLoginHoldingCode(ctx context.Context, holdingCode string) (string, error) {
	holdingCode = strings.TrimSpace(holdingCode)
	holdingCode, err := utils.NormalizeHoldingCode(holdingCode)
	if err != nil {
		return "", err
	}
	if holdingCode == "" {
		return holdingCode, nil
	}
	resolvedHoldingCode, err := svc.shopUserRepo.ResolveHoldingCodeByHoldingCode(ctx, holdingCode)
	if err != nil {
		return "", err
	}
	if holdingCode != "" && holdingCode != resolvedHoldingCode {
		return "", errors.New("holdingcode mismatch")
	}
	return resolvedHoldingCode, nil
}

func (svc AuthenticationService) findShopUser(ctx context.Context, holdingCode string, userUID string) (auth_models.ShopUser, error) {
	if strings.TrimSpace(userUID) == "" {
		return auth_models.ShopUser{}, errors.New("user identity invalid")
	}
	return svc.shopUserRepo.FindByHoldingCodeAndUserUID(ctx, holdingCode, userUID)
}

func (svc *AuthenticationService) ensureShopAccessAllowed(shopUser auth_models.ShopUser) error {
	expired := !shopUser.AccessExpiryDate.IsZero() && !svc.timeNow().Before(shopUser.AccessExpiryDate)
	if !shopUser.IsDeleted && !shopUser.IsAccessDisabled && !expired {
		return nil
	}

	if expired {
		return errors.New("user_access_expired")
	}
	if shopUser.IsDeleted {
		return errors.New("user_access_revoked")
	}
	return errors.New("user_access_disabled")
}

func (svc AuthenticationService) RefreshToken(tokenRequest models.TokenLoginRequest) (models.TokenLoginResponse, error) {

	token, refreshToken, _, err := svc.authService.RefreshToken(tokenRequest.Token)

	if err != nil {
		return models.TokenLoginResponse{}, err
	}

	return models.TokenLoginResponse{
		Token:   token,
		Refresh: refreshToken,
	}, nil
}

func (svc AuthenticationService) Register(userEmailRequest auth_models.RegisterEmailRequest) (string, error) {

	userEmailRequest.Email = utils.NormalizeEmail(userEmailRequest.Email)

	userFind, err := svc.authRepo.FindByIdentity(context.Background(), "email", userEmailRequest.Email)
	if err != nil && err.Error() != "mongo: no documents in result" {
		return "", err
	}

	if len(userFind.Username) > 0 {
		return "", errors.New("username is exists")
	}

	hashPassword, err := svc.passwordEncoder(userEmailRequest.Password)

	if err != nil {
		return "", err
	}

	user := auth_models.UserDoc{}

	user.UserDetail = userEmailRequest.UserDetail

	user.UID = svc.generateGUID()
	user.Username = userEmailRequest.Email
	user.Email = userEmailRequest.Email
	user.Password = hashPassword
	user.CreatedAt = svc.timeNow()

	idx, err := svc.authRepo.CreateUser(context.Background(), user)

	if err != nil {
		return "", err
	}

	return idx.Hex(), nil
}

// RegisterByUsername — สมัครสมาชิกด้วยรหัสพนักงาน + รหัสผ่าน (ไม่ต้องมี email)
func (svc AuthenticationService) RegisterByUsername(userRequest auth_models.RegisterUsernameRequest) (string, error) {

	userRequest.Username = auth_models.NormalizeUsercode(userRequest.Username)
	if !auth_models.IsValidUsercode(userRequest.Username) {
		return "", apperr.Validation("username", "usercode must contain 3 to 64 lowercase letters, digits, dots, hyphens, or underscores")
	}
	if !auth_models.IsValidPasswordLength(userRequest.Password) {
		return "", apperr.Validation("password", "password must contain 15 to 64 characters")
	}

	userFind, err := svc.authRepo.FindByIdentity(context.Background(), "username", userRequest.Username)
	if err != nil && err.Error() != "mongo: no documents in result" {
		return "", err
	}

	if len(userFind.Username) > 0 {
		return "", errors.New("username is exists")
	}

	hashPassword, err := svc.passwordEncoder(userRequest.Password)
	if err != nil {
		return "", err
	}

	user := auth_models.UserDoc{}
	user.UserDetail = userRequest.UserDetail
	user.UID = svc.generateGUID()
	user.Username = userRequest.Username
	user.Password = hashPassword
	user.CreatedAt = svc.timeNow()

	idx, err := svc.authRepo.CreateUser(context.Background(), user)
	if err != nil {
		return "", err
	}

	return idx.Hex(), nil
}

func (svc AuthenticationService) CheckExistsUsername(username string) (bool, error) {

	username = auth_models.NormalizeUsercode(username)
	if !auth_models.IsValidUsercode(username) {
		return false, apperr.Validation("username", "usercode must contain 3 to 64 lowercase letters, digits, dots, hyphens, or underscores")
	}

	userFind, err := svc.authRepo.FindByIdentity(context.Background(), "username", username)
	if err != nil && err.Error() != "mongo: no documents in result" {
		return true, err
	}

	if len(userFind.Username) > 0 {
		return true, nil
	}

	return false, nil
}

func (svc AuthenticationService) CheckExistsPhonenumber(phoneNumber string) (bool, error) {

	phoneNumber = utils.NormalizePhonenumber(phoneNumber)

	userPhonenumberFind, err := svc.authRepo.FindByIdentity(context.Background(), "phonenumber", phoneNumber)
	if err != nil && err.Error() != "mongo: no documents in result" {
		return true, err
	}

	if len(userPhonenumberFind.PhoneNumber) > 0 {
		return true, nil
	}

	return false, nil
}

func (svc AuthenticationService) SendPhonenumberOTP(otpRequest auth_models.OTPRequest) (auth_models.OTPResponse, error) {

	otpRequest.PhoneNumber = utils.NormalizePhonenumber(otpRequest.PhoneNumber)

	fullPhoneNumber := fmt.Sprintf("%s%s", otpRequest.CountryCode, otpRequest.PhoneNumber)
	result, err := svc.smsRepo.SendOTPViaLink(fullPhoneNumber)

	if err != nil {
		return auth_models.OTPResponse{}, err
	}

	return result, nil
}

func (svc AuthenticationService) RegisterByPhonenumber(userRequest auth_models.RegisterPhoneNumberRequest) (string, error) {

	isOtpPass, err := svc.smsRepo.VerifyOTPViaLink(userRequest.OTPToken, userRequest.OTPRefCode, userRequest.OTPPin)

	if err != nil {
		return "", err
	}

	if !isOtpPass {
		return "", errors.New("otp invalid")
	}

	userRequest.PhoneNumber = utils.NormalizePhonenumber(userRequest.PhoneNumber)

	if exists, err := svc.CheckExistsUsername(userRequest.Username); err != nil {
		return "", err
	} else if exists {
		return "", errors.New("username is exists")
	}

	if exists, err := svc.CheckExistsPhonenumber(userRequest.PhoneNumber); err != nil {
		return "", err
	} else if exists {
		return "", errors.New("phonenumber is exists")
	}

	hashPassword, err := svc.passwordEncoder(userRequest.Password)

	if err != nil {
		return "", err
	}

	user := auth_models.UserDoc{}

	user.UserDetail = userRequest.UserDetail

	user.UID = svc.generateGUID()
	user.Username = userRequest.Username
	user.Email = ""
	user.Password = hashPassword
	user.PhoneNumber = userRequest.PhoneNumber
	user.RegisterType = "phone_number"

	user.CreatedAt = svc.timeNow()

	idx, err := svc.authRepo.CreateUser(context.Background(), user)

	if err != nil {
		return "", err
	}

	return idx.Hex(), nil
}

func (svc AuthenticationService) ForgotPasswordByPhonenumber(userRequest auth_models.ForgotPasswordPhoneNumberRequest) error {

	isOtpPass, err := svc.smsRepo.VerifyOTPViaLink(userRequest.OTPToken, userRequest.OTPRefCode, userRequest.OTPPin)

	if err != nil {
		return err
	}

	if !isOtpPass {
		return errors.New("otp invalid")
	}

	userRequest.PhoneNumber = utils.NormalizePhonenumber(userRequest.PhoneNumber)

	if exists, err := svc.CheckExistsPhonenumber(userRequest.PhoneNumber); err != nil {
		return err
	} else if exists {
		return errors.New("phonenumber is exists")
	}

	userFind, err := svc.authRepo.FindByPhonenumber(context.Background(), userRequest.PhoneNumberField)
	if err != nil && err.Error() != "mongo: no documents in result" {
		return err
	}

	if len(userFind.PhoneNumber) < 1 {
		return errors.New("phone number is not exists")
	}

	hashPassword, err := svc.passwordEncoder(userRequest.Password)

	if err != nil {
		return err
	}

	userFind.Password = hashPassword

	err = svc.authRepo.UpdateUser(context.Background(), userFind.Username, *userFind)

	if err != nil {
		return err
	}

	return nil
}

func (svc AuthenticationService) Update(userUID string, userRequest auth_models.UserProfileRequest) error {

	if strings.TrimSpace(userUID) == "" {
		return errors.New("user identity invalid")
	}

	userFind, err := svc.authRepo.FindUserByUID(context.Background(), userUID)
	if err != nil && err.Error() != "mongo: no documents in result" {
		return err
	}

	if strings.TrimSpace(userFind.UID) == "" {
		return errors.New("user is not exists")
	}

	stableUID := userFind.UID
	stableRegisterType := userFind.RegisterType
	userFind.UserDetail = userRequest.UserDetail
	userFind.UID = stableUID
	userFind.RegisterType = stableRegisterType
	userFind.UpdatedAt = svc.timeNow()

	err = svc.authRepo.UpdateUserByUID(context.Background(), userUID, *userFind)

	if err != nil {
		return err
	}

	return nil
}

func (svc AuthenticationService) UpdatePassword(username string, currentPassword string, newPassword string) error {

	if username == "" {
		return errors.New("username invalid")
	}
	if !auth_models.IsValidPasswordLength(newPassword) {
		return apperr.Validation("newpassword", "password must contain 15 to 64 characters")
	}
	if auth_models.IsKnownCompromisedPassword(newPassword) {
		return apperr.Validation("newpassword", "password is known to be compromised")
	}

	userFind, err := svc.authRepo.FindUser(context.Background(), username)
	if err != nil && err.Error() != "mongo: no documents in result" {
		return err
	}

	if len(userFind.Username) < 1 {
		return errors.New("username is not exists")
	}

	passwordInvalid := !svc.checkHashPassword(currentPassword, userFind.Password)

	if passwordInvalid {
		return errors.New("current password invalid")
	}

	hashPassword, err := svc.passwordEncoder(newPassword)

	if err != nil {
		return err
	}

	userFind.Password = hashPassword
	userFind.UpdatedAt = svc.timeNow()

	err = svc.authRepo.UpdateUser(context.Background(), username, *userFind)

	if err != nil {
		return err
	}

	return svc.authService.RevokeUserTokens(username)
}

func (svc AuthenticationService) ResetPasswordToDefault(holdingCode string, authUsername string, targetUsername string) error {
	return apperr.ErrForbidden.
		WithMessage("administrator password reset is disabled").
		WithThaiMessage("ผู้ดูแลไม่สามารถตั้งหรือรีเซ็ตรหัสผ่านแทนผู้ใช้ได้")
}

func (svc AuthenticationService) Logout(authorizationHeader string) error {
	return svc.authService.RevokeSession(authorizationHeader)
}

func (svc AuthenticationService) Profile(username string, userUID string) (auth_models.UserProfile, error) {

	userProfile := auth_models.UserProfile{}
	var user *auth_models.UserDoc
	var err error
	if strings.TrimSpace(userUID) != "" {
		user, err = svc.authRepo.FindByIdentity(context.Background(), "uid", userUID)
	}
	if err != nil || strings.TrimSpace(userUID) == "" {
		user, err = svc.authRepo.FindUser(context.Background(), username)
	}
	if err != nil {
		return userProfile, err
	}
	userProfile.Username = user.Username
	userProfile.Email = user.Email
	userProfile.UserDetail = user.UserDetail
	userProfile.LineUserID = user.LineUserID
	userProfile.LineDisplayName = user.LineDisplayName
	userProfile.LinePictureURL = user.LinePictureURL
	return userProfile, nil
}

func (svc AuthenticationService) AccessShop(holdingCode string, businessCode string, branchUID string, username string, userUID string, authorizationHeader string, authContext models.AuthenticationContext) error {

	holdingCode = strings.TrimSpace(holdingCode)
	if holdingCode == "" {
		return errors.New("holdingcode invalid")
	}

	tokenStr, err := svc.authService.GetTokenFromAuthorizationHeader(microservice.AUTHTYPE_BEARER, authorizationHeader)

	if err != nil {
		return err
	}

	if len(tokenStr) < 1 {
		return errors.New("token invalid")
	}

	shopUser, err := svc.findShopUser(context.Background(), holdingCode, userUID)
	if err != nil {
		resolvedHoldingCode, resolveErr := svc.resolveLoginHoldingCode(context.Background(), holdingCode)
		if resolveErr == nil && resolvedHoldingCode != "" && resolvedHoldingCode != holdingCode {
			holdingCode = resolvedHoldingCode
			shopUser, err = svc.findShopUser(context.Background(), holdingCode, userUID)
		}
	}

	if err != nil {
		return err
	}

	if shopUser.ID == primitive.NilObjectID {
		return errors.New("holdingcode invalid")
	}

	if err = svc.ensureShopAccessAllowed(shopUser); err != nil {
		return err
	}
	businessCode = utils.NormalizeBusinessCode(businessCode)
	branchUID = strings.TrimSpace(branchUID)
	if businessCode != "" {
		companyUID, resolveErr := svc.shopUserRepo.ResolveCompanyUID(context.Background(), holdingCode, businessCode)
		allowed := auth_models.ScopesAllowCompanySelection(shopUser.AccessScopes, companyUID)
		if branchUID != "" {
			allowed = auth_models.ScopesAllowBranchSelection(shopUser.AccessScopes, companyUID, branchUID)
		}
		if resolveErr != nil || !allowed {
			return apperr.ErrForbidden.WithMessage("company access denied").WithThaiMessage("ไม่มีสิทธิ์ใช้งานบริษัทนี้")
		}
	} else if branchUID != "" {
		return apperr.ErrForbidden.WithMessage("branch requires company workspace").WithThaiMessage("ต้องเลือกบริษัทของสาขาก่อน")
	}

	err = svc.authService.SelectShop(microservice.AUTHTYPE_BEARER, tokenStr, holdingCode, businessCode, branchUID, shopUser.Role)

	if err != nil {
		return errors.New("failed shop select")
	}

	lastAccessedAt := svc.timeNow()
	err = svc.shopUserRepo.UpdateLastAccess(context.Background(), holdingCode, shopUser.UserUID, lastAccessedAt)
	if err != nil {
		logger.GetLogger().Error(err.Error())
	}

	err = svc.shopUserAccessLogRepo.Create(
		context.Background(),
		auth_models.ShopUserAccessLog{
			HoldingCode:    holdingCode,
			BusinessCode:   businessCode,
			Username:       shopUser.Username,
			Ip:             authContext.Ip,
			LastAccessedAt: lastAccessedAt,
		})

	if err != nil {
		logger.GetLogger().Error(err.Error())
	}

	return nil
}

func (svc AuthenticationService) UpdateFavoriteShop(holdingCode string, username string, userUID string, isFavorite bool) error {

	if holdingCode == "" {
		return errors.New("holdingcode invalid")
	}

	if strings.TrimSpace(userUID) == "" {
		return errors.New("user identity invalid")
	}

	shopUser, err := svc.shopUserRepo.FindByHoldingCodeAndUserUID(context.Background(), holdingCode, userUID)

	if err != nil {
		return err
	}

	if shopUser.ID == primitive.NilObjectID {
		return errors.New("shop invalid")
	}

	err = svc.shopUserRepo.SaveFavorite(context.Background(), holdingCode, shopUser.UserUID, isFavorite)
	if err != nil {
		return errors.New("favorite failed")
	}

	return nil
}

func (svc AuthenticationService) LoginWithFirebaseToken(token string) (string, error) {

	userInfo, err := svc.firebaseAdapter.ValidateToken(token)
	if err != nil {
		return "", err
	}

	// find
	userFind, err := svc.authRepo.FindUser(context.Background(), userInfo.Email)
	if err != nil && err.Error() != "mongo: no documents in result" {
		return "", err
	}

	if len(userFind.Username) == 0 {
		// register
		user := auth_models.UserDoc{}

		user.UID = svc.generateGUID()
		user.Username = userInfo.Email
		user.Email = userInfo.Email
		user.Password = ""
		user.UserDetail.Name = userInfo.Name
		user.CreatedAt = svc.timeNow()

		_, err := svc.authRepo.CreateUser(context.Background(), user)
		if err != nil {
			return "", err
		}
		userFind, err = svc.authRepo.FindUser(context.Background(), userInfo.Email)
		if err != nil && err.Error() != "mongo: no documents in result" {
			return "", err
		}
	}

	if !userFind.DisabledAt.IsZero() {
		return "", &auth_models.UserDisableLoginError{}
	}

	tokenString, err := svc.authService.GenerateTokenWithRedis(microservice.AUTHTYPE_BEARER, svc.tokenUserInfo(*userFind))

	if err != nil {
		return "", errors.New("generate token error")
	}

	return tokenString, nil
}

// LoginWithGoogleIdentity resolves returning users exclusively by the stable OIDC
// issuer+subject pair. Email is only recorded during the first verified link.
func (svc AuthenticationService) LoginWithGoogleIdentity(issuer string, subject string, email string, displayName string) (models.TokenLoginResponse, error) {
	issuer = normalizeGoogleIssuer(issuer)
	subject = strings.TrimSpace(subject)
	email = strings.ToLower(strings.TrimSpace(email))
	if issuer == "" || subject == "" || email == "" {
		return models.TokenLoginResponse{}, errors.New("google identity is invalid")
	}

	ctx := context.Background()
	identity, err := svc.authRepo.FindGoogleIdentity(ctx, issuer, subject)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return models.TokenLoginResponse{}, err
	}
	if errors.Is(err, mongo.ErrNoDocuments) {
		identity = &auth_models.GoogleIdentity{}
	}
	if identity.IdentityUID != "" {
		return svc.loginWithLinkedGoogleIdentity(ctx, *identity)
	}

	now := svc.timeNow().UTC()
	user := auth_models.UserDoc{
		GuidFixed: svc.generateGUID(),
		CreatedAt: now,
	}
	user.UID = svc.generateGUID()
	user.Email = email
	user.Name = strings.TrimSpace(displayName)
	identity = &auth_models.GoogleIdentity{
		IdentityUID:   svc.generateGUID(),
		UserUID:       user.UID,
		Issuer:        issuer,
		Subject:       subject,
		VerifiedEmail: email,
		IsActive:      true,
		LinkedAt:      now,
	}
	audit := auth_models.AuthAudit{
		AuditUID:   svc.generateGUID(),
		UserUID:    user.UID,
		Action:     "GOOGLE_IDENTITY_LINK",
		Outcome:    "SUCCESS",
		OccurredAt: now,
	}
	linkedUser, err := svc.authRepo.CreateGoogleUserIdentity(ctx, user, *identity, audit)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			linked, findErr := svc.authRepo.FindGoogleIdentity(ctx, issuer, subject)
			if findErr == nil && linked.IdentityUID != "" {
				return svc.loginWithLinkedGoogleIdentity(ctx, *linked)
			}
		}
		return models.TokenLoginResponse{}, err
	}
	return svc.processUserLogin(linkedUser, "", models.AuthenticationContext{})
}

func (svc AuthenticationService) loginWithLinkedGoogleIdentity(ctx context.Context, identity auth_models.GoogleIdentity) (models.TokenLoginResponse, error) {
	if !identity.IsActive || identity.RevokedAt != nil || strings.TrimSpace(identity.UserUID) == "" {
		return models.TokenLoginResponse{}, errors.New("google identity is inactive")
	}
	user, err := svc.authRepo.FindUserByUID(ctx, identity.UserUID)
	if err != nil {
		return models.TokenLoginResponse{}, err
	}
	if user.UID == "" || user.IsDeleted {
		return models.TokenLoginResponse{}, errors.New("google identity user not found")
	}
	if !user.DisabledAt.IsZero() {
		return models.TokenLoginResponse{}, &auth_models.UserDisableLoginError{}
	}
	return svc.processUserLogin(*user, "", models.AuthenticationContext{})
}

func normalizeGoogleIssuer(issuer string) string {
	switch strings.TrimSpace(issuer) {
	case "accounts.google.com", "https://accounts.google.com":
		return "https://accounts.google.com"
	default:
		return ""
	}
}

func (svc AuthenticationService) LoginWithLineToken(token string) (string, error) {

	userInfo, err := svc.lineAdapter.ValidateToken(token)
	if err != nil {
		return "", err
	}

	// find user by line user id (we'll use line user id as username for simplicity)
	userFind, err := svc.authRepo.FindUser(context.Background(), userInfo.UserId)
	if err != nil && err.Error() != "mongo: no documents in result" {
		return "", err
	}

	if len(userFind.Username) == 0 {
		// register new user
		user := auth_models.UserDoc{}

		user.UID = svc.generateGUID()
		user.Username = userInfo.UserId
		user.Password = ""
		user.UserDetail.Name = userInfo.DisplayName
		user.UserDetail.Avatar = userInfo.PictureUrl
		user.CreatedAt = svc.timeNow()

		_, err := svc.authRepo.CreateUser(context.Background(), user)
		if err != nil {
			return "", err
		}
		userFind, err = svc.authRepo.FindUser(context.Background(), userInfo.UserId)
		if err != nil && err.Error() != "mongo: no documents in result" {
			return "", err
		}
	}

	if !userFind.DisabledAt.IsZero() {
		return "", &auth_models.UserDisableLoginError{}
	}

	tokenString, err := svc.authService.GenerateTokenWithRedis(microservice.AUTHTYPE_BEARER, svc.tokenUserInfo(*userFind))

	if err != nil {
		return "", errors.New("generate token error")
	}

	return tokenString, nil
}

// LoginWithLineUserID — สำหรับ QR code / LIFF login flow
// ค้นหา user ที่เชื่อมต่อ LINE ไว้แล้ว (จาก users collection — ระดับ user ไม่ใช่ per-shop)
// แล้ว login ด้วย username ของ user นั้น
func (svc AuthenticationService) LoginWithLineUserID(lineUserID string, displayName string, pictureUrl string, email string) (string, string, error) {

	if lineUserID == "" {
		return "", "", errors.New("lineuserid is required")
	}

	// ค้นหา user ที่เชื่อมต่อ LINE นี้ไว้ (จาก users collection)
	userFind, err := svc.authRepo.FindByLineUserID(context.Background(), lineUserID)
	if err != nil {
		// fallback: ค้นจาก shopusers (สำหรับ backward compatibility กับข้อมูลเก่า)
		shopUser, shopErr := svc.shopUserRepo.FindByLineUserID(context.Background(), lineUserID)
		if shopErr != nil || shopUser.Username == "" {
			return "", "", errors.New("ไม่พบบัญชีที่เชื่อมต่อ LINE นี้ กรุณาเชื่อมต่อ LINE กับบัญชีก่อน")
		}
		// หา auth user จาก username ที่ได้จาก shopUser
		userFind, err = svc.authRepo.FindUser(context.Background(), shopUser.Username)
		if err != nil {
			return "", "", errors.New("ไม่พบข้อมูลผู้ใช้ในระบบ")
		}
	}

	if userFind.Username == "" {
		return "", "", errors.New("ไม่พบบัญชีที่เชื่อมต่อ LINE นี้ กรุณาเชื่อมต่อ LINE กับบัญชีก่อน")
	}

	if !userFind.DisabledAt.IsZero() {
		return "", "", &auth_models.UserDisableLoginError{}
	}

	tokenString, err := svc.authService.GenerateTokenWithRedis(microservice.AUTHTYPE_BEARER, svc.tokenUserInfo(*userFind))

	if err != nil {
		return "", "", errors.New("generate token error")
	}

	// return token + username ของ user จริง (เช่น email) ไม่ใช่ LINE display name
	return tokenString, userFind.Username, nil
}

// LinkLine — เชื่อมต่อ LINE กับ user profile (ระดับ user ไม่ใช่ per-shop)
func (svc AuthenticationService) LinkLine(username string, req auth_models.LinkLineRequest) error {

	if username == "" {
		return errors.New("username invalid")
	}

	// ตรวจสอบว่า LINE User ID นี้ถูกเชื่อมต่อกับ user อื่นหรือไม่
	existingUser, err := svc.authRepo.FindByLineUserID(context.Background(), req.LineUserID)
	if err == nil && existingUser != nil && existingUser.Username != "" && existingUser.Username != username {
		return errors.New("LINE นี้เชื่อมต่อกับบัญชีอื่นแล้ว (" + existingUser.Username + ")")
	}

	userFind, err := svc.authRepo.FindUser(context.Background(), username)
	if err != nil {
		return errors.New("ไม่พบข้อมูลผู้ใช้")
	}

	if len(userFind.Username) < 1 {
		return errors.New("ไม่พบข้อมูลผู้ใช้")
	}

	userFind.LineUserID = req.LineUserID
	userFind.LineDisplayName = req.LineDisplayName
	userFind.LinePictureURL = req.LinePictureURL
	userFind.UpdatedAt = svc.timeNow()

	err = svc.authRepo.UpdateUser(context.Background(), username, *userFind)
	if err != nil {
		return err
	}

	return nil
}

// UnlinkLine — ยกเลิกการเชื่อมต่อ LINE จาก user profile
func (svc AuthenticationService) UnlinkLine(username string) error {

	if username == "" {
		return errors.New("username invalid")
	}

	userFind, err := svc.authRepo.FindUser(context.Background(), username)
	if err != nil {
		return errors.New("ไม่พบข้อมูลผู้ใช้")
	}

	if len(userFind.Username) < 1 {
		return errors.New("ไม่พบข้อมูลผู้ใช้")
	}

	userFind.LineUserID = ""
	userFind.LineDisplayName = ""
	userFind.LinePictureURL = ""
	userFind.UpdatedAt = svc.timeNow()

	err = svc.authRepo.UpdateUser(context.Background(), username, *userFind)
	if err != nil {
		return err
	}

	return nil
}

func (svc AuthenticationService) DisableUser(userUID string) error {

	if strings.TrimSpace(userUID) == "" {
		return errors.New("user identity invalid")
	}

	userFind, err := svc.authRepo.FindUserByUID(context.Background(), userUID)
	if err != nil && err.Error() != "mongo: no documents in result" {
		return err
	}

	if strings.TrimSpace(userFind.UID) == "" {
		return errors.New("user is not exists")
	}

	userFind.DisabledAt = svc.timeNow()

	err = svc.authRepo.UpdateUserByUID(context.Background(), userUID, *userFind)
	if err != nil {
		return err
	}

	return svc.authService.RevokeUserTokensByUID(userUID)

}

func (svc AuthenticationService) DeleteUser(username string) error {

	if username == "" {
		return errors.New("username invalid")
	}

	userFind, err := svc.authRepo.FindUser(context.Background(), username)
	if err != nil && err.Error() != "mongo: no documents in result" {
		return err
	}

	if len(userFind.Username) < 1 {
		return errors.New("username is not exists")
	}

	if userFind.DisabledAt.IsZero() {
		return errors.New("user is not disabled")
	}

	shopFind, err := svc.shopUserRepo.FindByUsername(context.Background(), username)

	if err != nil {
		return err
	}

	for _, shopUser := range *shopFind {
		err = svc.shopUserRepo.Delete(context.Background(), shopUser.HoldingCode, username)
		if err != nil {
			return err
		}
	}

	err = svc.authRepo.DeleteUser(context.Background(), username)

	if err != nil {
		return err
	}

	return nil
}
