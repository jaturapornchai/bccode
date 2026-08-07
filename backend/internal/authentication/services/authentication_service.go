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
)

type IAuthenticationService interface {
	LoginWithPhoneNumber(userLoginReq *auth_models.UserLoginPhoneNumberRequest, authContext models.AuthenticationContext) (models.TokenLoginResponse, error)
	LoginWithPhoneNumberOTP(userLoginReq *auth_models.PhoneNumberOTPRequest, authContext models.AuthenticationContext) (models.TokenLoginResponse, error)
	Login(userReq *auth_models.UserLoginRequest, authContext models.AuthenticationContext) (models.TokenLoginResponse, error)
	Poslogin(userReq *auth_models.PosLoginRequest, authContext models.AuthenticationContext) (models.TokenLoginResponse, error)
	LoginEmail(userReq *auth_models.PosLoginRequest, authContext models.AuthenticationContext) (string, error)
	Register(userRequest auth_models.RegisterEmailRequest) (string, error)
	ForgotPasswordByPhonenumber(userRequest auth_models.ForgotPasswordPhoneNumberRequest) error
	Update(username string, userRequest auth_models.UserProfileRequest) error
	UpdatePassword(username string, currentPassword string, newPassword string) error
	ResetPasswordToDefault(holdingCode string, authUsername string, targetUsername string) error
	Logout(authorizationHeader string) error
	Profile(username string, userUID string) (auth_models.UserProfile, error)
	AccessShop(holdingCode string, businessCode string, username string, userUID string, authorizationHeader string, authContext models.AuthenticationContext) error
	UpdateFavoriteShop(holdingCode string, username string, userUID string, isFavorite bool) error
	LoginWithFirebaseToken(token string) (string, error)
	LoginWithLineToken(token string) (string, error)
	LoginWithLineUserID(lineUserID string, displayName string, pictureUrl string, email string) (string, string, error)
	LoginWithGoogleEmail(email string, displayName string) (string, error)
	RefreshToken(tokenRequest models.TokenLoginRequest) (models.TokenLoginResponse, error)

	LinkLine(username string, req auth_models.LinkLineRequest) error
	UnlinkLine(username string) error

	CheckExistsUsername(username string) (bool, error)
	CheckExistsPhonenumber(phoneNumber string) (bool, error)
	SendPhonenumberOTP(otpRequest auth_models.OTPRequest) (auth_models.OTPResponse, error)
	RegisterByPhonenumber(userRequest auth_models.RegisterPhoneNumberRequest) (string, error)
	RegisterByUsername(userRequest auth_models.RegisterUsernameRequest) (string, error)

	DisableUser(username string) error
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
	tokenString, err := svc.authService.GenerateTokenWithRedis(microservice.AUTHTYPE_BEARER, userInfo)

	if err != nil {
		return models.TokenLoginResponse{}, errors.New("login failed")
	}

	refreshTokenString, err := svc.authService.GenerateTokenWithRedis(microservice.AUTHTYPE_REFRESH, userInfo)

	if err != nil {
		svc.authService.DeleteToken(microservice.AUTHTYPE_BEARER, tokenString)
		return models.TokenLoginResponse{}, errors.New("login failed")
	}

	return models.TokenLoginResponse{
		Token:              tokenString,
		Refresh:            refreshTokenString,
		MustChangePassword: userInfo.MustChangePassword,
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

	userLoginReq.Username = utils.NormalizeUsername(userLoginReq.Username)

	userLoginReq.Username = strings.TrimSpace(userLoginReq.Username)
	userLoginReq.HoldingCode = strings.TrimSpace(userLoginReq.HoldingCode)

	findUser, err := svc.authRepo.FindUser(context.Background(), userLoginReq.Username)

	if err != nil && err.Error() != "mongo: no documents in result" {
		// svc.ms.Log("Authentication service", err.Error())
		return models.TokenLoginResponse{}, errors.New("auth: database connect error")
	}

	// Email-as-login fallback: if the username lookup did not match, retry by email.
	// The input may be an email address (e.g. owner logging in with jaturapornchai@gmail.com).
	// FindUser is built on PersisterMongo.FindOne which swallows "no documents" and returns
	// a zero struct + nil err (see go-expert "Mongo not-found semantics"), so the empty
	// result test below is the real "not found" signal — same pattern as the select-holding fix.
	if len(findUser.Username) < 1 && strings.Contains(userLoginReq.Username, "@") {
		findUserByEmail, emailErr := svc.authRepo.FindByIdentity(context.Background(), "email", userLoginReq.Username)
		if emailErr == nil && findUserByEmail != nil && len(findUserByEmail.Username) > 0 {
			findUser = findUserByEmail
		}
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

func (svc *AuthenticationService) processUserLogin(findUser auth_models.UserDoc, holdingCode string, authContext models.AuthenticationContext) (models.TokenLoginResponse, error) {
	userInfo := svc.tokenUserInfo(findUser)
	mustChangePassword := userInfo.MustChangePassword
	tokenString, err := svc.authService.GenerateTokenWithRedis(microservice.AUTHTYPE_BEARER, userInfo)

	if err != nil {
		return models.TokenLoginResponse{}, errors.New("login failed")
	}

	refreshTokenString, err := svc.authService.GenerateTokenWithRedis(microservice.AUTHTYPE_REFRESH, userInfo)

	if err != nil {
		svc.authService.DeleteToken(microservice.AUTHTYPE_BEARER, tokenString)
		return models.TokenLoginResponse{}, errors.New("login failed")
	}

	if len(holdingCode) > 0 {
		var shopUser auth_models.ShopUser
		var err error
		if strings.TrimSpace(findUser.UID) != "" {
			shopUser, err = svc.shopUserRepo.FindByHoldingCodeAndUserUID(context.Background(), holdingCode, findUser.UID)
		}
		// Also fall back when the uid query matched nothing (FindOne returns a zero struct + nil
		// error on no-match). A by-email membership created before first login keeps an empty/old
		// useruid, so the uid lookup misses and we must match by the stable username.
		if strings.TrimSpace(findUser.UID) == "" || err != nil || shopUser.ID == primitive.NilObjectID {
			shopUser, err = svc.shopUserRepo.FindByHoldingCodeAndUsername(context.Background(), holdingCode, findUser.Username)
		}

		if err != nil {
			return models.TokenLoginResponse{}, err
		}

		if shopUser.ID == primitive.NilObjectID {
			return models.TokenLoginResponse{}, errors.New("holdingcode invalid")
		}

		if err = svc.ensureShopAccessAllowed(context.Background(), holdingCode, shopUser); err != nil {
			svc.authService.DeleteToken(microservice.AUTHTYPE_BEARER, tokenString)
			svc.authService.DeleteToken(microservice.AUTHTYPE_REFRESH, refreshTokenString)
			return models.TokenLoginResponse{}, err
		}

		err = svc.authService.SelectShop(microservice.AUTHTYPE_BEARER, tokenString, holdingCode, "", shopUser.Role)

		if err != nil {
			return models.TokenLoginResponse{}, errors.New("failed shop select")
		}

		lastAccessedAt := svc.timeNow()

		err = svc.shopUserRepo.UpdateLastAccess(context.Background(), holdingCode, shopUser.Username, lastAccessedAt)
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

	return models.TokenLoginResponse{Token: tokenString, Refresh: refreshTokenString, MustChangePassword: mustChangePassword}, nil
}

func (svc AuthenticationService) tokenUserInfo(user auth_models.UserDoc) micromodel.UserInfo {
	return micromodel.UserInfo{
		Username:           user.Username,
		Name:               user.Name,
		UID:                user.UID,
		MustChangePassword: user.Password != "" && svc.checkHashPassword(models.DefaultUserPassword, user.Password),
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

func (svc AuthenticationService) findShopUser(ctx context.Context, holdingCode string, username string, userUID string) (auth_models.ShopUser, error) {
	var shopUser auth_models.ShopUser
	var err error
	if strings.TrimSpace(userUID) != "" {
		shopUser, err = svc.shopUserRepo.FindByHoldingCodeAndUserUID(ctx, holdingCode, userUID)
	}
	// Fall back to a username lookup when there is no userUID, the userUID query errored,
	// OR it matched nothing (PersisterMongo.FindOne swallows "no documents" and returns a
	// zero-valued struct with nil error). The token's uid can drift from the useruid stored
	// on the membership — e.g. a membership created by email before that person logged in,
	// or a token minted with an older uid — so the uid query misses and we must match by the
	// stable email username instead. Without this, AccessShop wrongly reports "holdingcode invalid".
	if strings.TrimSpace(userUID) == "" || err != nil || shopUser.ID == primitive.NilObjectID {
		shopUser, err = svc.shopUserRepo.FindByHoldingCodeAndUsername(ctx, holdingCode, username)
	}
	return shopUser, err
}

func (svc *AuthenticationService) ensureShopAccessAllowed(ctx context.Context, holdingCode string, shopUser auth_models.ShopUser) error {
	expired := !shopUser.AccessExpiryDate.IsZero() && time.Now().After(shopUser.AccessExpiryDate)
	if !shopUser.IsAccessDisabled && !expired {
		return nil
	}

	// Access is disabled (manual) or expired (offboarding) — the shop creator is always exempt.
	createdBy, err := svc.shopUserRepo.FindShopCreatedBy(ctx, holdingCode)
	if err != nil {
		return err
	}

	if strings.EqualFold(utils.NormalizeUsername(shopUser.Username), utils.NormalizeUsername(createdBy)) {
		return nil
	}

	if expired {
		return errors.New("user_access_expired")
	}
	return errors.New("user_access_disabled")
}

func (svc AuthenticationService) RefreshToken(tokenRequest models.TokenLoginRequest) (models.TokenLoginResponse, error) {

	token, refreshToken, mustChangePassword, err := svc.authService.RefreshToken(tokenRequest.Token)

	if err != nil {
		return models.TokenLoginResponse{}, err
	}

	return models.TokenLoginResponse{
		Token:              token,
		Refresh:            refreshToken,
		MustChangePassword: mustChangePassword,
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

	userRequest.Username = utils.NormalizeUsername(userRequest.Username)

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

	username = utils.NormalizeUsername(username)

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

func (svc AuthenticationService) Update(username string, userRequest auth_models.UserProfileRequest) error {

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

	userFind.UserDetail = userRequest.UserDetail
	userFind.UpdatedAt = svc.timeNow()

	err = svc.authRepo.UpdateUser(context.Background(), username, *userFind)

	if err != nil {
		return err
	}

	return nil
}

func (svc AuthenticationService) UpdatePassword(username string, currentPassword string, newPassword string) error {

	if username == "" {
		return errors.New("username invalid")
	}
	if newPassword == models.DefaultUserPassword {
		return apperr.Validation("newpassword", "new password must not be the default password")
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
	holdingCode = strings.TrimSpace(holdingCode)
	authUsername = utils.NormalizeUsername(authUsername)
	targetUsername = utils.NormalizeUsername(targetUsername)

	if holdingCode == "" {
		return errors.New("shop invalid")
	}
	if authUsername == "" || targetUsername == "" {
		return errors.New("username invalid")
	}
	if authUsername == targetUsername {
		return errors.New("use change password for self")
	}

	authUser, err := svc.shopUserRepo.FindByHoldingCodeAndUsername(context.Background(), holdingCode, authUsername)
	if err != nil {
		return err
	}
	expired := !authUser.AccessExpiryDate.IsZero() && svc.timeNow().After(authUser.AccessExpiryDate)
	if authUser.IsAccessDisabled || expired || (authUser.Role != models.ROLE_OWNER && authUser.Role != models.ROLE_ADMIN) {
		return errors.New("permission denied")
	}

	targetShopUser, err := svc.shopUserRepo.FindByHoldingCodeAndUsername(context.Background(), holdingCode, targetUsername)
	if err != nil {
		return err
	}
	if targetShopUser.Username == "" {
		return errors.New("user not found")
	}
	if authUser.Role == models.ROLE_ADMIN && targetShopUser.Role == models.ROLE_OWNER {
		return errors.New("permission denied")
	}

	hashPassword, err := svc.passwordEncoder(models.DefaultUserPassword)
	if err != nil {
		return err
	}

	userFind, err := svc.authRepo.FindUser(context.Background(), targetUsername)
	if err != nil {
		if !isMongoNotFoundError(err) {
			return err
		}
		user := auth_models.UserDoc{}
		user.UID = svc.generateGUID()
		user.Username = targetUsername
		user.Password = hashPassword
		user.UserDetail.Name = targetUsername
		user.CreatedAt = svc.timeNow()
		_, err = svc.authRepo.CreateUser(context.Background(), user)
		return err
	}
	if userFind.Username == "" {
		user := auth_models.UserDoc{}
		user.UID = svc.generateGUID()
		user.Username = targetUsername
		user.Password = hashPassword
		user.UserDetail.Name = targetUsername
		user.CreatedAt = svc.timeNow()
		_, err = svc.authRepo.CreateUser(context.Background(), user)
		return err
	}

	userFind.Password = hashPassword
	userFind.UpdatedAt = svc.timeNow()

	if err := svc.authRepo.UpdateUser(context.Background(), targetUsername, *userFind); err != nil {
		return err
	}
	return svc.authService.RevokeUserTokens(targetUsername)
}

func isMongoNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "no documents") || strings.Contains(message, "not found")
}

func (svc AuthenticationService) Logout(authorizationHeader string) error {
	return svc.authService.ExpireToken(microservice.AUTHTYPE_BEARER, authorizationHeader)
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
	userProfile.IsDefaultPassword = user.Password != "" && svc.checkHashPassword(models.DefaultUserPassword, user.Password)

	return userProfile, nil
}

func (svc AuthenticationService) AccessShop(holdingCode string, businessCode string, username string, userUID string, authorizationHeader string, authContext models.AuthenticationContext) error {

	holdingCode = strings.TrimSpace(holdingCode)
	if holdingCode == "" {
		return errors.New("holdingcode invalid")
	}

	if username == "" {
		return errors.New("username invalid")
	}

	tokenStr, err := svc.authService.GetTokenFromAuthorizationHeader(microservice.AUTHTYPE_BEARER, authorizationHeader)

	if err != nil {
		return err
	}

	if len(tokenStr) < 1 {
		return errors.New("token invalid")
	}

	shopUser, err := svc.findShopUser(context.Background(), holdingCode, username, userUID)
	if err != nil {
		resolvedHoldingCode, resolveErr := svc.resolveLoginHoldingCode(context.Background(), holdingCode)
		if resolveErr == nil && resolvedHoldingCode != "" && resolvedHoldingCode != holdingCode {
			holdingCode = resolvedHoldingCode
			shopUser, err = svc.findShopUser(context.Background(), holdingCode, username, userUID)
		}
	}

	if err != nil {
		return err
	}

	if shopUser.ID == primitive.NilObjectID {
		return errors.New("holdingcode invalid")
	}

	if err = svc.ensureShopAccessAllowed(context.Background(), holdingCode, shopUser); err != nil {
		return err
	}
	businessCode = utils.NormalizeBusinessCode(businessCode)
	if businessCode != "" && !auth_models.ScopesAllowCompanySelection(shopUser.AccessScopes, businessCode) {
		return apperr.ErrForbidden.WithMessage("company access denied").WithThaiMessage("ไม่มีสิทธิ์ใช้งานบริษัทนี้")
	}

	err = svc.authService.SelectShop(microservice.AUTHTYPE_BEARER, tokenStr, holdingCode, businessCode, shopUser.Role)

	if err != nil {
		return errors.New("failed shop select")
	}

	lastAccessedAt := svc.timeNow()
	err = svc.shopUserRepo.UpdateLastAccess(context.Background(), holdingCode, shopUser.Username, lastAccessedAt)
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

	if username == "" {
		return errors.New("username invalid")
	}

	var shopUser auth_models.ShopUser
	var err error
	if strings.TrimSpace(userUID) != "" {
		shopUser, err = svc.shopUserRepo.FindByHoldingCodeAndUserUID(context.Background(), holdingCode, userUID)
	}
	// Fall back to username when the uid query matched nothing (FindOne returns a zero struct +
	// nil error on no-match) so uid drift / by-email memberships still resolve.
	if strings.TrimSpace(userUID) == "" || err != nil || shopUser.ID == primitive.NilObjectID {
		shopUser, err = svc.shopUserRepo.FindByHoldingCodeAndUsername(context.Background(), holdingCode, username)
	}

	if err != nil {
		return err
	}

	if shopUser.ID == primitive.NilObjectID {
		return errors.New("shop invalid")
	}

	err = svc.shopUserRepo.SaveFavorite(context.Background(), holdingCode, shopUser.Username, isFavorite)
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

// LoginWithGoogleEmail — สำหรับ mobile Google OAuth (Android/iOS)
// ค้นหา user ด้วย email หรือสร้างใหม่ถ้ายังไม่มี แล้ว generate JWT
func (svc AuthenticationService) LoginWithGoogleEmail(email string, displayName string) (string, error) {
	if email == "" {
		return "", errors.New("email is required")
	}

	// ค้นหา user ด้วย email
	userFind, err := svc.authRepo.FindUser(context.Background(), email)
	if err != nil && err.Error() != "mongo: no documents in result" {
		return "", err
	}

	// ถ้าไม่มี user ให้สร้างใหม่
	if len(userFind.Username) == 0 {
		user := auth_models.UserDoc{}
		user.UID = svc.generateGUID()
		user.Username = email
		user.Email = email
		user.Password = ""
		user.UserDetail.Name = displayName
		user.CreatedAt = svc.timeNow()

		_, err = svc.authRepo.CreateUser(context.Background(), user)
		if err != nil {
			return "", err
		}
		userFind, err = svc.authRepo.FindUser(context.Background(), email)
		if err != nil {
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

func (svc AuthenticationService) DisableUser(username string) error {

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

	userFind.DisabledAt = svc.timeNow()

	err = svc.authRepo.UpdateUser(context.Background(), username, *userFind)
	if err != nil {
		return err
	}

	return svc.authService.RevokeUserTokens(username)

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
