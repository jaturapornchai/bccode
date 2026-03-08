package datatransfer

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/stockpickupproduct/models"
	stockPickupProductRepository "smlcloudplatform/internal/transaction/stockpickupproduct/repositories"
	"smlcloudplatform/pkg/microservice"
	msModels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StockPickupProductDataTransfer struct {
	transferConnection IDataTransferConnection
}

type IStockPickupProductDataTransferRepository interface {
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]models.StockPickupProductDoc, mongopagination.PaginationData, error)
}

type StockPickupProductDataTransferRepository struct {
	pst microservice.IPersisterMongo
	repositories.SearchRepository[models.StockPickupProductDoc]
}

func NewStockPickupProductDataTransferRepository(mongodbPersister microservice.IPersisterMongo) IStockPickupProductDataTransferRepository {
	repo := &StockPickupProductDataTransferRepository{
		pst: mongodbPersister,
	}

	repo.SearchRepository = repositories.NewSearchRepository[models.StockPickupProductDoc](mongodbPersister)
	return repo
}

func (repo StockPickupProductDataTransferRepository) FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]models.StockPickupProductDoc, mongopagination.PaginationData, error) {

	results, pagination, err := repo.SearchRepository.FindPage(ctx, shopID, searchInFields, pageable)

	if err != nil {
		return nil, mongopagination.PaginationData{}, err
	}

	return results, pagination, nil
}

func NewStockPickupProductDataTransfer(transferConnection IDataTransferConnection) IDataTransfer {
	return &StockPickupProductDataTransfer{
		transferConnection: transferConnection,
	}
}

func (pdt *StockPickupProductDataTransfer) StartTransfer(ctx context.Context, shopID string, targetShopID string) error {

	sourceRepository := NewStockPickupProductDataTransferRepository(pdt.transferConnection.GetSourceConnection())
	targetRepository := stockPickupProductRepository.NewStockPickupProductRepository(pdt.transferConnection.GetTargetConnection())

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
