package shop_test

import (
	"context"
	auth_model "smlcloudplatform/internal/authentication/models"
	common "smlcloudplatform/internal/models"
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

	shopRepo.On("Create", models.ShopDoc{

		ShopInfo: models.ShopInfo{
			DocIdentity: common.DocIdentity{
				GuidFixed: utilmock.MockGUID(),
			},
			Shop: models.Shop{
				Name1:     "shop_name",
				Telephone: "0000000000",
			},
		},
		ActivityDoc: common.ActivityDoc{
			CreatedBy: "user_create",
			CreatedAt: utilmock.MockTime(),
		},
	}).Return("", nil)

	shopUserRepo.On("Save", utilmock.MockGUID(), "user_create", auth_model.ROLE_OWNER).Return(nil)

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
					Name1:     "shop_name",
					Telephone: "0000000000",
				},
			},
			wantErr:  false,
			wantData: utilmock.MockGUID(),
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

func (m *ShopUserRepositoryMock) Update(ctx context.Context, id primitive.ObjectID, shopID string, username string, role auth_model.UserRole) error {
	args := m.Called(ctx, id, shopID, username, role)
	return args.Error(0)
}

func (m *ShopUserRepositoryMock) Save(ctx context.Context, shopID string, username string, role auth_model.UserRole) error {
	args := m.Called(ctx, shopID, username, role)
	return args.Error(0)
}

func (m *ShopUserRepositoryMock) UpdateLastAccess(ctx context.Context, shopID string, username string, lastAccessedAt time.Time) error {
	args := m.Called(ctx, shopID, username, lastAccessedAt)
	return args.Error(0)
}

func (m *ShopUserRepositoryMock) SaveFavorite(ctx context.Context, shopID string, username string, isFavorite bool) error {
	args := m.Called(ctx, shopID, username, isFavorite)
	return args.Error(0)
}

func (m *ShopUserRepositoryMock) Delete(ctx context.Context, shopID string, username string) error {
	args := m.Called(ctx, shopID, username)
	return args.Error(0)
}

func (m *ShopUserRepositoryMock) FindByShopIDAndUsernameInfo(ctx context.Context, shopID string, username string) (auth_model.ShopUserInfo, error) {
	args := m.Called(ctx, shopID, username)
	return args.Get(0).(auth_model.ShopUserInfo), args.Error(1)
}

func (m *ShopUserRepositoryMock) FindByShopIDAndUsername(ctx context.Context, shopID string, username string) (auth_model.ShopUser, error) {
	args := m.Called(ctx, shopID, username)
	return args.Get(0).(auth_model.ShopUser), args.Error(1)
}

func (m *ShopUserRepositoryMock) FindRole(ctx context.Context, shopID string, username string) (auth_model.UserRole, error) {
	args := m.Called(ctx, shopID, username)
	return args.Get(0).(auth_model.UserRole), args.Error(1)
}

func (m *ShopUserRepositoryMock) FindByShopID(ctx context.Context, shopID string) (*[]auth_model.ShopUser, error) {
	args := m.Called(ctx, shopID)
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

func (m *ShopUserRepositoryMock) FindByUserInShopPage(ctx context.Context, shopID string, pageable micromodels.Pageable) ([]auth_model.ShopUser, mongopagination.PaginationData, error) {
	args := m.Called(ctx, shopID, pageable)
	return args.Get(0).([]auth_model.ShopUser), args.Get(1).(mongopagination.PaginationData), args.Error(2)
}

func (m *ShopUserRepositoryMock) FindUserProfileByUsernames(ctx context.Context, usernames []string) ([]auth_model.UserProfile, error) {
	args := m.Called(ctx, usernames)
	return args.Get(0).([]auth_model.UserProfile), args.Error(1)
}
