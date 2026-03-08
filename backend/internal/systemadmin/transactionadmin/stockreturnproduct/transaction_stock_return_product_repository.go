package stockreturnproduct

import (
	"context"
	stockreturnproductmodels "smlcloudplatform/internal/transaction/stockreturnproduct/models"
	"smlcloudplatform/pkg/microservice"

	"go.mongodb.org/mongo-driver/bson"
)

type IStockReturnProductTransactionAdminRepository interface {
	FindStockReturnProductDocByShopID(ctx context.Context, shopID string) ([]stockreturnproductmodels.StockReturnProductDoc, error)
	FindStockReturnProductDeleteDocByShopID(ctx context.Context, shopID string) ([]stockreturnproductmodels.StockReturnProductDoc, error)
}

type StockReturnProductTransactionAdminRepository struct {
	pst microservice.IPersisterMongo
}

func NewStockReturnProductTransactionAdminRepository(pst microservice.IPersisterMongo) IStockReturnProductTransactionAdminRepository {
	return &StockReturnProductTransactionAdminRepository{
		pst: pst,
	}
}

func (r *StockReturnProductTransactionAdminRepository) FindStockReturnProductDocByShopID(ctx context.Context, shopID string) ([]stockreturnproductmodels.StockReturnProductDoc, error) {
	docList := []stockreturnproductmodels.StockReturnProductDoc{}

	err := r.pst.Find(ctx, &stockreturnproductmodels.StockReturnProductDoc{},
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

func (r *StockReturnProductTransactionAdminRepository) FindStockReturnProductDeleteDocByShopID(ctx context.Context, shopID string) ([]stockreturnproductmodels.StockReturnProductDoc, error) {
	docList := []stockreturnproductmodels.StockReturnProductDoc{}

	err := r.pst.Find(ctx, &stockreturnproductmodels.StockReturnProductDoc{},
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
