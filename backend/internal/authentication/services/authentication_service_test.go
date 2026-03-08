package services_test

import (
	"context"
	"errors"
	"smlcloudplatform/internal/authentication/models"
	"smlcloudplatform/internal/authentication/services"
	"smlcloudplatform/internal/firebase"
	"smlcloudplatform/internal/line"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/smlsoft/mongopagination"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func mockLoginData(authRepo *AuthenticationRepositoryMock, shopUserRepo *ShopUserRepositoryMock, microAuthServiceMock *AuthServiceMock) {

	shopID := "SHOP_ID_TEST"
	role := uint8(2)

	tokenMock := "TOKEN_MOCK"
	//authRepo FindUser
	userDoc1 := models.UserDoc{}
	userDoc1.Username = "tester1"
	userDoc1.Password = "tester1"
	userDoc1.Name = "tester1"

	authRepo.On("FindUser", userDoc1.Username).Return(&userDoc1, nil)
	authRepo.On("FindUser", "").Return(&models.UserDoc{}, nil)
	authRepo.On("FindUser", "tester2").Return(&models.UserDoc{}, nil)

	authRepo.On("FindUser", "user_register").Return(&models.UserDoc{}, nil)

	//authRepo CreateUser
	userDoc2 := models.UserDoc{}
	userDoc2.Username = "user_register"
	userDoc2.Password = "register_password_success"
	userDoc2.Name = "user_register"
	userDoc2.CreatedAt = MockTime()

	authRepo.On("CreateUser", userDoc2).Return(MockObjectID(), nil)

	//microAuth
	microAuthServiceMock.On("GenerateTokenWithRedis", micromodels.UserInfo{
		Username: userDoc1.Username,
		Name:     userDoc1.Name,
	}).Return(tokenMock, nil)

	microAuthServiceMock.On("SelectShop", tokenMock, shopID, role).Return(nil)

	shopUser := models.ShopUser{}
	shopUser.Username = userDoc1.Username
	shopUser.ShopID = shopID
	shopUser.Role = role

	//shopUser
	shopUserRepo.On("FindByShopIDAndUsername", shopID, userDoc1.Username).Return(shopUser, nil)

	shopUserRepo.On("FindByShopIDAndUsername", "SHOP_ID_INVALID", userDoc1.Username).Return(models.ShopUser{}, nil)
}

func TestAuthService_Login(t *testing.T) {
	shopID := "SHOP_ID_TEST"

	authRepo := new(AuthenticationRepositoryMock)
	shopUserRepo := new(ShopUserRepositoryMock)
	shopUserAccessLogRepo := new(ShopUserAccessLogRepositoryMock)
	smsRepo := new(SMSRepositoryMock)
	microAuthServiceMock := &AuthServiceMock{}

	mockLoginData(authRepo, shopUserRepo, microAuthServiceMock)

	type args struct {
		username string
		password string
		shopID   string
	}

	cases := []struct {
		name     string
		args     args
		wantErr  bool
		wantData string
	}{
		{
			name: "login success",
			args: args{
				shopID:   shopID,
				username: "tester1",
				password: "tester1",
			},
			wantErr:  false,
			wantData: "TOKEN_MOCK",
		},
		{
			name: "login success without shop id",
			args: args{
				username: "tester1",
				password: "tester1",
			},
			wantErr:  false,
			wantData: "TOKEN_MOCK",
		},
		{
			name: "login failure invalid shop id",
			args: args{
				shopID:   "SHOP_ID_INVALID",
				username: "tester1",
				password: "tester1",
			},
			wantErr:  true,
			wantData: "TOKEN_MOCK",
		},
		{
			name: "login failure password invalid",
			args: args{
				shopID:   shopID,
				username: "tester1",
				password: "invalidpassword",
			},
			wantErr:  true,
			wantData: "TOKEN_MOCK",
		},
		{
			name: "login failure username empty",
			args: args{
				shopID:   shopID,
				username: "",
				password: "invalidpassword",
			},
			wantErr:  true,
			wantData: "TOKEN_MOCK",
		},
		{
			name: "login failure username and password empty",
			args: args{
				shopID:   shopID,
				username: "",
				password: "",
			},
			wantErr:  true,
			wantData: "TOKEN_MOCK",
		},
	}

	authService := services.NewAuthenticationService(
		authRepo,
		shopUserRepo,
		shopUserAccessLogRepo,
		smsRepo,
		microAuthServiceMock,
		MockRandomString,
		MockRandomNumber,
		MockGUID,
		MockHashPassword,
		MockCheckPasswordHash,
		MockTime,
		MockFirebaseAdapter(),
		MockLineAdapter())
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {

			userReq := &models.UserLoginRequest{}
			userReq.Username = tt.args.username
			userReq.Password = tt.args.password
			userReq.ShopID = tt.args.shopID

			authContext := models.AuthenticationContext{
				Ip: "localhost",
			}

			tokenResult, err := authService.Login(userReq, authContext)

			if tt.wantErr {
				assert.NotNil(t, err)
				assert.Empty(t, tokenResult)
			} else {
				assert.Nil(t, err)
				assert.NotEmpty(t, tokenResult)
				assert.EqualValues(t, tt.wantData, tokenResult)
			}
		})
	}
}

