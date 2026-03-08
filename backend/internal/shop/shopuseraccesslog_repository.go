package shop

import (
	"context"
	"smlcloudplatform/internal/authentication/models"
	"smlcloudplatform/pkg/microservice"
)

type IShopUserAccessLogRepository interface {
	Create(ctx context.Context, shopUserAccessLog models.ShopUserAccessLog) error
}

type ShopUserAccessLogRepository struct {
	pst microservice.IPersisterMongo
}

func NewShopUserAccessLogRepository(pst microservice.IPersisterMongo) ShopUserAccessLogRepository {
	return ShopUserAccessLogRepository{
		pst: pst,
	}
}

func (svc ShopUserAccessLogRepository) Create(ctx context.Context, shopUserAccessLog models.ShopUserAccessLog) error {

	_, err := svc.pst.Create(ctx, &models.ShopUserAccessLog{}, shopUserAccessLog)

	if err != nil {
		return err
	}

	return nil
}
