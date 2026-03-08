package saleinvoicereturn

import (
	"context"
	saleInvoiceReturnModels "smlcloudplatform/internal/transaction/saleinvoicereturn/models"
	"smlcloudplatform/pkg/microservice"

	"go.mongodb.org/mongo-driver/bson"
)

type ISaleInvoiceReturnTransactionAdminRepositories interface {
	FindSaleInvoiceReturnDocByShopID(ctx context.Context, shopID string) ([]saleInvoiceReturnModels.SaleInvoiceReturnDoc, error)
	FindSaleInvoiceReturnDeleteDocByShopID(ctx context.Context, shopID string) ([]saleInvoiceReturnModels.SaleInvoiceReturnDoc, error)
}

type SaleInvoiceReturnTransactionAdminRepositories struct {
	pst microservice.IPersisterMongo
}

func NewSaleInvoiceReturnTransactionAdminRepositories(pst microservice.IPersisterMongo) ISaleInvoiceReturnTransactionAdminRepositories {
	return &SaleInvoiceReturnTransactionAdminRepositories{
		pst: pst,
	}
}

func (r SaleInvoiceReturnTransactionAdminRepositories) FindSaleInvoiceReturnDocByShopID(ctx context.Context, shopID string) ([]saleInvoiceReturnModels.SaleInvoiceReturnDoc, error) {

	docList := []saleInvoiceReturnModels.SaleInvoiceReturnDoc{}

	err := r.pst.Find(ctx, &saleInvoiceReturnModels.SaleInvoiceReturnDoc{},
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

func (r SaleInvoiceReturnTransactionAdminRepositories) FindSaleInvoiceReturnDeleteDocByShopID(ctx context.Context, shopID string) ([]saleInvoiceReturnModels.SaleInvoiceReturnDoc, error) {
	docList := []saleInvoiceReturnModels.SaleInvoiceReturnDoc{}

	err := r.pst.Find(ctx, &saleInvoiceReturnModels.SaleInvoiceReturnDoc{},
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
