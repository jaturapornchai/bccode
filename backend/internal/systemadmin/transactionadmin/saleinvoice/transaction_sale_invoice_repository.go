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
	FindSaleInvoiceByHoldingCode(ctx context.Context, holdingCode string) ([]saleInvoiceModels.SaleInvoiceDoc, error)
	FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable msModels.Pageable) ([]saleInvoiceModels.SaleInvoiceDoc, mongopagination.PaginationData, error)
	FindPageFilter(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, pageable msModels.Pageable) ([]saleInvoiceModels.SaleInvoiceDoc, mongopagination.PaginationData, error)
	FindSaleInvoiceDeleteByHoldingCode(ctx context.Context, holdingCode string) ([]saleInvoiceModels.SaleInvoiceDoc, error)
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

func (r SaleInvoiceTransactionAdminRepository) FindSaleInvoiceByHoldingCode(ctx context.Context, holdingCode string) ([]saleInvoiceModels.SaleInvoiceDoc, error) {
	docList := []saleInvoiceModels.SaleInvoiceDoc{}

	err := r.pst.Find(ctx, &saleInvoiceModels.SaleInvoiceDoc{},
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

func (r SaleInvoiceTransactionAdminRepository) FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable msModels.Pageable) ([]saleInvoiceModels.SaleInvoiceDoc, mongopagination.PaginationData, error) {

	results, pagination, err := r.SearchRepository.FindPage(ctx, holdingCode, searchInFields, pageable)

	if err != nil {
		return nil, mongopagination.PaginationData{}, err
	}

	return results, pagination, nil
}

func (r SaleInvoiceTransactionAdminRepository) FindPageFilter(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, pageable msModels.Pageable) ([]saleInvoiceModels.SaleInvoiceDoc, mongopagination.PaginationData, error) {

	results, pagination, err := r.SearchRepository.FindPageFilter(ctx, holdingCode, filters, searchInFields, pageable)

	if err != nil {
		return nil, mongopagination.PaginationData{}, err
	}

	return results, pagination, nil
}

func (r SaleInvoiceTransactionAdminRepository) FindSaleInvoiceDeleteByHoldingCode(ctx context.Context, holdingCode string) ([]saleInvoiceModels.SaleInvoiceDoc, error) {
	docList := []saleInvoiceModels.SaleInvoiceDoc{}

	err := r.pst.Find(ctx, &saleInvoiceModels.SaleInvoiceDoc{},
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
