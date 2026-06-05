package stockadjustment

import (
	"context"
	stockadjustmentmodels "smlcloudplatform/internal/transaction/stockadjustment/models"
	"smlcloudplatform/pkg/microservice"

	"go.mongodb.org/mongo-driver/bson"
)

type IStockAdjustmentTransactionAdminRepository interface {
	FindStockAdjustmentDocByHoldingCode(ctx context.Context, holdingCode string) ([]stockadjustmentmodels.StockAdjustmentDoc, error)
	FindStockAdjustmentDocDeleteByHoldingCode(ctx context.Context, holdingCode string) ([]stockadjustmentmodels.StockAdjustmentDoc, error)
}

type StockAdjustmentTransactionAdminRepository struct {
	pst microservice.IPersisterMongo
}

func NewStockAdjustmentTransactionAdminRepository(pst microservice.IPersisterMongo) IStockAdjustmentTransactionAdminRepository {
	return &StockAdjustmentTransactionAdminRepository{
		pst: pst,
	}
}

func (r *StockAdjustmentTransactionAdminRepository) FindStockAdjustmentDocByHoldingCode(ctx context.Context, holdingCode string) ([]stockadjustmentmodels.StockAdjustmentDoc, error) {
	docList := []stockadjustmentmodels.StockAdjustmentDoc{}

	err := r.pst.Find(ctx, &stockadjustmentmodels.StockAdjustmentDoc{},
		bson.M{
			"holdingcode": holdingCode,
			"deletedat":   bson.M{"$exists": false},
		},
		&docList)
	if err != nil {
		return nil, err
	}

	return docList, nil
}

func (r *StockAdjustmentTransactionAdminRepository) FindStockAdjustmentDocDeleteByHoldingCode(ctx context.Context, holdingCode string) ([]stockadjustmentmodels.StockAdjustmentDoc, error) {
	docList := []stockadjustmentmodels.StockAdjustmentDoc{}

	err := r.pst.Find(ctx, &stockadjustmentmodels.StockAdjustmentDoc{},
		bson.M{
			"holdingcode": holdingCode,
			"deletedat":   bson.M{"$exists": true},
		},
		&docList)
	if err != nil {
		return nil, err
	}

	return docList, nil
}
