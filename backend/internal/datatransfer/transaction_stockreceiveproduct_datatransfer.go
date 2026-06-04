package datatransfer

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/stockreceiveproduct/models"
	stockReceiveProductRepository "smlcloudplatform/internal/transaction/stockreceiveproduct/repositories"
	"smlcloudplatform/pkg/microservice"
	msModels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StockReceiveProductDataTransfer struct {
	transferConnection IDataTransferConnection
}

type IStockReceiveProductDataTransferRepository interface {
	FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable msModels.Pageable) ([]models.StockReceiveProductDoc, mongopagination.PaginationData, error)
}

type StockReceiveProductDataTransferRepository struct {
	pst microservice.IPersisterMongo
	repositories.SearchRepository[models.StockReceiveProductDoc]
}

func NewStockReceiveProductDataTransferRepository(mongodbPersister microservice.IPersisterMongo) IStockReceiveProductDataTransferRepository {
	repo := &StockReceiveProductDataTransferRepository{
		pst: mongodbPersister,
	}

	repo.SearchRepository = repositories.NewSearchRepository[models.StockReceiveProductDoc](mongodbPersister)
	return repo
}

func (repo StockReceiveProductDataTransferRepository) FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable msModels.Pageable) ([]models.StockReceiveProductDoc, mongopagination.PaginationData, error) {

	results, pagination, err := repo.SearchRepository.FindPage(ctx, holdingCode, searchInFields, pageable)

	if err != nil {
		return nil, mongopagination.PaginationData{}, err
	}

	return results, pagination, nil
}

func NewStockReceiveProductDataTransfer(transferConnection IDataTransferConnection) IDataTransfer {
	return &StockReceiveProductDataTransfer{
		transferConnection: transferConnection,
	}
}

func (pdt *StockReceiveProductDataTransfer) StartTransfer(ctx context.Context, holdingCode string, targetHoldingCode string) error {

	sourceRepository := NewStockReceiveProductDataTransferRepository(pdt.transferConnection.GetSourceConnection())
	targetRepository := stockReceiveProductRepository.NewStockReceiveProductRepository(pdt.transferConnection.GetTargetConnection())

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
