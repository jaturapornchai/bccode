package datatransfer

import (
	"context"
	"smlcloudplatform/internal/product/ordertype/models"
	orderTypeRepository "smlcloudplatform/internal/product/ordertype/repositories"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	msModels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrderTypeDataTransfer struct {
	transferConnection IDataTransferConnection
}

type IOrderTypeDataTransferRepository interface {
	FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable msModels.Pageable) ([]models.OrderTypeDoc, mongopagination.PaginationData, error)
}

type OrderTypeDataTransferRepository struct {
	pst microservice.IPersisterMongo
	repositories.SearchRepository[models.OrderTypeDoc]
}

func NewOrderTypeDataTransferRepository(mongodbPersister microservice.IPersisterMongo) IOrderTypeDataTransferRepository {
	repo := &OrderTypeDataTransferRepository{
		pst: mongodbPersister,
	}

	repo.SearchRepository = repositories.NewSearchRepository[models.OrderTypeDoc](mongodbPersister)
	return repo
}

func (repo OrderTypeDataTransferRepository) FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable msModels.Pageable) ([]models.OrderTypeDoc, mongopagination.PaginationData, error) {

	results, pagination, err := repo.SearchRepository.FindPage(ctx, holdingCode, searchInFields, pageable)

	if err != nil {
		return nil, mongopagination.PaginationData{}, err
	}

	return results, pagination, nil
}

func NewOrderTypeDataTransfer(transferConnection IDataTransferConnection) IDataTransfer {
	return &OrderTypeDataTransfer{
		transferConnection: transferConnection,
	}
}

func (pdt *OrderTypeDataTransfer) StartTransfer(ctx context.Context, holdingCode string, targetHoldingCode string) error {

	sourceRepository := NewOrderTypeDataTransferRepository(pdt.transferConnection.GetSourceConnection())
	targetRepository := orderTypeRepository.NewOrderTypeRepository(pdt.transferConnection.GetTargetConnection())

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
