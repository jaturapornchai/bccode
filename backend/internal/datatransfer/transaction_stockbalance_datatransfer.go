package datatransfer

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/stockbalance/models"
	stockbalancerepository "smlcloudplatform/internal/transaction/stockbalance/repositories"
	"smlcloudplatform/pkg/microservice"
	msModels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StockBalanceDataTransfer struct {
	transferConnection IDataTransferConnection
}

type IStockBalanceDataTransferRepository interface {
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]models.StockBalanceDoc, mongopagination.PaginationData, error)
}

type StockBalanceDataTransferRepository struct {
	pst microservice.IPersisterMongo
	repositories.SearchRepository[models.StockBalanceDoc]
}

func NewStockBalanceDataTransferRepository(mongodbPersister microservice.IPersisterMongo) IStockBalanceDataTransferRepository {

	repo := &StockBalanceDataTransferRepository{
		pst: mongodbPersister,
	}

	repo.SearchRepository = repositories.NewSearchRepository[models.StockBalanceDoc](mongodbPersister)
	return repo
}

func (repo StockBalanceDataTransferRepository) FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]models.StockBalanceDoc, mongopagination.PaginationData, error) {

	results, pagination, err := repo.SearchRepository.FindPage(ctx, shopID, searchInFields, pageable)

	if err != nil {
		return nil, mongopagination.PaginationData{}, err
	}

	return results, pagination, nil
}

func NewStockBalanceDataTransfer(transferConnection IDataTransferConnection) IDataTransfer {

	return &StockBalanceDataTransfer{
		transferConnection: transferConnection,
	}
}

func (pdt *StockBalanceDataTransfer) StartTransfer(ctx context.Context, shopID string, targetShopID string) error {

	sourceRepository := NewStockBalanceDataTransferRepository(pdt.transferConnection.GetSourceConnection())
	targetRepository := stockbalancerepository.NewStockBalanceRepository(pdt.transferConnection.GetTargetConnection())

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
