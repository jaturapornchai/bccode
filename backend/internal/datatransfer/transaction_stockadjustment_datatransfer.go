package datatransfer

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/stockadjustment/models"
	stockAdjustmentRepository "smlcloudplatform/internal/transaction/stockadjustment/repositories"
	"smlcloudplatform/pkg/microservice"
	msModels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StockAdjustmentDataTransfer struct {
	transferConnection IDataTransferConnection
}

type IStockAdjustmentDataTransferRepository interface {
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]models.StockAdjustmentDoc, mongopagination.PaginationData, error)
}

type StockAdjustmentDataTransferRepository struct {
	pst microservice.IPersisterMongo
	repositories.SearchRepository[models.StockAdjustmentDoc]
}

func NewStockAdjustmentDataTransferRepository(mongodbPersister microservice.IPersisterMongo) IStockAdjustmentDataTransferRepository {
	repo := &StockAdjustmentDataTransferRepository{
		pst: mongodbPersister,
	}

	repo.SearchRepository = repositories.NewSearchRepository[models.StockAdjustmentDoc](mongodbPersister)
	return repo
}

func (repo StockAdjustmentDataTransferRepository) FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]models.StockAdjustmentDoc, mongopagination.PaginationData, error) {

	results, pagination, err := repo.SearchRepository.FindPage(ctx, shopID, searchInFields, pageable)

	if err != nil {
		return nil, mongopagination.PaginationData{}, err
	}

	return results, pagination, nil
}

func NewStockAdjustmentDataTransfer(transferConnection IDataTransferConnection) IDataTransfer {
	return &StockAdjustmentDataTransfer{
		transferConnection: transferConnection,
	}
}

func (pdt *StockAdjustmentDataTransfer) StartTransfer(ctx context.Context, shopID string, targetShopID string) error {

	sourceRepository := NewStockAdjustmentDataTransferRepository(pdt.transferConnection.GetSourceConnection())
	targetRepository := stockAdjustmentRepository.NewStockAdjustmentRepository(pdt.transferConnection.GetTargetConnection())

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
