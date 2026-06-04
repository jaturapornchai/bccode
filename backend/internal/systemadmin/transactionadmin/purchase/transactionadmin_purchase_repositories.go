package purchase

import (
	"context"
	purchaseModels "smlcloudplatform/internal/transaction/purchase/models"
	"smlcloudplatform/pkg/microservice"

	"go.mongodb.org/mongo-driver/bson"
)

type IPurchaseTransactionAdminRepositories interface {
	FindPurchaseDocByHoldingCode(ctx context.Context, holdingCode string) ([]purchaseModels.PurchaseDoc, error)
	FindPurchaseDocDeleteByHoldingCode(ctx context.Context, holdingCode string) ([]purchaseModels.PurchaseDoc, error)
}

type PurchaseTransactionAdminRepositories struct {
	pst microservice.IPersisterMongo
}

func NewPurchaseTransactionAdminRepositories(pst microservice.IPersisterMongo) IPurchaseTransactionAdminRepositories {
	return &PurchaseTransactionAdminRepositories{
		pst: pst,
	}
}

func (r PurchaseTransactionAdminRepositories) FindPurchaseDocByHoldingCode(ctx context.Context, holdingCode string) ([]purchaseModels.PurchaseDoc, error) {

	docList := []purchaseModels.PurchaseDoc{}

	err := r.pst.Find(ctx, &purchaseModels.PurchaseDoc{},
		bson.M{
			"holding_code": holdingCode,
			"deleted_at":   bson.M{"$exists": false},
		},
		&docList)
	if err != nil {
		return nil, err
	}

	return docList, nil
}

func (r PurchaseTransactionAdminRepositories) FindPurchaseDocDeleteByHoldingCode(ctx context.Context, holdingCode string) ([]purchaseModels.PurchaseDoc, error) {
	docList := []purchaseModels.PurchaseDoc{}

	err := r.pst.Find(ctx, &purchaseModels.PurchaseDoc{},
		bson.M{
			"holding_code": holdingCode,
			"deleted_at":   bson.M{"$exists": true},
		},
		&docList)
	if err != nil {
		return nil, err
	}

	return docList, nil
}
