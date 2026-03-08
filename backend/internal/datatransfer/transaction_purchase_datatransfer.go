package datatransfer

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/purchase/models"
	transactionPurchaserRepository "smlcloudplatform/internal/transaction/purchase/repositories"
	"smlcloudplatform/pkg/microservice"
	msModels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TransactionPurchaseDataTransfer struct {
	transferConnection IDataTransferConnection
}

type ITransactionPurchaseDataTransferRepository interface {
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]models.PurchaseDoc, mongopagination.PaginationData, error)
}

type TransactionPurchaseDataTransferRepository struct {
	pst microservice.IPersisterMongo
	repositories.SearchRepository[models.PurchaseDoc]
}

func NewTransactionPurchaseDataTransferRepository(mongodbPersister microservice.IPersisterMongo) ITransactionPurchaseDataTransferRepository {
	repo := &TransactionPurchaseDataTransferRepository{
		pst: mongodbPersister,
	}

	repo.SearchRepository = repositories.NewSearchRepository[models.PurchaseDoc](mongodbPersister)
	return repo
}

func (repo TransactionPurchaseDataTransferRepository) FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]models.PurchaseDoc, mongopagination.PaginationData, error) {

	results, pagination, err := repo.SearchRepository.FindPage(ctx, shopID, searchInFields, pageable)

	if err != nil {
		return nil, mongopagination.PaginationData{}, err
	}

	return results, pagination, nil
}

func NewTransactionPurchaseDataTransfer(transferConnection IDataTransferConnection) IDataTransfer {
	return &TransactionPurchaseDataTransfer{
		transferConnection: transferConnection,
	}
}

func (pdt *TransactionPurchaseDataTransfer) StartTransfer(ctx context.Context, shopID string, targetShopID string) error {

	sourceRepository := NewTransactionPurchaseDataTransferRepository(pdt.transferConnection.GetSourceConnection())
	targetRepository := transactionPurchaserRepository.NewPurchaseRepository(pdt.transferConnection.GetTargetConnection())

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
