package purchasereturn

import (
	"context"
	purchaseReturnModel "smlcloudplatform/internal/transaction/purchasereturn/models"
	"smlcloudplatform/pkg/microservice"

	"go.mongodb.org/mongo-driver/bson"
)

type IPurchaseReturnTransactionAdminRepository interface {
	FindPurchaseReturnDocByShopID(ctx context.Context, shopID string) ([]purchaseReturnModel.PurchaseReturnDoc, error)
	FindPurchaseReturnDeleteDocByShopID(ctx context.Context, shopID string) ([]purchaseReturnModel.PurchaseReturnDoc, error)
}

type PurchaseReturnTransactionAdminRepository struct {
	pst microservice.IPersisterMongo
}

func NewPurchaseReturnTransactionAdminRepository(pst microservice.IPersisterMongo) IPurchaseReturnTransactionAdminRepository {
	return &PurchaseReturnTransactionAdminRepository{
		pst: pst,
	}
}

func (r *PurchaseReturnTransactionAdminRepository) FindPurchaseReturnDocByShopID(ctx context.Context, shopID string) ([]purchaseReturnModel.PurchaseReturnDoc, error) {
	docList := []purchaseReturnModel.PurchaseReturnDoc{}

	err := r.pst.Find(ctx, &purchaseReturnModel.PurchaseReturnDoc{},
		bson.M{
			"shopid":    shopID,
			"deletedat": bson.M{"$exists": false}},
		&docList)
	if err != nil {
		return nil, err
	}

	return docList, nil
}

func (r *PurchaseReturnTransactionAdminRepository) FindPurchaseReturnDeleteDocByShopID(ctx context.Context, shopID string) ([]purchaseReturnModel.PurchaseReturnDoc, error) {
	docList := []purchaseReturnModel.PurchaseReturnDoc{}

	err := r.pst.Find(ctx, &purchaseReturnModel.PurchaseReturnDoc{},
		bson.M{
			"shopid":    shopID,
			"deletedat": bson.M{"$exists": true},
		},
		&docList)
	if err != nil {
		return nil, err
	}

	return docList, nil
}
