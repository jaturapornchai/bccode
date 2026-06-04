package stockbalance

import (
	"context"
	stockBalanceProductModels "smlcloudplatform/internal/transaction/stockbalance/models"
	"smlcloudplatform/pkg/microservice"

	"go.mongodb.org/mongo-driver/bson"
)

type IStockBalanceTransactionAdminRepository interface {
	FindStockBalanceDocByHoldingCode(ctx context.Context, holdingCode string) ([]stockBalanceProductModels.StockBalanceDoc, error)
	FindStockBalanceDocDeleteByHoldingCode(ctx context.Context, holdingCode string) ([]stockBalanceProductModels.StockBalanceDoc, error)
}

type StockBalanceTransactionAdminRepository struct {
	pst microservice.IPersisterMongo
}

func NewStockBalanceTransactionAdminRepository(pst microservice.IPersisterMongo) IStockBalanceTransactionAdminRepository {
	return &StockBalanceTransactionAdminRepository{
		pst: pst,
	}
}

func (r *StockBalanceTransactionAdminRepository) FindStockBalanceDocByHoldingCode(ctx context.Context, holdingCode string) ([]stockBalanceProductModels.StockBalanceDoc, error) {
	docList := []stockBalanceProductModels.StockBalanceDoc{}

	err := r.pst.Find(ctx, &stockBalanceProductModels.StockBalanceDoc{},
		bson.M{"holding_code": holdingCode,
			"deleted_at": bson.M{"$exists": false},
		},
		&docList)
	if err != nil {
		return nil, err
	}

	return docList, nil
}

func (r *StockBalanceTransactionAdminRepository) FindStockBalanceDocDeleteByHoldingCode(ctx context.Context, holdingCode string) ([]stockBalanceProductModels.StockBalanceDoc, error) {
	docList := []stockBalanceProductModels.StockBalanceDoc{}

	err := r.pst.Find(ctx, &stockBalanceProductModels.StockBalanceDoc{},
		bson.M{"holding_code": holdingCode,
			"deleted_at": bson.M{"$exists": true},
		},
		&docList)
	if err != nil {
		return nil, err
	}

	return docList, nil
}
