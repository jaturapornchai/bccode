package datatransfer

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/saleinvoicebomprice/models"
	saleInvoiceBomPriceRepository "smlcloudplatform/internal/transaction/saleinvoicebomprice/repositories"
	"smlcloudplatform/pkg/microservice"
	msModels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SaleInvoiceBomPricesDataTransfer struct {
	transferConnection IDataTransferConnection
}

type ISaleInvoiceBomPricesDataTransferRepository interface {
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]models.SaleInvoiceBomPriceDoc, mongopagination.PaginationData, error)
}

type SaleInvoiceBomPricesDataTransferRepository struct {
	pst microservice.IPersisterMongo
	repositories.SearchRepository[models.SaleInvoiceBomPriceDoc]
}

func NewSaleInvoiceBomPricesDataTransferRepository(mongodbPersister microservice.IPersisterMongo) ISaleInvoiceBomPricesDataTransferRepository {
	repo := &SaleInvoiceBomPricesDataTransferRepository{
		pst: mongodbPersister,
	}

	repo.SearchRepository = repositories.NewSearchRepository[models.SaleInvoiceBomPriceDoc](mongodbPersister)
	return repo
}

func (repo SaleInvoiceBomPricesDataTransferRepository) FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]models.SaleInvoiceBomPriceDoc, mongopagination.PaginationData, error) {

	results, pagination, err := repo.SearchRepository.FindPage(ctx, shopID, searchInFields, pageable)

	if err != nil {
		return nil, mongopagination.PaginationData{}, err
	}

	return results, pagination, nil
}

func NewSaleInvoiceBomPricesDataTransfer(transferConnection IDataTransferConnection) IDataTransfer {

	return &SaleInvoiceBomPricesDataTransfer{
		transferConnection: transferConnection,
	}
}

func (pdt *SaleInvoiceBomPricesDataTransfer) StartTransfer(ctx context.Context, shopID string, targetShopID string) error {

	sourceRepository := NewSaleInvoiceBomPricesDataTransferRepository(pdt.transferConnection.GetSourceConnection())
	targetRepository := saleInvoiceBomPriceRepository.NewSaleInvoiceBomPriceRepository(pdt.transferConnection.GetTargetConnection())

	pageRequest := msModels.Pageable{
		Limit: 100,
		Page:  1,
	}

	for {

		docs, pages, err := sourceRepository.FindPage(ctx, shopID, nil, pageRequest)
		if err != nil {
			return err
		}

		if len(docs) > 0 {

			if targetShopID != "" {
				for i := range docs {
					docs[i].ShopID = targetShopID
					docs[i].ID = primitive.NewObjectID()
				}
			}

			err = targetRepository.CreateInBatch(ctx, docs)
			if err != nil {
				return err
			}
		}

		if pages.TotalPage > int64(pageRequest.Page) {
			pageRequest.Page++
		} else {
			break
		}
	}

	return nil

}
