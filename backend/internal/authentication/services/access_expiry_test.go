package services_test

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"smlcloudplatform/internal/authentication/models"
	authmodels "smlcloudplatform/internal/authentication/services"
	"smlcloudplatform/pkg/apperr"
	"smlcloudplatform/pkg/microservice"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func accessExpiryService(t *testing.T, expiry time.Time) (*AuthServiceMock, func() error) {
	t.Helper()
	shopUserRepo := new(ShopUserRepositoryMock)
	accessLogRepo := new(ShopUserAccessLogRepositoryMock)
	microAuth := &AuthServiceMock{}
	member := models.ShopUser{ID: "membership-1", AccessExpiryDate: expiry}
	member.Username = "somchai01"
	member.UserUID = "user-uid-1"
	member.HoldingCode = "rungrueng"
	shopUserRepo.On("FindByHoldingCodeAndUserUID", "rungrueng", "user-uid-1").Return(member, nil)
	shopUserRepo.On("UpdateLastAccess", "rungrueng", "user-uid-1", MockTime()).Return(nil)
	accessLogRepo.On("Create", mock.Anything).Return(nil)
	microAuth.On("GetTokenFromAuthorizationHeader", microservice.AUTHTYPE_BEARER, "Bearer token").Return("token", nil)
	microAuth.On("SelectShop", microservice.AUTHTYPE_BEARER, "token", "rungrueng", "", "", uint8(0)).Return(nil)
	svc := authmodels.NewAuthenticationService(
		new(AuthenticationRepositoryMock), shopUserRepo, accessLogRepo,
		new(SMSRepositoryMock), microAuth, MockRandomString, MockRandomNumber,
		MockGUID, MockHashPassword, MockCheckPasswordHash, MockTime, MockFirebaseAdapter(), MockLineAdapter())
	return microAuth, func() error {
		return svc.AccessShop("rungrueng", "", "", "somchai01", "user-uid-1", "Bearer token", models.AuthenticationContext{Ip: "127.0.0.1"})
	}
}

// ShopUser.AccessExpiryDate is the instant access ends: 00:00 of the day AFTER the expiry date
// (usable through its end, Holding timezone — models.AccessEndsAt). Selecting the Holding at that
// instant is refused with a 403 that names the reason (save-audit finding 4/5).
func TestAccessShopRefusesMembershipWhenAccessEnds(t *testing.T) {
	microAuth, access := accessExpiryService(t, MockTime())

	err := access()

	require.Error(t, err)
	require.True(t, errors.Is(err, authmodels.ErrShopAccessExpired))
	require.Equal(t, "user_access_expired", authmodels.ShopAccessDeniedKey(err))
	appErr := apperr.FromError(err)
	require.NotNil(t, appErr)
	require.Equal(t, http.StatusForbidden, appErr.StatusCode())
	microAuth.AssertNotCalled(t, "SelectShop", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestAccessShopAllowsMembershipBeforeAccessEnds(t *testing.T) {
	microAuth, access := accessExpiryService(t, MockTime().Add(time.Second))

	require.NoError(t, access())
	microAuth.AssertCalled(t, "SelectShop", microservice.AUTHTYPE_BEARER, "token", "rungrueng", "", "", uint8(0))
}

func TestShopAccessDeniedIsLocalizedAndKeepsTheReason(t *testing.T) {
	err := authmodels.ShopAccessDenied(authmodels.ErrShopAccessDisabled, "en")
	require.Equal(t, http.StatusForbidden, err.StatusCode())
	require.True(t, errors.Is(err, authmodels.ErrShopAccessDisabled))
	require.Equal(t, "user_access_disabled", authmodels.ShopAccessDeniedKey(err))
	require.Equal(t, "", authmodels.ShopAccessDeniedKey(errors.New("holdingcode invalid")))
}

// A first Google login may attach to an admin-created account by email, so the email must be
// one Google verified; an unverified one is refused before any lookup or account creation.
func TestLoginWithGoogleIdentityRequiresVerifiedEmail(t *testing.T) {
	authRepo := new(AuthenticationRepositoryMock)
	svc := authmodels.NewAuthenticationService(
		authRepo, new(ShopUserRepositoryMock), new(ShopUserAccessLogRepositoryMock),
		new(SMSRepositoryMock), &AuthServiceMock{}, MockRandomString, MockRandomNumber,
		MockGUID, MockHashPassword, MockCheckPasswordHash, MockTime, MockFirebaseAdapter(), MockLineAdapter())

	_, err := svc.LoginWithGoogleIdentity("accounts.google.com", "subject-1", "somchai@company.co.th", false, "Somchai")

	require.EqualError(t, err, "google email not verified")
	authRepo.AssertNotCalled(t, "FindGoogleIdentity", mock.Anything, mock.Anything)
	authRepo.AssertNotCalled(t, "CreateGoogleUserIdentity", mock.Anything, mock.Anything, mock.Anything)
}
