package datatransfer

import (
	"context"
	"smlcloudplatform/internal/order/setting/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	msModels "smlcloudplatform/pkg/microservice/models"

	orderDeviceSettingRepository "smlcloudplatform/internal/order/setting/repositories"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrderDeviceSettingDataTransfer struct {
	transferConnection IDataTransferConnection
}

type IOrderDeviceSettingDataTransferRepository interface {
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]models.SettingDoc, mongopagination.PaginationData, error)
}

type OrderDeviceSettingDataTransferRepository struct {
	pst microservice.IPersisterMongo
	repositories.SearchRepository[models.SettingDoc]
}

func NewOrderDeviceSettingDataTransferRepository(mongodbPersister microservice.IPersisterMongo) IOrderDeviceSettingDataTransferRepository {
	repo := &OrderDeviceSettingDataTransferRepository{
		pst: mongodbPersister,
	}

	repo.SearchRepository = repositories.NewSearchRepository[models.SettingDoc](mongodbPersister)
	return repo
}

func (repo OrderDeviceSettingDataTransferRepository) FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]models.SettingDoc, mongopagination.PaginationData, error) {

	results, pagination, err := repo.SearchRepository.FindPage(ctx, shopID, searchInFields, pageable)

	if err != nil {
		return nil, mongopagination.PaginationData{}, err
	}

	return results, pagination, nil
}

func NewOrderDeviceSettingDataTransfer(transferConnection IDataTransferConnection) IDataTransfer {
	return &OrderDeviceSettingDataTransfer{
		transferConnection: transferConnection,
	}
}

func (dt *OrderDeviceSettingDataTransfer) StartTransfer(ctx context.Context, shopID string, targetShopID string) error {

	sourceRepository := NewOrderDeviceSettingDataTransferRepository(dt.transferConnection.GetSourceConnection())
	targetRepository := orderDeviceSettingRepository.NewSettingRepository(dt.transferConnection.GetTargetConnection())

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
