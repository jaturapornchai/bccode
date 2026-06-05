package saleinvoicereturn

import (
	"context"
	saleInvoiceReturnModels "smlcloudplatform/internal/transaction/saleinvoicereturn/models"
	"smlcloudplatform/pkg/microservice"

	"go.mongodb.org/mongo-driver/bson"
)

type ISaleInvoiceReturnTransactionAdminRepositories interface {
	FindSaleInvoiceReturnDocByHoldingCode(ctx context.Context, holdingCode string) ([]saleInvoiceReturnModels.SaleInvoiceReturnDoc, error)
	FindSaleInvoiceReturnDeleteDocByHoldingCode(ctx context.Context, holdingCode string) ([]saleInvoiceReturnModels.SaleInvoiceReturnDoc, error)
}

type SaleInvoiceReturnTransactionAdminRepositories struct {
	pst microservice.IPersisterMongo
}

func NewSaleInvoiceReturnTransactionAdminRepositories(pst microservice.IPersisterMongo) ISaleInvoiceReturnTransactionAdminRepositories {
	return &SaleInvoiceReturnTransactionAdminRepositories{
		pst: pst,
	}
}

func (r SaleInvoiceReturnTransactionAdminRepositories) FindSaleInvoiceReturnDocByHoldingCode(ctx context.Context, holdingCode string) ([]saleInvoiceReturnModels.SaleInvoiceReturnDoc, error) {

	docList := []saleInvoiceReturnModels.SaleInvoiceReturnDoc{}

	err := r.pst.Find(ctx, &saleInvoiceReturnModels.SaleInvoiceReturnDoc{},
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

func (r SaleInvoiceReturnTransactionAdminRepositories) FindSaleInvoiceReturnDeleteDocByHoldingCode(ctx context.Context, holdingCode string) ([]saleInvoiceReturnModels.SaleInvoiceReturnDoc, error) {
	docList := []saleInvoiceReturnModels.SaleInvoiceReturnDoc{}

	err := r.pst.Find(ctx, &saleInvoiceReturnModels.SaleInvoiceReturnDoc{},
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
