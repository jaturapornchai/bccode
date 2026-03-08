package stockreceiveproduct

import (
	"context"
	stockReceiveProductModels "smlcloudplatform/internal/transaction/stockreceiveproduct/models"
	"smlcloudplatform/pkg/microservice"

	"go.mongodb.org/mongo-driver/bson"
)

type IStockReceiveTransactionAdminRepository interface {
	FindStockReceiveDocByShopID(ctx context.Context, shopID string) ([]stockReceiveProductModels.StockReceiveProductDoc, error)
	FindStockReceiveDeleteDocByShopID(ctx context.Context, shopID string) ([]stockReceiveProductModels.StockReceiveProductDoc, error)
}

type StockReceiveTransactionAdminRepository struct {
	pst microservice.IPersisterMongo
}

func NewStockReceiveTransactionAdminRepository(pst microservice.IPersisterMongo) IStockReceiveTransactionAdminRepository {
	return &StockReceiveTransactionAdminRepository{
		pst: pst,
	}
}

func (r *StockReceiveTransactionAdminRepository) FindStockReceiveDocByShopID(ctx context.Context, shopID string) ([]stockReceiveProductModels.StockReceiveProductDoc, error) {
	docList := []stockReceiveProductModels.StockReceiveProductDoc{}

	err := r.pst.Find(ctx, &stockReceiveProductModels.StockReceiveProductDoc{},
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

func (r *StockReceiveTransactionAdminRepository) FindStockReceiveDeleteDocByShopID(ctx context.Context, shopID string) ([]stockReceiveProductModels.StockReceiveProductDoc, error) {
	docList := []stockReceiveProductModels.StockReceiveProductDoc{}

	err := r.pst.Find(ctx, &stockReceiveProductModels.StockReceiveProductDoc{},
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
