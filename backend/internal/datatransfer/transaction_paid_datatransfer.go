package datatransfer

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/paid/models"
	transactionPaidRepository "smlcloudplatform/internal/transaction/paid/repositories"
	"smlcloudplatform/pkg/microservice"
	msModels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TransactionPaidDataTransfer struct {
	transferConnection IDataTransferConnection
}

type ITransactionPaidDataTransferRepository interface {
	FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable msModels.Pageable) ([]models.PaidDoc, mongopagination.PaginationData, error)
}

type TransactionPaidDataTransferRepository struct {
	pst microservice.IPersisterMongo
	repositories.SearchRepository[models.PaidDoc]
}

func NewTransactionPaidDataTransferRepository(mongodbPersister microservice.IPersisterMongo) ITransactionPaidDataTransferRepository {
	repo := &TransactionPaidDataTransferRepository{
		pst: mongodbPersister,
	}

	repo.SearchRepository = repositories.NewSearchRepository[models.PaidDoc](mongodbPersister)
	return repo
}

func (repo TransactionPaidDataTransferRepository) FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable msModels.Pageable) ([]models.PaidDoc, mongopagination.PaginationData, error) {

	results, pagination, err := repo.SearchRepository.FindPage(ctx, holdingCode, searchInFields, pageable)

	if err != nil {
		return nil, mongopagination.PaginationData{}, err
	}

	return results, pagination, nil
}

func NewTransactionPaidDataTransfer(transferConnection IDataTransferConnection) IDataTransfer {
	return &TransactionPaidDataTransfer{
		transferConnection: transferConnection,
	}
}

func (pdt *TransactionPaidDataTransfer) StartTransfer(ctx context.Context, holdingCode string, targetHoldingCode string) error {

	sourceRepository := NewTransactionPaidDataTransferRepository(pdt.transferConnection.GetSourceConnection())
	targetRepository := transactionPaidRepository.NewPaidRepository(pdt.transferConnection.GetTargetConnection())

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
