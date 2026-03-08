package datatransfer

import (
	"context"
	"smlcloudplatform/internal/debtaccount/debtor/models"
	debtorRepository "smlcloudplatform/internal/debtaccount/debtor/repositories"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	msModels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DebtorDataTransfer struct {
	transferConnection IDataTransferConnection
}

type IDebtorDataTransferRepository interface {
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]models.DebtorDoc, mongopagination.PaginationData, error)
}

type DebtorDataTransferRepository struct {
	pst microservice.IPersisterMongo
	repositories.SearchRepository[models.DebtorDoc]
}

func NewDebtorDataTransferRepository(mongodbPersister microservice.IPersisterMongo) IDebtorDataTransferRepository {
	repo := &DebtorDataTransferRepository{
		pst: mongodbPersister,
	}

	repo.SearchRepository = repositories.NewSearchRepository[models.DebtorDoc](mongodbPersister)
	return repo
}

func (repo DebtorDataTransferRepository) FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]models.DebtorDoc, mongopagination.PaginationData, error) {

	results, pagination, err := repo.SearchRepository.FindPage(ctx, shopID, searchInFields, pageable)

	if err != nil {
		return nil, mongopagination.PaginationData{}, err
	}

	return results, pagination, nil
}

func NewDebtorDataTransfer(transferConnection IDataTransferConnection) IDataTransfer {

	return &DebtorDataTransfer{
		transferConnection: transferConnection,
	}
}

func (pdt *DebtorDataTransfer) StartTransfer(ctx context.Context, shopID string, targetShopID string) error {

	sourceRepository := NewDebtorDataTransferRepository(pdt.transferConnection.GetSourceConnection())
	targetRepository := debtorRepository.NewDebtorRepository(pdt.transferConnection.GetTargetConnection())

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
