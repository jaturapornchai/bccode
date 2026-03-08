package datatransfer

import (
	"context"
	shopModule "smlcloudplatform/internal/shop"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ShopUserDataTransfer struct {
	transferConnection IDataTransferConnection
}

func NewShopUserDataTransfer(transferConnection IDataTransferConnection) IDataTransfer {
	return &ShopUserDataTransfer{
		transferConnection: transferConnection,
	}
}

func (sdt *ShopUserDataTransfer) StartTransfer(ctx context.Context, shopID string, targetShopID string) error {

	shopSourceRepository := shopModule.NewShopUserRepository(sdt.transferConnection.GetSourceConnection())

	shopUserList, err := shopSourceRepository.FindByShopID(ctx, shopID)
	if err != nil {
		return err
	}

	shopTargetRepository := shopModule.NewShopUserRepository(sdt.transferConnection.GetTargetConnection())

	for _, shopUser := range *shopUserList {

		if targetShopID != "" {
			shopUser.ID = primitive.NewObjectID()
			shopUser.ShopID = targetShopID
		}
		err := shopTargetRepository.Create(ctx, &shopUser)
		if err != nil {
			return err
		}
	}

	return nil

}
