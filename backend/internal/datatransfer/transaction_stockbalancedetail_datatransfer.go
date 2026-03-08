package datatransfer

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/stockbalancedetail/models"
	stockbalancedetailrepository "smlcloudplatform/internal/transaction/stockbalancedetail/repositories"
	"smlcloudplatform/pkg/microservice"
	msModels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StockBalanceDetailDataTransfer struct {
	transferConnection IDataTransferConnection
}

type IStockBalanceDetailDataTransferRepository interface {
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]models.StockBalanceDetailDoc, mongopagination.PaginationData, error)
}

type StockBalanceDetailDataTransferRepository struct {
	pst microservice.IPersisterMongo
	repositories.SearchRepository[models.StockBalanceDetailDoc]
}

func NewStockBalanceDetailDataTransferRepository(mongodbPersister microservice.IPersisterMongo) IStockBalanceDetailDataTransferRepository {

	repo := &StockBalanceDetailDataTransferRepository{
		pst: mongodbPersister,
	}

	repo.SearchRepository = repositories.NewSearchRepository[models.StockBalanceDetailDoc](mongodbPersister)
	return repo
}

func (repo StockBalanceDetailDataTransferRepository) FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]models.StockBalanceDetailDoc, mongopagination.PaginationData, error) {

	results, pagination, err := repo.SearchRepository.FindPage(ctx, shopID, searchInFields, pageable)

	if err != nil {
		return nil, mongopagination.PaginationData{}, err
	}

	return results, pagination, nil
}

func NewStockBalanceDetailDataTransfer(transferConnection IDataTransferConnection) IDataTransfer {

	return &StockBalanceDetailDataTransfer{
		transferConnection: transferConnection,
	}
}

func (pdt *StockBalanceDetailDataTransfer) StartTransfer(ctx context.Context, shopID string, targetShopID string) error {

	sourceRepository := NewStockBalanceDetailDataTransferRepository(pdt.transferConnection.GetSourceConnection())
	targetRepository := stockbalancedetailrepository.NewStockBalanceDetailRepository(pdt.transferConnection.GetTargetConnection())

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
