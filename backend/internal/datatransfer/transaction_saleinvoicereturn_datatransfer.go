package datatransfer

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/saleinvoicereturn/models"
	saleInvoiceReturnRepository "smlcloudplatform/internal/transaction/saleinvoicereturn/repositories"
	"smlcloudplatform/pkg/microservice"
	msModels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SaleInvoiceReturnDataTransfer struct {
	transferConnection IDataTransferConnection
}

type ISaleInvoiceReturnDataTransferRepository interface {
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]models.SaleInvoiceReturnDoc, mongopagination.PaginationData, error)
}

type SaleInvoiceReturnDataTransferRepository struct {
	pst microservice.IPersisterMongo
	repositories.SearchRepository[models.SaleInvoiceReturnDoc]
}

func NewSaleInvoiceReturnDataTransferRepository(mongodbPersister microservice.IPersisterMongo) ISaleInvoiceReturnDataTransferRepository {
	repo := &SaleInvoiceReturnDataTransferRepository{
		pst: mongodbPersister,
	}

	repo.SearchRepository = repositories.NewSearchRepository[models.SaleInvoiceReturnDoc](mongodbPersister)
	return repo
}

func (repo SaleInvoiceReturnDataTransferRepository) FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]models.SaleInvoiceReturnDoc, mongopagination.PaginationData, error) {

	results, pagination, err := repo.SearchRepository.FindPage(ctx, shopID, searchInFields, pageable)

	if err != nil {
		return nil, mongopagination.PaginationData{}, err
	}

	return results, pagination, nil
}

func NewSaleInvoiceReturnDataTransfer(transferConnection IDataTransferConnection) IDataTransfer {

	return &SaleInvoiceReturnDataTransfer{
		transferConnection: transferConnection,
	}
}

func (pdt *SaleInvoiceReturnDataTransfer) StartTransfer(ctx context.Context, shopID string, targetShopID string) error {

	sourceRepository := NewSaleInvoiceReturnDataTransferRepository(pdt.transferConnection.GetSourceConnection())
	targetRepository := saleInvoiceReturnRepository.NewSaleInvoiceReturnRepository(pdt.transferConnection.GetTargetConnection())

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
