package shop_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/stretchr/testify/mock"
	"github.com/tj/assert"

	auth_model "smlcloudplatform/internal/authentication/models"
	common "smlcloudplatform/internal/models"
	orgaccess "smlcloudplatform/internal/organization"
	"smlcloudplatform/internal/shop"
	"smlcloudplatform/internal/shop/models"
	"smlcloudplatform/pkg/apperr"
	micromodels "smlcloudplatform/pkg/microservice/models"
)

func mockTime() time.Time {
	return time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
}

func runTransaction(t *testing.T, shopRepo *ShopRepositoryMock, result error) {
	shopRepo.On("Transaction", mock.Anything, mock.Anything).Return(result).Run(func(args mock.Arguments) {
		if result != nil {
			return
		}
		if err := args.Get(1).(func(context.Context) error)(context.Background()); err != nil {
			t.Fatalf("transaction callback: %v", err)
		}
	})
}

func TestShop_Create(t *testing.T) {
	shopRepo := new(ShopRepositoryMock)
	shopUserRepo := new(ShopUserRepositoryMock)

	shopRepo.On("FindByHoldingCode", mock.Anything, "shoptest").Return(models.ShopDoc{}, errors.New("not found"))
	runTransaction(t, shopRepo, nil)
	shopRepo.On("Create", mock.Anything, mock.MatchedBy(func(doc models.ShopDoc) bool {
		return doc.HoldingCode == "shoptest" && doc.HoldingUID == "shoptest" && doc.GuidFixed == "shoptest" &&
			doc.IsActive && doc.CreatedBy == "USERUID001" && doc.CreatedAt.Equal(mockTime())
	})).Return(nil)
	shopRepo.On("RecordAudit", mock.Anything, mock.MatchedBy(func(audit orgaccess.Audit) bool {
		return audit.Action == "holding.created" && audit.HoldingCode == "shoptest" && audit.ActorUID == "USERUID001"
	})).Return(nil)
	shopUserRepo.On("SaveStable", mock.Anything, "shoptest", "USERUID001", auth_model.ROLE_OWNER).Return(nil)

	cases := []struct {
		name    string
		shop    models.Shop
		wantErr bool
	}{
		{name: "success create shop", shop: models.Shop{HoldingCode: "shoptest", Name1: "shop_name", Names: validHoldingNames(), Telephone: "0000000000"}},
		{name: "reject missing holdingcode", shop: models.Shop{Name1: "shop_name", Names: validHoldingNames()}, wantErr: true},
		{name: "reject invalid character holdingcode", shop: models.Shop{HoldingCode: "shop@test", Names: validHoldingNames()}, wantErr: true},
		{name: "reject missing holding name", shop: models.Shop{HoldingCode: "shoptest"}, wantErr: true},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			shopSvc := shop.NewShopService(shopRepo, shopUserRepo, mockTime)
			holdingUID, err := shopSvc.CreateShop("USERUID001", "user_create", tt.shop)
			if tt.wantErr {
				assert.NotNil(t, err)
				assert.Empty(t, holdingUID)
				return
			}
			assert.Nil(t, err)
			assert.EqualValues(t, "shoptest", holdingUID)
		})
	}
	shopUserRepo.AssertCalled(t, "SaveStable", mock.Anything, "shoptest", "USERUID001", auth_model.ROLE_OWNER)
}

func TestShopCreateReturnsConflictForExistingHoldingCode(t *testing.T) {
	shopRepo := new(ShopRepositoryMock)
	shopRepo.On("FindByHoldingCode", mock.Anything, "duplicate").Return(models.ShopDoc{}, nil)

	_, err := shop.NewShopService(shopRepo, new(ShopUserRepositoryMock), mockTime).
		CreateShop("user-uid", "", models.Shop{HoldingCode: "duplicate", Names: validHoldingNames()})

	assertDuplicateHoldingError(t, err)
	shopRepo.AssertNotCalled(t, "Transaction", mock.Anything, mock.Anything)
}

