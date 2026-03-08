package purchase

import (
	"context"
	purchaseModels "smlcloudplatform/internal/transaction/purchase/models"
	"smlcloudplatform/pkg/microservice"

	"go.mongodb.org/mongo-driver/bson"
)

type IPurchaseTransactionAdminRepositories interface {
	FindPurchaseDocByShopID(ctx context.Context, shopID string) ([]purchaseModels.PurchaseDoc, error)
	FindPurchaseDocDeleteByShopID(ctx context.Context, shopID string) ([]purchaseModels.PurchaseDoc, error)
}

type PurchaseTransactionAdminRepositories struct {
	pst microservice.IPersisterMongo
}

func NewPurchaseTransactionAdminRepositories(pst microservice.IPersisterMongo) IPurchaseTransactionAdminRepositories {
	return &PurchaseTransactionAdminRepositories{
		pst: pst,
	}
}

func (r PurchaseTransactionAdminRepositories) FindPurchaseDocByShopID(ctx context.Context, shopID string) ([]purchaseModels.PurchaseDoc, error) {

	docList := []purchaseModels.PurchaseDoc{}

	err := r.pst.Find(ctx, &purchaseModels.PurchaseDoc{},
		bson.M{
			"shopid":    shopID,
			"deletedat": bson.M{"$exists": false},
		},
		&docList)
	if err != nil {
		return nil, err
	}

	return docList, nil
}

func (r PurchaseTransactionAdminRepositories) FindPurchaseDocDeleteByShopID(ctx context.Context, shopID string) ([]purchaseModels.PurchaseDoc, error) {
	docList := []purchaseModels.PurchaseDoc{}

	err := r.pst.Find(ctx, &purchaseModels.PurchaseDoc{},
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
