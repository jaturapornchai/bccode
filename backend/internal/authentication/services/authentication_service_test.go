package services_test

import (
	"context"
	"errors"
	"smlcloudplatform/internal/authentication/models"
	"smlcloudplatform/internal/authentication/services"
	"smlcloudplatform/internal/firebase"
	"smlcloudplatform/internal/line"
	"smlcloudplatform/pkg/apperr"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/smlsoft/mongopagination"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func mockLoginData(authRepo *AuthenticationRepositoryMock, shopUserRepo *ShopUserRepositoryMock, microAuthServiceMock *AuthServiceMock) {

	holdingCode := "holdingtest"
	role := uint8(2)

	tokenMock := "TOKEN_MOCK"
	//authRepo FindUser
	userDoc1 := models.UserDoc{}
	userDoc1.UID = "uid-tester1"
	userDoc1.Username = "tester1"
	userDoc1.Password = "valid_password_123"
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
	loginUserInfo := micromodels.UserInfo{
		Username: userDoc1.Username,
		Name:     userDoc1.Name,
		UID:      userDoc1.UID,
	}
	microAuthServiceMock.On("CreateSession", loginUserInfo).Return(tokenMock, tokenMock, nil)
	microAuthServiceMock.On("RevokeSession", "Bearer "+tokenMock).Return(nil).Maybe()

	microAuthServiceMock.On("SelectShop", microservice.AUTHTYPE_BEARER, tokenMock, holdingCode, "", "", role).Return(nil)

	shopUser := models.ShopUser{}
	shopUser.ID = MockObjectID()
	shopUser.Username = userDoc1.Username
	shopUser.UserUID = userDoc1.UID
	shopUser.HoldingCode = holdingCode
	shopUser.Role = role

	//shopUser
	shopUserRepo.On("FindByHoldingCodeAndUserUID", holdingCode, userDoc1.UID).Return(shopUser, nil)
	shopUserRepo.On("FindByHoldingCodeAndUsername", holdingCode, userDoc1.Username).Return(shopUser, nil)
	shopUserRepo.On("ResolveHoldingCodeByHoldingCode", holdingCode).Return(holdingCode, nil)

	shopUserRepo.On("ResolveHoldingCodeByHoldingCode", "holdingmissing").Return("holdingmissing", nil)
	shopUserRepo.On("FindByHoldingCodeAndUserUID", "holdingmissing", userDoc1.UID).Return(models.ShopUser{}, nil)
	shopUserRepo.On("FindByHoldingCodeAndUsername", "holdingmissing", userDoc1.Username).Return(models.ShopUser{}, nil)
	shopUserRepo.On("UpdateLastAccess", holdingCode, userDoc1.UID, MockTime()).Return(nil)
}

