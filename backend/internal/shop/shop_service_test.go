package shop_test

import (
	"context"
	"errors"
	auth_model "smlcloudplatform/internal/authentication/models"
	"smlcloudplatform/internal/shop"
	"smlcloudplatform/internal/shop/models"
	utilmock "smlcloudplatform/mock"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"testing"
	"time"

	"github.com/smlsoft/mongopagination"
	"github.com/stretchr/testify/mock"
	"github.com/tj/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestShop_Create(t *testing.T) {
	shopRepo := new(ShopRepositoryMock)
	shopUserRepo := new(ShopUserRepositoryMock)

	shopRepo.On("FindByHoldingCode", mock.Anything, "shoptest").Return(models.ShopDoc{}, errors.New("not found"))
	shopRepo.On("Create", mock.Anything, mock.MatchedBy(func(doc models.ShopDoc) bool {
		return doc.HoldingCode == "shoptest" && doc.GuidFixed == "shoptest"
	})).Return("", nil)
	shopUserRepo.On("Save", mock.Anything, "shoptest", "user_create", auth_model.ROLE_OWNER).Return(nil)

	type args struct {
		username string
		shop     models.Shop
	}

	cases := []struct {
		name     string
		args     args
		wantErr  bool
		wantData string
	}{
		{
			name: "success create shop",
			args: args{
				username: "user_create",
				shop: models.Shop{
					HoldingCode: "shoptest",
					Name1:       "shop_name",
					Telephone:   "0000000000",
				},
			},
			wantErr:  false,
			wantData: "shoptest",
		},
		{
			name: "reject missing holdingcode",
			args: args{
				username: "user_create",
				shop: models.Shop{
					Name1:     "shop_name",
					Telephone: "0000000000",
				},
			},
			wantErr: true,
		},
		{
			name: "reject underscore holdingcode",
			args: args{
				username: "user_create",
				shop: models.Shop{
					HoldingCode: "shop_test",
					Name1:       "shop_name",
					Telephone:   "0000000000",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			shopSvc := shop.NewShopService(shopRepo, shopUserRepo, utilmock.MockGUID, utilmock.MockTime)

			shopGUID, err := shopSvc.CreateShop(tt.args.username, tt.args.shop)

			if tt.wantErr {
				assert.NotNil(t, err)
				assert.Empty(t, shopGUID)
			} else {
				assert.Nil(t, err)
				assert.NotEmpty(t, shopGUID)
				assert.EqualValues(t, tt.wantData, shopGUID)
			}
		})
	}

}

type ShopRepositoryMock struct {
	mock.Mock
}

func (m *ShopRepositoryMock) Create(ctx context.Context, shop models.ShopDoc) (string, error) {
	args := m.Called(ctx, shop)

	return args.String(0), args.Error(1)
}

func (m *ShopRepositoryMock) Update(ctx context.Context, guid string, shop models.ShopDoc) error {
	args := m.Called(ctx, guid, shop)

	return args.Error(0)
}
func (m *ShopRepositoryMock) FindByGuid(ctx context.Context, guid string) (models.ShopDoc, error) {
	args := m.Called(ctx, guid)
	return args.Get(0).(models.ShopDoc), args.Error(0)
}
func (m *ShopRepositoryMock) FindByHoldingCode(ctx context.Context, holdingCode string) (models.ShopDoc, error) {
	args := m.Called(ctx, holdingCode)
	return args.Get(0).(models.ShopDoc), args.Error(1)
}
func (m *ShopRepositoryMock) FindPage(ctx context.Context, pageable micromodels.Pageable) ([]models.ShopInfo, mongopagination.PaginationData, error) {
	args := m.Called(ctx, pageable)

	return args.Get(0).([]models.ShopInfo), args.Get(1).(mongopagination.PaginationData), args.Error(2)
}
func (m *ShopRepositoryMock) Delete(ctx context.Context, guid string, username string) error {
	args := m.Called(ctx, guid, username)

	return args.Error(0)
}

type ShopUserRepositoryMock struct {
	mock.Mock
}

func (m *ShopUserRepositoryMock) Create(ctx context.Context, shopUser *auth_model.ShopUser) error {
	args := m.Called(ctx, shopUser)
	return args.Error(0)
}

func (m *ShopUserRepositoryMock) Update(ctx context.Context, id primitive.ObjectID, holdingCode string, username string, role auth_model.UserRole) error {
	args := m.Called(ctx, id, holdingCode, username, role)
	return args.Error(0)
}

func (m *ShopUserRepositoryMock) Save(ctx context.Context, holdingCode string, username string, role auth_model.UserRole) error {
	args := m.Called(ctx, holdingCode, username, role)
	return args.Error(0)
}

func (m *ShopUserRepositoryMock) SaveFullProfile(ctx context.Context, holdingCode string, req *auth_model.UserRoleRequest) error {
	args := m.Called(ctx, holdingCode, req)
	return args.Error(0)
}

func (m *ShopUserRepositoryMock) UpdateLastAccess(ctx context.Context, holdingCode string, username string, lastAccessedAt time.Time) error {
	args := m.Called(ctx, holdingCode, username, lastAccessedAt)
	return args.Error(0)
}

func (m *ShopUserRepositoryMock) SaveFavorite(ctx context.Context, holdingCode string, username string, isFavorite bool) error {
	args := m.Called(ctx, holdingCode, username, isFavorite)
	return args.Error(0)
}

func (m *ShopUserRepositoryMock) Delete(ctx context.Context, holdingCode string, username string) error {
	args := m.Called(ctx, holdingCode, username)
	return args.Error(0)
}

func (m *ShopUserRepositoryMock) DeleteEmptyUsernames(ctx context.Context, holdingCode string) (int64, error) {
	args := m.Called(ctx, holdingCode)
	return args.Get(0).(int64), args.Error(1)
}

func (m *ShopUserRepositoryMock) FindByHoldingCodeAndUsernameInfo(ctx context.Context, holdingCode string, username string) (auth_model.ShopUserInfo, error) {
	args := m.Called(ctx, holdingCode, username)
	return args.Get(0).(auth_model.ShopUserInfo), args.Error(1)
}

func (m *ShopUserRepositoryMock) FindByHoldingCodeAndUserUIDInfo(ctx context.Context, holdingCode string, userUID string) (auth_model.ShopUserInfo, error) {
	args := m.Called(ctx, holdingCode, userUID)
	return args.Get(0).(auth_model.ShopUserInfo), args.Error(1)
}

func (m *ShopUserRepositoryMock) FindByHoldingCodeAndUserUID(ctx context.Context, holdingCode string, userUID string) (auth_model.ShopUser, error) {
	args := m.Called(ctx, holdingCode, userUID)
	return args.Get(0).(auth_model.ShopUser), args.Error(1)
}

func (m *ShopUserRepositoryMock) ResolveHoldingCodeByHoldingCode(ctx context.Context, holdingCode string) (string, error) {
	args := m.Called(ctx, holdingCode)
	return args.String(0), args.Error(1)
}

func (m *ShopUserRepositoryMock) FindByHoldingCodeAndUsername(ctx context.Context, holdingCode string, username string) (auth_model.ShopUser, error) {
	args := m.Called(ctx, holdingCode, username)
	return args.Get(0).(auth_model.ShopUser), args.Error(1)
}

func (m *ShopUserRepositoryMock) FindByHoldingCodeAndLineUserID(ctx context.Context, holdingCode string, lineUserID string) (auth_model.ShopUser, error) {
	args := m.Called(ctx, holdingCode, lineUserID)
	return args.Get(0).(auth_model.ShopUser), args.Error(1)
}

func (m *ShopUserRepositoryMock) FindByLineUserID(ctx context.Context, lineUserID string) (auth_model.ShopUser, error) {
	args := m.Called(ctx, lineUserID)
	return args.Get(0).(auth_model.ShopUser), args.Error(1)
}

func (m *ShopUserRepositoryMock) FindShopCreatedBy(ctx context.Context, holdingCode string) (string, error) {
	args := m.Called(ctx, holdingCode)
	return args.String(0), args.Error(1)
}

func (m *ShopUserRepositoryMock) FindRole(ctx context.Context, holdingCode string, username string) (auth_model.UserRole, error) {
	args := m.Called(ctx, holdingCode, username)
	return args.Get(0).(auth_model.UserRole), args.Error(1)
}

func (m *ShopUserRepositoryMock) FindByHoldingCode(ctx context.Context, holdingCode string) (*[]auth_model.ShopUser, error) {
	args := m.Called(ctx, holdingCode)
	return args.Get(0).(*[]auth_model.ShopUser), args.Error(1)
}
func (m *ShopUserRepositoryMock) FindByUsername(ctx context.Context, username string) (*[]auth_model.ShopUser, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(*[]auth_model.ShopUser), args.Error(1)
}

func (m *ShopUserRepositoryMock) FindByUsernamePage(ctx context.Context, username string, pageable micromodels.Pageable) ([]auth_model.ShopUserInfo, mongopagination.PaginationData, error) {
	args := m.Called(ctx, username, pageable)
	return args.Get(0).([]auth_model.ShopUserInfo), args.Get(1).(mongopagination.PaginationData), args.Error(2)
}

func (m *ShopUserRepositoryMock) FindByUserUIDPage(ctx context.Context, userUID string, pageable micromodels.Pageable) ([]auth_model.ShopUserInfo, mongopagination.PaginationData, error) {
	args := m.Called(ctx, userUID, pageable)
	return args.Get(0).([]auth_model.ShopUserInfo), args.Get(1).(mongopagination.PaginationData), args.Error(2)
}

func (m *ShopUserRepositoryMock) FindByUserInShopPage(ctx context.Context, holdingCode string, pageable micromodels.Pageable) ([]auth_model.ShopUser, mongopagination.PaginationData, error) {
	args := m.Called(ctx, holdingCode, pageable)
	return args.Get(0).([]auth_model.ShopUser), args.Get(1).(mongopagination.PaginationData), args.Error(2)
}

func (m *ShopUserRepositoryMock) FindByUserInShopPageWithProfileMatches(ctx context.Context, holdingCode string, pageable micromodels.Pageable, profileUsernames []string) ([]auth_model.ShopUser, mongopagination.PaginationData, error) {
	args := m.Called(ctx, holdingCode, pageable, profileUsernames)
	return args.Get(0).([]auth_model.ShopUser), args.Get(1).(mongopagination.PaginationData), args.Error(2)
}

func (m *ShopUserRepositoryMock) FindUsernamesByProfileQuery(ctx context.Context, query string) ([]string, error) {
	args := m.Called(ctx, query)
	return args.Get(0).([]string), args.Error(1)
}

func (m *ShopUserRepositoryMock) FindUserProfileByUsernames(ctx context.Context, usernames []string) ([]auth_model.UserProfile, error) {
	args := m.Called(ctx, usernames)
	return args.Get(0).([]auth_model.UserProfile), args.Error(1)
}
