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

	holdingCode := "HOLDING_CODE_TEST"
	role := uint8(2)

	tokenMock := "TOKEN_MOCK"
	//authRepo FindUser
	userDoc1 := models.UserDoc{}
	userDoc1.UID = "uid-tester1"
	userDoc1.Username = "tester1"
	userDoc1.Password = "tester1"
	userDoc1.Name = "tester1"

	authRepo.On("FindUser", userDoc1.Username).Return(&userDoc1, nil)
	authRepo.On("FindUser", "").Return(&models.UserDoc{}, nil)
	authRepo.On("FindUser", "tester2").Return(&models.UserDoc{}, nil)

	authRepo.On("FindUser", "user_register").Return(&models.UserDoc{}, nil)
	authRepo.On("FindByIdentity", "email", "user_register").Return(&models.UserDoc{}, nil)
	authRepo.On("FindByIdentity", "email", userDoc1.Username).Return(&userDoc1, nil)

	//authRepo CreateUser
	userDoc2 := models.UserDoc{}
	userDoc2.UID = MockGUID()
	userDoc2.Username = "user_register"
	userDoc2.Email = "user_register"
	userDoc2.Password = "register_password_success"
	userDoc2.Name = "user_register"
	userDoc2.CreatedAt = MockTime()

	authRepo.On("CreateUser", mock.MatchedBy(func(doc models.UserDoc) bool {
		return doc.UID == userDoc2.UID &&
			doc.Username == userDoc2.Username &&
			doc.Email == userDoc2.Email &&
			doc.Password == userDoc2.Password &&
			doc.Name == userDoc2.Name &&
			doc.CreatedAt.Equal(userDoc2.CreatedAt)
	})).Return(MockObjectID(), nil)

	//microAuth
	microAuthServiceMock.On("GenerateTokenWithRedis", microservice.AUTHTYPE_BEARER, micromodels.UserInfo{
		Username: userDoc1.Username,
		Name:     userDoc1.Name,
		UID:      userDoc1.UID,
	}).Return(tokenMock, nil)
	microAuthServiceMock.On("GenerateTokenWithRedis", microservice.AUTHTYPE_REFRESH, micromodels.UserInfo{
		Username: userDoc1.Username,
		Name:     userDoc1.Name,
		UID:      userDoc1.UID,
	}).Return(tokenMock, nil)

	microAuthServiceMock.On("SelectShop", microservice.AUTHTYPE_BEARER, tokenMock, holdingCode, role).Return(nil)

	shopUser := models.ShopUser{}
	shopUser.ID = MockObjectID()
	shopUser.Username = userDoc1.Username
	shopUser.UserUID = userDoc1.UID
	shopUser.HoldingCode = holdingCode
	shopUser.Role = role

	//shopUser
	shopUserRepo.On("FindByHoldingCodeAndUserUID", holdingCode, userDoc1.UID).Return(shopUser, nil)
	shopUserRepo.On("FindByHoldingCodeAndUsername", holdingCode, userDoc1.Username).Return(shopUser, nil)

	shopUserRepo.On("FindByHoldingCodeAndUserUID", "HOLDING_CODE_INVALID", userDoc1.UID).Return(models.ShopUser{}, nil)
	shopUserRepo.On("FindByHoldingCodeAndUsername", "HOLDING_CODE_INVALID", userDoc1.Username).Return(models.ShopUser{}, nil)
	shopUserRepo.On("UpdateLastAccess", holdingCode, userDoc1.Username, MockTime()).Return(nil)
}