func TestAuthService_Login(t *testing.T) {
	holdingCode := "holdingtest"

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
				password:    "valid_password_123",
			},
			wantErr:  false,
			wantData: "TOKEN_MOCK",
		},
		{
			name: "login success without holding code",
			args: args{
				username: "tester1",
				password: "valid_password_123",
			},
			wantErr:  false,
			wantData: "TOKEN_MOCK",
		},
		{
			name: "login failure invalid holding code",
			args: args{
				holdingCode: "holdingmissing",
				username:    "tester1",
				password:    "valid_password_123",
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

func TestAuthService_LoginRevokesSessionWhenStableMembershipIsMissing(t *testing.T) {
	authRepo := new(AuthenticationRepositoryMock)
	shopUserRepo := new(ShopUserRepositoryMock)
	microAuthServiceMock := &AuthServiceMock{}
	user := &models.UserDoc{}
	user.UID = "uid-member"
	user.Username = "member_user"
	user.Password = "valid_password_123"
	authRepo.On("FindUser", user.Username).Return(user, nil)
	shopUserRepo.On("ResolveHoldingCodeByHoldingCode", "holdingtest").Return("holdingtest", nil)
	shopUserRepo.On("FindByHoldingCodeAndUserUID", "holdingtest", user.UID).Return(models.ShopUser{}, nil)
	microAuthServiceMock.On("CreateSession", micromodels.UserInfo{Username: user.Username, UID: user.UID}).
		Return("access-token", "refresh-token", nil)
	microAuthServiceMock.On("RevokeSession", "Bearer access-token").Return(nil)
	authService := services.NewAuthenticationService(
		authRepo, shopUserRepo, new(ShopUserAccessLogRepositoryMock), new(SMSRepositoryMock),
		microAuthServiceMock, MockRandomString, MockRandomNumber, MockGUID, MockHashPassword,
		MockCheckPasswordHash, MockTime, MockFirebaseAdapter(), MockLineAdapter())

	_, err := authService.Login(&models.UserLoginRequest{
		UsernameField: models.UsernameField{Username: user.Username},
		UserPassword:  models.UserPassword{Password: user.Password},
		HoldingCode:   "holdingtest",
	}, models.AuthenticationContext{})

	assert.EqualError(t, err, "holdingcode invalid")
	microAuthServiceMock.AssertCalled(t, "RevokeSession", "Bearer access-token")
	shopUserRepo.AssertNotCalled(t, "FindByHoldingCodeAndUsername", mock.Anything, mock.Anything)
}

func TestAuthService_LoginDoesNotUseSharedPasswordState(t *testing.T) {
	authRepo := new(AuthenticationRepositoryMock)
	shopUserRepo := new(ShopUserRepositoryMock)
	shopUserAccessLogRepo := new(ShopUserAccessLogRepositoryMock)
	smsRepo := new(SMSRepositoryMock)
	microAuthServiceMock := &AuthServiceMock{}

	user := &models.UserDoc{}
	user.UID = "uid-default-user"
	user.Username = "default-user"
	user.Name = "ผู้ใช้เริ่มต้น"
	user.Password = "valid_password_123"
	authRepo.On("FindUser", user.Username).Return(user, nil)

	userInfo := micromodels.UserInfo{
		Username: user.Username,
		Name:     user.Name,
		UID:      user.UID,
	}
	microAuthServiceMock.On("CreateSession", userInfo).Return("bearer", "refresh", nil)

	authService := services.NewAuthenticationService(
		authRepo, shopUserRepo, shopUserAccessLogRepo, smsRepo, microAuthServiceMock,
		MockRandomString, MockRandomNumber, MockGUID, MockHashPassword, MockCheckPasswordHash,
		MockTime, MockFirebaseAdapter(), MockLineAdapter())

	result, err := authService.Login(&models.UserLoginRequest{
		UsernameField: models.UsernameField{Username: user.Username},
		UserPassword:  models.UserPassword{Password: user.Password},
	}, models.AuthenticationContext{Ip: "localhost"})

	assert.NoError(t, err)
	assert.Equal(t, "bearer", result.Token)
	assert.Equal(t, "refresh", result.Refresh)
}

func TestAuthService_OTPLoginDoesNotUseSharedPasswordState(t *testing.T) {
	authRepo := new(AuthenticationRepositoryMock)
	shopUserRepo := new(ShopUserRepositoryMock)
	shopUserAccessLogRepo := new(ShopUserAccessLogRepositoryMock)
	smsRepo := new(SMSRepositoryMock)
	microAuthServiceMock := &AuthServiceMock{}

	user := &models.UserDoc{}
	user.UID = "uid-default-otp"
	user.Username = "default-otp"
	user.PhoneNumber = "0812345678"
	user.Password = "valid_password_123"
	smsRepo.On("VerifyOTP", "ref", "123456").Return(true, nil)
	authRepo.On("FindByIdentity", "phonenumber", user.PhoneNumber).Return(user, nil)
	userInfo := micromodels.UserInfo{Username: user.Username, UID: user.UID}
	microAuthServiceMock.On("CreateSession", userInfo).Return("bearer", "refresh", nil)

	authService := services.NewAuthenticationService(
		authRepo, shopUserRepo, shopUserAccessLogRepo, smsRepo, microAuthServiceMock,
		MockRandomString, MockRandomNumber, MockGUID, MockHashPassword, MockCheckPasswordHash,
		MockTime, MockFirebaseAdapter(), MockLineAdapter())

	result, err := authService.LoginWithPhoneNumberOTP(&models.PhoneNumberOTPRequest{
		PhoneNumber: user.PhoneNumber,
		RefCode:     "ref",
		OTP:         "123456",
	}, models.AuthenticationContext{})

	assert.NoError(t, err)
	assert.Equal(t, "bearer", result.Token)
	assert.Equal(t, "refresh", result.Refresh)
}

func TestAuthService_RefreshTokenReturnsRotatedTokens(t *testing.T) {
	microAuthServiceMock := &AuthServiceMock{}
	microAuthServiceMock.On("RefreshToken", "refresh-in").Return("bearer", "refresh-out", true, nil)
	authService := services.NewAuthenticationService(
		new(AuthenticationRepositoryMock), new(ShopUserRepositoryMock), new(ShopUserAccessLogRepositoryMock),
		new(SMSRepositoryMock), microAuthServiceMock, MockRandomString, MockRandomNumber, MockGUID,
		MockHashPassword, MockCheckPasswordHash, MockTime, MockFirebaseAdapter(), MockLineAdapter())

	result, err := authService.RefreshToken(models.TokenLoginRequest{Token: "refresh-in"})

	assert.NoError(t, err)
	assert.Equal(t, "bearer", result.Token)
	assert.Equal(t, "refresh-out", result.Refresh)
}

func TestAuthService_DevLoginByUIDCreatesSessionAndAudit(t *testing.T) {
	authRepo := new(AuthenticationRepositoryMock)
	microAuthServiceMock := &AuthServiceMock{}
	user := &models.UserDoc{}
	user.UID = "dev-user-uid"
	user.Username = "dev_user"
	user.Email = " JATURAPORNCHAI@GMAIL.COM "
	user.Name = "Dev User"
	authRepo.On("FindUserByUID", user.UID).Return(user, nil)
	microAuthServiceMock.On("CreateSession", micromodels.UserInfo{
		Username: user.Username,
		Name:     user.Name,
		UID:      user.UID,
	}).Return("access-token", "refresh-token", nil)
	authRepo.On("CreateAuthAudit", mock.MatchedBy(func(audit models.AuthAudit) bool {
		return audit.AuditUID == MockGUID() && audit.UserUID == user.UID &&
			audit.Action == "DEV_LOGIN" && audit.Outcome == "SUCCESS" &&
			audit.OccurredAt.Equal(MockTime().UTC()) && audit.Metadata == nil && audit.SessionUID == ""
	})).Return(nil)
	authService := services.NewAuthenticationService(
		authRepo, new(ShopUserRepositoryMock), new(ShopUserAccessLogRepositoryMock),
		new(SMSRepositoryMock), microAuthServiceMock, MockRandomString, MockRandomNumber,
		MockGUID, MockHashPassword, MockCheckPasswordHash, MockTime, MockFirebaseAdapter(), MockLineAdapter())

	result, err := authService.DevLoginByUID(user.UID, models.AuthenticationContext{Ip: "127.0.0.1"})

	assert.NoError(t, err)
	assert.Equal(t, "access-token", result.Token)
	assert.Equal(t, "refresh-token", result.Refresh)
	microAuthServiceMock.AssertNotCalled(t, "RevokeSession", mock.Anything)
}

func TestAuthService_DevLoginByUIDRevokesSessionWhenAuditFails(t *testing.T) {
	authRepo := new(AuthenticationRepositoryMock)
	microAuthServiceMock := &AuthServiceMock{}
	user := &models.UserDoc{}
	user.UID = "dev-user-uid"
	user.Username = "dev_user"
	user.Email = "jaturapornchai@gmail.com"
	authRepo.On("FindUserByUID", user.UID).Return(user, nil)
	microAuthServiceMock.On("CreateSession", micromodels.UserInfo{Username: user.Username, UID: user.UID}).Return("access-token", "refresh-token", nil)
	authRepo.On("CreateAuthAudit", mock.Anything).Return(errors.New("audit unavailable"))
	microAuthServiceMock.On("RevokeSession", "Bearer access-token").Return(nil)
	authService := services.NewAuthenticationService(
		authRepo, new(ShopUserRepositoryMock), new(ShopUserAccessLogRepositoryMock),
		new(SMSRepositoryMock), microAuthServiceMock, MockRandomString, MockRandomNumber,
		MockGUID, MockHashPassword, MockCheckPasswordHash, MockTime, MockFirebaseAdapter(), MockLineAdapter())

	result, err := authService.DevLoginByUID(user.UID, models.AuthenticationContext{})

	assert.Empty(t, result.Token)
	assert.Empty(t, result.Refresh)
	appErr := apperr.FromError(err)
	assert.NotNil(t, appErr)
	assert.Equal(t, "INTERNAL_ERROR", appErr.Code)
	microAuthServiceMock.AssertCalled(t, "RevokeSession", "Bearer access-token")
}

func TestAuthService_DevLoginByUIDRejectsUnavailableUsers(t *testing.T) {
	tests := []struct {
		name    string
		userUID string
		user    *models.UserDoc
		err     error
	}{
		{name: "missing", userUID: "missing-user", err: mongo.ErrNoDocuments},
		{name: "deleted", userUID: "deleted-user", user: &models.UserDoc{UserDetail: models.UserDetail{UID: "deleted-user"}, IsDeleted: true}},
		{name: "wrong email", userUID: "wrong-email-user", user: &models.UserDoc{UserDetail: models.UserDetail{UID: "wrong-email-user"}, EmailField: models.EmailField{Email: "other@example.com"}}},
		{name: "disabled", userUID: "disabled-user", user: &models.UserDoc{UserDetail: models.UserDetail{UID: "disabled-user"}, EmailField: models.EmailField{Email: "jaturapornchai@gmail.com"}, DisabledAt: MockTime()}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authRepo := new(AuthenticationRepositoryMock)
			microAuthServiceMock := &AuthServiceMock{}
			authRepo.On("FindUserByUID", tt.userUID).Return(tt.user, tt.err)
			authService := services.NewAuthenticationService(
				authRepo, new(ShopUserRepositoryMock), new(ShopUserAccessLogRepositoryMock),
				new(SMSRepositoryMock), microAuthServiceMock, MockRandomString, MockRandomNumber,
				MockGUID, MockHashPassword, MockCheckPasswordHash, MockTime, MockFirebaseAdapter(), MockLineAdapter())

			_, err := authService.DevLoginByUID(tt.userUID, models.AuthenticationContext{})

			assert.Error(t, err)
			microAuthServiceMock.AssertNotCalled(t, "CreateSession", mock.Anything)
			authRepo.AssertNotCalled(t, "CreateAuthAudit", mock.Anything)
		})
	}
}

func TestAuthService_GoogleLoginResolvesLinkedUserByIssuerAndSubject(t *testing.T) {
	authRepo := new(AuthenticationRepositoryMock)
	microAuthServiceMock := &AuthServiceMock{}
	user := &models.UserDoc{}
	user.UID = "uid-google-user"
	user.Username = "google_user"
	user.Email = "old@example.com"
	user.Name = "Google User"
	identity := &models.GoogleIdentity{
		IdentityUID: "identity-1",
		UserUID:     user.UID,
		Issuer:      "https://accounts.google.com",
		Subject:     "google-subject",
		IsActive:    true,
	}
	authRepo.On("FindGoogleIdentity", identity.Issuer, identity.Subject).Return(identity, nil)
	authRepo.On("FindUserByUID", user.UID).Return(user, nil)
	userInfo := micromodels.UserInfo{Username: user.Username, Name: user.Name, UID: user.UID}
	microAuthServiceMock.On("CreateSession", userInfo).Return("access-token", "refresh-token", nil)

	authService := services.NewAuthenticationService(
		authRepo, new(ShopUserRepositoryMock), new(ShopUserAccessLogRepositoryMock),
		new(SMSRepositoryMock), microAuthServiceMock, MockRandomString, MockRandomNumber,
		MockGUID, MockHashPassword, MockCheckPasswordHash, MockTime, MockFirebaseAdapter(), MockLineAdapter())

	result, err := authService.LoginWithGoogleIdentity("accounts.google.com", identity.Subject, "new@example.com", user.Name)

	assert.NoError(t, err)
	assert.Equal(t, "access-token", result.Token)
	assert.Equal(t, "refresh-token", result.Refresh)
	authRepo.AssertNotCalled(t, "FindUser", mock.Anything)
	authRepo.AssertNotCalled(t, "CreateGoogleUserIdentity", mock.Anything, mock.Anything, mock.Anything)
}

func TestAuthService_GoogleLoginCreatesUserIdentityAndAuditAtomically(t *testing.T) {
	authRepo := new(AuthenticationRepositoryMock)
	microAuthServiceMock := &AuthServiceMock{}
	issuer := "https://accounts.google.com"
	subject := "new-google-subject"
	email := "new@example.com"
	authRepo.On("FindGoogleIdentity", issuer, subject).Return((*models.GoogleIdentity)(nil), mongo.ErrNoDocuments)
	authRepo.On("CreateGoogleUserIdentity",
		mock.MatchedBy(func(user models.UserDoc) bool {
			return user.GuidFixed == MockGUID() && user.UID == MockGUID() && user.Username == "" &&
				user.Email == email && user.Name == "New Google User" && user.CreatedAt.Equal(MockTime().UTC())
		}),
		mock.MatchedBy(func(identity models.GoogleIdentity) bool {
			return identity.IdentityUID == MockGUID() && identity.UserUID == MockGUID() &&
				identity.Issuer == issuer && identity.Subject == subject && identity.VerifiedEmail == email &&
				identity.IsActive && identity.LinkedAt.Equal(MockTime().UTC())
		}),
		mock.MatchedBy(func(audit models.AuthAudit) bool {
			return audit.AuditUID == MockGUID() && audit.UserUID == MockGUID() &&
				audit.Action == "GOOGLE_IDENTITY_LINK" && audit.Outcome == "SUCCESS" &&
				audit.OccurredAt.Equal(MockTime().UTC())
		}),
	).Return(models.UserDoc{
		UserDetail: models.UserDetail{UID: MockGUID(), Name: "New Google User"},
	}, nil)
	userInfo := micromodels.UserInfo{Name: "New Google User", UID: MockGUID()}
	microAuthServiceMock.On("CreateSession", userInfo).Return("access-token", "refresh-token", nil)

	authService := services.NewAuthenticationService(
		authRepo, new(ShopUserRepositoryMock), new(ShopUserAccessLogRepositoryMock),
		new(SMSRepositoryMock), microAuthServiceMock, MockRandomString, MockRandomNumber,
		MockGUID, MockHashPassword, MockCheckPasswordHash, MockTime, MockFirebaseAdapter(), MockLineAdapter())

	result, err := authService.LoginWithGoogleIdentity(issuer, subject, " NEW@EXAMPLE.COM ", " New Google User ")

	assert.NoError(t, err)
	assert.Equal(t, "access-token", result.Token)
	assert.Equal(t, "refresh-token", result.Refresh)
}

func TestAuthService_GoogleLoginRejectsInactiveIdentity(t *testing.T) {
	authRepo := new(AuthenticationRepositoryMock)
	authRepo.On("FindGoogleIdentity", "https://accounts.google.com", "revoked-subject").Return(&models.GoogleIdentity{
		IdentityUID: "identity-revoked",
		UserUID:     "uid-revoked",
		IsActive:    false,
	}, nil)
	microAuthServiceMock := &AuthServiceMock{}
	authService := services.NewAuthenticationService(
		authRepo, new(ShopUserRepositoryMock), new(ShopUserAccessLogRepositoryMock),
		new(SMSRepositoryMock), microAuthServiceMock, MockRandomString, MockRandomNumber,
		MockGUID, MockHashPassword, MockCheckPasswordHash, MockTime, MockFirebaseAdapter(), MockLineAdapter())

	_, err := authService.LoginWithGoogleIdentity("accounts.google.com", "revoked-subject", "user@example.com", "User")

	assert.EqualError(t, err, "google identity is inactive")
	authRepo.AssertNotCalled(t, "FindUserByUID", mock.Anything)
	microAuthServiceMock.AssertNotCalled(t, "CreateSession", mock.Anything)
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
	userDoc.UID = "user-update-uid"
	userDoc.Name = "user_update"
	userDoc.UpdatedAt = MockTime()

	authRepo.On("FindUserByUID", "user-update-uid").Return(userDoc, nil)

	userDocUpdate := models.UserDoc{}
	userDocUpdate.Username = "user_update"
	userDocUpdate.UID = "user-update-uid"
	userDocUpdate.Name = "new name"
	userDocUpdate.UpdatedAt = MockTime()

	authRepo.On("UpdateUserByUID", "user-update-uid", userDocUpdate).Return(nil)

	type args struct {
		username string
		userUID  string
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
				userUID:  "user-update-uid",
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

			err := authService.Update(tt.args.userUID, userReq)

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
	userDocUpdate.Password = "new_password_secure"
	userDocUpdate.UpdatedAt = MockTime()

	authRepo.On("UpdateUser", "user_update", userDocUpdate).Return(nil)
	microAuthServiceMock.On("RevokeUserTokens", "user_update").Return(nil)

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
				newPassword:     "new_password_secure",
			},
			wantErr: false,
		},
		{
			name: "update password failure",
			args: args{
				username:        "user_update",
				currentPassword: "current_password_invalid",
				newPassword:     "new_password_secure",
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

func TestAuthService_UpdatePasswordRejectsShortPassword(t *testing.T) {
	authService := services.NewAuthenticationService(
		new(AuthenticationRepositoryMock),
		new(ShopUserRepositoryMock),
		new(ShopUserAccessLogRepositoryMock),
		new(SMSRepositoryMock),
		&AuthServiceMock{},
		MockRandomString,
		MockRandomNumber,
		MockGUID,
		MockHashPassword,
		MockCheckPasswordHash,
		MockTime,
		MockFirebaseAdapter(),
		MockLineAdapter())

	err := authService.UpdatePassword("default-user", "current_password", "12345")
	appErr := apperr.FromError(err)
	assert.NotNil(t, appErr)
	assert.Equal(t, "VALIDATION_FAILED", appErr.Code)
	assert.Equal(t, "newpassword", appErr.Field)
}

func TestAuthService_ProfileDoesNotExposeSharedPasswordState(t *testing.T) {
	authRepo := new(AuthenticationRepositoryMock)
	shopUserRepo := new(ShopUserRepositoryMock)
	shopUserAccessLogRepo := new(ShopUserAccessLogRepositoryMock)
	smsRepo := new(SMSRepositoryMock)
	microAuthServiceMock := &AuthServiceMock{}

	userDoc := &models.UserDoc{}
	userDoc.Username = "default_user"
	userDoc.Email = "default_user@example.com"
	userDoc.Password = "hashed_password"
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
	assert.Empty(t, profile.Password)
}

func TestAuthService_ResetPasswordToDefaultIsDisabled(t *testing.T) {
	authRepo := new(AuthenticationRepositoryMock)
	shopUserRepo := new(ShopUserRepositoryMock)

	authService := services.NewAuthenticationService(
		authRepo,
		shopUserRepo,
		new(ShopUserAccessLogRepositoryMock),
		new(SMSRepositoryMock),
		&AuthServiceMock{},
		MockRandomString,
		MockRandomNumber,
		MockGUID,
		MockHashPassword,
		MockCheckPasswordHash,
		MockTime,
		MockFirebaseAdapter(),
		MockLineAdapter())

	err := authService.ResetPasswordToDefault("shoptest", "owner_user", "target_user")

	appErr := apperr.FromError(err)
	assert.NotNil(t, appErr)
	assert.Equal(t, "FORBIDDEN", appErr.Code)
	authRepo.AssertNotCalled(t, "FindUser", "target_user")
	shopUserRepo.AssertNotCalled(t, "FindByHoldingCodeAndUsername", "shoptest", "target_user")
}

func TestAuthService_LogoutRevokesWholeSession(t *testing.T) {
	microAuthServiceMock := &AuthServiceMock{}
	microAuthServiceMock.On("RevokeSession", "Bearer access-token").Return(nil)
	authService := services.NewAuthenticationService(
		new(AuthenticationRepositoryMock), new(ShopUserRepositoryMock), new(ShopUserAccessLogRepositoryMock),
		new(SMSRepositoryMock), microAuthServiceMock, MockRandomString, MockRandomNumber, MockGUID,
		MockHashPassword, MockCheckPasswordHash, MockTime, MockFirebaseAdapter(), MockLineAdapter())

	err := authService.Logout("Bearer access-token")

	assert.NoError(t, err)
	microAuthServiceMock.AssertCalled(t, "RevokeSession", "Bearer access-token")
	microAuthServiceMock.AssertNotCalled(t, "ExpireToken", mock.Anything, mock.Anything)
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
	shopUser.UserUID = "user-access-shop-uid"
	shopUser.HoldingCode = "shoptest"
	shopUser.Role = uint8(0)
	shopUser.AccessScopes = []models.AccessScope{{ScopeType: "company", CompanyUID: "company-a-uid"}}

	shopUserRepo.On("FindByHoldingCodeAndUserUID", "shoptest", shopUser.UserUID).Return(shopUser, nil)
	shopUserRepo.On("ResolveCompanyUID", "shoptest", "COMP-A").Return("company-a-uid", nil)
	shopUserRepo.On("ResolveCompanyUID", "shoptest", "COMP-B").Return("company-b-uid", nil)
	branchOnlyShopUser := shopUser
	branchOnlyShopUser.Username = "branch_only_user"
	branchOnlyShopUser.UserUID = "branch-only-user-uid"
	branchOnlyShopUser.AccessScopes = []models.AccessScope{{ScopeType: "branch", CompanyUID: "company-a-uid", BranchUID: "branch-1-uid"}}
	shopUserRepo.On("FindByHoldingCodeAndUserUID", "shoptest", branchOnlyShopUser.UserUID).Return(branchOnlyShopUser, nil)

	shopUserRepo.On("FindByHoldingCodeAndUserUID", "shoptestinvalid", shopUser.UserUID).Return(models.ShopUser{}, nil)

	disabledShopUser := shopUser
	disabledShopUser.Username = "disabled_user"
	disabledShopUser.UserUID = "disabled-user-uid"
	disabledShopUser.HoldingCode = "shopdisabled"
	disabledShopUser.IsAccessDisabled = true

	disabledCreator := shopUser
	disabledCreator.Username = "creator_user"
	disabledCreator.UserUID = "creator-user-uid"
	disabledCreator.HoldingCode = "shopdisabledcreator"
	disabledCreator.IsAccessDisabled = true

	shopUserRepo.On("FindByHoldingCodeAndUserUID", "shopdisabled", disabledShopUser.UserUID).Return(disabledShopUser, nil)
	shopUserRepo.On("FindByHoldingCodeAndUserUID", "shopdisabledcreator", disabledCreator.UserUID).Return(disabledCreator, nil)
	shopUserRepo.On("UpdateLastAccess", "shoptest", shopUser.UserUID, MockTime()).Return(nil)
	shopUserRepo.On("UpdateLastAccess", "shoptest", branchOnlyShopUser.UserUID, MockTime()).Return(nil)
	shopUserAccessLogRepo.On("Create", mock.MatchedBy(func(log models.ShopUserAccessLog) bool {
		return log.HoldingCode == "shoptest" &&
			log.Username == "user_access_shop" &&
			log.Ip == "localhost" &&
			log.LastAccessedAt.Equal(MockTime())
	})).Return(nil)
	shopUserAccessLogRepo.On("Create", mock.MatchedBy(func(log models.ShopUserAccessLog) bool {
		return log.HoldingCode == "shoptest" && log.Username == "branch_only_user" && log.LastAccessedAt.Equal(MockTime())
	})).Return(nil)
	microAuthServiceMock.On("SelectShop", microservice.AUTHTYPE_BEARER, "valid_token", "shoptest", "", "", uint8(0)).Return(nil)
	microAuthServiceMock.On("SelectShop", microservice.AUTHTYPE_BEARER, "valid_token", "shoptest", "COMP-A", "", uint8(0)).Return(nil)
	microAuthServiceMock.On("SelectShop", microservice.AUTHTYPE_BEARER, "valid_token", "shoptest", "COMP-A", "branch-1-uid", uint8(0)).Return(nil)
	microAuthServiceMock.On("SelectShop", microservice.AUTHTYPE_BEARER, "valid_token_invalid", "shoptestinvalid", "", "", uint8(0)).Return(errors.New("select shop failed"))

	type args struct {
		holdingCode         string
		businessCode        string
		branchUID           string
		username            string
		userUID             string
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
				holdingCode:         "shoptest",
				username:            "user_access_shop",
				userUID:             shopUser.UserUID,
				authorizationHeader: "authorization_header_valid",
			},
			wantErr: false,
		},
		{
			name: "success access allowed company",
			args: args{
				holdingCode:         "shoptest",
				businessCode:        "comp-a",
				username:            "user_access_shop",
				userUID:             shopUser.UserUID,
				authorizationHeader: "authorization_header_valid",
			},
			wantErr: false,
		},
		{
			name: "failure company scope denied",
			args: args{
				holdingCode:         "shoptest",
				businessCode:        "COMP-B",
				username:            "user_access_shop",
				userUID:             shopUser.UserUID,
				authorizationHeader: "authorization_header_valid",
			},
			wantErr: true,
		},
		{
			name: "success access exact branch scope",
			args: args{
				holdingCode:         "shoptest",
				businessCode:        "COMP-A",
				branchUID:           "branch-1-uid",
				username:            "branch_only_user",
				userUID:             branchOnlyShopUser.UserUID,
				authorizationHeader: "authorization_header_valid",
			},
			wantErr: false,
		},
		{
			name: "failure branch-only scope cannot select company",
			args: args{
				holdingCode:         "shoptest",
				businessCode:        "COMP-A",
				username:            "branch_only_user",
				userUID:             branchOnlyShopUser.UserUID,
				authorizationHeader: "authorization_header_valid",
			},
			wantErr: true,
		},
		{
			name: "failure authorization empty",
			args: args{
				holdingCode:         "shoptest",
				username:            "user_access_shop",
				userUID:             shopUser.UserUID,
				authorizationHeader: "",
			},
			wantErr: true,
		},
		{
			name: "failure shop invalid",
			args: args{
				holdingCode:         "shoptestinvalid",
				username:            "user_access_shop",
				userUID:             shopUser.UserUID,
				authorizationHeader: "authorization_header_valid",
			},
			wantErr: true,
		},
		{
			name: "failure access shop failed",
			args: args{
				holdingCode:         "shoptestinvalid",
				username:            "user_access_shop",
				userUID:             shopUser.UserUID,
				authorizationHeader: "authorization_header_valid",
			},
			wantErr: true,
		},
		{
			name: "failure access disabled user",
			args: args{
				holdingCode:         "shopdisabled",
				username:            "disabled_user",
				userUID:             disabledShopUser.UserUID,
				authorizationHeader: "authorization_header_valid",
			},
			wantErr: true,
		},
		{
			name: "failure disabled creator cannot bypass membership state",
			args: args{
				holdingCode:         "shopdisabledcreator",
				username:            "creator_user",
				userUID:             disabledCreator.UserUID,
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
			err := authService.AccessShop(tt.args.holdingCode, tt.args.businessCode, tt.args.branchUID, tt.args.username, tt.args.userUID, tt.args.authorizationHeader, authContext)

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

func (m *AuthenticationRepositoryMock) UpdateUserByUID(ctx context.Context, userUID string, userDoc models.UserDoc) error {
	args := m.Called(userUID, userDoc)
	return args.Error(0)
}

func (m *AuthenticationRepositoryMock) DeleteUser(ctx context.Context, username string) error {
	args := m.Called(ctx, username)
	return args.Error(0)
}

func (m *AuthenticationRepositoryMock) FindGoogleIdentity(ctx context.Context, issuer string, subject string) (*models.GoogleIdentity, error) {
	args := m.Called(issuer, subject)
	identity, _ := args.Get(0).(*models.GoogleIdentity)
	return identity, args.Error(1)
}

func (m *AuthenticationRepositoryMock) FindUserByUID(ctx context.Context, userUID string) (*models.UserDoc, error) {
	args := m.Called(userUID)
	user, _ := args.Get(0).(*models.UserDoc)
	return user, args.Error(1)
}

func (m *AuthenticationRepositoryMock) CreateAuthAudit(ctx context.Context, audit models.AuthAudit) error {
	args := m.Called(audit)
	return args.Error(0)
}

func (m *AuthenticationRepositoryMock) CreateGoogleUserIdentity(ctx context.Context, user models.UserDoc, identity models.GoogleIdentity, audit models.AuthAudit) (models.UserDoc, error) {
	args := m.Called(user, identity, audit)
	linkedUser, _ := args.Get(0).(models.UserDoc)
	return linkedUser, args.Error(1)
}

func (m *AuthenticationRepositoryMock) EnsureGoogleIdentityIndexes(ctx context.Context) error {
	args := m.Called()
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

func (m *ShopUserRepositoryMock) SaveStable(ctx context.Context, holdingCode string, holdingUID string, userUID string, username string, role models.UserRole, createdAt time.Time) error {
	args := m.Called(ctx, holdingCode, holdingUID, userUID, username, role, createdAt)
	return args.Error(0)
}

func (m *ShopUserRepositoryMock) SaveFullProfile(ctx context.Context, holdingCode string, req *models.UserRoleRequest) error {
	args := m.Called(holdingCode, req)
	return args.Error(0)
}

func (m *ShopUserRepositoryMock) UpdateLineFields(ctx context.Context, holdingCode string, userUID string, lineUserID string, lineDisplayName string, linePictureURL string) error {
	args := m.Called(holdingCode, userUID, lineUserID, lineDisplayName, linePictureURL)
	return args.Error(0)
}

func (m *ShopUserRepositoryMock) UpdateLastAccess(ctx context.Context, holdingCode string, userUID string, lastAccessedAt time.Time) error {
	args := m.Called(holdingCode, userUID, lastAccessedAt)
	return args.Error(0)
}

func (m *ShopUserRepositoryMock) SaveFavorite(ctx context.Context, holdingCode string, userUID string, isFavorite bool) error {
	args := m.Called(holdingCode, userUID, isFavorite)
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

func (m *ShopUserRepositoryMock) ResolveCompanyUID(ctx context.Context, holdingCode string, businessCode string) (string, error) {
	args := m.Called(holdingCode, businessCode)
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

func (m *AuthServiceMock) CreateSession(userInfo micromodels.UserInfo) (string, string, error) {
	args := m.Called(userInfo)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *AuthServiceMock) SelectShop(tokenType microservice.TokenType, tokenStr string, holdingCode string, businessCode string, branchUID string, role uint8) error {

	args := m.Called(tokenType, tokenStr, holdingCode, businessCode, branchUID, role)
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

func (m *AuthServiceMock) RefreshToken(token string) (string, string, bool, error) {
	args := m.Called(token)
	return args.String(0), args.String(1), args.Bool(2), args.Error(3)
}

func (m *AuthServiceMock) RevokeSession(tokenAuthorizationHeader string) error {
	args := m.Called(tokenAuthorizationHeader)
	return args.Error(0)
}

func (m *AuthServiceMock) RevokeUserTokens(username string) error {
	args := m.Called(username)
	return args.Error(0)
}

func (m *AuthServiceMock) RevokeUserTokensByUID(userUID string) error {
	args := m.Called(userUID)
	return args.Error(0)
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