func TestAuthService_Register(t *testing.T) {
	authRepo := new(AuthenticationRepositoryMock)
	shopUserRepo := new(ShopUserRepositoryMock)
	shopUserAccessLogRepo := new(ShopUserAccessLogRepositoryMock)
	smsRepo := new(SMSRepositoryMock)
	microAuthServiceMock := &AuthServiceMock{}

	mockLoginData(authRepo, shopUserRepo, microAuthServiceMock)

	type args struct {
		username string
		password string
		name     string
	}

	cases := []struct {
		name     string
		args     args
		wantErr  bool
		wantData string
	}{
		{
			name: "register success",
			args: args{
				username: "user_register",
				password: "register_password_success",
				name:     "user_register",
			},
			wantErr:  false,
			wantData: "62f9cb12c76fd9e83ac1b2ff",
		},
		{
			name: "register failure user is exist",
			args: args{
				username: "tester1",
				password: "register_password_failure",
				name:     "user_register",
			},
			wantErr:  true,
			wantData: "",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			authService := services.NewAuthenticationService(
				authRepo,
				shopUserRepo,
				shopUserAccessLogRepo,
				smsRepo,
				microAuthServiceMock,
				MockRandomString,
				MockRandomNumber,
				MockGUID,
				MockHashPassword,
				MockCheckPasswordHash,
				MockTime,
				MockFirebaseAdapter(),
				MockLineAdapter())

			userReq := models.RegisterEmailRequest{}
			userReq.Email = tt.args.username
			userReq.Password = tt.args.password
			userReq.Name = tt.args.name

			idx, err := authService.Register(userReq)

			if tt.wantErr {
				assert.NotNil(t, err)
				assert.Empty(t, "")
			} else {
				assert.Nil(t, err)
				assert.NotEmpty(t, idx)
				assert.EqualValues(t, tt.wantData, idx)
			}
		})
	}
}

