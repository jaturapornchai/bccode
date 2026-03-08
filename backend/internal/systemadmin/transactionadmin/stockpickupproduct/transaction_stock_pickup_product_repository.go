package stockpickupproduct

import (
	"context"
	stockPickupProductModels "smlcloudplatform/internal/transaction/stockpickupproduct/models"
	"smlcloudplatform/pkg/microservice"

	"go.mongodb.org/mongo-driver/bson"
)

type IStockPickupTransactionAdminRepository interface {
	FindStockPickupDocByShopID(ctx context.Context, shopID string) ([]stockPickupProductModels.StockPickupProductDoc, error)
	FindStockPickupDocDeleteByShopID(ctx context.Context, shopID string) ([]stockPickupProductModels.StockPickupProductDoc, error)
}

type StockPickupTransactionAdminRepository struct {
	pst microservice.IPersisterMongo
}

func NewStockPickupTransactionAdminRepository(pst microservice.IPersisterMongo) IStockPickupTransactionAdminRepository {
	return &StockPickupTransactionAdminRepository{
		pst: pst,
	}
}

func (r *StockPickupTransactionAdminRepository) FindStockPickupDocByShopID(ctx context.Context, shopID string) ([]stockPickupProductModels.StockPickupProductDoc, error) {
	docList := []stockPickupProductModels.StockPickupProductDoc{}

	err := r.pst.Find(ctx, &stockPickupProductModels.StockPickupProductDoc{},
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

func (r *StockPickupTransactionAdminRepository) FindStockPickupDocDeleteByShopID(ctx context.Context, shopID string) ([]stockPickupProductModels.StockPickupProductDoc, error) {
	docList := []stockPickupProductModels.StockPickupProductDoc{}

	err := r.pst.Find(ctx, &stockPickupProductModels.StockPickupProductDoc{},
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
