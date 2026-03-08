package datatransfer

import (
	"context"
	"smlcloudplatform/internal/repositories"
	restaurantRepository "smlcloudplatform/internal/restaurant/settings"
	"smlcloudplatform/internal/restaurant/settings/models"
	"smlcloudplatform/pkg/microservice"
	msModels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RestaurantSettingDataTransfer struct {
	transferConnection IDataTransferConnection
}

type IRestaurentSettingDataTransferRepository interface {
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]models.RestaurantSettingsDoc, mongopagination.PaginationData, error)
}

type RestaurentSettingDataTransferRepository struct {
	pst microservice.IPersisterMongo
	repositories.SearchRepository[models.RestaurantSettingsDoc]
}

func NewRestaurantSettingDataTransfer(transferConnection IDataTransferConnection) IDataTransfer {
	return &RestaurantSettingDataTransfer{
		transferConnection: transferConnection,
	}
}

func NewRestaurentSettingDataTransferRepository(mongodbPersister microservice.IPersisterMongo) IRestaurentSettingDataTransferRepository {
	repo := &RestaurentSettingDataTransferRepository{
		pst: mongodbPersister,
	}

	repo.SearchRepository = repositories.NewSearchRepository[models.RestaurantSettingsDoc](mongodbPersister)
	return repo
}

func (repo RestaurentSettingDataTransferRepository) FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]models.RestaurantSettingsDoc, mongopagination.PaginationData, error) {

	results, pagination, err := repo.SearchRepository.FindPage(ctx, shopID, searchInFields, pageable)

	if err != nil {
		return nil, mongopagination.PaginationData{}, err
	}

	return results, pagination, nil
}

func (sdt *RestaurantSettingDataTransfer) StartTransfer(ctx context.Context, shopID string, targetShopID string) error {

	sourceRepository := NewRestaurentSettingDataTransferRepository(sdt.transferConnection.GetSourceConnection())
	targetRepository := restaurantRepository.NewRestaurantSettingsRepository(sdt.transferConnection.GetTargetConnection())

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
