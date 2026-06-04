package stocktransfer

import (
	"context"
	stocktransfermodels "smlcloudplatform/internal/transaction/stocktransfer/models"
	"smlcloudplatform/pkg/microservice"

	"go.mongodb.org/mongo-driver/bson"
)

type IStockTransferTransactionAdminRepository interface {
	FindStockTransferDocByHoldingCode(ctx context.Context, holdingCode string) ([]stocktransfermodels.StockTransferDoc, error)
	FindStockTransferDocDeleteByHoldingCode(ctx context.Context, holdingCode string) ([]stocktransfermodels.StockTransferDoc, error)
}

type StockTransferTransactionAdminRepository struct {
	pst microservice.IPersisterMongo
}

func NewStockTransferTransactionAdminRepository(pst microservice.IPersisterMongo) IStockTransferTransactionAdminRepository {
	return &StockTransferTransactionAdminRepository{
		pst: pst,
	}
}

func (r *StockTransferTransactionAdminRepository) FindStockTransferDocByHoldingCode(ctx context.Context, holdingCode string) ([]stocktransfermodels.StockTransferDoc, error) {
	docList := []stocktransfermodels.StockTransferDoc{}

	err := r.pst.Find(ctx, &stocktransfermodels.StockTransferDoc{},
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

func (r *StockTransferTransactionAdminRepository) FindStockTransferDocDeleteByHoldingCode(ctx context.Context, holdingCode string) ([]stocktransfermodels.StockTransferDoc, error) {
	docList := []stocktransfermodels.StockTransferDoc{}

	err := r.pst.Find(ctx, &stocktransfermodels.StockTransferDoc{},
		bson.M{"holding_code": holdingCode,
			"deleted_at": bson.M{"$exists": true},
		},
		&docList)
	if err != nil {
		return nil, err
	}

	return docList, nil
}
