package stockbalance

import (
	"context"
	stockBalanceProductModels "smlcloudplatform/internal/transaction/stockbalance/models"
	"smlcloudplatform/pkg/microservice"

	"go.mongodb.org/mongo-driver/bson"
)

type IStockBalanceTransactionAdminRepository interface {
	FindStockBalanceDocByShopID(ctx context.Context, shopID string) ([]stockBalanceProductModels.StockBalanceDoc, error)
	FindStockBalanceDocDeleteByShopID(ctx context.Context, shopID string) ([]stockBalanceProductModels.StockBalanceDoc, error)
}

type StockBalanceTransactionAdminRepository struct {
	pst microservice.IPersisterMongo
}

func NewStockBalanceTransactionAdminRepository(pst microservice.IPersisterMongo) IStockBalanceTransactionAdminRepository {
	return &StockBalanceTransactionAdminRepository{
		pst: pst,
	}
}

func (r *StockBalanceTransactionAdminRepository) FindStockBalanceDocByShopID(ctx context.Context, shopID string) ([]stockBalanceProductModels.StockBalanceDoc, error) {
	docList := []stockBalanceProductModels.StockBalanceDoc{}

	err := r.pst.Find(ctx, &stockBalanceProductModels.StockBalanceDoc{},
		bson.M{"shopid": shopID,
			"deletedat": bson.M{"$exists": false},
		},
		&docList)
	if err != nil {
		return nil, err
	}

	return docList, nil
}

func (r *StockBalanceTransactionAdminRepository) FindStockBalanceDocDeleteByShopID(ctx context.Context, shopID string) ([]stockBalanceProductModels.StockBalanceDoc, error) {
	docList := []stockBalanceProductModels.StockBalanceDoc{}

	err := r.pst.Find(ctx, &stockBalanceProductModels.StockBalanceDoc{},
		bson.M{"shopid": shopID,
			"deletedat": bson.M{"$exists": true},
		},
		&docList)
	if err != nil {
		return nil, err
	}

	return docList, nil
}