func TestShopCreateReturnsConflictForUniqueViolation(t *testing.T) {
	shopRepo := new(ShopRepositoryMock)
	shopRepo.On("FindByHoldingCode", mock.Anything, "duplicate").Return(models.ShopDoc{}, errors.New("not found"))
	runTransaction(t, shopRepo, &pq.Error{Code: "23505", Message: "duplicate key"})

	_, err := shop.NewShopService(shopRepo, new(ShopUserRepositoryMock), mockTime).
		CreateShop("user-uid", "", models.Shop{HoldingCode: "duplicate", Names: validHoldingNames()})

	assertDuplicateHoldingError(t, err)
}

func assertDuplicateHoldingError(t *testing.T, err error) {
	t.Helper()
	var appError *apperr.AppError
	if !errors.As(err, &appError) {
		t.Fatalf("CreateShop error = %T %v, want *apperr.AppError", err, err)
	}
	assert.Equal(t, "DUPLICATE", appError.Code)
	assert.Equal(t, http.StatusConflict, appError.StatusCode())
	assert.Equal(t, "holdingcode", appError.Field)
}

func validHoldingNames() []common.NameX {
	code, name := "th", "กลุ่มกิจการรุ่งเรืองกรุ๊ป"
	return []common.NameX{{Code: &code, Name: &name}}
}

