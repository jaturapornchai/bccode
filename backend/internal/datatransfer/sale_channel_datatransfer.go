package datatransfer

import (
	"context"
	"smlcloudplatform/internal/channel/salechannel/models"
	saleChannelRepository "smlcloudplatform/internal/channel/salechannel/repositories"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	msModels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SaleChannelDataTransfer struct {
	transferConnection IDataTransferConnection
}

type ISaleChannelDataTransferRepository interface {
	FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable msModels.Pageable) ([]models.SaleChannelDoc, mongopagination.PaginationData, error)
}

type SaleChannelDataTransferRepository struct {
	pst microservice.IPersisterMongo
	repositories.SearchRepository[models.SaleChannelDoc]
}

func NewSaleChannelDataTransferRepository(mongodbPersister microservice.IPersisterMongo) ISaleChannelDataTransferRepository {
	repo := &SaleChannelDataTransferRepository{
		pst: mongodbPersister,
	}

	repo.SearchRepository = repositories.NewSearchRepository[models.SaleChannelDoc](mongodbPersister)
	return repo
}

func (repo SaleChannelDataTransferRepository) FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable msModels.Pageable) ([]models.SaleChannelDoc, mongopagination.PaginationData, error) {

	results, pagination, err := repo.SearchRepository.FindPage(ctx, holdingCode, searchInFields, pageable)

	if err != nil {
		return nil, mongopagination.PaginationData{}, err
	}

	return results, pagination, nil
}

func NewSaleChannelDataTransfer(transferConnection IDataTransferConnection) IDataTransfer {
	return &SaleChannelDataTransfer{
		transferConnection: transferConnection,
	}
}

func (pdt *SaleChannelDataTransfer) StartTransfer(ctx context.Context, holdingCode string, targetHoldingCode string) error {

	sourceRepository := NewSaleChannelDataTransferRepository(pdt.transferConnection.GetSourceConnection())
	targetRepository := saleChannelRepository.NewSaleChannelRepository(pdt.transferConnection.GetTargetConnection())

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
