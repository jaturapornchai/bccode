package shop_test

import (
	"errors"
	"net/http"
	"testing"

	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/shop"
	"smlcloudplatform/internal/shop/models"
	utilmock "smlcloudplatform/mock"
	"smlcloudplatform/pkg/apperr"

	"github.com/stretchr/testify/mock"
	"github.com/tj/assert"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestShopCreateReturnsConflictForExistingHoldingCode(t *testing.T) {
	shopRepo := new(ShopRepositoryMock)
	shopUserRepo := new(ShopUserRepositoryMock)
	shopRepo.On("FindByHoldingCode", mock.Anything, "duplicate").Return(models.ShopDoc{
		ShopInfo: models.ShopInfo{DocIdentity: common.DocIdentity{GuidFixed: "holding-uid"}},
	}, nil)

	service := shop.NewShopService(shopRepo, shopUserRepo, utilmock.MockGUID, utilmock.MockTime)
	_, err := service.CreateShop("user-uid", "", models.Shop{
		HoldingCode: "duplicate",
		Names:       validHoldingNames(),
	})

	assertDuplicateHoldingError(t, err)
	shopRepo.AssertNotCalled(t, "EnsureBootstrapIndexes", mock.Anything)
}

func TestShopCreateReturnsConflictForAtomicCodeClaimDuplicate(t *testing.T) {
	shopRepo := new(ShopRepositoryMock)
	shopUserRepo := new(ShopUserRepositoryMock)
	shopRepo.On("FindByHoldingCode", mock.Anything, "duplicate").Return(models.ShopDoc{}, errors.New("not found"))
	shopRepo.On("EnsureBootstrapIndexes", mock.Anything).Return(nil)
	shopRepo.On("Transaction", mock.Anything).Return(mongo.WriteException{
		WriteErrors: mongo.WriteErrors{{Code: 11000, Message: "duplicate key"}},
	})

	service := shop.NewShopService(shopRepo, shopUserRepo, utilmock.MockGUID, utilmock.MockTime)
	_, err := service.CreateShop("user-uid", "", models.Shop{
		HoldingCode: "duplicate",
		Names:       validHoldingNames(),
	})

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
