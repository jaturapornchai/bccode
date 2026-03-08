package datatransfer

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/stockreturnproduct/models"
	stockReturnProductRepository "smlcloudplatform/internal/transaction/stockreturnproduct/repositories"
	"smlcloudplatform/pkg/microservice"
	msModels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StockReturnProductDataTransfer struct {
	transferConnection IDataTransferConnection
}

type IStockReturnProductDataTransferRepository interface {
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]models.StockReturnProductDoc, mongopagination.PaginationData, error)
}

type StockReturnProductDataTransferRepository struct {
	pst microservice.IPersisterMongo
	repositories.SearchRepository[models.StockReturnProductDoc]
}

func NewStockReturnProductDataTransferRepository(mongodbPersister microservice.IPersisterMongo) IStockReturnProductDataTransferRepository {
	repo := &StockReturnProductDataTransferRepository{
		pst: mongodbPersister,
	}

	repo.SearchRepository = repositories.NewSearchRepository[models.StockReturnProductDoc](mongodbPersister)
	return repo
}

func (repo StockReturnProductDataTransferRepository) FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]models.StockReturnProductDoc, mongopagination.PaginationData, error) {

	results, pagination, err := repo.SearchRepository.FindPage(ctx, shopID, searchInFields, pageable)

	if err != nil {
		return nil, mongopagination.PaginationData{}, err
	}

	return results, pagination, nil
}

func NewStockReturnProductDataTransfer(transferConnection IDataTransferConnection) IDataTransfer {
	return &StockReturnProductDataTransfer{
		transferConnection: transferConnection,
	}
}

func (pdt *StockReturnProductDataTransfer) StartTransfer(ctx context.Context, shopID string, targetShopID string) error {

	sourceRepository := NewStockReturnProductDataTransferRepository(pdt.transferConnection.GetSourceConnection())
	targetRepository := stockReturnProductRepository.NewStockReturnProductRepository(pdt.transferConnection.GetTargetConnection())

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
