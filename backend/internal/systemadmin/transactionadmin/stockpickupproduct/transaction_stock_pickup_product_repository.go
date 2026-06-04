package stockpickupproduct

import (
	"context"
	stockPickupProductModels "smlcloudplatform/internal/transaction/stockpickupproduct/models"
	"smlcloudplatform/pkg/microservice"

	"go.mongodb.org/mongo-driver/bson"
)

type IStockPickupTransactionAdminRepository interface {
	FindStockPickupDocByHoldingCode(ctx context.Context, holdingCode string) ([]stockPickupProductModels.StockPickupProductDoc, error)
	FindStockPickupDocDeleteByHoldingCode(ctx context.Context, holdingCode string) ([]stockPickupProductModels.StockPickupProductDoc, error)
}

type StockPickupTransactionAdminRepository struct {
	pst microservice.IPersisterMongo
}

func NewStockPickupTransactionAdminRepository(pst microservice.IPersisterMongo) IStockPickupTransactionAdminRepository {
	return &StockPickupTransactionAdminRepository{
		pst: pst,
	}
}

func (r *StockPickupTransactionAdminRepository) FindStockPickupDocByHoldingCode(ctx context.Context, holdingCode string) ([]stockPickupProductModels.StockPickupProductDoc, error) {
	docList := []stockPickupProductModels.StockPickupProductDoc{}

	err := r.pst.Find(ctx, &stockPickupProductModels.StockPickupProductDoc{},
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

func (r *StockPickupTransactionAdminRepository) FindStockPickupDocDeleteByHoldingCode(ctx context.Context, holdingCode string) ([]stockPickupProductModels.StockPickupProductDoc, error) {
	docList := []stockPickupProductModels.StockPickupProductDoc{}

	err := r.pst.Find(ctx, &stockPickupProductModels.StockPickupProductDoc{},
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
