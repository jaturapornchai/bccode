package datatransfer

import (
	"context"
	"smlcloudplatform/internal/debtaccount/debtorgroup/models"
	debtorGroupRepositories "smlcloudplatform/internal/debtaccount/debtorgroup/repositories"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	msModels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DebtorGroupDataTransfer struct {
	transferConnection IDataTransferConnection
}

type IDebtorGroupDataTransferRepository interface {
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]models.DebtorGroupDoc, mongopagination.PaginationData, error)
}

type DebtorGroupDataTransferRepository struct {
	pst microservice.IPersisterMongo
	repositories.SearchRepository[models.DebtorGroupDoc]
}

func NewDebtorGroupDataTransferRepository(mongodbPersister microservice.IPersisterMongo) IDebtorGroupDataTransferRepository {
	repo := &DebtorGroupDataTransferRepository{
		pst: mongodbPersister,
	}

	repo.SearchRepository = repositories.NewSearchRepository[models.DebtorGroupDoc](mongodbPersister)
	return repo
}

func (repo DebtorGroupDataTransferRepository) FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]models.DebtorGroupDoc, mongopagination.PaginationData, error) {

	results, pagination, err := repo.SearchRepository.FindPage(ctx, shopID, searchInFields, pageable)

	if err != nil {
		return nil, mongopagination.PaginationData{}, err
	}

	return results, pagination, nil
}

func NewDebtorGroupDataTransfer(transferConnection IDataTransferConnection) IDataTransfer {

	return &DebtorGroupDataTransfer{
		transferConnection: transferConnection,
	}
}

func (pdt *DebtorGroupDataTransfer) StartTransfer(ctx context.Context, shopID string, targetShopID string) error {

	sourceRepository := NewDebtorGroupDataTransferRepository(pdt.transferConnection.GetSourceConnection())
	targetRepository := debtorGroupRepositories.NewDebtorGroupRepository(pdt.transferConnection.GetTargetConnection())

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
