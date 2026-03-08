package datatransfer

import (
	"context"
	"smlcloudplatform/internal/order/device/models"
	orderDeviceRepository "smlcloudplatform/internal/order/device/repositories"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	msModels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrderDeviceDataTransfer struct {
	transferConnection IDataTransferConnection
}

type IOrderDeviceDataTransferRepository interface {
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]models.OrderDeviceDoc, mongopagination.PaginationData, error)
}

type OrderDeviceDataTransferRepository struct {
	pst microservice.IPersisterMongo
	repositories.SearchRepository[models.OrderDeviceDoc]
}

func NewOrderDeviceDataTransferRepository(mongodbPersister microservice.IPersisterMongo) IOrderDeviceDataTransferRepository {
	repo := &OrderDeviceDataTransferRepository{
		pst: mongodbPersister,
	}

	repo.SearchRepository = repositories.NewSearchRepository[models.OrderDeviceDoc](mongodbPersister)
	return repo
}

func (repo OrderDeviceDataTransferRepository) FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]models.OrderDeviceDoc, mongopagination.PaginationData, error) {

	results, pagination, err := repo.SearchRepository.FindPage(ctx, shopID, searchInFields, pageable)

	if err != nil {
		return nil, mongopagination.PaginationData{}, err
	}

	return results, pagination, nil
}

func NewOrderDeviceDataTransfer(transferConnection IDataTransferConnection) IDataTransfer {
	return &OrderDeviceDataTransfer{
		transferConnection: transferConnection,
	}
}

func (dt *OrderDeviceDataTransfer) StartTransfer(ctx context.Context, shopID string, targetShopID string) error {

	sourceRepository := NewOrderDeviceDataTransferRepository(dt.transferConnection.GetSourceConnection())
	targetRepository := orderDeviceRepository.NewDeviceRepository(dt.transferConnection.GetTargetConnection())

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
