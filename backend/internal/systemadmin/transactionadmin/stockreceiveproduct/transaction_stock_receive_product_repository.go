package stockreceiveproduct

import (
	"context"
	stockReceiveProductModels "smlcloudplatform/internal/transaction/stockreceiveproduct/models"
	"smlcloudplatform/pkg/microservice"

	"go.mongodb.org/mongo-driver/bson"
)

type IStockReceiveTransactionAdminRepository interface {
	FindStockReceiveDocByHoldingCode(ctx context.Context, holdingCode string) ([]stockReceiveProductModels.StockReceiveProductDoc, error)
	FindStockReceiveDeleteDocByHoldingCode(ctx context.Context, holdingCode string) ([]stockReceiveProductModels.StockReceiveProductDoc, error)
}

type StockReceiveTransactionAdminRepository struct {
	pst microservice.IPersisterMongo
}

func NewStockReceiveTransactionAdminRepository(pst microservice.IPersisterMongo) IStockReceiveTransactionAdminRepository {
	return &StockReceiveTransactionAdminRepository{
		pst: pst,
	}
}

func (r *StockReceiveTransactionAdminRepository) FindStockReceiveDocByHoldingCode(ctx context.Context, holdingCode string) ([]stockReceiveProductModels.StockReceiveProductDoc, error) {
	docList := []stockReceiveProductModels.StockReceiveProductDoc{}

	err := r.pst.Find(ctx, &stockReceiveProductModels.StockReceiveProductDoc{},
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

func (r *StockReceiveTransactionAdminRepository) FindStockReceiveDeleteDocByHoldingCode(ctx context.Context, holdingCode string) ([]stockReceiveProductModels.StockReceiveProductDoc, error) {
	docList := []stockReceiveProductModels.StockReceiveProductDoc{}

	err := r.pst.Find(ctx, &stockReceiveProductModels.StockReceiveProductDoc{},
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
