package datatransfer

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/restaurant/staff/models"
	restaurantStaffRepositories "smlcloudplatform/internal/restaurant/staff/repositories"
	"smlcloudplatform/pkg/microservice"
	msModels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RestaurantStaffDataTransfer struct {
	transferConnection IDataTransferConnection
}

type IRestaurantStaffDataTransferRepository interface {
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]models.StaffDoc, mongopagination.PaginationData, error)
}

type RestaurantStaffDataTransferRepository struct {
	pst microservice.IPersisterMongo
	repositories.SearchRepository[models.StaffDoc]
}

func NewRestaurantStaffDataTransferRepository(mongodbPersister microservice.IPersisterMongo) IRestaurantStaffDataTransferRepository {
	repo := &RestaurantStaffDataTransferRepository{
		pst: mongodbPersister,
	}

	repo.SearchRepository = repositories.NewSearchRepository[models.StaffDoc](mongodbPersister)
	return repo
}

func (repo RestaurantStaffDataTransferRepository) FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]models.StaffDoc, mongopagination.PaginationData, error) {

	results, pagination, err := repo.SearchRepository.FindPage(ctx, shopID, searchInFields, pageable)

	if err != nil {
		return nil, mongopagination.PaginationData{}, err
	}

	return results, pagination, nil
}

func NewRestaurantStaffDataTransfer(transferConnection IDataTransferConnection) IDataTransfer {

	return &RestaurantStaffDataTransfer{
		transferConnection: transferConnection,
	}
}

func (pdt *RestaurantStaffDataTransfer) StartTransfer(ctx context.Context, shopID string, targetShopID string) error {

	sourceRepository := NewRestaurantStaffDataTransferRepository(pdt.transferConnection.GetSourceConnection())
	targetRepository := restaurantStaffRepositories.NewStaffRepository(pdt.transferConnection.GetTargetConnection())

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
