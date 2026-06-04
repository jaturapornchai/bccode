package datatransfer

import (
	"context"
	"fmt"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/saleinvoice/models"
	saleInvoiceRepository "smlcloudplatform/internal/transaction/saleinvoice/repositories"
	"smlcloudplatform/pkg/microservice"
	msModels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SaleInvoiceDataTransfer struct {
	transferConnection IDataTransferConnection
}

type ISaleInvoiceDataTransferRepository interface {
	FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable msModels.Pageable) ([]models.SaleInvoiceDoc, mongopagination.PaginationData, error)
}

type SaleInvoiceDataTransferRepository struct {
	pst microservice.IPersisterMongo
	repositories.SearchRepository[models.SaleInvoiceDoc]
}

func NewSaleInvoiceDataTransferRepository(mongodbPersister microservice.IPersisterMongo) ISaleInvoiceDataTransferRepository {
	repo := &SaleInvoiceDataTransferRepository{
		pst: mongodbPersister,
	}

	repo.SearchRepository = repositories.NewSearchRepository[models.SaleInvoiceDoc](mongodbPersister)
	return repo
}

func (repo SaleInvoiceDataTransferRepository) FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable msModels.Pageable) ([]models.SaleInvoiceDoc, mongopagination.PaginationData, error) {

	results, pagination, err := repo.SearchRepository.FindPage(ctx, holdingCode, searchInFields, pageable)

	if err != nil {
		return nil, mongopagination.PaginationData{}, err
	}

	return results, pagination, nil
}

func NewSaleInvoiceDataTransfer(transferConnection IDataTransferConnection) IDataTransfer {
	return &SaleInvoiceDataTransfer{
		transferConnection: transferConnection,
	}
}

func (pdt *SaleInvoiceDataTransfer) StartTransfer(ctx context.Context, holdingCode string, targetHoldingCode string) error {

	sourceRepository := NewSaleInvoiceDataTransferRepository(pdt.transferConnection.GetSourceConnection())
	targetRepository := saleInvoiceRepository.NewSaleInvoiceRepository(pdt.transferConnection.GetTargetConnection())

	pageRequest := msModels.Pageable{
		Limit: 100,
		Page:  1,
	}

	for {

		fmt.Println("Transfering page: ", pageRequest.Page)
		docs, pages, err := sourceRepository.FindPage(ctx, holdingCode, nil, pageRequest)
		if err != nil {
			return err
		}

		if len(docs) > 0 {

			if targetHoldingCode != "" {
				for i := range docs {
					docs[i].HoldingCode = targetHoldingCode
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
