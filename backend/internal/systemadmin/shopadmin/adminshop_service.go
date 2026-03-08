package shopadmin

import (
	"context"
	shopModels "smlcloudplatform/internal/shop/models"
	"smlcloudplatform/pkg/microservice"
	"time"
)

type IShopAdminService interface {
	ListShop() ([]ShopDoc, error)
	CreateShop(doc shopModels.ShopDoc) error
	FindShopByProjectNo(projectNo string) (shopModels.ShopDoc, error)
	ListShopUsersAll() ([]ShopUserDoc, error)
	ListShopUsersByShopId(shopId string) ([]ShopUserDoc, error)
}

type ShopAdminService struct {
	repo            IShopAdminRepository
	timeoutDuration time.Duration
}

func NewShopAdminService(persister microservice.IPersisterMongo) IShopAdminService {
	repo := NewShopAdminRepository(persister)
	return &ShopAdminService{
		repo:            repo,
		timeoutDuration: time.Duration(30) * time.Second,
	}
}

func (s *ShopAdminService) ListShop() ([]ShopDoc, error) {

	ctx, cancel := context.WithTimeout(context.Background(), s.timeoutDuration)
	defer cancel()

	return s.repo.ListAllShop(ctx)
}

func (s *ShopAdminService) CreateShop(doc shopModels.ShopDoc) error {

	ctx, cancel := context.WithTimeout(context.Background(), s.timeoutDuration)
	defer cancel()

	return s.repo.CreateShop(ctx, doc)
}

func (s *ShopAdminService) FindShopByProjectNo(projectNo string) (shopModels.ShopDoc, error) {

	ctx, cancel := context.WithTimeout(context.Background(), s.timeoutDuration)
	defer cancel()

	return s.repo.FindShopByProjectNo(ctx, projectNo)
}

func (s *ShopAdminService) ListShopUsersAll() ([]ShopUserDoc, error) {

	ctx, cancel := context.WithTimeout(context.Background(), s.timeoutDuration)
	defer cancel()

	return s.repo.ListShopUsersAll(ctx)
}

func (s *ShopAdminService) ListShopUsersByShopId(shopId string) ([]ShopUserDoc, error) {

	ctx, cancel := context.WithTimeout(context.Background(), s.timeoutDuration)
	defer cancel()

	return s.repo.ListShopUsersByShopId(ctx, shopId)
}
