package stocktransfer

import (
	"context"
	stocktransfermodels "smlcloudplatform/internal/transaction/stocktransfer/models"
	"smlcloudplatform/pkg/microservice"

	"go.mongodb.org/mongo-driver/bson"
)

type IStockTransferTransactionAdminRepository interface {
	FindStockTransferDocByShopID(ctx context.Context, shopID string) ([]stocktransfermodels.StockTransferDoc, error)
	FindStockTransferDocDeleteByShopID(ctx context.Context, shopID string) ([]stocktransfermodels.StockTransferDoc, error)
}

type StockTransferTransactionAdminRepository struct {
	pst microservice.IPersisterMongo
}

func NewStockTransferTransactionAdminRepository(pst microservice.IPersisterMongo) IStockTransferTransactionAdminRepository {
	return &StockTransferTransactionAdminRepository{
		pst: pst,
	}
}

func (r *StockTransferTransactionAdminRepository) FindStockTransferDocByShopID(ctx context.Context, shopID string) ([]stocktransfermodels.StockTransferDoc, error) {
	docList := []stocktransfermodels.StockTransferDoc{}

	err := r.pst.Find(ctx, &stocktransfermodels.StockTransferDoc{},
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

func (r *StockTransferTransactionAdminRepository) FindStockTransferDocDeleteByShopID(ctx context.Context, shopID string) ([]stocktransfermodels.StockTransferDoc, error) {
	docList := []stocktransfermodels.StockTransferDoc{}

	err := r.pst.Find(ctx, &stocktransfermodels.StockTransferDoc{},
		bson.M{"shopid": shopID,
			"deletedat": bson.M{"$exists": true},
		},
		&docList)
	if err != nil {
		return nil, err
	}

	return docList, nil
}