func TestAuthService_Update(t *testing.T) {
	authRepo := new(AuthenticationRepositoryMock)
	shopUserRepo := new(ShopUserRepositoryMock)
	shopUserAccessLogRepo := new(ShopUserAccessLogRepositoryMock)
	smsRepo := new(SMSRepositoryMock)
	microAuthServiceMock := &AuthServiceMock{}

	userDoc := &models.UserDoc{}

	userDoc.Username = "user_update"
	userDoc.Name = "user_update"
	userDoc.UpdatedAt = MockTime()

	authRepo.On("FindUser", "user_update").Return(userDoc, nil)

	userDocUpdate := models.UserDoc{}
	userDocUpdate.Username = "user_update"
	userDocUpdate.Name = "new name"
	userDocUpdate.UpdatedAt = MockTime()

	authRepo.On("UpdateUser", "user_update", userDocUpdate).Return(nil)

	type args struct {
		username string
		name     string
	}

	cases := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "update success",
			args: args{
				username: "user_update",
				name:     "new name",
			},
			wantErr: false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			authService := services.NewAuthenticationService(
				authRepo,
				shopUserRepo,
				shopUserAccessLogRepo,
				smsRepo,
				microAuthServiceMock,
				MockRandomString,
				MockRandomNumber,
				MockGUID,
				MockHashPassword,
				MockCheckPasswordHash,
				MockTime,
				MockFirebaseAdapter(),
				MockLineAdapter())

			userReq := models.UserProfileRequest{}
			userReq.Name = tt.args.name

			err := authService.Update(tt.args.username, userReq)

			if tt.wantErr {
				assert.NotNil(t, err)
				assert.Empty(t, "")
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestAuthService_UpdatePassword(t *testing.T) {
	authRepo := new(AuthenticationRepositoryMock)
	shopUserRepo := new(ShopUserRepositoryMock)
	shopUserAccessLogRepo := new(ShopUserAccessLogRepositoryMock)
	smsRepo := new(SMSRepositoryMock)
	microAuthServiceMock := &AuthServiceMock{}

	userDoc := &models.UserDoc{}

	userDoc.Username = "user_update"
	userDoc.Password = "current_password"
	userDoc.UpdatedAt = MockTime()

	authRepo.On("FindUser", "user_update").Return(userDoc, nil)

	userDocUpdate := models.UserDoc{}
	userDocUpdate.Username = "user_update"
	userDocUpdate.Password = "new_password"
	userDocUpdate.UpdatedAt = MockTime()

	authRepo.On("UpdateUser", "user_update", userDocUpdate).Return(nil)

	type args struct {
		username        string
		currentPassword string
		newPassword     string
	}

	cases := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "update password success",
			args: args{
				username:        "user_update",
				currentPassword: "current_password",
				newPassword:     "new_password",
			},
			wantErr: false,
		},
		{
			name: "update password failure",
			args: args{
				username:        "user_update",
				currentPassword: "current_password_invalid",
				newPassword:     "new_password",
			},
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			authService := services.NewAuthenticationService(
				authRepo,
				shopUserRepo,
				shopUserAccessLogRepo,
				smsRepo,
				microAuthServiceMock,
				MockRandomString,
				MockRandomNumber,
				MockGUID,
				MockHashPassword,
				MockCheckPasswordHash,
				MockTime,
				MockFirebaseAdapter(),
				MockLineAdapter())
			err := authService.UpdatePassword(tt.args.username, tt.args.currentPassword, tt.args.newPassword)

			if tt.wantErr {
				assert.NotNil(t, err)
				assert.Empty(t, "")
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestAuthService_AccessShop(t *testing.T) {
	authRepo := new(AuthenticationRepositoryMock)
	shopUserRepo := new(ShopUserRepositoryMock)
	shopUserAccessLogRepo := new(ShopUserAccessLogRepositoryMock)
	smsRepo := new(SMSRepositoryMock)
	microAuthServiceMock := &AuthServiceMock{}

	microAuthServiceMock.On("GetTokenFromAuthorizationHeader", "authorization_header_valid").Return("valid_token", nil)
	microAuthServiceMock.On("GetTokenFromAuthorizationHeader", "").Return("", errors.New("authorization is not empty"))

	shopUser := models.ShopUser{}
	shopUser.ID = MockObjectID()
	shopUser.Username = "user_access_shop"
	shopUser.ShopID = "shop_test"
	shopUser.Role = uint8(0)

	shopUserRepo.On("FindByShopIDAndUsername", "shop_test", "user_access_shop").Return(shopUser, nil)

	shopUserRepo.On("FindByShopIDAndUsername", "shop_test_invalid", "user_access_shop").Return(models.ShopUser{}, nil)

	microAuthServiceMock.On("SelectShop", "valid_token", "shop_test", uint8(0)).Return(nil)
	microAuthServiceMock.On("SelectShop", "valid_token_invalid", "shop_test_invalid", uint8(0)).Return(errors.New("select shop failed"))

	type args struct {
		shopID              string
		username            string
		authorizationHeader string
	}

	cases := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success access shop ",
			args: args{
				shopID:              "shop_test",
				username:            "user_access_shop",
				authorizationHeader: "authorization_header_valid",
			},
			wantErr: false,
		},
		{
			name: "failure authorization empty",
			args: args{
				shopID:              "shop_test",
				username:            "user_access_shop",
				authorizationHeader: "",
			},
			wantErr: true,
		},
		{
			name: "failure shop invalid",
			args: args{
				shopID:              "shop_test_invalid",
				username:            "user_access_shop",
				authorizationHeader: "authorization_header_valid",
			},
			wantErr: true,
		},
		{
			name: "failure access shop failed",
			args: args{
				shopID:              "shop_test_invalid",
				username:            "user_access_shop",
				authorizationHeader: "authorization_header_valid",
			},
			wantErr: true,
		},
	}

	authContext := models.AuthenticationContext{
		Ip: "localhost",
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			authService := services.NewAuthenticationService(
				authRepo,
				shopUserRepo,
				shopUserAccessLogRepo,
				smsRepo,
				microAuthServiceMock,
				MockRandomString,
				MockRandomNumber,
				MockGUID,
				MockHashPassword,
				MockCheckPasswordHash,
				MockTime,
				MockFirebaseAdapter(),
				MockLineAdapter())
			err := authService.AccessShop(tt.args.shopID, tt.args.username, tt.args.authorizationHeader, authContext)

			if tt.wantErr {
				assert.NotNil(t, err)
				assert.Empty(t, "")
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

type AuthenticationRepositoryMock struct {
	mock.Mock
}

func (m *AuthenticationRepositoryMock) FindByIdentity(ctx context.Context, fieldName string, value string) (*models.UserDoc, error) {

	args := m.Called(fieldName, value)

	return args.Get(0).(*models.UserDoc), args.Error(1)
}

func (m *AuthenticationRepositoryMock) FindUser(ctx context.Context, id string) (*models.UserDoc, error) {
	args := m.Called(id)
	return args.Get(0).(*models.UserDoc), args.Error(1)
}

func (m *AuthenticationRepositoryMock) FindByPhonenumber(ctx context.Context, phonenumber models.PhoneNumberField) (*models.UserDoc, error) {
	args := m.Called(phonenumber)
	return args.Get(0).(*models.UserDoc), args.Error(1)
}

func (m *AuthenticationRepositoryMock) CreateUser(ctx context.Context, doc models.UserDoc) (primitive.ObjectID, error) {
	args := m.Called(doc)
	return args.Get(0).(primitive.ObjectID), args.Error(1)
}

func (m *AuthenticationRepositoryMock) UpdateUser(ctx context.Context, username string, userDoc models.UserDoc) error {
	args := m.Called(username, userDoc)
	return args.Error(0)
}

func (m *AuthenticationRepositoryMock) DeleteUser(ctx context.Context, username string) error {
	args := m.Called(ctx, username)
	return args.Error(0)
}

type ShopUserRepositoryMock struct {
	mock.Mock
}

func (m *ShopUserRepositoryMock) Create(ctx context.Context, shopUser *models.ShopUser) error {
	args := m.Called(ctx, shopUser)
	return args.Error(0)
}

func (m *ShopUserRepositoryMock) Update(ctx context.Context, id primitive.ObjectID, shopID string, username string, role models.UserRole) error {
	args := m.Called(id, shopID, username, role)
	return args.Error(0)
}

func (m *ShopUserRepositoryMock) Save(ctx context.Context, shopID string, username string, role models.UserRole) error {
	args := m.Called(shopID, username, role)
	return args.Error(0)
}

func (m *ShopUserRepositoryMock) UpdateLastAccess(ctx context.Context, shopID string, username string, lastAccessedAt time.Time) error {
	args := m.Called(shopID, username, lastAccessedAt)
	return args.Error(0)
}

func (m *ShopUserRepositoryMock) SaveFavorite(ctx context.Context, shopID string, username string, isFavorite bool) error {
	args := m.Called(shopID, username, isFavorite)
	return args.Error(0)
}

func (m *ShopUserRepositoryMock) Delete(ctx context.Context, shopID string, username string) error {
	args := m.Called(shopID, username)
	return args.Error(0)
}

func (m *ShopUserRepositoryMock) FindByShopIDAndUsernameInfo(ctx context.Context, shopID string, username string) (models.ShopUserInfo, error) {
	args := m.Called(shopID, username)
	return args.Get(0).(models.ShopUserInfo), args.Error(1)
}

func (m *ShopUserRepositoryMock) FindByShopIDAndUsername(ctx context.Context, shopID string, username string) (models.ShopUser, error) {
	args := m.Called(shopID, username)
	return args.Get(0).(models.ShopUser), args.Error(1)
}

func (m *ShopUserRepositoryMock) FindRole(ctx context.Context, shopID string, username string) (models.UserRole, error) {
	args := m.Called(shopID, username)
	return args.Get(0).(models.UserRole), args.Error(1)
}
func (m *ShopUserRepositoryMock) FindByShopID(ctx context.Context, shopID string) (*[]models.ShopUser, error) {
	args := m.Called(shopID)
	return args.Get(0).(*[]models.ShopUser), args.Error(1)
}

func (m *ShopUserRepositoryMock) FindByUsername(ctx context.Context, username string) (*[]models.ShopUser, error) {
	args := m.Called(username)
	return args.Get(0).(*[]models.ShopUser), args.Error(1)
}
func (m *ShopUserRepositoryMock) FindByUsernamePage(ctx context.Context, username string, pageable micromodels.Pageable) ([]models.ShopUserInfo, mongopagination.PaginationData, error) {
	args := m.Called(username, pageable)
	return args.Get(0).([]models.ShopUserInfo), args.Get(1).(mongopagination.PaginationData), args.Error(2)
}
func (m *ShopUserRepositoryMock) FindByUserInShopPage(ctx context.Context, shopID string, pageable micromodels.Pageable) ([]models.ShopUser, mongopagination.PaginationData, error) {
	args := m.Called(shopID, pageable)
	return args.Get(0).([]models.ShopUser), args.Get(1).(mongopagination.PaginationData), args.Error(2)
}

func (m *ShopUserRepositoryMock) FindUserProfileByUsernames(ctx context.Context, usernames []string) ([]models.UserProfile, error) {
	args := m.Called(usernames)
	return args.Get(0).([]models.UserProfile), args.Error(1)
}

func (m *ShopUserRepositoryMock) FindByShopIDAndUsernameAndRole(ctx context.Context, shopID string, username string, role models.UserRole) (models.ShopUser, error) {
	args := m.Called(shopID, username, role)
	return args.Get(0).(models.ShopUser), args.Error(1)
}

// Shop User Access Log
type ShopUserAccessLogRepositoryMock struct {
	mock.Mock
}

func (m *ShopUserAccessLogRepositoryMock) Create(ctx context.Context, shopUserAccessLog models.ShopUserAccessLog) error {
	args := m.Called(shopUserAccessLog)
	return args.Error(0)
}

type AuthServiceMock struct {
	mock.Mock
}

func (m *AuthServiceMock) MWFuncWithRedisMixShop(cacher microservice.ICacher, shopPath []string, publicPath ...string) echo.MiddlewareFunc {
	args := m.Called(cacher, shopPath, publicPath)
	return args.Get(0).(echo.MiddlewareFunc)
}

func (m *AuthServiceMock) MWFuncWithRedis(cacher microservice.ICacher, publicPath ...string) echo.MiddlewareFunc {
	args := m.Called(cacher, publicPath)
	return args.Get(0).(echo.MiddlewareFunc)
}

func (m *AuthServiceMock) MWFuncWithShop(cacher microservice.ICacher, publicPath ...string) echo.MiddlewareFunc {
	args := m.Called(cacher, publicPath)
	return args.Get(0).(echo.MiddlewareFunc)
}

func (m *AuthServiceMock) GetPrefixCacheKey(tokenType microservice.TokenType) string {
	args := m.Called(tokenType)
	return args.String(0)
}

func (m *AuthServiceMock) GetTokenFromContext(c echo.Context) (*microservice.TokenContext, error) {

	args := m.Called(c)

	return args.Get(0).(*microservice.TokenContext), args.Error(1)
}

func (m *AuthServiceMock) GetTokenFromAuthorizationHeader(tokenType microservice.TokenType, tokenAuthorization string) (string, error) {

	args := m.Called(tokenType, tokenAuthorization)

	return args.String(0), args.Error(1)
}

func (m *AuthServiceMock) GenerateTokenWithRedis(tokenType microservice.TokenType, userInfo micromodels.UserInfo) (string, error) {

	args := m.Called(tokenType, userInfo)
	return args.String(0), args.Error(1)
}

func (m *AuthServiceMock) GenerateTokenWithRedisExpire(tokenType microservice.TokenType, userInfo micromodels.UserInfo, expireTime time.Duration) (string, error) {

	args := m.Called(tokenType, userInfo, expireTime)
	return args.String(0), args.Error(1)
}

func (m *AuthServiceMock) SelectShop(tokenType microservice.TokenType, tokenStr string, shopID string, role uint8) error {

	args := m.Called(tokenType, tokenStr, shopID, role)
	return args.Error(0)
}

func (m *AuthServiceMock) ExpireToken(tokenType microservice.TokenType, tokenAuthorizationHeader string) error {
	args := m.Called(tokenType, tokenAuthorizationHeader)
	return args.Error(0)
}

func (m *AuthServiceMock) DeleteToken(tokenType microservice.TokenType, tokenStr string) error {
	args := m.Called(tokenType, tokenStr)
	return args.Error(0)
}

func (m *AuthServiceMock) RefreshToken(token string) (string, string, error) {
	args := m.Called(token)
	return args.String(0), args.String(1), args.Error(2)
}

type SMSRepositoryMock struct {
	mock.Mock
}

func (m *SMSRepositoryMock) SendSMS(phoneNumber string, message string, expire time.Duration) error {
	args := m.Called(phoneNumber, message, expire)
	return args.Error(0)
}

func (m *SMSRepositoryMock) SendOTP(phoneNumber string, refCode string, otpCode string, expire time.Duration) error {
	args := m.Called(phoneNumber, refCode, otpCode, expire)
	return args.Error(0)
}

func (m *SMSRepositoryMock) VerifyOTP(refCode string, otpCode string) (bool, error) {
	args := m.Called(refCode, otpCode)
	return args.Bool(0), args.Error(1)
}

func (m *SMSRepositoryMock) SendOTPViaLink(fullPhoneNumber string) (models.OTPResponse, error) {
	args := m.Called(fullPhoneNumber)
	return args.Get(0).(models.OTPResponse), args.Error(1)
}

func (m *SMSRepositoryMock) VerifyOTPViaLink(otpToken, optRefCode, otpPin string) (bool, error) {
	args := m.Called(otpToken, optRefCode, otpPin)
	return args.Bool(0), args.Error(1)
}

func MockObjectID() primitive.ObjectID {
	idx, _ := primitive.ObjectIDFromHex("62f9cb12c76fd9e83ac1b2ff")
	return idx
}

func MockHashPassword(password string) (string, error) {
	return password, nil
}

func MockCheckPasswordHash(password string, hash string) bool {
	return password == hash
}

func MockTime() time.Time {
	timeVal, _ := time.Parse("2006-01-02 15:04:05", "2022-08-30 00:00:00")
	return timeVal
}

func MockFirebaseAdapter() firebase.IFirebaseAdapter {
	return &firebase.FirebaseAdapter{}
}

func MockLineAdapter() line.ILineAdapter {
	return &line.LineAdapter{}
}

func MockGUID() string {
	return "mock_guid"
}

func MockRandomString(n int) string {
	return "123456"
}

func MockRandomNumber(n int) string {
	return "123456"
}