func TestAuthService_Login(t *testing.T) {
	holdingCode := "HOLDING_CODE_TEST"

	authRepo := new(AuthenticationRepositoryMock)
	shopUserRepo := new(ShopUserRepositoryMock)
	shopUserAccessLogRepo := new(ShopUserAccessLogRepositoryMock)
	smsRepo := new(SMSRepositoryMock)
	microAuthServiceMock := &AuthServiceMock{}

	mockLoginData(authRepo, shopUserRepo, microAuthServiceMock)

	type args struct {
		username    string
		password    string
		holdingCode string
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
				holdingCode: holdingCode,
				username:    "tester1",
				password:    "tester1",
			},
			wantErr:  false,
			wantData: "TOKEN_MOCK",
		},
		{
			name: "login success without holding code",
			args: args{
				username: "tester1",
				password: "tester1",
			},
			wantErr:  false,
			wantData: "TOKEN_MOCK",
		},
		{
			name: "login failure invalid holding code",
			args: args{
				holdingCode: "HOLDING_CODE_INVALID",
				username:    "tester1",
				password:    "tester1",
			},
			wantErr:  true,
			wantData: "TOKEN_MOCK",
		},
		{
			name: "login failure password invalid",
			args: args{
				holdingCode: holdingCode,
				username:    "tester1",
				password:    "invalidpassword",
			},
			wantErr:  true,
			wantData: "TOKEN_MOCK",
		},
		{
			name: "login failure username empty",
			args: args{
				holdingCode: holdingCode,
				username:    "",
				password:    "invalidpassword",
			},
			wantErr:  true,
			wantData: "TOKEN_MOCK",
		},
		{
			name: "login failure username and password empty",
			args: args{
				holdingCode: holdingCode,
				username:    "",
				password:    "",
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
			userReq.HoldingCode = tt.args.holdingCode

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
				assert.EqualValues(t, tt.wantData, tokenResult.Token)
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

func TestAuthService_ProfileFlagsDefaultPassword(t *testing.T) {
	authRepo := new(AuthenticationRepositoryMock)
	shopUserRepo := new(ShopUserRepositoryMock)
	shopUserAccessLogRepo := new(ShopUserAccessLogRepositoryMock)
	smsRepo := new(SMSRepositoryMock)
	microAuthServiceMock := &AuthServiceMock{}

	userDoc := &models.UserDoc{}
	userDoc.Username = "default_user"
	userDoc.Email = "default_user@example.com"
	userDoc.Password = models.DefaultUserPassword
	userDoc.Name = "Default User"

	authRepo.On("FindUser", "default_user").Return(userDoc, nil)

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

	profile, err := authService.Profile("default_user", "")

	assert.Nil(t, err)
	assert.Equal(t, "default_user@example.com", profile.Email)
	assert.True(t, profile.IsDefaultPassword)
	assert.Empty(t, profile.Password)
}

func TestAuthService_ResetPasswordToDefault(t *testing.T) {
	authRepo := new(AuthenticationRepositoryMock)
	shopUserRepo := new(ShopUserRepositoryMock)
	shopUserAccessLogRepo := new(ShopUserAccessLogRepositoryMock)
	smsRepo := new(SMSRepositoryMock)
	microAuthServiceMock := &AuthServiceMock{}

	owner := models.ShopUser{}
	owner.HoldingCode = "shop_test"
	owner.Username = "owner_user"
	owner.Role = models.ROLE_OWNER

	targetShopUser := models.ShopUser{}
	targetShopUser.HoldingCode = "shop_test"
	targetShopUser.Username = "target_user"
	targetShopUser.Role = models.ROLE_USER

	targetUser := &models.UserDoc{}
	targetUser.Username = "target_user"
	targetUser.Password = "old_hash"

	expectedUser := models.UserDoc{}
	expectedUser.Username = "target_user"
	expectedUser.Password = models.DefaultUserPassword
	expectedUser.UpdatedAt = MockTime()

	shopUserRepo.On("FindByHoldingCodeAndUsername", "shop_test", "owner_user").Return(owner, nil)
	shopUserRepo.On("FindByHoldingCodeAndUsername", "shop_test", "target_user").Return(targetShopUser, nil)
	authRepo.On("FindUser", "target_user").Return(targetUser, nil)
	authRepo.On("UpdateUser", "target_user", expectedUser).Return(nil)

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

	err := authService.ResetPasswordToDefault("shop_test", "owner_user", "target_user")

	assert.Nil(t, err)
}

func TestAuthService_AccessShop(t *testing.T) {
	authRepo := new(AuthenticationRepositoryMock)
	shopUserRepo := new(ShopUserRepositoryMock)
	shopUserAccessLogRepo := new(ShopUserAccessLogRepositoryMock)
	smsRepo := new(SMSRepositoryMock)
	microAuthServiceMock := &AuthServiceMock{}

	microAuthServiceMock.On("GetTokenFromAuthorizationHeader", microservice.AUTHTYPE_BEARER, "authorization_header_valid").Return("valid_token", nil)
	microAuthServiceMock.On("GetTokenFromAuthorizationHeader", microservice.AUTHTYPE_BEARER, "").Return("", errors.New("authorization is not empty"))

	shopUser := models.ShopUser{}
	shopUser.ID = MockObjectID()
	shopUser.Username = "user_access_shop"
	shopUser.HoldingCode = "shop_test"
	shopUser.Role = uint8(0)

	shopUserRepo.On("FindByHoldingCodeAndUsername", "shop_test", "user_access_shop").Return(shopUser, nil)

	shopUserRepo.On("FindByHoldingCodeAndUsername", "shop_test_invalid", "user_access_shop").Return(models.ShopUser{}, nil)

	disabledShopUser := shopUser
	disabledShopUser.Username = "disabled_user"
	disabledShopUser.HoldingCode = "shop_disabled"
	disabledShopUser.IsAccessDisabled = true

	disabledCreator := shopUser
	disabledCreator.Username = "creator_user"
	disabledCreator.HoldingCode = "shop_disabled_creator"
	disabledCreator.IsAccessDisabled = true

	shopUserRepo.On("FindByHoldingCodeAndUsername", "shop_disabled", "disabled_user").Return(disabledShopUser, nil)
	shopUserRepo.On("FindShopCreatedBy", "shop_disabled").Return("creator_user", nil)
	shopUserRepo.On("FindByHoldingCodeAndUsername", "shop_disabled_creator", "creator_user").Return(disabledCreator, nil)
	shopUserRepo.On("FindShopCreatedBy", "shop_disabled_creator").Return("creator_user", nil)
	shopUserRepo.On("UpdateLastAccess", "shop_test", "user_access_shop", MockTime()).Return(nil)
	shopUserRepo.On("UpdateLastAccess", "shop_disabled_creator", "creator_user", MockTime()).Return(nil)
	shopUserAccessLogRepo.On("Create", mock.MatchedBy(func(log models.ShopUserAccessLog) bool {
		return log.HoldingCode == "shop_test" &&
			log.Username == "user_access_shop" &&
			log.Ip == "localhost" &&
			log.LastAccessedAt.Equal(MockTime())
	})).Return(nil)
	shopUserAccessLogRepo.On("Create", mock.MatchedBy(func(log models.ShopUserAccessLog) bool {
		return log.HoldingCode == "shop_disabled_creator" &&
			log.Username == "creator_user" &&
			log.Ip == "localhost" &&
			log.LastAccessedAt.Equal(MockTime())
	})).Return(nil)

	microAuthServiceMock.On("SelectShop", microservice.AUTHTYPE_BEARER, "valid_token", "shop_test", uint8(0)).Return(nil)
	microAuthServiceMock.On("SelectShop", microservice.AUTHTYPE_BEARER, "valid_token", "shop_disabled_creator", uint8(0)).Return(nil)
	microAuthServiceMock.On("SelectShop", microservice.AUTHTYPE_BEARER, "valid_token_invalid", "shop_test_invalid", uint8(0)).Return(errors.New("select shop failed"))

	type args struct {
		holdingCode         string
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
				holdingCode:         "shop_test",
				username:            "user_access_shop",
				authorizationHeader: "authorization_header_valid",
			},
			wantErr: false,
		},
		{
			name: "failure authorization empty",
			args: args{
				holdingCode:         "shop_test",
				username:            "user_access_shop",
				authorizationHeader: "",
			},
			wantErr: true,
		},
		{
			name: "failure shop invalid",
			args: args{
				holdingCode:         "shop_test_invalid",
				username:            "user_access_shop",
				authorizationHeader: "authorization_header_valid",
			},
			wantErr: true,
		},
		{
			name: "failure access shop failed",
			args: args{
				holdingCode:         "shop_test_invalid",
				username:            "user_access_shop",
				authorizationHeader: "authorization_header_valid",
			},
			wantErr: true,
		},
		{
			name: "failure access disabled user",
			args: args{
				holdingCode:         "shop_disabled",
				username:            "disabled_user",
				authorizationHeader: "authorization_header_valid",
			},
			wantErr: true,
		},
		{
			name: "success disabled creator still access",
			args: args{
				holdingCode:         "shop_disabled_creator",
				username:            "creator_user",
				authorizationHeader: "authorization_header_valid",
			},
			wantErr: false,
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
			err := authService.AccessShop(tt.args.holdingCode, tt.args.username, "", tt.args.authorizationHeader, authContext)

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

func (m *AuthenticationRepositoryMock) FindByLineUserID(ctx context.Context, lineUserID string) (*models.UserDoc, error) {
	args := m.Called(lineUserID)
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

func (m *ShopUserRepositoryMock) Update(ctx context.Context, id primitive.ObjectID, holdingCode string, username string, role models.UserRole) error {
	args := m.Called(id, holdingCode, username, role)
	return args.Error(0)
}

func (m *ShopUserRepositoryMock) Save(ctx context.Context, holdingCode string, username string, role models.UserRole) error {
	args := m.Called(holdingCode, username, role)
	return args.Error(0)
}

func (m *ShopUserRepositoryMock) SaveFullProfile(ctx context.Context, holdingCode string, req *models.UserRoleRequest) error {
	args := m.Called(holdingCode, req)
	return args.Error(0)
}

func (m *ShopUserRepositoryMock) UpdateLastAccess(ctx context.Context, holdingCode string, username string, lastAccessedAt time.Time) error {
	args := m.Called(holdingCode, username, lastAccessedAt)
	return args.Error(0)
}

func (m *ShopUserRepositoryMock) SaveFavorite(ctx context.Context, holdingCode string, username string, isFavorite bool) error {
	args := m.Called(holdingCode, username, isFavorite)
	return args.Error(0)
}

func (m *ShopUserRepositoryMock) Delete(ctx context.Context, holdingCode string, username string) error {
	args := m.Called(holdingCode, username)
	return args.Error(0)
}

func (m *ShopUserRepositoryMock) DeleteEmptyUsernames(ctx context.Context, holdingCode string) (int64, error) {
	args := m.Called(holdingCode)
	return args.Get(0).(int64), args.Error(1)
}

func (m *ShopUserRepositoryMock) FindByHoldingCodeAndUsernameInfo(ctx context.Context, holdingCode string, username string) (models.ShopUserInfo, error) {
	args := m.Called(holdingCode, username)
	return args.Get(0).(models.ShopUserInfo), args.Error(1)
}

func (m *ShopUserRepositoryMock) FindByHoldingCodeAndUserUIDInfo(ctx context.Context, holdingCode string, userUID string) (models.ShopUserInfo, error) {
	args := m.Called(holdingCode, userUID)
	return args.Get(0).(models.ShopUserInfo), args.Error(1)
}

func (m *ShopUserRepositoryMock) FindByHoldingCodeAndUserUID(ctx context.Context, holdingCode string, userUID string) (models.ShopUser, error) {
	args := m.Called(holdingCode, userUID)
	return args.Get(0).(models.ShopUser), args.Error(1)
}

func (m *ShopUserRepositoryMock) ResolveHoldingCodeByHoldingCode(ctx context.Context, holdingCode string) (string, error) {
	args := m.Called(holdingCode)
	return args.String(0), args.Error(1)
}

func (m *ShopUserRepositoryMock) FindByHoldingCodeAndUsername(ctx context.Context, holdingCode string, username string) (models.ShopUser, error) {
	args := m.Called(holdingCode, username)
	return args.Get(0).(models.ShopUser), args.Error(1)
}

func (m *ShopUserRepositoryMock) FindByHoldingCodeAndLineUserID(ctx context.Context, holdingCode string, lineUserID string) (models.ShopUser, error) {
	args := m.Called(holdingCode, lineUserID)
	return args.Get(0).(models.ShopUser), args.Error(1)
}

func (m *ShopUserRepositoryMock) FindByLineUserID(ctx context.Context, lineUserID string) (models.ShopUser, error) {
	args := m.Called(lineUserID)
	return args.Get(0).(models.ShopUser), args.Error(1)
}

func (m *ShopUserRepositoryMock) FindShopCreatedBy(ctx context.Context, holdingCode string) (string, error) {
	args := m.Called(holdingCode)
	return args.String(0), args.Error(1)
}

func (m *ShopUserRepositoryMock) FindRole(ctx context.Context, holdingCode string, username string) (models.UserRole, error) {
	args := m.Called(holdingCode, username)
	return args.Get(0).(models.UserRole), args.Error(1)
}
func (m *ShopUserRepositoryMock) FindByHoldingCode(ctx context.Context, holdingCode string) (*[]models.ShopUser, error) {
	args := m.Called(holdingCode)
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

func (m *ShopUserRepositoryMock) FindByUserUIDPage(ctx context.Context, userUID string, pageable micromodels.Pageable) ([]models.ShopUserInfo, mongopagination.PaginationData, error) {
	args := m.Called(userUID, pageable)
	return args.Get(0).([]models.ShopUserInfo), args.Get(1).(mongopagination.PaginationData), args.Error(2)
}

func (m *ShopUserRepositoryMock) FindByUserInShopPage(ctx context.Context, holdingCode string, pageable micromodels.Pageable) ([]models.ShopUser, mongopagination.PaginationData, error) {
	args := m.Called(holdingCode, pageable)
	return args.Get(0).([]models.ShopUser), args.Get(1).(mongopagination.PaginationData), args.Error(2)
}

func (m *ShopUserRepositoryMock) FindByUserInShopPageWithProfileMatches(ctx context.Context, holdingCode string, pageable micromodels.Pageable, profileUsernames []string) ([]models.ShopUser, mongopagination.PaginationData, error) {
	args := m.Called(holdingCode, pageable, profileUsernames)
	return args.Get(0).([]models.ShopUser), args.Get(1).(mongopagination.PaginationData), args.Error(2)
}

func (m *ShopUserRepositoryMock) FindUsernamesByProfileQuery(ctx context.Context, query string) ([]string, error) {
	args := m.Called(query)
	return args.Get(0).([]string), args.Error(1)
}

func (m *ShopUserRepositoryMock) FindUserProfileByUsernames(ctx context.Context, usernames []string) ([]models.UserProfile, error) {
	args := m.Called(usernames)
	return args.Get(0).([]models.UserProfile), args.Error(1)
}

func (m *ShopUserRepositoryMock) FindByHoldingCodeAndUsernameAndRole(ctx context.Context, holdingCode string, username string, role models.UserRole) (models.ShopUser, error) {
	args := m.Called(holdingCode, username, role)
	return args.Get(0).(models.ShopUser), args.Error(1)
}

// Shop User Access Log
type ShopUserAccessLogRepositoryMock struct {
	mock.Mock
}

func (m *ShopUserAccessLogRepositoryMock) Create(ctx context.Context, shopUserAccessLog models.ShopUserAccessLog) error {
	if len(m.ExpectedCalls) == 0 {
		return nil
	}
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

func (m *AuthServiceMock) SelectShop(tokenType microservice.TokenType, tokenStr string, holdingCode string, role uint8) error {

	args := m.Called(tokenType, tokenStr, holdingCode, role)
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
	return "mockguid"
}

func MockRandomString(n int) string {
	return "123456"
}

func MockRandomNumber(n int) string {
	return "123456"
}
