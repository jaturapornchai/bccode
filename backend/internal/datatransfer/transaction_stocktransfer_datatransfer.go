package datatransfer

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/stocktransfer/models"
	stockTransferRepository "smlcloudplatform/internal/transaction/stocktransfer/repositories"
	"smlcloudplatform/pkg/microservice"
	msModels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StockTransferDataTransfer struct {
	transferConnection IDataTransferConnection
}

type IStockTransferDataTransferRepository interface {
	FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable msModels.Pageable) ([]models.StockTransferDoc, mongopagination.PaginationData, error)
}

type StockTransferDataTransferRepository struct {
	pst microservice.IPersisterMongo
	repositories.SearchRepository[models.StockTransferDoc]
}

func NewStockTransferDataTransferRepository(mongodbPersister microservice.IPersisterMongo) IStockTransferDataTransferRepository {
	repo := &StockTransferDataTransferRepository{
		pst: mongodbPersister,
	}

	repo.SearchRepository = repositories.NewSearchRepository[models.StockTransferDoc](mongodbPersister)
	return repo
}

func (repo StockTransferDataTransferRepository) FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable msModels.Pageable) ([]models.StockTransferDoc, mongopagination.PaginationData, error) {

	results, pagination, err := repo.SearchRepository.FindPage(ctx, holdingCode, searchInFields, pageable)

	if err != nil {
		return nil, mongopagination.PaginationData{}, err
	}

	return results, pagination, nil
}

func NewStockTransferDataTransfer(transferConnection IDataTransferConnection) IDataTransfer {
	return &StockTransferDataTransfer{
		transferConnection: transferConnection,
	}
}

func (pdt *StockTransferDataTransfer) StartTransfer(ctx context.Context, holdingCode string, targetHoldingCode string) error {

	sourceRepository := NewStockTransferDataTransferRepository(pdt.transferConnection.GetSourceConnection())
	targetRepository := stockTransferRepository.NewStockTransferRepository(pdt.transferConnection.GetTargetConnection())

	pageRequest := msModels.Pageable{
		Limit: 100,
		Page:  1,
	}

	for {

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
