package saleinvoice

import (
	"context"
	"smlcloudplatform/internal/repositories"
	saleInvoiceModels "smlcloudplatform/internal/transaction/saleinvoice/models"
	"smlcloudplatform/pkg/microservice"
	msModels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type ISaleInvoiceTransactionAdminRepository interface {
	FindSaleInvoiceByShopID(ctx context.Context, shopID string) ([]saleInvoiceModels.SaleInvoiceDoc, error)
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]saleInvoiceModels.SaleInvoiceDoc, mongopagination.PaginationData, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable msModels.Pageable) ([]saleInvoiceModels.SaleInvoiceDoc, mongopagination.PaginationData, error)
	FindSaleInvoiceDeleteByShopID(ctx context.Context, shopID string) ([]saleInvoiceModels.SaleInvoiceDoc, error)
}

type SaleInvoiceTransactionAdminRepository struct {
	pst microservice.IPersisterMongo
	repositories.SearchRepository[saleInvoiceModels.SaleInvoiceDoc]
}

func NewSaleInvoiceTransactionAdminRepository(pst microservice.IPersisterMongo) ISaleInvoiceTransactionAdminRepository {
	return &SaleInvoiceTransactionAdminRepository{
		pst:              pst,
		SearchRepository: repositories.NewSearchRepository[saleInvoiceModels.SaleInvoiceDoc](pst),
	}
}

func (r SaleInvoiceTransactionAdminRepository) FindSaleInvoiceByShopID(ctx context.Context, shopID string) ([]saleInvoiceModels.SaleInvoiceDoc, error) {
	docList := []saleInvoiceModels.SaleInvoiceDoc{}

	err := r.pst.Find(ctx, &saleInvoiceModels.SaleInvoiceDoc{},
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

func (r SaleInvoiceTransactionAdminRepository) FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]saleInvoiceModels.SaleInvoiceDoc, mongopagination.PaginationData, error) {

	results, pagination, err := r.SearchRepository.FindPage(ctx, shopID, searchInFields, pageable)

	if err != nil {
		return nil, mongopagination.PaginationData{}, err
	}

	return results, pagination, nil
}

func (r SaleInvoiceTransactionAdminRepository) FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable msModels.Pageable) ([]saleInvoiceModels.SaleInvoiceDoc, mongopagination.PaginationData, error) {

	results, pagination, err := r.SearchRepository.FindPageFilter(ctx, shopID, filters, searchInFields, pageable)

	if err != nil {
		return nil, mongopagination.PaginationData{}, err
	}

	return results, pagination, nil
}

func (r SaleInvoiceTransactionAdminRepository) FindSaleInvoiceDeleteByShopID(ctx context.Context, shopID string) ([]saleInvoiceModels.SaleInvoiceDoc, error) {
	docList := []saleInvoiceModels.SaleInvoiceDoc{}

	err := r.pst.Find(ctx, &saleInvoiceModels.SaleInvoiceDoc{},
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
