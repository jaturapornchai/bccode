package datatransfer

import (
	"context"
	"smlcloudplatform/internal/channel/transportchannel/models"
	transportChannelRepository "smlcloudplatform/internal/channel/transportchannel/repositories"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	msModels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SaleTransportDataTransfer struct {
	transferConnection IDataTransferConnection
}

type ISaleTransportDataTransferRepository interface {
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]models.TransportChannelDoc, mongopagination.PaginationData, error)
}

type SaleTransportDataTransferRepository struct {
	pst microservice.IPersisterMongo
	repositories.SearchRepository[models.TransportChannelDoc]
}

func NewSaleTransportDataTransferRepository(mongodbPersister microservice.IPersisterMongo) ISaleTransportDataTransferRepository {
	repo := &SaleTransportDataTransferRepository{
		pst: mongodbPersister,
	}

	repo.SearchRepository = repositories.NewSearchRepository[models.TransportChannelDoc](mongodbPersister)
	return repo
}

func (repo SaleTransportDataTransferRepository) FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]models.TransportChannelDoc, mongopagination.PaginationData, error) {

	results, pagination, err := repo.SearchRepository.FindPage(ctx, shopID, searchInFields, pageable)

	if err != nil {
		return nil, mongopagination.PaginationData{}, err
	}

	return results, pagination, nil
}

func NewSaleTransportDataTransfer(transferConnection IDataTransferConnection) IDataTransfer {
	return &SaleTransportDataTransfer{
		transferConnection: transferConnection,
	}
}

func (pdt *SaleTransportDataTransfer) StartTransfer(ctx context.Context, shopID string, targetShopID string) error {

	sourceRepository := NewSaleTransportDataTransferRepository(pdt.transferConnection.GetSourceConnection())
	targetRepository := transportChannelRepository.NewTransportChannelRepository(pdt.transferConnection.GetTargetConnection())

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
