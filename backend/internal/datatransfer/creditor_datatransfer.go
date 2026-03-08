package datatransfer

import (
	"context"
	"smlcloudplatform/internal/debtaccount/creditor/models"
	creditorRepositories "smlcloudplatform/internal/debtaccount/creditor/repositories"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	msModels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CreditorDataTransfer struct {
	transferConnection IDataTransferConnection
}

type ICreditorDataTransferRepository interface {
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]models.CreditorDoc, mongopagination.PaginationData, error)
}

type CreditorDataTransferRepository struct {
	pst microservice.IPersisterMongo
	repositories.SearchRepository[models.CreditorDoc]
}

func NewCreditorDataTransferRepository(mongodbPersister microservice.IPersisterMongo) ICreditorDataTransferRepository {
	repo := &CreditorDataTransferRepository{
		pst: mongodbPersister,
	}

	repo.SearchRepository = repositories.NewSearchRepository[models.CreditorDoc](mongodbPersister)
	return repo
}

func (repo CreditorDataTransferRepository) FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]models.CreditorDoc, mongopagination.PaginationData, error) {

	results, pagination, err := repo.SearchRepository.FindPage(ctx, shopID, searchInFields, pageable)

	if err != nil {
		return nil, mongopagination.PaginationData{}, err
	}

	return results, pagination, nil

}

func NewCreditorDataTransfer(transferConnection IDataTransferConnection) IDataTransfer {

	return &CreditorDataTransfer{
		transferConnection: transferConnection,
	}
}

func (pdt *CreditorDataTransfer) StartTransfer(ctx context.Context, shopID string, targetShopID string) error {

	sourceRepository := NewCreditorDataTransferRepository(pdt.transferConnection.GetSourceConnection())
	targetRepository := creditorRepositories.NewCreditorRepository(pdt.transferConnection.GetTargetConnection())

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