func TestShopUpdateRejectsHoldingCodeChangeUntilAtomicRenameExists(t *testing.T) {
	shopRepo := new(ShopRepositoryMock)
	shopRepo.On("FindByHoldingCode", mock.Anything, "oldcode").Return(models.ShopDoc{
		ShopInfo: models.ShopInfo{
			DocIdentity: common.DocIdentity{GuidFixed: "oldcode"},
			Shop:        models.Shop{HoldingCode: "oldcode", IsActive: true},
		},
	}, nil)

	service := shop.NewShopService(shopRepo, new(ShopUserRepositoryMock), mockTime)
	err := service.UpdateShop("oldcode", "actor", models.Shop{HoldingCode: "newcode"})
	if !errors.Is(err, shop.ErrHoldingCodeRenameUnavailable) {
		t.Fatalf("UpdateShop error = %v", err)
	}
	shopRepo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

type ShopRepositoryMock struct {
	mock.Mock
}

func (m *ShopRepositoryMock) Transaction(ctx context.Context, fn func(context.Context) error) error {
	return m.Called(ctx, fn).Error(0)
}

func (m *ShopRepositoryMock) Create(ctx context.Context, doc models.ShopDoc) error {
	return m.Called(ctx, doc).Error(0)
}

func (m *ShopRepositoryMock) Update(ctx context.Context, holdingCode string, expectedActive bool, doc models.ShopDoc) error {
	return m.Called(ctx, holdingCode, expectedActive, doc).Error(0)
}

func (m *ShopRepositoryMock) UpdateStatus(ctx context.Context, holdingCode string, expectedActive bool, active bool, username string, now time.Time) error {
	return m.Called(ctx, holdingCode, expectedActive, active, username, now).Error(0)
}

func (m *ShopRepositoryMock) FindByHoldingCode(ctx context.Context, holdingCode string) (models.ShopDoc, error) {
	args := m.Called(ctx, holdingCode)
	return args.Get(0).(models.ShopDoc), args.Error(1)
}

func (m *ShopRepositoryMock) RecordAudit(ctx context.Context, audit orgaccess.Audit) error {
	return m.Called(ctx, audit).Error(0)
}

type ShopUserRepositoryMock struct {
	mock.Mock
}

func (m *ShopUserRepositoryMock) FindByHoldingCodeAndUsername(ctx context.Context, holdingCode string, username string) (auth_model.ShopUser, error) {
	args := m.Called(ctx, holdingCode, username)
	return args.Get(0).(auth_model.ShopUser), args.Error(1)
}

func (m *ShopUserRepositoryMock) FindByHoldingCodeAndUserUID(ctx context.Context, holdingCode string, userUID string) (auth_model.ShopUser, error) {
	args := m.Called(ctx, holdingCode, userUID)
	return args.Get(0).(auth_model.ShopUser), args.Error(1)
}

func (m *ShopUserRepositoryMock) FindByUserUIDPage(ctx context.Context, userUID string, pageable micromodels.Pageable) ([]auth_model.ShopUserInfo, common.PaginationData, error) {
	args := m.Called(ctx, userUID, pageable)
	return args.Get(0).([]auth_model.ShopUserInfo), args.Get(1).(common.PaginationData), args.Error(2)
}

func (m *ShopUserRepositoryMock) FindByUsernamePage(ctx context.Context, username string, pageable micromodels.Pageable) ([]auth_model.ShopUserInfo, common.PaginationData, error) {
	args := m.Called(ctx, username, pageable)
	return args.Get(0).([]auth_model.ShopUserInfo), args.Get(1).(common.PaginationData), args.Error(2)
}

func (m *ShopUserRepositoryMock) FindByUserInShopPageWithProfileMatches(ctx context.Context, holdingCode string, pageable micromodels.Pageable, profileUsernames []string) ([]auth_model.ShopUser, common.PaginationData, error) {
	args := m.Called(ctx, holdingCode, pageable, profileUsernames)
	return args.Get(0).([]auth_model.ShopUser), args.Get(1).(common.PaginationData), args.Error(2)
}

func (m *ShopUserRepositoryMock) FindUsernamesByProfileQuery(ctx context.Context, query string) ([]string, error) {
	args := m.Called(ctx, query)
	return args.Get(0).([]string), args.Error(1)
}

func (m *ShopUserRepositoryMock) FindUserProfileByUsernames(ctx context.Context, usernames []string) ([]auth_model.UserProfile, error) {
	args := m.Called(ctx, usernames)
	return args.Get(0).([]auth_model.UserProfile), args.Error(1)
}

func (m *ShopUserRepositoryMock) FindShopCreatedBy(ctx context.Context, holdingCode string) (string, error) {
	args := m.Called(ctx, holdingCode)
	return args.String(0), args.Error(1)
}

func (m *ShopUserRepositoryMock) SaveFullProfile(ctx context.Context, holdingCode string, req *auth_model.UserRoleRequest) error {
	return m.Called(ctx, holdingCode, req).Error(0)
}

func (m *ShopUserRepositoryMock) SaveStable(ctx context.Context, holdingCode string, userUID string, role auth_model.UserRole) error {
	return m.Called(ctx, holdingCode, userUID, role).Error(0)
}

func (m *ShopUserRepositoryMock) Delete(ctx context.Context, holdingCode string, username string, actorUID string) error {
	return m.Called(ctx, holdingCode, username, actorUID).Error(0)
}

func (m *ShopUserRepositoryMock) UpdateLastAccess(ctx context.Context, holdingCode string, userUID string, lastAccessedAt time.Time) error {
	return m.Called(ctx, holdingCode, userUID, lastAccessedAt).Error(0)
}

func (m *ShopUserRepositoryMock) SaveFavorite(ctx context.Context, holdingCode string, userUID string, isFavorite bool) error {
	return m.Called(ctx, holdingCode, userUID, isFavorite).Error(0)
}

func (m *ShopUserRepositoryMock) ResolveHoldingCodeByHoldingCode(ctx context.Context, holdingCode string) (string, error) {
	args := m.Called(ctx, holdingCode)
	return args.String(0), args.Error(1)
}

func (m *ShopUserRepositoryMock) ResolveCompanyUID(ctx context.Context, holdingCode string, businessCode string) (string, error) {
	args := m.Called(ctx, holdingCode, businessCode)
	return args.String(0), args.Error(1)
}
